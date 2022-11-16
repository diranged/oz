package builders

import (
	"context"
	"fmt"

	"github.com/diranged/oz/interfaces"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Builder interface {
	GetClient() client.Client
	GetCtx() context.Context
	GetScheme() *runtime.Scheme

	GetRequest() interfaces.OzRequestResource
	GetTemplate() interfaces.OzTemplateResource

	// Returns back the PodName that the user is being granted direct access to.
	GeneratePodName() (podName string, err error)

	// Generates all of the resources required to fulfill the access request.
	GenerateAccessResources() (statusString string, accessString string, err error)
}

type BaseBuilder struct {
	Builder

	Client client.Client
	Ctx    context.Context
	Scheme *runtime.Scheme

	// ApiReader should be generated with mgr.GetAPIReader() to create a non-cached client object. This is used
	// for certain Get() calls where we need to ensure we are getting the latest version from the API, and not a cached
	// object.
	//
	// See https://github.com/kubernetes-sigs/controller-runtime/issues/585#issuecomment-528102351
	//
	ApiReader client.Reader

	// Generic struct that satisfies the OzRequestResource interface. This is used for the common
	// functions inside the BaseBuilder struct.
	Request interfaces.OzRequestResource

	// Generic struct that satisfies the OzTemplateREsource interface. This is used for the common
	// functions inside the BaseBuilder struct.
	Template interfaces.OzTemplateResource
}

func (t *BaseBuilder) GetClient() client.Client {
	return t.Client
}

func (t *BaseBuilder) GetCtx() context.Context {
	return t.Ctx
}

func (t *BaseBuilder) GetScheme() *runtime.Scheme {
	return t.Scheme
}

func (t *BaseBuilder) GetTemplate() interfaces.OzTemplateResource {
	return t.Template
}

func (t *BaseBuilder) GetRequest() interfaces.OzRequestResource {
	return t.Request
}

func (t *BaseBuilder) GetTargetRefResource() (client.Object, error) {
	// https://blog.gripdev.xyz/2020/07/20/k8s-operator-with-dynamic-crds-using-controller-runtime-no-structs/
	obj := t.Template.GetTargetRef().GetObject()
	err := t.Client.Get(t.Ctx, types.NamespacedName{
		Name:      t.Template.GetTargetRef().GetName(),
		Namespace: t.Template.GetNamespace(),
	}, obj)
	return obj, err
}

func (t *BaseBuilder) VerifyPodExists(name string, namespace string) error {
	logger := log.FromContext(t.Ctx)
	logger.Info(fmt.Sprintf("Verifying that Pod %s still exists...", name))

	// Search for the Pod
	pod := &corev1.Pod{}
	err := t.Client.Get(t.Ctx, types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, pod)

	// On any failure, update the pod status with the failure...
	if err != nil {
		return fmt.Errorf("pod %s (ns: %s) is not found: %s", name, namespace, err)
	}
	return nil
}
