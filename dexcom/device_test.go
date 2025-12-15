package dexcom_test

import (
	"encoding/json"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/dexcom"
	dexcomTest "github.com/tidepool-org/platform/dexcom/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureNormalizer "github.com/tidepool-org/platform/structure/normalizer"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Device", func() {
	It("DevicesResponseRecordType returns expected", func() {
		Expect(dexcom.DevicesResponseRecordType).To(Equal("device"))
	})

	It("DevicesResponseRecordVersion returns expected", func() {
		Expect(dexcom.DevicesResponseRecordVersion).To(Equal("3.0"))
	})

	It("DeviceDisplayDeviceUnknown returns expected", func() {
		Expect(dexcom.DeviceDisplayDeviceUnknown).To(Equal("unknown"))
	})

	It("DeviceDisplayDeviceAndroid returns expected", func() {
		Expect(dexcom.DeviceDisplayDeviceAndroid).To(Equal("android"))
	})

	It("DeviceDisplayDeviceIOS returns expected", func() {
		Expect(dexcom.DeviceDisplayDeviceIOS).To(Equal("iOS"))
	})

	It("DeviceDisplayDeviceReceiver returns expected", func() {
		Expect(dexcom.DeviceDisplayDeviceReceiver).To(Equal("receiver"))
	})

	It("DeviceDisplayDeviceShareReceiver returns expected", func() {
		Expect(dexcom.DeviceDisplayDeviceShareReceiver).To(Equal("shareReceiver"))
	})

	It("DeviceDisplayDeviceTouchscreenReceiver returns expected", func() {
		Expect(dexcom.DeviceDisplayDeviceTouchscreenReceiver).To(Equal("touchscreenReceiver"))
	})

	It("DeviceDisplayAppUnknown returns expected", func() {
		Expect(dexcom.DeviceDisplayAppUnknown).To(Equal("unknown"))
	})

	It("DeviceDisplayAppG5 returns expected", func() {
		Expect(dexcom.DeviceDisplayAppG5).To(Equal("G5"))
	})

	It("DeviceDisplayAppG6 returns expected", func() {
		Expect(dexcom.DeviceDisplayAppG6).To(Equal("G6"))
	})

	It("DeviceDisplayAppG7 returns expected", func() {
		Expect(dexcom.DeviceDisplayAppG7).To(Equal("G7"))
	})

	It("DeviceDisplayAppReceiver returns expected", func() {
		Expect(dexcom.DeviceDisplayAppReceiver).To(Equal("receiver"))
	})

	It("DeviceDisplayAppWatch returns expected", func() {
		Expect(dexcom.DeviceDisplayAppWatch).To(Equal("Watch"))
	})

	It("DeviceTransmitterGenerationUnknown returns expected", func() {
		Expect(dexcom.DeviceTransmitterGenerationUnknown).To(Equal("unknown"))
	})

	It("DeviceTransmitterGenerationG4 returns expected", func() {
		Expect(dexcom.DeviceTransmitterGenerationG4).To(Equal("g4"))
	})

	It("DeviceTransmitterGenerationG5 returns expected", func() {
		Expect(dexcom.DeviceTransmitterGenerationG5).To(Equal("g5"))
	})

	It("DeviceTransmitterGenerationG6 returns expected", func() {
		Expect(dexcom.DeviceTransmitterGenerationG6).To(Equal("g6"))
	})

	It("DeviceTransmitterGenerationG6Pro returns expected", func() {
		Expect(dexcom.DeviceTransmitterGenerationG6Pro).To(Equal("g6 pro"))
	})

	It("DeviceTransmitterGenerationG6Plus returns expected", func() {
		Expect(dexcom.DeviceTransmitterGenerationG6Plus).To(Equal("g6+"))
	})

	It("DeviceTransmitterGenerationPro returns expected", func() {
		Expect(dexcom.DeviceTransmitterGenerationPro).To(Equal("dexcomPro"))
	})

	It("DeviceTransmitterGenerationG7 returns expected", func() {
		Expect(dexcom.DeviceTransmitterGenerationG7).To(Equal("g7"))
	})

	It("DeviceTransmitterGenerationG715Day returns expected", func() {
		Expect(dexcom.DeviceTransmitterGenerationG715Day).To(Equal("g715day"))
	})

	It("DeviceDisplayDevices returns expected", func() {
		Expect(dexcom.DeviceDisplayDevices()).To(Equal([]string{"unknown", "android", "iOS", "receiver", "shareReceiver", "touchscreenReceiver"}))
	})

	It("DeviceDisplayApps returns expected", func() {
		Expect(dexcom.DeviceDisplayApps()).To(Equal([]string{"unknown", "G5", "G6", "G7", "receiver", "Watch"}))
	})

	It("DeviceTransmitterGenerations returns expected", func() {
		Expect(dexcom.DeviceTransmitterGenerations()).To(Equal([]string{"unknown", "g4", "g5", "g6", "g6 pro", "g6+", "dexcomPro", "g7", "g715day"}))
	})

	Context("ParseDevicesResponse", func() {
		It("returns nil if the object is nil", func() {
			parser := structureParser.NewObject(logTest.NewLogger(), nil)
			Expect(dexcom.ParseDevicesResponse(parser)).To(BeNil())
		})

		It("returns the parsed object", func() {
			expectedDatum := dexcomTest.RandomDevicesResponse()
			object := dexcomTest.NewObjectFromDevicesResponse(expectedDatum, test.ObjectFormatJSON)
			parser := structureParser.NewObject(logTest.NewLogger(), &object)
			Expect(dexcom.ParseDevicesResponse(parser)).To(Equal(expectedDatum))
		})
	})

	Context("DevicesResponse", func() {
		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dexcom.DevicesResponse), expectedErrors ...error) {
					expectedDatum := dexcomTest.RandomDevicesResponse()
					object := dexcomTest.NewObjectFromDevicesResponse(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					for index, expectedError := range expectedErrors {
						expectedErrors[index] = errors.WithMeta(expectedError, expectedDatum)
					}
					datum := &dexcom.DevicesResponse{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dexcom.DevicesResponse) {},
				),
				Entry("recordType invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.DevicesResponse) {
						object["recordType"] = true
						expectedDatum.RecordType = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/recordType"),
				),
				Entry("recordVersion invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.DevicesResponse) {
						object["recordVersion"] = true
						expectedDatum.RecordVersion = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/recordVersion"),
				),
				Entry("userId invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.DevicesResponse) {
						object["userId"] = true
						expectedDatum.UserID = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/userId"),
				),
				Entry("records invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.DevicesResponse) {
						object["records"] = true
						expectedDatum.Records = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/records"),
				),
				Entry("records element invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.DevicesResponse) {
						object["records"] = []interface{}{false}
						expectedDatum.Records = &dexcom.Devices{nil}
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/records/0"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dexcom.DevicesResponse), expectedErrors ...error) {
					datum := dexcomTest.RandomDevicesResponse()
					mutator(datum)
					for index, expectedError := range expectedErrors {
						expectedErrors[index] = errors.WithMeta(expectedError, datum)
					}
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dexcom.DevicesResponse) {},
				),
				Entry("recordType missing",
					func(datum *dexcom.DevicesResponse) {
						datum.RecordType = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/recordType"),
				),
				Entry("recordType invalid",
					func(datum *dexcom.DevicesResponse) {
						datum.RecordType = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalid", dexcom.DevicesResponseRecordType), "/recordType"),
				),
				Entry("recordVersion missing",
					func(datum *dexcom.DevicesResponse) {
						datum.RecordVersion = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/recordVersion"),
				),
				Entry("recordVersion invalid",
					func(datum *dexcom.DevicesResponse) {
						datum.RecordVersion = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalid", dexcom.DevicesResponseRecordVersion), "/recordVersion"),
				),
				Entry("userId missing",
					func(datum *dexcom.DevicesResponse) {
						datum.UserID = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/userId"),
				),
				Entry("userId empty",
					func(datum *dexcom.DevicesResponse) {
						datum.UserID = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/userId"),
				),
				Entry("records missing",
					func(datum *dexcom.DevicesResponse) {
						datum.Records = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/records"),
				),
				Entry("records invalid does not report an error",
					func(datum *dexcom.DevicesResponse) {
						(*(*datum.Records)[0].AlertSchedules)[0].AlertScheduleSettings.IsDefaultSchedule = nil
					},
				),
				Entry("multiple errors",
					func(datum *dexcom.DevicesResponse) {
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

	Context("ParseDevice", func() {
		It("returns nil if the object is nil", func() {
			parser := structureParser.NewObject(logTest.NewLogger(), nil)
			Expect(dexcom.ParseDevice(parser)).To(BeNil())
		})

		It("returns the parsed object", func() {
			expectedDatum := dexcomTest.RandomDevice()
			object := dexcomTest.NewObjectFromDevice(expectedDatum, test.ObjectFormatJSON)
			parser := structureParser.NewObject(logTest.NewLogger(), &object)
			Expect(dexcom.ParseDevice(parser)).To(Equal(expectedDatum))
		})
	})

	Context("Device", func() {
		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dexcom.Device), expectedErrors ...error) {
					expectedDatum := dexcomTest.RandomDevice()
					object := dexcomTest.NewObjectFromDevice(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					for index, expectedError := range expectedErrors {
						expectedErrors[index] = errors.WithMeta(expectedError, expectedDatum)
					}
					datum := &dexcom.Device{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dexcom.Device) {},
				),
				Entry("lastUploadDate invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.Device) {
						object["lastUploadDate"] = true
						expectedDatum.LastUploadDate = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/lastUploadDate"),
				),
				Entry("lastUploadDate invalid time",
					func(object map[string]interface{}, expectedDatum *dexcom.Device) {
						object["lastUploadDate"] = "invalid"
						expectedDatum.LastUploadDate = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorValueTimeNotParsable("invalid", time.RFC3339Nano), "/lastUploadDate"),
				),
				Entry("alertSchedules invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.Device) {
						object["alertSchedules"] = true
						expectedDatum.AlertSchedules = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/alertSchedules"),
				),
				Entry("transmitterGeneration invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.Device) {
						object["transmitterGeneration"] = true
						expectedDatum.TransmitterGeneration = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/transmitterGeneration"),
				),
				Entry("transmitterId invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.Device) {
						object["transmitterId"] = true
						expectedDatum.TransmitterID = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/transmitterId"),
				),
				Entry("displayDevice invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.Device) {
						object["displayDevice"] = true
						expectedDatum.DisplayDevice = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/displayDevice"),
				),
				Entry("displayApp invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.Device) {
						object["displayApp"] = true
						expectedDatum.DisplayApp = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/displayApp"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dexcom.Device), expectedErrors ...error) {
					datum := dexcomTest.RandomDevice()
					mutator(datum)
					for index, expectedError := range expectedErrors {
						expectedErrors[index] = errors.WithMeta(expectedError, datum)
					}
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dexcom.Device) {},
				),
				Entry("lastUploadDate missing",
					func(datum *dexcom.Device) {
						datum.LastUploadDate = nil
					},
				),
				Entry("lastUploadDate zero",
					func(datum *dexcom.Device) {
						datum.LastUploadDate.Time = time.Time{}
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/lastUploadDate"),
				),
				Entry("lastUploadDate after now",
					func(datum *dexcom.Device) {
						datum.LastUploadDate.Time = time.Unix(3000000000, 0)
					},
				),
				Entry("alertSchedules missing",
					func(datum *dexcom.Device) {
						datum.AlertSchedules = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/alertSchedules"),
				),
				Entry("alertSchedules invalid",
					func(datum *dexcom.Device) {
						(*datum.AlertSchedules)[0].AlertScheduleSettings.IsEnabled = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/alertSchedules/0/alertScheduleSettings/isEnabled"),
				),
				Entry("transmitterGeneration missing",
					func(datum *dexcom.Device) {
						datum.TransmitterGeneration = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/transmitterGeneration"),
				),
				Entry("transmitterGeneration invalid",
					func(datum *dexcom.Device) {
						datum.TransmitterGeneration = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dexcom.DeviceTransmitterGenerations()), "/transmitterGeneration"),
				),
				Entry("transmitterId missing",
					func(datum *dexcom.Device) {
						datum.TransmitterID = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/transmitterId"),
				),
				Entry("transmitterId invalid",
					func(datum *dexcom.Device) {
						datum.TransmitterID = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(dexcom.ErrorValueStringAsTransmitterIDNotValid("invalid"), "/transmitterId"),
				),
				Entry("transmitterId empty",
					func(datum *dexcom.Device) {
						datum.TransmitterID = pointer.FromString("")
					},
				),
				Entry("displayDevice missing",
					func(datum *dexcom.Device) {
						datum.DisplayDevice = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/displayDevice"),
				),
				Entry("displayDevice invalid",
					func(datum *dexcom.Device) {
						datum.DisplayDevice = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dexcom.DeviceDisplayDevices()), "/displayDevice"),
				),
				Entry("displayApp missing",
					func(datum *dexcom.Device) {
						datum.DisplayApp = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/displayApp"),
				),
				Entry("displayApp invalid",
					func(datum *dexcom.Device) {
						datum.DisplayApp = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dexcom.DeviceDisplayApps()), "/displayApp"),
				),
				Entry("multiple errors",
					func(datum *dexcom.Device) {
						datum.LastUploadDate = dexcom.TimeFromRaw(time.Time{})
						datum.AlertSchedules = nil
						datum.TransmitterGeneration = nil
						datum.TransmitterID = nil
						datum.DisplayDevice = nil
						datum.DisplayApp = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/lastUploadDate"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/alertSchedules"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/transmitterGeneration"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/transmitterId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/displayDevice"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/displayApp"),
				),
			)
		})
	})

	Context("Normalize", func() {
		DescribeTable("normalizes the datum",
			func(mutator func(datum *dexcom.Device), expectator func(datum *dexcom.Device, expectedDatum *dexcom.Device)) {
				for _, origin := range structure.Origins() {
					datum := dexcomTest.RandomDevice()
					mutator(datum)
					expectedDatum := dexcomTest.CloneDevice(datum)
					normalizer := structureNormalizer.New(logTest.NewLogger())
					Expect(normalizer).ToNot(BeNil())
					datum.Normalize(normalizer.WithOrigin(origin))
					Expect(normalizer.Error()).To(BeNil())
					if expectator != nil {
						expectator(datum, expectedDatum)
					}
					Expect(datum).To(Equal(expectedDatum))
				}
			},
			Entry("does not modify the datum",
				func(datum *dexcom.Device) {
					datum.AlertSchedules = nil
				},
				nil,
			),
			Entry("normalizes the alertSchedules",
				func(datum *dexcom.Device) {},
				func(datum *dexcom.Device, expectedDatum *dexcom.Device) {
					expectedDatum.AlertSchedules.Normalize(structureNormalizer.New(logTest.NewLogger()))
				},
			),
		)
	})

	Context("ID", func() {
		It("uses the transmitterGeneration if the transmitterId is missing", func() {
			datum := dexcomTest.RandomDevice()
			datum.TransmitterID = nil
			Expect(datum.ID()).To(Equal(*datum.TransmitterGeneration))
		})

		It("uses the transmitterGeneration if the transmitterId is empty", func() {
			datum := dexcomTest.RandomDevice()
			datum.TransmitterID = pointer.FromString("")
			Expect(datum.ID()).To(Equal(*datum.TransmitterGeneration))
		})

		It("uses the transmitterGeneration if the transmitterId is valid", func() {
			datum := dexcomTest.RandomDevice()
			Expect(datum.ID()).To(Equal(*datum.TransmitterID))
		})
	})

	Context("Hash", func() {
		It("returns the hash", func() {
			serializedDatum := `
{
  "lastUploadDate": "2024-05-04T02:58:02Z",
  "alertSchedules": [
    {
      "alertScheduleSettings": {
        "isDefaultSchedule": true,
        "isEnabled": false,
        "isActive": true,
        "alertScheduleName": "",
        "startTime": "00:00",
        "endTime": "00:00",
        "daysOfWeek": [
          "sunday",
          "monday",
          "tuesday",
          "wednesday",
          "thursday",
          "friday",
          "saturday"
        ]
      },
      "alertSettings": [
        {
          "systemTime": "2017-12-14T20:08:32Z",
          "displayTime": "2020-06-03T21:18:50Z",
          "alertName": "urgentLowSoon",
          "unit": "mmol/L",
          "value": 3.571621888156606,
          "snooze": 315,
          "enabled": true,
          "secondaryTriggerCondition": 1340732657,
          "soundTheme": "classic",
          "soundOutputMode": "unknown"
        }
      ]
    }
  ],
  "transmitterGeneration": "dexcomPro",
  "transmitterId": "mn9w7dfqgsgm9xrlycb03exhd637lw4gg76hiihvvvqw2tad7m6pk7620css4e21",
  "displayDevice": "touchscreenReceiver",
  "displayApp": "G7"
}
`
			datum := &dexcom.Device{}
			Expect(json.Unmarshal([]byte(serializedDatum), datum)).ToNot(HaveOccurred())
			Expect(datum.Hash()).To(Equal("5c544ec70cdab9fc521003d594404373"))
		})
	})
})
