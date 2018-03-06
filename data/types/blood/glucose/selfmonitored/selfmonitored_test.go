package selfmonitored_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	testDataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose/test"
	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/factory"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/parser"
	testData "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/blood/glucose/selfmonitored"
	testDataTypesBloodGlucose "github.com/tidepool-org/platform/data/types/blood/glucose/test"
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
	return &types.Meta{
		Type: "smbg",
	}
}

func NewSelfMonitored(units *string) *selfmonitored.SelfMonitored {
	datum := selfmonitored.New()
	datum.Glucose = *testDataTypesBloodGlucose.NewGlucose(units)
	datum.Type = "smbg"
	datum.SubType = pointer.String(test.RandomStringFromStringArray(selfmonitored.SubTypes()))
	return datum
}

func CloneSelfMonitored(datum *selfmonitored.SelfMonitored) *selfmonitored.SelfMonitored {
	if datum == nil {
		return nil
	}
	clone := selfmonitored.New()
	clone.Glucose = *testDataTypesBloodGlucose.CloneGlucose(&datum.Glucose)
	clone.SubType = test.CloneString(datum.SubType)
	return clone
}

func NewTestSelfMonitored(sourceTime interface{}, sourceUnits interface{}, sourceValue interface{}, sourceSubType interface{}) *selfmonitored.SelfMonitored {
	datum := selfmonitored.Init()
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
	if val, ok := sourceSubType.(string); ok {
		datum.SubType = &val
	}
	return datum
}

var _ = Describe("SelfMonitored", func() {
	It("Type is expected", func() {
		Expect(selfmonitored.Type).To(Equal("smbg"))
	})

	It("SubTypeLinked is expected", func() {
		Expect(selfmonitored.SubTypeLinked).To(Equal("linked"))
	})

	It("SubTypeManual is expected", func() {
		Expect(selfmonitored.SubTypeManual).To(Equal("manual"))
	})

	It("SubTypes returns expected", func() {
		Expect(selfmonitored.SubTypes()).To(Equal([]string{"linked", "manual"}))
	})

	Context("NewDatum", func() {
		It("returns the expected datum", func() {
			Expect(selfmonitored.NewDatum()).To(Equal(&selfmonitored.SelfMonitored{}))
		})
	})

	Context("New", func() {
		It("returns the expected datum", func() {
			Expect(selfmonitored.New()).To(Equal(&selfmonitored.SelfMonitored{}))
		})
	})

	Context("Init", func() {
		It("returns the expected datum", func() {
			datum := selfmonitored.Init()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("smbg"))
			Expect(datum.Units).To(BeNil())
			Expect(datum.Value).To(BeNil())
			Expect(datum.SubType).To(BeNil())
		})
	})

	Context("with new datum", func() {
		var datum *selfmonitored.SelfMonitored

		BeforeEach(func() {
			datum = NewSelfMonitored(pointer.String("mmol/L"))
		})

		Context("Init", func() {
			It("initializes the datum", func() {
				datum.Init()
				Expect(datum.Type).To(Equal("smbg"))
				Expect(datum.Units).To(BeNil())
				Expect(datum.Value).To(BeNil())
				Expect(datum.SubType).To(BeNil())
			})
		})
	})

	Context("SelfMonitored", func() {
		Context("Parse", func() {
			var datum *selfmonitored.SelfMonitored

			BeforeEach(func() {
				datum = selfmonitored.Init()
				Expect(datum).ToNot(BeNil())
			})

			DescribeTable("parses the datum",
				func(sourceObject *map[string]interface{}, expectedDatum *selfmonitored.SelfMonitored, expectedErrors []*service.Error) {
					testContext, err := context.NewStandard(null.NewLogger())
					Expect(err).ToNot(HaveOccurred())
					Expect(testContext).ToNot(BeNil())
					testFactory, err := factory.NewStandard()
					Expect(err).ToNot(HaveOccurred())
					Expect(testFactory).ToNot(BeNil())
					testParser, err := parser.NewStandardObject(testContext, testFactory, sourceObject, parser.AppendErrorNotParsed)
					Expect(err).ToNot(HaveOccurred())
					Expect(testParser).ToNot(BeNil())
					Expect(datum.Parse(testParser)).To(Succeed())
					Expect(datum.Time).To(Equal(expectedDatum.Time))
					Expect(datum.Units).To(Equal(expectedDatum.Units))
					Expect(datum.Value).To(Equal(expectedDatum.Value))
					Expect(datum.SubType).To(Equal(expectedDatum.SubType))
					Expect(testContext.Errors()).To(ConsistOf(expectedErrors))
				},
				Entry("parses object that is nil",
					nil,
					NewTestSelfMonitored(nil, nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that is empty",
					&map[string]interface{}{},
					NewTestSelfMonitored(nil, nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that has valid time",
					&map[string]interface{}{"time": "2016-09-06T13:45:58-07:00"},
					NewTestSelfMonitored("2016-09-06T13:45:58-07:00", nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that has invalid time",
					&map[string]interface{}{"time": 0},
					NewTestSelfMonitored(nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotString(0), "/time", NewMeta()),
					}),
				Entry("parses object that has valid units",
					&map[string]interface{}{"units": "mmol/L"},
					NewTestSelfMonitored(nil, "mmol/L", nil, nil),
					[]*service.Error{}),
				Entry("parses object that has invalid units",
					&map[string]interface{}{"units": 0},
					NewTestSelfMonitored(nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotString(0), "/units", NewMeta()),
					}),
				Entry("parses object that has valid value",
					&map[string]interface{}{"value": 10.0},
					NewTestSelfMonitored(nil, nil, 10.0, nil),
					[]*service.Error{}),
				Entry("parses object that has invalid value",
					&map[string]interface{}{"value": "invalid"},
					NewTestSelfMonitored(nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/value", NewMeta()),
					}),
				Entry("parses object that has valid sub type",
					&map[string]interface{}{"subType": "linked"},
					NewTestSelfMonitored(nil, nil, nil, "linked"),
					[]*service.Error{}),
				Entry("parses object that has invalid sub type",
					&map[string]interface{}{"subType": 0},
					NewTestSelfMonitored(nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotString(0), "/subType", NewMeta()),
					}),
				Entry("parses object that has multiple valid fields",
					&map[string]interface{}{"time": "2016-09-06T13:45:58-07:00", "units": "mmol/L", "value": 10.0, "subType": "linked"},
					NewTestSelfMonitored("2016-09-06T13:45:58-07:00", "mmol/L", 10.0, "linked"),
					[]*service.Error{}),
				Entry("parses object that has multiple invalid fields",
					&map[string]interface{}{"time": 0, "units": 0, "value": "invalid", "subType": 0},
					NewTestSelfMonitored(nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotString(0), "/time", NewMeta()),
						testData.ComposeError(service.ErrorTypeNotString(0), "/units", NewMeta()),
						testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/value", NewMeta()),
						testData.ComposeError(service.ErrorTypeNotString(0), "/subType", NewMeta()),
					}),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(units *string, mutator func(datum *selfmonitored.SelfMonitored, units *string), expectedErrors ...error) {
					datum := NewSelfMonitored(units)
					mutator(datum, units)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					pointer.String("mmol/L"),
					func(datum *selfmonitored.SelfMonitored, units *string) {},
				),
				Entry("type missing",
					pointer.String("mmol/L"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Type = "" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &types.Meta{}),
				),
				Entry("type invalid",
					pointer.String("mmol/L"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Type = "invalidType" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "smbg"), "/type", &types.Meta{Type: "invalidType"}),
				),
				Entry("type smbg",
					pointer.String("mmol/L"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Type = "smbg" },
				),
				Entry("units missing; value missing",
					nil,
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
				),
				Entry("units missing; value out of range (lower)",
					nil,
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = pointer.Float64(-0.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; value in range (lower)",
					nil,
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = pointer.Float64(0.0) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; value in range (upper)",
					nil,
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = pointer.Float64(55.0) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; value out of range (upper)",
					nil,
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = pointer.Float64(1000.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units invalid; value missing",
					pointer.String("invalid"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
				),
				Entry("units invalid; value out of range (lower)",
					pointer.String("invalid"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = pointer.Float64(-0.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; value in range (lower)",
					pointer.String("invalid"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = pointer.Float64(0.0) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; value in range (upper)",
					pointer.String("invalid"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = pointer.Float64(55.0) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; value out of range (upper)",
					pointer.String("invalid"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = pointer.Float64(1000.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units mmol/L; value missing",
					pointer.String("mmol/L"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
				),
				Entry("units mmol/L; value out of range (lower)",
					pointer.String("mmol/L"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = pointer.Float64(-0.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 55.0), "/value", NewMeta()),
				),
				Entry("units mmol/L; value in range (lower)",
					pointer.String("mmol/L"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = pointer.Float64(0.0) },
				),
				Entry("units mmol/L; value in range (upper)",
					pointer.String("mmol/L"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = pointer.Float64(55.0) },
				),
				Entry("units mmol/L; value out of range (upper)",
					pointer.String("mmol/L"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = pointer.Float64(55.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(55.1, 0.0, 55.0), "/value", NewMeta()),
				),
				Entry("units mmol/l; value missing",
					pointer.String("mmol/l"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
				),
				Entry("units mmol/l; value out of range (lower)",
					pointer.String("mmol/l"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = pointer.Float64(-0.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 55.0), "/value", NewMeta()),
				),
				Entry("units mmol/l; value in range (lower)",
					pointer.String("mmol/l"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = pointer.Float64(0.0) },
				),
				Entry("units mmol/l; value in range (upper)",
					pointer.String("mmol/l"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = pointer.Float64(55.0) },
				),
				Entry("units mmol/l; value out of range (upper)",
					pointer.String("mmol/l"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = pointer.Float64(55.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(55.1, 0.0, 55.0), "/value", NewMeta()),
				),
				Entry("units mg/dL; value missing",
					pointer.String("mg/dL"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
				),
				Entry("units mg/dL; value out of range (lower)",
					pointer.String("mg/dL"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = pointer.Float64(-0.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/value", NewMeta()),
				),
				Entry("units mg/dL; value in range (lower)",
					pointer.String("mg/dL"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = pointer.Float64(0.0) },
				),
				Entry("units mg/dL; value in range (upper)",
					pointer.String("mg/dL"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = pointer.Float64(1000.0) },
				),
				Entry("units mg/dL; value out of range (upper)",
					pointer.String("mg/dL"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = pointer.Float64(1000.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/value", NewMeta()),
				),
				Entry("units mg/dl; value missing",
					pointer.String("mg/dl"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
				),
				Entry("units mg/dl; value out of range (lower)",
					pointer.String("mg/dL"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = pointer.Float64(-0.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/value", NewMeta()),
				),
				Entry("units mg/dl; value in range (lower)",
					pointer.String("mg/dL"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = pointer.Float64(0.0) },
				),
				Entry("units mg/dl; value in range (upper)",
					pointer.String("mg/dL"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = pointer.Float64(1000.0) },
				),
				Entry("units mg/dl; value out of range (upper)",
					pointer.String("mg/dL"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = pointer.Float64(1000.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/value", NewMeta()),
				),
				Entry("sub type missing",
					pointer.String("mmol/L"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.SubType = nil },
				),
				Entry("sub type invalid",
					pointer.String("mmol/L"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.SubType = pointer.String("invalid") },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"linked", "manual"}), "/subType", NewMeta()),
				),
				Entry("sub type linked",
					pointer.String("mmol/L"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.SubType = pointer.String("linked") },
				),
				Entry("sub type manual",
					pointer.String("mmol/L"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.SubType = pointer.String("manual") },
				),
				Entry("multiple errors",
					nil,
					func(datum *selfmonitored.SelfMonitored, units *string) {
						datum.Type = ""
						datum.Value = nil
						datum.SubType = pointer.String("invalid")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &types.Meta{}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", &types.Meta{}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", &types.Meta{}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"linked", "manual"}), "/subType", &types.Meta{}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(units *string, mutator func(datum *selfmonitored.SelfMonitored, units *string), expectator func(datum *selfmonitored.SelfMonitored, expectedDatum *selfmonitored.SelfMonitored, units *string)) {
					for _, origin := range structure.Origins() {
						datum := NewSelfMonitored(units)
						mutator(datum, units)
						expectedDatum := CloneSelfMonitored(datum)
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
					func(datum *selfmonitored.SelfMonitored, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units missing; value missing",
					nil,
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = nil },
					nil,
				),
				Entry("does not modify the datum; units invalid",
					pointer.String("invalid"),
					func(datum *selfmonitored.SelfMonitored, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units invalid; value missing",
					pointer.String("invalid"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = nil },
					nil,
				),
				Entry("does not modify the datum; sub type missing",
					pointer.String("mmol/L"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.SubType = nil },
					nil,
				),
				Entry("does not modify the datum; sub type invalid",
					pointer.String("mmol/L"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.SubType = pointer.String("invalid") },
					nil,
				),
				Entry("does not modify the datum; sub type linked",
					pointer.String("mmol/L"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.SubType = pointer.String("linked") },
					nil,
				),
				Entry("does not modify the datum; sub type manual",
					pointer.String("mmol/L"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.SubType = pointer.String("manual") },
					nil,
				),
			)

			DescribeTable("normalizes the datum with origin external",
				func(units *string, mutator func(datum *selfmonitored.SelfMonitored, units *string), expectator func(datum *selfmonitored.SelfMonitored, expectedDatum *selfmonitored.SelfMonitored, units *string)) {
					datum := NewSelfMonitored(units)
					mutator(datum, units)
					expectedDatum := CloneSelfMonitored(datum)
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
					func(datum *selfmonitored.SelfMonitored, units *string) {},
					func(datum *selfmonitored.SelfMonitored, expectedDatum *selfmonitored.SelfMonitored, units *string) {
						testDataBloodGlucose.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
						testDataBloodGlucose.ExpectNormalizedValue(datum.Value, expectedDatum.Value, units)
					},
				),
				Entry("does not modify the datum; units mmol/L; value missing",
					pointer.String("mmol/L"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = nil },
					func(datum *selfmonitored.SelfMonitored, expectedDatum *selfmonitored.SelfMonitored, units *string) {
						testDataBloodGlucose.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					},
				),
				Entry("modifies the datum; units mmol/l",
					pointer.String("mmol/l"),
					func(datum *selfmonitored.SelfMonitored, units *string) {},
					func(datum *selfmonitored.SelfMonitored, expectedDatum *selfmonitored.SelfMonitored, units *string) {
						testDataBloodGlucose.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
						testDataBloodGlucose.ExpectNormalizedValue(datum.Value, expectedDatum.Value, units)
					},
				),
				Entry("modifies the datum; units mmol/l; value missing",
					pointer.String("mmol/l"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = nil },
					func(datum *selfmonitored.SelfMonitored, expectedDatum *selfmonitored.SelfMonitored, units *string) {
						testDataBloodGlucose.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					},
				),
				Entry("modifies the datum; units mg/dL",
					pointer.String("mg/dL"),
					func(datum *selfmonitored.SelfMonitored, units *string) {},
					func(datum *selfmonitored.SelfMonitored, expectedDatum *selfmonitored.SelfMonitored, units *string) {
						testDataBloodGlucose.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
						testDataBloodGlucose.ExpectNormalizedValue(datum.Value, expectedDatum.Value, units)
					},
				),
				Entry("modifies the datum; units mg/dL; value missing",
					pointer.String("mg/dL"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = nil },
					func(datum *selfmonitored.SelfMonitored, expectedDatum *selfmonitored.SelfMonitored, units *string) {
						testDataBloodGlucose.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					},
				),
				Entry("modifies the datum; units mg/dl",
					pointer.String("mg/dl"),
					func(datum *selfmonitored.SelfMonitored, units *string) {},
					func(datum *selfmonitored.SelfMonitored, expectedDatum *selfmonitored.SelfMonitored, units *string) {
						testDataBloodGlucose.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
						testDataBloodGlucose.ExpectNormalizedValue(datum.Value, expectedDatum.Value, units)
					},
				),
				Entry("modifies the datum; units mg/dl; value missing",
					pointer.String("mg/dl"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = nil },
					func(datum *selfmonitored.SelfMonitored, expectedDatum *selfmonitored.SelfMonitored, units *string) {
						testDataBloodGlucose.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					},
				),
			)

			DescribeTable("normalizes the datum with origin internal/store",
				func(units *string, mutator func(datum *selfmonitored.SelfMonitored, units *string), expectator func(datum *selfmonitored.SelfMonitored, expectedDatum *selfmonitored.SelfMonitored, units *string)) {
					for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
						datum := NewSelfMonitored(units)
						mutator(datum, units)
						expectedDatum := CloneSelfMonitored(datum)
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
					func(datum *selfmonitored.SelfMonitored, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/L; value missing",
					pointer.String("mmol/L"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = nil },
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.String("mmol/l"),
					func(datum *selfmonitored.SelfMonitored, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l; value missing",
					pointer.String("mmol/l"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = nil },
					nil,
				),
				Entry("does not modify the datum; units mg/dL",
					pointer.String("mg/dL"),
					func(datum *selfmonitored.SelfMonitored, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dL; value missing",
					pointer.String("mg/dL"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = nil },
					nil,
				),
				Entry("does not modify the datum; units mg/dl",
					pointer.String("mg/dl"),
					func(datum *selfmonitored.SelfMonitored, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dl; value missing",
					pointer.String("mg/dl"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = nil },
					nil,
				),
			)
		})
	})
})
