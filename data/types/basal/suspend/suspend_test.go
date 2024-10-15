package suspend_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/basal"
	dataTypesBasalAutomated "github.com/tidepool-org/platform/data/types/basal/automated"
	dataTypesBasalAutomatedTest "github.com/tidepool-org/platform/data/types/basal/automated/test"
	dataTypesBasalScheduled "github.com/tidepool-org/platform/data/types/basal/scheduled"
	dataTypesBasalScheduledTest "github.com/tidepool-org/platform/data/types/basal/scheduled/test"
	"github.com/tidepool-org/platform/data/types/basal/suspend"
	dataTypesBasalTemporary "github.com/tidepool-org/platform/data/types/basal/temporary"
	dataTypesBasalTemporaryTest "github.com/tidepool-org/platform/data/types/basal/temporary/test"
	dataTypesBasalTest "github.com/tidepool-org/platform/data/types/basal/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewMeta() interface{} {
	return &basal.Meta{
		Type:         "basal",
		DeliveryType: "suspend",
	}
}

func NewSuspend() *suspend.Suspend {
	datum := suspend.New()
	datum.Basal = *dataTypesBasalTest.RandomBasal()
	datum.DeliveryType = "suspend"
	datum.Duration = pointer.FromInt(test.RandomIntFromRange(suspend.DurationMinimum, suspend.DurationMaximum))
	datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, suspend.DurationMaximum))
	datum.Suppressed = dataTypesBasalTemporaryTest.RandomSuppressedTemporary(dataTypesBasalScheduledTest.RandomSuppressedScheduled())
	return datum
}

func CloneSuspend(datum *suspend.Suspend) *suspend.Suspend {
	if datum == nil {
		return nil
	}
	clone := suspend.New()
	clone.Basal = *dataTypesBasalTest.CloneBasal(&datum.Basal)
	clone.Duration = pointer.CloneInt(datum.Duration)
	clone.DurationExpected = pointer.CloneInt(datum.DurationExpected)
	if datum.Suppressed != nil {
		switch suppressed := datum.Suppressed.(type) {
		case *dataTypesBasalAutomated.SuppressedAutomated:
			clone.Suppressed = dataTypesBasalAutomatedTest.CloneSuppressedAutomated(suppressed)
		case *dataTypesBasalScheduled.SuppressedScheduled:
			clone.Suppressed = dataTypesBasalScheduledTest.CloneSuppressedScheduled(suppressed)
		case *dataTypesBasalTemporary.SuppressedTemporary:
			clone.Suppressed = dataTypesBasalTemporaryTest.CloneSuppressedTemporary(suppressed)
		}
	}
	return clone
}

var _ = Describe("Suspend", func() {
	It("DeliveryType is expected", func() {
		Expect(suspend.DeliveryType).To(Equal("suspend"))
	})

	It("DurationMaximum is expected", func() {
		Expect(suspend.DurationMaximum).To(Equal(604800000))
	})

	It("DurationMinimum is expected", func() {
		Expect(suspend.DurationMinimum).To(Equal(0))
	})

	Context("New", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := suspend.New()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("basal"))
			Expect(datum.DeliveryType).To(Equal("suspend"))
			Expect(datum.Duration).To(BeNil())
			Expect(datum.DurationExpected).To(BeNil())
			Expect(datum.Suppressed).To(BeNil())
		})
	})

	Context("Suspend", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *suspend.Suspend), expectedErrors ...error) {
					datum := NewSuspend()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *suspend.Suspend) {},
				),
				Entry("type missing",
					func(datum *suspend.Suspend) { datum.Type = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &basal.Meta{DeliveryType: "suspend"}),
				),
				Entry("type invalid",
					func(datum *suspend.Suspend) { datum.Type = "invalidType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "basal"), "/type", &basal.Meta{Type: "invalidType", DeliveryType: "suspend"}),
				),
				Entry("type basal",
					func(datum *suspend.Suspend) { datum.Type = "basal" },
				),
				Entry("delivery type missing",
					func(datum *suspend.Suspend) { datum.DeliveryType = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/deliveryType", &basal.Meta{Type: "basal"}),
				),
				Entry("delivery type invalid",
					func(datum *suspend.Suspend) { datum.DeliveryType = "invalidDeliveryType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidDeliveryType", "suspend"), "/deliveryType", &basal.Meta{Type: "basal", DeliveryType: "invalidDeliveryType"}),
				),
				Entry("delivery type suspend",
					func(datum *suspend.Suspend) { datum.DeliveryType = "suspend" },
				),
				Entry("duration missing; duration expected missing",
					func(datum *suspend.Suspend) {
						datum.Duration = nil
						datum.DurationExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
				),
				Entry("duration missing; duration expected out of range (lower)",
					func(datum *suspend.Suspend) {
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration missing; duration expected in range (lower)",
					func(datum *suspend.Suspend) {
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
				),
				Entry("duration missing; duration expected in range (upper)",
					func(datum *suspend.Suspend) {
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(604800000)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
				),
				Entry("duration missing; duration expected out of range (upper)",
					func(datum *suspend.Suspend) {
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(604800001)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected missing",
					func(datum *suspend.Suspend) {
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/duration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected out of range (lower)",
					func(datum *suspend.Suspend) {
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected in range (lower)",
					func(datum *suspend.Suspend) {
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = pointer.FromInt(0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/duration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected in range (upper)",
					func(datum *suspend.Suspend) {
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = pointer.FromInt(604800000)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/duration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected out of range (upper)",
					func(datum *suspend.Suspend) {
						datum.Duration = pointer.FromInt(-1)
						datum.DurationExpected = pointer.FromInt(604800001)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (lower); duration expected missing",
					func(datum *suspend.Suspend) {
						datum.Duration = pointer.FromInt(0)
						datum.DurationExpected = nil
					},
				),
				Entry("duration in range (lower); duration expected out of range (lower)",
					func(datum *suspend.Suspend) {
						datum.Duration = pointer.FromInt(0)
						datum.DurationExpected = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (lower); duration expected in range (lower)",
					func(datum *suspend.Suspend) {
						datum.Duration = pointer.FromInt(0)
						datum.DurationExpected = pointer.FromInt(0)
					},
				),
				Entry("duration in range (lower); duration expected in range (upper)",
					func(datum *suspend.Suspend) {
						datum.Duration = pointer.FromInt(0)
						datum.DurationExpected = pointer.FromInt(604800000)
					},
				),
				Entry("duration in range (lower); duration expected out of range (upper)",
					func(datum *suspend.Suspend) {
						datum.Duration = pointer.FromInt(0)
						datum.DurationExpected = pointer.FromInt(604800001)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (upper); duration expected missing",
					func(datum *suspend.Suspend) {
						datum.Duration = pointer.FromInt(604800000)
						datum.DurationExpected = nil
					},
				),
				Entry("duration in range (upper); duration expected out of range (lower)",
					func(datum *suspend.Suspend) {
						datum.Duration = pointer.FromInt(604800000)
						datum.DurationExpected = pointer.FromInt(604799999)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604799999, 604800000, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (upper); duration expected in range (lower)",
					func(datum *suspend.Suspend) {
						datum.Duration = pointer.FromInt(604800000)
						datum.DurationExpected = pointer.FromInt(604800000)
					},
				),
				Entry("duration in range (upper); duration expected in range (upper)",
					func(datum *suspend.Suspend) {
						datum.Duration = pointer.FromInt(604800000)
						datum.DurationExpected = pointer.FromInt(604800000)
					},
				),
				Entry("duration in range (upper); duration expected out of range (upper)",
					func(datum *suspend.Suspend) {
						datum.Duration = pointer.FromInt(604800000)
						datum.DurationExpected = pointer.FromInt(604800001)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 604800000, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected missing",
					func(datum *suspend.Suspend) {
						datum.Duration = pointer.FromInt(604800001)
						datum.DurationExpected = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/duration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected out of range (lower)",
					func(datum *suspend.Suspend) {
						datum.Duration = pointer.FromInt(604800001)
						datum.DurationExpected = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected in range (lower)",
					func(datum *suspend.Suspend) {
						datum.Duration = pointer.FromInt(604800001)
						datum.DurationExpected = pointer.FromInt(0)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/duration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected in range (upper)",
					func(datum *suspend.Suspend) {
						datum.Duration = pointer.FromInt(604800001)
						datum.DurationExpected = pointer.FromInt(604800000)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/duration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected out of range (upper)",
					func(datum *suspend.Suspend) {
						datum.Duration = pointer.FromInt(604800001)
						datum.DurationExpected = pointer.FromInt(604800001)
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/duration", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("suppressed missing",
					func(datum *suspend.Suspend) { datum.Suppressed = nil },
				),
				Entry("suppressed automated",
					func(datum *suspend.Suspend) {
						datum.Suppressed = dataTypesBasalAutomatedTest.RandomSuppressedAutomated()
					},
				),
				Entry("suppressed scheduled",
					func(datum *suspend.Suspend) {
						datum.Suppressed = dataTypesBasalScheduledTest.RandomSuppressedScheduled()
					},
				),
				Entry("suppressed temporary with suppressed missing",
					func(datum *suspend.Suspend) {
						datum.Suppressed = dataTypesBasalTemporaryTest.RandomSuppressedTemporary(nil)
					},
				),
				Entry("suppressed temporary with suppressed automated",
					func(datum *suspend.Suspend) {
						datum.Suppressed = dataTypesBasalTemporaryTest.RandomSuppressedTemporary(dataTypesBasalAutomatedTest.RandomSuppressedAutomated())
					},
				),
				Entry("suppressed temporary with suppressed scheduled",
					func(datum *suspend.Suspend) {
						datum.Suppressed = dataTypesBasalTemporaryTest.RandomSuppressedTemporary(dataTypesBasalScheduledTest.RandomSuppressedScheduled())
					},
				),
				Entry("suppressed temporary with suppressed temporary with suppressed missing",
					func(datum *suspend.Suspend) {
						datum.Suppressed = dataTypesBasalTemporaryTest.RandomSuppressedTemporary(dataTypesBasalTemporaryTest.RandomSuppressedTemporary(nil))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotValid(), "/suppressed/suppressed", NewMeta()),
				),
				Entry("multiple errors",
					func(datum *suspend.Suspend) {
						datum.Type = "invalidType"
						datum.DeliveryType = "invalidDeliveryType"
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(604800001)
						datum.Suppressed = dataTypesBasalTemporaryTest.RandomSuppressedTemporary(dataTypesBasalTemporaryTest.RandomSuppressedTemporary(nil))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "basal"), "/type", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidDeliveryType", "suspend"), "/deliveryType", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/expectedDuration", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotValid(), "/suppressed/suppressed", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *suspend.Suspend)) {
					for _, origin := range structure.Origins() {
						datum := NewSuspend()
						mutator(datum)
						expectedDatum := CloneSuspend(datum)
						normalizer := dataNormalizer.New(logTest.NewLogger())
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *suspend.Suspend) {},
				),
				Entry("does not modify the datum; type missing",
					func(datum *suspend.Suspend) { datum.Type = "" },
				),
				Entry("does not modify the datum; delivery type missing",
					func(datum *suspend.Suspend) { datum.DeliveryType = "" },
				),
				Entry("does not modify the datum; duration missing",
					func(datum *suspend.Suspend) { datum.Duration = nil },
				),
				Entry("does not modify the datum; duration expected missing",
					func(datum *suspend.Suspend) { datum.DurationExpected = nil },
				),
				Entry("does not modify the datum; suppressed missing",
					func(datum *suspend.Suspend) { datum.Suppressed = nil },
				),
			)
		})
	})
})
