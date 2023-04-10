package podwatcher

import (
	"context"
	"fmt"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// HandleAttach is a placeholder for future logic that will validate whether or
// not a user has the appropriate permissions to attach to a new pod. Currently
// this function logs the event, and that is it.
func (w *PodWatcher) HandleAttach(ctx context.Context, req admission.Request) admission.Response {
	logger := log.FromContext(ctx)

	opts := &corev1.PodAttachOptions{}
	err := w.decoder.Decode(req, opts)
	if err != nil {
		logger.Error(err, "Couldnt decode")
		return admission.Errored(http.StatusBadRequest, err)
	}

	// Create a Pod reference object for the Pod being attached, so that we can
	// attach events to the Pod.
	pod := getPod(req)

	// Format a message
	eventMsg := fmt.Sprintf(
		"%s/%s operation on Pod %s (container: %s) by %s (interactive: %t)",
		req.Operation,
		req.SubResource,
		req.Name,
		opts.Container,
		req.UserInfo.Username,
		opts.TTY,
	)

	// Log and Record the event
	w.recorder.Event(pod, "Normal", "PodAttach", eventMsg)
	logger.Info(eventMsg)

	return admission.Allowed("");
}
