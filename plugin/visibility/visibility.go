package main

import (
	"fmt"
	"os"

	abbottPlugin "github.com/tidepool-org/platform-plugin-abbott/abbott/plugin"
	tandemPlugin "github.com/tidepool-org/platform-plugin-tandem/tandem/plugin"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: visibility <plugin>")
		os.Exit(1)
	}

	switch plugin := os.Args[1]; plugin {
	case "abbott":
		fmt.Println(abbottPlugin.Visibility())
	case "tandem":
		fmt.Println(tandemPlugin.Visibility())
	default:
		fmt.Fprintf(os.Stderr, "unknown plugin %q\n", plugin)
		os.Exit(1)
	}
}
