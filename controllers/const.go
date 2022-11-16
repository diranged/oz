package controllers

const (
	fieldSelectorMetadataName string = "metadata.name"
	fieldSelectorStatusPhase  string = "status.phase"
)

const (
	DEFAULT_RECONCILIATION_INTERVAL int = 5
)

type OzResourceConditionTypes string

const (
	// Both
	ConditionDurationsValid OzResourceConditionTypes = "AccessDurationsValid"

	// Access Requests
	ConditionTargetTemplateExists   OzResourceConditionTypes = "TargetTemplateExists"
	ConditionRoleCreated            OzResourceConditionTypes = "RoleCreated"
	ConditionRoleBindingCreated     OzResourceConditionTypes = "RoleBindingCreated"
	ConditionAccessStillValid       OzResourceConditionTypes = "AccessStillValid"
	ConditionAccessResourcesCreated OzResourceConditionTypes = "AccessResourcesCreated"

	// TODO: maybe get ridof?
	ConditionTargetPodSelected OzResourceConditionTypes = "TargetPodSelected"

	// Access Templates
	ConditionTargetRefExists OzResourceConditionTypes = "TargetRefExists"
)
