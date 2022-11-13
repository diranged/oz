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

func init() {
	SchemeBuilder.Register(&ExecAccessTemplate{}, &ExecAccessTemplateList{})
}
