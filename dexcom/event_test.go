package dexcom_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/dexcom"
)

var _ = Describe("Event", func() {
	It("EventTypeCarbs is expected", func() {
		Expect(dexcom.EventTypeCarbs).To(Equal("carbs"))
	})

	It("EventTypeExercise is expected", func() {
		Expect(dexcom.EventTypeExercise).To(Equal("exercise"))
	})

	It("EventTypeHealth is expected", func() {
		Expect(dexcom.EventTypeHealth).To(Equal("health"))
	})

	It("EventTypeInsulin is expected", func() {
		Expect(dexcom.EventTypeInsulin).To(Equal("insulin"))
	})

	It("EventUnitCarbsGrams is expected", func() {
		Expect(dexcom.EventUnitCarbsGrams).To(Equal("grams"))
	})

	It("EventValueCarbsGramsMaximum is expected", func() {
		Expect(dexcom.EventValueCarbsGramsMaximum).To(Equal(250.0))
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
		Expect(dexcom.EventValueExerciseMinutesMaximum).To(Equal(360.0))
	})

	It("EventValueExerciseMinutesMinimum is expected", func() {
		Expect(dexcom.EventValueExerciseMinutesMinimum).To(Equal(0.0))
	})

	It("EventSubTypeHealthAlcohol is expected", func() {
		Expect(dexcom.EventSubTypeHealthAlcohol).To(Equal("alcohol"))
	})

	It("EventSubTypeHealthCycle is expected", func() {
		Expect(dexcom.EventSubTypeHealthCycle).To(Equal("cycle"))
	})

	It("EventSubTypeHealthHighSymptoms is expected", func() {
		Expect(dexcom.EventSubTypeHealthHighSymptoms).To(Equal("highSymptoms"))
	})

	It("EventSubTypeHealthIllness is expected", func() {
		Expect(dexcom.EventSubTypeHealthIllness).To(Equal("illness"))
	})

	It("EventSubTypeHealthLowSymptoms is expected", func() {
		Expect(dexcom.EventSubTypeHealthLowSymptoms).To(Equal("lowSymptoms"))
	})

	It("EventSubTypeHealthStress is expected", func() {
		Expect(dexcom.EventSubTypeHealthStress).To(Equal("stress"))
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

	It("EventStatusCreated is expected", func() {
		Expect(dexcom.EventStatusCreated).To(Equal("created"))
	})

	It("EventStatusDeleted is expected", func() {
		Expect(dexcom.EventStatusDeleted).To(Equal("deleted"))
	})

	It("EventTypes returns expected", func() {
		Expect(dexcom.EventTypes()).To(Equal([]string{"carbs", "exercise", "health", "insulin"}))
	})

	It("EventSubTypesExercise returns expected", func() {
		Expect(dexcom.EventSubTypesExercise()).To(Equal([]string{"light", "medium", "heavy"}))
	})

	It("EventSubTypesHealth returns expected", func() {
		Expect(dexcom.EventSubTypesHealth()).To(Equal([]string{"alcohol", "cycle", "highSymptoms", "illness", "lowSymptoms", "stress"}))
	})

	It("EventSubTypesInsulin returns expected", func() {
		Expect(dexcom.EventSubTypesInsulin()).To(Equal([]string{"fastActing", "longActing"}))
	})

	It("EventStatuses returns expected", func() {
		Expect(dexcom.EventStatuses()).To(Equal([]string{"created", "deleted"}))
	})
})
