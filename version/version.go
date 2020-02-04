package version

import "fmt"

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func Version() string {
	fmt.Printf("%v, commit %v, built at %v\n", version, commit, date)
}
