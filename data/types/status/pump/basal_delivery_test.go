package pump_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

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
)

var _ = Describe("BasalDelivery", func() {
	It("BasalDeliveryStateCancelingTemporary is expected", func() {
		Expect(dataTypesStatusPump.BasalDeliveryStateCancelingTemporary).To(Equal("cancelingTemporary"))
	})

	It("BasalDeliveryStateInitiatingTemporary is expected", func() {
		Expect(dataTypesStatusPump.BasalDeliveryStateInitiatingTemporary).To(Equal("initiatingTemporary"))
	})

	It("BasalDeliveryStateResuming is expected", func() {
		Expect(dataTypesStatusPump.BasalDeliveryStateResuming).To(Equal("resuming"))
	})

	It("BasalDeliveryStateScheduled is expected", func() {
		Expect(dataTypesStatusPump.BasalDeliveryStateScheduled).To(Equal("scheduled"))
	})

	It("BasalDeliveryStateSuspended is expected", func() {
		Expect(dataTypesStatusPump.BasalDeliveryStateSuspended).To(Equal("suspended"))
	})

	It("BasalDeliveryStateSuspending is expected", func() {
		Expect(dataTypesStatusPump.BasalDeliveryStateSuspending).To(Equal("suspending"))
	})

	It("BasalDeliveryStateTemporary is expected", func() {
		Expect(dataTypesStatusPump.BasalDeliveryStateTemporary).To(Equal("temporary"))
	})

	It("BasalDoseAmountDeliveredMaximum is expected", func() {
		Expect(dataTypesStatusPump.BasalDoseAmountDeliveredMaximum).To(Equal(1000))
	})

	It("BasalDoseAmountDeliveredMinimum is expected", func() {
		Expect(dataTypesStatusPump.BasalDoseAmountDeliveredMinimum).To(Equal(0))
	})

	It("BasalDoseRateMaximum is expected", func() {
		Expect(dataTypesStatusPump.BasalDoseRateMaximum).To(Equal(100))
	})

	It("BasalDoseRateMinimum is expected", func() {
		Expect(dataTypesStatusPump.BasalDoseRateMinimum).To(Equal(0))
	})

	It("BasalDeliveryStates returns expected", func() {
		Expect(dataTypesStatusPump.BasalDeliveryStates()).To(Equal([]string{"cancelingTemporary", "initiatingTemporary", "resuming", "scheduled", "suspended", "suspending", "temporary"}))
	})

	Context("BasalDelivery", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesStatusPump.BasalDelivery)) {
				datum := dataTypesStatusPumpTest.RandomBasalDelivery()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataTypesStatusPumpTest.NewObjectFromBasalDelivery(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataTypesStatusPumpTest.NewObjectFromBasalDelivery(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataTypesStatusPump.BasalDelivery) {},
			),
			Entry("empty",
				func(datum *dataTypesStatusPump.BasalDelivery) {
					*datum = *dataTypesStatusPump.NewBasalDelivery()
				},
			),
			Entry("all",
				func(datum *dataTypesStatusPump.BasalDelivery) {
					datum.State = pointer.FromString(test.RandomStringFromArray(dataTypesStatusPump.BasalDeliveryStates()))
					datum.Time = pointer.FromTime(test.RandomTime())
					datum.Dose = dataTypesStatusPumpTest.RandomBasalDose()
				},
			),
		)

		Context("ParseBasalDelivery", func() {
			It("returns nil when the object is missing", func() {
				Expect(dataTypesStatusPump.ParseBasalDelivery(structureParser.NewObject(logTest.NewLogger(), nil))).To(BeNil())
			})

			It("returns new datum when the object is valid", func() {
				datum := dataTypesStatusPumpTest.RandomBasalDelivery()
				object := dataTypesStatusPumpTest.NewObjectFromBasalDelivery(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(logTest.NewLogger(), &object)
				Expect(dataTypesStatusPump.ParseBasalDelivery(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("NewBasalDelivery", func() {
			It("returns the expected datum with all values initialized", func() {
				datum := dataTypesStatusPump.NewBasalDelivery()
				Expect(datum).ToNot(BeNil())
				Expect(datum.State).To(BeNil())
				Expect(datum.Time).To(BeNil())
				Expect(datum.Dose).To(BeNil())
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dataTypesStatusPump.BasalDelivery), expectedErrors ...error) {
					expectedDatum := dataTypesStatusPumpTest.RandomBasalDelivery()
					object := dataTypesStatusPumpTest.NewObjectFromBasalDelivery(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := dataTypesStatusPump.NewBasalDelivery()
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dataTypesStatusPump.BasalDelivery) {},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *dataTypesStatusPump.BasalDelivery) {
						object["state"] = true
						object["time"] = true
						object["dose"] = true
						expectedDatum.State = nil
						expectedDatum.Time = nil
						expectedDatum.Dose = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/state"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/time"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/dose"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("return the expected results when the input",
				func(mutator func(datum *dataTypesStatusPump.BasalDelivery), expectedErrors ...error) {
					datum := dataTypesStatusPumpTest.RandomBasalDelivery()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesStatusPump.BasalDelivery) {},
				),
				Entry("state missing",
					func(datum *dataTypesStatusPump.BasalDelivery) {
						datum.State = nil
						datum.Time = nil
						datum.Dose = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/state"),
				),
				Entry("state invalid",
					func(datum *dataTypesStatusPump.BasalDelivery) {
						datum.State = pointer.FromString("invalid")
						datum.Time = nil
						datum.Dose = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"cancelingTemporary", "initiatingTemporary", "resuming", "scheduled", "suspended", "suspending", "temporary"}), "/state"),
				),
				Entry("state cancelingTemporary",
					func(datum *dataTypesStatusPump.BasalDelivery) {
						datum.State = pointer.FromString("cancelingTemporary")
						datum.Time = nil
						datum.Dose = nil
					},
				),
				Entry("state cancelingTemporary; time exists",
					func(datum *dataTypesStatusPump.BasalDelivery) {
						datum.State = pointer.FromString("cancelingTemporary")
						datum.Time = pointer.FromTime(test.RandomTime())
						datum.Dose = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/time"),
				),
				Entry("state cancelingTemporary; dose exists",
					func(datum *dataTypesStatusPump.BasalDelivery) {
						datum.State = pointer.FromString("cancelingTemporary")
						datum.Time = nil
						datum.Dose = dataTypesStatusPumpTest.RandomBasalDose()
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/dose"),
				),
				Entry("state initiatingTemporary",
					func(datum *dataTypesStatusPump.BasalDelivery) {
						datum.State = pointer.FromString("initiatingTemporary")
						datum.Time = nil
						datum.Dose = nil
					},
				),
				Entry("state initiatingTemporary; time exists",
					func(datum *dataTypesStatusPump.BasalDelivery) {
						datum.State = pointer.FromString("initiatingTemporary")
						datum.Time = pointer.FromTime(test.RandomTime())
						datum.Dose = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/time"),
				),
				Entry("state initiatingTemporary; dose exists",
					func(datum *dataTypesStatusPump.BasalDelivery) {
						datum.State = pointer.FromString("initiatingTemporary")
						datum.Time = nil
						datum.Dose = dataTypesStatusPumpTest.RandomBasalDose()
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/dose"),
				),
				Entry("state resuming",
					func(datum *dataTypesStatusPump.BasalDelivery) {
						datum.State = pointer.FromString("resuming")
						datum.Time = nil
						datum.Dose = nil
					},
				),
				Entry("state resuming; time exists",
					func(datum *dataTypesStatusPump.BasalDelivery) {
						datum.State = pointer.FromString("resuming")
						datum.Time = pointer.FromTime(test.RandomTime())
						datum.Dose = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/time"),
				),
				Entry("state resuming; dose exists",
					func(datum *dataTypesStatusPump.BasalDelivery) {
						datum.State = pointer.FromString("resuming")
						datum.Time = nil
						datum.Dose = dataTypesStatusPumpTest.RandomBasalDose()
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/dose"),
				),
				Entry("state scheduled",
					func(datum *dataTypesStatusPump.BasalDelivery) {
						datum.State = pointer.FromString("scheduled")
						datum.Time = pointer.FromTime(test.RandomTime())
						datum.Dose = nil
					},
				),
				Entry("state scheduled; time not exists",
					func(datum *dataTypesStatusPump.BasalDelivery) {
						datum.State = pointer.FromString("scheduled")
						datum.Time = nil
						datum.Dose = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/time"),
				),
				Entry("state scheduled; dose exists",
					func(datum *dataTypesStatusPump.BasalDelivery) {
						datum.State = pointer.FromString("scheduled")
						datum.Time = pointer.FromTime(test.RandomTime())
						datum.Dose = dataTypesStatusPumpTest.RandomBasalDose()
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/dose"),
				),
				Entry("state suspended",
					func(datum *dataTypesStatusPump.BasalDelivery) {
						datum.State = pointer.FromString("suspended")
						datum.Time = pointer.FromTime(test.RandomTime())
						datum.Dose = nil
					},
				),
				Entry("state suspended; time not exists",
					func(datum *dataTypesStatusPump.BasalDelivery) {
						datum.State = pointer.FromString("suspended")
						datum.Time = nil
						datum.Dose = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/time"),
				),
				Entry("state suspended; dose exists",
					func(datum *dataTypesStatusPump.BasalDelivery) {
						datum.State = pointer.FromString("suspended")
						datum.Time = pointer.FromTime(test.RandomTime())
						datum.Dose = dataTypesStatusPumpTest.RandomBasalDose()
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/dose"),
				),
				Entry("state suspending",
					func(datum *dataTypesStatusPump.BasalDelivery) {
						datum.State = pointer.FromString("suspending")
						datum.Time = nil
						datum.Dose = nil
					},
				),
				Entry("state suspending; time exists",
					func(datum *dataTypesStatusPump.BasalDelivery) {
						datum.State = pointer.FromString("suspending")
						datum.Time = pointer.FromTime(test.RandomTime())
						datum.Dose = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/time"),
				),
				Entry("state suspending; dose exists",
					func(datum *dataTypesStatusPump.BasalDelivery) {
						datum.State = pointer.FromString("suspending")
						datum.Time = nil
						datum.Dose = dataTypesStatusPumpTest.RandomBasalDose()
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/dose"),
				),
				Entry("state temporary",
					func(datum *dataTypesStatusPump.BasalDelivery) {
						datum.State = pointer.FromString("temporary")
						datum.Time = nil
						datum.Dose = dataTypesStatusPumpTest.RandomBasalDose()
					},
				),
				Entry("state temporary; time exists",
					func(datum *dataTypesStatusPump.BasalDelivery) {
						datum.State = pointer.FromString("temporary")
						datum.Time = pointer.FromTime(test.RandomTime())
						datum.Dose = dataTypesStatusPumpTest.RandomBasalDose()
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/time"),
				),
				Entry("state temporary; dose not exists",
					func(datum *dataTypesStatusPump.BasalDelivery) {
						datum.State = pointer.FromString("temporary")
						datum.Time = nil
						datum.Dose = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/dose"),
				),
				Entry("state temporary; dose invalid",
					func(datum *dataTypesStatusPump.BasalDelivery) {
						datum.State = pointer.FromString("temporary")
						datum.Time = nil
						datum.Dose = dataTypesStatusPumpTest.RandomBasalDose()
						datum.Dose.Rate = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/dose/rate"),
				),
				Entry("multiple errors",
					func(datum *dataTypesStatusPump.BasalDelivery) {
						datum.State = nil
						datum.Time = pointer.FromTime(test.RandomTime())
						datum.Dose = dataTypesStatusPumpTest.RandomBasalDose()
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/state"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/time"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/dose"),
				),
			)
		})
	})

	Context("BasalDose", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataTypesStatusPump.BasalDose)) {
				datum := dataTypesStatusPumpTest.RandomBasalDose()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataTypesStatusPumpTest.NewObjectFromBasalDose(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataTypesStatusPumpTest.NewObjectFromBasalDose(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataTypesStatusPump.BasalDose) {},
			),
			Entry("empty",
				func(datum *dataTypesStatusPump.BasalDose) {
					*datum = *dataTypesStatusPump.NewBasalDose()
				},
			),
			Entry("all",
				func(datum *dataTypesStatusPump.BasalDose) {
					datum.StartTime = pointer.FromTime(test.RandomTime())
					datum.EndTime = pointer.FromTime(test.RandomTimeAfter(*datum.StartTime))
					datum.Rate = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesStatusPump.BasalDoseRateMinimum, dataTypesStatusPump.BasalDoseRateMaximum))
					datum.AmountDelivered = pointer.FromFloat64(test.RandomFloat64FromRange(dataTypesStatusPump.BasalDoseAmountDeliveredMinimum, dataTypesStatusPump.BasalDoseAmountDeliveredMaximum))
				},
			),
		)

		Context("ParseBasalDose", func() {
			It("returns nil when the object is missing", func() {
				Expect(dataTypesStatusPump.ParseBasalDose(structureParser.NewObject(logTest.NewLogger(), nil))).To(BeNil())
			})

			It("returns new datum when the object is valid", func() {
				datum := dataTypesStatusPumpTest.RandomBasalDose()
				object := dataTypesStatusPumpTest.NewObjectFromBasalDose(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(logTest.NewLogger(), &object)
				Expect(dataTypesStatusPump.ParseBasalDose(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("NewBasalDose", func() {
			It("returns the expected datum with all values initialized", func() {
				datum := dataTypesStatusPump.NewBasalDose()
				Expect(datum).ToNot(BeNil())
				Expect(datum.StartTime).To(BeNil())
				Expect(datum.EndTime).To(BeNil())
				Expect(datum.Rate).To(BeNil())
				Expect(datum.AmountDelivered).To(BeNil())
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dataTypesStatusPump.BasalDose), expectedErrors ...error) {
					expectedDatum := dataTypesStatusPumpTest.RandomBasalDose()
					object := dataTypesStatusPumpTest.NewObjectFromBasalDose(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := dataTypesStatusPump.NewBasalDose()
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dataTypesStatusPump.BasalDose) {},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *dataTypesStatusPump.BasalDose) {
						object["startTime"] = true
						object["endTime"] = true
						object["rate"] = true
						object["amountDelivered"] = true
						expectedDatum.StartTime = nil
						expectedDatum.EndTime = nil
						expectedDatum.Rate = nil
						expectedDatum.AmountDelivered = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/startTime"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/endTime"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotFloat64(true), "/rate"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotFloat64(true), "/amountDelivered"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("return the expected results when the input",
				func(mutator func(datum *dataTypesStatusPump.BasalDose), expectedErrors ...error) {
					datum := dataTypesStatusPumpTest.RandomBasalDose()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesStatusPump.BasalDose) {},
				),
				Entry("start time missing; end time missing",
					func(datum *dataTypesStatusPump.BasalDose) {
						datum.StartTime = nil
						datum.EndTime = nil
					},
				),
				Entry("start time missing; end time exists",
					func(datum *dataTypesStatusPump.BasalDose) {
						datum.StartTime = nil
						datum.EndTime = pointer.FromTime(test.RandomTime())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/endTime"),
				),
				Entry("start time exists; end time missing",
					func(datum *dataTypesStatusPump.BasalDose) {
						datum.StartTime = pointer.FromTime(test.RandomTime())
						datum.EndTime = nil
					},
				),
				Entry("start time exists; end time exists before start time",
					func(datum *dataTypesStatusPump.BasalDose) {
						datum.StartTime = pointer.FromTime(test.PastNearTime())
						datum.EndTime = pointer.FromTime(test.PastFarTime())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotAfter(test.PastFarTime(), test.PastNearTime()), "/endTime"),
				),
				Entry("start time exists; end time exists after start time",
					func(datum *dataTypesStatusPump.BasalDose) {
						datum.StartTime = pointer.FromTime(test.PastFarTime())
						datum.EndTime = pointer.FromTime(test.PastNearTime())
					},
				),
				Entry("rate missing",
					func(datum *dataTypesStatusPump.BasalDose) { datum.Rate = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
				),
				Entry("rate out of range (lower)",
					func(datum *dataTypesStatusPump.BasalDose) { datum.Rate = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0, 100), "/rate"),
				),
				Entry("rate in range (lower)",
					func(datum *dataTypesStatusPump.BasalDose) { datum.Rate = pointer.FromFloat64(0) },
				),
				Entry("rate in range (upper)",
					func(datum *dataTypesStatusPump.BasalDose) { datum.Rate = pointer.FromFloat64(100) },
				),
				Entry("rate out of range (upper)",
					func(datum *dataTypesStatusPump.BasalDose) { datum.Rate = pointer.FromFloat64(100.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(100.1, 0, 100), "/rate"),
				),
				Entry("amount delivered missing",
					func(datum *dataTypesStatusPump.BasalDose) { datum.AmountDelivered = nil },
				),
				Entry("amount delivered out of range (lower)",
					func(datum *dataTypesStatusPump.BasalDose) { datum.AmountDelivered = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0, 1000), "/amountDelivered"),
				),
				Entry("amount delivered in range (lower)",
					func(datum *dataTypesStatusPump.BasalDose) { datum.AmountDelivered = pointer.FromFloat64(0) },
				),
				Entry("amount delivered in range (upper)",
					func(datum *dataTypesStatusPump.BasalDose) { datum.AmountDelivered = pointer.FromFloat64(1000) },
				),
				Entry("amount delivered out of range (upper)",
					func(datum *dataTypesStatusPump.BasalDose) { datum.AmountDelivered = pointer.FromFloat64(1000.1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(1000.1, 0, 1000), "/amountDelivered"),
				),
				Entry("multiple errors",
					func(datum *dataTypesStatusPump.BasalDose) {
						datum.StartTime = nil
						datum.EndTime = pointer.FromTime(test.RandomTime())
						datum.Rate = nil
						datum.AmountDelivered = pointer.FromFloat64(-0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/endTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rate"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(-0.1, 0, 1000), "/amountDelivered"),
				),
			)
		})
	})
})
