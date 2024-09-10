package pump_test

import (
	dataTypesStatusPump "github.com/tidepool-org/platform/data/types/status/pump"
	dataTypesStatusPumpTest "github.com/tidepool-org/platform/data/types/status/pump/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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

	Context("BolusDelivery", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesStatusPump.BolusDelivery)) {
				datum := dataTypesStatusPumpTest.RandomBolusDelivery()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataTypesStatusPumpTest.NewObjectFromBolusDelivery(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataTypesStatusPumpTest.NewObjectFromBolusDelivery(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataTypesStatusPump.BolusDelivery) {},
			),
			Entry("empty",
				func(datum *dataTypesStatusPump.BolusDelivery) {
					*datum = *dataTypesStatusPump.NewBolusDelivery()
				},
			),
			Entry("all",
				func(datum *dataTypesStatusPump.BolusDelivery) {
					datum.State = pointer.FromString(test.RandomStringFromArray(dataTypesStatusPump.BolusDeliveryStates()))
					datum.Dose = dataTypesStatusPumpTest.RandomBolusDose()
				},
			),
		)

		Context("ParseBolusDelivery", func() {
			It("returns nil when the object is missing", func() {
				Expect(dataTypesStatusPump.ParseBolusDelivery(structureParser.NewObject(logTest.NewLogger(), nil))).To(BeNil())
			})

			It("returns new datum when the object is valid", func() {
				datum := dataTypesStatusPumpTest.RandomBolusDelivery()
				object := dataTypesStatusPumpTest.NewObjectFromBolusDelivery(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(logTest.NewLogger(), &object)
				Expect(dataTypesStatusPump.ParseBolusDelivery(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("NewBolusDelivery", func() {
			It("returns the expected datum with all values initialized", func() {
				datum := dataTypesStatusPump.NewBolusDelivery()
				Expect(datum).ToNot(BeNil())
				Expect(datum.State).To(BeNil())
				Expect(datum.Dose).To(BeNil())
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dataTypesStatusPump.BolusDelivery), expectedErrors ...error) {
					expectedDatum := dataTypesStatusPumpTest.RandomBolusDelivery()
					object := dataTypesStatusPumpTest.NewObjectFromBolusDelivery(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := dataTypesStatusPump.NewBolusDelivery()
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dataTypesStatusPump.BolusDelivery) {},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *dataTypesStatusPump.BolusDelivery) {
						object["state"] = true
						object["dose"] = true
						expectedDatum.State = nil
						expectedDatum.Dose = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/state"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/dose"),
				),
			)
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

	Context("BolusDose", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesStatusPump.BolusDose)) {
				datum := dataTypesStatusPumpTest.RandomBolusDose()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataTypesStatusPumpTest.NewObjectFromBolusDose(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataTypesStatusPumpTest.NewObjectFromBolusDose(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataTypesStatusPump.BolusDose) {},
			),
			Entry("empty",
				func(datum *dataTypesStatusPump.BolusDose) {
					*datum = *dataTypesStatusPump.NewBolusDose()
				},
			),
			Entry("all",
				func(datum *dataTypesStatusPump.BolusDose) {
					datum.StartTime = pointer.FromTime(test.RandomTime())
					datum.Amount = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesStatusPump.BolusDoseAmountMinimum, dataTypesStatusPump.BolusDoseAmountMaximum))
					datum.AmountDelivered = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesStatusPump.BolusDoseAmountDeliveredMinimum, dataTypesStatusPump.BolusDoseAmountDeliveredMaximum))
				},
			),
		)

		Context("ParseBolusDose", func() {
			It("returns nil when the object is missing", func() {
				Expect(dataTypesStatusPump.ParseBolusDose(structureParser.NewObject(logTest.NewLogger(), nil))).To(BeNil())
			})

			It("returns new datum when the object is valid", func() {
				datum := dataTypesStatusPumpTest.RandomBolusDose()
				object := dataTypesStatusPumpTest.NewObjectFromBolusDose(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(logTest.NewLogger(), &object)
				Expect(dataTypesStatusPump.ParseBolusDose(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("NewBolusDose", func() {
			It("returns the expected datum with all values initialized", func() {
				datum := dataTypesStatusPump.NewBolusDose()
				Expect(datum).ToNot(BeNil())
				Expect(datum.StartTime).To(BeNil())
				Expect(datum.Amount).To(BeNil())
				Expect(datum.AmountDelivered).To(BeNil())
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dataTypesStatusPump.BolusDose), expectedErrors ...error) {
					expectedDatum := dataTypesStatusPumpTest.RandomBolusDose()
					object := dataTypesStatusPumpTest.NewObjectFromBolusDose(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := dataTypesStatusPump.NewBolusDose()
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dataTypesStatusPump.BolusDose) {},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *dataTypesStatusPump.BolusDose) {
						object["startTime"] = true
						object["amount"] = true
						object["amountDelivered"] = true
						expectedDatum.StartTime = nil
						expectedDatum.Amount = nil
						expectedDatum.AmountDelivered = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/startTime"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotFloat64(true), "/amount"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotFloat64(true), "/amountDelivered"),
				),
			)
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
				Entry("amount out of range (lower)",
					func(datum *dataTypesStatusPump.BolusDose) {
						datum.Amount = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0, 1000), "/amount"),
				),
				Entry("amount in range (lower)",
					func(datum *dataTypesStatusPump.BolusDose) {
						datum.Amount = pointer.FromFloat64(0)
					},
				),
				Entry("amount in range (upper)",
					func(datum *dataTypesStatusPump.BolusDose) {
						datum.Amount = pointer.FromFloat64(1000)
					},
				),
				Entry("amount out of range (upper)",
					func(datum *dataTypesStatusPump.BolusDose) {
						datum.Amount = pointer.FromFloat64(1000.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(1000.1, 0, 1000), "/amount"),
				),
				Entry("amount delivered out of range (lower)",
					func(datum *dataTypesStatusPump.BolusDose) {
						datum.AmountDelivered = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0, 1000), "/amountDelivered"),
				),
				Entry("amount delivered in range (lower)",
					func(datum *dataTypesStatusPump.BolusDose) {
						datum.AmountDelivered = pointer.FromFloat64(0)
					},
				),
				Entry("amount delivered in range (upper)",
					func(datum *dataTypesStatusPump.BolusDose) {
						datum.AmountDelivered = pointer.FromFloat64(1000)
					},
				),
				Entry("amount delivered out of range (upper)",
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
