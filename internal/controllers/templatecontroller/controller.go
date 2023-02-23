package templatecontroller

import (
	"context"
	"fmt"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/diranged/oz/internal/controllers/internal/ctrlrequeue"
	"github.com/diranged/oz/internal/controllers/internal/status"
)

// Annotation for generating RBAC role for writing Events
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

//+kubebuilder:rbac:groups=crds.wizardofoz.co,resources=execaccesstemplates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=crds.wizardofoz.co,resources=execaccesstemplates/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=crds.wizardofoz.co,resources=execaccesstemplates/finalizers,verbs=update

//+kubebuilder:rbac:groups=crds.wizardofoz.co,resources=podaccesstemplates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=crds.wizardofoz.co,resources=podaccesstemplates/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=crds.wizardofoz.co,resources=podaccesstemplates/finalizers,verbs=update

//+kubebuilder:rbac:groups=apps,resources=deployments;daemonsets;statefulsets,verbs=get;list;watch

// Reconcile is a high level entrypoint triggered by Watches on particular
// Custom Resources within the cluster. This wrapper handles a few common
// startup behaviors, and introduces reconcile timing logging.
//
// The real business-logic is in the reconcile() (lowercased) function.
func (r *TemplateReconciler) Reconcile(
	ctx context.Context,
	req ctrl.Request,
) (ctrl.Result, error) {
	var result ctrl.Result
	var err error

	// Build a RequestContext for this reconciliation loop
	rctx := newRequestContext(ctx, r.TemplateType, req)

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
func (r *TemplateReconciler) reconcile(rctx *RequestContext) (ctrl.Result, error) {
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

	// VERIFICATION: Make sure that the TargetRef is valid and points to an active controller
	//
	// An error is only returned if the conditions update fails. Otherwise we
	// continue to move on.
	err := r.verifyTargetRef(rctx)
	if err != nil {
		return ctrlrequeue.RequeueError(err)
	}

	// VERIFICATION: Make sure the DefaultDuration and MaxDuration settings are valid.
	//
	// An error is only returned if the conditions update fails. Otherwise we
	// continue to move on.
	err = r.verifyDuration(rctx)
	if err != nil {
		return ctrlrequeue.RequeueError(err)
	}

	// TODO:
	// VERIFICATION: Ensure that the allowedGroups match valid group name strings

	// FINAL: Set Status.Ready state
	err = status.SetReadyStatus(rctx, r, rctx.obj)
	if err != nil {
		return ctrlrequeue.RequeueError(err)
	}

	// Exit Reconciliation Loop
	rctx.log.Info("Ending reconcile loop")
	return ctrlrequeue.RequeueAfter(r.ReconciliationInterval)
}
