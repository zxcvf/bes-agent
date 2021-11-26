package py

import (
	"fmt"
	"hash/fnv"
	"strings"
)

// ID is the representation of the unique ID of a Check instance
type ID string

// Identify returns an unique ID for a check and its configuration
func Identify(check *PythonCheck, instance interface{}, initConfig interface{}) ID {
	return BuildID(check.ModuleName, instance, initConfig)
}

// BuildID returns an unique ID for a check name and its configuration
func BuildID(checkName string, instance, initConfig interface{}) ID {
	h := fnv.New64()

	id := fmt.Sprintf("%s:%x", checkName, h.Sum64())
	return ID(id)
}

// IDToCheckName returns the check name from a check ID
func IDToCheckName(id ID) string {
	return strings.SplitN(string(id), ":", 2)[0]
}
