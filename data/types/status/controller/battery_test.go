package controller_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataTypesStatusController "github.com/tidepool-org/platform/data/types/status/controller"
	dataTypesStatusControllerTest "github.com/tidepool-org/platform/data/types/status/controller/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Battery", func() {
	It("BatteryRemainingPercentMaximum is expected", func() {
		Expect(dataTypesStatusController.BatteryRemainingPercentMaximum).To(Equal(1.0))
	})

	It("BatteryRemainingPercentMinimum is expected", func() {
		Expect(dataTypesStatusController.BatteryRemainingPercentMinimum).To(Equal(0.0))
	})

	It("BatteryStateCharging is expected", func() {
		Expect(dataTypesStatusController.BatteryStateCharging).To(Equal("charging"))
	})

	It("BatteryStateFull is expected", func() {
		Expect(dataTypesStatusController.BatteryStateFull).To(Equal("full"))
	})

	It("BatteryStateUnplugged is expected", func() {
		Expect(dataTypesStatusController.BatteryStateUnplugged).To(Equal("unplugged"))
	})

	It("BatteryUnitsPercent is expected", func() {
		Expect(dataTypesStatusController.BatteryUnitsPercent).To(Equal("percent"))
	})

	It("BatteryStates returns expected", func() {
		Expect(dataTypesStatusController.BatteryStates()).To(Equal([]string{"charging", "full", "unplugged"}))
	})

	It("BatteryUnits returns expected", func() {
		Expect(dataTypesStatusController.BatteryUnits()).To(Equal([]string{"percent"}))
	})

	Context("Battery", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesStatusController.Battery)) {
				datum := dataTypesStatusControllerTest.RandomBattery()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataTypesStatusControllerTest.NewObjectFromBattery(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataTypesStatusControllerTest.NewObjectFromBattery(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataTypesStatusController.Battery) {},
			),
			Entry("empty",
				func(datum *dataTypesStatusController.Battery) {
					*datum = *dataTypesStatusController.NewBattery()
				},
			),
			Entry("all",
				func(datum *dataTypesStatusController.Battery) {
					datum.Time = pointer.FromTime(test.RandomTime())
					datum.State = pointer.FromString(test.RandomStringFromArray(dataTypesStatusController.BatteryStates()))
					datum.Remaining = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesStatusController.BatteryRemainingPercentMinimum, dataTypesStatusController.BatteryRemainingPercentMaximum))
					datum.Units = pointer.FromString(test.RandomStringFromArray(dataTypesStatusController.BatteryUnits()))
				},
			),
		)

		Context("ParseBattery", func() {
			It("returns nil when the object is missing", func() {
				Expect(dataTypesStatusController.ParseBattery(structureParser.NewObject(nil))).To(BeNil())
			})

			It("returns new datum when the object is valid", func() {
				datum := dataTypesStatusControllerTest.RandomBattery()
				object := dataTypesStatusControllerTest.NewObjectFromBattery(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(&object)
				Expect(dataTypesStatusController.ParseBattery(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("NewBattery", func() {
			It("returns the expected datum with all values initialized", func() {
				datum := dataTypesStatusController.NewBattery()
				Expect(datum).ToNot(BeNil())
				Expect(datum.Time).To(BeNil())
				Expect(datum.State).To(BeNil())
				Expect(datum.Remaining).To(BeNil())
				Expect(datum.Units).To(BeNil())
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dataTypesStatusController.Battery), expectedErrors ...error) {
					expectedDatum := dataTypesStatusControllerTest.RandomBattery()
					object := dataTypesStatusControllerTest.NewObjectFromBattery(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := dataTypesStatusController.NewBattery()
					errorsTest.ExpectEqual(structureParser.NewObject(&object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dataTypesStatusController.Battery) {},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *dataTypesStatusController.Battery) {
						object["time"] = true
						object["state"] = true
						object["remaining"] = true
						object["units"] = true
						expectedDatum.Time = nil
						expectedDatum.State = nil
						expectedDatum.Remaining = nil
						expectedDatum.Units = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/time"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/state"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotFloat64(true), "/remaining"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/units"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("return the expected results when the input",
				func(mutator func(datum *dataTypesStatusController.Battery), expectedErrors ...error) {
					datum := dataTypesStatusControllerTest.RandomBattery()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesStatusController.Battery) {},
				),
				Entry("time missing",
					func(datum *dataTypesStatusController.Battery) { datum.Time = nil },
				),
				Entry("state missing",
					func(datum *dataTypesStatusController.Battery) { datum.State = nil },
				),
				Entry("state invalid",
					func(datum *dataTypesStatusController.Battery) {
						datum.State = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dataTypesStatusController.BatteryStates()), "/state"),
				),
				Entry("state charging",
					func(datum *dataTypesStatusController.Battery) {
						datum.State = pointer.FromString("charging")
					},
				),
				Entry("state full",
					func(datum *dataTypesStatusController.Battery) {
						datum.State = pointer.FromString("full")
					},
				),
				Entry("state unplugged",
					func(datum *dataTypesStatusController.Battery) {
						datum.State = pointer.FromString("unplugged")
					},
				),
				Entry("remaining missing; units missing",
					func(datum *dataTypesStatusController.Battery) {
						datum.Remaining = nil
						datum.Units = nil
					},
				),
				Entry("remaining missing; units exists",
					func(datum *dataTypesStatusController.Battery) {
						datum.Remaining = nil
						datum.Units = pointer.FromString("percent")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/units"),
				),
				Entry("remaining out of range (lower)",
					func(datum *dataTypesStatusController.Battery) {
						datum.Remaining = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0, 1.0), "/remaining"),
				),
				Entry("remaining in range (lower)",
					func(datum *dataTypesStatusController.Battery) {
						datum.Remaining = pointer.FromFloat64(0.0)
					},
				),
				Entry("remaining in range (upper)",
					func(datum *dataTypesStatusController.Battery) {
						datum.Remaining = pointer.FromFloat64(1.0)
					},
				),
				Entry("remaining out of range (upper)",
					func(datum *dataTypesStatusController.Battery) {
						datum.Remaining = pointer.FromFloat64(1.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(1.1, 0, 1.0), "/remaining"),
				),
				Entry("units missing",
					func(datum *dataTypesStatusController.Battery) { datum.Units = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/units"),
				),
				Entry("units invalid",
					func(datum *dataTypesStatusController.Battery) { datum.Units = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"percent"}), "/units"),
				),
				Entry("units percent",
					func(datum *dataTypesStatusController.Battery) { datum.Units = pointer.FromString("percent") },
				),
				Entry("one of required missing",
					func(datum *dataTypesStatusController.Battery) {
						datum.State = nil
						datum.Remaining = nil
						datum.Units = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValuesNotExistForAny("state", "remaining"), ""),
				),
				Entry("multiple errors",
					func(datum *dataTypesStatusController.Battery) {
						datum.State = pointer.FromString("invalid")
						datum.Remaining = nil
						datum.Units = pointer.FromString(dataTypesStatusController.BatteryUnitsPercent)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dataTypesStatusController.BatteryStates()), "/state"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/units"),
				),
			)
		})
	})
})
