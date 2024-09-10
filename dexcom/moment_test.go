package dexcom_test

import (
	"sort"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/dexcom"
	dexcomTest "github.com/tidepool-org/platform/dexcom/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Moment", func() {
	Context("ParseMoment", func() {
		It("returns nil if the parser does not exist", func() {
			Expect(dexcom.ParseMoment(structureParser.NewObject(logTest.NewLogger(), nil))).To(BeNil())
		})

		DescribeTable("parses the datum",
			func(mutator func(object map[string]interface{}, expectedDatum *dexcom.Moment), expectedErrors ...error) {
				expectedDatum := dexcomTest.RandomMomentFromRange(test.PastFarTime(), test.FutureFarTime())
				object := dexcomTest.NewObjectFromMoment(expectedDatum, test.ObjectFormatJSON)
				mutator(object, expectedDatum)
				parser := structureParser.NewObject(logTest.NewLogger(), &object)
				datum := dexcom.ParseMoment(parser)
				errorsTest.ExpectEqual(parser.Error(), expectedErrors...)
				Expect(datum).To(Equal(expectedDatum))
			},
			Entry("succeeds",
				func(object map[string]interface{}, expectedDatum *dexcom.Moment) {},
			),
			Entry("systemTime invalid type",
				func(object map[string]interface{}, expectedDatum *dexcom.Moment) {
					object["systemTime"] = true
					expectedDatum.SystemTime = nil
				},
				errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/systemTime"),
			),
			Entry("systemTime invalid time",
				func(object map[string]interface{}, expectedDatum *dexcom.Moment) {
					object["systemTime"] = "invalid"
					expectedDatum.SystemTime = nil
				},
				errorsTest.WithPointerSource(structureParser.ErrorValueTimeNotParsable("invalid", time.RFC3339Nano), "/systemTime"),
			),
			Entry("displayTime invalid type",
				func(object map[string]interface{}, expectedDatum *dexcom.Moment) {
					object["displayTime"] = true
					expectedDatum.DisplayTime = nil
				},
				errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/displayTime"),
			),
			Entry("displayTime invalid time",
				func(object map[string]interface{}, expectedDatum *dexcom.Moment) {
					object["displayTime"] = "invalid"
					expectedDatum.DisplayTime = nil
				},
				errorsTest.WithPointerSource(structureParser.ErrorValueTimeNotParsable("invalid", time.RFC3339Nano), "/displayTime"),
			),
		)
	})

	Context("NewMoment", func() {
		It("returns an empty Moment", func() {
			moment := dexcom.NewMoment()
			Expect(moment).ToNot(BeNil())
			Expect(moment.SystemTime).To(BeZero())
			Expect(moment.DisplayTime).To(BeZero())
		})
	})

	Context("Moment", func() {
		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dexcom.Moment), expectedErrors ...error) {
					expectedDatum := dexcomTest.RandomMomentFromRange(test.PastFarTime(), test.FutureFarTime())
					object := dexcomTest.NewObjectFromMoment(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := &dexcom.Moment{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dexcom.Moment) {},
				),
				Entry("systemTime invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.Moment) {
						object["systemTime"] = true
						expectedDatum.SystemTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/systemTime"),
				),
				Entry("systemTime invalid time",
					func(object map[string]interface{}, expectedDatum *dexcom.Moment) {
						object["systemTime"] = "invalid"
						expectedDatum.SystemTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorValueTimeNotParsable("invalid", time.RFC3339Nano), "/systemTime"),
				),
				Entry("displayTime invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.Moment) {
						object["displayTime"] = true
						expectedDatum.DisplayTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/displayTime"),
				),
				Entry("displayTime invalid time",
					func(object map[string]interface{}, expectedDatum *dexcom.Moment) {
						object["displayTime"] = "invalid"
						expectedDatum.DisplayTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorValueTimeNotParsable("invalid", time.RFC3339Nano), "/displayTime"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dexcom.Moment), expectedErrors ...error) {
					datum := dexcomTest.RandomMomentFromRange(test.PastFarTime(), test.FutureFarTime())
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dexcom.Moment) {},
				),
				Entry("systemTime missing",
					func(datum *dexcom.Moment) {
						datum.SystemTime = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/systemTime"),
				),
				Entry("systemTime zero",
					func(datum *dexcom.Moment) {
						datum.SystemTime.Time = time.Time{}
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/systemTime"),
				),
				Entry("displayTime missing",
					func(datum *dexcom.Moment) {
						datum.DisplayTime = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/displayTime"),
				),
				Entry("displayTime zero",
					func(datum *dexcom.Moment) {
						datum.DisplayTime.Time = time.Time{}
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/displayTime"),
				),
				Entry("multiple errors",
					func(datum *dexcom.Moment) {
						datum.SystemTime = nil
						datum.DisplayTime = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/systemTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/displayTime"),
				),
			)
		})

		Context("SystemTimeRaw", func() {
			It("returns nil when systemTime is nil", func() {
				datum := dexcomTest.RandomMomentFromRange(test.PastFarTime(), test.FutureFarTime())
				datum.SystemTime = nil
				Expect(datum.SystemTimeRaw()).To(BeNil())
			})

			It("returns nil when systemTime is zero", func() {
				datum := dexcomTest.RandomMomentFromRange(test.PastFarTime(), test.FutureFarTime())
				datum.SystemTime.Time = time.Time{}
				Expect(datum.SystemTimeRaw()).To(BeNil())
			})

			It("returns systemTime when systemTime is valid", func() {
				datum := dexcomTest.RandomMomentFromRange(test.PastFarTime(), test.FutureFarTime())
				systemTimeRaw := datum.SystemTimeRaw()
				Expect(systemTimeRaw).ToNot(BeNil())
				Expect(*systemTimeRaw).To(Equal(datum.SystemTime.Time))
			})
		})
	})

	Context("Moment", func() {
		Context("Compact", func() {
			It("removes nil moments from array", func() {
				moment1 := dexcomTest.RandomMomentFromRange(test.PastFarTime(), test.FutureFarTime())
				moment2 := dexcomTest.RandomMomentFromRange(test.PastFarTime(), test.FutureFarTime())
				moments := dexcom.Moments{nil, moment1, nil, moment2, nil}
				Expect(moments.Compact()).To(Equal(dexcom.Moments{moment1, moment2}))
			})
		})
	})

	Context("MomentsBySystemTimeRaw", func() {
		It("sorts moments by systemTime with nils at end", func() {
			moment1 := dexcomTest.RandomMomentFromTime(time.Unix(1600000000, 0))
			moment2 := dexcomTest.RandomMomentFromTime(time.Unix(1700000000, 0))
			moment3 := dexcomTest.RandomMomentFromTime(time.Unix(1800000000, 0))
			moments := dexcom.Moments{nil, moment2, nil, moment1, nil, moment3, nil}
			sort.Stable(dexcom.MomentsBySystemTimeRaw(moments))
			Expect(moments).To(Equal(dexcom.Moments{moment1, moment2, moment3, nil, nil, nil, nil}))
		})
	})
})
