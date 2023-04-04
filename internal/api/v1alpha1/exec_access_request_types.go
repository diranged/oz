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
	CoreStatus `json:",inline"`

	// The Target Pod Name where access has been granted
	PodName string `json:"podName,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ExecAccessRequest is the Schema for the execaccessrequests API
//
// +kubebuilder:printcolumn:name="Template",type="string",JSONPath=".spec.templateName",description="Access Template"
// +kubebuilder:printcolumn:name="Pod",type="string",JSONPath=".status.podName",description="Target Pod Name"
// +kubebuilder:printcolumn:name="Ready",type="boolean",JSONPath=".status.ready",description="Is request ready?"
type ExecAccessRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ExecAccessRequestSpec   `json:"spec,omitempty"`
	Status ExecAccessRequestStatus `json:"status,omitempty"`
}

// https://stackoverflow.com/questions/33089523/how-to-mark-golang-struct-as-implementing-interface
var (
	_ IPodRequestResource = &ExecAccessRequest{}
	_ IPodRequestResource = (*ExecAccessRequest)(nil)
)

// GetStatus implements the ICoreResource interface
func (r *ExecAccessRequest) GetStatus() ICoreStatus {
	return &r.Status
}

// GetTemplate returns a populated ExecAccessTemplate that this ExecAccessRequest is referencing.
func (r *ExecAccessRequest) GetTemplate(
	ctx context.Context,
	cl client.Client,
) (ITemplateResource, error) {
	return GetExecAccessTemplate(ctx, cl, r.Spec.TemplateName, r.Namespace)
}

// GetTemplateName returns the user supplied Spec.templateName field
func (r *ExecAccessRequest) GetTemplateName() string {
	return r.Spec.TemplateName
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

// GetAccessCommand conforms to the interfaces.OzRequestResource interface
func (r *ExecAccessRequest) GetAccessCommand() string {
	return r.Status.AccessMessage
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

// GetExecAccessRequest returns back an ExecAccessRequest resource matching the request supplied to
// the reconciler loop, or returns back an error.
func GetExecAccessRequest(
	ctx context.Context,
	cl client.Client,
	name string,
	namespace string,
) (*ExecAccessRequest, error) {
	tmpl := &ExecAccessRequest{}
	err := cl.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, tmpl)
	return tmpl, err
}

//+kubebuilder:object:root=true

// ExecAccessRequestList contains a list of ExecAccessRequest
type ExecAccessRequestList struct {
	metav1.TypeMeta `                    json:",inline"`
	metav1.ListMeta `                    json:"metadata,omitempty"`
	Items           []ExecAccessRequest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ExecAccessRequest{}, &ExecAccessRequestList{})
}
