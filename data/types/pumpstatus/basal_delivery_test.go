package pumpstatus_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataTypesPumpStatus "github.com/tidepool-org/platform/data/types/pumpstatus"
	dataTypesPumpStatusTest "github.com/tidepool-org/platform/data/types/pumpstatus/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("BasalDelivery", func() {
	It("BasalDeliveryStateCancelingTemporary is expected", func() {
		Expect(dataTypesPumpStatus.BasalDeliveryStateCancelingTemporary).To(Equal("cancelingTemporary"))
	})

	It("BasalDeliveryStateInitiatingTemporary is expected", func() {
		Expect(dataTypesPumpStatus.BasalDeliveryStateInitiatingTemporary).To(Equal("initiatingTemporary"))
	})

	It("BasalDeliveryStateResuming is expected", func() {
		Expect(dataTypesPumpStatus.BasalDeliveryStateResuming).To(Equal("resuming"))
	})

	It("BasalDeliveryStateScheduled is expected", func() {
		Expect(dataTypesPumpStatus.BasalDeliveryStateScheduled).To(Equal("scheduled"))
	})

	It("BasalDeliveryStateSuspended is expected", func() {
		Expect(dataTypesPumpStatus.BasalDeliveryStateSuspended).To(Equal("suspended"))
	})

	It("BasalDeliveryStateSuspending is expected", func() {
		Expect(dataTypesPumpStatus.BasalDeliveryStateSuspending).To(Equal("suspending"))
	})

	It("BasalDeliveryStateTemporary is expected", func() {
		Expect(dataTypesPumpStatus.BasalDeliveryStateTemporary).To(Equal("temporary"))
	})

	It("BasalDoseAmountDeliveredMaximum is expected", func() {
		Expect(dataTypesPumpStatus.BasalDoseAmountDeliveredMaximum).To(Equal(1000))
	})

	It("BasalDoseAmountDeliveredMinimum is expected", func() {
		Expect(dataTypesPumpStatus.BasalDoseAmountDeliveredMinimum).To(Equal(0))
	})

	It("BasalDoseRateMaximum is expected", func() {
		Expect(dataTypesPumpStatus.BasalDoseRateMaximum).To(Equal(100))
	})

	It("BasalDoseRateMinimum is expected", func() {
		Expect(dataTypesPumpStatus.BasalDoseRateMinimum).To(Equal(0))
	})

	It("BasalDeliveryStates returns expected", func() {
		Expect(dataTypesPumpStatus.BasalDeliveryStates()).To(Equal([]string{"cancelingTemporary", "initiatingTemporary", "resuming", "scheduled", "suspended", "suspending", "temporary"}))
	})

	Context("ParseBasalDelivery", func() {
		// TODO
	})

	Context("NewBasalDelivery", func() {
		It("is successful", func() {
			Expect(dataTypesPumpStatus.NewBasalDelivery()).To(Equal(&dataTypesPumpStatus.BasalDelivery{}))
		})
	})

	Context("BasalDelivery", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("return the expected results when the input",
				func(mutator func(datum *dataTypesPumpStatus.BasalDelivery), expectedErrors ...error) {
					datum := dataTypesPumpStatusTest.RandomBasalDelivery()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesPumpStatus.BasalDelivery) {},
				),
				Entry("state missing",
					func(datum *dataTypesPumpStatus.BasalDelivery) {
						datum.State = nil
						datum.Time = nil
						datum.Dose = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/state"),
				),
				Entry("state invalid",
					func(datum *dataTypesPumpStatus.BasalDelivery) {
						datum.State = pointer.FromString("invalid")
						datum.Time = nil
						datum.Dose = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"cancelingTemporary", "initiatingTemporary", "resuming", "scheduled", "suspended", "suspending", "temporary"}), "/state"),
				),
				Entry("state cancelingTemporary",
					func(datum *dataTypesPumpStatus.BasalDelivery) {
						datum.State = pointer.FromString("cancelingTemporary")
						datum.Time = nil
						datum.Dose = nil
					},
				),
				Entry("state cancelingTemporary; time exists",
					func(datum *dataTypesPumpStatus.BasalDelivery) {
						datum.State = pointer.FromString("cancelingTemporary")
						datum.Time = pointer.FromTime(test.RandomTime())
						datum.Dose = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/time"),
				),
				Entry("state cancelingTemporary; dose exists",
					func(datum *dataTypesPumpStatus.BasalDelivery) {
						datum.State = pointer.FromString("cancelingTemporary")
						datum.Time = nil
						datum.Dose = dataTypesPumpStatusTest.RandomBasalDose()
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/dose"),
				),
				Entry("state initiatingTemporary",
					func(datum *dataTypesPumpStatus.BasalDelivery) {
						datum.State = pointer.FromString("initiatingTemporary")
						datum.Time = nil
						datum.Dose = nil
					},
				),
				Entry("state initiatingTemporary; time exists",
					func(datum *dataTypesPumpStatus.BasalDelivery) {
						datum.State = pointer.FromString("initiatingTemporary")
						datum.Time = pointer.FromTime(test.RandomTime())
						datum.Dose = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/time"),
				),
				Entry("state initiatingTemporary; dose exists",
					func(datum *dataTypesPumpStatus.BasalDelivery) {
						datum.State = pointer.FromString("initiatingTemporary")
						datum.Time = nil
						datum.Dose = dataTypesPumpStatusTest.RandomBasalDose()
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/dose"),
				),
				Entry("state resuming",
					func(datum *dataTypesPumpStatus.BasalDelivery) {
						datum.State = pointer.FromString("resuming")
						datum.Time = nil
						datum.Dose = nil
					},
				),
				Entry("state resuming; time exists",
					func(datum *dataTypesPumpStatus.BasalDelivery) {
						datum.State = pointer.FromString("resuming")
						datum.Time = pointer.FromTime(test.RandomTime())
						datum.Dose = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/time"),
				),
				Entry("state resuming; dose exists",
					func(datum *dataTypesPumpStatus.BasalDelivery) {
						datum.State = pointer.FromString("resuming")
						datum.Time = nil
						datum.Dose = dataTypesPumpStatusTest.RandomBasalDose()
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/dose"),
				),
				Entry("state scheduled",
					func(datum *dataTypesPumpStatus.BasalDelivery) {
						datum.State = pointer.FromString("scheduled")
						datum.Time = pointer.FromTime(test.RandomTime())
						datum.Dose = nil
					},
				),
				Entry("state scheduled; time not exists",
					func(datum *dataTypesPumpStatus.BasalDelivery) {
						datum.State = pointer.FromString("scheduled")
						datum.Time = nil
						datum.Dose = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/time"),
				),
				Entry("state scheduled; dose exists",
					func(datum *dataTypesPumpStatus.BasalDelivery) {
						datum.State = pointer.FromString("scheduled")
						datum.Time = pointer.FromTime(test.RandomTime())
						datum.Dose = dataTypesPumpStatusTest.RandomBasalDose()
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/dose"),
				),
				Entry("state suspended",
					func(datum *dataTypesPumpStatus.BasalDelivery) {
						datum.State = pointer.FromString("suspended")
						datum.Time = pointer.FromTime(test.RandomTime())
						datum.Dose = nil
					},
				),
				Entry("state suspended; time not exists",
					func(datum *dataTypesPumpStatus.BasalDelivery) {
						datum.State = pointer.FromString("suspended")
						datum.Time = nil
						datum.Dose = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/time"),
				),
				Entry("state suspended; dose exists",
					func(datum *dataTypesPumpStatus.BasalDelivery) {
						datum.State = pointer.FromString("suspended")
						datum.Time = pointer.FromTime(test.RandomTime())
						datum.Dose = dataTypesPumpStatusTest.RandomBasalDose()
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/dose"),
				),
				Entry("state suspending",
					func(datum *dataTypesPumpStatus.BasalDelivery) {
						datum.State = pointer.FromString("suspending")
						datum.Time = nil
						datum.Dose = nil
					},
				),
				Entry("state suspending; time exists",
					func(datum *dataTypesPumpStatus.BasalDelivery) {
						datum.State = pointer.FromString("suspending")
						datum.Time = pointer.FromTime(test.RandomTime())
						datum.Dose = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/time"),
				),
				Entry("state suspending; dose exists",
					func(datum *dataTypesPumpStatus.BasalDelivery) {
						datum.State = pointer.FromString("suspending")
						datum.Time = nil
						datum.Dose = dataTypesPumpStatusTest.RandomBasalDose()
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/dose"),
				),
				Entry("state temporary",
					func(datum *dataTypesPumpStatus.BasalDelivery) {
						datum.State = pointer.FromString("temporary")
						datum.Time = nil
						datum.Dose = dataTypesPumpStatusTest.RandomBasalDose()
					},
				),
				Entry("state temporary; time exists",
					func(datum *dataTypesPumpStatus.BasalDelivery) {
						datum.State = pointer.FromString("temporary")
						datum.Time = pointer.FromTime(test.RandomTime())
						datum.Dose = dataTypesPumpStatusTest.RandomBasalDose()
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/time"),
				),
				Entry("state temporary; dose not exists",
					func(datum *dataTypesPumpStatus.BasalDelivery) {
						datum.State = pointer.FromString("temporary")
						datum.Time = nil
						datum.Dose = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/dose"),
				),
				Entry("state temporary; dose invalid",
					func(datum *dataTypesPumpStatus.BasalDelivery) {
						datum.State = pointer.FromString("temporary")
						datum.Time = nil
						datum.Dose = dataTypesPumpStatusTest.RandomBasalDose()
						datum.Dose.Rate = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/dose/rate"),
				),
				Entry("multiple errors",
					func(datum *dataTypesPumpStatus.BasalDelivery) {
						datum.State = nil
						datum.Time = pointer.FromTime(test.RandomTime())
						datum.Dose = dataTypesPumpStatusTest.RandomBasalDose()
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/state"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/time"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/dose"),
				),
			)
		})
	})

	Context("ParseBasalDose", func() {
		// TODO
	})

	Context("NewBasalDose", func() {
		It("is successful", func() {
			Expect(dataTypesPumpStatus.NewBasalDose()).To(Equal(&dataTypesPumpStatus.BasalDose{}))
		})
	})

	Context("BasalDose", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("return the expected results when the input",
				func(mutator func(datum *dataTypesPumpStatus.BasalDose), expectedErrors ...error) {
					datum := dataTypesPumpStatusTest.RandomBasalDose()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesPumpStatus.BasalDose) {},
				),
				Entry("endTime before startTime",
					func(datum *dataTypesPumpStatus.BasalDose) {
						datum.StartTime = pointer.FromTime(test.PastNearTime())
						datum.EndTime = pointer.FromTime(test.PastFarTime())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotAfter(test.PastFarTime(), test.PastNearTime()), "/endTime"),
				),
				Entry("rate missing",
					func(datum *dataTypesPumpStatus.BasalDose) { datum.Rate = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("rate below minimum",
					func(datum *dataTypesPumpStatus.BasalDose) { datum.Rate = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0, 100), "/rate"),
				),
				Entry("rate above maximum",
					func(datum *dataTypesPumpStatus.BasalDose) { datum.Rate = pointer.FromFloat64(100.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(100.1, 0, 100), "/rate"),
				),
				Entry("amount delivered below minimum",
					func(datum *dataTypesPumpStatus.BasalDose) { datum.AmountDelivered = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0, 1000), "/amountDelivered"),
				),
				Entry("amount delivered above maximum",
					func(datum *dataTypesPumpStatus.BasalDose) { datum.AmountDelivered = pointer.FromFloat64(1000.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(1000.1, 0, 1000), "/amountDelivered"),
				),
				Entry("multiple errors",
					func(datum *dataTypesPumpStatus.BasalDose) {
						datum.StartTime = pointer.FromTime(test.PastNearTime())
						datum.EndTime = pointer.FromTime(test.PastFarTime())
						datum.Rate = nil
						datum.AmountDelivered = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotAfter(test.PastFarTime(), test.PastNearTime()), "/endTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0, 1000), "/amountDelivered"),
				),
			)
		})
	})
})
