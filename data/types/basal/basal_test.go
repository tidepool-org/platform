package basal_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/parser"
	testData "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/basal"
	testDataTypesBasal "github.com/tidepool-org/platform/data/types/basal/test"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func NewTestBasal(sourceTime interface{}, sourceDeliveryType interface{}) *basal.Basal {
	datum := basal.New("")
	datum.DeviceID = pointer.FromString(id.New())
	if val, ok := sourceTime.(string); ok {
		datum.Time = &val
	}
	if val, ok := sourceDeliveryType.(string); ok {
		datum.DeliveryType = val
	}
	return &datum
}

var _ = Describe("Basal", func() {
	It("Type is expected", func() {
		Expect(basal.Type).To(Equal("basal"))
	})

	Context("New", func() {
		It("creates a new datum with all values initialized", func() {
			deliveryType := testDataTypes.NewType()
			datum := basal.New(deliveryType)
			Expect(datum.Type).To(Equal("basal"))
			Expect(datum.DeliveryType).To(Equal(deliveryType))
		})
	})

	Context("with new datum", func() {
		var deliveryType string
		var datum basal.Basal

		BeforeEach(func() {
			deliveryType = testDataTypes.NewType()
			datum = basal.New(deliveryType)
		})

		Context("Meta", func() {
			It("returns the meta with delivery type", func() {
				Expect(datum.Meta()).To(Equal(&basal.Meta{Type: "basal", DeliveryType: deliveryType}))
			})
		})
	})

	Context("Basal", func() {
		Context("Parse", func() {
			var datum *basal.Basal

			BeforeEach(func() {
				datum = NewTestBasal("basal", nil)
			})

			DescribeTable("parses the datum",
				func(sourceObject *map[string]interface{}, expectedDatum *basal.Basal, expectedErrors []*service.Error) {
					testContext, err := context.NewStandard(null.NewLogger())
					Expect(err).ToNot(HaveOccurred())
					Expect(testContext).ToNot(BeNil())
					testParser, err := parser.NewStandardObject(testContext, sourceObject, parser.AppendErrorNotParsed)
					Expect(err).ToNot(HaveOccurred())
					Expect(testParser).ToNot(BeNil())
					Expect(datum.Parse(testParser)).To(Succeed())
					Expect(datum.Time).To(Equal(expectedDatum.Time))
					Expect(datum.DeliveryType).To(Equal(expectedDatum.DeliveryType))
					Expect(testContext.Errors()).To(ConsistOf(expectedErrors))
				},
				Entry("parses object that is nil",
					nil,
					NewTestBasal(nil, nil),
					[]*service.Error{}),
				Entry("parses object that is empty",
					&map[string]interface{}{},
					NewTestBasal(nil, nil),
					[]*service.Error{}),
				Entry("parses object that has valid time",
					&map[string]interface{}{"time": "2016-09-06T13:45:58-07:00"},
					NewTestBasal("2016-09-06T13:45:58-07:00", nil),
					[]*service.Error{}),
				Entry("parses object that has invalid time",
					&map[string]interface{}{"time": 0},
					NewTestBasal(nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotString(0), "/time", &basal.Meta{Type: "basal"}),
					}),
				Entry("does not parse delivery type",
					&map[string]interface{}{"deliveryType": "scheduled"},
					NewTestBasal(nil, nil),
					[]*service.Error{}),
				Entry("parses object that has multiple valid fields",
					&map[string]interface{}{"time": "2016-09-06T13:45:58-07:00", "deliveryType": "scheduled"},
					NewTestBasal("2016-09-06T13:45:58-07:00", nil),
					[]*service.Error{}),
				Entry("parses object that has multiple invalid fields",
					&map[string]interface{}{"time": 0, "deliveryType": 0},
					NewTestBasal(nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotString(0), "/time", &basal.Meta{Type: "basal"}),
					}),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *basal.Basal), expectedErrors ...error) {
					datum := testDataTypesBasal.NewBasal()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *basal.Basal) {},
				),
				Entry("type missing",
					func(datum *basal.Basal) { datum.Type = "" },
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/type"),
				),
				Entry("type invalid",
					func(datum *basal.Basal) { datum.Type = "invalid" },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalid", "basal"), "/type"),
				),
				Entry("type basal",
					func(datum *basal.Basal) { datum.Type = "basal" },
				),
				Entry("delivery type missing",
					func(datum *basal.Basal) { datum.DeliveryType = "" },
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/deliveryType"),
				),
				Entry("delivery type valid",
					func(datum *basal.Basal) { datum.DeliveryType = testDataTypes.NewType() },
				),
				Entry("multiple errors",
					func(datum *basal.Basal) {
						datum.Type = "invalid"
						datum.DeliveryType = ""
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalid", "basal"), "/type"),
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/deliveryType"),
				),
			)
		})

		Context("IdentityFields", func() {
			var datum *basal.Basal

			BeforeEach(func() {
				datum = testDataTypesBasal.NewBasal()
			})

			It("returns error if user id is missing", func() {
				datum.UserID = nil
				identityFields, err := datum.IdentityFields()
				Expect(err).To(MatchError("user id is missing"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns error if user id is empty", func() {
				datum.UserID = pointer.FromString("")
				identityFields, err := datum.IdentityFields()
				Expect(err).To(MatchError("user id is empty"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns error if delivery type is empty", func() {
				datum.DeliveryType = ""
				identityFields, err := datum.IdentityFields()
				Expect(err).To(MatchError("delivery type is empty"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns the expected identity fields", func() {
				identityFields, err := datum.IdentityFields()
				Expect(err).ToNot(HaveOccurred())
				Expect(identityFields).To(Equal([]string{*datum.UserID, *datum.DeviceID, *datum.Time, datum.Type, datum.DeliveryType}))
			})
		})
	})

	Context("ParseDeliveryType", func() {
		// TODO
	})
})
