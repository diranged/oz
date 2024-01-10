package podaccessbuilder

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	rolloutsv1alpha1 "github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/diranged/oz/internal/api/v1alpha1"
	bldutil "github.com/diranged/oz/internal/builders/utils"
	"github.com/diranged/oz/internal/testing/utils"
)

var _ = Describe("RequestReconciler", Ordered, func() {
	Context("CreateAccessResources()", func() {
		var (
			ctx             = context.Background()
			ns              *corev1.Namespace
			deployment      *appsv1.Deployment
			request         *v1alpha1.PodAccessRequest
			rolloutRequest  *v1alpha1.PodAccessRequest
			template        *v1alpha1.PodAccessTemplate
			rolloutTemplate *v1alpha1.PodAccessTemplate
			builder         = PodAccessBuilder{}
		)

		BeforeAll(func() {
			By("Should have a namespace to execute tests in")
			ns = &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: utils.RandomString(8),
				},
			}
			err := k8sClient.Create(ctx, ns)
			Expect(err).ToNot(HaveOccurred())

			By("Creating a Deployment to reference for the test")
			deployment = &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "deployment-test",
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

			By("Creating a Rollout to reference for the test")
			rollout := &rolloutsv1alpha1.Rollout{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "rollout-test",
					Namespace: ns.Name,
				},
				Spec: rolloutsv1alpha1.RolloutSpec{
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

			err = k8sClient.Create(ctx, rollout)
			Expect(err).To(Not(HaveOccurred()))

			By("Should have an PodAccessTemplate to test against")
			cpuReq, _ := resource.ParseQuantity("1")
			template = &v1alpha1.PodAccessTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      utils.RandomString(8),
					Namespace: ns.GetName(),
				},
				Spec: v1alpha1.PodAccessTemplateSpec{
					AccessConfig: v1alpha1.AccessConfig{
						AllowedGroups:   []string{"testGroupA"},
						DefaultDuration: "1h",
						MaxDuration:     "2h",
					},
					ControllerTargetRef: &v1alpha1.CrossVersionObjectReference{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
						Name:       deployment.Name,
					},
					ControllerTargetMutationConfig: &v1alpha1.PodTemplateSpecMutationConfig{
						DefaultContainerName: "test",
						Command:              &[]string{"/bin/sleep"},
						Args:                 &[]string{"100"},
						Env: []corev1.EnvVar{
							{Name: "FOO", Value: "BAR"},
						},
						Resources: corev1.ResourceRequirements{
							Requests: map[corev1.ResourceName]resource.Quantity{
								"cpu": cpuReq,
							},
						},
						PatchSpecOperations: []map[string]string{
							{
								"op":    "replace",
								"path":  "/spec/containers/0/name",
								"value": "oz",
							},
						},
					},
				},
			}
			err = k8sClient.Create(ctx, template)
			Expect(err).ToNot(HaveOccurred())

			rolloutTemplate = &v1alpha1.PodAccessTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      utils.RandomString(8),
					Namespace: ns.GetName(),
				},
				Spec: v1alpha1.PodAccessTemplateSpec{
					AccessConfig: v1alpha1.AccessConfig{
						AllowedGroups:   []string{"testGroupA"},
						DefaultDuration: "1h",
						MaxDuration:     "2h",
					},
					ControllerTargetRef: &v1alpha1.CrossVersionObjectReference{
						APIVersion: "argoproj.io/v1alpha1",
						Kind:       "Rollout",
						Name:       "rollout-test",
					},
					ControllerTargetMutationConfig: &v1alpha1.PodTemplateSpecMutationConfig{
						DefaultContainerName: "test",
						Command:              &[]string{"/bin/sleep"},
						Args:                 &[]string{"100"},
						Env: []corev1.EnvVar{
							{Name: "FOO", Value: "BAR"},
						},
						Resources: corev1.ResourceRequirements{
							Requests: map[corev1.ResourceName]resource.Quantity{
								"cpu": cpuReq,
							},
						},
					},
				},
			}
			err = k8sClient.Create(ctx, rolloutTemplate)
			Expect(err).ToNot(HaveOccurred())

			By("Should have an PodAccessRequest built to test against")
			request = &v1alpha1.PodAccessRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "createaccessresource-test",
					Namespace: ns.GetName(),
				},
				Spec: v1alpha1.PodAccessRequestSpec{
					TemplateName: template.GetName(),
				},
			}
			err = k8sClient.Create(ctx, request)
			Expect(err).ToNot(HaveOccurred())

			// verify podaccess request with Rollout
			rolloutRequest = &v1alpha1.PodAccessRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "createaccessresource-rollout-test",
					Namespace: ns.GetName(),
				},
				Spec: v1alpha1.PodAccessRequestSpec{
					TemplateName: rolloutTemplate.GetName(),
				},
			}
			err = k8sClient.Create(ctx, rolloutRequest)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterAll(func() {
			By("Should delete the namespace")
			err := k8sClient.Delete(ctx, ns)
			Expect(err).ToNot(HaveOccurred())
		})

		It("CreateAccessResources() should succeed", func() {
			request.Status.PodName = ""

			// Execute
			ret, err := builder.CreateAccessResources(ctx, k8sClient, request, template)

			// VERIFY: No error returned
			Expect(err).ToNot(HaveOccurred())

			// VERIFY: Proper status string returned
			Expect(ret).To(MatchRegexp(fmt.Sprintf(
				"Success. Pod %s-.*, Role %s-.*, RoleBinding %s.* created",
				request.GetName(),
				request.GetName(),
				request.GetName(),
			)))

			// VERIFY: Pod Created as expected
			foundPod := &corev1.Pod{}
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      bldutil.GenerateResourceName(request),
				Namespace: ns.GetName(),
			}, foundPod)
			Expect(err).ToNot(HaveOccurred())
			Expect(foundPod.GetOwnerReferences()).ToNot(BeNil())
			Expect(foundPod.Spec.Containers[0].Command[0]).To(Equal("/bin/sleep"))
			Expect(foundPod.Spec.Containers[0].Args[0]).To(Equal("100"))
			Expect(foundPod.Spec.Containers[0].Name).To(Equal("oz"))

			// VERIFY: Role Created as expected
			foundRole := &rbacv1.Role{}
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      bldutil.GenerateResourceName(request),
				Namespace: ns.GetName(),
			}, foundRole)
			Expect(err).ToNot(HaveOccurred())
			Expect(foundRole.GetOwnerReferences()).ToNot(BeNil())
			Expect(foundRole.Rules[0].ResourceNames[0]).To(Equal(foundPod.GetName()))
			Expect(foundRole.Rules[1].ResourceNames[0]).To(Equal(foundPod.GetName()))

			// VERIFY: RoleBinding Created as expected
			foundRoleBinding := &rbacv1.RoleBinding{}
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      bldutil.GenerateResourceName(request),
				Namespace: ns.GetName(),
			}, foundRoleBinding)
			Expect(err).ToNot(HaveOccurred())
			Expect(foundRoleBinding.GetOwnerReferences()).ToNot(BeNil())
			Expect(foundRoleBinding.RoleRef.Name).To(Equal(foundRole.GetName()))
			Expect(foundRoleBinding.Subjects[0].Name).To(Equal("testGroupA"))
		})

		It("CreateAccessResources() should succeed with Rollout", func() {
			rolloutRequest.Status.PodName = ""

			// Execute
			ret, err := builder.CreateAccessResources(ctx, k8sClient, rolloutRequest, rolloutTemplate)

			// VERIFY: No error returned
			Expect(err).ToNot(HaveOccurred())

			// VERIFY: Proper status string returned
			Expect(ret).To(MatchRegexp(fmt.Sprintf(
				"Success. Pod %s-.*, Role %s-.*, RoleBinding %s.* created",
				rolloutRequest.GetName(),
				rolloutRequest.GetName(),
				rolloutRequest.GetName(),
			)))

			// VERIFY: Pod Created as expected
			foundPod := &corev1.Pod{}
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      bldutil.GenerateResourceName(rolloutRequest),
				Namespace: ns.GetName(),
			}, foundPod)
			Expect(err).ToNot(HaveOccurred())
			Expect(foundPod.GetOwnerReferences()).ToNot(BeNil())
			Expect(foundPod.Spec.Containers[0].Command[0]).To(Equal("/bin/sleep"))
			Expect(foundPod.Spec.Containers[0].Args[0]).To(Equal("100"))

			// VERIFY: Role Created as expected
			foundRole := &rbacv1.Role{}
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      bldutil.GenerateResourceName(rolloutRequest),
				Namespace: ns.GetName(),
			}, foundRole)
			Expect(err).ToNot(HaveOccurred())
			Expect(foundRole.GetOwnerReferences()).ToNot(BeNil())
			Expect(foundRole.Rules[0].ResourceNames[0]).To(Equal(foundPod.GetName()))
			Expect(foundRole.Rules[1].ResourceNames[0]).To(Equal(foundPod.GetName()))

			// VERIFY: RoleBinding Created as expected
			foundRoleBinding := &rbacv1.RoleBinding{}
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      bldutil.GenerateResourceName(rolloutRequest),
				Namespace: ns.GetName(),
			}, foundRoleBinding)
			Expect(err).ToNot(HaveOccurred())
			Expect(foundRoleBinding.GetOwnerReferences()).ToNot(BeNil())
			Expect(foundRoleBinding.RoleRef.Name).To(Equal(foundRole.GetName()))
			Expect(foundRoleBinding.Subjects[0].Name).To(Equal("testGroupA"))
		})
	})
})
