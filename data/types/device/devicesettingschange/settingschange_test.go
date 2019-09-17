package devicesettingschange_test

import (
	. "github.com/onsi/ginkgo"

	"github.com/tidepool-org/platform/data/types/device/devicesettingschange"
	"github.com/tidepool-org/platform/pointer"
)

func NewSettingsChange() *devicesettingschange.SettingsChange {
	datum := devicesettingschange.NewSettingsChange()
	return datum
}

func CloneSettingsChange(datum *devicesettingschange.SettingsChange) *devicesettingschange.SettingsChange {
	if datum == nil {
		return nil
	}
	clone := devicesettingschange.NewSettingsChange()
	clone.From = pointer.CloneString(datum.From)
	clone.To = pointer.CloneString(datum.To)
	return clone
}

var _ = Describe("SettingsChange", func() {
})
