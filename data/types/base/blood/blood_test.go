package blood_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/factory"
	"github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/parser"
	"github.com/tidepool-org/platform/data/types/base"
	"github.com/tidepool-org/platform/data/types/base/blood"
	"github.com/tidepool-org/platform/data/types/base/testing"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/service"
)

func NewMeta() interface{} {
	return &base.Meta{
		Type: "testBlood",
	}
}

func NewTestBlood(sourceTime interface{}, sourceUnits interface{}, sourceValue interface{}) *blood.Blood {
	testBlood := &blood.Blood{}
	testBlood.Init()
	testBlood.Type = "testBlood"
	testBlood.DeviceID = app.StringAsPointer(app.NewID())
	if value, ok := sourceTime.(string); ok {
		testBlood.Time = app.StringAsPointer(value)
	}
	if value, ok := sourceUnits.(string); ok {
		testBlood.Units = app.StringAsPointer(value)
	}
	if value, ok := sourceValue.(float64); ok {
		testBlood.Value = app.FloatAsPointer(value)
	}
	return testBlood
}

var _ = Describe("Blood", func() {
	Context("with new blood", func() {
		var testBlood *blood.Blood

		BeforeEach(func() {
			testBlood = &blood.Blood{}
		})

		Context("Init", func() {
			It("initializes the blood", func() {
				testBlood.Init()
				Expect(testBlood.Units).To(BeNil())
				Expect(testBlood.Value).To(BeNil())
			})
		})

		Context("with initialized", func() {
			BeforeEach(func() {
				testBlood.Init()
				testBlood.Type = "testBlood"
			})

			DescribeTable("Parse",
				func(sourceObject *map[string]interface{}, expectedBlood *blood.Blood, expectedErrors []*service.Error) {
					testContext, err := context.NewStandard(log.NewNull())
					Expect(err).ToNot(HaveOccurred())
					Expect(testContext).ToNot(BeNil())
					testFactory, err := factory.NewStandard()
					Expect(err).ToNot(HaveOccurred())
					Expect(testFactory).ToNot(BeNil())
					testParser, err := parser.NewStandardObject(testContext, testFactory, sourceObject, parser.AppendErrorNotParsed)
					Expect(err).ToNot(HaveOccurred())
					Expect(testParser).ToNot(BeNil())
					Expect(testBlood.Parse(testParser)).To(Succeed())
					Expect(testBlood.Time).To(Equal(expectedBlood.Time))
					Expect(testBlood.Units).To(Equal(expectedBlood.Units))
					Expect(testBlood.Value).To(Equal(expectedBlood.Value))
					Expect(testContext.Errors()).To(ConsistOf(expectedErrors))
				},
				Entry("parses object that is nil",
					nil,
					NewTestBlood(nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that is empty",
					&map[string]interface{}{},
					NewTestBlood(nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that has valid time",
					&map[string]interface{}{"time": "2016-09-06T13:45:58-07:00"},
					NewTestBlood("2016-09-06T13:45:58-07:00", nil, nil),
					[]*service.Error{}),
				Entry("parses object that has invalid time",
					&map[string]interface{}{"time": 0},
					NewTestBlood(nil, nil, nil),
					[]*service.Error{
						testing.ComposeError(service.ErrorTypeNotString(0), "/time", NewMeta()),
					}),
				Entry("parses object that has valid units",
					&map[string]interface{}{"units": "mmol/L"},
					NewTestBlood(nil, "mmol/L", nil),
					[]*service.Error{}),
				Entry("parses object that has invalid units",
					&map[string]interface{}{"units": 0},
					NewTestBlood(nil, nil, nil),
					[]*service.Error{
						testing.ComposeError(service.ErrorTypeNotString(0), "/units", NewMeta()),
					}),
				Entry("parses object that has valid value",
					&map[string]interface{}{"value": 1.0},
					NewTestBlood(nil, nil, 1.0),
					[]*service.Error{}),
				Entry("parses object that has invalid value",
					&map[string]interface{}{"value": "invalid"},
					NewTestBlood(nil, nil, nil),
					[]*service.Error{
						testing.ComposeError(service.ErrorTypeNotFloat("invalid"), "/value", NewMeta()),
					}),
				Entry("parses object that has multiple valid fields",
					&map[string]interface{}{"time": "2016-09-06T13:45:58-07:00", "units": "mmol/L", "value": 1.0},
					NewTestBlood("2016-09-06T13:45:58-07:00", "mmol/L", 1.0),
					[]*service.Error{}),
				Entry("parses object that has multiple invalid fields",
					&map[string]interface{}{"time": 0, "units": 0, "value": "invalid"},
					NewTestBlood(nil, nil, nil),
					[]*service.Error{
						testing.ComposeError(service.ErrorTypeNotString(0), "/time", NewMeta()),
						testing.ComposeError(service.ErrorTypeNotString(0), "/units", NewMeta()),
						testing.ComposeError(service.ErrorTypeNotFloat("invalid"), "/value", NewMeta()),
					}),
			)

			DescribeTable("Validate",
				func(sourceBlood *blood.Blood, expectedErrors []*service.Error) {
					testContext, err := context.NewStandard(log.NewNull())
					Expect(err).ToNot(HaveOccurred())
					Expect(testContext).ToNot(BeNil())
					testValidator, err := validator.NewStandard(testContext)
					Expect(err).ToNot(HaveOccurred())
					Expect(testValidator).ToNot(BeNil())
					Expect(sourceBlood.Validate(testValidator)).To(Succeed())
					Expect(testContext.Errors()).To(ConsistOf(expectedErrors))
				},
				Entry("all valid",
					NewTestBlood("2016-09-06T13:45:58-07:00", "mmol/L", 1.0),
					[]*service.Error{}),
				Entry("missing time",
					NewTestBlood(nil, "mmol/L", 1.0),
					[]*service.Error{
						testing.ComposeError(service.ErrorValueNotExists(), "/time", NewMeta()),
					}),
				Entry("missing units",
					NewTestBlood("2016-09-06T13:45:58-07:00", nil, 1.0),
					[]*service.Error{
						testing.ComposeError(service.ErrorValueNotExists(), "/units", NewMeta()),
					}),
				Entry("missing value",
					NewTestBlood("2016-09-06T13:45:58-07:00", "mmol/L", nil),
					[]*service.Error{
						testing.ComposeError(service.ErrorValueNotExists(), "/value", NewMeta()),
					}),
				Entry("multiple",
					NewTestBlood(nil, nil, nil),
					[]*service.Error{
						testing.ComposeError(service.ErrorValueNotExists(), "/time", NewMeta()),
						testing.ComposeError(service.ErrorValueNotExists(), "/units", NewMeta()),
						testing.ComposeError(service.ErrorValueNotExists(), "/value", NewMeta()),
					}),
			)

			Context("Normalize", func() {
				It("succeeds", func() {
					testContext, err := context.NewStandard(log.NewNull())
					Expect(err).ToNot(HaveOccurred())
					Expect(testContext).ToNot(BeNil())
					testNormalizer, err := normalizer.NewStandard(testContext)
					Expect(err).ToNot(HaveOccurred())
					Expect(testNormalizer).ToNot(BeNil())
					Expect(testBlood.Normalize(testNormalizer)).To(Succeed())
				})
			})

			Context("IdentityFields", func() {
				var userID string
				var deviceID string

				BeforeEach(func() {
					userID = app.NewID()
					deviceID = app.NewID()
					testBlood.UserID = userID
					testBlood.DeviceID = &deviceID
					testBlood.Time = app.StringAsPointer("2016-09-06T13:45:58-07:00")
					testBlood.Units = app.StringAsPointer("mmol/L")
					testBlood.Value = app.FloatAsPointer(1)
				})

				It("returns error if user id is empty", func() {
					testBlood.UserID = ""
					identityFields, err := testBlood.IdentityFields()
					Expect(err).To(MatchError("base: user id is empty"))
					Expect(identityFields).To(BeEmpty())
				})

				It("returns error if units is missing", func() {
					testBlood.Units = nil
					identityFields, err := testBlood.IdentityFields()
					Expect(err).To(MatchError("blood: units is missing"))
					Expect(identityFields).To(BeEmpty())
				})

				It("returns error if value is missing", func() {
					testBlood.Value = nil
					identityFields, err := testBlood.IdentityFields()
					Expect(err).To(MatchError("blood: value is missing"))
					Expect(identityFields).To(BeEmpty())
				})

				It("returns the expected identity fields", func() {
					identityFields, err := testBlood.IdentityFields()
					Expect(err).ToNot(HaveOccurred())
					Expect(identityFields).To(Equal([]string{userID, deviceID, "2016-09-06T13:45:58-07:00", "testBlood", "mmol/L", "1"}))
				})
			})
		})
	})
})
