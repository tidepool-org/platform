package devicestatus_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"

	"github.com/tidepool-org/platform/structure"

	"github.com/tidepool-org/platform/data/types/devicestatus"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	InvalidType  = "invalidType"
	ValidVersion = "1.0"
)

func NewMeta() interface{} {
	return &types.Meta{
		Type: devicestatus.Type,
	}
}

func NewDeviceStatus() *devicestatus.DeviceStatus {
	datum := devicestatus.NewDeviceStatus()
	datum.Base = *dataTypesTest.NewBase()
	datum.DeviceType = pointer.FromString(test.RandomStringFromArray(devicestatus.DeviceTypes()))
	datum.Type = devicestatus.Type
	datum.Version = pointer.FromString(ValidVersion)
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

	Context("DeviceStatus", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("return the expected results when the input",

				func(mutator func(datum *devicestatus.DeviceStatus), expectedErrors ...error) {
					datum := NewDeviceStatus()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *devicestatus.DeviceStatus) {},
				),
				Entry("invalid Device Type",
					func(datum *devicestatus.DeviceStatus) {
						datum.DeviceType = pointer.FromString(InvalidType)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf(InvalidType, devicestatus.DeviceTypes()), "/deviceType", NewMeta()),
				),
				Entry("version missing",
					func(datum *devicestatus.DeviceStatus) { datum.Version = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/version", NewMeta()),
				),
			)
		})
	})
})
