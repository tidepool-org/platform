package pump_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataTypesStatusPump "github.com/tidepool-org/platform/data/types/status/pump"
	dataTypesStatusPumpTest "github.com/tidepool-org/platform/data/types/status/pump/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("BolusDelivery", func() {
	It("BolusDeliveryStateCanceling is expected", func() {
		Expect(dataTypesStatusPump.BolusDeliveryStateCanceling).To(Equal("canceling"))
	})

	It("BolusDeliveryStateDelivering is expected", func() {
		Expect(dataTypesStatusPump.BolusDeliveryStateDelivering).To(Equal("delivering"))
	})

	It("BolusDeliveryStateInitiating is expected", func() {
		Expect(dataTypesStatusPump.BolusDeliveryStateInitiating).To(Equal("initiating"))
	})

	It("BolusDeliveryStateNone is expected", func() {
		Expect(dataTypesStatusPump.BolusDeliveryStateNone).To(Equal("none"))
	})

	It("BolusDoseAmountMaximum is expected", func() {
		Expect(dataTypesStatusPump.BolusDoseAmountMaximum).To(Equal(1000))
	})

	It("BolusDoseAmountMinimum is expected", func() {
		Expect(dataTypesStatusPump.BolusDoseAmountMinimum).To(Equal(0))
	})

	It("BolusDoseAmountDeliveredMaximum is expected", func() {
		Expect(dataTypesStatusPump.BolusDoseAmountDeliveredMaximum).To(Equal(1000))
	})

	It("BolusDoseAmountDeliveredMinimum is expected", func() {
		Expect(dataTypesStatusPump.BolusDoseAmountDeliveredMinimum).To(Equal(0))
	})

	It("BolusDeliveryStates returns expected", func() {
		Expect(dataTypesStatusPump.BolusDeliveryStates()).To(Equal([]string{"canceling", "delivering", "initiating", "none"}))
	})

	Context("ParseBolusDelivery", func() {
		// TODO
	})

	Context("NewBolusDelivery", func() {
		It("is successful", func() {
			Expect(dataTypesStatusPump.NewBolusDelivery()).To(Equal(&dataTypesStatusPump.BolusDelivery{}))
		})
	})

	Context("BolusDelivery", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("return the expected results when the input",
				func(mutator func(datum *dataTypesStatusPump.BolusDelivery), expectedErrors ...error) {
					datum := dataTypesStatusPumpTest.RandomBolusDelivery()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesStatusPump.BolusDelivery) {},
				),
				Entry("state missing",
					func(datum *dataTypesStatusPump.BolusDelivery) {
						datum.State = nil
						datum.Dose = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/state"),
				),
				Entry("state invalid",
					func(datum *dataTypesStatusPump.BolusDelivery) {
						datum.State = pointer.FromString("invalid")
						datum.Dose = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"canceling", "delivering", "initiating", "none"}), "/state"),
				),
				Entry("state canceling",
					func(datum *dataTypesStatusPump.BolusDelivery) {
						datum.State = pointer.FromString("canceling")
						datum.Dose = nil
					},
				),
				Entry("state canceling; dose exists",
					func(datum *dataTypesStatusPump.BolusDelivery) {
						datum.State = pointer.FromString("canceling")
						datum.Dose = dataTypesStatusPumpTest.RandomBolusDose()
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/dose"),
				),
				Entry("state delivering",
					func(datum *dataTypesStatusPump.BolusDelivery) {
						datum.State = pointer.FromString("delivering")
						datum.Dose = dataTypesStatusPumpTest.RandomBolusDose()
					},
				),
				Entry("state delivering; dose not exists",
					func(datum *dataTypesStatusPump.BolusDelivery) {
						datum.State = pointer.FromString("delivering")
						datum.Dose = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/dose"),
				),
				Entry("state initiating",
					func(datum *dataTypesStatusPump.BolusDelivery) {
						datum.State = pointer.FromString("initiating")
						datum.Dose = nil
					},
				),
				Entry("state initiating; dose exists",
					func(datum *dataTypesStatusPump.BolusDelivery) {
						datum.State = pointer.FromString("initiating")
						datum.Dose = dataTypesStatusPumpTest.RandomBolusDose()
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/dose"),
				),
				Entry("state none",
					func(datum *dataTypesStatusPump.BolusDelivery) {
						datum.State = pointer.FromString("none")
						datum.Dose = nil
					},
				),
				Entry("state none; dose exists",
					func(datum *dataTypesStatusPump.BolusDelivery) {
						datum.State = pointer.FromString("none")
						datum.Dose = dataTypesStatusPumpTest.RandomBolusDose()
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/dose"),
				),
				Entry("multiple errors",
					func(datum *dataTypesStatusPump.BolusDelivery) {
						datum.State = nil
						datum.Dose = dataTypesStatusPumpTest.RandomBolusDose()
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
			Expect(dataTypesStatusPump.NewBolusDose()).To(Equal(&dataTypesStatusPump.BolusDose{}))
		})
	})

	Context("BolusDose", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("return the expected results when the input",
				func(mutator func(datum *dataTypesStatusPump.BolusDose), expectedErrors ...error) {
					datum := dataTypesStatusPumpTest.RandomBolusDose()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesStatusPump.BolusDose) {},
				),
				Entry("amount missing",
					func(datum *dataTypesStatusPump.BolusDose) { datum.Amount = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/amount"),
				),
				Entry("amount below minimum",
					func(datum *dataTypesStatusPump.BolusDose) {
						datum.Amount = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0, 1000), "/amount"),
				),
				Entry("amount above maximum",
					func(datum *dataTypesStatusPump.BolusDose) {
						datum.Amount = pointer.FromFloat64(1000.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(1000.1, 0, 1000), "/amount"),
				),
				Entry("amount delivered below minimum",
					func(datum *dataTypesStatusPump.BolusDose) {
						datum.AmountDelivered = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0, 1000), "/amountDelivered"),
				),
				Entry("amount delivered above maximum",
					func(datum *dataTypesStatusPump.BolusDose) {
						datum.AmountDelivered = pointer.FromFloat64(1000.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(1000.1, 0, 1000), "/amountDelivered"),
				),
				Entry("multiple errors",
					func(datum *dataTypesStatusPump.BolusDose) {
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
