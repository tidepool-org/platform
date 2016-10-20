package basal_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/factory"
	"github.com/tidepool-org/platform/data/parser"
	"github.com/tidepool-org/platform/data/types/base/basal"
	"github.com/tidepool-org/platform/data/types/base/testing"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/service"
)

func NewTestSuppressed(sourceType interface{}, sourceDeliveryType interface{}, sourceRate interface{}, sourceScheduleName interface{}, sourceSuppressed *basal.Suppressed) *basal.Suppressed {
	testSuppressed := &basal.Suppressed{}
	if value, ok := sourceType.(string); ok {
		testSuppressed.Type = app.StringAsPointer(value)
	}
	if value, ok := sourceDeliveryType.(string); ok {
		testSuppressed.DeliveryType = app.StringAsPointer(value)
	}
	if value, ok := sourceRate.(float64); ok {
		testSuppressed.Rate = app.FloatAsPointer(value)
	}
	if value, ok := sourceScheduleName.(string); ok {
		testSuppressed.ScheduleName = app.StringAsPointer(value)
	}
	testSuppressed.Suppressed = sourceSuppressed
	return testSuppressed
}

var _ = Describe("Target", func() {
	DescribeTable("ParseSuppressed",
		func(sourceObject *map[string]interface{}, expectedSuppressed *basal.Suppressed, expectedErrors []*service.Error) {
			testContext, err := context.NewStandard(log.NewNull())
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
			NewTestSuppressed(nil, nil, nil, nil, nil),
			[]*service.Error{}),
		Entry("parses object that has multiple valid fields",
			&map[string]interface{}{"type": "basal", "deliveryType": "temp", "rate": 2.0, "suppressed": map[string]interface{}{"type": "basal", "deliveryType": "scheduled", "rate": 1.0, "scheduleName": "Weekday"}},
			NewTestSuppressed("basal", "temp", 2.0, nil, NewTestSuppressed("basal", "scheduled", 1.0, "Weekday", nil)),
			[]*service.Error{}),
		Entry("parses object that has multiple invalid fields",
			&map[string]interface{}{"type": 0, "deliveryType": 0, "rate": "invalid", "scheduleName": 0, "suppressed": 0},
			NewTestSuppressed(nil, nil, nil, nil, nil),
			[]*service.Error{
				testing.ComposeError(service.ErrorTypeNotString(0), "/type", nil),
				testing.ComposeError(service.ErrorTypeNotString(0), "/deliveryType", nil),
				testing.ComposeError(service.ErrorTypeNotFloat("invalid"), "/rate", nil),
				testing.ComposeError(service.ErrorTypeNotString(0), "/scheduleName", nil),
				testing.ComposeError(service.ErrorTypeNotObject(0), "/suppressed", nil),
			}),
		Entry("parses object that has additional fields",
			&map[string]interface{}{"type": "basal", "deliveryType": "temp", "rate": 2.0, "suppressed": map[string]interface{}{"type": "basal", "deliveryType": "scheduled", "rate": 1.0, "scheduleName": "Weekday"}, "additional": 0.0},
			NewTestSuppressed("basal", "temp", 2.0, nil, NewTestSuppressed("basal", "scheduled", 1.0, "Weekday", nil)),
			[]*service.Error{
				testing.ComposeError(parser.ErrorNotParsed(), "/additional", nil),
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
				testContext, err := context.NewStandard(log.NewNull())
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
				NewTestSuppressed(nil, nil, nil, nil, nil),
				[]*service.Error{}),
			Entry("parses object that is empty",
				&map[string]interface{}{},
				NewTestSuppressed(nil, nil, nil, nil, nil),
				[]*service.Error{}),
			Entry("parses object that has valid type",
				&map[string]interface{}{"type": "basal"},
				NewTestSuppressed("basal", nil, nil, nil, nil),
				[]*service.Error{}),
			Entry("parses object that has invalid type",
				&map[string]interface{}{"type": 0},
				NewTestSuppressed(nil, nil, nil, nil, nil),
				[]*service.Error{
					testing.ComposeError(service.ErrorTypeNotString(0), "/type", nil),
				}),
			Entry("parses object that has valid delivery type",
				&map[string]interface{}{"deliveryType": "temp"},
				NewTestSuppressed(nil, "temp", nil, nil, nil),
				[]*service.Error{}),
			Entry("parses object that has invalid delivery type",
				&map[string]interface{}{"deliveryType": 0},
				NewTestSuppressed(nil, nil, nil, nil, nil),
				[]*service.Error{
					testing.ComposeError(service.ErrorTypeNotString(0), "/deliveryType", nil),
				}),
			Entry("parses object that has valid rate",
				&map[string]interface{}{"rate": 2.0},
				NewTestSuppressed(nil, nil, 2.0, nil, nil),
				[]*service.Error{}),
			Entry("parses object that has invalid rate",
				&map[string]interface{}{"rate": "invalid"},
				NewTestSuppressed(nil, nil, nil, nil, nil),
				[]*service.Error{
					testing.ComposeError(service.ErrorTypeNotFloat("invalid"), "/rate", nil),
				}),
			Entry("parses object that has valid schedule name",
				&map[string]interface{}{"scheduleName": "Weekday"},
				NewTestSuppressed(nil, nil, nil, "Weekday", nil),
				[]*service.Error{}),
			Entry("parses object that has invalid schedule name",
				&map[string]interface{}{"scheduleName": 0},
				NewTestSuppressed(nil, nil, nil, nil, nil),
				[]*service.Error{
					testing.ComposeError(service.ErrorTypeNotString(0), "/scheduleName", nil),
				}),
			Entry("parses object that has valid suppressed",
				&map[string]interface{}{"suppressed": map[string]interface{}{}},
				NewTestSuppressed(nil, nil, nil, nil, NewTestSuppressed(nil, nil, nil, nil, nil)),
				[]*service.Error{}),
			Entry("parses object that has valid suppressed that has multiple valid fields",
				&map[string]interface{}{"suppressed": map[string]interface{}{"type": "basal", "deliveryType": "scheduled", "rate": 1.0, "scheduleName": "Weekday"}},
				NewTestSuppressed(nil, nil, nil, nil, NewTestSuppressed("basal", "scheduled", 1.0, "Weekday", nil)),
				[]*service.Error{}),
			Entry("parses object that has invalid suppressed",
				&map[string]interface{}{"suppressed": 0},
				NewTestSuppressed(nil, nil, nil, nil, nil),
				[]*service.Error{
					testing.ComposeError(service.ErrorTypeNotObject(0), "/suppressed", nil),
				}),
			Entry("parses object that has multiple valid fields",
				&map[string]interface{}{"type": "basal", "deliveryType": "temp", "rate": 2.0, "suppressed": map[string]interface{}{"type": "basal", "deliveryType": "scheduled", "rate": 1.0, "scheduleName": "Weekday"}},
				NewTestSuppressed("basal", "temp", 2.0, nil, NewTestSuppressed("basal", "scheduled", 1.0, "Weekday", nil)),
				[]*service.Error{}),
			Entry("parses object that has multiple invalid fields",
				&map[string]interface{}{"type": 0, "deliveryType": 0, "rate": "invalid", "scheduleName": 0, "suppressed": 0},
				NewTestSuppressed(nil, nil, nil, nil, nil),
				[]*service.Error{
					testing.ComposeError(service.ErrorTypeNotString(0), "/type", nil),
					testing.ComposeError(service.ErrorTypeNotString(0), "/deliveryType", nil),
					testing.ComposeError(service.ErrorTypeNotFloat("invalid"), "/rate", nil),
					testing.ComposeError(service.ErrorTypeNotString(0), "/scheduleName", nil),
					testing.ComposeError(service.ErrorTypeNotObject(0), "/suppressed", nil),
				}),
		)

		DescribeTable("Validate",
			func(sourceSuppressed *basal.Suppressed, allowedDeliveryTypes []string, expectedErrors []*service.Error) {
				testContext, err := context.NewStandard(log.NewNull())
				Expect(err).ToNot(HaveOccurred())
				Expect(testContext).ToNot(BeNil())
				testValidator, err := validator.NewStandard(testContext)
				Expect(err).ToNot(HaveOccurred())
				Expect(testValidator).ToNot(BeNil())
				sourceSuppressed.Validate(testValidator, allowedDeliveryTypes)
				Expect(testContext.Errors()).To(ConsistOf(expectedErrors))
			},
			Entry("validates a suppressed with type scheduled; all valid",
				NewTestSuppressed("basal", "scheduled", 1.0, "Weekday", nil), []string{"scheduled"},
				[]*service.Error{}),
			Entry("validates a suppressed with type scheduled; type missing",
				NewTestSuppressed(nil, "scheduled", 1.0, "Weekday", nil), []string{"scheduled"},
				[]*service.Error{
					testing.ComposeError(service.ErrorValueNotExists(), "/type", nil),
				}),
			Entry("validates a suppressed with type scheduled; type not basal",
				NewTestSuppressed("invalid", "scheduled", 1.0, "Weekday", nil), []string{"scheduled"},
				[]*service.Error{
					testing.ComposeError(service.ErrorValueNotEqualTo("invalid", "basal"), "/type", nil),
				}),
			Entry("validates a suppressed with type scheduled; deliveryType missing",
				NewTestSuppressed("basal", nil, 1.0, "Weekday", nil), []string{"scheduled"},
				[]*service.Error{
					testing.ComposeError(service.ErrorValueNotExists(), "/deliveryType", nil),
				}),
			Entry("validates a suppressed with type scheduled; deliveryType not allowed",
				NewTestSuppressed("basal", "invalid", 1.0, "Weekday", nil), []string{"scheduled"},
				[]*service.Error{
					testing.ComposeError(service.ErrorValueStringNotOneOf("invalid", []string{"scheduled"}), "/deliveryType", nil),
				}),
			Entry("validates a suppressed with type scheduled; rate missing",
				NewTestSuppressed("basal", "scheduled", nil, "Weekday", nil), []string{"scheduled"},
				[]*service.Error{
					testing.ComposeError(service.ErrorValueNotExists(), "/rate", nil),
				}),
			Entry("validates a suppressed with type scheduled; rate out of range (lower)",
				NewTestSuppressed("basal", "scheduled", -0.1, "Weekday", nil), []string{"scheduled"},
				[]*service.Error{
					testing.ComposeError(service.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/rate", nil),
				}),
			Entry("validates a suppressed with type scheduled; rate at limit (lower)",
				NewTestSuppressed("basal", "scheduled", 0.0, "Weekday", nil), []string{"scheduled"},
				[]*service.Error{}),
			Entry("validates a suppressed with type scheduled; rate at limit (upper)",
				NewTestSuppressed("basal", "scheduled", 100.0, "Weekday", nil), []string{"scheduled"},
				[]*service.Error{}),
			Entry("validates a suppressed with type scheduled; rate out of range (upper)",
				NewTestSuppressed("basal", "scheduled", 100.1, "Weekday", nil), []string{"scheduled"},
				[]*service.Error{
					testing.ComposeError(service.ErrorValueNotInRange(100.1, 0.0, 100.0), "/rate", nil),
				}),
			Entry("validates a suppressed with type scheduled; schedule name empty",
				NewTestSuppressed("basal", "scheduled", 1.0, "", nil), []string{"scheduled"},
				[]*service.Error{
					testing.ComposeError(service.ErrorValueEmpty(), "/scheduleName", nil),
				}),
			Entry("validates a suppressed with type scheduled; suppressed exists",
				NewTestSuppressed("basal", "scheduled", 1.0, "Weekday", NewTestSuppressed(nil, nil, nil, nil, nil)), []string{"scheduled"},
				[]*service.Error{
					testing.ComposeError(service.ErrorValueExists(), "/suppressed", nil),
				}),
			Entry("validates a suppressed with type scheduled; multiple",
				NewTestSuppressed("invalid", "scheduled", 100.1, "", NewTestSuppressed(nil, nil, nil, nil, nil)), []string{"scheduled"},
				[]*service.Error{
					testing.ComposeError(service.ErrorValueNotEqualTo("invalid", "basal"), "/type", nil),
					testing.ComposeError(service.ErrorValueNotInRange(100.1, 0.0, 100.0), "/rate", nil),
					testing.ComposeError(service.ErrorValueEmpty(), "/scheduleName", nil),
					testing.ComposeError(service.ErrorValueExists(), "/suppressed", nil),
				}),
			Entry("validates a suppressed with type temp; all valid",
				NewTestSuppressed("basal", "temp", 2.0, nil, NewTestSuppressed("basal", "scheduled", 1.0, "Weekday", nil)), []string{"scheduled", "temp"},
				[]*service.Error{}),
			Entry("validates a suppressed with type temp; type missing",
				NewTestSuppressed(nil, "temp", 2.0, nil, NewTestSuppressed("basal", "scheduled", 1.0, "Weekday", nil)), []string{"scheduled", "temp"},
				[]*service.Error{
					testing.ComposeError(service.ErrorValueNotExists(), "/type", nil),
				}),
			Entry("validates a suppressed with type temp; type not basal",
				NewTestSuppressed("invalid", "temp", 2.0, nil, NewTestSuppressed("basal", "scheduled", 1.0, "Weekday", nil)), []string{"scheduled", "temp"},
				[]*service.Error{
					testing.ComposeError(service.ErrorValueNotEqualTo("invalid", "basal"), "/type", nil),
				}),
			Entry("validates a suppressed with type temp; deliveryType missing",
				NewTestSuppressed("basal", nil, 2.0, nil, NewTestSuppressed("basal", "scheduled", 1.0, "Weekday", nil)), []string{"scheduled", "temp"},
				[]*service.Error{
					testing.ComposeError(service.ErrorValueNotExists(), "/deliveryType", nil),
				}),
			Entry("validates a suppressed with type temp; deliveryType not allowed",
				NewTestSuppressed("basal", "invalid", 2.0, "Weekday", NewTestSuppressed("basal", "scheduled", 1.0, "Weekday", nil)), []string{"scheduled", "temp"},
				[]*service.Error{
					testing.ComposeError(service.ErrorValueStringNotOneOf("invalid", []string{"scheduled", "temp"}), "/deliveryType", nil),
				}),
			Entry("validates a suppressed with type temp; rate missing",
				NewTestSuppressed("basal", "temp", nil, nil, NewTestSuppressed("basal", "scheduled", 1.0, "Weekday", nil)), []string{"scheduled", "temp"},
				[]*service.Error{
					testing.ComposeError(service.ErrorValueNotExists(), "/rate", nil),
				}),
			Entry("validates a suppressed with type temp; rate out of range (lower)",
				NewTestSuppressed("basal", "temp", -0.1, nil, NewTestSuppressed("basal", "scheduled", 1.0, "Weekday", nil)), []string{"scheduled", "temp"},
				[]*service.Error{
					testing.ComposeError(service.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/rate", nil),
				}),
			Entry("validates a suppressed with type temp; rate at limit (lower)",
				NewTestSuppressed("basal", "temp", 0.0, nil, NewTestSuppressed("basal", "scheduled", 1.0, "Weekday", nil)), []string{"scheduled", "temp"},
				[]*service.Error{}),
			Entry("validates a suppressed with type temp; rate at limit (upper)",
				NewTestSuppressed("basal", "temp", 100.0, nil, NewTestSuppressed("basal", "scheduled", 1.0, "Weekday", nil)), []string{"scheduled", "temp"},
				[]*service.Error{}),
			Entry("validates a suppressed with type temp; rate out of range (upper)",
				NewTestSuppressed("basal", "temp", 100.1, nil, NewTestSuppressed("basal", "scheduled", 1.0, "Weekday", nil)), []string{"scheduled", "temp"},
				[]*service.Error{
					testing.ComposeError(service.ErrorValueNotInRange(100.1, 0.0, 100.0), "/rate", nil),
				}),
			Entry("validates a suppressed with type temp; schedule name exists",
				NewTestSuppressed("basal", "temp", 2.0, "Weekday", NewTestSuppressed("basal", "scheduled", 1.0, "Weekday", nil)), []string{"scheduled", "temp"},
				[]*service.Error{
					testing.ComposeError(service.ErrorValueExists(), "/scheduleName", nil),
				}),
			Entry("validates a suppressed with type temp; suppressed missing",
				NewTestSuppressed("basal", "temp", 2.0, nil, nil), []string{"scheduled", "temp"},
				[]*service.Error{
					testing.ComposeError(service.ErrorValueNotExists(), "/suppressed", nil),
				}),
			Entry("validates a suppressed with type temp; multiple",
				NewTestSuppressed("invalid", "temp", 100.1, "Weekday", nil), []string{"scheduled", "temp"},
				[]*service.Error{
					testing.ComposeError(service.ErrorValueNotEqualTo("invalid", "basal"), "/type", nil),
					testing.ComposeError(service.ErrorValueNotInRange(100.1, 0.0, 100.0), "/rate", nil),
					testing.ComposeError(service.ErrorValueExists(), "/scheduleName", nil),
					testing.ComposeError(service.ErrorValueNotExists(), "/suppressed", nil),
				}),
			Entry("validates a suppressed with type temp; suppressed type missing",
				NewTestSuppressed("basal", "temp", 2.0, nil, NewTestSuppressed(nil, "scheduled", 1.0, "Weekday", nil)), []string{"scheduled", "temp"},
				[]*service.Error{
					testing.ComposeError(service.ErrorValueNotExists(), "/suppressed/type", nil),
				}),
			Entry("validates a suppressed with type temp; suppressed type not basal",
				NewTestSuppressed("basal", "temp", 2.0, nil, NewTestSuppressed("invalid", "scheduled", 1.0, "Weekday", nil)), []string{"scheduled", "temp"},
				[]*service.Error{
					testing.ComposeError(service.ErrorValueNotEqualTo("invalid", "basal"), "/suppressed/type", nil),
				}),
			Entry("validates a suppressed with type temp; suppressed deliveryType missing",
				NewTestSuppressed("basal", "temp", 2.0, nil, NewTestSuppressed("basal", nil, 1.0, "Weekday", nil)), []string{"scheduled", "temp"},
				[]*service.Error{
					testing.ComposeError(service.ErrorValueNotExists(), "/suppressed/deliveryType", nil),
				}),
			Entry("validates a suppressed with type temp; suppressed deliveryType not allowed",
				NewTestSuppressed("basal", "temp", 2.0, nil, NewTestSuppressed("basal", "temp", 1.0, "Weekday", nil)), []string{"scheduled", "temp"},
				[]*service.Error{
					testing.ComposeError(service.ErrorValueStringNotOneOf("temp", []string{"scheduled"}), "/suppressed/deliveryType", nil),
				}),
			Entry("validates a suppressed with type temp; suppressed rate missing",
				NewTestSuppressed("basal", "temp", 2.0, nil, NewTestSuppressed("basal", "scheduled", nil, "Weekday", nil)), []string{"scheduled", "temp"},
				[]*service.Error{
					testing.ComposeError(service.ErrorValueNotExists(), "/suppressed/rate", nil),
				}),
			Entry("validates a suppressed with type temp; suppressed rate out of range (lower)",
				NewTestSuppressed("basal", "temp", 2.0, nil, NewTestSuppressed("basal", "scheduled", -0.1, "Weekday", nil)), []string{"scheduled", "temp"},
				[]*service.Error{
					testing.ComposeError(service.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/suppressed/rate", nil),
				}),
			Entry("validates a suppressed with type temp; suppressed rate at limit (lower)",
				NewTestSuppressed("basal", "temp", 2.0, nil, NewTestSuppressed("basal", "scheduled", 0.0, "Weekday", nil)), []string{"scheduled", "temp"},
				[]*service.Error{}),
			Entry("validates a suppressed with type temp; suppressed rate at limit (upper)",
				NewTestSuppressed("basal", "temp", 2.0, nil, NewTestSuppressed("basal", "scheduled", 100.0, "Weekday", nil)), []string{"scheduled", "temp"},
				[]*service.Error{}),
			Entry("validates a suppressed with type temp; suppressed rate out of range (upper)",
				NewTestSuppressed("basal", "temp", 2.0, nil, NewTestSuppressed("basal", "scheduled", 100.1, "Weekday", nil)), []string{"scheduled", "temp"},
				[]*service.Error{
					testing.ComposeError(service.ErrorValueNotInRange(100.1, 0.0, 100.0), "/suppressed/rate", nil),
				}),
			Entry("validates a suppressed with type temp; suppressed schedule name empty",
				NewTestSuppressed("basal", "temp", 2.0, nil, NewTestSuppressed("basal", "scheduled", 1.0, "", nil)), []string{"scheduled", "temp"},
				[]*service.Error{
					testing.ComposeError(service.ErrorValueEmpty(), "/suppressed/scheduleName", nil),
				}),
			Entry("validates a suppressed with type temp; suppressed suppressed exists",
				NewTestSuppressed("basal", "temp", 2.0, nil, NewTestSuppressed("basal", "scheduled", 1.0, "Weekday", NewTestSuppressed(nil, nil, nil, nil, nil))), []string{"scheduled", "temp"},
				[]*service.Error{
					testing.ComposeError(service.ErrorValueExists(), "/suppressed/suppressed", nil),
				}),
			Entry("validates a suppressed with type temp; suppressed multiple",
				NewTestSuppressed("basal", "temp", 2.0, nil, NewTestSuppressed("invalid", "scheduled", 100.1, "", NewTestSuppressed(nil, nil, nil, nil, nil))), []string{"scheduled", "temp"},
				[]*service.Error{
					testing.ComposeError(service.ErrorValueNotEqualTo("invalid", "basal"), "/suppressed/type", nil),
					testing.ComposeError(service.ErrorValueNotInRange(100.1, 0.0, 100.0), "/suppressed/rate", nil),
					testing.ComposeError(service.ErrorValueEmpty(), "/suppressed/scheduleName", nil),
					testing.ComposeError(service.ErrorValueExists(), "/suppressed/suppressed", nil),
				}),
		)
	})
})
