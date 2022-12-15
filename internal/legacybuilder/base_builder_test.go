package legacybuilder

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	api "github.com/diranged/oz/internal/api/v1alpha1"
	"github.com/diranged/oz/internal/testing/utils"
)

var _ = Describe("BaseBuilder", Ordered, func() {
	Context("Functions()", func() {
		var (
			namespace  *corev1.Namespace
			deployment *appsv1.Deployment
			ctx        = context.Background()
			request    *api.PodAccessRequest
			template   *api.PodAccessTemplate
			builder    *BaseBuilder
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

			// Create the PodAccessBuilder finally - fully populated with the
			// Request, Template and fake clients.
			builder = &BaseBuilder{
				Client:    k8sClient,
				Ctx:       ctx,
				APIReader: k8sClient,
				Request:   request,
				Template:  template,
			}
		})

		AfterEach(func() {
			err := k8sClient.Delete(ctx, deployment)
			Expect(err).To(Not(HaveOccurred()))
			err = k8sClient.Delete(ctx, request)
			Expect(err).To(Not(HaveOccurred()))
			err = k8sClient.Delete(ctx, template)
			Expect(err).To(Not(HaveOccurred()))
		})

		It("Get* Funcs should work", func() {
			Expect(builder.GetClient()).To(Equal(k8sClient))
			Expect(builder.GetCtx()).To(Equal(ctx))
			Expect(builder.GetScheme()).To(Equal(k8sClient.Scheme()))
			Expect(builder.GetTemplate()).To(Equal(template))
			Expect(builder.GetRequest()).To(Equal(request))
		})

		It("getShortUID should work", func() {
			ret := getShortUID(request)
			Expect(len(ret)).To(Equal(8))
		})

		It("generateResourceName should work", func() {
			ret := generateResourceName(request)
			Expect(len(ret)).To(Equal(17))
		})

		It("GetTargetRefResource() should return a valid Client.Object", func() {
			ret, err := builder.GetTargetRefResource()
			Expect(err).To(Not(HaveOccurred()))
			Expect(ret.GetName()).To(Equal("test-dep"))
		})

		It("createPod() should work sanely", func() {
			// Get the PodTemplateSpec
			pts, err := builder.getPodTemplateFromController()
			Expect(err).To(Not(HaveOccurred()))

			// First, we should create the pod and return it.
			pod, err := builder.createPod(pts)
			Expect(err).To(Not(HaveOccurred()))

			// Store the original resourceVersino
			origResourceVersion := pod.ResourceVersion

			// Mutate the pod ourslves. This simulates a third party resource,
			// eg, "istio", mutating the pod.
			pod.ObjectMeta.SetAnnotations(map[string]string{
				"MyAnnotation":      "bar",
				"MyOtherAnnotation": "baz",
			})
			err = k8sClient.Update(ctx, pod)
			Expect(err).To(Not(HaveOccurred()))

			// VERIFY: The resourceVersion should have changed
			postAnnotationUpdateVersion := pod.ResourceVersion
			Expect(origResourceVersion).To(Not(Equal(postAnnotationUpdateVersion)))

			// Next, re-run the createPod function. We want this function to
			// never re-create the Pod object once it's been created, or update
			// it.
			_, err = builder.createPod(pts)
			Expect(err).To(Not(HaveOccurred()))

			// Re-get the pod from the API
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      pod.Name,
				Namespace: pod.Namespace,
			}, pod)
			Expect(err).To(Not(HaveOccurred()))

			// VERIFY: The Pod resourceVersion has not changed
			Expect(pod.ObjectMeta.ResourceVersion).To(Equal(postAnnotationUpdateVersion))
		})
	})
})
