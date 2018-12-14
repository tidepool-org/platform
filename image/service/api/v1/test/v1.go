package test

import (
	"fmt"
	"strings"

	"github.com/tidepool-org/platform/image"
	imageTest "github.com/tidepool-org/platform/image/test"
	"github.com/tidepool-org/platform/test"
	testHttp "github.com/tidepool-org/platform/test/http"
)

func RandomPathSuffix() string {
	parts := make([]string, test.RandomIntFromRange(0, 2))
	for index := range parts {
		parts[index] = testHttp.RandomPathPart()
	}
	extension, _ := image.ExtensionFromMediaType(imageTest.RandomMediaType())
	parts = append(parts, fmt.Sprintf("%s.%s", testHttp.RandomPathPart(), extension))
	return strings.Join(parts, "/")
}
