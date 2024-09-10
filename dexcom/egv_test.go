package dexcom_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/dexcom"
	dexcomTest "github.com/tidepool-org/platform/dexcom/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("EGV", func() {
	It("EGVsResponseRecordType is expected", func() {
		Expect(dexcom.EGVsResponseRecordType).To(Equal("egv"))
	})

	It("EGVsResponseRecordVersion is expected", func() {
		Expect(dexcom.EGVsResponseRecordVersion).To(Equal("3.0"))
	})

	It("EGVUnitUnknown is expected", func() {
		Expect(dexcom.EGVUnitUnknown).To(Equal("unknown"))
	})

	It("EGVUnitMgdL is expected", func() {
		Expect(dexcom.EGVUnitMgdL).To(Equal("mg/dL"))
	})

	It("EGVUnitMmolL is expected", func() {
		Expect(dexcom.EGVUnitMmolL).To(Equal("mmol/L"))
	})

	It("EGVRateUnitUnknown is expected", func() {
		Expect(dexcom.EGVRateUnitUnknown).To(Equal("unknown"))
	})

	It("EGVRateUnitMgdLMinute is expected", func() {
		Expect(dexcom.EGVRateUnitMgdLMinute).To(Equal("mg/dL/min"))
	})

	It("EGVRateUnitMmolLMinute is expected", func() {
		Expect(dexcom.EGVRateUnitMmolLMinute).To(Equal("mmol/L/min"))
	})

	It("EGVValueMgdLMaximum is expected", func() {
		Expect(dexcom.EGVValueMgdLMaximum).To(Equal(1000.0))
	})

	It("EGVValueMgdLMinimum is expected", func() {
		Expect(dexcom.EGVValueMgdLMinimum).To(Equal(0.0))
	})

	It("EGVValueMmolLMaximum is expected", func() {
		Expect(dexcom.EGVValueMmolLMaximum).To(Equal(55.0))
	})

	It("EGVValueMmolLMinimum is expected", func() {
		Expect(dexcom.EGVValueMmolLMinimum).To(Equal(0.0))
	})

	It("EGVStatusUnknown is expected", func() {
		Expect(dexcom.EGVStatusUnknown).To(Equal("unknown"))
	})

	It("EGVStatusHigh is expected", func() {
		Expect(dexcom.EGVStatusHigh).To(Equal("high"))
	})

	It("EGVStatusLow is expected", func() {
		Expect(dexcom.EGVStatusLow).To(Equal("low"))
	})

	It("EGVStatusOK is expected", func() {
		Expect(dexcom.EGVStatusOK).To(Equal("ok"))
	})

	It("EGVTrendNone is expected", func() {
		Expect(dexcom.EGVTrendNone).To(Equal("none"))
	})

	It("EGVTrendUnknown is expected", func() {
		Expect(dexcom.EGVTrendUnknown).To(Equal("unknown"))
	})

	It("EGVTrendDoubleUp is expected", func() {
		Expect(dexcom.EGVTrendDoubleUp).To(Equal("doubleUp"))
	})

	It("EGVTrendSingleUp is expected", func() {
		Expect(dexcom.EGVTrendSingleUp).To(Equal("singleUp"))
	})

	It("EGVTrendFortyFiveUp is expected", func() {
		Expect(dexcom.EGVTrendFortyFiveUp).To(Equal("fortyFiveUp"))
	})

	It("EGVTrendFlat is expected", func() {
		Expect(dexcom.EGVTrendFlat).To(Equal("flat"))
	})

	It("EGVTrendFortyFiveDown is expected", func() {
		Expect(dexcom.EGVTrendFortyFiveDown).To(Equal("fortyFiveDown"))
	})

	It("EGVTrendSingleDown is expected", func() {
		Expect(dexcom.EGVTrendSingleDown).To(Equal("singleDown"))
	})

	It("EGVTrendDoubleDown is expected", func() {
		Expect(dexcom.EGVTrendDoubleDown).To(Equal("doubleDown"))
	})

	It("EGVTrendNotComputable is expected", func() {
		Expect(dexcom.EGVTrendNotComputable).To(Equal("notComputable"))
	})

	It("EGVTrendRateOutOfRange is expected", func() {
		Expect(dexcom.EGVTrendRateOutOfRange).To(Equal("rateOutOfRange"))
	})

	It("EGVTransmitterTickMinimum is expected", func() {
		Expect(dexcom.EGVTransmitterTickMinimum).To(Equal(0))
	})

	It("EGVValuePinnedMgdLMaximum is expected", func() {
		Expect(dexcom.EGVValuePinnedMgdLMaximum).To(Equal(400.0))
	})

	It("EGVValuePinnedMgdLMinimum is expected", func() {
		Expect(dexcom.EGVValuePinnedMgdLMinimum).To(Equal(40.0))
	})

	It("EGVValuePinnedMmolLMaximum is expected", func() {
		Expect(dexcom.EGVValuePinnedMmolLMaximum).To(Equal(22.20299))
	})

	It("EGVValuePinnedMmolLMinimum is expected", func() {
		Expect(dexcom.EGVValuePinnedMmolLMinimum).To(Equal(2.22030))
	})

	It("EGVUnits returns expected", func() {
		Expect(dexcom.EGVUnits()).To(Equal([]string{"unknown", "mg/dL", "mmol/L"}))
	})

	It("EGVRateUnits returns expected", func() {
		Expect(dexcom.EGVRateUnits()).To(Equal([]string{"unknown", "mg/dL/min", "mmol/L/min"}))
	})

	It("EGVStatuses returns expected", func() {
		Expect(dexcom.EGVStatuses()).To(Equal([]string{"unknown", "high", "low", "ok"}))
	})

	It("EGVTrends returns expected", func() {
		Expect(dexcom.EGVTrends()).To(Equal([]string{"unknown", "none", "doubleUp", "singleUp", "fortyFiveUp", "flat", "fortyFiveDown", "singleDown", "doubleDown", "notComputable", "rateOutOfRange"}))
	})

	Context("ParseEGVsResponse", func() {
		It("returns nil if the object is nil", func() {
			parser := structureParser.NewObject(logTest.NewLogger(), nil)
			Expect(dexcom.ParseEGVsResponse(parser)).To(BeNil())
		})

		It("returns the parsed object", func() {
			expectedDatum := dexcomTest.RandomEGVsResponse()
			object := dexcomTest.NewObjectFromEGVsResponse(expectedDatum, test.ObjectFormatJSON)
			parser := structureParser.NewObject(logTest.NewLogger(), &object)
			Expect(dexcom.ParseEGVsResponse(parser)).To(Equal(expectedDatum))
		})
	})

	Context("EGVsResponse", func() {
		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dexcom.EGVsResponse), expectedErrors ...error) {
					expectedDatum := dexcomTest.RandomEGVsResponse()
					object := dexcomTest.NewObjectFromEGVsResponse(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					for index, expectedError := range expectedErrors {
						expectedErrors[index] = errors.WithMeta(expectedError, expectedDatum)
					}
					datum := &dexcom.EGVsResponse{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dexcom.EGVsResponse) {},
				),
				Entry("recordType invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.EGVsResponse) {
						object["recordType"] = true
						expectedDatum.RecordType = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/recordType"),
				),
				Entry("recordVersion invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.EGVsResponse) {
						object["recordVersion"] = true
						expectedDatum.RecordVersion = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/recordVersion"),
				),
				Entry("userId invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.EGVsResponse) {
						object["userId"] = true
						expectedDatum.UserID = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/userId"),
				),
				Entry("records invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.EGVsResponse) {
						object["records"] = true
						expectedDatum.Records = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/records"),
				),
				Entry("records element invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.EGVsResponse) {
						object["records"] = []interface{}{false}
						expectedDatum.Records = &dexcom.EGVs{nil}
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/records/0"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dexcom.EGVsResponse), expectedErrors ...error) {
					datum := dexcomTest.RandomEGVsResponse()
					mutator(datum)
					for index, expectedError := range expectedErrors {
						expectedErrors[index] = errors.WithMeta(expectedError, datum)
					}
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dexcom.EGVsResponse) {},
				),
				Entry("recordType missing",
					func(datum *dexcom.EGVsResponse) {
						datum.RecordType = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/recordType"),
				),
				Entry("recordType invalid",
					func(datum *dexcom.EGVsResponse) {
						datum.RecordType = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalid", dexcom.EGVsResponseRecordType), "/recordType"),
				),
				Entry("recordVersion missing",
					func(datum *dexcom.EGVsResponse) {
						datum.RecordVersion = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/recordVersion"),
				),
				Entry("recordVersion invalid",
					func(datum *dexcom.EGVsResponse) {
						datum.RecordVersion = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalid", dexcom.EGVsResponseRecordVersion), "/recordVersion"),
				),
				Entry("userId missing",
					func(datum *dexcom.EGVsResponse) {
						datum.UserID = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/userId"),
				),
				Entry("userId empty",
					func(datum *dexcom.EGVsResponse) {
						datum.UserID = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/userId"),
				),
				Entry("records missing",
					func(datum *dexcom.EGVsResponse) {
						datum.Records = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/records"),
				),
				Entry("records invalid does not report an error",
					func(datum *dexcom.EGVsResponse) {
						(*datum.Records)[0].RecordID = nil
					},
				),
				Entry("multiple errors",
					func(datum *dexcom.EGVsResponse) {
						datum.RecordType = nil
						datum.RecordVersion = nil
						datum.UserID = nil
						datum.Records = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/recordType"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/recordVersion"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/userId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/records"),
				),
			)
		})
	})

	Context("ParseEGV", func() {
		It("returns nil if the object is nil", func() {
			parser := structureParser.NewObject(logTest.NewLogger(), nil)
			Expect(dexcom.ParseEGV(parser)).To(BeNil())
		})

		It("returns the parsed object", func() {
			expectedDatum := dexcomTest.RandomEGV()
			object := dexcomTest.NewObjectFromEGV(expectedDatum, test.ObjectFormatJSON)
			parser := structureParser.NewObject(logTest.NewLogger(), &object)
			Expect(dexcom.ParseEGV(parser)).To(Equal(expectedDatum))
		})
	})

	Context("EGV", func() {
		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dexcom.EGV), expectedErrors ...error) {
					expectedDatum := dexcomTest.RandomEGV()
					object := dexcomTest.NewObjectFromEGV(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					for index, expectedError := range expectedErrors {
						expectedErrors[index] = errors.WithMeta(expectedError, expectedDatum)
					}
					datum := &dexcom.EGV{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dexcom.EGV) {},
				),
				Entry("recordId invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.EGV) {
						object["recordId"] = true
						expectedDatum.RecordID = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/recordId"),
				),
				Entry("systemTime invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.EGV) {
						object["systemTime"] = true
						expectedDatum.SystemTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/systemTime"),
				),
				Entry("systemTime invalid time",
					func(object map[string]interface{}, expectedDatum *dexcom.EGV) {
						object["systemTime"] = "invalid"
						expectedDatum.SystemTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorValueTimeNotParsable("invalid", time.RFC3339Nano), "/systemTime"),
				),
				Entry("displayTime invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.EGV) {
						object["displayTime"] = true
						expectedDatum.DisplayTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/displayTime"),
				),
				Entry("displayTime invalid time",
					func(object map[string]interface{}, expectedDatum *dexcom.EGV) {
						object["displayTime"] = "invalid"
						expectedDatum.DisplayTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorValueTimeNotParsable("invalid", time.RFC3339Nano), "/displayTime"),
				),
				Entry("unit invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.EGV) {
						object["unit"] = true
						expectedDatum.Unit = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/unit"),
				),
				Entry("value invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.EGV) {
						object["value"] = true
						expectedDatum.Value = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotFloat64(true), "/value"),
				),
				Entry("rateUnit invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.EGV) {
						object["rateUnit"] = true
						expectedDatum.RateUnit = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/rateUnit"),
				),
				Entry("trendRate invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.EGV) {
						object["trendRate"] = true
						expectedDatum.TrendRate = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotFloat64(true), "/trendRate"),
				),
				Entry("status invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.EGV) {
						object["status"] = true
						expectedDatum.Status = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/status"),
				),
				Entry("trend invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.EGV) {
						object["trend"] = true
						expectedDatum.Trend = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/trend"),
				),
				Entry("transmitterGeneration invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.EGV) {
						object["transmitterGeneration"] = true
						expectedDatum.TransmitterGeneration = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/transmitterGeneration"),
				),
				Entry("transmitterId invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.EGV) {
						object["transmitterId"] = true
						expectedDatum.TransmitterID = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/transmitterId"),
				),
				Entry("transmitterTicks invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.EGV) {
						object["transmitterTicks"] = true
						expectedDatum.TransmitterTicks = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotInt(true), "/transmitterTicks"),
				),
				Entry("displayDevice invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.EGV) {
						object["displayDevice"] = true
						expectedDatum.DisplayDevice = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/displayDevice"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dexcom.EGV), expectedErrors ...error) {
					datum := dexcomTest.RandomEGV()
					mutator(datum)
					for index, expectedError := range expectedErrors {
						expectedErrors[index] = errors.WithMeta(expectedError, datum)
					}
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dexcom.EGV) {},
				),
				Entry("recordId missing",
					func(datum *dexcom.EGV) {
						datum.RecordID = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/recordId"),
				),
				Entry("recordId empty",
					func(datum *dexcom.EGV) {
						datum.RecordID = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/recordId"),
				),
				Entry("systemTime missing",
					func(datum *dexcom.EGV) {
						datum.SystemTime = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/systemTime"),
				),
				Entry("systemTime zero",
					func(datum *dexcom.EGV) {
						datum.SystemTime.Time = time.Time{}
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/systemTime"),
				),
				Entry("displayTime missing",
					func(datum *dexcom.EGV) {
						datum.DisplayTime = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/displayTime"),
				),
				Entry("displayTime zero",
					func(datum *dexcom.EGV) {
						datum.DisplayTime.Time = time.Time{}
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/displayTime"),
				),
				Entry("unit missing",
					func(datum *dexcom.EGV) {
						datum.Unit = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/unit"),
				),
				Entry("unit invalid",
					func(datum *dexcom.EGV) {
						datum.Unit = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dexcom.EGVUnits()), "/unit"),
				),
				Entry("value missing",
					func(datum *dexcom.EGV) {
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("unit mg/dL; value out of range (lower)",
					func(datum *dexcom.EGV) {
						datum.Unit = pointer.FromString(dexcom.EGVUnitMgdL)
						datum.Value = pointer.FromFloat64(dexcom.EGVValueMgdLMinimum - 0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(dexcom.EGVValueMgdLMinimum-0.1, dexcom.EGVValueMgdLMinimum, dexcom.EGVValueMgdLMaximum), "/value"),
				),
				Entry("unit mg/dL; value in range (lower)",
					func(datum *dexcom.EGV) {
						datum.Unit = pointer.FromString(dexcom.EGVUnitMgdL)
						datum.Value = pointer.FromFloat64(dexcom.EGVValueMgdLMinimum)
					},
				),
				Entry("unit mg/dL; value in range (upper)",
					func(datum *dexcom.EGV) {
						datum.Unit = pointer.FromString(dexcom.EGVUnitMgdL)
						datum.Value = pointer.FromFloat64(dexcom.EGVValueMgdLMaximum)
					},
				),
				Entry("unit mg/dL; value out of range (upper)",
					func(datum *dexcom.EGV) {
						datum.Unit = pointer.FromString(dexcom.EGVUnitMgdL)
						datum.Value = pointer.FromFloat64(dexcom.EGVValueMgdLMaximum + 0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(dexcom.EGVValueMgdLMaximum+0.1, dexcom.EGVValueMgdLMinimum, dexcom.EGVValueMgdLMaximum), "/value"),
				),
				Entry("unit mmol/L; value out of range (lower)",
					func(datum *dexcom.EGV) {
						datum.Unit = pointer.FromString(dexcom.EGVUnitMmolL)
						datum.Value = pointer.FromFloat64(dexcom.EGVValueMmolLMinimum - 0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(dexcom.EGVValueMmolLMinimum-0.1, dexcom.EGVValueMmolLMinimum, dexcom.EGVValueMmolLMaximum), "/value"),
				),
				Entry("unit mmol/L; value in range (lower)",
					func(datum *dexcom.EGV) {
						datum.Unit = pointer.FromString(dexcom.EGVUnitMmolL)
						datum.Value = pointer.FromFloat64(dexcom.EGVValueMmolLMinimum)
					},
				),
				Entry("unit mmol/L; value in range (upper)",
					func(datum *dexcom.EGV) {
						datum.Unit = pointer.FromString(dexcom.EGVUnitMmolL)
						datum.Value = pointer.FromFloat64(dexcom.EGVValueMmolLMaximum)
					},
				),
				Entry("unit mmol/L; value out of range (upper)",
					func(datum *dexcom.EGV) {
						datum.Unit = pointer.FromString(dexcom.EGVUnitMmolL)
						datum.Value = pointer.FromFloat64(dexcom.EGVValueMmolLMaximum + 0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(dexcom.EGVValueMmolLMaximum+0.1, dexcom.EGVValueMmolLMinimum, dexcom.EGVValueMmolLMaximum), "/value"),
				),
				Entry("rateUnit missing",
					func(datum *dexcom.EGV) {
						datum.RateUnit = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rateUnit"),
				),
				Entry("rateUnit invalid",
					func(datum *dexcom.EGV) {
						datum.RateUnit = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dexcom.EGVRateUnits()), "/rateUnit"),
				),
				Entry("trendRate missing",
					func(datum *dexcom.EGV) {
						datum.TrendRate = nil
					},
				),
				Entry("status missing",
					func(datum *dexcom.EGV) {
						datum.Status = nil
					},
				),
				Entry("status invalid",
					func(datum *dexcom.EGV) {
						datum.Status = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dexcom.EGVStatuses()), "/status"),
				),
				Entry("trend missing",
					func(datum *dexcom.EGV) {
						datum.Trend = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/trend"),
				),
				Entry("trend invalid",
					func(datum *dexcom.EGV) {
						datum.Trend = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dexcom.EGVTrends()), "/trend"),
				),
				Entry("transmitterGeneration missing",
					func(datum *dexcom.EGV) {
						datum.TransmitterGeneration = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/transmitterGeneration"),
				),
				Entry("transmitterGeneration invalid",
					func(datum *dexcom.EGV) {
						datum.TransmitterGeneration = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dexcom.DeviceTransmitterGenerations()), "/transmitterGeneration"),
				),
				Entry("transmitterId missing",
					func(datum *dexcom.EGV) {
						datum.TransmitterID = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/transmitterId"),
				),
				Entry("transmitterId invalid",
					func(datum *dexcom.EGV) {
						datum.TransmitterID = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(dexcom.ErrorValueStringAsTransmitterIDNotValid("invalid"), "/transmitterId"),
				),
				Entry("transmitterId empty",
					func(datum *dexcom.EGV) {
						datum.TransmitterID = pointer.FromString("")
					},
				),
				Entry("transmitterTicks missing",
					func(datum *dexcom.EGV) {
						datum.TransmitterTicks = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/transmitterTicks"),
				),
				Entry("transmitterTicks invalid",
					func(datum *dexcom.EGV) {
						datum.TransmitterTicks = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, dexcom.EGVTransmitterTickMinimum), "/transmitterTicks"),
				),
				Entry("displayDevice missing",
					func(datum *dexcom.EGV) {
						datum.DisplayDevice = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/displayDevice"),
				),
				Entry("displayDevice invalid",
					func(datum *dexcom.EGV) {
						datum.DisplayDevice = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dexcom.DeviceDisplayDevices()), "/displayDevice"),
				),
				Entry("multiple errors",
					func(datum *dexcom.EGV) {
						datum.SystemTime = nil
						datum.DisplayTime = nil
						datum.Unit = nil
						datum.Value = nil
						datum.RateUnit = nil
						datum.Status = pointer.FromString("invalid")
						datum.Trend = nil
						datum.TransmitterGeneration = nil
						datum.TransmitterID = nil
						datum.TransmitterTicks = nil
						datum.DisplayDevice = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/systemTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/displayTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/unit"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rateUnit"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dexcom.EGVStatuses()), "/status"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/trend"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/transmitterGeneration"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/transmitterId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/transmitterTicks"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/displayDevice"),
				),
			)
		})
	})
})
