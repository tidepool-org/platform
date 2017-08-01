package ketone_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"math"

	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/normalizer"
	testData "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/blood/ketone"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/service"
)

func NewMeta() interface{} {
	return &types.Meta{
		Type: "bloodKetone",
	}
}

func NewTestKetone(sourceTime interface{}, sourceUnits interface{}, sourceValue interface{}) *ketone.Ketone {
	testKetone := ketone.Init()
	testKetone.DeviceID = pointer.String(id.New())
	if value, ok := sourceTime.(string); ok {
		testKetone.Time = pointer.String(value)
	}
	if value, ok := sourceUnits.(string); ok {
		testKetone.Units = pointer.String(value)
	}
	if value, ok := sourceValue.(float64); ok {
		testKetone.Value = pointer.Float(value)
	}
	return testKetone
}

var _ = Describe("Ketone", func() {
	Context("Type", func() {
		It("returns the expected type", func() {
			Expect(ketone.Type()).To(Equal("bloodKetone"))
		})
	})

	Context("NewDatum", func() {
		It("returns the expected datum", func() {
			Expect(ketone.NewDatum()).To(Equal(&ketone.Ketone{}))
		})
	})

	Context("New", func() {
		It("returns the expected ketone", func() {
			Expect(ketone.New()).To(Equal(&ketone.Ketone{}))
		})
	})

	Context("Init", func() {
		It("returns the expected ketone", func() {
			testKetone := ketone.Init()
			Expect(testKetone).ToNot(BeNil())
			Expect(testKetone.ID).ToNot(BeEmpty())
			Expect(testKetone.Type).To(Equal("bloodKetone"))
		})
	})

	Context("with new ketone", func() {
		var testKetone *ketone.Ketone

		BeforeEach(func() {
			testKetone = ketone.New()
			Expect(testKetone).ToNot(BeNil())
		})

		Context("Init", func() {
			It("initializes the ketone", func() {
				testKetone.Init()
				Expect(testKetone.ID).ToNot(BeEmpty())
				Expect(testKetone.Type).To(Equal("bloodKetone"))
			})
		})

		Context("with initialized", func() {
			BeforeEach(func() {
				testKetone.Init()
			})

			DescribeTable("Validate",
				func(sourceKetone *ketone.Ketone, expectedErrors []*service.Error) {
					testContext, err := context.NewStandard(log.NewNull())
					Expect(err).ToNot(HaveOccurred())
					Expect(testContext).ToNot(BeNil())
					testValidator, err := validator.NewStandard(testContext)
					Expect(err).ToNot(HaveOccurred())
					Expect(testValidator).ToNot(BeNil())
					Expect(sourceKetone.Validate(testValidator)).To(Succeed())
					Expect(testContext.Errors()).To(ConsistOf(expectedErrors))
				},
				Entry("all valid",
					NewTestKetone("2016-09-06T13:45:58-07:00", "mmol/L", 1.0),
					[]*service.Error{}),
				Entry("missing time",
					NewTestKetone(nil, "mmol/L", 1.0),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotExists(), "/time", NewMeta()),
					}),
				Entry("missing units",
					NewTestKetone("2016-09-06T13:45:58-07:00", nil, 1.0),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotExists(), "/units", NewMeta()),
					}),
				Entry("unknown units",
					NewTestKetone("2016-09-06T13:45:58-07:00", "unknown", 1.0),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueStringNotOneOf("unknown", []string{"mmol/L", "mmol/l"}), "/units", NewMeta()),
					}),
				Entry("mmol/L units",
					NewTestKetone("2016-09-06T13:45:58-07:00", "mmol/L", 1.0),
					[]*service.Error{}),
				Entry("mmol/l units",
					NewTestKetone("2016-09-06T13:45:58-07:00", "mmol/l", 1.0),
					[]*service.Error{}),
				Entry("mg/dL units",
					NewTestKetone("2016-09-06T13:45:58-07:00", "mg/dL", 1.0),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueStringNotOneOf("mg/dL", []string{"mmol/L", "mmol/l"}), "/units", NewMeta()),
					}),
				Entry("mg/dl units",
					NewTestKetone("2016-09-06T13:45:58-07:00", "mg/dl", 1.0),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueStringNotOneOf("mg/dl", []string{"mmol/L", "mmol/l"}), "/units", NewMeta()),
					}),
				Entry("missing value",
					NewTestKetone("2016-09-06T13:45:58-07:00", "mmol/L", nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotExists(), "/value", NewMeta()),
					}),
				Entry("unknown units; value in range (lower)",
					NewTestKetone("2016-09-06T13:45:58-07:00", "unknown", -math.MaxFloat64),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueStringNotOneOf("unknown", []string{"mmol/L", "mmol/l"}), "/units", NewMeta()),
					}),
				Entry("unknown units; value in range (upper)",
					NewTestKetone("2016-09-06T13:45:58-07:00", "unknown", math.MaxFloat64),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueStringNotOneOf("unknown", []string{"mmol/L", "mmol/l"}), "/units", NewMeta()),
					}),
				Entry("mmol/L units; value out of range (lower)",
					NewTestKetone("2016-09-06T13:45:58-07:00", "mmol/L", -0.1),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotInRange(-0.1, 0.0, 10.0), "/value", NewMeta()),
					}),
				Entry("mmol/L units; value in range (lower)",
					NewTestKetone("2016-09-06T13:45:58-07:00", "mmol/L", 0.0),
					[]*service.Error{}),
				Entry("mmol/L units; value in range (upper)",
					NewTestKetone("2016-09-06T13:45:58-07:00", "mmol/L", 10.0),
					[]*service.Error{}),
				Entry("mmol/L units; value out of range (upper)",
					NewTestKetone("2016-09-06T13:45:58-07:00", "mmol/L", 10.1),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotInRange(10.1, 0.0, 10.0), "/value", NewMeta()),
					}),
				Entry("mmol/l units; value out of range (lower)",
					NewTestKetone("2016-09-06T13:45:58-07:00", "mmol/l", -0.1),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotInRange(-0.1, 0.0, 10.0), "/value", NewMeta()),
					}),
				Entry("mmol/l units; value in range (lower)",
					NewTestKetone("2016-09-06T13:45:58-07:00", "mmol/l", 0.0),
					[]*service.Error{}),
				Entry("mmol/l units; value in range (upper)",
					NewTestKetone("2016-09-06T13:45:58-07:00", "mmol/l", 10.0),
					[]*service.Error{}),
				Entry("mmol/l units; value out of range (upper)",
					NewTestKetone("2016-09-06T13:45:58-07:00", "mmol/l", 10.1),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotInRange(10.1, 0.0, 10.0), "/value", NewMeta()),
					}),
				Entry("mg/dL units; value in range (lower)",
					NewTestKetone("2016-09-06T13:45:58-07:00", "mg/dL", -math.MaxFloat64),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueStringNotOneOf("mg/dL", []string{"mmol/L", "mmol/l"}), "/units", NewMeta()),
					}),
				Entry("mg/dL units; value in range (upper)",
					NewTestKetone("2016-09-06T13:45:58-07:00", "mg/dL", math.MaxFloat64),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueStringNotOneOf("mg/dL", []string{"mmol/L", "mmol/l"}), "/units", NewMeta()),
					}),
				Entry("mg/dl units; value in range (lower)",
					NewTestKetone("2016-09-06T13:45:58-07:00", "mg/dl", -math.MaxFloat64),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueStringNotOneOf("mg/dl", []string{"mmol/L", "mmol/l"}), "/units", NewMeta()),
					}),
				Entry("mg/dl units; value in range (upper)",
					NewTestKetone("2016-09-06T13:45:58-07:00", "mg/dl", math.MaxFloat64),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueStringNotOneOf("mg/dl", []string{"mmol/L", "mmol/l"}), "/units", NewMeta()),
					}),
				Entry("multiple",
					NewTestKetone(nil, "unknown", nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotExists(), "/time", NewMeta()),
						testData.ComposeError(service.ErrorValueStringNotOneOf("unknown", []string{"mmol/L", "mmol/l"}), "/units", NewMeta()),
						testData.ComposeError(service.ErrorValueNotExists(), "/value", NewMeta()),
					}),
			)

			DescribeTable("Normalize",
				func(sourceKetone *ketone.Ketone, expectedKetone *ketone.Ketone) {
					sourceKetone.GUID = expectedKetone.GUID
					sourceKetone.ID = expectedKetone.ID
					sourceKetone.DeviceID = expectedKetone.DeviceID
					testContext, err := context.NewStandard(log.NewNull())
					Expect(err).ToNot(HaveOccurred())
					Expect(testContext).ToNot(BeNil())
					testNormalizer, err := normalizer.NewStandard(testContext)
					Expect(err).ToNot(HaveOccurred())
					Expect(testNormalizer).ToNot(BeNil())
					Expect(sourceKetone.Normalize(testNormalizer)).To(Succeed())
					Expect(sourceKetone).To(Equal(expectedKetone))
				},
				Entry("unknown units",
					NewTestKetone("2016-09-06T13:45:58-07:00", "unknown", 1.0),
					NewTestKetone("2016-09-06T13:45:58-07:00", "unknown", 1.0),
				),
				Entry("mmol/L units",
					NewTestKetone("2016-09-06T13:45:58-07:00", "mmol/L", 1.0),
					NewTestKetone("2016-09-06T13:45:58-07:00", "mmol/L", 1.0),
				),
				Entry("mmol/l units",
					NewTestKetone("2016-09-06T13:45:58-07:00", "mmol/l", 1.0),
					NewTestKetone("2016-09-06T13:45:58-07:00", "mmol/L", 1.0),
				),
				Entry("mg/dL units",
					NewTestKetone("2016-09-06T13:45:58-07:00", "mg/dL", 180.0),
					NewTestKetone("2016-09-06T13:45:58-07:00", "mg/dL", 180.0),
				),
				Entry("mg/dl units",
					NewTestKetone("2016-09-06T13:45:58-07:00", "mg/dl", 180.0),
					NewTestKetone("2016-09-06T13:45:58-07:00", "mg/dl", 180.0),
				),
			)
		})
	})
})
