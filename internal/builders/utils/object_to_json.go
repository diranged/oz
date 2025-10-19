package bldutil

import (
	"encoding/json"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ObjectToJSON is a quick helper function for pretty-printing an entire K8S object in JSON form.
// Used in certain debug log statements primarily.
func ObjectToJSON(obj client.Object) string {
	jsonData, err := json.Marshal(obj)
	if err != nil {
		fmt.Printf("could not marshal json: %s\n", err)
		return ""
	}
	return string(jsonData)
}
