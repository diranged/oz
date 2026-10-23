package podaccessbuilder

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/diranged/oz/internal/api/v1alpha1"
	bldutil "github.com/diranged/oz/internal/builders/utils"
	"github.com/diranged/oz/internal/imagepolicy"
)

// CreateAccessResources implements the IBuilder interface
//
// revive:disable:cyclomatic A linear sequence of build steps, each with its own error check
func (b *PodAccessBuilder) CreateAccessResources(
	ctx context.Context,
	client client.Client,
	req v1alpha1.IRequestResource,
	tmpl v1alpha1.ITemplateResource,
) (statusString string, err error) {
	log := logf.FromContext(ctx).WithName("CreateAccessResources")

	// Cast the Request into an PodAccessRequest.
	podReq := req.(*v1alpha1.PodAccessRequest)
	// Cast the Template into an PodAccessTemplate.
	podTmpl := tmpl.(*v1alpha1.PodAccessTemplate)

	// First, get the desired PodSpec. If there's a failure at this point, return it.
	podTemplateSpec, err := bldutil.GetPodTemplateFromController(ctx, client, tmpl)
	if err != nil {
		log.Error(err, "Failed to generate PodSpec for PodAccessRequest")
		return "", err
	}

	// Keep a reference to the unmutated spec. An image override needs it to
	// resolve the "default" container the same way PatchPodTemplateSpec() does
	// - the mutation config's JSON patches are free to rename containers, so
	// resolving the name against the mutated spec could fail to find a
	// container that existed when the template author named it.
	// PatchPodTemplateSpec() deep-copies its input, so this stays intact.
	origPodTemplateSpec := podTemplateSpec

	// Run the PodSpec through the optional mutation config
	mutator := podTmpl.Spec.ControllerTargetMutationConfig
	if mutator != nil {
		podTemplateSpec, err = mutator.PatchPodTemplateSpec(ctx, podTemplateSpec)
		if err != nil {
			log.Error(err, "Failed to mutate PodSpec for PodAccessRequest")
			return statusString, err
		}
	}

	// Finally, apply the requester's image override, if they asked for one.
	// This happens last so that it wins over anything the template's mutation
	// config did to the image.
	if err := applyImageOverride(ctx, podReq, mutator, origPodTemplateSpec, &podTemplateSpec); err != nil {
		log.Error(err, "Failed to apply image override for PodAccessRequest")
		return statusString, err
	}

	// Generate a Pod for the user to access
	pod, err := bldutil.CreatePod(ctx, client, podReq, podTemplateSpec)
	if err != nil {
		log.Error(err, "Failed to create Pod for AccessRequest")
		return statusString, err
	}

	// Define the permissions the access request will grant.
	//
	// TODO: Implement the ability to tune this in the PodAccessTemplate settings.
	rules := []rbacv1.PolicyRule{
		{
			APIGroups:     []string{corev1.GroupName},
			Resources:     []string{"pods"},
			ResourceNames: []string{pod.GetName()},
			Verbs:         []string{"get", "list", "watch"},
		},
		{
			APIGroups:     []string{corev1.GroupName},
			Resources:     []string{"pods/exec"},
			ResourceNames: []string{pod.GetName()},
			Verbs:         []string{"create", "update", "delete", "get", "list"},
		},
	}

	// Get the Role, or error out
	role, err := bldutil.CreateRole(ctx, client, podReq, rules)
	if err != nil {
		return statusString, err
	}

	// Get the Binding, or error out
	rb, err := bldutil.CreateRoleBinding(ctx, client, podReq, tmpl, role)
	if err != nil {
		return statusString, err
	}

	accessString, err := bldutil.CreateAccessCommand(podTmpl.Spec.AccessConfig.AccessCommand, pod.ObjectMeta)
	if err != nil {
		return "", err
	}
	podReq.Status.SetAccessMessage(accessString)

	// Set the podName (note, just in the local object). If this fails (for
	// example, its already set on the object), then we also bail out. This
	// only fails if the Status.PodName field has already been set, which would
	// indicate some kind of a reconcile loop conflict.
	//
	// Writing back into the cluster is not handled here - must be handled by
	// the caller of this method.
	if err := podReq.SetPodName(pod.GetName()); err != nil {
		return "", err
	}

	// We've been mutating the podReq Status throughout this build. Need to
	// push the update back to the cluster here.
	if err := client.Status().Update(ctx, podReq); err != nil {
		return "", err
	}

	statusString = fmt.Sprintf("Success. Pod %s, Role %s, RoleBinding %s created",
		pod.Name,
		role.Name,
		rb.Name,
	)
	return statusString, nil
}

// applyImageOverride replaces the image of the "default" container with the one
// requested in Spec.image, after re-checking it against the cluster's image
// policy.
//
// The policy is deliberately re-validated here rather than trusted from the
// admission webhook. The ValidatingWebhookConfiguration is an optional part of
// the deployment (see the chart's `webhook.create` value), so this is the only
// check that is guaranteed to run - and it is the last point before an
// arbitrary image would be handed to the Kubernetes API.
func applyImageOverride(
	ctx context.Context,
	req *v1alpha1.PodAccessRequest,
	mutator *v1alpha1.PodTemplateSpecMutationConfig,
	origPodTemplateSpec corev1.PodTemplateSpec,
	podTemplateSpec *corev1.PodTemplateSpec,
) error {
	if req.Spec.Image == "" {
		return nil
	}

	log := logf.FromContext(ctx).WithName("applyImageOverride")

	if err := imagepolicy.Active().Validate(req.Spec.Image); err != nil {
		return fmt.Errorf("spec.image is not allowed: %w", err)
	}

	// A PodAccessTemplate is not required to define a mutation config, so fall
	// back to an empty name and let the shared helper resolve the container
	// from the well-known annotation or position.
	var defaultContainerName string
	if mutator != nil {
		defaultContainerName = mutator.DefaultContainerName
	}

	id, err := v1alpha1.GetDefaultContainerID(ctx, origPodTemplateSpec, defaultContainerName)
	if err != nil {
		return err
	}

	// The container was resolved against the pre-mutation PodSpec, and a
	// template's patchSpecOperations are free to add or remove containers, so
	// the index is not guaranteed to still be in range.
	if id < 0 || id >= len(podTemplateSpec.Spec.Containers) {
		return fmt.Errorf(
			"cannot apply spec.image: container %d no longer exists after the template's mutations were applied",
			id,
		)
	}

	log.Info(
		"Overriding container image",
		"container", podTemplateSpec.Spec.Containers[id].Name,
		"from", podTemplateSpec.Spec.Containers[id].Image,
		"to", req.Spec.Image,
	)
	podTemplateSpec.Spec.Containers[id].Image = req.Spec.Image

	return nil
}
