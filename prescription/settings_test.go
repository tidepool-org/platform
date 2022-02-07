package prescription_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/structure/validator"

	"github.com/tidepool-org/platform/prescription"
	"github.com/tidepool-org/platform/prescription/test"
)

var _ = Describe("Initial Settings", func() {
	var settings *prescription.InitialSettings
	var validate structure.Validator

	BeforeEach(func() {
		settings = test.RandomInitialSettings()
		validate = validator.New()
		Expect(validate.Validate(settings)).ToNot(HaveOccurred())
	})

	Describe("ValidateSubmittedPrescription", func() {
		BeforeEach(func() {
			validate = validator.New()
		})

		It("fails with empty basal rate schedule", func() {
			settings.BasalRateSchedule = nil
			settings.ValidateSubmittedPrescription(validate)
			Expect(validate.Error()).To(HaveOccurred())
		})

		It("fails with empty blood glucose target schedule", func() {
			settings.BloodGlucoseTargetSchedule = nil
			settings.ValidateSubmittedPrescription(validate)
			Expect(validate.Error()).To(HaveOccurred())
		})

		It("fails with empty glucose safety limit", func() {
			settings.GlucoseSafetyLimit = nil
			settings.ValidateSubmittedPrescription(validate)
			Expect(validate.Error()).To(HaveOccurred())
		})

		It("fails with empty carbohydrate ratio schedule", func() {
			settings.CarbohydrateRatioSchedule = nil
			settings.ValidateSubmittedPrescription(validate)
			Expect(validate.Error()).To(HaveOccurred())
		})

		It("fails with empty insulin sensitivity schedule", func() {
			settings.InsulinSensitivitySchedule = nil
			settings.ValidateSubmittedPrescription(validate)
			Expect(validate.Error()).To(HaveOccurred())
		})

		It("fails with empty insulin model", func() {
			settings.InsulinModel = nil
			settings.ValidateSubmittedPrescription(validate)
			Expect(validate.Error()).To(HaveOccurred())
		})

		It("fails with empty basal rate maximum", func() {
			settings.BasalRateMaximum = nil
			settings.ValidateSubmittedPrescription(validate)
			Expect(validate.Error()).To(HaveOccurred())
		})

		It("fails with empty bolus amount maximum", func() {
			settings.BolusAmountMaximum = nil
			settings.ValidateSubmittedPrescription(validate)
			Expect(validate.Error()).To(HaveOccurred())
		})

		It("fails with nil pump id", func() {
			settings.PumpID = nil
			settings.ValidateSubmittedPrescription(validate)
			Expect(validate.Error()).To(HaveOccurred())
		})

		It("fails with nil cgm id", func() {
			settings.CgmID = nil
			settings.ValidateSubmittedPrescription(validate)
			Expect(validate.Error()).To(HaveOccurred())
		})
	})
})
