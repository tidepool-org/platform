package dexcom_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

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
				Expect(tm.MarshalText()).To(Equal([]byte(tm.Time.Format(dexcom.TimeFormatMilli))))
			})
		})

		Context("MarshalJSON", func() {
			It("returns successfully", func() {
				Expect(tm.MarshalJSON()).To(Equal([]byte(fmt.Sprintf("%q", tm.Time.Format(dexcom.TimeFormatMilli)))))
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

	Context("TimeFromString", func() {
		It("returns nil if raw is nil", func() {
			Expect(dexcom.TimeFromString(nil)).To(BeNil())
		})

		It("returns successfully if includes numeric zone", func() {
			raw := pointer.FromString("2023-04-05T16:57:06.696-07:00")
			tm := dexcom.TimeFromString(raw)
			Expect(tm).ToNot(BeNil())
			expectedTm, _ := time.Parse(dexcom.TimeFormatMilliZ, *raw)
			Expect(expectedTm).ToNot(BeNil())
			Expect(tm.Time).To(Equal(expectedTm))
		})
		It("returns successfully if includes UTC", func() {
			raw := pointer.FromString("2023-04-05T23:57:06.696Z")
			tm := dexcom.TimeFromString(raw)
			Expect(tm).ToNot(BeNil())
			expectedTm, _ := time.Parse(dexcom.TimeFormatMilliUTC, *raw)
			Expect(expectedTm).ToNot(BeNil())
			Expect(tm.Time).To(Equal(expectedTm))
		})
		It("returns successfully if includes no zone", func() {
			raw := pointer.FromString("2023-04-05T23:57:06.696")
			tm := dexcom.TimeFromString(raw)
			Expect(tm).ToNot(BeNil())
			expectedTm, _ := time.Parse(dexcom.TimeFormatMilli, *raw)
			Expect(expectedTm).ToNot(BeNil())
			Expect(tm.Time).To(Equal(expectedTm))
		})
		It("returns successfully if  minimal timestamp includes HH:MM:SS", func() {
			raw := pointer.FromString("2023-04-05T23:57:06")
			tm := dexcom.TimeFromString(raw)
			Expect(tm).ToNot(BeNil())
			expectedTm, _ := time.Parse(dexcom.TimeFormat, *raw)
			Expect(expectedTm).ToNot(BeNil())
			Expect(tm.Time).To(Equal(expectedTm))
		})
		It("returns successfully if minimal timestamp includes only HH:MM", func() {
			raw := pointer.FromString("2023-04-05T23:57")
			tm := dexcom.TimeFromString(raw)
			Expect(tm).ToNot(BeNil())
			expectedTm, _ := time.Parse(dexcom.TimeMinimalFormat, *raw)
			Expect(expectedTm).ToNot(BeNil())
			Expect(tm.Time).To(Equal(expectedTm))
		})
		It("returns successfully if minimal timestamp includes numeric zone", func() {
			raw := pointer.FromString("2023-04-05T23:57:06-07:00")
			tm := dexcom.TimeFromString(raw)
			Expect(tm).ToNot(BeNil())
			expectedTm, _ := time.Parse(dexcom.TimeFormatZ, *raw)
			Expect(expectedTm).ToNot(BeNil())
			Expect(tm.Time).To(Equal(expectedTm))
		})
		It("returns successfully if minimal timestamp includes UTC", func() {
			raw := pointer.FromString("2023-04-05T23:57:06Z")
			tm := dexcom.TimeFromString(raw)
			Expect(tm).ToNot(BeNil())
			expectedTm, _ := time.Parse(dexcom.TimeFormatUTC, *raw)
			Expect(expectedTm).ToNot(BeNil())
			Expect(tm.Time).To(Equal(expectedTm))
		})
	})

})
