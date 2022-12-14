// Package podwatcher provides a Webhook handler for Pod Exec/Debug events for auditing purposes
package podwatcher

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// example code: https://github.com/kubernetes-sigs/controller-runtime/blob/master/examples/builtins/validatingwebhook.go

// PodExecWatcher is a ValidatingWebhookEndpoint that receives calls from the
// Kubernetes API just before Pod's "exec" subresource is written into the
// cluster. The intention for this resource is to perform audit-logging type
// actions in the short term, and in the long term provide a more granular
// layer of security for Pod Exec access.
type PodExecWatcher struct {
	Client  client.Client
	decoder *admission.Decoder
}

// +kubebuilder:webhook:path=/watch-v1-pod,mutating=false,failurePolicy=fail,sideEffects=None,groups="",resources=pods/exec;pods/attach,verbs=create;update;connect,versions=v1,name=vpod.kb.io,admissionReviewVersions=v1

// Handle logs out each time an Exec/Attach call is made on a pod.
//
// Right now this is purely an informative log event. When we take care of
// https://github.com/diranged/oz/issues/24, we can use this handler to push
// events onto the Pods (and Access Requests) for audit purposes.
//
// Additionally, https://github.com/diranged/oz/issues/50 and
// https://github.com/diranged/oz/issues/51 will be handled through this
// endpoint in the future.
func (w *PodExecWatcher) Handle(ctx context.Context, req admission.Request) admission.Response {
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

	exec := &corev1.PodExecOptions{}
	err := w.decoder.Decode(req, exec)
	if err != nil {
		logger.Error(err, "Couldnt decode")
		return admission.Errored(http.StatusBadRequest, err)
	}

	return admission.Allowed("")
}

// PodWatcher implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (w *PodExecWatcher) InjectDecoder(d *admission.Decoder) error {
	w.decoder = d
	return nil
}

// ObjectToJSON is a quick helper function for pretty-printing an entire K8S object in JSON form.
// Used in certain debug log statements primarily.
func ObjectToJSON(obj any) string {
	jsonData, err := json.Marshal(obj)
	if err != nil {
		fmt.Printf("could not marshal json: %s\n", err)
		return ""
	}
	return string(jsonData)
}
