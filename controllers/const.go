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
	// conditionDurationsValid is used by both AccessTemplate and AccessRequest resources. It
	// indicates whether or not the various duration fields are valid.
	conditionDurationsValid OzResourceConditionTypes = "AccessDurationsValid"

	// Access Requests
	conditionTargetTemplateExists   OzResourceConditionTypes = "TargetTemplateExists"
	conditionRoleCreated            OzResourceConditionTypes = "RoleCreated"
	conditionRoleBindingCreated     OzResourceConditionTypes = "RoleBindingCreated"
	conditionAccessStillValid       OzResourceConditionTypes = "AccessStillValid"
	conditionAccessResourcesCreated OzResourceConditionTypes = "AccessResourcesCreated"

	// TODO: maybe get ridof?
	conditionTargetPodSelected OzResourceConditionTypes = "TargetPodSelected"

	// Access Templates
	conditionTargetRefExists OzResourceConditionTypes = "TargetRefExists"
)
