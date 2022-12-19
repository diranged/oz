package utils

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/diranged/oz/internal/api/v1alpha1"
)

// CreatePod creates a new Pod based on the supplied PodTemplateSpec, ensuring
// that the OwnerReference is set appropriately before the creation to
// guarantee proper cleanup.
func CreatePod(
	ctx context.Context,
	client client.Client,
	req v1alpha1.IRequestResource,
	podTemplateSpec corev1.PodTemplateSpec,
) (*corev1.Pod, error) {
	logger := logf.FromContext(ctx)

	// We'll populate this pod object
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GenerateResourceName(req),
			Namespace: req.GetNamespace(),
		},
	}

	// Verify first whether or not a pod already exists with this name. If it
	// does, we just return it back. The issue here is that updating a Pod is
	// an unusual thing to do once it's alive, and can cause race condition
	// issues if you do not do the updates properly.
	//
	// https://github.com/diranged/oz/issues/27
	err := client.Get(ctx, types.NamespacedName{
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
	if err := ctrlutil.SetControllerReference(req, pod, client.Scheme()); err != nil {
		return nil, err
	}

	// https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/controller/controllerutil#CreateOrUpdate
	//
	// In an update event, wec an only update the Annotations and the OwnerReference. Nothing else.
	logger.V(1).Info(fmt.Sprintf("Creating Pod %s (ns: %s)", pod.Name, pod.Namespace))
	logger.V(5).Info("Pod Json", "json", ObjectToJSON(pod))
	if err := client.Create(ctx, pod); err != nil {
		return nil, err
	}

	return pod, nil
}
