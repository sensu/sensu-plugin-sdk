package version

import "fmt"

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// Version returns the plugin version, along with information about the commit
// the plugin was built at, along with the date.
func Version() string {
	return fmt.Sprintf("%v, commit %v, built at %v", version, commit, date)
}
