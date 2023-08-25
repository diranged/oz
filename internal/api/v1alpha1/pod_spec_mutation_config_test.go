package v1alpha1

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("PodSpecMutationConfig", Ordered, func() {
	var podTemplateSpec corev1.PodTemplateSpec

	BeforeEach(func() {
		// Create a fake podTemplateSpec target
		termPeriod := int64(300)
		podTemplateSpec = v1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					"Foo": "bar",
				},
				Labels: map[string]string{
					"testLabel": "testValue",
				},
			},
			Spec: corev1.PodSpec{
				TerminationGracePeriodSeconds: &termPeriod,
				TopologySpreadConstraints: []corev1.TopologySpreadConstraint{
					{
						TopologyKey:       "topology.kubernetes.io/zone",
						MaxSkew:           1,
						WhenUnsatisfiable: corev1.DoNotSchedule,
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"test": "test",
							},
						},
					},
				},
				SecurityContext: &corev1.PodSecurityContext{},
				Containers: []corev1.Container{
					{
						Name:  "contA",
						Image: "nginx:latest",
						LivenessProbe: &v1.Probe{
							ProbeHandler: v1.ProbeHandler{
								Exec: &v1.ExecAction{
									Command: []string{"/bin/true"},
								},
							},
							InitialDelaySeconds: 5,
							TimeoutSeconds:      5,
							PeriodSeconds:       30,
							SuccessThreshold:    1,
							FailureThreshold:    3,
						},
						ReadinessProbe: &v1.Probe{
							ProbeHandler: v1.ProbeHandler{
								Exec: &v1.ExecAction{
									Command: []string{"/bin/true"},
								},
							},
							InitialDelaySeconds: 5,
							TimeoutSeconds:      5,
							PeriodSeconds:       30,
							SuccessThreshold:    1,
							FailureThreshold:    3,
						},
						StartupProbe: &v1.Probe{
							ProbeHandler: v1.ProbeHandler{
								Exec: &v1.ExecAction{
									Command: []string{"/bin/true"},
								},
							},
							InitialDelaySeconds: 5,
							TimeoutSeconds:      5,
							PeriodSeconds:       30,
							SuccessThreshold:    1,
							FailureThreshold:    3,
						},
					},
					{
						Name:  "contB",
						Image: "nginx:latest",
					},
				},
			},
		}
	})

	Context("Functions()", func() {
		It("getDefaultContainerID should return 0 by default", func() {
			// Basic resource with no mutation config
			config := &PodTemplateSpecMutationConfig{}

			// Run it
			ret, err := config.getDefaultContainerID(ctx, podTemplateSpec)
			Expect(err).To(Not(HaveOccurred()))

			// VERIFY: We got back '0' by default
			Expect(ret).To(Equal(0))
		})
		It(
			"getDefaultContainerID should return 1 by if we set the container name to contB",
			func() {
				// Basic resource with simple container identifier
				config := &PodTemplateSpecMutationConfig{DefaultContainerName: "contB"}

				// Run it
				ret, err := config.getDefaultContainerID(ctx, podTemplateSpec)
				Expect(err).To(Not(HaveOccurred()))

				// VERIFY: We got back '1'
				Expect(ret).To(Equal(1))
			},
		)

		It(
			"getDefaultContainerID should return 1 by if the annotation is set",
			func() {
				// Patch the deployment
				podTemplateSpec.Annotations[DefaultContainerAnnotationKey] = "contB"

				// Basic resource with no mutation config
				config := &PodTemplateSpecMutationConfig{}

				// Run it
				ret, err := config.getDefaultContainerID(ctx, podTemplateSpec)
				Expect(err).To(Not(HaveOccurred()))

				// VERIFY: We got back '1'
				Expect(ret).To(Equal(1))
			},
		)

		It("PatchPodTemplateSpec should return only default mutations normally", func() {
			// Basic resource with no mutation config
			config := &PodTemplateSpecMutationConfig{}

			// Run it
			ret, err := config.PatchPodTemplateSpec(ctx, podTemplateSpec)
			Expect(err).To(Not(HaveOccurred()))

			// Apply the default mutations to the original pod template spec,
			// these are the mutations we expect because they are hard-coded.
			expectedPodTemplateSpec := podTemplateSpec.DeepCopy()
			// Wipe: TerminationGracePeriodSeconds
			expectedPodTemplateSpec.Spec.TerminationGracePeriodSeconds = nil
			// Wipe: livenessProbe
			expectedPodTemplateSpec.Spec.Containers[0].LivenessProbe = nil
			// Wipe: readinessProbe
			expectedPodTemplateSpec.Spec.Containers[0].ReadinessProbe = nil
			// Wipe: startupProbe
			expectedPodTemplateSpec.Spec.Containers[0].StartupProbe = nil
			// Wipe: metadata.labels
			expectedPodTemplateSpec.ObjectMeta.Labels = map[string]string{}
			// Wipe: topologySpreadConstraints
			expectedPodTemplateSpec.Spec.TopologySpreadConstraints = nil

			// VERIFY: Unmutated by default
			Expect(ret.DeepCopy()).To(Equal(expectedPodTemplateSpec))
		})

		It("PatchPodTemplateSpec should allow skipping the default mutations", func() {
			// Basic resource with no mutation config
			config := &PodTemplateSpecMutationConfig{
				KeepTerminationGracePeriod:    true,
				KeepLivenessProbe:             true,
				KeepStartupProbe:              true,
				KeepReadinessProbe:            true,
				KeepTopologySpreadConstraints: true,
			}

			// Run it
			ret, err := config.PatchPodTemplateSpec(ctx, podTemplateSpec)
			Expect(err).To(Not(HaveOccurred()))

			// Wipe: metadata.labels (not optional)
			expectedPodTemplateSpec := podTemplateSpec.DeepCopy()
			expectedPodTemplateSpec.ObjectMeta.Labels = map[string]string{}

			// VERIFY: Unmutated by default
			Expect(ret.DeepCopy()).To(Equal(expectedPodTemplateSpec))
		})

		It("PatchPodTemplateSpec should purge default annotations if requested", func() {
			// Basic resource with no mutation config
			config := &PodTemplateSpecMutationConfig{
				PurgeAnnotations: true,
				PodAnnotations: &map[string]string{
					"TestAnnotation": "value",
				},
			}

			// Run it
			ret, err := config.PatchPodTemplateSpec(ctx, podTemplateSpec)
			Expect(err).To(Not(HaveOccurred()))

			// VERIFY: Only one annotation found
			Expect(ret.ObjectMeta.Annotations).To(Equal(
				map[string]string{
					"TestAnnotation": "value",
				},
			))
		})

		It("PatchPodTemplateSpec should mutate command", func() {
			// Basic resource with no mutation config
			config := &PodTemplateSpecMutationConfig{
				Command: &[]string{"/bin/sleep"},
				Args:    &[]string{"100"},
				Resources: corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						corev1.ResourceCPU: *resource.NewQuantity(1, resource.Format("DecimalExponent")),
					},
				},
				PodAnnotations: &map[string]string{
					"TestAnnotation": "value",
				},
				PodLabels: &map[string]string{
					"TestLabelTwo": "bar",
				},
				Env: []corev1.EnvVar{
					{Name: "FOO", Value: "BAR"},
				},
			}

			// Run it
			ret, err := config.PatchPodTemplateSpec(ctx, podTemplateSpec)
			Expect(err).To(Not(HaveOccurred()))

			// VERIFY: Command/Args is set
			Expect(ret.Spec.Containers[0].Command[0]).To(Equal("/bin/sleep"))
			Expect(ret.Spec.Containers[0].Args[0]).To(Equal("100"))

			// VERIFY: New annotation is inserted
			Expect(ret.ObjectMeta.Annotations["TestAnnotation"]).To(Equal("value"))
			Expect(ret.ObjectMeta.Labels["TestLabelTwo"]).To(Equal("bar"))

			// VERIFY: Resources are set
			Expect(ret.Spec.Containers[0].Resources.Limits.Cpu().String()).To(Equal("1"))

			// VERIFY: EnvVar is set
			Expect(len(ret.Spec.Containers[0].Env)).To(Equal(1))
		})

		It("PatchPodTemplateSpec should fail if invalid container name supplied", func() {
			// Basic resource with no mutation config
			config := &PodTemplateSpecMutationConfig{DefaultContainerName: "bogus"}

			// Run it
			ret, err := config.PatchPodTemplateSpec(ctx, podTemplateSpec)
			Expect(err).To(HaveOccurred())

			// VERIFY: Unmutated
			Expect(ret).To(Equal(podTemplateSpec))
		})

		It("PatchPodTemplateSpec should add node selectors if requested (not initially present)", func() {
			// Basic resource with no mutation config
			config := &PodTemplateSpecMutationConfig{
				PurgeAnnotations: true,
				NodeSelector: &map[string]string{
					"selector": "value",
				},
			}

			// Run it
			ret, err := config.PatchPodTemplateSpec(ctx, podTemplateSpec)
			Expect(err).To(Not(HaveOccurred()))

			// VERIFY: Only one annotation found
			Expect(ret.Spec.NodeSelector).To(Equal(
				map[string]string{
					"selector": "value",
				},
			))
		})

		It("PatchPodTemplateSpec should add node selectors if requested (initially present)", func() {
			// Basic resource with no mutation config
			config := &PodTemplateSpecMutationConfig{
				PurgeAnnotations: true,
				NodeSelector: &map[string]string{
					"selector": "value",
				},
			}

			// Run it
			podTemplateSpec.Spec.NodeSelector = map[string]string{
				"already": "there",
			}
			ret, err := config.PatchPodTemplateSpec(ctx, podTemplateSpec)
			Expect(err).To(Not(HaveOccurred()))

			// VERIFY: Only one annotation found
			Expect(ret.Spec.NodeSelector).To(Equal(
				map[string]string{
					"already":  "there",
					"selector": "value",
				},
			))
		})

		It("PatchPodTemplateSpec should pod security context if supplied", func() {
			// Basic resource with no mutation config
			sc := &corev1.SecurityContext{
				Capabilities: &corev1.Capabilities{
					Add: []corev1.Capability{"SYS_PTRACE"},
				},
			}
			config := &PodTemplateSpecMutationConfig{
				SecurityContext: sc,
			}

			// Run it
			ret, err := config.PatchPodTemplateSpec(ctx, podTemplateSpec)
			Expect(err).To(Not(HaveOccurred()))

			// VERIFY: container security context is modified
			Expect(ret.Spec.Containers[0].SecurityContext).To(Equal(sc))
		})
	})
})
