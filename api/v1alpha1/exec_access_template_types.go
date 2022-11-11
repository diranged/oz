/*
Copyright 2022 Matt Wise.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"context"
	"errors"
	"fmt"
	"math/rand"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// ExecAccessTemplateSpec defines the desired state of ExecAccessTemplate
type ExecAccessTemplateSpec struct {
	// TargetRef provides a pattern for referencing objects from another API in a generic way.
	// +kubebuilder:validation:Required
	TargetRef CrossVersionObjectReference `json:"targetRef"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ExecAccessTemplate is the Schema for the execaccesstemplates API
type ExecAccessTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ExecAccessTemplateSpec `json:"spec,omitempty"`
	Status BaseTemplateStatus     `json:"status,omitempty"`
}

func (t *ExecAccessTemplate) GetDeployment(cl client.Client, ctx context.Context) (*appsv1.Deployment, error) {
	found := &appsv1.Deployment{}

	err := cl.Get(ctx, types.NamespacedName{
		Name:      *t.Spec.TargetRef.Name,
		Namespace: t.Namespace,
	}, found)
	return found, err
}

func (t *ExecAccessTemplate) GetDaemonSet(cl client.Client, ctx context.Context) (*appsv1.DaemonSet, error) {
	found := &appsv1.DaemonSet{}
	err := cl.Get(ctx, types.NamespacedName{
		Name:      *t.Spec.TargetRef.Name,
		Namespace: t.Namespace,
	}, found)
	return found, err
}

func (t *ExecAccessTemplate) GetStatefulSet(cl client.Client, ctx context.Context) (*appsv1.StatefulSet, error) {
	found := &appsv1.StatefulSet{}
	err := cl.Get(ctx, types.NamespacedName{
		Name:      *t.Spec.TargetRef.Name,
		Namespace: t.Namespace,
	}, found)
	return found, err
}

// https://medium.com/coding-kubernetes/using-k8s-label-selectors-in-go-the-right-way-733cde7e8630
func (t *ExecAccessTemplate) GetTargetPodSelectorLabels(cl client.Client, ctx context.Context) (labels.Selector, error) {
	// https://sdk.operatorframework.io/docs/building-operators/golang/references/logging/
	logger := ctrllog.FromContext(ctx)

	// Get the controller - if there's any error, return.
	if t.Spec.TargetRef.Kind == DeploymentController {
		controller, err := t.GetDeployment(cl, ctx)
		if err != nil {
			logger.Error(err, "Failed to find target Deployment")
			return nil, err
		}
		return metav1.LabelSelectorAsSelector(controller.Spec.Selector)
	} else if t.Spec.TargetRef.Kind == DaemonSetController {
		controller, err := t.GetDaemonSet(cl, ctx)
		if err != nil {
			logger.Error(err, "Failed to find target DaemonSet")
			return nil, err
		}
		return metav1.LabelSelectorAsSelector(controller.Spec.Selector)
	} else if t.Spec.TargetRef.Kind == StatefulSetController {
		controller, err := t.GetStatefulSet(cl, ctx)
		if err != nil {
			logger.Error(err, "Failed to find target StatefulSet")
			return nil, err
		}
		return metav1.LabelSelectorAsSelector(controller.Spec.Selector)
	}
	return nil, errors.New("Invalid input")
}

func (t *ExecAccessTemplate) GetRandomPod(cl client.Client, ctx context.Context) (*corev1.Pod, error) {
	// https://sdk.operatorframework.io/docs/building-operators/golang/references/logging/
	logger := ctrllog.FromContext(ctx)

	// Will populate this further down
	pod := &v1.Pod{}

	// Discover all running pods for the controller by polling for matching labels
	//logger.Info(fmt.Sprintf("Discovering Running Pods in %s %s...", controller.GetObjectKind(), controller.GetName()))

	// https://medium.com/coding-kubernetes/using-k8s-label-selectors-in-go-the-right-way-733cde7e8630
	selector, err := t.GetTargetPodSelectorLabels(cl, ctx)
	if err != nil {
		logger.Error(err, "Faild to find label selector, cannot automatically discover pods")
		return nil, err
	}

	// List all of the pods in the Deployment by searching for matching pods with the current Label
	// Selector.
	podList := &v1.PodList{}
	opts := []client.ListOption{
		client.InNamespace(t.Namespace),
		client.MatchingLabelsSelector{
			Selector: selector,
		},
		// TODO: Figure this out...
		//client.MatchingFields{"status.phase": "Running"},
	}
	if err := cl.List(ctx, podList, opts...); err != nil {
		logger.Error(err, "Failed to retrieve Pod list")
		return nil, err
	}

	// Randomly generate a number from within the length of the returned pod list...
	randomIndex := rand.Intn(len(podList.Items))

	// Return the randomly generated Pod
	logger.Info(fmt.Sprintf("Returning Pod %s", pod.Name))
	pod = &podList.Items[randomIndex]

	return pod, err
}

//+kubebuilder:object:root=true

// ExecAccessTemplateList contains a list of ExecAccessTemplate
type ExecAccessTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ExecAccessTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ExecAccessTemplate{}, &ExecAccessTemplateList{})
}
