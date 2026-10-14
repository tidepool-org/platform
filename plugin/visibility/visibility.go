package main

import (
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"

	abbottPlugin "github.com/tidepool-org/platform-plugin-abbott/abbott/plugin"
)

var pluginVisibilityFuncs = map[string]func() string{
	"abbott": abbottPlugin.Visibility,
}

func main() {
	if argsLength := len(os.Args); argsLength < 2 {
		fail("missing argument: <plugin>")
	} else if argsLength > 2 {
		fail(fmt.Sprintf("unexpected arguments: %s", strings.Join(os.Args[2:], " ")))
	}

	pluginVisibilityFunc, ok := pluginVisibilityFuncs[os.Args[1]]
	if !ok {
		fail(fmt.Sprintf("plugin not registered: %s", os.Args[1]))
	}

	fmt.Println(pluginVisibilityFunc())
}

func fail(message string) {
	fmt.Fprintf(os.Stderr, "ERROR: %s\n", message)
	fmt.Fprintf(os.Stderr, "usage: visibility <plugin>\n")
	fmt.Fprintf(os.Stderr, "available plugins: %s\n", strings.Join(slices.Sorted(maps.Keys(pluginVisibilityFuncs)), ", "))
	os.Exit(1)
}
