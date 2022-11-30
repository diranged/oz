package webhook

import (
	"context"
	"errors"
	"net/http"

	v1 "k8s.io/api/admission/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// IContextuallyValidatableObject implements a similar pattern to the
// [`controller-runtime`](https://github.com/kubernetes-sigs/controller-runtime/tree/v0.13.1/pkg/webhook)
// webhook pattern. The difference is that the `Default()` function is not only
// supplied the request resource, but also the request context in the form of
// an
// [`admission.Request`](https://github.com/kubernetes-sigs/controller-runtime/blob/v0.13.1/pkg/webhook/admission/webhook.go#L43-L66)
// object.
//
// Modified from https://github.com/kubernetes-sigs/controller-runtime/blob/v0.13.1/pkg/webhook/admission/defaulter_custom.go#L29-L32
type IContextuallyValidatableObject interface {
	runtime.Object
	ValidateCreate(req admission.Request) error
	ValidateUpdate(req admission.Request, old runtime.Object) error
	ValidateDelete(req admission.Request) error
}

// RegisterContextualValidator leverages many of the patterns and code from the
// Controller-Runtime Admission package, but is one level _less_ abstracted.
// Rather than calling the `Default()` function on the target resource type,
func RegisterContextualValidator(
	obj IContextuallyValidatableObject,
	mgr ctrl.Manager,
) error {
	// Get the GroupVersionKind for the target schema object.
	gvk, err := apiutil.GVKForObject(obj, mgr.GetScheme())
	if err != nil {
		return err
	}
	path := generateValidatePath(gvk)

	// Create a Webhook{} resource with our Handler.
	mwh := &admission.Webhook{
		Handler: &validatorForType{object: obj},
	}

	// Insert the path into the webhook server and point it at our mutating
	// webhook handler. This must take place before the default controller
	// NewWebhookManagedBy().Complete() function is called.
	mgr.GetWebhookServer().Register(path, mwh)

	return nil
}

// A validatorForType mimics the
// [`validatorForType`](https://github.com/kubernetes-sigs/controller-runtime/blob/v0.13.1/pkg/webhook/admission/defaulter_custom.go)
// code, but understands to pass the `admission.Request` object into the `Default()` function.
//
// https://github.com/kubernetes-sigs/controller-runtime/blob/v0.13.1/pkg/webhook/admission/defaulter_custom.go#L41-L45
type validatorForType struct {
	object  IContextuallyValidatableObject
	decoder *admission.Decoder
}

// InjectDecoder injects the decoder into a mutatingHandler.
//
// https://github.com/kubernetes-sigs/controller-runtime/blob/v0.13.1/pkg/webhook/admission/inject.go
func (h *validatorForType) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

var _ admission.DecoderInjector = &validatorForType{}

// Handle manages the inbound request from the API server. It's responsible for
// decoding the request into an
// [`admission.Request`](https://pkg.go.dev/k8s.io/api/admission/v1#AdmissionRequest)
// object, calling the `Default()` function on that object, and then returning
// back the patched response to the API server.
// Handle handles admission requests.
//
// revive:disable:cyclomatic Replication of existing code in Controller-Runtime
func (h *validatorForType) Handle(_ context.Context, req admission.Request) admission.Response {
	// https://github.com/kubernetes-sigs/controller-runtime/blob/v0.13.1/pkg/webhook/admission/validator.go#L59-L62
	if h.object == nil {
		panic("object should never be nil")
	}

	// Get the object in the request
	//
	// https://github.com/kubernetes-sigs/controller-runtime/blob/v0.13.1/pkg/webhook/admission/validator.go#L63-L79
	obj := h.object.DeepCopyObject().(IContextuallyValidatableObject)
	if req.Operation == v1.Create {
		err := h.decoder.Decode(req, obj)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}

		err = obj.ValidateCreate(req)
		if err != nil {
			var apiStatus apierrors.APIStatus
			if errors.As(err, &apiStatus) {
				return validationResponseFromStatus(false, apiStatus.Status())
			}
			return admission.Denied(err.Error())
		}
	}
	if req.Operation == v1.Update {
		oldObj := obj.DeepCopyObject()

		err := h.decoder.DecodeRaw(req.Object, obj)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		err = h.decoder.DecodeRaw(req.OldObject, oldObj)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}

		err = obj.ValidateUpdate(req, oldObj)
		if err != nil {
			var apiStatus apierrors.APIStatus
			if errors.As(err, &apiStatus) {
				return validationResponseFromStatus(false, apiStatus.Status())
			}
			return admission.Denied(err.Error())
		}
	}

	if req.Operation == v1.Delete {
		// In reference to PR: https://github.com/kubernetes/kubernetes/pull/76346
		// OldObject contains the object being deleted
		err := h.decoder.DecodeRaw(req.OldObject, obj)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}

		err = obj.ValidateDelete(req)
		if err != nil {
			var apiStatus apierrors.APIStatus
			if errors.As(err, &apiStatus) {
				return validationResponseFromStatus(false, apiStatus.Status())
			}
			return admission.Denied(err.Error())
		}
	}

	return admission.Allowed("")
}
