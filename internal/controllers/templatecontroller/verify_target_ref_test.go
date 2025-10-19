package templatecontroller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/diranged/oz/internal/api/v1alpha1"
	"github.com/diranged/oz/internal/testing/utils"
)

var _ = Describe("TemplateReconciler", Ordered, func() {
	Context("verifyTargetRef()", func() {
		var (
			ctx        = context.Background()
			ns         *corev1.Namespace
			reconciler *TemplateReconciler
			deployment *appsv1.Deployment
		)

		BeforeAll(func() {
			By("Should have a namespace to execute tests in")
			ns = &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: testutil.RandomString(8),
				},
			}
			err := k8sClient.Create(ctx, ns)
			Expect(err).ToNot(HaveOccurred())

			By("Creating the RequestReconciler")
			reconciler = &TemplateReconciler{
				Client:                 k8sClient,
				APIReader:              k8sClient,
				Scheme:                 k8sClient.Scheme(),
				TemplateType:           &v1alpha1.ExecAccessTemplate{},
				recorder:               record.NewFakeRecorder(50),
				ReconciliationInterval: 0,
			}

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
		})

		AfterAll(func() {
			By("Should delete the namespace")
			err := k8sClient.Delete(ctx, ns)
			Expect(err).ToNot(HaveOccurred())
		})

		It("verifgyTargetRef() should work", func() {
			By("Should have an ExecAccessTemplate built to test against")
			template := &v1alpha1.ExecAccessTemplate{
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
			err := k8sClient.Create(ctx, template)
			Expect(err).ToNot(HaveOccurred())

			By("Populating the RequestContext")
			rctx := newRequestContext(
				ctx,
				reconciler.TemplateType,
				reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      template.GetName(),
						Namespace: template.GetNamespace(),
					},
				},
			)
			err = reconciler.fetchRequestObject(rctx)
			Expect(err).ToNot(HaveOccurred())

			By("Executing the test")
			err = reconciler.verifyTargetRef(rctx)

			// VERIFY: No error returned
			Expect(err).ToNot(HaveOccurred())

			// VERIFY: ConditionTemplateTargetRefExists = False
			cond := meta.FindStatusCondition(
				*rctx.obj.GetStatus().GetConditions(),
				string(v1alpha1.ConditionTargetRefExists.String()),
			)
			Expect(cond).ToNot(BeNil())
			Expect(cond.Status).To(Equal(metav1.ConditionTrue))
			Expect(cond.Reason).To(Equal(string(metav1.StatusSuccess)))
		})

		It("verifgyTargetRef() should fail with missing deployment", func() {
			By("Should have an ExecAccessTemplate built to test against")
			template := &v1alpha1.ExecAccessTemplate{
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
						Name:       "invalid",
					},
				},
			}
			err := k8sClient.Create(ctx, template)
			Expect(err).ToNot(HaveOccurred())

			By("Populating the RequestContext")
			rctx := newRequestContext(
				ctx,
				reconciler.TemplateType,
				reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      template.GetName(),
						Namespace: template.GetNamespace(),
					},
				},
			)
			err = reconciler.fetchRequestObject(rctx)
			Expect(err).ToNot(HaveOccurred())

			By("Executing the test")
			err = reconciler.verifyTargetRef(rctx)

			// VERIFY: No error returned
			Expect(err).ToNot(HaveOccurred())

			// VERIFY: ConditionTemplateTargetRefExists = False
			cond := meta.FindStatusCondition(
				*rctx.obj.GetStatus().GetConditions(),
				string(v1alpha1.ConditionTargetRefExists.String()),
			)
			Expect(cond).ToNot(BeNil())
			Expect(cond.Status).To(Equal(metav1.ConditionFalse))
			Expect(cond.Reason).To(Equal(string(metav1.StatusReasonNotFound)))
		})
	})
})
