package status

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/diranged/oz/internal/api/v1alpha1"
	api "github.com/diranged/oz/internal/api/v1alpha1"
)

// UpdateCondition provides a simple way to update the .Status.Conditions field
// of a given resource. The resource must match the ResourceWithConditions
// interface - which exposes the GetConditions() method.
//
// When an updateCondition() call is made, we retrieve the current list of
// conditions first from the request object. From there, we insert in a new
// Condition into the resource. Finally we call the UpdateStatus() function to
// push the update to Kubernetes.
//
// revive:disable:argument-limit long but reasonable
func UpdateCondition(
	ctx context.Context,
	rec hasStatusReconciler,
	res api.ICoreResource,
	conditionType v1alpha1.IConditionType,
	conditionStatus metav1.ConditionStatus,
	reason string,
	message string,
) error {
	logger := log.FromContext(ctx)
	logger.V(1).
		Info(fmt.Sprintf("Updating condition %s to %s", conditionType, conditionStatus))

	meta.SetStatusCondition(res.GetStatus().GetConditions(), metav1.Condition{
		Type:               conditionType.String(),
		Status:             conditionStatus,
		ObservedGeneration: res.GetGeneration(),
		LastTransitionTime: metav1.Time{},
		Reason:             reason,
		Message:            message,
	})

	// Save the object into Kubernetes, and return any error that might have happened.
	return UpdateStatus(ctx, rec, res)
}
