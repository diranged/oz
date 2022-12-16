package v1alpha1

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ICoreStatus is used to define the core common status functions that all Status structs in this
// API must adhere to. These common functions simplify the reconciler() functions so that they can
// easily get/set status on the resources in a common way.
//
// +kubebuilder:object:generate=false
type ICoreStatus interface {
	IsReady() bool
	SetReady(bool)
	GetConditions() *[]metav1.Condition
}

// IRequestStatus is a more specific Status interface that enables getting and
// setting access instruction methods.
//
// +kubebuilder:object:generate=false
type IRequestStatus interface {
	ICoreStatus
	SetAccessMessage(string)
	GetAccessMessage() string
}

// ITemplateStatus provides a more specific Status interface for Access
// Templates. Functionality to come in the future.
//
// +kubebuilder:object:generate=false
type ITemplateStatus interface {
	ICoreStatus
}

// The ICoreResource interface wraps a standard client.Object resource (metav1.Object + runtime.Object)
// with a few additional requirements for common methods that we use throughout our reconciliation process.
//
// +kubebuilder:object:generate=false
type ICoreResource interface {
	// Common client.Object stuff
	metav1.Object
	runtime.Object

	// Returns a Status object that matches our ICoreStatus interface.
	GetStatus() ICoreStatus
}

// ITemplateResource represents a common "AccessTemplate" resource for the Oz Controller. These
// templates provide different types of access into resources (eg, "Exec" vs "Debug" vs "launch me a
// dedicated pod"). A set of common methods are required though that are used by the
// OzTemplateReconciler.
//
// +kubebuilder:object:generate=false
type ITemplateResource interface {
	ICoreResource

	// Returns a CrossVersionObjectReference to the controller target for the template. Eg Deployment, StatefulSet, etc.
	GetTargetRef() *CrossVersionObjectReference

	// Returns the Spec.accessConfig
	GetAccessConfig() *AccessConfig
}

// IRequestResource represents a common "AccesRequest" resource for the Oz Controller. These requests
// have a common set of required methods that are used by the OzRequestReconciler.
//
// +kubebuilder:object:generate=false
type IRequestResource interface {
	ICoreResource

	// Returns an unpopulated ITemplateResource that this IRequestResource points to
	GetTemplate(context.Context, client.Client) (ITemplateResource, error)

	// Returns the user-supplied Spec.templateName field
	GetTemplateName() string

	// Returns the Spec.duration in time.Duration() format, or nil.
	GetDuration() (time.Duration, error)

	// Returns the uptime in time.Duration() format
	GetUptime() time.Duration
}

// IPodRequestResource is a Pod-access specific request interface that exposes a few more functions
// for storing references to specific Pods that the requestor is being granted access to.
//
// +kubebuilder:object:generate=false
type IPodRequestResource interface {
	IRequestResource

	// Sets the Status.PodName field if it is empty. If it is set, returns an error.
	SetPodName(string) error

	// Gets the Status.PodName field, or returns an empty string.
	GetPodName() string
}
