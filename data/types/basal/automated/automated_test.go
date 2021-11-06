package automated_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataTypesBasal "github.com/tidepool-org/platform/data/types/basal"
	dataTypesBasalAutomated "github.com/tidepool-org/platform/data/types/basal/automated"
	dataTypesBasalAutomatedTest "github.com/tidepool-org/platform/data/types/basal/automated/test"
	dataTypesBasalScheduledTest "github.com/tidepool-org/platform/data/types/basal/scheduled/test"
	dataTypesBasalTemporaryTest "github.com/tidepool-org/platform/data/types/basal/temporary/test"
	dataTypesBasalTest "github.com/tidepool-org/platform/data/types/basal/test"
	dataTypesInsulinTest "github.com/tidepool-org/platform/data/types/insulin/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	metadataTest "github.com/tidepool-org/platform/metadata/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewMeta() interface{} {
	return &dataTypesBasal.Meta{
		Type:         "basal",
		DeliveryType: "automated",
	}
}

var _ = Describe("Automated", func() {
	It("DeliveryType is expected", func() {
		Expect(dataTypesBasalAutomated.DeliveryType).To(Equal("automated"))
	})

	It("DurationMaximum is expected", func() {
		Expect(dataTypesBasalAutomated.DurationMaximum).To(Equal(604800000))
	})

	It("DurationMinimum is expected", func() {
		Expect(dataTypesBasalAutomated.DurationMinimum).To(Equal(0))
	})

	It("RateMaximum is expected", func() {
		Expect(dataTypesBasalAutomated.RateMaximum).To(Equal(100.0))
	})

	It("RateMinimum is expected", func() {
		Expect(dataTypesBasalAutomated.RateMinimum).To(Equal(0.0))
	})

	Context("Automated", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesBasalAutomated.Automated)) {
				datum := dataTypesBasalAutomatedTest.RandomAutomated()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataTypesBasalAutomatedTest.NewObjectFromAutomated(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataTypesBasalAutomatedTest.NewObjectFromAutomated(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataTypesBasalAutomated.Automated) {},
			),
			Entry("empty",
				func(datum *dataTypesBasalAutomated.Automated) {
					*datum = *dataTypesBasalAutomated.New()
				},
			),
			Entry("all",
				func(datum *dataTypesBasalAutomated.Automated) {
					datum.Duration = pointer.FromInt(test.RandomIntFromRange(dataTypesBasalAutomated.DurationMinimum, dataTypesBasalAutomated.DurationMaximum))
					datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, dataTypesBasalAutomated.DurationMaximum))
					datum.InsulinFormulation = dataTypesInsulinTest.RandomFormulation(3)
					datum.Rate = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesBasalAutomated.RateMinimum, dataTypesBasalAutomated.RateMaximum))
					datum.ScheduleName = pointer.FromString(dataTypesBasalTest.RandomScheduleName())
					datum.Suppressed = dataTypesBasalScheduledTest.RandomSuppressedScheduled()
				},
			),
		)

		Context("New", func() {
			It("returns the expected datum with all values initialized", func() {
				datum := dataTypesBasalAutomated.New()
				Expect(datum).ToNot(BeNil())
				Expect(datum.Type).To(Equal("basal"))
				Expect(datum.DeliveryType).To(Equal("automated"))
				Expect(datum.Duration).To(BeNil())
				Expect(datum.DurationExpected).To(BeNil())
				Expect(datum.InsulinFormulation).To(BeNil())
				Expect(datum.Rate).To(BeNil())
				Expect(datum.ScheduleName).To(BeNil())
				Expect(datum.Suppressed).To(BeNil())
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dataTypesBasalAutomated.Automated), expectedErrors ...error) {
					expectedDatum := dataTypesBasalAutomatedTest.RandomAutomatedForParser()
					object := dataTypesBasalAutomatedTest.NewObjectFromAutomated(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := dataTypesBasalAutomated.New()
					errorsTest.ExpectEqual(structureParser.NewObject(&object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dataTypesBasalAutomated.Automated) {},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *dataTypesBasalAutomated.Automated) {
						object["duration"] = true
						object["expectedDuration"] = true
						object["insulinFormulation"] = true
						object["rate"] = true
						object["scheduleName"] = true
						object["suppressed"] = dataTypesBasalTemporaryTest.NewObjectFromSuppressedTemporary(dataTypesBasalTemporaryTest.RandomSuppressedTemporary(nil), test.ObjectFormatJSON)
						expectedDatum.Duration = nil
						expectedDatum.DurationExpected = nil
						expectedDatum.InsulinFormulation = nil
						expectedDatum.Rate = nil
						expectedDatum.ScheduleName = nil
						expectedDatum.Suppressed = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotInt(true), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotInt(true), "/expectedDuration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotObject(true), "/insulinFormulation", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotFloat64(true), "/rate", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureParser.ErrorTypeNotString(true), "/scheduleName", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("temp", []string{"scheduled"}), "/suppressed/deliveryType", NewMeta()),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataTypesBasalAutomated.Automated), expectedErrors ...error) {
					datum := dataTypesBasalAutomatedTest.RandomAutomated()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesBasalAutomated.Automated) {},
				),
				Entry("type missing",
					func(datum *dataTypesBasalAutomated.Automated) { datum.Type = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &dataTypesBasal.Meta{DeliveryType: "automated"}),
				),
				Entry("type invalid",
					func(datum *dataTypesBasalAutomated.Automated) { datum.Type = "invalidType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "basal"), "/type", &dataTypesBasal.Meta{Type: "invalidType", DeliveryType: "automated"}),
				),
				Entry("type basal",
					func(datum *dataTypesBasalAutomated.Automated) { datum.Type = "basal" },
				),
				Entry("delivery type missing",
					func(datum *dataTypesBasalAutomated.Automated) { datum.DeliveryType = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/deliveryType", &dataTypesBasal.Meta{Type: "basal"}),
				),
				Entry("delivery type invalid",
					func(datum *dataTypesBasalAutomated.Automated) { datum.DeliveryType = "invalidDeliveryType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidDeliveryType", "automated"), "/deliveryType", &dataTypesBasal.Meta{Type: "basal", DeliveryType: "invalidDeliveryType"}),
				),
				Entry("delivery type automated",
					func(datum *dataTypesBasalAutomated.Automated) { datum.DeliveryType = "automated" },
				),
				Entry("duration missing; duration expected missing",
					func(datum *dataTypesBasalAutomated.Automated) {
						datum.Duration = nil
						datum.DurationExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
				),
				Entry("duration missing; duration expected out of range (lower)",
					func(datum *dataTypesBasalAutomated.Automated) {
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration missing; duration expected in range (lower)",
					func(datum *dataTypesBasalAutomated.Automated) {
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
				),
				Entry("duration missing; duration expected in range (upper)",
					func(datum *dataTypesBasalAutomated.Automated) {
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(604800000)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
				),
				Entry("duration missing; duration expected out of range (upper)",
					func(datum *dataTypesBasalAutomated.Automated) {
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(604800001)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected missing",
					func(datum *dataTypesBasalAutomated.Automated) {
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/duration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected out of range (lower)",
					func(datum *dataTypesBasalAutomated.Automated) {
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected in range (lower)",
					func(datum *dataTypesBasalAutomated.Automated) {
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = pointer.FromInt(0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/duration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected in range (upper)",
					func(datum *dataTypesBasalAutomated.Automated) {
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = pointer.FromInt(604800000)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/duration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected out of range (upper)",
					func(datum *dataTypesBasalAutomated.Automated) {
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = pointer.FromInt(604800001)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (lower); duration expected missing",
					func(datum *dataTypesBasalAutomated.Automated) {
						datum.Duration = pointer.FromInt(0)
						datum.DurationExpected = nil
					},
				),
				Entry("duration in range (lower); duration expected out of range (lower)",
					func(datum *dataTypesBasalAutomated.Automated) {
						datum.Duration = pointer.FromInt(0)
						datum.DurationExpected = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (lower); duration expected in range (lower)",
					func(datum *dataTypesBasalAutomated.Automated) {
						datum.Duration = pointer.FromInt(0)
						datum.DurationExpected = pointer.FromInt(0)
					},
				),
				Entry("duration in range (lower); duration expected in range (upper)",
					func(datum *dataTypesBasalAutomated.Automated) {
						datum.Duration = pointer.FromInt(0)
						datum.DurationExpected = pointer.FromInt(604800000)
					},
				),
				Entry("duration in range (lower); duration expected out of range (upper)",
					func(datum *dataTypesBasalAutomated.Automated) {
						datum.Duration = pointer.FromInt(0)
						datum.DurationExpected = pointer.FromInt(604800001)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (upper); duration expected missing",
					func(datum *dataTypesBasalAutomated.Automated) {
						datum.Duration = pointer.FromInt(604800000)
						datum.DurationExpected = nil
					},
				),
				Entry("duration in range (upper); duration expected out of range (lower)",
					func(datum *dataTypesBasalAutomated.Automated) {
						datum.Duration = pointer.FromInt(604800000)
						datum.DurationExpected = pointer.FromInt(604799999)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604799999, 604800000, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (upper); duration expected in range (lower)",
					func(datum *dataTypesBasalAutomated.Automated) {
						datum.Duration = pointer.FromInt(604800000)
						datum.DurationExpected = pointer.FromInt(604800000)
					},
				),
				Entry("duration in range (upper); duration expected in range (upper)",
					func(datum *dataTypesBasalAutomated.Automated) {
						datum.Duration = pointer.FromInt(604800000)
						datum.DurationExpected = pointer.FromInt(604800000)
					},
				),
				Entry("duration in range (upper); duration expected out of range (upper)",
					func(datum *dataTypesBasalAutomated.Automated) {
						datum.Duration = pointer.FromInt(604800000)
						datum.DurationExpected = pointer.FromInt(604800001)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 604800000, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected missing",
					func(datum *dataTypesBasalAutomated.Automated) {
						datum.Duration = pointer.FromInt(604800001)
						datum.DurationExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/duration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected out of range (lower)",
					func(datum *dataTypesBasalAutomated.Automated) {
						datum.Duration = pointer.FromInt(604800001)
						datum.DurationExpected = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected in range (lower)",
					func(datum *dataTypesBasalAutomated.Automated) {
						datum.Duration = pointer.FromInt(604800001)
						datum.DurationExpected = pointer.FromInt(0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/duration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected in range (upper)",
					func(datum *dataTypesBasalAutomated.Automated) {
						datum.Duration = pointer.FromInt(604800001)
						datum.DurationExpected = pointer.FromInt(604800000)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/duration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected out of range (upper)",
					func(datum *dataTypesBasalAutomated.Automated) {
						datum.Duration = pointer.FromInt(604800001)
						datum.DurationExpected = pointer.FromInt(604800001)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("insulin formulation missing",
					func(datum *dataTypesBasalAutomated.Automated) { datum.InsulinFormulation = nil },
				),
				Entry("insulin formulation invalid",
					func(datum *dataTypesBasalAutomated.Automated) {
						datum.InsulinFormulation.Compounds = nil
						datum.InsulinFormulation.Name = nil
						datum.InsulinFormulation.Simple = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/insulinFormulation/simple", NewMeta()),
				),
				Entry("insulin formulation valid",
					func(datum *dataTypesBasalAutomated.Automated) {
						datum.InsulinFormulation = dataTypesInsulinTest.RandomFormulation(3)
					},
				),
				Entry("rate missing",
					func(datum *dataTypesBasalAutomated.Automated) { datum.Rate = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/rate", NewMeta()),
				),
				Entry("rate out of range (lower)",
					func(datum *dataTypesBasalAutomated.Automated) { datum.Rate = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/rate", NewMeta()),
				),
				Entry("rate in range (lower)",
					func(datum *dataTypesBasalAutomated.Automated) { datum.Rate = pointer.FromFloat64(0.0) },
				),
				Entry("rate in range (upper)",
					func(datum *dataTypesBasalAutomated.Automated) { datum.Rate = pointer.FromFloat64(100.0) },
				),
				Entry("rate out of range (upper)",
					func(datum *dataTypesBasalAutomated.Automated) { datum.Rate = pointer.FromFloat64(100.1) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/rate", NewMeta()),
				),
				Entry("schedule name empty",
					func(datum *dataTypesBasalAutomated.Automated) { datum.ScheduleName = pointer.FromString("") },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/scheduleName", NewMeta()),
				),
				Entry("schedule name length; in range (upper)",
					func(datum *dataTypesBasalAutomated.Automated) {
						datum.ScheduleName = pointer.FromString(test.RandomStringFromRange(1000, 1000))
					},
				),
				Entry("schedule name length; out of range (upper)",
					func(datum *dataTypesBasalAutomated.Automated) {
						datum.ScheduleName = pointer.FromString(test.RandomStringFromRange(1001, 1001))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotLessThanOrEqualTo(1001, 1000), "/scheduleName", NewMeta()),
				),
				Entry("schedule name valid",
					func(datum *dataTypesBasalAutomated.Automated) {
						datum.ScheduleName = pointer.FromString(dataTypesBasalTest.RandomScheduleName())
					},
				),
				Entry("suppressed missing",
					func(datum *dataTypesBasalAutomated.Automated) { datum.Suppressed = nil },
				),
				Entry("suppressed invalid",
					func(datum *dataTypesBasalAutomated.Automated) {
						datum.Suppressed = dataTypesBasalTemporaryTest.RandomSuppressedTemporary(nil)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotValid(), "/suppressed", NewMeta()),
				),
				Entry("suppressed valid",
					func(datum *dataTypesBasalAutomated.Automated) {
						datum.Suppressed = dataTypesBasalScheduledTest.RandomSuppressedScheduled()
					},
				),
				Entry("multiple errors",
					func(datum *dataTypesBasalAutomated.Automated) {
						datum.Type = "invalidType"
						datum.DeliveryType = "invalidDeliveryType"
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(604800001)
						datum.InsulinFormulation.Compounds = nil
						datum.InsulinFormulation.Name = nil
						datum.InsulinFormulation.Simple = nil
						datum.Rate = pointer.FromFloat64(100.1)
						datum.ScheduleName = pointer.FromString("")
						datum.Suppressed = dataTypesBasalTemporaryTest.RandomSuppressedTemporary(nil)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "basal"), "/type", &dataTypesBasal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidDeliveryType", "automated"), "/deliveryType", &dataTypesBasal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", &dataTypesBasal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/expectedDuration", &dataTypesBasal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/insulinFormulation/simple", &dataTypesBasal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/rate", &dataTypesBasal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/scheduleName", &dataTypesBasal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotValid(), "/suppressed", &dataTypesBasal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"})),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *dataTypesBasalAutomated.Automated)) {
					for _, origin := range structure.Origins() {
						datum := dataTypesBasalAutomatedTest.RandomAutomated()
						mutator(datum)
						expectedDatum := dataTypesBasalAutomatedTest.CloneAutomated(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *dataTypesBasalAutomated.Automated) {},
				),
				Entry("does not modify the datum; type missing",
					func(datum *dataTypesBasalAutomated.Automated) { datum.Type = "" },
				),
				Entry("does not modify the datum; delivery type missing",
					func(datum *dataTypesBasalAutomated.Automated) { datum.DeliveryType = "" },
				),
				Entry("does not modify the datum; duration missing",
					func(datum *dataTypesBasalAutomated.Automated) { datum.Duration = nil },
				),
				Entry("does not modify the datum; duration expected missing",
					func(datum *dataTypesBasalAutomated.Automated) { datum.DurationExpected = nil },
				),
				Entry("does not modify the datum; insulin formulation missing",
					func(datum *dataTypesBasalAutomated.Automated) { datum.InsulinFormulation = nil },
				),
				Entry("does not modify the datum; rate missing",
					func(datum *dataTypesBasalAutomated.Automated) { datum.Rate = nil },
				),
				Entry("does not modify the datum; schedule name missing",
					func(datum *dataTypesBasalAutomated.Automated) { datum.ScheduleName = nil },
				),
				Entry("does not modify the datum; suppressed missing",
					func(datum *dataTypesBasalAutomated.Automated) { datum.Suppressed = nil },
				),
			)
		})
	})

	Context("ParseSuppressedAutomated", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesBasalAutomated.SuppressedAutomated)) {
				datum := dataTypesBasalAutomatedTest.RandomSuppressedAutomated()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataTypesBasalAutomatedTest.NewObjectFromSuppressedAutomated(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataTypesBasalAutomatedTest.NewObjectFromSuppressedAutomated(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataTypesBasalAutomated.SuppressedAutomated) {},
			),
			Entry("empty",
				func(datum *dataTypesBasalAutomated.SuppressedAutomated) {
					*datum = *dataTypesBasalAutomated.NewSuppressedAutomated()
				},
			),
			Entry("all",
				func(datum *dataTypesBasalAutomated.SuppressedAutomated) {
					datum.InsulinFormulation = dataTypesInsulinTest.RandomFormulation(3)
					datum.Rate = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesBasalAutomated.RateMinimum, dataTypesBasalAutomated.RateMaximum))
					datum.ScheduleName = pointer.FromString(dataTypesBasalTest.RandomScheduleName())
					datum.Suppressed = dataTypesBasalScheduledTest.RandomSuppressedScheduled()
				},
			),
		)

		Context("ParseSuppressedAutomated", func() {
			It("returns nil when the object is missing", func() {
				Expect(dataTypesBasalAutomated.ParseSuppressedAutomated(structureParser.NewObject(nil))).To(BeNil())
			})

			It("returns new datum when the object is valid", func() {
				datum := dataTypesBasalAutomatedTest.RandomSuppressedAutomated()
				object := dataTypesBasalAutomatedTest.NewObjectFromSuppressedAutomated(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(&object)
				Expect(dataTypesBasalAutomated.ParseSuppressedAutomated(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("NewSuppressedAutomated", func() {
			It("returns the expected datum with all values initialized", func() {
				datum := dataTypesBasalAutomated.NewSuppressedAutomated()
				Expect(datum).ToNot(BeNil())
				Expect(datum.Type).To(Equal(pointer.FromString("basal")))
				Expect(datum.DeliveryType).To(Equal(pointer.FromString("automated")))
				Expect(datum.InsulinFormulation).To(BeNil())
				Expect(datum.Rate).To(BeNil())
				Expect(datum.ScheduleName).To(BeNil())
				Expect(datum.Suppressed).To(BeNil())
			})
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataTypesBasalAutomated.SuppressedAutomated), expectedErrors ...error) {
					datum := dataTypesBasalAutomatedTest.RandomSuppressedAutomated()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesBasalAutomated.SuppressedAutomated) {},
				),
				Entry("type missing",
					func(datum *dataTypesBasalAutomated.SuppressedAutomated) { datum.Type = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/type"),
				),
				Entry("type invalid",
					func(datum *dataTypesBasalAutomated.SuppressedAutomated) {
						datum.Type = pointer.FromString("invalidType")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalidType", "basal"), "/type"),
				),
				Entry("type basal",
					func(datum *dataTypesBasalAutomated.SuppressedAutomated) { datum.Type = pointer.FromString("basal") },
				),
				Entry("delivery type missing",
					func(datum *dataTypesBasalAutomated.SuppressedAutomated) { datum.DeliveryType = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/deliveryType"),
				),
				Entry("delivery type invalid",
					func(datum *dataTypesBasalAutomated.SuppressedAutomated) {
						datum.DeliveryType = pointer.FromString("invalidDeliveryType")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalidDeliveryType", "automated"), "/deliveryType"),
				),
				Entry("delivery type automated",
					func(datum *dataTypesBasalAutomated.SuppressedAutomated) {
						datum.DeliveryType = pointer.FromString("automated")
					},
				),
				Entry("annotations missing",
					func(datum *dataTypesBasalAutomated.SuppressedAutomated) { datum.Annotations = nil },
				),
				Entry("annotations valid",
					func(datum *dataTypesBasalAutomated.SuppressedAutomated) {
						datum.Annotations = metadataTest.RandomMetadataArray()
					},
				),
				Entry("insulin formulation missing",
					func(datum *dataTypesBasalAutomated.SuppressedAutomated) { datum.InsulinFormulation = nil },
				),
				Entry("insulin formulation invalid",
					func(datum *dataTypesBasalAutomated.SuppressedAutomated) {
						datum.InsulinFormulation.Compounds = nil
						datum.InsulinFormulation.Name = nil
						datum.InsulinFormulation.Simple = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/insulinFormulation/simple"),
				),
				Entry("insulin formulation valid",
					func(datum *dataTypesBasalAutomated.SuppressedAutomated) {
						datum.InsulinFormulation = dataTypesInsulinTest.RandomFormulation(3)
					},
				),
				Entry("rate missing",
					func(datum *dataTypesBasalAutomated.SuppressedAutomated) { datum.Rate = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("rate out of range (lower)",
					func(datum *dataTypesBasalAutomated.SuppressedAutomated) { datum.Rate = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/rate"),
				),
				Entry("rate in range (lower)",
					func(datum *dataTypesBasalAutomated.SuppressedAutomated) { datum.Rate = pointer.FromFloat64(0.0) },
				),
				Entry("rate in range (upper)",
					func(datum *dataTypesBasalAutomated.SuppressedAutomated) { datum.Rate = pointer.FromFloat64(100.0) },
				),
				Entry("rate out of range (upper)",
					func(datum *dataTypesBasalAutomated.SuppressedAutomated) { datum.Rate = pointer.FromFloat64(100.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/rate"),
				),
				Entry("schedule name empty",
					func(datum *dataTypesBasalAutomated.SuppressedAutomated) { datum.ScheduleName = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/scheduleName"),
				),
				Entry("schedule name length; in range (upper)",
					func(datum *dataTypesBasalAutomated.SuppressedAutomated) {
						datum.ScheduleName = pointer.FromString(test.RandomStringFromRange(1000, 1000))
					},
				),
				Entry("schedule name length; out of range (upper)",
					func(datum *dataTypesBasalAutomated.SuppressedAutomated) {
						datum.ScheduleName = pointer.FromString(test.RandomStringFromRange(1001, 1001))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(1001, 1000), "/scheduleName"),
				),
				Entry("schedule name valid",
					func(datum *dataTypesBasalAutomated.SuppressedAutomated) {
						datum.ScheduleName = pointer.FromString(dataTypesBasalTest.RandomScheduleName())
					},
				),
				Entry("suppressed missing",
					func(datum *dataTypesBasalAutomated.SuppressedAutomated) { datum.Suppressed = nil },
				),
				Entry("suppressed invalid",
					func(datum *dataTypesBasalAutomated.SuppressedAutomated) {
						datum.Suppressed = dataTypesBasalTemporaryTest.RandomSuppressedTemporary(nil)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/suppressed"),
				),
				Entry("suppressed valid",
					func(datum *dataTypesBasalAutomated.SuppressedAutomated) {
						datum.Suppressed = dataTypesBasalScheduledTest.RandomSuppressedScheduled()
					},
				),
				Entry("multiple errors",
					func(datum *dataTypesBasalAutomated.SuppressedAutomated) {
						datum.Type = pointer.FromString("invalidType")
						datum.DeliveryType = pointer.FromString("invalidDeliveryType")
						datum.InsulinFormulation.Compounds = nil
						datum.InsulinFormulation.Name = nil
						datum.InsulinFormulation.Simple = nil
						datum.Rate = pointer.FromFloat64(100.1)
						datum.ScheduleName = pointer.FromString("")
						datum.Suppressed = dataTypesBasalTemporaryTest.RandomSuppressedTemporary(nil)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalidType", "basal"), "/type"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalidDeliveryType", "automated"), "/deliveryType"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/insulinFormulation/simple"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/rate"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/scheduleName"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/suppressed"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *dataTypesBasalAutomated.SuppressedAutomated)) {
					for _, origin := range structure.Origins() {
						datum := dataTypesBasalAutomatedTest.RandomSuppressedAutomated()
						mutator(datum)
						expectedDatum := dataTypesBasalAutomatedTest.CloneSuppressedAutomated(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *dataTypesBasalAutomated.SuppressedAutomated) {},
				),
				Entry("does not modify the datum; type missing",
					func(datum *dataTypesBasalAutomated.SuppressedAutomated) { datum.Type = nil },
				),
				Entry("does not modify the datum; delivery type missing",
					func(datum *dataTypesBasalAutomated.SuppressedAutomated) { datum.DeliveryType = nil },
				),
				Entry("does not modify the datum; annotations missing",
					func(datum *dataTypesBasalAutomated.SuppressedAutomated) { datum.Annotations = nil },
				),
				Entry("does not modify the datum; insulin formulation missing",
					func(datum *dataTypesBasalAutomated.SuppressedAutomated) { datum.InsulinFormulation = nil },
				),
				Entry("does not modify the datum; rate missing",
					func(datum *dataTypesBasalAutomated.SuppressedAutomated) { datum.Rate = nil },
				),
				Entry("does not modify the datum; schedule name missing",
					func(datum *dataTypesBasalAutomated.SuppressedAutomated) { datum.ScheduleName = nil },
				),
				Entry("does not modify the datum; suppressed missing",
					func(datum *dataTypesBasalAutomated.SuppressedAutomated) { datum.Suppressed = nil },
				),
			)
		})
	})
})
