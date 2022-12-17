package utils

import (
	"context"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/diranged/oz/internal/api/v1alpha1"
)

// CreateRole will create a Kubernetes Role for a specific Access Request with
// the supplied permissions. The OwnerReference is set to ensure proper
// cleanup.
func CreateRole(
	ctx context.Context,
	client client.Client,
	req v1alpha1.IRequestResource,
	rules []rbacv1.PolicyRule,
) (*rbacv1.Role, error) {
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      generateResourceName(req),
			Namespace: req.GetNamespace(),
		},
		Rules: rules,
	}

	// Set the OwnerRef before we try to create the object
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/owners-dependents/
	if err := ctrlutil.SetControllerReference(req, role, client.Scheme()); err != nil {
		return nil, err
	}

	// Generate an empty role resource. This role resource will be filled-in by the CreateOrUpdate() call when
	// it checks the Kubernetes API for the existing role. Our update function will then update the appropriate
	// values from the desired role object above.
	emptyRole := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{Name: role.Name, Namespace: role.Namespace},
	}

	// https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/controller/controllerutil#CreateOrUpdate
	if _, err := ctrlutil.CreateOrUpdate(ctx, client, emptyRole, func() error {
		emptyRole.ObjectMeta = role.ObjectMeta
		emptyRole.Rules = role.Rules
		emptyRole.OwnerReferences = role.OwnerReferences
		return nil
	}); err != nil {
		return nil, err
	}

	return role, nil
}
