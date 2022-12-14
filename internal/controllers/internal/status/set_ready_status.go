package status

import (
	"context"
	"fmt"
	"strconv"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"

	api "github.com/diranged/oz/internal/api/v1alpha1"
)

// SetReadyStatus flips the Status.Ready field to true or false. This is used at the end of a reconciliation loop
// when all of the conditions of the resource are known to have been populated. If all Conditions are in the
// ConditionSuccess status, then Status.Ready is set to true. Otherwise, it is set to False.
//
// Status.Ready is used by the 'ozctl' commandline tool to inform users when their access request
// has been approved and configured.
func SetReadyStatus(ctx context.Context, rec hasStatusReconciler, res api.ICoreResource) error {
	logger := log.FromContext(ctx)
	logger.V(1).Info("Checking final condition state")

	// Default to everything being ready. We'll iterate though all conditions and then flip this to false if any
	// of those conditions are not true.
	ready := true

	// Get the pointer to the conditions list
	conditions := res.GetStatus().GetConditions()

	// Iterate. If any are not true, then we flip the ready flag to false.
	for _, cond := range *conditions {
		if cond.Status != metav1.ConditionTrue {
			ready = false
		}
	}

	// Save the flag, and update the object. Return the result of the object update (if its an error).
	logger.Info(fmt.Sprintf("Setting ready state to %s", strconv.FormatBool(ready)))
	res.GetStatus().SetReady(ready)
	return UpdateStatus(ctx, rec, res)
}
