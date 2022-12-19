package utils

import "sigs.k8s.io/controller-runtime/pkg/client"

const shortUIDLength = 8

// getShortUID returns back a shortened version of the UID that the Kubernetes cluster used to store
// the AccessRequest internally. This is used by the Builders to create unique names for the
// resources they manage (Roles, RoleBindings, etc).
//
// Returns:
//
//	shortUID: A 10-digit long shortened UID
func getShortUID(obj client.Object) string {
	// TODO: If the UID isn't there, we should generate something random OR throw an error.
	return string(obj.GetUID())[0:shortUIDLength]
}
