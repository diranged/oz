package builders

import (
	"fmt"
	"math/rand"

	api "github.com/diranged/oz/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ExecAccessBuilder implements the required resources for the api.ExecAccessTemplate CRD.
//
// An "ExecAccessRequest" is used to generate access that has been defined through an "ExecAccessTemplate".
//
// An "ExecAccessTemplate" allows a group to "kubectl exec" into an already running Pod in a
// specific Controller (DaemonSet, Deployment, StatefulSet). This privileged access is generally
// only used when it is critical to troubleshoot a live Pod that is serving a particular workload.
type ExecAccessBuilder struct {
	BaseBuilder

	Request  *api.ExecAccessRequest
	Template *api.ExecAccessTemplate
}

// https://stackoverflow.com/questions/33089523/how-to-mark-golang-struct-as-implementing-interface
var (
	_ IBuilder = &ExecAccessBuilder{}
	_ IBuilder = (*ExecAccessBuilder)(nil)
)

// GenerateAccessResources is the primary function called by the reconciler to this Builder object. This function
// is responsible for building all of the temporary access resources, and returning back information about them
// to the user. Any error causes this function to stop and fail.
//
// Returns:
//
//	statusString: A string representing the status of all of the resources created. This is applied to the
//	conditions of the AccessRequest by the reconciler loop.
//
//	accessString: A string representing how the end-user can use the resources. Eg: "kubectl exec ...". This
//	string may go away.
//
//	err: Any errors during the building and application of these resources.
func (b *ExecAccessBuilder) GenerateAccessResources() (statusString string, err error) {
	var accessString string

	// Get the target Pod Name that the user is going to have access to
	targetPodName, err := b.getPodName()
	if err != nil {
		return statusString, err
	}

	// Get the Role, or error out
	role, err := b.createAccessRole(targetPodName)
	if err != nil {
		return statusString, err
	}

	// Get the Binding, or error out
	rb, err := b.createAccessRoleBinding()
	if err != nil {
		return statusString, err
	}

	statusString = fmt.Sprintf("Success. Role %s, RoleBinding %s created", role.Name, rb.Name)
	accessString = fmt.Sprintf(
		"kubectl exec -ti -n %s %s -- /bin/sh",
		b.Template.Namespace,
		targetPodName,
	)

	// TODO: check err return from SetPodName
	_ = b.Request.SetPodName(targetPodName)
	b.Request.Status.SetAccessMessage(accessString)

	return statusString, err
}

// generatePodName is used to discover the target pod that the user is going to have access to. This
// function is designed to be idempotent - so once a podName has been selected, it will be used on
// each and every reconcile going forward.
//
//   - If status.podName is set? Return that value Else? Continue.
//   - If request.targetPod...
//     ... is set, call getSpecificPod() to verify that the pod exists and is valid for the request
//     ... is not set, call getRandomPod() to pick a random pod from the target controller
//   - Save the picked podName into the request status and update the request object

// Returns:
//
//	podname: A string with the pod name (or an empty string in a failure)
//	error: Any errors generating the podName.
func (b *ExecAccessBuilder) getPodName() (podName string, err error) {
	logger := log.FromContext(b.Ctx)

	// If this resource already has a status.podName field set, then we respect that no matter what.
	// We never mutate the pod that this access request was originally created for. Otherwise, pick
	// a Pod and populate that status field.
	if b.GetRequest().GetPodName() != "" {
		logger.Info(fmt.Sprintf("Pod already assigned - %s", b.GetRequest().GetPodName()))
		return b.GetRequest().GetPodName(), nil
	}

	// If the user supplied their own Pod, then get that Pod back to make sure it exists. Otherwise,
	// randomly select a pod.
	var pod *corev1.Pod
	if b.Request.Spec.TargetPod == "" {
		pod, err = b.getRandomPod()
		if err != nil {
			logger.Error(err, "Failed to retrieve Pod from ExecAccessTemplate")
			return "", err
		}
	} else {
		pod, err = b.getSpecificPod()

		// Informative for the operator for now. The verification step below truly let the user know about the problem.
		if err != nil {
			logger.Info("Error looking up Pod")
			return "", err
		}
	}

	// Set the podName (note, just in the local object). If this fails (for example, its already set
	// on the object), then we also bail out. This only fails if the Status.PodName field has already been set,
	// which would indicate some kind of a reconcile loop conflict.
	//
	// The responsibility of pushing the .Status.PodName field back to Kubernetes is in the reconciliation loop,
	// where it will call UpdateCondition (which calls UpdateStatus) at the end of this succesful method. In this
	// way, we do not update the AccessRequest with a PodName status until we have confidence that all of the access
	// resources have indeed been created.
	if err := b.Request.SetPodName(pod.Name); err != nil {
		return "", err
	}

	// Return the podName string.
	return pod.Name, nil
}

func (b *ExecAccessBuilder) getRandomPod() (*corev1.Pod, error) {
	// https://sdk.operatorframework.io/docs/building-operators/golang/references/logging/
	logger := log.FromContext(b.Ctx)
	logger.Info("Finding Pods...")

	// Will populate this further down
	pod := &corev1.Pod{}

	// https://medium.com/coding-kubernetes/using-k8s-label-selectors-in-go-the-right-way-733cde7e8630
	selector, err := b.getTargetPodSelectorLabels()
	if err != nil {
		logger.Error(err, "Failed to find label selector, cannot automatically discover pods")
		return nil, err
	}

	// List all of the pods in the Deployment by searching for matching pods with the current Label
	// Selector.
	podList := &corev1.PodList{}
	opts := []client.ListOption{
		client.InNamespace(b.Template.Namespace),
		client.MatchingLabelsSelector{
			Selector: selector,
		},
		// TODO: Figure this out...
		client.MatchingFields{"status.phase": "Running"},
	}
	if err := b.Client.List(b.Ctx, podList, opts...); err != nil {
		logger.Error(err, "Failed to retrieve Pod list")
		return nil, err
	}

	if len(podList.Items) < 1 {
		return nil, fmt.Errorf("no pods found maching selector")
	}

	// Randomly generate a number from within the length of the returned pod list...
	randomIndex := rand.Intn(len(podList.Items))

	// Return the randomly generated Pod
	logger.Info(fmt.Sprintf("Returning Pod %s", pod.Name))
	pod = &podList.Items[randomIndex]

	return pod, err
}

func (b *ExecAccessBuilder) getSpecificPod() (*corev1.Pod, error) {
	podName := b.Request.Spec.TargetPod

	// https://sdk.operatorframework.io/docs/building-operators/golang/references/logging/
	logger := log.FromContext(b.Ctx)
	logger.Info(fmt.Sprintf("Looking for Pod %s", podName))

	// https://medium.com/coding-kubernetes/using-k8s-label-selectors-in-go-the-right-way-733cde7e8630
	selector, err := b.getTargetPodSelectorLabels()
	if err != nil {
		logger.Error(err, "Failed to find label selector, cannot automatically discover pods")
		return nil, err
	}

	// List all of the pods in the Deployment by searching for matching pods with the current Label
	// Selector.
	podList := &corev1.PodList{}
	opts := []client.ListOption{
		client.InNamespace(b.Template.Namespace),
		client.MatchingLabelsSelector{
			Selector: selector,
		},
		client.MatchingFields{"metadata.name": podName, "status.phase": "Running"},
		// TODO: Figure this out...
		// client.MatchingFields{"status.phase": "Running"},
	}
	// if err := b.Client.List(b.Ctx, podList, opts...); err != nil {
	if err := b.APIReader.List(b.Ctx, podList, opts...); err != nil {
		logger.Error(err, "Failed to retrieve Pod list")
		return nil, err
	}
	if len(podList.Items) < 1 {
		return nil, fmt.Errorf("pod named %s not found", podName)
	}
	if len(podList.Items) > 1 {
		return nil, fmt.Errorf("multiple pods matching %s returned - critical failure", podName)
	}

	// Return the first element from the list
	return &podList.Items[0], err
}
