package continuous_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	dataBloodGlucoseTest "github.com/tidepool-org/platform/data/blood/glucose/test"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	dataTypesBloodGlucoseTest "github.com/tidepool-org/platform/data/types/blood/glucose/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/metadata"
	metadataTest "github.com/tidepool-org/platform/metadata/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewMeta() any {
	return &types.Meta{
		Type: "cbg",
	}
}

func RandomContinuous(units *string, rateUnits *string) *continuous.Continuous {
	datum := continuous.New()
	datum.Glucose = *dataTypesBloodGlucoseTest.NewGlucose(units)
	datum.Type = "cbg"
	datum.Trend = pointer.FromString(test.RandomStringFromArray(continuous.Trends()))
	datum.TrendRateUnits = rateUnits
	if datum.TrendRateUnits != nil {
		datum.TrendRate = pointer.FromFloat64(test.RandomFloat64FromRange(dataBloodGlucose.ValueRangeForRateUnits(datum.TrendRateUnits)))
	}
	datum.SampleInterval = pointer.FromInt(test.RandomIntFromRange(continuous.SampleIntervalMinimum, continuous.SampleIntervalMaximum))
	datum.Backfilled = pointer.FromBool(test.RandomBool())
	return datum
}

func CloneContinuous(datum *continuous.Continuous) *continuous.Continuous {
	if datum == nil {
		return nil
	}
	clone := continuous.New()
	clone.Glucose = *dataTypesBloodGlucoseTest.CloneGlucose(&datum.Glucose)
	clone.Trend = pointer.CloneString(datum.Trend)
	clone.TrendRateUnits = pointer.CloneString(datum.TrendRateUnits)
	clone.TrendRate = pointer.CloneFloat64(datum.TrendRate)
	clone.SampleInterval = pointer.CloneInt(datum.SampleInterval)
	clone.Backfilled = pointer.CloneBool(datum.Backfilled)
	return clone
}

var _ = Describe("Continuous", func() {
	It("Type is expected", func() {
		Expect(continuous.Type).To(Equal("cbg"))
	})

	Context("New", func() {
		It("returns the expected datum", func() {
			datum := continuous.New()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("cbg"))
			Expect(datum.Units).To(BeNil())
			Expect(datum.Value).To(BeNil())
			Expect(datum.Trend).To(BeNil())
			Expect(datum.TrendRateUnits).To(BeNil())
			Expect(datum.TrendRate).To(BeNil())
			Expect(datum.SampleInterval).To(BeNil())
			Expect(datum.Backfilled).To(BeNil())
		})
	})

	Context("Validate", func() {
		DescribeTable("validates the datum",
			func(units *string, rateUnits *string, mutator func(datum *continuous.Continuous, units *string, rateUnits *string), expectedErrors ...error) {
				datum := RandomContinuous(units, rateUnits)
				mutator(datum, units, rateUnits)
				dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
			},
			Entry("succeeds",
				pointer.FromString("mmol/L"),
				pointer.FromString("mmol/L/minute"),
				func(datum *continuous.Continuous, units *string, rateUnits *string) {},
			),
			Entry("type missing",
				pointer.FromString("mmol/L"),
				pointer.FromString("mmol/L/minute"),
				func(datum *continuous.Continuous, units *string, rateUnits *string) { datum.Type = "" },
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &types.Meta{}),
			),
			Entry("type invalid",
				pointer.FromString("mmol/L"),
				pointer.FromString("mmol/L/minute"),
				func(datum *continuous.Continuous, units *string, rateUnits *string) { datum.Type = "invalidType" },
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "cbg"), "/type", &types.Meta{Type: "invalidType"}),
			),
			Entry("type cbg",
				pointer.FromString("mmol/L"),
				pointer.FromString("mmol/L/minute"),
				func(datum *continuous.Continuous, units *string, rateUnits *string) { datum.Type = "cbg" },
			),
			Entry("units missing; value missing",
				nil,
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) { datum.Value = nil },
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
			),
			Entry("units missing; value out of range (lower)",
				nil,
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Value = pointer.FromFloat64(-0.1)
				},
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
			),
			Entry("units missing; value in range (lower)",
				nil,
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Value = pointer.FromFloat64(0.0)
				},
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
			),
			Entry("units missing; value in range (upper)",
				nil,
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Value = pointer.FromFloat64(55.0)
				},
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
			),
			Entry("units missing; value out of range (upper)",
				nil,
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Value = pointer.FromFloat64(1000.1)
				},
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
			),
			Entry("units invalid; value missing",
				pointer.FromString("invalid"),
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) { datum.Value = nil },
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
			),
			Entry("units invalid; value out of range (lower)",
				pointer.FromString("invalid"),
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Value = pointer.FromFloat64(-0.1)
				},
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
			),
			Entry("units invalid; value in range (lower)",
				pointer.FromString("invalid"),
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Value = pointer.FromFloat64(0.0)
				},
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
			),
			Entry("units invalid; value in range (upper)",
				pointer.FromString("invalid"),
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Value = pointer.FromFloat64(55.0)
				},
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
			),
			Entry("units invalid; value out of range (upper)",
				pointer.FromString("invalid"),
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Value = pointer.FromFloat64(1000.1)
				},
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
			),
			Entry("units mmol/L; value missing",
				pointer.FromString("mmol/L"),
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Value = nil
				},
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
			),
			Entry("units mmol/L; value out of range (lower)",
				pointer.FromString("mmol/L"),
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Value = pointer.FromFloat64(-0.1)
				},
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 55.0), "/value", NewMeta()),
			),
			Entry("units mmol/L; value in range (lower)",
				pointer.FromString("mmol/L"),
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Value = pointer.FromFloat64(0.0)
				},
			),
			Entry("units mmol/L; value in range (upper)",
				pointer.FromString("mmol/L"),
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Value = pointer.FromFloat64(55.0)
				},
			),
			Entry("units mmol/L; value out of range (upper)",
				pointer.FromString("mmol/L"),
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Value = pointer.FromFloat64(55.1)
				},
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(55.1, 0.0, 55.0), "/value", NewMeta()),
			),
			Entry("units mmol/l; value missing",
				pointer.FromString("mmol/l"),
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Value = nil
				},
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
			),
			Entry("units mmol/l; value out of range (lower)",
				pointer.FromString("mmol/l"),
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Value = pointer.FromFloat64(-0.1)
				},
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 55.0), "/value", NewMeta()),
			),
			Entry("units mmol/l; value in range (lower)",
				pointer.FromString("mmol/l"),
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Value = pointer.FromFloat64(0.0)
				},
			),
			Entry("units mmol/l; value in range (upper)",
				pointer.FromString("mmol/l"),
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Value = pointer.FromFloat64(55.0)
				},
			),
			Entry("units mmol/l; value out of range (upper)",
				pointer.FromString("mmol/l"),
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Value = pointer.FromFloat64(55.1)
				},
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(55.1, 0.0, 55.0), "/value", NewMeta()),
			),
			Entry("units mg/dL; value missing",
				pointer.FromString("mg/dL"),
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Value = nil
				},
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
			),
			Entry("units mg/dL; value out of range (lower)",
				pointer.FromString("mg/dL"),
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Value = pointer.FromFloat64(-0.1)
				},
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/value", NewMeta()),
			),
			Entry("units mg/dL; value in range (lower)",
				pointer.FromString("mg/dL"),
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Value = pointer.FromFloat64(0.0)
				},
			),
			Entry("units mg/dL; value in range (upper)",
				pointer.FromString("mg/dL"),
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Value = pointer.FromFloat64(1000.0)
				},
			),
			Entry("units mg/dL; value out of range (upper)",
				pointer.FromString("mg/dL"),
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Value = pointer.FromFloat64(1000.1)
				},
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/value", NewMeta()),
			),
			Entry("units mg/dl; value missing",
				pointer.FromString("mg/dl"),
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Value = nil
				},
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
			),
			Entry("units mg/dl; value out of range (lower)",
				pointer.FromString("mg/dl"),
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Value = pointer.FromFloat64(-0.1)
				},
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/value", NewMeta()),
			),
			Entry("units mg/dl; value in range (lower)",
				pointer.FromString("mg/dl"),
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Value = pointer.FromFloat64(0.0)
				},
			),
			Entry("units mg/dl; value in range (upper)",
				pointer.FromString("mg/dl"),
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Value = pointer.FromFloat64(1000.0)
				},
			),
			Entry("units mg/dl; value out of range (upper)",
				pointer.FromString("mg/dl"),
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Value = pointer.FromFloat64(1000.1)
				},
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/value", NewMeta()),
			),
			Entry("trend missing",
				pointer.FromString("mg/dl"),
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) { datum.Trend = nil },
			),
			Entry("trend empty",
				pointer.FromString("mg/dl"),
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Trend = pointer.FromString("")
				},
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("", continuous.Trends()), "/trend", NewMeta()),
			),
			Entry("trend invalid",
				pointer.FromString("mg/dl"),
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Trend = pointer.FromString("invalid")
				},
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", continuous.Trends()), "/trend", NewMeta()),
			),
			Entry("trend constant",
				pointer.FromString("mg/dl"),
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Trend = pointer.FromString("constant")
				},
			),
			Entry("trend slowFall",
				pointer.FromString("mg/dl"),
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Trend = pointer.FromString("slowFall")
				},
			),
			Entry("trend slowRise",
				pointer.FromString("mg/dl"),
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Trend = pointer.FromString("slowRise")
				},
			),
			Entry("trend moderateFall",
				pointer.FromString("mg/dl"),
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Trend = pointer.FromString("moderateFall")
				},
			),
			Entry("trend moderateRise",
				pointer.FromString("mg/dl"),
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Trend = pointer.FromString("moderateRise")
				},
			),
			Entry("trend rapidFall",
				pointer.FromString("mg/dl"),
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Trend = pointer.FromString("rapidFall")
				},
			),
			Entry("trend rapidRise",
				pointer.FromString("mg/dl"),
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Trend = pointer.FromString("rapidRise")
				},
			),
			Entry("trend rate units missing; trend rate missing",
				pointer.FromString("mmol/L"),
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) { datum.TrendRate = nil },
			),
			Entry("trend rate units invalid; trend rate missing",
				pointer.FromString("mmol/L"),
				pointer.FromString("invalid"),
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.TrendRate = nil
				},
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/trendRateUnits", NewMeta()),
			),
			Entry("trend rate units invalid; trend rate out of range (lower)",
				pointer.FromString("mmol/L"),
				pointer.FromString("invalid"),
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.TrendRate = pointer.FromFloat64(-5.51)
				},
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L/minute", "mg/dL/minute"}), "/trendRateUnits", NewMeta()),
			),
			Entry("trend rate units invalid; trend rate in range (lower)",
				pointer.FromString("mmol/L"),
				pointer.FromString("invalid"),
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.TrendRate = pointer.FromFloat64(-5.5)
				},
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L/minute", "mg/dL/minute"}), "/trendRateUnits", NewMeta()),
			),
			Entry("trend rate units invalid; trend rate in range (upper)",
				pointer.FromString("mmol/L"),
				pointer.FromString("invalid"),
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.TrendRate = pointer.FromFloat64(5.5)
				},
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L/minute", "mg/dL/minute"}), "/trendRateUnits", NewMeta()),
			),
			Entry("trend rate units invalid; trend rate out of range (upper)",
				pointer.FromString("mmol/L"),
				pointer.FromString("invalid"),
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.TrendRate = pointer.FromFloat64(5.51)
				},
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L/minute", "mg/dL/minute"}), "/trendRateUnits", NewMeta()),
			),
			Entry("trend rate units mmol/L/minute; trend rate missing",
				pointer.FromString("mmol/L"),
				pointer.FromString("mmol/L/minute"),
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.TrendRate = nil
				},
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/trendRateUnits", NewMeta()),
			),
			Entry("trend rate units mmol/L/minute; trend rate out of range (lower)",
				pointer.FromString("mmol/L"),
				pointer.FromString("mmol/L/minute"),
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.TrendRate = pointer.FromFloat64(-5.51)
				},
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-5.51, -5.5, 5.5), "/trendRate", NewMeta()),
			),
			Entry("trend rate units mmol/L/minute; trend rate in range (lower)",
				pointer.FromString("mmol/L"),
				pointer.FromString("mmol/L/minute"),
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.TrendRate = pointer.FromFloat64(-5.5)
				},
			),
			Entry("trend rate units mmol/L/minute; trend rate in range (upper)",
				pointer.FromString("mmol/L"),
				pointer.FromString("mmol/L/minute"),
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.TrendRate = pointer.FromFloat64(5.5)
				},
			),
			Entry("trend rate units mmol/L/minute; trend rate out of range (upper)",
				pointer.FromString("mmol/L"),
				pointer.FromString("mmol/L/minute"),
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.TrendRate = pointer.FromFloat64(5.51)
				},
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(5.51, -5.5, 5.5), "/trendRate", NewMeta()),
			),
			Entry("trend rate units mg/dL/minute; trend rate missing",
				pointer.FromString("mg/dL"),
				pointer.FromString("mg/dL/minute"),
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.TrendRate = nil
				},
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/trendRateUnits", NewMeta()),
			),
			Entry("trend rate units mg/dL/minute; trend rate out of range (lower)",
				pointer.FromString("mg/dL"),
				pointer.FromString("mg/dL/minute"),
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.TrendRate = pointer.FromFloat64(-100.1)
				},
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-100.1, -100.0, 100.0), "/trendRate", NewMeta()),
			),
			Entry("trend rate units mg/dL/minute; trend rate in range (lower)",
				pointer.FromString("mg/dL"),
				pointer.FromString("mg/dL/minute"),
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.TrendRate = pointer.FromFloat64(-100.0)
				},
			),
			Entry("trend rate units mg/dL/minute; trend rate in range (upper)",
				pointer.FromString("mg/dL"),
				pointer.FromString("mg/dL/minute"),
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.TrendRate = pointer.FromFloat64(100.0)
				},
			),
			Entry("trend rate units mg/dL/minute; trend rate out of range (upper)",
				pointer.FromString("mg/dL"),
				pointer.FromString("mg/dL/minute"),
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.TrendRate = pointer.FromFloat64(100.1)
				},
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, -100.0, 100.0), "/trendRate", NewMeta()),
			),
			Entry("sample interval missing",
				pointer.FromString("mg/dl"),
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) { datum.SampleInterval = nil },
			),
			Entry("sample interval out of range (lower)",
				pointer.FromString("mg/dl"),
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.SampleInterval = pointer.FromInt(-1)
				},
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/sampleInterval", NewMeta()),
			),
			Entry("sample interval in range (lower)",
				pointer.FromString("mg/dl"),
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.SampleInterval = pointer.FromInt(0)
				},
			),
			Entry("sample interval in range (upper)",
				pointer.FromString("mg/dl"),
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.SampleInterval = pointer.FromInt(86400000)
				},
			),
			Entry("sample interval out of range (upper)",
				pointer.FromString("mg/dl"),
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.SampleInterval = pointer.FromInt(86400001)
				},
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(86400001, 0, 86400000), "/sampleInterval", NewMeta()),
			),
			Entry("multiple errors",
				nil,
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Type = ""
					datum.Value = nil
					datum.Trend = nil
					datum.TrendRate = pointer.FromFloat64(-100.1)
					datum.SampleInterval = pointer.FromInt(-1)
				},
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &types.Meta{}),
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", &types.Meta{}),
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", &types.Meta{}),
				errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 86400000), "/sampleInterval", &types.Meta{}),
			),
		)
	})

	Context("Normalize", func() {
		DescribeTable("normalizes the datum",
			func(units *string, rateUnits *string, mutator func(datum *continuous.Continuous, units *string, rateUnits *string), expectator func(datum *continuous.Continuous, expectedDatum *continuous.Continuous, units *string, rateUnits *string)) {
				for _, origin := range structure.Origins() {
					datum := RandomContinuous(units, rateUnits)
					mutator(datum, units, rateUnits)
					expectedDatum := CloneContinuous(datum)
					normalizer := dataNormalizer.New(logTest.NewLogger())
					Expect(normalizer).ToNot(BeNil())
					datum.Normalize(normalizer.WithOrigin(origin))
					Expect(normalizer.Error()).To(BeNil())
					Expect(normalizer.Data()).To(BeEmpty())
					expectedDatum.Raw = metadataTest.CloneMetadata(datum.Raw)
					if expectator != nil {
						expectator(datum, expectedDatum, units, rateUnits)
					}
					Expect(datum).To(Equal(expectedDatum))
				}
			},
			Entry("does not modify the datum; units missing",
				nil,
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) {},
				nil,
			),
			Entry("does not modify the datum; units missing; value missing",
				nil,
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) { datum.Value = nil },
				nil,
			),
			Entry("does not modify the datum; units invalid",
				pointer.FromString("invalid"),
				pointer.FromString("invalid"),
				func(datum *continuous.Continuous, units *string, rateUnits *string) {},
				nil,
			),
			Entry("does not modify the datum; units invalid; value missing",
				pointer.FromString("invalid"),
				pointer.FromString("invalid"),
				func(datum *continuous.Continuous, units *string, rateUnits *string) { datum.Value = nil },
				nil,
			),
			Entry("does not modify the datum; trend missing",
				nil,
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) { datum.Trend = nil },
				nil,
			),
			Entry("does not modify the datum; trend rate units invalid",
				nil,
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.TrendRateUnits = pointer.FromString("invalid")
				},
				nil,
			),
			Entry("does not modify the datum; trend rate units invalid; trend rate missing",
				nil,
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.TrendRateUnits = pointer.FromString("invalid")
					datum.TrendRate = nil
				},
				nil,
			),
			Entry("does not modify the datum; trend rate missing",
				nil,
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) { datum.TrendRate = nil },
				nil,
			),
		)

		DescribeTable("normalizes the datum with origin external",
			func(units *string, rateUnits *string, mutator func(datum *continuous.Continuous, units *string, rateUnits *string), expectator func(datum *continuous.Continuous, expectedDatum *continuous.Continuous, units *string, value *float64, rateUnits *string)) {
				datum := RandomContinuous(units, rateUnits)
				mutator(datum, units, rateUnits)
				originalValue := pointer.CloneFloat64(datum.Value)
				expectedDatum := CloneContinuous(datum)
				normalizer := dataNormalizer.New(logTest.NewLogger())
				Expect(normalizer).ToNot(BeNil())
				datum.Normalize(normalizer.WithOrigin(structure.OriginExternal))
				Expect(normalizer.Error()).To(BeNil())
				Expect(normalizer.Data()).To(BeEmpty())
				expectedDatum.Raw = metadataTest.CloneMetadata(datum.Raw)
				if expectator != nil {
					expectator(datum, expectedDatum, units, originalValue, rateUnits)
				}
				Expect(datum).To(Equal(expectedDatum))
			},
			Entry("does not modify the datum; units invalid; value missing",
				pointer.FromString("invalid"),
				pointer.FromString("invalid"),
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Value = nil
					datum.TrendRate = nil
				},
				nil,
			),
			Entry("does not modify the datum; units mmol/L",
				pointer.FromString("mmol/L"),
				pointer.FromString("mmol/L/minute"),
				func(datum *continuous.Continuous, units *string, rateUnits *string) {},
				nil,
			),
			Entry("does not modify the datum; units mmol/L; value missing",
				pointer.FromString("mmol/L"),
				pointer.FromString("mmol/L/minute"),
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Value = nil
					datum.TrendRate = nil
				},
				nil,
			),
			Entry("modifies the datum; units mmol/l",
				pointer.FromString("mmol/l"),
				pointer.FromString("mmol/L/minute"),
				func(datum *continuous.Continuous, units *string, rateUnits *string) {},
				func(datum *continuous.Continuous, expectedDatum *continuous.Continuous, units *string, value *float64, rateUnits *string) {
					dataBloodGlucoseTest.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					dataBloodGlucoseTest.ExpectRaw(datum.Raw, &metadata.Metadata{"units": *units, "value": *value})
				},
			),
			Entry("modifies the datum; units mmol/l; value missing",
				pointer.FromString("mmol/l"),
				pointer.FromString("mmol/L/minute"),
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Value = nil
					datum.TrendRate = nil
				},
				func(datum *continuous.Continuous, expectedDatum *continuous.Continuous, units *string, value *float64, rateUnits *string) {
					dataBloodGlucoseTest.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					dataBloodGlucoseTest.ExpectRaw(datum.Raw, &metadata.Metadata{"units": *units, "value": nil})
				},
			),
			Entry("modifies the datum; units mg/dL",
				pointer.FromString("mg/dL"),
				pointer.FromString("mg/dL/minute"),
				func(datum *continuous.Continuous, units *string, rateUnits *string) {},
				func(datum *continuous.Continuous, expectedDatum *continuous.Continuous, units *string, value *float64, rateUnits *string) {
					dataBloodGlucoseTest.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					dataBloodGlucoseTest.ExpectNormalizedValue(datum.Value, expectedDatum.Value, units)
					dataBloodGlucoseTest.ExpectNormalizedRateUnits(datum.TrendRateUnits, expectedDatum.TrendRateUnits)
					dataBloodGlucoseTest.ExpectNormalizedRateValue(datum.TrendRate, expectedDatum.TrendRate, rateUnits)
					dataBloodGlucoseTest.ExpectRaw(datum.Raw, &metadata.Metadata{"units": *units, "value": *value})
				},
			),
			Entry("modifies the datum; units mg/dL; value missing",
				pointer.FromString("mg/dL"),
				pointer.FromString("mg/dL/minute"),
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Value = nil
					datum.TrendRate = nil
				},
				func(datum *continuous.Continuous, expectedDatum *continuous.Continuous, units *string, value *float64, rateUnits *string) {
					dataBloodGlucoseTest.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					dataBloodGlucoseTest.ExpectNormalizedRateUnits(datum.TrendRateUnits, expectedDatum.TrendRateUnits)
					dataBloodGlucoseTest.ExpectRaw(datum.Raw, &metadata.Metadata{"units": *units, "value": nil})
				},
			),
			Entry("modifies the datum; units mg/dl",
				pointer.FromString("mg/dl"),
				pointer.FromString("mg/dL/minute"),
				func(datum *continuous.Continuous, units *string, rateUnits *string) {},
				func(datum *continuous.Continuous, expectedDatum *continuous.Continuous, units *string, value *float64, rateUnits *string) {
					dataBloodGlucoseTest.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					dataBloodGlucoseTest.ExpectNormalizedValue(datum.Value, expectedDatum.Value, units)
					dataBloodGlucoseTest.ExpectNormalizedRateUnits(datum.TrendRateUnits, expectedDatum.TrendRateUnits)
					dataBloodGlucoseTest.ExpectNormalizedRateValue(datum.TrendRate, expectedDatum.TrendRate, rateUnits)
					dataBloodGlucoseTest.ExpectRaw(datum.Raw, &metadata.Metadata{"units": *units, "value": *value})
				},
			),
			Entry("modifies the datum; units mg/dl; value missing",
				pointer.FromString("mg/dl"),
				pointer.FromString("mg/dL/minute"),
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Value = nil
					datum.TrendRate = nil
				},
				func(datum *continuous.Continuous, expectedDatum *continuous.Continuous, units *string, value *float64, rateUnits *string) {
					dataBloodGlucoseTest.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					dataBloodGlucoseTest.ExpectNormalizedRateUnits(datum.TrendRateUnits, expectedDatum.TrendRateUnits)
					dataBloodGlucoseTest.ExpectRaw(datum.Raw, &metadata.Metadata{"units": *units, "value": nil})
				},
			),
		)

		DescribeTable("normalizes the datum with origin internal/store",
			func(units *string, rateUnits *string, mutator func(datum *continuous.Continuous, units *string, rateUnits *string), expectator func(datum *continuous.Continuous, expectedDatum *continuous.Continuous, units *string, rateUnits *string)) {
				for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
					datum := RandomContinuous(units, rateUnits)
					mutator(datum, units, rateUnits)
					expectedDatum := CloneContinuous(datum)
					normalizer := dataNormalizer.New(logTest.NewLogger())
					Expect(normalizer).ToNot(BeNil())
					datum.Normalize(normalizer.WithOrigin(origin))
					Expect(normalizer.Error()).To(BeNil())
					Expect(normalizer.Data()).To(BeEmpty())
					if expectator != nil {
						expectator(datum, expectedDatum, units, rateUnits)
					}
					Expect(datum).To(Equal(expectedDatum))
				}
			},
			Entry("does not modify the datum; units missing",
				nil,
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) {},
				nil,
			),
			Entry("does not modify the datum; units missing; value missing",
				nil,
				nil,
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Value = nil
					datum.TrendRate = nil
				},
				nil,
			),
			Entry("does not modify the datum; units invalid",
				pointer.FromString("invalid"),
				pointer.FromString("invalid"),
				func(datum *continuous.Continuous, units *string, rateUnits *string) {},
				nil,
			),
			Entry("does not modify the datum; units invalid; value missing",
				pointer.FromString("invalid"),
				pointer.FromString("invalid"),
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Value = nil
					datum.TrendRate = nil
				},
				nil,
			),
			Entry("does not modify the datum; units mmol/L",
				pointer.FromString("mmol/L"),
				pointer.FromString("mmol/L/minute"),
				func(datum *continuous.Continuous, units *string, rateUnits *string) {},
				nil,
			),
			Entry("does not modify the datum; units mmol/L; value missing",
				pointer.FromString("mmol/L"),
				pointer.FromString("mmol/L/minute"),
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Value = nil
					datum.TrendRate = nil
				},
				nil,
			),
			Entry("does not modify the datum; units mmol/l",
				pointer.FromString("mmol/l"),
				pointer.FromString("mmol/L/minute"),
				func(datum *continuous.Continuous, units *string, rateUnits *string) {},
				nil,
			),
			Entry("does not modify the datum; units mmol/l; value missing",
				pointer.FromString("mmol/l"),
				pointer.FromString("mmol/L/minute"),
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Value = nil
					datum.TrendRate = nil
				},
				nil,
			),
			Entry("does not modify the datum; units mg/dL",
				pointer.FromString("mg/dL"),
				pointer.FromString("mg/dL/minute"),
				func(datum *continuous.Continuous, units *string, rateUnits *string) {},
				nil,
			),
			Entry("does not modify the datum; units mg/dL; value missing",
				pointer.FromString("mg/dL"),
				pointer.FromString("mg/dL/minute"),
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Value = nil
					datum.TrendRate = nil
				},
				nil,
			),
			Entry("does not modify the datum; units mg/dl",
				pointer.FromString("mg/dl"),
				pointer.FromString("mg/dL/minute"),
				func(datum *continuous.Continuous, units *string, rateUnits *string) {},
				nil,
			),
			Entry("does not modify the datum; units mg/dl; value missing",
				pointer.FromString("mg/dl"),
				pointer.FromString("mg/dL/minute"),
				func(datum *continuous.Continuous, units *string, rateUnits *string) {
					datum.Value = nil
					datum.TrendRate = nil
				},
				nil,
			),
		)
	})
})
