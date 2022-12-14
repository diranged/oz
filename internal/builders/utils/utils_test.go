package utils

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/diranged/oz/internal/api/v1alpha1"
	"github.com/diranged/oz/internal/testing/utils"
)

var _ = Describe("IBuilder / Utils", Ordered, func() {
	Context("Functions()", func() {
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
					Name: utils.RandomString(8),
				},
			}
			err := k8sClient.Create(ctx, namespace)
			Expect(err).To(Not(HaveOccurred()))
		})

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

		AfterEach(func() {
			err := k8sClient.Delete(ctx, deployment)
			Expect(err).To(Not(HaveOccurred()))
			err = k8sClient.Delete(ctx, request)
			Expect(err).To(Not(HaveOccurred()))
			err = k8sClient.Delete(ctx, template)
			Expect(err).To(Not(HaveOccurred()))
		})

		It("getShortUID should work", func() {
			ret := getShortUID(request)
			Expect(len(ret)).To(Equal(8))
		})

		It("generateResourceName should work", func() {
			ret := GenerateResourceName(request)
			Expect(len(ret)).To(Equal(17))
		})
	})
})
