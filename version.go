package doxiego

import "fmt"

// Semantic versioning - http://semver.org/
const (
	Major = 2
	Minor = 0
	Patch = 0
)

// Version returns library version.
func Version() string {
	return fmt.Sprintf("%d.%d.%d", Major, Minor, Patch)
}
