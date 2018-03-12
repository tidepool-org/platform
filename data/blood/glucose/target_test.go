package glucose_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"math"

	"github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/data/context"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/parser"
	testData "github.com/tidepool-org/platform/data/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/service"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func AsStringPointer(source interface{}) *string {
	if sourceString, ok := source.(string); ok {
		return pointer.String(sourceString)
	}
	return nil
}

func NewTarget(high interface{}, low interface{}, rng interface{}, target interface{}) *glucose.Target {
	datum := glucose.NewTarget()
	if value, ok := high.(float64); ok {
		datum.High = &value
	}
	if value, ok := low.(float64); ok {
		datum.Low = &value
	}
	if value, ok := rng.(float64); ok {
		datum.Range = &value
	}
	if value, ok := target.(float64); ok {
		datum.Target = &value
	}
	return datum
}

func NewTestTarget(sourceTarget interface{}, sourceRange interface{}, sourceLow interface{}, sourceHigh interface{}) *glucose.Target {
	testTarget := glucose.NewTarget()
	if value, ok := sourceTarget.(float64); ok {
		testTarget.Target = pointer.Float64(value)
	}
	if value, ok := sourceRange.(float64); ok {
		testTarget.Range = pointer.Float64(value)
	}
	if value, ok := sourceLow.(float64); ok {
		testTarget.Low = pointer.Float64(value)
	}
	if value, ok := sourceHigh.(float64); ok {
		testTarget.High = pointer.Float64(value)
	}
	return testTarget
}

var _ = Describe("Target", func() {
	DescribeTable("ParseTarget",
		func(sourceObject *map[string]interface{}, expectedTarget *glucose.Target, expectedErrors []*service.Error) {
			testContext, err := context.NewStandard(null.NewLogger())
			Expect(err).ToNot(HaveOccurred())
			Expect(testContext).ToNot(BeNil())
			testParser, err := parser.NewStandardObject(testContext, sourceObject, parser.AppendErrorNotParsed)
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
				testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/target", nil),
				testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/range", nil),
				testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/low", nil),
				testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/high", nil),
			}),
		Entry("parses object that has additional fields", &map[string]interface{}{"target": 120.0, "range": 10.0, "low": 110.0, "high": 130.0, "additional": 0.0},
			NewTestTarget(120.0, 10.0, 110.0, 130.0), []*service.Error{
				testData.ComposeError(parser.ErrorNotParsed(), "/additional", nil),
			}),
	)

	Context("NewTarget", func() {
		It("is successful", func() {
			Expect(glucose.NewTarget()).To(Equal(&glucose.Target{}))
		})
	})

	Context("Target", func() {
		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(sourceObject *map[string]interface{}, expectedDatum *glucose.Target, expectedErrors []*service.Error) {
					testContext, err := context.NewStandard(null.NewLogger())
					Expect(err).ToNot(HaveOccurred())
					Expect(testContext).ToNot(BeNil())
					testParser, err := parser.NewStandardObject(testContext, sourceObject, parser.AppendErrorNotParsed)
					Expect(err).ToNot(HaveOccurred())
					Expect(testParser).ToNot(BeNil())
					sourceTarget := &glucose.Target{}
					sourceTarget.Parse(testParser)
					Expect(sourceTarget).To(Equal(expectedDatum))
					Expect(testContext.Errors()).To(ConsistOf(expectedErrors))
				},
				Entry("succeeds", &map[string]interface{}{"target": 120.0, "range": 10.0, "low": 110.0, "high": 130.0},
					NewTestTarget(120.0, 10.0, 110.0, 130.0), []*service.Error{}),
				Entry("nil", nil, NewTestTarget(nil, nil, nil, nil), []*service.Error{}),
				Entry("empty", &map[string]interface{}{}, NewTestTarget(nil, nil, nil, nil), []*service.Error{}),
				Entry("target valid", &map[string]interface{}{"target": 120.0}, NewTestTarget(120.0, nil, nil, nil), []*service.Error{}),
				Entry("target invalid", &map[string]interface{}{"target": "invalid"}, NewTestTarget(nil, nil, nil, nil), []*service.Error{
					testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/target", nil),
				}),
				Entry("range valid", &map[string]interface{}{"range": 10.0}, NewTestTarget(nil, 10.0, nil, nil), []*service.Error{}),
				Entry("range invalid", &map[string]interface{}{"range": "invalid"}, NewTestTarget(nil, nil, nil, nil), []*service.Error{
					testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/range", nil),
				}),
				Entry("low valid", &map[string]interface{}{"low": 110.0}, NewTestTarget(nil, nil, 110.0, nil), []*service.Error{}),
				Entry("low invalid", &map[string]interface{}{"low": "invalid"}, NewTestTarget(nil, nil, nil, nil), []*service.Error{
					testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/low", nil),
				}),
				Entry("high valid", &map[string]interface{}{"high": 130.0}, NewTestTarget(nil, nil, nil, 130.0), []*service.Error{}),
				Entry("high invalid", &map[string]interface{}{"high": "invalid"}, NewTestTarget(nil, nil, nil, nil), []*service.Error{
					testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/high", nil),
				}),
				Entry("multiple errors", &map[string]interface{}{"target": "invalid", "range": "invalid", "low": "invalid", "high": "invalid"},
					NewTestTarget(nil, nil, nil, nil), []*service.Error{
						testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/target", nil),
						testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/range", nil),
						testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/low", nil),
						testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/high", nil),
					}),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(datum *glucose.Target, sourceUnits string, expectedErrors ...error) {
					validator := structureValidator.New()
					Expect(validator).ToNot(BeNil())
					datum.Validate(validator, &sourceUnits)
					testErrors.ExpectEqual(validator.Error(), expectedErrors...)
				},
				Entry("units mmol/L; target/range; all valid",
					NewTarget(nil, nil, 1.0, 6.6), "mmol/L"),
				Entry("units mmol/L; target/range; high exists",
					NewTarget(7.6, nil, 1.0, 6.6), "mmol/L",
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/high"),
				),
				Entry("units mmol/L; target/range; low exists",
					NewTarget(nil, 5.6, 1.0, 6.6), "mmol/L",
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/low"),
				),
				Entry("units mmol/L; target/range; range out of range",
					NewTarget(nil, nil, 6.7, 6.6), "mmol/L",
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(6.7, 0, 6.6), "/range"),
				),
				Entry("units mmol/L; target/range; target out of range",
					NewTarget(nil, nil, 1.0, 55.1), "mmol/L",
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(55.1, 0, 55), "/target"),
				),
				Entry("units mmol/L; target/range; multiple",
					NewTarget(7.6, 5.6, 6.7, 6.6), "mmol/L",
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/high"),
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/low"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(6.7, 0, 6.6), "/range"),
				),
				Entry("units mmol/L; target/high; all valid",
					NewTarget(7.6, nil, nil, 6.6), "mmol/L"),
				Entry("units mmol/L; target/high; high out of range (lower)",
					NewTarget(6.5, nil, nil, 6.6), "mmol/L",
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(6.5, 6.6, 55), "/high"),
				),
				Entry("units mmol/L; target/high; high out of range (upper)",
					NewTarget(55.1, nil, nil, 6.6), "mmol/L",
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(55.1, 6.6, 55), "/high"),
				),
				Entry("units mmol/L; target/high; low exists",
					NewTarget(7.6, 5.6, nil, 6.6), "mmol/L",
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/low"),
				),
				Entry("units mmol/L; target/high; target out of range",
					NewTarget(7.6, nil, nil, 55.1), "mmol/L",
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(55.1, 0, 55), "/target"),
				),
				Entry("units mmol/L; target/high; multiple",
					NewTarget(55.1, 5.6, nil, 6.6), "mmol/L",
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(55.1, 6.6, 55), "/high"),
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/low"),
				),
				Entry("units mmol/L; target; all valid",
					NewTarget(nil, nil, nil, 6.6), "mmol/L"),
				Entry("units mmol/L; target; low exists",
					NewTarget(nil, 5.6, nil, 6.6), "mmol/L",
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/low"),
				),
				Entry("units mmol/L; target; target out of range",
					NewTarget(nil, nil, nil, 55.1), "mmol/L",
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(55.1, 0, 55), "/target"),
				),
				Entry("units mmol/L; target; multiple",
					NewTarget(nil, 5.6, nil, 55.1), "mmol/L",
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/low"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(55.1, 0, 55), "/target"),
				),
				Entry("units mmol/L; low/high; all valid",
					NewTarget(7.6, 5.6, nil, nil), "mmol/L"),
				Entry("units mmol/L; low/high; high out of range (lower)",
					NewTarget(5.5, 5.6, nil, nil), "mmol/L",
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(5.5, 5.6, 55), "/high"),
				),
				Entry("units mmol/L; low/high; high out of range (upper)",
					NewTarget(55.1, 5.6, nil, nil), "mmol/L",
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(55.1, 5.6, 55), "/high"),
				),
				Entry("units mmol/L; low/high; low out of range",
					NewTarget(7.6, -0.1, nil, nil), "mmol/L",
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0, 55), "/low"),
				),
				Entry("units mmol/L; low",
					NewTarget(nil, 5.6, nil, nil), "mmol/L",
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/high"),
				),
				Entry("units mmol/L; none",
					NewTarget(nil, nil, nil, nil), nil,
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/target"),
				),
				Entry("units mmol/l; target/range; all valid",
					NewTarget(nil, nil, 1.0, 6.6), "mmol/l"),
				Entry("units mmol/l; target/range; high exists",
					NewTarget(7.6, nil, 1.0, 6.6), "mmol/l",
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/high"),
				),
				Entry("units mmol/l; target/range; low exists",
					NewTarget(nil, 5.6, 1.0, 6.6), "mmol/l",
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/low"),
				),
				Entry("units mmol/l; target/range; range out of range",
					NewTarget(nil, nil, 6.7, 6.6), "mmol/l",
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(6.7, 0, 6.6), "/range"),
				),
				Entry("units mmol/l; target/range; target out of range",
					NewTarget(nil, nil, 1.0, 55.1), "mmol/l",
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(55.1, 0, 55), "/target"),
				),
				Entry("units mmol/l; target/range; multiple",
					NewTarget(7.6, 5.6, 6.7, 6.6), "mmol/l",
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/high"),
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/low"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(6.7, 0, 6.6), "/range"),
				),
				Entry("units mmol/l; target/high; all valid",
					NewTarget(7.6, nil, nil, 6.6), "mmol/l"),
				Entry("units mmol/l; target/high; high out of range (lower)",
					NewTarget(6.5, nil, nil, 6.6), "mmol/l",
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(6.5, 6.6, 55), "/high"),
				),
				Entry("units mmol/l; target/high; high out of range (upper)",
					NewTarget(55.1, nil, nil, 6.6), "mmol/l",
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(55.1, 6.6, 55), "/high"),
				),
				Entry("units mmol/l; target/high; low exists",
					NewTarget(7.6, 5.6, nil, 6.6), "mmol/l",
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/low"),
				),
				Entry("units mmol/l; target/high; target out of range",
					NewTarget(7.6, nil, nil, 55.1), "mmol/l",
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(55.1, 0, 55), "/target"),
				),
				Entry("units mmol/l; target/high; multiple",
					NewTarget(55.1, 5.6, nil, 6.6), "mmol/l",
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(55.1, 6.6, 55), "/high"),
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/low"),
				),
				Entry("units mmol/l; target; all valid",
					NewTarget(nil, nil, nil, 6.6), "mmol/l"),
				Entry("units mmol/l; target; low exists",
					NewTarget(nil, 5.6, nil, 6.6), "mmol/l",
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/low"),
				),
				Entry("units mmol/l; target; target out of range",
					NewTarget(nil, nil, nil, 55.1), "mmol/l",
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(55.1, 0, 55), "/target"),
				),
				Entry("units mmol/l; target; multiple",
					NewTarget(nil, 5.6, nil, 55.1), "mmol/l",
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/low"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(55.1, 0, 55), "/target"),
				),
				Entry("units mmol/l; low/high; all valid",
					NewTarget(7.6, 5.6, nil, nil), "mmol/l"),
				Entry("units mmol/l; low/high; high out of range (lower)",
					NewTarget(5.5, 5.6, nil, nil), "mmol/l",
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(5.5, 5.6, 55), "/high"),
				),
				Entry("units mmol/l; low/high; high out of range (upper)",
					NewTarget(55.1, 5.6, nil, nil), "mmol/l",
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(55.1, 5.6, 55), "/high"),
				),
				Entry("units mmol/l; low/high; low out of range",
					NewTarget(7.6, -0.1, nil, nil), "mmol/l",
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0, 55), "/low"),
				),
				Entry("units mmol/l; low",
					NewTarget(nil, 5.6, nil, nil), "mmol/l",
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/high"),
				),
				Entry("units mmol/l; none",
					NewTarget(nil, nil, nil, nil), nil,
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/target"),
				),
				Entry("units mg/dL; target/range; all valid",
					NewTarget(nil, nil, 10.0, 120.0), "mg/dL"),
				Entry("units mg/dL; target/range; high exists",
					NewTarget(130.0, nil, 10.0, 120.0), "mg/dL",
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/high"),
				),
				Entry("units mg/dL; target/range; low exists",
					NewTarget(nil, 110.0, 10.0, 120.0), "mg/dL",
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/low"),
				),
				Entry("units mg/dL; target/range; range out of range",
					NewTarget(nil, nil, 130.0, 120.0), "mg/dL",
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(130, 0, 120), "/range"),
				),
				Entry("units mg/dL; target/range; target out of range",
					NewTarget(nil, nil, 10.0, 1001.0), "mg/dL",
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(1001, 0, 1000), "/target"),
				),
				Entry("units mg/dL; target/range; multiple",
					NewTarget(130.0, 110.0, 130.0, 120.0), "mg/dL",
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/high"),
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/low"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(130, 0, 120), "/range"),
				),
				Entry("units mg/dL; target/high; all valid",
					NewTarget(130, nil, nil, 120.0), "mg/dL"),
				Entry("units mg/dL; target/high; high out of range (lower)",
					NewTarget(119.0, nil, nil, 120.0), "mg/dL",
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(119, 120, 1000), "/high"),
				),
				Entry("units mg/dL; target/high; high out of range (upper)",
					NewTarget(1001.0, nil, nil, 120.0), "mg/dL",
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(1001, 120, 1000), "/high"),
				),
				Entry("units mg/dL; target/high; low exists",
					NewTarget(130.0, 110.0, nil, 120.0), "mg/dL",
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/low"),
				),
				Entry("units mg/dL; target/high; target out of range",
					NewTarget(130.0, nil, nil, 1001.0), "mg/dL",
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(1001, 0, 1000), "/target"),
				),
				Entry("units mg/dL; target/high; multiple",
					NewTarget(1001.0, 110.0, nil, 120.0), "mg/dL",
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(1001, 120, 1000), "/high"),
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/low"),
				),
				Entry("units mg/dL; target; all valid",
					NewTarget(nil, nil, nil, 120.0), "mg/dL"),
				Entry("units mg/dL; target; low exists",
					NewTarget(nil, 110.0, nil, 120.0), "mg/dL",
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/low"),
				),
				Entry("units mg/dL; target; target out of range",
					NewTarget(nil, nil, nil, 1001.0), "mg/dL",
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(1001, 0, 1000), "/target"),
				),
				Entry("units mg/dL; target; multiple",
					NewTarget(nil, 110.0, nil, 1001.0), "mg/dL",
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/low"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(1001, 0, 1000), "/target"),
				),
				Entry("units mg/dL; low/high; all valid",
					NewTarget(130.0, 110.0, nil, nil), "mg/dL"),
				Entry("units mg/dL; low/high; high out of range (lower)",
					NewTarget(109.0, 110.0, nil, nil), "mg/dL",
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(109, 110, 1000), "/high"),
				),
				Entry("units mg/dL; low/high; high out of range (upper)",
					NewTarget(1001.0, 110.0, nil, nil), "mg/dL",
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(1001, 110, 1000), "/high"),
				),
				Entry("units mg/dL; low/high; low out of range",
					NewTarget(130.0, -1.0, nil, nil), "mg/dL",
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(-1, 0, 1000), "/low"),
				),
				Entry("units mg/dL; low",
					NewTarget(nil, 110.0, nil, nil), "mg/dL",
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/high"),
				),
				Entry("units mg/dL; none",
					NewTarget(nil, nil, nil, nil), nil,
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/target"),
				),
				Entry("units mg/dl; target/range; all valid",
					NewTarget(nil, nil, 10.0, 120.0), "mg/dl"),
				Entry("units mg/dl; target/range; high exists",
					NewTarget(130.0, nil, 10.0, 120.0), "mg/dl",
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/high"),
				),
				Entry("units mg/dl; target/range; low exists",
					NewTarget(nil, 110.0, 10.0, 120.0), "mg/dl",
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/low"),
				),
				Entry("units mg/dl; target/range; range out of range",
					NewTarget(nil, nil, 130.0, 120.0), "mg/dl",
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(130, 0, 120), "/range"),
				),
				Entry("units mg/dl; target/range; target out of range",
					NewTarget(nil, nil, 10.0, 1001.0), "mg/dl",
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(1001, 0, 1000), "/target"),
				),
				Entry("units mg/dl; target/range; multiple",
					NewTarget(130.0, 110.0, 130.0, 120.0), "mg/dl",
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/high"),
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/low"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(130, 0, 120), "/range"),
				),
				Entry("units mg/dl; target/high; all valid",
					NewTarget(130, nil, nil, 120.0), "mg/dl"),
				Entry("units mg/dl; target/high; high out of range (lower)",
					NewTarget(119.0, nil, nil, 120.0), "mg/dl",
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(119, 120, 1000), "/high"),
				),
				Entry("units mg/dl; target/high; high out of range (upper)",
					NewTarget(1001.0, nil, nil, 120.0), "mg/dl",
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(1001, 120, 1000), "/high"),
				),
				Entry("units mg/dl; target/high; low exists",
					NewTarget(130.0, 110.0, nil, 120.0), "mg/dl",
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/low"),
				),
				Entry("units mg/dl; target/high; target out of range",
					NewTarget(130.0, nil, nil, 1001.0), "mg/dl",
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(1001, 0, 1000), "/target"),
				),
				Entry("units mg/dl; target/high; multiple",
					NewTarget(1001.0, 110.0, nil, 120.0), "mg/dl",
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(1001, 120, 1000), "/high"),
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/low"),
				),
				Entry("units mg/dl; target; all valid",
					NewTarget(nil, nil, nil, 120.0), "mg/dl"),
				Entry("units mg/dl; target; low exists",
					NewTarget(nil, 110.0, nil, 120.0), "mg/dl",
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/low"),
				),
				Entry("units mg/dl; target; target out of range",
					NewTarget(nil, nil, nil, 1001.0), "mg/dl",
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(1001, 0, 1000), "/target"),
				),
				Entry("units mg/dl; target; multiple",
					NewTarget(nil, 110.0, nil, 1001.0), "mg/dl",
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/low"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(1001, 0, 1000), "/target"),
				),
				Entry("units mg/dl; low/high; all valid",
					NewTarget(130.0, 110.0, nil, nil), "mg/dl"),
				Entry("units mg/dl; low/high; high out of range (lower)",
					NewTarget(109.0, 110.0, nil, nil), "mg/dl",
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(109, 110, 1000), "/high"),
				),
				Entry("units mg/dl; low/high; high out of range (upper)",
					NewTarget(1001.0, 110.0, nil, nil), "mg/dl",
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(1001, 110, 1000), "/high"),
				),
				Entry("units mg/dl; low/high; low out of range",
					NewTarget(130.0, -1.0, nil, nil), "mg/dl",
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(-1, 0, 1000), "/low"),
				),
				Entry("units mg/dl; low",
					NewTarget(nil, 110.0, nil, nil), "mg/dl",
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/high"),
				),
				Entry("units mg/dl; none",
					NewTarget(nil, nil, nil, nil), nil,
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/target"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(datum *glucose.Target, sourceUnits interface{}, expectedDatum *glucose.Target) {
					normalizer := dataNormalizer.New()
					Expect(normalizer).ToNot(BeNil())
					datum.Normalize(normalizer, AsStringPointer(sourceUnits))
					Expect(normalizer.Error()).ToNot(HaveOccurred())
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("normalizes a target with units of nil", NewTarget(130.0, 110.0, 10.0, 120.0), nil, NewTarget(130.0, 110.0, 10.0, 120.0)),
				Entry("normalizes a target with units of nil and values are nil", NewTarget(nil, nil, nil, nil), nil, NewTarget(nil, nil, nil, nil)),
				Entry("normalizes a target with units of unknown", NewTarget(130.0, 110.0, 10.0, 120.0), "unknown", NewTarget(130.0, 110.0, 10.0, 120.0)),
				Entry("normalizes a target with units of unknown and values are nil", NewTarget(nil, nil, nil, nil), "unknown", NewTarget(nil, nil, nil, nil)),
				Entry("normalizes a target with units of mmol/L", NewTarget(7.6, 5.6, 1.0, 6.6), "mmol/L", NewTarget(7.6, 5.6, 1.0, 6.6)),
				Entry("normalizes a target with units of mmol/L and values are nil", NewTarget(nil, nil, nil, nil), "mmol/L", NewTarget(nil, nil, nil, nil)),
				Entry("normalizes a target with units of mmol/l", NewTarget(7.6, 5.6, 1.0, 6.6), "mmol/l", NewTarget(7.6, 5.6, 1.0, 6.6)),
				Entry("normalizes a target with units of mmol/l and values are nil", NewTarget(nil, nil, nil, nil), "mmol/l", NewTarget(nil, nil, nil, nil)),
				Entry("normalizes a target with units of mg/dL", NewTarget(130.0, 110.0, 10.0, 120.0), "mg/dL", NewTarget(7.21597, 6.10582, 0.55507, 6.66090)),
				Entry("normalizes a target with units of mg/dL and values are nil", NewTarget(nil, nil, nil, nil), "mg/dL", NewTarget(nil, nil, nil, nil)),
				Entry("normalizes a target with units of mg/dl", NewTarget(130.0, 110.0, 10.0, 120.0), "mg/dl", NewTarget(7.21597, 6.10582, 0.55507, 6.66090)),
				Entry("normalizes a target with units of mg/dl and values are nil", NewTarget(nil, nil, nil, nil), "mg/dl", NewTarget(nil, nil, nil, nil)),
			)
		})
	})

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

	DescribeTable("LowRangeForUnits",
		func(units *string, expectedLower float64, expectedUpper float64) {
			actualLower, actualUpper := glucose.LowRangeForUnits(units)
			Expect(actualLower).To(Equal(expectedLower))
			Expect(actualUpper).To(Equal(expectedUpper))
		},
		Entry("returns no range for nil", nil, -math.MaxFloat64, math.MaxFloat64),
		Entry("returns no range for unknown units", pointer.String("unknown"), -math.MaxFloat64, math.MaxFloat64),
		Entry("returns expected range for mmol/L units", pointer.String("mmol/L"), 0.0, 55.0),
		Entry("returns expected range for mmol/l units", pointer.String("mmol/l"), 0.0, 55.0),
		Entry("returns expected range for mg/dL units", pointer.String("mg/dL"), 0.0, 1000.0),
		Entry("returns expected range for mg/dl units", pointer.String("mg/dl"), 0.0, 1000.0),
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

	DescribeTable("TargetRangeForUnits",
		func(units *string, expectedLower float64, expectedUpper float64) {
			actualLower, actualUpper := glucose.TargetRangeForUnits(units)
			Expect(actualLower).To(Equal(expectedLower))
			Expect(actualUpper).To(Equal(expectedUpper))
		},
		Entry("returns no range for nil", nil, -math.MaxFloat64, math.MaxFloat64),
		Entry("returns no range for unknown units", pointer.String("unknown"), -math.MaxFloat64, math.MaxFloat64),
		Entry("returns expected range for mmol/L units", pointer.String("mmol/L"), 0.0, 55.0),
		Entry("returns expected range for mmol/l units", pointer.String("mmol/l"), 0.0, 55.0),
		Entry("returns expected range for mg/dL units", pointer.String("mg/dL"), 0.0, 1000.0),
		Entry("returns expected range for mg/dl units", pointer.String("mg/dl"), 0.0, 1000.0),
	)
})
