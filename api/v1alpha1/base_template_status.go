package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// BaseTemplateStatus is the core set of status fields that we expect to be in each and every one of
// our template (AccessTemplate, ExecAccessTemplate, etc) resources.
type BaseTemplateStatus struct {
	// Available refers to whether or not the ExecAccessTemplate resource has been validated and is
	// available for use.
	// Available bool `json:"available,omitempty"`

	// Conditions represent the latest state of the resource
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}
