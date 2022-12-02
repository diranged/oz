package webhook

import (
	"strings"

	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const delimiter = "-"

// Copy-Pasta from https://github.com/kubernetes-sigs/controller-runtime/blob/v0.13.1/pkg/builder/webhook.go#L208-L216
func generateMutatePath(gvk schema.GroupVersionKind) string {
	return "/mutate" + delimiter + strings.ReplaceAll(gvk.Group, ".", delimiter) + delimiter +
		gvk.Version + delimiter + strings.ToLower(gvk.Kind)
}

// Copy-Pasta from https://github.com/kubernetes-sigs/controller-runtime/blob/v0.13.1/pkg/builder/webhook.go#L208-L216
func generateValidatePath(gvk schema.GroupVersionKind) string {
	return "/validate" + delimiter + strings.ReplaceAll(gvk.Group, ".", delimiter) + delimiter +
		gvk.Version + delimiter + strings.ToLower(gvk.Kind)
}

// validationResponseFromStatus returns a response for admitting a request with provided Status object.
//
// https://github.com/kubernetes-sigs/controller-runtime/blob/v0.13.1/pkg/webhook/admission/response.go#L105-L114
func validationResponseFromStatus(allowed bool, status metav1.Status) admission.Response {
	resp := admission.Response{
		AdmissionResponse: admissionv1.AdmissionResponse{
			Allowed: allowed,
			Result:  &status,
		},
	}
	return resp
}
