package pumpstatus_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"time"

	dataTypes "github.com/tidepool-org/platform/data/types"
	dataTypesPumpStatus "github.com/tidepool-org/platform/data/types/pumpstatus"
	dataTypesPumpStatusTest "github.com/tidepool-org/platform/data/types/pumpstatus/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func NewMeta() interface{} {
	return &dataTypes.Meta{
		Type: "pumpStatus",
	}
}

var _ = Describe("PumpStatus", func() {
	It("Type is expected", func() {
		Expect(dataTypesPumpStatus.Type).To(Equal("pumpStatus"))
	})

	It("TimeFormat is expected", func() {
		Expect(dataTypesPumpStatus.TimeFormat).To(Equal(time.RFC3339Nano))
	})

	Context("New", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := dataTypesPumpStatus.New()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("pumpStatus"))
			Expect(datum.BasalDelivery).To(BeNil())
			Expect(datum.Battery).To(BeNil())
			Expect(datum.BolusDelivery).To(BeNil())
			Expect(datum.Device).To(BeNil())
			Expect(datum.Reservoir).To(BeNil())
		})
	})

	Context("PumpStatus", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("return the expected results when the input",
				func(mutator func(datum *dataTypesPumpStatus.PumpStatus), expectedErrors ...error) {
					datum := dataTypesPumpStatusTest.RandomPumpStatus()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesPumpStatus.PumpStatus) {},
				),
				Entry("type missing",
					func(datum *dataTypesPumpStatus.PumpStatus) { datum.Type = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &dataTypes.Meta{}),
				),
				Entry("type invalid",
					func(datum *dataTypesPumpStatus.PumpStatus) { datum.Type = "invalidType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "pumpStatus"), "/type", &dataTypes.Meta{Type: "invalidType"}),
				),
				Entry("type pumpStatus",
					func(datum *dataTypesPumpStatus.PumpStatus) { datum.Type = "pumpStatus" },
				),
				Entry("basal delivery invalid",
					func(datum *dataTypesPumpStatus.PumpStatus) {
						datum.BasalDelivery.State = nil
						datum.BasalDelivery.Time = nil
						datum.BasalDelivery.Dose = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/basalDelivery/state", NewMeta()),
				),
				Entry("battery invalid",
					func(datum *dataTypesPumpStatus.PumpStatus) { datum.Battery.Remaining = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/battery/remaining", NewMeta()),
				),
				Entry("bolus delivery invalid",
					func(datum *dataTypesPumpStatus.PumpStatus) {
						datum.BolusDelivery.State = nil
						datum.BolusDelivery.Dose = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bolusDelivery/state", NewMeta()),
				),
				Entry("device invalid",
					func(datum *dataTypesPumpStatus.PumpStatus) { datum.Device.ID = pointer.FromString("") },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotInRange(0, 1, 1000), "/device/id", NewMeta()),
				),
				Entry("reservoir invalid",
					func(datum *dataTypesPumpStatus.PumpStatus) { datum.Reservoir.Remaining = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/reservoir/remaining", NewMeta()),
				),
				Entry("multiple errors",
					func(datum *dataTypesPumpStatus.PumpStatus) {
						datum.BasalDelivery.State = nil
						datum.BasalDelivery.Time = nil
						datum.BasalDelivery.Dose = nil
						datum.Battery.Remaining = nil
						datum.BolusDelivery.State = nil
						datum.BolusDelivery.Dose = nil
						datum.Device.ID = pointer.FromString("")
						datum.Reservoir.Remaining = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/basalDelivery/state", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/battery/remaining", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bolusDelivery/state", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotInRange(0, 1, 1000), "/device/id", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/reservoir/remaining", NewMeta()),
				),
			)
		})
	})
})
