package dexcom_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/dexcom"
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
})
