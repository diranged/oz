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
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ExecAccessRequestSpec defines the desired state of ExecAccessRequest
type ExecAccessRequestSpec struct {
	// Defines the name of the `ExecAcessTemplate` that should be used to grant access to the target
	// resource.
	//
	// +kubebuilder:validation:Required
	TemplateName string `json:"templateName"`

	// TargetPod is used to explicitly define the target pod that the Exec privilges should be
	// granted to. If not supplied, then a random pod is chosen.
	TargetPod string `json:"targetPod,omitempty"`

	// Duration sets the length of time from the `spec.creationTimestamp` that this object will live. After the
	// time has expired, the resouce will be automatically deleted on the next reconcilliation loop.
	//
	// If omitted, the spec.defautlDuration from the ExecAccessTemplate is used.
	//
	// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	Duration string `json:"duration,omitempty"`
}

// ExecAccessRequestStatus defines the observed state of ExecAccessRequest
type ExecAccessRequestStatus struct {
	// Current status of the Access Request
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`

	// The Target Pod Name where access has been granted
	PodName string `json:"podName,omitempty"`

	// The name of the Role created for this temporary access request
	RoleName string `json:"roleName,omitempty"`

	// The name of th RoleBinding created for this temporary access request
	RoleBindingName string `json:"roleBindingName,omitempty"`

	// Simple boolean to let us know if the resource is ready for use or not
	Ready bool `json:"ready,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ExecAccessRequest is the Schema for the execaccessrequests API
type ExecAccessRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ExecAccessRequestSpec   `json:"spec,omitempty"`
	Status ExecAccessRequestStatus `json:"status,omitempty"`
}

// GetDuration conforms to the interfaces.OzRequestResource interface
func (r *ExecAccessRequest) GetDuration() (time.Duration, error) {
	if r.Spec.Duration != "" {
		return time.ParseDuration(r.Spec.Duration)
	}
	return time.Duration(0), nil
}

// GetUptime conforms to the interfaces.OzRequestResource interface
func (r *ExecAccessRequest) GetUptime() time.Duration {
	now := time.Now()
	creation := r.CreationTimestamp.Time
	return now.Sub(creation)
}

// SetPodName conforms to the interfaces.OzRequestResource interface
func (r *ExecAccessRequest) SetPodName(name string) error {
	if r.Status.PodName != "" {
		return fmt.Errorf("Status.PodName arlready set: %s", r.Status.PodName)
	}
	r.Status.PodName = name
	return nil
}

// GetPodName conforms to the interfaces.OzRequestResource interface
func (r *ExecAccessRequest) GetPodName() string {
	return r.Status.PodName
}

// GetConditions returns a pointer to the list of conditions in the ExecAccessRequestStatus object.
//
// Conform to the interfaces.OzResource interface
func (r *ExecAccessRequest) GetConditions() *[]metav1.Condition {
	return &r.Status.Conditions
}

// IsReady conforms to the interfaces.OzResource interface
func (r *ExecAccessRequest) IsReady() bool {
	return r.Status.Ready
}

// SetReady conforms to the interfaces.OzResource interface
func (r *ExecAccessRequest) SetReady(ready bool) {
	r.Status.Ready = ready
}

// GetExecAccessRequest returns back an ExecAccessRequest resource matching the request supplied to
// the reconciler loop, or returns back an error.
func GetExecAccessRequest(ctx context.Context, cl client.Reader, name string, namespace string) (*ExecAccessRequest, error) {
	tmpl := &ExecAccessRequest{}
	err := cl.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, tmpl)
	return tmpl, err
}

//+kubebuilder:object:root=true

// ExecAccessRequestList contains a list of ExecAccessRequest
type ExecAccessRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ExecAccessRequest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ExecAccessRequest{}, &ExecAccessRequestList{})
}
