package automated_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	testData "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/basal"
	"github.com/tidepool-org/platform/data/types/basal/automated"
	testDataTypesBasalAutomated "github.com/tidepool-org/platform/data/types/basal/automated/test"
	testDataTypesBasal "github.com/tidepool-org/platform/data/types/basal/test"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewMeta() interface{} {
	return &basal.Meta{
		Type:         "basal",
		DeliveryType: "automated",
	}
}

func NewAutomated() *automated.Automated {
	datum := automated.New()
	datum.Basal = *testDataTypesBasal.NewBasal()
	datum.DeliveryType = "automated"
	datum.Duration = pointer.Int(test.RandomIntFromRange(automated.DurationMinimum, automated.DurationMaximum))
	datum.DurationExpected = pointer.Int(test.RandomIntFromRange(*datum.Duration, automated.DurationMaximum))
	datum.Rate = pointer.Float64(test.RandomFloat64FromRange(automated.RateMinimum, automated.RateMaximum))
	datum.ScheduleName = pointer.String(testDataTypesBasal.NewScheduleName())
	return datum
}

func CloneAutomated(datum *automated.Automated) *automated.Automated {
	if datum == nil {
		return nil
	}
	clone := automated.New()
	clone.Basal = *testDataTypesBasal.CloneBasal(&datum.Basal)
	clone.Duration = test.CloneInt(datum.Duration)
	clone.DurationExpected = test.CloneInt(datum.DurationExpected)
	clone.Rate = test.CloneFloat64(datum.Rate)
	clone.ScheduleName = test.CloneString(datum.ScheduleName)
	return clone
}

var _ = Describe("Automated", func() {
	It("DeliveryType is expected", func() {
		Expect(automated.DeliveryType).To(Equal("automated"))
	})

	It("DurationMaximum is expected", func() {
		Expect(automated.DurationMaximum).To(Equal(604800000))
	})

	It("DurationMinimum is expected", func() {
		Expect(automated.DurationMinimum).To(Equal(0))
	})

	It("RateMaximum is expected", func() {
		Expect(automated.RateMaximum).To(Equal(100.0))
	})

	It("RateMinimum is expected", func() {
		Expect(automated.RateMinimum).To(Equal(0.0))
	})

	Context("New", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := automated.New()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("basal"))
			Expect(datum.DeliveryType).To(Equal("automated"))
			Expect(datum.Duration).To(BeNil())
			Expect(datum.DurationExpected).To(BeNil())
			Expect(datum.Rate).To(BeNil())
			Expect(datum.ScheduleName).To(BeNil())
		})
	})

	Context("Automated", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *automated.Automated), expectedErrors ...error) {
					datum := NewAutomated()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *automated.Automated) {},
				),
				Entry("type missing",
					func(datum *automated.Automated) { datum.Type = "" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &basal.Meta{DeliveryType: "automated"}),
				),
				Entry("type invalid",
					func(datum *automated.Automated) { datum.Type = "invalidType" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "basal"), "/type", &basal.Meta{Type: "invalidType", DeliveryType: "automated"}),
				),
				Entry("type basal",
					func(datum *automated.Automated) { datum.Type = "basal" },
				),
				Entry("delivery type missing",
					func(datum *automated.Automated) { datum.DeliveryType = "" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/deliveryType", &basal.Meta{Type: "basal"}),
				),
				Entry("delivery type invalid",
					func(datum *automated.Automated) { datum.DeliveryType = "invalidDeliveryType" },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidDeliveryType", "automated"), "/deliveryType", &basal.Meta{Type: "basal", DeliveryType: "invalidDeliveryType"}),
				),
				Entry("delivery type automated",
					func(datum *automated.Automated) { datum.DeliveryType = "automated" },
				),
				Entry("duration missing; duration expected missing",
					func(datum *automated.Automated) {
						datum.Duration = nil
						datum.DurationExpected = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
				),
				Entry("duration missing; duration expected out of range (lower)",
					func(datum *automated.Automated) {
						datum.Duration = nil
						datum.DurationExpected = pointer.Int(-1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration missing; duration expected in range (lower)",
					func(datum *automated.Automated) {
						datum.Duration = nil
						datum.DurationExpected = pointer.Int(0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
				),
				Entry("duration missing; duration expected in range (upper)",
					func(datum *automated.Automated) {
						datum.Duration = nil
						datum.DurationExpected = pointer.Int(604800000)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
				),
				Entry("duration missing; duration expected out of range (upper)",
					func(datum *automated.Automated) {
						datum.Duration = nil
						datum.DurationExpected = pointer.Int(604800001)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected missing",
					func(datum *automated.Automated) {
						datum.Duration = pointer.Int(-1)
						datum.DurationExpected = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/duration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected out of range (lower)",
					func(datum *automated.Automated) {
						datum.Duration = pointer.Int(-1)
						datum.DurationExpected = pointer.Int(-1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/duration", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected in range (lower)",
					func(datum *automated.Automated) {
						datum.Duration = pointer.Int(-1)
						datum.DurationExpected = pointer.Int(0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/duration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected in range (upper)",
					func(datum *automated.Automated) {
						datum.Duration = pointer.Int(-1)
						datum.DurationExpected = pointer.Int(604800000)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/duration", NewMeta()),
				),
				Entry("duration out of range (lower); duration expected out of range (upper)",
					func(datum *automated.Automated) {
						datum.Duration = pointer.Int(-1)
						datum.DurationExpected = pointer.Int(604800001)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/duration", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (lower); duration expected missing",
					func(datum *automated.Automated) {
						datum.Duration = pointer.Int(0)
						datum.DurationExpected = nil
					},
				),
				Entry("duration in range (lower); duration expected out of range (lower)",
					func(datum *automated.Automated) {
						datum.Duration = pointer.Int(0)
						datum.DurationExpected = pointer.Int(-1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (lower); duration expected in range (lower)",
					func(datum *automated.Automated) {
						datum.Duration = pointer.Int(0)
						datum.DurationExpected = pointer.Int(0)
					},
				),
				Entry("duration in range (lower); duration expected in range (upper)",
					func(datum *automated.Automated) {
						datum.Duration = pointer.Int(0)
						datum.DurationExpected = pointer.Int(604800000)
					},
				),
				Entry("duration in range (lower); duration expected out of range (upper)",
					func(datum *automated.Automated) {
						datum.Duration = pointer.Int(0)
						datum.DurationExpected = pointer.Int(604800001)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (upper); duration expected missing",
					func(datum *automated.Automated) {
						datum.Duration = pointer.Int(604800000)
						datum.DurationExpected = nil
					},
				),
				Entry("duration in range (upper); duration expected out of range (lower)",
					func(datum *automated.Automated) {
						datum.Duration = pointer.Int(604800000)
						datum.DurationExpected = pointer.Int(604799999)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604799999, 604800000, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration in range (upper); duration expected in range (lower)",
					func(datum *automated.Automated) {
						datum.Duration = pointer.Int(604800000)
						datum.DurationExpected = pointer.Int(604800000)
					},
				),
				Entry("duration in range (upper); duration expected in range (upper)",
					func(datum *automated.Automated) {
						datum.Duration = pointer.Int(604800000)
						datum.DurationExpected = pointer.Int(604800000)
					},
				),
				Entry("duration in range (upper); duration expected out of range (upper)",
					func(datum *automated.Automated) {
						datum.Duration = pointer.Int(604800000)
						datum.DurationExpected = pointer.Int(604800001)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 604800000, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected missing",
					func(datum *automated.Automated) {
						datum.Duration = pointer.Int(604800001)
						datum.DurationExpected = nil
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/duration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected out of range (lower)",
					func(datum *automated.Automated) {
						datum.Duration = pointer.Int(604800001)
						datum.DurationExpected = pointer.Int(-1)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/duration", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-1, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected in range (lower)",
					func(datum *automated.Automated) {
						datum.Duration = pointer.Int(604800001)
						datum.DurationExpected = pointer.Int(0)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/duration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected in range (upper)",
					func(datum *automated.Automated) {
						datum.Duration = pointer.Int(604800001)
						datum.DurationExpected = pointer.Int(604800000)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/duration", NewMeta()),
				),
				Entry("duration out of range (upper); duration expected out of range (upper)",
					func(datum *automated.Automated) {
						datum.Duration = pointer.Int(604800001)
						datum.DurationExpected = pointer.Int(604800001)
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/duration", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/expectedDuration", NewMeta()),
				),
				Entry("rate missing",
					func(datum *automated.Automated) { datum.Rate = nil },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/rate", NewMeta()),
				),
				Entry("rate out of range (lower)",
					func(datum *automated.Automated) { datum.Rate = pointer.Float64(-0.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/rate", NewMeta()),
				),
				Entry("rate in range (lower)",
					func(datum *automated.Automated) { datum.Rate = pointer.Float64(0.0) },
				),
				Entry("rate in range (upper)",
					func(datum *automated.Automated) { datum.Rate = pointer.Float64(100.0) },
				),
				Entry("rate out of range (upper)",
					func(datum *automated.Automated) { datum.Rate = pointer.Float64(100.1) },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/rate", NewMeta()),
				),
				Entry("schedule name empty",
					func(datum *automated.Automated) { datum.ScheduleName = pointer.String("") },
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/scheduleName", NewMeta()),
				),
				Entry("schedule name valid",
					func(datum *automated.Automated) {
						datum.ScheduleName = pointer.String(testDataTypesBasal.NewScheduleName())
					},
				),
				Entry("multiple errors",
					func(datum *automated.Automated) {
						datum.Type = "invalidType"
						datum.DeliveryType = "invalidDeliveryType"
						datum.Duration = nil
						datum.DurationExpected = pointer.Int(604800001)
						datum.Rate = pointer.Float64(100.1)
						datum.ScheduleName = pointer.String("")
					},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "basal"), "/type", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidDeliveryType", "automated"), "/deliveryType", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/expectedDuration", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/rate", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/scheduleName", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *automated.Automated)) {
					for _, origin := range structure.Origins() {
						datum := NewAutomated()
						mutator(datum)
						expectedDatum := CloneAutomated(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *automated.Automated) {},
				),
				Entry("does not modify the datum; type missing",
					func(datum *automated.Automated) { datum.Type = "" },
				),
				Entry("does not modify the datum; delivery type missing",
					func(datum *automated.Automated) { datum.DeliveryType = "" },
				),
				Entry("does not modify the datum; duration missing",
					func(datum *automated.Automated) { datum.Duration = nil },
				),
				Entry("does not modify the datum; duration expected missing",
					func(datum *automated.Automated) { datum.DurationExpected = nil },
				),
				Entry("does not modify the datum; rate missing",
					func(datum *automated.Automated) { datum.Rate = nil },
				),
				Entry("does not modify the datum; schedule name missing",
					func(datum *automated.Automated) { datum.ScheduleName = nil },
				),
			)
		})
	})

	Context("ParseSuppressedAutomated", func() {
		// TODO
	})

	Context("NewSuppressedAutomated", func() {
		It("returns the expected datum", func() {
			Expect(automated.NewSuppressedAutomated()).To(Equal(&automated.SuppressedAutomated{
				Type:         pointer.String("basal"),
				DeliveryType: pointer.String("automated"),
			}))
		})
	})

	Context("SuppressedAutomated", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *automated.SuppressedAutomated), expectedErrors ...error) {
					datum := testDataTypesBasalAutomated.NewSuppressedAutomated()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *automated.SuppressedAutomated) {},
				),
				Entry("type missing",
					func(datum *automated.SuppressedAutomated) { datum.Type = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/type"),
				),
				Entry("type invalid",
					func(datum *automated.SuppressedAutomated) { datum.Type = pointer.String("invalidType") },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalidType", "basal"), "/type"),
				),
				Entry("type basal",
					func(datum *automated.SuppressedAutomated) { datum.Type = pointer.String("basal") },
				),
				Entry("delivery type missing",
					func(datum *automated.SuppressedAutomated) { datum.DeliveryType = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/deliveryType"),
				),
				Entry("delivery type invalid",
					func(datum *automated.SuppressedAutomated) { datum.DeliveryType = pointer.String("invalidDeliveryType") },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalidDeliveryType", "automated"), "/deliveryType"),
				),
				Entry("delivery type automated",
					func(datum *automated.SuppressedAutomated) { datum.DeliveryType = pointer.String("automated") },
				),
				Entry("annotations missing",
					func(datum *automated.SuppressedAutomated) { datum.Annotations = nil },
				),
				Entry("annotations valid",
					func(datum *automated.SuppressedAutomated) { datum.Annotations = testData.NewBlobArray() },
				),
				Entry("rate missing",
					func(datum *automated.SuppressedAutomated) { datum.Rate = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("rate out of range (lower)",
					func(datum *automated.SuppressedAutomated) { datum.Rate = pointer.Float64(-0.1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 100.0), "/rate"),
				),
				Entry("rate in range (lower)",
					func(datum *automated.SuppressedAutomated) { datum.Rate = pointer.Float64(0.0) },
				),
				Entry("rate in range (upper)",
					func(datum *automated.SuppressedAutomated) { datum.Rate = pointer.Float64(100.0) },
				),
				Entry("rate out of range (upper)",
					func(datum *automated.SuppressedAutomated) { datum.Rate = pointer.Float64(100.1) },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/rate"),
				),
				Entry("schedule name empty",
					func(datum *automated.SuppressedAutomated) { datum.ScheduleName = pointer.String("") },
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/scheduleName"),
				),
				Entry("schedule name valid",
					func(datum *automated.SuppressedAutomated) {
						datum.ScheduleName = pointer.String(testDataTypesBasal.NewScheduleName())
					},
				),
				Entry("multiple errors",
					func(datum *automated.SuppressedAutomated) {
						datum.Type = pointer.String("invalidType")
						datum.DeliveryType = pointer.String("invalidDeliveryType")
						datum.Rate = pointer.Float64(100.1)
						datum.ScheduleName = pointer.String("")
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalidType", "basal"), "/type"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalidDeliveryType", "automated"), "/deliveryType"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(100.1, 0.0, 100.0), "/rate"),
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/scheduleName"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *automated.SuppressedAutomated)) {
					for _, origin := range structure.Origins() {
						datum := testDataTypesBasalAutomated.NewSuppressedAutomated()
						mutator(datum)
						expectedDatum := testDataTypesBasalAutomated.CloneSuppressedAutomated(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *automated.SuppressedAutomated) {},
				),
				Entry("does not modify the datum; type missing",
					func(datum *automated.SuppressedAutomated) { datum.Type = nil },
				),
				Entry("does not modify the datum; delivery type missing",
					func(datum *automated.SuppressedAutomated) { datum.DeliveryType = nil },
				),
				Entry("does not modify the datum; annotations missing",
					func(datum *automated.SuppressedAutomated) { datum.Annotations = nil },
				),
				Entry("does not modify the datum; rate missing",
					func(datum *automated.SuppressedAutomated) { datum.Rate = nil },
				),
				Entry("does not modify the datum; schedule name missing",
					func(datum *automated.SuppressedAutomated) { datum.ScheduleName = nil },
				),
			)
		})
	})
})
