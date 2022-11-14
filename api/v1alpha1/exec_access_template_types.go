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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ExecAccessTemplate is the Schema for the execaccesstemplates API
type ExecAccessTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ExecAccessTemplateSpec `json:"spec,omitempty"`
	Status BaseTemplateStatus     `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ExecAccessTemplateList contains a list of ExecAccessTemplate
type ExecAccessTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ExecAccessTemplate `json:"items"`
}

func (t *ExecAccessTemplate) GetDefaultDuration() (time.Duration, error) {
	return time.ParseDuration(t.Spec.DefaultDuration)
}

func (t *ExecAccessTemplate) GetMaxDuration() (time.Duration, error) {
	return time.ParseDuration(t.Spec.MaxDuration)
}

func init() {
	SchemeBuilder.Register(&ExecAccessTemplate{}, &ExecAccessTemplateList{})
}
