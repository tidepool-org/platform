package raw_test

import (
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"github.com/tidepool-org/platform/data"
	dataRaw "github.com/tidepool-org/platform/data/raw"
	dataRawTest "github.com/tidepool-org/platform/data/raw/test"
	dataTest "github.com/tidepool-org/platform/data/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/net"
	"github.com/tidepool-org/platform/pointer"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Raw", func() {
	It("FilterCreatedDateFormat is expected", func() {
		Expect(dataRaw.FilterCreatedDateFormat).To(Equal(time.DateOnly))
	})

	It("FilterDataSetIDsLengthMaximum is expected", func() {
		Expect(dataRaw.FilterDataSetIDsLengthMaximum).To(Equal(100))
	})

	It("DataSizeMaximum is expected", func() {
		Expect(dataRaw.DataSizeMaximum).To(Equal(8388608))
	})

	It("MetadataLengthMaximum is expected", func() {
		Expect(dataRaw.MetadataLengthMaximum).To(Equal(4096))
	})

	It("MediaTypeDefault is expected", func() {
		Expect(dataRaw.MediaTypeDefault).To(Equal("application/octet-stream"))
	})

	Context("Filter", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataRaw.Filter)) {
				datum := dataRawTest.RandomFilter()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataRawTest.NewObjectFromFilter(datum, test.ObjectFormatJSON))
			},
			Entry("succeeds",
				func(datum *dataRaw.Filter) {},
			),
			Entry("empty",
				func(datum *dataRaw.Filter) {
					*datum = dataRaw.Filter{}
				},
			),
		)

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]any, expectedDatum *dataRaw.Filter), expectedErrors ...error) {
					expectedDatum := dataRawTest.RandomFilter()
					dataRawTest.RandomFilter()
					object := dataRawTest.NewObjectFromFilter(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := &dataRaw.Filter{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]any, expectedDatum *dataRaw.Filter) {},
				),
				Entry("dataSetIds invalid element type",
					func(object map[string]any, expectedDatum *dataRaw.Filter) {
						object["dataSetIds"] = []any{true}
						expectedDatum.DataSetIDs = pointer.FromStringArray([]string{""})
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/dataSetIds/0"),
				),
				Entry("multiple",
					func(object map[string]any, expectedDatum *dataRaw.Filter) {
						object["createdDate"] = true
						object["dataSetIds"] = true
						object["processed"] = ""
						expectedDatum.CreatedDate = nil
						expectedDatum.DataSetIDs = nil
						expectedDatum.Processed = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/createdDate"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/dataSetIds"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotBool(""), "/processed"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataRaw.Filter), expectedErrors ...error) {
					datum := dataRawTest.RandomFilter()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataRaw.Filter) {},
				),
				Entry("createdDate missing",
					func(datum *dataRaw.Filter) { datum.CreatedDate = nil },
				),
				Entry("createdDate invalid",
					func(datum *dataRaw.Filter) {
						datum.CreatedDate = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", dataRaw.FilterCreatedDateFormat), "/createdDate"),
				),
				Entry("createdDate zero",
					func(datum *dataRaw.Filter) {
						datum.CreatedDate = pointer.FromString(time.Time{}.Format(dataRaw.FilterCreatedDateFormat))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/createdDate"),
				),
				Entry("dataSetIds missing",
					func(datum *dataRaw.Filter) { datum.DataSetIDs = nil },
				),
				Entry("dataSetIds empty",
					func(datum *dataRaw.Filter) { datum.DataSetIDs = pointer.FromStringArray([]string{}) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/dataSetIds"),
				),
				Entry("dataSetIds too many elements",
					func(datum *dataRaw.Filter) {
						datum.DataSetIDs = pointer.FromStringArray(test.RandomStringArrayFromRangeAndGeneratorWithoutDuplicates(dataRaw.FilterDataSetIDsLengthMaximum+1, dataRaw.FilterDataSetIDsLengthMaximum+1, dataTest.RandomDataSetID))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(dataRaw.FilterDataSetIDsLengthMaximum+1, dataRaw.FilterDataSetIDsLengthMaximum), "/dataSetIds"),
				),
				Entry("dataSetIds invalid element",
					func(datum *dataRaw.Filter) { datum.DataSetIDs = pointer.FromStringArray([]string{"invalid"}) },
					errorsTest.WithPointerSource(data.ErrorValueStringAsSetIDNotValid("invalid"), "/dataSetIds/0"),
				),
				Entry("dataSetIds duplicate element",
					func(datum *dataRaw.Filter) {
						dataSetID := dataTest.RandomDataSetID()
						datum.DataSetIDs = pointer.FromStringArray([]string{dataSetID, dataSetID})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueDuplicate(), "/dataSetIds/1"),
				),
				Entry("multiple errors",
					func(datum *dataRaw.Filter) {
						datum.CreatedDate = pointer.FromString("invalid")
						datum.DataSetIDs = pointer.FromStringArray([]string{})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", dataRaw.FilterCreatedDateFormat), "/createdDate"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/dataSetIds"),
				),
			)
		})

		Context("with new filter", func() {
			var filter *dataRaw.Filter

			BeforeEach(func() {
				filter = dataRawTest.RandomFilter()
			})

			Context("CreatedTime", func() {
				It("returns nil if createdDate is nil", func() {
					filter.CreatedDate = nil
					Expect(filter.CreatedTime()).To(BeNil())
				})

				It("returns nil if createdDate is invalid", func() {
					filter.CreatedDate = pointer.FromString("invalid")
					Expect(filter.CreatedTime()).To(BeNil())
				})

				It("returns date-only time", func() {
					expectedCreatedTime, err := time.Parse(dataRaw.FilterCreatedDateFormat, *filter.CreatedDate)
					Expect(err).ToNot(HaveOccurred())
					Expect(filter.CreatedTime()).To(PointTo(Equal(expectedCreatedTime)))
				})
			})
		})
	})

	Context("Create", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataRaw.Create)) {
				datum := dataRawTest.RandomCreate()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataRawTest.NewObjectFromCreate(datum, test.ObjectFormatJSON))
			},
			Entry("succeeds",
				func(datum *dataRaw.Create) {},
			),
			Entry("empty",
				func(datum *dataRaw.Create) {
					*datum = dataRaw.Create{}
				},
			),
		)

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]any, expectedDatum *dataRaw.Create), expectedErrors ...error) {
					expectedDatum := dataRawTest.RandomCreate()
					dataRawTest.RandomCreate()
					object := dataRawTest.NewObjectFromCreate(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := &dataRaw.Create{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]any, expectedDatum *dataRaw.Create) {},
				),
				Entry("metadata nil",
					func(object map[string]any, expectedDatum *dataRaw.Create) {
						object["metadata"] = nil
						expectedDatum.Metadata = nil
					},
				),
				Entry("multiple",
					func(object map[string]any, expectedDatum *dataRaw.Create) {
						object["metadata"] = true
						object["digestMD5"] = true
						object["mediaType"] = true
						expectedDatum.Metadata = nil
						expectedDatum.DigestMD5 = nil
						expectedDatum.MediaType = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/metadata"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/digestMD5"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/mediaType"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataRaw.Create), expectedErrors ...error) {
					datum := dataRawTest.RandomCreate()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataRaw.Create) {},
				),
				Entry("metadata missing",
					func(datum *dataRaw.Create) { datum.Metadata = nil },
				),
				Entry("metadata invalid",
					func(datum *dataRaw.Create) {
						datum.Metadata = map[string]any{"invalid": strings.Repeat("X", dataRaw.MetadataLengthMaximum)}
					},
					errorsTest.WithPointerSource(structureValidator.ErrorSizeNotLessThanOrEqualTo(4110, dataRaw.MetadataLengthMaximum), "/metadata"),
				),
				Entry("digestMD5 missing",
					func(datum *dataRaw.Create) { datum.DigestMD5 = nil },
				),
				Entry("digestMD5 empty",
					func(datum *dataRaw.Create) { datum.DigestMD5 = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/digestMD5"),
				),
				Entry("digestMD5 invalid",
					func(datum *dataRaw.Create) { datum.DigestMD5 = pointer.FromString("#") },
					errorsTest.WithPointerSource(net.ErrorValueStringAsDigestMD5NotValid("#"), "/digestMD5"),
				),
				Entry("mediaType missing",
					func(datum *dataRaw.Create) { datum.MediaType = nil },
				),
				Entry("mediaType empty",
					func(datum *dataRaw.Create) { datum.MediaType = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/mediaType"),
				),
				Entry("mediaType invalid",
					func(datum *dataRaw.Create) { datum.MediaType = pointer.FromString("/") },
					errorsTest.WithPointerSource(net.ErrorValueStringAsMediaTypeNotValid("/"), "/mediaType"),
				),
				Entry("multiple errors",
					func(datum *dataRaw.Create) {
						datum.Metadata = map[string]any{"invalid": strings.Repeat("X", dataRaw.MetadataLengthMaximum)}
						datum.DigestMD5 = pointer.FromString("")
						datum.MediaType = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorSizeNotLessThanOrEqualTo(4110, dataRaw.MetadataLengthMaximum), "/metadata"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/digestMD5"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/mediaType"),
				),
			)
		})
	})

	Context("Content", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataRaw.Content)) {
				datum := dataRawTest.RandomContent()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataRawTest.NewObjectFromContent(datum, test.ObjectFormatJSON))
			},
			Entry("succeeds",
				func(datum *dataRaw.Content) {},
			),
			Entry("empty",
				func(datum *dataRaw.Content) {
					*datum = dataRaw.Content{}
				},
			),
		)

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataRaw.Content), expectedErrors ...error) {
					datum := dataRawTest.RandomContent()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("digestMD5 empty",
					func(datum *dataRaw.Content) { datum.DigestMD5 = "" },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/digestMD5"),
				),
				Entry("digestMD5 invalid",
					func(datum *dataRaw.Content) { datum.DigestMD5 = "#" },
					errorsTest.WithPointerSource(net.ErrorValueStringAsDigestMD5NotValid("#"), "/digestMD5"),
				),
				Entry("mediaType empty",
					func(datum *dataRaw.Content) { datum.MediaType = "" },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/mediaType"),
				),
				Entry("mediaType invalid",
					func(datum *dataRaw.Content) { datum.MediaType = "/" },
					errorsTest.WithPointerSource(net.ErrorValueStringAsMediaTypeNotValid("/"), "/mediaType"),
				),
				Entry("readCloser missing",
					func(datum *dataRaw.Content) { datum.ReadCloser = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/readCloser"),
				),
				Entry("multiple errors",
					func(datum *dataRaw.Content) {
						datum.DigestMD5 = ""
						datum.MediaType = ""
						datum.ReadCloser = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/digestMD5"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/mediaType"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/readCloser"),
				),
			)
		})
	})

	Context("Update", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataRaw.Update)) {
				datum := dataRawTest.RandomUpdate()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataRawTest.NewObjectFromUpdate(datum, test.ObjectFormatJSON))
			},
			Entry("succeeds",
				func(datum *dataRaw.Update) {},
			),
			Entry("empty",
				func(datum *dataRaw.Update) {
					*datum = dataRaw.Update{}
				},
			),
		)

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]any, expectedDatum *dataRaw.Update), expectedErrors ...error) {
					expectedDatum := dataRawTest.RandomUpdate()
					dataRawTest.RandomUpdate()
					object := dataRawTest.NewObjectFromUpdate(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := &dataRaw.Update{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]any, expectedDatum *dataRaw.Update) {},
				),
				Entry("processedTime invalid type",
					func(object map[string]any, expectedDatum *dataRaw.Update) {
						object["processedTime"] = true
						expectedDatum.ProcessedTime = time.Time{}
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/processedTime"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataRaw.Update), expectedErrors ...error) {
					datum := dataRawTest.RandomUpdate()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataRaw.Update) {},
				),
				Entry("processedTime zero",
					func(datum *dataRaw.Update) { datum.ProcessedTime = time.Time{} },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/processedTime"),
				),
				Entry("processedTime not before now",
					func(datum *dataRaw.Update) {
						datum.ProcessedTime = test.FutureFarTime()
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(test.FutureFarTime()), "/processedTime"),
				),
				Entry("multiple errors",
					func(datum *dataRaw.Update) {
						datum.ProcessedTime = time.Time{}
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/processedTime"),
				),
			)
		})
	})
})
