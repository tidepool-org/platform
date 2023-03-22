package dexcom_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/dexcom"
	"github.com/tidepool-org/platform/dexcom/test"
	"github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("Calibration", func() {

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
		Expect(dexcom.CalibrationUnits()).To(Equal([]string{dexcom.CalibrationUnitUnknown, dexcom.CalibrationUnitMgdL, dexcom.CalibrationUnitMmolL}))
	})

	Describe("Validate", func() {
		Describe("requires", func() {
			It("recordId", func() {
				calibration := test.RandomCalibration()
				calibration.ID = nil
				validator := validator.New()
				calibration.Validate(validator)
				Expect(validator.Error()).To(HaveOccurred())
			})
			It("systemTime", func() {
				calibration := test.RandomCalibration()
				calibration.SystemTime = nil
				validator := validator.New()
				calibration.Validate(validator)
				Expect(validator.Error()).To(HaveOccurred())
			})
			It("displayTime", func() {
				calibration := test.RandomCalibration()
				calibration.DisplayTime = nil
				validator := validator.New()
				calibration.Validate(validator)
				Expect(validator.Error()).To(HaveOccurred())
			})
			It("displayDevice", func() {
				calibration := test.RandomCalibration()
				calibration.DisplayDevice = nil
				validator := validator.New()
				calibration.Validate(validator)
				Expect(validator.Error()).To(HaveOccurred())
			})
			It("value", func() {
				calibration := test.RandomCalibration()
				calibration.Value = nil
				validator := validator.New()
				calibration.Validate(validator)
				Expect(validator.Error()).To(HaveOccurred())
			})
			It("transmitterId", func() {
				calibration := test.RandomCalibration()
				calibration.TransmitterID = nil
				validator := validator.New()
				calibration.Validate(validator)
				Expect(validator.Error()).To(HaveOccurred())
			})
			It("transmitterTicks", func() {
				calibration := test.RandomCalibration()
				calibration.TransmitterTicks = nil
				validator := validator.New()
				calibration.Validate(validator)
				Expect(validator.Error()).To(HaveOccurred())
			})
			It("transmitterGeneration", func() {
				calibration := test.RandomCalibration()
				calibration.TransmitterGeneration = nil
				validator := validator.New()
				calibration.Validate(validator)
				Expect(validator.Error()).To(HaveOccurred())
			})
		})
		Describe("does not require", func() {
			It("unit", func() {
				calibration := test.RandomCalibration()
				calibration.Unit = nil
				validator := validator.New()
				calibration.Validate(validator)
				Expect(validator.Error()).ToNot(HaveOccurred())
			})

		})
	})

})
