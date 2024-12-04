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

var _ = Describe("Event", func() {
	It("EventsResponseRecordType is expected", func() {
		Expect(dexcom.EventsResponseRecordType).To(Equal("event"))
	})

	It("EventsResponseRecordVersion is expected", func() {
		Expect(dexcom.EventsResponseRecordVersion).To(Equal("3.0"))
	})

	It("EventStatusCreated is expected", func() {
		Expect(dexcom.EventStatusCreated).To(Equal("created"))
	})

	It("EventStatusUpdated is expected", func() {
		Expect(dexcom.EventStatusUpdated).To(Equal("updated"))
	})

	It("EventStatusDeleted is expected", func() {
		Expect(dexcom.EventStatusDeleted).To(Equal("deleted"))
	})

	It("EventTypeUnknown is expected", func() {
		Expect(dexcom.EventTypeUnknown).To(Equal("unknown"))
	})

	It("EventTypeInsulin is expected", func() {
		Expect(dexcom.EventTypeInsulin).To(Equal("insulin"))
	})

	It("EventTypeCarbs is expected", func() {
		Expect(dexcom.EventTypeCarbs).To(Equal("carbs"))
	})

	It("EventTypeExercise is expected", func() {
		Expect(dexcom.EventTypeExercise).To(Equal("exercise"))
	})

	It("EventTypeHealth is expected", func() {
		Expect(dexcom.EventTypeHealth).To(Equal("health"))
	})

	It("EventTypeBloodGlucose is expected", func() {
		Expect(dexcom.EventTypeBloodGlucose).To(Equal("bloodGlucose"))
	})

	It("EventTypeNotes is expected", func() {
		Expect(dexcom.EventTypeNotes).To(Equal("notes"))
	})

	It("EventTypeNotesAlternate is expected", func() {
		Expect(dexcom.EventTypeNotesAlternate).To(Equal("note"))
	})

	It("EventSubTypeInsulinFastActing is expected", func() {
		Expect(dexcom.EventSubTypeInsulinFastActing).To(Equal("fastActing"))
	})

	It("EventSubTypeInsulinLongActing is expected", func() {
		Expect(dexcom.EventSubTypeInsulinLongActing).To(Equal("longActing"))
	})

	It("EventUnitInsulinUnits is expected", func() {
		Expect(dexcom.EventUnitInsulinUnits).To(Equal("units"))
	})

	It("EventValueInsulinUnitsMaximum is expected", func() {
		Expect(dexcom.EventValueInsulinUnitsMaximum).To(Equal(250.0))
	})

	It("EventValueInsulinUnitsMinimum is expected", func() {
		Expect(dexcom.EventValueInsulinUnitsMinimum).To(Equal(0.0))
	})

	It("EventValueInsulinUnitsDefault is expected", func() {
		Expect(dexcom.EventValueInsulinUnitsDefault).To(Equal("0"))
	})

	It("EventUnitCarbsGrams is expected", func() {
		Expect(dexcom.EventUnitCarbsGrams).To(Equal("grams"))
	})

	It("EventValueCarbsGramsMaximum is expected", func() {
		Expect(dexcom.EventValueCarbsGramsMaximum).To(Equal(1000.0))
	})

	It("EventValueCarbsGramsMinimum is expected", func() {
		Expect(dexcom.EventValueCarbsGramsMinimum).To(Equal(0.0))
	})

	It("EventValueCarbsGramsDefault is expected", func() {
		Expect(dexcom.EventValueCarbsGramsDefault).To(Equal("0"))
	})

	It("EventSubTypeExerciseLight is expected", func() {
		Expect(dexcom.EventSubTypeExerciseLight).To(Equal("light"))
	})

	It("EventSubTypeExerciseMedium is expected", func() {
		Expect(dexcom.EventSubTypeExerciseMedium).To(Equal("medium"))
	})

	It("EventSubTypeExerciseHeavy is expected", func() {
		Expect(dexcom.EventSubTypeExerciseHeavy).To(Equal("heavy"))
	})

	It("EventUnitExerciseMinutes is expected", func() {
		Expect(dexcom.EventUnitExerciseMinutes).To(Equal("minutes"))
	})

	It("EventValueExerciseMinutesMaximum is expected", func() {
		Expect(dexcom.EventValueExerciseMinutesMaximum).To(Equal(10080.0))
	})

	It("EventValueExerciseMinutesMinimum is expected", func() {
		Expect(dexcom.EventValueExerciseMinutesMinimum).To(Equal(0.0))
	})

	It("EventValueExerciseMinutesDefault is expected", func() {
		Expect(dexcom.EventValueExerciseMinutesDefault).To(Equal("0"))
	})

	It("EventSubTypeHealthIllness is expected", func() {
		Expect(dexcom.EventSubTypeHealthIllness).To(Equal("illness"))
	})

	It("EventSubTypeHealthStress is expected", func() {
		Expect(dexcom.EventSubTypeHealthStress).To(Equal("stress"))
	})

	It("EventSubTypeHealthHighSymptoms is expected", func() {
		Expect(dexcom.EventSubTypeHealthHighSymptoms).To(Equal("highSymptoms"))
	})

	It("EventSubTypeHealthLowSymptoms is expected", func() {
		Expect(dexcom.EventSubTypeHealthLowSymptoms).To(Equal("lowSymptoms"))
	})

	It("EventSubTypeHealthCycle is expected", func() {
		Expect(dexcom.EventSubTypeHealthCycle).To(Equal("cycle"))
	})

	It("EventSubTypeHealthAlcohol is expected", func() {
		Expect(dexcom.EventSubTypeHealthAlcohol).To(Equal("alcohol"))
	})

	It("EventUnitBloodGlucoseMgdL is expected", func() {
		Expect(dexcom.EventUnitBloodGlucoseMgdL).To(Equal("mg/dL"))
	})

	It("EventUnitBloodGlucoseMmolL is expected", func() {
		Expect(dexcom.EventUnitBloodGlucoseMmolL).To(Equal("mmol/L"))
	})

	It("EventValueBloodGlucoseMgdLMaximum is expected", func() {
		Expect(dexcom.EventValueBloodGlucoseMgdLMaximum).To(Equal(1000.0))
	})

	It("EventValueBloodGlucoseMgdLMinimum is expected", func() {
		Expect(dexcom.EventValueBloodGlucoseMgdLMinimum).To(Equal(0.0))
	})

	It("EventValueBloodGlucoseMmolLMaximum is expected", func() {
		Expect(dexcom.EventValueBloodGlucoseMmolLMaximum).To(Equal(55.0))
	})

	It("EventValueBloodGlucoseMmolLMinimum is expected", func() {
		Expect(dexcom.EventValueBloodGlucoseMmolLMinimum).To(Equal(0.0))
	})

	It("EventStatuses returns expected", func() {
		Expect(dexcom.EventStatuses()).To(Equal([]string{"created", "updated", "deleted"}))
	})

	It("EventTypes returns expected", func() {
		Expect(dexcom.EventTypes()).To(Equal([]string{"unknown", "insulin", "carbs", "exercise", "health", "bloodGlucose", "notes"}))
	})

	It("EventSubTypesInsulin returns expected", func() {
		Expect(dexcom.EventSubTypesInsulin()).To(Equal([]string{"fastActing", "longActing"}))
	})

	It("EventUnitsInsulin returns expected", func() {
		Expect(dexcom.EventUnitsInsulin()).To(Equal([]string{"units"}))
	})

	It("EventUnitsCarbs returns expected", func() {
		Expect(dexcom.EventUnitsCarbs()).To(Equal([]string{"grams"}))
	})

	It("EventSubTypesExercise returns expected", func() {
		Expect(dexcom.EventSubTypesExercise()).To(Equal([]string{"light", "medium", "heavy"}))
	})

	It("EventUnitsExercise returns expected", func() {
		Expect(dexcom.EventUnitsExercise()).To(Equal([]string{"minutes"}))
	})

	It("EventSubTypesHealth returns expected", func() {
		Expect(dexcom.EventSubTypesHealth()).To(Equal([]string{"illness", "stress", "highSymptoms", "lowSymptoms", "cycle", "alcohol"}))
	})

	It("EventUnitsBloodGlucose returns expected", func() {
		Expect(dexcom.EventUnitsBloodGlucose()).To(Equal([]string{"mg/dL", "mmol/L"}))
	})

	Context("ParseEventsResponse", func() {
		It("returns nil if the object is nil", func() {
			parser := structureParser.NewObject(logTest.NewLogger(), nil)
			Expect(dexcom.ParseEventsResponse(parser)).To(BeNil())
		})

		It("returns the parsed object", func() {
			expectedDatum := dexcomTest.RandomEventsResponse()
			object := dexcomTest.NewObjectFromEventsResponse(expectedDatum, test.ObjectFormatJSON)
			parser := structureParser.NewObject(logTest.NewLogger(), &object)
			Expect(dexcom.ParseEventsResponse(parser)).To(Equal(expectedDatum))
		})
	})

	Context("EventsResponse", func() {
		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dexcom.EventsResponse), expectedErrors ...error) {
					expectedDatum := dexcomTest.RandomEventsResponse()
					object := dexcomTest.NewObjectFromEventsResponse(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					for index, expectedError := range expectedErrors {
						expectedErrors[index] = errors.WithMeta(expectedError, expectedDatum)
					}
					datum := &dexcom.EventsResponse{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dexcom.EventsResponse) {},
				),
				Entry("recordType invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.EventsResponse) {
						object["recordType"] = true
						expectedDatum.RecordType = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/recordType"),
				),
				Entry("recordVersion invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.EventsResponse) {
						object["recordVersion"] = true
						expectedDatum.RecordVersion = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/recordVersion"),
				),
				Entry("userId invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.EventsResponse) {
						object["userId"] = true
						expectedDatum.UserID = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/userId"),
				),
				Entry("records invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.EventsResponse) {
						object["records"] = true
						expectedDatum.Records = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/records"),
				),
				Entry("records element invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.EventsResponse) {
						object["records"] = []interface{}{false}
						expectedDatum.Records = &dexcom.Events{nil}
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/records/0"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dexcom.EventsResponse), expectedErrors ...error) {
					datum := dexcomTest.RandomEventsResponse()
					mutator(datum)
					for index, expectedError := range expectedErrors {
						expectedErrors[index] = errors.WithMeta(expectedError, datum)
					}
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dexcom.EventsResponse) {},
				),
				Entry("recordType missing",
					func(datum *dexcom.EventsResponse) {
						datum.RecordType = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/recordType"),
				),
				Entry("recordType invalid",
					func(datum *dexcom.EventsResponse) {
						datum.RecordType = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalid", dexcom.EventsResponseRecordType), "/recordType"),
				),
				Entry("recordVersion missing",
					func(datum *dexcom.EventsResponse) {
						datum.RecordVersion = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/recordVersion"),
				),
				Entry("recordVersion invalid",
					func(datum *dexcom.EventsResponse) {
						datum.RecordVersion = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalid", dexcom.EventsResponseRecordVersion), "/recordVersion"),
				),
				Entry("userId missing",
					func(datum *dexcom.EventsResponse) {
						datum.UserID = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/userId"),
				),
				Entry("userId empty",
					func(datum *dexcom.EventsResponse) {
						datum.UserID = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/userId"),
				),
				Entry("records missing",
					func(datum *dexcom.EventsResponse) {
						datum.Records = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/records"),
				),
				Entry("records invalid does not report an error",
					func(datum *dexcom.EventsResponse) {
						(*datum.Records)[0].RecordID = nil
					},
				),
				Entry("multiple errors",
					func(datum *dexcom.EventsResponse) {
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

	Context("ParseEvent", func() {
		It("returns nil if the object is nil", func() {
			parser := structureParser.NewObject(logTest.NewLogger(), nil)
			Expect(dexcom.ParseEvent(parser)).To(BeNil())
		})

		It("returns the parsed object", func() {
			expectedDatum := dexcomTest.RandomEvent()
			object := dexcomTest.NewObjectFromEvent(expectedDatum, test.ObjectFormatJSON)
			parser := structureParser.NewObject(logTest.NewLogger(), &object)
			Expect(dexcom.ParseEvent(parser)).To(Equal(expectedDatum))
		})
	})

	Context("Event", func() {
		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dexcom.Event), expectedErrors ...error) {
					expectedDatum := dexcomTest.RandomEvent()
					object := dexcomTest.NewObjectFromEvent(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					for index, expectedError := range expectedErrors {
						expectedErrors[index] = errors.WithMeta(expectedError, expectedDatum)
					}
					datum := &dexcom.Event{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dexcom.Event) {},
				),
				Entry("recordId invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.Event) {
						object["recordId"] = true
						expectedDatum.RecordID = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/recordId"),
				),
				Entry("systemTime invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.Event) {
						object["systemTime"] = true
						expectedDatum.SystemTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/systemTime"),
				),
				Entry("systemTime invalid time",
					func(object map[string]interface{}, expectedDatum *dexcom.Event) {
						object["systemTime"] = "invalid"
						expectedDatum.SystemTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorValueTimeNotParsable("invalid", time.RFC3339Nano), "/systemTime"),
				),
				Entry("displayTime invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.Event) {
						object["displayTime"] = true
						expectedDatum.DisplayTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/displayTime"),
				),
				Entry("displayTime invalid time",
					func(object map[string]interface{}, expectedDatum *dexcom.Event) {
						object["displayTime"] = "invalid"
						expectedDatum.DisplayTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorValueTimeNotParsable("invalid", time.RFC3339Nano), "/displayTime"),
				),
				Entry("eventStatus invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.Event) {
						object["eventStatus"] = true
						expectedDatum.EventStatus = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/eventStatus"),
				),
				Entry("eventType invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.Event) {
						object["eventType"] = true
						expectedDatum.EventType = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/eventType"),
				),
				Entry("eventSubType invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.Event) {
						object["eventSubType"] = true
						expectedDatum.EventSubType = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/eventSubType"),
				),
				Entry("unit invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.Event) {
						object["unit"] = true
						expectedDatum.Unit = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/unit"),
				),
				Entry("value invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.Event) {
						object["value"] = true
						expectedDatum.Value = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/value"),
				),
				Entry("transmitterGeneration invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.Event) {
						object["transmitterGeneration"] = true
						expectedDatum.TransmitterGeneration = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/transmitterGeneration"),
				),
				Entry("transmitterId invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.Event) {
						object["transmitterId"] = true
						expectedDatum.TransmitterID = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/transmitterId"),
				),
				Entry("displayDevice invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.Event) {
						object["displayDevice"] = true
						expectedDatum.DisplayDevice = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/displayDevice"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the common datum",
				func(mutator func(datum *dexcom.Event), expectedErrors ...error) {
					datum := dexcomTest.RandomEvent()
					mutator(datum)
					for index, expectedError := range expectedErrors {
						expectedErrors[index] = errors.WithMeta(expectedError, datum)
					}
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dexcom.Event) {},
				),
				Entry("recordId missing",
					func(datum *dexcom.Event) {
						datum.RecordID = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/recordId"),
				),
				Entry("recordId empty",
					func(datum *dexcom.Event) {
						datum.RecordID = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/recordId"),
				),
				Entry("systemTime missing",
					func(datum *dexcom.Event) {
						datum.SystemTime = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/systemTime"),
				),
				Entry("systemTime zero",
					func(datum *dexcom.Event) {
						datum.SystemTime.Time = time.Time{}
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/systemTime"),
				),
				Entry("displayTime missing",
					func(datum *dexcom.Event) {
						datum.DisplayTime = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/displayTime"),
				),
				Entry("displayTime zero",
					func(datum *dexcom.Event) {
						datum.DisplayTime.Time = time.Time{}
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/displayTime"),
				),
				Entry("eventStatus missing",
					func(datum *dexcom.Event) {
						datum.EventStatus = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/eventStatus"),
				),
				Entry("eventStatus invalid",
					func(datum *dexcom.Event) {
						datum.EventStatus = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dexcom.EventStatuses()), "/eventStatus"),
				),
				Entry("eventType missing",
					func(datum *dexcom.Event) {
						datum.EventType = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/eventType"),
				),
				Entry("eventType invalid",
					func(datum *dexcom.Event) {
						datum.EventType = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dexcom.EventTypes()), "/eventType"),
				),
				Entry("eventType note",
					func(datum *dexcom.Event) {
						datum.EventType = pointer.FromString("note")
						datum.EventSubType = nil
						datum.Unit = nil
						datum.Value = pointer.FromString(test.RandomString())
					},
				),
				Entry("transmitterGeneration missing",
					func(datum *dexcom.Event) {
						datum.TransmitterGeneration = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/transmitterGeneration"),
				),
				Entry("transmitterGeneration invalid",
					func(datum *dexcom.Event) {
						datum.TransmitterGeneration = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dexcom.DeviceTransmitterGenerations()), "/transmitterGeneration"),
				),
				Entry("transmitterId missing",
					func(datum *dexcom.Event) {
						datum.TransmitterID = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/transmitterId"),
				),
				Entry("transmitterId invalid",
					func(datum *dexcom.Event) {
						datum.TransmitterID = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(dexcom.ErrorValueStringAsTransmitterIDNotValid("invalid"), "/transmitterId"),
				),
				Entry("transmitterId empty",
					func(datum *dexcom.Event) {
						datum.TransmitterID = pointer.FromString("")
					},
				),
				Entry("displayDevice missing",
					func(datum *dexcom.Event) {
						datum.DisplayDevice = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/displayDevice"),
				),
				Entry("displayDevice invalid",
					func(datum *dexcom.Event) {
						datum.DisplayDevice = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dexcom.DeviceDisplayDevices()), "/displayDevice"),
				),
				Entry("multiple errors",
					func(datum *dexcom.Event) {
						datum.SystemTime = nil
						datum.DisplayTime = nil
						datum.TransmitterGeneration = nil
						datum.TransmitterID = nil
						datum.DisplayDevice = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/systemTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/displayTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/transmitterGeneration"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/transmitterId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/displayDevice"),
				),
			)

			DescribeTable("validates the datum with eventType of unknown",
				func(mutator func(datum *dexcom.Event), expectedErrors ...error) {
					datum := dexcomTest.RandomEventWithType("unknown")
					mutator(datum)
					for index, expectedError := range expectedErrors {
						expectedErrors[index] = errors.WithMeta(expectedError, datum)
					}
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dexcom.Event) {},
				),
				Entry("eventSubType missing",
					func(datum *dexcom.Event) {
						datum.EventSubType = nil
					},
				),
				Entry("eventSubType present",
					func(datum *dexcom.Event) {
						datum.EventSubType = pointer.FromString(test.RandomString())
					},
				),
				Entry("unit missing",
					func(datum *dexcom.Event) {
						datum.Unit = nil
					},
				),
				Entry("unit present",
					func(datum *dexcom.Event) {
						datum.Unit = pointer.FromString(test.RandomString())
					},
				),
				Entry("value missing",
					func(datum *dexcom.Event) {
						datum.Value = nil
					},
				),
				Entry("value present",
					func(datum *dexcom.Event) {
						datum.Value = pointer.FromString(test.RandomString())
					},
				),
			)

			DescribeTable("validates the datum with eventType of insulin",
				func(mutator func(datum *dexcom.Event), expectedErrors ...error) {
					datum := dexcomTest.RandomEventWithType("insulin")
					mutator(datum)
					for index, expectedError := range expectedErrors {
						expectedErrors[index] = errors.WithMeta(expectedError, datum)
					}
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dexcom.Event) {},
				),
				Entry("eventSubType missing",
					func(datum *dexcom.Event) {
						datum.EventSubType = nil
					},
				),
				Entry("eventSubType invalid",
					func(datum *dexcom.Event) {
						datum.EventSubType = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dexcom.EventSubTypesInsulin()), "/eventSubType"),
				),
				Entry("unit missing",
					func(datum *dexcom.Event) {
						datum.Unit = nil
					},
				),
				Entry("unit empty",
					func(datum *dexcom.Event) {
						datum.Unit = pointer.FromString("")
					},
				),
				Entry("unit invalid",
					func(datum *dexcom.Event) {
						datum.Unit = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dexcom.EventUnitsInsulin()), "/unit"),
				),
				Entry("value missing",
					func(datum *dexcom.Event) {
						datum.Value = nil
					},
				),
				Entry("value empty",
					func(datum *dexcom.Event) {
						datum.Value = pointer.FromString("")
					},
				),
				Entry("value invalid",
					func(datum *dexcom.Event) {
						datum.Value = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(dexcom.ErrorValueFloat64NotParsable("invalid"), "/value"),
				),
				Entry("unit units; value out of range (lower)",
					func(datum *dexcom.Event) {
						datum.Unit = pointer.FromString(dexcom.EventUnitInsulinUnits)
						datum.Value = dexcomTest.StringPointerFromFloat64(dexcom.EventValueInsulinUnitsMinimum - 0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(dexcom.EventValueInsulinUnitsMinimum-0.1, dexcom.EventValueInsulinUnitsMinimum, dexcom.EventValueInsulinUnitsMaximum), "/value"),
				),
				Entry("unit units; value in range (lower)",
					func(datum *dexcom.Event) {
						datum.Unit = pointer.FromString(dexcom.EventUnitInsulinUnits)
						datum.Value = dexcomTest.StringPointerFromFloat64(dexcom.EventValueInsulinUnitsMinimum)
					},
				),
				Entry("unit units; value in range (upper)",
					func(datum *dexcom.Event) {
						datum.Unit = pointer.FromString(dexcom.EventUnitInsulinUnits)
						datum.Value = dexcomTest.StringPointerFromFloat64(dexcom.EventValueInsulinUnitsMaximum)
					},
				),
				Entry("unit units; value out of range (upper)",
					func(datum *dexcom.Event) {
						datum.Unit = pointer.FromString(dexcom.EventUnitInsulinUnits)
						datum.Value = dexcomTest.StringPointerFromFloat64(dexcom.EventValueInsulinUnitsMaximum + 0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(dexcom.EventValueInsulinUnitsMaximum+0.1, dexcom.EventValueInsulinUnitsMinimum, dexcom.EventValueInsulinUnitsMaximum), "/value"),
				),
				Entry("multiple errors",
					func(datum *dexcom.Event) {
						datum.EventSubType = pointer.FromString("invalid")
						datum.Unit = pointer.FromString("invalid")
						datum.Value = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dexcom.EventSubTypesInsulin()), "/eventSubType"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dexcom.EventUnitsInsulin()), "/unit"),
					errorsTest.WithPointerSource(dexcom.ErrorValueFloat64NotParsable("invalid"), "/value"),
				),
			)

			DescribeTable("validates the datum with eventType of carbs",
				func(mutator func(datum *dexcom.Event), expectedErrors ...error) {
					datum := dexcomTest.RandomEventWithType("carbs")
					mutator(datum)
					for index, expectedError := range expectedErrors {
						expectedErrors[index] = errors.WithMeta(expectedError, datum)
					}
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dexcom.Event) {},
				),
				Entry("eventSubType present",
					func(datum *dexcom.Event) {
						datum.EventSubType = pointer.FromString(test.RandomString())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/eventSubType"),
				),
				Entry("unit missing",
					func(datum *dexcom.Event) {
						datum.Unit = nil
					},
				),
				Entry("unit empty",
					func(datum *dexcom.Event) {
						datum.Unit = pointer.FromString("")
					},
				),
				Entry("unit invalid",
					func(datum *dexcom.Event) {
						datum.Unit = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dexcom.EventUnitsCarbs()), "/unit"),
				),
				Entry("value missing",
					func(datum *dexcom.Event) {
						datum.Value = nil
					},
				),
				Entry("value empty",
					func(datum *dexcom.Event) {
						datum.Value = pointer.FromString("")
					},
				),
				Entry("value invalid",
					func(datum *dexcom.Event) {
						datum.Value = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(dexcom.ErrorValueFloat64NotParsable("invalid"), "/value"),
				),
				Entry("unit grams; value out of range (lower)",
					func(datum *dexcom.Event) {
						datum.Unit = pointer.FromString(dexcom.EventUnitCarbsGrams)
						datum.Value = dexcomTest.StringPointerFromFloat64(dexcom.EventValueCarbsGramsMinimum - 0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(dexcom.EventValueCarbsGramsMinimum-0.1, dexcom.EventValueCarbsGramsMinimum, dexcom.EventValueCarbsGramsMaximum), "/value"),
				),
				Entry("unit grams; value in range (lower)",
					func(datum *dexcom.Event) {
						datum.Unit = pointer.FromString(dexcom.EventUnitCarbsGrams)
						datum.Value = dexcomTest.StringPointerFromFloat64(dexcom.EventValueCarbsGramsMinimum)
					},
				),
				Entry("unit grams; value in range (upper)",
					func(datum *dexcom.Event) {
						datum.Unit = pointer.FromString(dexcom.EventUnitCarbsGrams)
						datum.Value = dexcomTest.StringPointerFromFloat64(dexcom.EventValueCarbsGramsMaximum)
					},
				),
				Entry("unit grams; value out of range (upper)",
					func(datum *dexcom.Event) {
						datum.Unit = pointer.FromString(dexcom.EventUnitCarbsGrams)
						datum.Value = dexcomTest.StringPointerFromFloat64(dexcom.EventValueCarbsGramsMaximum + 0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(dexcom.EventValueCarbsGramsMaximum+0.1, dexcom.EventValueCarbsGramsMinimum, dexcom.EventValueCarbsGramsMaximum), "/value"),
				),
				Entry("multiple errors",
					func(datum *dexcom.Event) {
						datum.EventSubType = pointer.FromString(test.RandomString())
						datum.Unit = pointer.FromString("invalid")
						datum.Value = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/eventSubType"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dexcom.EventUnitsCarbs()), "/unit"),
					errorsTest.WithPointerSource(dexcom.ErrorValueFloat64NotParsable("invalid"), "/value"),
				),
			)

			DescribeTable("validates the datum with eventType of exercise",
				func(mutator func(datum *dexcom.Event), expectedErrors ...error) {
					datum := dexcomTest.RandomEventWithType("exercise")
					mutator(datum)
					for index, expectedError := range expectedErrors {
						expectedErrors[index] = errors.WithMeta(expectedError, datum)
					}
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dexcom.Event) {},
				),
				Entry("eventSubType missing",
					func(datum *dexcom.Event) {
						datum.EventSubType = nil
					},
				),
				Entry("eventSubType invalid",
					func(datum *dexcom.Event) {
						datum.EventSubType = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dexcom.EventSubTypesExercise()), "/eventSubType"),
				),
				Entry("unit missing",
					func(datum *dexcom.Event) {
						datum.Unit = nil
					},
				),
				Entry("unit empty",
					func(datum *dexcom.Event) {
						datum.Unit = pointer.FromString("")
					},
				),
				Entry("unit invalid",
					func(datum *dexcom.Event) {
						datum.Unit = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dexcom.EventUnitsExercise()), "/unit"),
				),
				Entry("value missing",
					func(datum *dexcom.Event) {
						datum.Value = nil
					},
				),
				Entry("value empty",
					func(datum *dexcom.Event) {
						datum.Value = pointer.FromString("")
					},
				),
				Entry("value -1",
					func(datum *dexcom.Event) {
						datum.Value = pointer.FromString("-1")
					},
				),
				Entry("value -1.00",
					func(datum *dexcom.Event) {
						datum.Value = pointer.FromString("-1.00")
					},
				),
				Entry("value invalid",
					func(datum *dexcom.Event) {
						datum.Value = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(dexcom.ErrorValueFloat64NotParsable("invalid"), "/value"),
				),
				Entry("unit minutes; value out of range (lower)",
					func(datum *dexcom.Event) {
						datum.Unit = pointer.FromString(dexcom.EventUnitExerciseMinutes)
						datum.Value = dexcomTest.StringPointerFromFloat64(dexcom.EventValueExerciseMinutesMinimum - 0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(dexcom.EventValueExerciseMinutesMinimum-0.1, dexcom.EventValueExerciseMinutesMinimum, dexcom.EventValueExerciseMinutesMaximum), "/value"),
				),
				Entry("unit minutes; value in range (lower)",
					func(datum *dexcom.Event) {
						datum.Unit = pointer.FromString(dexcom.EventUnitExerciseMinutes)
						datum.Value = dexcomTest.StringPointerFromFloat64(dexcom.EventValueExerciseMinutesMinimum)
					},
				),
				Entry("unit minutes; value in range (upper)",
					func(datum *dexcom.Event) {
						datum.Unit = pointer.FromString(dexcom.EventUnitExerciseMinutes)
						datum.Value = dexcomTest.StringPointerFromFloat64(dexcom.EventValueExerciseMinutesMaximum)
					},
				),
				Entry("unit minutes; value out of range (upper)",
					func(datum *dexcom.Event) {
						datum.Unit = pointer.FromString(dexcom.EventUnitExerciseMinutes)
						datum.Value = dexcomTest.StringPointerFromFloat64(dexcom.EventValueExerciseMinutesMaximum + 0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(dexcom.EventValueExerciseMinutesMaximum+0.1, dexcom.EventValueExerciseMinutesMinimum, dexcom.EventValueExerciseMinutesMaximum), "/value"),
				),
				Entry("multiple errors",
					func(datum *dexcom.Event) {
						datum.EventSubType = pointer.FromString("invalid")
						datum.Unit = pointer.FromString("invalid")
						datum.Value = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dexcom.EventSubTypesExercise()), "/eventSubType"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dexcom.EventUnitsExercise()), "/unit"),
					errorsTest.WithPointerSource(dexcom.ErrorValueFloat64NotParsable("invalid"), "/value"),
				),
			)

			DescribeTable("validates the datum with eventType of health",
				func(mutator func(datum *dexcom.Event), expectedErrors ...error) {
					datum := dexcomTest.RandomEventWithType("health")
					mutator(datum)
					for index, expectedError := range expectedErrors {
						expectedErrors[index] = errors.WithMeta(expectedError, datum)
					}
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dexcom.Event) {},
				),
				Entry("eventSubType missing",
					func(datum *dexcom.Event) {
						datum.EventSubType = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/eventSubType"),
				),
				Entry("eventSubType invalid",
					func(datum *dexcom.Event) {
						datum.EventSubType = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dexcom.EventSubTypesHealth()), "/eventSubType"),
				),
				Entry("unit present",
					func(datum *dexcom.Event) {
						datum.Unit = pointer.FromString(test.RandomString())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/unit"),
				),
				Entry("value present",
					func(datum *dexcom.Event) {
						datum.Value = pointer.FromString(test.RandomString())
					},
				),
				Entry("multiple errors",
					func(datum *dexcom.Event) {
						datum.EventSubType = nil
						datum.Unit = pointer.FromString(test.RandomString())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/eventSubType"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/unit"),
				),
			)

			DescribeTable("validates the datum with eventType of bloodGlucose",
				func(mutator func(datum *dexcom.Event), expectedErrors ...error) {
					datum := dexcomTest.RandomEventWithType("bloodGlucose")
					mutator(datum)
					for index, expectedError := range expectedErrors {
						expectedErrors[index] = errors.WithMeta(expectedError, datum)
					}
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dexcom.Event) {},
				),
				Entry("eventSubType present",
					func(datum *dexcom.Event) {
						datum.EventSubType = pointer.FromString(test.RandomString())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/eventSubType"),
				),
				Entry("unit missing",
					func(datum *dexcom.Event) {
						datum.Unit = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/unit"),
				),
				Entry("unit invalid",
					func(datum *dexcom.Event) {
						datum.Unit = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dexcom.EventUnitsBloodGlucose()), "/unit"),
				),
				Entry("value missing",
					func(datum *dexcom.Event) {
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("value invalid",
					func(datum *dexcom.Event) {
						datum.Value = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(dexcom.ErrorValueFloat64NotParsable("invalid"), "/value"),
				),
				Entry("unit mg/dL; value out of range (lower)",
					func(datum *dexcom.Event) {
						datum.Unit = pointer.FromString(dexcom.EventUnitBloodGlucoseMgdL)
						datum.Value = dexcomTest.StringPointerFromFloat64(dexcom.EventValueBloodGlucoseMgdLMinimum - 0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(dexcom.EventValueBloodGlucoseMgdLMinimum-0.1, dexcom.EventValueBloodGlucoseMgdLMinimum, dexcom.EventValueBloodGlucoseMgdLMaximum), "/value"),
				),
				Entry("unit mg/dL; value in range (lower)",
					func(datum *dexcom.Event) {
						datum.Unit = pointer.FromString(dexcom.EventUnitBloodGlucoseMgdL)
						datum.Value = dexcomTest.StringPointerFromFloat64(dexcom.EventValueBloodGlucoseMgdLMinimum)
					},
				),
				Entry("unit mg/dL; value in range (upper)",
					func(datum *dexcom.Event) {
						datum.Unit = pointer.FromString(dexcom.EventUnitBloodGlucoseMgdL)
						datum.Value = dexcomTest.StringPointerFromFloat64(dexcom.EventValueBloodGlucoseMgdLMaximum)
					},
				),
				Entry("unit mg/dL; value out of range (upper)",
					func(datum *dexcom.Event) {
						datum.Unit = pointer.FromString(dexcom.EventUnitBloodGlucoseMgdL)
						datum.Value = dexcomTest.StringPointerFromFloat64(dexcom.EventValueBloodGlucoseMgdLMaximum + 0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(dexcom.EventValueBloodGlucoseMgdLMaximum+0.1, dexcom.EventValueBloodGlucoseMgdLMinimum, dexcom.EventValueBloodGlucoseMgdLMaximum), "/value"),
				),
				Entry("unit mmol/L; value out of range (lower)",
					func(datum *dexcom.Event) {
						datum.Unit = pointer.FromString(dexcom.EventUnitBloodGlucoseMmolL)
						datum.Value = dexcomTest.StringPointerFromFloat64(dexcom.EventValueBloodGlucoseMmolLMinimum - 0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(dexcom.EventValueBloodGlucoseMmolLMinimum-0.1, dexcom.EventValueBloodGlucoseMmolLMinimum, dexcom.EventValueBloodGlucoseMmolLMaximum), "/value"),
				),
				Entry("unit mmol/L; value in range (lower)",
					func(datum *dexcom.Event) {
						datum.Unit = pointer.FromString(dexcom.EventUnitBloodGlucoseMmolL)
						datum.Value = dexcomTest.StringPointerFromFloat64(dexcom.EventValueBloodGlucoseMmolLMinimum)
					},
				),
				Entry("unit mmol/L; value in range (upper)",
					func(datum *dexcom.Event) {
						datum.Unit = pointer.FromString(dexcom.EventUnitBloodGlucoseMmolL)
						datum.Value = dexcomTest.StringPointerFromFloat64(dexcom.EventValueBloodGlucoseMmolLMaximum)
					},
				),
				Entry("unit mmol/L; value out of range (upper)",
					func(datum *dexcom.Event) {
						datum.Unit = pointer.FromString(dexcom.EventUnitBloodGlucoseMmolL)
						datum.Value = dexcomTest.StringPointerFromFloat64(dexcom.EventValueBloodGlucoseMmolLMaximum + 0.1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(dexcom.EventValueBloodGlucoseMmolLMaximum+0.1, dexcom.EventValueBloodGlucoseMmolLMinimum, dexcom.EventValueBloodGlucoseMmolLMaximum), "/value"),
				),
				Entry("multiple errors",
					func(datum *dexcom.Event) {
						datum.EventSubType = pointer.FromString(test.RandomString())
						datum.Unit = pointer.FromString("invalid")
						datum.Value = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/eventSubType"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dexcom.EventUnitsBloodGlucose()), "/unit"),
					errorsTest.WithPointerSource(dexcom.ErrorValueFloat64NotParsable("invalid"), "/value"),
				),
			)

			DescribeTable("validates the datum with eventType of notes",
				func(mutator func(datum *dexcom.Event), expectedErrors ...error) {
					datum := dexcomTest.RandomEventWithType("notes")
					mutator(datum)
					for index, expectedError := range expectedErrors {
						expectedErrors[index] = errors.WithMeta(expectedError, datum)
					}
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dexcom.Event) {},
				),
				Entry("eventSubType present",
					func(datum *dexcom.Event) {
						datum.EventSubType = pointer.FromString(test.RandomString())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/eventSubType"),
				),
				Entry("unit present",
					func(datum *dexcom.Event) {
						datum.Unit = pointer.FromString(test.RandomString())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/unit"),
				),
				Entry("value missing",
					func(datum *dexcom.Event) {
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
				Entry("value empty",
					func(datum *dexcom.Event) {
						datum.Value = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/value"),
				),
				Entry("multiple errors",
					func(datum *dexcom.Event) {
						datum.EventSubType = pointer.FromString(test.RandomString())
						datum.Unit = pointer.FromString(test.RandomString())
						datum.Value = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/eventSubType"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/unit"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
				),
			)
		})

		DescribeTable("validates the datum with eventType of note (alternate form of notes)",
			func(mutator func(datum *dexcom.Event), expectedErrors ...error) {
				datum := dexcomTest.RandomEventWithType("notes")
				datum.EventType = pointer.FromString("note")
				mutator(datum)
				for index, expectedError := range expectedErrors {
					expectedErrors[index] = errors.WithMeta(expectedError, datum)
				}
				errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
			},
			Entry("succeeds",
				func(datum *dexcom.Event) {},
			),
			Entry("eventSubType present",
				func(datum *dexcom.Event) {
					datum.EventSubType = pointer.FromString(test.RandomString())
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/eventSubType"),
			),
			Entry("unit present",
				func(datum *dexcom.Event) {
					datum.Unit = pointer.FromString(test.RandomString())
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/unit"),
			),
			Entry("value missing",
				func(datum *dexcom.Event) {
					datum.Value = nil
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
			),
			Entry("value empty",
				func(datum *dexcom.Event) {
					datum.Value = pointer.FromString("")
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/value"),
			),
			Entry("multiple errors",
				func(datum *dexcom.Event) {
					datum.EventSubType = pointer.FromString(test.RandomString())
					datum.Unit = pointer.FromString(test.RandomString())
					datum.Value = nil
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/eventSubType"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/unit"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/value"),
			),
		)
	})
})
