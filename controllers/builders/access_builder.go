package builders

import (
	"context"

	"github.com/diranged/oz/interfaces"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Builder interface {
	GetClient() client.Client
	GetCtx() context.Context
	GetScheme() *runtime.Scheme

	GetRequest() interfaces.OzRequestResource
	GetTemplate() interfaces.OzTemplateResource

	// Returns back the PodName that the user is being granted direct access to.
	GeneratePodName() (string, error)
}

type BaseBuilder struct {
	Builder

	Client client.Client
	Ctx    context.Context
	Scheme *runtime.Scheme

	Request  interfaces.OzRequestResource
	Template interfaces.OzTemplateResource
}

func (t *BaseBuilder) GetClient() client.Client {
	return t.Client
}

func (t *BaseBuilder) GetCtx() context.Context {
	return t.Ctx
}

func (t *BaseBuilder) GetScheme() *runtime.Scheme {
	return t.Scheme
}

func (t *BaseBuilder) GetTemplate() interfaces.OzTemplateResource {
	return t.Template
}

func (t *BaseBuilder) GetRequest() interfaces.OzRequestResource {
	return t.Request
}

func (t *BaseBuilder) GetTargetRefResource() (client.Object, error) {
	// https://blog.gripdev.xyz/2020/07/20/k8s-operator-with-dynamic-crds-using-controller-runtime-no-structs/
	obj := t.Template.GetTargetRef().GetObject()
	err := t.Client.Get(t.Ctx, types.NamespacedName{
		Name:      t.Template.GetTargetRef().GetName(),
		Namespace: t.Template.GetNamespace(),
	}, obj)
	return obj, err
}

type AccessBuilder struct {
	*BaseBuilder
}

// TODO: GeneratePodName needs to figure out the PodName after it has created the target pod in the first place? Or
// it could just generate a static name with a clean function and return that.
func (t *AccessBuilder) GeneratePodName() (string, error) {
	return "junk", nil
}
