package scheduled_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/basal"
	"github.com/tidepool-org/platform/data/types/basal/scheduled"
	dataTypesBasalScheduledTest "github.com/tidepool-org/platform/data/types/basal/scheduled/test"
	dataTypesBasalTest "github.com/tidepool-org/platform/data/types/basal/test"
	dataTypesInsulinTest "github.com/tidepool-org/platform/data/types/insulin/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	metadataTest "github.com/tidepool-org/platform/metadata/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewMeta() interface{} {
	return &basal.Meta{
		Type:         "basal",
		DeliveryType: "scheduled",
	}
}

func NewScheduled() *scheduled.Scheduled {
	datum := scheduled.New()
	datum.Basal = *dataTypesBasalTest.NewBasal()
	datum.DeliveryType = "scheduled"
	datum.Duration = pointer.FromInt(test.RandomIntFromRange(scheduled.DurationMinimum, scheduled.DurationMaximum))
	datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, scheduled.DurationMaximum))
	datum.InsulinFormulation = dataTypesInsulinTest.NewFormulation(3)
	datum.Rate = pointer.FromFloat64(test.RandomFloat64FromRange(scheduled.RateMinimum, scheduled.RateMaximum))
	datum.ScheduleName = pointer.FromString(dataTypesBasalTest.NewScheduleName())
	return datum
}

func CloneScheduled(datum *scheduled.Scheduled) *scheduled.Scheduled {
	if datum == nil {
		return nil
	}
	clone := scheduled.New()
	clone.Basal = *dataTypesBasalTest.CloneBasal(&datum.Basal)
	clone.Duration = pointer.CloneInt(datum.Duration)
	clone.DurationExpected = pointer.CloneInt(datum.DurationExpected)
	clone.InsulinFormulation = dataTypesInsulinTest.CloneFormulation(datum.InsulinFormulation)
	clone.Rate = pointer.CloneFloat64(datum.Rate)
	clone.ScheduleName = pointer.CloneString(datum.ScheduleName)
	return clone
}

var _ = Describe("Scheduled", func() {
	It("DeliveryType is expected", func() {
		Expect(scheduled.DeliveryType).To(Equal("scheduled"))
	})

	It("DurationMaximum is expected", func() {
		Expect(scheduled.DurationMaximum).To(Equal(604800000))
	})

	It("DurationMinimum is expected", func() {
		Expect(scheduled.DurationMinimum).To(Equal(0))
	})

	It("RateMaximum is expected", func() {
		Expect(scheduled.RateMaximum).To(Equal(100.0))
	})

	It("RateMinimum is expected", func() {
		Expect(scheduled.RateMinimum).To(Equal(0.0))
	})

	Context("New", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := scheduled.New()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("basal"))
			Expect(datum.DeliveryType).To(Equal("scheduled"))
			Expect(datum.Duration).To(BeNil())
			Expect(datum.DurationExpected).To(BeNil())
			Expect(datum.InsulinFormulation).To(BeNil())
			Expect(datum.Rate).To(BeNil())
			Expect(datum.ScheduleName).To(BeNil())
		})
	})

	Context("Scheduled", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *scheduled.Scheduled), expectedErrors ...error) {
					datum := NewScheduled()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *scheduled.Scheduled) {},
				),
				Entry("type missing",
					func(datum *scheduled.Scheduled) { datum.Type = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &basal.Meta{DeliveryType: "scheduled"}),
				),
				Entry("type invalid",
					func(datum *scheduled.Scheduled) { datum.Type = "invalidType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "basal"), "/type", &basal.Meta{Type: "invalidType", DeliveryType: "scheduled"}),
				),
				Entry("type basal",
					func(datum *scheduled.Scheduled) { datum.Type = "basal" },
				),
				Entry("delivery type missing",
					func(datum *scheduled.Scheduled) { datum.DeliveryType = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/deliveryType", &basal.Meta{Type: "basal"}),
				),
				Entry("delivery type invalid",
					func(datum *scheduled.Scheduled) { datum.DeliveryType = "invalidDeliveryType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidDeliveryType", "scheduled"), "/deliveryType", &basal.Meta{Type: "basal", DeliveryType: "invalidDeliveryType"}),
				),
				Entry("delivery type scheduled",
					func(datum *scheduled.Scheduled) { datum.DeliveryType = "scheduled" },
				),
				Entry("duration missing; duration expected missing",
					func(datum *scheduled.Scheduled) {
						datum.Duration = nil
						datum.DurationExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
				),
				Entry("duration missing; duration expected out of range (lower)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
				),
				Entry("duration missing; duration expected in range (lower)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
				),
				Entry("duration missing; duration expected in range (upper)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(604800000)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
				),
				Entry("duration missing; duration expected out of range (upper)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(604800001)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected missing",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = nil
					},
				),
				Entry("duration out of range (lower); duration expected out of range (lower)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = pointer.FromInt(-1)
					},
				),
				Entry("duration out of range (lower); duration expected in range (lower)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = pointer.FromInt(0)
					},
				),
				Entry("duration out of range (lower); duration expected in range (upper)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = pointer.FromInt(604800000)
					},
				),
				Entry("duration out of range (lower); duration expected out of range (upper)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = pointer.FromInt(604800001)
					},
				),
				Entry("duration in range (lower); duration expected missing",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.FromInt(0)
						datum.DurationExpected = nil
					},
				),
				Entry("duration in range (lower); duration expected out of range (lower)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.FromInt(0)
						datum.DurationExpected = pointer.FromInt(-1)
					},
				),
				Entry("duration in range (lower); duration expected in range (lower)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.FromInt(0)
						datum.DurationExpected = pointer.FromInt(0)
					},
				),
				Entry("duration in range (lower); duration expected in range (upper)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.FromInt(0)
						datum.DurationExpected = pointer.FromInt(604800000)
					},
				),
				Entry("duration in range (lower); duration expected out of range (upper)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.FromInt(0)
						datum.DurationExpected = pointer.FromInt(604800001)
					},
				),
				Entry("duration in range (upper); duration expected missing",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.FromInt(604800000)
						datum.DurationExpected = nil
					},
				),
				Entry("duration in range (upper); duration expected out of range (lower)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.FromInt(604800000)
						datum.DurationExpected = pointer.FromInt(604799999)
					},
				),
				Entry("duration in range (upper); duration expected in range (lower)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.FromInt(604800000)
						datum.DurationExpected = pointer.FromInt(604800000)
					},
				),
				Entry("duration in range (upper); duration expected in range (upper)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.FromInt(604800000)
						datum.DurationExpected = pointer.FromInt(604800000)
					},
				),
				Entry("duration in range (upper); duration expected out of range (upper)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.FromInt(604800000)
						datum.DurationExpected = pointer.FromInt(604800001)
					},
				),
				Entry("duration out of range (upper); duration expected missing",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.FromInt(604800001)
						datum.DurationExpected = nil
					},
				),
				Entry("duration out of range (upper); duration expected out of range (lower)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.FromInt(604800001)
						datum.DurationExpected = pointer.FromInt(-1)
					},
				),
				Entry("duration out of range (upper); duration expected in range (lower)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.FromInt(604800001)
						datum.DurationExpected = pointer.FromInt(0)
					},
				),
				Entry("duration out of range (upper); duration expected in range (upper)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.FromInt(604800001)
						datum.DurationExpected = pointer.FromInt(604800000)
					},
				),
				Entry("duration out of range (upper); duration expected out of range (upper)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.FromInt(604800001)
						datum.DurationExpected = pointer.FromInt(604800001)
					},
				),
				Entry("insulin formulation missing",
					func(datum *scheduled.Scheduled) { datum.InsulinFormulation = nil },
				),
				Entry("insulin formulation invalid",
					func(datum *scheduled.Scheduled) {
						datum.InsulinFormulation.Compounds = nil
						datum.InsulinFormulation.Name = nil
						datum.InsulinFormulation.Simple = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/insulinFormulation/simple", NewMeta()),
				),
				Entry("insulin formulation valid",
					func(datum *scheduled.Scheduled) { datum.InsulinFormulation = dataTypesInsulinTest.NewFormulation(3) },
				),
				Entry("rate missing",
					func(datum *scheduled.Scheduled) { datum.Rate = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/rate", NewMeta()),
				),
				Entry("rate out of range (lower)",
					func(datum *scheduled.Scheduled) { datum.Rate = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/rate", NewMeta()),
				),
				Entry("rate in range (lower)",
					func(datum *scheduled.Scheduled) { datum.Rate = pointer.FromFloat64(0.0) },
				),
				Entry("rate in range (upper)",
					func(datum *scheduled.Scheduled) { datum.Rate = pointer.FromFloat64(100.0) },
				),
				Entry("rate out of range (upper)",
					func(datum *scheduled.Scheduled) { datum.Rate = pointer.FromFloat64(100.1) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/rate", NewMeta()),
				),
				Entry("schedule name empty",
					func(datum *scheduled.Scheduled) { datum.ScheduleName = pointer.FromString("") },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/scheduleName", NewMeta()),
				),
				Entry("schedule name valid",
					func(datum *scheduled.Scheduled) {
						datum.ScheduleName = pointer.FromString(dataTypesBasalTest.NewScheduleName())
					},
				),
				Entry("multiple errors",
					func(datum *scheduled.Scheduled) {
						datum.Type = "invalidType"
						datum.DeliveryType = "invalidDeliveryType"
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(604800001)
						datum.InsulinFormulation.Compounds = nil
						datum.InsulinFormulation.Name = nil
						datum.InsulinFormulation.Simple = nil
						datum.Rate = pointer.FromFloat64(100.1)
						datum.ScheduleName = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "basal"), "/type", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidDeliveryType", "scheduled"), "/deliveryType", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/insulinFormulation/simple", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/rate", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/scheduleName", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
				),
			)
			DescribeTable("validates the warnings on datum",
				func(mutator func(datum *scheduled.Scheduled), expectedErrors ...error) {
					datum := NewScheduled()
					mutator(datum)
					dataTypesTest.ValidateWarningsWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("duration missing; duration warning expected out of range (upper)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(604800001)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration missing; duration expected out of range (lower)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration missing; duration expected warning out of range (upper)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(604800001)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (lower); duration warning expected missing",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/duration", NewMeta()),
				),
				Entry("duration out of range (lower); duration warning expected out of range (lower)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (lower); duration warning expected in range (lower)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = pointer.FromInt(0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/duration", NewMeta()),
				),
				Entry("duration out of range (lower); duration warning expected in range (upper)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = pointer.FromInt(604800000)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/duration", NewMeta()),
				),
				Entry("duration out of range (lower); duration warning expected out of range (upper)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = pointer.FromInt(604800001)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (lower); duration warning expected out of range (lower)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.FromInt(0)
						datum.DurationExpected = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (lower); duration warning expected out of range (upper)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.FromInt(0)
						datum.DurationExpected = pointer.FromInt(604800001)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (upper); duration warning expected out of range (lower)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.FromInt(604800000)
						datum.DurationExpected = pointer.FromInt(604799999)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604799999, 604800000, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (upper); duration warning expected out of range (upper)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.FromInt(604800000)
						datum.DurationExpected = pointer.FromInt(604800001)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 604800000, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (upper); duration warning expected missing",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.FromInt(604800001)
						datum.DurationExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/duration", NewMeta()),
				),
				Entry("duration out of range (upper); duration warning expected out of range (lower)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.FromInt(604800001)
						datum.DurationExpected = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (upper); duration warning expected in range (lower)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.FromInt(604800001)
						datum.DurationExpected = pointer.FromInt(0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/duration", NewMeta()),
				),
				Entry("duration out of range (upper); duration warning expected in range (upper)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.FromInt(604800001)
						datum.DurationExpected = pointer.FromInt(604800000)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/duration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected out of range (upper)",
					func(datum *scheduled.Scheduled) {
						datum.Duration = pointer.FromInt(604800001)
						datum.DurationExpected = pointer.FromInt(604800001)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("multiple errors with a single warning",
					func(datum *scheduled.Scheduled) {
						datum.Type = "invalidType"
						datum.DeliveryType = "invalidDeliveryType"
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(604800001)
						datum.InsulinFormulation.Compounds = nil
						datum.InsulinFormulation.Name = nil
						datum.InsulinFormulation.Simple = nil
						datum.Rate = pointer.FromFloat64(100.1)
						datum.ScheduleName = pointer.FromString("")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/expectedDuration", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *scheduled.Scheduled)) {
					for _, origin := range structure.Origins() {
						datum := NewScheduled()
						mutator(datum)
						expectedDatum := CloneScheduled(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *scheduled.Scheduled) {},
				),
				Entry("does not modify the datum; type missing",
					func(datum *scheduled.Scheduled) { datum.Type = "" },
				),
				Entry("does not modify the datum; delivery type missing",
					func(datum *scheduled.Scheduled) { datum.DeliveryType = "" },
				),
				Entry("does not modify the datum; duration missing",
					func(datum *scheduled.Scheduled) { datum.Duration = nil },
				),
				Entry("does not modify the datum; duration expected missing",
					func(datum *scheduled.Scheduled) { datum.DurationExpected = nil },
				),
				Entry("does not modify the datum; insulin formulation missing",
					func(datum *scheduled.Scheduled) { datum.InsulinFormulation = nil },
				),
				Entry("does not modify the datum; rate missing",
					func(datum *scheduled.Scheduled) { datum.Rate = nil },
				),
				Entry("does not modify the datum; schedule name missing",
					func(datum *scheduled.Scheduled) { datum.ScheduleName = nil },
				),
			)
		})
	})

	Context("ParseSuppressedScheduled", func() {
		// TODO
	})

	Context("NewSuppressedScheduled", func() {
		It("returns the expected datum", func() {
			Expect(scheduled.NewSuppressedScheduled()).To(Equal(&scheduled.SuppressedScheduled{
				Type:         pointer.FromString("basal"),
				DeliveryType: pointer.FromString("scheduled"),
			}))
		})
	})

	Context("SuppressedScheduled", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *scheduled.SuppressedScheduled), expectedErrors ...error) {
					datum := dataTypesBasalScheduledTest.NewSuppressedScheduled()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *scheduled.SuppressedScheduled) {},
				),
				Entry("type missing",
					func(datum *scheduled.SuppressedScheduled) { datum.Type = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/type"),
				),
				Entry("type invalid",
					func(datum *scheduled.SuppressedScheduled) { datum.Type = pointer.FromString("invalidType") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalidType", "basal"), "/type"),
				),
				Entry("type basal",
					func(datum *scheduled.SuppressedScheduled) { datum.Type = pointer.FromString("basal") },
				),
				Entry("delivery type missing",
					func(datum *scheduled.SuppressedScheduled) { datum.DeliveryType = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/deliveryType"),
				),
				Entry("delivery type invalid",
					func(datum *scheduled.SuppressedScheduled) {
						datum.DeliveryType = pointer.FromString("invalidDeliveryType")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalidDeliveryType", "scheduled"), "/deliveryType"),
				),
				Entry("delivery type scheduled",
					func(datum *scheduled.SuppressedScheduled) { datum.DeliveryType = pointer.FromString("scheduled") },
				),
				Entry("annotations missing",
					func(datum *scheduled.SuppressedScheduled) { datum.Annotations = nil },
				),
				Entry("annotations valid",
					func(datum *scheduled.SuppressedScheduled) {
						datum.Annotations = metadataTest.RandomMetadataArray()
					},
				),
				Entry("insulin formulation missing",
					func(datum *scheduled.SuppressedScheduled) { datum.InsulinFormulation = nil },
				),
				Entry("insulin formulation invalid",
					func(datum *scheduled.SuppressedScheduled) {
						datum.InsulinFormulation.Compounds = nil
						datum.InsulinFormulation.Name = nil
						datum.InsulinFormulation.Simple = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/insulinFormulation/simple"),
				),
				Entry("insulin formulation valid",
					func(datum *scheduled.SuppressedScheduled) {
						datum.InsulinFormulation = dataTypesInsulinTest.NewFormulation(3)
					},
				),
				Entry("rate missing",
					func(datum *scheduled.SuppressedScheduled) { datum.Rate = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("rate out of range (lower)",
					func(datum *scheduled.SuppressedScheduled) { datum.Rate = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/rate"),
				),
				Entry("rate in range (lower)",
					func(datum *scheduled.SuppressedScheduled) { datum.Rate = pointer.FromFloat64(0.0) },
				),
				Entry("rate in range (upper)",
					func(datum *scheduled.SuppressedScheduled) { datum.Rate = pointer.FromFloat64(100.0) },
				),
				Entry("rate out of range (upper)",
					func(datum *scheduled.SuppressedScheduled) { datum.Rate = pointer.FromFloat64(100.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/rate"),
				),
				Entry("schedule name empty",
					func(datum *scheduled.SuppressedScheduled) { datum.ScheduleName = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/scheduleName"),
				),
				Entry("schedule name valid",
					func(datum *scheduled.SuppressedScheduled) {
						datum.ScheduleName = pointer.FromString(dataTypesBasalTest.NewScheduleName())
					},
				),
				Entry("multiple errors",
					func(datum *scheduled.SuppressedScheduled) {
						datum.Type = pointer.FromString("invalidType")
						datum.DeliveryType = pointer.FromString("invalidDeliveryType")
						datum.InsulinFormulation.Compounds = nil
						datum.InsulinFormulation.Name = nil
						datum.InsulinFormulation.Simple = nil
						datum.Rate = pointer.FromFloat64(100.1)
						datum.ScheduleName = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalidType", "basal"), "/type"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalidDeliveryType", "scheduled"), "/deliveryType"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/insulinFormulation/simple"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/rate"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/scheduleName"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *scheduled.SuppressedScheduled)) {
					for _, origin := range structure.Origins() {
						datum := dataTypesBasalScheduledTest.NewSuppressedScheduled()
						mutator(datum)
						expectedDatum := dataTypesBasalScheduledTest.CloneSuppressedScheduled(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *scheduled.SuppressedScheduled) {},
				),
				Entry("does not modify the datum; type missing",
					func(datum *scheduled.SuppressedScheduled) { datum.Type = nil },
				),
				Entry("does not modify the datum; delivery type missing",
					func(datum *scheduled.SuppressedScheduled) { datum.DeliveryType = nil },
				),
				Entry("does not modify the datum; annotations missing",
					func(datum *scheduled.SuppressedScheduled) { datum.Annotations = nil },
				),
				Entry("does not modify the datum; insulin formulation missing",
					func(datum *scheduled.SuppressedScheduled) { datum.InsulinFormulation = nil },
				),
				Entry("does not modify the datum; rate missing",
					func(datum *scheduled.SuppressedScheduled) { datum.Rate = nil },
				),
				Entry("does not modify the datum; schedule name missing",
					func(datum *scheduled.SuppressedScheduled) { datum.ScheduleName = nil },
				),
			)
		})
	})
})
