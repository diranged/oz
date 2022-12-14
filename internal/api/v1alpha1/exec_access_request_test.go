package v1alpha1

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	admissionv1 "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// The ExecAccessRequest tests are primarily testing the behavior of the
// ExecAccessRequest struct - but because some of our tests are testing against
// the real Kubernetes API, they indirectly are also acting as tests of the
// ExecAccessRequestController Reconciler().
var _ = Describe("ExecAccessRequest", Ordered, func() {
	var namespace *corev1.Namespace
	var deployment *appsv1.Deployment
	var template *ExecAccessTemplate

	// These tests create real ExecAccessRequest{} objects in the cluster and
	// validate behavior. This indirectly tests both the reconciler code, AND
	// directly is testing the Webhook code. These are functional tests to
	// ensure basic functionality.
	//
	// Explicit test-cases for function logic is handled in the next Context()
	// below.
	Context("Reconciliation / Webhook Tests", func() {
		requestName := "test"

		It("Creation of the ExecAccessRequest to work - Passes Webhook", func() {
			request := &ExecAccessRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      requestName,
					Namespace: template.Namespace,
				},
				Spec: ExecAccessRequestSpec{
					TemplateName: template.Name,
					Duration:     "1h",
				},
			}
			err := k8sClient.Create(ctx, request)
			Expect(err).To(Not(HaveOccurred()))
		})

		It("ExecAccessRequest becomes ready - Passes Reconciliation", func() {
			request := &ExecAccessRequest{}
			Eventually(func() error {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      requestName,
					Namespace: template.Namespace,
				}, request)
				Expect(err).To(Not(HaveOccurred()))
				if request.GetStatus().IsReady() {
					return nil
				}
				return fmt.Errorf(
					"Failed to reconcile resource: %s",
					strconv.FormatBool(request.GetStatus().IsReady()),
				)
			}, time.Minute, time.Second).Should(HaveOccurred())
		})

		It("ExecAccessRequest Update - Passes Webhook ValidateUpdate() Call", func() {
			// Get the request first
			request := &ExecAccessRequest{}
			err := k8sClient.Get(ctx, types.NamespacedName{
				Name:      requestName,
				Namespace: template.Namespace,
			}, request)
			Expect(err).To(Not(HaveOccurred()))

			// Update it and push it
			request.ObjectMeta.SetAnnotations(map[string]string{"foo": "bar"})
			err = k8sClient.Update(ctx, request)
			Expect(err).To(Not(HaveOccurred()))
		})
		It("ExecAccessRequest Delete - Passes Webhook ValidateUpdate() Call", func() {
			// Get the request first
			request := &ExecAccessRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      requestName,
					Namespace: template.Namespace,
				},
			}
			err := k8sClient.Delete(ctx, request)
			Expect(err).To(Not(HaveOccurred()))
		})
	})

	// This Context() tests specific functions - no real calls against the API
	// are made here. Fake data is passed into the functions to verify that
	// they each behave correctly under different conditions.
	Context("Functional Unit Tests", func() {
		var (
			err              error
			requestName      = "test"
			admissionRequest *admission.Request

			request = &ExecAccessRequest{
				Spec: ExecAccessRequestSpec{
					TemplateName: "",
					Duration:     "",
				},
			}

			gvr = metav1.GroupVersionResource{
				Group:    api.SchemeGroupVersion.Group,
				Version:  api.SchemeGroupVersion.Version,
				Resource: "execaccessrequest",
			}
			gvk = metav1.GroupVersionKind{
				Group:   api.SchemeGroupVersion.Group,
				Version: api.SchemeGroupVersion.Version,
				Kind:    ExecAccessRequest{}.Kind,
			}
		)

		// TODO: The "Userinfo" checks should move into an authentication
		// package so that we can write one set of tests for all of the
		// Validate* functions.
		It("Create with UserInfo...", func() {
			requestBytes, _ := json.Marshal(request)
			admissionRequest = &admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					Resource:        gvr,
					RequestKind:     &gvk,
					RequestResource: &gvr,
					Name:            requestName,
					Namespace:       namespace.Name,
					Operation:       "CREATE",
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
			err = request.ValidateCreate(*admissionRequest)
			Expect(err).To(Not(HaveOccurred()))
		})

		It("Create without UserInfo...", func() {
			requestBytes, _ := json.Marshal(request)
			admissionRequest = &admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					Resource:        gvr,
					RequestKind:     &gvk,
					RequestResource: &gvr,
					Name:            requestName,
					Namespace:       namespace.Name,
					Operation:       "CREATE",
					UserInfo:        authenticationv1.UserInfo{},
					Object: runtime.RawExtension{
						Raw: requestBytes,
					},
				},
			}
			err = request.ValidateCreate(*admissionRequest)
			Expect(err).To(Not(HaveOccurred()))
		})

		It("Update with UserInfo...", func() {
			requestBytes, _ := json.Marshal(request)
			admissionRequest = &admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					Resource:        gvr,
					RequestKind:     &gvk,
					RequestResource: &gvr,
					Name:            requestName,
					Namespace:       namespace.Name,
					Operation:       "UPDATE",
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
			err = request.ValidateUpdate(*admissionRequest, request)
			Expect(err).To(Not(HaveOccurred()))
		})

		It("Update without UserInfo...", func() {
			requestBytes, _ := json.Marshal(request)
			admissionRequest = &admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					Resource:        gvr,
					RequestKind:     &gvk,
					RequestResource: &gvr,
					Name:            requestName,
					Namespace:       namespace.Name,
					Operation:       "UPDATE",
					UserInfo:        authenticationv1.UserInfo{},
					Object: runtime.RawExtension{
						Raw: requestBytes,
					},
				},
			}
			err = request.ValidateUpdate(*admissionRequest, request)
			Expect(err).To(Not(HaveOccurred()))
		})
	})

	// Setup code below here - this code rarely changes, the tests above are
	// much more important.
	BeforeAll(func() {
		By("Creating the Namespace to perform the tests")
		namespace = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: randomString(8),
			},
		}
		err := k8sClient.Create(ctx, namespace)
		Expect(err).To(Not(HaveOccurred()))

		By("Creating the Deployment to perform the tests")
		deployment = &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-deployment",
				Namespace: namespace.Name,
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"testLabel": "testValue",
					},
				},
				Template: v1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"Foo": "bar",
						},
						Labels: map[string]string{
							"testLabel": "testValue",
						},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "conta",
								Image: "nginx:latest",
							},
						},
					},
				},
			},
		}
		err = k8sClient.Create(ctx, deployment)
		Expect(err).To(Not(HaveOccurred()))
	})

	AfterAll(func() {
		By("Deleting the Namespace for tests")
		err := k8sClient.Delete(ctx, namespace)
		Expect(err).To(Not(HaveOccurred()))
	})

	BeforeEach(func() {
		var err error

		// Create the execaccesstemplate
		template = &ExecAccessTemplate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: namespace.Name,
			},
			Spec: ExecAccessTemplateSpec{
				AccessConfig: AccessConfig{
					AllowedGroups:   []string{"admins"},
					DefaultDuration: "1h",
					MaxDuration:     "24h",
				},
				ControllerTargetRef: &CrossVersionObjectReference{
					APIVersion: "apps/v1",
					Kind:       "Deployment",
					Name:       deployment.Name,
				},
			},
		}
		err = k8sClient.Create(ctx, template)
		Expect(err).To(Not(HaveOccurred()))

		// Verify the template is ready
		Eventually(func() error {
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      template.Name,
				Namespace: template.Namespace,
			}, template)
			Expect(err).To(Not(HaveOccurred()))

			if template.GetStatus().IsReady() {
				return nil
			}
			return fmt.Errorf(
				"Failed to reconcile resource: %s",
				strconv.FormatBool(template.GetStatus().IsReady()),
			)
		}, time.Minute, time.Second).Should(HaveOccurred())
	})

	AfterEach(func() {
		// Clear out the Template after each test
		err := k8sClient.Delete(ctx, template)
		Expect(err).To(Not(HaveOccurred()))
	})
})
