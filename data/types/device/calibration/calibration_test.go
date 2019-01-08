package calibration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	dataBloodGlucoseTest "github.com/tidepool-org/platform/data/blood/glucose/test"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/device"
	"github.com/tidepool-org/platform/data/types/device/calibration"
	dataTypesDeviceTest "github.com/tidepool-org/platform/data/types/device/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewMeta() interface{} {
	return &device.Meta{
		Type:    "deviceEvent",
		SubType: "calibration",
	}
}

func NewCalibration(units *string) *calibration.Calibration {
	datum := calibration.New()
	datum.Device = *dataTypesDeviceTest.NewDevice()
	datum.SubType = "calibration"
	datum.Units = units
	datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(dataBloodGlucose.ValueRangeForUnits(units)))
	return datum
}

func CloneCalibration(datum *calibration.Calibration) *calibration.Calibration {
	if datum == nil {
		return nil
	}
	clone := calibration.New()
	clone.Device = *dataTypesDeviceTest.CloneDevice(&datum.Device)
	clone.Units = test.CloneString(datum.Units)
	clone.Value = test.CloneFloat64(datum.Value)
	return clone
}

var _ = Describe("Calibration", func() {
	It("SubType is expected", func() {
		Expect(calibration.SubType).To(Equal("calibration"))
	})

	Context("New", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := calibration.New()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("deviceEvent"))
			Expect(datum.SubType).To(Equal("calibration"))
			Expect(datum.Units).To(BeNil())
			Expect(datum.Value).To(BeNil())
		})
	})

	Context("Calibration", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(units *string, mutator func(datum *calibration.Calibration, units *string), expectedErrors ...error) {
					datum := NewCalibration(units)
					mutator(datum, units)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					pointer.FromString("mmol/L"),
					func(datum *calibration.Calibration, units *string) {},
				),
				Entry("type missing",
					pointer.FromString("mmol/L"),
					func(datum *calibration.Calibration, units *string) { datum.Type = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &device.Meta{SubType: "calibration"}),
				),
				Entry("type invalid",
					pointer.FromString("mmol/L"),
					func(datum *calibration.Calibration, units *string) { datum.Type = "invalidType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "deviceEvent"), "/type", &device.Meta{Type: "invalidType", SubType: "calibration"}),
				),
				Entry("type device",
					pointer.FromString("mmol/L"),
					func(datum *calibration.Calibration, units *string) { datum.Type = "deviceEvent" },
				),
				Entry("sub type missing",
					pointer.FromString("mmol/L"),
					func(datum *calibration.Calibration, units *string) { datum.SubType = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/subType", &device.Meta{Type: "deviceEvent"}),
				),
				Entry("sub type invalid",
					pointer.FromString("mmol/L"),
					func(datum *calibration.Calibration, units *string) { datum.SubType = "invalidSubType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "calibration"), "/subType", &device.Meta{Type: "deviceEvent", SubType: "invalidSubType"}),
				),
				Entry("sub type calibration",
					pointer.FromString("mmol/L"),
					func(datum *calibration.Calibration, units *string) { datum.SubType = "calibration" },
				),
				Entry("units missing; value missing",
					nil,
					func(datum *calibration.Calibration, units *string) { datum.Value = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
				),
				Entry("units missing; value out of range (lower)",
					nil,
					func(datum *calibration.Calibration, units *string) { datum.Value = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; value in range (lower)",
					nil,
					func(datum *calibration.Calibration, units *string) { datum.Value = pointer.FromFloat64(0.0) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; value in range (upper)",
					nil,
					func(datum *calibration.Calibration, units *string) { datum.Value = pointer.FromFloat64(55.0) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; value out of range (upper)",
					nil,
					func(datum *calibration.Calibration, units *string) { datum.Value = pointer.FromFloat64(1000.1) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units invalid; value missing",
					pointer.FromString("invalid"),
					func(datum *calibration.Calibration, units *string) { datum.Value = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
				),
				Entry("units invalid; value out of range (lower)",
					pointer.FromString("invalid"),
					func(datum *calibration.Calibration, units *string) { datum.Value = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; value in range (lower)",
					pointer.FromString("invalid"),
					func(datum *calibration.Calibration, units *string) { datum.Value = pointer.FromFloat64(0.0) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; value in range (upper)",
					pointer.FromString("invalid"),
					func(datum *calibration.Calibration, units *string) { datum.Value = pointer.FromFloat64(55.0) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; value out of range (upper)",
					pointer.FromString("invalid"),
					func(datum *calibration.Calibration, units *string) { datum.Value = pointer.FromFloat64(1000.1) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units mmol/L; value missing",
					pointer.FromString("mmol/L"),
					func(datum *calibration.Calibration, units *string) { datum.Value = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
				),
				Entry("units mmol/L; value out of range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *calibration.Calibration, units *string) { datum.Value = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 55.0), "/value", NewMeta()),
				),
				Entry("units mmol/L; value in range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *calibration.Calibration, units *string) { datum.Value = pointer.FromFloat64(0.0) },
				),
				Entry("units mmol/L; value in range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *calibration.Calibration, units *string) { datum.Value = pointer.FromFloat64(55.0) },
				),
				Entry("units mmol/L; value out of range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *calibration.Calibration, units *string) { datum.Value = pointer.FromFloat64(55.1) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(55.1, 0.0, 55.0), "/value", NewMeta()),
				),
				Entry("units mmol/l; value missing",
					pointer.FromString("mmol/l"),
					func(datum *calibration.Calibration, units *string) { datum.Value = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
				),
				Entry("units mmol/l; value out of range (lower)",
					pointer.FromString("mmol/l"),
					func(datum *calibration.Calibration, units *string) { datum.Value = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 55.0), "/value", NewMeta()),
				),
				Entry("units mmol/l; value in range (lower)",
					pointer.FromString("mmol/l"),
					func(datum *calibration.Calibration, units *string) { datum.Value = pointer.FromFloat64(0.0) },
				),
				Entry("units mmol/l; value in range (upper)",
					pointer.FromString("mmol/l"),
					func(datum *calibration.Calibration, units *string) { datum.Value = pointer.FromFloat64(55.0) },
				),
				Entry("units mmol/l; value out of range (upper)",
					pointer.FromString("mmol/l"),
					func(datum *calibration.Calibration, units *string) { datum.Value = pointer.FromFloat64(55.1) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(55.1, 0.0, 55.0), "/value", NewMeta()),
				),
				Entry("units mg/dL; value missing",
					pointer.FromString("mg/dL"),
					func(datum *calibration.Calibration, units *string) { datum.Value = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
				),
				Entry("units mg/dL; value out of range (lower)",
					pointer.FromString("mg/dL"),
					func(datum *calibration.Calibration, units *string) { datum.Value = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/value", NewMeta()),
				),
				Entry("units mg/dL; value in range (lower)",
					pointer.FromString("mg/dL"),
					func(datum *calibration.Calibration, units *string) { datum.Value = pointer.FromFloat64(0.0) },
				),
				Entry("units mg/dL; value in range (upper)",
					pointer.FromString("mg/dL"),
					func(datum *calibration.Calibration, units *string) { datum.Value = pointer.FromFloat64(1000.0) },
				),
				Entry("units mg/dL; value out of range (upper)",
					pointer.FromString("mg/dL"),
					func(datum *calibration.Calibration, units *string) { datum.Value = pointer.FromFloat64(1000.1) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/value", NewMeta()),
				),
				Entry("units mg/dl; value missing",
					pointer.FromString("mg/dl"),
					func(datum *calibration.Calibration, units *string) { datum.Value = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
				),
				Entry("units mg/dl; value out of range (lower)",
					pointer.FromString("mg/dl"),
					func(datum *calibration.Calibration, units *string) { datum.Value = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/value", NewMeta()),
				),
				Entry("units mg/dl; value in range (lower)",
					pointer.FromString("mg/dl"),
					func(datum *calibration.Calibration, units *string) { datum.Value = pointer.FromFloat64(0.0) },
				),
				Entry("units mg/dl; value in range (upper)",
					pointer.FromString("mg/dl"),
					func(datum *calibration.Calibration, units *string) { datum.Value = pointer.FromFloat64(1000.0) },
				),
				Entry("units mg/dl; value out of range (upper)",
					pointer.FromString("mg/dl"),
					func(datum *calibration.Calibration, units *string) { datum.Value = pointer.FromFloat64(1000.1) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/value", NewMeta()),
				),
				Entry("multiple errors",
					nil,
					func(datum *calibration.Calibration, units *string) {
						datum.Type = "invalidType"
						datum.SubType = "invalidSubType"
						datum.Value = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "deviceEvent"), "/type", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "calibration"), "/subType", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(units *string, mutator func(datum *calibration.Calibration, units *string), expectator func(datum *calibration.Calibration, expectedDatum *calibration.Calibration, units *string)) {
					for _, origin := range structure.Origins() {
						datum := NewCalibration(units)
						mutator(datum, units)
						expectedDatum := CloneCalibration(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						if expectator != nil {
							expectator(datum, expectedDatum, units)
						}
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum; units missing",
					nil,
					func(datum *calibration.Calibration, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units missing; value missing",
					nil,
					func(datum *calibration.Calibration, units *string) { datum.Value = nil },
					nil,
				),
				Entry("does not modify the datum; units invalid",
					pointer.FromString("invalid"),
					func(datum *calibration.Calibration, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units invalid; value missing",
					pointer.FromString("invalid"),
					func(datum *calibration.Calibration, units *string) { datum.Value = nil },
					nil,
				),
			)

			DescribeTable("normalizes the datum with origin external",
				func(units *string, mutator func(datum *calibration.Calibration, units *string), expectator func(datum *calibration.Calibration, expectedDatum *calibration.Calibration, units *string)) {
					datum := NewCalibration(units)
					mutator(datum, units)
					expectedDatum := CloneCalibration(datum)
					normalizer := dataNormalizer.New()
					Expect(normalizer).ToNot(BeNil())
					datum.Normalize(normalizer.WithOrigin(structure.OriginExternal))
					Expect(normalizer.Error()).To(BeNil())
					Expect(normalizer.Data()).To(BeEmpty())
					if expectator != nil {
						expectator(datum, expectedDatum, units)
					}
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("does not modify the datum; units mmol/L",
					pointer.FromString("mmol/L"),
					func(datum *calibration.Calibration, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/L; value missing",
					pointer.FromString("mmol/L"),
					func(datum *calibration.Calibration, units *string) { datum.Value = nil },
					nil,
				),
				Entry("modifies the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *calibration.Calibration, units *string) {},
					func(datum *calibration.Calibration, expectedDatum *calibration.Calibration, units *string) {
						dataBloodGlucoseTest.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					},
				),
				Entry("modifies the datum; units mmol/l; value missing",
					pointer.FromString("mmol/l"),
					func(datum *calibration.Calibration, units *string) { datum.Value = nil },
					func(datum *calibration.Calibration, expectedDatum *calibration.Calibration, units *string) {
						dataBloodGlucoseTest.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					},
				),
				Entry("modifies the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *calibration.Calibration, units *string) {},
					func(datum *calibration.Calibration, expectedDatum *calibration.Calibration, units *string) {
						dataBloodGlucoseTest.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
						dataBloodGlucoseTest.ExpectNormalizedValue(datum.Value, expectedDatum.Value, units)
					},
				),
				Entry("modifies the datum; units mg/dL; value missing",
					pointer.FromString("mg/dL"),
					func(datum *calibration.Calibration, units *string) { datum.Value = nil },
					func(datum *calibration.Calibration, expectedDatum *calibration.Calibration, units *string) {
						dataBloodGlucoseTest.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					},
				),
				Entry("modifies the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *calibration.Calibration, units *string) {},
					func(datum *calibration.Calibration, expectedDatum *calibration.Calibration, units *string) {
						dataBloodGlucoseTest.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
						dataBloodGlucoseTest.ExpectNormalizedValue(datum.Value, expectedDatum.Value, units)
					},
				),
				Entry("modifies the datum; units mg/dl; value missing",
					pointer.FromString("mg/dl"),
					func(datum *calibration.Calibration, units *string) { datum.Value = nil },
					func(datum *calibration.Calibration, expectedDatum *calibration.Calibration, units *string) {
						dataBloodGlucoseTest.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					},
				),
			)

			DescribeTable("normalizes the datum with origin internal/store",
				func(units *string, mutator func(datum *calibration.Calibration, units *string), expectator func(datum *calibration.Calibration, expectedDatum *calibration.Calibration, units *string)) {
					for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
						datum := NewCalibration(units)
						mutator(datum, units)
						expectedDatum := CloneCalibration(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						if expectator != nil {
							expectator(datum, expectedDatum, units)
						}
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum; units mmol/L",
					pointer.FromString("mmol/L"),
					func(datum *calibration.Calibration, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/L; value missing",
					pointer.FromString("mmol/L"),
					func(datum *calibration.Calibration, units *string) { datum.Value = nil },
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *calibration.Calibration, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l; value missing",
					pointer.FromString("mmol/l"),
					func(datum *calibration.Calibration, units *string) { datum.Value = nil },
					nil,
				),
				Entry("does not modify the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *calibration.Calibration, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dL; value missing",
					pointer.FromString("mg/dL"),
					func(datum *calibration.Calibration, units *string) { datum.Value = nil },
					nil,
				),
				Entry("does not modify the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *calibration.Calibration, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dl; value missing",
					pointer.FromString("mg/dl"),
					func(datum *calibration.Calibration, units *string) { datum.Value = nil },
					nil,
				),
			)
		})
	})
})
