package dexcom_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/dexcom"
	dexcomTest "github.com/tidepool-org/platform/dexcom/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("DataRange", func() {
	It("DataRangesResponseRecordType is expected", func() {
		Expect(dexcom.DataRangesResponseRecordType).To(Equal("dataRange"))
	})

	It("DataRangesResponseRecordVersion is expected", func() {
		Expect(dexcom.DataRangesResponseRecordVersion).To(Equal("3.0"))
	})

	Context("ParseDataRangesResponse", func() {
		It("returns nil if the object is nil", func() {
			parser := structureParser.NewObject(logTest.NewLogger(), nil)
			Expect(dexcom.ParseDataRangesResponse(parser)).To(BeNil())
		})

		It("returns the parsed object", func() {
			expectedDatum := dexcomTest.RandomDataRangesResponse()
			object := dexcomTest.NewObjectFromDataRangesResponse(expectedDatum, test.ObjectFormatJSON)
			parser := structureParser.NewObject(logTest.NewLogger(), &object)
			Expect(dexcom.ParseDataRangesResponse(parser)).To(Equal(expectedDatum))
		})
	})

	Context("DataRangesResponse", func() {
		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dexcom.DataRangesResponse), expectedErrors ...error) {
					expectedDatum := dexcomTest.RandomDataRangesResponse()
					object := dexcomTest.NewObjectFromDataRangesResponse(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					for index, expectedError := range expectedErrors {
						expectedErrors[index] = errors.WithMeta(expectedError, expectedDatum)
					}
					datum := &dexcom.DataRangesResponse{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dexcom.DataRangesResponse) {},
				),
				Entry("recordType invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.DataRangesResponse) {
						object["recordType"] = true
						expectedDatum.RecordType = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/recordType"),
				),
				Entry("recordVersion invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.DataRangesResponse) {
						object["recordVersion"] = true
						expectedDatum.RecordVersion = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/recordVersion"),
				),
				Entry("userId invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.DataRangesResponse) {
						object["userId"] = true
						expectedDatum.UserID = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/userId"),
				),
				Entry("calibrations invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.DataRangesResponse) {
						object["calibrations"] = true
						expectedDatum.Calibrations = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/calibrations"),
				),
				Entry("egvs invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.DataRangesResponse) {
						object["egvs"] = true
						expectedDatum.EGVs = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/egvs"),
				),
				Entry("events invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.DataRangesResponse) {
						object["events"] = true
						expectedDatum.Events = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/events"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dexcom.DataRangesResponse), expectedErrors ...error) {
					datum := dexcomTest.RandomDataRangesResponse()
					mutator(datum)
					for index, expectedError := range expectedErrors {
						expectedErrors[index] = errors.WithMeta(expectedError, datum)
					}
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dexcom.DataRangesResponse) {},
				),
				Entry("recordType missing",
					func(datum *dexcom.DataRangesResponse) {
						datum.RecordType = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/recordType"),
				),
				Entry("recordType invalid",
					func(datum *dexcom.DataRangesResponse) {
						datum.RecordType = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalid", dexcom.DataRangesResponseRecordType), "/recordType"),
				),
				Entry("recordVersion missing",
					func(datum *dexcom.DataRangesResponse) {
						datum.RecordVersion = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/recordVersion"),
				),
				Entry("recordVersion invalid",
					func(datum *dexcom.DataRangesResponse) {
						datum.RecordVersion = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalid", dexcom.DataRangesResponseRecordVersion), "/recordVersion"),
				),
				Entry("userId missing",
					func(datum *dexcom.DataRangesResponse) {
						datum.UserID = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/userId"),
				),
				Entry("userId empty",
					func(datum *dexcom.DataRangesResponse) {
						datum.UserID = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/userId"),
				),
				Entry("calibrations missing",
					func(datum *dexcom.DataRangesResponse) {
						datum.Calibrations = nil
					},
				),
				Entry("calibrations invalid",
					func(datum *dexcom.DataRangesResponse) {
						datum.Calibrations.Start = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/calibrations/start"),
				),
				Entry("egvs missing",
					func(datum *dexcom.DataRangesResponse) {
						datum.EGVs = nil
					},
				),
				Entry("egvs invalid",
					func(datum *dexcom.DataRangesResponse) {
						datum.EGVs.Start = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/egvs/start"),
				),
				Entry("events missing",
					func(datum *dexcom.DataRangesResponse) {
						datum.Events = nil
					},
				),
				Entry("events invalid",
					func(datum *dexcom.DataRangesResponse) {
						datum.Events.Start = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/events/start"),
				),
				Entry("multiple errors",
					func(datum *dexcom.DataRangesResponse) {
						datum.RecordType = nil
						datum.RecordVersion = nil
						datum.UserID = nil
						datum.Calibrations.Start = nil
						datum.EGVs.Start = nil
						datum.Events.Start = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/recordType"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/recordVersion"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/userId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/calibrations/start"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/egvs/start"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/events/start"),
				),
			)
		})

		Context("DataRange", func() {
			It("returns nil if all moments are nil", func() {
				datum := dexcomTest.RandomDataRangesResponse()
				datum.Calibrations = nil
				datum.EGVs = nil
				datum.Events = nil
				Expect(datum.DataRange()).To(BeNil())
			})

			It("returns nil if all moment starts are nil", func() {
				datum := dexcomTest.RandomDataRangesResponse()
				datum.Calibrations.Start = nil
				datum.EGVs.Start = nil
				datum.Events.Start = nil
				Expect(datum.DataRange()).To(BeNil())
			})

			It("returns nil if all moment ends are nil", func() {
				datum := dexcomTest.RandomDataRangesResponse()
				datum.Calibrations.End = nil
				datum.EGVs.End = nil
				datum.Events.End = nil
				Expect(datum.DataRange()).To(BeNil())
			})

			It("returns data range successfully", func() {
				datum := dexcomTest.RandomDataRangesResponse()
				datum.Calibrations.Start = dexcomTest.RandomMomentFromTime(time.Unix(1730000000, 0))
				datum.Calibrations.End = dexcomTest.RandomMomentFromTime(time.Unix(1790000000, 0))
				datum.EGVs.Start = dexcomTest.RandomMomentFromTime(time.Unix(1710000000, 0))
				datum.EGVs.End = dexcomTest.RandomMomentFromTime(time.Unix(1780000000, 0))
				datum.Events.Start = dexcomTest.RandomMomentFromTime(time.Unix(1750000000, 0))
				datum.Events.End = dexcomTest.RandomMomentFromTime(time.Unix(1770000000, 0))
				expectedDataRange := &dexcom.DataRange{
					Start: datum.EGVs.Start,
					End:   datum.Calibrations.End,
				}
				Expect(datum.DataRange()).To(Equal(expectedDataRange))
			})
		})
	})

	Context("ParseDataRange", func() {
		It("returns nil if the object is nil", func() {
			parser := structureParser.NewObject(logTest.NewLogger(), nil)
			Expect(dexcom.ParseDataRange(parser)).To(BeNil())
		})

		It("returns the parsed object", func() {
			expectedDatum := dexcomTest.RandomDataRange()
			object := dexcomTest.NewObjectFromDataRange(expectedDatum, test.ObjectFormatJSON)
			parser := structureParser.NewObject(logTest.NewLogger(), &object)
			Expect(dexcom.ParseDataRange(parser)).To(Equal(expectedDatum))
		})
	})

	Context("DataRange", func() {
		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dexcom.DataRange), expectedErrors ...error) {
					expectedDatum := dexcomTest.RandomDataRange()
					object := dexcomTest.NewObjectFromDataRange(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := &dexcom.DataRange{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dexcom.DataRange) {},
				),
				Entry("start invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.DataRange) {
						object["start"] = true
						expectedDatum.Start = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/start"),
				),
				Entry("end invalid type",
					func(object map[string]interface{}, expectedDatum *dexcom.DataRange) {
						object["end"] = true
						expectedDatum.End = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/end"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dexcom.DataRange), expectedErrors ...error) {
					datum := dexcomTest.RandomDataRange()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dexcom.DataRange) {},
				),
				Entry("start missing",
					func(datum *dexcom.DataRange) {
						datum.Start = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/start"),
				),
				Entry("start invalid",
					func(datum *dexcom.DataRange) {
						datum.Start.SystemTime = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/start/systemTime"),
				),
				Entry("end missing",
					func(datum *dexcom.DataRange) {
						datum.End = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/end"),
				),
				Entry("end invalid",
					func(datum *dexcom.DataRange) {
						datum.End.SystemTime = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/end/systemTime"),
				),
				Entry("multiple errors",
					func(datum *dexcom.DataRange) {
						datum.Start = nil
						datum.End = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/start"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/end"),
				),
			)
		})
	})
})
