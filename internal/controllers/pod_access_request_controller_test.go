package controllers

import (
	"context"

	api "github.com/diranged/oz/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("PodAccessRequestController", Ordered, func() {
	Context("Controller Test", func() {
		const TestName = "podaccesstemplatecontroller"

		var (
			namespace  *corev1.Namespace
			deployment *appsv1.Deployment
			ctx        = context.Background()
			request    *api.PodAccessRequest
			template   *api.PodAccessTemplate
		)

		// NOTE: We use a real k8sClient for these tests beacuse we need to
		// verify things like UID generation happening in the backend, as well
		// as generation spec updates.
		BeforeAll(func() {
			By("Creating the Namespace to perform the tests")
			namespace = &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: randomString(8),
				},
			}
			err := k8sClient.Create(ctx, namespace)
			Expect(err).To(Not(HaveOccurred()))
		})

		// Before each test case, we create a new Deployment, PodAccessRequest and PodAccessTemplate.
		BeforeEach(func() {
			// Create a fake deployment target
			deployment = &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-dep",
					Namespace: namespace.GetName(),
				},
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"testLabel": "testValue",
						},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{
								api.DefaultContainerAnnotationKey: "contb",
								"Foo":                             "bar",
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
								{
									Name:  "contb",
									Image: "nginx:latest",
								},
							},
						},
					},
				},
			}
			err := k8sClient.Create(ctx, deployment)
			Expect(err).To(Not(HaveOccurred()))

			// Create a default PodAccessTemplate. We'll mutate it for specific tests.
			template = &api.PodAccessTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-template",
					Namespace: deployment.Namespace,
				},
				Spec: api.PodAccessTemplateSpec{
					AccessConfig: api.AccessConfig{
						AllowedGroups:   []string{"testGroupA"},
						DefaultDuration: "1h",
						MaxDuration:     "2h",
					},
					ControllerTargetRef: &api.CrossVersionObjectReference{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
						Name:       deployment.Name,
					},
					ControllerTargetMutationConfig: &api.PodTemplateSpecMutationConfig{},
				},
			}
			err = k8sClient.Create(ctx, template)
			Expect(err).To(Not(HaveOccurred()))

			// Create a simple PodAccessRequest resource to test the template with
			request = &api.PodAccessRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-req",
					Namespace: template.Namespace,
				},
				Spec: api.PodAccessRequestSpec{
					TemplateName: template.Name,
					Duration:     "5m",
				},
			}
			err = k8sClient.Create(ctx, request)
			Expect(err).To(Not(HaveOccurred()))
		})

		// After each test, we wipe out the Deployment, Template and Request
		AfterEach(func() {
			err := k8sClient.Delete(ctx, deployment)
			Expect(err).To(Not(HaveOccurred()))
			err = k8sClient.Delete(ctx, request)
			Expect(err).To(Not(HaveOccurred()))
			err = k8sClient.Delete(ctx, template)
			Expect(err).To(Not(HaveOccurred()))
		})

		// One conceptual test here... our goal is primarily to test
		// reconciliation, not the deep internals about the Builders and the
		// particular settings that they are applying.
		It("Should successfully reconcile a simple Deployment and Template", func() {
			// Verify that reconciliation of the PodAccessTemplate succeeds on
			// the first attempt, without any failures reported.
			By("Reconciling the custom PodAccessTemplate first")
			tmplReconciler := &PodAccessTemplateReconciler{
				BaseTemplateReconciler: BaseTemplateReconciler{
					BaseReconciler: BaseReconciler{
						Client:    k8sClient,
						Scheme:    k8sClient.Scheme(),
						APIReader: k8sClient,
					},
				},
			}
			_, err := tmplReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      TestName,
					Namespace: namespace.Name,
				},
			})
			Expect(err).To(Not(HaveOccurred()))

			// Verify that the PodAccessRequest reconciliation actually does
			// NOT pass the first time. It should get about half way through,
			// and report an error (and requeue) while waiting for the desired
			// Pod to come up.
			By("Reconciling the custom PodAccessRequest first, which should error out")
			reqReconciler := &PodAccessRequestReconciler{
				BaseRequestReconciler: BaseRequestReconciler{
					BaseReconciler: BaseReconciler{
						Client:    k8sClient,
						Scheme:    k8sClient.Scheme(),
						APIReader: k8sClient,
					},
				},
			}
			_, err = reqReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      request.Name,
					Namespace: request.Namespace,
				},
			})
			Expect(err).To(Not(HaveOccurred()))

			// Verify that the Request is still in not-ready state, even though
			// the reconcile didn't error out.
			Expect(request.Status.IsReady()).To(BeFalse())

			// Refetch our Request object... reconiliation has mutated its
			// .Status fields.
			By("Refetching our Request...")
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      request.Name,
				Namespace: request.Namespace,
			}, request)
			Expect(err).To(Not(HaveOccurred()))

			// Verify that the request access message was set
			By("Verifying the Status.AccessMessage is set, but ready state is false")
			Expect(
				request.Status.AccessMessage,
			).To(MatchRegexp("kubectl exec -ti -n .* .* -- /bin/sh"))
			Expect(request.Status.IsReady()).To(BeFalse())

			// Patch the pod status state to "Running" so we can simulate that
			// the pod is actually up now.
			By("Patching the Pod State...")
			pod := &corev1.Pod{}
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      request.Status.PodName,
				Namespace: request.Namespace,
			}, pod)
			Expect(err).To(Not(HaveOccurred()))

			// Update it
			pod.Status.Phase = corev1.PodRunning

			// Update the status, handle failure.
			err = k8sClient.Status().Update(ctx, pod)
			Expect(err).To(Not(HaveOccurred()))

			// Refetch it
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      pod.Name,
				Namespace: pod.Namespace,
			}, pod)
			Expect(pod.Status.Phase).To(Equal(corev1.PodRunning))
			Expect(err).To(Not(HaveOccurred()))

			// Now we verify that on the second reconciliation there are no
			// errors and we get all the way through the process.
			By("Reconciling again, expecting success")
			_, err = reqReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      request.Name,
					Namespace: request.Namespace,
				},
			})
			Expect(err).To(Not(HaveOccurred()))

			// Lastly, make sure we updated the request status with
			// Status.Ready=True.
			By("Verifying that the PodAccessRequest IS ready")
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      request.Name,
				Namespace: request.Namespace,
			}, request)
			Expect(err).To(Not(HaveOccurred()))
			Expect(request.Status.IsReady()).To(BeTrue())
		})
	})
})
