package bldutil

import (
	"context"
	"errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/diranged/oz/internal/api/v1alpha1"
)

// GetSelectorLabels understands how to return a labels.Selector struct from
// a supplied controller object - as long as it is one of the following:
//
//   - Deployment
//   - DaemonSet
//   - StatefulSet
//   - Rollout
//
// https://medium.com/coding-kubernetes/using-k8s-label-selectors-in-go-the-right-way-733cde7e8630
//
// Returns:
//
//   - labels.Selector: A populated labels.Selector which can be used when searching for Pods
//   - error
//
// revive:disable:cyclomatic
func GetSelectorLabels(
	ctx context.Context,
	client client.Client,
	tmpl v1alpha1.ITemplateResource,
) (labels.Selector, error) {
	// https://sdk.operatorframework.io/docs/building-operators/golang/references/logging/
	log := logf.FromContext(ctx)

	targetController, err := GetTargetRefResource(ctx, client, tmpl)
	if err != nil {
		return nil, err
	}

	// TODO: Figure out a more generic way to do this that doesn't involve a bunch of checks like this
	switch kind := targetController.GetObjectKind().GroupVersionKind().Kind; kind {
	case "Deployment":
		controller, err := getDeployment(ctx, client, targetController)
		if err != nil {
			log.Error(err, "Failed to find target Deployment")
			return nil, err
		}
		return metav1.LabelSelectorAsSelector(controller.Spec.Selector)

	case "Rollout":
		controller, err := getRollout(ctx, client, targetController)
		if err != nil {
			log.Error(err, "Failed to find target Rollout")
			return nil, err
		}
		return metav1.LabelSelectorAsSelector(controller.Spec.Selector)

	case "DaemonSet":
		controller, err := getDaemonSet(ctx, client, targetController)
		if err != nil {
			log.Error(err, "Failed to find target DaemonSet")
			return nil, err
		}
		return metav1.LabelSelectorAsSelector(controller.Spec.Selector)

	case "StatefulSet":
		controller, err := getStatefulSet(ctx, client, targetController)
		if err != nil {
			log.Error(err, "Failed to find target StatefulSet")
			return nil, err
		}
		return metav1.LabelSelectorAsSelector(controller.Spec.Selector)

	default:
		return nil, errors.New("invalid input")
	}
}
