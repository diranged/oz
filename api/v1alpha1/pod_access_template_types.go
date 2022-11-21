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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// AccessTemplateSpec defines the desired state of AccessTemplate
type AccessTemplateSpec struct {
	// AccessConfig provides a common struct for defining who has access to the resources this
	// template controls, how long they have access, etc.
	AccessConfig AccessConfig `json:"accessConfig"`

	// TargetRef provides a pattern for referencing objects from another API in a generic way.
	// +kubebuilder:validation:Required
	TargetRef CrossVersionObjectReference `json:"targetRef"`

	// Command is used to override the .Spec.containers[0].command field for the target Pod and Container. This can
	// be handy in ensuring that the default application does not start up and do any work. If set, this overrides the
	// Spec.conatiners[0].args property as well.
	Command []string `json:"command,omitempty"`

	// If supplied these resource requirements will override the default .Spec.containers[0].resource requested for the
	// the pod. Note though that we do not override all of the resource requests in the Pod because there may be many
	// containers.
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Upper bound of the ephemeral storage that an AccessRequest can make against this template for
	// the primary container.
	MaxStorage resource.Quantity `json:"maxStorage,omitempty"`

	// Upper bound of the CPU that an AccessRequest can make against this tmemplate for the primary container.
	MaxCPU resource.Quantity `json:"maxCpu,omitempty"`

	// Upper bound of the memory that an AccessRequest can make against this template for the primary container.
	MaxMemory resource.Quantity `json:"maxMemory,omitempty"`
}

// AccessTemplateStatus defines the observed state of AccessRequest
type AccessTemplateStatus struct {
	CoreStatus `json:",inline"`

	// The Target Pod Name where access has been granted
	PodName string `json:"podName,omitempty"`

	// The name of the Role created for this temporary access request
	RoleName string `json:"roleName,omitempty"`

	// The name of th RoleBinding created for this temporary access request
	RoleBindingName string `json:"roleBindingName,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AccessTemplate is the Schema for the accesstemplates API
type AccessTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AccessTemplateSpec   `json:"spec,omitempty"`
	Status AccessTemplateStatus `json:"status,omitempty"`
}

// https://stackoverflow.com/questions/33089523/how-to-mark-golang-struct-as-implementing-interface
var _ ITemplateResource = &AccessTemplate{}
var _ ITemplateResource = (*AccessTemplate)(nil)

// GetStatus returns the core Status field for this resource.
//
// Returns:
//
//	AccessRequestStatus
func (t *AccessTemplate) GetStatus() ICoreStatus {
	return &t.Status
}

// GetTargetRef conforms to the controllers.OzTemplateResource interface.
func (t *AccessTemplate) GetTargetRef() *CrossVersionObjectReference {
	return &t.Spec.TargetRef
}

// GetAccessConfig returns the Spec.accessConfig field for this resource in an AccessConfig object form.
func (t *AccessTemplate) GetAccessConfig() *AccessConfig {
	return &t.Spec.AccessConfig
}

//+kubebuilder:object:root=true

// AccessTemplateList contains a list of AccessTemplate
type AccessTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AccessTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AccessTemplate{}, &AccessTemplateList{})
}

// GetAccessTemplate returns back an AccessTemplate resource matching the request supplied to the
// reconciler loop, or returns back an error.
func GetAccessTemplate(ctx context.Context, cl client.Client, name string, namespace string) (*AccessTemplate, error) {
	tmpl := &AccessTemplate{}
	err := cl.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, tmpl)
	return tmpl, err
}
