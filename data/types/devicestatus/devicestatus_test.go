package devicestatus_test

import (
	. "github.com/onsi/ginkgo"

	"github.com/tidepool-org/platform/data/types/devicestatus"
)

func NewDeviceStatus() *devicestatus.DeviceStatus {
	datum := devicestatus.NewDeviceStatus()
	return datum
}

func CloneDeviceStatus(datum *devicestatus.DeviceStatus) *devicestatus.DeviceStatus {
	if datum == nil {
		return nil
	}
	clone := devicestatus.NewDeviceStatus()
	return clone
}

var _ = Describe("DeviceStatus", func() {

	Context("ParseDeviceStatus", func() {
		// TODO
	})

	Context("NewDeviceStatus", func() {
	})

	Context("DeviceStatus", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
		})
	})
})
