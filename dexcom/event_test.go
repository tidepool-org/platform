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
		It("Allows health events value to be 0", func() {
			event := test.RandomEvent()
			event.Type = pointer.FromString(dexcom.EventTypeHealth)
			event.SubType = pointer.FromString(dexcom.EventSubTypeHealthIllness)
			event.Unit = nil
			event.Value = pointer.FromFloat64(0)

			validator := validator.New()
			event.Validate(validator)

			Expect(validator.Error()).ToNot(HaveOccurred())
		})
	})
})
