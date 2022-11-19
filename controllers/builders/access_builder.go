package builders

// AccessBuilder implements the required resources for the api.AccessTemplate CRD.
//
// An "AccessRequest" is used to generate access that has been defined through an "AccessTemplate".
//
// An "AccessTemplate" defines a mode of access into a Pod by which a PodSpec is copied out of an
// existing Deployment (or StatefulSet, DaemonSet), mutated so that the Pod is not in the path of
// live traffic, and then Role and RoleBindings are created to grant the developer access into the
// Pod.
type AccessBuilder struct {
	*BaseBuilder
}

// GeneratePodName returns back the PodName field which will be populated into the AccessRequest.
//
// TODO: GeneratePodName needs to figure out the PodName after it has created the target pod in the first place? Or
// it could just generate a static name with a clean function and return that.
func (t *AccessBuilder) GeneratePodName() (string, error) {
	return "junk", nil
}
