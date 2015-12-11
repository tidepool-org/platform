package main

import (
	"fmt"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/version"
)

func main() {
	fmt.Println(version.String)
	fmt.Println(data.GetData())
}
