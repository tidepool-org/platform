package pump_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataTypes "github.com/tidepool-org/platform/data/types"
	dataTypesStatusPump "github.com/tidepool-org/platform/data/types/status/pump"
	dataTypesStatusPumpTest "github.com/tidepool-org/platform/data/types/status/pump/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewMeta() interface{} {
	return &dataTypes.Meta{
		Type: "pumpStatus",
	}
}

var _ = Describe("Pump", func() {
	It("Type is expected", func() {
		Expect(dataTypesStatusPump.Type).To(Equal("pumpStatus"))
	})

	Context("Pump", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesStatusPump.Pump)) {
				datum := dataTypesStatusPumpTest.RandomPump()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataTypesStatusPumpTest.NewObjectFromPump(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataTypesStatusPumpTest.NewObjectFromPump(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataTypesStatusPump.Pump) {},
			),
			Entry("empty",
				func(datum *dataTypesStatusPump.Pump) {
					*datum = *dataTypesStatusPump.New()
				},
			),
			Entry("all",
				func(datum *dataTypesStatusPump.Pump) {
					datum.BasalDelivery = dataTypesStatusPumpTest.RandomBasalDelivery()
					datum.Battery = dataTypesStatusPumpTest.RandomBattery()
					datum.BolusDelivery = dataTypesStatusPumpTest.RandomBolusDelivery()
					datum.DeliveryIndeterminant = pointer.FromBool(test.RandomBool())
					datum.Reservoir = dataTypesStatusPumpTest.RandomReservoir()
				},
			),
		)

		Context("New", func() {
			It("returns the expected datum with all values initialized", func() {
				datum := dataTypesStatusPump.New()
				Expect(datum).ToNot(BeNil())
				Expect(datum.Type).To(Equal("pumpStatus"))
				Expect(datum.BasalDelivery).To(BeNil())
				Expect(datum.Battery).To(BeNil())
				Expect(datum.BolusDelivery).To(BeNil())
				Expect(datum.DeliveryIndeterminant).To(BeNil())
				Expect(datum.Reservoir).To(BeNil())
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dataTypesStatusPump.Pump), expectedErrors ...error) {
					expectedDatum := dataTypesStatusPumpTest.RandomPumpForParser()
					object := dataTypesStatusPumpTest.NewObjectFromPump(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := dataTypesStatusPump.New()
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dataTypesStatusPump.Pump) {},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *dataTypesStatusPump.Pump) {
						object["basalDelivery"] = true
						object["battery"] = true
						object["bolusDelivery"] = true
						object["deliveryIndeterminant"] = ""
						object["reservoir"] = true
						expectedDatum.BasalDelivery = nil
						expectedDatum.Battery = nil
						expectedDatum.BolusDelivery = nil
						expectedDatum.DeliveryIndeterminant = nil
						expectedDatum.Reservoir = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotObject(true), "/basalDelivery", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotObject(true), "/battery", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotObject(true), "/bolusDelivery", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotBool(""), "/deliveryIndeterminant", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotObject(true), "/reservoir", NewMeta()),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("return the expected results when the input",
				func(mutator func(datum *dataTypesStatusPump.Pump), expectedErrors ...error) {
					datum := dataTypesStatusPumpTest.RandomPump()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesStatusPump.Pump) {},
				),
				Entry("type missing",
					func(datum *dataTypesStatusPump.Pump) { datum.Type = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &dataTypes.Meta{}),
				),
				Entry("type invalid",
					func(datum *dataTypesStatusPump.Pump) { datum.Type = "invalidType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "pumpStatus"), "/type", &dataTypes.Meta{Type: "invalidType"}),
				),
				Entry("type pumpStatus",
					func(datum *dataTypesStatusPump.Pump) { datum.Type = "pumpStatus" },
				),
				Entry("basal delivery invalid",
					func(datum *dataTypesStatusPump.Pump) {
						datum.BasalDelivery.State = nil
						datum.BasalDelivery.Time = nil
						datum.BasalDelivery.Dose = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/basalDelivery/state", NewMeta()),
				),
				Entry("battery invalid",
					func(datum *dataTypesStatusPump.Pump) { datum.Battery.State = pointer.FromString("invalid") },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", dataTypesStatusPump.BatteryStates()), "/battery/state", NewMeta()),
				),
				Entry("bolus delivery invalid",
					func(datum *dataTypesStatusPump.Pump) {
						datum.BolusDelivery.State = nil
						datum.BolusDelivery.Dose = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bolusDelivery/state", NewMeta()),
				),
				Entry("reservoir invalid",
					func(datum *dataTypesStatusPump.Pump) { datum.Reservoir.Remaining = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/reservoir/remaining", NewMeta()),
				),
				Entry("one of required missing",
					func(datum *dataTypesStatusPump.Pump) {
						datum.BasalDelivery = nil
						datum.Battery = nil
						datum.BolusDelivery = nil
						datum.DeliveryIndeterminant = nil
						datum.Reservoir = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValuesNotExistForAny("basalDelivery", "battery", "bolusDelivery", "deliveryIndeterminant", "reservoir"), "", NewMeta()),
				),
				Entry("multiple errors",
					func(datum *dataTypesStatusPump.Pump) {
						datum.Type = "invalidType"
						datum.BasalDelivery.State = nil
						datum.BasalDelivery.Time = nil
						datum.BasalDelivery.Dose = nil
						datum.Battery.State = pointer.FromString("invalid")
						datum.BolusDelivery.State = nil
						datum.BolusDelivery.Dose = nil
						datum.Reservoir.Remaining = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "pumpStatus"), "/type", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/basalDelivery/state", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", dataTypesStatusPump.BatteryStates()), "/battery/state", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/bolusDelivery/state", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/reservoir/remaining", &dataTypes.Meta{Type: "invalidType"}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *dataTypesStatusPump.Pump), expectator func(datum *dataTypesStatusPump.Pump, expectedDatum *dataTypesStatusPump.Pump)) {
					for _, origin := range structure.Origins() {
						datum := dataTypesStatusPumpTest.RandomPump()
						mutator(datum)
						expectedDatum := dataTypesStatusPumpTest.ClonePump(datum)
						normalizer := dataNormalizer.New(logTest.NewLogger())
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						if expectator != nil {
							expectator(datum, expectedDatum)
						}
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *dataTypesStatusPump.Pump) {},
					nil,
				),
			)
		})
	})
})
