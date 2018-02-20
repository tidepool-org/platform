package basal_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/context"
	"github.com/tidepool-org/platform/data/factory"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/parser"
	testData "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/basal"
	testDataTypesBasal "github.com/tidepool-org/platform/data/types/basal/test"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("Suppressed", func() {
	It("RateMaximum is expected", func() {
		Expect(basal.RateMaximum).To(Equal(100.0))
	})

	It("RateMinimum is expected", func() {
		Expect(basal.RateMinimum).To(Equal(0.0))
	})

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
			testDataTypesBasal.NewTestSuppressed(nil, nil, nil, nil, nil, nil),
			[]*service.Error{}),
		Entry("parses object that has multiple valid fields",
			&map[string]interface{}{"type": "basal", "deliveryType": "temp", "rate": 2.0, "suppressed": map[string]interface{}{"type": "basal", "deliveryType": "scheduled", "rate": 1.0, "scheduleName": "Weekday", "annotations": []interface{}{map[string]interface{}{"one": "two"}}}},
			testDataTypesBasal.NewTestSuppressed("basal", "temp", nil, 2.0, nil, testDataTypesBasal.NewTestSuppressed("basal", "scheduled", &data.BlobArray{{"one": "two"}}, 1.0, "Weekday", nil)),
			[]*service.Error{}),
		Entry("parses object that has multiple invalid fields",
			&map[string]interface{}{"type": 0, "deliveryType": 0, "rate": "invalid", "scheduleName": 0, "suppressed": 0},
			testDataTypesBasal.NewTestSuppressed(nil, nil, nil, nil, nil, nil),
			[]*service.Error{
				testData.ComposeError(service.ErrorTypeNotString(0), "/type", nil),
				testData.ComposeError(service.ErrorTypeNotString(0), "/deliveryType", nil),
				testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/rate", nil),
				testData.ComposeError(service.ErrorTypeNotString(0), "/scheduleName", nil),
				testData.ComposeError(service.ErrorTypeNotObject(0), "/suppressed", nil),
			}),
		Entry("parses object that has additional fields",
			&map[string]interface{}{"type": "basal", "deliveryType": "temp", "rate": 2.0, "suppressed": map[string]interface{}{"type": "basal", "deliveryType": "scheduled", "rate": 1.0, "scheduleName": "Weekday", "annotations": []interface{}{map[string]interface{}{"one": "two"}}}, "additional": 0.0},
			testDataTypesBasal.NewTestSuppressed("basal", "temp", nil, 2.0, nil, testDataTypesBasal.NewTestSuppressed("basal", "scheduled", &data.BlobArray{{"one": "two"}}, 1.0, "Weekday", nil)),
			[]*service.Error{
				testData.ComposeError(parser.ErrorNotParsed(), "/additional", nil),
			}),
	)

	Context("NewSuppressed", func() {
		It("is successful", func() {
			Expect(basal.NewSuppressed()).To(Equal(&basal.Suppressed{}))
		})
	})

	Context("Suppressed", func() {
		Context("Parse", func() {
			DescribeTable("parses the datum",
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
					datum := &basal.Suppressed{}
					datum.Parse(testParser)
					Expect(datum).To(Equal(expectedSuppressed))
					Expect(testContext.Errors()).To(ConsistOf(expectedErrors))
				},
				Entry("parses object that is nil",
					nil,
					testDataTypesBasal.NewTestSuppressed(nil, nil, nil, nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that is empty",
					&map[string]interface{}{},
					testDataTypesBasal.NewTestSuppressed(nil, nil, nil, nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that has valid type",
					&map[string]interface{}{"type": "basal"},
					testDataTypesBasal.NewTestSuppressed("basal", nil, nil, nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that has invalid type",
					&map[string]interface{}{"type": 0},
					testDataTypesBasal.NewTestSuppressed(nil, nil, nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotString(0), "/type", nil),
					}),
				Entry("parses object that has valid delivery type",
					&map[string]interface{}{"deliveryType": "temp"},
					testDataTypesBasal.NewTestSuppressed(nil, "temp", nil, nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that has invalid delivery type",
					&map[string]interface{}{"deliveryType": 0},
					testDataTypesBasal.NewTestSuppressed(nil, nil, nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotString(0), "/deliveryType", nil),
					}),
				Entry("parses object that has valid rate",
					&map[string]interface{}{"rate": 2.0},
					testDataTypesBasal.NewTestSuppressed(nil, nil, nil, 2.0, nil, nil),
					[]*service.Error{}),
				Entry("parses object that has invalid rate",
					&map[string]interface{}{"rate": "invalid"},
					testDataTypesBasal.NewTestSuppressed(nil, nil, nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/rate", nil),
					}),
				Entry("parses object that has valid schedule name",
					&map[string]interface{}{"scheduleName": "Weekday"},
					testDataTypesBasal.NewTestSuppressed(nil, nil, nil, nil, "Weekday", nil),
					[]*service.Error{}),
				Entry("parses object that has invalid schedule name",
					&map[string]interface{}{"scheduleName": 0},
					testDataTypesBasal.NewTestSuppressed(nil, nil, nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotString(0), "/scheduleName", nil),
					}),
				Entry("parses object that has valid annotations",
					&map[string]interface{}{"annotations": []interface{}{map[string]interface{}{"a": "b"}}},
					testDataTypesBasal.NewTestSuppressed(nil, nil, &data.BlobArray{{"a": "b"}}, nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that has valid suppressed",
					&map[string]interface{}{"suppressed": map[string]interface{}{}},
					testDataTypesBasal.NewTestSuppressed(nil, nil, nil, nil, nil, testDataTypesBasal.NewTestSuppressed(nil, nil, nil, nil, nil, nil)),
					[]*service.Error{}),
				Entry("parses object that has valid suppressed that has multiple valid fields",
					&map[string]interface{}{"suppressed": map[string]interface{}{"type": "basal", "deliveryType": "scheduled", "rate": 1.0, "scheduleName": "Weekday"}},
					testDataTypesBasal.NewTestSuppressed(nil, nil, nil, nil, nil, testDataTypesBasal.NewTestSuppressed("basal", "scheduled", nil, 1.0, "Weekday", nil)),
					[]*service.Error{}),
				Entry("parses object that has invalid suppressed",
					&map[string]interface{}{"suppressed": 0},
					testDataTypesBasal.NewTestSuppressed(nil, nil, nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotObject(0), "/suppressed", nil),
					}),
				Entry("parses object that has multiple valid fields",
					&map[string]interface{}{"type": "basal", "deliveryType": "temp", "rate": 2.0, "annotations": []interface{}{map[string]interface{}{"a": "b"}}, "suppressed": map[string]interface{}{"type": "basal", "deliveryType": "scheduled", "rate": 1.0, "scheduleName": "Weekday", "annotations": []interface{}{map[string]interface{}{"d": "e"}}}},
					testDataTypesBasal.NewTestSuppressed("basal", "temp", &data.BlobArray{{"a": "b"}}, 2.0, nil, testDataTypesBasal.NewTestSuppressed("basal", "scheduled", &data.BlobArray{{"d": "e"}}, 1.0, "Weekday", nil)),
					[]*service.Error{}),
				Entry("parses object that has multiple invalid fields",
					&map[string]interface{}{"type": 0, "deliveryType": 0, "rate": "invalid", "scheduleName": 0, "suppressed": 0},
					testDataTypesBasal.NewTestSuppressed(nil, nil, nil, nil, nil, nil),
					[]*service.Error{
						testData.ComposeError(service.ErrorTypeNotString(0), "/type", nil),
						testData.ComposeError(service.ErrorTypeNotString(0), "/deliveryType", nil),
						testData.ComposeError(service.ErrorTypeNotFloat("invalid"), "/rate", nil),
						testData.ComposeError(service.ErrorTypeNotString(0), "/scheduleName", nil),
						testData.ComposeError(service.ErrorTypeNotObject(0), "/suppressed", nil),
					}),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum; delivery type temporary",
				func(mutator func(datum *basal.Suppressed), expectedErrors ...error) {
					datum := testDataTypesBasal.NewSuppressedTemporary()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(structureValidator.NewValidatableWithStringArrayAdapter(datum, &[]string{"scheduled", "temp"}), structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *basal.Suppressed) {},
				),
				Entry("delivery type missing",
					func(datum *basal.Suppressed) { datum.DeliveryType = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/deliveryType"),
				),
				Entry("delivery type invalid",
					func(datum *basal.Suppressed) { datum.DeliveryType = pointer.String("invalid") },
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"scheduled", "temp"}), "/deliveryType"),
				),
				Entry("type missing",
					func(datum *basal.Suppressed) { datum.Type = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/type"),
				),
				Entry("type invalid",
					func(datum *basal.Suppressed) { datum.Type = pointer.String("invalid") },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalid", "basal"), "/type"),
				),
				Entry("type basal",
					func(datum *basal.Suppressed) { datum.Type = pointer.String("basal") },
				),
				Entry("annotations missing",
					func(datum *basal.Suppressed) { datum.Annotations = nil },
				),
				Entry("annotations valid",
					func(datum *basal.Suppressed) { datum.Annotations = testData.NewBlobArray() },
				),
				Entry("rate missing",
					func(datum *basal.Suppressed) { datum.Rate = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("rate out of range (lower)",
					func(datum *basal.Suppressed) { datum.Rate = pointer.Float64(-0.1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/rate"),
				),
				Entry("rate at limit (lower)",
					func(datum *basal.Suppressed) { datum.Rate = pointer.Float64(0.0) },
				),
				Entry("rate at limit (upper)",
					func(datum *basal.Suppressed) { datum.Rate = pointer.Float64(100.0) },
				),
				Entry("rate out of range (upper)",
					func(datum *basal.Suppressed) { datum.Rate = pointer.Float64(100.1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/rate"),
				),
				Entry("schedule name missing",
					func(datum *basal.Suppressed) { datum.ScheduleName = nil },
				),
				Entry("schedule name exists",
					func(datum *basal.Suppressed) {
						datum.ScheduleName = pointer.String(testDataTypesBasal.NewScheduleName())
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/scheduleName"),
				),
				Entry("suppressed missing",
					func(datum *basal.Suppressed) { datum.Suppressed = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/suppressed"),
				),
				Entry("suppressed invalid",
					func(datum *basal.Suppressed) { datum.Suppressed.Type = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/suppressed/type"),
				),
				Entry("suppressed valid",
					func(datum *basal.Suppressed) { datum.Suppressed = testDataTypesBasal.NewSuppressedScheduled() },
				),
				Entry("multiple errors",
					func(datum *basal.Suppressed) {
						datum.Type = pointer.String("invalid")
						datum.Rate = pointer.Float64(100.1)
						datum.ScheduleName = pointer.String(testDataTypesBasal.NewScheduleName())
						datum.Suppressed = nil
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalid", "basal"), "/type"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/rate"),
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/scheduleName"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/suppressed"),
				),
			)

			DescribeTable("validates the datum; delivery type scheduled",
				func(mutator func(datum *basal.Suppressed), expectedErrors ...error) {
					datum := testDataTypesBasal.NewSuppressedScheduled()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(structureValidator.NewValidatableWithStringArrayAdapter(datum, &[]string{"scheduled"}), structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *basal.Suppressed) {},
				),
				Entry("delivery type missing",
					func(datum *basal.Suppressed) { datum.DeliveryType = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/deliveryType"),
				),
				Entry("delivery type invalid",
					func(datum *basal.Suppressed) { datum.DeliveryType = pointer.String("invalid") },
					testErrors.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"scheduled"}), "/deliveryType"),
				),
				Entry("type missing",
					func(datum *basal.Suppressed) { datum.Type = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/type"),
				),
				Entry("type invalid",
					func(datum *basal.Suppressed) { datum.Type = pointer.String("invalid") },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalid", "basal"), "/type"),
				),
				Entry("type basal",
					func(datum *basal.Suppressed) { datum.Type = pointer.String("basal") },
				),
				Entry("rate missing",
					func(datum *basal.Suppressed) { datum.Rate = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("rate out of range (lower)",
					func(datum *basal.Suppressed) { datum.Rate = pointer.Float64(-0.1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/rate"),
				),
				Entry("rate at limit (lower)",
					func(datum *basal.Suppressed) { datum.Rate = pointer.Float64(0.0) },
				),
				Entry("rate at limit (upper)",
					func(datum *basal.Suppressed) { datum.Rate = pointer.Float64(100.0) },
				),
				Entry("rate out of range (upper)",
					func(datum *basal.Suppressed) { datum.Rate = pointer.Float64(100.1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/rate"),
				),
				Entry("schedule name missing",
					func(datum *basal.Suppressed) { datum.ScheduleName = nil },
				),
				Entry("schedule name empty",
					func(datum *basal.Suppressed) { datum.ScheduleName = pointer.String("") },
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/scheduleName"),
				),
				Entry("schedule name valid",
					func(datum *basal.Suppressed) {
						datum.ScheduleName = pointer.String(testDataTypesBasal.NewScheduleName())
					},
				),
				Entry("suppressed missing",
					func(datum *basal.Suppressed) { datum.Suppressed = nil },
				),
				Entry("suppressed exists",
					func(datum *basal.Suppressed) { datum.Suppressed = testDataTypesBasal.NewSuppressedScheduled() },
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/suppressed"),
				),
				Entry("multiple errors",
					func(datum *basal.Suppressed) {
						datum.Type = pointer.String("invalid")
						datum.Rate = pointer.Float64(100.1)
						datum.ScheduleName = pointer.String("")
						datum.Suppressed = testDataTypesBasal.NewSuppressedScheduled()
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalid", "basal"), "/type"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/rate"),
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/scheduleName"),
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/suppressed"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *basal.Suppressed)) {
					for _, origin := range structure.Origins() {
						datum := testDataTypesBasal.NewSuppressedTemporary()
						mutator(datum)
						expectedDatum := testDataTypesBasal.CloneSuppressed(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *basal.Suppressed) {},
				),
				Entry("does not modify the datum; suppressed missing",
					func(datum *basal.Suppressed) { datum.Suppressed = nil },
				),
			)
		})
	})

	Context("FindAndRemoveDeliveryType", func() {
		DescribeTable("returns expected result when",
			func(allowedDeliveryTypes []string, deliveryType string, expectedResult []string, expectedBool bool) {
				var originalAllowedDeliveryTypes []string
				if allowedDeliveryTypes != nil {
					originalAllowedDeliveryTypes = append([]string{}, allowedDeliveryTypes...)
				}
				actualResult, actualBool := basal.FindAndRemoveDeliveryType(allowedDeliveryTypes, deliveryType)
				Expect(actualBool).To(Equal(expectedBool))
				Expect(actualResult).To(Equal(expectedResult))
				Expect(allowedDeliveryTypes).To(Equal(originalAllowedDeliveryTypes))
			},
			Entry("is an nil array", nil, "zero", nil, false),
			Entry("is an empty array ", []string{}, "zero", []string{}, false),
			Entry("is not found", []string{"one", "two", "three"}, "zero", []string{"one", "two", "three"}, false),
			Entry("is found at first position", []string{"one", "two", "three"}, "one", []string{"two", "three"}, true),
			Entry("is found at middle position", []string{"one", "two", "three"}, "two", []string{"one", "three"}, true),
			Entry("is found at last position", []string{"one", "two", "three"}, "three", []string{"one", "two"}, true),
		)
	})
})
