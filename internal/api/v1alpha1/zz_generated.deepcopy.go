//go:build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AccessConfig) DeepCopyInto(out *AccessConfig) {
	*out = *in
	if in.AllowedGroups != nil {
		in, out := &in.AllowedGroups, &out.AllowedGroups
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AccessConfig.
func (in *AccessConfig) DeepCopy() *AccessConfig {
	if in == nil {
		return nil
	}
	out := new(AccessConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CoreStatus.
func (in *CoreStatus) DeepCopy() *CoreStatus {
	if in == nil {
		return nil
	}
	out := new(CoreStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CrossVersionObjectReference) DeepCopyInto(out *CrossVersionObjectReference) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CrossVersionObjectReference.
func (in *CrossVersionObjectReference) DeepCopy() *CrossVersionObjectReference {
	if in == nil {
		return nil
	}
	out := new(CrossVersionObjectReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExecAccessRequest) DeepCopyInto(out *ExecAccessRequest) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExecAccessRequest.
func (in *ExecAccessRequest) DeepCopy() *ExecAccessRequest {
	if in == nil {
		return nil
	}
	out := new(ExecAccessRequest)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ExecAccessRequest) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExecAccessRequestList) DeepCopyInto(out *ExecAccessRequestList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ExecAccessRequest, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExecAccessRequestList.
func (in *ExecAccessRequestList) DeepCopy() *ExecAccessRequestList {
	if in == nil {
		return nil
	}
	out := new(ExecAccessRequestList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ExecAccessRequestList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExecAccessRequestSpec) DeepCopyInto(out *ExecAccessRequestSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExecAccessRequestSpec.
func (in *ExecAccessRequestSpec) DeepCopy() *ExecAccessRequestSpec {
	if in == nil {
		return nil
	}
	out := new(ExecAccessRequestSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExecAccessRequestStatus) DeepCopyInto(out *ExecAccessRequestStatus) {
	*out = *in
	in.CoreStatus.DeepCopyInto(&out.CoreStatus)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExecAccessRequestStatus.
func (in *ExecAccessRequestStatus) DeepCopy() *ExecAccessRequestStatus {
	if in == nil {
		return nil
	}
	out := new(ExecAccessRequestStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExecAccessTemplate) DeepCopyInto(out *ExecAccessTemplate) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExecAccessTemplate.
func (in *ExecAccessTemplate) DeepCopy() *ExecAccessTemplate {
	if in == nil {
		return nil
	}
	out := new(ExecAccessTemplate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ExecAccessTemplate) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExecAccessTemplateList) DeepCopyInto(out *ExecAccessTemplateList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ExecAccessTemplate, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExecAccessTemplateList.
func (in *ExecAccessTemplateList) DeepCopy() *ExecAccessTemplateList {
	if in == nil {
		return nil
	}
	out := new(ExecAccessTemplateList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ExecAccessTemplateList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExecAccessTemplateSpec) DeepCopyInto(out *ExecAccessTemplateSpec) {
	*out = *in
	in.AccessConfig.DeepCopyInto(&out.AccessConfig)
	if in.ControllerTargetRef != nil {
		in, out := &in.ControllerTargetRef, &out.ControllerTargetRef
		*out = new(CrossVersionObjectReference)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExecAccessTemplateSpec.
func (in *ExecAccessTemplateSpec) DeepCopy() *ExecAccessTemplateSpec {
	if in == nil {
		return nil
	}
	out := new(ExecAccessTemplateSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExecAccessTemplateStatus) DeepCopyInto(out *ExecAccessTemplateStatus) {
	*out = *in
	in.CoreStatus.DeepCopyInto(&out.CoreStatus)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExecAccessTemplateStatus.
func (in *ExecAccessTemplateStatus) DeepCopy() *ExecAccessTemplateStatus {
	if in == nil {
		return nil
	}
	out := new(ExecAccessTemplateStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *JSONPatchOperation) DeepCopyInto(out *JSONPatchOperation) {
	*out = *in
	out.Value = in.Value
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new JSONPatchOperation.
func (in *JSONPatchOperation) DeepCopy() *JSONPatchOperation {
	if in == nil {
		return nil
	}
	out := new(JSONPatchOperation)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PodAccessRequest) DeepCopyInto(out *PodAccessRequest) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PodAccessRequest.
func (in *PodAccessRequest) DeepCopy() *PodAccessRequest {
	if in == nil {
		return nil
	}
	out := new(PodAccessRequest)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PodAccessRequest) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PodAccessRequestList) DeepCopyInto(out *PodAccessRequestList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]PodAccessRequest, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PodAccessRequestList.
func (in *PodAccessRequestList) DeepCopy() *PodAccessRequestList {
	if in == nil {
		return nil
	}
	out := new(PodAccessRequestList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PodAccessRequestList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PodAccessRequestSpec) DeepCopyInto(out *PodAccessRequestSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PodAccessRequestSpec.
func (in *PodAccessRequestSpec) DeepCopy() *PodAccessRequestSpec {
	if in == nil {
		return nil
	}
	out := new(PodAccessRequestSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PodAccessRequestStatus) DeepCopyInto(out *PodAccessRequestStatus) {
	*out = *in
	in.CoreStatus.DeepCopyInto(&out.CoreStatus)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PodAccessRequestStatus.
func (in *PodAccessRequestStatus) DeepCopy() *PodAccessRequestStatus {
	if in == nil {
		return nil
	}
	out := new(PodAccessRequestStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PodAccessTemplate) DeepCopyInto(out *PodAccessTemplate) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PodAccessTemplate.
func (in *PodAccessTemplate) DeepCopy() *PodAccessTemplate {
	if in == nil {
		return nil
	}
	out := new(PodAccessTemplate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PodAccessTemplate) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PodAccessTemplateList) DeepCopyInto(out *PodAccessTemplateList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]PodAccessTemplate, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PodAccessTemplateList.
func (in *PodAccessTemplateList) DeepCopy() *PodAccessTemplateList {
	if in == nil {
		return nil
	}
	out := new(PodAccessTemplateList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PodAccessTemplateList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PodAccessTemplateSpec) DeepCopyInto(out *PodAccessTemplateSpec) {
	*out = *in
	in.AccessConfig.DeepCopyInto(&out.AccessConfig)
	if in.ControllerTargetRef != nil {
		in, out := &in.ControllerTargetRef, &out.ControllerTargetRef
		*out = new(CrossVersionObjectReference)
		**out = **in
	}
	if in.ControllerTargetMutationConfig != nil {
		in, out := &in.ControllerTargetMutationConfig, &out.ControllerTargetMutationConfig
		*out = new(PodTemplateSpecMutationConfig)
		(*in).DeepCopyInto(*out)
	}
	if in.PodSpec != nil {
		in, out := &in.PodSpec, &out.PodSpec
		*out = new(v1.PodSpec)
		(*in).DeepCopyInto(*out)
	}
	out.MaxStorage = in.MaxStorage.DeepCopy()
	out.MaxCPU = in.MaxCPU.DeepCopy()
	out.MaxMemory = in.MaxMemory.DeepCopy()
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PodAccessTemplateSpec.
func (in *PodAccessTemplateSpec) DeepCopy() *PodAccessTemplateSpec {
	if in == nil {
		return nil
	}
	out := new(PodAccessTemplateSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PodAccessTemplateStatus) DeepCopyInto(out *PodAccessTemplateStatus) {
	*out = *in
	in.CoreStatus.DeepCopyInto(&out.CoreStatus)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PodAccessTemplateStatus.
func (in *PodAccessTemplateStatus) DeepCopy() *PodAccessTemplateStatus {
	if in == nil {
		return nil
	}
	out := new(PodAccessTemplateStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PodTemplateSpecMutationConfig) DeepCopyInto(out *PodTemplateSpecMutationConfig) {
	*out = *in
	if in.Command != nil {
		in, out := &in.Command, &out.Command
		*out = new([]string)
		if **in != nil {
			in, out := *in, *out
			*out = make([]string, len(*in))
			copy(*out, *in)
		}
	}
	if in.Args != nil {
		in, out := &in.Args, &out.Args
		*out = new([]string)
		if **in != nil {
			in, out := *in, *out
			*out = make([]string, len(*in))
			copy(*out, *in)
		}
	}
	if in.Env != nil {
		in, out := &in.Env, &out.Env
		*out = make([]v1.EnvVar, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.Resources.DeepCopyInto(&out.Resources)
	if in.PodAnnotations != nil {
		in, out := &in.PodAnnotations, &out.PodAnnotations
		*out = new(map[string]string)
		if **in != nil {
			in, out := *in, *out
			*out = make(map[string]string, len(*in))
			for key, val := range *in {
				(*out)[key] = val
			}
		}
	}
	if in.PodLabels != nil {
		in, out := &in.PodLabels, &out.PodLabels
		*out = new(map[string]string)
		if **in != nil {
			in, out := *in, *out
			*out = make(map[string]string, len(*in))
			for key, val := range *in {
				(*out)[key] = val
			}
		}
	}
	if in.PatchSpecOperations != nil {
		in, out := &in.PatchSpecOperations, &out.PatchSpecOperations
		*out = make([]JSONPatchOperation, len(*in))
		copy(*out, *in)
	}
	if in.NodeSelector != nil {
		in, out := &in.NodeSelector, &out.NodeSelector
		*out = new(map[string]string)
		if **in != nil {
			in, out := *in, *out
			*out = make(map[string]string, len(*in))
			for key, val := range *in {
				(*out)[key] = val
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PodTemplateSpecMutationConfig.
func (in *PodTemplateSpecMutationConfig) DeepCopy() *PodTemplateSpecMutationConfig {
	if in == nil {
		return nil
	}
	out := new(PodTemplateSpecMutationConfig)
	in.DeepCopyInto(out)
	return out
}
