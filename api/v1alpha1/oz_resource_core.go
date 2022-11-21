package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type ozResourceCore struct {
	Status ozResourceCoreStatus `json:"status,omitempty"`
}

// GetConditions conforms to the interfaces.OzResource interface.
func (c *ozResourceCore) GetConditions() *[]metav1.Condition {
	return c.Status.GetConditions()
}

// IsReady conforms to the interfaces.OzResource interface
func (c *ozResourceCore) IsReady() bool {
	return c.Status.IsReady()
}

// SetReady conforms to the interfaces.OzResource interface
func (c *ozResourceCore) SetReady(ready bool) {
	c.Status.SetReady(ready)
}

// DeepCopyInto is a noop. We hold no real data in this ozResourceCore struct, just common functions
// we want to apply to all of our core resources.
func (c *ozResourceCore) DeepCopyInto(out *ozResourceCore) {
	*out = *c
}
