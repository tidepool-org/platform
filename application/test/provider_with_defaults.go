package test

import (
	"fmt"
	"strings"

	"github.com/tidepool-org/platform/config"
	configTest "github.com/tidepool-org/platform/config/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
	"github.com/tidepool-org/platform/version"
	versionTest "github.com/tidepool-org/platform/version/test"
)

func NewProviderWithDefaults() *Provider {
	var versionReporter version.Reporter = versionTest.NewReporter()
	var configReporter config.Reporter = configTest.NewReporter()
	var logger log.Logger = logTest.NewLogger()
	provider := NewProvider()
	provider.VersionReporterOutput = &versionReporter
	provider.ConfigReporterOutput = &configReporter
	provider.LoggerOutput = &logger
	provider.PrefixOutput = pointer.FromString(test.RandomStringFromRangeAndCharset(4, 8, test.CharsetUppercase))
	provider.NameOutput = pointer.FromString(test.RandomStringFromRangeAndCharset(4, 16, test.CharsetAlphaNumeric))
	provider.UserAgentOutput = pointer.FromString(fmt.Sprintf("%s-%s/%s",
		strings.ToTitle(strings.ToLower(*provider.PrefixOutput)),
		strings.ToTitle(strings.ToLower(*provider.NameOutput)),
		versionReporter.Base(),
	))
	return provider
}
