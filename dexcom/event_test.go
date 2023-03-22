package dexcom_test

import (
	. "github.com/onsi/ginkgo"
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
		Expect(dexcom.EventTypes()).To(Equal([]string{"bloodGlucose", "carbs", "exercise", "health", "insulin", "notes", "unknown"}))
		Expect(dexcom.EventTypes()).To(Equal([]string{
			dexcom.EventTypeBG,
			dexcom.EventTypeCarbs,
			dexcom.EventTypeExercise,
			dexcom.EventTypeHealth,
			dexcom.EventTypeInsulin,
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
			event := test.RandomEvent()
			event.Type = pointer.FromString(dexcom.EventTypeHealth)
			event.SubType = pointer.FromString(dexcom.EventSubTypeHealthIllness)
			event.Unit = nil
			event.Value = pointer.FromString("stuff")
			validator := validator.New()
			event.Validate(validator)
			Expect(validator.Error()).ToNot(HaveOccurred())
		})
		Describe("requires", func() {
			It("systemTime", func() {
				event := test.RandomEvent()
				event.SystemTime = nil
				validator := validator.New()
				event.Validate(validator)
				Expect(validator.Error()).To(HaveOccurred())
			})
			It("displayTime", func() {
				event := test.RandomEvent()
				event.DisplayTime = nil
				validator := validator.New()
				event.Validate(validator)
				Expect(validator.Error()).To(HaveOccurred())
			})
			It("recordId", func() {
				event := test.RandomEvent()
				event.ID = nil
				validator := validator.New()
				event.Validate(validator)
				Expect(validator.Error()).To(HaveOccurred())
			})
			It("eventStatus", func() {
				event := test.RandomEvent()
				event.Status = nil
				validator := validator.New()
				event.Validate(validator)
				Expect(validator.Error()).To(HaveOccurred())
			})
			It("eventType", func() {
				event := test.RandomEvent()
				event.Type = nil
				validator := validator.New()
				event.Validate(validator)
				Expect(validator.Error()).To(HaveOccurred())
			})
			It("value", func() {
				event := test.RandomEvent()
				event.Value = nil
				validator := validator.New()
				event.Validate(validator)
				Expect(validator.Error()).To(HaveOccurred())
			})
			It("transmitterId", func() {
				event := test.RandomEvent()
				event.TransmitterID = nil
				validator := validator.New()
				event.Validate(validator)
				Expect(validator.Error()).To(HaveOccurred())
			})
			It("transmitterGeneration", func() {
				event := test.RandomEvent()
				event.TransmitterGeneration = nil
				validator := validator.New()
				event.Validate(validator)
				Expect(validator.Error()).To(HaveOccurred())
			})
			It("displayDevice", func() {
				event := test.RandomEvent()
				event.DisplayDevice = nil
				validator := validator.New()
				event.Validate(validator)
				Expect(validator.Error()).To(HaveOccurred())
			})
		})
		Describe("does not require", func() {
			It("eventSubType", func() {
				event := test.RandomEvent()
				event.SubType = nil
				validator := validator.New()
				event.Validate(validator)
				Expect(validator.Error()).ToNot(HaveOccurred())
			})
			It("unit when type unknown", func() {
				event := test.RandomEvent()
				event.Type = pointer.FromString(dexcom.EventTypeUnknown)
				event.Unit = nil
				validator := validator.New()
				event.Validate(validator)
				Expect(validator.Error()).ToNot(HaveOccurred())
			})
			It("unit when type notes", func() {
				event := test.RandomEvent()
				event.Type = pointer.FromString(dexcom.EventTypeNotes)
				event.Unit = nil
				validator := validator.New()
				event.Validate(validator)
				Expect(validator.Error()).ToNot(HaveOccurred())
			})
			It("unit when type health", func() {
				event := test.RandomEvent()
				event.Type = pointer.FromString(dexcom.EventTypeHealth)
				event.Unit = nil
				validator := validator.New()
				event.Validate(validator)
				Expect(validator.Error()).ToNot(HaveOccurred())
			})
		})
	})
})
