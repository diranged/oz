package controllers

import (
	"context"
	"fmt"
	"strconv"
	"time"

	api "github.com/diranged/oz/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("ExecAccessTemplateController", Ordered, func() {
	Context("Controller Test", func() {
		const TestName = "execaccesstemplatecontroller"
		var namespace *corev1.Namespace

		// Logger for our tests - makes it easier for us to debug sometimes
		ctx := context.Background()
		logger := log.FromContext(ctx)

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

		// Keep track of resources created within the tests.
		//
		// https://github.com/kubernetes-sigs/controller-runtime/issues/880#issuecomment-634493086
		var createdResources []client.Object
		deleteResourceAfterTest := func(o client.Object) {
			logger.V(1).
				Info(fmt.Sprintf("Adding %s/%s to list of objects to clean up", o.GetNamespace(), o.GetName()))
			createdResources = append(createdResources, o)
		}
		BeforeEach(func() {
			logger.Info("resetting created resources list")
			createdResources = nil
		})
		AfterEach(func() {
			for i := len(createdResources) - 1; i >= 0; i-- {
				r := createdResources[i]
				key := client.ObjectKeyFromObject(r)
				logger.Info(
					"deleting resource",
					"namespace",
					key.Namespace,
					"name",
					key.Name,
					"test",
					CurrentSpecReport().FullText,
				)
				Expect(k8sClient.Delete(ctx, r)).To(Succeed())

				_, isNamespace := r.(*corev1.Namespace)
				if !isNamespace {
					logger.Info(
						"waiting for resource to disappear",
						"namespace",
						key.Namespace,
						"name",
						key.Name,
						"test",
						CurrentSpecReport().FullText,
					)
					Eventually(func() error {
						return k8sClient.Get(ctx, key, r)
					}, time.Minute, time.Second).Should(HaveOccurred())
					logger.Info(
						"deleted resource",
						"namespace",
						key.Namespace,
						"name",
						key.Name,
						"test",
						CurrentSpecReport().FullText,
					)
				}
			}
		})

		It("Should successfully reconcile a custom resource", func() {
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      TestName,
					Namespace: namespace.Name,
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

			By("Creating a Deployment to reference for tests")
			err := k8sClient.Create(ctx, deployment)
			deleteResourceAfterTest(deployment)
			Expect(err).To(Not(HaveOccurred()))

			By("Creating the custom resource")
			template := &api.ExecAccessTemplate{}
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      TestName,
				Namespace: namespace.Name,
			}, template)

			if err != nil && errors.IsNotFound(err) {
				template := &api.ExecAccessTemplate{
					ObjectMeta: metav1.ObjectMeta{
						Name:      TestName,
						Namespace: namespace.Name,
					},
					Spec: api.ExecAccessTemplateSpec{
						ControllerTargetRef: &api.CrossVersionObjectReference{
							APIVersion: "apps/v1",
							Kind:       "Deployment",
							Name:       deployment.Name,
						},
						AccessConfig: api.AccessConfig{
							AllowedGroups:   []string{"testGroupA"},
							DefaultDuration: "1h",
							MaxDuration:     "2h",
						},
					},
				}
				err = k8sClient.Create(ctx, template)
				deleteResourceAfterTest(template)
				Expect(err).To(Not(HaveOccurred()))
			}

			By("Checking if the custom resource was successfully created")
			Eventually(func() error {
				found := &api.ExecAccessTemplate{}
				return k8sClient.Get(ctx, types.NamespacedName{
					Name:      TestName,
					Namespace: namespace.Name,
				}, found)
			}, time.Minute, time.Second).Should(Succeed())

			By("Reconciling the custom resource")
			reconciler := &ExecAccessTemplateReconciler{
				BaseTemplateReconciler: BaseTemplateReconciler{
					BaseReconciler: BaseReconciler{
						Client:    k8sClient,
						Scheme:    k8sClient.Scheme(),
						APIReader: k8sClient,
					},
				},
			}
			_, err = reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      TestName,
					Namespace: namespace.Name,
				},
			})
			Expect(err).To(Not(HaveOccurred()))

			By("Verifying the resource became ready")
			Eventually(func() error {
				found := &api.ExecAccessTemplate{}
				_ = k8sClient.Get(ctx, types.NamespacedName{
					Name:      TestName,
					Namespace: namespace.Name,
				}, found)

				// Wait until the 2 conditions are met before checking the ready status. This ensures
				// a full reconciliation loop.
				if found.GetStatus().IsReady() {
					return nil
				}
				return fmt.Errorf(
					fmt.Sprintf(
						"Failed to reconcile resource: %s",
						strconv.FormatBool(found.GetStatus().IsReady()),
					),
				)
			}, 10*time.Second, time.Second).Should(Succeed())
		})

		It("Should fail to reconcile a resource with an invalid target", func() {
			By("Creating the custom resource")
			template := &api.ExecAccessTemplate{}
			err := k8sClient.Get(ctx, types.NamespacedName{
				Name:      TestName,
				Namespace: namespace.Name,
			}, template)

			if err != nil && errors.IsNotFound(err) {
				template := &api.ExecAccessTemplate{
					ObjectMeta: metav1.ObjectMeta{
						Name:      TestName,
						Namespace: namespace.Name,
					},
					Spec: api.ExecAccessTemplateSpec{
						AccessConfig: api.AccessConfig{
							// VALID
							AllowedGroups: []string{"testGroupA"},

							// INVALID: DefaultDuraiton cannot be longer than MaxDuration
							DefaultDuration: "2h",
							MaxDuration:     "1h",
						},

						// INVALID: This target does not exist
						ControllerTargetRef: &api.CrossVersionObjectReference{
							APIVersion: "apps/v1",
							Kind:       "Deployment",
							Name:       "invalid-name",
						},
					},
				}
				err = k8sClient.Create(ctx, template)
				deleteResourceAfterTest(template)
				Expect(err).To(Not(HaveOccurred()))
			}

			By("Checking if the custom resource was successfully created")
			Eventually(func() error {
				found := &api.ExecAccessTemplate{}
				return k8sClient.Get(ctx, types.NamespacedName{
					Name:      TestName,
					Namespace: namespace.Name,
				}, found)
			}, time.Minute, time.Second).Should(Succeed())

			By("Reconciling the custom resource")
			reconciler := &ExecAccessTemplateReconciler{
				BaseTemplateReconciler: BaseTemplateReconciler{
					BaseReconciler: BaseReconciler{
						Client:    k8sClient,
						Scheme:    k8sClient.Scheme(),
						APIReader: k8sClient,
					},
				},
			}
			_, err = reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      TestName,
					Namespace: namespace.Name,
				},
			})
			Expect(err).To(Not(HaveOccurred()))

			By("Verify that the TargetRefExists condition is False")
			Eventually(func() error {
				found := &api.ExecAccessTemplate{}
				_ = k8sClient.Get(ctx, types.NamespacedName{
					Name:      TestName,
					Namespace: namespace.Name,
				}, found)

				if meta.IsStatusConditionPresentAndEqual(
					*found.GetStatus().GetConditions(),
					string(ConditionTargetRefExists),
					metav1.ConditionFalse,
				) {
					// If the condition is set, and its set to False, then we can return success. We
					// failed appropriately.
					return nil
				}
				// Return a failure. We'll loop over this a few times before giving up.
				return fmt.Errorf(
					"Expected %s to be %s",
					ConditionTargetRefExists,
					metav1.ConditionFalse,
				)
			}, 10*time.Second, time.Second).Should(Succeed())

			By("Verify that the TargetDuration condition is False")
			Eventually(func() error {
				found := &api.ExecAccessTemplate{}
				_ = k8sClient.Get(ctx, types.NamespacedName{
					Name:      TestName,
					Namespace: namespace.Name,
				}, found)

				if meta.IsStatusConditionPresentAndEqual(
					*found.GetStatus().GetConditions(),
					string(ConditionDurationsValid),
					metav1.ConditionFalse,
				) {
					// If the condition is set, and its set to False, then we can return success. We
					// failed appropriately.
					logger.V(1).Info("shit")
					return nil
				}
				// Return a failure. We'll loop over this a few times before giving up.
				return fmt.Errorf(
					"Expected %s to be %s",
					ConditionTargetRefExists,
					metav1.ConditionFalse,
				)
			}, 10*time.Second, time.Second).Should(Succeed())
		})
	})
})
