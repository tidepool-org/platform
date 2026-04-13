package times_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
	timeZone "github.com/tidepool-org/platform/time/zone"
	"github.com/tidepool-org/platform/times"
	timesTest "github.com/tidepool-org/platform/times/test"
)

var _ = Describe("time_range", func() {
	Context("TimeRange", func() {
		Context("ParseTimeRange", func() {
			It("returns nil when the object is missing", func() {
				Expect(times.ParseTimeRange(structureParser.NewObject(logTest.NewLogger(), nil))).To(BeNil())
			})

			It("returns new datum when the object is valid", func() {
				datum := timesTest.RandomTimeRange(test.AllowOptional())
				object := timesTest.NewObjectFromTimeRange(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(logTest.NewLogger(), &object)
				Expect(times.ParseTimeRange(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("TimeRange", func() {
			DescribeTable("serializes the datum as expected",
				func(mutator func(datum *times.TimeRange)) {
					datum := timesTest.RandomTimeRange(test.AllowOptional())
					mutator(datum)
					test.ExpectSerializedObjectJSON(datum, timesTest.NewObjectFromTimeRange(datum, test.ObjectFormatJSON))
					test.ExpectSerializedObjectBSON(datum, timesTest.NewObjectFromTimeRange(datum, test.ObjectFormatBSON))
				},
				Entry("succeeds",
					func(datum *times.TimeRange) {},
				),
				Entry("empty",
					func(datum *times.TimeRange) {
						*datum = times.TimeRange{}
					},
				),
				Entry("all",
					func(datum *times.TimeRange) {
						datum.From = pointer.From(test.RandomTime())
						datum.To = pointer.From(test.RandomTimeBefore(*datum.From))
					},
				),
			)

			Context("Parse", func() {
				DescribeTable("parses the datum",
					func(mutator func(object map[string]any, expectedDatum *times.TimeRange), expectedErrors ...error) {
						expectedDatum := timesTest.RandomTimeRange(test.AllowOptional())
						object := timesTest.NewObjectFromTimeRange(expectedDatum, test.ObjectFormatJSON)
						mutator(object, expectedDatum)
						result := &times.TimeRange{}
						errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(result), expectedErrors...)
						Expect(result).To(Equal(expectedDatum))
					},
					Entry("succeeds",
						func(object map[string]any, expectedDatum *times.TimeRange) {},
					),
					Entry("empty",
						func(object map[string]any, expectedDatum *times.TimeRange) {
							clear(object)
							*expectedDatum = times.TimeRange{}
						},
					),
					Entry("multiple errors",
						func(object map[string]any, expectedDatum *times.TimeRange) {
							object["from"] = true
							object["to"] = true
							expectedDatum.From = nil
							expectedDatum.To = nil
						},
						errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/from"),
						errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/to"),
					),
				)
			})

			Context("Validate", func() {
				DescribeTable("validates the datum",
					func(mutator func(datum *times.TimeRange), expectedErrors ...error) {
						datum := timesTest.RandomTimeRange(test.AllowOptional())
						mutator(datum)
						errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
					},
					Entry("succeeds",
						func(datum *times.TimeRange) {},
					),
					Entry("from missing",
						func(datum *times.TimeRange) {
							datum.From = nil
						},
					),
					Entry("from zero",
						func(datum *times.TimeRange) {
							datum.From = pointer.From(time.Time{})
						},
						errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/from"),
					),
					Entry("to missing",
						func(datum *times.TimeRange) {
							datum.To = nil
						},
					),
					Entry("to zero",
						func(datum *times.TimeRange) {
							datum.From = nil
							datum.To = pointer.From(time.Time{})
						},
						errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/to"),
					),
					Entry("to before from",
						func(datum *times.TimeRange) {
							datum.From = pointer.From(test.FutureNearTime())
							datum.To = pointer.From(test.PastNearTime())
						},
						errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotAfter(test.PastNearTime(), test.FutureNearTime()), "/to"),
					),
					Entry("multiple errors",
						func(datum *times.TimeRange) {
							datum.From = pointer.From(time.Time{})
							datum.To = pointer.From(time.Time{})
						},
						errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/from"),
						errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/to"),
					),
				)
			})

			Context("InLocation", func() {
				var timeRange times.TimeRange

				BeforeEach(func() {
					fromLocation, err := time.LoadLocation(test.RandomStringFromArray(timeZone.Names()))
					Expect(err).ToNot(HaveOccurred())
					Expect(fromLocation).ToNot(BeNil())

					toLocation, err := time.LoadLocation(test.RandomStringFromArray(timeZone.Names()))
					Expect(err).ToNot(HaveOccurred())
					Expect(toLocation).ToNot(BeNil())

					timeRange = *timesTest.RandomTimeRange()
					timeRange.From = pointer.FromTime(timeRange.From.In(fromLocation))
					timeRange.To = pointer.FromTime(timeRange.To.In(toLocation))
				})

				It("returns the time range unchanged when the location is nil", func() {
					Expect(timeRange.InLocation(nil)).To(Equal(timeRange))
				})

				It("returns a time range in the specified location", func() {
					targetLocation, err := time.LoadLocation(test.RandomStringFromArray(timeZone.Names()))
					Expect(err).ToNot(HaveOccurred())
					Expect(targetLocation).ToNot(BeNil())
					expectedTimeRange := times.TimeRange{
						From: pointer.FromTime(timeRange.From.In(targetLocation)),
						To:   pointer.FromTime(timeRange.To.In(targetLocation)),
					}
					Expect(timeRange.InLocation(targetLocation)).To(Equal(expectedTimeRange))
				})
			})

			Context("Clamped", func() {
				var first = test.RandomTime()
				var second = test.RandomTimeAfter(first)
				var third = test.RandomTimeAfter(second)

				DescribeTable("returns a clamped time range",
					func(timeRange times.TimeRange, minimum time.Time, maximum time.Time, expected times.TimeRange) {
						original := timeRange
						Expect(timeRange.Clamped(minimum, maximum)).To(Equal(expected))
						Expect(timeRange).To(Equal(original))
					},
					Entry("empty", times.TimeRange{}, time.Time{}, time.Time{}, times.TimeRange{}),
					Entry("from before minimum", times.TimeRange{From: pointer.From(first)}, second, third, times.TimeRange{From: pointer.From(second)}),
					Entry("from between minimum and maximum", times.TimeRange{From: pointer.From(second)}, first, third, times.TimeRange{From: pointer.From(second)}),
					Entry("from after maximum", times.TimeRange{From: pointer.From(third)}, first, second, times.TimeRange{From: pointer.From(second)}),
					Entry("to before minimum", times.TimeRange{To: pointer.From(first)}, second, third, times.TimeRange{To: pointer.From(second)}),
					Entry("to between minimum and maximum", times.TimeRange{To: pointer.From(second)}, first, third, times.TimeRange{To: pointer.From(second)}),
					Entry("to after maximum", times.TimeRange{To: pointer.From(third)}, first, second, times.TimeRange{To: pointer.From(second)}),
					Entry("multiple", times.TimeRange{From: pointer.From(first), To: pointer.From(third)}, second, third, times.TimeRange{From: pointer.From(second), To: pointer.From(third)}),
				)
			})

			Context("Truncated", func() {
				var from = test.RandomTime()
				var to = test.RandomTimeAfter(from)

				DescribeTable("returns a truncated time range",
					func(timeRange times.TimeRange, duration time.Duration, expected times.TimeRange) {
						original := timeRange
						Expect(timeRange.Truncated(duration)).To(Equal(expected))
						Expect(timeRange).To(Equal(original))
					},
					Entry("empty", times.TimeRange{}, time.Minute, times.TimeRange{}),
					Entry("from", times.TimeRange{From: pointer.From(from)}, time.Millisecond, times.TimeRange{From: pointer.From(from.Truncate(time.Millisecond))}),
					Entry("to", times.TimeRange{To: pointer.From(to)}, time.Minute, times.TimeRange{To: pointer.From(to.Truncate(time.Minute))}),
					Entry("multiple", times.TimeRange{From: pointer.From(from), To: pointer.From(to)}, time.Hour, times.TimeRange{From: pointer.From(from.Truncate(time.Hour)), To: pointer.From(to.Truncate(time.Hour))}),
				)
			})

			Context("Date", func() {
				It("returns a time range truncated to the date", func() {
					timeRange := timesTest.RandomTimeRange()
					expectedTimeRange := times.TimeRange{
						From: pointer.From(time.Date(timeRange.From.Year(), timeRange.From.Month(), timeRange.From.Day(), 0, 0, 0, 0, timeRange.From.Location())),
						To:   pointer.From(time.Date(timeRange.To.Year(), timeRange.To.Month(), timeRange.To.Day(), 0, 0, 0, 0, timeRange.To.Location())),
					}
					Expect(timeRange.Date()).To(Equal(expectedTimeRange))
				})
			})

			Context("String", func() {
				var from = test.RandomTime()
				var to = test.RandomTimeAfter(from)

				DescribeTable("returns the expected string",
					func(timeRange times.TimeRange, expected string) {
						Expect(timeRange.String(time.RFC3339)).To(Equal(expected))
					},
					Entry("empty", times.TimeRange{}, "-"),
					Entry("from", times.TimeRange{From: pointer.From(from)}, from.Format(time.RFC3339)+"-"),
					Entry("to", times.TimeRange{To: pointer.From(to)}, "-"+to.Format(time.RFC3339)),
					Entry("multiple", times.TimeRange{From: pointer.From(from), To: pointer.From(to)}, from.Format(time.RFC3339)+"-"+to.Format(time.RFC3339)),
				)
			})
		})
	})

	Context("TimeRangeMetadata", func() {
		Context("MetadataKeyTimeRange", func() {
			It("returns expected value", func() {
				Expect(times.MetadataKeyTimeRange).To(Equal("timeRange"))
			})
		})

		Context("TimeRangeMetadata", func() {
			DescribeTable("serializes the datum as expected",
				func(mutator func(datum *times.TimeRangeMetadata)) {
					datum := timesTest.RandomTimeRangeMetadata(test.AllowOptional())
					mutator(datum)
					test.ExpectSerializedObjectJSON(datum, timesTest.NewObjectFromTimeRangeMetadata(datum, test.ObjectFormatJSON))
					test.ExpectSerializedObjectBSON(datum, timesTest.NewObjectFromTimeRangeMetadata(datum, test.ObjectFormatBSON))
				},
				Entry("succeeds",
					func(datum *times.TimeRangeMetadata) {},
				),
				Entry("empty",
					func(datum *times.TimeRangeMetadata) {
						*datum = times.TimeRangeMetadata{}
					},
				),
				Entry("all",
					func(datum *times.TimeRangeMetadata) {
						datum.TimeRange = timesTest.RandomTimeRange(test.AllowOptional())
					},
				),
			)

			Context("Parse", func() {
				DescribeTable("parses the datum",
					func(mutator func(object map[string]any, expectedDatum *times.TimeRangeMetadata), expectedErrors ...error) {
						expectedDatum := timesTest.RandomTimeRangeMetadata(test.AllowOptional())
						object := timesTest.NewObjectFromTimeRangeMetadata(expectedDatum, test.ObjectFormatJSON)
						mutator(object, expectedDatum)
						result := &times.TimeRangeMetadata{}
						errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(result), expectedErrors...)
						Expect(result).To(Equal(expectedDatum))
					},
					Entry("succeeds",
						func(object map[string]any, expectedDatum *times.TimeRangeMetadata) {},
					),
					Entry("empty",
						func(object map[string]any, expectedDatum *times.TimeRangeMetadata) {
							clear(object)
							*expectedDatum = times.TimeRangeMetadata{}
						},
					),
					Entry("multiple errors",
						func(object map[string]any, expectedDatum *times.TimeRangeMetadata) {
							object["timeRange"] = true
							expectedDatum.TimeRange = nil
						},
						errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/timeRange"),
					),
				)
			})

			Context("Validate", func() {
				DescribeTable("validates the datum",
					func(mutator func(datum *times.TimeRangeMetadata), expectedErrors ...error) {
						datum := timesTest.RandomTimeRangeMetadata(test.AllowOptional())
						mutator(datum)
						errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
					},
					Entry("succeeds",
						func(datum *times.TimeRangeMetadata) {},
					),
					Entry("time range missing",
						func(datum *times.TimeRangeMetadata) {
							datum.TimeRange = nil
						},
					),
					Entry("time range empty",
						func(datum *times.TimeRangeMetadata) {
							datum.TimeRange = &times.TimeRange{}
						},
					),
					Entry("time range invalid",
						func(datum *times.TimeRangeMetadata) {
							datum.TimeRange = &times.TimeRange{
								From: pointer.From(time.Time{}),
							}
						},
						errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/timeRange/from"),
					),
					Entry("time range valid",
						func(datum *times.TimeRangeMetadata) {
							datum.TimeRange = timesTest.RandomTimeRange(test.AllowOptional())
						},
					),
					Entry("multiple errors",
						func(datum *times.TimeRangeMetadata) {
							datum.TimeRange = &times.TimeRange{
								From: pointer.From(time.Time{}),
							}
						},
						errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/timeRange/from"),
					),
				)
			})
		})
	})
})
