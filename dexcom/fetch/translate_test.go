package fetch_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	dataTypes "github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/dexcom"
	dexcomFetch "github.com/tidepool-org/platform/dexcom/fetch"
	dexcomTest "github.com/tidepool-org/platform/dexcom/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
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

				ctx := log.NewContextWithLogger(context.Background(), logTest.NewLogger())

				dexcomSystemTime, err = dexcom.TimeFromString(systemTime)
				Expect(err).ToNot(HaveOccurred())
				Expect(dexcomSystemTime).ToNot(BeNil())
				if displayTime != nil {
					dexcomDisplayTime, err = dexcom.TimeFromString(*displayTime)
					Expect(err).ToNot(HaveOccurred())
					Expect(dexcomDisplayTime).ToNot(BeNil())
				}
				datum = &dataTypes.Base{}

				dexcomFetch.TranslateTime(ctx, dexcomSystemTime, dexcomDisplayTime, datum)
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
			Entry("systemTime is missing zone; displayTime is missing",
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
			Entry("systemTime is missing zone; displayTime is the same as systemTime UTC",
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
			Entry("systemTime is missing zone; displayTime is ahead systemTime and missing zone",
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
			Entry("systemTime is missing zone; displayTime is ahead systemTime and UTC",
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
			Entry("systemTime is missing zone; displayTime is ahead systemTime and not UTC",
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
			Entry("systemTime is missing zone; displayTime is behind systemTime and missing zone",
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
			Entry("systemTime is missing zone; displayTime is behind systemTime and UTC",
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
			Entry("systemTime is missing zone; displayTime is behind systemTime and not UTC",
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
			Entry("systemTime is UTC; displayTime is missing",
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
			Entry("systemTime is UTC; displayTime is the same as systemTime UTC",
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
			Entry("systemTime is UTC; displayTime is ahead systemTime and missing zone",
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
			Entry("systemTime is UTC; displayTime is ahead systemTime and UTC",
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
			Entry("systemTime is UTC; displayTime is ahead systemTime and not UTC",
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
			Entry("systemTime is UTC; displayTime is behind systemTime and missing zone",
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
			Entry("systemTime is UTC; displayTime is behind systemTime and UTC",
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
			Entry("systemTime is UTC; displayTime is behind systemTime and not UTC",
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
			Entry("systemTime is not UTC; displayTime is missing",
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
			Entry("systemTime is not UTC; displayTime is the same as systemTime",
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
			Entry("systemTime is not UTC; displayTime is ahead systemTime and missing zone",
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
			Entry("systemTime is not UTC; displayTime is ahead systemTime and UTC",
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
			Entry("systemTime is not UTC; displayTime is ahead systemTime and not UTC",
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
			Entry("systemTime is not UTC; displayTime is behind systemTime and missing zone",
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
			Entry("systemTime is not UTC; displayTime is behind systemTime and UTC",
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
			Entry("systemTime is not UTC; displayTime is behind systemTime and not UTC",
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

	Context("TranslateDeviceIDFromTransmitter", func() {
		It("returns nil if the transmitter generation is nil", func() {
			Expect(dexcomFetch.TranslateDeviceIDFromTransmitter(nil, pointer.FromString(dexcomTest.RandomTransmitterID()))).To(BeNil())
		})

		It("returns nil if the transmitter generation is invalid", func() {
			Expect(dexcomFetch.TranslateDeviceIDFromTransmitter(pointer.FromString("invalid"), pointer.FromString(dexcomTest.RandomTransmitterID()))).To(BeNil())
		})

		It("returns nil if the transmitter id is nil", func() {
			Expect(dexcomFetch.TranslateDeviceIDFromTransmitter(pointer.FromString(dexcom.DeviceTransmitterGenerationUnknown), nil)).To(BeNil())
		})

		It("returns nil if the transmitter id is empty", func() {
			Expect(dexcomFetch.TranslateDeviceIDFromTransmitter(pointer.FromString(dexcom.DeviceTransmitterGenerationUnknown), pointer.FromString(""))).To(BeNil())
		})

		DescribeTable("returns translated device id if transmitter generation is",
			func(transmitterGeneration string, expectedDeviceIDPrefix string) {
				transmitterID := dexcomTest.RandomTransmitterID()
				deviceID := dexcomFetch.TranslateDeviceIDFromTransmitter(pointer.FromString(transmitterGeneration), pointer.FromString(transmitterID))
				Expect(deviceID).ToNot(BeNil())
				Expect(*deviceID).To(Equal(fmt.Sprintf("%s_%s", expectedDeviceIDPrefix, transmitterID)))
			},
			Entry("DeviceTransmitterGenerationUnknown", dexcom.DeviceTransmitterGenerationUnknown, "Dexcom"),
			Entry("DeviceTransmitterGenerationG4", dexcom.DeviceTransmitterGenerationG4, "DexcomG4"),
			Entry("DeviceTransmitterGenerationG5", dexcom.DeviceTransmitterGenerationG5, "DexcomG5"),
			Entry("DeviceTransmitterGenerationG6", dexcom.DeviceTransmitterGenerationG6, "DexcomG6"),
			Entry("DeviceTransmitterGenerationG6Pro", dexcom.DeviceTransmitterGenerationG6Pro, "DexcomG6Pro"),
			Entry("DeviceTransmitterGenerationG6Plus", dexcom.DeviceTransmitterGenerationG6Plus, "DexcomG6Plus"),
			Entry("DeviceTransmitterGenerationPro", dexcom.DeviceTransmitterGenerationPro, "DexcomPro"),
			Entry("DeviceTransmitterGenerationG7", dexcom.DeviceTransmitterGenerationG7, "DexcomG7"),
			Entry("DeviceTransmitterGenerationG715Day", dexcom.DeviceTransmitterGenerationG715Day, "DexcomG7-15Day"),
		)
	})

	Context("TranslateDeviceIDPrefixFromTransmitterGeneration", func() {
		It("returns nil if the transmitter generation is nil", func() {
			Expect(dexcomFetch.TranslateDeviceIDPrefixFromTransmitterGeneration(nil)).To(BeNil())
		})

		DescribeTable("returns translated device id if transmitter generation is",
			func(transmitterGeneration string, expectedDeviceIDPrefix string) {
				deviceIDPrefix := dexcomFetch.TranslateDeviceIDPrefixFromTransmitterGeneration(pointer.FromString(transmitterGeneration))
				Expect(deviceIDPrefix).ToNot(BeNil())
				Expect(*deviceIDPrefix).To(Equal(expectedDeviceIDPrefix))
			},
			Entry("DeviceTransmitterGenerationUnknown", dexcom.DeviceTransmitterGenerationUnknown, "Dexcom"),
			Entry("DeviceTransmitterGenerationG4", dexcom.DeviceTransmitterGenerationG4, "DexcomG4"),
			Entry("DeviceTransmitterGenerationG5", dexcom.DeviceTransmitterGenerationG5, "DexcomG5"),
			Entry("DeviceTransmitterGenerationG6", dexcom.DeviceTransmitterGenerationG6, "DexcomG6"),
			Entry("DeviceTransmitterGenerationG6Pro", dexcom.DeviceTransmitterGenerationG6Pro, "DexcomG6Pro"),
			Entry("DeviceTransmitterGenerationG6Plus", dexcom.DeviceTransmitterGenerationG6Plus, "DexcomG6Plus"),
			Entry("DeviceTransmitterGenerationPro", dexcom.DeviceTransmitterGenerationPro, "DexcomPro"),
			Entry("DeviceTransmitterGenerationG7", dexcom.DeviceTransmitterGenerationG7, "DexcomG7"),
		)

		It("returns nil if the transmitter generation is invalid", func() {
			Expect(dexcomFetch.TranslateDeviceIDPrefixFromTransmitterGeneration(pointer.FromString("invalid"))).To(BeNil())
		})
	})
})
