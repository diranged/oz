package builders

import (
	"context"

	"github.com/diranged/oz/interfaces"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AccessBuilder struct {
	Client client.Client
	Ctx    context.Context
	Scheme *runtime.Scheme

	Request  interfaces.OzRequestResource
	Template interfaces.OzTemplateResource
}

func (t *AccessBuilder) GetTargetResource() (client.Object, error) {
	// https://blog.gripdev.xyz/2020/07/20/k8s-operator-with-dynamic-crds-using-controller-runtime-no-structs/
	obj := t.Template.GetTemplateTarget().GetObject()
	err := t.Client.Get(t.Ctx, types.NamespacedName{
		Name:      t.Template.GetTemplateTarget().GetName(),
		Namespace: t.Template.GetNamespace(),
	}, obj)
	return obj, err
}
