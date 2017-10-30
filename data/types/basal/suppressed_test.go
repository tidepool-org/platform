package basal_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/factory"
	"github.com/tidepool-org/platform/data/parser"
	testData "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/basal"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/service"
)

func NewTestSuppressed(sourceType interface{}, sourceDeliveryType interface{}, sourceRate interface{}, sourceScheduleName interface{}, sourceAnnotations *[]interface{}, sourceSuppressed *basal.Suppressed) *basal.Suppressed {
	testSuppressed := &basal.Suppressed{}
	if value, ok := sourceType.(string); ok {
		testSuppressed.Type = pointer.String(value)
	}
	if value, ok := sourceDeliveryType.(string); ok {
		testSuppressed.DeliveryType = pointer.String(value)
	}
	if value, ok := sourceRate.(float64); ok {
		testSuppressed.Rate = pointer.Float64(value)
	}
	if value, ok := sourceScheduleName.(string); ok {
		testSuppressed.ScheduleName = pointer.String(value)
	}
	testSuppressed.Annotations = sourceAnnotations
	testSuppressed.Suppressed = sourceSuppressed
	return testSuppressed
}

var _ = Describe("Target", func() {
	DescribeTable("ParseSuppressed",
		func(sourceObject *map[string]interface{}, expectedSuppressed *basal.Suppressed, expectedErrors []*service.Error) {
			testContext, err := context.NewStandard(null.NewLogger())
			Expect(err).ToNot(HaveOccurred())
			Expect(testContext).ToNot(BeNil())
			testFactory, err := factory.NewStandard()
			Expect(err).ToNot(HaveOccurred())
			Expect(testFactory).ToNot(BeNil())
			testParser, err := parser.NewStandardObject(testContext, testFactory, sourceObject, parser.AppendErrorNotParsed)
			Expect(err).ToNot(HaveOccurred())
			Expect(testParser).ToNot(BeNil())
			Expect(basal.ParseSuppressed(testParser)).To(Equal(expectedSuppressed))
			Expect(testContext.Errors()).To(ConsistOf(expectedErrors))
		},
		Entry("parses object that is nil", nil, nil, []*service.Error{}),
		Entry("parses object that is empty",
			&map[string]interface{}{},
			NewTestSuppressed(nil, nil, nil, nil, nil, nil),
			[]*service.Error{}),
		Entry("parses object that has multiple valid fields",
			&map[string]interface{}{"type": "basal", "deliveryType": "temp", "rate": 2.0, "suppressed": map[string]interface{}{"type": "basal", "deliveryType": "scheduled", "rate": 1.0, "scheduleName": "Weekday", "annotations": []interface{}{"one", "two", "three"}}},
			NewTestSuppressed("basal", "temp", 2.0, nil, nil, NewTestSuppressed("basal", "scheduled", 1.0, "Weekday", &[]interface{}{"one", "two", "three"}, nil)),
			[]*service.Error{}),
		Entry("parses object that has multiple invalid fields",
			&map[string]interface{}{"type": 0, "deliveryType": 0, "rate": "invalid", "scheduleName": 0, "suppressed": 0},
			NewTestSuppressed(nil, nil, nil, nil, nil, nil),
			[]*service.Error{
				testData.ComposeError(service.ErrorTypeNotString(0), "/type", nil),
				testData.ComposeError(service.ErrorTypeNotString(0), "/deliveryType", nil),
				testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/rate", nil),
				testData.ComposeError(service.ErrorTypeNotString(0), "/scheduleName", nil),
				testData.ComposeError(service.ErrorTypeNotObject(0), "/suppressed", nil),
			}),
		Entry("parses object that has additional fields",
			&map[string]interface{}{"type": "basal", "deliveryType": "temp", "rate": 2.0, "suppressed": map[string]interface{}{"type": "basal", "deliveryType": "scheduled", "rate": 1.0, "scheduleName": "Weekday", "annotations": []interface{}{"one", "two", "three"}}, "additional": 0.0},
			NewTestSuppressed("basal", "temp", 2.0, nil, nil, NewTestSuppressed("basal", "scheduled", 1.0, "Weekday", &[]interface{}{"one", "two", "three"}, nil)),
			[]*service.Error{
				testData.ComposeError(parser.ErrorNotParsed(), "/additional", nil),
			}),
	)

	Context("NewSuppressed", func() {
		It("is successful", func() {
			Expect(basal.NewSuppressed()).To(Equal(&basal.Suppressed{}))
		})
	})

	Context("with new suppressed", func() {
		DescribeTable("Parse",
			func(sourceObject *map[string]interface{}, expectedSuppressed *basal.Suppressed, expectedErrors []*service.Error) {
				testContext, err := context.NewStandard(null.NewLogger())
				Expect(err).ToNot(HaveOccurred())
				Expect(testContext).ToNot(BeNil())
				testFactory, err := factory.NewStandard()
				Expect(err).ToNot(HaveOccurred())
				Expect(testFactory).ToNot(BeNil())
				testParser, err := parser.NewStandardObject(testContext, testFactory, sourceObject, parser.AppendErrorNotParsed)
				Expect(err).ToNot(HaveOccurred())
				Expect(testParser).ToNot(BeNil())
				sourceSuppressed := &basal.Suppressed{}
				sourceSuppressed.Parse(testParser)
				Expect(sourceSuppressed).To(Equal(expectedSuppressed))
				Expect(testContext.Errors()).To(ConsistOf(expectedErrors))
			},
			Entry("parses object that is nil",
				nil,
				NewTestSuppressed(nil, nil, nil, nil, nil, nil),
				[]*service.Error{}),
			Entry("parses object that is empty",
				&map[string]interface{}{},
				NewTestSuppressed(nil, nil, nil, nil, nil, nil),
				[]*service.Error{}),
			Entry("parses object that has valid type",
				&map[string]interface{}{"type": "basal"},
				NewTestSuppressed("basal", nil, nil, nil, nil, nil),
				[]*service.Error{}),
			Entry("parses object that has invalid type",
				&map[string]interface{}{"type": 0},
				NewTestSuppressed(nil, nil, nil, nil, nil, nil),
				[]*service.Error{
					testData.ComposeError(service.ErrorTypeNotString(0), "/type", nil),
				}),
			Entry("parses object that has valid delivery type",
				&map[string]interface{}{"deliveryType": "temp"},
				NewTestSuppressed(nil, "temp", nil, nil, nil, nil),
				[]*service.Error{}),
			Entry("parses object that has invalid delivery type",
				&map[string]interface{}{"deliveryType": 0},
				NewTestSuppressed(nil, nil, nil, nil, nil, nil),
				[]*service.Error{
					testData.ComposeError(service.ErrorTypeNotString(0), "/deliveryType", nil),
				}),
			Entry("parses object that has valid rate",
				&map[string]interface{}{"rate": 2.0},
				NewTestSuppressed(nil, nil, 2.0, nil, nil, nil),
				[]*service.Error{}),
			Entry("parses object that has invalid rate",
				&map[string]interface{}{"rate": "invalid"},
				NewTestSuppressed(nil, nil, nil, nil, nil, nil),
				[]*service.Error{
					testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/rate", nil),
				}),
			Entry("parses object that has valid schedule name",
				&map[string]interface{}{"scheduleName": "Weekday"},
				NewTestSuppressed(nil, nil, nil, "Weekday", nil, nil),
				[]*service.Error{}),
			Entry("parses object that has invalid schedule name",
				&map[string]interface{}{"scheduleName": 0},
				NewTestSuppressed(nil, nil, nil, nil, nil, nil),
				[]*service.Error{
					testData.ComposeError(service.ErrorTypeNotString(0), "/scheduleName", nil),
				}),
			Entry("parses object that has valid annotations",
				&map[string]interface{}{"annotations": []interface{}{"a", "b", "c"}},
				NewTestSuppressed(nil, nil, nil, nil, &[]interface{}{"a", "b", "c"}, nil),
				[]*service.Error{}),
			Entry("parses object that has valid suppressed",
				&map[string]interface{}{"suppressed": map[string]interface{}{}},
				NewTestSuppressed(nil, nil, nil, nil, nil, NewTestSuppressed(nil, nil, nil, nil, nil, nil)),
				[]*service.Error{}),
			Entry("parses object that has valid suppressed that has multiple valid fields",
				&map[string]interface{}{"suppressed": map[string]interface{}{"type": "basal", "deliveryType": "scheduled", "rate": 1.0, "scheduleName": "Weekday"}},
				NewTestSuppressed(nil, nil, nil, nil, nil, NewTestSuppressed("basal", "scheduled", 1.0, "Weekday", nil, nil)),
				[]*service.Error{}),
			Entry("parses object that has invalid suppressed",
				&map[string]interface{}{"suppressed": 0},
				NewTestSuppressed(nil, nil, nil, nil, nil, nil),
				[]*service.Error{
					testData.ComposeError(service.ErrorTypeNotObject(0), "/suppressed", nil),
				}),
			Entry("parses object that has multiple valid fields",
				&map[string]interface{}{"type": "basal", "deliveryType": "temp", "rate": 2.0, "annotations": []interface{}{"a", "b", "c"}, "suppressed": map[string]interface{}{"type": "basal", "deliveryType": "scheduled", "rate": 1.0, "scheduleName": "Weekday", "annotations": []interface{}{"d", "e", "f"}}},
				NewTestSuppressed("basal", "temp", 2.0, nil, &[]interface{}{"a", "b", "c"}, NewTestSuppressed("basal", "scheduled", 1.0, "Weekday", &[]interface{}{"d", "e", "f"}, nil)),
				[]*service.Error{}),
			Entry("parses object that has multiple invalid fields",
				&map[string]interface{}{"type": 0, "deliveryType": 0, "rate": "invalid", "scheduleName": 0, "suppressed": 0},
				NewTestSuppressed(nil, nil, nil, nil, nil, nil),
				[]*service.Error{
					testData.ComposeError(service.ErrorTypeNotString(0), "/type", nil),
					testData.ComposeError(service.ErrorTypeNotString(0), "/deliveryType", nil),
					testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/rate", nil),
					testData.ComposeError(service.ErrorTypeNotString(0), "/scheduleName", nil),
					testData.ComposeError(service.ErrorTypeNotObject(0), "/suppressed", nil),
				}),
		)

		DescribeTable("Validate",
			func(sourceSuppressed *basal.Suppressed, allowedDeliveryTypes []string, expectedErrors []*service.Error) {
				testContext, err := context.NewStandard(null.NewLogger())
				Expect(err).ToNot(HaveOccurred())
				Expect(testContext).ToNot(BeNil())
				testValidator, err := validator.NewStandard(testContext)
				Expect(err).ToNot(HaveOccurred())
				Expect(testValidator).ToNot(BeNil())
				sourceSuppressed.Validate(testValidator, allowedDeliveryTypes)
				Expect(testContext.Errors()).To(ConsistOf(expectedErrors))
			},
			Entry("validates a suppressed with type scheduled; all valid",
				NewTestSuppressed("basal", "scheduled", 1.0, "Weekday", nil, nil), []string{"scheduled"},
				[]*service.Error{}),
			Entry("validates a suppressed with type scheduled; type missing",
				NewTestSuppressed(nil, "scheduled", 1.0, "Weekday", nil, nil), []string{"scheduled"},
				[]*service.Error{
					testData.ComposeError(service.ErrorValueNotExists(), "/type", nil),
				}),
			Entry("validates a suppressed with type scheduled; type not basal",
				NewTestSuppressed("invalid", "scheduled", 1.0, "Weekday", nil, nil), []string{"scheduled"},
				[]*service.Error{
					testData.ComposeError(service.ErrorValueNotEqualTo("invalid", "basal"), "/type", nil),
				}),
			Entry("validates a suppressed with type scheduled; deliveryType missing",
				NewTestSuppressed("basal", nil, 1.0, "Weekday", nil, nil), []string{"scheduled"},
				[]*service.Error{
					testData.ComposeError(service.ErrorValueNotExists(), "/deliveryType", nil),
				}),
			Entry("validates a suppressed with type scheduled; deliveryType not allowed",
				NewTestSuppressed("basal", "invalid", 1.0, "Weekday", nil, nil), []string{"scheduled"},
				[]*service.Error{
					testData.ComposeError(service.ErrorValueStringNotOneOf("invalid", []string{"scheduled"}), "/deliveryType", nil),
				}),
			Entry("validates a suppressed with type scheduled; rate missing",
				NewTestSuppressed("basal", "scheduled", nil, "Weekday", nil, nil), []string{"scheduled"},
				[]*service.Error{
					testData.ComposeError(service.ErrorValueNotExists(), "/rate", nil),
				}),
			Entry("validates a suppressed with type scheduled; rate out of range (lower)",
				NewTestSuppressed("basal", "scheduled", -0.1, "Weekday", nil, nil), []string{"scheduled"},
				[]*service.Error{
					testData.ComposeError(service.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/rate", nil),
				}),
			Entry("validates a suppressed with type scheduled; rate at limit (lower)",
				NewTestSuppressed("basal", "scheduled", 0.0, "Weekday", nil, nil), []string{"scheduled"},
				[]*service.Error{}),
			Entry("validates a suppressed with type scheduled; rate at limit (upper)",
				NewTestSuppressed("basal", "scheduled", 100.0, "Weekday", nil, nil), []string{"scheduled"},
				[]*service.Error{}),
			Entry("validates a suppressed with type scheduled; rate out of range (upper)",
				NewTestSuppressed("basal", "scheduled", 100.1, "Weekday", nil, nil), []string{"scheduled"},
				[]*service.Error{
					testData.ComposeError(service.ErrorValueNotInRange(100.1, 0.0, 100.0), "/rate", nil),
				}),
			Entry("validates a suppressed with type scheduled; schedule name empty",
				NewTestSuppressed("basal", "scheduled", 1.0, "", nil, nil), []string{"scheduled"},
				[]*service.Error{
					testData.ComposeError(service.ErrorValueEmpty(), "/scheduleName", nil),
				}),
			Entry("validates a suppressed with type scheduled; suppressed exists",
				NewTestSuppressed("basal", "scheduled", 1.0, "Weekday", nil, NewTestSuppressed(nil, nil, nil, nil, nil, nil)), []string{"scheduled"},
				[]*service.Error{
					testData.ComposeError(service.ErrorValueExists(), "/suppressed", nil),
				}),
			Entry("validates a suppressed with type scheduled; multiple",
				NewTestSuppressed("invalid", "scheduled", 100.1, "", nil, NewTestSuppressed(nil, nil, nil, nil, nil, nil)), []string{"scheduled"},
				[]*service.Error{
					testData.ComposeError(service.ErrorValueNotEqualTo("invalid", "basal"), "/type", nil),
					testData.ComposeError(service.ErrorValueNotInRange(100.1, 0.0, 100.0), "/rate", nil),
					testData.ComposeError(service.ErrorValueEmpty(), "/scheduleName", nil),
					testData.ComposeError(service.ErrorValueExists(), "/suppressed", nil),
				}),
			Entry("validates a suppressed with type temp; all valid",
				NewTestSuppressed("basal", "temp", 2.0, nil, nil, NewTestSuppressed("basal", "scheduled", 1.0, "Weekday", nil, nil)), []string{"scheduled", "temp"},
				[]*service.Error{}),
			Entry("validates a suppressed with type temp; type missing",
				NewTestSuppressed(nil, "temp", 2.0, nil, nil, NewTestSuppressed("basal", "scheduled", 1.0, "Weekday", nil, nil)), []string{"scheduled", "temp"},
				[]*service.Error{
					testData.ComposeError(service.ErrorValueNotExists(), "/type", nil),
				}),
			Entry("validates a suppressed with type temp; type not basal",
				NewTestSuppressed("invalid", "temp", 2.0, nil, nil, NewTestSuppressed("basal", "scheduled", 1.0, "Weekday", nil, nil)), []string{"scheduled", "temp"},
				[]*service.Error{
					testData.ComposeError(service.ErrorValueNotEqualTo("invalid", "basal"), "/type", nil),
				}),
			Entry("validates a suppressed with type temp; deliveryType missing",
				NewTestSuppressed("basal", nil, 2.0, nil, nil, NewTestSuppressed("basal", "scheduled", 1.0, "Weekday", nil, nil)), []string{"scheduled", "temp"},
				[]*service.Error{
					testData.ComposeError(service.ErrorValueNotExists(), "/deliveryType", nil),
				}),
			Entry("validates a suppressed with type temp; deliveryType not allowed",
				NewTestSuppressed("basal", "invalid", 2.0, "Weekday", nil, NewTestSuppressed("basal", "scheduled", 1.0, "Weekday", nil, nil)), []string{"scheduled", "temp"},
				[]*service.Error{
					testData.ComposeError(service.ErrorValueStringNotOneOf("invalid", []string{"scheduled", "temp"}), "/deliveryType", nil),
				}),
			Entry("validates a suppressed with type temp; rate missing",
				NewTestSuppressed("basal", "temp", nil, nil, nil, NewTestSuppressed("basal", "scheduled", 1.0, "Weekday", nil, nil)), []string{"scheduled", "temp"},
				[]*service.Error{
					testData.ComposeError(service.ErrorValueNotExists(), "/rate", nil),
				}),
			Entry("validates a suppressed with type temp; rate out of range (lower)",
				NewTestSuppressed("basal", "temp", -0.1, nil, nil, NewTestSuppressed("basal", "scheduled", 1.0, "Weekday", nil, nil)), []string{"scheduled", "temp"},
				[]*service.Error{
					testData.ComposeError(service.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/rate", nil),
				}),
			Entry("validates a suppressed with type temp; rate at limit (lower)",
				NewTestSuppressed("basal", "temp", 0.0, nil, nil, NewTestSuppressed("basal", "scheduled", 1.0, "Weekday", nil, nil)), []string{"scheduled", "temp"},
				[]*service.Error{}),
			Entry("validates a suppressed with type temp; rate at limit (upper)",
				NewTestSuppressed("basal", "temp", 100.0, nil, nil, NewTestSuppressed("basal", "scheduled", 1.0, "Weekday", nil, nil)), []string{"scheduled", "temp"},
				[]*service.Error{}),
			Entry("validates a suppressed with type temp; rate out of range (upper)",
				NewTestSuppressed("basal", "temp", 100.1, nil, nil, NewTestSuppressed("basal", "scheduled", 1.0, "Weekday", nil, nil)), []string{"scheduled", "temp"},
				[]*service.Error{
					testData.ComposeError(service.ErrorValueNotInRange(100.1, 0.0, 100.0), "/rate", nil),
				}),
			Entry("validates a suppressed with type temp; schedule name exists",
				NewTestSuppressed("basal", "temp", 2.0, "Weekday", nil, NewTestSuppressed("basal", "scheduled", 1.0, "Weekday", nil, nil)), []string{"scheduled", "temp"},
				[]*service.Error{
					testData.ComposeError(service.ErrorValueExists(), "/scheduleName", nil),
				}),
			Entry("validates a suppressed with type temp; suppressed missing",
				NewTestSuppressed("basal", "temp", 2.0, nil, nil, nil), []string{"scheduled", "temp"},
				[]*service.Error{
					testData.ComposeError(service.ErrorValueNotExists(), "/suppressed", nil),
				}),
			Entry("validates a suppressed with type temp; multiple",
				NewTestSuppressed("invalid", "temp", 100.1, "Weekday", nil, nil), []string{"scheduled", "temp"},
				[]*service.Error{
					testData.ComposeError(service.ErrorValueNotEqualTo("invalid", "basal"), "/type", nil),
					testData.ComposeError(service.ErrorValueNotInRange(100.1, 0.0, 100.0), "/rate", nil),
					testData.ComposeError(service.ErrorValueExists(), "/scheduleName", nil),
					testData.ComposeError(service.ErrorValueNotExists(), "/suppressed", nil),
				}),
			Entry("validates a suppressed with type temp; suppressed type missing",
				NewTestSuppressed("basal", "temp", 2.0, nil, nil, NewTestSuppressed(nil, "scheduled", 1.0, "Weekday", nil, nil)), []string{"scheduled", "temp"},
				[]*service.Error{
					testData.ComposeError(service.ErrorValueNotExists(), "/suppressed/type", nil),
				}),
			Entry("validates a suppressed with type temp; suppressed type not basal",
				NewTestSuppressed("basal", "temp", 2.0, nil, nil, NewTestSuppressed("invalid", "scheduled", 1.0, "Weekday", nil, nil)), []string{"scheduled", "temp"},
				[]*service.Error{
					testData.ComposeError(service.ErrorValueNotEqualTo("invalid", "basal"), "/suppressed/type", nil),
				}),
			Entry("validates a suppressed with type temp; suppressed deliveryType missing",
				NewTestSuppressed("basal", "temp", 2.0, nil, nil, NewTestSuppressed("basal", nil, 1.0, "Weekday", nil, nil)), []string{"scheduled", "temp"},
				[]*service.Error{
					testData.ComposeError(service.ErrorValueNotExists(), "/suppressed/deliveryType", nil),
				}),
			Entry("validates a suppressed with type temp; suppressed deliveryType not allowed",
				NewTestSuppressed("basal", "temp", 2.0, nil, nil, NewTestSuppressed("basal", "temp", 1.0, "Weekday", nil, nil)), []string{"scheduled", "temp"},
				[]*service.Error{
					testData.ComposeError(service.ErrorValueStringNotOneOf("temp", []string{"scheduled"}), "/suppressed/deliveryType", nil),
				}),
			Entry("validates a suppressed with type temp; suppressed rate missing",
				NewTestSuppressed("basal", "temp", 2.0, nil, nil, NewTestSuppressed("basal", "scheduled", nil, "Weekday", nil, nil)), []string{"scheduled", "temp"},
				[]*service.Error{
					testData.ComposeError(service.ErrorValueNotExists(), "/suppressed/rate", nil),
				}),
			Entry("validates a suppressed with type temp; suppressed rate out of range (lower)",
				NewTestSuppressed("basal", "temp", 2.0, nil, nil, NewTestSuppressed("basal", "scheduled", -0.1, "Weekday", nil, nil)), []string{"scheduled", "temp"},
				[]*service.Error{
					testData.ComposeError(service.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/suppressed/rate", nil),
				}),
			Entry("validates a suppressed with type temp; suppressed rate at limit (lower)",
				NewTestSuppressed("basal", "temp", 2.0, nil, nil, NewTestSuppressed("basal", "scheduled", 0.0, "Weekday", nil, nil)), []string{"scheduled", "temp"},
				[]*service.Error{}),
			Entry("validates a suppressed with type temp; suppressed rate at limit (upper)",
				NewTestSuppressed("basal", "temp", 2.0, nil, nil, NewTestSuppressed("basal", "scheduled", 100.0, "Weekday", nil, nil)), []string{"scheduled", "temp"},
				[]*service.Error{}),
			Entry("validates a suppressed with type temp; suppressed rate out of range (upper)",
				NewTestSuppressed("basal", "temp", 2.0, nil, nil, NewTestSuppressed("basal", "scheduled", 100.1, "Weekday", nil, nil)), []string{"scheduled", "temp"},
				[]*service.Error{
					testData.ComposeError(service.ErrorValueNotInRange(100.1, 0.0, 100.0), "/suppressed/rate", nil),
				}),
			Entry("validates a suppressed with type temp; suppressed schedule name empty",
				NewTestSuppressed("basal", "temp", 2.0, nil, nil, NewTestSuppressed("basal", "scheduled", 1.0, "", nil, nil)), []string{"scheduled", "temp"},
				[]*service.Error{
					testData.ComposeError(service.ErrorValueEmpty(), "/suppressed/scheduleName", nil),
				}),
			Entry("validates a suppressed with type temp; suppressed suppressed exists",
				NewTestSuppressed("basal", "temp", 2.0, nil, nil, NewTestSuppressed("basal", "scheduled", 1.0, "Weekday", nil, NewTestSuppressed(nil, nil, nil, nil, nil, nil))), []string{"scheduled", "temp"},
				[]*service.Error{
					testData.ComposeError(service.ErrorValueExists(), "/suppressed/suppressed", nil),
				}),
			Entry("validates a suppressed with type temp; suppressed multiple",
				NewTestSuppressed("basal", "temp", 2.0, nil, nil, NewTestSuppressed("invalid", "scheduled", 100.1, "", nil, NewTestSuppressed(nil, nil, nil, nil, nil, nil))), []string{"scheduled", "temp"},
				[]*service.Error{
					testData.ComposeError(service.ErrorValueNotEqualTo("invalid", "basal"), "/suppressed/type", nil),
					testData.ComposeError(service.ErrorValueNotInRange(100.1, 0.0, 100.0), "/suppressed/rate", nil),
					testData.ComposeError(service.ErrorValueEmpty(), "/suppressed/scheduleName", nil),
					testData.ComposeError(service.ErrorValueExists(), "/suppressed/suppressed", nil),
				}),
		)
	})

	Context("HasDeliveryTypeOneOf", func() {
		var suppressed *basal.Suppressed

		BeforeEach(func() {
			suppressed = basal.NewSuppressed()
		})

		It("returns false if suppressed delivery type is nil", func() {
			Expect(suppressed.HasDeliveryTypeOneOf([]string{"one", "two", "three"})).To(BeFalse())
		})

		DescribeTable("returns expected result when",
			func(suppressedDeliveryType string, deliveryTypes []string, expectedResult bool) {
				suppressed.DeliveryType = pointer.String(suppressedDeliveryType)
				Expect(suppressed.HasDeliveryTypeOneOf(deliveryTypes)).To(Equal(expectedResult))
			},
			Entry("is nil delivery type string array", "two", nil, false),
			Entry("is single delivery type string array", "two", []string{}, false),
			Entry("is single invalid delivery type string array", "two", []string{"one"}, false),
			Entry("is single valid delivery type string array", "two", []string{"two"}, true),
			Entry("is multiple invalid delivery type string array", "two", []string{"one", "three"}, false),
			Entry("is multiple invalid and valid delivery type string array", "two", []string{"one", "two", "three", "four"}, true),
		)
	})
})
