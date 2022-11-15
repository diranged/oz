package v1alpha1

type ControllerKind string

const (
	DeploymentController  ControllerKind = "Deployment"
	DaemonSetController   ControllerKind = "DaemonSet"
	StatefulSetController ControllerKind = "StatefulSet"
)

const (
	// TemplateAvailability is the string used for the primary status condition that indicates
	// whether or not an `AccessTemplate` or `ExecAccessTemplate` is ready for use.
	TemplateAvailability = "TemplateAvailable"

	// TemplateAvailabilityStatusAvailable represents the status of the Template when it is healthy and ready to use.
	TemplateAvailabilityStatusAvailable = "Available"

	// TemplateAvailabilityStatusDegraded indicates that the Template is unable to be used
	TemplateAvailabilityStatusDegraded = "Degraded"
)
