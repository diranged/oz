// Package legacybuilder provides a set of Access Builder structs and methods for dynamically
// generating Kubernetes resources for a particular type of Access Request.
package legacybuilder

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	api "github.com/diranged/oz/internal/api/v1alpha1"
)

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
