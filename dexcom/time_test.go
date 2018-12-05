package dexcom_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"

	"github.com/tidepool-org/platform/dexcom"
	dexcomTest "github.com/tidepool-org/platform/dexcom/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Time", func() {
	Context("NewTime", func() {
		It("returns successfully", func() {
			tm := dexcom.NewTime()
			Expect(tm).ToNot(BeNil())
			Expect(tm.Time).To(BeZero())
		})
	})

	Context("with new time", func() {
		var tm *dexcom.Time

		BeforeEach(func() {
			tm = dexcomTest.RandomTime()
		})

		Context("Raw", func() {
			It("returns nil if time is nil", func() {
				tm = nil
				Expect(tm.Raw()).To(BeNil())
			})

			It("returns successfully", func() {
				raw := test.RandomTime()
				tm.Time = raw
				Expect(*tm.Raw()).To(Equal(raw))
			})
		})

		Context("MarshalText", func() {
			It("returns successfully", func() {
				Expect(tm.MarshalText()).To(Equal([]byte(tm.Time.Format(dexcom.TimeFormat))))
			})
		})

		Context("MarshalJSON", func() {
			It("returns successfully", func() {
				Expect(tm.MarshalJSON()).To(Equal([]byte(fmt.Sprintf("%q", tm.Time.Format(dexcom.TimeFormat)))))
			})
		})
	})

	Context("TimeFromRaw", func() {
		It("returns nil if raw is nil", func() {
			Expect(dexcom.TimeFromRaw(nil)).To(BeNil())
		})

		It("returns successfully", func() {
			raw := pointer.FromTime(test.RandomTime())
			tm := dexcom.TimeFromRaw(raw)
			Expect(tm).ToNot(BeNil())
			Expect(tm.Time).To(Equal(*raw))
		})
	})
})
