package controllers

const (
	fieldSelectorMetadataName string = "metadata.name"
	fieldSelectorStatusPhase  string = "status.phase"
)

const (
	// DefaultReconciliationInterval defines the number of minutes inbetween regular scheduled
	// checks of the target resources that our controllers are managing.
	DefaultReconciliationInterval int = 5

	// ErrorReconciliationInterval defines how long (in seconds) in between a failed reconciliation
	// loop before the next one should kick off.
	ErrorReconciliationInterval int = 30
)

// OzResourceConditionTypes defines a set of known Status.Condition[].ConditionType fields that are
// used throughout the AccessRequest and AccessTemplate reconcilers.
type OzResourceConditionTypes string

const (
	// ConditionDurationsValid is used by both AccessTemplate and AccessRequest resources. It
	// indicates whether or not the various duration fields are valid.
	ConditionDurationsValid OzResourceConditionTypes = "AccessDurationsValid"

	// ConditionTargetTemplateExists indicates that the Access Request is pointing to a valid Access
	// Template.
	ConditionTargetTemplateExists OzResourceConditionTypes = "TargetTemplateExists"

	// ConditionAccessStillValid is continaully updated based on whether or not the Access Request
	// has timed out.
	ConditionAccessStillValid OzResourceConditionTypes = "AccessStillValid"

	// ConditionAccessResourcesCreated indicates whether or not the target access request resources
	// have been properly created.
	ConditionAccessResourcesCreated OzResourceConditionTypes = "AccessResourcesCreated"

	// ConditionTargetRefExists indicates whether or not an AccessTemplate is pointing to a valid
	// Controller.
	ConditionTargetRefExists OzResourceConditionTypes = "TargetRefExists"
)
