package timechange_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataTypesDevice "github.com/tidepool-org/platform/data/types/device"
	dataTypesDeviceTimechange "github.com/tidepool-org/platform/data/types/device/timechange"
	dataTypesDeviceTimechangeTest "github.com/tidepool-org/platform/data/types/device/timechange/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("TimeChange", func() {
	It("MethodAutomatic is expected", func() {
		Expect(dataTypesDeviceTimechange.MethodAutomatic).To(Equal("automatic"))
	})

	It("MethodManual is expected", func() {
		Expect(dataTypesDeviceTimechange.MethodManual).To(Equal("manual"))
	})

	It("SubType is expected", func() {
		Expect(dataTypesDeviceTimechange.SubType).To(Equal("timeChange"))
	})

	It("Methods returns expected", func() {
		Expect(dataTypesDeviceTimechange.Methods()).To(Equal([]string{"automatic", "manual"}))
	})

	Context("New", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := dataTypesDeviceTimechange.New()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("deviceEvent"))
			Expect(datum.SubType).To(Equal("timeChange"))
			Expect(datum.From).To(BeNil())
			Expect(datum.Method).To(BeNil())
			Expect(datum.To).To(BeNil())
			Expect(datum.Change).To(BeNil())
		})
	})

	Context("TimeChange", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			Context("non-deprecated", func() {
				DescribeTable("validates the datum",
					func(mutator func(datum *dataTypesDeviceTimechange.TimeChange), expectedErrors ...error) {
						datum := dataTypesDeviceTimechangeTest.RandomTimeChange(false)
						mutator(datum)
						dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
					},
					Entry("succeeds",
						func(datum *dataTypesDeviceTimechange.TimeChange) {},
					),
					Entry("type missing",
						func(datum *dataTypesDeviceTimechange.TimeChange) { datum.Type = "" },
						errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &dataTypesDevice.Meta{SubType: "timeChange"}),
					),
					Entry("type invalid",
						func(datum *dataTypesDeviceTimechange.TimeChange) { datum.Type = "invalidType" },
						errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "deviceEvent"), "/type", &dataTypesDevice.Meta{Type: "invalidType", SubType: "timeChange"}),
					),
					Entry("type device",
						func(datum *dataTypesDeviceTimechange.TimeChange) { datum.Type = "deviceEvent" },
					),
					Entry("sub type missing",
						func(datum *dataTypesDeviceTimechange.TimeChange) { datum.SubType = "" },
						errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/subType", &dataTypesDevice.Meta{Type: "deviceEvent"}),
					),
					Entry("sub type invalid",
						func(datum *dataTypesDeviceTimechange.TimeChange) { datum.SubType = "invalidSubType" },
						errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "timeChange"), "/subType", &dataTypesDevice.Meta{Type: "deviceEvent", SubType: "invalidSubType"}),
					),
					Entry("sub type timeChange",
						func(datum *dataTypesDeviceTimechange.TimeChange) { datum.SubType = "timeChange" },
					),
					Entry("from missing",
						func(datum *dataTypesDeviceTimechange.TimeChange) { datum.From = nil },
					),
					Entry("from invalid",
						func(datum *dataTypesDeviceTimechange.TimeChange) { datum.From.Time = pointer.FromTime(time.Time{}) },
						errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/from/time", dataTypesDeviceTimechangeTest.NewMeta()),
					),
					Entry("from valid",
						func(datum *dataTypesDeviceTimechange.TimeChange) {
							datum.From = dataTypesDeviceTimechangeTest.RandomInfo()
						},
					),
					Entry("method missing",
						func(datum *dataTypesDeviceTimechange.TimeChange) { datum.Method = nil },
					),
					Entry("method empty",
						func(datum *dataTypesDeviceTimechange.TimeChange) { datum.Method = pointer.FromString("") },
						errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("", []string{"automatic", "manual"}), "/method", dataTypesDeviceTimechangeTest.NewMeta()),
					),
					Entry("method invalid",
						func(datum *dataTypesDeviceTimechange.TimeChange) { datum.Method = pointer.FromString("invalid") },
						errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"automatic", "manual"}), "/method", dataTypesDeviceTimechangeTest.NewMeta()),
					),
					Entry("method automatic",
						func(datum *dataTypesDeviceTimechange.TimeChange) { datum.Method = pointer.FromString("automatic") },
					),
					Entry("method manual",
						func(datum *dataTypesDeviceTimechange.TimeChange) { datum.Method = pointer.FromString("manual") },
					),
					Entry("to missing",
						func(datum *dataTypesDeviceTimechange.TimeChange) { datum.To = nil },
						errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/to", dataTypesDeviceTimechangeTest.NewMeta()),
					),
					Entry("to invalid",
						func(datum *dataTypesDeviceTimechange.TimeChange) { datum.To.Time = pointer.FromTime(time.Time{}) },
						errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/to/time", dataTypesDeviceTimechangeTest.NewMeta()),
					),
					Entry("to valid",
						func(datum *dataTypesDeviceTimechange.TimeChange) {
							datum.To = dataTypesDeviceTimechangeTest.RandomInfo()
						},
					),
					Entry("multiple errors",
						func(datum *dataTypesDeviceTimechange.TimeChange) {
							datum.Type = "invalidType"
							datum.SubType = "invalidSubType"
							datum.From.Time = pointer.FromTime(time.Time{})
							datum.Method = pointer.FromString("")
							datum.To = nil
						},
						errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "deviceEvent"), "/type", &dataTypesDevice.Meta{Type: "invalidType", SubType: "invalidSubType"}),
						errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "timeChange"), "/subType", &dataTypesDevice.Meta{Type: "invalidType", SubType: "invalidSubType"}),
						errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/from/time", &dataTypesDevice.Meta{Type: "invalidType", SubType: "invalidSubType"}),
						errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("", []string{"automatic", "manual"}), "/method", &dataTypesDevice.Meta{Type: "invalidType", SubType: "invalidSubType"}),
						errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/to", &dataTypesDevice.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					),
				)
			})

			Context("deprecated", func() {
				DescribeTable("validates the datum",
					func(mutator func(datum *dataTypesDeviceTimechange.TimeChange), expectedErrors ...error) {
						datum := dataTypesDeviceTimechangeTest.RandomTimeChange(true)
						mutator(datum)
						dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
					},
					Entry("succeeds",
						func(datum *dataTypesDeviceTimechange.TimeChange) {},
					),
					Entry("type missing",
						func(datum *dataTypesDeviceTimechange.TimeChange) { datum.Type = "" },
						errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &dataTypesDevice.Meta{SubType: "timeChange"}),
					),
					Entry("type invalid",
						func(datum *dataTypesDeviceTimechange.TimeChange) { datum.Type = "invalidType" },
						errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "deviceEvent"), "/type", &dataTypesDevice.Meta{Type: "invalidType", SubType: "timeChange"}),
					),
					Entry("type device",
						func(datum *dataTypesDeviceTimechange.TimeChange) { datum.Type = "deviceEvent" },
					),
					Entry("sub type missing",
						func(datum *dataTypesDeviceTimechange.TimeChange) { datum.SubType = "" },
						errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/subType", &dataTypesDevice.Meta{Type: "deviceEvent"}),
					),
					Entry("sub type invalid",
						func(datum *dataTypesDeviceTimechange.TimeChange) { datum.SubType = "invalidSubType" },
						errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "timeChange"), "/subType", &dataTypesDevice.Meta{Type: "deviceEvent", SubType: "invalidSubType"}),
					),
					Entry("sub type timeChange",
						func(datum *dataTypesDeviceTimechange.TimeChange) { datum.SubType = "timeChange" },
					),
					Entry("change missing",
						func(datum *dataTypesDeviceTimechange.TimeChange) { datum.Change = nil },
						errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValuesNotExistForOne("to", "change"), "", dataTypesDeviceTimechangeTest.NewMeta()),
					),
					Entry("change invalid",
						func(datum *dataTypesDeviceTimechange.TimeChange) { datum.Change.Agent = nil },
						errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/change/agent", dataTypesDeviceTimechangeTest.NewMeta()),
					),
					Entry("change valid",
						func(datum *dataTypesDeviceTimechange.TimeChange) {
							datum.Change = dataTypesDeviceTimechangeTest.RandomChange()
						},
					),
					Entry("multiple errors",
						func(datum *dataTypesDeviceTimechange.TimeChange) {
							datum.Type = "invalidType"
							datum.SubType = "invalidSubType"
							datum.Change = nil
						},
						errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "deviceEvent"), "/type", &dataTypesDevice.Meta{Type: "invalidType", SubType: "invalidSubType"}),
						errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "timeChange"), "/subType", &dataTypesDevice.Meta{Type: "invalidType", SubType: "invalidSubType"}),
						errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValuesNotExistForOne("to", "change"), "", &dataTypesDevice.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					),
				)
			})
		})

		Context("Normalize", func() {
			normalizeValidations := func(deprecated bool) {
				DescribeTable("normalizes the datum",
					func(mutator func(datum *dataTypesDeviceTimechange.TimeChange)) {
						for _, origin := range structure.Origins() {
							datum := dataTypesDeviceTimechangeTest.RandomTimeChange(deprecated)
							mutator(datum)
							expectedDatum := dataTypesDeviceTimechangeTest.CloneTimeChange(datum)
							normalizer := dataNormalizer.New()
							Expect(normalizer).ToNot(BeNil())
							datum.Normalize(normalizer.WithOrigin(origin))
							Expect(normalizer.Error()).To(BeNil())
							Expect(normalizer.Data()).To(BeEmpty())
							Expect(datum).To(Equal(expectedDatum))
						}
					},
					Entry("does not modify the datum",
						func(datum *dataTypesDeviceTimechange.TimeChange) {},
					),
					Entry("does not modify the datum; from missing",
						func(datum *dataTypesDeviceTimechange.TimeChange) { datum.From = nil },
					),
					Entry("does not modify the datum; method missing",
						func(datum *dataTypesDeviceTimechange.TimeChange) { datum.Method = nil },
					),
					Entry("does not modify the datum; to missing",
						func(datum *dataTypesDeviceTimechange.TimeChange) { datum.To = nil },
					),
					Entry("does not modify the datum; change missing",
						func(datum *dataTypesDeviceTimechange.TimeChange) { datum.Change = nil },
					),
				)
			}

			Context("non-deprecated", func() {
				normalizeValidations(false)
			})

			Context("deprecated", func() {
				normalizeValidations(true)
			})
		})
	})
})
