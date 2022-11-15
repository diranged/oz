package controllers

const (
	fieldSelectorMetadataName string = "metadata.name"
	fieldSelectorStatusPhase  string = "status.phase"
)

type RequestConditionTypes string

const (
	// Both
	ConditionDurationsValid RequestConditionTypes = "AccessDurationsValid"

	// Access Requests
	ConditionTargetTemplateExists RequestConditionTypes = "TargetTemplateExists"
	ConditionTargetPodExists      RequestConditionTypes = "TargetPodExists"
	ConditionTargetPodSelected    RequestConditionTypes = "TargetPodSelected"
	ConditionRoleCreated          RequestConditionTypes = "RoleCreated"
	ConditionRoleBindingCreated   RequestConditionTypes = "RoleBindingCreated"
	ConditionAccessStillValid     RequestConditionTypes = "AccessStillValid"

	// Access Templates
	ConditionTargetRefExists RequestConditionTypes = "TargetRefExists"
)
