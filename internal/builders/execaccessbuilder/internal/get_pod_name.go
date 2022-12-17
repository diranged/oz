package internal

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/diranged/oz/internal/api/v1alpha1"
)

// GetPodName is used to discover the target pod that the user is going to have access to. This
// function is designed to be idempotent - so once a podName has been selected, it will be used on
// each and every reconcile going forward.
//
//   - If status.podName is set? Return that value Else? Continue.
//   - If request.targetPod...
//     ... is set, call getSpecificPod() to verify that the pod exists and is valid for the request
//     ... is not set, call getRandomPod() to pick a random pod from the target controller
//   - Save the picked podName into the request status and update the request object
//
// Returns:
//
//	podname: A string with the pod name (or an empty string in a failure)
//	error: Any errors generating the podName.
func GetPodName(
	ctx context.Context,
	client client.Client,
	req *v1alpha1.ExecAccessRequest,
	tmpl *v1alpha1.ExecAccessTemplate,
) (podName string, err error) {
	log := logf.FromContext(ctx)
	var pod *corev1.Pod

	// If this resource already has a status.podName field set, then we respect
	// that no matter what. We never mutate the pod that this access request
	// was originally created for. Otherwise, pick a Pod and populate that
	// status field.
	if req.GetPodName() != "" {
		log.Info(fmt.Sprintf("Pod already assigned - %s", req.GetPodName()))
		return req.GetPodName(), nil
	}

	// If the user supplied their own Pod, then get that Pod back to make sure
	// it exists. Otherwise, randomly select a pod.
	switch req.Spec.TargetPod {
	case "":
		pod, err = getRandomPod(ctx, client, tmpl)
		if err != nil {
			log.Error(err, "Failed to retrieve Pod from ExecAccessTemplate")
			return "", err
		}
	default:
		pod, err = getSpecificPod(ctx, client, req.Spec.TargetPod, tmpl)

		// Informative for the operator for now. The verification step below
		// truly let the user know about the problem.
		if err != nil {
			log.Info("Error looking up Pod")
			return "", err
		}
	}

	// Set the podName (note, just in the local object). If this fails (for
	// example, its already set on the object), then we also bail out. This
	// only fails if the Status.PodName field has already been set, which would
	// indicate some kind of a reconcile loop conflict.
	//
	// The responsibility of pushing the .Status.PodName field back to
	// Kubernetes is in the reconciliation loop, where it will call
	// UpdateCondition (which calls UpdateStatus) at the end of this succesful
	// method. In this way, we do not update the AccessRequest with a PodName
	// status until we have confidence that all of the access resources have
	// indeed been created.
	if err := req.SetPodName(pod.GetName()); err != nil {
		return "", err
	}

	// Return the podName string.
	return pod.Name, nil
}
