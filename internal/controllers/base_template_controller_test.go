package controllers

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/diranged/oz/internal/api/v1alpha1"
	"github.com/diranged/oz/internal/legacybuilder"
)

var _ = Describe("BaseTemplateReconciler", Ordered, func() {
	Context("VerifyTargetRef", func() {
		var (
			deployment *appsv1.Deployment
			template   *v1alpha1.ExecAccessTemplate
			builder    legacybuilder.IBuilder
			r          *BaseTemplateReconciler
			fakeClient client.Client
			ctx        = context.Background()
			logger     = log.FromContext(ctx)
			err        error
		)

		BeforeEach(func() {
			logger.Info("BeforeEach...")

			// NOTE: Fake Client used here to make it easier to keep state separate between each It() test.
			fakeClient = fake.NewClientBuilder().WithRuntimeObjects().Build()

			deployment = &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "targetDeployment",
					Namespace: "fake",
				},
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"deployment": "true",
						},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"deployment": "true",
							},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "test",
									Image: "ubuntu:latest",
								},
							},
							ServiceAccountName: "default",
						},
					},
				},
			}
			err = fakeClient.Create(ctx, deployment)
			Expect(err).To(Not(HaveOccurred()))

			r = &BaseTemplateReconciler{
				BaseReconciler: BaseReconciler{
					Client:                  fakeClient,
					APIReader:               fakeClient,
					logger:                  logger,
					ReconcililationInterval: 0,
				},
			}
		})

		It("Should work if the deployment target is valid", func() {
			// Create the ExecAccessTemplate object that points to the valid Deployment
			template = &v1alpha1.ExecAccessTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testingTemplate",
					Namespace: "fake",
				},
				Spec: v1alpha1.ExecAccessTemplateSpec{
					AccessConfig: v1alpha1.AccessConfig{
						AllowedGroups:   []string{"foo", "bar"},
						DefaultDuration: "1h",
						MaxDuration:     "2h",
					},
					ControllerTargetRef: &v1alpha1.CrossVersionObjectReference{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
						Name:       "targetDeployment",
					},
				},
			}

			// Create the template resource for real in the fake kubernetes client
			err := fakeClient.Create(ctx, template)
			Expect(err).To(Not(HaveOccurred()))

			// Create the Builder that we'll be testing
			builder = &legacybuilder.BaseBuilder{
				Client:    fakeClient,
				Ctx:       ctx,
				APIReader: fakeClient,
				Template:  template,
			}

			// Call the method.. it should succeed.
			err = r.VerifyTargetRef(builder)
			Expect(err).To(Not(HaveOccurred()))

			// Now check that the condition was set as True
			Expect(meta.IsStatusConditionPresentAndEqual(
				template.Status.Conditions,
				string(v1alpha1.ConditionTargetRefExists),
				metav1.ConditionTrue)).To(BeTrue())
		})

		It("Should set condition if the target is invalid", func() {
			// Create the ExecAccessTemplate object that points to the valid Deployment
			template = &v1alpha1.ExecAccessTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testingTemplate",
					Namespace: "fake",
				},
				Spec: v1alpha1.ExecAccessTemplateSpec{
					AccessConfig: v1alpha1.AccessConfig{
						AllowedGroups:   []string{"foo", "bar"},
						DefaultDuration: "1h",
						MaxDuration:     "2h",
					},
					ControllerTargetRef: &v1alpha1.CrossVersionObjectReference{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
						Name:       "invalidDeploymentName",
					},
				},
			}

			// Create the template resource for real in the fake kubernetes client
			err := fakeClient.Create(ctx, template)
			Expect(err).To(Not(HaveOccurred()))

			r = &BaseTemplateReconciler{
				BaseReconciler: BaseReconciler{
					Client:                  fakeClient,
					APIReader:               fakeClient,
					logger:                  logger,
					ReconcililationInterval: 0,
				},
			}

			// Create the Builder that we'll be testing
			builder = &legacybuilder.BaseBuilder{
				Client:    fakeClient,
				Ctx:       ctx,
				APIReader: fakeClient,
				Template:  template,
			}

			// Call the method.. it should still succeed
			err = r.VerifyTargetRef(builder)
			Expect(err).To(Not(HaveOccurred()))

			// Now check that the condition was set though
			Expect(meta.IsStatusConditionPresentAndEqual(
				template.Status.Conditions,
				string(v1alpha1.ConditionTargetRefExists),
				metav1.ConditionFalse)).To(BeTrue())
		})
	})

	Context("VerifyMiscSettings", func() {
		var (
			template   *v1alpha1.ExecAccessTemplate
			builder    *legacybuilder.BaseBuilder
			r          *BaseTemplateReconciler
			fakeClient client.Client
			ctx        = context.Background()
			logger     = log.FromContext(ctx)
			err        error
		)

		BeforeEach(func() {
			logger.Info("BeforeEach...")

			// NOTE: Fake Client used here to make it easier to keep state separate between each It() test.
			fakeClient = fake.NewClientBuilder().WithRuntimeObjects().Build()

			r = &BaseTemplateReconciler{
				BaseReconciler: BaseReconciler{
					Client:                  fakeClient,
					Scheme:                  fakeClient.Scheme(),
					APIReader:               fakeClient,
					logger:                  logger,
					ReconcililationInterval: 0,
				},
			}
		})
		It("Should Update Conditions to true if settings are valid", func() {
			// Create the ExecAccessTemplate object that points to the valid Deployment
			template = &v1alpha1.ExecAccessTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testingTemplate",
					Namespace: "fake",
				},
				Spec: v1alpha1.ExecAccessTemplateSpec{
					AccessConfig: v1alpha1.AccessConfig{
						AllowedGroups:   []string{"foo", "bar"},
						DefaultDuration: "1h",
						MaxDuration:     "2h",
					},
					ControllerTargetRef: &v1alpha1.CrossVersionObjectReference{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
						Name:       "invalidDeploymentName",
					},
				},
			}

			// Create the template resource for real in the fake kubernetes client
			err = fakeClient.Create(ctx, template)
			Expect(err).To(Not(HaveOccurred()))

			// Create the Builder that we'll be testing
			builder = &legacybuilder.BaseBuilder{
				Client:    fakeClient,
				Ctx:       ctx,
				APIReader: fakeClient,
				Template:  template,
			}

			// Run the verification
			err = r.VerifyMiscSettings(builder)
			Expect(err).To(Not(HaveOccurred()))

			// Now check that the condition was set though
			Expect(meta.IsStatusConditionPresentAndEqual(
				template.Status.Conditions,
				string(v1alpha1.ConditionTemplateDurationsValid),
				metav1.ConditionTrue)).To(BeTrue())
			cond := meta.FindStatusCondition(
				template.Status.Conditions,
				string(v1alpha1.ConditionTemplateDurationsValid),
			)
			Expect(cond.Message).To(Equal("spec.defaultDuration and spec.maxDuration valid"))
		})

		It("Should Update Condition to False if DefaultDuration is invalid", func() {
			// Create the ExecAccessTemplate object that points to the valid Deployment
			template = &v1alpha1.ExecAccessTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testingTemplate",
					Namespace: "fake",
				},
				Spec: v1alpha1.ExecAccessTemplateSpec{
					AccessConfig: v1alpha1.AccessConfig{
						AllowedGroups:   []string{"foo", "bar"},
						DefaultDuration: "1invalidtimeframe",
						MaxDuration:     "2h",
					},
					ControllerTargetRef: &v1alpha1.CrossVersionObjectReference{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
						Name:       "invalidDeploymentName",
					},
				},
			}

			// Create the template resource for real in the fake kubernetes client
			err = fakeClient.Create(ctx, template)
			Expect(err).To(Not(HaveOccurred()))

			// Create the Builder that we'll be testing
			builder = &legacybuilder.BaseBuilder{
				Client:    fakeClient,
				Ctx:       ctx,
				APIReader: fakeClient,
				Template:  template,
			}

			// Run the verification
			err = r.VerifyMiscSettings(builder)
			Expect(err).To(Not(HaveOccurred()))

			// Now check that the condition was set though
			Expect(meta.IsStatusConditionPresentAndEqual(
				template.Status.Conditions,
				string(v1alpha1.ConditionTemplateDurationsValid),
				metav1.ConditionFalse)).To(BeTrue())
			cond := meta.FindStatusCondition(
				template.Status.Conditions,
				string(v1alpha1.ConditionTemplateDurationsValid),
			)
			Expect(
				cond.Message,
			).To(Equal("Error on spec.defaultDuration: time: unknown unit \"invalidtimeframe\" in duration \"1invalidtimeframe\""))
		})

		It("Should Update Condition to False if MaxDuration is invalid", func() {
			// Create the ExecAccessTemplate object that points to the valid Deployment
			template = &v1alpha1.ExecAccessTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testingTemplate",
					Namespace: "fake",
				},
				Spec: v1alpha1.ExecAccessTemplateSpec{
					AccessConfig: v1alpha1.AccessConfig{
						AllowedGroups:   []string{"foo", "bar"},
						DefaultDuration: "1h",
						MaxDuration:     "1invalidtimeframe",
					},
					ControllerTargetRef: &v1alpha1.CrossVersionObjectReference{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
						Name:       "invalidDeploymentName",
					},
				},
			}

			// Create the template resource for real in the fake kubernetes client
			err = fakeClient.Create(ctx, template)
			Expect(err).To(Not(HaveOccurred()))

			// Create the Builder that we'll be testing
			builder = &legacybuilder.BaseBuilder{
				Client:    fakeClient,
				Ctx:       ctx,
				APIReader: fakeClient,
				Template:  template,
			}

			// Run the verification
			err = r.VerifyMiscSettings(builder)
			Expect(err).To(Not(HaveOccurred()))

			// Now check that the condition was set though
			Expect(meta.IsStatusConditionPresentAndEqual(
				template.Status.Conditions,
				string(v1alpha1.ConditionTemplateDurationsValid),
				metav1.ConditionFalse)).To(BeTrue())
			cond := meta.FindStatusCondition(
				template.Status.Conditions,
				string(v1alpha1.ConditionTemplateDurationsValid),
			)
			Expect(
				cond.Message,
			).To(Equal("Error on spec.maxDuration: time: unknown unit \"invalidtimeframe\" in duration \"1invalidtimeframe\""))
		})

		It("Should Update Condition to False if DefaultDuration > MaxDuration", func() {
			// Create the ExecAccessTemplate object that points to the valid Deployment
			template = &v1alpha1.ExecAccessTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testingTemplate",
					Namespace: "fake",
				},
				Spec: v1alpha1.ExecAccessTemplateSpec{
					AccessConfig: v1alpha1.AccessConfig{
						AllowedGroups:   []string{"foo", "bar"},
						DefaultDuration: "1h",
						MaxDuration:     "1m",
					},
					ControllerTargetRef: &v1alpha1.CrossVersionObjectReference{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
						Name:       "invalidDeploymentName",
					},
				},
			}

			// Create the template resource for real in the fake kubernetes client
			err = fakeClient.Create(ctx, template)
			Expect(err).To(Not(HaveOccurred()))

			// Create the Builder that we'll be testing
			builder = &legacybuilder.BaseBuilder{
				Client:    fakeClient,
				Ctx:       ctx,
				APIReader: fakeClient,
				Template:  template,
			}

			// Run the verification
			err = r.VerifyMiscSettings(builder)
			Expect(err).To(Not(HaveOccurred()))

			// Now check that the condition was set though
			Expect(meta.IsStatusConditionPresentAndEqual(
				template.Status.Conditions,
				string(v1alpha1.ConditionTemplateDurationsValid),
				metav1.ConditionFalse)).To(BeTrue())
			cond := meta.FindStatusCondition(
				template.Status.Conditions,
				string(v1alpha1.ConditionTemplateDurationsValid),
			)
			Expect(
				cond.Message,
			).To(Equal("Error: spec.defaultDuration can not be greater than spec.maxDuration"))
		})
	})
})
