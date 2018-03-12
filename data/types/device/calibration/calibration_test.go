package calibration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	testDataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose/test"
	"github.com/tidepool-org/platform/data/context"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/parser"
	testData "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/device"
	"github.com/tidepool-org/platform/data/types/device/calibration"
	testDataTypesDevice "github.com/tidepool-org/platform/data/types/device/test"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/service"
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
	datum.Device = *testDataTypesDevice.NewDevice()
	datum.SubType = "calibration"
	datum.Units = units
	datum.Value = pointer.Float64(test.RandomFloat64FromRange(dataBloodGlucose.ValueRangeForUnits(units)))
	return datum
}

func CloneCalibration(datum *calibration.Calibration) *calibration.Calibration {
	if datum == nil {
		return nil
	}
	clone := calibration.New()
	clone.Device = *testDataTypesDevice.CloneDevice(&datum.Device)
	clone.Units = test.CloneString(datum.Units)
	clone.Value = test.CloneFloat64(datum.Value)
	return clone
}

func NewTestCalibration(sourceTime interface{}, sourceUnits interface{}, sourceValue interface{}) *calibration.Calibration {
	datum := calibration.Init()
	datum.DeviceID = pointer.String(id.New())
	if val, ok := sourceTime.(string); ok {
		datum.Time = &val
	}
	if val, ok := sourceUnits.(string); ok {
		datum.Units = &val
	}
	if val, ok := sourceValue.(float64); ok {
		datum.Value = &val
	}
	return datum
}

var _ = Describe("Calibration", func() {
	It("SubType is expected", func() {
		Expect(calibration.SubType).To(Equal("calibration"))
	})

	Context("NewDatum", func() {
		It("returns the expected datum", func() {
			Expect(calibration.NewDatum()).To(Equal(&calibration.Calibration{}))
		})
	})

	Context("New", func() {
		It("returns the expected datum", func() {
			Expect(calibration.New()).To(Equal(&calibration.Calibration{}))
		})
	})

	Context("Init", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := calibration.Init()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("deviceEvent"))
			Expect(datum.SubType).To(Equal("calibration"))
			Expect(datum.Units).To(BeNil())
			Expect(datum.Value).To(BeNil())
		})
	})

	Context("with new datum", func() {
		var datum *calibration.Calibration

		BeforeEach(func() {
			datum = NewCalibration(pointer.String("mmol/L"))
		})

		Context("Init", func() {
			It("initializes the datum", func() {
				datum.Init()
				Expect(datum.Type).To(Equal("deviceEvent"))
				Expect(datum.SubType).To(Equal("calibration"))
				Expect(datum.Units).To(BeNil())
				Expect(datum.Value).To(BeNil())
			})
		})
	})

	Context("Calibration", func() {
		Context("Parse", func() {
			var datum *calibration.Calibration

			BeforeEach(func() {
				datum = calibration.Init()
				Expect(datum).ToNot(BeNil())
			})

			DescribeTable("parses the datum",
				func(sourceObject *map[string]interface{}, expectedDatum *calibration.Calibration, expectedErrors []*service.Error) {
					testContext, err := context.NewStandard(null.NewLogger())
					Expect(err).ToNot(HaveOccurred())
					Expect(testContext).ToNot(BeNil())
					testParser, err := parser.NewStandardObject(testContext, sourceObject, parser.AppendErrorNotParsed)
					Expect(err).ToNot(HaveOccurred())
					Expect(testParser).ToNot(BeNil())
					Expect(datum.Parse(testParser)).To(Succeed())
					Expect(datum.Time).To(Equal(expectedDatum.Time))
					Expect(datum.Units).To(Equal(expectedDatum.Units))
					Expect(datum.Value).To(Equal(expectedDatum.Value))
					Expect(testContext.Errors()).To(ConsistOf(expectedErrors))
				},
				Entry("parses object that is nil",
					nil,
					NewTestCalibration(nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that is empty",
					&map[string]interface{}{},
					NewTestCalibration(nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that has valid time",
					&map[string]interface{}{"time": "2016-09-06T13:45:58-07:00"},
					NewTestCalibration("2016-09-06T13:45:58-07:00", nil, nil),
					[]*service.Error{}),
				Entry("parses object that has invalid time",
					&map[string]interface{}{"time": 0},
					NewTestCalibration(nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotString(0), "/time", NewMeta()),
					}),
				Entry("parses object that has valid units",
					&map[string]interface{}{"units": "mmol/L"},
					NewTestCalibration(nil, "mmol/L", nil),
					[]*service.Error{}),
				Entry("parses object that has invalid units",
					&map[string]interface{}{"units": 123},
					NewTestCalibration(nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotString(123), "/units", NewMeta()),
					}),
				Entry("parses object that has valid value",
					&map[string]interface{}{"value": 9.4},
					NewTestCalibration(nil, nil, 9.4),
					[]*service.Error{}),
				Entry("parses object that has invalid value",
					&map[string]interface{}{"value": "invalid"},
					NewTestCalibration(nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/value", NewMeta()),
					}),
				Entry("parses object that has multiple valid fields",
					&map[string]interface{}{"time": "2016-09-06T13:45:58-07:00", "units": "mmol/L", "value": 9.4},
					NewTestCalibration("2016-09-06T13:45:58-07:00", "mmol/L", 9.4),
					[]*service.Error{}),
				Entry("parses object that has multiple invalid fields",
					&map[string]interface{}{"time": 0, "units": 123, "value": "invalid"},
					NewTestCalibration(nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotString(0), "/time", NewMeta()),
						testData.ComposeError(service.ErrorTypeNotString(123), "/units", NewMeta()),
						testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/value", NewMeta()),
					}),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(units *string, mutator func(datum *calibration.Calibration, units *string), expectedErrors ...error) {
					datum := NewCalibration(units)
					mutator(datum, units)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					pointer.String("mmol/L"),
					func(datum *calibration.Calibration, units *string) {},
				),
				Entry("type missing",
					pointer.String("mmol/L"),
					func(datum *calibration.Calibration, units *string) { datum.Type = "" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &device.Meta{SubType: "calibration"}),
				),
				Entry("type invalid",
					pointer.String("mmol/L"),
					func(datum *calibration.Calibration, units *string) { datum.Type = "invalidType" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "deviceEvent"), "/type", &device.Meta{Type: "invalidType", SubType: "calibration"}),
				),
				Entry("type device",
					pointer.String("mmol/L"),
					func(datum *calibration.Calibration, units *string) { datum.Type = "deviceEvent" },
				),
				Entry("sub type missing",
					pointer.String("mmol/L"),
					func(datum *calibration.Calibration, units *string) { datum.SubType = "" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/subType", &device.Meta{Type: "deviceEvent"}),
				),
				Entry("sub type invalid",
					pointer.String("mmol/L"),
					func(datum *calibration.Calibration, units *string) { datum.SubType = "invalidSubType" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "calibration"), "/subType", &device.Meta{Type: "deviceEvent", SubType: "invalidSubType"}),
				),
				Entry("sub type calibration",
					pointer.String("mmol/L"),
					func(datum *calibration.Calibration, units *string) { datum.SubType = "calibration" },
				),
				Entry("units missing; value missing",
					nil,
					func(datum *calibration.Calibration, units *string) { datum.Value = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
				),
				Entry("units missing; value out of range (lower)",
					nil,
					func(datum *calibration.Calibration, units *string) { datum.Value = pointer.Float64(-0.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; value in range (lower)",
					nil,
					func(datum *calibration.Calibration, units *string) { datum.Value = pointer.Float64(0.0) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; value in range (upper)",
					nil,
					func(datum *calibration.Calibration, units *string) { datum.Value = pointer.Float64(55.0) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; value out of range (upper)",
					nil,
					func(datum *calibration.Calibration, units *string) { datum.Value = pointer.Float64(1000.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units invalid; value missing",
					pointer.String("invalid"),
					func(datum *calibration.Calibration, units *string) { datum.Value = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
				),
				Entry("units invalid; value out of range (lower)",
					pointer.String("invalid"),
					func(datum *calibration.Calibration, units *string) { datum.Value = pointer.Float64(-0.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; value in range (lower)",
					pointer.String("invalid"),
					func(datum *calibration.Calibration, units *string) { datum.Value = pointer.Float64(0.0) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; value in range (upper)",
					pointer.String("invalid"),
					func(datum *calibration.Calibration, units *string) { datum.Value = pointer.Float64(55.0) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; value out of range (upper)",
					pointer.String("invalid"),
					func(datum *calibration.Calibration, units *string) { datum.Value = pointer.Float64(1000.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units mmol/L; value missing",
					pointer.String("mmol/L"),
					func(datum *calibration.Calibration, units *string) { datum.Value = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
				),
				Entry("units mmol/L; value out of range (lower)",
					pointer.String("mmol/L"),
					func(datum *calibration.Calibration, units *string) { datum.Value = pointer.Float64(-0.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 55.0), "/value", NewMeta()),
				),
				Entry("units mmol/L; value in range (lower)",
					pointer.String("mmol/L"),
					func(datum *calibration.Calibration, units *string) { datum.Value = pointer.Float64(0.0) },
				),
				Entry("units mmol/L; value in range (upper)",
					pointer.String("mmol/L"),
					func(datum *calibration.Calibration, units *string) { datum.Value = pointer.Float64(55.0) },
				),
				Entry("units mmol/L; value out of range (upper)",
					pointer.String("mmol/L"),
					func(datum *calibration.Calibration, units *string) { datum.Value = pointer.Float64(55.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(55.1, 0.0, 55.0), "/value", NewMeta()),
				),
				Entry("units mmol/l; value missing",
					pointer.String("mmol/l"),
					func(datum *calibration.Calibration, units *string) { datum.Value = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
				),
				Entry("units mmol/l; value out of range (lower)",
					pointer.String("mmol/l"),
					func(datum *calibration.Calibration, units *string) { datum.Value = pointer.Float64(-0.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 55.0), "/value", NewMeta()),
				),
				Entry("units mmol/l; value in range (lower)",
					pointer.String("mmol/l"),
					func(datum *calibration.Calibration, units *string) { datum.Value = pointer.Float64(0.0) },
				),
				Entry("units mmol/l; value in range (upper)",
					pointer.String("mmol/l"),
					func(datum *calibration.Calibration, units *string) { datum.Value = pointer.Float64(55.0) },
				),
				Entry("units mmol/l; value out of range (upper)",
					pointer.String("mmol/l"),
					func(datum *calibration.Calibration, units *string) { datum.Value = pointer.Float64(55.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(55.1, 0.0, 55.0), "/value", NewMeta()),
				),
				Entry("units mg/dL; value missing",
					pointer.String("mg/dL"),
					func(datum *calibration.Calibration, units *string) { datum.Value = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
				),
				Entry("units mg/dL; value out of range (lower)",
					pointer.String("mg/dL"),
					func(datum *calibration.Calibration, units *string) { datum.Value = pointer.Float64(-0.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/value", NewMeta()),
				),
				Entry("units mg/dL; value in range (lower)",
					pointer.String("mg/dL"),
					func(datum *calibration.Calibration, units *string) { datum.Value = pointer.Float64(0.0) },
				),
				Entry("units mg/dL; value in range (upper)",
					pointer.String("mg/dL"),
					func(datum *calibration.Calibration, units *string) { datum.Value = pointer.Float64(1000.0) },
				),
				Entry("units mg/dL; value out of range (upper)",
					pointer.String("mg/dL"),
					func(datum *calibration.Calibration, units *string) { datum.Value = pointer.Float64(1000.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/value", NewMeta()),
				),
				Entry("units mg/dl; value missing",
					pointer.String("mg/dl"),
					func(datum *calibration.Calibration, units *string) { datum.Value = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
				),
				Entry("units mg/dl; value out of range (lower)",
					pointer.String("mg/dl"),
					func(datum *calibration.Calibration, units *string) { datum.Value = pointer.Float64(-0.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/value", NewMeta()),
				),
				Entry("units mg/dl; value in range (lower)",
					pointer.String("mg/dl"),
					func(datum *calibration.Calibration, units *string) { datum.Value = pointer.Float64(0.0) },
				),
				Entry("units mg/dl; value in range (upper)",
					pointer.String("mg/dl"),
					func(datum *calibration.Calibration, units *string) { datum.Value = pointer.Float64(1000.0) },
				),
				Entry("units mg/dl; value out of range (upper)",
					pointer.String("mg/dl"),
					func(datum *calibration.Calibration, units *string) { datum.Value = pointer.Float64(1000.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/value", NewMeta()),
				),
				Entry("multiple errors",
					nil,
					func(datum *calibration.Calibration, units *string) {
						datum.Type = "invalidType"
						datum.SubType = "invalidSubType"
						datum.Value = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "deviceEvent"), "/type", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidSubType", "calibration"), "/subType", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", &device.Meta{Type: "invalidType", SubType: "invalidSubType"}),
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
					pointer.String("invalid"),
					func(datum *calibration.Calibration, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units invalid; value missing",
					pointer.String("invalid"),
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
					pointer.String("mmol/L"),
					func(datum *calibration.Calibration, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/L; value missing",
					pointer.String("mmol/L"),
					func(datum *calibration.Calibration, units *string) { datum.Value = nil },
					nil,
				),
				Entry("modifies the datum; units mmol/l",
					pointer.String("mmol/l"),
					func(datum *calibration.Calibration, units *string) {},
					func(datum *calibration.Calibration, expectedDatum *calibration.Calibration, units *string) {
						testDataBloodGlucose.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					},
				),
				Entry("modifies the datum; units mmol/l; value missing",
					pointer.String("mmol/l"),
					func(datum *calibration.Calibration, units *string) { datum.Value = nil },
					func(datum *calibration.Calibration, expectedDatum *calibration.Calibration, units *string) {
						testDataBloodGlucose.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					},
				),
				Entry("modifies the datum; units mg/dL",
					pointer.String("mg/dL"),
					func(datum *calibration.Calibration, units *string) {},
					func(datum *calibration.Calibration, expectedDatum *calibration.Calibration, units *string) {
						testDataBloodGlucose.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
						testDataBloodGlucose.ExpectNormalizedValue(datum.Value, expectedDatum.Value, units)
					},
				),
				Entry("modifies the datum; units mg/dL; value missing",
					pointer.String("mg/dL"),
					func(datum *calibration.Calibration, units *string) { datum.Value = nil },
					func(datum *calibration.Calibration, expectedDatum *calibration.Calibration, units *string) {
						testDataBloodGlucose.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					},
				),
				Entry("modifies the datum; units mg/dl",
					pointer.String("mg/dl"),
					func(datum *calibration.Calibration, units *string) {},
					func(datum *calibration.Calibration, expectedDatum *calibration.Calibration, units *string) {
						testDataBloodGlucose.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
						testDataBloodGlucose.ExpectNormalizedValue(datum.Value, expectedDatum.Value, units)
					},
				),
				Entry("modifies the datum; units mg/dl; value missing",
					pointer.String("mg/dl"),
					func(datum *calibration.Calibration, units *string) { datum.Value = nil },
					func(datum *calibration.Calibration, expectedDatum *calibration.Calibration, units *string) {
						testDataBloodGlucose.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
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
					pointer.String("mmol/L"),
					func(datum *calibration.Calibration, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/L; value missing",
					pointer.String("mmol/L"),
					func(datum *calibration.Calibration, units *string) { datum.Value = nil },
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.String("mmol/l"),
					func(datum *calibration.Calibration, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l; value missing",
					pointer.String("mmol/l"),
					func(datum *calibration.Calibration, units *string) { datum.Value = nil },
					nil,
				),
				Entry("does not modify the datum; units mg/dL",
					pointer.String("mg/dL"),
					func(datum *calibration.Calibration, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dL; value missing",
					pointer.String("mg/dL"),
					func(datum *calibration.Calibration, units *string) { datum.Value = nil },
					nil,
				),
				Entry("does not modify the datum; units mg/dl",
					pointer.String("mg/dl"),
					func(datum *calibration.Calibration, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dl; value missing",
					pointer.String("mg/dl"),
					func(datum *calibration.Calibration, units *string) { datum.Value = nil },
					nil,
				),
			)
		})
	})
})
