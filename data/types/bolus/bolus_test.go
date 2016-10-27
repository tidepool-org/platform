package bolus_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/factory"
	"github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/parser"
	testData "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/bolus"
	"github.com/tidepool-org/platform/data/validator"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/service"
)

func NewMeta(subType string) interface{} {
	return &bolus.Meta{
		Type:    "bolus",
		SubType: subType,
	}
}

func NewTestBolus(sourceTime interface{}, sourceSubType interface{}) *bolus.Bolus {
	testBolus := &bolus.Bolus{}
	testBolus.Init()
	testBolus.DeviceID = app.StringAsPointer(app.NewID())
	if value, ok := sourceTime.(string); ok {
		testBolus.Time = app.StringAsPointer(value)
	}
	if value, ok := sourceSubType.(string); ok {
		testBolus.SubType = value
	}
	return testBolus
}

var _ = Describe("Bolus", func() {
	Context("Type", func() {
		It("returns the expected type", func() {
			Expect(bolus.Type()).To(Equal("bolus"))
		})
	})

	Context("with new bolus", func() {
		var testBolus *bolus.Bolus

		BeforeEach(func() {
			testBolus = &bolus.Bolus{}
		})

		Context("Init", func() {
			It("initializes the bolus", func() {
				testBolus.Init()
				Expect(testBolus.ID).ToNot(BeEmpty())
				Expect(testBolus.Type).To(Equal("bolus"))
				Expect(testBolus.SubType).To(BeEmpty())
			})
		})

		Context("with initialized", func() {
			BeforeEach(func() {
				testBolus.Init()
			})

			Context("Meta", func() {
				It("returns the meta with no sub type", func() {
					testBolus.Init()
					Expect(testBolus.Meta()).To(Equal(NewMeta("")))
				})

				It("returns the meta with sub type", func() {
					testBolus.Init()
					testBolus.SubType = "dual/square"
					Expect(testBolus.Meta()).To(Equal(NewMeta("dual/square")))
				})
			})

			DescribeTable("Parse",
				func(sourceObject *map[string]interface{}, expectedBolus *bolus.Bolus, expectedErrors []*service.Error) {
					testContext, err := context.NewStandard(log.NewNull())
					Expect(err).ToNot(HaveOccurred())
					Expect(testContext).ToNot(BeNil())
					testFactory, err := factory.NewStandard()
					Expect(err).ToNot(HaveOccurred())
					Expect(testFactory).ToNot(BeNil())
					testParser, err := parser.NewStandardObject(testContext, testFactory, sourceObject, parser.AppendErrorNotParsed)
					Expect(err).ToNot(HaveOccurred())
					Expect(testParser).ToNot(BeNil())
					Expect(testBolus.Parse(testParser)).To(Succeed())
					Expect(testBolus.Time).To(Equal(expectedBolus.Time))
					Expect(testBolus.SubType).To(Equal(expectedBolus.SubType))
					Expect(testContext.Errors()).To(ConsistOf(expectedErrors))
				},
				Entry("parses object that is nil",
					nil,
					NewTestBolus(nil, nil),
					[]*service.Error{}),
				Entry("parses object that is empty",
					&map[string]interface{}{},
					NewTestBolus(nil, nil),
					[]*service.Error{}),
				Entry("parses object that has valid time",
					&map[string]interface{}{"time": "2016-09-06T13:45:58-07:00"},
					NewTestBolus("2016-09-06T13:45:58-07:00", nil),
					[]*service.Error{}),
				Entry("parses object that has invalid time",
					&map[string]interface{}{"time": 0},
					NewTestBolus(nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotString(0), "/time", NewMeta("")),
					}),
				Entry("does not parse sub type",
					&map[string]interface{}{"subType": "dual/square"},
					NewTestBolus(nil, nil),
					[]*service.Error{}),
				Entry("parses object that has multiple valid fields",
					&map[string]interface{}{"time": "2016-09-06T13:45:58-07:00", "subType": "dual/square"},
					NewTestBolus("2016-09-06T13:45:58-07:00", nil),
					[]*service.Error{}),
				Entry("parses object that has multiple invalid fields",
					&map[string]interface{}{"time": 0, "subType": 0},
					NewTestBolus(nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotString(0), "/time", NewMeta("")),
					}),
			)

			DescribeTable("Validate",
				func(sourceBolus *bolus.Bolus, expectedErrors []*service.Error) {
					testContext, err := context.NewStandard(log.NewNull())
					Expect(err).ToNot(HaveOccurred())
					Expect(testContext).ToNot(BeNil())
					testValidator, err := validator.NewStandard(testContext)
					Expect(err).ToNot(HaveOccurred())
					Expect(testValidator).ToNot(BeNil())
					Expect(sourceBolus.Validate(testValidator)).To(Succeed())
					Expect(testContext.Errors()).To(ConsistOf(expectedErrors))
				},
				Entry("all valid",
					NewTestBolus("2016-09-06T13:45:58-07:00", "dual/square"),
					[]*service.Error{}),
				Entry("missing time",
					NewTestBolus(nil, "dual/square"),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotExists(), "/time", NewMeta("dual/square")),
					}),
				Entry("missing sub type",
					NewTestBolus("2016-09-06T13:45:58-07:00", nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueEmpty(), "/subType", NewMeta("")),
					}),
				Entry("specified sub type",
					NewTestBolus("2016-09-06T13:45:58-07:00", "specified"),
					[]*service.Error{}),
				Entry("multiple",
					NewTestBolus(nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorValueNotExists(), "/time", NewMeta("")),
						testData.ComposeError(service.ErrorValueEmpty(), "/subType", NewMeta("")),
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
					Expect(testBolus.Normalize(testNormalizer)).To(Succeed())
				})
			})

			Context("IdentityFields", func() {
				var userID string
				var deviceID string

				BeforeEach(func() {
					userID = app.NewID()
					deviceID = app.NewID()
					testBolus.UserID = userID
					testBolus.DeviceID = &deviceID
					testBolus.Time = app.StringAsPointer("2016-09-06T13:45:58-07:00")
					testBolus.SubType = "dual/square"
				})

				It("returns error if user id is empty", func() {
					testBolus.UserID = ""
					identityFields, err := testBolus.IdentityFields()
					Expect(err).To(MatchError("base: user id is empty"))
					Expect(identityFields).To(BeEmpty())
				})

				It("returns error if sub type is empty", func() {
					testBolus.SubType = ""
					identityFields, err := testBolus.IdentityFields()
					Expect(err).To(MatchError("bolus: sub type is empty"))
					Expect(identityFields).To(BeEmpty())
				})

				It("returns the expected identity fields", func() {
					identityFields, err := testBolus.IdentityFields()
					Expect(err).ToNot(HaveOccurred())
					Expect(identityFields).To(Equal([]string{userID, deviceID, "2016-09-06T13:45:58-07:00", "bolus", "dual/square"}))
				})
			})
		})
	})
})
