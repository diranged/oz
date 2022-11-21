package v1alpha1

import (
	"fmt"
	"reflect"

	corev1 "k8s.io/api/core/v1"
)

// PodSpecMutationConfig provides a common pattern for describing mutations to an existing PodSpec
// that should be applied. The primary use case is in the PodAccessTemplate, where an existing
// controller (Deployment, DaemonSet, StatefulSet) can be used as the reference for the PodSpec
// that is launched for the user. However, the operator may want to make modifications to the
// PodSpec at launch time (eg, change the entrypoint command or arguments).
//
// TODO: Add podAnnotations
// TODO: Add podLabels
// TODO: Add nodeSelector
// TODO: Add affinity
type PodSpecMutationConfig struct {
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
}

// getDefaultContainerID returns the numerical identifier of the container within the
// PodSpec.Containers[] list that the mutation configuration should apply to.
//
// Returns:
//
//	int: The identifier in the PodSpec.Containers[] list of the "default" container to mutate.
func (c *PodSpecMutationConfig) getDefaultContainerID(pod corev1.PodSpec) (int, error) {
	// By default, return 0.
	if c.DefaultContainerName == "" {
		return 0, nil
	}

	// Iterate through the containers
	for i, container := range pod.Containers {
		if container.Name == c.DefaultContainerName {
			return i, nil
		}
	}

	// Finally, return 0 if no match found
	return -1, fmt.Errorf("could not find container named %s in PodSpec", c.DefaultContainerName)
}

// PatchPodSpec returns a mutated new PodSpec object based on the supplied spec, and the parameters
// in the PodSpecMutationConfig struct.
//
// Returns:
//
//	corev1.PodSpec: A new PodSpec object with the mutated configuration.
func (c *PodSpecMutationConfig) PatchPodSpec(orig corev1.PodSpec) (corev1.PodSpec, error) {
	n := *orig.DeepCopy()

	defContainerID, err := c.getDefaultContainerID(orig)
	if err != nil {
		return orig, err
	}

	if c.Command != nil {
		n.Containers[defContainerID].Command = *c.Command
		n.Containers[defContainerID].Args = []string{}
	}

	if c.Args != nil {
		n.Containers[defContainerID].Args = *c.Args
	}

	if !reflect.DeepEqual(c.Resources, corev1.ResourceRequirements{}) {
		n.Containers[defContainerID].Resources = c.Resources
	}

	if len(c.Env) > 0 {
		n.Containers[defContainerID].Env = append(n.Containers[defContainerID].Env, c.Env...)
	}

	return n, nil
}
