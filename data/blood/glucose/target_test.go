package glucose_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"math"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/factory"
	"github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/parser"
	"github.com/tidepool-org/platform/data/types/testing"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/service"
)

func AsStringPointer(source interface{}) *string {
	if sourceString, ok := source.(string); ok {
		return app.StringAsPointer(sourceString)
	}
	return nil
}

func NewTestTarget(sourceTarget interface{}, sourceRange interface{}, sourceLow interface{}, sourceHigh interface{}) *glucose.Target {
	testTarget := &glucose.Target{}
	if value, ok := sourceTarget.(float64); ok {
		testTarget.Target = app.FloatAsPointer(value)
	}
	if value, ok := sourceRange.(float64); ok {
		testTarget.Range = app.FloatAsPointer(value)
	}
	if value, ok := sourceLow.(float64); ok {
		testTarget.Low = app.FloatAsPointer(value)
	}
	if value, ok := sourceHigh.(float64); ok {
		testTarget.High = app.FloatAsPointer(value)
	}
	return testTarget
}

var _ = Describe("Target", func() {
	DescribeTable("ParseTarget",
		func(sourceObject *map[string]interface{}, expectedTarget *glucose.Target, expectedErrors []*service.Error) {
			testContext, err := context.NewStandard(log.NewNull())
			Expect(err).ToNot(HaveOccurred())
			Expect(testContext).ToNot(BeNil())
			testFactory, err := factory.NewStandard()
			Expect(err).ToNot(HaveOccurred())
			Expect(testFactory).ToNot(BeNil())
			testParser, err := parser.NewStandardObject(testContext, testFactory, sourceObject, parser.AppendErrorNotParsed)
			Expect(err).ToNot(HaveOccurred())
			Expect(testParser).ToNot(BeNil())
			Expect(glucose.ParseTarget(testParser)).To(Equal(expectedTarget))
			Expect(testContext.Errors()).To(ConsistOf(expectedErrors))
		},
		Entry("parses object that is nil", nil, nil, []*service.Error{}),
		Entry("parses object that is empty", &map[string]interface{}{}, NewTestTarget(nil, nil, nil, nil), []*service.Error{}),
		Entry("parses object that has multiple valid fields", &map[string]interface{}{"target": 120.0, "range": 10.0, "low": 110.0, "high": 130.0},
			NewTestTarget(120.0, 10.0, 110.0, 130.0), []*service.Error{}),
		Entry("parses object that has multiple invalid fields", &map[string]interface{}{"target": "invalid", "range": "invalid", "low": "invalid", "high": "invalid"},
			NewTestTarget(nil, nil, nil, nil), []*service.Error{
				testing.ComposeError(service.ErrorTypeNotFloat("invalid"), "/target", nil),
				testing.ComposeError(service.ErrorTypeNotFloat("invalid"), "/range", nil),
				testing.ComposeError(service.ErrorTypeNotFloat("invalid"), "/low", nil),
				testing.ComposeError(service.ErrorTypeNotFloat("invalid"), "/high", nil),
			}),
		Entry("parses object that has additional fields", &map[string]interface{}{"target": 120.0, "range": 10.0, "low": 110.0, "high": 130.0, "additional": 0.0},
			NewTestTarget(120.0, 10.0, 110.0, 130.0), []*service.Error{
				testing.ComposeError(parser.ErrorNotParsed(), "/additional", nil),
			}),
	)

	Context("NewTarget", func() {
		It("is successful", func() {
			Expect(glucose.NewTarget()).To(Equal(&glucose.Target{}))
		})
	})

	Context("with new target", func() {
		DescribeTable("Parse",
			func(sourceObject *map[string]interface{}, expectedTarget *glucose.Target, expectedErrors []*service.Error) {
				testContext, err := context.NewStandard(log.NewNull())
				Expect(err).ToNot(HaveOccurred())
				Expect(testContext).ToNot(BeNil())
				testFactory, err := factory.NewStandard()
				Expect(err).ToNot(HaveOccurred())
				Expect(testFactory).ToNot(BeNil())
				testParser, err := parser.NewStandardObject(testContext, testFactory, sourceObject, parser.AppendErrorNotParsed)
				Expect(err).ToNot(HaveOccurred())
				Expect(testParser).ToNot(BeNil())
				sourceTarget := &glucose.Target{}
				sourceTarget.Parse(testParser)
				Expect(sourceTarget).To(Equal(expectedTarget))
				Expect(testContext.Errors()).To(ConsistOf(expectedErrors))
			},
			Entry("parses object that is nil", nil, NewTestTarget(nil, nil, nil, nil), []*service.Error{}),
			Entry("parses object that is empty", &map[string]interface{}{}, NewTestTarget(nil, nil, nil, nil), []*service.Error{}),
			Entry("parses object that has valid target", &map[string]interface{}{"target": 120.0}, NewTestTarget(120.0, nil, nil, nil), []*service.Error{}),
			Entry("parses object that has invalid target", &map[string]interface{}{"target": "invalid"}, NewTestTarget(nil, nil, nil, nil), []*service.Error{
				testing.ComposeError(service.ErrorTypeNotFloat("invalid"), "/target", nil),
			}),
			Entry("parses object that has valid range", &map[string]interface{}{"range": 10.0}, NewTestTarget(nil, 10.0, nil, nil), []*service.Error{}),
			Entry("parses object that has invalid range", &map[string]interface{}{"range": "invalid"}, NewTestTarget(nil, nil, nil, nil), []*service.Error{
				testing.ComposeError(service.ErrorTypeNotFloat("invalid"), "/range", nil),
			}),
			Entry("parses object that has valid low", &map[string]interface{}{"low": 110.0}, NewTestTarget(nil, nil, 110.0, nil), []*service.Error{}),
			Entry("parses object that has invalid low", &map[string]interface{}{"low": "invalid"}, NewTestTarget(nil, nil, nil, nil), []*service.Error{
				testing.ComposeError(service.ErrorTypeNotFloat("invalid"), "/low", nil),
			}),
			Entry("parses object that has valid high", &map[string]interface{}{"high": 130.0}, NewTestTarget(nil, nil, nil, 130.0), []*service.Error{}),
			Entry("parses object that has invalid high", &map[string]interface{}{"high": "invalid"}, NewTestTarget(nil, nil, nil, nil), []*service.Error{
				testing.ComposeError(service.ErrorTypeNotFloat("invalid"), "/high", nil),
			}),
			Entry("parses object that has multiple valid fields", &map[string]interface{}{"target": 120.0, "range": 10.0, "low": 110.0, "high": 130.0},
				NewTestTarget(120.0, 10.0, 110.0, 130.0), []*service.Error{}),
			Entry("parses object that has multiple invalid fields", &map[string]interface{}{"target": "invalid", "range": "invalid", "low": "invalid", "high": "invalid"},
				NewTestTarget(nil, nil, nil, nil), []*service.Error{
					testing.ComposeError(service.ErrorTypeNotFloat("invalid"), "/target", nil),
					testing.ComposeError(service.ErrorTypeNotFloat("invalid"), "/range", nil),
					testing.ComposeError(service.ErrorTypeNotFloat("invalid"), "/low", nil),
					testing.ComposeError(service.ErrorTypeNotFloat("invalid"), "/high", nil),
				}),
		)

		DescribeTable("Validate",
			func(sourceTarget *glucose.Target, sourceUnits interface{}, expectedErrors []*service.Error) {
				testContext, err := context.NewStandard(log.NewNull())
				Expect(err).ToNot(HaveOccurred())
				Expect(testContext).ToNot(BeNil())
				testValidator, err := validator.NewStandard(testContext)
				Expect(err).ToNot(HaveOccurred())
				Expect(testValidator).ToNot(BeNil())
				sourceTarget.Validate(testValidator, AsStringPointer(sourceUnits))
				Expect(testContext.Errors()).To(ConsistOf(expectedErrors))
			},
			Entry("validates a target with units of mmol/L; target/range; all valid", NewTestTarget(6.6, 1.0, nil, nil), "mmol/L", []*service.Error{}),
			Entry("validates a target with units of mmol/L; target/range; target out of range", NewTestTarget(55.1, 1.0, nil, nil), "mmol/L", []*service.Error{
				testing.ComposeError(service.ErrorValueNotInRange(55.1, 0, 55), "/target", nil),
			}),
			Entry("validates a target with units of mmol/L; target/range; range out of range", NewTestTarget(6.6, 6.7, nil, nil), "mmol/L", []*service.Error{
				testing.ComposeError(service.ErrorValueNotInRange(6.7, 0, 6.6), "/range", nil),
			}),
			Entry("validates a target with units of mmol/L; target/range; low exists", NewTestTarget(6.6, 1.0, 5.6, nil), "mmol/L", []*service.Error{
				testing.ComposeError(service.ErrorValueExists(), "/low", nil),
			}),
			Entry("validates a target with units of mmol/L; target/range; high exists", NewTestTarget(6.6, 1.0, nil, 7.6), "mmol/L", []*service.Error{
				testing.ComposeError(service.ErrorValueExists(), "/high", nil),
			}),
			Entry("validates a target with units of mmol/L; target/range; multiple", NewTestTarget(6.6, 6.7, 5.6, 7.6), "mmol/L", []*service.Error{
				testing.ComposeError(service.ErrorValueNotInRange(6.7, 0, 6.6), "/range", nil),
				testing.ComposeError(service.ErrorValueExists(), "/low", nil),
				testing.ComposeError(service.ErrorValueExists(), "/high", nil),
			}),
			Entry("validates a target with units of mmol/L; target/high; all valid", NewTestTarget(6.6, nil, nil, 7.6), "mmol/L", []*service.Error{}),
			Entry("validates a target with units of mmol/L; target/high; target out of range", NewTestTarget(55.1, nil, nil, 7.6), "mmol/L", []*service.Error{
				testing.ComposeError(service.ErrorValueNotInRange(55.1, 0, 55), "/target", nil),
			}),
			Entry("validates a target with units of mmol/L; target/high; low exists", NewTestTarget(6.6, nil, 5.6, 7.6), "mmol/L", []*service.Error{
				testing.ComposeError(service.ErrorValueExists(), "/low", nil),
			}),
			Entry("validates a target with units of mmol/L; target/high; high out of range (lower)", NewTestTarget(6.6, nil, nil, 6.5), "mmol/L", []*service.Error{
				testing.ComposeError(service.ErrorValueNotInRange(6.5, 6.6, 55), "/high", nil),
			}),
			Entry("validates a target with units of mmol/L; target/high; high out of range (upper)", NewTestTarget(6.6, nil, nil, 55.1), "mmol/L", []*service.Error{
				testing.ComposeError(service.ErrorValueNotInRange(55.1, 6.6, 55), "/high", nil),
			}),
			Entry("validates a target with units of mmol/L; target/high; multiple", NewTestTarget(6.6, nil, 5.6, 55.1), "mmol/L", []*service.Error{
				testing.ComposeError(service.ErrorValueExists(), "/low", nil),
				testing.ComposeError(service.ErrorValueNotInRange(55.1, 6.6, 55), "/high", nil),
			}),
			Entry("validates a target with units of mmol/L; target; all valid", NewTestTarget(6.6, nil, nil, nil), "mmol/L", []*service.Error{}),
			Entry("validates a target with units of mmol/L; target; target out of range", NewTestTarget(55.1, nil, nil, nil), "mmol/L", []*service.Error{
				testing.ComposeError(service.ErrorValueNotInRange(55.1, 0, 55), "/target", nil),
			}),
			Entry("validates a target with units of mmol/L; target; low exists", NewTestTarget(6.6, nil, 5.6, nil), "mmol/L", []*service.Error{
				testing.ComposeError(service.ErrorValueExists(), "/low", nil),
			}),
			Entry("validates a target with units of mmol/L; target; multiple", NewTestTarget(55.1, nil, 5.6, nil), "mmol/L", []*service.Error{
				testing.ComposeError(service.ErrorValueNotInRange(55.1, 0, 55), "/target", nil),
				testing.ComposeError(service.ErrorValueExists(), "/low", nil),
			}),
			Entry("validates a target with units of mmol/L; low/high; all valid", NewTestTarget(nil, nil, 5.6, 7.6), "mmol/L", []*service.Error{}),
			Entry("validates a target with units of mmol/L; low/high; low out of range", NewTestTarget(nil, nil, -0.1, 7.6), "mmol/L", []*service.Error{
				testing.ComposeError(service.ErrorValueNotInRange(-0.1, 0, 55), "/low", nil),
			}),
			Entry("validates a target with units of mmol/L; low/high; high out of range (lower)", NewTestTarget(nil, nil, 5.6, 5.5), "mmol/L", []*service.Error{
				testing.ComposeError(service.ErrorValueNotInRange(5.5, 5.6, 55), "/high", nil),
			}),
			Entry("validates a target with units of mmol/L; low/high; high out of range (upper)", NewTestTarget(nil, nil, 5.6, 55.1), "mmol/L", []*service.Error{
				testing.ComposeError(service.ErrorValueNotInRange(55.1, 5.6, 55), "/high", nil),
			}),
			Entry("validates a target with units of mmol/L; low", NewTestTarget(nil, nil, 5.6, nil), "mmol/L", []*service.Error{
				testing.ComposeError(service.ErrorValueNotExists(), "/high", nil),
			}),
			Entry("validates a target with units of mmol/L; none", NewTestTarget(nil, nil, nil, nil), nil, []*service.Error{
				testing.ComposeError(service.ErrorValueNotExists(), "/target", nil),
			}),
			Entry("validates a target with units of mg/dL; target/range; all valid", NewTestTarget(120.0, 10.0, nil, nil), "mg/dL", []*service.Error{}),
			Entry("validates a target with units of mg/dL; target/range; target out of range", NewTestTarget(1001.0, 10.0, nil, nil), "mg/dL", []*service.Error{
				testing.ComposeError(service.ErrorValueNotInRange(1001, 0, 1000), "/target", nil),
			}),
			Entry("validates a target with units of mg/dL; target/range; range out of range", NewTestTarget(120.0, 130.0, nil, nil), "mg/dL", []*service.Error{
				testing.ComposeError(service.ErrorValueNotInRange(130, 0, 120), "/range", nil),
			}),
			Entry("validates a target with units of mg/dL; target/range; low exists", NewTestTarget(120.0, 10.0, 110.0, nil), "mg/dL", []*service.Error{
				testing.ComposeError(service.ErrorValueExists(), "/low", nil),
			}),
			Entry("validates a target with units of mg/dL; target/range; high exists", NewTestTarget(120.0, 10.0, nil, 130.0), "mg/dL", []*service.Error{
				testing.ComposeError(service.ErrorValueExists(), "/high", nil),
			}),
			Entry("validates a target with units of mg/dL; target/range; multiple", NewTestTarget(120.0, 130.0, 110.0, 130.0), "mg/dL", []*service.Error{
				testing.ComposeError(service.ErrorValueNotInRange(130, 0, 120), "/range", nil),
				testing.ComposeError(service.ErrorValueExists(), "/low", nil),
				testing.ComposeError(service.ErrorValueExists(), "/high", nil),
			}),
			Entry("validates a target with units of mg/dL; target/high; all valid", NewTestTarget(120.0, nil, nil, 130.0), "mg/dL", []*service.Error{}),
			Entry("validates a target with units of mg/dL; target/high; target out of range", NewTestTarget(1001.0, nil, nil, 130.0), "mg/dL", []*service.Error{
				testing.ComposeError(service.ErrorValueNotInRange(1001, 0, 1000), "/target", nil),
			}),
			Entry("validates a target with units of mg/dL; target/high; low exists", NewTestTarget(120.0, nil, 110.0, 130.0), "mg/dL", []*service.Error{
				testing.ComposeError(service.ErrorValueExists(), "/low", nil),
			}),
			Entry("validates a target with units of mg/dL; target/high; high out of range (lower)", NewTestTarget(120.0, nil, nil, 119.0), "mg/dL", []*service.Error{
				testing.ComposeError(service.ErrorValueNotInRange(119, 120, 1000), "/high", nil),
			}),
			Entry("validates a target with units of mg/dL; target/high; high out of range (upper)", NewTestTarget(120.0, nil, nil, 1001.0), "mg/dL", []*service.Error{
				testing.ComposeError(service.ErrorValueNotInRange(1001, 120, 1000), "/high", nil),
			}),
			Entry("validates a target with units of mg/dL; target/high; multiple", NewTestTarget(120.0, nil, 110.0, 1001.0), "mg/dL", []*service.Error{
				testing.ComposeError(service.ErrorValueExists(), "/low", nil),
				testing.ComposeError(service.ErrorValueNotInRange(1001, 120, 1000), "/high", nil),
			}),
			Entry("validates a target with units of mg/dL; target; all valid", NewTestTarget(120.0, nil, nil, nil), "mg/dL", []*service.Error{}),
			Entry("validates a target with units of mg/dL; target; target out of range", NewTestTarget(1001.0, nil, nil, nil), "mg/dL", []*service.Error{
				testing.ComposeError(service.ErrorValueNotInRange(1001, 0, 1000), "/target", nil),
			}),
			Entry("validates a target with units of mg/dL; target; low exists", NewTestTarget(120.0, nil, 110.0, nil), "mg/dL", []*service.Error{
				testing.ComposeError(service.ErrorValueExists(), "/low", nil),
			}),
			Entry("validates a target with units of mg/dL; target; multiple", NewTestTarget(1001.0, nil, 110.0, nil), "mg/dL", []*service.Error{
				testing.ComposeError(service.ErrorValueNotInRange(1001, 0, 1000), "/target", nil),
				testing.ComposeError(service.ErrorValueExists(), "/low", nil),
			}),
			Entry("validates a target with units of mg/dL; low/high; all valid", NewTestTarget(nil, nil, 110.0, 130.0), "mg/dL", []*service.Error{}),
			Entry("validates a target with units of mg/dL; low/high; low out of range", NewTestTarget(nil, nil, -1.0, 130.0), "mg/dL", []*service.Error{
				testing.ComposeError(service.ErrorValueNotInRange(-1, 0, 1000), "/low", nil),
			}),
			Entry("validates a target with units of mg/dL; low/high; high out of range (lower)", NewTestTarget(nil, nil, 110.0, 109.0), "mg/dL", []*service.Error{
				testing.ComposeError(service.ErrorValueNotInRange(109, 110, 1000), "/high", nil),
			}),
			Entry("validates a target with units of mg/dL; low/high; high out of range (upper)", NewTestTarget(nil, nil, 110.0, 1001.0), "mg/dL", []*service.Error{
				testing.ComposeError(service.ErrorValueNotInRange(1001, 110, 1000), "/high", nil),
			}),
			Entry("validates a target with units of mg/dL; low", NewTestTarget(nil, nil, 110.0, nil), "mg/dL", []*service.Error{
				testing.ComposeError(service.ErrorValueNotExists(), "/high", nil),
			}),
			Entry("validates a target with units of mg/dL; none", NewTestTarget(nil, nil, nil, nil), nil, []*service.Error{
				testing.ComposeError(service.ErrorValueNotExists(), "/target", nil),
			}),
		)

		DescribeTable("Normalize",
			func(sourceTarget *glucose.Target, sourceUnits interface{}, expectedTarget *glucose.Target) {
				testContext, err := context.NewStandard(log.NewNull())
				Expect(err).ToNot(HaveOccurred())
				Expect(testContext).ToNot(BeNil())
				testNormalizer, err := normalizer.NewStandard(testContext)
				Expect(err).ToNot(HaveOccurred())
				Expect(testNormalizer).ToNot(BeNil())
				sourceTarget.Normalize(testNormalizer, AsStringPointer(sourceUnits))
				Expect(sourceTarget).To(Equal(expectedTarget))
			},
			Entry("normalizes a target with units of nil", NewTestTarget(120.0, 10.0, 110.0, 130.0), nil, NewTestTarget(120.0, 10.0, 110.0, 130.0)),
			Entry("normalizes a target with units of nil and values are nil", NewTestTarget(nil, nil, nil, nil), nil, NewTestTarget(nil, nil, nil, nil)),
			Entry("normalizes a target with units of unknown", NewTestTarget(120.0, 10.0, 110.0, 130.0), "unknown", NewTestTarget(120.0, 10.0, 110.0, 130.0)),
			Entry("normalizes a target with units of unknown and values are nil", NewTestTarget(nil, nil, nil, nil), "unknown", NewTestTarget(nil, nil, nil, nil)),
			Entry("normalizes a target with units of mmol/L", NewTestTarget(6.6, 1.0, 5.6, 7.6), "mmol/L", NewTestTarget(6.6, 1.0, 5.6, 7.6)),
			Entry("normalizes a target with units of mmol/L and values are nil", NewTestTarget(nil, nil, nil, nil), "mmol/L", NewTestTarget(nil, nil, nil, nil)),
			Entry("normalizes a target with units of mg/dL", NewTestTarget(120.0, 10.0, 110.0, 130.0), "mg/dL", NewTestTarget(6.66090, 0.55507, 6.10582, 7.21597)),
			Entry("normalizes a target with units of mg/dL and values are nil", NewTestTarget(nil, nil, nil, nil), "mg/dL", NewTestTarget(nil, nil, nil, nil)),
		)
	})

	DescribeTable("TargetRangeForUnits",
		func(units *string, expectedLower float64, expectedUpper float64) {
			actualLower, actualUpper := glucose.TargetRangeForUnits(units)
			Expect(actualLower).To(Equal(expectedLower))
			Expect(actualUpper).To(Equal(expectedUpper))
		},
		Entry("returns no range for nil", nil, -math.MaxFloat64, math.MaxFloat64),
		Entry("returns no range for unknown units", app.StringAsPointer("unknown"), -math.MaxFloat64, math.MaxFloat64),
		Entry("returns expected range for mmol/L units", app.StringAsPointer("mmol/L"), 0.0, 55.0),
		Entry("returns expected range for mmol/l units", app.StringAsPointer("mmol/l"), 0.0, 55.0),
		Entry("returns expected range for mg/dL units", app.StringAsPointer("mg/dL"), 0.0, 1000.0),
		Entry("returns expected range for mg/dl units", app.StringAsPointer("mg/dl"), 0.0, 1000.0),
	)

	DescribeTable("RangeRangeForUnits",
		func(sourceTarget float64, sourceUnits interface{}, expectedLower float64, expectedUpper float64) {
			actualLower, actualUpper := glucose.RangeRangeForUnits(sourceTarget, AsStringPointer(sourceUnits))
			Expect(actualLower).To(BeNumerically("~", expectedLower))
			Expect(actualUpper).To(BeNumerically("~", expectedUpper))
		},
		Entry("returns range where units are nil", 120.0, nil, -math.MaxFloat64, math.MaxFloat64),
		Entry("returns range where units are unknown", 120.0, "unknown", -math.MaxFloat64, math.MaxFloat64),
		Entry("returns range where units are mmol/L and target < lower limit", -0.1, "mmol/L", -math.MaxFloat64, math.MaxFloat64),
		Entry("returns range where units are mmol/L and target is lower limit", 0.0, "mmol/L", 0.0, 0.0),
		Entry("returns range where units are mmol/L and target is just above lower limit", 0.1, "mmol/L", 0.0, 0.1),
		Entry("returns range where units are mmol/L and target is a reasonable value", 6.6, "mmol/L", 0.0, 6.6),
		Entry("returns range where units are mmol/L and target is just below upper limit", 54.9, "mmol/L", 0.0, 0.1),
		Entry("returns range where units are mmol/L and target is upper limit", 55.0, "mmol/L", 0.0, 0.0),
		Entry("returns range where units are mmol/L and target > upper limit", 55.1, "mmol/L", -math.MaxFloat64, math.MaxFloat64),
		Entry("returns range where units are mmol/l and target < lower limit", -0.1, "mmol/l", -math.MaxFloat64, math.MaxFloat64),
		Entry("returns range where units are mmol/l and target is lower limit", 0.0, "mmol/l", 0.0, 0.0),
		Entry("returns range where units are mmol/l and target is just above lower limit", 0.1, "mmol/l", 0.0, 0.1),
		Entry("returns range where units are mmol/l and target is a reasonable value", 6.6, "mmol/l", 0.0, 6.6),
		Entry("returns range where units are mmol/l and target is just below upper limit", 54.9, "mmol/l", 0.0, 0.1),
		Entry("returns range where units are mmol/l and target is upper limit", 55.0, "mmol/l", 0.0, 0.0),
		Entry("returns range where units are mmol/l and target > upper limit", 55.1, "mmol/l", -math.MaxFloat64, math.MaxFloat64),
		Entry("returns range where units are mg/dL and target < lower limit", -1.0, "mg/dL", -math.MaxFloat64, math.MaxFloat64),
		Entry("returns range where units are mg/dL and target is lower limit", 0.0, "mg/dL", 0.0, 0.0),
		Entry("returns range where units are mg/dL and target is just above lower limit", 1.0, "mg/dL", 0.0, 1.0),
		Entry("returns range where units are mg/dL and target is a reasonable value", 120.0, "mg/dL", 0.0, 120.0),
		Entry("returns range where units are mg/dL and target is just below upper limit", 999.0, "mg/dL", 0.0, 1.0),
		Entry("returns range where units are mg/dL and target is upper limit", 1000.0, "mg/dL", 0.0, 0.0),
		Entry("returns range where units are mg/dL and target > upper limit", 1001.0, "mg/dL", -math.MaxFloat64, math.MaxFloat64),
		Entry("returns range where units are mg/dl and target < lower limit", -1.0, "mg/dl", -math.MaxFloat64, math.MaxFloat64),
		Entry("returns range where units are mg/dl and target is lower limit", 0.0, "mg/dl", 0.0, 0.0),
		Entry("returns range where units are mg/dl and target is just above lower limit", 1.0, "mg/dl", 0.0, 1.0),
		Entry("returns range where units are mg/dl and target is a reasonable value", 120.0, "mg/dl", 0.0, 120.0),
		Entry("returns range where units are mg/dl and target is just below upper limit", 999.0, "mg/dl", 0.0, 1.0),
		Entry("returns range where units are mg/dl and target is upper limit", 1000.0, "mg/dl", 0.0, 0.0),
		Entry("returns range where units are mg/dl and target > upper limit", 1001.0, "mg/dl", -math.MaxFloat64, math.MaxFloat64),
	)

	DescribeTable("LowRangeForUnits",
		func(units *string, expectedLower float64, expectedUpper float64) {
			actualLower, actualUpper := glucose.LowRangeForUnits(units)
			Expect(actualLower).To(Equal(expectedLower))
			Expect(actualUpper).To(Equal(expectedUpper))
		},
		Entry("returns no range for nil", nil, -math.MaxFloat64, math.MaxFloat64),
		Entry("returns no range for unknown units", app.StringAsPointer("unknown"), -math.MaxFloat64, math.MaxFloat64),
		Entry("returns expected range for mmol/L units", app.StringAsPointer("mmol/L"), 0.0, 55.0),
		Entry("returns expected range for mmol/l units", app.StringAsPointer("mmol/l"), 0.0, 55.0),
		Entry("returns expected range for mg/dL units", app.StringAsPointer("mg/dL"), 0.0, 1000.0),
		Entry("returns expected range for mg/dl units", app.StringAsPointer("mg/dl"), 0.0, 1000.0),
	)

	DescribeTable("HighRangeForUnits",
		func(sourceLow float64, sourceUnits interface{}, expectedLower float64, expectedUpper float64) {
			actualLower, actualUpper := glucose.HighRangeForUnits(sourceLow, AsStringPointer(sourceUnits))
			Expect(actualLower).To(BeNumerically("~", expectedLower))
			Expect(actualUpper).To(BeNumerically("~", expectedUpper))
		},
		Entry("returns range where units are nil", 120.0, nil, -math.MaxFloat64, math.MaxFloat64),
		Entry("returns range where units are unknown", 120.0, "unknown", -math.MaxFloat64, math.MaxFloat64),
		Entry("returns range where units are mmol/L and low < lower limit", -0.1, "mmol/L", -math.MaxFloat64, math.MaxFloat64),
		Entry("returns range where units are mmol/L and low is lower limit", 0.0, "mmol/L", 0.0, 55.0),
		Entry("returns range where units are mmol/L and low is just above lower limit", 0.1, "mmol/L", 0.1, 55.0),
		Entry("returns range where units are mmol/L and low is a reasonable value", 6.6, "mmol/L", 6.6, 55.0),
		Entry("returns range where units are mmol/L and low is just below upper limit", 54.9, "mmol/L", 54.9, 55.0),
		Entry("returns range where units are mmol/L and low is upper limit", 55.0, "mmol/L", 55.0, 55.0),
		Entry("returns range where units are mmol/L and low > upper limit", 55.1, "mmol/L", -math.MaxFloat64, math.MaxFloat64),
		Entry("returns range where units are mmol/l and low < lower limit", -0.1, "mmol/l", -math.MaxFloat64, math.MaxFloat64),
		Entry("returns range where units are mmol/l and low is lower limit", 0.0, "mmol/l", 0.0, 55.0),
		Entry("returns range where units are mmol/l and low is just above lower limit", 0.1, "mmol/l", 0.1, 55.0),
		Entry("returns range where units are mmol/l and low is a reasonable value", 6.6, "mmol/l", 6.6, 55.0),
		Entry("returns range where units are mmol/l and low is just below upper limit", 54.9, "mmol/l", 54.9, 55.0),
		Entry("returns range where units are mmol/l and low is upper limit", 55.0, "mmol/l", 55.0, 55.0),
		Entry("returns range where units are mmol/l and low > upper limit", 55.1, "mmol/l", -math.MaxFloat64, math.MaxFloat64),
		Entry("returns range where units are mg/dL and low < lower limit", -1.0, "mg/dL", -math.MaxFloat64, math.MaxFloat64),
		Entry("returns range where units are mg/dL and low is lower limit", 0.0, "mg/dL", 0.0, 1000.0),
		Entry("returns range where units are mg/dL and low is just above lower limit", 1.0, "mg/dL", 1.0, 1000.0),
		Entry("returns range where units are mg/dL and low is a reasonable value", 120.0, "mg/dL", 120.0, 1000.0),
		Entry("returns range where units are mg/dL and low is just below upper limit", 999.0, "mg/dL", 999.0, 1000.0),
		Entry("returns range where units are mg/dL and low is upper limit", 1000.0, "mg/dL", 1000.0, 1000.0),
		Entry("returns range where units are mg/dL and low > upper limit", 1001.0, "mg/dL", -math.MaxFloat64, math.MaxFloat64),
		Entry("returns range where units are mg/dl and low < lower limit", -1.0, "mg/dl", -math.MaxFloat64, math.MaxFloat64),
		Entry("returns range where units are mg/dl and low is lower limit", 0.0, "mg/dl", 0.0, 1000.0),
		Entry("returns range where units are mg/dl and low is just above lower limit", 1.0, "mg/dl", 1.0, 1000.0),
		Entry("returns range where units are mg/dl and low is a reasonable value", 120.0, "mg/dl", 120.0, 1000.0),
		Entry("returns range where units are mg/dl and low is just below upper limit", 999.0, "mg/dl", 999.0, 1000.0),
		Entry("returns range where units are mg/dl and low is upper limit", 1000.0, "mg/dl", 1000.0, 1000.0),
		Entry("returns range where units are mg/dl and low > upper limit", 1001.0, "mg/dl", -math.MaxFloat64, math.MaxFloat64),
	)
})
