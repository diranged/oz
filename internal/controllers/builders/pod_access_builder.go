package builders

import (
	"errors"
	"fmt"

	api "github.com/diranged/oz/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// PodAccessBuilder implements the required resources for the api.AccessTemplate CRD.
//
// An "AccessRequest" is used to generate access that has been defined through an "AccessTemplate".
//
// An "AccessTemplate" defines a mode of access into a Pod by which a PodSpec is copied out of an
// existing Deployment (or StatefulSet, DaemonSet), mutated so that the Pod is not in the path of
// live traffic, and then Role and RoleBindings are created to grant the developer access into the
// Pod.
type PodAccessBuilder struct {
	BaseBuilder

	Request  *api.PodAccessRequest
	Template *api.PodAccessTemplate
}

// https://stackoverflow.com/questions/33089523/how-to-mark-golang-struct-as-implementing-interface
var (
	_ IBuilder = &PodAccessBuilder{}
	_ IBuilder = (*PodAccessBuilder)(nil)
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
func (b *PodAccessBuilder) GenerateAccessResources() (statusString string, err error) {
	logger := log.FromContext(b.Ctx)
	var accessString string

	// First, get the desired PodSpec. If there's a failure at this point, return it.
	podTemplateSpec, err := b.generatePodTemplateSpec()
	if err != nil {
		logger.Error(err, "Failed to generate PodSpec for PodAccessRequest")
		return statusString, err
	}

	// Run the PodSpec through the optional mutation config
	mutator := b.Template.Spec.ControllerTargetMutationConfig
	podTemplateSpec, err = mutator.PatchPodTemplateSpec(b.Ctx, podTemplateSpec)
	if err != nil {
		logger.Error(err, "Failed to mutate PodSpec for PodAccessRequest")
		return statusString, err
	}

	// Generate a Pod for the user to access
	pod, err := b.createPod(podTemplateSpec)
	if err != nil {
		logger.Error(err, "Failed to create Pod for AccessRequest")
		return statusString, err
	}

	// Get the Role, or error out
	role, err := b.createAccessRole(pod.GetName())
	if err != nil {
		return statusString, err
	}

	// Get the Binding, or error out
	rb, err := b.createAccessRoleBinding()
	if err != nil {
		return statusString, err
	}

	statusString = fmt.Sprintf(
		"Success. Pod %s, Role %s, RoleBinding %s created",
		pod.Name,
		role.Name,
		rb.Name,
	)
	accessString = fmt.Sprintf(
		"kubectl exec -ti -n %s %s -- /bin/sh",
		pod.GetNamespace(),
		pod.GetName(),
	)

	// TODO: check err return from SetPodName
	_ = b.Request.SetPodName(pod.GetName())
	b.Request.Status.SetAccessMessage(accessString)

	return statusString, err
}

// VerifyAccessResources verifies that the Pod created in the
// GenerateAccessResources() function is up and in the "Running" phase.
func (b *PodAccessBuilder) VerifyAccessResources() (statusString string, err error) {
	// First, verify whether or not the PodName field has been set. If not,
	// then some part of the reconciliation has previously failed.
	if b.Request.GetPodName() == "" {
		return "No Pod Assigned Yet", errors.New("status.podName not yet set")
	}

	// Next, get the Pod. If the pod-get fails, then we need to return that failure.
	pod := &corev1.Pod{}
	err = b.APIReader.Get(b.Ctx, types.NamespacedName{
		Name:      b.Request.GetPodName(),
		Namespace: b.Request.Namespace,
	}, pod)
	if err != nil {
		return "Error Fetching Pod", err
	}

	// Now, check the Pod ready status
	if pod.Status.Phase != corev1.PodRunning {
		statusMsg := fmt.Sprintf("Pod in %s Phase", pod.Status.Phase)
		return statusMsg, errors.New(statusMsg)
	}

	// Finally, return the pod phase
	return fmt.Sprintf("Pod is %s", pod.Status.Phase), nil
}

func (b *PodAccessBuilder) generatePodTemplateSpec() (corev1.PodTemplateSpec, error) {
	return b.getPodTemplateFromController()
}
