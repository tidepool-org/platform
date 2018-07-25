package blood_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"math"
	"strconv"

	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/parser"
	testData "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/blood"
	testDataTypesBlood "github.com/tidepool-org/platform/data/types/blood/test"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewTestBlood(sourceTime interface{}, sourceUnits interface{}, sourceValue interface{}) *blood.Blood {
	datum := blood.New("blood")
	datum.DeviceID = pointer.String(id.New())
	if val, ok := sourceTime.(string); ok {
		datum.Time = &val
	}
	if val, ok := sourceUnits.(string); ok {
		datum.Units = &val
	}
	if val, ok := sourceValue.(float64); ok {
		datum.Value = &val
	}
	return &datum
}

var _ = Describe("Blood", func() {
	Context("New", func() {
		It("creates a new datum with all values initialized", func() {
			typ := testDataTypes.NewType()
			datum := blood.New(typ)
			Expect(datum.Type).To(Equal(typ))
			Expect(datum.Units).To(BeNil())
			Expect(datum.Value).To(BeNil())
		})
	})

	Context("with new datum", func() {
		var typ string
		var datum blood.Blood

		BeforeEach(func() {
			typ = testDataTypes.NewType()
			datum = blood.New(typ)
		})

		Context("Meta", func() {
			It("returns the meta with delivery type", func() {
				Expect(datum.Meta()).To(Equal(&types.Meta{Type: typ}))
			})
		})
	})

	Context("Blood", func() {
		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(sourceObject *map[string]interface{}, expectedDatum *blood.Blood, expectedErrors []*service.Error) {
					testContext, err := context.NewStandard(null.NewLogger())
					Expect(err).ToNot(HaveOccurred())
					Expect(testContext).ToNot(BeNil())
					testParser, err := parser.NewStandardObject(testContext, sourceObject, parser.AppendErrorNotParsed)
					Expect(err).ToNot(HaveOccurred())
					Expect(testParser).ToNot(BeNil())
					datum := &blood.Blood{}
					datum.Type = "blood"
					Expect(datum.Parse(testParser)).To(Succeed())
					Expect(datum.Time).To(Equal(expectedDatum.Time))
					Expect(datum.Units).To(Equal(expectedDatum.Units))
					Expect(datum.Value).To(Equal(expectedDatum.Value))
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
						testData.ComposeError(service.ErrorTypeNotString(0), "/time", &types.Meta{Type: "blood"}),
					}),
				Entry("parses object that has valid units",
					&map[string]interface{}{"units": "mmol/L"},
					NewTestBlood(nil, "mmol/L", nil),
					[]*service.Error{}),
				Entry("parses object that has invalid units",
					&map[string]interface{}{"units": 0},
					NewTestBlood(nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotString(0), "/units", &types.Meta{Type: "blood"}),
					}),
				Entry("parses object that has valid value",
					&map[string]interface{}{"value": 1.0},
					NewTestBlood(nil, nil, 1.0),
					[]*service.Error{}),
				Entry("parses object that has invalid value",
					&map[string]interface{}{"value": "invalid"},
					NewTestBlood(nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/value", &types.Meta{Type: "blood"}),
					}),
				Entry("parses object that has multiple valid fields",
					&map[string]interface{}{"time": "2016-09-06T13:45:58-07:00", "units": "mmol/L", "value": 1.0},
					NewTestBlood("2016-09-06T13:45:58-07:00", "mmol/L", 1.0),
					[]*service.Error{}),
				Entry("parses object that has multiple invalid fields",
					&map[string]interface{}{"time": 0, "units": 0, "value": "invalid"},
					NewTestBlood(nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotString(0), "/time", &types.Meta{Type: "blood"}),
						testData.ComposeError(service.ErrorTypeNotString(0), "/units", &types.Meta{Type: "blood"}),
						testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/value", &types.Meta{Type: "blood"}),
					}),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *blood.Blood), expectedErrors ...error) {
					datum := testDataTypesBlood.NewBlood()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *blood.Blood) {},
				),
				Entry("type missing",
					func(datum *blood.Blood) { datum.Type = "" },
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/type"),
				),
				Entry("type exists",
					func(datum *blood.Blood) { datum.Type = testDataTypes.NewType() },
				),
				Entry("units missing",
					func(datum *blood.Blood) { datum.Units = nil },
				),
				Entry("units exists",
					func(datum *blood.Blood) { datum.Units = pointer.String(testDataTypes.NewType()) },
				),
				Entry("value missing",
					func(datum *blood.Blood) { datum.Value = nil },
				),
				Entry("value exists",
					func(datum *blood.Blood) {
						datum.Value = pointer.Float64(test.RandomFloat64FromRange(-math.MaxFloat64, math.MaxFloat64))
					},
				),
			)
		})

		Context("IdentityFields", func() {
			var datum *blood.Blood

			BeforeEach(func() {
				datum = testDataTypesBlood.NewBlood()
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

			It("returns error if units is missing", func() {
				datum.Units = nil
				identityFields, err := datum.IdentityFields()
				Expect(err).To(MatchError("units is missing"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns error if value is missing", func() {
				datum.Value = nil
				identityFields, err := datum.IdentityFields()
				Expect(err).To(MatchError("value is missing"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns the expected identity fields", func() {
				identityFields, err := datum.IdentityFields()
				Expect(err).ToNot(HaveOccurred())
				Expect(identityFields).To(Equal([]string{*datum.UserID, *datum.DeviceID, *datum.Time, datum.Type, *datum.Units, strconv.FormatFloat(*datum.Value, 'f', -1, 64)}))
			})
		})
	})
})
