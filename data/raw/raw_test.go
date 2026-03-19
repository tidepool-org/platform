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
	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/net"
	"github.com/tidepool-org/platform/pointer"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureTest "github.com/tidepool-org/platform/structure/test"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
	"github.com/tidepool-org/platform/user"
)

var _ = Describe("Raw", func() {
	It("FilterCreatedDateFormat is expected", func() {
		Expect(dataRaw.FilterCreatedDateFormat).To(Equal(time.DateOnly))
	})

	It("DataSizeMaximum is expected", func() {
		Expect(dataRaw.DataSizeMaximum).To(Equal(8388608))
	})

	It("MetadataSizeMaximum is expected", func() {
		Expect(dataRaw.MetadataSizeMaximum).To(Equal(4096))
	})

	It("MediaTypeDefault is expected", func() {
		Expect(dataRaw.MediaTypeDefault).To(Equal("application/octet-stream"))
	})

	Context("Filter", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataRaw.Filter)) {
				datum := dataRawTest.RandomFilter(test.AllowOptional())
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataRawTest.NewObjectFromFilter(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataRawTest.NewObjectFromFilter(datum, test.ObjectFormatJSON))
			},
			Entry("succeeds",
				func(datum *dataRaw.Filter) {},
			),
			Entry("empty",
				func(datum *dataRaw.Filter) {
					*datum = dataRaw.Filter{}
				},
			),
			Entry("all",
				func(datum *dataRaw.Filter) {
					*datum = *dataRawTest.RandomFilter()
				},
			),
		)

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]any, expectedDatum *dataRaw.Filter), expectedErrors ...error) {
					expectedDatum := dataRawTest.RandomFilter(test.AllowOptional())
					object := dataRawTest.NewObjectFromFilter(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := &dataRaw.Filter{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]any, expectedDatum *dataRaw.Filter) {},
				),
				Entry("multiple",
					func(object map[string]any, expectedDatum *dataRaw.Filter) {
						object["createdDate"] = true
						object["dataSetId"] = true
						object["processed"] = ""
						object["archivable"] = ""
						object["archived"] = ""
						expectedDatum.CreatedDate = nil
						expectedDatum.DataSetID = nil
						expectedDatum.Processed = nil
						expectedDatum.Archivable = nil
						expectedDatum.Archived = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/createdDate"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/dataSetId"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotBool(""), "/processed"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotBool(""), "/archivable"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotBool(""), "/archived"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataRaw.Filter), expectedErrors ...error) {
					datum := dataRawTest.RandomFilter(test.AllowOptional())
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
				Entry("dataSetId missing",
					func(datum *dataRaw.Filter) { datum.DataSetID = nil },
				),
				Entry("dataSetId empty",
					func(datum *dataRaw.Filter) { datum.DataSetID = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/dataSetId"),
				),
				Entry("dataSetId invalid",
					func(datum *dataRaw.Filter) { datum.DataSetID = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(data.ErrorValueStringAsSetIDNotValid("invalid"), "/dataSetId"),
				),
				Entry("multiple errors",
					func(datum *dataRaw.Filter) {
						datum.CreatedDate = pointer.FromString("invalid")
						datum.DataSetID = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", dataRaw.FilterCreatedDateFormat), "/createdDate"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/dataSetId"),
				),
			)
		})

		Context("with new filter", func() {
			var filter *dataRaw.Filter

			BeforeEach(func() {
				filter = dataRawTest.RandomFilter(test.AllowOptional())
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
					filter.CreatedDate = pointer.FromString(test.RandomTime().Format(time.DateOnly))
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
				datum := dataRawTest.RandomCreate(test.AllowOptional())
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataRawTest.NewObjectFromCreate(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataRawTest.NewObjectFromCreate(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataRaw.Create) {},
			),
			Entry("empty",
				func(datum *dataRaw.Create) {
					*datum = dataRaw.Create{}
				},
			),
			Entry("all",
				func(datum *dataRaw.Create) {
					*datum = *dataRawTest.RandomCreate()
				},
			),
		)

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]any, expectedDatum *dataRaw.Create), expectedErrors ...error) {
					expectedDatum := dataRawTest.RandomCreate(test.AllowOptional())
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
						object["archivableTime"] = true
						expectedDatum.Metadata = nil
						expectedDatum.DigestMD5 = nil
						expectedDatum.MediaType = nil
						expectedDatum.ArchivableTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/metadata"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/digestMD5"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/mediaType"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/archivableTime"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataRaw.Create), expectedErrors ...error) {
					datum := dataRawTest.RandomCreate(test.AllowOptional())
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
						datum.Metadata = map[string]any{"invalid": strings.Repeat("X", dataRaw.MetadataSizeMaximum)}
					},
					errorsTest.WithPointerSource(structureValidator.ErrorSizeNotLessThanOrEqualTo(4110, dataRaw.MetadataSizeMaximum), "/metadata"),
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
				Entry("archivableTime missing",
					func(datum *dataRaw.Create) { datum.ArchivableTime = nil },
				),
				Entry("archivableTime zero",
					func(datum *dataRaw.Create) { datum.ArchivableTime = pointer.FromTime(time.Time{}) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/archivableTime"),
				),
				Entry("multiple errors",
					func(datum *dataRaw.Create) {
						datum.Metadata = map[string]any{"invalid": strings.Repeat("X", dataRaw.MetadataSizeMaximum)}
						datum.DigestMD5 = pointer.FromString("")
						datum.MediaType = pointer.FromString("")
						datum.ArchivableTime = pointer.FromTime(time.Time{})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorSizeNotLessThanOrEqualTo(4110, dataRaw.MetadataSizeMaximum), "/metadata"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/digestMD5"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/mediaType"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/archivableTime"),
				),
			)
		})
	})

	Context("Content", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataRaw.Content)) {
				datum := dataRawTest.RandomContent(test.AllowOptional())
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataRawTest.NewObjectFromContent(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataRawTest.NewObjectFromContent(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataRaw.Content) {},
			),
			Entry("empty",
				func(datum *dataRaw.Content) {
					*datum = dataRaw.Content{}
				},
			),
			Entry("all",
				func(datum *dataRaw.Content) {
					*datum = *dataRawTest.RandomContent()
				},
			),
		)

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataRaw.Content), expectedErrors ...error) {
					datum := dataRawTest.RandomContent(test.AllowOptional())
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
				datum := dataRawTest.RandomUpdate(test.AllowOptional())
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataRawTest.NewObjectFromUpdate(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataRawTest.NewObjectFromUpdate(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataRaw.Update) {},
			),
			Entry("empty",
				func(datum *dataRaw.Update) {
					*datum = dataRaw.Update{}
				},
			),
			Entry("all",
				func(datum *dataRaw.Update) {
					*datum = *dataRawTest.RandomUpdate()
				},
			),
		)

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]any, expectedDatum *dataRaw.Update), expectedErrors ...error) {
					expectedDatum := dataRawTest.RandomUpdate(test.AllowOptional())
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
						object["archivableTime"] = true
						object["archivedTime"] = true
						object["metadata"] = true
						expectedDatum.ProcessedTime = nil
						expectedDatum.ArchivableTime = nil
						expectedDatum.ArchivedTime = nil
						expectedDatum.Metadata = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/processedTime"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/archivableTime"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/archivedTime"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/metadata"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataRaw.Update), expectedErrors ...error) {
					datum := dataRawTest.RandomUpdate(test.AllowOptional())
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataRaw.Update) {},
				),
				Entry("processedTime zero",
					func(datum *dataRaw.Update) { datum.ProcessedTime = pointer.FromTime(time.Time{}) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/processedTime"),
				),
				Entry("processedTime not before now",
					func(datum *dataRaw.Update) {
						datum.ProcessedTime = pointer.FromTime(test.FutureFarTime())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(test.FutureFarTime()), "/processedTime"),
				),
				Entry("archivableTime zero",
					func(datum *dataRaw.Update) { datum.ArchivableTime = pointer.FromTime(time.Time{}) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/archivableTime"),
				),
				Entry("archivedTime zero",
					func(datum *dataRaw.Update) { datum.ArchivedTime = pointer.FromTime(time.Time{}) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/archivedTime"),
				),
				Entry("archivedTime not before now",
					func(datum *dataRaw.Update) {
						datum.ArchivedTime = pointer.FromTime(test.FutureFarTime())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(test.FutureFarTime()), "/archivedTime"),
				),
				Entry("metadata invalid",
					func(datum *dataRaw.Update) {
						datum.Metadata = &map[string]any{"invalid": strings.Repeat("X", dataRaw.MetadataSizeMaximum)}
					},
					errorsTest.WithPointerSource(structureValidator.ErrorSizeNotLessThanOrEqualTo(4110, dataRaw.MetadataSizeMaximum), "/metadata"),
				),
				Entry("all missing",
					func(datum *dataRaw.Update) {
						datum.ProcessedTime = nil
						datum.ArchivableTime = nil
						datum.ArchivedTime = nil
						datum.Metadata = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValuesNotExistForAny("archivableTime", "archivedTime", "metadata", "processedTime"), ""),
				),
				Entry("multiple errors",
					func(datum *dataRaw.Update) {
						datum.ProcessedTime = pointer.FromTime(time.Time{})
						datum.ArchivableTime = pointer.FromTime(time.Time{})
						datum.ArchivedTime = pointer.FromTime(time.Time{})
						datum.Metadata = pointer.FromAny(map[string]any{"invalid": strings.Repeat("X", dataRaw.MetadataSizeMaximum)})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/processedTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/archivableTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/archivedTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorSizeNotLessThanOrEqualTo(4110, dataRaw.MetadataSizeMaximum), "/metadata"),
				),
			)
		})
	})

	Context("Raw", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataRaw.Raw)) {
				datum := dataRawTest.RandomRaw(test.AllowOptional())
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataRawTest.NewObjectFromRaw(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataRawTest.NewObjectFromRaw(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataRaw.Raw) {},
			),
			Entry("empty",
				func(datum *dataRaw.Raw) {
					*datum = dataRaw.Raw{}
				},
			),
			Entry("all",
				func(datum *dataRaw.Raw) {
					*datum = *dataRawTest.RandomRaw()
				},
			),
		)

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]any, expectedDatum *dataRaw.Raw), expectedErrors ...error) {
					expectedDatum := dataRawTest.RandomRaw(test.AllowOptional())
					object := dataRawTest.NewObjectFromRaw(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := &dataRaw.Raw{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]any, expectedDatum *dataRaw.Raw) {},
				),
				Entry("multiple",
					func(object map[string]any, expectedDatum *dataRaw.Raw) {
						object["id"] = true
						object["userId"] = true
						object["dataSetId"] = true
						object["metadata"] = true
						object["digestMD5"] = true
						object["mediaType"] = true
						object["size"] = true
						object["processedTime"] = true
						object["archivableTime"] = true
						object["archivedTime"] = true
						object["createdTime"] = true
						object["modifiedTime"] = true
						object["revision"] = true
						expectedDatum.ID = ""
						expectedDatum.UserID = ""
						expectedDatum.DataSetID = ""
						expectedDatum.Metadata = nil
						expectedDatum.DigestMD5 = ""
						expectedDatum.MediaType = ""
						expectedDatum.Size = 0
						expectedDatum.ProcessedTime = nil
						expectedDatum.ArchivableTime = nil
						expectedDatum.ArchivedTime = nil
						expectedDatum.CreatedTime = time.Time{}
						expectedDatum.ModifiedTime = nil
						expectedDatum.Revision = 0
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/id"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/userId"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/dataSetId"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/metadata"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/digestMD5"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/mediaType"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotInt(true), "/size"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/processedTime"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/archivableTime"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/archivedTime"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/createdTime"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/modifiedTime"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotInt(true), "/revision"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataRaw.Raw), expectedErrors ...error) {
					datum := dataRawTest.RandomRaw(test.AllowOptional())
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataRaw.Raw) {},
				),
				Entry("id missing",
					func(datum *dataRaw.Raw) { datum.ID = "" },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/id"),
				),
				Entry("id invalid",
					func(datum *dataRaw.Raw) { datum.ID = "invalid" },
					errorsTest.WithPointerSource(dataRaw.ErrorValueStringAsDataRawIDNotValid("invalid"), "/id"),
				),
				Entry("userId missing",
					func(datum *dataRaw.Raw) { datum.UserID = "" },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/userId"),
				),
				Entry("userId invalid",
					func(datum *dataRaw.Raw) { datum.UserID = "invalid" },
					errorsTest.WithPointerSource(user.ErrorValueStringAsIDNotValid("invalid"), "/userId"),
				),
				Entry("dataSetId missing",
					func(datum *dataRaw.Raw) { datum.DataSetID = "" },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/dataSetId"),
				),
				Entry("dataSetId invalid",
					func(datum *dataRaw.Raw) { datum.DataSetID = "invalid" },
					errorsTest.WithPointerSource(data.ErrorValueStringAsSetIDNotValid("invalid"), "/dataSetId"),
				),
				Entry("metadata missing",
					func(datum *dataRaw.Raw) { datum.Metadata = nil },
				),
				Entry("metadata invalid",
					func(datum *dataRaw.Raw) {
						datum.Metadata = map[string]any{"invalid": strings.Repeat("X", dataRaw.MetadataSizeMaximum)}
					},
					errorsTest.WithPointerSource(structureValidator.ErrorSizeNotLessThanOrEqualTo(4110, dataRaw.MetadataSizeMaximum), "/metadata"),
				),
				Entry("digestMD5 empty",
					func(datum *dataRaw.Raw) { datum.DigestMD5 = "" },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/digestMD5"),
				),
				Entry("digestMD5 invalid",
					func(datum *dataRaw.Raw) { datum.DigestMD5 = "#" },
					errorsTest.WithPointerSource(net.ErrorValueStringAsDigestMD5NotValid("#"), "/digestMD5"),
				),
				Entry("mediaType empty",
					func(datum *dataRaw.Raw) { datum.MediaType = "" },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/mediaType"),
				),
				Entry("mediaType invalid",
					func(datum *dataRaw.Raw) { datum.MediaType = "/" },
					errorsTest.WithPointerSource(net.ErrorValueStringAsMediaTypeNotValid("/"), "/mediaType"),
				),
				Entry("size invalid",
					func(datum *dataRaw.Raw) { datum.Size = -1 },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "/size"),
				),
				Entry("processedTime before createdTime",
					func(datum *dataRaw.Raw) {
						datum.ProcessedTime = pointer.FromTime(test.PastFarTime())
						datum.ArchivableTime = nil
						datum.ArchivedTime = nil
						datum.CreatedTime = test.PastNearTime()
						datum.ModifiedTime = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotAfter(test.PastFarTime(), test.PastNearTime()), "/processedTime"),
				),
				Entry("processedTime after now",
					func(datum *dataRaw.Raw) {
						datum.ProcessedTime = pointer.FromTime(test.FutureNearTime())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(test.FutureNearTime()), "/processedTime"),
				),
				Entry("archivableTime before createdTime",
					func(datum *dataRaw.Raw) {
						datum.ProcessedTime = nil
						datum.ArchivableTime = pointer.FromTime(test.PastFarTime())
						datum.ArchivedTime = nil
						datum.CreatedTime = test.PastNearTime()
						datum.ModifiedTime = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotAfter(test.PastFarTime(), test.PastNearTime()), "/archivableTime"),
				),
				Entry("archivedTime before createdTime",
					func(datum *dataRaw.Raw) {
						datum.ProcessedTime = nil
						datum.ArchivableTime = nil
						datum.ArchivedTime = pointer.FromTime(test.PastFarTime())
						datum.CreatedTime = test.PastNearTime()
						datum.ModifiedTime = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotAfter(test.PastFarTime(), test.PastNearTime()), "/archivedTime"),
				),
				Entry("archivedTime after now",
					func(datum *dataRaw.Raw) {
						datum.ArchivedTime = pointer.FromTime(test.FutureNearTime())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(test.FutureNearTime()), "/archivedTime"),
				),
				Entry("createdTime zero",
					func(datum *dataRaw.Raw) { datum.CreatedTime = time.Time{} },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/createdTime"),
				),
				Entry("createdTime after now",
					func(datum *dataRaw.Raw) {
						datum.ProcessedTime = nil
						datum.ArchivableTime = nil
						datum.ArchivedTime = nil
						datum.CreatedTime = test.FutureNearTime()
						datum.ModifiedTime = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(test.FutureNearTime()), "/createdTime"),
				),
				Entry("modifiedTime before createdTime",
					func(datum *dataRaw.Raw) {
						datum.ProcessedTime = nil
						datum.ArchivableTime = nil
						datum.ArchivedTime = nil
						datum.CreatedTime = test.PastNearTime()
						datum.ModifiedTime = pointer.FromTime(test.PastFarTime())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotAfter(test.PastFarTime(), test.PastNearTime()), "/modifiedTime"),
				),
				Entry("modifiedTime after now",
					func(datum *dataRaw.Raw) {
						datum.ModifiedTime = pointer.FromTime(test.FutureNearTime())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(test.FutureNearTime()), "/modifiedTime"),
				),
				Entry("revision invalid",
					func(datum *dataRaw.Raw) { datum.Revision = -1 },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "/revision"),
				),
				Entry("multiple errors",
					func(datum *dataRaw.Raw) {
						datum.ID = ""
						datum.UserID = ""
						datum.DataSetID = ""
						datum.Metadata = map[string]any{"invalid": strings.Repeat("X", dataRaw.MetadataSizeMaximum)}
						datum.DigestMD5 = ""
						datum.MediaType = ""
						datum.Size = -1
						datum.ProcessedTime = pointer.FromTime(test.PastNearTime())
						datum.ArchivableTime = pointer.FromTime(test.PastNearTime())
						datum.ArchivedTime = pointer.FromTime(test.PastNearTime())
						datum.CreatedTime = test.FutureNearTime()
						datum.ModifiedTime = pointer.FromTime(test.PastNearTime())
						datum.Revision = -1
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/id"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/userId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/dataSetId"),
					errorsTest.WithPointerSource(structureValidator.ErrorSizeNotLessThanOrEqualTo(4110, dataRaw.MetadataSizeMaximum), "/metadata"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/digestMD5"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/mediaType"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "/size"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotAfter(test.PastNearTime(), test.FutureNearTime()), "/processedTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotAfter(test.PastNearTime(), test.FutureNearTime()), "/archivableTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotAfter(test.PastNearTime(), test.FutureNearTime()), "/archivedTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(test.FutureNearTime()), "/createdTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotAfter(test.PastNearTime(), test.FutureNearTime()), "/modifiedTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "/revision"),
				),
			)
		})

		Context("with new raw", func() {
			var datum *dataRaw.Raw

			BeforeEach(func() {
				datum = dataRawTest.RandomRaw(test.AllowOptional())
			})

			Context("IsProcessed", func() {
				It("returns false if processedTime is nil", func() {
					datum.ProcessedTime = nil
					Expect(datum.IsProcessed()).To(BeFalse())
				})

				It("returns true if processedTime is not nil", func() {
					datum.ProcessedTime = pointer.FromTime(test.RandomTime())
					Expect(datum.IsProcessed()).To(BeTrue())
				})
			})

			Context("IsArchivable", func() {
				It("returns false if archivableTime is nil", func() {
					datum.ArchivableTime = nil
					Expect(datum.IsArchivable()).To(BeFalse())
				})

				It("returns false if archivableTime is after now", func() {
					datum.ArchivableTime = pointer.FromTime(time.Now().Add(time.Minute))
					Expect(datum.IsArchivable()).To(BeFalse())
				})

				It("returns true if archivableTime is before now", func() {
					datum.ArchivableTime = pointer.FromTime(time.Now().Add(-time.Minute))
					Expect(datum.IsArchivable()).To(BeTrue())
				})
			})

			Context("IsArchived", func() {
				It("returns false if archivableTime is nil", func() {
					datum.ArchivedTime = nil
					Expect(datum.IsArchived()).To(BeFalse())
				})

				It("returns true if archivableTime is not nil", func() {
					datum.ArchivedTime = pointer.FromTime(test.RandomTime())
					Expect(datum.IsArchived()).To(BeTrue())
				})
			})
		})
	})

	Context("IsValidDataRawID, DataRawIDValidator, ValidateDataRawID, and ErrorValueStringAsDataRawIDNotValid", func() {
		DescribeTable("return the expected results when the input",
			func(value string, expectedErrors ...error) {
				Expect(dataRaw.IsValidDataRawID(value)).To(Equal(len(expectedErrors) == 0))
				errorReporter := structureTest.NewErrorReporter()
				dataRaw.DataRawIDValidator(value, errorReporter)
				errorsTest.ExpectEqual(errorReporter.Error(), expectedErrors...)
				errorsTest.ExpectEqual(dataRaw.ValidateDataRawID(value), expectedErrors...)
			},
			Entry("is an empty string", "", structureValidator.ErrorValueEmpty()),
			Entry("has id length out of range (lower)", "0123456789abcdef0123456:2026-01-15", dataRaw.ErrorValueStringAsDataRawIDNotValid("0123456789abcdef0123456:2026-01-15")),
			Entry("has id length in range", "0123456789abcdef01234567:2026-01-15"),
			Entry("has id length out of range (upper)", "0123456789abcdef012345678:2026-01-15", dataRaw.ErrorValueStringAsDataRawIDNotValid("0123456789abcdef012345678:2026-01-15")),
			Entry("has id length in range with invalid date", "0123456789abcdef01234567:2026-01-32", dataRaw.ErrorValueStringAsDataRawIDNotValid("0123456789abcdef01234567:2026-01-32")),
			Entry("has uppercase characters", "0123456789ABCDEF01234567:2026-01-15", dataRaw.ErrorValueStringAsDataRawIDNotValid("0123456789ABCDEF01234567:2026-01-15")),
			Entry("has symbols", "0123456789^%$#abcdef0123:2026-01-15", dataRaw.ErrorValueStringAsDataRawIDNotValid("0123456789^%$#abcdef0123:2026-01-15")),
			Entry("has whitespace", "0123456789    abcdef0123:2026-01-15", dataRaw.ErrorValueStringAsDataRawIDNotValid("0123456789    abcdef0123:2026-01-15")),
		)
	})
})
