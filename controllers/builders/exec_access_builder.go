package builders

import (
	"fmt"
	"math/rand"

	api "github.com/diranged/oz/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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
	*BaseBuilder

	Request  *api.ExecAccessRequest
	Template *api.ExecAccessTemplate
}

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
func (b *ExecAccessBuilder) GenerateAccessResources() (statusString string, accessString string, err error) {
	// Get the target Pod Name that the user is going to have access to
	targetPodName, err := b.generatePodName()
	if err != nil {
		return statusString, accessString, err
	}

	// Get the Role, or error out
	role, err := b.applyAccessRole(targetPodName)
	if err != nil {
		return statusString, accessString, err
	}

	// Get the Binding, or error out
	rb, err := b.applyAccessRoleBinding()
	if err != nil {
		return statusString, accessString, err
	}

	statusString = fmt.Sprintf("Success. Role %s, RoleBinding %s created", role.Name, rb.Name)
	accessString = fmt.Sprintf("kubectl exec -ti -n %s %s -- /bin/sh", b.Template.Namespace, "asdf")

	return statusString, accessString, err
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
//
// Returns:
//
//	podname: A string with the pod name (or an empty string in a failure)
//	error: Any errors generating the podName.
func (b *ExecAccessBuilder) generatePodName() (podName string, err error) {
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
		//client.MatchingFields{"status.phase": "Running"},
	}
	//if err := b.Client.List(b.Ctx, podList, opts...); err != nil {
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

func (b *ExecAccessBuilder) applyAccessRole(podName string) (*rbacv1.Role, error) {
	role := &rbacv1.Role{}

	role.Name = fmt.Sprintf("%s-%s", b.Request.Name, b.Request.GetShortUID())
	role.Namespace = b.Template.Namespace
	role.Rules = []rbacv1.PolicyRule{
		{
			APIGroups:     []string{corev1.GroupName},
			Resources:     []string{"pods"},
			ResourceNames: []string{podName},
			Verbs:         []string{"get", "list", "watch"},
		},
		{
			APIGroups:     []string{corev1.GroupName},
			Resources:     []string{"pods/exec"},
			ResourceNames: []string{podName},
			Verbs:         []string{"create", "update", "delete", "get", "list"},
		},
	}

	// Set the ownerRef for the Deployment
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/owners-dependents/
	if err := ctrlutil.SetControllerReference(b.Request, role, b.Scheme); err != nil {
		return nil, err
	}

	// Generate an empty role resource. This role resource will be filled-in by the CreateOrUpdate() call when
	// it checks the Kubernetes API for the existing role. Our update function will then update the appropriate
	// values from the desired role object above.
	emptyRole := &rbacv1.Role{ObjectMeta: metav1.ObjectMeta{Name: role.Name, Namespace: role.Namespace}}

	// https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/controller/controllerutil#CreateOrUpdate
	if _, err := ctrlutil.CreateOrUpdate(b.Ctx, b.Client, emptyRole, func() error {
		emptyRole.ObjectMeta = role.ObjectMeta
		emptyRole.Rules = role.Rules
		emptyRole.OwnerReferences = role.OwnerReferences
		return nil
	}); err != nil {
		return nil, err
	}

	return role, nil
}

func (b *ExecAccessBuilder) applyAccessRoleBinding() (*rbacv1.RoleBinding, error) {
	rb := &rbacv1.RoleBinding{}

	rb.Name = fmt.Sprintf("%s-%s", b.Request.Name, b.Request.GetShortUID())
	rb.Namespace = b.Template.Namespace
	rb.RoleRef = rbacv1.RoleRef{
		APIGroup: rbacv1.GroupName,
		Kind:     "Role",
		Name:     rb.Name,
	}
	rb.Subjects = []rbacv1.Subject{}

	for _, group := range b.Template.Spec.AllowedGroups {
		rb.Subjects = append(rb.Subjects, rbacv1.Subject{
			APIGroup: rbacv1.SchemeGroupVersion.Group,
			Kind:     rbacv1.GroupKind,
			Name:     group,
		})
	}

	// Set the ownerRef for the Deployment
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/owners-dependents/
	if err := ctrlutil.SetControllerReference(b.Request, rb, b.Scheme); err != nil {
		return nil, err
	}

	// Generate an empty role resource. This role resource will be filled-in by the CreateOrUpdate() call when
	// it checks the Kubernetes API for the existing role. Our update function will then update the appropriate
	// values from the desired role object above.
	emptyRb := &rbacv1.RoleBinding{ObjectMeta: metav1.ObjectMeta{Name: rb.Name, Namespace: rb.Namespace}}

	// https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/controller/controllerutil#CreateOrUpdate
	if _, err := ctrlutil.CreateOrUpdate(b.Ctx, b.Client, emptyRb, func() error {
		emptyRb.ObjectMeta = rb.ObjectMeta
		emptyRb.RoleRef = rb.RoleRef
		emptyRb.Subjects = rb.Subjects
		emptyRb.OwnerReferences = rb.OwnerReferences
		return nil
	}); err != nil {
		return nil, err
	}

	return rb, nil
}
