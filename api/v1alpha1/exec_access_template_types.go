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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ExecAccessTemplateSpec defines the desired state of ExecAccessTemplate
type ExecAccessTemplateSpec struct {
	// AccessConfig provides a common struct for defining who has access to the resources this
	// template controls, how long they have access, etc.
	AccessConfig AccessConfig `json:"accessConfig"`

	// ControllerTargetRef provides a pattern for referencing objects from another API in a generic way.
	//
	// +kubebuilder:validation:Required
	ControllerTargetRef *CrossVersionObjectReference `json:"controllerTargetRef"`
}

// ExecAccessTemplateStatus is the core set of status fields that we expect to be in each and every one of
// our template (AccessTemplate, ExecAccessTemplate, etc) resources.
type ExecAccessTemplateStatus struct {
	CoreStatus `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ExecAccessTemplate is the Schema for the execaccesstemplates API
//
// +kubebuilder:printcolumn:name="Ready",type="boolean",JSONPath=".status.ready",description="Is template ready?"
type ExecAccessTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ExecAccessTemplateSpec   `json:"spec,omitempty"`
	Status ExecAccessTemplateStatus `json:"status,omitempty"`
}

// https://stackoverflow.com/questions/33089523/how-to-mark-golang-struct-as-implementing-interface
var (
	_ ITemplateResource = &ExecAccessTemplate{}
	_ ITemplateResource = (*ExecAccessTemplate)(nil)
)

// GetStatus returns the core Status field for this resource.
//
// Returns:
//
//	AccessRequestStatus
func (t *ExecAccessTemplate) GetStatus() ICoreStatus {
	return &t.Status
}

// GetAccessConfig returns the Spec.accessConfig field for this resource in an AccessConfig object form.
func (t *ExecAccessTemplate) GetAccessConfig() *AccessConfig {
	return &t.Spec.AccessConfig
}

// GetTargetRef conforms to the controllers.OzTemplateResource interface.
func (t *ExecAccessTemplate) GetTargetRef() *CrossVersionObjectReference {
	return t.Spec.ControllerTargetRef
}

// GetExecAccessTemplate returns back an ExecAccessTemplate resource matching the request supplied to the reconciler loop, or returns back an error.
func GetExecAccessTemplate(
	ctx context.Context,
	cl client.Client,
	name string,
	namespace string,
) (*ExecAccessTemplate, error) {
	tmpl := &ExecAccessTemplate{}
	err := cl.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, tmpl)
	return tmpl, err
}

//+kubebuilder:object:root=true

// ExecAccessTemplateList contains a list of ExecAccessTemplate
type ExecAccessTemplateList struct {
	metav1.TypeMeta `                     json:",inline"`
	metav1.ListMeta `                     json:"metadata,omitempty"`
	Items           []ExecAccessTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ExecAccessTemplate{}, &ExecAccessTemplateList{})
}
