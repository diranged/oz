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
	"errors"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// PodAccessTemplateSpec defines the desired state of AccessTemplate
type PodAccessTemplateSpec struct {
	// AccessConfig provides a common struct for defining who has access to the resources this
	// template controls, how long they have access, etc.
	AccessConfig AccessConfig `json:"accessConfig"`

	// ControllerTargetRef provides a pattern for referencing objects from another API in a generic way.
	//
	// +kubebuilder:validation:Optional
	ControllerTargetRef *CrossVersionObjectReference `json:"controllerTargetRef"`

	// ControllerTargetMutationConfig contains parameters that allow for customizing the copy of a
	// controller-sourced PodSpec. This setting is only valid if controllerTargetRef is set.
	//
	// +kubebuilder:validation:Optional
	ControllerTargetMutationConfig *PodTemplateSpecMutationConfig `json:"controllerTargetMutationConfig,omitempty"`

	// PodSpec ...
	//
	// +kubebuilder:validation:Optional
	PodSpec *corev1.PodSpec `json:"podSpec,omitempty"`

	// Upper bound of the ephemeral storage that an AccessRequest can make against this template for
	// the primary container.
	//
	// +kubebuilder:validation:Optional
	MaxStorage resource.Quantity `json:"maxStorage,omitempty"`

	// Upper bound of the CPU that an AccessRequest can make against this tmemplate for the primary container.
	//
	// +kubebuilder:validation:Optional
	MaxCPU resource.Quantity `json:"maxCpu,omitempty"`

	// Upper bound of the memory that an AccessRequest can make against this template for the primary container.
	//
	// +kubebuilder:validation:Optional
	MaxMemory resource.Quantity `json:"maxMemory,omitempty"`
}

// PodAccessTemplateStatus defines the observed state of PodAccessTemplate
type PodAccessTemplateStatus struct {
	TemplateStatus `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// PodAccessTemplate is the Schema for the accesstemplates API
//
// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Ready",type="boolean",JSONPath=".status.ready",description="Is template ready?"
type PodAccessTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PodAccessTemplateSpec   `json:"spec,omitempty"`
	Status PodAccessTemplateStatus `json:"status,omitempty"`
}

// https://stackoverflow.com/questions/33089523/how-to-mark-golang-struct-as-implementing-interface
var (
	_ ITemplateResource = &PodAccessTemplate{}
	_ ITemplateResource = (*PodAccessTemplate)(nil)
)

// GetConditions implements the ICoreResource Interface
func (t *PodAccessTemplate) GetConditions() *[]metav1.Condition {
	return &t.Status.Conditions
}

// IsReady implements the ICoreResource Interface
func (t *PodAccessTemplate) IsReady() bool {
	return t.Status.Ready
}

// SetReady implements the ICoreResource Interface
func (t *PodAccessTemplate) SetReady(ready bool) {
	t.Status.Ready = ready
}

// GetTargetRef conforms to the controllers.OzTemplateResource interface.
func (t *PodAccessTemplate) GetTargetRef() *CrossVersionObjectReference {
	return t.Spec.ControllerTargetRef
}

// GetAccessConfig returns the Spec.accessConfig field for this resource in an AccessConfig object form.
func (t *PodAccessTemplate) GetAccessConfig() *AccessConfig {
	return &t.Spec.AccessConfig
}

// Validate the inputs
func (t *PodAccessTemplate) Validate() error {
	if (*t.Spec.ControllerTargetRef != CrossVersionObjectReference{}) &&
		reflect.DeepEqual(t.Spec.PodSpec, corev1.PodSpec{}) {
		return errors.New(
			"cannot set both Spec.controllerTargetRef and spec.podSpec - use one or the other",
		)
	}

	if (*t.Spec.ControllerTargetRef == CrossVersionObjectReference{}) &&
		reflect.DeepEqual(t.Spec.ControllerTargetMutationConfig, PodTemplateSpecMutationConfig{}) {
		return errors.New(
			"cannot set Spec.controllerTargetMutationConfig if Spec.controllerTargetRef is not also set",
		)
	}

	return nil
}

//+kubebuilder:object:root=true

// PodAccessTemplateList contains a list of AccessTemplate
type PodAccessTemplateList struct {
	metav1.TypeMeta `                    json:",inline"`
	metav1.ListMeta `                    json:"metadata,omitempty"`
	Items           []PodAccessTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PodAccessTemplate{}, &PodAccessTemplateList{})
}

// GetPodAccessTemplate returns back an AccessTemplate resource matching the request supplied to the
// reconciler loop, or returns back an error.
func GetPodAccessTemplate(
	ctx context.Context,
	cl client.Client,
	name string,
	namespace string,
) (*PodAccessTemplate, error) {
	tmpl := &PodAccessTemplate{}
	err := cl.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, tmpl)
	return tmpl, err
}
