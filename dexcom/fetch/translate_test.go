package fetch_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	dataTypes "github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/dexcom"
	dexcomFetch "github.com/tidepool-org/platform/dexcom/fetch"
	"github.com/tidepool-org/platform/pointer"
)

var _ = Describe("Translate", func() {
	Context("TranslateTime", func() {
		systemTimeMissingZone := "2023-10-31T15:31:42.123"
		systemTimeUTC := "2023-10-31T15:31:42.123Z"
		systemTimeNotUTC := "2023-10-31T10:31:42.123-05:00"
		displayTimeAheadMissingZone := "2023-11-03T20:42:56.468"
		displayTimeAheadUTC := "2023-11-03T20:42:56.468Z"
		displayTimeAheadNotUTC := "2023-11-03T20:42:56.468-06:00"
		displayTimeBehindMissingZone := "2023-10-29T09:02:29.468"
		displayTimeBehindUTC := "2023-10-29T09:02:29.468Z"
		displayTimeBehindNotUTC := "2023-10-29T09:02:29.468-06:00"

		DescribeTable("translates various time fields",
			func(systemTime string, displayTime *string, expectedDatum *dataTypes.Base) {
				var dexcomSystemTime *dexcom.Time
				var dexcomDisplayTime *dexcom.Time
				var datum *dataTypes.Base
				var err error

				dexcomSystemTime, err = dexcom.TimeFromString(systemTime)
				Expect(err).ToNot(HaveOccurred())
				Expect(dexcomSystemTime).ToNot(BeNil())
				if displayTime != nil {
					dexcomDisplayTime, err = dexcom.TimeFromString(*displayTime)
					Expect(err).ToNot(HaveOccurred())
					Expect(dexcomDisplayTime).ToNot(BeNil())
				}
				datum = &dataTypes.Base{}

				dexcomFetch.TranslateTime(dexcomSystemTime, dexcomDisplayTime, datum)
				Expect(datum.Time).To(Equal(expectedDatum.Time))
				Expect(datum.DeviceTime).To(Equal(expectedDatum.DeviceTime))
				Expect(datum.TimeZoneOffset).To(Equal(expectedDatum.TimeZoneOffset))
				Expect(datum.ClockDriftOffset).To(Equal(expectedDatum.ClockDriftOffset))
				Expect(datum.ConversionOffset).To(Equal(expectedDatum.ConversionOffset))
				Expect(datum.Payload).ToNot(BeNil())
				Expect(datum.Payload.Get("systemTime")).To(Equal(dexcomSystemTime))
				if displayTime != nil {
					Expect(datum.Payload.Get("displayTime")).To(Equal(dexcomDisplayTime))
				} else {
					Expect(datum.Payload.Get("displayTime")).To(BeNil())
				}

				// Ensure all calculations are correct
				if datum.DeviceTime != nil {
					calculatedDeviceTime := datum.Time.Add(time.Duration(*datum.TimeZoneOffset) * time.Minute)
					if datum.ClockDriftOffset != nil {
						calculatedDeviceTime = calculatedDeviceTime.Add(time.Duration(*datum.ClockDriftOffset) * time.Millisecond)
					}
					if datum.ConversionOffset != nil {
						calculatedDeviceTime = calculatedDeviceTime.Add(time.Duration(*datum.ConversionOffset) * time.Millisecond)
					}
					Expect(*datum.DeviceTime).To(Equal(calculatedDeviceTime.Format(data.DeviceTimeFormat)))
				}
			},
			Entry("system time is missing zone; display time is missing",
				systemTimeMissingZone,
				nil,
				&dataTypes.Base{
					Time:             pointer.FromTime(time.Date(2023, 10, 31, 15, 31, 42, 123000000, time.UTC)),
					DeviceTime:       nil,
					TimeZoneOffset:   nil,
					ClockDriftOffset: nil,
					ConversionOffset: nil,
				},
			),
			Entry("system time is missing zone; display time is the same as system time UTC",
				systemTimeMissingZone,
				&systemTimeUTC,
				&dataTypes.Base{
					Time:             pointer.FromTime(time.Date(2023, 10, 31, 15, 31, 42, 123000000, time.UTC)),
					DeviceTime:       pointer.FromString("2023-10-31T15:31:42"),
					TimeZoneOffset:   pointer.FromInt(0),
					ClockDriftOffset: nil,
					ConversionOffset: nil,
				},
			),
			Entry("system time is missing zone; display time is ahead system time and missing zone",
				systemTimeMissingZone,
				&displayTimeAheadMissingZone,
				&dataTypes.Base{
					Time:             pointer.FromTime(time.Date(2023, 10, 31, 15, 31, 42, 123000000, time.UTC)),
					DeviceTime:       pointer.FromString("2023-11-03T20:42:56"),
					TimeZoneOffset:   pointer.FromInt(5 * 60),
					ClockDriftOffset: pointer.FromInt(674345),
					ConversionOffset: pointer.FromInt(3 * 24 * 60 * 60 * 1000),
				},
			),
			Entry("system time is missing zone; display time is ahead system time and UTC",
				systemTimeMissingZone,
				&displayTimeAheadUTC,
				&dataTypes.Base{
					Time:             pointer.FromTime(time.Date(2023, 10, 31, 15, 31, 42, 123000000, time.UTC)),
					DeviceTime:       pointer.FromString("2023-11-03T20:42:56"),
					TimeZoneOffset:   pointer.FromInt(0),
					ClockDriftOffset: pointer.FromInt(674345),
					ConversionOffset: pointer.FromInt(3*24*60*60*1000 + 5*60*60*1000),
				},
			),
			Entry("system time is missing zone; display time is ahead system time and not UTC",
				systemTimeMissingZone,
				&displayTimeAheadNotUTC,
				&dataTypes.Base{
					Time:             pointer.FromTime(time.Date(2023, 10, 31, 15, 31, 42, 123000000, time.UTC)),
					DeviceTime:       pointer.FromString("2023-11-03T20:42:56"),
					TimeZoneOffset:   pointer.FromInt(-6 * 60),
					ClockDriftOffset: pointer.FromInt(674345),
					ConversionOffset: pointer.FromInt(3*24*60*60*1000 + 11*60*60*1000),
				},
			),
			Entry("system time is missing zone; display time is behind system time and missing zone",
				systemTimeMissingZone,
				&displayTimeBehindMissingZone,
				&dataTypes.Base{
					Time:             pointer.FromTime(time.Date(2023, 10, 31, 15, 31, 42, 123000000, time.UTC)),
					DeviceTime:       pointer.FromString("2023-10-29T09:02:29"),
					TimeZoneOffset:   pointer.FromInt(-6.5 * 60),
					ClockDriftOffset: pointer.FromInt(47345),
					ConversionOffset: pointer.FromInt(-2 * 24 * 60 * 60 * 1000),
				},
			),
			Entry("system time is missing zone; display time is behind system time and UTC",
				systemTimeMissingZone,
				&displayTimeBehindUTC,
				&dataTypes.Base{
					Time:             pointer.FromTime(time.Date(2023, 10, 31, 15, 31, 42, 123000000, time.UTC)),
					DeviceTime:       pointer.FromString("2023-10-29T09:02:29"),
					TimeZoneOffset:   pointer.FromInt(0),
					ClockDriftOffset: pointer.FromInt(47345),
					ConversionOffset: pointer.FromInt(-2*24*60*60*1000 - 6.5*60*60*1000),
				},
			),
			Entry("system time is missing zone; display time is behind system time and not UTC",
				systemTimeMissingZone,
				&displayTimeBehindNotUTC,
				&dataTypes.Base{
					Time:             pointer.FromTime(time.Date(2023, 10, 31, 15, 31, 42, 123000000, time.UTC)),
					DeviceTime:       pointer.FromString("2023-10-29T09:02:29"),
					TimeZoneOffset:   pointer.FromInt(-6 * 60),
					ClockDriftOffset: pointer.FromInt(47345),
					ConversionOffset: pointer.FromInt(-2*24*60*60*1000 - 30*60*1000),
				},
			),
			Entry("system time is UTC; display time is missing",
				systemTimeUTC,
				nil,
				&dataTypes.Base{
					Time:             pointer.FromTime(time.Date(2023, 10, 31, 15, 31, 42, 123000000, time.UTC)),
					DeviceTime:       nil,
					TimeZoneOffset:   nil,
					ClockDriftOffset: nil,
					ConversionOffset: nil,
				},
			),
			Entry("system time is UTC; display time is the same as system time UTC",
				systemTimeUTC,
				&systemTimeUTC,
				&dataTypes.Base{
					Time:             pointer.FromTime(time.Date(2023, 10, 31, 15, 31, 42, 123000000, time.UTC)),
					DeviceTime:       pointer.FromString("2023-10-31T15:31:42"),
					TimeZoneOffset:   pointer.FromInt(0),
					ClockDriftOffset: nil,
					ConversionOffset: nil,
				},
			),
			Entry("system time is UTC; display time is ahead system time and missing zone",
				systemTimeUTC,
				&displayTimeAheadMissingZone,
				&dataTypes.Base{
					Time:             pointer.FromTime(time.Date(2023, 10, 31, 15, 31, 42, 123000000, time.UTC)),
					DeviceTime:       pointer.FromString("2023-11-03T20:42:56"),
					TimeZoneOffset:   pointer.FromInt(5 * 60),
					ClockDriftOffset: pointer.FromInt(674345),
					ConversionOffset: pointer.FromInt(3 * 24 * 60 * 60 * 1000),
				},
			),
			Entry("system time is UTC; display time is ahead system time and UTC",
				systemTimeUTC,
				&displayTimeAheadUTC,
				&dataTypes.Base{
					Time:             pointer.FromTime(time.Date(2023, 10, 31, 15, 31, 42, 123000000, time.UTC)),
					DeviceTime:       pointer.FromString("2023-11-03T20:42:56"),
					TimeZoneOffset:   pointer.FromInt(0),
					ClockDriftOffset: pointer.FromInt(674345),
					ConversionOffset: pointer.FromInt(3*24*60*60*1000 + 5*60*60*1000),
				},
			),
			Entry("system time is UTC; display time is ahead system time and not UTC",
				systemTimeUTC,
				&displayTimeAheadNotUTC,
				&dataTypes.Base{
					Time:             pointer.FromTime(time.Date(2023, 10, 31, 15, 31, 42, 123000000, time.UTC)),
					DeviceTime:       pointer.FromString("2023-11-03T20:42:56"),
					TimeZoneOffset:   pointer.FromInt(-6 * 60),
					ClockDriftOffset: pointer.FromInt(674345),
					ConversionOffset: pointer.FromInt(3*24*60*60*1000 + 11*60*60*1000),
				},
			),
			Entry("system time is UTC; display time is behind system time and missing zone",
				systemTimeUTC,
				&displayTimeBehindMissingZone,
				&dataTypes.Base{
					Time:             pointer.FromTime(time.Date(2023, 10, 31, 15, 31, 42, 123000000, time.UTC)),
					DeviceTime:       pointer.FromString("2023-10-29T09:02:29"),
					TimeZoneOffset:   pointer.FromInt(-6.5 * 60),
					ClockDriftOffset: pointer.FromInt(47345),
					ConversionOffset: pointer.FromInt(-2 * 24 * 60 * 60 * 1000),
				},
			),
			Entry("system time is UTC; display time is behind system time and UTC",
				systemTimeUTC,
				&displayTimeBehindUTC,
				&dataTypes.Base{
					Time:             pointer.FromTime(time.Date(2023, 10, 31, 15, 31, 42, 123000000, time.UTC)),
					DeviceTime:       pointer.FromString("2023-10-29T09:02:29"),
					TimeZoneOffset:   pointer.FromInt(0),
					ClockDriftOffset: pointer.FromInt(47345),
					ConversionOffset: pointer.FromInt(-2*24*60*60*1000 - 6.5*60*60*1000),
				},
			),
			Entry("system time is UTC; display time is behind system time and not UTC",
				systemTimeUTC,
				&displayTimeBehindNotUTC,
				&dataTypes.Base{
					Time:             pointer.FromTime(time.Date(2023, 10, 31, 15, 31, 42, 123000000, time.UTC)),
					DeviceTime:       pointer.FromString("2023-10-29T09:02:29"),
					TimeZoneOffset:   pointer.FromInt(-6 * 60),
					ClockDriftOffset: pointer.FromInt(47345),
					ConversionOffset: pointer.FromInt(-2*24*60*60*1000 - 30*60*1000),
				},
			),
			Entry("system time is not UTC; display time is missing",
				systemTimeNotUTC,
				nil,
				&dataTypes.Base{
					Time:             pointer.FromTime(time.Date(2023, 10, 31, 15, 31, 42, 123000000, time.UTC)),
					DeviceTime:       nil,
					TimeZoneOffset:   nil,
					ClockDriftOffset: nil,
					ConversionOffset: nil,
				},
			),
			Entry("system time is not UTC; display time is the same as system time",
				systemTimeNotUTC,
				&systemTimeNotUTC,
				&dataTypes.Base{
					Time:             pointer.FromTime(time.Date(2023, 10, 31, 15, 31, 42, 123000000, time.UTC)),
					DeviceTime:       pointer.FromString("2023-10-31T10:31:42"),
					TimeZoneOffset:   pointer.FromInt(-5 * 60),
					ClockDriftOffset: nil,
					ConversionOffset: nil,
				},
			),
			Entry("system time is not UTC; display time is ahead system time and missing zone",
				systemTimeNotUTC,
				&displayTimeAheadMissingZone,
				&dataTypes.Base{
					Time:             pointer.FromTime(time.Date(2023, 10, 31, 15, 31, 42, 123000000, time.UTC)),
					DeviceTime:       pointer.FromString("2023-11-03T20:42:56"),
					TimeZoneOffset:   pointer.FromInt(5 * 60),
					ClockDriftOffset: pointer.FromInt(674345),
					ConversionOffset: pointer.FromInt(3 * 24 * 60 * 60 * 1000),
				},
			),
			Entry("system time is not UTC; display time is ahead system time and UTC",
				systemTimeNotUTC,
				&displayTimeAheadUTC,
				&dataTypes.Base{
					Time:             pointer.FromTime(time.Date(2023, 10, 31, 15, 31, 42, 123000000, time.UTC)),
					DeviceTime:       pointer.FromString("2023-11-03T20:42:56"),
					TimeZoneOffset:   pointer.FromInt(0),
					ClockDriftOffset: pointer.FromInt(674345),
					ConversionOffset: pointer.FromInt(3*24*60*60*1000 + 5*60*60*1000),
				},
			),
			Entry("system time is not UTC; display time is ahead system time and not UTC",
				systemTimeNotUTC,
				&displayTimeAheadNotUTC,
				&dataTypes.Base{
					Time:             pointer.FromTime(time.Date(2023, 10, 31, 15, 31, 42, 123000000, time.UTC)),
					DeviceTime:       pointer.FromString("2023-11-03T20:42:56"),
					TimeZoneOffset:   pointer.FromInt(-6 * 60),
					ClockDriftOffset: pointer.FromInt(674345),
					ConversionOffset: pointer.FromInt(3*24*60*60*1000 + 11*60*60*1000),
				},
			),
			Entry("system time is not UTC; display time is behind system time and missing zone",
				systemTimeNotUTC,
				&displayTimeBehindMissingZone,
				&dataTypes.Base{
					Time:             pointer.FromTime(time.Date(2023, 10, 31, 15, 31, 42, 123000000, time.UTC)),
					DeviceTime:       pointer.FromString("2023-10-29T09:02:29"),
					TimeZoneOffset:   pointer.FromInt(-6.5 * 60),
					ClockDriftOffset: pointer.FromInt(47345),
					ConversionOffset: pointer.FromInt(-2 * 24 * 60 * 60 * 1000),
				},
			),
			Entry("system time is not UTC; display time is behind system time and UTC",
				systemTimeNotUTC,
				&displayTimeBehindUTC,
				&dataTypes.Base{
					Time:             pointer.FromTime(time.Date(2023, 10, 31, 15, 31, 42, 123000000, time.UTC)),
					DeviceTime:       pointer.FromString("2023-10-29T09:02:29"),
					TimeZoneOffset:   pointer.FromInt(0),
					ClockDriftOffset: pointer.FromInt(47345),
					ConversionOffset: pointer.FromInt(-2*24*60*60*1000 - 6.5*60*60*1000),
				},
			),
			Entry("system time is not UTC; display time is behind system time and not UTC",
				systemTimeNotUTC,
				&displayTimeBehindNotUTC,
				&dataTypes.Base{
					Time:             pointer.FromTime(time.Date(2023, 10, 31, 15, 31, 42, 123000000, time.UTC)),
					DeviceTime:       pointer.FromString("2023-10-29T09:02:29"),
					TimeZoneOffset:   pointer.FromInt(-6 * 60),
					ClockDriftOffset: pointer.FromInt(47345),
					ConversionOffset: pointer.FromInt(-2*24*60*60*1000 - 30*60*1000),
				},
			),
		)
	})
})
