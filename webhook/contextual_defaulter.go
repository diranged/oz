package webhook

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// IContextuallyDefaultableObject implements a similar pattern to the
// [`controller-runtime`](https://github.com/kubernetes-sigs/controller-runtime/tree/v0.13.1/pkg/webhook)
// webhook pattern. The difference is that the `Default()` function is not only
// supplied the request resource, but also the request context in the form of
// an
// [`admission.Request`](https://github.com/kubernetes-sigs/controller-runtime/blob/v0.13.1/pkg/webhook/admission/webhook.go#L43-L66)
// object.
//
// Modified from https://github.com/kubernetes-sigs/controller-runtime/blob/v0.13.1/pkg/webhook/admission/defaulter_custom.go#L29-L32
type IContextuallyDefaultableObject interface {
	runtime.Object
	Default(req admission.Request) error
}

// RegisterContextualDefaulter leverages many of the patterns and code from the
// Controller-Runtime Admission package, but is one level _less_ abstracted.
// Rather than calling the `Default()` function on the target resource type,
func RegisterContextualDefaulter(
	obj IContextuallyDefaultableObject,
	mgr ctrl.Manager,
) error {
	// Get the GroupVersionKind for the target schema object.
	gvk, err := apiutil.GVKForObject(obj, mgr.GetScheme())
	if err != nil {
		return err
	}
	path := generateMutatePath(gvk)

	// Create a Webhook{} resource with our Handler.
	mwh := &admission.Webhook{
		Handler: &defaulterForType{object: obj},
	}

	// Insert the path into the webhook server and point it at our mutating
	// webhook handler. This must take place before the default controller
	// NewWebhookManagedBy().Complete() function is called.
	mgr.GetWebhookServer().Register(path, mwh)

	return nil
}

// A defaulterForType mimics the
// [`defaulterForType`](https://github.com/kubernetes-sigs/controller-runtime/blob/v0.13.1/pkg/webhook/admission/defaulter_custom.go)
// code, but understands to pass the `admission.Request` object into the `Default()` function.
//
// https://github.com/kubernetes-sigs/controller-runtime/blob/v0.13.1/pkg/webhook/admission/defaulter_custom.go#L41-L45
type defaulterForType struct {
	object  IContextuallyDefaultableObject
	decoder *admission.Decoder
}

// InjectDecoder injects the decoder into a mutatingHandler.
//
// https://github.com/kubernetes-sigs/controller-runtime/blob/v0.13.1/pkg/webhook/admission/inject.go
func (h *defaulterForType) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

var _ admission.DecoderInjector = &defaulterForType{}

// Handle manages the inbound request from the API server. It's responsible for
// decoding the request into an
// [`admission.Request`](https://pkg.go.dev/k8s.io/api/admission/v1#AdmissionRequest)
// object, calling the `Default()` function on that object, and then returning
// back the patched response to the API server.
func (h *defaulterForType) Handle(_ context.Context, req admission.Request) admission.Response {
	// https://github.com/kubernetes-sigs/controller-runtime/blob/v0.13.1/pkg/webhook/admission/defaulter.go#L57-L59
	if h.object == nil {
		panic("object should never be nil")
	}

	// always skip when a DELETE operation received in mutation handler
	// describe in https://github.com/kubernetes-sigs/controller-runtime/issues/1762
	// https://github.com/kubernetes-sigs/controller-runtime/blob/v0.13.1/pkg/webhook/admission/defaulter.go#L61-L70
	if req.Operation == admissionv1.Delete {
		return admission.Response{AdmissionResponse: admissionv1.AdmissionResponse{
			Allowed: true,
			Result: &metav1.Status{
				Code: http.StatusOK,
			},
		}}
	}

	// Get the object in the request
	// https://github.com/kubernetes-sigs/controller-runtime/blob/v0.13.1/pkg/webhook/admission/defaulter.go#L72-L76
	obj := h.object.DeepCopyObject().(IContextuallyDefaultableObject)
	if err := h.decoder.Decode(req, obj); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	// Default the object
	//
	// orig: https://github.com/kubernetes-sigs/controller-runtime/blob/v0.13.1/pkg/webhook/admission/defaulter.go#L78-L83
	err := obj.Default(req)
	if err != nil {
		var apiStatus apierrors.APIStatus
		if errors.As(err, &apiStatus) {
			return validationResponseFromStatus(false, apiStatus.Status())
		}
		return admission.Denied(err.Error())
	}

	marshalled, err := json.Marshal(obj)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	// Create the patch
	return admission.PatchResponseFromRaw(req.Object.Raw, marshalled)
}
