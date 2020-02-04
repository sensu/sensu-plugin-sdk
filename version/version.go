package version

import "fmt"

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func Version() string {
	return fmt.Printf("%v, commit %v, built at %v", version, commit, date)
}
