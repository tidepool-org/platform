package app_settings_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	app_settings "github.com/tidepool-org/platform/data/types/settings/app"

	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"

	"github.com/tidepool-org/platform/test"
)

func RandomAppSettings() *app_settings.AppSettings {
	settings := app_settings.NewAppSettings()
	settings.LoopAppVersion = pointer.FromString(test.RandomStringFromRange(app_settings.MinVersionLength, app_settings.MaxVersionLength))
	settings.Name = pointer.FromString(test.RandomStringFromRange(app_settings.MinNameLength, app_settings.MaxNameLength))

	return settings
}

var _ = Describe("App Settings", func() {
	Context("App Settings", func() {
		DescribeTable("return the expected results when the input",
			func(mutator func(datum *app_settings.AppSettings), expectedErrors ...error) {
				datum := RandomAppSettings()
				mutator(datum)
				dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
			},
			Entry("succeeds",
				func(datum *app_settings.AppSettings) {},
			),
			Entry("name invalid",
				func(datum *app_settings.AppSettings) { datum.Name = pointer.FromString("") },
				errorsTest.WithPointerSource(structureValidator.ErrorLengthNotInRange(0, app_settings.MinNameLength, app_settings.MaxNameLength), "/name"),
			),
			Entry("loop app version invalid",
				func(datum *app_settings.AppSettings) { datum.LoopAppVersion = pointer.FromString("") },
				errorsTest.WithPointerSource(structureValidator.ErrorLengthNotInRange(0, app_settings.MinVersionLength, app_settings.MaxVersionLength), "/loopAppVersion"),
			),
		)
	})

})
