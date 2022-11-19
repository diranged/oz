package builders

import (
	"fmt"

	api "github.com/diranged/oz/api/v1alpha1"
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

// GeneratePodName returns back the PodName field which will be populated into the AccessRequest.
//
// TODO: GeneratePodName needs to figure out the PodName after it has created the target pod in the first place? Or
// it could just generate a static name with a clean function and return that.
func (b *AccessBuilder) generatePodName() (string, error) {
	return "junk", nil
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
