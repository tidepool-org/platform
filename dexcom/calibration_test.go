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

var _ = Describe("Calibration", func() {
	It("CalibrationsResponseRecordType is expected", func() {
		Expect(dexcom.CalibrationsResponseRecordType).To(Equal("calibration"))
	})

	It("CalibrationsResponseRecordVersion is expected", func() {
		Expect(dexcom.CalibrationsResponseRecordVersion).To(Equal("3.0"))
	})

	It("CalibrationUnitUnknown is expected", func() {
		Expect(dexcom.CalibrationUnitUnknown).To(Equal("unknown"))
	})

	It("CalibrationUnitMgdL is expected", func() {
		Expect(dexcom.CalibrationUnitMgdL).To(Equal("mg/dL"))
	})

	It("CalibrationUnitMmolL is expected", func() {
		Expect(dexcom.CalibrationUnitMmolL).To(Equal("mmol/L"))
	})

	It("CalibrationValueMgdLMaximum is expected", func() {
		Expect(dexcom.CalibrationValueMgdLMaximum).To(Equal(1000.0))
	})

	It("CalibrationValueMgdLMinimum is expected", func() {
		Expect(dexcom.CalibrationValueMgdLMinimum).To(Equal(0.0))
	})

	It("CalibrationValueMmolLMaximum is expected", func() {
		Expect(dexcom.CalibrationValueMmolLMaximum).To(Equal(55.0))
	})

	It("CalibrationValueMmolLMinimum is expected", func() {
		Expect(dexcom.CalibrationValueMmolLMinimum).To(Equal(0.0))
	})

	It("CalibrationUnits returns expected", func() {
		Expect(dexcom.CalibrationUnits()).To(Equal([]string{"unknown", "mg/dL", "mmol/L"}))
	})

	Context("ParseCalibrationsResponse", func() {
		It("returns nil if the object is nil", func() {
			parser := structureParser.NewObject(logTest.NewLogger(), nil)
			Expect(dexcom.ParseCalibrationsResponse(parser)).To(BeNil())
		})

		It("returns the parsed object", func() {
			expectedDatum := dexcomTest.RandomCalibrationsResponse()
			object := dexcomTest.NewObjectFromCalibrationsResponse(expectedDatum, test.ObjectFormatJSON)
			parser := structureParser.NewObject(logTest.NewLogger(), &object)
			Expect(dexcom.ParseCalibrationsResponse(parser)).To(Equal(expectedDatum))
		})
	})

	Context("CalibrationsResponse", func() {
		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dexcom.CalibrationsResponse), expectedErrors ...error) {
					expectedDatum := dexcomTest.RandomCalibrationsResponse()
					object := dexcomTest.NewObjectFromCalibrationsResponse(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					for index, expectedError := range expectedErrors {
						expectedErrors[index] = errors.WithMeta(expectedError, expectedDatum)
					}
					datum := &dexcom.CalibrationsResponse{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dexcom.CalibrationsResponse) {},
				),
				Entry("recordType invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.CalibrationsResponse) {
						object["recordType"] = true
						expectedDatum.RecordType = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/recordType"),
				),
				Entry("recordVersion invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.CalibrationsResponse) {
						object["recordVersion"] = true
						expectedDatum.RecordVersion = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/recordVersion"),
				),
				Entry("userId invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.CalibrationsResponse) {
						object["userId"] = true
						expectedDatum.UserID = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/userId"),
				),
				Entry("records invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.CalibrationsResponse) {
						object["records"] = true
						expectedDatum.Records = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/records"),
				),
				Entry("records element invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.CalibrationsResponse) {
						object["records"] = []interface{}{false}
						expectedDatum.Records = &dexcom.Calibrations{nil}
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/records/0"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dexcom.CalibrationsResponse), expectedErrors ...error) {
					datum := dexcomTest.RandomCalibrationsResponse()
					mutator(datum)
					for index, expectedError := range expectedErrors {
						expectedErrors[index] = errors.WithMeta(expectedError, datum)
					}
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dexcom.CalibrationsResponse) {},
				),
				Entry("recordType missing",
					func(datum *dexcom.CalibrationsResponse) {
						datum.RecordType = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/recordType"),
				),
				Entry("recordType invalid",
					func(datum *dexcom.CalibrationsResponse) {
						datum.RecordType = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalid", dexcom.CalibrationsResponseRecordType), "/recordType"),
				),
				Entry("recordVersion missing",
					func(datum *dexcom.CalibrationsResponse) {
						datum.RecordVersion = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/recordVersion"),
				),
				Entry("recordVersion invalid",
					func(datum *dexcom.CalibrationsResponse) {
						datum.RecordVersion = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalid", dexcom.CalibrationsResponseRecordVersion), "/recordVersion"),
				),
				Entry("userId missing",
					func(datum *dexcom.CalibrationsResponse) {
						datum.UserID = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/userId"),
				),
				Entry("userId empty",
					func(datum *dexcom.CalibrationsResponse) {
						datum.UserID = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/userId"),
				),
				Entry("records missing",
					func(datum *dexcom.CalibrationsResponse) {
						datum.Records = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/records"),
				),
				Entry("records invalid does not report an error",
					func(datum *dexcom.CalibrationsResponse) {
						(*datum.Records)[0].RecordID = nil
					},
				),
				Entry("multiple errors",
					func(datum *dexcom.CalibrationsResponse) {
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

	Context("ParseCalibration", func() {
		It("returns nil if the object is nil", func() {
			parser := structureParser.NewObject(logTest.NewLogger(), nil)
			Expect(dexcom.ParseCalibration(parser)).To(BeNil())
		})

		It("returns the parsed object", func() {
			expectedDatum := dexcomTest.RandomCalibration()
			object := dexcomTest.NewObjectFromCalibration(expectedDatum, test.ObjectFormatJSON)
			parser := structureParser.NewObject(logTest.NewLogger(), &object)
			Expect(dexcom.ParseCalibration(parser)).To(Equal(expectedDatum))
		})
	})

	Context("Calibration", func() {
		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dexcom.Calibration), expectedErrors ...error) {
					expectedDatum := dexcomTest.RandomCalibration()
					object := dexcomTest.NewObjectFromCalibration(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					for index, expectedError := range expectedErrors {
						expectedErrors[index] = errors.WithMeta(expectedError, expectedDatum)
					}
					datum := &dexcom.Calibration{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dexcom.Calibration) {},
				),
				Entry("recordId invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.Calibration) {
						object["recordId"] = true
						expectedDatum.RecordID = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/recordId"),
				),
				Entry("systemTime invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.Calibration) {
						object["systemTime"] = true
						expectedDatum.SystemTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/systemTime"),
				),
				Entry("systemTime invalid time",
					func(object map[string]interface{}, expectedDatum *dexcom.Calibration) {
						object["systemTime"] = "invalid"
						expectedDatum.SystemTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorValueTimeNotParsable("invalid", time.RFC3339Nano), "/systemTime"),
				),
				Entry("displayTime invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.Calibration) {
						object["displayTime"] = true
						expectedDatum.DisplayTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/displayTime"),
				),
				Entry("displayTime invalid time",
					func(object map[string]interface{}, expectedDatum *dexcom.Calibration) {
						object["displayTime"] = "invalid"
						expectedDatum.DisplayTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorValueTimeNotParsable("invalid", time.RFC3339Nano), "/displayTime"),
				),
				Entry("unit invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.Calibration) {
						object["unit"] = true
						expectedDatum.Unit = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/unit"),
				),
				Entry("value invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.Calibration) {
						object["value"] = true
						expectedDatum.Value = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotFloat64(true), "/value"),
				),
				Entry("transmitterGeneration invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.Calibration) {
						object["transmitterGeneration"] = true
						expectedDatum.TransmitterGeneration = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/transmitterGeneration"),
				),
				Entry("transmitterId invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.Calibration) {
						object["transmitterId"] = true
						expectedDatum.TransmitterID = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/transmitterId"),
				),
				Entry("transmitterTicks invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.Calibration) {
						object["transmitterTicks"] = true
						expectedDatum.TransmitterTicks = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotInt(true), "/transmitterTicks"),
				),
				Entry("displayDevice invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.Calibration) {
						object["displayDevice"] = true
						expectedDatum.DisplayDevice = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/displayDevice"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dexcom.Calibration), expectedErrors ...error) {
					datum := dexcomTest.RandomCalibration()
					mutator(datum)
					for index, expectedError := range expectedErrors {
						expectedErrors[index] = errors.WithMeta(expectedError, datum)
					}
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dexcom.Calibration) {},
				),
				Entry("recordId missing",
					func(datum *dexcom.Calibration) {
						datum.RecordID = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/recordId"),
				),
				Entry("recordId empty",
					func(datum *dexcom.Calibration) {
						datum.RecordID = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/recordId"),
				),
				Entry("systemTime missing",
					func(datum *dexcom.Calibration) {
						datum.SystemTime = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/systemTime"),
				),
				Entry("systemTime zero",
					func(datum *dexcom.Calibration) {
						datum.SystemTime.Time = time.Time{}
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/systemTime"),
				),
				Entry("displayTime missing",
					func(datum *dexcom.Calibration) {
						datum.DisplayTime = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/displayTime"),
				),
				Entry("displayTime zero",
					func(datum *dexcom.Calibration) {
						datum.DisplayTime.Time = time.Time{}
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/displayTime"),
				),
				Entry("unit missing",
					func(datum *dexcom.Calibration) {
						datum.Unit = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/unit"),
				),
				Entry("unit invalid",
					func(datum *dexcom.Calibration) {
						datum.Unit = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dexcom.CalibrationUnits()), "/unit"),
				),
				Entry("value missing",
					func(datum *dexcom.Calibration) {
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("unit mg/dL; value out of range (lower)",
					func(datum *dexcom.Calibration) {
						datum.Unit = pointer.FromString(dexcom.CalibrationUnitMgdL)
						datum.Value = pointer.FromFloat64(dexcom.CalibrationValueMgdLMinimum - 0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(dexcom.CalibrationValueMgdLMinimum-0.1, dexcom.CalibrationValueMgdLMinimum, dexcom.CalibrationValueMgdLMaximum), "/value"),
				),
				Entry("unit mg/dL; value in range (lower)",
					func(datum *dexcom.Calibration) {
						datum.Unit = pointer.FromString(dexcom.CalibrationUnitMgdL)
						datum.Value = pointer.FromFloat64(dexcom.CalibrationValueMgdLMinimum)
					},
				),
				Entry("unit mg/dL; value in range (upper)",
					func(datum *dexcom.Calibration) {
						datum.Unit = pointer.FromString(dexcom.CalibrationUnitMgdL)
						datum.Value = pointer.FromFloat64(dexcom.CalibrationValueMgdLMaximum)
					},
				),
				Entry("unit mg/dL; value out of range (upper)",
					func(datum *dexcom.Calibration) {
						datum.Unit = pointer.FromString(dexcom.CalibrationUnitMgdL)
						datum.Value = pointer.FromFloat64(dexcom.CalibrationValueMgdLMaximum + 0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(dexcom.CalibrationValueMgdLMaximum+0.1, dexcom.CalibrationValueMgdLMinimum, dexcom.CalibrationValueMgdLMaximum), "/value"),
				),
				Entry("unit mmol/L; value out of range (lower)",
					func(datum *dexcom.Calibration) {
						datum.Unit = pointer.FromString(dexcom.CalibrationUnitMmolL)
						datum.Value = pointer.FromFloat64(dexcom.CalibrationValueMmolLMinimum - 0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(dexcom.CalibrationValueMmolLMinimum-0.1, dexcom.CalibrationValueMmolLMinimum, dexcom.CalibrationValueMmolLMaximum), "/value"),
				),
				Entry("unit mmol/L; value in range (lower)",
					func(datum *dexcom.Calibration) {
						datum.Unit = pointer.FromString(dexcom.CalibrationUnitMmolL)
						datum.Value = pointer.FromFloat64(dexcom.CalibrationValueMmolLMinimum)
					},
				),
				Entry("unit mmol/L; value in range (upper)",
					func(datum *dexcom.Calibration) {
						datum.Unit = pointer.FromString(dexcom.CalibrationUnitMmolL)
						datum.Value = pointer.FromFloat64(dexcom.CalibrationValueMmolLMaximum)
					},
				),
				Entry("unit mmol/L; value out of range (upper)",
					func(datum *dexcom.Calibration) {
						datum.Unit = pointer.FromString(dexcom.CalibrationUnitMmolL)
						datum.Value = pointer.FromFloat64(dexcom.CalibrationValueMmolLMaximum + 0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(dexcom.CalibrationValueMmolLMaximum+0.1, dexcom.CalibrationValueMmolLMinimum, dexcom.CalibrationValueMmolLMaximum), "/value"),
				),
				Entry("transmitterGeneration missing",
					func(datum *dexcom.Calibration) {
						datum.TransmitterGeneration = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/transmitterGeneration"),
				),
				Entry("transmitterGeneration invalid",
					func(datum *dexcom.Calibration) {
						datum.TransmitterGeneration = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dexcom.DeviceTransmitterGenerations()), "/transmitterGeneration"),
				),
				Entry("transmitterId missing",
					func(datum *dexcom.Calibration) {
						datum.TransmitterID = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/transmitterId"),
				),
				Entry("transmitterId invalid",
					func(datum *dexcom.Calibration) {
						datum.TransmitterID = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(dexcom.ErrorValueStringAsTransmitterIDNotValid("invalid"), "/transmitterId"),
				),
				Entry("transmitterId empty",
					func(datum *dexcom.Calibration) {
						datum.TransmitterID = pointer.FromString("")
					},
				),
				Entry("transmitterTicks missing",
					func(datum *dexcom.Calibration) {
						datum.TransmitterTicks = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/transmitterTicks"),
				),
				Entry("transmitterTicks invalid",
					func(datum *dexcom.Calibration) {
						datum.TransmitterTicks = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, dexcom.EGVTransmitterTickMinimum), "/transmitterTicks"),
				),
				Entry("displayDevice missing",
					func(datum *dexcom.Calibration) {
						datum.DisplayDevice = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/displayDevice"),
				),
				Entry("displayDevice invalid",
					func(datum *dexcom.Calibration) {
						datum.DisplayDevice = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dexcom.DeviceDisplayDevices()), "/displayDevice"),
				),
				Entry("multiple errors",
					func(datum *dexcom.Calibration) {
						datum.SystemTime = nil
						datum.DisplayTime = nil
						datum.Unit = nil
						datum.Value = nil
						datum.TransmitterGeneration = nil
						datum.TransmitterID = nil
						datum.TransmitterTicks = nil
						datum.DisplayDevice = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/systemTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/displayTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/unit"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/transmitterGeneration"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/transmitterId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/transmitterTicks"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/displayDevice"),
				),
			)
		})
	})
})
