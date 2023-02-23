package templatecontroller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/diranged/oz/internal/api/v1alpha1"
	"github.com/diranged/oz/internal/testing/utils"
)

var _ = Describe("TemplateReconciler", Ordered, func() {
	Context("verifyDuration()", func() {
		var (
			ctx        = context.Background()
			ns         *v1.Namespace
			reconciler *TemplateReconciler
			recorder   = record.NewFakeRecorder(50)
		)

		BeforeAll(func() {
			By("Should have a namespace to execute tests in")
			ns = &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: utils.RandomString(8),
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
				recorder:               recorder,
				ReconciliationInterval: 0,
			}
		})

		AfterAll(func() {
			By("Should delete the namespace")
			err := k8sClient.Delete(ctx, ns)
			Expect(err).ToNot(HaveOccurred())
		})

		It("verifyDuration() should work", func() {
			By("Should have an ExecAccessTemplate built to test against")
			template := &v1alpha1.ExecAccessTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      utils.RandomString(8),
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
						Name:       "junk",
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
			err = reconciler.verifyDuration(rctx)
			Expect(err).ToNot(HaveOccurred())

			// VERIFY: ConditionTemplateDurationsValid = True
			cond := meta.FindStatusCondition(
				*rctx.obj.GetStatus().GetConditions(),
				string(v1alpha1.ConditionTemplateDurationsValid.String()),
			)
			Expect(cond).ToNot(BeNil())
			Expect(cond.Status).To(Equal(metav1.ConditionTrue))
			Expect(cond.Reason).To(Equal(string(metav1.StatusSuccess)))
		})

		It("verifyDuration() should return error if defaultDuration is invalid", func() {
			By("Should have an ExecAccessTemplate built to test against")
			template := &v1alpha1.ExecAccessTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      utils.RandomString(8),
					Namespace: ns.GetName(),
				},
				Spec: v1alpha1.ExecAccessTemplateSpec{
					AccessConfig: v1alpha1.AccessConfig{
						AllowedGroups:   []string{"foo"},
						DefaultDuration: "1hour",
						MaxDuration:     "2h",
					},
					ControllerTargetRef: &v1alpha1.CrossVersionObjectReference{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
						Name:       "junk",
					},
				},
			}
			err := k8sClient.Create(ctx, template)
			Expect(err).ToNot(HaveOccurred())
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      template.GetName(),
				Namespace: template.GetNamespace(),
			}, template)
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

			// Run the verifyDuration
			By("Executing the test")
			err = reconciler.verifyDuration(rctx)

			// VERIFY: No error returned, but the status of the template should be updated
			Expect(err).ToNot(HaveOccurred())

			// VERIFY: ConditionAccessResourcesCreated = False
			cond := meta.FindStatusCondition(
				*rctx.obj.GetStatus().GetConditions(),
				string(v1alpha1.ConditionTemplateDurationsValid.String()),
			)
			Expect(cond).ToNot(BeNil())
			Expect(cond.Status).To(Equal(metav1.ConditionFalse))
			Expect(cond.Reason).To(Equal(string(metav1.StatusReasonNotAcceptable)))
			Expect(cond.Message).To(MatchRegexp("unknown unit"))
		})

		It("verifyDuration() should return error if maxDuration is invalid", func() {
			By("Should have an ExecAccessTemplate built to test against")
			template := &v1alpha1.ExecAccessTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      utils.RandomString(8),
					Namespace: ns.GetName(),
				},
				Spec: v1alpha1.ExecAccessTemplateSpec{
					AccessConfig: v1alpha1.AccessConfig{
						AllowedGroups:   []string{"foo"},
						DefaultDuration: "1h",
						MaxDuration:     "2hour",
					},
					ControllerTargetRef: &v1alpha1.CrossVersionObjectReference{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
						Name:       "junk",
					},
				},
			}
			err := k8sClient.Create(ctx, template)
			Expect(err).ToNot(HaveOccurred())
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      template.GetName(),
				Namespace: template.GetNamespace(),
			}, template)
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

			// Run the verifyDuration
			By("Executing the test")
			err = reconciler.verifyDuration(rctx)

			// VERIFY: No error returned, but the status of the template should be updated
			Expect(err).ToNot(HaveOccurred())

			// VERIFY: ConditionAccessResourcesCreated = False
			cond := meta.FindStatusCondition(
				*rctx.obj.GetStatus().GetConditions(),
				string(v1alpha1.ConditionTemplateDurationsValid.String()),
			)
			Expect(cond).ToNot(BeNil())
			Expect(cond.Status).To(Equal(metav1.ConditionFalse))
			Expect(cond.Reason).To(Equal(string(metav1.StatusReasonNotAcceptable)))
			Expect(cond.Message).To(MatchRegexp("unknown unit"))
		})

		It("verifyDuration() should return error if defaultDuration > maxDuration", func() {
			By("Should have an ExecAccessTemplate built to test against")
			template := &v1alpha1.ExecAccessTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      utils.RandomString(8),
					Namespace: ns.GetName(),
				},
				Spec: v1alpha1.ExecAccessTemplateSpec{
					AccessConfig: v1alpha1.AccessConfig{
						AllowedGroups:   []string{"foo"},
						DefaultDuration: "2h",
						MaxDuration:     "1h",
					},
					ControllerTargetRef: &v1alpha1.CrossVersionObjectReference{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
						Name:       "junk",
					},
				},
			}
			err := k8sClient.Create(ctx, template)
			Expect(err).ToNot(HaveOccurred())
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      template.GetName(),
				Namespace: template.GetNamespace(),
			}, template)
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

			// Run the verifyDuration
			By("Executing the test")
			err = reconciler.verifyDuration(rctx)

			// VERIFY: No error returned, but the status of the template should be updated
			Expect(err).ToNot(HaveOccurred())

			// VERIFY: ConditionAccessResourcesCreated = False
			cond := meta.FindStatusCondition(
				*rctx.obj.GetStatus().GetConditions(),
				string(v1alpha1.ConditionTemplateDurationsValid.String()),
			)
			Expect(cond).ToNot(BeNil())
			Expect(cond.Status).To(Equal(metav1.ConditionFalse))
			Expect(cond.Reason).To(Equal(string(metav1.StatusReasonNotAcceptable)))
			Expect(cond.Message).To(MatchRegexp("can not be greater than"))
		})
	})
})
