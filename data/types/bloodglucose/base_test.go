package bloodglucose_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/types/bloodglucose"
)

var _ = Describe("Base", func() {

	mmolL := "mmol/L"
	mmoll := "mmol/l"
	mgdL := "mg/dL"
	mgdl := "mg/dl"

	Context("convert value", func() {

		It("returns same value if already mmol/L", func() {
			fiveFive := 5.5
			Expect(bloodglucose.ConvertMgToMmol(&fiveFive, &mmolL)).To(Equal(&fiveFive))
		})

		It("returns same value if already mmol/L", func() {
			fiveFive := 5.5
			Expect(bloodglucose.ConvertMgToMmol(&fiveFive, &mmoll)).To(Equal(&fiveFive))
		})

		It("returns value in mmol/L if mg/dL", func() {
			threeSixty := 360.0
			expected := threeSixty / 18.01559
			Expect(bloodglucose.ConvertMgToMmol(&threeSixty, &mgdL)).To(Equal(&expected))
		})

		It("returns value in mmol/L if mg/dl", func() {
			threeSixty := 360.0
			expected := threeSixty / 18.01559
			Expect(bloodglucose.ConvertMgToMmol(&threeSixty, &mgdl)).To(Equal(&expected))
		})
	})

	// Context("convert units", func() {

	// 	It("keeps as mmol/L if already set as that", func() {
	// 		Expect(bloodglucose.NormalizeUnitName(&mmolL)).To(Equal(&mmol))
	// 	})

	// 	It("chages to mmol/L if mmol/l", func() {
	// 		Expect(bloodglucose.NormalizeUnitName(&mmoll)).To(Equal(&mmol))
	// 	})

	// 	It("keeps as mg/dL if already set as that", func() {
	// 		Expect(bloodglucose.NormalizeUnitName(&mgdL)).To(Equal(&mg))
	// 	})

	// 	It("chages to mg/dl if mg/dL", func() {
	// 		Expect(bloodglucose.NormalizeUnitName(&mgdl)).To(Equal(&mg))
	// 	})

	// 	It("does nothing if units are not what we expect", func() {
	// 		random := "flying pigs"
	// 		Expect(bloodglucose.NormalizeUnitName(&random)).To(Equal(&random))
	// 	})
	// })
})
