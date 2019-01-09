package reservoirchange_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataTest "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/device"
	"github.com/tidepool-org/platform/data/types/device/reservoirchange"
	dataTypesDeviceStatus "github.com/tidepool-org/platform/data/types/device/status"
	dataTypesDeviceStatusTest "github.com/tidepool-org/platform/data/types/device/status/test"
	dataTypesDeviceTest "github.com/tidepool-org/platform/data/types/device/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func NewMeta() interface{} {
	return &device.Meta{
		Type:    "deviceEvent",
		SubType: "reservoirChange",
	}
}

func NewReservoirChange() *reservoirchange.ReservoirChange {
	datum := reservoirchange.New()
	datum.Device = *dataTypesDeviceTest.NewDevice()
	datum.SubType = "reservoirChange"
	return datum
}

func NewReservoirChangeWithStatus() *reservoirchange.ReservoirChange {
	var status data.Datum
	status = dataTypesDeviceStatusTest.NewStatus()
	datum := NewReservoirChange()
	datum.Status = &status
	return datum
}

func NewReservoirChangeWithStatusID() *reservoirchange.ReservoirChange {
	datum := NewReservoirChange()
	datum.StatusID = pointer.FromString(dataTest.RandomID())
	return datum
}

func CloneReservoirChange(datum *reservoirchange.ReservoirChange) *reservoirchange.ReservoirChange {
	if datum == nil {
		return nil
	}
	clone := reservoirchange.New()
	clone.Device = *dataTypesDeviceTest.CloneDevice(&datum.Device)
	if datum.Status != nil {
		switch status := (*datum.Status).(type) {
		case *dataTypesDeviceStatus.Status:
			clone.Status = data.DatumAsPointer(dataTypesDeviceStatusTest.CloneStatus(status))
		}
	}
	clone.StatusID = pointer.CloneString(datum.StatusID)
	return clone
}

var _ = Describe("Change", func() {
	It("SubType is expected", func() {
		Expect(reservoirchange.SubType).To(Equal("reservoirChange"))
	})

	Context("New", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := reservoirchange.New()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("deviceEvent"))
			Expect(datum.SubType).To(Equal("reservoirChange"))
			Expect(datum.Status).To(BeNil())
			Expect(datum.StatusID).To(BeNil())
		})
	})

	Context("ReservoirChange", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *reservoirchange.ReservoirChange), expectedErrors ...error) {
					datum := NewReservoirChange()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *reservoirchange.ReservoirChange) {},
				),
				Entry("type missing",
					func(datum *reservoirchange.ReservoirChange) { datum.Type = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &device.Meta{SubType: "reservoirChange"}),
				),
				Entry("type invalid",
					func(datum *reservoirchange.ReservoirChange) { datum.Type = "invalidType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "deviceEvent"), "/type", &device.Meta{Type: "invalidType", SubType: "reservoirChange"}),
				),
				Entry("type device",
					func(datum *reservoirchange.ReservoirChange) { datum.Type = "deviceEvent" },
				),
				Entry("sub type missing",
					func(datum *reservoirchange.ReservoirChange) { datum.SubType = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/subType", &device.Meta{Type: "deviceEvent"}),
				),
				Entry("sub type invalid",
					func(datum *reservoirchange.ReservoirChange) { datum.SubType = "invalidSubType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "reservoirChange"), "/subType", &device.Meta{Type: "deviceEvent", SubType: "invalidSubType"}),
				),
				Entry("sub type reservoir change",
					func(datum *reservoirchange.ReservoirChange) { datum.SubType = "reservoirChange" },
				),
				Entry("multiple errors",
					func(datum *reservoirchange.ReservoirChange) {
						datum.Type = "invalidType"
						datum.SubType = "invalidSubType"
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "deviceEvent"), "/type", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "reservoirChange"), "/subType", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
				),
			)

			DescribeTable("validates the datum with origin external",
				func(mutator func(datum *reservoirchange.ReservoirChange), expectedErrors ...error) {
					datum := NewReservoirChangeWithStatus()
					mutator(datum)
					dataTypesTest.ValidateWithOrigin(datum, structure.OriginExternal, expectedErrors...)
				},
				Entry("succeeds",
					func(datum *reservoirchange.ReservoirChange) {},
				),
				Entry("status missing",
					func(datum *reservoirchange.ReservoirChange) { datum.Status = nil },
				),
				Entry("status valid",
					func(datum *reservoirchange.ReservoirChange) {
						datum.Status = data.DatumAsPointer(dataTypesDeviceStatusTest.NewStatus())
					},
				),
				Entry("status id missing",
					func(datum *reservoirchange.ReservoirChange) { datum.StatusID = nil },
				),
				Entry("status id exists",
					func(datum *reservoirchange.ReservoirChange) { datum.StatusID = pointer.FromString(dataTest.RandomID()) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/statusId", NewMeta()),
				),
				Entry("multiple errors",
					func(datum *reservoirchange.ReservoirChange) {
						datum.StatusID = pointer.FromString(dataTest.RandomID())
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/statusId", NewMeta()),
				),
			)

			DescribeTable("validates the datum with origin internal/store",
				func(mutator func(datum *reservoirchange.ReservoirChange), expectedErrors ...error) {
					datum := NewReservoirChangeWithStatusID()
					mutator(datum)
					dataTypesTest.ValidateWithOrigin(datum, structure.OriginInternal, expectedErrors...)
					dataTypesTest.ValidateWithOrigin(datum, structure.OriginStore, expectedErrors...)
				},
				Entry("succeeds",
					func(datum *reservoirchange.ReservoirChange) {},
				),
				Entry("status missing",
					func(datum *reservoirchange.ReservoirChange) { datum.Status = nil },
				),
				Entry("status exists",
					func(datum *reservoirchange.ReservoirChange) {
						datum.Status = data.DatumAsPointer(dataTypesDeviceStatusTest.NewStatus())
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/status", NewMeta()),
				),
				Entry("status id missing",
					func(datum *reservoirchange.ReservoirChange) { datum.StatusID = nil },
				),
				Entry("status id invalid",
					func(datum *reservoirchange.ReservoirChange) { datum.StatusID = pointer.FromString("invalid") },
					errorsTest.WithPointerSourceAndMeta(data.ErrorValueStringAsIDNotValid("invalid"), "/statusId", NewMeta()),
				),
				Entry("status id valid",
					func(datum *reservoirchange.ReservoirChange) { datum.StatusID = pointer.FromString(dataTest.RandomID()) },
				),
				Entry("multiple errors",
					func(datum *reservoirchange.ReservoirChange) {
						datum.Status = data.DatumAsPointer(dataTypesDeviceStatusTest.NewStatus())
						datum.StatusID = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/status", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(data.ErrorValueStringAsIDNotValid("invalid"), "/statusId", NewMeta()),
				),
			)
		})

		Context("Normalize", func() {
			It("does not modify datum if status is missing", func() {
				datum := NewReservoirChangeWithStatusID()
				expectedDatum := CloneReservoirChange(datum)
				normalizer := dataNormalizer.New()
				Expect(normalizer).ToNot(BeNil())
				datum.Normalize(normalizer)
				Expect(normalizer.Error()).To(BeNil())
				Expect(normalizer.Data()).To(BeEmpty())
				Expect(datum).To(Equal(expectedDatum))
			})

			It("normalizes the datum and replaces status with status id", func() {
				datumStatus := dataTypesDeviceStatusTest.NewStatus()
				datum := NewReservoirChangeWithStatusID()
				datum.Status = data.DatumAsPointer(datumStatus)
				expectedDatum := CloneReservoirChange(datum)
				normalizer := dataNormalizer.New()
				Expect(normalizer).ToNot(BeNil())
				datum.Normalize(normalizer)
				Expect(normalizer.Error()).To(BeNil())
				Expect(normalizer.Data()).To(Equal([]data.Datum{datumStatus}))
				expectedDatum.Status = nil
				expectedDatum.StatusID = pointer.FromString(*datumStatus.ID)
				Expect(datum).To(Equal(expectedDatum))
			})
		})
	})
})
