package bloodglucose

import (
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"
)

func TestingDatumBase() map[string]interface{} {
	return map[string]interface{}{
		"userId":           "b676436f60",
		"groupId":          "43099shgs55",
		"uploadId":         "upid_b856b0e6e519",
		"deviceTime":       "2014-06-11T06:00:00.000Z",
		"time":             "2014-06-11T06:00:00.000Z",
		"timezoneOffset":   0,
		"conversionOffset": 0,
		"clockDriftOffset": 0,
		"deviceId":         "InsOmn-111111111",
	}
}

var _ = Describe("Base", func() {

	mmolL := "mmol/L"
	mmoll := "mmol/l"
	mgdL := "mg/dL"
	mgdl := "mg/dl"

	Context("convert value", func() {

		It("returns same value if already mmol/L", func() {
			fiveFive := 5.5
			Expect(convertMgToMmol(&fiveFive, &mmolL)).To(Equal(&fiveFive))
		})

		It("returns same value if already mmol/L", func() {
			fiveFive := 5.5
			Expect(convertMgToMmol(&fiveFive, &mmoll)).To(Equal(&fiveFive))
		})

		It("returns value in mmol/L if mg/dL", func() {
			threeSixty := 360.0
			expected := threeSixty / 18.01559
			Expect(convertMgToMmol(&threeSixty, &mgdL)).To(Equal(&expected))
		})

		It("returns value in mmol/L if mg/dl", func() {
			threeSixty := 360.0
			expected := threeSixty / 18.01559
			Expect(convertMgToMmol(&threeSixty, &mgdl)).To(Equal(&expected))
		})

	})

	Context("convert units", func() {

		It("keeps as mmol/L if already set as that", func() {
			Expect(normalizeUnitName(&mmolL)).To(Equal(&mmol))
		})

		It("chages to mmol/L if mmol/l", func() {
			Expect(normalizeUnitName(&mmoll)).To(Equal(&mmol))
		})

		It("keeps as mg/dL if already set as that", func() {
			Expect(normalizeUnitName(&mgdL)).To(Equal(&mg))
		})

		It("chages to mg/dl if mg/dL", func() {
			Expect(normalizeUnitName(&mgdl)).To(Equal(&mg))
		})

		It("does nothing if units are not what we expect", func() {
			random := "flying pigs"
			Expect(normalizeUnitName(&random)).To(Equal(&random))
		})

	})

})
