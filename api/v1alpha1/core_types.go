package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type ControllerKind string

const (
	DeploymentController  ControllerKind = "Deployment"
	DaemonSetController   ControllerKind = "DaemonSet"
	StatefulSetController ControllerKind = "StatefulSet"
)

const (
	// TemplateAvailability is the string used for the primary status condition that indicates
	// whether or not an `AccessTemplate` or `ExecAccessTemplate` is ready for use.
	TemplateAvailability = "TemplateAvailable"

	// TemplateAvailabilityStatusAvailable represents the status of the Template when it is healthy and ready to use.
	TemplateAvailabilityStatusAvailable = "Available"

	// TemplateAvailabilityStatusDegraded indicates that the Template is unable to be used
	TemplateAvailabilityStatusDegraded = "Degraded"

	RequestValidated        = "RequestValidated"
	RequestValidatedSuccess = "Success"
	RequestValidatedFailed  = "Failed"
)

// Important: Run "make" to regenerate code after modifying this file
type CrossVersionObjectReference struct {
	// Defines the "APIVersion" of the resource being referred to. Eg, "apps/v1".
	// +kubebuilder:validation:Required
	APIVersion *string `json:"apiVersion"`

	// Defines the "Kind" of resource being referred to.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=Deployment;DaemonSet;StatefulSet
	Kind ControllerKind `json:"kind"`

	// Defines the "metadata.name" of the target resource.
	// +kubebuilder:validation:Required
	Name *string `json:"name"`
}

// BaseTemplateStatus is the core set of status fields that we expect to be in each and every one of
// our template (AccessTemplate, ExecAccessTemplate, etc) resources.
type BaseTemplateStatus struct {
	// Available refers to whether or not the ExecAccessTemplate resource has been validated and is
	// available for use.
	// Available bool `json:"available,omitempty"`

	// Conditions represent the latest state of the resource
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}
