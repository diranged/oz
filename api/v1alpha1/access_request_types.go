/*
Copyright 2022 Matt Wise.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AccessRequestSpec defines the desired state of AccessRequest
type AccessRequestSpec struct {
	// Defines the name of the `ExecAcessTemplate` that should be used to grant access to the target
	// resource.
	//
	// +kubebuilder:validation:Required
	TemplateName string `json:"templateName"`

	// Duration sets the length of time from the `spec.creationTimestamp` that this object will live. After the
	// time has expired, the resouce will be automatically deleted on the next reconcilliation loop.
	//
	// If omitted, the spec.defautlDuration from the ExecAccessTemplate is used.
	//
	// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	Duration string `json:"duration,omitempty"`
}

// AccessRequestStatus defines the observed state of AccessRequest
type AccessRequestStatus struct {
	ozResourceCoreStatus `json:",inline"`

	// The Target Pod Name where access has been granted
	PodName string `json:"podName,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AccessRequest is the Schema for the accessrequests API
type AccessRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AccessRequestSpec   `json:"spec,omitempty"`
	Status AccessRequestStatus `json:"status,omitempty"`
}

// GetDuration conform to the interfaces.OzRequestResource interface
func (t *AccessRequest) GetDuration() (time.Duration, error) {
	if t.Spec.Duration != "" {
		return time.ParseDuration(t.Spec.Duration)
	}
	return time.Duration(0), nil
}

// GetUptime conform to the interfaces.OzRequestResource interface
func (t *AccessRequest) GetUptime() time.Duration {
	now := time.Now()
	creation := t.CreationTimestamp.Time
	return now.Sub(creation)
}

// GetConditions returns back a pointer to the list of conditions in the ExecAccessRequestStatus
// object.
//
// Conform to the interfaces.OzResource interface
func (t *AccessRequest) GetConditions() *[]metav1.Condition {
	return &t.Status.Conditions
}

// IsReady conforms to the interfaces.OzResource interface
func (t *AccessRequest) IsReady() bool {
	return t.Status.Ready
}

// SetReady conforms to the interfaces.OzResource interface
func (t *AccessRequest) SetReady(ready bool) {
	t.Status.Ready = ready
}

// SetPodName conforms to the interfaces.OzRequestResource interface
func (t *AccessRequest) SetPodName(name string) error {
	t.Status.PodName = name
	return nil
}

// GetPodName returns the PodName that has been assigned to the Status field within this AccessRequest.
func (t *AccessRequest) GetPodName() string {
	return t.Status.PodName
}

// GetAccessRequest returns back an ExecAccessRequest resource matching the request supplied to the
// reconciler loop, or returns back an error.
func GetAccessRequest(ctx context.Context, cl client.Reader, name string, namespace string) (*AccessRequest, error) {
	tmpl := &AccessRequest{}
	err := cl.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, tmpl)
	return tmpl, err
}

//+kubebuilder:object:root=true

// AccessRequestList contains a list of AccessRequest
type AccessRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AccessRequest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AccessRequest{}, &AccessRequestList{})
}
