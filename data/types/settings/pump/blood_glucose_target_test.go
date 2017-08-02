package pump_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/factory"
	"github.com/tidepool-org/platform/data/parser"
	testData "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/service"
)

func AsStringPointer(source interface{}) *string {
	if sourceString, ok := source.(string); ok {
		return pointer.String(sourceString)
	}
	return nil
}

func NewTestBloodGlucoseTarget(sourceTarget interface{}, sourceRange interface{}, sourceLow interface{}, sourceHigh interface{}, sourceStart interface{}) *pump.BloodGlucoseTarget {
	testTarget := &pump.BloodGlucoseTarget{}
	if value, ok := sourceTarget.(float64); ok {
		testTarget.Target.Target = pointer.Float(value)
	}
	if value, ok := sourceRange.(float64); ok {
		testTarget.Range = pointer.Float(value)
	}
	if value, ok := sourceLow.(float64); ok {
		testTarget.Low = pointer.Float(value)
	}
	if value, ok := sourceHigh.(float64); ok {
		testTarget.High = pointer.Float(value)
	}
	if value, ok := sourceStart.(int); ok {
		testTarget.Start = pointer.Integer(value)
	}
	return testTarget
}

var _ = Describe("BloodGlucoseTarget", func() {
	DescribeTable("ParseBloodGlucoseTarget",
		func(sourceObject *map[string]interface{}, expectedTarget *pump.BloodGlucoseTarget, expectedErrors []*service.Error) {
			testContext, err := context.NewStandard(null.NewLogger())
			Expect(err).ToNot(HaveOccurred())
			Expect(testContext).ToNot(BeNil())
			testFactory, err := factory.NewStandard()
			Expect(err).ToNot(HaveOccurred())
			Expect(testFactory).ToNot(BeNil())
			testParser, err := parser.NewStandardObject(testContext, testFactory, sourceObject, parser.AppendErrorNotParsed)
			Expect(err).ToNot(HaveOccurred())
			Expect(testParser).ToNot(BeNil())
			Expect(pump.ParseBloodGlucoseTarget(testParser)).To(Equal(expectedTarget))
			Expect(testContext.Errors()).To(ConsistOf(expectedErrors))
		},
		Entry("parses object that is nil", nil, nil, []*service.Error{}),
		Entry("parses object that is empty", &map[string]interface{}{}, NewTestBloodGlucoseTarget(nil, nil, nil, nil, nil), []*service.Error{}),
		Entry("parses object that has multiple valid fields", &map[string]interface{}{"target": 120.0, "range": 10.0, "low": 110.0, "high": 130.0, "start": 3600},
			NewTestBloodGlucoseTarget(120.0, 10.0, 110.0, 130.0, 3600), []*service.Error{}),
		Entry("parses object that has multiple invalid fields", &map[string]interface{}{"target": "invalid", "range": "invalid", "low": "invalid", "high": "invalid", "start": "invalid"},
			NewTestBloodGlucoseTarget(nil, nil, nil, nil, nil), []*service.Error{
				testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/target", nil),
				testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/range", nil),
				testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/low", nil),
				testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/high", nil),
				testData.ComposeError(service.ErrorTypeNotInteger("invalid"), "/start", nil),
			}),
		Entry("parses object that has additional fields", &map[string]interface{}{"target": 120.0, "range": 10.0, "low": 110.0, "high": 130.0, "start": 3600, "additional": 0.0},
			NewTestBloodGlucoseTarget(120.0, 10.0, 110.0, 130.0, 3600), []*service.Error{
				testData.ComposeError(parser.ErrorNotParsed(), "/additional", nil),
			}),
	)

	DescribeTable("ParseBloodGlucoseTargetArray",
		func(sourceArray *[]interface{}, expectedArray *[]*pump.BloodGlucoseTarget, expectedErrors []*service.Error) {
			testContext, err := context.NewStandard(null.NewLogger())
			Expect(err).ToNot(HaveOccurred())
			Expect(testContext).ToNot(BeNil())
			testFactory, err := factory.NewStandard()
			Expect(err).ToNot(HaveOccurred())
			Expect(testFactory).ToNot(BeNil())
			testParser, err := parser.NewStandardArray(testContext, testFactory, sourceArray, parser.AppendErrorNotParsed)
			Expect(err).ToNot(HaveOccurred())
			Expect(testParser).ToNot(BeNil())
			Expect(pump.ParseBloodGlucoseTargetArray(testParser)).To(Equal(expectedArray))
			Expect(testContext.Errors()).To(ConsistOf(expectedErrors))
		},
		Entry("parses array that is nil", nil, nil, []*service.Error{}),
		Entry("parses array that is empty", &[]interface{}{}, &[]*pump.BloodGlucoseTarget{}, []*service.Error{}),
		Entry("parses array that has one valid",
			&[]interface{}{
				map[string]interface{}{"target": 120.0, "range": 10.0, "low": 110.0, "high": 130.0, "start": 3600},
			},
			&[]*pump.BloodGlucoseTarget{
				NewTestBloodGlucoseTarget(120.0, 10.0, 110.0, 130.0, 3600),
			}, []*service.Error{}),
		Entry("parses array that has more than one valid",
			&[]interface{}{
				map[string]interface{}{"target": 120.0, "range": 10.0, "low": 110.0, "high": 130.0, "start": 3600},
				map[string]interface{}{"target": 121.0, "range": 11.0, "low": 111.0, "high": 131.0, "start": 3601},
			},
			&[]*pump.BloodGlucoseTarget{
				NewTestBloodGlucoseTarget(120.0, 10.0, 110.0, 130.0, 3600),
				NewTestBloodGlucoseTarget(121.0, 11.0, 111.0, 131.0, 3601),
			}, []*service.Error{}),
		Entry("parses array that has one valid and one invalid",
			&[]interface{}{
				map[string]interface{}{"target": "invalid", "range": "invalid", "low": "invalid", "high": "invalid", "start": "invalid"},
				map[string]interface{}{"target": 120.0, "range": 10.0, "low": 110.0, "high": 130.0, "start": 3600},
			},
			&[]*pump.BloodGlucoseTarget{
				NewTestBloodGlucoseTarget(nil, nil, nil, nil, nil),
				NewTestBloodGlucoseTarget(120.0, 10.0, 110.0, 130.0, 3600),
			}, []*service.Error{
				testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/0/target", nil),
				testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/0/range", nil),
				testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/0/low", nil),
				testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/0/high", nil),
				testData.ComposeError(service.ErrorTypeNotInteger("invalid"), "/0/start", nil),
			}),
		Entry("parses array that has more than one invalid",
			&[]interface{}{
				map[string]interface{}{"target": "invalid-0", "range": "invalid-0", "low": "invalid-0", "high": "invalid-0", "start": "invalid-0"},
				map[string]interface{}{"target": "invalid-1", "range": "invalid-1", "low": "invalid-1", "high": "invalid-1", "start": "invalid-1"},
			},
			&[]*pump.BloodGlucoseTarget{
				NewTestBloodGlucoseTarget(nil, nil, nil, nil, nil),
				NewTestBloodGlucoseTarget(nil, nil, nil, nil, nil),
			}, []*service.Error{
				testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/0/target", nil),
				testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/0/range", nil),
				testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/0/low", nil),
				testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/0/high", nil),
				testData.ComposeError(service.ErrorTypeNotInteger("invalid"), "/0/start", nil),
				testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/1/target", nil),
				testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/1/range", nil),
				testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/1/low", nil),
				testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/1/high", nil),
				testData.ComposeError(service.ErrorTypeNotInteger("invalid"), "/1/start", nil),
			}),
		Entry("parses array that has more than one valid with additional field",
			&[]interface{}{
				map[string]interface{}{"target": 120.0, "range": 10.0, "low": 110.0, "high": 130.0, "start": 3600},
				map[string]interface{}{"target": 121.0, "range": 11.0, "low": 111.0, "high": 131.0, "start": 3601, "additional": 0.0},
			},
			&[]*pump.BloodGlucoseTarget{
				NewTestBloodGlucoseTarget(120.0, 10.0, 110.0, 130.0, 3600),
				NewTestBloodGlucoseTarget(121.0, 11.0, 111.0, 131.0, 3601),
			}, []*service.Error{
				testData.ComposeError(parser.ErrorNotParsed(), "/1/additional", nil),
			}),
	)

	Context("NewBloodGlucoseTarget", func() {
		It("is successful", func() {
			Expect(pump.NewBloodGlucoseTarget()).To(Equal(&pump.BloodGlucoseTarget{}))
		})
	})

	Context("with new blood glucose target", func() {
		DescribeTable("Parse",
			func(sourceObject *map[string]interface{}, expectedTarget *pump.BloodGlucoseTarget, expectedErrors []*service.Error) {
				testContext, err := context.NewStandard(null.NewLogger())
				Expect(err).ToNot(HaveOccurred())
				Expect(testContext).ToNot(BeNil())
				testFactory, err := factory.NewStandard()
				Expect(err).ToNot(HaveOccurred())
				Expect(testFactory).ToNot(BeNil())
				testParser, err := parser.NewStandardObject(testContext, testFactory, sourceObject, parser.AppendErrorNotParsed)
				Expect(err).ToNot(HaveOccurred())
				Expect(testParser).ToNot(BeNil())
				sourceTarget := &pump.BloodGlucoseTarget{}
				sourceTarget.Parse(testParser)
				Expect(sourceTarget).To(Equal(expectedTarget))
				Expect(testContext.Errors()).To(ConsistOf(expectedErrors))
			},
			Entry("parses object that is nil", nil, NewTestBloodGlucoseTarget(nil, nil, nil, nil, nil), []*service.Error{}),
			Entry("parses object that is empty", &map[string]interface{}{}, NewTestBloodGlucoseTarget(nil, nil, nil, nil, nil), []*service.Error{}),
			Entry("parses object that has valid start", &map[string]interface{}{"start": 3600}, NewTestBloodGlucoseTarget(nil, nil, nil, nil, 3600), []*service.Error{}),
			Entry("parses object that has invalid start", &map[string]interface{}{"start": "invalid"}, NewTestBloodGlucoseTarget(nil, nil, nil, nil, nil), []*service.Error{
				testData.ComposeError(service.ErrorTypeNotInteger("invalid"), "/start", nil),
			}),
			Entry("parses object that has multiple valid fields", &map[string]interface{}{"target": 120.0, "range": 10.0, "low": 110.0, "high": 130.0, "start": 3600},
				NewTestBloodGlucoseTarget(120.0, 10.0, 110.0, 130.0, 3600), []*service.Error{}),
			Entry("parses object that has multiple invalid fields", &map[string]interface{}{"target": "invalid", "range": "invalid", "low": "invalid", "high": "invalid", "start": "invalid"},
				NewTestBloodGlucoseTarget(nil, nil, nil, nil, nil), []*service.Error{
					testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/target", nil),
					testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/range", nil),
					testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/low", nil),
					testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/high", nil),
					testData.ComposeError(service.ErrorTypeNotInteger("invalid"), "/start", nil),
				}),
		)

		DescribeTable("Validate",
			func(sourceTarget *pump.BloodGlucoseTarget, sourceUnits interface{}, expectedErrors []*service.Error) {
				testContext, err := context.NewStandard(null.NewLogger())
				Expect(err).ToNot(HaveOccurred())
				Expect(testContext).ToNot(BeNil())
				testValidator, err := validator.NewStandard(testContext)
				Expect(err).ToNot(HaveOccurred())
				Expect(testValidator).ToNot(BeNil())
				sourceTarget.Validate(testValidator, AsStringPointer(sourceUnits))
				Expect(testContext.Errors()).To(ConsistOf(expectedErrors))
			},
			Entry("validates a target with units of mmol/L; target/range; all valid", NewTestBloodGlucoseTarget(6.6, 1.0, nil, nil, 3600), "mmol/L", []*service.Error{}),
			Entry("validates a target with units of mmol/L; target/range; start missing", NewTestBloodGlucoseTarget(6.6, 1.0, nil, nil, nil), "mmol/L", []*service.Error{
				testData.ComposeError(service.ErrorValueNotExists(), "/start", nil),
			}),
			Entry("validates a target with units of mmol/L; target/range; start at lower", NewTestBloodGlucoseTarget(6.6, 1.0, nil, nil, 0), "mmol/L", []*service.Error{}),
			Entry("validates a target with units of mmol/L; target/range; start at upper", NewTestBloodGlucoseTarget(6.6, 1.0, nil, nil, 8640000), "mmol/L", []*service.Error{}),
			Entry("validates a target with units of mmol/L; target/range; start out of range (lower)", NewTestBloodGlucoseTarget(6.6, 1.0, nil, nil, -1), "mmol/L", []*service.Error{
				testData.ComposeError(service.ErrorValueNotInRange(-1, 0, 86400000), "/start", nil),
			}),
			Entry("validates a target with units of mmol/L; target/range; start out of range (upper)", NewTestBloodGlucoseTarget(6.6, 1.0, nil, nil, 86400001), "mmol/L", []*service.Error{
				testData.ComposeError(service.ErrorValueNotInRange(86400001, 0, 86400000), "/start", nil),
			}),
			Entry("validates a target with units of mg/dL; target/range; all valid", NewTestBloodGlucoseTarget(120.0, 10.0, nil, nil, 3600), "mg/dL", []*service.Error{}),
			Entry("validates a target with units of mg/dL; target/range; start at lower", NewTestBloodGlucoseTarget(120.0, 10.0, nil, nil, 0), "mg/dL", []*service.Error{}),
			Entry("validates a target with units of mg/dL; target/range; start at upper", NewTestBloodGlucoseTarget(120.0, 10.0, nil, nil, 86400000), "mg/dL", []*service.Error{}),
			Entry("validates a target with units of mg/dL; target/range; start out of range (lower)", NewTestBloodGlucoseTarget(120.0, 10.0, nil, nil, -1), "mg/dL", []*service.Error{
				testData.ComposeError(service.ErrorValueNotInRange(-1, 0, 86400000), "/start", nil),
			}),
			Entry("validates a target with units of mg/dL; target/range; start out of range (upper)", NewTestBloodGlucoseTarget(120.0, 10.0, nil, nil, 86400001), "mg/dL", []*service.Error{
				testData.ComposeError(service.ErrorValueNotInRange(86400001, 0, 86400000), "/start", nil),
			}),
		)
	})
})
