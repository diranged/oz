package controllers

import (
	"context"
	"errors"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log"

	api "github.com/diranged/oz/internal/api/v1alpha1"
	"github.com/diranged/oz/internal/builders"
	"github.com/diranged/oz/internal/controllers/internal/conditions"
)

var _ = Describe("BaseRequestReconciler", Ordered, func() {
	Context("verifyAccessResources", func() {
		var (
			request    *api.ExecAccessRequest
			builder    *FakeBuilder
			r          *BaseRequestReconciler
			fakeClient client.Client
			ctx        = context.Background()
		)

		BeforeEach(func() {
			// NOTE: Fake Client used here to make it easier to keep state separate between each It() test.
			fakeClient = fake.NewClientBuilder().WithRuntimeObjects().Build()

			// Create the template that can be referenced by the request
			r = &BaseRequestReconciler{
				BaseReconciler: BaseReconciler{
					Client:                  fakeClient,
					Scheme:                  fakeClient.Scheme(),
					APIReader:               fakeClient,
					ReconcililationInterval: 0,
				},
			}

			// Create an empty request to test against
			request = &api.ExecAccessRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "expiredRequest",
					Namespace: "namespace",
				},
				Spec: api.ExecAccessRequestSpec{
					TemplateName: "bogus",
				},
				Status: api.ExecAccessRequestStatus{
					CoreStatus: api.CoreStatus{
						Conditions:    []metav1.Condition{},
						Ready:         false,
						AccessMessage: "",
					},
				},
			}

			// Create the Builder that we'll be testing
			builder = &FakeBuilder{
				BaseBuilder: builders.BaseBuilder{
					Client:    fakeClient,
					Ctx:       ctx,
					APIReader: fakeClient,
					Request:   request,
				},
			}

			// Create the template resource for real in the fake kubernetes client
			err := fakeClient.Create(ctx, request)
			Expect(err).To(Not(HaveOccurred()))
		})

		It("Should return clean access message set condition to true", func() {
			// Configure FakeBuilder to return success
			builder.retAccessString = "here you go"
			builder.retErr = nil
			builder.retStatusString = "success"

			// Build the resources
			err := r.verifyAccessResourcesBuilt(builder)

			// VERIFY: no error
			Expect(err).To(Not(HaveOccurred()))

			// VERIFY: the conditions in the request object were updated
			Expect(meta.IsStatusConditionPresentAndEqual(
				request.Status.Conditions,
				string(conditions.ConditionAccessResourcesCreated),
				metav1.ConditionTrue)).To(BeTrue())
			cond := meta.FindStatusCondition(
				request.Status.Conditions,
				string(conditions.ConditionAccessResourcesCreated),
			)
			Expect(cond.Message).To(Equal("success"))
		})

		It(
			"Should return access message set condition to false if error creating resources",
			func() {
				// Configure FakeBuilder to return success
				builder.retAccessString = "an error happened"
				builder.retErr = errors.New("i failed")
				builder.retStatusString = "failure"

				// Build the resources
				err := r.verifyAccessResourcesBuilt(builder)

				// VERIFY: error occurred
				Expect(err).To(HaveOccurred())

				// VERIFY: the conditions in the request object were updated
				Expect(meta.IsStatusConditionPresentAndEqual(
					request.Status.Conditions,
					string(conditions.ConditionAccessResourcesCreated),
					metav1.ConditionFalse)).To(BeTrue())
				cond := meta.FindStatusCondition(
					request.Status.Conditions,
					string(conditions.ConditionAccessResourcesCreated),
				)
				Expect(cond.Message).To(Equal("ERROR: i failed"))
			},
		)

		It("Should return an error if the UpdateCondition fails on success", func() {
			// Configure FakeBuilder to return success
			builder.retAccessString = "here you go"
			builder.retErr = nil
			builder.retStatusString = "success"

			// Break the "request" object by changing its name to something that doesn't exist,
			// to cause the UpdateCondition() to fail.
			request.Name = "bogus"

			// Build the resources
			err := r.verifyAccessResourcesBuilt(builder)

			// VERIFY: error occurred
			Expect(err).To(HaveOccurred())
			Expect(
				err.Error(),
			).To(Equal("execaccessrequests.crds.wizardofoz.co \"bogus\" not found"))
		})
	})

	Context("isAccessExpired", func() {
		var (
			request    *api.ExecAccessRequest
			builder    *builders.BaseBuilder
			r          *BaseRequestReconciler
			fakeClient client.Client
			ctx        = context.Background()
		)

		BeforeEach(func() {
			// NOTE: Fake Client used here to make it easier to keep state separate between each It() test.
			fakeClient = fake.NewClientBuilder().WithRuntimeObjects().Build()

			// Create the template that can be referenced by the request
			r = &BaseRequestReconciler{
				BaseReconciler: BaseReconciler{
					Client:                  fakeClient,
					Scheme:                  fakeClient.Scheme(),
					APIReader:               fakeClient,
					ReconcililationInterval: 0,
				},
			}
		})

		It("Should return False if the access valid condition is True", func() {
			// Create a fake request with a condition already populated that indicates we've been expired
			request = &api.ExecAccessRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "expiredRequest",
					Namespace: "namespace",
				},
				Spec: api.ExecAccessRequestSpec{
					TemplateName: "bogus",
				},
				Status: api.ExecAccessRequestStatus{
					CoreStatus: api.CoreStatus{
						Conditions: []metav1.Condition{
							{
								Type:    string(conditions.ConditionAccessStillValid),
								Status:  metav1.ConditionTrue,
								Reason:  "Valid",
								Message: "Valid",
							},
						},
						Ready:         false,
						AccessMessage: "",
					},
				},
			}

			// Create the template resource for real in the fake kubernetes client
			err := fakeClient.Create(ctx, request)
			Expect(err).To(Not(HaveOccurred()))

			// Create the Builder that we'll be testing
			builder = &builders.BaseBuilder{
				Client:    fakeClient,
				Ctx:       ctx,
				APIReader: fakeClient,
				Request:   request,
			}

			// VERIFY: isExpired returned False
			isExpired, err := r.isAccessExpired(builder)
			Expect(err).To(Not(HaveOccurred()))
			Expect(isExpired).To(BeFalse())
		})

		It("Should return True if the access valid condition is false", func() {
			// Create a fake request with a condition already populated that indicates we've been expired
			request = &api.ExecAccessRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "expiredRequest",
					Namespace: "namespace",
				},
				Spec: api.ExecAccessRequestSpec{
					TemplateName: "bogus",
				},
				Status: api.ExecAccessRequestStatus{
					CoreStatus: api.CoreStatus{
						Conditions: []metav1.Condition{
							{
								Type:    string(conditions.ConditionAccessStillValid),
								Status:  metav1.ConditionFalse,
								Reason:  "Expired",
								Message: "Expired",
							},
						},
						Ready:         false,
						AccessMessage: "",
					},
				},
			}

			// Create the template resource for real in the fake kubernetes client
			err := fakeClient.Create(ctx, request)
			Expect(err).To(Not(HaveOccurred()))

			// Create the Builder that we'll be testing
			builder = &builders.BaseBuilder{
				Client:    fakeClient,
				Ctx:       ctx,
				APIReader: fakeClient,
				Request:   request,
			}

			// VERIFY: isExpired returned True
			isExpired, err := r.isAccessExpired(builder)
			Expect(err).To(Not(HaveOccurred()))
			Expect(isExpired).To(BeTrue())

			// VERIFY: the AccessRequest is deleted
			found := &api.ExecAccessRequest{}
			err = fakeClient.Get(ctx, types.NamespacedName{
				Name:      request.Name,
				Namespace: request.Namespace,
			}, found)
			Expect(
				err.Error(),
			).To(Equal("execaccessrequests.crds.wizardofoz.co \"expiredRequest\" not found"))
		})

		It("Should return False if the AccessValid condition is missing", func() {
			// Create a fake request with a condition already populated that indicates we've been expired
			request = &api.ExecAccessRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "expiredRequest",
					Namespace: "namespace",
				},
				Spec: api.ExecAccessRequestSpec{
					TemplateName: "bogus",
				},
				Status: api.ExecAccessRequestStatus{
					CoreStatus: api.CoreStatus{
						Conditions:    []metav1.Condition{},
						Ready:         false,
						AccessMessage: "",
					},
				},
			}

			// Create the template resource for real in the fake kubernetes client
			err := fakeClient.Create(ctx, request)
			Expect(err).To(Not(HaveOccurred()))

			// Create the Builder that we'll be testing
			builder = &builders.BaseBuilder{
				Client:    fakeClient,
				Ctx:       ctx,
				APIReader: fakeClient,
				Request:   request,
			}

			// VERIFY: isExpired returns false
			isExpired, err := r.isAccessExpired(builder)
			Expect(err).To(Not(HaveOccurred()))
			Expect(isExpired).To(BeFalse())
		})
	})

	Context("verifyDuration", func() {
		var (
			template   *api.ExecAccessTemplate
			request    *api.ExecAccessRequest
			builder    *builders.BaseBuilder
			r          *BaseRequestReconciler
			fakeClient client.Client
			ctx        = context.Background()
			logger     = log.FromContext(ctx)
		)

		BeforeEach(func() {
			logger.Info("BeforeEach...")

			// NOTE: Fake Client used here to make it easier to keep state separate between each It() test.
			fakeClient = fake.NewClientBuilder().WithRuntimeObjects().Build()

			// Create a common ExecAccessTemplate used to test the request against
			template = &api.ExecAccessTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testingTemplate",
					Namespace: "fake",
				},
				Spec: api.ExecAccessTemplateSpec{
					AccessConfig: api.AccessConfig{
						AllowedGroups:   []string{"foo", "bar"},
						DefaultDuration: "1h",
						MaxDuration:     "2h",
					},
					ControllerTargetRef: &api.CrossVersionObjectReference{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
						Name:       "targetDeployment",
					},
				},
			}

			// Create the template that can be referenced by the request
			err := fakeClient.Create(ctx, template)
			Expect(err).To(Not(HaveOccurred()))

			r = &BaseRequestReconciler{
				BaseReconciler: BaseReconciler{
					Client:                  fakeClient,
					Scheme:                  fakeClient.Scheme(),
					APIReader:               fakeClient,
					logger:                  logger,
					ReconcililationInterval: 0,
				},
			}
		})

		It("Should update conditions to True in success", func() {
			// Create the ExecAccessTemplate object that points to the valid Deployment
			request = &api.ExecAccessRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testingRequest",
					Namespace: "fake",
					// Set the creation timestamp so that the verification works, the fake kubeclient doesn't do that.
					CreationTimestamp: metav1.NewTime(time.Now()),
				},
				Spec: api.ExecAccessRequestSpec{
					TemplateName: "testingTemplate",
					Duration:     "30m",
				},
			}

			// Create the template resource for real in the fake kubernetes client
			err := fakeClient.Create(ctx, request)
			Expect(err).To(Not(HaveOccurred()))

			// Create the Builder that we'll be testing
			builder = &builders.BaseBuilder{
				Client:    fakeClient,
				Ctx:       ctx,
				APIReader: fakeClient,
				Template:  template,
				Request:   request,
			}

			// Call the method.. it should succeed.
			err = r.verifyDuration(builder)
			Expect(err).To(Not(HaveOccurred()))

			// VERIFY: The ConditionDurationsValid is True
			Expect(meta.IsStatusConditionPresentAndEqual(
				request.Status.Conditions,
				string(conditions.ConditionDurationsValid),
				metav1.ConditionTrue)).To(BeTrue())
			cond := meta.FindStatusCondition(
				request.Status.Conditions,
				string(conditions.ConditionDurationsValid),
			)
			Expect(cond.Message).To(Equal("Access requested custom duration (30m0s)"))

			// VERIFY: The conditionAccessStillValid is True
			cond = meta.FindStatusCondition(
				request.Status.Conditions,
				string(conditions.ConditionAccessStillValid),
			)
			Expect(cond.Message).To(Equal("Access still valid"))
		})

		It("Should use template default duration if none is supplied", func() {
			// Create the ExecAccessTemplate object that points to the valid Deployment
			request = &api.ExecAccessRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testingRequest",
					Namespace: "fake",
					// Set the creation timestamp so that the verification works, the fake kubeclient doesn't do that.
					CreationTimestamp: metav1.NewTime(time.Now()),
				},
				Spec: api.ExecAccessRequestSpec{
					TemplateName: "testingTemplate",
					// Duration:     "30m",
				},
			}

			// Create the template resource for real in the fake kubernetes client
			err := fakeClient.Create(ctx, request)
			Expect(err).To(Not(HaveOccurred()))

			// Create the Builder that we'll be testing
			builder = &builders.BaseBuilder{
				Client:    fakeClient,
				Ctx:       ctx,
				APIReader: fakeClient,
				Template:  template,
				Request:   request,
			}

			// Call the method.. it should succeed.
			err = r.verifyDuration(builder)
			Expect(err).To(Not(HaveOccurred()))

			// VERIFY: The ConditionDurationsValid is True
			Expect(meta.IsStatusConditionPresentAndEqual(
				request.Status.Conditions,
				string(conditions.ConditionDurationsValid),
				metav1.ConditionTrue)).To(BeTrue())
			cond := meta.FindStatusCondition(
				request.Status.Conditions,
				string(conditions.ConditionDurationsValid),
			)
			Expect(
				cond.Message,
			).To(Equal("Access request duration defaulting to template duration time (1h0m0s)"))

			// VERIFY: The conditionAccessStillValid is True
			cond = meta.FindStatusCondition(
				request.Status.Conditions,
				string(conditions.ConditionAccessStillValid),
			)
			Expect(cond.Message).To(Equal("Access still valid"))
		})

		It("Should use template max duration if requested duration is too long", func() {
			// Create the ExecAccessTemplate object that points to the valid Deployment
			request = &api.ExecAccessRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testingRequest",
					Namespace: "fake",
					// Set the creation timestamp so that the verification works, the fake kubeclient doesn't do that.
					CreationTimestamp: metav1.NewTime(time.Now()),
				},
				Spec: api.ExecAccessRequestSpec{
					TemplateName: "testingTemplate",
					Duration:     "24h",
				},
			}

			// Create the template resource for real in the fake kubernetes client
			err := fakeClient.Create(ctx, request)
			Expect(err).To(Not(HaveOccurred()))

			// Create the Builder that we'll be testing
			builder = &builders.BaseBuilder{
				Client:    fakeClient,
				Ctx:       ctx,
				APIReader: fakeClient,
				Template:  template,
				Request:   request,
			}

			// Call the method.. it should succeed.
			err = r.verifyDuration(builder)
			Expect(err).To(Not(HaveOccurred()))

			// VERIFY: The ConditionDurationsValid is True
			Expect(meta.IsStatusConditionPresentAndEqual(
				request.Status.Conditions,
				string(conditions.ConditionDurationsValid),
				metav1.ConditionTrue)).To(BeTrue())
			cond := meta.FindStatusCondition(
				request.Status.Conditions,
				string(conditions.ConditionDurationsValid),
			)
			Expect(
				cond.Message,
			).To(Equal("Access requested duration (24h0m0s) larger than template maximum duration (2h0m0s)"))

			// VERIFY: The conditionAccessStillValid is True
			cond = meta.FindStatusCondition(
				request.Status.Conditions,
				string(conditions.ConditionAccessStillValid),
			)
			Expect(cond.Message).To(Equal("Access still valid"))
		})

		It("Should set condition if the spec.Duration is invalid", func() {
			// Create the ExecAccessTemplate object that points to the valid Deployment
			request = &api.ExecAccessRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testingRequest",
					Namespace: "fake",
					// Set the creation timestamp so that the verification works, the fake kubeclient doesn't do that.
					CreationTimestamp: metav1.NewTime(time.Now()),
				},
				Spec: api.ExecAccessRequestSpec{
					TemplateName: "testingTemplate",
					Duration:     "30minutes",
				},
			}

			// Create the template resource for real in the fake kubernetes client
			err := fakeClient.Create(ctx, request)
			Expect(err).To(Not(HaveOccurred()))

			// Create the Builder that we'll be testing
			builder = &builders.BaseBuilder{
				Client:    fakeClient,
				Ctx:       ctx,
				APIReader: fakeClient,
				Template:  template,
				Request:   request,
			}

			// Call the method.. it should succeed.
			err = r.verifyDuration(builder)

			// VERIFY: The proper Error was returned
			Expect(err).To(HaveOccurred())
			Expect(
				err.Error(),
			).To(Equal("time: unknown unit \"minutes\" in duration \"30minutes\""))

			// VERIFY: The ConditionDurationsValid is False
			Expect(meta.IsStatusConditionPresentAndEqual(
				request.Status.Conditions,
				string(conditions.ConditionDurationsValid),
				metav1.ConditionFalse)).To(BeTrue())

			// VERIFY: The Condition was updated properly in the object even though an error was returned
			cond := meta.FindStatusCondition(
				request.Status.Conditions,
				string(conditions.ConditionDurationsValid),
			)
			Expect(
				cond.Message,
			).To(Equal("spec.duration error: time: unknown unit \"minutes\" in duration \"30minutes\""))
		})

		It(
			"Should set condition if the referenced template spec.accessConfig.defaultDuration is invalid",
			func() {
				// Get the template, and update its defaultDuration to something invalid
				err := fakeClient.Get(ctx, types.NamespacedName{
					Name:      template.Name,
					Namespace: template.Namespace,
				}, template)
				Expect(err).To(Not(HaveOccurred()))
				template.Spec.AccessConfig.DefaultDuration = "1hour"
				err = fakeClient.Update(ctx, template)
				Expect(err).To(Not(HaveOccurred()))

				// Create the ExecAccessTemplate object that points to the valid Deployment
				request = &api.ExecAccessRequest{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "testingRequest",
						Namespace: "fake",
						// Set the creation timestamp so that the verification works, the fake kubeclient doesn't do that.
						CreationTimestamp: metav1.NewTime(time.Now()),
					},
					Spec: api.ExecAccessRequestSpec{
						TemplateName: "testingTemplate",
						Duration:     "30m",
					},
				}

				// Create the template resource for real in the fake kubernetes client
				err = fakeClient.Create(ctx, request)
				Expect(err).To(Not(HaveOccurred()))

				// Create the Builder that we'll be testing
				builder = &builders.BaseBuilder{
					Client:    fakeClient,
					Ctx:       ctx,
					APIReader: fakeClient,
					Template:  template,
					Request:   request,
				}

				// Call the method.. it should succeed.
				err = r.verifyDuration(builder)

				// VERIFY: The proper Error was returned
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("time: unknown unit \"hour\" in duration \"1hour\""))

				// VERIFY: The ConditionDurationsValid is False
				Expect(meta.IsStatusConditionPresentAndEqual(
					request.Status.Conditions,
					string(conditions.ConditionDurationsValid),
					metav1.ConditionFalse)).To(BeTrue())

				// VERIFY: The Condition was updated properly in the object even though an error was returned
				cond := meta.FindStatusCondition(
					request.Status.Conditions,
					string(conditions.ConditionDurationsValid),
				)
				Expect(
					cond.Message,
				).To(Equal("Template Error, spec.defaultDuration error: time: unknown unit \"hour\" in duration \"1hour\""))
			},
		)

		It(
			"Should set condition if the referenced template spec.accessConfig.maxDuration is invalid",
			func() {
				// Get the template, and update its defaultDuration to something invalid
				err := fakeClient.Get(ctx, types.NamespacedName{
					Name:      template.Name,
					Namespace: template.Namespace,
				}, template)
				Expect(err).To(Not(HaveOccurred()))
				template.Spec.AccessConfig.MaxDuration = "1hour"
				err = fakeClient.Update(ctx, template)
				Expect(err).To(Not(HaveOccurred()))

				// Create the ExecAccessTemplate object that points to the valid Deployment
				request = &api.ExecAccessRequest{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "testingRequest",
						Namespace: "fake",
						// Set the creation timestamp so that the verification works, the fake kubeclient doesn't do that.
						CreationTimestamp: metav1.NewTime(time.Now()),
					},
					Spec: api.ExecAccessRequestSpec{
						TemplateName: "testingTemplate",
						Duration:     "30m",
					},
				}

				// Create the template resource for real in the fake kubernetes client
				err = fakeClient.Create(ctx, request)
				Expect(err).To(Not(HaveOccurred()))

				// Create the Builder that we'll be testing
				builder = &builders.BaseBuilder{
					Client:    fakeClient,
					Ctx:       ctx,
					APIReader: fakeClient,
					Template:  template,
					Request:   request,
				}

				// Call the method.. it should succeed.
				err = r.verifyDuration(builder)

				// VERIFY: The proper Error was returned
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("time: unknown unit \"hour\" in duration \"1hour\""))

				// VERIFY: The ConditionDurationsValid is False
				Expect(meta.IsStatusConditionPresentAndEqual(
					request.Status.Conditions,
					string(conditions.ConditionDurationsValid),
					metav1.ConditionFalse)).To(BeTrue())

				// VERIFY: The Condition was updated properly in the object even though an error was returned
				cond := meta.FindStatusCondition(
					request.Status.Conditions,
					string(conditions.ConditionDurationsValid),
				)
				Expect(
					cond.Message,
				).To(Equal("Template Error, spec.maxDuration error: time: unknown unit \"hour\" in duration \"1hour\""))
			},
		)

		It("Should set access expired if uptime > duration", func() {
			// Create the ExecAccessTemplate object that points to the valid Deployment
			request = &api.ExecAccessRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testingRequest",
					Namespace: "fake",
					// Set the creation timestamp so that the verification works, the fake kubeclient doesn't do that.
					CreationTimestamp: metav1.NewTime(time.Now().Add(time.Minute * -5)),
				},
				Spec: api.ExecAccessRequestSpec{
					TemplateName: "testingTemplate",
					Duration:     "1m",
				},
			}

			// Create the template resource for real in the fake kubernetes client
			err := fakeClient.Create(ctx, request)
			Expect(err).To(Not(HaveOccurred()))

			// Create the Builder that we'll be testing
			builder = &builders.BaseBuilder{
				Client:    fakeClient,
				Ctx:       ctx,
				APIReader: fakeClient,
				Template:  template,
				Request:   request,
			}

			// Call the method.. it should succeed.
			err = r.verifyDuration(builder)
			Expect(err).To(Not(HaveOccurred()))

			// VERIFY: The ConditionDurationsValid is True
			Expect(meta.IsStatusConditionPresentAndEqual(
				request.Status.Conditions,
				string(conditions.ConditionDurationsValid),
				metav1.ConditionTrue)).To(BeTrue())
			cond := meta.FindStatusCondition(
				request.Status.Conditions,
				string(conditions.ConditionDurationsValid),
			)
			Expect(cond.Message).To(Equal("Access requested custom duration (1m0s)"))

			// VERIFY: The conditionAccessStillValid is True
			cond = meta.FindStatusCondition(
				request.Status.Conditions,
				string(conditions.ConditionAccessStillValid),
			)
			Expect(cond.Message).To(Equal("Access expired"))
		})
	})
})
