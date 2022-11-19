// Package builders provides a set of Access Builder structs and methods for dynamically
// generating Kubernetes resources for a particular type of Access Request.
package builders

import (
	"context"
	"errors"
	"fmt"

	"github.com/diranged/oz/interfaces"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Builder defines the interface for a particular "access builder". An "access builder" is typically
// paired with an "access template" struct in the api.v1alpha1 package. Each unique type of access
// template will have its own access builder that is used to implement the goals of that particular
// template.
//
// Common interface functions are used to keep the reconiliation loop code in the individual
// controllers package clean.
type Builder interface {
	GetClient() client.Client
	GetCtx() context.Context
	GetScheme() *runtime.Scheme

	GetRequest() interfaces.OzRequestResource
	GetTemplate() interfaces.OzTemplateResource

	// Returns back the PodName that the user is being granted direct access to.
	GeneratePodName() (podName string, err error)

	// Generates all of the resources required to fulfill the access request.
	GenerateAccessResources() (statusString string, accessString string, err error)
}

// BaseBuilder provides a starting point struct with a set of common methods. These methods are used
// by template specific builders to reduce the amount of code we re-write.
type BaseBuilder struct {
	Builder

	Client client.Client
	Ctx    context.Context
	Scheme *runtime.Scheme

	// APIReader should be generated with mgr.GetAPIReader() to create a non-cached client object. This is used
	// for certain Get() calls where we need to ensure we are getting the latest version from the API, and not a cached
	// object.
	//
	// See https://github.com/kubernetes-sigs/controller-runtime/issues/585#issuecomment-528102351
	//
	APIReader client.Reader

	// Generic struct that satisfies the OzRequestResource interface. This is used for the common
	// functions inside the BaseBuilder struct.
	Request interfaces.OzRequestResource

	// Generic struct that satisfies the OzTemplateREsource interface. This is used for the common
	// functions inside the BaseBuilder struct.
	Template interfaces.OzTemplateResource
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
	return b.Scheme
}

// GetTemplate provides an access method to the generic interfaces.OzTemplateResource interface
// which is used to access common methods that each Access Template must expose.
//
// Returns:
//
//	interfaces.OzTemplateResource
func (b *BaseBuilder) GetTemplate() interfaces.OzTemplateResource {
	return b.Template
}

// GetRequest provides an access method to the generic interfaces.OzRequestResource interface
// which is used to access common methods that each Access Request must expose.
//
// Returns:
//
//	interfaces.OzRequestResource
func (b *BaseBuilder) GetRequest() interfaces.OzRequestResource {
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

// VerifyPodExists may be deleted soon
func (b *BaseBuilder) VerifyPodExists(name string, namespace string) error {
	logger := log.FromContext(b.Ctx)
	logger.Info(fmt.Sprintf("Verifying that Pod %s still exists...", name))

	// Search for the Pod
	pod := &corev1.Pod{}
	err := b.Client.Get(b.Ctx, types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, pod)

	// On any failure, update the pod status with the failure...
	if err != nil {
		return fmt.Errorf("pod %s (ns: %s) is not found: %s", name, namespace, err)
	}
	return nil
}

// getTargetPodSelectorLabels understands how to return a labels.Selector struct from
// a supplied controller object - as long as it is one of the following:
//
//   - Deployment
//   - DaemonSet
//   - StatefulSet
//
// https://medium.com/coding-kubernetes/using-k8s-label-selectors-in-go-the-right-way-733cde7e8630
//
// Returns:
//
//   - labels.Selector: A populated labels.Selector which can be used when searching for Pods
//   - error
func (b *BaseBuilder) getTargetPodSelectorLabels() (labels.Selector, error) {
	// https://sdk.operatorframework.io/docs/building-operators/golang/references/logging/
	logger := log.FromContext(b.Ctx)

	targetController, err := b.GetTargetRefResource()
	if err != nil {
		return nil, err
	}

	// TODO: Figure out a more generic way to do this that doesn't involve a bunch of checks like this
	switch kind := targetController.GetObjectKind().GroupVersionKind().Kind; kind {
	case "Deployment":
		controller, err := b.getDeployment(targetController)
		if err != nil {
			logger.Error(err, "Failed to find target Deployment")
			return nil, err
		}
		return metav1.LabelSelectorAsSelector(controller.Spec.Selector)

	case "DaemonSet":
		controller, err := b.getDaemonSet(targetController)
		if err != nil {
			logger.Error(err, "Failed to find target DaemonSet")
			return nil, err
		}
		return metav1.LabelSelectorAsSelector(controller.Spec.Selector)

	case "StatefulSet":
		controller, err := b.getStatefulSet(targetController)
		if err != nil {
			logger.Error(err, "Failed to find target StatefulSet")
			return nil, err
		}
		return metav1.LabelSelectorAsSelector(controller.Spec.Selector)

	default:
		return nil, errors.New("invalid input")
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
