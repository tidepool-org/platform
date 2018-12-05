package dexcom_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/dexcom"
)

var _ = Describe("Calibration", func() {
	It("CalibrationUnitMgdL is expected", func() {
		Expect(dexcom.CalibrationUnitMgdL).To(Equal("mg/dL"))
	})

	It("CalibrationValueMgdLMaximum is expected", func() {
		Expect(dexcom.CalibrationValueMgdLMaximum).To(Equal(600.0))
	})

	It("CalibrationValueMgdLMinimum is expected", func() {
		Expect(dexcom.CalibrationValueMgdLMinimum).To(Equal(20.0))
	})

	It("CalibrationUnits returns expected", func() {
		Expect(dexcom.CalibrationUnits()).To(Equal([]string{"mg/dL"}))
	})
})
