package dexcom_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
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
		DescribeTable("errors when",
			func(setupFunc func() *dexcom.Calibration) {
				testCalibration := setupFunc()
				validator := validator.New()
				testCalibration.Validate(validator)
				Expect(validator.Error()).To(HaveOccurred())
			},
			Entry("required id is not set", func() *dexcom.Calibration {
				device := test.RandomCalibration()
				device.ID = nil
				return device
			}),
			Entry("required id is not set", func() *dexcom.Calibration {
				device := test.RandomCalibration()
				device.ID = nil
				return device
			}),
			Entry("required systemTime is not set", func() *dexcom.Calibration {
				device := test.RandomCalibration()
				device.SystemTime = nil
				return device
			}),
			Entry("required displayTime is not set", func() *dexcom.Calibration {
				device := test.RandomCalibration()
				device.DisplayTime = nil
				return device
			}),
			Entry("required displayDevice is not set", func() *dexcom.Calibration {
				device := test.RandomCalibration()
				device.DisplayDevice = nil
				return device
			}),
			Entry("required value is not set", func() *dexcom.Calibration {
				device := test.RandomCalibration()
				device.Value = nil
				return device
			}),
			Entry("required transmitterID is not set", func() *dexcom.Calibration {
				device := test.RandomCalibration()
				device.TransmitterID = nil
				return device
			}),
			Entry("required transmitterTicks is not set", func() *dexcom.Calibration {
				device := test.RandomCalibration()
				device.TransmitterTicks = nil
				return device
			}),
			Entry("required transmitterGeneration is not set", func() *dexcom.Calibration {
				device := test.RandomCalibration()
				device.TransmitterGeneration = nil
				return device
			}),
		)
		DescribeTable("does not error when",
			func(setupFunc func() *dexcom.Calibration) {
				testCalibration := setupFunc()
				validator := validator.New()
				testCalibration.Validate(validator)
				Expect(validator.Error()).ToNot(HaveOccurred())
			},
			Entry("unit is not set", func() *dexcom.Calibration {
				device := test.RandomCalibration()
				device.Unit = nil
				return device
			}),
		)
	})

})
