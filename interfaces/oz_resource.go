package interfaces

import (
	"time"

	api "github.com/diranged/oz/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type isReadyAble interface {
	IsReady() bool
	SetReady(bool)
}

type hasStatusWithConditions interface {
	// Returns a pointer to a list of conditions. The pointer is important so that the returned value can be
	// updated and then the resource can be saved with the updated conditions.
	GetConditions() *[]metav1.Condition
}

// This interface wraps the standard client.Object resource (metav1.Object + runtime.Object) with a requirement for
// a `GetConditions()` function that returns back the nested Status.Conditions list. This is used by
// BaseReconciler.UpdateCondition()
type OzResource interface {
	metav1.Object
	runtime.Object

	hasStatusWithConditions
	isReadyAble
}

// This interface represents common "AccessTemplate" resources for the Oz operator. These templates
// provide different types of access into resources (eg, "Exec" vs "Debug" vs "launch me a dedicated pod"),
// but all provide a common set of functions allowing our OzTemplateReconciler to be used as a common starting
// point for the corresponding reconciliation classes.
type OzTemplateResource interface {
	OzResource

	// Returns a CrossVersionObjectReference to the controller target for the template. Eg Deployment, StatefulSet, etc.
	GetTargetRef() *api.CrossVersionObjectReference

	// Returns the Spec.defaultDuration in time.Duration() format
	GetDefaultDuration() (time.Duration, error)

	// Returns the Spec.maxduration in time.Duration() format
	GetMaxDuration() (time.Duration, error)
}

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
