package request_controller

import (
	"context"
	"fmt"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/diranged/oz/internal/controllers/internal/ctrlrequeue"
)

//+kubebuilder:rbac:groups=crds.wizardofoz.co,resources=execaccessrequests,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=crds.wizardofoz.co,resources=execaccessrequests/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=crds.wizardofoz.co,resources=execaccessrequests/finalizers,verbs=update

//+kubebuilder:rbac:groups=crds.wizardofoz.co,resources=podaccessrequests,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=crds.wizardofoz.co,resources=podaccessrequests/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=crds.wizardofoz.co,resources=podaccessrequests/finalizers,verbs=update

//+kubebuilder:rbac:groups=apps,resources=deployments;daemonsets;statefulsets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch

// Reconcile is a high level entrypoint triggered by Watches on particular
// Custom Resources within the cluster. This wrapper handles a few common
// startup behaviors, and introduces reconcile timing logging.
//
// The real business-logic is in the reconcile() (lowercased) function.
func (r *RequestReconciler) Reconcile(
	ctx context.Context,
	req ctrl.Request,
) (ctrl.Result, error) {
	var result ctrl.Result
	var err error

	// Build a RequestContext for this reconciliation loop
	rctx := newRequestContext(ctx, r.RequestType, req)

	// Boilerplate. Report back on every reconcile how long it took.
	start := time.Now()
	defer func() {
		msg := fmt.Sprintf("reconciliation finished in %s", time.Since(start))
		if result.RequeueAfter > 0 {
			msg = fmt.Sprintf(
				"%s, next run in %s (%s)",
				msg, result.RequeueAfter, time.Now().Add(result.RequeueAfter).Format(time.RFC3339),
			)
		}
		if err != nil {
			rctx.log.Error(err, msg)
		} else {
			rctx.log.Info(msg)
		}
	}()

	// Run the actual reconciliation an return that result. Pass in the
	// Component object that's already been populated by the cache.
	result, err = r.reconcile(rctx)
	return result, err
}

// reconcile() manages the state for a Component through the generic Installers package.
//
// revive:disable:confusing-naming
func (r *RequestReconciler) reconcile(rctx *RequestContext) (ctrl.Result, error) {
	rctx.log.Info("Starting reconcile loop")

	// First make sure we use the ApiReader (non-cached) client to go and
	// figure out if the resource exists or not. If it doesn't come back, we
	// exit out beacuse it is likely the object has been deleted and we no
	// longer need to worry about it.
	rctx.log.V(1).Info("Verifying still exists")

	// VERIFICATION: Does the resource exist anymore at all? If the component
	// no longer exists, then there is no work for us to do.
	if err := r.fetchRequestObject(rctx); err != nil {
		if apierrors.IsNotFound(err) {
			rctx.log.V(2).Info(fmt.Sprintf("Request %q not found, must be deleted", rctx.req.Name))
			return ctrlrequeue.NoRequeue()
		}
		// Error reading the object, requeue the request.
		return ctrlrequeue.RequeueError(err)
	}
	rctx.log.V(2).Info("Found request", "request", rctx.obj)

	// VERIFICATION: Check that the Builder can find the template the Request references
	if err := r.verifyTemplate(rctx); err != nil {
		rctx.log.Error(err, "Error - will requeue")
		return ctrlrequeue.RequeueError(err)
	}

	// UPDATE: Set the OwnerReference for the request - so if the template is
	// deleted, all requests are deleted.
	if err := r.Builder.SetOwnerReference(rctx.Context, r.Client, rctx.obj); err != nil {
		rctx.log.Error(err, "Error setting owner reference - will requeue")
		return ctrlrequeue.RequeueError(err)
	}

	//
	return ctrlrequeue.RequeueAfter(r.ReconcilliationInterval)
}
