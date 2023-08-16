package podwatcher

import (
	"context"
	"fmt"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// HandleExec monitors for CONNECT events on existing Pods and logs events
// about them.
func (w *PodWatcher) HandleExec(ctx context.Context, req admission.Request) admission.Response {
	logger := log.FromContext(ctx)

	// Parse the request into a set of PodExecOptions that we can understand
	opts := &corev1.PodExecOptions{}
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
	w.recorder.Event(pod, "Normal", "PodExec", eventMsg)
	logger.Info(eventMsg)

	return admission.Allowed("")
}
