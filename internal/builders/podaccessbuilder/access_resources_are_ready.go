package podaccessbuilder

import (
	"context"
	"errors"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/diranged/oz/internal/api/v1alpha1"
)

// AccessResourcesAreReady implements the IBuilder interface
//
// TODO: Implement a waiter loop
func (b *PodAccessBuilder) AccessResourcesAreReady(
	ctx context.Context,
	client client.Client,
	req v1alpha1.IRequestResource,
	_ v1alpha1.ITemplateResource,
) (bool, error) {
	// Cast the Request into an PodAccessRequest.
	podReq := req.(*v1alpha1.PodAccessRequest)

	// First, verify whether or not the PodName field has been set. If not,
	// then some part of the reconciliation has previously failed.
	if podReq.GetPodName() == "" {
		return false, errors.New("status.podName not yet set")
	}

	// Next, get the Pod. If the pod-get fails, then we need to return that failure.
	pod := &corev1.Pod{}
	err := client.Get(ctx, types.NamespacedName{
		Name:      podReq.GetPodName(),
		Namespace: podReq.GetNamespace(),
	}, pod)
	if err != nil {
		return false, err
	}

	// Now, check the Pod ready status
	return pod.Status.Phase == PodPhaseRunning, nil
}
