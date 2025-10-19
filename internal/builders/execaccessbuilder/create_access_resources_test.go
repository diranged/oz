package execaccessbuilder

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/diranged/oz/internal/api/v1alpha1"
	"github.com/diranged/oz/internal/builders/execaccessbuilder/internal"
	bldutil "github.com/diranged/oz/internal/builders/utils"
	"github.com/diranged/oz/internal/testing/utils"
)

var _ = Describe("RequestReconciler", Ordered, func() {
	Context("CreateAccessResources()", func() {
		var (
			ctx        = context.Background()
			ns         *corev1.Namespace
			deployment *appsv1.Deployment
			pod        *corev1.Pod
			request    *v1alpha1.ExecAccessRequest
			template   *v1alpha1.ExecAccessTemplate
			builder    = ExecAccessBuilder{}
		)

		// For Envtest
		podselection.PodPhaseRunning = "Pending"

		BeforeAll(func() {
			By("Should have a namespace to execute tests in")
			ns = &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: testutil.RandomString(8),
				},
			}
			err := k8sClient.Create(ctx, ns)
			Expect(err).ToNot(HaveOccurred())

			By("Creating a Deployment to reference for the test")
			deployment = &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      testutil.RandomString(4),
					Namespace: ns.Name,
				},
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"testLabel": "testValue",
						},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"testLabel": "testValue",
							},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "test",
									Image: "nginx:latest",
								},
							},
						},
					},
				},
			}
			err = k8sClient.Create(ctx, deployment)
			Expect(err).To(Not(HaveOccurred()))

			By("Create a single Pod that should match the Deployment spec above for testing")
			pod = &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      testutil.RandomString(8),
					Namespace: ns.GetName(),
					Labels:    deployment.Spec.Selector.MatchLabels,
				},
				Spec: deployment.Spec.Template.Spec,
				Status: corev1.PodStatus{
					Phase: "Running",
				},
			}
			err = k8sClient.Create(ctx, pod)
			Expect(err).To(Not(HaveOccurred()))

			By("Should have an ExecAccessTemplate to test against")
			template = &v1alpha1.ExecAccessTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      testutil.RandomString(8),
					Namespace: ns.GetName(),
				},
				Spec: v1alpha1.ExecAccessTemplateSpec{
					AccessConfig: v1alpha1.AccessConfig{
						AllowedGroups:   []string{"foo"},
						DefaultDuration: "1h",
						MaxDuration:     "2h",
					},
					ControllerTargetRef: &v1alpha1.CrossVersionObjectReference{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
						Name:       deployment.GetName(),
					},
				},
			}
			err = k8sClient.Create(ctx, template)
			Expect(err).ToNot(HaveOccurred())

			By("Should have an ExecAccessRequest built to test against")
			request = &v1alpha1.ExecAccessRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "createaccessresource-test",
					Namespace: ns.GetName(),
				},
				Spec: v1alpha1.ExecAccessRequestSpec{
					TemplateName: template.GetName(),
				},
			}
			err = k8sClient.Create(ctx, request)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterAll(func() {
			By("Should delete the namespace")
			err := k8sClient.Delete(ctx, ns)
			Expect(err).ToNot(HaveOccurred())
		})

		It(
			"CreateAccessResources() should return status.podName regardless of requested target pod",
			func() {
				// Create a test pod that we're going to slide in as the
				// already-assigned pod for the access request.
				p := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testutil.RandomString(8),
						Namespace: ns.GetName(),
					},
					Spec: deployment.Spec.Template.Spec,
					Status: corev1.PodStatus{
						Phase: "Running",
					},
				}
				err := k8sClient.Create(ctx, p)
				Expect(err).To(Not(HaveOccurred()))

				// Now hack the request and override its assigned pod
				request.Status.PodName = p.GetName()

				// But set the TargetPod to some arbitrary string
				request.Spec.TargetPod = "junkPod"

				// Execute
				ret, err := builder.CreateAccessResources(ctx, k8sClient, request, template)

				// VERIFY: No errors
				Expect(err).ToNot(HaveOccurred())
				Expect(ret).To(MatchRegexp("Success"))

				// VERIFY: Status string looks roughly right
				Expect(ret).To(MatchRegexp(fmt.Sprintf(
					"Success. Role %s-.*, RoleBinding %s.* created",
					request.GetName(),
					request.GetName(),
				)))
			},
		)

		It("CreateAccessResources() should succeed with user-specified pod", func() {
			request.Status.PodName = ""
			request.Spec.TargetPod = pod.GetName()
			ret, err := builder.CreateAccessResources(ctx, k8sClient, request, template)
			Expect(err).ToNot(HaveOccurred())

			// VERIFY: Status string looks roughly right
			Expect(ret).To(MatchRegexp(fmt.Sprintf(
				"Success. Role %s-.*, RoleBinding %s.* created",
				request.GetName(),
				request.GetName(),
			)))
		})

		It("CreateAccessResources() should fail if pod is missing", func() {
			request.Status.PodName = ""
			request.Spec.TargetPod = "testPod"
			_, err := builder.CreateAccessResources(ctx, k8sClient, request, template)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp("not found"))
		})

		It("CreateAccessResources() should succeed with random pod selection", func() {
			request.Status.PodName = ""
			request.Spec.TargetPod = ""

			ret, err := builder.CreateAccessResources(ctx, k8sClient, request, template)

			// VERIFY: No error returned
			Expect(err).ToNot(HaveOccurred())

			// VERIFY: Proper status string returned
			Expect(ret).To(MatchRegexp(fmt.Sprintf(
				"Success. Role %s-.*, RoleBinding %s.* created",
				request.GetName(),
				request.GetName(),
			)))

			// VERIFY: Role Created as expected
			foundRole := &rbacv1.Role{}
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      bldutil.GenerateResourceName(request),
				Namespace: ns.GetName(),
			}, foundRole)
			Expect(err).ToNot(HaveOccurred())
			Expect(foundRole.GetOwnerReferences()).ToNot(BeNil())
			Expect(foundRole.Rules[0].ResourceNames[0]).To(Equal(pod.GetName()))
			Expect(foundRole.Rules[1].ResourceNames[0]).To(Equal(pod.GetName()))

			// VERIFY: RoleBinding Created as expected
			foundRoleBinding := &rbacv1.RoleBinding{}
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      bldutil.GenerateResourceName(request),
				Namespace: ns.GetName(),
			}, foundRoleBinding)
			Expect(err).ToNot(HaveOccurred())
			Expect(foundRoleBinding.GetOwnerReferences()).ToNot(BeNil())
			Expect(foundRoleBinding.RoleRef.Name).To(Equal(foundRole.GetName()))
			Expect(foundRoleBinding.Subjects[0].Name).To(Equal("foo"))
		})
	})
})
