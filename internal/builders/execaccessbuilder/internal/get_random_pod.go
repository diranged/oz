package internal

import (
	"context"
	"fmt"
	"math/rand"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/diranged/oz/internal/api/v1alpha1"
	"github.com/diranged/oz/internal/builders/utils"
)

func getRandomPod(
	ctx context.Context,
	cl client.Client,
	tmpl *v1alpha1.ExecAccessTemplate,
) (*corev1.Pod, error) {
	log := logf.FromContext(ctx)
	log.Info("Finding Pods...")

	// https://medium.com/coding-kubernetes/using-k8s-label-selectors-in-go-the-right-way-733cde7e8630
	selector, err := utils.GetSelectorLabels(ctx, cl, tmpl)
	if err != nil {
		log.Error(err, "Failed to find label selector, cannot automatically discover pods")
		return nil, err
	}

	// List all of the pods in the Deployment by searching for matching pods with the current Label
	// Selector.
	podList := &corev1.PodList{}
	opts := []client.ListOption{
		client.InNamespace(tmpl.Namespace),
		client.MatchingLabelsSelector{
			Selector: selector,
		},
		client.MatchingFields{
			v1alpha1.FieldSelectorStatusPhase: PodPhaseRunning,
		},
	}
	if err := cl.List(ctx, podList, opts...); err != nil {
		log.Error(err, "Failed to retrieve Pod list")
		return nil, err
	}

	if len(podList.Items) < 1 {
		return nil, fmt.Errorf("no pods found maching selector")
	}

	// Randomly generate a number from within the length of the returned pod list...
	randomIndex := rand.Intn(len(podList.Items))

	// Return the randomly generated Pod
	pod := &podList.Items[randomIndex]
	log.Info(fmt.Sprintf("Returning Pod %s", pod.Name))

	return pod, err
}
