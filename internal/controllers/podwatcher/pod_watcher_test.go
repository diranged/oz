package podwatcher

import (
	"context"
	"encoding/json"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	admissionv1 "k8s.io/api/admission/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var _ = Describe("PodWatcher", Ordered, func() {
	decoder, _ := admission.NewDecoder(scheme.Scheme)

	Context("Functional Unit Tests", func() {
		var (
			admissionRequest *admission.Request
			ctx              = context.Background()
			requestName      = "test"
			watcher          = &PodExecWatcher{Client: k8sClient, decoder: decoder}
			request          = &corev1.PodExecOptions{
				TypeMeta: metav1.TypeMeta{
					Kind:       "",
					APIVersion: "",
				},
				Stdin:     false,
				Stdout:    false,
				Stderr:    false,
				TTY:       false,
				Container: "",
				Command:   []string{},
			}
			resource = metav1.GroupVersionResource{
				Group:    corev1.SchemeGroupVersion.Group,
				Version:  corev1.SchemeGroupVersion.Version,
				Resource: corev1.Pod{}.Kind,
			}
			kind = metav1.GroupVersionKind{
				Group:   api.SchemeGroupVersion.Group,
				Version: api.SchemeGroupVersion.Version,
				Kind:    corev1.PodExecOptions{}.Kind,
			}
		)

		It("Handle() calls with invalid data should return an error", func() {
			admissionRequest = &admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					Resource:        resource,
					RequestKind:     &kind,
					RequestResource: &resource,
					Name:            requestName,
					Namespace:       requestName,
					Operation:       "CONNET",
					UserInfo: authenticationv1.UserInfo{
						Username: "admin",
						UID:      "",
						Groups:   []string{},
						Extra: map[string]authenticationv1.ExtraValue{
							"": {},
						},
					},
					Object: runtime.RawExtension{
						Raw: []byte("asdf"),
					},
				},
			}
			resp := watcher.Handle(ctx, *admissionRequest)
			Expect(resp.Result.Code).To(Equal(int32(http.StatusBadRequest)))
			Expect(resp.Allowed).To(BeFalse())
		})

		// TODO: The "Userinfo" checks should move into an authentication
		// package so that we can write one set of tests for all of the
		// Validate* functions.
		It("Create calls with req.Userinfo{} should suceed", func() {
			requestBytes, _ := json.Marshal(request)
			admissionRequest = &admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					Resource:        resource,
					RequestKind:     &kind,
					RequestResource: &resource,
					Name:            requestName,
					Namespace:       requestName,
					Operation:       "CONNET",
					UserInfo: authenticationv1.UserInfo{
						Username: "admin",
						UID:      "",
						Groups:   []string{},
						Extra: map[string]authenticationv1.ExtraValue{
							"": {},
						},
					},
					Object: runtime.RawExtension{
						Raw: requestBytes,
					},
				},
			}
			resp := watcher.Handle(ctx, *admissionRequest)
			Expect(resp.Allowed).To(BeTrue())
		})
	})
})
