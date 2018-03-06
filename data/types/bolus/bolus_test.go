package bolus_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/factory"
	"github.com/tidepool-org/platform/data/parser"
	testData "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/bolus"
	testDataTypesBolus "github.com/tidepool-org/platform/data/types/bolus/test"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func NewTestBolus(sourceTime interface{}, sourceSubType interface{}) *bolus.Bolus {
	datum := &bolus.Bolus{}
	datum.Init()
	datum.DeviceID = pointer.String(id.New())
	if val, ok := sourceTime.(string); ok {
		datum.Time = &val
	}
	if val, ok := sourceSubType.(string); ok {
		datum.SubType = val
	}
	return datum
}

var _ = Describe("Bolus", func() {
	It("Type is expected", func() {
		Expect(bolus.Type).To(Equal("bolus"))
	})

	Context("with new datum", func() {
		var datum *bolus.Bolus

		BeforeEach(func() {
			datum = testDataTypesBolus.NewBolus()
		})

		Context("Init", func() {
			It("initializes the datum", func() {
				datum.Init()
				Expect(datum.Type).To(Equal("bolus"))
				Expect(datum.SubType).To(BeEmpty())
			})
		})

		Context("with initialized", func() {
			BeforeEach(func() {
				datum.Init()
			})

			Context("Meta", func() {
				It("returns the meta with no sub type", func() {
					Expect(datum.Meta()).To(Equal(&bolus.Meta{Type: "bolus"}))
				})

				It("returns the meta with sub type", func() {
					datum.SubType = testDataTypes.NewType()
					Expect(datum.Meta()).To(Equal(&bolus.Meta{Type: "bolus", SubType: datum.SubType}))
				})
			})
		})
	})

	Context("Bolus", func() {
		Context("Parse", func() {
			var datum *bolus.Bolus

			BeforeEach(func() {
				datum = &bolus.Bolus{}
				datum.Init()
			})

			DescribeTable("parses the datum",
				func(sourceObject *map[string]interface{}, expectedDatum *bolus.Bolus, expectedErrors []*service.Error) {
					testContext, err := context.NewStandard(null.NewLogger())
					Expect(err).ToNot(HaveOccurred())
					Expect(testContext).ToNot(BeNil())
					testFactory, err := factory.NewStandard()
					Expect(err).ToNot(HaveOccurred())
					Expect(testFactory).ToNot(BeNil())
					testParser, err := parser.NewStandardObject(testContext, testFactory, sourceObject, parser.AppendErrorNotParsed)
					Expect(err).ToNot(HaveOccurred())
					Expect(testParser).ToNot(BeNil())
					Expect(datum.Parse(testParser)).To(Succeed())
					Expect(datum.Time).To(Equal(expectedDatum.Time))
					Expect(datum.SubType).To(Equal(expectedDatum.SubType))
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
						testData.ComposeError(service.ErrorTypeNotString(0), "/time", &bolus.Meta{Type: "bolus"}),
					}),
				Entry("does not parse sub type",
					&map[string]interface{}{"subType": "normal"},
					NewTestBolus(nil, nil),
					[]*service.Error{}),
				Entry("parses object that has multiple valid fields",
					&map[string]interface{}{"time": "2016-09-06T13:45:58-07:00", "subType": "normal"},
					NewTestBolus("2016-09-06T13:45:58-07:00", nil),
					[]*service.Error{}),
				Entry("parses object that has multiple invalid fields",
					&map[string]interface{}{"time": 0, "subType": 0},
					NewTestBolus(nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotString(0), "/time", &bolus.Meta{Type: "bolus"}),
					}),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *bolus.Bolus), expectedErrors ...error) {
					datum := testDataTypesBolus.NewBolus()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *bolus.Bolus) {},
				),
				Entry("type missing",
					func(datum *bolus.Bolus) { datum.Type = "" },
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/type"),
				),
				Entry("type invalid",
					func(datum *bolus.Bolus) { datum.Type = "invalid" },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalid", "bolus"), "/type"),
				),
				Entry("type bolus",
					func(datum *bolus.Bolus) { datum.Type = "bolus" },
				),
				Entry("sub type missing",
					func(datum *bolus.Bolus) { datum.SubType = "" },
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/subType"),
				),
				Entry("sub type valid",
					func(datum *bolus.Bolus) { datum.SubType = testDataTypes.NewType() },
				),
				Entry("multiple errors",
					func(datum *bolus.Bolus) {
						datum.Type = "invalid"
						datum.SubType = ""
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalid", "bolus"), "/type"),
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/subType"),
				),
			)
		})

		Context("IdentityFields", func() {
			var datum *bolus.Bolus

			BeforeEach(func() {
				datum = testDataTypesBolus.NewBolus()
			})

			It("returns error if user id is missing", func() {
				datum.UserID = nil
				identityFields, err := datum.IdentityFields()
				Expect(err).To(MatchError("user id is missing"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns error if user id is empty", func() {
				datum.UserID = pointer.String("")
				identityFields, err := datum.IdentityFields()
				Expect(err).To(MatchError("user id is empty"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns error if sub type is empty", func() {
				datum.SubType = ""
				identityFields, err := datum.IdentityFields()
				Expect(err).To(MatchError("sub type is empty"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns the expected identity fields", func() {
				identityFields, err := datum.IdentityFields()
				Expect(err).ToNot(HaveOccurred())
				Expect(identityFields).To(Equal([]string{*datum.UserID, *datum.DeviceID, *datum.Time, datum.Type, datum.SubType}))
			})
		})
	})
})
