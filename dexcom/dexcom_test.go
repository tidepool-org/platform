package dexcom_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/dexcom"
)

var _ = Describe("Device", func() {
	It("AlertNameFixedLow is expected", func() {
		Expect(dexcom.AlertNameFixedLow).To(Equal("fixedLow"))
	})

	It("AlertNameLow is expected", func() {
		Expect(dexcom.AlertNameLow).To(Equal("low"))
	})

	It("AlertNameHigh is expected", func() {
		Expect(dexcom.AlertNameHigh).To(Equal("high"))
	})

	It("AlertNameRise is expected", func() {
		Expect(dexcom.AlertNameRise).To(Equal("rise"))
	})

	It("AlertNameFall is expected", func() {
		Expect(dexcom.AlertNameFall).To(Equal("fall"))
	})

	It("AlertNameOutOfRange is expected", func() {
		Expect(dexcom.AlertNameOutOfRange).To(Equal("outOfRange"))
	})

	It("AlertNames is expected", func() {
		Expect(dexcom.AlertNames()).To(Equal([]string{"fixedLow", "low", "high", "rise", "fall", "outOfRange"}))
	})
})
