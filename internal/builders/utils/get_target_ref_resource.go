package utils

import (
	"context"

	"github.com/diranged/oz/internal/api/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetTargetRefResource returns a generic client.Object resource from the Kubernetes API that points
// to the Access Template Spec.targetRef configured resource. This generic function allows us (in
// the future) to have AccessTemplates understand how to point to all kinds of different Pods via
// different controllers.
//
// Returns:
//
//	client.Object: An unstructured.Unstructured{} object pointing to the target controller.
func GetTargetRefResource(
	ctx context.Context,
	client client.Client,
	tmpl v1alpha1.ITemplateResource,
) (client.Object, error) {
	// https://blog.gripdev.xyz/2020/07/20/k8s-operator-with-dynamic-crds-using-controller-runtime-no-structs/
	obj := tmpl.GetTargetRef().GetObject()
	err := client.Get(ctx, types.NamespacedName{
		Name:      tmpl.GetTargetRef().GetName(),
		Namespace: tmpl.GetNamespace(),
	}, obj)
	return obj, err
}
