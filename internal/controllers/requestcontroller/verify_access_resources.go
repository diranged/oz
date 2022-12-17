package requestcontroller

import (
	"fmt"
	"time"

	"github.com/diranged/oz/internal/api/v1alpha1"
	"github.com/diranged/oz/internal/controllers/internal/status"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	resourceWaitRequeueInterval    = (5 * time.Second)
	resourceWaitRequeueInitialWait = (5 * time.Second)
)

func (r *RequestReconciler) verifyAccessResources(
	rctx *RequestContext,
	tmpl v1alpha1.ITemplateResource,
) (shouldReturn bool, result ctrl.Result, resultErr error) {
	{ // Create the resources
		var statusStr string
		var err error

		rctx.log.V(1).Info("Making sure Access Resources have been created")
		if statusStr, err = r.Builder.CreateAccessResources(rctx.Context, r.Client, rctx.obj, tmpl); err != nil {
			// NOTE: Blindly ignoring the error return here because we are already
			// returning an error which will fail the reconciliation.
			_ = status.SetAccessResourcesNotCreated(rctx.Context, r, rctx.obj, err)
			return true, result, err
		}
		if err := status.SetAccessResourcesCreated(rctx.Context, r, rctx.obj, statusStr); err != nil {
			return true, result, err
		}
	}

	{ // Sleep a few seconds

		rctx.log.Info(
			fmt.Sprintf(
				"Waiting at least %s for them to become ready",
				resourceWaitRequeueInitialWait,
			),
		)
		time.Sleep(resourceWaitRequeueInitialWait)
	}

	{ // Check if the resources are ready
		rctx.log.V(1).Info("Checking if Access Resources are ready")
		if areReady, err := r.Builder.AccessResourcesAreReady(rctx.Context, r.Client, rctx.obj, tmpl); err != nil {
			// NOTE: Blindly ignoring the error return here because we are already
			// returning an error which will fail the reconciliation.
			_ = status.SetAccessResourcesNotReady(rctx.Context, r, rctx.obj, err)
			return true, result, err
		} else if !areReady {
			// NOTE: Blindly ignoring the error return here because we are already
			// returning an error which will fail the reconciliation.
			_ = status.SetAccessResourcesNotReady(rctx.Context, r, rctx.obj,
				fmt.Errorf("Resources not yet available... will check in %s", resourceWaitRequeueInterval))
			return true, ctrl.Result{RequeueAfter: resourceWaitRequeueInterval}, nil
		}

		rctx.log.V(1).Info("Builder indicates Access Resources are ready!")
		if err := status.SetAccessResourcesReady(rctx.Context, r, rctx.obj, "Ready"); err != nil {
			return true, result, err
		}
	}

	// Finally, do not requeue, do not end reconciliation. Move forward.
	return false, result, nil
}
