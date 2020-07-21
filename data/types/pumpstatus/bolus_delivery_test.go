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
)

var _ = Describe("BolusDelivery", func() {
	It("BolusDeliveryStateCanceling is expected", func() {
		Expect(dataTypesPumpStatus.BolusDeliveryStateCanceling).To(Equal("canceling"))
	})

	It("BolusDeliveryStateDelivering is expected", func() {
		Expect(dataTypesPumpStatus.BolusDeliveryStateDelivering).To(Equal("delivering"))
	})

	It("BolusDeliveryStateInitiating is expected", func() {
		Expect(dataTypesPumpStatus.BolusDeliveryStateInitiating).To(Equal("initiating"))
	})

	It("BolusDeliveryStateNone is expected", func() {
		Expect(dataTypesPumpStatus.BolusDeliveryStateNone).To(Equal("none"))
	})

	It("BolusDoseAmountMaximum is expected", func() {
		Expect(dataTypesPumpStatus.BolusDoseAmountMaximum).To(Equal(1000))
	})

	It("BolusDoseAmountMinimum is expected", func() {
		Expect(dataTypesPumpStatus.BolusDoseAmountMinimum).To(Equal(0))
	})

	It("BolusDoseAmountDeliveredMaximum is expected", func() {
		Expect(dataTypesPumpStatus.BolusDoseAmountDeliveredMaximum).To(Equal(1000))
	})

	It("BolusDoseAmountDeliveredMinimum is expected", func() {
		Expect(dataTypesPumpStatus.BolusDoseAmountDeliveredMinimum).To(Equal(0))
	})

	It("BolusDeliveryStates returns expected", func() {
		Expect(dataTypesPumpStatus.BolusDeliveryStates()).To(Equal([]string{"canceling", "delivering", "initiating", "none"}))
	})

	Context("ParseBolusDelivery", func() {
		// TODO
	})

	Context("NewBolusDelivery", func() {
		It("is successful", func() {
			Expect(dataTypesPumpStatus.NewBolusDelivery()).To(Equal(&dataTypesPumpStatus.BolusDelivery{}))
		})
	})

	Context("BolusDelivery", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("return the expected results when the input",
				func(mutator func(datum *dataTypesPumpStatus.BolusDelivery), expectedErrors ...error) {
					datum := dataTypesPumpStatusTest.RandomBolusDelivery()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesPumpStatus.BolusDelivery) {},
				),
				Entry("state missing",
					func(datum *dataTypesPumpStatus.BolusDelivery) {
						datum.State = nil
						datum.Dose = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/state"),
				),
				Entry("state invalid",
					func(datum *dataTypesPumpStatus.BolusDelivery) {
						datum.State = pointer.FromString("invalid")
						datum.Dose = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"canceling", "delivering", "initiating", "none"}), "/state"),
				),
				Entry("state canceling",
					func(datum *dataTypesPumpStatus.BolusDelivery) {
						datum.State = pointer.FromString("canceling")
						datum.Dose = nil
					},
				),
				Entry("state canceling; dose exists",
					func(datum *dataTypesPumpStatus.BolusDelivery) {
						datum.State = pointer.FromString("canceling")
						datum.Dose = dataTypesPumpStatusTest.RandomBolusDose()
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/dose"),
				),
				Entry("state delivering",
					func(datum *dataTypesPumpStatus.BolusDelivery) {
						datum.State = pointer.FromString("delivering")
						datum.Dose = dataTypesPumpStatusTest.RandomBolusDose()
					},
				),
				Entry("state delivering; dose not exists",
					func(datum *dataTypesPumpStatus.BolusDelivery) {
						datum.State = pointer.FromString("delivering")
						datum.Dose = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/dose"),
				),
				Entry("state initiating",
					func(datum *dataTypesPumpStatus.BolusDelivery) {
						datum.State = pointer.FromString("initiating")
						datum.Dose = nil
					},
				),
				Entry("state initiating; dose exists",
					func(datum *dataTypesPumpStatus.BolusDelivery) {
						datum.State = pointer.FromString("initiating")
						datum.Dose = dataTypesPumpStatusTest.RandomBolusDose()
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/dose"),
				),
				Entry("state none",
					func(datum *dataTypesPumpStatus.BolusDelivery) {
						datum.State = pointer.FromString("none")
						datum.Dose = nil
					},
				),
				Entry("state none; dose exists",
					func(datum *dataTypesPumpStatus.BolusDelivery) {
						datum.State = pointer.FromString("none")
						datum.Dose = dataTypesPumpStatusTest.RandomBolusDose()
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/dose"),
				),
				Entry("multiple errors",
					func(datum *dataTypesPumpStatus.BolusDelivery) {
						datum.State = nil
						datum.Dose = dataTypesPumpStatusTest.RandomBolusDose()
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/state"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/dose"),
				),
			)
		})
	})

	Context("ParseBolusDose", func() {
		// TODO
	})

	Context("NewBolusDose", func() {
		It("is successful", func() {
			Expect(dataTypesPumpStatus.NewBolusDose()).To(Equal(&dataTypesPumpStatus.BolusDose{}))
		})
	})

	Context("BolusDose", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("return the expected results when the input",
				func(mutator func(datum *dataTypesPumpStatus.BolusDose), expectedErrors ...error) {
					datum := dataTypesPumpStatusTest.RandomBolusDose()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesPumpStatus.BolusDose) {},
				),
				Entry("amount missing",
					func(datum *dataTypesPumpStatus.BolusDose) { datum.Amount = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/amount"),
				),
				Entry("amount below minimum",
					func(datum *dataTypesPumpStatus.BolusDose) {
						datum.Amount = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0, 1000), "/amount"),
				),
				Entry("amount above maximum",
					func(datum *dataTypesPumpStatus.BolusDose) {
						datum.Amount = pointer.FromFloat64(1000.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(1000.1, 0, 1000), "/amount"),
				),
				Entry("amount delivered below minimum",
					func(datum *dataTypesPumpStatus.BolusDose) {
						datum.AmountDelivered = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0, 1000), "/amountDelivered"),
				),
				Entry("amount delivered above maximum",
					func(datum *dataTypesPumpStatus.BolusDose) {
						datum.AmountDelivered = pointer.FromFloat64(1000.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(1000.1, 0, 1000), "/amountDelivered"),
				),
				Entry("multiple errors",
					func(datum *dataTypesPumpStatus.BolusDose) {
						datum.Amount = nil
						datum.AmountDelivered = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/amount"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0, 1000), "/amountDelivered"),
				),
			)
		})
	})
})
