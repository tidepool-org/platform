package reservoirchange_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/device"
	"github.com/tidepool-org/platform/data/types/device/reservoirchange"
	dataTypesDeviceStatus "github.com/tidepool-org/platform/data/types/device/status"
	testDataTypesDeviceStatus "github.com/tidepool-org/platform/data/types/device/status/test"
	testDataTypesDevice "github.com/tidepool-org/platform/data/types/device/test"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewMeta() interface{} {
	return &device.Meta{
		Type:    "deviceEvent",
		SubType: "reservoirChange",
	}
}

func NewReservoirChange() *reservoirchange.ReservoirChange {
	datum := reservoirchange.New()
	datum.Device = *testDataTypesDevice.NewDevice()
	datum.SubType = "reservoirChange"
	return datum
}

func NewReservoirChangeWithStatus() *reservoirchange.ReservoirChange {
	var status data.Datum
	status = testDataTypesDeviceStatus.NewStatus()
	datum := NewReservoirChange()
	datum.Status = &status
	return datum
}

func NewReservoirChangeWithStatusID() *reservoirchange.ReservoirChange {
	datum := NewReservoirChange()
	datum.StatusID = pointer.String(id.New())
	return datum
}

func CloneReservoirChange(datum *reservoirchange.ReservoirChange) *reservoirchange.ReservoirChange {
	if datum == nil {
		return nil
	}
	clone := reservoirchange.New()
	clone.Device = *testDataTypesDevice.CloneDevice(&datum.Device)
	if datum.Status != nil {
		switch status := (*datum.Status).(type) {
		case *dataTypesDeviceStatus.Status:
			clone.Status = data.DatumAsPointer(testDataTypesDeviceStatus.CloneStatus(status))
		}
	}
	clone.StatusID = test.CloneString(datum.StatusID)
	return clone
}

var _ = Describe("Change", func() {
	It("SubType is expected", func() {
		Expect(reservoirchange.SubType).To(Equal("reservoirChange"))
	})

	Context("NewDatum", func() {
		It("returns the expected datum", func() {
			Expect(reservoirchange.NewDatum()).To(Equal(&reservoirchange.ReservoirChange{}))
		})
	})

	Context("New", func() {
		It("returns the expected datum", func() {
			Expect(reservoirchange.New()).To(Equal(&reservoirchange.ReservoirChange{}))
		})
	})

	Context("Init", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := reservoirchange.Init()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("deviceEvent"))
			Expect(datum.SubType).To(Equal("reservoirChange"))
			Expect(datum.Status).To(BeNil())
			Expect(datum.StatusID).To(BeNil())
		})
	})

	Context("with new datum", func() {
		var datum *reservoirchange.ReservoirChange

		BeforeEach(func() {
			datum = NewReservoirChange()
		})

		Context("Init", func() {
			It("initializes the datum", func() {
				datum.Init()
				Expect(datum.Type).To(Equal("deviceEvent"))
				Expect(datum.SubType).To(Equal("reservoirChange"))
				Expect(datum.Status).To(BeNil())
				Expect(datum.StatusID).To(BeNil())
			})
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
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *reservoirchange.ReservoirChange) {},
				),
				Entry("type missing",
					func(datum *reservoirchange.ReservoirChange) { datum.Type = "" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &device.Meta{SubType: "reservoirChange"}),
				),
				Entry("type invalid",
					func(datum *reservoirchange.ReservoirChange) { datum.Type = "invalidType" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "deviceEvent"), "/type", &device.Meta{Type: "invalidType", SubType: "reservoirChange"}),
				),
				Entry("type device",
					func(datum *reservoirchange.ReservoirChange) { datum.Type = "deviceEvent" },
				),
				Entry("sub type missing",
					func(datum *reservoirchange.ReservoirChange) { datum.SubType = "" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/subType", &device.Meta{Type: "deviceEvent"}),
				),
				Entry("sub type invalid",
					func(datum *reservoirchange.ReservoirChange) { datum.SubType = "invalidSubType" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "reservoirChange"), "/subType", &device.Meta{Type: "deviceEvent", SubType: "invalidSubType"}),
				),
				Entry("sub type reservoir change",
					func(datum *reservoirchange.ReservoirChange) { datum.SubType = "reservoirChange" },
				),
				Entry("multiple errors",
					func(datum *reservoirchange.ReservoirChange) {
						datum.Type = "invalidType"
						datum.SubType = "invalidSubType"
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "deviceEvent"), "/type", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "reservoirChange"), "/subType", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
				),
			)

			DescribeTable("validates the datum with origin external",
				func(mutator func(datum *reservoirchange.ReservoirChange), expectedErrors ...error) {
					datum := NewReservoirChangeWithStatus()
					mutator(datum)
					testDataTypes.ValidateWithOrigin(datum, structure.OriginExternal, expectedErrors...)
				},
				Entry("succeeds",
					func(datum *reservoirchange.ReservoirChange) {},
				),
				Entry("status missing",
					func(datum *reservoirchange.ReservoirChange) { datum.Status = nil },
				),
				Entry("status valid",
					func(datum *reservoirchange.ReservoirChange) {
						datum.Status = data.DatumAsPointer(testDataTypesDeviceStatus.NewStatus())
					},
				),
				Entry("status id missing",
					func(datum *reservoirchange.ReservoirChange) { datum.StatusID = nil },
				),
				Entry("status id exists",
					func(datum *reservoirchange.ReservoirChange) { datum.StatusID = pointer.String(id.New()) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/statusId", NewMeta()),
				),
				Entry("multiple errors",
					func(datum *reservoirchange.ReservoirChange) {
						datum.StatusID = pointer.String(id.New())
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/statusId", NewMeta()),
				),
			)

			DescribeTable("validates the datum with origin internal/store",
				func(mutator func(datum *reservoirchange.ReservoirChange), expectedErrors ...error) {
					datum := NewReservoirChangeWithStatusID()
					mutator(datum)
					testDataTypes.ValidateWithOrigin(datum, structure.OriginInternal, expectedErrors...)
					testDataTypes.ValidateWithOrigin(datum, structure.OriginStore, expectedErrors...)
				},
				Entry("succeeds",
					func(datum *reservoirchange.ReservoirChange) {},
				),
				Entry("status missing",
					func(datum *reservoirchange.ReservoirChange) { datum.Status = nil },
				),
				Entry("status exists",
					func(datum *reservoirchange.ReservoirChange) {
						datum.Status = data.DatumAsPointer(testDataTypesDeviceStatus.NewStatus())
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/status", NewMeta()),
				),
				Entry("status id missing",
					func(datum *reservoirchange.ReservoirChange) { datum.StatusID = nil },
				),
				Entry("status id invalid",
					func(datum *reservoirchange.ReservoirChange) { datum.StatusID = pointer.String("invalid") },
					testErrors.WithPointerSourceAndMeta(id.ErrorValueStringAsIDNotValid("invalid"), "/statusId", NewMeta()),
				),
				Entry("status id valid",
					func(datum *reservoirchange.ReservoirChange) { datum.StatusID = pointer.String(id.New()) },
				),
				Entry("multiple errors",
					func(datum *reservoirchange.ReservoirChange) {
						datum.Status = data.DatumAsPointer(testDataTypesDeviceStatus.NewStatus())
						datum.StatusID = pointer.String("invalid")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/status", NewMeta()),
					testErrors.WithPointerSourceAndMeta(id.ErrorValueStringAsIDNotValid("invalid"), "/statusId", NewMeta()),
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
				datumStatus := testDataTypesDeviceStatus.NewStatus()
				datum := NewReservoirChangeWithStatusID()
				datum.Status = data.DatumAsPointer(datumStatus)
				expectedDatum := CloneReservoirChange(datum)
				normalizer := dataNormalizer.New()
				Expect(normalizer).ToNot(BeNil())
				datum.Normalize(normalizer)
				Expect(normalizer.Error()).To(BeNil())
				Expect(normalizer.Data()).To(Equal([]data.Datum{datumStatus}))
				expectedDatum.Status = nil
				expectedDatum.StatusID = pointer.String(*datumStatus.ID)
				Expect(datum).To(Equal(expectedDatum))
			})
		})
	})
})
