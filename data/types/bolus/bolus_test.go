package bolus_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/context"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/parser"
	testData "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/bolus"
	testDataTypesBolus "github.com/tidepool-org/platform/data/types/bolus/test"
	testDataTypesInsulin "github.com/tidepool-org/platform/data/types/insulin/test"
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
	datum := bolus.New("")
	datum.DeviceID = pointer.String(id.New())
	if val, ok := sourceTime.(string); ok {
		datum.Time = &val
	}
	if val, ok := sourceSubType.(string); ok {
		datum.SubType = val
	}
	return &datum
}

var _ = Describe("Bolus", func() {
	It("Type is expected", func() {
		Expect(bolus.Type).To(Equal("bolus"))
	})

	Context("New", func() {
		It("creates a new datum with all values initialized", func() {
			subType := testDataTypes.NewType()
			datum := bolus.New(subType)
			Expect(datum.Type).To(Equal("bolus"))
			Expect(datum.SubType).To(Equal(subType))
			Expect(datum.InsulinFormulation).To(BeNil())
		})
	})

	Context("with new datum", func() {
		var subType string
		var datum bolus.Bolus

		BeforeEach(func() {
			subType = testDataTypes.NewType()
			datum = bolus.New(subType)
		})

		Context("Meta", func() {
			It("returns the meta with delivery type", func() {
				Expect(datum.Meta()).To(Equal(&bolus.Meta{Type: "bolus", SubType: subType}))
			})
		})
	})

	Context("Bolus", func() {
		Context("Parse", func() {
			var datum *bolus.Bolus

			BeforeEach(func() {
				datum = NewTestBolus("bolus", nil)
			})

			DescribeTable("parses the datum",
				func(sourceObject *map[string]interface{}, expectedDatum *bolus.Bolus, expectedErrors []*service.Error) {
					testContext, err := context.NewStandard(null.NewLogger())
					Expect(err).ToNot(HaveOccurred())
					Expect(testContext).ToNot(BeNil())
					testParser, err := parser.NewStandardObject(testContext, sourceObject, parser.AppendErrorNotParsed)
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
				Entry("insulin formulation missing",
					func(datum *bolus.Bolus) { datum.InsulinFormulation = nil },
				),
				Entry("insulin formulation invalid",
					func(datum *bolus.Bolus) {
						datum.InsulinFormulation.Compounds = nil
						datum.InsulinFormulation.Name = nil
						datum.InsulinFormulation.Simple = nil
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/insulinFormulation/simple"),
				),
				Entry("insulin formulation valid",
					func(datum *bolus.Bolus) { datum.InsulinFormulation = testDataTypesInsulin.NewFormulation(3) },
				),
				Entry("multiple errors",
					func(datum *bolus.Bolus) {
						datum.Type = "invalid"
						datum.SubType = ""
						datum.InsulinFormulation.Compounds = nil
						datum.InsulinFormulation.Name = nil
						datum.InsulinFormulation.Simple = nil
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalid", "bolus"), "/type"),
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/subType"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/insulinFormulation/simple"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *bolus.Bolus)) {
					for _, origin := range structure.Origins() {
						datum := testDataTypesBolus.NewBolus()
						mutator(datum)
						expectedDatum := testDataTypesBolus.CloneBolus(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *bolus.Bolus) {},
				),
				Entry("does not modify the datum; type missing",
					func(datum *bolus.Bolus) { datum.Type = "" },
				),
				Entry("does not modify the datum; sub type missing",
					func(datum *bolus.Bolus) { datum.SubType = "" },
				),
				Entry("does not modify the datum; insulin formulation missing",
					func(datum *bolus.Bolus) { datum.InsulinFormulation = nil },
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
