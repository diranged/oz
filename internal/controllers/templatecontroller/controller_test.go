package templatecontroller

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/diranged/oz/internal/api/v1alpha1"
	"github.com/diranged/oz/internal/testing/utils"
)

var _ = Describe("TemplateReconciler", Ordered, func() {
	Context("TemplateReconciler()", func() {
		/*
			Reconcile() tests
		*/
		Context("Reconcile()", func() {
			var (
				ctx        = context.Background()
				ns         *corev1.Namespace
				deployment *appsv1.Deployment
				template   *v1alpha1.ExecAccessTemplate
				reconciler *TemplateReconciler
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

				By("Should have an ExecAccessTemplate built to test against")
				template = &v1alpha1.ExecAccessTemplate{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "reconcile-test",
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

				By("Creating the RequestReconciler")
				reconciler = &TemplateReconciler{
					Client:                 k8sClient,
					Scheme:                 k8sClient.Scheme(),
					APIReader:              k8sClient,
					TemplateType:           &v1alpha1.ExecAccessTemplate{},
					ReconciliationInterval: time.Minute,
				}
			})

			It("Reconcile() should return if the Request object is gone", func() {
				result, err := reconciler.Reconcile(
					ctx,
					reconcile.Request{
						NamespacedName: types.NamespacedName{
							Name:      "missing",
							Namespace: template.GetNamespace(),
						},
					},
				)
				// VERIFY: No error returned
				Expect(err).ToNot(HaveOccurred())

				// VERIFY: No Requeue
				Expect(result.Requeue).To(BeFalse())
			})

			It("Reconcile() should work", func() {
				By("Running reconcile()")
				result, err := reconciler.Reconcile(
					ctx,
					reconcile.Request{
						NamespacedName: types.NamespacedName{
							Name:      template.GetName(),
							Namespace: template.GetNamespace(),
						},
					},
				)
				// VERIFY: The result is that we WILL requeue in a few minutes
				Expect(result.RequeueAfter).To(Equal(reconciler.ReconciliationInterval))
				Expect(err).ToNot(HaveOccurred())

				// Refetch our Request object... reconiliation has mutated its
				// .Status fields.
				By("Refetching our Template...")
				err = k8sClient.Get(ctx, types.NamespacedName{
					Name:      template.Name,
					Namespace: template.Namespace,
				}, template)
				Expect(err).To(Not(HaveOccurred()))

				// VERIFY: The condition was updated appropriately
				By("Checking the resulting conditions")

				// ConditionTemplateDurationsValid = True
				cond := meta.FindStatusCondition(
					*template.GetStatus().GetConditions(),
					v1alpha1.ConditionTemplateDurationsValid.String(),
				)
				Expect(cond).ToNot(BeNil())
				Expect(cond.Status).To(Equal(metav1.ConditionTrue))
				Expect(cond.Reason).To(Equal(string(metav1.StatusSuccess)))

				// ConditionTargetRefExists = True
				cond = meta.FindStatusCondition(
					*template.GetStatus().GetConditions(),
					v1alpha1.ConditionTargetRefExists.String(),
				)
				Expect(cond).ToNot(BeNil())
				Expect(cond.Status).To(Equal(metav1.ConditionTrue))
				Expect(cond.Reason).To(Equal(string(metav1.StatusSuccess)))

				// Ready Status was set to true
				Expect(template.Status.IsReady()).To(BeTrue())
			})

			It(
				"Reconcile() should mark resource as not ready if conditions fail, and not requeue",
				func() {
					By("Pointing the Template to an invalid Deployment")
					template.Spec.ControllerTargetRef.Name = "invalid"
					err := k8sClient.Update(ctx, template)
					Expect(err).ToNot(HaveOccurred())

					By("Running reconcile()")
					result, err := reconciler.Reconcile(
						ctx,
						reconcile.Request{
							NamespacedName: types.NamespacedName{
								Name:      template.GetName(),
								Namespace: template.GetNamespace(),
							},
						},
					)
					// VERIFY: The result is that we will NOT requeue
					Expect(result.Requeue).To(BeFalse())
					Expect(err).ToNot(HaveOccurred())

					// Refetch our Request object... reconiliation has mutated its
					// .Status fields.
					By("Refetching our Request...")
					err = k8sClient.Get(ctx, types.NamespacedName{
						Name:      template.Name,
						Namespace: template.Namespace,
					}, template)
					Expect(err).To(Not(HaveOccurred()))

					// VERIFY: The ConditionRequestDurationsValid = False
					By("Checking the resulting conditions")
					cond := meta.FindStatusCondition(
						*template.GetStatus().GetConditions(),
						v1alpha1.ConditionTargetRefExists.String(),
					)
					Expect(cond).ToNot(BeNil())
					Expect(cond.Status).To(Equal(metav1.ConditionFalse))
					Expect(cond.Reason).To(Equal(string(metav1.StatusReasonNotFound)))

					// Ready Status was set to false
					Expect(template.Status.IsReady()).To(BeFalse())
				},
			)
		})
	})
})
