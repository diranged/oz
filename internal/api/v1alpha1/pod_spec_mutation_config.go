package v1alpha1

import (
	"context"
	"fmt"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	// DefaultContainerAnnotationKey is the name of the Key in the Pod
	// Annotations that notates which container in the PodSpec is considered
	// the "default" container for kubectl. This annotation is also used to
	// determine which container is mutated by the
	// PodTemplateSpecMutationConfig struct.
	DefaultContainerAnnotationKey = "kubectl.kubernetes.io/default-container"
)

// PodTemplateSpecMutationConfig provides a common pattern for describing mutations to an existing PodSpec
// that should be applied. The primary use case is in the PodAccessTemplate, where an existing
// controller (Deployment, DaemonSet, StatefulSet) can be used as the reference for the PodSpec
// that is launched for the user. However, the operator may want to make modifications to the
// PodSpec at launch time (eg, change the entrypoint command or arguments).
//
// TODO: Add podAnnotations
// TODO: Add podLabels
// TODO: Add nodeSelector
// TODO: Add affinity
type PodTemplateSpecMutationConfig struct {
	// DefaultContainerName allows the operator to define which container is considered the default
	// container, and that is the container that this mutation configuration applies to. If not set,
	// then the first container defined in the spec.containers[] list is patched.
	DefaultContainerName string `json:"defaultContainerName,omitempty"`

	// Command is used to override the .Spec.containers[0].command field for the target Pod and
	// Container. This can be handy in ensuring that the default application does not start up and
	// do any work. If set, this overrides the Spec.conatiners[0].args property as well.
	Command *[]string `json:"command,omitempty"`

	// Args will override the Spec.containers[0].args property.
	Args *[]string `json:"args,omitempty"`

	// Env allows overriding specific environment variables (or adding new ones). Note, we do not
	// purge the original environmnt variables.
	Env []corev1.EnvVar `json:"env,omitempty"`

	// If supplied these resource requirements will override the default .Spec.containers[0].resource requested for the
	// the pod. Note though that we do not override all of the resource requests in the Pod because there may be many
	// containers.
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// If supplied, these
	// [annotations](https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/)
	// are applied to the target
	// [`PodTemplateSpec`](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#podtemplatespec-v1-core).
	// These are merged into the final Annotations. If you want to _replace_
	// the annotations, make sure to set the `purgeAnnotations` flag to `true`.
	PodAnnotations *map[string]string `json:"podAnnotations,omitempty"`

	// If supplied, Oz will insert these
	// [labels](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/)
	// into the target
	// [`PodTemplateSpec`](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#podtemplatespec-v1-core).
	// By default Oz purges all Labels from pods (to prevent the new Pod from
	// having traffic routed to it), so this is effectively a new set of labels
	// applied to the Pod.
	PodLabels *map[string]string `json:"podLabels,omitempty"`

	// By default, Oz keeps the original
	// [`PodTemplateSpec`](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#podtemplatespec-v1-core)
	// `metadata.annotations` field. If you want to purge this, set this flag
	// to `true.`
	//
	// +kubebuilder:default:=false
	PurgeAnnotations bool `json:"purgeAnnotations,omitempty"`

	// By default, Oz wipes out the PodSpec
	// [`terminationGracePeriodSeconds`](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#podspec-v1-core)
	// setting on Pods to ensure that they can be killed as soon as the
	// AccessRequest expires. This flag overrides that behavior.
	//
	// +kubebuilder:default:=false
	KeepTerminationGracePeriod bool `json:"keepTerminationGracePeriod,omitempty"`

	// By default, Oz wipes out the PodSpec
	// [`livenessProbe`](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#podspec-v1-core)
	// configuration for the default container so that the container does not
	// get terminated if the main application is not running or passing checks.
	// This setting overrides that behavior.
	//
	// +kubebuilder:default:=false
	KeepLivenessProbe bool `json:"keepLivenessProbe,omitempty"`

	// By default, Oz wipes out the PodSpec
	// [`readinessProbe`](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#podspec-v1-core)
	// configuration for the default container so that the container does not
	// get terminated if the main application is not running or passing checks.
	// This setting overrides that behavior.
	//
	// +kubebuilder:default:=false
	KeepReadinessProbe bool `json:"keepReadinessProbe,omitempty"`

	// By default, Oz wipes out the PodSpec
	// [`startupProbe`](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#podspec-v1-core)
	// configuration for the default container so that the container does not
	// get terminated if the main application is not running or passing checks.
	// This setting overrides that behavior.
	//
	// +kubebuilder:default:=false
	KeepStartupProbe bool `json:"keepStartupProbe,omitempty"`
}

// getDefaultContainerID returns the numerical identifier of the container within the
// PodSpec.Containers[] list that the mutation configuration should apply to.
//
// Returns:
//
//	int: The identifier in the PodSpec.Containers[] list of the "default" container to mutate.
func (c *PodTemplateSpecMutationConfig) getDefaultContainerID(
	ctx context.Context,
	pod corev1.PodTemplateSpec,
) (int, error) {
	logger := log.FromContext(ctx)
	logger.V(1).Info("Determining \"default\" container ID from PodTemplateSpec...")

	// Temporary placeholder for the default container name we're going to look for.
	var defContName string

	// If the user did not supply a DefaultContainerName spec, then try to find
	// the well known annotation.
	if c.DefaultContainerName == "" {
		if val, ok := pod.ObjectMeta.Annotations[DefaultContainerAnnotationKey]; ok {
			if ok {
				logger.V(1).
					Info(fmt.Sprintf("%s annotation detected, using %s", DefaultContainerAnnotationKey, val))
				defContName = val
			}
		}
	} else {
		logger.V(1).Info(fmt.Sprintf("Using template-supplied value %s", c.DefaultContainerName))
		defContName = c.DefaultContainerName
	}

	// At this point, if we didn't find the user supplied value OR the default
	// annotation field, we return 0.
	if defContName == "" {
		// Return 0 if no annotation was found either
		logger.V(1).Info("No configuration detected, returning container ID 0")
		return 0, nil
	}

	// Iterate through the containers
	for i, container := range pod.Spec.Containers {
		if container.Name == defContName {
			logger.V(1).
				Info(fmt.Sprintf("Discovered default container, returning container ID %d", i))
			return i, nil
		}
	}

	// Finally, return 0 if no match found
	return -1, fmt.Errorf("could not find container named %s in PodSpec", defContName)
}

// PatchPodTemplateSpec returns a mutated new PodSpec object based on the
// supplied spec, and the parameters in the PodSpecMutationConfig struct.
//
// Returns:
//
//	corev1.PodSpec: A new PodSpec object with the mutated configuration.
//
// revive:disable:cyclomatic High complexity score but easy to understand
func (c *PodTemplateSpecMutationConfig) PatchPodTemplateSpec(
	ctx context.Context,
	orig corev1.PodTemplateSpec,
) (corev1.PodTemplateSpec, error) {
	logger := log.FromContext(ctx)
	n := *orig.DeepCopy()

	defContainerID, err := c.getDefaultContainerID(ctx, orig)
	if err != nil {
		return orig, err
	}

	// By default we purge the Spec.terminationGracePeriodSeconds value.
	if !c.KeepTerminationGracePeriod {
		logger.V(1).Info("Purging spec.terminationGracePeriodSeconds...")
		n.Spec.TerminationGracePeriodSeconds = nil
	}
	if !c.KeepLivenessProbe {
		logger.V(1).
			Info(fmt.Sprintf("Purging spec.containers[%d].livenessProbe...", defContainerID))
		n.Spec.Containers[defContainerID].LivenessProbe = nil
	}

	if !c.KeepReadinessProbe {
		logger.V(1).
			Info(fmt.Sprintf("Purging spec.containers[%d].readinessProbe...", defContainerID))
		n.Spec.Containers[defContainerID].ReadinessProbe = nil
	}

	if !c.KeepStartupProbe {
		logger.V(1).
			Info(fmt.Sprintf("Purging spec.containers[%d].startupProbe...", defContainerID))
		n.Spec.Containers[defContainerID].StartupProbe = nil
	}

	if c.PurgeAnnotations {
		logger.V(1).Info("Purging metadata.annotations...")
		n.ObjectMeta.Annotations = map[string]string{}
	}

	if c.PodAnnotations != nil {
		for k, v := range *c.PodAnnotations {
			logger.V(1).Info(fmt.Sprintf("Setting metadata.annotations.%s: %s", k, v))
			n.ObjectMeta.Annotations[k] = v
		}
	}

	// Always purge the metadata.labels before moving forward. We do this to
	// ensure that we never launch a Pod that is going to accept traffic for
	// part of a service.
	//
	// TODO: Figure out how to use controller selector labels to purge more
	// selectively in the future.
	n.ObjectMeta.Labels = map[string]string{}

	if c.PodLabels != nil {
		for k, v := range *c.PodLabels {
			logger.V(1).Info(fmt.Sprintf("Setting metadata.labels.%s: %s", k, v))
			n.ObjectMeta.Labels[k] = v
		}
	}

	if c.Command != nil {
		logger.V(1).Info(fmt.Sprintf("Overriding spec.containers[%d].command...", defContainerID))
		n.Spec.Containers[defContainerID].Command = *c.Command
		n.Spec.Containers[defContainerID].Args = []string{}
	}

	if c.Args != nil {
		logger.V(1).Info(fmt.Sprintf("Overriding spec.containers[%d].args...", defContainerID))
		n.Spec.Containers[defContainerID].Args = *c.Args
	}

	if !reflect.DeepEqual(c.Resources, corev1.ResourceRequirements{}) {
		logger.V(1).Info(fmt.Sprintf("Overriding spec.containers[%d].resources...", defContainerID))
		n.Spec.Containers[defContainerID].Resources = c.Resources
	}

	if len(c.Env) > 0 {
		logger.V(1).Info(fmt.Sprintf("Adding spec.containers[%d].env...", defContainerID))
		n.Spec.Containers[defContainerID].Env = append(
			n.Spec.Containers[defContainerID].Env,
			c.Env...)
	}

	return n, nil
}
