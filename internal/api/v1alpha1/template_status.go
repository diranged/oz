package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TemplateStatus provides a common set of .Status fields and functions. The goal is to
// conform to the interfaces.OzResource interface commonly across all of our core CRDs.
type TemplateStatus struct {
	// Current status of the Access Template
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`

	// Simple boolean to let us know if the resource is ready for use or not
	Ready bool `json:"ready,omitempty"`
}

// DeepCopyInto is typically auto-generated by controller-gen. However, it seems that controller-gen
// fails when we include the ozResourceTemplateStatus.Conditions field. Implementing our own DeepCopyInto function
// resolves this, but does put the responsibility on us to keep this updated.
func (in *TemplateStatus) DeepCopyInto(out *TemplateStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]metav1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}
