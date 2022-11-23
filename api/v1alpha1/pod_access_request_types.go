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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PodAccessRequestSpec defines the desired state of AccessRequest
type PodAccessRequestSpec struct {
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
	// Valid time units are "s", "m", "h".
	//
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern="^[0-9]+(s|m|h)$"
	Duration string `json:"duration,omitempty"`
}

// PodAccessRequestStatus defines the observed state of AccessRequest
type PodAccessRequestStatus struct {
	CoreStatus `json:",inline"`

	// The Target Pod Name where access has been granted
	PodName string `json:"podName,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// PodAccessRequest is the Schema for the accessrequests API
//
// +kubebuilder:printcolumn:name="Ready",type="boolean",JSONPath=".status.ready",description="Is request ready?"
// +kubebuilder:printcolumn:name="Template",type="string",JSONPath=".spec.templateName",description="Access Template"
// +kubebuilder:printcolumn:name="Pod",type="string",JSONPath=".status.podName",description="Target Pod Name"
type PodAccessRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PodAccessRequestSpec   `json:"spec,omitempty"`
	Status PodAccessRequestStatus `json:"status,omitempty"`
}

// https://stackoverflow.com/questions/33089523/how-to-mark-golang-struct-as-implementing-interface
var (
	_ IPodRequestResource = &PodAccessRequest{}
	_ IPodRequestResource = (*PodAccessRequest)(nil)
)

// GetStatus returns the core Status field for this resource.
//
// Returns:
//
//	AccessRequestStatus
func (r *PodAccessRequest) GetStatus() ICoreStatus {
	return &r.Status
}

// GetTemplate returns a populated PodAccessTemplate that this PodAccessRequest is referencing.
func (r *PodAccessRequest) GetTemplate(
	ctx context.Context,
	cl client.Client,
) (ITemplateResource, error) {
	return GetPodAccessTemplate(ctx, cl, r.Spec.TemplateName, r.Namespace)
}

// GetTemplateName returns the user supplied Spec.templateName field
func (r *PodAccessRequest) GetTemplateName() string {
	return r.Spec.TemplateName
}

// GetDuration conform to the interfaces.OzRequestResource interface
func (r *PodAccessRequest) GetDuration() (time.Duration, error) {
	if r.Spec.Duration != "" {
		return time.ParseDuration(r.Spec.Duration)
	}
	return time.Duration(0), nil
}

// GetUptime conform to the interfaces.OzRequestResource interface
func (r *PodAccessRequest) GetUptime() time.Duration {
	now := time.Now()
	creation := r.CreationTimestamp.Time
	return now.Sub(creation)
}

// SetPodName conforms to the interfaces.OzRequestResource interface
func (r *PodAccessRequest) SetPodName(name string) error {
	if r.Status.PodName != "" {
		return fmt.Errorf(
			"immutable field Status.PodName already set (%s), cannot update to %s",
			r.Status.PodName,
			name,
		)
	}
	r.Status.PodName = name
	return nil
}

// GetPodName returns the PodName that has been assigned to the Status field within this AccessRequest.
func (r *PodAccessRequest) GetPodName() string {
	return r.Status.PodName
}

// GetPodAccessRequest returns back an ExecAccessRequest resource matching the request supplied to the
// reconciler loop, or returns back an error.
func GetPodAccessRequest(
	ctx context.Context,
	cl client.Reader,
	name string,
	namespace string,
) (*PodAccessRequest, error) {
	tmpl := &PodAccessRequest{}
	err := cl.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, tmpl)
	return tmpl, err
}

//+kubebuilder:object:root=true

// PodAccessRequestList contains a list of AccessRequest
type PodAccessRequestList struct {
	metav1.TypeMeta `                   json:",inline"`
	metav1.ListMeta `                   json:"metadata,omitempty"`
	Items           []PodAccessRequest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PodAccessRequest{}, &PodAccessRequestList{})
}
