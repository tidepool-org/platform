package blob_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/blob"
	blobTest "github.com/tidepool-org/platform/blob/test"
	"github.com/tidepool-org/platform/crypto"
	cryptoTest "github.com/tidepool-org/platform/crypto/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/net"
	netTest "github.com/tidepool-org/platform/net/test"
	"github.com/tidepool-org/platform/pointer"
	requestTest "github.com/tidepool-org/platform/request/test"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureTest "github.com/tidepool-org/platform/structure/test"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
	testHttp "github.com/tidepool-org/platform/test/http"
	"github.com/tidepool-org/platform/user"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("Blob", func() {
	It("SizeMaximum is expected", func() {
		Expect(blob.SizeMaximum).To(Equal(104857600))
	})

	It("StatusAvailable is expected", func() {
		Expect(blob.StatusAvailable).To(Equal("available"))
	})

	It("StatusCreated is expected", func() {
		Expect(blob.StatusCreated).To(Equal("created"))
	})

	It("Statuses returns expected", func() {
		Expect(blob.Statuses()).To(Equal([]string{"available", "created"}))
	})

	Context("NewFilter", func() {
		It("returns successfully with default values", func() {
			Expect(blob.NewFilter()).To(Equal(&blob.Filter{}))
		})
	})

	Context("Filter", func() {
		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *blob.Filter), expectedErrors ...error) {
					expectedDatum := blobTest.RandomFilter()
					object := blobTest.NewObjectFromFilter(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := &blob.Filter{}
					errorsTest.ExpectEqual(structureParser.NewObject(&object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *blob.Filter) {},
				),
				Entry("media type missing",
					func(object map[string]interface{}, expectedDatum *blob.Filter) {
						delete(object, "mediaType")
						expectedDatum.MediaType = nil
					},
				),
				Entry("media type invalid type",
					func(object map[string]interface{}, expectedDatum *blob.Filter) {
						object["mediaType"] = true
						expectedDatum.MediaType = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/mediaType"),
				),
				Entry("media type empty",
					func(object map[string]interface{}, expectedDatum *blob.Filter) {
						object["mediaType"] = []string{}
						expectedDatum.MediaType = pointer.FromStringArray([]string{})
					},
				),
				Entry("media type valid",
					func(object map[string]interface{}, expectedDatum *blob.Filter) {
						valid := netTest.RandomMediaTypes(1, 3)
						object["mediaType"] = valid
						expectedDatum.MediaType = pointer.FromStringArray(valid)
					},
				),
				Entry("status missing",
					func(object map[string]interface{}, expectedDatum *blob.Filter) {
						delete(object, "status")
						expectedDatum.Status = nil
					},
				),
				Entry("status invalid type",
					func(object map[string]interface{}, expectedDatum *blob.Filter) {
						object["status"] = true
						expectedDatum.Status = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/status"),
				),
				Entry("status empty",
					func(object map[string]interface{}, expectedDatum *blob.Filter) {
						object["status"] = []string{}
						expectedDatum.Status = pointer.FromStringArray([]string{})
					},
				),
				Entry("status valid",
					func(object map[string]interface{}, expectedDatum *blob.Filter) {
						valid := blobTest.RandomStatuses()
						object["status"] = valid
						expectedDatum.Status = pointer.FromStringArray(valid)
					},
				),
				Entry("multiple",
					func(object map[string]interface{}, expectedDatum *blob.Filter) {
						object["mediaType"] = true
						object["status"] = true
						expectedDatum.MediaType = nil
						expectedDatum.Status = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/mediaType"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/status"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *blob.Filter), expectedErrors ...error) {
					datum := blobTest.RandomFilter()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New().Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *blob.Filter) {},
				),
				Entry("media type missing",
					func(datum *blob.Filter) { datum.MediaType = nil },
				),
				Entry("media type empty",
					func(datum *blob.Filter) {
						datum.MediaType = pointer.FromStringArray([]string{})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/mediaType"),
				),
				Entry("media type element empty",
					func(datum *blob.Filter) {
						datum.MediaType = pointer.FromStringArray([]string{netTest.RandomMediaType(), ""})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/mediaType/1"),
				),
				Entry("media type element invalid",
					func(datum *blob.Filter) {
						datum.MediaType = pointer.FromStringArray([]string{netTest.RandomMediaType(), "/"})
					},
					errorsTest.WithPointerSource(net.ErrorValueStringAsMediaTypeNotValid("/"), "/mediaType/1"),
				),
				Entry("media type element duplicate",
					func(datum *blob.Filter) {
						mediaType := netTest.RandomMediaType()
						datum.MediaType = pointer.FromStringArray([]string{mediaType, mediaType})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueDuplicate(), "/mediaType/1"),
				),
				Entry("media type valid",
					func(datum *blob.Filter) { datum.MediaType = pointer.FromStringArray(netTest.RandomMediaTypes(1, 3)) },
				),
				Entry("status missing",
					func(datum *blob.Filter) { datum.Status = nil },
				),
				Entry("status empty",
					func(datum *blob.Filter) { datum.Status = pointer.FromStringArray([]string{}) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/status"),
				),
				Entry("status element empty",
					func(datum *blob.Filter) { datum.Status = pointer.FromStringArray([]string{"created", ""}) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", blob.Statuses()), "/status/1"),
				),
				Entry("status element invalid",
					func(datum *blob.Filter) { datum.Status = pointer.FromStringArray([]string{"created", "invalid"}) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", blob.Statuses()), "/status/1"),
				),
				Entry("status element duplicate",
					func(datum *blob.Filter) { datum.Status = pointer.FromStringArray([]string{"created", "created"}) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueDuplicate(), "/status/1"),
				),
				Entry("status available",
					func(datum *blob.Filter) { datum.Status = pointer.FromStringArray([]string{"available"}) },
				),
				Entry("status created",
					func(datum *blob.Filter) { datum.Status = pointer.FromStringArray([]string{"created"}) },
				),
				Entry("status available and created",
					func(datum *blob.Filter) { datum.Status = pointer.FromStringArray([]string{"available", "created"}) },
				),
				Entry("multiple errors",
					func(datum *blob.Filter) {
						datum.MediaType = pointer.FromStringArray([]string{})
						datum.Status = pointer.FromStringArray([]string{})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/mediaType"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/status"),
				),
			)
		})

		Context("with new filter", func() {
			var filter *blob.Filter

			BeforeEach(func() {
				filter = blobTest.RandomFilter()
			})

			Context("MutateRequest", func() {
				var req *http.Request

				BeforeEach(func() {
					req = testHttp.NewRequest()
				})

				It("returns an error when the request is missing", func() {
					errorsTest.ExpectEqual(filter.MutateRequest(nil), errors.New("request is missing"))
				})

				It("sets request query as expected", func() {
					Expect(filter.MutateRequest(req)).To(Succeed())
					Expect(req.URL.Query()).To(Equal(url.Values{
						"mediaType": *filter.MediaType,
						"status":    *filter.Status,
					}))
				})

				It("does not set request query when the filter is empty", func() {
					filter.MediaType = nil
					filter.Status = nil
					Expect(filter.MutateRequest(req)).To(Succeed())
					Expect(req.URL.Query()).To(BeEmpty())
				})
			})
		})
	})

	Context("NewContent", func() {
		It("returns successfully with default values", func() {
			Expect(blob.NewContent()).To(Equal(&blob.Content{}))
		})
	})

	Context("Content", func() {
		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *blob.Content), expectedErrors ...error) {
					datum := blobTest.RandomContent()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New().Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *blob.Content) {},
				),
				Entry("body missing",
					func(datum *blob.Content) { datum.Body = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/body"),
				),
				Entry("body valid",
					func(datum *blob.Content) { datum.Body = ioutil.NopCloser(bytes.NewReader(test.RandomBytes())) },
				),
				Entry("digest MD5 missing",
					func(datum *blob.Content) { datum.DigestMD5 = nil },
				),
				Entry("digest MD5 empty",
					func(datum *blob.Content) { datum.DigestMD5 = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/digestMD5"),
				),
				Entry("digest MD5 invalid",
					func(datum *blob.Content) { datum.DigestMD5 = pointer.FromString("#") },
					errorsTest.WithPointerSource(crypto.ErrorValueStringAsBase64EncodedMD5HashNotValid("#"), "/digestMD5"),
				),
				Entry("digest MD5 valid",
					func(datum *blob.Content) {
						datum.DigestMD5 = pointer.FromString(cryptoTest.RandomBase64EncodedMD5Hash())
					},
				),
				Entry("media type missing",
					func(datum *blob.Content) { datum.MediaType = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/mediaType"),
				),
				Entry("media type empty",
					func(datum *blob.Content) { datum.MediaType = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/mediaType"),
				),
				Entry("media type invalid",
					func(datum *blob.Content) { datum.MediaType = pointer.FromString("/") },
					errorsTest.WithPointerSource(net.ErrorValueStringAsMediaTypeNotValid("/"), "/mediaType"),
				),
				Entry("media type valid",
					func(datum *blob.Content) { datum.MediaType = pointer.FromString(netTest.RandomMediaType()) },
				),
				Entry("multiple errors",
					func(datum *blob.Content) {
						datum.Body = nil
						datum.DigestMD5 = pointer.FromString("")
						datum.MediaType = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/body"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/digestMD5"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/mediaType"),
				),
			)
		})
	})

	Context("Blob", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *blob.Blob)) {
				datum := blobTest.RandomBlob()
				mutator(datum)
				test.ExpectSerializedBSON(datum, blobTest.NewObjectFromBlob(datum, test.ObjectFormatBSON))
				test.ExpectSerializedJSON(datum, blobTest.NewObjectFromBlob(datum, test.ObjectFormatJSON))
			},
			Entry("succeeds",
				func(datum *blob.Blob) {},
			),
			Entry("empty",
				func(datum *blob.Blob) { *datum = blob.Blob{} },
			),
		)

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *blob.Blob), expectedErrors ...error) {
					expectedDatum := blobTest.RandomBlob()
					object := blobTest.NewObjectFromBlob(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := &blob.Blob{}
					errorsTest.ExpectEqual(structureParser.NewObject(&object).Parse(datum), expectedErrors...)
					Expect(datum).To(blobTest.MatchBlob(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *blob.Blob) {},
				),
				Entry("id missing",
					func(object map[string]interface{}, expectedDatum *blob.Blob) {
						delete(object, "id")
						expectedDatum.ID = nil
					},
				),
				Entry("id invalid type",
					func(object map[string]interface{}, expectedDatum *blob.Blob) {
						object["id"] = true
						expectedDatum.ID = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/id"),
				),
				Entry("id empty",
					func(object map[string]interface{}, expectedDatum *blob.Blob) {
						object["id"] = ""
						expectedDatum.ID = pointer.FromString("")
					},
				),
				Entry("id valid",
					func(object map[string]interface{}, expectedDatum *blob.Blob) {
						valid := blobTest.RandomID()
						object["id"] = valid
						expectedDatum.ID = pointer.FromString(valid)
					},
				),
				Entry("user id missing",
					func(object map[string]interface{}, expectedDatum *blob.Blob) {
						delete(object, "userId")
						expectedDatum.UserID = nil
					},
				),
				Entry("user id invalid type",
					func(object map[string]interface{}, expectedDatum *blob.Blob) {
						object["userId"] = true
						expectedDatum.UserID = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/userId"),
				),
				Entry("user id empty",
					func(object map[string]interface{}, expectedDatum *blob.Blob) {
						object["userId"] = ""
						expectedDatum.UserID = pointer.FromString("")
					},
				),
				Entry("user id valid",
					func(object map[string]interface{}, expectedDatum *blob.Blob) {
						valid := userTest.RandomID()
						object["userId"] = valid
						expectedDatum.UserID = pointer.FromString(valid)
					},
				),
				Entry("digest MD5 missing",
					func(object map[string]interface{}, expectedDatum *blob.Blob) {
						delete(object, "digestMD5")
						expectedDatum.DigestMD5 = nil
					},
				),
				Entry("digest MD5 invalid type",
					func(object map[string]interface{}, expectedDatum *blob.Blob) {
						object["digestMD5"] = true
						expectedDatum.DigestMD5 = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/digestMD5"),
				),
				Entry("digest MD5 empty",
					func(object map[string]interface{}, expectedDatum *blob.Blob) {
						object["digestMD5"] = ""
						expectedDatum.DigestMD5 = pointer.FromString("")
					},
				),
				Entry("digest MD5 valid",
					func(object map[string]interface{}, expectedDatum *blob.Blob) {
						valid := cryptoTest.RandomBase64EncodedMD5Hash()
						object["digestMD5"] = valid
						expectedDatum.DigestMD5 = pointer.FromString(valid)
					},
				),
				Entry("media type missing",
					func(object map[string]interface{}, expectedDatum *blob.Blob) {
						delete(object, "mediaType")
						expectedDatum.MediaType = nil
					},
				),
				Entry("media type invalid type",
					func(object map[string]interface{}, expectedDatum *blob.Blob) {
						object["mediaType"] = true
						expectedDatum.MediaType = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/mediaType"),
				),
				Entry("media type empty",
					func(object map[string]interface{}, expectedDatum *blob.Blob) {
						object["mediaType"] = ""
						expectedDatum.MediaType = pointer.FromString("")
					},
				),
				Entry("media type valid",
					func(object map[string]interface{}, expectedDatum *blob.Blob) {
						valid := netTest.RandomMediaType()
						object["mediaType"] = valid
						expectedDatum.MediaType = pointer.FromString(valid)
					},
				),
				Entry("size missing",
					func(object map[string]interface{}, expectedDatum *blob.Blob) {
						delete(object, "size")
						expectedDatum.Size = nil
					},
				),
				Entry("size invalid type",
					func(object map[string]interface{}, expectedDatum *blob.Blob) {
						object["size"] = true
						expectedDatum.Size = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotInt(true), "/size"),
				),
				Entry("size valid",
					func(object map[string]interface{}, expectedDatum *blob.Blob) {
						valid := test.RandomIntFromRange(1, 100*1024*1024)
						object["size"] = valid
						expectedDatum.Size = pointer.FromInt(valid)
					},
				),
				Entry("status missing",
					func(object map[string]interface{}, expectedDatum *blob.Blob) {
						delete(object, "status")
						expectedDatum.Status = nil
					},
				),
				Entry("status invalid type",
					func(object map[string]interface{}, expectedDatum *blob.Blob) {
						object["status"] = true
						expectedDatum.Status = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/status"),
				),
				Entry("status empty",
					func(object map[string]interface{}, expectedDatum *blob.Blob) {
						object["status"] = ""
						expectedDatum.Status = pointer.FromString("")
					},
				),
				Entry("status valid",
					func(object map[string]interface{}, expectedDatum *blob.Blob) {
						valid := test.RandomStringFromArray(blob.Statuses())
						object["status"] = valid
						expectedDatum.Status = pointer.FromString(valid)
					},
				),
				Entry("created time missing",
					func(object map[string]interface{}, expectedDatum *blob.Blob) {
						delete(object, "createdTime")
						expectedDatum.CreatedTime = nil
					},
				),
				Entry("created time invalid type",
					func(object map[string]interface{}, expectedDatum *blob.Blob) {
						object["createdTime"] = true
						expectedDatum.CreatedTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/createdTime"),
				),
				Entry("created time invalid",
					func(object map[string]interface{}, expectedDatum *blob.Blob) {
						object["createdTime"] = "invalid"
						expectedDatum.CreatedTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorValueTimeNotParsable("invalid", time.RFC3339Nano), "/createdTime"),
				),
				Entry("created time valid",
					func(object map[string]interface{}, expectedDatum *blob.Blob) {
						valid := test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now())
						object["createdTime"] = valid.Format(time.RFC3339Nano)
						expectedDatum.CreatedTime = pointer.FromTime(valid)
					},
				),
				Entry("modified time missing",
					func(object map[string]interface{}, expectedDatum *blob.Blob) {
						delete(object, "modifiedTime")
						expectedDatum.ModifiedTime = nil
					},
				),
				Entry("modified time invalid type",
					func(object map[string]interface{}, expectedDatum *blob.Blob) {
						object["modifiedTime"] = true
						expectedDatum.ModifiedTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/modifiedTime"),
				),
				Entry("modified time invalid",
					func(object map[string]interface{}, expectedDatum *blob.Blob) {
						object["modifiedTime"] = "invalid"
						expectedDatum.ModifiedTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorValueTimeNotParsable("invalid", time.RFC3339Nano), "/modifiedTime"),
				),
				Entry("modified time valid",
					func(object map[string]interface{}, expectedDatum *blob.Blob) {
						valid := test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now())
						object["modifiedTime"] = valid.Format(time.RFC3339Nano)
						expectedDatum.ModifiedTime = pointer.FromTime(valid)
					},
				),
				Entry("revision missing",
					func(object map[string]interface{}, expectedDatum *blob.Blob) {
						delete(object, "revision")
						expectedDatum.Revision = nil
					},
				),
				Entry("revision invalid type",
					func(object map[string]interface{}, expectedDatum *blob.Blob) {
						object["revision"] = true
						expectedDatum.Revision = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotInt(true), "/revision"),
				),
				Entry("revision valid",
					func(object map[string]interface{}, expectedDatum *blob.Blob) {
						valid := requestTest.RandomRevision()
						object["revision"] = valid
						expectedDatum.Revision = pointer.FromInt(valid)
					},
				),
				Entry("multiple",
					func(object map[string]interface{}, expectedDatum *blob.Blob) {
						object["id"] = true
						object["userId"] = true
						object["digestMD5"] = true
						object["mediaType"] = true
						object["size"] = true
						object["status"] = true
						object["createdTime"] = true
						object["modifiedTime"] = true
						object["revision"] = true
						expectedDatum.ID = nil
						expectedDatum.UserID = nil
						expectedDatum.DigestMD5 = nil
						expectedDatum.MediaType = nil
						expectedDatum.Size = nil
						expectedDatum.Status = nil
						expectedDatum.CreatedTime = nil
						expectedDatum.ModifiedTime = nil
						expectedDatum.Revision = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/id"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/userId"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/digestMD5"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/mediaType"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotInt(true), "/size"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/status"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/createdTime"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/modifiedTime"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotInt(true), "/revision"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *blob.Blob), expectedErrors ...error) {
					datum := blobTest.RandomBlob()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New().Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *blob.Blob) {},
				),
				Entry("id missing",
					func(datum *blob.Blob) { datum.ID = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/id"),
				),
				Entry("id empty",
					func(datum *blob.Blob) { datum.ID = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/id"),
				),
				Entry("id invalid",
					func(datum *blob.Blob) { datum.ID = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(blob.ErrorValueStringAsIDNotValid("invalid"), "/id"),
				),
				Entry("id valid",
					func(datum *blob.Blob) { datum.ID = pointer.FromString(blobTest.RandomID()) },
				),
				Entry("user id missing",
					func(datum *blob.Blob) { datum.UserID = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/userId"),
				),
				Entry("user id empty",
					func(datum *blob.Blob) { datum.UserID = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/userId"),
				),
				Entry("user id invalid",
					func(datum *blob.Blob) { datum.UserID = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(user.ErrorValueStringAsIDNotValid("invalid"), "/userId"),
				),
				Entry("user id valid",
					func(datum *blob.Blob) { datum.UserID = pointer.FromString(userTest.RandomID()) },
				),
				Entry("digest MD5 missing",
					func(datum *blob.Blob) { datum.DigestMD5 = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/digestMD5"),
				),
				Entry("digest MD5 empty",
					func(datum *blob.Blob) { datum.DigestMD5 = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/digestMD5"),
				),
				Entry("digest MD5 invalid",
					func(datum *blob.Blob) { datum.DigestMD5 = pointer.FromString("#") },
					errorsTest.WithPointerSource(crypto.ErrorValueStringAsBase64EncodedMD5HashNotValid("#"), "/digestMD5"),
				),
				Entry("digest MD5 valid",
					func(datum *blob.Blob) { datum.DigestMD5 = pointer.FromString(cryptoTest.RandomBase64EncodedMD5Hash()) },
				),
				Entry("media type missing",
					func(datum *blob.Blob) { datum.MediaType = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/mediaType"),
				),
				Entry("media type empty",
					func(datum *blob.Blob) { datum.MediaType = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/mediaType"),
				),
				Entry("media type invalid",
					func(datum *blob.Blob) { datum.MediaType = pointer.FromString("/") },
					errorsTest.WithPointerSource(net.ErrorValueStringAsMediaTypeNotValid("/"), "/mediaType"),
				),
				Entry("media type valid",
					func(datum *blob.Blob) { datum.MediaType = pointer.FromString(netTest.RandomMediaType()) },
				),
				Entry("size missing",
					func(datum *blob.Blob) { datum.Size = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/size"),
				),
				Entry("size out of range (lower)",
					func(datum *blob.Blob) { datum.Size = pointer.FromInt(-1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "/size"),
				),
				Entry("size in range (lower)",
					func(datum *blob.Blob) { datum.Size = pointer.FromInt(0) },
				),
				Entry("status missing",
					func(datum *blob.Blob) { datum.Status = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/status"),
				),
				Entry("status empty",
					func(datum *blob.Blob) { datum.Status = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", blob.Statuses()), "/status"),
				),
				Entry("status invalid",
					func(datum *blob.Blob) { datum.Status = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", blob.Statuses()), "/status"),
				),
				Entry("status created",
					func(datum *blob.Blob) { datum.Status = pointer.FromString("created") },
				),
				Entry("status available",
					func(datum *blob.Blob) { datum.Status = pointer.FromString("available") },
				),
				Entry("created time missing",
					func(datum *blob.Blob) { datum.CreatedTime = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/createdTime"),
				),
				Entry("created time zero",
					func(datum *blob.Blob) { datum.CreatedTime = pointer.FromTime(time.Time{}) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/createdTime"),
				),
				Entry("created time after now",
					func(datum *blob.Blob) {
						datum.CreatedTime = pointer.FromTime(test.FutureFarTime())
						datum.ModifiedTime = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(test.FutureFarTime()), "/createdTime"),
				),
				Entry("created time valid",
					func(datum *blob.Blob) {
						datum.CreatedTime = pointer.FromTime(test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now()))
						datum.ModifiedTime = nil
					},
				),
				Entry("modified time missing",
					func(datum *blob.Blob) { datum.ModifiedTime = nil },
				),
				Entry("modified time before created time",
					func(datum *blob.Blob) {
						datum.CreatedTime = pointer.FromTime(test.PastNearTime())
						datum.ModifiedTime = pointer.FromTime(test.PastFarTime())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotAfter(test.PastFarTime(), test.PastNearTime()), "/modifiedTime"),
				),
				Entry("modified time after now",
					func(datum *blob.Blob) { datum.ModifiedTime = pointer.FromTime(test.FutureFarTime()) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(test.FutureFarTime()), "/modifiedTime"),
				),
				Entry("modified time valid",
					func(datum *blob.Blob) {
						datum.ModifiedTime = pointer.FromTime(test.RandomTimeFromRange(*datum.CreatedTime, time.Now()))
					},
				),
				Entry("revision missing",
					func(datum *blob.Blob) {
						datum.Revision = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/revision"),
				),
				Entry("revision out of range (lower)",
					func(datum *blob.Blob) {
						datum.Revision = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "/revision"),
				),
				Entry("revision in range (lower)",
					func(datum *blob.Blob) {
						datum.Revision = pointer.FromInt(0)
					},
				),
				Entry("multiple errors",
					func(datum *blob.Blob) {
						datum.ID = nil
						datum.UserID = nil
						datum.DigestMD5 = nil
						datum.MediaType = nil
						datum.Size = nil
						datum.Status = nil
						datum.CreatedTime = nil
						datum.ModifiedTime = pointer.FromTime(test.FutureFarTime())
						datum.Revision = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/id"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/userId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/digestMD5"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/mediaType"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/size"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/status"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/createdTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(test.FutureFarTime()), "/modifiedTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "/revision"),
				),
			)
		})
	})

	Context("NewID", func() {
		It("returns a string of 32 lowercase hexidecimal characters", func() {
			Expect(blob.NewID()).To(MatchRegexp("^[0-9a-f]{32}$"))
		})

		It("returns different IDs for each invocation", func() {
			Expect(blob.NewID()).ToNot(Equal(blob.NewID()))
		})
	})

	Context("IsValidID, IDValidator, and ValidateID", func() {
		DescribeTable("return the expected results when the input",
			func(value string, expectedErrors ...error) {
				Expect(blob.IsValidID(value)).To(Equal(len(expectedErrors) == 0))
				errorReporter := structureTest.NewErrorReporter()
				blob.IDValidator(value, errorReporter)
				errorsTest.ExpectEqual(errorReporter.Error(), expectedErrors...)
				errorsTest.ExpectEqual(blob.ValidateID(value), expectedErrors...)
			},
			Entry("is an empty string", "", structureValidator.ErrorValueEmpty()),
			Entry("has string length out of range (lower)", "0123456789abcdefghijklmnopqrstu", blob.ErrorValueStringAsIDNotValid("0123456789abcdefghijklmnopqrstu")),
			Entry("has string length in range", test.RandomStringFromRangeAndCharset(32, 32, test.CharsetLowercase+test.CharsetNumeric)),
			Entry("has string length out of range (upper)", "0123456789abcdefghijklmnopqrstuvw", blob.ErrorValueStringAsIDNotValid("0123456789abcdefghijklmnopqrstuvw")),
			Entry("has uppercase characters", "0123456789ABCDEFGHIJKLMNOPQRSTUV", blob.ErrorValueStringAsIDNotValid("0123456789ABCDEFGHIJKLMNOPQRSTUV")),
			Entry("has symbols", "0123456789!@#$%^abcdefghijklmnop", blob.ErrorValueStringAsIDNotValid("0123456789!@#$%^abcdefghijklmnop")),
			Entry("has whitespace", "0123456789      abcdefghijklmnop", blob.ErrorValueStringAsIDNotValid("0123456789      abcdefghijklmnop")),
		)
	})
})
