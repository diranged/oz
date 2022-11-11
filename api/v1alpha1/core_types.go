package v1alpha1

type Kind string

const (
	Deployment  Kind = "Deployment"
	DaemonSet   Kind = "DaemonSet"
	StatefulSet Kind = "StatefulSet"
)

// Important: Run "make" to regenerate code after modifying this file
type CrossVersionObjectReference struct {
	// Defines the "APIVersion" of the resource being referred to. Eg, "apps/v1".
	// +kubebuilder:validation:Required
	APIVersion *string `json:"apiVersion"`

	// Defines the "Kind" of resource being referred to.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=Deployment;DaemonSet;StatefulSet
	Kind Kind `json:"kind"`

	// Defines the "metadata.name" of the target resource.
	// +kubebuilder:validation:Required
	Name *string `json:"name"`
}
