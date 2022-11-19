package builders

import (
	"fmt"

	api "github.com/diranged/oz/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// AccessBuilder implements the required resources for the api.AccessTemplate CRD.
//
// An "AccessRequest" is used to generate access that has been defined through an "AccessTemplate".
//
// An "AccessTemplate" defines a mode of access into a Pod by which a PodSpec is copied out of an
// existing Deployment (or StatefulSet, DaemonSet), mutated so that the Pod is not in the path of
// live traffic, and then Role and RoleBindings are created to grant the developer access into the
// Pod.
type AccessBuilder struct {
	BaseBuilder

	Request  *api.AccessRequest
	Template *api.AccessTemplate
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
func (b *AccessBuilder) GenerateAccessResources() (statusString string, accessString string, err error) {
	logger := log.FromContext(b.Ctx)

	// First, get the desired PodSpec. If there's a failure at this point, return it.
	podSpec, err := b.generatePodSpec()
	if err != nil {
		logger.Error(err, "Failed to generate PodSpec for AccessRequest")
		return statusString, accessString, err
	}

	// Generate a Pod for the user to access
	pod, err := b.createPod(podSpec)
	if err != nil {
		logger.Error(err, "Failed to create Pod for AccessRequest")
		return statusString, accessString, err
	}

	// Get the Role, or error out
	role, err := b.createAccessRole(pod.GetName())
	if err != nil {
		return statusString, accessString, err
	}

	// Get the Binding, or error out
	rb, err := b.createAccessRoleBinding()
	if err != nil {
		return statusString, accessString, err
	}

	statusString = fmt.Sprintf("Success. Pod %s, Role %s, RoleBinding %s created", pod.Name, role.Name, rb.Name)
	accessString = fmt.Sprintf("kubectl exec -ti -n %s %s -- /bin/sh", pod.GetNamespace(), pod.GetName())

	b.Request.Status.PodName = pod.GetName()

	return statusString, accessString, err
}

func (b *AccessBuilder) generatePodSpec() (corev1.PodSpec, error) {
	return b.getPodSpecFromController()
}
