package v1alpha1

import (
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Important: Run "make" to regenerate code after modifying this file
type CrossVersionObjectReference struct {
	// Defines the "ApiVersion" of the resource being referred to. Eg, "apps/v1".
	//
	// TODO: Figure out how to regex validate that it has a "/" in it
	//
	// +kubebuilder:validation:Required
	ApiVersion string `json:"apiVersion"`

	// Defines the "Kind" of resource being referred to.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=Deployment;DaemonSet;StatefulSet
	Kind ControllerKind `json:"kind"`

	// Defines the "metadata.Name" of the target resource.
	// +kubebuilder:validation:Required
	Name string `json:"name"`
}

func (r *CrossVersionObjectReference) GetGroup() string {
	return strings.Split(r.ApiVersion, "/")[0]
}

func (r *CrossVersionObjectReference) GetVersion() string {
	return strings.Split(r.ApiVersion, "/")[1]
}

func (r *CrossVersionObjectReference) GetKind() string {
	return string(r.Kind)
}

func (r *CrossVersionObjectReference) GetName() string {
	return r.Name
}

func (r *CrossVersionObjectReference) GetGroupVersionKind() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   r.GetGroup(),
		Version: r.GetVersion(),
		Kind:    r.GetKind(),
	}
}

func (r *CrossVersionObjectReference) GetObject() client.Object {
	groupVersionKind := r.GetGroupVersionKind()
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(groupVersionKind)
	return obj
}
