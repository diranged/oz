package requestcontroller

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/diranged/oz/internal/api/v1alpha1"
)

func (r *RequestReconciler) isAccessExpired(
	rctx *RequestContext,
) (shouldEndReconcile bool, result ctrl.Result, resultErr error) {
	rctx.log.V(1).Info("Checking if access has expired...")
	conditions := rctx.obj.GetStatus().GetConditions()
	cond := meta.FindStatusCondition(*conditions, v1alpha1.ConditionAccessStillValid.String())
	if cond == nil {
		rctx.log.V(1).Info(
			fmt.Sprintf(
				"Missing Condition %s, skipping deletion",
				v1alpha1.ConditionAccessStillValid,
			),
		)
		shouldEndReconcile = false
		resultErr = nil
	} else if cond.Status == metav1.ConditionFalse {
		rctx.log.Info(
			fmt.Sprintf(
				"Found Condition %s in state %s, terminating request",
				v1alpha1.ConditionAccessStillValid,
				cond.Status,
			),
		)
		shouldEndReconcile = true
		result = ctrl.Result{}
		resultErr = r.Delete(rctx.Context, rctx.obj)
	} else {
		rctx.log.V(1).Info(
			fmt.Sprintf(
				"Found Condition %s in state %s, leaving alone",
				v1alpha1.ConditionAccessStillValid,
				cond.Status,
			),
		)
		shouldEndReconcile = false
		result = ctrl.Result{}
		resultErr = nil
	}

	return shouldEndReconcile, result, resultErr
}
