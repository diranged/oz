package builders

import (
	"context"
	"errors"
	"fmt"
	"math/rand"

	api "github.com/diranged/oz/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

type ExecAccessBuilder struct {
	Client client.Client
	Ctx    context.Context
	Scheme *runtime.Scheme

	Request  *api.ExecAccessRequest
	Template *api.ExecAccessTemplate
}

func (t *ExecAccessBuilder) GetDeployment() (*appsv1.Deployment, error) {
	found := &appsv1.Deployment{}

	err := t.Client.Get(t.Ctx, types.NamespacedName{
		Name:      *t.Template.Spec.TargetRef.Name,
		Namespace: t.Template.Namespace,
	}, found)
	return found, err
}

func (t *ExecAccessBuilder) GetDaemonSet() (*appsv1.DaemonSet, error) {
	found := &appsv1.DaemonSet{}
	err := t.Client.Get(t.Ctx, types.NamespacedName{
		Name:      *t.Template.Spec.TargetRef.Name,
		Namespace: t.Template.Namespace,
	}, found)
	return found, err
}

func (t *ExecAccessBuilder) GetStatefulSet() (*appsv1.StatefulSet, error) {
	found := &appsv1.StatefulSet{}
	err := t.Client.Get(t.Ctx, types.NamespacedName{
		Name:      *t.Template.Spec.TargetRef.Name,
		Namespace: t.Template.Namespace,
	}, found)
	return found, err
}

// https://medium.com/coding-kubernetes/using-k8s-label-selectors-in-go-the-right-way-733cde7e8630
func (t *ExecAccessBuilder) GetTargetPodSelectorLabels(cl client.Client, ctx context.Context) (labels.Selector, error) {
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

func (t *ExecAccessBuilder) GetRandomPod() (*corev1.Pod, error) {
	// https://sdk.operatorframework.io/docs/building-operators/golang/references/logging/
	logger := ctrllog.FromContext(t.Ctx)
	logger.Info("Finding Pods...")

	// Will populate this further down
	pod := &corev1.Pod{}

	// https://medium.com/coding-kubernetes/using-k8s-label-selectors-in-go-the-right-way-733cde7e8630
	selector, err := t.GetTargetPodSelectorLabels(t.Client, t.Ctx)
	if err != nil {
		logger.Error(err, "Failed to find label selector, cannot automatically discover pods")
		return nil, err
	}

	// List all of the pods in the Deployment by searching for matching pods with the current Label
	// Selector.
	podList := &corev1.PodList{}
	opts := []client.ListOption{
		client.InNamespace(t.Template.Namespace),
		client.MatchingLabelsSelector{
			Selector: selector,
		},
		// TODO: Figure this out...
		client.MatchingFields{"status.phase": "Running"},
	}
	if err := t.Client.List(t.Ctx, podList, opts...); err != nil {
		logger.Error(err, "Failed to retrieve Pod list")
		return nil, err
	}

	if len(podList.Items) < 1 {
		return nil, fmt.Errorf("no pods found maching selector")
	}

	// Randomly generate a number from within the length of the returned pod list...
	randomIndex := rand.Intn(len(podList.Items))

	// Return the randomly generated Pod
	logger.Info(fmt.Sprintf("Returning Pod %s", pod.Name))
	pod = &podList.Items[randomIndex]

	return pod, err
}

func (b *ExecAccessBuilder) GetSpecificPod() (*corev1.Pod, error) {
	podName := b.Request.Spec.TargetPod

	// https://sdk.operatorframework.io/docs/building-operators/golang/references/logging/
	logger := ctrllog.FromContext(b.Ctx)
	logger.Info(fmt.Sprintf("Looking for Pod %s", podName))

	// https://medium.com/coding-kubernetes/using-k8s-label-selectors-in-go-the-right-way-733cde7e8630
	selector, err := b.GetTargetPodSelectorLabels(b.Client, b.Ctx)
	if err != nil {
		logger.Error(err, "Failed to find label selector, cannot automatically discover pods")
		return nil, err
	}

	// List all of the pods in the Deployment by searching for matching pods with the current Label
	// Selector.
	podList := &corev1.PodList{}
	opts := []client.ListOption{
		client.InNamespace(b.Template.Namespace),
		client.MatchingLabelsSelector{
			Selector: selector,
		},
		client.MatchingFields{"metadata.name": podName},
		// TODO: Figure this out...
		//client.MatchingFields{"status.phase": "Running"},
	}
	if err := b.Client.List(b.Ctx, podList, opts...); err != nil {
		logger.Error(err, "Failed to retrieve Pod list")
		return nil, err
	}
	if len(podList.Items) < 1 {
		return nil, fmt.Errorf("pod named %s not found", podName)
	}
	if len(podList.Items) > 1 {
		return nil, fmt.Errorf("multiple pods matching %s returned - critical failure", podName)
	}

	// Return the first element from the list
	return &podList.Items[0], err
}

func (b *ExecAccessBuilder) GetTargetPodName() (string, error) {
	logger := ctrllog.FromContext(b.Ctx)

	// If the user supplied their own Pod, then get that Pod back to make sure it exists. Otherwise,
	// randomly select a pod.
	if b.Request.Spec.TargetPod == "" {
		pod, err := b.GetRandomPod()
		if err != nil {
			logger.Error(err, "Failed to retrieve Pod from ExecAccessTemplate")
			return "", err
		}
		return pod.Name, nil
	} else {
		pod, err := b.GetSpecificPod()

		// Informative for the operator for now. The verification step below truly let the user know about the problem.
		if err != nil {
			logger.Info("Error looking up Pod")
			return "", err
		}

		return pod.Name, nil
	}

}

func (b *ExecAccessBuilder) GenerateAccessRole() (*rbacv1.Role, error) {
	role := &rbacv1.Role{}

	role.Name = fmt.Sprintf("%s-%s", b.Request.Name, b.Request.GetUniqueId())
	role.Namespace = b.Template.Namespace
	role.Rules = []rbacv1.PolicyRule{
		{
			APIGroups:     []string{corev1.GroupName},
			Resources:     []string{"pods"},
			ResourceNames: []string{b.Request.Status.PodName},
			Verbs:         []string{"get", "list", "watch"},
		},
		{
			APIGroups:     []string{corev1.GroupName},
			Resources:     []string{"pods/exec"},
			ResourceNames: []string{b.Request.Status.PodName},
			Verbs:         []string{"create", "update", "delete", "get", "list"},
		},
	}

	// Set the ownerRef for the Deployment
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/owners-dependents/
	if err := ctrl.SetControllerReference(b.Request, role, b.Scheme); err != nil {
		return nil, err
	}

	return role, nil
}

func (b *ExecAccessBuilder) GenerateAccessRoleBinding() (*rbacv1.RoleBinding, error) {
	rb := &rbacv1.RoleBinding{}

	rb.Name = fmt.Sprintf("%s-%s", b.Request.Name, b.Request.GetUniqueId())
	rb.Namespace = b.Template.Namespace
	rb.RoleRef = rbacv1.RoleRef{
		APIGroup: rbacv1.GroupName,
		Kind:     "Role",
		Name:     rb.Name,
	}
	rb.Subjects = []rbacv1.Subject{}

	for _, group := range b.Template.Spec.AllowedGroups {
		rb.Subjects = append(rb.Subjects, rbacv1.Subject{
			APIGroup: rbacv1.SchemeGroupVersion.Group,
			Kind:     rbacv1.GroupKind,
			Name:     group,
		})
	}

	// Set the ownerRef for the Deployment
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/owners-dependents/
	if err := ctrl.SetControllerReference(b.Request, rb, b.Scheme); err != nil {
		return nil, err
	}

	return rb, nil
}
