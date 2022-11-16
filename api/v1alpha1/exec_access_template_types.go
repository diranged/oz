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

type TemplateConditionTypes string

const (
	ConditionTargetRefVerified TemplateConditionTypes = "TargetReferenceVerified"
)

// ExecAccessTemplateSpec defines the desired state of ExecAccessTemplate
type ExecAccessTemplateSpec struct {
	// TargetRef provides a pattern for referencing objects from another API in a generic way.
	// +kubebuilder:validation:Required
	TargetRef CrossVersionObjectReference `json:"targetRef"`

	// AllowedGroups lists out the groups (in string name form) that will be allowed to Exec into
	// the target pod.
	//
	// +kubebuilder:validation:Required
	AllowedGroups []string `json:"allowedGroups"`

	// DefaultDuration sets the default time that an `ExecAccessRequest` resource will live. Must
	// be set below MaxDuration.
	//
	// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	//
	// +kubebuilder:default:="1h"
	DefaultDuration string `json:"defaultDuration"`

	// MaxDuration sets the maximum duration that an `ExecAccessRequest` resource can request to
	// stick around.
	//
	// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	//
	// +kubebuilder:default:="24h"
	MaxDuration string `json:"maxDuration"`
}

// ExecAccessTemplateStatus is the core set of status fields that we expect to be in each and every one of
// our template (AccessTemplate, ExecAccessTemplate, etc) resources.
type ExecAccessTemplateStatus struct {
	// Available refers to whether or not the ExecAccessTemplate resource has been validated and is
	// available for use.
	// Available bool `json:"available,omitempty"`

	// Conditions represent the latest state of the resource
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`

	// Simple boolean to let us know if the resource is ready for use or not
	Ready bool `json:"ready,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ExecAccessTemplate is the Schema for the execaccesstemplates API
type ExecAccessTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ExecAccessTemplateSpec   `json:"spec,omitempty"`
	Status ExecAccessTemplateStatus `json:"status,omitempty"`
}

// Conform to the controllers.OzResource interface.
func (t *ExecAccessTemplate) GetConditions() *[]metav1.Condition {
	return &t.Status.Conditions
}

// Conform to the interfaces.OzResource interface
func (t *ExecAccessTemplate) IsReady() bool {
	return t.Status.Ready
}

// Conform to the interfaces.OzResource interface
func (t *ExecAccessTemplate) SetReady(ready bool) {
	t.Status.Ready = ready
}

// Conform to the controllers.OzTemplateResource interface.
func (t *ExecAccessTemplate) GetTargetRef() *CrossVersionObjectReference {
	return &t.Spec.TargetRef
}

// TODO: Decide if this is good
func (t *ExecAccessTemplate) GetDefaultDuration() (time.Duration, error) {
	return time.ParseDuration(t.Spec.DefaultDuration)
}

func (t *ExecAccessTemplate) GetMaxDuration() (time.Duration, error) {
	return time.ParseDuration(t.Spec.MaxDuration)
}

// GetResource returns back an ExecAccessTemplate resource matching the request supplied to the reconciler loop, or
// returns back an error.
func GetExecAccessTemplate(cl client.Client, ctx context.Context, name string, namespace string) (*ExecAccessTemplate, error) {
	tmpl := &ExecAccessTemplate{}
	err := cl.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, tmpl)
	return tmpl, err
}

//+kubebuilder:object:root=true

// ExecAccessTemplateList contains a list of ExecAccessTemplate
type ExecAccessTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ExecAccessTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ExecAccessTemplate{}, &ExecAccessTemplateList{})
}
