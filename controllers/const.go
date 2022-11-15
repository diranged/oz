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
	ConditionTargetTemplateExists OzResourceConditionTypes = "TargetTemplateExists"
	ConditionTargetPodExists      OzResourceConditionTypes = "TargetPodExists"
	ConditionTargetPodSelected    OzResourceConditionTypes = "TargetPodSelected"
	ConditionRoleCreated          OzResourceConditionTypes = "RoleCreated"
	ConditionRoleBindingCreated   OzResourceConditionTypes = "RoleBindingCreated"
	ConditionAccessStillValid     OzResourceConditionTypes = "AccessStillValid"

	// Access Templates
	ConditionTargetRefExists OzResourceConditionTypes = "TargetRefExists"
)
