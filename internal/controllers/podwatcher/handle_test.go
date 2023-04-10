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
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var _ = Describe("PodWatcher", Ordered, func() {
	decoder, _ := admission.NewDecoder(scheme.Scheme)

	Context("Functional Unit Tests", func() {
		var (
			admissionRequest *admission.Request
			ctx              = context.Background()
			requestName      = "test"
			recorder         = record.NewFakeRecorder(50)
			watcher          = &PodWatcher{
				Client:   k8sClient,
				decoder:  decoder,
				recorder: recorder,
			}
			resource = metav1.GroupVersionResource{
				Group:    corev1.SchemeGroupVersion.Group,
				Version:  corev1.SchemeGroupVersion.Version,
				Resource: corev1.Pod{}.Kind,
			}
		)

		It("Handle() with non-CONNECT calls should just pass (for now)", func() {
			admissionRequest = &admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					Operation: admissionv1.Create,
				},
			}
			resp := watcher.Handle(ctx, *admissionRequest)
			Expect(resp.Result.Code).To(Equal(int32(http.StatusOK)))
			Expect(resp.Allowed).To(BeTrue())
		})

		It("Handle() CONNECT calls with invalid data should just pass (for now)", func() {
			admissionRequest = &admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					Operation: admissionv1.Connect,
					Object: runtime.RawExtension{
						Raw: []byte("asdf"),
					},
				},
			}
			resp := watcher.Handle(ctx, *admissionRequest)
			Expect(resp.Result.Code).To(Equal(int32(http.StatusOK)))
			Expect(resp.Allowed).To(BeTrue())
		})

		It("Handle() CONNECT calls for a pods/exec action should call HandleExec() and work", func() {
			// Create a fresh recorder that can accept ONE event
			recorder := record.NewFakeRecorder(1)
			watcher.recorder = recorder

			// Pod Exec Options for the request
			request := &corev1.PodExecOptions{
				TypeMeta: metav1.TypeMeta{
					Kind:       "PodExecOptions",
					APIVersion: "v1",
				},
				Stdin:     false,
				Stdout:    false,
				Stderr:    false,
				TTY:       true,
				Container: "fooContainer",
				Command:   []string{},
			}
			requestBytes, _ := json.Marshal(request)

			// The admission request itself with the PodExecOptions baked in
			admissionRequest = &admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					Resource: resource,
					Kind: metav1.GroupVersionKind{
						Group:   request.GroupVersionKind().Group,
						Version: request.APIVersion,
						Kind:    request.Kind,
					},
					RequestKind: &metav1.GroupVersionKind{
						Group:   request.GroupVersionKind().Group,
						Version: request.APIVersion,
						Kind:    request.Kind,
					},
					RequestResource: &resource,
					Name:            requestName,
					Namespace:       requestName,
					Operation:       admissionv1.Connect,
					SubResource:     "exec",
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

			// Handle the event
			resp := watcher.Handle(ctx, *admissionRequest)

			// Close out our events recorder
			close(recorder.Events)

			// Make sure we passed
			Expect(resp.Allowed).To(BeTrue())

			//
			msg := <-recorder.Events
			Expect(msg).To(Equal("Normal PodExec CONNECT/exec operation on Pod test (container: fooContainer) by admin (interactive: true)"))
		})

		It("Handle() CONNECT calls for a pods/attach action should call HandleAttach() and work", func() {
			// Create a fresh recorder that can accept ONE event
			recorder := record.NewFakeRecorder(1)
			watcher.recorder = recorder

			// Pod Exec Options for the request
			request := &corev1.PodAttachOptions{
				TypeMeta: metav1.TypeMeta{
					Kind:       "PodAttachOptions",
					APIVersion: "v1",
				},
				Stdin:     false,
				Stdout:    false,
				Stderr:    false,
				TTY:       true,
				Container: "fooContainer",
			}
			requestBytes, _ := json.Marshal(request)

			// The admission request itself with the PodExecOptions baked in
			admissionRequest = &admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					Resource: resource,
					Kind: metav1.GroupVersionKind{
						Group:   request.GroupVersionKind().Group,
						Version: request.APIVersion,
						Kind:    request.Kind,
					},
					RequestKind: &metav1.GroupVersionKind{
						Group:   request.GroupVersionKind().Group,
						Version: request.APIVersion,
						Kind:    request.Kind,
					},
					RequestResource: &resource,
					Name:            requestName,
					Namespace:       requestName,
					Operation:       admissionv1.Connect,
					SubResource:     "attach",
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

			// Handle the event
			resp := watcher.Handle(ctx, *admissionRequest)

			// Close out our events recorder
			close(recorder.Events)

			// Make sure we passed
			Expect(resp.Allowed).To(BeTrue())

			//
			msg := <-recorder.Events
			Expect(msg).To(Equal("Normal PodAttach CONNECT/attach operation on Pod test (container: fooContainer) by admin (interactive: true)"))
		})
	})
})
