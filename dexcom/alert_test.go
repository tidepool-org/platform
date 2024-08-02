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

var _ = Describe("Alert", func() {
	It("AlertsResponseRecordType returns expected", func() {
		Expect(dexcom.AlertsResponseRecordType).To(Equal("alert"))
	})

	It("AlertsResponseRecordVersion returns expected", func() {
		Expect(dexcom.AlertsResponseRecordVersion).To(Equal("3.0"))
	})

	It("AlertNameUnknown returns expected", func() {
		Expect(dexcom.AlertNameUnknown).To(Equal("unknown"))
	})

	It("AlertNameHigh returns expected", func() {
		Expect(dexcom.AlertNameHigh).To(Equal("high"))
	})

	It("AlertNameLow returns expected", func() {
		Expect(dexcom.AlertNameLow).To(Equal("low"))
	})

	It("AlertNameRise returns expected", func() {
		Expect(dexcom.AlertNameRise).To(Equal("rise"))
	})

	It("AlertNameFall returns expected", func() {
		Expect(dexcom.AlertNameFall).To(Equal("fall"))
	})

	It("AlertNameOutOfRange returns expected", func() {
		Expect(dexcom.AlertNameOutOfRange).To(Equal("outOfRange"))
	})

	It("AlertNameUrgentLow returns expected", func() {
		Expect(dexcom.AlertNameUrgentLow).To(Equal("urgentLow"))
	})

	It("AlertNameUrgentLowSoon returns expected", func() {
		Expect(dexcom.AlertNameUrgentLowSoon).To(Equal("urgentLowSoon"))
	})

	It("AlertNameNoReadings returns expected", func() {
		Expect(dexcom.AlertNameNoReadings).To(Equal("noReadings"))
	})

	It("AlertNameFixedLow returns expected", func() {
		Expect(dexcom.AlertNameFixedLow).To(Equal("fixedLow"))
	})

	It("AlertStateUnknown returns expected", func() {
		Expect(dexcom.AlertStateUnknown).To(Equal("unknown"))
	})

	It("AlertStateInactive returns expected", func() {
		Expect(dexcom.AlertStateInactive).To(Equal("inactive"))
	})

	It("AlertStateActiveSnoozed returns expected", func() {
		Expect(dexcom.AlertStateActiveSnoozed).To(Equal("activeSnoozed"))
	})

	It("AlertStateActiveAlarming returns expected", func() {
		Expect(dexcom.AlertStateActiveAlarming).To(Equal("activeAlarming"))
	})

	It("AlertNames returns expected", func() {
		Expect(dexcom.AlertNames()).To(Equal([]string{"unknown", "high", "low", "rise", "fall", "outOfRange", "urgentLow", "urgentLowSoon", "noReadings", "fixedLow"}))
	})

	It("AlertStates returns expected", func() {
		Expect(dexcom.AlertStates()).To(Equal([]string{"unknown", "inactive", "activeSnoozed", "activeAlarming"}))
	})

	Context("ParseAlertsResponse", func() {
		It("returns nil if the object is nil", func() {
			parser := structureParser.NewObject(logTest.NewLogger(), nil)
			Expect(dexcom.ParseAlertsResponse(parser)).To(BeNil())
		})

		It("returns the parsed object", func() {
			expectedDatum := dexcomTest.RandomAlertsResponse()
			object := dexcomTest.NewObjectFromAlertsResponse(expectedDatum, test.ObjectFormatJSON)
			parser := structureParser.NewObject(logTest.NewLogger(), &object)
			Expect(dexcom.ParseAlertsResponse(parser)).To(Equal(expectedDatum))
		})
	})

	Context("AlertsResponse", func() {
		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dexcom.AlertsResponse), expectedErrors ...error) {
					expectedDatum := dexcomTest.RandomAlertsResponse()
					object := dexcomTest.NewObjectFromAlertsResponse(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					for index, expectedError := range expectedErrors {
						expectedErrors[index] = errors.WithMeta(expectedError, expectedDatum)
					}
					datum := &dexcom.AlertsResponse{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dexcom.AlertsResponse) {},
				),
				Entry("recordType invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.AlertsResponse) {
						object["recordType"] = true
						expectedDatum.RecordType = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/recordType"),
				),
				Entry("recordVersion invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.AlertsResponse) {
						object["recordVersion"] = true
						expectedDatum.RecordVersion = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/recordVersion"),
				),
				Entry("userId invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.AlertsResponse) {
						object["userId"] = true
						expectedDatum.UserID = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/userId"),
				),
				Entry("records invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.AlertsResponse) {
						object["records"] = true
						expectedDatum.Records = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/records"),
				),
				Entry("records element invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.AlertsResponse) {
						object["records"] = []interface{}{false}
						expectedDatum.Records = &dexcom.Alerts{nil}
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/records/0"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dexcom.AlertsResponse), expectedErrors ...error) {
					datum := dexcomTest.RandomAlertsResponse()
					mutator(datum)
					for index, expectedError := range expectedErrors {
						expectedErrors[index] = errors.WithMeta(expectedError, datum)
					}
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dexcom.AlertsResponse) {},
				),
				Entry("recordType missing",
					func(datum *dexcom.AlertsResponse) {
						datum.RecordType = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/recordType"),
				),
				Entry("recordType invalid",
					func(datum *dexcom.AlertsResponse) {
						datum.RecordType = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalid", dexcom.AlertsResponseRecordType), "/recordType"),
				),
				Entry("recordVersion missing",
					func(datum *dexcom.AlertsResponse) {
						datum.RecordVersion = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/recordVersion"),
				),
				Entry("recordVersion invalid",
					func(datum *dexcom.AlertsResponse) {
						datum.RecordVersion = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalid", dexcom.AlertsResponseRecordVersion), "/recordVersion"),
				),
				Entry("userId missing",
					func(datum *dexcom.AlertsResponse) {
						datum.UserID = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/userId"),
				),
				Entry("userId empty",
					func(datum *dexcom.AlertsResponse) {
						datum.UserID = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/userId"),
				),
				Entry("records missing",
					func(datum *dexcom.AlertsResponse) {
						datum.Records = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/records"),
				),
				Entry("records invalid does not report an error",
					func(datum *dexcom.AlertsResponse) {
						(*datum.Records)[0].RecordID = nil
					},
				),
				Entry("multiple errors",
					func(datum *dexcom.AlertsResponse) {
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

	Context("ParseAlert", func() {
		It("returns nil if the object is nil", func() {
			parser := structureParser.NewObject(logTest.NewLogger(), nil)
			Expect(dexcom.ParseAlert(parser)).To(BeNil())
		})

		It("returns the parsed object", func() {
			expectedDatum := dexcomTest.RandomAlert()
			object := dexcomTest.NewObjectFromAlert(expectedDatum, test.ObjectFormatJSON)
			parser := structureParser.NewObject(logTest.NewLogger(), &object)
			Expect(dexcom.ParseAlert(parser)).To(Equal(expectedDatum))
		})
	})

	Context("Alert", func() {
		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dexcom.Alert), expectedErrors ...error) {
					expectedDatum := dexcomTest.RandomAlert()
					object := dexcomTest.NewObjectFromAlert(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					for index, expectedError := range expectedErrors {
						expectedErrors[index] = errors.WithMeta(expectedError, expectedDatum)
					}
					datum := &dexcom.Alert{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dexcom.Alert) {},
				),
				Entry("recordId invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.Alert) {
						object["recordId"] = true
						expectedDatum.RecordID = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/recordId"),
				),
				Entry("systemTime invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.Alert) {
						object["systemTime"] = true
						expectedDatum.SystemTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/systemTime"),
				),
				Entry("systemTime invalid time",
					func(object map[string]interface{}, expectedDatum *dexcom.Alert) {
						object["systemTime"] = "invalid"
						expectedDatum.SystemTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorValueTimeNotParsable("invalid", time.RFC3339Nano), "/systemTime"),
				),
				Entry("displayTime invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.Alert) {
						object["displayTime"] = true
						expectedDatum.DisplayTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/displayTime"),
				),
				Entry("displayTime invalid time",
					func(object map[string]interface{}, expectedDatum *dexcom.Alert) {
						object["displayTime"] = "invalid"
						expectedDatum.DisplayTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorValueTimeNotParsable("invalid", time.RFC3339Nano), "/displayTime"),
				),
				Entry("alertName invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.Alert) {
						object["alertName"] = true
						expectedDatum.AlertName = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/alertName"),
				),
				Entry("alertState invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.Alert) {
						object["alertState"] = true
						expectedDatum.AlertState = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/alertState"),
				),
				Entry("transmitterGeneration invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.Alert) {
						object["transmitterGeneration"] = true
						expectedDatum.TransmitterGeneration = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/transmitterGeneration"),
				),
				Entry("transmitterId invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.Alert) {
						object["transmitterId"] = true
						expectedDatum.TransmitterID = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/transmitterId"),
				),
				Entry("displayDevice invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.Alert) {
						object["displayDevice"] = true
						expectedDatum.DisplayDevice = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/displayDevice"),
				),
				Entry("displayApp invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.Alert) {
						object["displayApp"] = true
						expectedDatum.DisplayApp = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/displayApp"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dexcom.Alert), expectedErrors ...error) {
					datum := dexcomTest.RandomAlert()
					mutator(datum)
					for index, expectedError := range expectedErrors {
						expectedErrors[index] = errors.WithMeta(expectedError, datum)
					}
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dexcom.Alert) {},
				),
				Entry("recordId missing",
					func(datum *dexcom.Alert) {
						datum.RecordID = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/recordId"),
				),
				Entry("recordId empty",
					func(datum *dexcom.Alert) {
						datum.RecordID = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/recordId"),
				),
				Entry("systemTime missing",
					func(datum *dexcom.Alert) {
						datum.SystemTime = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/systemTime"),
				),
				Entry("systemTime zero",
					func(datum *dexcom.Alert) {
						datum.SystemTime.Time = time.Time{}
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/systemTime"),
				),
				Entry("displayTime missing",
					func(datum *dexcom.Alert) {
						datum.DisplayTime = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/displayTime"),
				),
				Entry("displayTime zero",
					func(datum *dexcom.Alert) {
						datum.DisplayTime.Time = time.Time{}
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/displayTime"),
				),
				Entry("alertName missing",
					func(datum *dexcom.Alert) {
						datum.AlertName = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/alertName"),
				),
				Entry("alertName invalid",
					func(datum *dexcom.Alert) {
						datum.AlertName = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dexcom.AlertNames()), "/alertName"),
				),
				Entry("alertState missing",
					func(datum *dexcom.Alert) {
						datum.AlertState = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/alertState"),
				),
				Entry("alertState invalid",
					func(datum *dexcom.Alert) {
						datum.AlertState = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dexcom.AlertStates()), "/alertState"),
				),
				Entry("transmitterGeneration missing",
					func(datum *dexcom.Alert) {
						datum.TransmitterGeneration = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/transmitterGeneration"),
				),
				Entry("transmitterGeneration invalid",
					func(datum *dexcom.Alert) {
						datum.TransmitterGeneration = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dexcom.DeviceTransmitterGenerations()), "/transmitterGeneration"),
				),
				Entry("transmitterId missing",
					func(datum *dexcom.Alert) {
						datum.TransmitterID = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/transmitterId"),
				),
				Entry("transmitterId invalid",
					func(datum *dexcom.Alert) {
						datum.TransmitterID = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(dexcom.ErrorValueStringAsTransmitterIDNotValid("invalid"), "/transmitterId"),
				),
				Entry("transmitterId empty",
					func(datum *dexcom.Alert) {
						datum.TransmitterID = pointer.FromString("")
					},
				),
				Entry("displayDevice missing",
					func(datum *dexcom.Alert) {
						datum.DisplayDevice = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/displayDevice"),
				),
				Entry("displayDevice invalid",
					func(datum *dexcom.Alert) {
						datum.DisplayDevice = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dexcom.DeviceDisplayDevices()), "/displayDevice"),
				),
				Entry("displayApp missing",
					func(datum *dexcom.Alert) {
						datum.DisplayApp = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/displayApp"),
				),
				Entry("displayApp invalid",
					func(datum *dexcom.Alert) {
						datum.DisplayApp = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dexcom.DeviceDisplayApps()), "/displayApp"),
				),
				Entry("multiple errors",
					func(datum *dexcom.Alert) {
						datum.SystemTime = nil
						datum.DisplayTime = nil
						datum.AlertName = nil
						datum.AlertState = nil
						datum.TransmitterGeneration = nil
						datum.TransmitterID = nil
						datum.DisplayDevice = nil
						datum.DisplayApp = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/systemTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/displayTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/alertName"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/alertState"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/transmitterGeneration"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/transmitterId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/displayDevice"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/displayApp"),
				),
			)
		})
	})
})
