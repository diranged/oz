package utils

import (
	"context"
	"errors"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/diranged/oz/internal/api/v1alpha1"
)

// GetPodTemplateFromController will return a PodTemplate resource from an
// understood controller type (Deployment, DaemonSet or StatefulSet).
func GetPodTemplateFromController(
	ctx context.Context,
	client client.Client,
	tmpl v1alpha1.ITemplateResource,
) (corev1.PodTemplateSpec, error) {
	// https://sdk.operatorframework.io/docs/building-operators/golang/references/logging/
	log := logf.FromContext(ctx)

	targetController, err := GetTargetRefResource(ctx, client, tmpl)
	if err != nil {
		return corev1.PodTemplateSpec{}, err
	}

	// TODO: Figure out a more generic way to do this that doesn't involve a bunch of checks like this
	switch kind := targetController.GetObjectKind().GroupVersionKind().Kind; kind {
	case "Deployment":
		controller, err := getDeployment(ctx, client, targetController)
		if err != nil {
			log.Error(err, "Failed to find target Deployment")
			return corev1.PodTemplateSpec{}, err
		}
		return *controller.Spec.Template.DeepCopy(), nil

	case "DaemonSet":
		controller, err := getDaemonSet(ctx, client, targetController)
		if err != nil {
			log.Error(err, "Failed to find target DaemonSet")
			return corev1.PodTemplateSpec{}, err
		}
		return *controller.Spec.Template.DeepCopy(), nil

	case "StatefulSet":
		controller, err := getStatefulSet(ctx, client, targetController)
		if err != nil {
			log.Error(err, "Failed to find target StatefulSet")
			return corev1.PodTemplateSpec{}, err
		}
		return *controller.Spec.Template.DeepCopy(), nil

	default:
		return corev1.PodTemplateSpec{}, errors.New("invalid input")
	}
}
