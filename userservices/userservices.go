package main

import (
	"fmt"

	"github.com/tidepool-org/platform/user"
	"github.com/tidepool-org/platform/version"
)

func main() {
	fmt.Println(version.String)
	fmt.Println(user.GetUser())
}
