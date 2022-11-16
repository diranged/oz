package v1alpha1

import (
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CrossVersionObjectReference provides us a generic way to define a reference to an APIGroup, Kind
// and Name of a particular resource. Primarily used for the AccessTemplate and ExecAccessTemplate,
// but generic enough to be used in other resources down the road.
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

// GetGroup returns the APIGroup name only (eg "apps")
func (r *CrossVersionObjectReference) GetGroup() string {
	return strings.Split(r.ApiVersion, "/")[0]
}

// GetVersion returns the API "Version" only (eg "v1")
func (r *CrossVersionObjectReference) GetVersion() string {
	return strings.Split(r.ApiVersion, "/")[1]
}

// GetKind returns the resource Kind (eg "Deployment")
func (r *CrossVersionObjectReference) GetKind() string {
	return string(r.Kind)
}

// GetName returns the Name of the resource (eg "MyDeploymentThing")
func (r *CrossVersionObjectReference) GetName() string {
	return r.Name
}

// GetGroupVersionKind returns a populated schema object thta can be used by the unstructured
// Kubernetes API client to get the final target object from the API.
func (r *CrossVersionObjectReference) GetGroupVersionKind() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   r.GetGroup(),
		Version: r.GetVersion(),
		Kind:    r.GetKind(),
	}
}

// GetObject returns a generic unstrucutred resource that points to the desired API object. Because
// this is unstructured (for now), you can really only use this to get metadata back from the API
// about the resource.
//
// TODO: Figure out if we can cast this into a desired object type in some way that would provide us
// access to the Spec.
func (r *CrossVersionObjectReference) GetObject() client.Object {
	groupVersionKind := r.GetGroupVersionKind()
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(groupVersionKind)
	return obj
}