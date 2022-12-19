// Package legacybuilder provides a set of Access Builder structs and methods for dynamically
// generating Kubernetes resources for a particular type of Access Request.
package legacybuilder

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	api "github.com/diranged/oz/internal/api/v1alpha1"
)

const shortUIDLength = 8

// BaseBuilder provides a starting point struct with a set of common methods. These methods are used
// by template specific builders to reduce the amount of code we re-write.
type BaseBuilder struct {
	IBuilder

	Client client.Client
	Ctx    context.Context

	// APIReader should be generated with mgr.GetAPIReader() to create a non-cached client object. This is used
	// for certain Get() calls where we need to ensure we are getting the latest version from the API, and not a cached
	// object.
	//
	// See https://github.com/kubernetes-sigs/controller-runtime/issues/585#issuecomment-528102351
	//
	APIReader client.Reader

	// Generic struct that satisfies the OzRequestResource interface. This is used for the common
	// functions inside the BaseBuilder struct.
	Request api.IPodRequestResource

	// Generic struct that satisfies the OzTemplateREsource interface. This is used for the common
	// functions inside the BaseBuilder struct.
	Template api.ITemplateResource
}

// GetClient provides an access method for the cached and default client.Client resource from the
// reconciliation loop.
//
// Returns:
//
//	client.Client: The default controller-runtime cached Client struct.
func (b *BaseBuilder) GetClient() client.Client {
	return b.Client
}

// GetCtx provides an access method for the context.Context resource from the reconciliation loop.
//
// Returns:
//
//	context.Context: The default controller-runtime context.Context struct.
func (b *BaseBuilder) GetCtx() context.Context {
	return b.Ctx
}

// GetScheme provides an access method for the runtime.Schema pointer from the reconciliation loop.
//
// Returns:
//
//	*runtime.Scheme: A pointer back to the runtime.Scheme from the controller-runtime struct.
func (b *BaseBuilder) GetScheme() *runtime.Scheme {
	return b.Client.Scheme()
}

// GetTemplate provides an access method to the generic api.OzTemplateResource interface
// which is used to access common methods that each Access Template must expose.
//
// Returns:
//
//	api.OzTemplateResource
func (b *BaseBuilder) GetTemplate() api.ITemplateResource {
	return b.Template
}

// GetRequest provides an access method to the generic api.OzRequestResource interface
// which is used to access common methods that each Access Request must expose.
//
// Returns:
//
//	api.OzRequestResource
func (b *BaseBuilder) GetRequest() api.IPodRequestResource {
	return b.Request
}

// GetTargetRefResource returns a generic client.Object resource from the Kubernetes API that points
// to the Access Template Spec.targetRef configured resource. This generic function allows us (in
// the future) to have AccessTemplates understand how to point to all kinds of different Pods via
// different controllers.
//
// Returns:
//
//	client.Object: An unstructured.Unstructured{} object pointing to the target controller.
func (b *BaseBuilder) GetTargetRefResource() (client.Object, error) {
	// https://blog.gripdev.xyz/2020/07/20/k8s-operator-with-dynamic-crds-using-controller-runtime-no-structs/
	obj := b.Template.GetTargetRef().GetObject()
	err := b.Client.Get(b.Ctx, types.NamespacedName{
		Name:      b.Template.GetTargetRef().GetName(),
		Namespace: b.Template.GetNamespace(),
	}, obj)
	return obj, err
}

func (b *BaseBuilder) getPodTemplateFromController() (corev1.PodTemplateSpec, error) {
	// https://sdk.operatorframework.io/docs/building-operators/golang/references/logging/
	logger := log.FromContext(b.Ctx)

	targetController, err := b.GetTargetRefResource()
	if err != nil {
		return corev1.PodTemplateSpec{}, err
	}

	// TODO: Figure out a more generic way to do this that doesn't involve a bunch of checks like this
	switch kind := targetController.GetObjectKind().GroupVersionKind().Kind; kind {
	case "Deployment":
		controller, err := b.getDeployment(targetController)
		if err != nil {
			logger.Error(err, "Failed to find target Deployment")
			return corev1.PodTemplateSpec{}, err
		}
		return *controller.Spec.Template.DeepCopy(), nil

	case "DaemonSet":
		controller, err := b.getDaemonSet(targetController)
		if err != nil {
			logger.Error(err, "Failed to find target DaemonSet")
			return corev1.PodTemplateSpec{}, err
		}
		return *controller.Spec.Template.DeepCopy(), nil

	case "StatefulSet":
		controller, err := b.getStatefulSet(targetController)
		if err != nil {
			logger.Error(err, "Failed to find target StatefulSet")
			return corev1.PodTemplateSpec{}, err
		}
		return *controller.Spec.Template.DeepCopy(), nil

	default:
		return corev1.PodTemplateSpec{}, errors.New("invalid input")
	}
}

// getDeployment returns a Deployment given the supplied generic client.Object resource
//
// Returns:
//
//	appsv1.Deployment: A populated deployment object
//	error: Any error that may have occurred
func (b *BaseBuilder) getDeployment(obj client.Object) (*appsv1.Deployment, error) {
	found := &appsv1.Deployment{}
	err := b.Client.Get(b.Ctx, types.NamespacedName{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}, found)
	return found, err
}

// getDaemonSet returns a DaemonSet given the supplied generic client.Object resource
//
// Returns:
//
//	appsv1.DaemonSet: A populated deployment object
//	error: Any error that may have occurred
func (b *BaseBuilder) getDaemonSet(obj client.Object) (*appsv1.DaemonSet, error) {
	found := &appsv1.DaemonSet{}
	err := b.Client.Get(b.Ctx, types.NamespacedName{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}, found)
	return found, err
}

// getStatefulSet returns a StatefulSet given the supplied generic client.Object resource
//
// Returns:
//
//	appsv1.StatefulSet: A populated deployment object
//	error: Any error that may have occurred
func (b *BaseBuilder) getStatefulSet(obj client.Object) (*appsv1.StatefulSet, error) {
	found := &appsv1.StatefulSet{}
	err := b.Client.Get(b.Ctx, types.NamespacedName{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}, found)
	return found, err
}

func (b *BaseBuilder) createAccessRole(podName string) (*rbacv1.Role, error) {
	role := &rbacv1.Role{}

	role.Name = generateResourceName(b.Request)
	role.Namespace = b.Template.GetNamespace()
	role.Rules = []rbacv1.PolicyRule{
		{
			APIGroups:     []string{corev1.GroupName},
			Resources:     []string{"pods"},
			ResourceNames: []string{podName},
			Verbs:         []string{"get", "list", "watch"},
		},
		{
			APIGroups:     []string{corev1.GroupName},
			Resources:     []string{"pods/exec"},
			ResourceNames: []string{podName},
			Verbs:         []string{"create", "update", "delete", "get", "list"},
		},
	}

	// Set the ownerRef for the Deployment
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/owners-dependents/
	if err := ctrlutil.SetControllerReference(b.Request, role, b.GetScheme()); err != nil {
		return nil, err
	}

	// Generate an empty role resource. This role resource will be filled-in by the CreateOrUpdate() call when
	// it checks the Kubernetes API for the existing role. Our update function will then update the appropriate
	// values from the desired role object above.
	emptyRole := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{Name: role.Name, Namespace: role.Namespace},
	}

	// https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/controller/controllerutil#CreateOrUpdate
	if _, err := ctrlutil.CreateOrUpdate(b.Ctx, b.Client, emptyRole, func() error {
		emptyRole.ObjectMeta = role.ObjectMeta
		emptyRole.Rules = role.Rules
		emptyRole.OwnerReferences = role.OwnerReferences
		return nil
	}); err != nil {
		return nil, err
	}

	return role, nil
}

func (b *BaseBuilder) createAccessRoleBinding() (*rbacv1.RoleBinding, error) {
	rb := &rbacv1.RoleBinding{}

	rb.Name = generateResourceName(b.Request)
	rb.Namespace = b.Template.GetNamespace()
	rb.RoleRef = rbacv1.RoleRef{
		APIGroup: rbacv1.GroupName,
		Kind:     "Role",
		Name:     rb.Name,
	}
	rb.Subjects = []rbacv1.Subject{}

	for _, group := range b.Template.GetAccessConfig().GetAllowedGroups() {
		rb.Subjects = append(rb.Subjects, rbacv1.Subject{
			APIGroup: rbacv1.SchemeGroupVersion.Group,
			Kind:     rbacv1.GroupKind,
			Name:     group,
		})
	}

	// Set the ownerRef for the Deployment
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/owners-dependents/
	if err := ctrlutil.SetControllerReference(b.Request, rb, b.GetScheme()); err != nil {
		return nil, err
	}

	// Generate an empty role resource. This role resource will be filled-in by the CreateOrUpdate() call when
	// it checks the Kubernetes API for the existing role. Our update function will then update the appropriate
	// values from the desired role object above.
	emptyRb := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: rb.Name, Namespace: rb.Namespace},
	}

	// https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/controller/controllerutil#CreateOrUpdate
	if _, err := ctrlutil.CreateOrUpdate(b.Ctx, b.Client, emptyRb, func() error {
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

func (b *BaseBuilder) createPod(podTemplateSpec corev1.PodTemplateSpec) (*corev1.Pod, error) {
	logger := log.FromContext(b.Ctx)

	// We'll populate this pod object
	pod := &corev1.Pod{}
	pod.Name = generateResourceName(b.Request)
	pod.Namespace = b.Template.GetNamespace()

	// Verify first whether or not a pod already exists with this name. If it
	// does, we just return it back. The issue here is that updating a Pod is
	// an unusual thing to do once it's alive, and can cause race condition
	// issues if you do not do the updates properly.
	//
	// https://github.com/diranged/oz/issues/27
	err := b.Client.Get(b.Ctx, types.NamespacedName{
		Name:      pod.Name,
		Namespace: pod.Namespace,
	}, pod)

	// If there was no error on this get, then the object already exists in K8S
	// and we need to just return that.
	if err == nil {
		return pod, err
	}

	// Finish filling out the desired PodSpec at this point.
	pod.Spec = *podTemplateSpec.Spec.DeepCopy()
	pod.ObjectMeta.Annotations = podTemplateSpec.ObjectMeta.Annotations
	pod.ObjectMeta.Labels = podTemplateSpec.ObjectMeta.Labels

	// Set the ownerRef for the Deployment
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/owners-dependents/
	if err := ctrlutil.SetControllerReference(b.Request, pod, b.GetScheme()); err != nil {
		return nil, err
	}

	// https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/controller/controllerutil#CreateOrUpdate
	//
	// In an update event, wec an only update the Annotations and the OwnerReference. Nothing else.
	logger.V(1).Info(fmt.Sprintf("Creating or Updating Pod %s (ns: %s)", pod.Name, pod.Namespace))
	logger.V(1).Info("Pod Json", "json", ObjectToJSON(pod))
	if err := b.Client.Create(b.Ctx, pod); err != nil {
		return nil, err
	}

	return pod, nil
}

// getShortUID returns back a shortened version of the UID that the Kubernetes cluster used to store
// the AccessRequest internally. This is used by the Builders to create unique names for the
// resources they manage (Roles, RoleBindings, etc).
//
// Returns:
//
//	shortUID: A 10-digit long shortened UID
func getShortUID(obj client.Object) string {
	// TODO: If the UID isn't there, we should generate something random OR throw an error.
	return string(obj.GetUID())[0:shortUIDLength]
}

// generateResourceName takes in an API.IRequestResource conforming object and returns a unique
// resource name string that can be used to safely create other resources (roles, bindings, etc).
//
// Returns:
//
//	string: A resource name string
func generateResourceName(req api.IRequestResource) string {
	return fmt.Sprintf("%s-%s", req.GetName(), getShortUID(req))
}

// ObjectToJSON is a quick helper function for pretty-printing an entire K8S object in JSON form.
// Used in certain debug log statements primarily.
func ObjectToJSON(obj client.Object) string {
	jsonData, err := json.Marshal(obj)
	if err != nil {
		fmt.Printf("could not marshal json: %s\n", err)
		return ""
	}
	return string(jsonData)
}
