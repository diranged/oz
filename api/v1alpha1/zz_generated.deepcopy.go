//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AccessRequest) DeepCopyInto(out *AccessRequest) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AccessRequest.
func (in *AccessRequest) DeepCopy() *AccessRequest {
	if in == nil {
		return nil
	}
	out := new(AccessRequest)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AccessRequest) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AccessRequestList) DeepCopyInto(out *AccessRequestList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]AccessRequest, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AccessRequestList.
func (in *AccessRequestList) DeepCopy() *AccessRequestList {
	if in == nil {
		return nil
	}
	out := new(AccessRequestList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AccessRequestList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AccessRequestSpec) DeepCopyInto(out *AccessRequestSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AccessRequestSpec.
func (in *AccessRequestSpec) DeepCopy() *AccessRequestSpec {
	if in == nil {
		return nil
	}
	out := new(AccessRequestSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AccessRequestStatus) DeepCopyInto(out *AccessRequestStatus) {
	*out = *in
	in.ozResourceCoreStatus.DeepCopyInto(&out.ozResourceCoreStatus)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AccessRequestStatus.
func (in *AccessRequestStatus) DeepCopy() *AccessRequestStatus {
	if in == nil {
		return nil
	}
	out := new(AccessRequestStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AccessTemplate) DeepCopyInto(out *AccessTemplate) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AccessTemplate.
func (in *AccessTemplate) DeepCopy() *AccessTemplate {
	if in == nil {
		return nil
	}
	out := new(AccessTemplate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AccessTemplate) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AccessTemplateList) DeepCopyInto(out *AccessTemplateList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]AccessTemplate, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AccessTemplateList.
func (in *AccessTemplateList) DeepCopy() *AccessTemplateList {
	if in == nil {
		return nil
	}
	out := new(AccessTemplateList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AccessTemplateList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AccessTemplateSpec) DeepCopyInto(out *AccessTemplateSpec) {
	*out = *in
	out.TargetRef = in.TargetRef
	if in.AllowedGroups != nil {
		in, out := &in.AllowedGroups, &out.AllowedGroups
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Command != nil {
		in, out := &in.Command, &out.Command
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	in.Resources.DeepCopyInto(&out.Resources)
	out.MaxStorage = in.MaxStorage.DeepCopy()
	out.MaxCPU = in.MaxCPU.DeepCopy()
	out.MaxMemory = in.MaxMemory.DeepCopy()
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AccessTemplateSpec.
func (in *AccessTemplateSpec) DeepCopy() *AccessTemplateSpec {
	if in == nil {
		return nil
	}
	out := new(AccessTemplateSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AccessTemplateStatus) DeepCopyInto(out *AccessTemplateStatus) {
	*out = *in
	in.ozResourceCoreStatus.DeepCopyInto(&out.ozResourceCoreStatus)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AccessTemplateStatus.
func (in *AccessTemplateStatus) DeepCopy() *AccessTemplateStatus {
	if in == nil {
		return nil
	}
	out := new(AccessTemplateStatus)
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
	in.ozResourceCoreStatus.DeepCopyInto(&out.ozResourceCoreStatus)
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
	out.TargetRef = in.TargetRef
	if in.AllowedGroups != nil {
		in, out := &in.AllowedGroups, &out.AllowedGroups
		*out = make([]string, len(*in))
		copy(*out, *in)
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
	in.ozResourceCoreStatus.DeepCopyInto(&out.ozResourceCoreStatus)
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
