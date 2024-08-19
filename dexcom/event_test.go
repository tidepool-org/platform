package dexcom_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/dexcom/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure/validator"

	"github.com/tidepool-org/platform/dexcom"
)

var _ = Describe("Event", func() {

	It("EventUnitUnknown is expected", func() {
		Expect(dexcom.EventUnitUnknown).To(Equal("unknown"))
	})

	It("EventUnitMgdL is expected", func() {
		Expect(dexcom.EventUnitMgdL).To(Equal("mg/dL"))
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

	It("EventUnitExerciseMinutes is expected", func() {
		Expect(dexcom.EventUnitExerciseMinutes).To(Equal("minutes"))
	})

	It("EventValueExerciseMinutesMaximum is expected", func() {
		Expect(dexcom.EventValueExerciseMinutesMaximum).To(Equal(10080.0))
	})

	It("EventValueExerciseMinutesMinimum is expected", func() {
		Expect(dexcom.EventValueExerciseMinutesMinimum).To(Equal(0.0))
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

	It("EventStatuses is expected", func() {
		Expect(dexcom.EventStatuses()).To(Equal([]string{"created", "updated", "deleted"}))
		Expect(dexcom.EventStatuses()).To(Equal([]string{
			dexcom.EventStatusCreated,
			dexcom.EventStatusUpdated,
			dexcom.EventStatusDeleted,
		}))
	})

	It("EventTypes returns expected", func() {
		Expect(dexcom.EventTypes()).To(Equal([]string{"bloodGlucose", "carbs", "exercise", "health", "insulin", "note", "notes", "unknown"}))
		Expect(dexcom.EventTypes()).To(Equal([]string{
			dexcom.EventTypeBG,
			dexcom.EventTypeCarbs,
			dexcom.EventTypeExercise,
			dexcom.EventTypeHealth,
			dexcom.EventTypeInsulin,
			dexcom.EventTypeNote,
			dexcom.EventTypeNotes,
			dexcom.EventTypeUnknown,
		}))
	})

	It("EventSubTypesExercise returns expected", func() {
		Expect(dexcom.EventSubTypesExercise()).To(Equal([]string{"light", "medium", "heavy"}))
		Expect(dexcom.EventSubTypesExercise()).To(Equal([]string{
			dexcom.EventSubTypeExerciseLight,
			dexcom.EventSubTypeExerciseMedium,
			dexcom.EventSubTypeExerciseHeavy,
		}))
	})

	It("EventSubTypesHealth returns expected", func() {
		Expect(dexcom.EventSubTypesHealth()).To(Equal([]string{"alcohol", "cycle", "highSymptoms", "illness", "lowSymptoms", "stress"}))
		Expect(dexcom.EventSubTypesHealth()).To(Equal([]string{
			dexcom.EventSubTypeHealthAlcohol,
			dexcom.EventSubTypeHealthCycle,
			dexcom.EventSubTypeHealthHighSymptoms,
			dexcom.EventSubTypeHealthIllness,
			dexcom.EventSubTypeHealthLowSymptoms,
			dexcom.EventSubTypeHealthStress,
		}))
	})

	It("EventSubTypesInsulin returns expected", func() {
		Expect(dexcom.EventSubTypesInsulin()).To(Equal([]string{"fastActing", "longActing"}))
		Expect(dexcom.EventSubTypesInsulin()).To(Equal([]string{
			dexcom.EventSubTypeInsulinFastActing,
			dexcom.EventSubTypeInsulinLongActing,
		}))
	})

	Describe("Validate", func() {
		It("Allows health events to have no units", func() {
			event := test.RandomEvent(pointer.FromString(dexcom.EventTypeHealth))
			event.Unit = nil
			event.Value = pointer.FromString("stuff")
			validator := validator.New()
			event.Validate(validator)
			Expect(validator.Error()).ToNot(HaveOccurred())
		})
		DescribeTable("requires",
			func(setupEventFunc func() *dexcom.Event) {
				testEvent := setupEventFunc()
				validator := validator.New()
				testEvent.Validate(validator)
				Expect(validator.Error()).To(HaveOccurred())
			},
			Entry("displayDevice to be set", func() *dexcom.Event {
				event := test.RandomEvent(nil)
				event.DisplayDevice = nil
				return event
			}),
			Entry("transmitterGeneration to be set", func() *dexcom.Event {
				event := test.RandomEvent(nil)
				event.TransmitterGeneration = nil
				return event
			}),
			Entry("transmitterId to be set", func() *dexcom.Event {
				event := test.RandomEvent(nil)
				event.TransmitterID = nil
				return event
			}),
			Entry("systemTime to be set", func() *dexcom.Event {
				Skip("systemTime will occassionally be unset that kills the whole upload process")
				event := test.RandomEvent(nil)
				event.SystemTime = nil
				return event
			}),
			Entry("displayTime to be set", func() *dexcom.Event {
				Skip("displayTime will occassionally be unset that kills the whole upload process")
				event := test.RandomEvent(nil)
				event.DisplayTime = nil
				return event
			}),
			Entry("id to be set", func() *dexcom.Event {
				event := test.RandomEvent(nil)
				event.ID = nil
				return event
			}),
			Entry("status to be set", func() *dexcom.Event {
				event := test.RandomEvent(nil)
				event.Status = nil
				return event
			}),
			Entry("type to be set", func() *dexcom.Event {
				event := test.RandomEvent(nil)
				event.Type = nil
				return event
			}),
			Entry("value to be set", func() *dexcom.Event {
				event := test.RandomEvent(nil)
				event.Value = nil
				return event
			}),
		)
		DescribeTable("expects value to be valid",
			func(setupEventFunc func() *dexcom.Event) {
				testEvent := setupEventFunc()
				validator := validator.New()
				testEvent.Validate(validator)
				Expect(validator.Error()).To(HaveOccurred())
			},
			Entry("when EventTypeBG", func() *dexcom.Event {
				event := test.RandomEvent(pointer.FromString(dexcom.EventTypeBG))
				event.Value = pointer.FromString("10s")
				return event
			}),
			Entry("when EventTypeCarbs", func() *dexcom.Event {
				event := test.RandomEvent(pointer.FromString(dexcom.EventTypeCarbs))
				event.Value = pointer.FromString("99.s")
				return event
			}),
			Entry("when EventTypeInsulin", func() *dexcom.Event {
				event := test.RandomEvent(pointer.FromString(dexcom.EventTypeInsulin))
				event.Value = pointer.FromString("10s")
				return event
			}),
			Entry("when EventTypeExercise", func() *dexcom.Event {
				event := test.RandomEvent(pointer.FromString(dexcom.EventTypeExercise))
				event.Value = pointer.FromString("100m")
				return event
			}),
		)
		DescribeTable("value is valid at minimum",
			func(setupEventFunc func() *dexcom.Event) {
				testEvent := setupEventFunc()
				validator := validator.New()
				testEvent.Validate(validator)
				Expect(validator.Error()).ToNot(HaveOccurred())
			},
			Entry("when EventTypeBG", func() *dexcom.Event {
				event := test.RandomEvent(pointer.FromString(dexcom.EventTypeBG))
				event.Value = pointer.FromString(fmt.Sprintf("%v", dexcom.EventValueMgdLMinimum))
				return event
			}),
			Entry("when EventTypeCarbs", func() *dexcom.Event {
				event := test.RandomEvent(pointer.FromString(dexcom.EventTypeCarbs))
				event.Value = pointer.FromString(fmt.Sprintf("%v", dexcom.EventValueCarbsGramsMinimum))
				return event
			}),
			Entry("when EventTypeInsulin", func() *dexcom.Event {
				event := test.RandomEvent(pointer.FromString(dexcom.EventTypeInsulin))
				event.Value = pointer.FromString(fmt.Sprintf("%v", dexcom.EventValueInsulinUnitsMinimum))
				return event
			}),
			Entry("when EventTypeExercise", func() *dexcom.Event {
				event := test.RandomEvent(pointer.FromString(dexcom.EventTypeExercise))
				event.Value = pointer.FromString(fmt.Sprintf("%v", dexcom.EventValueExerciseMinutesMinimum))
				return event
			}),
		)
		DescribeTable("does not require",
			func(setupEventFunc func() *dexcom.Event) {
				testEvent := setupEventFunc()
				validator := validator.New()
				testEvent.Validate(validator)
				Expect(validator.Error()).ToNot(HaveOccurred())
			},
			Entry("eventSubType to be set", func() *dexcom.Event {
				event := test.RandomEvent(nil)
				event.SubType = nil
				return event
			}),
			Entry("unit to be set when type is unknown", func() *dexcom.Event {
				event := test.RandomEvent(pointer.FromString(dexcom.EventTypeUnknown))
				event.Unit = nil
				return event
			}),
			Entry("unit to be set when type is notes", func() *dexcom.Event {
				event := test.RandomEvent(pointer.FromString(dexcom.EventTypeNote))
				event.Unit = nil
				return event
			}),
			Entry("unit to be set when type is health", func() *dexcom.Event {
				event := test.RandomEvent(pointer.FromString(dexcom.EventTypeHealth))
				event.Unit = nil
				return event
			}),
		)
	})
})
