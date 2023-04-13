package podaccessbuilder

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/diranged/oz/internal/api/v1alpha1"
	"github.com/diranged/oz/internal/builders/utils"
)

// CreateAccessResources implements the IBuilder interface
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
	podTemplateSpec, err := utils.GetPodTemplateFromController(ctx, client, tmpl)
	if err != nil {
		log.Error(err, "Failed to generate PodSpec for PodAccessRequest")
		return "", err
	}

	// Run the PodSpec through the optional mutation config
	mutator := podTmpl.Spec.ControllerTargetMutationConfig
	if mutator != nil {
		podTemplateSpec, err = mutator.PatchPodTemplateSpec(ctx, podTemplateSpec)
		if err != nil {
			log.Error(err, "Failed to mutate PodSpec for PodAccessRequest")
			return statusString, err
		}
	}

	// Generate a Pod for the user to access
	pod, err := utils.CreatePod(ctx, client, podReq, podTemplateSpec)
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
	role, err := utils.CreateRole(ctx, client, podReq, rules)
	if err != nil {
		return statusString, err
	}

	// Get the Binding, or error out
	rb, err := utils.CreateRoleBinding(ctx, client, podReq, tmpl, role)
	if err != nil {
		return statusString, err
	}

	accessString, err := utils.CreateAccessCommand(podTmpl.Spec.AccessConfig.AccessCommand, pod.ObjectMeta)
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
