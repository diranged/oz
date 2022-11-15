package interfaces

import (
	api "github.com/diranged/oz/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type HasStatusWithConditions interface {
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

	HasStatusWithConditions
}

type OzTemplateResource interface {
	OzResource
	GetTemplateTarget() *api.CrossVersionObjectReference
}

type OzRequestResource interface {
	OzResource
}
