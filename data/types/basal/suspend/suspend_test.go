package suspend_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/context"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/parser"
	dataTest "github.com/tidepool-org/platform/data/test"
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
	"github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/service"
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
	datum.Basal = *dataTypesBasalTest.NewBasal()
	datum.DeliveryType = "suspend"
	datum.Duration = pointer.FromInt(test.RandomIntFromRange(suspend.DurationMinimum, suspend.DurationMaximum))
	datum.DurationExpected = pointer.FromInt(test.RandomIntFromRange(*datum.Duration, suspend.DurationMaximum))
	datum.Suppressed = dataTypesBasalTemporaryTest.NewSuppressedTemporary(dataTypesBasalScheduledTest.NewSuppressedScheduled())
	return datum
}

func CloneSuspend(datum *suspend.Suspend) *suspend.Suspend {
	if datum == nil {
		return nil
	}
	clone := suspend.New()
	clone.Basal = *dataTypesBasalTest.CloneBasal(&datum.Basal)
	clone.Duration = test.CloneInt(datum.Duration)
	clone.DurationExpected = test.CloneInt(datum.DurationExpected)
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

func NewTestSuspend(sourceTime interface{}, sourceDuration interface{}, sourceDurationExpected interface{}, sourceSuppressed suspend.Suppressed) *suspend.Suspend {
	datum := suspend.New()
	datum.DeviceID = pointer.FromString(dataTest.NewDeviceID())
	if val, ok := sourceTime.(string); ok {
		datum.Time = &val
	}
	if val, ok := sourceDuration.(int); ok {
		datum.Duration = &val
	}
	if val, ok := sourceDurationExpected.(int); ok {
		datum.DurationExpected = &val
	}
	datum.Suppressed = sourceSuppressed
	return datum
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
			var datum *suspend.Suspend

			BeforeEach(func() {
				datum = suspend.New()
				Expect(datum).ToNot(BeNil())
			})

			DescribeTable("parses the datum",
				func(sourceObject *map[string]interface{}, expectedDatum *suspend.Suspend, expectedErrors []*service.Error) {
					testContext, err := context.NewStandard(null.NewLogger())
					Expect(err).ToNot(HaveOccurred())
					Expect(testContext).ToNot(BeNil())
					testParser, err := parser.NewStandardObject(testContext, sourceObject, parser.AppendErrorNotParsed)
					Expect(err).ToNot(HaveOccurred())
					Expect(testParser).ToNot(BeNil())
					Expect(datum.Parse(testParser)).To(Succeed())
					Expect(datum.Time).To(Equal(expectedDatum.Time))
					Expect(datum.Duration).To(Equal(expectedDatum.Duration))
					Expect(datum.DurationExpected).To(Equal(expectedDatum.DurationExpected))
					if expectedDatum.Suppressed != nil {
						Expect(datum.Suppressed).To(Equal(expectedDatum.Suppressed))
					} else {
						Expect(datum.Suppressed).To(BeNil())
					}
					Expect(testContext.Errors()).To(ConsistOf(expectedErrors))
				},
				Entry("parses object that is nil",
					nil,
					NewTestSuspend(nil, nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that is empty",
					&map[string]interface{}{},
					NewTestSuspend(nil, nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that has valid time",
					&map[string]interface{}{"time": "2016-09-06T13:45:58-07:00"},
					NewTestSuspend("2016-09-06T13:45:58-07:00", nil, nil, nil),
					[]*service.Error{}),
				Entry("parses object that has invalid time",
					&map[string]interface{}{"time": 0},
					NewTestSuspend(nil, nil, nil, nil),
					[]*service.Error{
						dataTest.ComposeError(service.ErrorTypeNotString(0), "/time", NewMeta()),
					}),
				Entry("parses object that has valid duration",
					&map[string]interface{}{"duration": 3600000},
					NewTestSuspend(nil, 3600000, nil, nil),
					[]*service.Error{}),
				Entry("parses object that has invalid duration",
					&map[string]interface{}{"duration": "invalid"},
					NewTestSuspend(nil, nil, nil, nil),
					[]*service.Error{
						dataTest.ComposeError(service.ErrorTypeNotInteger("invalid"), "/duration", NewMeta()),
					}),
				Entry("parses object that has valid duration expected",
					&map[string]interface{}{"expectedDuration": 7200000},
					NewTestSuspend(nil, nil, 7200000, nil),
					[]*service.Error{}),
				Entry("parses object that has invalid duration expected",
					&map[string]interface{}{"expectedDuration": "invalid"},
					NewTestSuspend(nil, nil, nil, nil),
					[]*service.Error{
						dataTest.ComposeError(service.ErrorTypeNotInteger("invalid"), "/expectedDuration", NewMeta()),
					}),
				Entry("parses object that has valid suppressed",
					&map[string]interface{}{"suppressed": map[string]interface{}{"type": "basal", "deliveryType": "scheduled", "rate": 1.0, "scheduleName": "Weekday"}},
					NewTestSuspend(nil, nil, nil, &dataTypesBasalScheduled.SuppressedScheduled{Type: pointer.FromString("basal"), DeliveryType: pointer.FromString("scheduled"), Rate: pointer.FromFloat64(1.0), ScheduleName: pointer.FromString("Weekday")}),
					[]*service.Error{}),
				Entry("parses object that has invalid suppressed",
					&map[string]interface{}{"suppressed": "invalid"},
					NewTestSuspend(nil, nil, nil, nil),
					[]*service.Error{
						dataTest.ComposeError(service.ErrorTypeNotObject("invalid"), "/suppressed", NewMeta()),
					}),
				Entry("parses object that has multiple valid fields",
					&map[string]interface{}{"time": "2016-09-06T13:45:58-07:00", "duration": 3600000, "expectedDuration": 7200000},
					NewTestSuspend("2016-09-06T13:45:58-07:00", 3600000, 7200000, nil),
					[]*service.Error{}),
				Entry("parses object that has multiple invalid fields",
					&map[string]interface{}{"time": 0, "duration": "invalid", "expectedDuration": "invalid", "suppressed": "invalid"},
					NewTestSuspend(nil, nil, nil, nil),
					[]*service.Error{
						dataTest.ComposeError(service.ErrorTypeNotString(0), "/time", NewMeta()),
						dataTest.ComposeError(service.ErrorTypeNotInteger("invalid"), "/duration", NewMeta()),
						dataTest.ComposeError(service.ErrorTypeNotInteger("invalid"), "/expectedDuration", NewMeta()),
						dataTest.ComposeError(service.ErrorTypeNotObject("invalid"), "/suppressed", NewMeta()),
					}),
			)
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
						datum.Suppressed = dataTypesBasalAutomatedTest.NewSuppressedAutomated()
					},
				),
				Entry("suppressed scheduled",
					func(datum *suspend.Suspend) {
						datum.Suppressed = dataTypesBasalScheduledTest.NewSuppressedScheduled()
					},
				),
				Entry("suppressed temporary with suppressed missing",
					func(datum *suspend.Suspend) {
						datum.Suppressed = dataTypesBasalTemporaryTest.NewSuppressedTemporary(nil)
					},
				),
				Entry("suppressed temporary with suppressed automated",
					func(datum *suspend.Suspend) {
						datum.Suppressed = dataTypesBasalTemporaryTest.NewSuppressedTemporary(dataTypesBasalAutomatedTest.NewSuppressedAutomated())
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/suppressed/suppressed", NewMeta()),
				),
				Entry("suppressed temporary with suppressed scheduled",
					func(datum *suspend.Suspend) {
						datum.Suppressed = dataTypesBasalTemporaryTest.NewSuppressedTemporary(dataTypesBasalScheduledTest.NewSuppressedScheduled())
					},
				),
				Entry("suppressed temporary with suppressed temporary with suppressed missing",
					func(datum *suspend.Suspend) {
						datum.Suppressed = dataTypesBasalTemporaryTest.NewSuppressedTemporary(dataTypesBasalTemporaryTest.NewSuppressedTemporary(nil))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/suppressed/suppressed", NewMeta()),
				),
				Entry("multiple errors",
					func(datum *suspend.Suspend) {
						datum.Type = "invalidType"
						datum.DeliveryType = "invalidDeliveryType"
						datum.Duration = nil
						datum.DurationExpected = pointer.FromInt(604800001)
						datum.Suppressed = dataTypesBasalTemporaryTest.NewSuppressedTemporary(dataTypesBasalTemporaryTest.NewSuppressedTemporary(nil))
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "basal"), "/type", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidDeliveryType", "suspend"), "/deliveryType", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/duration", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(604800001, 0, 604800000), "/expectedDuration", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueExists(), "/suppressed/suppressed", &basal.Meta{Type: "invalidType", DeliveryType: "invalidDeliveryType"}),
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
						normalizer := dataNormalizer.New()
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
