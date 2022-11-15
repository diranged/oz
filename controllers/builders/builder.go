package builders

import (
	"context"
	"errors"

	api "github.com/diranged/oz/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

type AccessBuilder struct {
	Client client.Client
	Ctx    context.Context
	Scheme *runtime.Scheme

	Request  *api.AccessRequest
	Template *api.AccessTemplate
}

func (t *AccessBuilder) GetDeployment() (*appsv1.Deployment, error) {
	found := &appsv1.Deployment{}

	err := t.Client.Get(t.Ctx, types.NamespacedName{
		Name:      *t.Template.Spec.TargetRef.Name,
		Namespace: t.Template.Namespace,
	}, found)
	return found, err
}

func (t *AccessBuilder) GetDaemonSet() (*appsv1.DaemonSet, error) {
	found := &appsv1.DaemonSet{}
	err := t.Client.Get(t.Ctx, types.NamespacedName{
		Name:      *t.Template.Spec.TargetRef.Name,
		Namespace: t.Template.Namespace,
	}, found)
	return found, err
}

func (t *AccessBuilder) GetStatefulSet() (*appsv1.StatefulSet, error) {
	found := &appsv1.StatefulSet{}
	err := t.Client.Get(t.Ctx, types.NamespacedName{
		Name:      *t.Template.Spec.TargetRef.Name,
		Namespace: t.Template.Namespace,
	}, found)
	return found, err
}

// https://medium.com/coding-kubernetes/using-k8s-label-selectors-in-go-the-right-way-733cde7e8630
func (t *AccessBuilder) GetTargetPodSelectorLabels(cl client.Client, ctx context.Context) (labels.Selector, error) {
	// https://sdk.operatorframework.io/docs/building-operators/golang/references/logging/
	logger := ctrllog.FromContext(ctx)

	// Get the controller - if there's any error, return.
	if t.Template.Spec.TargetRef.Kind == api.DeploymentController {
		controller, err := t.GetDeployment()
		if err != nil {
			logger.Error(err, "Failed to find target Deployment")
			return nil, err
		}
		return metav1.LabelSelectorAsSelector(controller.Spec.Selector)
	} else if t.Template.Spec.TargetRef.Kind == api.DaemonSetController {
		controller, err := t.GetDaemonSet()
		if err != nil {
			logger.Error(err, "Failed to find target DaemonSet")
			return nil, err
		}
		return metav1.LabelSelectorAsSelector(controller.Spec.Selector)
	} else if t.Template.Spec.TargetRef.Kind == api.StatefulSetController {
		controller, err := t.GetStatefulSet()
		if err != nil {
			logger.Error(err, "Failed to find target StatefulSet")
			return nil, err
		}
		return metav1.LabelSelectorAsSelector(controller.Spec.Selector)
	}
	return nil, errors.New("invalid input")
}

// func (b *AccessBuilder) GenerateAccessRole() (*rbacv1.Role, error) {
// 	role := &rbacv1.Role{}
//
// 	role.Name = fmt.Sprintf("%s-%s", b.Request.Name, b.Request.GetUniqueId())
// 	role.Namespace = b.Template.Namespace
// 	role.Rules = []rbacv1.PolicyRule{
// 		{
// 			APIGroups:     []string{corev1.GroupName},
// 			Resources:     []string{"pods"},
// 			ResourceNames: []string{b.Request.Status.PodName},
// 			Verbs:         []string{"get", "list", "watch"},
// 		},
// 		{
// 			APIGroups:     []string{corev1.GroupName},
// 			Resources:     []string{"pods/exec"},
// 			ResourceNames: []string{b.Request.Status.PodName},
// 			Verbs:         []string{"create", "update", "delete", "get", "list"},
// 		},
// 	}
//
// 	// Set the ownerRef for the Deployment
// 	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/owners-dependents/
// 	if err := ctrl.SetControllerReference(b.Request, role, b.Scheme); err != nil {
// 		return nil, err
// 	}
//
// 	return role, nil
// }
//
// func (b *AccessBuilder) GenerateAccessRoleBinding() (*rbacv1.RoleBinding, error) {
// 	rb := &rbacv1.RoleBinding{}
//
// 	rb.Name = fmt.Sprintf("%s-%s", b.Request.Name, b.Request.GetUniqueId())
// 	rb.Namespace = b.Template.Namespace
// 	rb.RoleRef = rbacv1.RoleRef{
// 		APIGroup: rbacv1.GroupName,
// 		Kind:     "Role",
// 		Name:     rb.Name,
// 	}
// 	rb.Subjects = []rbacv1.Subject{}
//
// 	for _, group := range b.Template.Spec.AllowedGroups {
// 		rb.Subjects = append(rb.Subjects, rbacv1.Subject{
// 			APIGroup: rbacv1.SchemeGroupVersion.Group,
// 			Kind:     rbacv1.GroupKind,
// 			Name:     group,
// 		})
// 	}
//
// 	// Set the ownerRef for the Deployment
// 	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/owners-dependents/
// 	if err := ctrl.SetControllerReference(b.Request, rb, b.Scheme); err != nil {
// 		return nil, err
// 	}
//
// 	return rb, nil
// }
//
