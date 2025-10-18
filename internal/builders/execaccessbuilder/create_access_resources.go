package execaccessbuilder

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/diranged/oz/internal/api/v1alpha1"
	"github.com/diranged/oz/internal/builders/execaccessbuilder/internal"
	"github.com/diranged/oz/internal/builders/utils"
)

// CreateAccessResources implements the IBuilder interface
func (b *ExecAccessBuilder) CreateAccessResources(
	ctx context.Context,
	client client.Client,
	req v1alpha1.IRequestResource,
	tmpl v1alpha1.ITemplateResource,
) (statusString string, err error) {
	// Cast the Request into an ExecAccessRequest.
	execReq := req.(*v1alpha1.ExecAccessRequest)
	// Cast the Template into an ExecAccessTemplate.
	execTmpl := tmpl.(*v1alpha1.ExecAccessTemplate)

	// Get the target Pod Name that the user is going to have access to
	targetPod, err := internal.GetPod(ctx, client, execReq, execTmpl)
	if err != nil {
		return statusString, err
	}

	// Define the permissions the access request will grant.
	//
	// TODO: Implement the ability to tune this in the ExecAccessTemplate settings.
	rules := []rbacv1.PolicyRule{
		{
			APIGroups:     []string{corev1.GroupName},
			Resources:     []string{"pods"},
			ResourceNames: []string{targetPod.Name},
			Verbs:         []string{"get", "list", "watch"},
		},
		{
			APIGroups:     []string{corev1.GroupName},
			Resources:     []string{"pods/exec"},
			ResourceNames: []string{targetPod.Name},
			Verbs:         []string{"create", "update", "delete", "get", "list"},
		},
	}

	// Get the Role, or error out
	role, err := utils.CreateRole(ctx, client, execReq, rules)
	if err != nil {
		return statusString, err
	}

	// Get the Binding, or error out
	rb, err := utils.CreateRoleBinding(ctx, client, execReq, tmpl, role)
	if err != nil {
		return statusString, err
	}

	accessString, err := utils.CreateAccessCommand(execTmpl.Spec.AccessConfig.AccessCommand, targetPod.ObjectMeta)
	if err != nil {
		return "", err
	}
	execReq.Status.SetAccessMessage(accessString)

	// We've been mutating the execReq Status throughout this build. Need to
	// push the update back to the cluster here.
	if err := client.Status().Update(ctx, execReq); err != nil {
		return "", err
	}

	statusString = fmt.Sprintf("Success. Role %s, RoleBinding %s created", role.Name, rb.Name)
	return statusString, nil
}
