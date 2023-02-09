package temporary_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/basal"
	dataTypesBasalAutomatedTest "github.com/tidepool-org/platform/data/types/basal/automated/test"
	dataTypesBasalScheduled "github.com/tidepool-org/platform/data/types/basal/scheduled"
	dataTypesBasalScheduledTest "github.com/tidepool-org/platform/data/types/basal/scheduled/test"
	"github.com/tidepool-org/platform/data/types/basal/temporary"
	dataTypesBasalTemporaryTest "github.com/tidepool-org/platform/data/types/basal/temporary/test"
	dataTypesBasalTest "github.com/tidepool-org/platform/data/types/basal/test"
	dataTypesInsulinTest "github.com/tidepool-org/platform/data/types/insulin/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewMeta() interface{} {
	return &basal.Meta{
		Type:         "basal",
		DeliveryType: "temp",
	}
}

func NewTemporary() *temporary.Temporary {
	datum := temporary.New()
	datum.Basal = *dataTypesBasalTest.RandomBasal()
	datum.DeliveryType = "temp"
	datum.Duration = pointer.FromInt(test.RandomIntFromRange(temporary.DurationMinimum, temporary.DurationMaximum))
	datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, temporary.DurationMaximum))
	datum.InsulinFormulation = dataTypesInsulinTest.RandomFormulation(3)
	datum.Percent = pointer.FromFloat64(test.RandomFloat64FromRange(temporary.PercentMinimum, temporary.PercentMaximum))
	datum.Rate = pointer.FromFloat64(test.RandomFloat64FromRange(temporary.RateMinimum, temporary.RateMaximum))
	datum.Suppressed = dataTypesBasalScheduledTest.RandomSuppressedScheduled()
	return datum
}

func CloneTemporary(datum *temporary.Temporary) *temporary.Temporary {
	if datum == nil {
		return nil
	}
	clone := temporary.New()
	clone.Basal = *dataTypesBasalTest.CloneBasal(&datum.Basal)
	clone.Duration = pointer.CloneInt(datum.Duration)
	clone.DurationExpected = pointer.CloneInt(datum.DurationExpected)
	clone.InsulinFormulation = dataTypesInsulinTest.CloneFormulation(datum.InsulinFormulation)
	clone.Percent = pointer.CloneFloat64(datum.Percent)
	clone.Rate = pointer.CloneFloat64(datum.Rate)
	if datum.Suppressed != nil {
		switch suppressed := datum.Suppressed.(type) {
		case *dataTypesBasalScheduled.SuppressedScheduled:
			clone.Suppressed = dataTypesBasalScheduledTest.CloneSuppressedScheduled(suppressed)
		}
	}
	return clone
}

var _ = Describe("Temporary", func() {
	It("DeliveryType is expected", func() {
		Expect(temporary.DeliveryType).To(Equal("temp"))
	})

	It("DurationMaximum is expected", func() {
		Expect(temporary.DurationMaximum).To(Equal(604800000))
	})

	It("DurationMinimum is expected", func() {
		Expect(temporary.DurationMinimum).To(Equal(0))
	})

	It("PercentMaximum is expected", func() {
		Expect(temporary.PercentMaximum).To(Equal(10.0))
	})

	It("PercentMinimum is expected", func() {
		Expect(temporary.PercentMinimum).To(Equal(0.0))
	})

	It("RateMaximum is expected", func() {
		Expect(temporary.RateMaximum).To(Equal(100.0))
	})

	It("RateMinimum is expected", func() {
		Expect(temporary.RateMinimum).To(Equal(0.0))
	})

	Context("New", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := temporary.New()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("basal"))
			Expect(datum.DeliveryType).To(Equal("temp"))
			Expect(datum.Duration).To(BeNil())
			Expect(datum.DurationExpected).To(BeNil())
			Expect(datum.InsulinFormulation).To(BeNil())
			Expect(datum.Percent).To(BeNil())
			Expect(datum.Rate).To(BeNil())
			Expect(datum.Suppressed).To(BeNil())
		})
	})

	Context("Temporary", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *temporary.Temporary), expectedErrors ...error) {
					datum := NewTemporary()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *temporary.Temporary) {},
				),
				Entry("type missing",
					func(datum *temporary.Temporary) { datum.Type = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &basal.Meta{DeliveryType: "temp"}),
				),
				Entry("type invalid",
					func(datum *temporary.Temporary) { datum.Type = "invalidType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "basal"), "/type", &basal.Meta{Type: "invalidType", DeliveryType: "temp"}),
				),
				Entry("type basal",
					func(datum *temporary.Temporary) { datum.Type = "basal" },
				),
				Entry("delivery type missing",
					func(datum *temporary.Temporary) { datum.DeliveryType = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/deliveryType", &basal.Meta{Type: "basal"}),
				),
				Entry("delivery type invalid",
					func(datum *temporary.Temporary) { datum.DeliveryType = "invalidDeliveryType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidDeliveryType", "temp"), "/deliveryType", &basal.Meta{Type: "basal", DeliveryType: "invalidDeliveryType"}),
				),
				Entry("delivery type temp",
					func(datum *temporary.Temporary) { datum.DeliveryType = "temp" },
				),
				Entry("duration missing; duration expected missing",
					func(datum *temporary.Temporary) {
						datum.Duration = nil
						datum.DurationExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
				),
				Entry("duration missing; duration expected out of range (lower)",
					func(datum *temporary.Temporary) {
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration missing; duration expected in range (lower)",
					func(datum *temporary.Temporary) {
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
				),
				Entry("duration missing; duration expected in range (upper)",
					func(datum *temporary.Temporary) {
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(604800000)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
				),
				Entry("duration missing; duration expected out of range (upper)",
					func(datum *temporary.Temporary) {
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(604800001)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected missing",
					func(datum *temporary.Temporary) {
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/duration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected out of range (lower)",
					func(datum *temporary.Temporary) {
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected in range (lower)",
					func(datum *temporary.Temporary) {
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = pointer.FromInt(0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/duration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected in range (upper)",
					func(datum *temporary.Temporary) {
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = pointer.FromInt(604800000)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/duration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected out of range (upper)",
					func(datum *temporary.Temporary) {
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = pointer.FromInt(604800001)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (lower); duration expected missing",
					func(datum *temporary.Temporary) {
						datum.Duration = pointer.FromInt(0)
						datum.DurationExpected = nil
					},
				),
				Entry("duration in range (lower); duration expected out of range (lower)",
					func(datum *temporary.Temporary) {
						datum.Duration = pointer.FromInt(0)
						datum.DurationExpected = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (lower); duration expected in range (lower)",
					func(datum *temporary.Temporary) {
						datum.Duration = pointer.FromInt(0)
						datum.DurationExpected = pointer.FromInt(0)
					},
				),
				Entry("duration in range (lower); duration expected in range (upper)",
					func(datum *temporary.Temporary) {
						datum.Duration = pointer.FromInt(0)
						datum.DurationExpected = pointer.FromInt(604800000)
					},
				),
				Entry("duration in range (lower); duration expected out of range (upper)",
					func(datum *temporary.Temporary) {
						datum.Duration = pointer.FromInt(0)
						datum.DurationExpected = pointer.FromInt(604800001)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (upper); duration expected missing",
					func(datum *temporary.Temporary) {
						datum.Duration = pointer.FromInt(604800000)
						datum.DurationExpected = nil
					},
				),
				Entry("duration in range (upper); duration expected out of range (lower)",
					func(datum *temporary.Temporary) {
						datum.Duration = pointer.FromInt(604800000)
						datum.DurationExpected = pointer.FromInt(604799999)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604799999, 604800000, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (upper); duration expected in range (lower)",
					func(datum *temporary.Temporary) {
						datum.Duration = pointer.FromInt(604800000)
						datum.DurationExpected = pointer.FromInt(604800000)
					},
				),
				Entry("duration in range (upper); duration expected in range (upper)",
					func(datum *temporary.Temporary) {
						datum.Duration = pointer.FromInt(604800000)
						datum.DurationExpected = pointer.FromInt(604800000)
					},
				),
				Entry("duration in range (upper); duration expected out of range (upper)",
					func(datum *temporary.Temporary) {
						datum.Duration = pointer.FromInt(604800000)
						datum.DurationExpected = pointer.FromInt(604800001)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 604800000, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected missing",
					func(datum *temporary.Temporary) {
						datum.Duration = pointer.FromInt(604800001)
						datum.DurationExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/duration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected out of range (lower)",
					func(datum *temporary.Temporary) {
						datum.Duration = pointer.FromInt(604800001)
						datum.DurationExpected = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected in range (lower)",
					func(datum *temporary.Temporary) {
						datum.Duration = pointer.FromInt(604800001)
						datum.DurationExpected = pointer.FromInt(0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/duration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected in range (upper)",
					func(datum *temporary.Temporary) {
						datum.Duration = pointer.FromInt(604800001)
						datum.DurationExpected = pointer.FromInt(604800000)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/duration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected out of range (upper)",
					func(datum *temporary.Temporary) {
						datum.Duration = pointer.FromInt(604800001)
						datum.DurationExpected = pointer.FromInt(604800001)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("insulin formulation missing",
					func(datum *temporary.Temporary) { datum.InsulinFormulation = nil },
				),
				Entry("insulin formulation invalid",
					func(datum *temporary.Temporary) {
						datum.InsulinFormulation.Compounds = nil
						datum.InsulinFormulation.Name = nil
						datum.InsulinFormulation.Simple = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/insulinFormulation/simple", NewMeta()),
				),
				Entry("insulin formulation valid",
					func(datum *temporary.Temporary) { datum.InsulinFormulation = dataTypesInsulinTest.RandomFormulation(3) },
				),
				Entry("percent missing",
					func(datum *temporary.Temporary) { datum.Percent = nil },
				),
				Entry("percent out of range (lower)",
					func(datum *temporary.Temporary) { datum.Percent = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 10.0), "/percent", NewMeta()),
				),
				Entry("percent in range (lower)",
					func(datum *temporary.Temporary) { datum.Percent = pointer.FromFloat64(0.0) },
				),
				Entry("percent in range (upper)",
					func(datum *temporary.Temporary) { datum.Percent = pointer.FromFloat64(10.0) },
				),
				Entry("percent out of range (upper)",
					func(datum *temporary.Temporary) { datum.Percent = pointer.FromFloat64(10.1) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(10.1, 0.0, 10.0), "/percent", NewMeta()),
				),
				Entry("rate missing",
					func(datum *temporary.Temporary) { datum.Rate = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/rate", NewMeta()),
				),
				Entry("rate out of range (lower)",
					func(datum *temporary.Temporary) { datum.Rate = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/rate", NewMeta()),
				),
				Entry("rate in range (lower)",
					func(datum *temporary.Temporary) { datum.Rate = pointer.FromFloat64(0.0) },
				),
				Entry("rate in range (upper)",
					func(datum *temporary.Temporary) { datum.Rate = pointer.FromFloat64(100.0) },
				),
				Entry("rate out of range (upper)",
					func(datum *temporary.Temporary) { datum.Rate = pointer.FromFloat64(100.1) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/rate", NewMeta()),
				),
				Entry("suppressed missing",
					func(datum *temporary.Temporary) { datum.Suppressed = nil },
				),
				Entry("suppressed automated",
					func(datum *temporary.Temporary) {
						datum.Suppressed = dataTypesBasalAutomatedTest.RandomSuppressedAutomated()
					},
				),
				Entry("suppressed scheduled",
					func(datum *temporary.Temporary) {
						datum.Suppressed = dataTypesBasalScheduledTest.RandomSuppressedScheduled()
					},
				),
				Entry("suppressed temporary with suppressed missing",
					func(datum *temporary.Temporary) {
						datum.Suppressed = dataTypesBasalTemporaryTest.RandomSuppressedTemporary(nil)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotValid(), "/suppressed", NewMeta()),
				),
				Entry("multiple errors",
					func(datum *temporary.Temporary) {
						datum.Type = "invalidType"
						datum.DeliveryType = "invalidDeliveryType"
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(604800001)
						datum.InsulinFormulation.Compounds = nil
						datum.InsulinFormulation.Name = nil
						datum.InsulinFormulation.Simple = nil
						datum.Percent = pointer.FromFloat64(10.1)
						datum.Rate = pointer.FromFloat64(100.1)
						datum.Suppressed = dataTypesBasalTemporaryTest.RandomSuppressedTemporary(nil)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "basal"), "/type", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidDeliveryType", "temp"), "/deliveryType", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/expectedDuration", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/insulinFormulation/simple", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(10.1, 0.0, 10.0), "/percent", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/rate", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotValid(), "/suppressed", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *temporary.Temporary)) {
					for _, origin := range structure.Origins() {
						datum := NewTemporary()
						mutator(datum)
						expectedDatum := CloneTemporary(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *temporary.Temporary) {},
				),
				Entry("does not modify the datum; type missing",
					func(datum *temporary.Temporary) { datum.Type = "" },
				),
				Entry("does not modify the datum; delivery type missing",
					func(datum *temporary.Temporary) { datum.DeliveryType = "" },
				),
				Entry("does not modify the datum; duration missing",
					func(datum *temporary.Temporary) { datum.Duration = nil },
				),
				Entry("does not modify the datum; duration expected missing",
					func(datum *temporary.Temporary) { datum.DurationExpected = nil },
				),
				Entry("does not modify the datum; insulin formulation missing",
					func(datum *temporary.Temporary) { datum.InsulinFormulation = nil },
				),
				Entry("does not modify the datum; percent missing",
					func(datum *temporary.Temporary) { datum.Percent = nil },
				),
				Entry("does not modify the datum; reate missing",
					func(datum *temporary.Temporary) { datum.Rate = nil },
				),
				Entry("does not modify the datum; suppressed missing",
					func(datum *temporary.Temporary) { datum.Suppressed = nil },
				),
			)
		})
	})

	Context("ParseSuppressedTemporary", func() {
		// TODO
	})

	Context("NewSuppressedTemporary", func() {
		It("returns the expected datum", func() {
			Expect(temporary.NewSuppressedTemporary()).To(Equal(&temporary.SuppressedTemporary{
				Type:         pointer.FromString("basal"),
				DeliveryType: pointer.FromString("temp"),
			}))
		})
	})

	Context("SuppressedTemporary", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *temporary.SuppressedTemporary), expectedErrors ...error) {
					datum := dataTypesBasalTemporaryTest.RandomSuppressedTemporary(dataTypesBasalScheduledTest.RandomSuppressedScheduled())
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *temporary.SuppressedTemporary) {},
				),
				Entry("type missing",
					func(datum *temporary.SuppressedTemporary) { datum.Type = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/type"),
				),
				Entry("type invalid",
					func(datum *temporary.SuppressedTemporary) { datum.Type = pointer.FromString("invalidType") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalidType", "basal"), "/type"),
				),
				Entry("type basal",
					func(datum *temporary.SuppressedTemporary) { datum.Type = pointer.FromString("basal") },
				),
				Entry("delivery type missing",
					func(datum *temporary.SuppressedTemporary) { datum.DeliveryType = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/deliveryType"),
				),
				Entry("delivery type invalid",
					func(datum *temporary.SuppressedTemporary) {
						datum.DeliveryType = pointer.FromString("invalidDeliveryType")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalidDeliveryType", "temp"), "/deliveryType"),
				),
				Entry("delivery type temp",
					func(datum *temporary.SuppressedTemporary) { datum.DeliveryType = pointer.FromString("temp") },
				),
				Entry("insulin formulation missing",
					func(datum *temporary.SuppressedTemporary) { datum.InsulinFormulation = nil },
				),
				Entry("insulin formulation invalid",
					func(datum *temporary.SuppressedTemporary) {
						datum.InsulinFormulation.Compounds = nil
						datum.InsulinFormulation.Name = nil
						datum.InsulinFormulation.Simple = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/insulinFormulation/simple"),
				),
				Entry("insulin formulation valid",
					func(datum *temporary.SuppressedTemporary) {
						datum.InsulinFormulation = dataTypesInsulinTest.RandomFormulation(3)
					},
				),
				Entry("percent missing",
					func(datum *temporary.SuppressedTemporary) { datum.Percent = nil },
				),
				Entry("percent out of range (lower)",
					func(datum *temporary.SuppressedTemporary) { datum.Percent = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 10.0), "/percent"),
				),
				Entry("percent in range (lower)",
					func(datum *temporary.SuppressedTemporary) { datum.Percent = pointer.FromFloat64(0.0) },
				),
				Entry("percent in range (upper)",
					func(datum *temporary.SuppressedTemporary) { datum.Percent = pointer.FromFloat64(10.0) },
				),
				Entry("percent out of range (upper)",
					func(datum *temporary.SuppressedTemporary) { datum.Percent = pointer.FromFloat64(10.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(10.1, 0.0, 10.0), "/percent"),
				),
				Entry("rate missing",
					func(datum *temporary.SuppressedTemporary) { datum.Rate = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("rate out of range (lower)",
					func(datum *temporary.SuppressedTemporary) { datum.Rate = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/rate"),
				),
				Entry("rate in range (lower)",
					func(datum *temporary.SuppressedTemporary) { datum.Rate = pointer.FromFloat64(0.0) },
				),
				Entry("rate in range (upper)",
					func(datum *temporary.SuppressedTemporary) { datum.Rate = pointer.FromFloat64(100.0) },
				),
				Entry("rate out of range (upper)",
					func(datum *temporary.SuppressedTemporary) { datum.Rate = pointer.FromFloat64(100.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/rate"),
				),
				Entry("suppressed missing",
					func(datum *temporary.SuppressedTemporary) { datum.Suppressed = nil },
				),
				Entry("suppressed automated",
					func(datum *temporary.SuppressedTemporary) {
						datum.Suppressed = dataTypesBasalAutomatedTest.RandomSuppressedAutomated()
					},
				),
				Entry("suppressed scheduled",
					func(datum *temporary.SuppressedTemporary) {
						datum.Suppressed = dataTypesBasalScheduledTest.RandomSuppressedScheduled()
					},
				),
				Entry("suppressed temporary with suppressed missing",
					func(datum *temporary.SuppressedTemporary) {
						datum.Suppressed = dataTypesBasalTemporaryTest.RandomSuppressedTemporary(nil)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/suppressed"),
				),
				Entry("multiple errors",
					func(datum *temporary.SuppressedTemporary) {
						datum.Type = pointer.FromString("invalidType")
						datum.DeliveryType = pointer.FromString("invalidDeliveryType")
						datum.InsulinFormulation.Compounds = nil
						datum.InsulinFormulation.Name = nil
						datum.InsulinFormulation.Simple = nil
						datum.Percent = pointer.FromFloat64(10.1)
						datum.Rate = pointer.FromFloat64(100.1)
						datum.Suppressed = dataTypesBasalTemporaryTest.RandomSuppressedTemporary(nil)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalidType", "basal"), "/type"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalidDeliveryType", "temp"), "/deliveryType"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/insulinFormulation/simple"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(10.1, 0.0, 10.0), "/percent"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/rate"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotValid(), "/suppressed"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *temporary.SuppressedTemporary)) {
					for _, origin := range structure.Origins() {
						datum := dataTypesBasalTemporaryTest.RandomSuppressedTemporary(dataTypesBasalScheduledTest.RandomSuppressedScheduled())
						mutator(datum)
						expectedDatum := dataTypesBasalTemporaryTest.CloneSuppressedTemporary(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *temporary.SuppressedTemporary) {},
				),
				Entry("does not modify the datum; type missing",
					func(datum *temporary.SuppressedTemporary) { datum.Type = nil },
				),
				Entry("does not modify the datum; delivery type missing",
					func(datum *temporary.SuppressedTemporary) { datum.DeliveryType = nil },
				),
				Entry("does not modify the datum; annotations missing",
					func(datum *temporary.SuppressedTemporary) { datum.Annotations = nil },
				),
				Entry("does not modify the datum; insulin formulation missing",
					func(datum *temporary.SuppressedTemporary) { datum.InsulinFormulation = nil },
				),
				Entry("does not modify the datum; percent missing",
					func(datum *temporary.SuppressedTemporary) { datum.Percent = nil },
				),
				Entry("does not modify the datum; reate missing",
					func(datum *temporary.SuppressedTemporary) { datum.Rate = nil },
				),
				Entry("does not modify the datum; suppressed missing",
					func(datum *temporary.SuppressedTemporary) { datum.Suppressed = nil },
				),
			)
		})
	})
})
