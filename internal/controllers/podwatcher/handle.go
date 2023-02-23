// Package podwatcher provides a Webhook handler for Pod Exec/Debug events for auditing purposes
package podwatcher

import (
	"context"
	"fmt"

	admissionv1 "k8s.io/api/admission/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/watch-v1-pod,mutating=false,failurePolicy=fail,sideEffects=None,groups="",resources=pods/exec;pods/attach,verbs=create;update;connect,versions=v1,name=vpod.kb.io,admissionReviewVersions=v1

// Handle is responsible for monitoring events that take place on a Pod
// (Attach, Execs, etc) and ultimately making decisions about whether or not
// those events can take place. The Handle() function primarily fires off
// requests to more explicit handlers for different event types and then
// returns the result.
//
// https://github.com/diranged/oz/issues/50 and
// https://github.com/diranged/oz/issues/51 will be handled through this
// endpoint in the future.
func (w *PodWatcher) Handle(ctx context.Context, req admission.Request) admission.Response {
	logger := log.FromContext(ctx)
	logger.Info(
		fmt.Sprintf(
			"Handling %s Operation on %s/%s by %s",
			req.Operation,
			req.Resource.Resource,
			req.Name,
			req.UserInfo.Username,
		), "request", ObjectToJSON(req),
	)

	// Ensure that this is a CONNECT Operation. Anything else, we just don't understand.
	if req.Operation != admissionv1.Connect {
		return admission.Allowed("")
	}

	// Decide what kind of attachment is being made
	switch reqKind := req.Kind.Kind; reqKind {
	case "PodAttachOptions":
		return w.HandleAttach(ctx, req)
	case "PodExecOptions":
		return w.HandleExec(ctx, req)
	}

	// If we got here, we don't understand the event type and just move on
	return admission.Allowed("")
}
