package version

import "fmt"

var Base string
var Commit string

func Short() string {
	return fmt.Sprintf("%s+%s", Base, Commit[0:8])
}

func Long() string {
	return fmt.Sprintf("%s+%s", Base, Commit)
}
