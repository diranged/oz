package legacybuilder

import (
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	api "github.com/diranged/oz/internal/api/v1alpha1"
)

// PodAccessBuilder implements the required resources for the api.AccessTemplate CRD.
//
// An "AccessRequest" is used to generate access that has been defined through an "AccessTemplate".
//
// An "AccessTemplate" defines a mode of access into a Pod by which a PodSpec is copied out of an
// existing Deployment (or StatefulSet, DaemonSet), mutated so that the Pod is not in the path of
// live traffic, and then Role and RoleBindings are created to grant the developer access into the
// Pod.
type PodAccessBuilder struct {
	BaseBuilder

	Request  *api.PodAccessRequest
	Template *api.PodAccessTemplate
}

// https://stackoverflow.com/questions/33089523/how-to-mark-golang-struct-as-implementing-interface
var (
	_ IBuilder = &PodAccessBuilder{}
	_ IBuilder = (*PodAccessBuilder)(nil)
)

// VerifyAccessResources verifies that the Pod created in the
// GenerateAccessResources() function is up and in the "Running" phase.
func (b *PodAccessBuilder) VerifyAccessResources() (statusString string, err error) {
	// First, verify whether or not the PodName field has been set. If not,
	// then some part of the reconciliation has previously failed.
	if b.Request.GetPodName() == "" {
		return "No Pod Assigned Yet", errors.New("status.podName not yet set")
	}

	// Next, get the Pod. If the pod-get fails, then we need to return that failure.
	pod := &corev1.Pod{}
	err = b.APIReader.Get(b.Ctx, types.NamespacedName{
		Name:      b.Request.GetPodName(),
		Namespace: b.Request.Namespace,
	}, pod)
	if err != nil {
		return "Error Fetching Pod", err
	}

	// Now, check the Pod ready status
	if pod.Status.Phase != corev1.PodRunning {
		statusMsg := fmt.Sprintf("Pod in %s Phase", pod.Status.Phase)
		return statusMsg, errors.New(statusMsg)
	}

	// Finally, return the pod phase
	return fmt.Sprintf("Pod is %s", pod.Status.Phase), nil
}

func (b *PodAccessBuilder) generatePodTemplateSpec() (corev1.PodTemplateSpec, error) {
	return b.getPodTemplateFromController()
}
