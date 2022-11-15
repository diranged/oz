package controllers

const (
	fieldSelectorMetadataName string = "metadata.name"
	fieldSelectorStatusPhase  string = "status.phase"
)

type BaseResourceConditionTypes string

const (
	// Both
	ConditionDurationsValid BaseResourceConditionTypes = "AccessDurationsValid"

	// Access Requests
	ConditionTargetTemplateExists BaseResourceConditionTypes = "TargetTemplateExists"
	ConditionTargetPodExists      BaseResourceConditionTypes = "TargetPodExists"
	ConditionTargetPodSelected    BaseResourceConditionTypes = "TargetPodSelected"
	ConditionRoleCreated          BaseResourceConditionTypes = "RoleCreated"
	ConditionRoleBindingCreated   BaseResourceConditionTypes = "RoleBindingCreated"
	ConditionAccessStillValid     BaseResourceConditionTypes = "AccessStillValid"

	// Access Templates
	ConditionTargetRefExists BaseResourceConditionTypes = "TargetRefExists"
)
