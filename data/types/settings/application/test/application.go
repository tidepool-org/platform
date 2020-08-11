package test

import (
	dataTypesSettingsApplication "github.com/tidepool-org/platform/data/types/settings/application"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomApplication() *dataTypesSettingsApplication.Application {
	datum := dataTypesSettingsApplication.New()
	datum.Base = *dataTypesTest.NewBase()
	datum.Type = "applicationSettings"
	datum.Name = pointer.FromString(test.RandomStringFromRange(dataTypesSettingsApplication.NameLengthMinimum, dataTypesSettingsApplication.NameLengthMaximum))
	datum.Version = pointer.FromString(test.RandomStringFromRange(dataTypesSettingsApplication.VersionLengthMinimum, dataTypesSettingsApplication.VersionLengthMaximum))
	return datum
}

func CloneApplication(datum *dataTypesSettingsApplication.Application) *dataTypesSettingsApplication.Application {
	if datum == nil {
		return nil
	}
	clone := dataTypesSettingsApplication.New()
	clone.Base = *dataTypesTest.CloneBase(&datum.Base)
	clone.Name = pointer.CloneString(datum.Name)
	clone.Version = pointer.CloneString(datum.Version)
	return clone
}
