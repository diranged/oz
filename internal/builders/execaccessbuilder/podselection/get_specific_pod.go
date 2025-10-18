package podselection

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/diranged/oz/internal/api/v1alpha1"
	"github.com/diranged/oz/internal/builders/utils"
)

func getSpecificPod(
	ctx context.Context,
	cl client.Client,
	podName string,
	tmpl *v1alpha1.ExecAccessTemplate,
) (*corev1.Pod, error) {
	log := logf.FromContext(ctx)
	log.Info(fmt.Sprintf("Looking for Pod %s", podName))

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
		client.InNamespace(tmpl.GetNamespace()),
		client.MatchingLabelsSelector{
			Selector: selector,
		},
		client.MatchingFields{
			v1alpha1.FieldSelectorMetadataName: podName,
			v1alpha1.FieldSelectorStatusPhase:  string(PodPhaseRunning),
		},
	}
	if err := cl.List(ctx, podList, opts...); err != nil {
		log.Error(err, "Failed to retrieve Pod list")
		return nil, err
	}
	if len(podList.Items) < 1 {
		return nil, fmt.Errorf("pod named %s not found", podName)
	}
	if len(podList.Items) > 1 {
		return nil, fmt.Errorf("multiple pods matching %s returned - critical failure", podName)
	}

	// Return the first element from the list
	return &podList.Items[0], err
}
