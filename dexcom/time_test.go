package dexcom_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/dexcom"
	structureParser "github.com/tidepool-org/platform/structure/parser"
)

var _ = Describe("Time", func() {
	Context("ParseTime", func() {
		It("returns nil if key not present in parser", func() {
			parser := structureParser.NewObject(&map[string]any{})
			tm := dexcom.ParseTime(parser, "test")
			Expect(tm).To(BeNil())
			Expect(parser.HasError()).To(BeFalse())
		})

		DescribeTable("does not parse an invalid time string",
			func(timeString string) {
				parser := structureParser.NewObject(&map[string]any{"test": timeString})
				tm := dexcom.ParseTime(parser, "test")
				Expect(parser.HasError()).To(BeTrue())
				Expect(parser.Error()).To(MatchError(fmt.Sprintf(`value "%s" is not a parsable time of format "2006-01-02T15:04:05.999999999Z07:00"`, timeString)))
				Expect(tm).To(BeNil())
			},
			Entry("empty string", ""),
			Entry("invalid string", "invalid"),
			Entry("invalid year", "200"),
			Entry("invalid month separator", "2001:0"),
			Entry("invalid month", "2001-0"),
			Entry("invalid day separator", "2001-02:0"),
			Entry("invalid day", "2001-02-0"),
			Entry("invalid date separator", "2001-02-03:"),
			Entry("invalid hour", "2001-02-03TH"),
			Entry("invalid hour separator", "2001-02-03T14-"),
			Entry("invalid minute", "2001-02-03T14:M"),
			Entry("invalid minute separator", "2001-02-03T14:15-"),
			Entry("invalid second", "2001-02-03T14:15:M"),
			Entry("invalid second separator", "2001-02-03T14:15:16-"),
			Entry("invalid nanosecond", "2001-02-03T14:15:16.N"),
			Entry("invalid remainder", "2001-02-03T14:15:16.1234567890"),
			Entry("invalid time string with alernate time zone format -00", "2001-02-03T14:15:16-05"),
			Entry("invalid time string with alernate time zone format -00:00:00", "2001-02-03T14:15:16-05:00:00"),
			Entry("invalid time string with alernate time zone format Z00:00", "2001-02-03T14:15:16Z05:00"),
			Entry("invalid time string with alernate time zone format -0000", "2001-02-03T14:15:16-0500"),
		)

		DescribeTable("parses the time appropriately",
			func(timeString string, expectedTime time.Time) {
				parser := structureParser.NewObject(&map[string]any{"test": timeString})
				tm := dexcom.ParseTime(parser, "test")
				Expect(parser.HasError()).To(BeFalse())
				Expect(tm).ToNot(BeNil())
				Expect(tm.Raw()).ToNot(BeNil())
				Expect(*tm.Raw()).To(Equal(expectedTime))
			},
			Entry("valid date string", "2001-02-03", time.Date(2001, 2, 3, 0, 0, 0, 0, time.UTC)),
			Entry("valid date string with Z time zone", "2001-02-03Z", time.Date(2001, 2, 3, 0, 0, 0, 0, time.UTC)),
			Entry("valid date string with -05:00 time zone", "2001-02-03-05:00", time.Date(2001, 2, 3, 0, 0, 0, 0, time.FixedZone("", -5*60*60))),
			Entry("valid date string with T", "2001-02-03T", time.Date(2001, 2, 3, 0, 0, 0, 0, time.UTC)),
			Entry("valid date string with T and Z time zone", "2001-02-03TZ", time.Date(2001, 2, 3, 0, 0, 0, 0, time.UTC)),
			Entry("valid date string with T and -05:00 time zone", "2001-02-03T-05:00", time.Date(2001, 2, 3, 0, 0, 0, 0, time.FixedZone("", -5*60*60))),
			Entry("valid date string with hour", "2001-02-03T14", time.Date(2001, 2, 3, 14, 0, 0, 0, time.UTC)),
			Entry("valid date string with hour and Z time zone", "2001-02-03T14Z", time.Date(2001, 2, 3, 14, 0, 0, 0, time.UTC)),
			Entry("valid date string with hour and -05:00 time zone", "2001-02-03T14-05:00", time.Date(2001, 2, 3, 14, 0, 0, 0, time.FixedZone("", -5*60*60))),
			Entry("valid date string with hour and minute", "2001-02-03T14:15", time.Date(2001, 2, 3, 14, 15, 0, 0, time.UTC)),
			Entry("valid date string with hour and minute and Z time zone", "2001-02-03T14:15Z", time.Date(2001, 2, 3, 14, 15, 0, 0, time.UTC)),
			Entry("valid date string with hour and minute and -05:00 time zone", "2001-02-03T14:15-05:00", time.Date(2001, 2, 3, 14, 15, 0, 0, time.FixedZone("", -5*60*60))),
			Entry("valid time string", "2001-02-03T14:15:16", time.Date(2001, 2, 3, 14, 15, 16, 0, time.UTC)),
			Entry("valid time string with Z time zone", "2001-02-03T14:15:16Z", time.Date(2001, 2, 3, 14, 15, 16, 0, time.UTC)),
			Entry("valid time string with -05:00 time zone", "2001-02-03T14:15:16-05:00", time.Date(2001, 2, 3, 14, 15, 16, 0, time.FixedZone("", -5*60*60))),
			Entry("valid time string with milliseconds", "2001-02-03T14:15:16.7531", time.Date(2001, 2, 3, 14, 15, 16, 753100000, time.UTC)),
			Entry("valid time string with milliseconds and Z time zone", "2001-02-03T14:15:16.7531Z", time.Date(2001, 2, 3, 14, 15, 16, 753100000, time.UTC)),
			Entry("valid time string with milliseconds and -05:00 time zone", "2001-02-03T14:15:16.7531-05:00", time.Date(2001, 2, 3, 14, 15, 16, 753100000, time.FixedZone("", -5*60*60))),
			Entry("valid time string with missing zero prefixes", "2001-2-3T4:0:6.", time.Date(2001, 2, 3, 4, 0, 6, 0, time.UTC)),
		)
	})

	Context("TimeFromString", func() {
		DescribeTable("does not parse an invalid time string",
			func(timeString string) {
				tm, err := dexcom.TimeFromString(timeString)
				Expect(err).To(MatchError("time is not parsable"))
				Expect(tm).To(BeNil())
			},
			Entry("empty string", ""),
			Entry("invalid string", "invalid"),
			Entry("invalid year", "200"),
			Entry("invalid month separator", "2001:0"),
			Entry("invalid month", "2001-0"),
			Entry("invalid day separator", "2001-02:0"),
			Entry("invalid day", "2001-02-0"),
			Entry("invalid date separator", "2001-02-03:"),
			Entry("invalid hour", "2001-02-03TH"),
			Entry("invalid hour separator", "2001-02-03T14-"),
			Entry("invalid minute", "2001-02-03T14:M"),
			Entry("invalid minute separator", "2001-02-03T14:15-"),
			Entry("invalid second", "2001-02-03T14:15:M"),
			Entry("invalid second separator", "2001-02-03T14:15:16-"),
			Entry("invalid nanosecond", "2001-02-03T14:15:16.N"),
			Entry("invalid remainder", "2001-02-03T14:15:16.1234567890"),
			Entry("invalid time string with alernate time zone format -00", "2001-02-03T14:15:16-05"),
			Entry("invalid time string with alernate time zone format -00:00:00", "2001-02-03T14:15:16-05:00:00"),
			Entry("invalid time string with alernate time zone format Z00:00", "2001-02-03T14:15:16Z05:00"),
			Entry("invalid time string with alernate time zone format -0000", "2001-02-03T14:15:16-0500"),
		)

		DescribeTable("parses the time appropriately",
			func(timeString string, expectedTime time.Time, expectedZoneParsed bool) {
				tm, err := dexcom.TimeFromString(timeString)
				Expect(err).ToNot(HaveOccurred())
				Expect(tm).ToNot(BeNil())
				Expect(tm.Raw()).ToNot(BeNil())
				Expect(*tm.Raw()).To(Equal(expectedTime))
				Expect(tm.ZoneParsed()).To(Equal(expectedZoneParsed))
				bites, err := tm.MarshalJSON()
				Expect(err).ToNot(HaveOccurred())
				Expect(bites).ToNot(BeNil())
				Expect(bites).To(Equal([]byte(fmt.Sprintf(`"%s"`, timeString))))
			},
			Entry("valid date string", "2001-02-03", time.Date(2001, 2, 3, 0, 0, 0, 0, time.UTC), false),
			Entry("valid date string with Z time zone", "2001-02-03Z", time.Date(2001, 2, 3, 0, 0, 0, 0, time.UTC), true),
			Entry("valid date string with -05:00 time zone", "2001-02-03-05:00", time.Date(2001, 2, 3, 0, 0, 0, 0, time.FixedZone("", -5*60*60)), true),
			Entry("valid date string with T", "2001-02-03T", time.Date(2001, 2, 3, 0, 0, 0, 0, time.UTC), false),
			Entry("valid date string with T and Z time zone", "2001-02-03TZ", time.Date(2001, 2, 3, 0, 0, 0, 0, time.UTC), true),
			Entry("valid date string with T and -05:00 time zone", "2001-02-03T-05:00", time.Date(2001, 2, 3, 0, 0, 0, 0, time.FixedZone("", -5*60*60)), true),
			Entry("valid date string with hour", "2001-02-03T14", time.Date(2001, 2, 3, 14, 0, 0, 0, time.UTC), false),
			Entry("valid date string with hour and Z time zone", "2001-02-03T14Z", time.Date(2001, 2, 3, 14, 0, 0, 0, time.UTC), true),
			Entry("valid date string with hour and -05:00 time zone", "2001-02-03T14-05:00", time.Date(2001, 2, 3, 14, 0, 0, 0, time.FixedZone("", -5*60*60)), true),
			Entry("valid date string with hour and minute", "2001-02-03T14:15", time.Date(2001, 2, 3, 14, 15, 0, 0, time.UTC), false),
			Entry("valid date string with hour and minute and Z time zone", "2001-02-03T14:15Z", time.Date(2001, 2, 3, 14, 15, 0, 0, time.UTC), true),
			Entry("valid date string with hour and minute and -05:00 time zone", "2001-02-03T14:15-05:00", time.Date(2001, 2, 3, 14, 15, 0, 0, time.FixedZone("", -5*60*60)), true),
			Entry("valid time string", "2001-02-03T14:15:16", time.Date(2001, 2, 3, 14, 15, 16, 0, time.UTC), false),
			Entry("valid time string with Z time zone", "2001-02-03T14:15:16Z", time.Date(2001, 2, 3, 14, 15, 16, 0, time.UTC), true),
			Entry("valid time string with -05:00 time zone", "2001-02-03T14:15:16-05:00", time.Date(2001, 2, 3, 14, 15, 16, 0, time.FixedZone("", -5*60*60)), true),
			Entry("valid time string with milliseconds", "2001-02-03T14:15:16.7531", time.Date(2001, 2, 3, 14, 15, 16, 753100000, time.UTC), false),
			Entry("valid time string with milliseconds and Z time zone", "2001-02-03T14:15:16.7531Z", time.Date(2001, 2, 3, 14, 15, 16, 753100000, time.UTC), true),
			Entry("valid time string with milliseconds and -05:00 time zone", "2001-02-03T14:15:16.7531-05:00", time.Date(2001, 2, 3, 14, 15, 16, 753100000, time.FixedZone("", -5*60*60)), true),
			Entry("valid time string with missing zero prefixes", "2001-2-3T4:0:6.", time.Date(2001, 2, 3, 4, 0, 6, 0, time.UTC), false),
		)
	})

	Context("TimeFromRaw", func() {
		It("successfully creates the time", func() {
			expectedTime := time.Date(2001, 2, 3, 14, 15, 16, 0, time.FixedZone("", -5*60*60))
			tm := dexcom.TimeFromRaw(expectedTime)
			Expect(tm).ToNot(BeNil())
			Expect(tm.Raw()).To(Equal(&expectedTime))
			Expect(tm.ZoneParsed()).To(BeTrue())
			bites, err := tm.MarshalJSON()
			Expect(err).ToNot(HaveOccurred())
			Expect(bites).ToNot(BeNil())
			Expect(bites).To(Equal([]byte(fmt.Sprintf(`"%s"`, expectedTime.Format(time.RFC3339Nano)))))
		})
	})

	Context("TimeFromTime", func() {
		It("returns nil if time is nil", func() {
			Expect(dexcom.TimeFromTime(nil)).To(BeNil())
		})

		It("successfully creates the time", func() {
			expectedTime := dexcom.TimeFromRaw(time.Date(2001, 2, 3, 14, 15, 16, 0, time.FixedZone("", -5*60*60)))
			tm := dexcom.TimeFromTime(expectedTime)
			Expect(tm).To(Equal(expectedTime))
		})
	})

	Context("Time created manually", func() {
		It("successfully creates the time", func() {
			expectedTime := time.Date(2001, 2, 3, 14, 15, 16, 0, time.FixedZone("", -5*60*60))
			tm := &dexcom.Time{Time: expectedTime}
			Expect(tm.Raw()).To(Equal(&expectedTime))
			Expect(tm.ZoneParsed()).To(BeTrue())
			bites, err := tm.MarshalJSON()
			Expect(err).ToNot(HaveOccurred())
			Expect(bites).ToNot(BeNil())
			Expect(bites).To(Equal([]byte(fmt.Sprintf(`"%s"`, expectedTime.Format(time.RFC3339Nano)))))
		})
	})
})
