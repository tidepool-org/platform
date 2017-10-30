package glucose_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"math"

	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/normalizer"
	testData "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/blood/glucose"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/service"
)

func NewMeta() interface{} {
	return &types.Meta{
		Type: "testGlucose",
	}
}

func NewTestGlucose(sourceTime interface{}, sourceUnits interface{}, sourceValue interface{}) *glucose.Glucose {
	testGlucose := &glucose.Glucose{}
	testGlucose.Init()
	testGlucose.Type = "testGlucose"
	testGlucose.DeviceID = pointer.String(id.New())
	if value, ok := sourceTime.(string); ok {
		testGlucose.Time = pointer.String(value)
	}
	if value, ok := sourceUnits.(string); ok {
		testGlucose.Units = pointer.String(value)
	}
	if value, ok := sourceValue.(float64); ok {
		testGlucose.Value = pointer.Float64(value)
	}
	return testGlucose
}

var _ = Describe("Glucose", func() {
	Context("with new glucose", func() {
		var testGlucose *glucose.Glucose

		BeforeEach(func() {
			testGlucose = &glucose.Glucose{}
		})

		Context("with initialized", func() {
			BeforeEach(func() {
				testGlucose.Init()
			})

			DescribeTable("Validate",
				func(sourceGlucose *glucose.Glucose, expectedErrors []*service.Error) {
					testContext, err := context.NewStandard(null.NewLogger())
					Expect(err).ToNot(HaveOccurred())
					Expect(testContext).ToNot(BeNil())
					testValidator, err := validator.NewStandard(testContext)
					Expect(err).ToNot(HaveOccurred())
					Expect(testValidator).ToNot(BeNil())
					Expect(sourceGlucose.Validate(testValidator)).To(Succeed())
					Expect(testContext.Errors()).To(ConsistOf(expectedErrors))
				},
				Entry("all valid",
					NewTestGlucose("2016-09-06T13:45:58-07:00", "mmol/L", 10.0),
					[]*service.Error{}),
				Entry("missing time",
					NewTestGlucose(nil, "mmol/L", 10.0),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotExists(), "/time", NewMeta()),
					}),
				Entry("missing units",
					NewTestGlucose("2016-09-06T13:45:58-07:00", nil, 10.0),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotExists(), "/units", NewMeta()),
					}),
				Entry("unknown units",
					NewTestGlucose("2016-09-06T13:45:58-07:00", "unknown", 10.0),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueStringNotOneOf("unknown", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
					}),
				Entry("mmol/L units",
					NewTestGlucose("2016-09-06T13:45:58-07:00", "mmol/L", 10.0),
					[]*service.Error{}),
				Entry("mmol/l units",
					NewTestGlucose("2016-09-06T13:45:58-07:00", "mmol/l", 10.0),
					[]*service.Error{}),
				Entry("mg/dL units",
					NewTestGlucose("2016-09-06T13:45:58-07:00", "mg/dL", 180.0),
					[]*service.Error{}),
				Entry("mg/dl units",
					NewTestGlucose("2016-09-06T13:45:58-07:00", "mg/dl", 180.0),
					[]*service.Error{}),
				Entry("missing value",
					NewTestGlucose("2016-09-06T13:45:58-07:00", "mmol/L", nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotExists(), "/value", NewMeta()),
					}),
				Entry("unknown units; value in range (lower)",
					NewTestGlucose("2016-09-06T13:45:58-07:00", "unknown", -math.MaxFloat64),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueStringNotOneOf("unknown", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
					}),
				Entry("unknown units; value in range (upper)",
					NewTestGlucose("2016-09-06T13:45:58-07:00", "unknown", math.MaxFloat64),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueStringNotOneOf("unknown", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
					}),
				Entry("mmol/L units; value out of range (lower)",
					NewTestGlucose("2016-09-06T13:45:58-07:00", "mmol/L", -0.1),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotInRange(-0.1, 0.0, 55.0), "/value", NewMeta()),
					}),
				Entry("mmol/L units; value in range (lower)",
					NewTestGlucose("2016-09-06T13:45:58-07:00", "mmol/L", 0.0),
					[]*service.Error{}),
				Entry("mmol/L units; value in range (upper)",
					NewTestGlucose("2016-09-06T13:45:58-07:00", "mmol/L", 55.0),
					[]*service.Error{}),
				Entry("mmol/L units; value out of range (upper)",
					NewTestGlucose("2016-09-06T13:45:58-07:00", "mmol/L", 55.1),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotInRange(55.1, 0.0, 55.0), "/value", NewMeta()),
					}),
				Entry("mmol/l units; value out of range (lower)",
					NewTestGlucose("2016-09-06T13:45:58-07:00", "mmol/l", -0.1),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotInRange(-0.1, 0.0, 55.0), "/value", NewMeta()),
					}),
				Entry("mmol/l units; value in range (lower)",
					NewTestGlucose("2016-09-06T13:45:58-07:00", "mmol/l", 0.0),
					[]*service.Error{}),
				Entry("mmol/l units; value in range (upper)",
					NewTestGlucose("2016-09-06T13:45:58-07:00", "mmol/l", 55.0),
					[]*service.Error{}),
				Entry("mmol/l units; value out of range (upper)",
					NewTestGlucose("2016-09-06T13:45:58-07:00", "mmol/l", 55.1),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotInRange(55.1, 0.0, 55.0), "/value", NewMeta()),
					}),
				Entry("mg/dL units; value out of range (lower)",
					NewTestGlucose("2016-09-06T13:45:58-07:00", "mg/dL", -0.1),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/value", NewMeta()),
					}),
				Entry("mg/dL units; value in range (lower)",
					NewTestGlucose("2016-09-06T13:45:58-07:00", "mg/dL", 0.0),
					[]*service.Error{}),
				Entry("mg/dL units; value in range (upper)",
					NewTestGlucose("2016-09-06T13:45:58-07:00", "mg/dL", 1000.0),
					[]*service.Error{}),
				Entry("mg/dL units; value out of range (upper)",
					NewTestGlucose("2016-09-06T13:45:58-07:00", "mg/dL", 1000.1),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/value", NewMeta()),
					}),
				Entry("mg/dl units; value out of range (lower)",
					NewTestGlucose("2016-09-06T13:45:58-07:00", "mg/dl", -0.1),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/value", NewMeta()),
					}),
				Entry("mg/dl units; value in range (lower)",
					NewTestGlucose("2016-09-06T13:45:58-07:00", "mg/dl", 0.0),
					[]*service.Error{}),
				Entry("mg/dl units; value in range (upper)",
					NewTestGlucose("2016-09-06T13:45:58-07:00", "mg/dl", 1000.0),
					[]*service.Error{}),
				Entry("mg/dl units; value out of range (upper)",
					NewTestGlucose("2016-09-06T13:45:58-07:00", "mg/dl", 1000.1),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/value", NewMeta()),
					}),
				Entry("multiple",
					NewTestGlucose(nil, "unknown", nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotExists(), "/time", NewMeta()),
						testData.ComposeError(service.ErrorValueStringNotOneOf("unknown", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
						testData.ComposeError(service.ErrorValueNotExists(), "/value", NewMeta()),
					}),
			)

			DescribeTable("Normalize",
				func(sourceGlucose *glucose.Glucose, expectedKetone *glucose.Glucose) {
					sourceGlucose.GUID = expectedKetone.GUID
					sourceGlucose.ID = expectedKetone.ID
					sourceGlucose.DeviceID = expectedKetone.DeviceID
					testContext, err := context.NewStandard(null.NewLogger())
					Expect(err).ToNot(HaveOccurred())
					Expect(testContext).ToNot(BeNil())
					testNormalizer, err := normalizer.NewStandard(testContext)
					Expect(err).ToNot(HaveOccurred())
					Expect(testNormalizer).ToNot(BeNil())
					Expect(sourceGlucose.Normalize(testNormalizer)).To(Succeed())
					Expect(sourceGlucose).To(Equal(expectedKetone))
				},
				Entry("unknown units",
					NewTestGlucose("2016-09-06T13:45:58-07:00", "unknown", 10.0),
					NewTestGlucose("2016-09-06T13:45:58-07:00", "unknown", 10.0),
				),
				Entry("mmol/L units",
					NewTestGlucose("2016-09-06T13:45:58-07:00", "mmol/L", 10.0),
					NewTestGlucose("2016-09-06T13:45:58-07:00", "mmol/L", 10.0),
				),
				Entry("mmol/l units",
					NewTestGlucose("2016-09-06T13:45:58-07:00", "mmol/l", 10.0),
					NewTestGlucose("2016-09-06T13:45:58-07:00", "mmol/L", 10.0),
				),
				Entry("mg/dL units",
					NewTestGlucose("2016-09-06T13:45:58-07:00", "mg/dL", 180.0),
					NewTestGlucose("2016-09-06T13:45:58-07:00", "mmol/L", 9.99135),
				),
				Entry("mg/dl units",
					NewTestGlucose("2016-09-06T13:45:58-07:00", "mg/dl", 180.0),
					NewTestGlucose("2016-09-06T13:45:58-07:00", "mmol/L", 9.99135),
				),
			)
		})
	})
})
