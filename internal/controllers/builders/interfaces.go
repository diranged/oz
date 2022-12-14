package builders

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	api "github.com/diranged/oz/internal/api/v1alpha1"
)

// IBuilder defines the interface for a particular "access builder". An "access builder" is typically
// paired with an "access template" struct in the api.v1alpha1 package. Each unique type of access
// template will have its own access builder that is used to implement the goals of that particular
// template.
//
// Common interface functions are used to keep the reconiliation loop code in the individual
// controllers package clean.
type IBuilder interface {
	// GetClient returns a Kubernetes client.Client object that can be used for making safe and
	// cached calls to the API.
	GetClient() client.Client

	// GetCtx returns the context.Context object that is used to hand off async API calls to the
	// system.
	GetCtx() context.Context

	// GetScheme returns the runtime.Scheme that is populated for the API client, ensuring that we
	// understand the local CRDs from this controller.
	GetScheme() *runtime.Scheme

	// GetRequest returns an Access Request resource that conforms to the api.IPodRequestResource
	// interface.
	//
	// TODO: Generalize this into just an api.IRequestResource interface, and use a PodRequestResource
	// more specifically for the PodAccessBuilder.
	GetRequest() api.IPodRequestResource

	// GetTemplate returns an Access Template resouce that conforms to the api.ITemplateResource
	// interface.
	GetTemplate() api.ITemplateResource

	// Generates all of the resources required to fulfill the access request.
	GenerateAccessResources() (statusString string, err error)

	// GetTargetRefResource returns a generic but populated client.Object resource from an Access
	// Template. Typically this is a Deployment, DaemonSet, etc.
	GetTargetRefResource() (client.Object, error)
}

// IPodAccessBuilder is an extended interface from the IBuilder that provides a few additional
// common methods that are specific to validating Access Templates that provide Pod-level access
// for developers.
type IPodAccessBuilder interface {
	IBuilder

	// Returns back the status of the various access resources. If they are not
	// ready yet, this stage will prevent the Ozctl tool from thinking the
	// access request has been fulfilled.
	VerifyAccessResources() (statusString string, err error)
}
