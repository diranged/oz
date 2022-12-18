package podaccessbuilder

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/diranged/oz/internal/api/v1alpha1"
)

// AccessResourcesAreReady implements the IBuilder interface by checking for
// the current state of the Pod for the user and returning True when it is
// ready, or False if it is not ready after a specified timeout.
//
// TODO: Implement a per-pod-access-template setting to tune this timeout.
func (b *PodAccessBuilder) AccessResourcesAreReady(
	ctx context.Context,
	client client.Client,
	req v1alpha1.IRequestResource,
	_ v1alpha1.ITemplateResource,
) (bool, error) {
	log := logf.FromContext(ctx).WithName("AccessResourcesAreReady")

	// Cast the Request into an PodAccessRequest.
	podReq := req.(*v1alpha1.PodAccessRequest)

	// First, verify whether or not the PodName field has been set. If not,
	// then some part of the reconciliation has previously failed.
	if podReq.GetPodName() == "" {
		return false, errors.New("status.podName not yet set")
	}

	// This empty Pod struct will be filled in by the isPodReady() function.
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podReq.GetPodName(),
			Namespace: podReq.GetNamespace(),
		},
	}

	log.Info(
		fmt.Sprintf(
			"Checking if pod %s is ready yet (timeout: %s)",
			pod.GetName(),
			defaultReadyWaitTime,
		),
	)

	// Store the ready/err states outside of the loop so that we can return
	// them at the end of the method.
	var ready bool
	var err error

	// In a loop, keep checking the Pod state. When it's ready, return. In an
	// error, just keep looping. After the timeout has occurrred, we simply
	// return the last known state.
	for stay, timeout := true, time.After(defaultReadyWaitTime); stay; {
		select {
		case <-timeout:
			log.Info(fmt.Sprintf("Timeout waiting for %s to become ready.", pod.GetName()))
			stay = false
		default:
			if ready, err = isPodReady(ctx, client, log, pod); err != nil {
				if apierrors.IsNotFound(err) {
					// Immediately bail out and let a requeue event happen
					return false, err
				}
				// For any other error, consider it transient and try again
				log.Error(err, "Error getting Pod status (will retry)")
			} else if ready {
				// return ready = true, and no error
				log.Info("Pod ready state", "phase", pod.Status.Phase)
				return ready, nil
			}
		}

		// Wait 1 second before trying again
		log.V(1).Info("Sleeping and trying again")
		time.Sleep(defaultReadyWaitInterval)
	}

	return ready, nil
}

func isPodReady(
	ctx context.Context,
	client client.Client,
	log logr.Logger,
	pod *corev1.Pod,
) (bool, error) {
	// Next, get the Pod. If the pod-get fails, then we need to return that failure.
	log.V(2).Info("Getting pod")
	err := client.Get(ctx, types.NamespacedName{
		Name:      pod.GetName(),
		Namespace: pod.GetNamespace(),
	}, pod)
	if err != nil {
		return false, err
	}

	// Now, check the Pod Phase first... the pod could be Pending and not yet
	// ready to even have a condition state.
	if pod.Status.Phase != PodPhaseRunning {
		log.V(2).Info(fmt.Sprintf("Pod Phase is %s, not %s", pod.Status.Phase, PodPhaseRunning))
		return false, nil
	}

	// Iterate through the PodConditions looking for the PodReady condition.
	// When we find it, return whether it's "True" or "False".
	conditions := pod.Status.Conditions
	for _, condition := range conditions {
		if condition.Type == corev1.PodReady {
			// val = condition.Status == corev1.ConditionTrue
			log.V(2).
				Info(fmt.Sprintf("Got to the inner condition... returning %s", condition.Status))
			return condition.Status == corev1.ConditionTrue, nil
		}
	}

	// Return ready=false at this point
	log.V(2).Info("Pod Ready Condition not yet True")
	return false, nil
}
