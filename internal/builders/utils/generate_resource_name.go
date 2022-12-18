package utils

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GenerateResourceName takes in an API.IRequestResource conforming object and returns a unique
// resource name string that can be used to safely create other resources (roles, bindings, etc).
//
// Returns:
//
//	string: A resource name string
func GenerateResourceName(req client.Object) string {
	return fmt.Sprintf("%s-%s", req.GetName(), getShortUID(req))
}
