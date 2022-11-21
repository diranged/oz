package v1alpha1

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// +kubebuilder:object:generate=false
type isReadyAble interface {
	IsReady() bool
	SetReady(bool)
}

// +kubebuilder:object:generate=false
type hasStatusWithConditions interface {
	// Returns a pointer to a list of conditions. The pointer is important so that the returned value can be
	// updated and then the resource can be saved with the updated conditions.
	GetConditions() *[]metav1.Condition
}

// The OzResource interface wraps a standard client.Object resource (metav1.Object + runtime.Object)
// with a few additional requirements for common methods that we use throughout our reconciliation process.
// +kubebuilder:object:generate=false
type OzResource interface {
	// Common client.Object stuff
	metav1.Object
	runtime.Object

	// Requires that object exposes a GetConditions() method for returning the Status.Conditions data.
	hasStatusWithConditions

	// Requires that the object exposes a SetReady() and IsReady() method for handling Status.Ready
	// field.
	isReadyAble
}

// OzTemplateResource represents a common "AccessTemplate" resource for the Oz operator. These
// templates provide different types of access into resources (eg, "Exec" vs "Debug" vs "launch me a
// dedicated pod"). A set of common methods are required though that are used by the
// OzTemplateReconciler.
// +kubebuilder:object:generate=false
type OzTemplateResource interface {
	OzResource

	// Returns a CrossVersionObjectReference to the controller target for the template. Eg Deployment, StatefulSet, etc.
	GetTargetRef() *CrossVersionObjectReference

	// Returns back the Spec.allowedGroups field
	GetAllowedGroups() []string

	// Returns the Spec.defaultDuration in time.Duration() format
	GetDefaultDuration() (time.Duration, error)

	// Returns the Spec.maxduration in time.Duration() format
	GetMaxDuration() (time.Duration, error)
}

// OzRequestResource represents a common "AccesRequest" resource for the Oz operator. These requests
// have a common set of required methods that are used by the OzRequestReconciler.
// +kubebuilder:object:generate=false
type OzRequestResource interface {
	OzResource

	// Returns the Spec.duration in time.Duration() format, or nil.
	GetDuration() (time.Duration, error)

	// Returns the uptime in time.Duration() format
	GetUptime() time.Duration

	// Sets the Status.PodName field if it is empty. If it is set, returns an error.
	SetPodName(string) error

	// Gets the Status.PodName field, or returns an empty string.
	GetPodName() string
}
