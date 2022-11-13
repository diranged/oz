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
	"crypto/md5"
	"encoding/hex"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	//
	// TODO: Implement this
	//
	TargetPod string `json:"targetPod,omitempty"`
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

func (r *ExecAccessRequest) GetUniqueId() string {
	idString := fmt.Sprintf("%s-%s-%s", r.Name, r.Namespace, r.CreationTimestamp)
	hash := md5.Sum([]byte(idString))
	return hex.EncodeToString(hash[:])[0:10]
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
