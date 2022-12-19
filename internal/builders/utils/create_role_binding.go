package utils

import (
	"context"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/diranged/oz/internal/api/v1alpha1"
)

// CreateRoleBinding will create a RoleBinding to a Role for a set of Groups
// defined in an Access Template.
func CreateRoleBinding(
	ctx context.Context,
	client client.Client,
	req v1alpha1.IRequestResource,
	tmpl v1alpha1.ITemplateResource,
	role *rbacv1.Role,
) (*rbacv1.RoleBinding, error) {
	rb := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GenerateResourceName(req),
			Namespace: req.GetNamespace(),
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "Role",
			Name:     role.Name,
		},
		Subjects: []rbacv1.Subject{},
	}

	for _, group := range tmpl.GetAccessConfig().GetAllowedGroups() {
		rb.Subjects = append(rb.Subjects, rbacv1.Subject{
			APIGroup: rbacv1.SchemeGroupVersion.Group,
			Kind:     rbacv1.GroupKind,
			Name:     group,
		})
	}

	// Set the ownerRef for the Deployment
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/owners-dependents/
	if err := ctrlutil.SetControllerReference(req, rb, client.Scheme()); err != nil {
		return nil, err
	}

	// Generate an empty role resource. This role resource will be filled-in by the CreateOrUpdate() call when
	// it checks the Kubernetes API for the existing role. Our update function will then update the appropriate
	// values from the desired role object above.
	emptyRb := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: rb.Name, Namespace: rb.Namespace},
	}

	// https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/controller/controllerutil#CreateOrUpdate
	if _, err := ctrlutil.CreateOrUpdate(ctx, client, emptyRb, func() error {
		emptyRb.ObjectMeta = rb.ObjectMeta
		emptyRb.RoleRef = rb.RoleRef
		emptyRb.Subjects = rb.Subjects
		emptyRb.OwnerReferences = rb.OwnerReferences
		return nil
	}); err != nil {
		return nil, err
	}

	return rb, nil
}
