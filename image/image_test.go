package image_test

import (
	"bytes"
	"fmt"
	"image/color"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/crypto"
	cryptoTest "github.com/tidepool-org/platform/crypto/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/image"
	imageTest "github.com/tidepool-org/platform/image/test"
	"github.com/tidepool-org/platform/net"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	requestTest "github.com/tidepool-org/platform/request/test"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureTest "github.com/tidepool-org/platform/structure/test"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
	testHttp "github.com/tidepool-org/platform/test/http"
	"github.com/tidepool-org/platform/user"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("Image", func() {
	It("ContentIntentAlternate is expected", func() {
		Expect(image.ContentIntentAlternate).To(Equal("alternate"))
	})

	It("ContentIntentOriginal is expected", func() {
		Expect(image.ContentIntentOriginal).To(Equal("original"))
	})

	It("HeightMaximum is expected", func() {
		Expect(image.HeightMaximum).To(Equal(10000))
	})

	It("HeightMinimum is expected", func() {
		Expect(image.HeightMinimum).To(Equal(1))
	})

	It("MediaTypeImageJPEG is expected", func() {
		Expect(image.MediaTypeImageJPEG).To(Equal("image/jpeg"))
	})

	It("MediaTypeImagePNG is expected", func() {
		Expect(image.MediaTypeImagePNG).To(Equal("image/png"))
	})

	It("ModeDefault is expected", func() {
		Expect(image.ModeDefault).To(Equal(image.ModeFit))
	})

	It("ModeFill is expected", func() {
		Expect(image.ModeFill).To(Equal("fill"))
	})

	It("ModeFillDown is expected", func() {
		Expect(image.ModeFillDown).To(Equal("fillDown"))
	})

	It("ModeFit is expected", func() {
		Expect(image.ModeFit).To(Equal("fit"))
	})

	It("ModeFitDown is expected", func() {
		Expect(image.ModeFitDown).To(Equal("fitDown"))
	})

	It("ModePad is expected", func() {
		Expect(image.ModePad).To(Equal("pad"))
	})

	It("ModePadDown is expected", func() {
		Expect(image.ModePadDown).To(Equal("padDown"))
	})

	It("ModeScale is expected", func() {
		Expect(image.ModeScale).To(Equal("scale"))
	})

	It("ModeScaleDown is expected", func() {
		Expect(image.ModeScaleDown).To(Equal("scaleDown"))
	})

	It("NameLengthMaximum is expected", func() {
		Expect(image.NameLengthMaximum).To(Equal(100))
	})

	It("QualityDefault is expected", func() {
		Expect(image.QualityDefault).To(Equal(95))
	})

	It("QualityMaximum is expected", func() {
		Expect(image.QualityMaximum).To(Equal(100))
	})

	It("QualityMinimum is expected", func() {
		Expect(image.QualityMinimum).To(Equal(1))
	})

	It("RenditionExtensionSeparator is expected", func() {
		Expect(image.RenditionExtensionSeparator).To(Equal("."))
	})

	It("RenditionFieldSeparator is expected", func() {
		Expect(image.RenditionFieldSeparator).To(Equal("_"))
	})

	It("RenditionKeyValueSeparator is expected", func() {
		Expect(image.RenditionKeyValueSeparator).To(Equal("="))
	})

	It("RenditionsLengthMaximum is expected", func() {
		Expect(image.RenditionsLengthMaximum).To(Equal(10))
	})

	It("SizeMaximum is expected", func() {
		Expect(image.SizeMaximum).To(Equal(100 * 1024 * 1024))
	})

	It("StatusAvailable is expected", func() {
		Expect(image.StatusAvailable).To(Equal("available"))
	})

	It("StatusCreated is expected", func() {
		Expect(image.StatusCreated).To(Equal("created"))
	})

	It("WidthMaximum is expected", func() {
		Expect(image.WidthMaximum).To(Equal(10000))
	})

	It("WidthMinimum is expected", func() {
		Expect(image.WidthMinimum).To(Equal(1))
	})

	It("BackgroundDefault returns expected", func() {
		Expect(image.BackgroundDefault()).To(Equal(&image.Color{NRGBA: color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}}))
	})

	It("ContentIntents returns expected", func() {
		Expect(image.ContentIntents()).To(Equal([]string{"alternate", "original"}))
	})

	It("MediaTypes returns expected", func() {
		Expect(image.MediaTypes()).To(Equal([]string{"image/jpeg", "image/png"}))
	})

	It("Modes returns expected", func() {
		Expect(image.Modes()).To(Equal([]string{"fill", "fillDown", "fit", "fitDown", "pad", "padDown", "scale", "scaleDown"}))
	})

	It("Statuses returns expected", func() {
		Expect(image.Statuses()).To(Equal([]string{"available", "created"}))
	})

	Context("Filter", func() {
		Context("NewFilter", func() {
			It("returns successfully with default values", func() {
				Expect(image.NewFilter()).To(Equal(&image.Filter{}))
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *image.Filter), expectedErrors ...error) {
					expectedDatum := imageTest.RandomFilter()
					object := imageTest.NewObjectFromFilter(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := &image.Filter{}
					errorsTest.ExpectEqual(structureParser.NewObject(&object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *image.Filter) {},
				),
				Entry("status missing",
					func(object map[string]interface{}, expectedDatum *image.Filter) {
						delete(object, "status")
						expectedDatum.Status = nil
					},
				),
				Entry("status invalid type",
					func(object map[string]interface{}, expectedDatum *image.Filter) {
						object["status"] = true
						expectedDatum.Status = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/status"),
				),
				Entry("status empty",
					func(object map[string]interface{}, expectedDatum *image.Filter) {
						object["status"] = []string{}
						expectedDatum.Status = pointer.FromStringArray([]string{})
					},
				),
				Entry("status valid",
					func(object map[string]interface{}, expectedDatum *image.Filter) {
						valid := imageTest.RandomStatuses()
						object["status"] = valid
						expectedDatum.Status = pointer.FromStringArray(valid)
					},
				),
				Entry("content intent missing",
					func(object map[string]interface{}, expectedDatum *image.Filter) {
						delete(object, "contentIntent")
						expectedDatum.ContentIntent = nil
					},
				),
				Entry("content intent invalid type",
					func(object map[string]interface{}, expectedDatum *image.Filter) {
						object["contentIntent"] = true
						expectedDatum.ContentIntent = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/contentIntent"),
				),
				Entry("content intent empty",
					func(object map[string]interface{}, expectedDatum *image.Filter) {
						object["contentIntent"] = []string{}
						expectedDatum.ContentIntent = pointer.FromStringArray([]string{})
					},
				),
				Entry("content intent valid",
					func(object map[string]interface{}, expectedDatum *image.Filter) {
						valid := imageTest.RandomContentIntents()
						object["contentIntent"] = valid
						expectedDatum.ContentIntent = pointer.FromStringArray(valid)
					},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *image.Filter) {
						object["status"] = true
						object["contentIntent"] = true
						expectedDatum.Status = nil
						expectedDatum.ContentIntent = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/status"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/contentIntent"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *image.Filter), expectedErrors ...error) {
					datum := imageTest.RandomFilter()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New().Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *image.Filter) {},
				),
				Entry("status missing",
					func(datum *image.Filter) { datum.Status = nil },
				),
				Entry("status empty",
					func(datum *image.Filter) { datum.Status = pointer.FromStringArray([]string{}) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/status"),
				),
				Entry("status element empty",
					func(datum *image.Filter) { datum.Status = pointer.FromStringArray([]string{"available", ""}) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", image.Statuses()), "/status/1"),
				),
				Entry("status element invalid",
					func(datum *image.Filter) { datum.Status = pointer.FromStringArray([]string{"available", "invalid"}) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", image.Statuses()), "/status/1"),
				),
				Entry("status element duplicate",
					func(datum *image.Filter) { datum.Status = pointer.FromStringArray([]string{"available", "available"}) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueDuplicate(), "/status/1"),
				),
				Entry("status available",
					func(datum *image.Filter) { datum.Status = pointer.FromStringArray([]string{"available"}) },
				),
				Entry("status created",
					func(datum *image.Filter) { datum.Status = pointer.FromStringArray([]string{"created"}) },
				),
				Entry("status available and created",
					func(datum *image.Filter) { datum.Status = pointer.FromStringArray([]string{"available", "created"}) },
				),
				Entry("content intent missing",
					func(datum *image.Filter) { datum.ContentIntent = nil },
				),
				Entry("content intent empty",
					func(datum *image.Filter) { datum.ContentIntent = pointer.FromStringArray([]string{}) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/contentIntent"),
				),
				Entry("content intent element empty",
					func(datum *image.Filter) { datum.ContentIntent = pointer.FromStringArray([]string{"alternate", ""}) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", image.ContentIntents()), "/contentIntent/1"),
				),
				Entry("content intent element invalid",
					func(datum *image.Filter) {
						datum.ContentIntent = pointer.FromStringArray([]string{"alternate", "invalid"})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", image.ContentIntents()), "/contentIntent/1"),
				),
				Entry("content intent element duplicate",
					func(datum *image.Filter) {
						datum.ContentIntent = pointer.FromStringArray([]string{"alternate", "alternate"})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueDuplicate(), "/contentIntent/1"),
				),
				Entry("content intent alternate",
					func(datum *image.Filter) { datum.ContentIntent = pointer.FromStringArray([]string{"alternate"}) },
				),
				Entry("content intent original",
					func(datum *image.Filter) { datum.ContentIntent = pointer.FromStringArray([]string{"original"}) },
				),
				Entry("content intent alternate and original",
					func(datum *image.Filter) {
						datum.ContentIntent = pointer.FromStringArray([]string{"alternate", "original"})
					},
				),
				Entry("multiple errors",
					func(datum *image.Filter) {
						datum.Status = pointer.FromStringArray([]string{})
						datum.ContentIntent = pointer.FromStringArray([]string{})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/status"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/contentIntent"),
				),
			)
		})

		Context("with new filter", func() {
			var datum *image.Filter

			BeforeEach(func() {
				datum = imageTest.RandomFilter()
			})

			Context("MutateRequest", func() {
				var req *http.Request

				BeforeEach(func() {
					req = testHttp.NewRequest()
				})

				It("returns an error when the request is missing", func() {
					errorsTest.ExpectEqual(datum.MutateRequest(nil), errors.New("request is missing"))
				})

				It("sets request query as expected", func() {
					Expect(datum.MutateRequest(req)).To(Succeed())
					Expect(req.URL.Query()).To(Equal(url.Values{
						"status":        *datum.Status,
						"contentIntent": *datum.ContentIntent,
					}))
				})

				It("does not set request query when the filter is empty", func() {
					datum.Status = nil
					datum.ContentIntent = nil
					Expect(datum.MutateRequest(req)).To(Succeed())
					Expect(req.URL.Query()).To(BeEmpty())
				})
			})
		})
	})

	Context("Metadata", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *image.Metadata)) {
				datum := imageTest.RandomMetadata()
				mutator(datum)
				test.ExpectSerializedJSON(datum, imageTest.NewObjectFromMetadata(datum, test.ObjectFormatJSON))
			},
			Entry("succeeds",
				func(datum *image.Metadata) {},
			),
			Entry("empty",
				func(datum *image.Metadata) { *datum = *image.NewMetadata() },
			),
		)

		Context("NewMetadata", func() {
			It("returns successfully with default values", func() {
				Expect(image.NewMetadata()).To(Equal(&image.Metadata{}))
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *image.Metadata), expectedErrors ...error) {
					expectedDatum := imageTest.RandomMetadata()
					object := imageTest.NewObjectFromMetadata(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := image.NewMetadata()
					errorsTest.ExpectEqual(structureParser.NewObject(&object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *image.Metadata) {},
				),
				Entry("name missing",
					func(object map[string]interface{}, expectedDatum *image.Metadata) {
						delete(object, "name")
						expectedDatum.Name = nil
					},
				),
				Entry("name invalid type",
					func(object map[string]interface{}, expectedDatum *image.Metadata) {
						object["name"] = true
						expectedDatum.Name = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/name"),
				),
				Entry("name valid",
					func(object map[string]interface{}, expectedDatum *image.Metadata) {
						valid := imageTest.RandomName()
						object["name"] = valid
						expectedDatum.Name = pointer.FromString(valid)
					},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *image.Metadata) {
						object["name"] = true
						expectedDatum.Name = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/name"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *image.Metadata), expectedErrors ...error) {
					datum := imageTest.RandomMetadata()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New().Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *image.Metadata) {},
				),
				Entry("name missing",
					func(datum *image.Metadata) { datum.Name = nil },
				),
				Entry("name empty",
					func(datum *image.Metadata) { datum.Name = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/name"),
				),
				Entry("name valid",
					func(datum *image.Metadata) {
						datum.Name = pointer.FromString(imageTest.RandomName())
					},
				),
				Entry("name valid; length in range (upper)",
					func(datum *image.Metadata) {
						datum.Name = pointer.FromString(test.RandomStringFromRange(100, 100))
					},
				),
				Entry("name invalid; length out of range (upper)",
					func(datum *image.Metadata) { datum.Name = pointer.FromString(test.RandomStringFromRange(101, 101)) },
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/name"),
				),
				Entry("multiple errors",
					func(datum *image.Metadata) { datum.Name = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/name"),
				),
			)
		})
	})

	Context("Content", func() {
		Context("NewContent", func() {
			It("returns successfully with default values", func() {
				Expect(image.NewContent()).To(Equal(&image.Content{}))
			})
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *image.Content), expectedErrors ...error) {
					datum := imageTest.RandomContent()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New().Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *image.Content) {},
				),
				Entry("body missing",
					func(datum *image.Content) { datum.Body = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/body"),
				),
				Entry("body valid",
					func(datum *image.Content) {
						datum.Body = ioutil.NopCloser(bytes.NewReader(imageTest.RandomContentBytes()))
					},
				),
				Entry("digest MD5 missing",
					func(datum *image.Content) { datum.DigestMD5 = nil },
				),
				Entry("digest MD5 empty",
					func(datum *image.Content) { datum.DigestMD5 = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/digestMD5"),
				),
				Entry("digest MD5 invalid",
					func(datum *image.Content) { datum.DigestMD5 = pointer.FromString("#") },
					errorsTest.WithPointerSource(crypto.ErrorValueStringAsBase64EncodedMD5HashNotValid("#"), "/digestMD5"),
				),
				Entry("digest MD5 valid",
					func(datum *image.Content) {
						datum.DigestMD5 = pointer.FromString(cryptoTest.RandomBase64EncodedMD5Hash())
					},
				),
				Entry("media type missing",
					func(datum *image.Content) { datum.MediaType = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/mediaType"),
				),
				Entry("media type empty",
					func(datum *image.Content) { datum.MediaType = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/mediaType"),
				),
				Entry("media type invalid",
					func(datum *image.Content) { datum.MediaType = pointer.FromString("/") },
					errorsTest.WithPointerSource(net.ErrorValueStringAsMediaTypeNotValid("/"), "/mediaType"),
				),
				Entry("media type unsupported",
					func(datum *image.Content) { datum.MediaType = pointer.FromString("application/octet-stream") },
					errorsTest.WithPointerSource(request.ErrorMediaTypeNotSupported("application/octet-stream"), "/mediaType"),
				),
				Entry("media type valid",
					func(datum *image.Content) { datum.MediaType = pointer.FromString(imageTest.RandomMediaType()) },
				),
				Entry("multiple errors",
					func(datum *image.Content) {
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

	Context("Rendition", func() {
		Context("ParseRenditionFromString", func() {
			var rendition *image.Rendition

			BeforeEach(func() {
				rendition = imageTest.RandomRendition()
			})

			It("returns an error when the string there are too many extensions", func() {
				renditionAsString := rendition.String() + ".jpeg"
				result, err := image.ParseRenditionFromString(renditionAsString)
				errorsTest.ExpectEqual(err, image.ErrorValueRenditionNotParsable(renditionAsString))
				Expect(result).To(BeNil())
			})

			It("returns an error when there is no media type", func() {
				rendition.MediaType = nil
				rendition.Quality = nil
				renditionAsString := rendition.String()
				result, err := image.ParseRenditionFromString(renditionAsString)
				errorsTest.ExpectEqual(err, errorsTest.WithParameterSource(structureValidator.ErrorValueNotExists(), "mediaType"))
				Expect(result).To(BeNil())
			})

			It("returns an error when there is the media type is not supported", func() {
				rendition.MediaType = nil
				renditionAsString := fmt.Sprintf("%s%s", rendition.String(), ".bin")
				result, err := image.ParseRenditionFromString(renditionAsString)
				errorsTest.ExpectEqual(err, image.ErrorValueRenditionNotParsable(renditionAsString))
				Expect(result).To(BeNil())
			})

			It("returns an error when a part does not have a proper key and value", func() {
				rendition.MediaType = nil
				rendition.Width = nil
				renditionAsString := fmt.Sprintf("%s%s%s%s", rendition.String(), image.RenditionFieldSeparator, "w", ".jpeg")
				result, err := image.ParseRenditionFromString(renditionAsString)
				errorsTest.ExpectEqual(err, image.ErrorValueRenditionNotParsable(renditionAsString))
				Expect(result).To(BeNil())
			})

			It("returns an error when there is an invalid part", func() {
				rendition.MediaType = nil
				renditionAsString := fmt.Sprintf("%s%s%s%s", rendition.String(), image.RenditionFieldSeparator, "invalid=invalid", ".jpeg")
				result, err := image.ParseRenditionFromString(renditionAsString)
				errorsTest.ExpectEqual(err, image.ErrorValueRenditionNotParsable(renditionAsString))
				Expect(result).To(BeNil())
			})

			It("returns an error when there is a duplicate part", func() {
				rendition.MediaType = nil
				renditionAsString := fmt.Sprintf("%s%s%s%s%s%s", rendition.String(), image.RenditionFieldSeparator, "w", image.RenditionKeyValueSeparator, "1", ".jpeg")
				result, err := image.ParseRenditionFromString(renditionAsString)
				errorsTest.ExpectEqual(err, errorsTest.WithParameterSource(structureParser.ErrorNotParsed(), "width"))
				Expect(result).To(BeNil())
			})

			It("returns successfully when width missing", func() {
				rendition.Width = nil
				result, err := image.ParseRenditionFromString(rendition.String())
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(Equal(rendition))
			})

			It("returns successfully when height missing", func() {
				rendition.Height = nil
				result, err := image.ParseRenditionFromString(rendition.String())
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(Equal(rendition))
			})

			It("returns successfully when mode missing", func() {
				rendition.Mode = nil
				result, err := image.ParseRenditionFromString(rendition.String())
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(Equal(rendition))
			})

			It("returns successfully when background missing", func() {
				rendition.Background = nil
				result, err := image.ParseRenditionFromString(rendition.String())
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(Equal(rendition))
			})

			It("returns successfully when quality missing", func() {
				rendition.Quality = nil
				result, err := image.ParseRenditionFromString(rendition.String())
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(Equal(rendition))
			})

			It("returns successfully", func() {
				result, err := image.ParseRenditionFromString(rendition.String())
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(Equal(rendition))
			})
		})

		Context("NewRendition", func() {
			It("returns successfully with default values", func() {
				Expect(image.NewRendition()).To(Equal(&image.Rendition{}))
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *image.Rendition), expectedErrors ...error) {
					expectedDatum := imageTest.RandomRendition()
					object := imageTest.NewObjectFromRendition(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := &image.Rendition{}
					errorsTest.ExpectEqual(structureParser.NewObject(&object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *image.Rendition) {},
				),
				Entry("media type missing",
					func(object map[string]interface{}, expectedDatum *image.Rendition) {
						delete(object, "mediaType")
						expectedDatum.MediaType = nil
					},
				),
				Entry("media type invalid type",
					func(object map[string]interface{}, expectedDatum *image.Rendition) {
						object["mediaType"] = true
						expectedDatum.MediaType = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/mediaType"),
				),
				Entry("media type empty",
					func(object map[string]interface{}, expectedDatum *image.Rendition) {
						object["mediaType"] = ""
						expectedDatum.MediaType = pointer.FromString("")
					},
				),
				Entry("media type valid",
					func(object map[string]interface{}, expectedDatum *image.Rendition) {
						valid := imageTest.RandomMediaType()
						object["mediaType"] = valid
						expectedDatum.MediaType = pointer.FromString(valid)
					},
				),
				Entry("width missing",
					func(object map[string]interface{}, expectedDatum *image.Rendition) {
						delete(object, "width")
						expectedDatum.Width = nil
					},
				),
				Entry("width invalid type",
					func(object map[string]interface{}, expectedDatum *image.Rendition) {
						object["width"] = true
						expectedDatum.Width = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotInt(true), "/width"),
				),
				Entry("width valid",
					func(object map[string]interface{}, expectedDatum *image.Rendition) {
						valid := imageTest.RandomWidth()
						object["width"] = valid
						expectedDatum.Width = pointer.FromInt(valid)
					},
				),
				Entry("height missing",
					func(object map[string]interface{}, expectedDatum *image.Rendition) {
						delete(object, "height")
						expectedDatum.Height = nil
					},
				),
				Entry("height invalid type",
					func(object map[string]interface{}, expectedDatum *image.Rendition) {
						object["height"] = true
						expectedDatum.Height = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotInt(true), "/height"),
				),
				Entry("height valid",
					func(object map[string]interface{}, expectedDatum *image.Rendition) {
						valid := imageTest.RandomHeight()
						object["height"] = valid
						expectedDatum.Height = pointer.FromInt(valid)
					},
				),
				Entry("mode missing",
					func(object map[string]interface{}, expectedDatum *image.Rendition) {
						delete(object, "mode")
						expectedDatum.Mode = nil
					},
				),
				Entry("mode invalid type",
					func(object map[string]interface{}, expectedDatum *image.Rendition) {
						object["mode"] = true
						expectedDatum.Mode = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/mode"),
				),
				Entry("mode empty",
					func(object map[string]interface{}, expectedDatum *image.Rendition) {
						object["mode"] = ""
						expectedDatum.Mode = pointer.FromString("")
					},
				),
				Entry("mode valid",
					func(object map[string]interface{}, expectedDatum *image.Rendition) {
						valid := imageTest.RandomMode()
						object["mode"] = valid
						expectedDatum.Mode = pointer.FromString(valid)
					},
				),
				Entry("background missing",
					func(object map[string]interface{}, expectedDatum *image.Rendition) {
						delete(object, "background")
						expectedDatum.Background = nil
					},
				),
				Entry("background invalid type",
					func(object map[string]interface{}, expectedDatum *image.Rendition) {
						object["background"] = true
						expectedDatum.Background = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/background"),
				),
				Entry("background invalid content",
					func(object map[string]interface{}, expectedDatum *image.Rendition) {
						valid := "invalid"
						object["background"] = valid
						expectedDatum.Background = nil
					},
					errorsTest.WithPointerSource(image.ErrorValueStringAsColorNotValid("invalid"), "/background"),
				),
				Entry("background valid",
					func(object map[string]interface{}, expectedDatum *image.Rendition) {
						valid := imageTest.RandomColor()
						object["background"] = valid.String()
						expectedDatum.Background = valid
					},
				),
				Entry("quality missing",
					func(object map[string]interface{}, expectedDatum *image.Rendition) {
						delete(object, "quality")
						expectedDatum.Quality = nil
					},
				),
				Entry("quality invalid type",
					func(object map[string]interface{}, expectedDatum *image.Rendition) {
						object["quality"] = true
						expectedDatum.Quality = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotInt(true), "/quality"),
				),
				Entry("quality valid",
					func(object map[string]interface{}, expectedDatum *image.Rendition) {
						valid := imageTest.RandomQuality()
						object["quality"] = valid
						expectedDatum.Quality = pointer.FromInt(valid)
					},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *image.Rendition) {
						object["mediaType"] = true
						object["width"] = true
						object["height"] = true
						object["mode"] = true
						object["background"] = true
						object["quality"] = true
						expectedDatum.MediaType = nil
						expectedDatum.Width = nil
						expectedDatum.Height = nil
						expectedDatum.Mode = nil
						expectedDatum.Background = nil
						expectedDatum.Quality = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/mediaType"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotInt(true), "/width"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotInt(true), "/height"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/mode"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/background"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotInt(true), "/quality"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *image.Rendition), expectedErrors ...error) {
					datum := imageTest.RandomRendition()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New().Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *image.Rendition) {},
				),
				Entry("media type missing",
					func(datum *image.Rendition) {
						datum.MediaType = nil
						datum.Quality = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/mediaType"),
				),
				Entry("media type empty",
					func(datum *image.Rendition) {
						datum.MediaType = pointer.FromString("")
						datum.Quality = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/mediaType"),
				),
				Entry("media type invalid",
					func(datum *image.Rendition) {
						datum.MediaType = pointer.FromString("/")
						datum.Quality = nil
					},
					errorsTest.WithPointerSource(net.ErrorValueStringAsMediaTypeNotValid("/"), "/mediaType"),
				),
				Entry("media type unsupported",
					func(datum *image.Rendition) {
						datum.MediaType = pointer.FromString("application/octet-stream")
						datum.Quality = nil
					},
					errorsTest.WithPointerSource(request.ErrorMediaTypeNotSupported("application/octet-stream"), "/mediaType"),
				),
				Entry("media type valid",
					func(datum *image.Rendition) {
						datum.MediaType = pointer.FromString(imageTest.RandomMediaType())
						datum.Quality = nil
					},
				),
				Entry("width missing",
					func(datum *image.Rendition) { datum.Width = nil },
				),
				Entry("width out of range (lower)",
					func(datum *image.Rendition) { datum.Width = pointer.FromInt(0) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(0, 1, 10000), "/width"),
				),
				Entry("width in range (lower)",
					func(datum *image.Rendition) { datum.Width = pointer.FromInt(1) },
				),
				Entry("width in range (upper)",
					func(datum *image.Rendition) { datum.Width = pointer.FromInt(10000) },
				),
				Entry("width out of range (upper)",
					func(datum *image.Rendition) { datum.Width = pointer.FromInt(10001) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(10001, 1, 10000), "/width"),
				),
				Entry("height missing",
					func(datum *image.Rendition) { datum.Height = nil },
				),
				Entry("height out of range (lower)",
					func(datum *image.Rendition) { datum.Height = pointer.FromInt(0) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(0, 1, 10000), "/height"),
				),
				Entry("height in range (lower)",
					func(datum *image.Rendition) { datum.Height = pointer.FromInt(1) },
				),
				Entry("height in range (upper)",
					func(datum *image.Rendition) { datum.Height = pointer.FromInt(10000) },
				),
				Entry("height out of range (upper)",
					func(datum *image.Rendition) { datum.Height = pointer.FromInt(10001) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(10001, 1, 10000), "/height"),
				),
				Entry("width and height missing",
					func(datum *image.Rendition) {
						datum.Width = nil
						datum.Height = nil
					},
					structureValidator.ErrorValuesNotExistForAny("width", "height"),
				),
				Entry("mode missing",
					func(datum *image.Rendition) { datum.Mode = nil },
				),
				Entry("mode empty",
					func(datum *image.Rendition) { datum.Mode = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", []string{"fill", "fillDown", "fit", "fitDown", "pad", "padDown", "scale", "scaleDown"}), "/mode"),
				),
				Entry("mode invalid",
					func(datum *image.Rendition) { datum.Mode = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"fill", "fillDown", "fit", "fitDown", "pad", "padDown", "scale", "scaleDown"}), "/mode"),
				),
				Entry("mode fill",
					func(datum *image.Rendition) { datum.Mode = pointer.FromString("fill") },
				),
				Entry("mode fill down",
					func(datum *image.Rendition) { datum.Mode = pointer.FromString("fillDown") },
				),
				Entry("mode fit",
					func(datum *image.Rendition) { datum.Mode = pointer.FromString("fit") },
				),
				Entry("mode fit down",
					func(datum *image.Rendition) { datum.Mode = pointer.FromString("fitDown") },
				),
				Entry("mode pad",
					func(datum *image.Rendition) { datum.Mode = pointer.FromString("pad") },
				),
				Entry("mode pad down",
					func(datum *image.Rendition) { datum.Mode = pointer.FromString("padDown") },
				),
				Entry("mode scale",
					func(datum *image.Rendition) { datum.Mode = pointer.FromString("scale") },
				),
				Entry("mode scale down",
					func(datum *image.Rendition) { datum.Mode = pointer.FromString("scaleDown") },
				),
				Entry("background missing",
					func(datum *image.Rendition) { datum.Background = nil },
				),
				Entry("background valid",
					func(datum *image.Rendition) {
						datum.Background = imageTest.RandomColor()
					},
				),
				Entry("media type missing; quality missing",
					func(datum *image.Rendition) {
						datum.MediaType = nil
						datum.Quality = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/mediaType"),
				),
				Entry("media type missing; quality out of range (lower)",
					func(datum *image.Rendition) {
						datum.MediaType = nil
						datum.Quality = pointer.FromInt(0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/mediaType"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/quality"),
				),
				Entry("media type missing; quality in range (lower)",
					func(datum *image.Rendition) {
						datum.MediaType = nil
						datum.Quality = pointer.FromInt(1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/mediaType"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/quality"),
				),
				Entry("media type missing; quality in range (upper)",
					func(datum *image.Rendition) {
						datum.MediaType = nil
						datum.Quality = pointer.FromInt(100)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/mediaType"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/quality"),
				),
				Entry("media type missing; quality out of range (upper)",
					func(datum *image.Rendition) {
						datum.MediaType = nil
						datum.Quality = pointer.FromInt(101)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/mediaType"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/quality"),
				),
				Entry("media type image/jpeg; quality missing",
					func(datum *image.Rendition) {
						datum.MediaType = pointer.FromString("image/jpeg")
						datum.Quality = nil
					},
				),
				Entry("media type image/jpeg; quality out of range (lower)",
					func(datum *image.Rendition) {
						datum.MediaType = pointer.FromString("image/jpeg")
						datum.Quality = pointer.FromInt(0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(0, 1, 100), "/quality"),
				),
				Entry("media type image/jpeg; quality in range (lower)",
					func(datum *image.Rendition) {
						datum.MediaType = pointer.FromString("image/jpeg")
						datum.Quality = pointer.FromInt(1)
					},
				),
				Entry("media type image/jpeg; quality in range (upper)",
					func(datum *image.Rendition) {
						datum.MediaType = pointer.FromString("image/jpeg")
						datum.Quality = pointer.FromInt(100)
					},
				),
				Entry("media type image/jpeg; quality out of range (upper)",
					func(datum *image.Rendition) {
						datum.MediaType = pointer.FromString("image/jpeg")
						datum.Quality = pointer.FromInt(101)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(101, 1, 100), "/quality"),
				),
				Entry("media type image/png; quality missing",
					func(datum *image.Rendition) {
						datum.MediaType = pointer.FromString("image/png")
						datum.Quality = nil
					},
				),
				Entry("media type image/png; quality out of range (lower)",
					func(datum *image.Rendition) {
						datum.MediaType = pointer.FromString("image/png")
						datum.Quality = pointer.FromInt(0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/quality"),
				),
				Entry("media type image/png; quality in range (lower)",
					func(datum *image.Rendition) {
						datum.MediaType = pointer.FromString("image/png")
						datum.Quality = pointer.FromInt(1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/quality"),
				),
				Entry("media type image/png; quality in range (upper)",
					func(datum *image.Rendition) {
						datum.MediaType = pointer.FromString("image/png")
						datum.Quality = pointer.FromInt(100)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/quality"),
				),
				Entry("media type image/png; quality out of range (upper)",
					func(datum *image.Rendition) {
						datum.MediaType = pointer.FromString("image/png")
						datum.Quality = pointer.FromInt(101)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/quality"),
				),
				Entry("multiple errors",
					func(datum *image.Rendition) {
						datum.MediaType = nil
						datum.Width = pointer.FromInt(0)
						datum.Height = pointer.FromInt(0)
						datum.Mode = pointer.FromString("")
						datum.Quality = pointer.FromInt(imageTest.RandomQuality())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/mediaType"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(0, 1, 10000), "/width"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(0, 1, 10000), "/height"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", []string{"fill", "fillDown", "fit", "fitDown", "pad", "padDown", "scale", "scaleDown"}), "/mode"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/quality"),
				),
			)
		})

		Context("with new rendition", func() {
			var datum *image.Rendition

			BeforeEach(func() {
				datum = imageTest.RandomRendition()
			})

			Context("SupportsQuality", func() {
				It("returns false when media type is missing", func() {
					datum.MediaType = nil
					Expect(datum.SupportsQuality()).To(BeFalse())
				})

				It("returns false when media type is empty", func() {
					datum.MediaType = pointer.FromString("")
					Expect(datum.SupportsQuality()).To(BeFalse())
				})

				It("returns false when media type is invalid", func() {
					datum.MediaType = pointer.FromString("invalid")
					Expect(datum.SupportsQuality()).To(BeFalse())
				})

				It("returns false when media type is unknown", func() {
					datum.MediaType = pointer.FromString("application/octet-stream")
					Expect(datum.SupportsQuality()).To(BeFalse())
				})

				It("returns true when media type is image/jpeg", func() {
					datum.MediaType = pointer.FromString("image/jpeg")
					Expect(datum.SupportsQuality()).To(BeTrue())
				})

				It("returns false when media type is image/png", func() {
					datum.MediaType = pointer.FromString("image/png")
					Expect(datum.SupportsQuality()).To(BeFalse())
				})
			})

			Context("SupportsTransparency", func() {
				It("returns false when media type is missing", func() {
					datum.MediaType = nil
					Expect(datum.SupportsTransparency()).To(BeFalse())
				})

				It("returns false when media type is empty", func() {
					datum.MediaType = pointer.FromString("")
					Expect(datum.SupportsTransparency()).To(BeFalse())
				})

				It("returns false when media type is invalid", func() {
					datum.MediaType = pointer.FromString("invalid")
					Expect(datum.SupportsTransparency()).To(BeFalse())
				})

				It("returns false when media type is unknown", func() {
					datum.MediaType = pointer.FromString("application/octet-stream")
					Expect(datum.SupportsTransparency()).To(BeFalse())
				})

				It("returns false when media type is image/jpeg", func() {
					datum.MediaType = pointer.FromString("image/jpeg")
					Expect(datum.SupportsTransparency()).To(BeFalse())
				})

				It("returns true when media type is image/png", func() {
					datum.MediaType = pointer.FromString("image/png")
					Expect(datum.SupportsTransparency()).To(BeTrue())
				})
			})

			Context("with aspect ratio", func() {
				var aspectRatio float64

				BeforeEach(func() {
					aspectRatio = test.RandomFloat64FromRange(0.01, 100.0)
				})

				Context("ConstrainWidth", func() {
					It("constrains the width based upon the height and aspect ratio", func() {
						clone := imageTest.CloneRendition(datum)
						datum.ConstrainWidth(aspectRatio)
						Expect(*datum.Width).To(Equal(int(math.Round(float64(*clone.Height) * aspectRatio))))
						Expect(*datum.Height).To(Equal(*clone.Height))
					})
				})

				Context("ConstrainHeight", func() {
					It("constrains the height based upon the width and aspect ratio", func() {
						clone := imageTest.CloneRendition(datum)
						datum.ConstrainHeight(aspectRatio)
						Expect(*datum.Width).To(Equal(*clone.Width))
						Expect(*datum.Height).To(Equal(int(math.Round(float64(*clone.Width) / aspectRatio))))
					})
				})

				Context("WithDefaults", func() {
					It("returns successfully", func() {
						clone := datum.WithDefaults(aspectRatio)
						Expect(clone).To(Equal(datum))
					})

					It("returns constrained width when width is missing", func() {
						datum.Width = nil
						clone := datum.WithDefaults(aspectRatio)
						datum.ConstrainWidth(aspectRatio)
						Expect(clone).To(Equal(datum))
					})

					It("returns constrained height when height is missing", func() {
						datum.Height = nil
						clone := datum.WithDefaults(aspectRatio)
						datum.ConstrainHeight(aspectRatio)
						Expect(clone).To(Equal(datum))
					})

					It("returns default mode when mode is missing", func() {
						datum.Mode = nil
						clone := datum.WithDefaults(aspectRatio)
						datum.Mode = pointer.FromString(image.ModeDefault)
						Expect(clone).To(Equal(datum))
					})

					It("returns default background when background is missing", func() {
						datum.Background = nil
						clone := datum.WithDefaults(aspectRatio)
						datum.Background = image.BackgroundDefault()
						Expect(clone).To(Equal(datum))
					})

					It("returns unchanged when quality is missing and media type is missing", func() {
						datum.MediaType = nil
						datum.Quality = nil
						clone := datum.WithDefaults(aspectRatio)
						Expect(clone).To(Equal(datum))
					})
					It("returns default quality when missing and media type is image/jpeg", func() {
						datum.MediaType = pointer.FromString("image/jpeg")
						datum.Quality = nil
						clone := datum.WithDefaults(aspectRatio)
						datum.Quality = pointer.FromInt(image.QualityDefault)
						Expect(clone).To(Equal(datum))
					})

					It("returns unchanged when quality is missing and media type is image/png", func() {
						datum.MediaType = pointer.FromString("image/png")
						datum.Quality = nil
						clone := datum.WithDefaults(aspectRatio)
						Expect(clone).To(Equal(datum))
					})
				})
			})
		})

		Context("String", func() {
			DescribeTable("serializes the datum as expected",
				func(mutator func(datum *image.Rendition)) {
					datum := imageTest.RandomRendition()
					mutator(datum)
					var parts []string
					if datum.Width != nil {
						parts = append(parts, fmt.Sprintf("w=%d", *datum.Width))
					}
					if datum.Height != nil {
						parts = append(parts, fmt.Sprintf("h=%d", *datum.Height))
					}
					if datum.Mode != nil {
						parts = append(parts, fmt.Sprintf("m=%s", *datum.Mode))
					}
					if datum.Background != nil {
						parts = append(parts, fmt.Sprintf("b=%s", datum.Background.String()))
					}
					if datum.Quality != nil {
						parts = append(parts, fmt.Sprintf("q=%d", *datum.Quality))
					}
					renditionAsString := strings.Join(parts, "_")
					if datum.MediaType != nil {
						if extension, valid := image.ExtensionFromMediaType(*datum.MediaType); valid {
							renditionAsString = fmt.Sprintf("%s.%s", renditionAsString, extension)
						}
					}
					Expect(datum.String()).To(Equal(renditionAsString))
				},
				Entry("succeeds",
					func(datum *image.Rendition) {},
				),
				Entry("without media type",
					func(datum *image.Rendition) { datum.MediaType = nil },
				),
				Entry("with unsupported media type",
					func(datum *image.Rendition) { datum.MediaType = pointer.FromString("application/octet-stream") },
				),
				Entry("without width",
					func(datum *image.Rendition) { datum.Width = nil },
				),
				Entry("without height",
					func(datum *image.Rendition) { datum.Height = nil },
				),
				Entry("without mode",
					func(datum *image.Rendition) { datum.Mode = nil },
				),
				Entry("without background",
					func(datum *image.Rendition) { datum.Background = nil },
				),
				Entry("without quality",
					func(datum *image.Rendition) { datum.Quality = nil },
				),
				Entry("without all",
					func(datum *image.Rendition) {
						datum.MediaType = nil
						datum.Width = nil
						datum.Height = nil
						datum.Mode = nil
						datum.Background = nil
						datum.Quality = nil
					},
				),
			)
		})
	})

	Context("ContentAttributes", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *image.ContentAttributes)) {
				datum := imageTest.RandomContentAttributes()
				mutator(datum)
				test.ExpectSerializedJSON(datum, imageTest.NewObjectFromContentAttributes(datum, test.ObjectFormatJSON))
				test.ExpectSerializedBSON(datum, imageTest.NewObjectFromContentAttributes(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *image.ContentAttributes) {},
			),
			Entry("empty",
				func(datum *image.ContentAttributes) { *datum = *image.NewContentAttributes() },
			),
			Entry("with modified time",
				func(datum *image.ContentAttributes) {
					datum.ModifiedTime = pointer.FromTime(test.RandomTimeFromRange(*datum.CreatedTime, time.Now()).Truncate(time.Second))
				},
			),
		)

		Context("ParseContentAttributes", func() {
			It("returns nil when the object is missing", func() {
				Expect(image.ParseContentAttributes(structureParser.NewObject(nil))).To(BeNil())
			})

			It("returns new datum when the object is valid", func() {
				datum := imageTest.RandomContentAttributes()
				object := imageTest.NewObjectFromContentAttributes(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(&object)
				Expect(image.ParseContentAttributes(parser)).To(imageTest.MatchContentAttributes(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("NewContentAttributes", func() {
			It("returns successfully with default values", func() {
				Expect(image.NewContentAttributes()).To(Equal(&image.ContentAttributes{}))
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *image.ContentAttributes), expectedErrors ...error) {
					expectedDatum := imageTest.RandomContentAttributes()
					object := imageTest.NewObjectFromContentAttributes(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := &image.ContentAttributes{}
					errorsTest.ExpectEqual(structureParser.NewObject(&object).Parse(datum), expectedErrors...)
					Expect(datum).To(imageTest.MatchContentAttributes(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *image.ContentAttributes) {},
				),
				Entry("digest MD5 missing",
					func(object map[string]interface{}, expectedDatum *image.ContentAttributes) {
						delete(object, "digestMD5")
						expectedDatum.DigestMD5 = nil
					},
				),
				Entry("digest MD5 invalid type",
					func(object map[string]interface{}, expectedDatum *image.ContentAttributes) {
						object["digestMD5"] = true
						expectedDatum.DigestMD5 = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/digestMD5"),
				),
				Entry("digest MD5 empty",
					func(object map[string]interface{}, expectedDatum *image.ContentAttributes) {
						object["digestMD5"] = ""
						expectedDatum.DigestMD5 = pointer.FromString("")
					},
				),
				Entry("digest MD5 valid",
					func(object map[string]interface{}, expectedDatum *image.ContentAttributes) {
						valid := cryptoTest.RandomBase64EncodedMD5Hash()
						object["digestMD5"] = valid
						expectedDatum.DigestMD5 = pointer.FromString(valid)
					},
				),
				Entry("media type missing",
					func(object map[string]interface{}, expectedDatum *image.ContentAttributes) {
						delete(object, "mediaType")
						expectedDatum.MediaType = nil
					},
				),
				Entry("media type invalid type",
					func(object map[string]interface{}, expectedDatum *image.ContentAttributes) {
						object["mediaType"] = true
						expectedDatum.MediaType = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/mediaType"),
				),
				Entry("media type empty",
					func(object map[string]interface{}, expectedDatum *image.ContentAttributes) {
						object["mediaType"] = ""
						expectedDatum.MediaType = pointer.FromString("")
					},
				),
				Entry("media type valid",
					func(object map[string]interface{}, expectedDatum *image.ContentAttributes) {
						valid := imageTest.RandomMediaType()
						object["mediaType"] = valid
						expectedDatum.MediaType = pointer.FromString(valid)
					},
				),
				Entry("width missing",
					func(object map[string]interface{}, expectedDatum *image.ContentAttributes) {
						delete(object, "width")
						expectedDatum.Width = nil
					},
				),
				Entry("width invalid type",
					func(object map[string]interface{}, expectedDatum *image.ContentAttributes) {
						object["width"] = true
						expectedDatum.Width = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotInt(true), "/width"),
				),
				Entry("width valid",
					func(object map[string]interface{}, expectedDatum *image.ContentAttributes) {
						valid := imageTest.RandomWidth()
						object["width"] = valid
						expectedDatum.Width = pointer.FromInt(valid)
					},
				),
				Entry("height missing",
					func(object map[string]interface{}, expectedDatum *image.ContentAttributes) {
						delete(object, "height")
						expectedDatum.Height = nil
					},
				),
				Entry("height invalid type",
					func(object map[string]interface{}, expectedDatum *image.ContentAttributes) {
						object["height"] = true
						expectedDatum.Height = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotInt(true), "/height"),
				),
				Entry("height valid",
					func(object map[string]interface{}, expectedDatum *image.ContentAttributes) {
						valid := imageTest.RandomHeight()
						object["height"] = valid
						expectedDatum.Height = pointer.FromInt(valid)
					},
				),
				Entry("size missing",
					func(object map[string]interface{}, expectedDatum *image.ContentAttributes) {
						delete(object, "size")
						expectedDatum.Size = nil
					},
				),
				Entry("size invalid type",
					func(object map[string]interface{}, expectedDatum *image.ContentAttributes) {
						object["size"] = true
						expectedDatum.Size = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotInt(true), "/size"),
				),
				Entry("size valid",
					func(object map[string]interface{}, expectedDatum *image.ContentAttributes) {
						valid := test.RandomIntFromRange(1, 100*1024*1024)
						object["size"] = valid
						expectedDatum.Size = pointer.FromInt(valid)
					},
				),
				Entry("created time missing",
					func(object map[string]interface{}, expectedDatum *image.ContentAttributes) {
						delete(object, "createdTime")
						expectedDatum.CreatedTime = nil
					},
				),
				Entry("created time invalid type",
					func(object map[string]interface{}, expectedDatum *image.ContentAttributes) {
						object["createdTime"] = true
						expectedDatum.CreatedTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/createdTime"),
				),
				Entry("created time invalid",
					func(object map[string]interface{}, expectedDatum *image.ContentAttributes) {
						object["createdTime"] = "invalid"
						expectedDatum.CreatedTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorValueTimeNotParsable("invalid", time.RFC3339), "/createdTime"),
				),
				Entry("created time valid",
					func(object map[string]interface{}, expectedDatum *image.ContentAttributes) {
						valid := test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now()).Truncate(time.Second)
						object["createdTime"] = valid.Format(time.RFC3339)
						expectedDatum.CreatedTime = pointer.FromTime(valid)
					},
				),
				Entry("modified time missing",
					func(object map[string]interface{}, expectedDatum *image.ContentAttributes) {
						delete(object, "modifiedTime")
						expectedDatum.ModifiedTime = nil
					},
				),
				Entry("modified time invalid type",
					func(object map[string]interface{}, expectedDatum *image.ContentAttributes) {
						object["modifiedTime"] = true
						expectedDatum.ModifiedTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/modifiedTime"),
				),
				Entry("modified time invalid",
					func(object map[string]interface{}, expectedDatum *image.ContentAttributes) {
						object["modifiedTime"] = "invalid"
						expectedDatum.ModifiedTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorValueTimeNotParsable("invalid", time.RFC3339), "/modifiedTime"),
				),
				Entry("modified time valid",
					func(object map[string]interface{}, expectedDatum *image.ContentAttributes) {
						valid := test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now()).Truncate(time.Second)
						object["modifiedTime"] = valid.Format(time.RFC3339)
						expectedDatum.ModifiedTime = pointer.FromTime(valid)
					},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *image.ContentAttributes) {
						object["digestMD5"] = true
						object["mediaType"] = true
						object["width"] = true
						object["height"] = true
						object["size"] = true
						object["createdTime"] = true
						object["modifiedTime"] = true
						expectedDatum.DigestMD5 = nil
						expectedDatum.MediaType = nil
						expectedDatum.Width = nil
						expectedDatum.Height = nil
						expectedDatum.Size = nil
						expectedDatum.CreatedTime = nil
						expectedDatum.ModifiedTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/digestMD5"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/mediaType"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotInt(true), "/width"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotInt(true), "/height"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotInt(true), "/size"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/createdTime"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/modifiedTime"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *image.ContentAttributes), expectedErrors ...error) {
					datum := imageTest.RandomContentAttributes()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New().Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *image.ContentAttributes) {},
				),
				Entry("digest MD5 missing",
					func(datum *image.ContentAttributes) { datum.DigestMD5 = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/digestMD5"),
				),
				Entry("digest MD5 empty",
					func(datum *image.ContentAttributes) { datum.DigestMD5 = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/digestMD5"),
				),
				Entry("digest MD5 invalid",
					func(datum *image.ContentAttributes) { datum.DigestMD5 = pointer.FromString("#") },
					errorsTest.WithPointerSource(crypto.ErrorValueStringAsBase64EncodedMD5HashNotValid("#"), "/digestMD5"),
				),
				Entry("digest MD5 valid",
					func(datum *image.ContentAttributes) {
						datum.DigestMD5 = pointer.FromString(cryptoTest.RandomBase64EncodedMD5Hash())
					},
				),
				Entry("media type missing",
					func(datum *image.ContentAttributes) { datum.MediaType = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/mediaType"),
				),
				Entry("media type empty",
					func(datum *image.ContentAttributes) { datum.MediaType = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/mediaType"),
				),
				Entry("media type invalid",
					func(datum *image.ContentAttributes) { datum.MediaType = pointer.FromString("/") },
					errorsTest.WithPointerSource(net.ErrorValueStringAsMediaTypeNotValid("/"), "/mediaType"),
				),
				Entry("media type unsupported",
					func(datum *image.ContentAttributes) { datum.MediaType = pointer.FromString("application/octet-stream") },
					errorsTest.WithPointerSource(request.ErrorMediaTypeNotSupported("application/octet-stream"), "/mediaType"),
				),
				Entry("media type valid",
					func(datum *image.ContentAttributes) {
						datum.MediaType = pointer.FromString(imageTest.RandomMediaType())
					},
				),
				Entry("width missing",
					func(datum *image.ContentAttributes) { datum.Width = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/width"),
				),
				Entry("width out of range (lower)",
					func(datum *image.ContentAttributes) { datum.Width = pointer.FromInt(0) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotGreaterThan(0, 0), "/width"),
				),
				Entry("width in range (lower)",
					func(datum *image.ContentAttributes) { datum.Width = pointer.FromInt(1) },
				),
				Entry("height missing",
					func(datum *image.ContentAttributes) { datum.Height = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/height"),
				),
				Entry("height out of range (lower)",
					func(datum *image.ContentAttributes) { datum.Height = pointer.FromInt(0) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotGreaterThan(0, 0), "/height"),
				),
				Entry("height in range (lower)",
					func(datum *image.ContentAttributes) { datum.Height = pointer.FromInt(1) },
				),
				Entry("size missing",
					func(datum *image.ContentAttributes) { datum.Size = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/size"),
				),
				Entry("size out of range (lower)",
					func(datum *image.ContentAttributes) { datum.Size = pointer.FromInt(0) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotGreaterThan(0, 0), "/size"),
				),
				Entry("size in range (lower)",
					func(datum *image.ContentAttributes) { datum.Size = pointer.FromInt(1) },
				),
				Entry("created time missing",
					func(datum *image.ContentAttributes) { datum.CreatedTime = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/createdTime"),
				),
				Entry("created time zero",
					func(datum *image.ContentAttributes) { datum.CreatedTime = pointer.FromTime(time.Time{}) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/createdTime"),
				),
				Entry("created time after now",
					func(datum *image.ContentAttributes) {
						datum.CreatedTime = pointer.FromTime(test.FutureFarTime())
						datum.ModifiedTime = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(test.FutureFarTime()), "/createdTime"),
				),
				Entry("created time valid",
					func(datum *image.ContentAttributes) {
						datum.CreatedTime = pointer.FromTime(test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now()).Truncate(time.Second))
						datum.ModifiedTime = pointer.FromTime(test.RandomTimeFromRange(*datum.CreatedTime, time.Now()).Truncate(time.Second))
					},
				),
				Entry("modified time missing",
					func(datum *image.ContentAttributes) { datum.ModifiedTime = nil },
				),
				Entry("modified time before created time",
					func(datum *image.ContentAttributes) {
						datum.CreatedTime = pointer.FromTime(test.PastNearTime())
						datum.ModifiedTime = pointer.FromTime(test.PastFarTime())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotAfter(test.PastFarTime(), test.PastNearTime()), "/modifiedTime"),
				),
				Entry("modified time after now",
					func(datum *image.ContentAttributes) { datum.ModifiedTime = pointer.FromTime(test.FutureFarTime()) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(test.FutureFarTime()), "/modifiedTime"),
				),
				Entry("modified time valid",
					func(datum *image.ContentAttributes) {
						datum.ModifiedTime = pointer.FromTime(test.RandomTimeFromRange(*datum.CreatedTime, time.Now()).Truncate(time.Second))
					},
				),
				Entry("multiple errors",
					func(datum *image.ContentAttributes) {
						datum.DigestMD5 = pointer.FromString("")
						datum.MediaType = nil
						datum.Width = pointer.FromInt(0)
						datum.Height = pointer.FromInt(0)
						datum.Size = pointer.FromInt(0)
						datum.CreatedTime = nil
						datum.ModifiedTime = pointer.FromTime(test.FutureFarTime())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/digestMD5"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/mediaType"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotGreaterThan(0, 0), "/width"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotGreaterThan(0, 0), "/height"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotGreaterThan(0, 0), "/size"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/createdTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(test.FutureFarTime()), "/modifiedTime"),
				),
			)
		})

		Context("with new content attributes", func() {
			var datum *image.ContentAttributes

			BeforeEach(func() {
				datum = imageTest.RandomContentAttributes()
			})

			Context("SupportsQuality", func() {
				It("returns false when media type is missing", func() {
					datum.MediaType = nil
					Expect(datum.SupportsQuality()).To(BeFalse())
				})

				It("returns false when media type is empty", func() {
					datum.MediaType = pointer.FromString("")
					Expect(datum.SupportsQuality()).To(BeFalse())
				})

				It("returns false when media type is invalid", func() {
					datum.MediaType = pointer.FromString("invalid")
					Expect(datum.SupportsQuality()).To(BeFalse())
				})

				It("returns false when media type is unknown", func() {
					datum.MediaType = pointer.FromString("application/octet-stream")
					Expect(datum.SupportsQuality()).To(BeFalse())
				})

				It("returns true when media type is image/jpeg", func() {
					datum.MediaType = pointer.FromString("image/jpeg")
					Expect(datum.SupportsQuality()).To(BeTrue())
				})

				It("returns false when media type is image/png", func() {
					datum.MediaType = pointer.FromString("image/png")
					Expect(datum.SupportsQuality()).To(BeFalse())
				})
			})

			Context("SupportsTransparency", func() {
				It("returns false when media type is missing", func() {
					datum.MediaType = nil
					Expect(datum.SupportsTransparency()).To(BeFalse())
				})

				It("returns false when media type is empty", func() {
					datum.MediaType = pointer.FromString("")
					Expect(datum.SupportsTransparency()).To(BeFalse())
				})

				It("returns false when media type is invalid", func() {
					datum.MediaType = pointer.FromString("invalid")
					Expect(datum.SupportsTransparency()).To(BeFalse())
				})

				It("returns false when media type is unknown", func() {
					datum.MediaType = pointer.FromString("application/octet-stream")
					Expect(datum.SupportsTransparency()).To(BeFalse())
				})

				It("returns false when media type is image/jpeg", func() {
					datum.MediaType = pointer.FromString("image/jpeg")
					Expect(datum.SupportsTransparency()).To(BeFalse())
				})

				It("returns true when media type is image/png", func() {
					datum.MediaType = pointer.FromString("image/png")
					Expect(datum.SupportsTransparency()).To(BeTrue())
				})
			})
		})
	})

	Context("Image", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *image.Image)) {
				datum := imageTest.RandomImage()
				mutator(datum)
				test.ExpectSerializedJSON(datum, imageTest.NewObjectFromImage(datum, test.ObjectFormatJSON))
				test.ExpectSerializedBSON(datum, imageTest.NewObjectFromImage(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *image.Image) {},
			),
			Entry("empty",
				func(datum *image.Image) { *datum = image.Image{} },
			),
			Entry("with modified time",
				func(datum *image.Image) {
					datum.ModifiedTime = pointer.FromTime(test.RandomTimeFromRange(*datum.CreatedTime, time.Now()).Truncate(time.Second))
				},
			),
			Entry("with deleted time",
				func(datum *image.Image) {
					datum.ModifiedTime = pointer.FromTime(test.RandomTimeFromRange(*datum.CreatedTime, time.Now()).Truncate(time.Second))
					datum.DeletedTime = pointer.CloneTime(datum.ModifiedTime)
				},
			),
		)

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *image.Image), expectedErrors ...error) {
					expectedDatum := imageTest.RandomImage()
					object := imageTest.NewObjectFromImage(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := &image.Image{}
					errorsTest.ExpectEqual(structureParser.NewObject(&object).Parse(datum), expectedErrors...)
					Expect(datum).To(imageTest.MatchImage(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *image.Image) {},
				),
				Entry("id missing",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						delete(object, "id")
						expectedDatum.ID = nil
					},
				),
				Entry("id invalid type",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						object["id"] = true
						expectedDatum.ID = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/id"),
				),
				Entry("id empty",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						object["id"] = ""
						expectedDatum.ID = pointer.FromString("")
					},
				),
				Entry("id valid",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						valid := imageTest.RandomID()
						object["id"] = valid
						expectedDatum.ID = pointer.FromString(valid)
					},
				),
				Entry("user id missing",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						delete(object, "userId")
						expectedDatum.UserID = nil
					},
				),
				Entry("user id invalid type",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						object["userId"] = true
						expectedDatum.UserID = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/userId"),
				),
				Entry("user id empty",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						object["userId"] = ""
						expectedDatum.UserID = pointer.FromString("")
					},
				),
				Entry("user id valid",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						valid := userTest.RandomID()
						object["userId"] = valid
						expectedDatum.UserID = pointer.FromString(valid)
					},
				),
				Entry("status missing",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						delete(object, "status")
						expectedDatum.Status = nil
					},
				),
				Entry("status invalid type",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						object["status"] = true
						expectedDatum.Status = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/status"),
				),
				Entry("status empty",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						object["status"] = ""
						expectedDatum.Status = pointer.FromString("")
					},
				),
				Entry("status valid",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						valid := imageTest.RandomStatus()
						object["status"] = valid
						expectedDatum.Status = pointer.FromString(valid)
					},
				),
				Entry("name missing",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						delete(object, "name")
						expectedDatum.Name = nil
					},
				),
				Entry("name empty",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						object["name"] = ""
						expectedDatum.Name = pointer.FromString("")
					},
				),
				Entry("name invalid type",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						object["name"] = true
						expectedDatum.Name = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/name"),
				),
				Entry("name valid",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						valid := imageTest.RandomName()
						object["name"] = valid
						expectedDatum.Name = pointer.FromString(valid)
					},
				),
				Entry("content id missing",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						delete(object, "contentId")
						expectedDatum.ContentID = nil
					},
				),
				Entry("content id invalid type",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						object["contentId"] = true
						expectedDatum.ContentID = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/contentId"),
				),
				Entry("content id empty",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						object["contentId"] = ""
						expectedDatum.ContentID = pointer.FromString("")
					},
				),
				Entry("content id valid",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						valid := imageTest.RandomContentID()
						object["contentId"] = valid
						expectedDatum.ContentID = pointer.FromString(valid)
					},
				),
				Entry("content intent missing",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						delete(object, "contentIntent")
						expectedDatum.ContentIntent = nil
					},
				),
				Entry("content intent invalid type",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						object["contentIntent"] = true
						expectedDatum.ContentIntent = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/contentIntent"),
				),
				Entry("content intent empty",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						object["contentIntent"] = ""
						expectedDatum.ContentIntent = pointer.FromString("")
					},
				),
				Entry("content intent valid",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						valid := imageTest.RandomContentIntent()
						object["contentIntent"] = valid
						expectedDatum.ContentIntent = pointer.FromString(valid)
					},
				),
				Entry("content attributes missing",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						delete(object, "contentAttributes")
						expectedDatum.ContentAttributes = nil
					},
				),
				Entry("content attributes invalid type",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						object["contentAttributes"] = true
						expectedDatum.ContentAttributes = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/contentAttributes"),
				),
				Entry("content attributes valid",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						valid := imageTest.RandomContentAttributes()
						object["contentAttributes"] = imageTest.NewObjectFromContentAttributes(valid, test.ObjectFormatJSON)
						expectedDatum.ContentAttributes = valid
					},
				),
				Entry("renditions id missing",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						delete(object, "renditionsId")
						expectedDatum.RenditionsID = nil
					},
				),
				Entry("renditions id invalid type",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						object["renditionsId"] = true
						expectedDatum.RenditionsID = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/renditionsId"),
				),
				Entry("renditions id empty",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						object["renditionsId"] = ""
						expectedDatum.RenditionsID = pointer.FromString("")
					},
				),
				Entry("renditions id valid",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						valid := imageTest.RandomRenditionsID()
						object["renditionsId"] = valid
						expectedDatum.RenditionsID = pointer.FromString(valid)
					},
				),
				Entry("renditions missing",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						delete(object, "renditions")
						expectedDatum.Renditions = nil
					},
				),
				Entry("renditions invalid type",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						object["renditions"] = true
						expectedDatum.Renditions = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/renditions"),
				),
				Entry("renditions valid",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						valid := imageTest.RandomRenditionsAsStrings()
						object["renditions"] = test.NewObjectFromStringArray(valid, test.ObjectFormatJSON)
						expectedDatum.Renditions = pointer.FromStringArray(valid)
					},
				),
				Entry("created time missing",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						delete(object, "createdTime")
						expectedDatum.CreatedTime = nil
					},
				),
				Entry("created time invalid type",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						object["createdTime"] = true
						expectedDatum.CreatedTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/createdTime"),
				),
				Entry("created time invalid",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						object["createdTime"] = "invalid"
						expectedDatum.CreatedTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorValueTimeNotParsable("invalid", time.RFC3339), "/createdTime"),
				),
				Entry("created time valid",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						valid := test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now()).Truncate(time.Second)
						object["createdTime"] = valid.Format(time.RFC3339)
						expectedDatum.CreatedTime = pointer.FromTime(valid)
					},
				),
				Entry("modified time missing",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						delete(object, "modifiedTime")
						expectedDatum.ModifiedTime = nil
					},
				),
				Entry("modified time invalid type",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						object["modifiedTime"] = true
						expectedDatum.ModifiedTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/modifiedTime"),
				),
				Entry("modified time invalid",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						object["modifiedTime"] = "invalid"
						expectedDatum.ModifiedTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorValueTimeNotParsable("invalid", time.RFC3339), "/modifiedTime"),
				),
				Entry("modified time valid",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						valid := test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now()).Truncate(time.Second)
						object["modifiedTime"] = valid.Format(time.RFC3339)
						expectedDatum.ModifiedTime = pointer.FromTime(valid)
					},
				),
				Entry("deleted time missing",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						delete(object, "deletedTime")
						expectedDatum.DeletedTime = nil
					},
				),
				Entry("deleted time invalid type",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						object["deletedTime"] = true
						expectedDatum.DeletedTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/deletedTime"),
				),
				Entry("deleted time invalid",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						object["deletedTime"] = "invalid"
						expectedDatum.DeletedTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorValueTimeNotParsable("invalid", time.RFC3339), "/deletedTime"),
				),
				Entry("deleted time valid",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						valid := test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now()).Truncate(time.Second)
						object["deletedTime"] = valid.Format(time.RFC3339)
						expectedDatum.DeletedTime = pointer.FromTime(valid)
					},
				),
				Entry("revision missing",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						delete(object, "revision")
						expectedDatum.Revision = nil
					},
				),
				Entry("revision invalid type",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						object["revision"] = true
						expectedDatum.Revision = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotInt(true), "/revision"),
				),
				Entry("revision valid",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						valid := requestTest.RandomRevision()
						object["revision"] = valid
						expectedDatum.Revision = pointer.FromInt(valid)
					},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *image.Image) {
						object["id"] = true
						object["userId"] = true
						object["status"] = true
						object["name"] = true
						object["contentId"] = true
						object["contentIntent"] = true
						object["contentAttributes"] = true
						object["renditionsId"] = true
						object["renditions"] = true
						object["createdTime"] = true
						object["modifiedTime"] = true
						object["deletedTime"] = true
						object["revision"] = true
						expectedDatum.ID = nil
						expectedDatum.UserID = nil
						expectedDatum.Status = nil
						expectedDatum.Name = nil
						expectedDatum.ContentID = nil
						expectedDatum.ContentIntent = nil
						expectedDatum.ContentAttributes = nil
						expectedDatum.RenditionsID = nil
						expectedDatum.Renditions = nil
						expectedDatum.CreatedTime = nil
						expectedDatum.ModifiedTime = nil
						expectedDatum.DeletedTime = nil
						expectedDatum.Revision = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/id"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/userId"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/status"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/name"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/contentId"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/contentIntent"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/contentAttributes"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/renditionsId"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/renditions"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/createdTime"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/modifiedTime"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/deletedTime"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotInt(true), "/revision"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *image.Image), expectedErrors ...error) {
					datum := imageTest.RandomImage()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New().Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *image.Image) {},
				),
				Entry("id missing",
					func(datum *image.Image) { datum.ID = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/id"),
				),
				Entry("id empty",
					func(datum *image.Image) { datum.ID = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/id"),
				),
				Entry("id invalid",
					func(datum *image.Image) { datum.ID = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(image.ErrorValueStringAsIDNotValid("invalid"), "/id"),
				),
				Entry("id valid",
					func(datum *image.Image) { datum.ID = pointer.FromString(imageTest.RandomID()) },
				),
				Entry("user id missing",
					func(datum *image.Image) { datum.UserID = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/userId"),
				),
				Entry("user id empty",
					func(datum *image.Image) { datum.UserID = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/userId"),
				),
				Entry("user id invalid",
					func(datum *image.Image) { datum.UserID = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(user.ErrorValueStringAsIDNotValid("invalid"), "/userId"),
				),
				Entry("user id valid",
					func(datum *image.Image) { datum.UserID = pointer.FromString(userTest.RandomID()) },
				),
				Entry("status missing; content id, content intent, content attributes, renditions id, and renditions missing",
					func(datum *image.Image) {
						datum.Status = nil
						datum.ContentID = nil
						datum.ContentIntent = nil
						datum.ContentAttributes = nil
						datum.RenditionsID = nil
						datum.Renditions = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/status"),
				),
				Entry("status missing; content id missing",
					func(datum *image.Image) {
						datum.Status = nil
						datum.ContentID = nil
						datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/status"),
				),
				Entry("status missing; content id invalid",
					func(datum *image.Image) {
						datum.Status = nil
						datum.ContentID = pointer.FromString("")
						datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/status"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/contentId"),
				),
				Entry("status missing; content intent missing",
					func(datum *image.Image) {
						datum.Status = nil
						datum.ContentID = pointer.FromString(imageTest.RandomContentID())
						datum.ContentIntent = nil
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/status"),
				),
				Entry("status missing; content intent invalid",
					func(datum *image.Image) {
						datum.Status = nil
						datum.ContentID = pointer.FromString(imageTest.RandomContentID())
						datum.ContentIntent = pointer.FromString("")
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/status"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", image.ContentIntents()), "/contentIntent"),
				),
				Entry("status missing; content attributes missing",
					func(datum *image.Image) {
						datum.Status = nil
						datum.ContentID = pointer.FromString(imageTest.RandomContentID())
						datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
						datum.ContentAttributes = nil
						datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/status"),
				),
				Entry("status missing; content attributes empty",
					func(datum *image.Image) {
						datum.Status = nil
						datum.ContentID = pointer.FromString(imageTest.RandomContentID())
						datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
						datum.ContentAttributes = image.NewContentAttributes()
						datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/status"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/contentAttributes/digestMD5"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/contentAttributes/mediaType"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/contentAttributes/width"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/contentAttributes/height"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/contentAttributes/size"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/contentAttributes/createdTime"),
				),
				Entry("status missing; content attributes invalid",
					func(datum *image.Image) {
						datum.Status = nil
						datum.ContentID = pointer.FromString(imageTest.RandomContentID())
						datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.ContentAttributes.DigestMD5 = pointer.FromString("")
						datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/status"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/contentAttributes/digestMD5"),
				),
				Entry("status missing; renditions id missing",
					func(datum *image.Image) {
						datum.Status = nil
						datum.ContentID = pointer.FromString(imageTest.RandomContentID())
						datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.RenditionsID = nil
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/status"),
				),
				Entry("status missing; renditions id invalid",
					func(datum *image.Image) {
						datum.Status = nil
						datum.ContentID = pointer.FromString(imageTest.RandomContentID())
						datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.RenditionsID = pointer.FromString("")
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/status"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/renditionsId"),
				),
				Entry("status missing; renditions missing",
					func(datum *image.Image) {
						datum.Status = nil
						datum.ContentID = pointer.FromString(imageTest.RandomContentID())
						datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
						datum.Renditions = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/status"),
				),
				Entry("status missing; renditions empty",
					func(datum *image.Image) {
						datum.Status = nil
						datum.ContentID = pointer.FromString(imageTest.RandomContentID())
						datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
						datum.Renditions = pointer.FromStringArray([]string{})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/status"),
				),
				Entry("status missing; renditions element empty",
					func(datum *image.Image) {
						datum.Status = nil
						datum.ContentID = pointer.FromString(imageTest.RandomContentID())
						datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
						(*datum.Renditions)[0] = ""
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/status"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/renditions/0"),
				),
				Entry("status missing; renditions element not unique",
					func(datum *image.Image) {
						datum.Status = nil
						datum.ContentID = pointer.FromString(imageTest.RandomContentID())
						datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
						(*datum.Renditions)[1] = (*datum.Renditions)[0]
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/status"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueDuplicate(), "/renditions/1"),
				),
				Entry("status missing; content id, content intent, content attributes, renditions id, and renditions valid",
					func(datum *image.Image) {
						datum.Status = nil
						datum.ContentID = pointer.FromString(imageTest.RandomContentID())
						datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/status"),
				),
				Entry("status invalid; content id, content intent, content attributes, renditions id, and renditions missing",
					func(datum *image.Image) {
						datum.Status = pointer.FromString("")
						datum.ContentID = nil
						datum.ContentIntent = nil
						datum.ContentAttributes = nil
						datum.RenditionsID = nil
						datum.Renditions = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", image.Statuses()), "/status"),
				),
				Entry("status invalid; content id missing",
					func(datum *image.Image) {
						datum.Status = pointer.FromString("")
						datum.ContentID = nil
						datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", image.Statuses()), "/status"),
				),
				Entry("status invalid; content id invalid",
					func(datum *image.Image) {
						datum.Status = pointer.FromString("")
						datum.ContentID = pointer.FromString("")
						datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", image.Statuses()), "/status"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/contentId"),
				),
				Entry("status invalid; content intent missing",
					func(datum *image.Image) {
						datum.Status = pointer.FromString("")
						datum.ContentID = pointer.FromString(imageTest.RandomContentID())
						datum.ContentIntent = nil
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", image.Statuses()), "/status"),
				),
				Entry("status invalid; content intent invalid",
					func(datum *image.Image) {
						datum.Status = pointer.FromString("")
						datum.ContentID = pointer.FromString(imageTest.RandomContentID())
						datum.ContentIntent = pointer.FromString("")
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", image.Statuses()), "/status"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", image.ContentIntents()), "/contentIntent"),
				),
				Entry("status invalid; content attributes missing",
					func(datum *image.Image) {
						datum.Status = pointer.FromString("")
						datum.ContentID = pointer.FromString(imageTest.RandomContentID())
						datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
						datum.ContentAttributes = nil
						datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", image.Statuses()), "/status"),
				),
				Entry("status invalid; content attributes empty",
					func(datum *image.Image) {
						datum.Status = pointer.FromString("")
						datum.ContentID = pointer.FromString(imageTest.RandomContentID())
						datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
						datum.ContentAttributes = image.NewContentAttributes()
						datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", image.Statuses()), "/status"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/contentAttributes/digestMD5"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/contentAttributes/mediaType"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/contentAttributes/width"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/contentAttributes/height"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/contentAttributes/size"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/contentAttributes/createdTime"),
				),
				Entry("status invalid; content attributes invalid",
					func(datum *image.Image) {
						datum.Status = pointer.FromString("")
						datum.ContentID = pointer.FromString(imageTest.RandomContentID())
						datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.ContentAttributes.DigestMD5 = pointer.FromString("")
						datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", image.Statuses()), "/status"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/contentAttributes/digestMD5"),
				),
				Entry("status invalid; renditions id missing",
					func(datum *image.Image) {
						datum.Status = pointer.FromString("")
						datum.ContentID = pointer.FromString(imageTest.RandomContentID())
						datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.RenditionsID = nil
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", image.Statuses()), "/status"),
				),
				Entry("status invalid; renditions id invalid",
					func(datum *image.Image) {
						datum.Status = pointer.FromString("")
						datum.ContentID = pointer.FromString(imageTest.RandomContentID())
						datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.RenditionsID = pointer.FromString("")
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", image.Statuses()), "/status"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/renditionsId"),
				),
				Entry("status invalid; renditions missing",
					func(datum *image.Image) {
						datum.Status = pointer.FromString("")
						datum.ContentID = pointer.FromString(imageTest.RandomContentID())
						datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
						datum.Renditions = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", image.Statuses()), "/status"),
				),
				Entry("status invalid; renditions empty",
					func(datum *image.Image) {
						datum.Status = pointer.FromString("")
						datum.ContentID = pointer.FromString(imageTest.RandomContentID())
						datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
						datum.Renditions = pointer.FromStringArray([]string{})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", image.Statuses()), "/status"),
				),
				Entry("status invalid; renditions element empty",
					func(datum *image.Image) {
						datum.Status = pointer.FromString("")
						datum.ContentID = pointer.FromString(imageTest.RandomContentID())
						datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
						(*datum.Renditions)[0] = ""
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", image.Statuses()), "/status"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/renditions/0"),
				),
				Entry("status invalid; renditions element not unique",
					func(datum *image.Image) {
						datum.Status = pointer.FromString("")
						datum.ContentID = pointer.FromString(imageTest.RandomContentID())
						datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
						(*datum.Renditions)[1] = (*datum.Renditions)[0]
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", image.Statuses()), "/status"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueDuplicate(), "/renditions/1"),
				),
				Entry("status invalid; content id, content intent, content attributes, renditions id, and renditions valid",
					func(datum *image.Image) {
						datum.Status = pointer.FromString("")
						datum.ContentID = pointer.FromString(imageTest.RandomContentID())
						datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", image.Statuses()), "/status"),
				),
				Entry("status available; content id, content intent, content attributes, renditions id, and renditions missing",
					func(datum *image.Image) {
						datum.Status = pointer.FromString("available")
						datum.ContentID = nil
						datum.ContentIntent = nil
						datum.ContentAttributes = nil
						datum.RenditionsID = nil
						datum.Renditions = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/contentIntent"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/contentAttributes"),
				),
				Entry("status available; content id missing",
					func(datum *image.Image) {
						datum.Status = pointer.FromString("available")
						datum.ContentID = nil
						datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
					},
				),
				Entry("status available; content id invalid",
					func(datum *image.Image) {
						datum.Status = pointer.FromString("available")
						datum.ContentID = pointer.FromString("")
						datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/contentId"),
				),
				Entry("status available; content intent missing",
					func(datum *image.Image) {
						datum.Status = pointer.FromString("available")
						datum.ContentID = pointer.FromString(imageTest.RandomContentID())
						datum.ContentIntent = nil
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/contentIntent"),
				),
				Entry("status available; content intent invalid",
					func(datum *image.Image) {
						datum.Status = pointer.FromString("available")
						datum.ContentID = pointer.FromString(imageTest.RandomContentID())
						datum.ContentIntent = pointer.FromString("")
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", image.ContentIntents()), "/contentIntent"),
				),
				Entry("status available; content attributes missing",
					func(datum *image.Image) {
						datum.Status = pointer.FromString("available")
						datum.ContentID = pointer.FromString(imageTest.RandomContentID())
						datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
						datum.ContentAttributes = nil
						datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/contentAttributes"),
				),
				Entry("status available; content attributes empty",
					func(datum *image.Image) {
						datum.Status = pointer.FromString("available")
						datum.ContentID = pointer.FromString(imageTest.RandomContentID())
						datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
						datum.ContentAttributes = image.NewContentAttributes()
						datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/contentAttributes/digestMD5"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/contentAttributes/mediaType"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/contentAttributes/width"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/contentAttributes/height"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/contentAttributes/size"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/contentAttributes/createdTime"),
				),
				Entry("status available; content attributes invalid",
					func(datum *image.Image) {
						datum.Status = pointer.FromString("available")
						datum.ContentID = pointer.FromString(imageTest.RandomContentID())
						datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.ContentAttributes.DigestMD5 = pointer.FromString("")
						datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/contentAttributes/digestMD5"),
				),
				Entry("status available; renditions id missing",
					func(datum *image.Image) {
						datum.Status = pointer.FromString("available")
						datum.ContentID = pointer.FromString(imageTest.RandomContentID())
						datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.RenditionsID = nil
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/renditions"),
				),
				Entry("status available; renditions id invalid",
					func(datum *image.Image) {
						datum.Status = pointer.FromString("available")
						datum.ContentID = pointer.FromString(imageTest.RandomContentID())
						datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.RenditionsID = pointer.FromString("")
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/renditionsId"),
				),
				Entry("status available; renditions missing",
					func(datum *image.Image) {
						datum.Status = pointer.FromString("available")
						datum.ContentID = pointer.FromString(imageTest.RandomContentID())
						datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
						datum.Renditions = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/renditions"),
				),
				Entry("status available; renditions empty",
					func(datum *image.Image) {
						datum.Status = pointer.FromString("available")
						datum.ContentID = pointer.FromString(imageTest.RandomContentID())
						datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
						datum.Renditions = pointer.FromStringArray([]string{})
					},
				),
				Entry("status available; renditions element empty",
					func(datum *image.Image) {
						datum.Status = pointer.FromString("available")
						datum.ContentID = pointer.FromString(imageTest.RandomContentID())
						datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
						(*datum.Renditions)[0] = ""
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/renditions/0"),
				),
				Entry("status available; renditions element not unique",
					func(datum *image.Image) {
						datum.Status = pointer.FromString("available")
						datum.ContentID = pointer.FromString(imageTest.RandomContentID())
						datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
						(*datum.Renditions)[1] = (*datum.Renditions)[0]
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueDuplicate(), "/renditions/1"),
				),
				Entry("status available; content id, content intent, content attributes, renditions id, and renditions valid",
					func(datum *image.Image) {
						datum.Status = pointer.FromString("available")
						datum.ContentID = pointer.FromString(imageTest.RandomContentID())
						datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
					},
				),
				Entry("status created; content id, content intent, content attributes, renditions id, and renditions missing",
					func(datum *image.Image) {
						datum.Status = pointer.FromString("created")
						datum.ContentID = nil
						datum.ContentIntent = nil
						datum.ContentAttributes = nil
						datum.RenditionsID = nil
						datum.Renditions = nil
					},
				),
				Entry("status created; content id missing",
					func(datum *image.Image) {
						datum.Status = pointer.FromString("created")
						datum.ContentID = nil
						datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentIntent"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentAttributes"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/renditionsId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/renditions"),
				),
				Entry("status created; content id invalid",
					func(datum *image.Image) {
						datum.Status = pointer.FromString("created")
						datum.ContentID = pointer.FromString("")
						datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentIntent"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentAttributes"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/renditionsId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/renditions"),
				),
				Entry("status created; content intent missing",
					func(datum *image.Image) {
						datum.Status = pointer.FromString("created")
						datum.ContentID = pointer.FromString(imageTest.RandomContentID())
						datum.ContentIntent = nil
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentAttributes"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/renditionsId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/renditions"),
				),
				Entry("status created; content intent invalid",
					func(datum *image.Image) {
						datum.Status = pointer.FromString("created")
						datum.ContentID = pointer.FromString(imageTest.RandomContentID())
						datum.ContentIntent = pointer.FromString("")
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentIntent"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentAttributes"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/renditionsId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/renditions"),
				),
				Entry("status created; content attributes missing",
					func(datum *image.Image) {
						datum.Status = pointer.FromString("created")
						datum.ContentID = pointer.FromString(imageTest.RandomContentID())
						datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
						datum.ContentAttributes = nil
						datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentIntent"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/renditionsId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/renditions"),
				),
				Entry("status created; content attributes empty",
					func(datum *image.Image) {
						datum.Status = pointer.FromString("created")
						datum.ContentID = pointer.FromString(imageTest.RandomContentID())
						datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
						datum.ContentAttributes = image.NewContentAttributes()
						datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentIntent"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentAttributes"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/renditionsId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/renditions"),
				),
				Entry("status created; content attributes invalid",
					func(datum *image.Image) {
						datum.Status = pointer.FromString("created")
						datum.ContentID = pointer.FromString(imageTest.RandomContentID())
						datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.ContentAttributes.DigestMD5 = pointer.FromString("")
						datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentIntent"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentAttributes"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/renditionsId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/renditions"),
				),
				Entry("status created; renditions id missing",
					func(datum *image.Image) {
						datum.Status = pointer.FromString("created")
						datum.ContentID = pointer.FromString(imageTest.RandomContentID())
						datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.RenditionsID = nil
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentIntent"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentAttributes"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/renditions"),
				),
				Entry("status created; renditions id invalid",
					func(datum *image.Image) {
						datum.Status = pointer.FromString("created")
						datum.ContentID = pointer.FromString(imageTest.RandomContentID())
						datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.RenditionsID = pointer.FromString("")
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentIntent"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentAttributes"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/renditionsId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/renditions"),
				),
				Entry("status created; renditions missing",
					func(datum *image.Image) {
						datum.Status = pointer.FromString("created")
						datum.ContentID = pointer.FromString(imageTest.RandomContentID())
						datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
						datum.Renditions = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentIntent"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentAttributes"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/renditionsId"),
				),
				Entry("status created; renditions empty",
					func(datum *image.Image) {
						datum.Status = pointer.FromString("created")
						datum.ContentID = pointer.FromString(imageTest.RandomContentID())
						datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
						datum.Renditions = pointer.FromStringArray([]string{})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentIntent"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentAttributes"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/renditionsId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/renditions"),
				),
				Entry("status created; renditions element empty",
					func(datum *image.Image) {
						datum.Status = pointer.FromString("created")
						datum.ContentID = pointer.FromString(imageTest.RandomContentID())
						datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
						(*datum.Renditions)[0] = ""
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentIntent"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentAttributes"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/renditionsId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/renditions"),
				),
				Entry("status created; renditions element not unique",
					func(datum *image.Image) {
						datum.Status = pointer.FromString("created")
						datum.ContentID = pointer.FromString(imageTest.RandomContentID())
						datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
						(*datum.Renditions)[1] = (*datum.Renditions)[0]
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentIntent"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentAttributes"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/renditionsId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/renditions"),
				),
				Entry("status created; content id, content intent, content attributes, renditions id, and renditions valid",
					func(datum *image.Image) {
						datum.Status = pointer.FromString("created")
						datum.ContentID = pointer.FromString(imageTest.RandomContentID())
						datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentIntent"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentAttributes"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/renditionsId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/renditions"),
				),
				Entry("name missing",
					func(datum *image.Image) { datum.Name = nil },
				),
				Entry("name empty",
					func(datum *image.Image) { datum.Name = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/name"),
				),
				Entry("name valid",
					func(datum *image.Image) {
						datum.Name = pointer.FromString(imageTest.RandomName())
					},
				),
				Entry("name valid; length in range (upper)",
					func(datum *image.Image) {
						datum.Name = pointer.FromString(test.RandomStringFromRange(100, 100))
					},
				),
				Entry("name invalid; length out of range (upper)",
					func(datum *image.Image) { datum.Name = pointer.FromString(test.RandomStringFromRange(101, 101)) },
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/name"),
				),
				Entry("created time missing",
					func(datum *image.Image) { datum.CreatedTime = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/createdTime"),
				),
				Entry("created time zero",
					func(datum *image.Image) { datum.CreatedTime = pointer.FromTime(time.Time{}) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/createdTime"),
				),
				Entry("created time after now",
					func(datum *image.Image) {
						datum.CreatedTime = pointer.FromTime(test.FutureFarTime())
						datum.ModifiedTime = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(test.FutureFarTime()), "/createdTime"),
				),
				Entry("created time valid",
					func(datum *image.Image) {
						datum.CreatedTime = pointer.FromTime(test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now()).Truncate(time.Second))
						datum.ModifiedTime = pointer.FromTime(test.RandomTimeFromRange(*datum.CreatedTime, time.Now()).Truncate(time.Second))
					},
				),
				Entry("modified time missing",
					func(datum *image.Image) { datum.ModifiedTime = nil },
				),
				Entry("modified time before created time",
					func(datum *image.Image) {
						datum.CreatedTime = pointer.FromTime(test.PastNearTime())
						datum.ModifiedTime = pointer.FromTime(test.PastFarTime())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotAfter(test.PastFarTime(), test.PastNearTime()), "/modifiedTime"),
				),
				Entry("modified time after now",
					func(datum *image.Image) { datum.ModifiedTime = pointer.FromTime(test.FutureFarTime()) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(test.FutureFarTime()), "/modifiedTime"),
				),
				Entry("modified time valid",
					func(datum *image.Image) {
						datum.ModifiedTime = pointer.FromTime(test.RandomTimeFromRange(*datum.CreatedTime, time.Now()).Truncate(time.Second))
					},
				),
				Entry("deleted time missing",
					func(datum *image.Image) { datum.DeletedTime = nil },
				),
				Entry("deleted time before created time",
					func(datum *image.Image) {
						datum.CreatedTime = pointer.FromTime(test.PastNearTime())
						datum.ModifiedTime = pointer.FromTime(test.RandomTimeFromRange(*datum.CreatedTime, time.Now()).Truncate(time.Second))
						datum.DeletedTime = pointer.FromTime(test.PastFarTime())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotAfter(test.PastFarTime(), test.PastNearTime()), "/deletedTime"),
				),
				Entry("deleted time after now",
					func(datum *image.Image) { datum.DeletedTime = pointer.FromTime(test.FutureFarTime()) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(test.FutureFarTime()), "/deletedTime"),
				),
				Entry("deleted time valid",
					func(datum *image.Image) {
						datum.DeletedTime = pointer.FromTime(test.RandomTimeFromRange(*datum.CreatedTime, time.Now()).Truncate(time.Second))
					},
				),
				Entry("revision missing",
					func(datum *image.Image) {
						datum.Revision = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/revision"),
				),
				Entry("revision out of range (lower)",
					func(datum *image.Image) {
						datum.Revision = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "/revision"),
				),
				Entry("revision in range (lower)",
					func(datum *image.Image) {
						datum.Revision = pointer.FromInt(0)
					},
				),
				Entry("multiple errors",
					func(datum *image.Image) {
						datum.ID = nil
						datum.UserID = nil
						datum.Status = nil
						datum.Name = pointer.FromString("")
						datum.ContentID = pointer.FromString("")
						datum.ContentIntent = pointer.FromString("")
						datum.ContentAttributes = imageTest.RandomContentAttributes()
						datum.ContentAttributes.DigestMD5 = pointer.FromString("")
						datum.RenditionsID = pointer.FromString("")
						datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
						(*datum.Renditions)[0] = ""
						datum.CreatedTime = nil
						datum.ModifiedTime = pointer.FromTime(test.FutureFarTime())
						datum.DeletedTime = pointer.FromTime(test.FutureFarTime())
						datum.Revision = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/id"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/userId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/status"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/name"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/contentId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", image.ContentIntents()), "/contentIntent"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/contentAttributes/digestMD5"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/renditionsId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/renditions/0"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/createdTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(test.FutureFarTime()), "/modifiedTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(test.FutureFarTime()), "/deletedTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "/revision"),
				),
			)
		})

		Context("with new image", func() {
			var datum *image.Image

			BeforeEach(func() {
				datum = imageTest.RandomImage()
			})

			Context("HasContent", func() {
				It("returns false when status is missing", func() {
					datum.Status = nil
					datum.ContentID = pointer.FromString(imageTest.RandomContentID())
					datum.ContentIntent = pointer.FromString(image.ContentIntentOriginal)
					datum.ContentAttributes = imageTest.RandomContentAttributes()
					Expect(datum.HasContent()).To(BeFalse())
				})

				It("returns false when status is not available", func() {
					datum.Status = pointer.FromString(image.StatusCreated)
					datum.ContentID = pointer.FromString(imageTest.RandomContentID())
					datum.ContentIntent = pointer.FromString(image.ContentIntentOriginal)
					datum.ContentAttributes = imageTest.RandomContentAttributes()
					Expect(datum.HasContent()).To(BeFalse())
				})

				It("returns false when content id is missing", func() {
					datum.Status = pointer.FromString(image.StatusAvailable)
					datum.ContentID = nil
					datum.ContentIntent = pointer.FromString(image.ContentIntentOriginal)
					datum.ContentAttributes = imageTest.RandomContentAttributes()
					Expect(datum.HasContent()).To(BeFalse())
				})

				It("returns false when content intent is missing", func() {
					datum.Status = pointer.FromString(image.StatusAvailable)
					datum.ContentID = pointer.FromString(imageTest.RandomContentID())
					datum.ContentIntent = nil
					datum.ContentAttributes = imageTest.RandomContentAttributes()
					Expect(datum.HasContent()).To(BeFalse())
				})

				It("returns false when content attributes is missing", func() {
					datum.Status = pointer.FromString(image.StatusAvailable)
					datum.ContentID = pointer.FromString(imageTest.RandomContentID())
					datum.ContentIntent = pointer.FromString(image.ContentIntentOriginal)
					datum.ContentAttributes = nil
					Expect(datum.HasContent()).To(BeFalse())
				})

				It("returns true as expected", func() {
					datum.Status = pointer.FromString(image.StatusAvailable)
					datum.ContentID = pointer.FromString(imageTest.RandomContentID())
					datum.ContentIntent = pointer.FromString(image.ContentIntentOriginal)
					datum.ContentAttributes = imageTest.RandomContentAttributes()
					Expect(datum.HasContent()).To(BeTrue())
				})
			})

			Context("HasRendition", func() {
				var rendition image.Rendition

				BeforeEach(func() {
					rendition = *imageTest.RandomRendition()
				})

				It("returns false when the renditions is missing", func() {
					datum.Renditions = nil
					Expect(datum.HasRendition(rendition)).To(BeFalse())
				})

				It("returns false when the renditions is empty", func() {
					datum.Renditions = pointer.FromStringArray([]string{})
					Expect(datum.HasRendition(rendition)).To(BeFalse())
				})

				It("returns false when the renditions is does not contain the rendition", func() {
					datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
					Expect(datum.HasRendition(rendition)).To(BeFalse())
				})

				It("returns false when the renditions is does contain the rendition", func() {
					datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
					(*datum.Renditions) = append(*datum.Renditions, rendition.String())
					Expect(datum.HasRendition(rendition)).To(BeTrue())
				})
			})

			Context("Sanitize", func() {
				var original *image.Image

				BeforeEach(func() {
					datum.Status = pointer.FromString(image.StatusAvailable)
					datum.ContentID = pointer.FromString(imageTest.RandomContentID())
					datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
					datum.ContentAttributes = imageTest.RandomContentAttributes()
					datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
					datum.Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
					original = imageTest.CloneImage(datum)
				})

				It("does sanitize renditions if details is missing", func() {
					original.ContentID = nil
					original.RenditionsID = nil
					original.Renditions = nil
					datum.Sanitize(nil)
					Expect(datum).To(Equal(original))
				})

				It("does sanitize renditions if details is not service", func() {
					original.ContentID = nil
					original.RenditionsID = nil
					original.Renditions = nil
					datum.Sanitize(request.NewDetails(request.MethodSessionToken, userTest.RandomID(), authTest.NewSessionToken()))
					Expect(datum).To(Equal(original))
				})

				It("does not sanitize renditions if details is service", func() {
					datum.Sanitize(request.NewDetails(request.MethodServiceSecret, "", authTest.NewServiceSecret()))
					Expect(datum).To(Equal(original))
				})
			})
		})
	})

	Context("Images", func() {
		Context("Sanitize", func() {
			var datum image.Images
			var original image.Images

			BeforeEach(func() {
				datum = imageTest.RandomImages(0, 2)
				for index := range datum {
					datum[index].Status = pointer.FromString(image.StatusAvailable)
					datum[index].ContentID = pointer.FromString(imageTest.RandomContentID())
					datum[index].ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
					datum[index].ContentAttributes = imageTest.RandomContentAttributes()
					datum[index].RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
					datum[index].Renditions = pointer.FromStringArray(imageTest.RandomRenditionsAsStrings())
				}
				original = imageTest.CloneImages(datum)
			})

			It("does sanitize renditions if details is missing", func() {
				for index := range original {
					original[index].ContentID = nil
					original[index].RenditionsID = nil
					original[index].Renditions = nil
				}
				datum.Sanitize(nil)
				Expect(datum).To(Equal(original))
			})

			It("does sanitize renditions if details is not service", func() {
				for index := range original {
					original[index].ContentID = nil
					original[index].RenditionsID = nil
					original[index].Renditions = nil
				}
				datum.Sanitize(request.NewDetails(request.MethodSessionToken, userTest.RandomID(), authTest.NewSessionToken()))
				Expect(datum).To(Equal(original))
			})

			It("does not sanitize renditions if details is service", func() {
				datum.Sanitize(request.NewDetails(request.MethodServiceSecret, "", authTest.NewServiceSecret()))
				Expect(datum).To(Equal(original))
			})
		})
	})

	Context("ID", func() {
		Context("NewID", func() {
			It("returns a string of 32 lowercase hexidecimal characters", func() {
				Expect(image.NewID()).To(MatchRegexp("^[0-9a-f]{32}$"))
			})

			It("returns different IDs for each invocation", func() {
				Expect(image.NewID()).ToNot(Equal(image.NewID()))
			})
		})

		Context("IsValidID, IDValidator, and ValidateID", func() {
			DescribeTable("return the expected results when the input",
				func(value string, expectedErrors ...error) {
					Expect(image.IsValidID(value)).To(Equal(len(expectedErrors) == 0))
					errorReporter := structureTest.NewErrorReporter()
					image.IDValidator(value, errorReporter)
					errorsTest.ExpectEqual(errorReporter.Error(), expectedErrors...)
					errorsTest.ExpectEqual(image.ValidateID(value), expectedErrors...)
				},
				Entry("is an empty", "", structureValidator.ErrorValueEmpty()),
				Entry("has string length out of range (lower)", "0123456789abcdefghijklmnopqrstu", image.ErrorValueStringAsIDNotValid("0123456789abcdefghijklmnopqrstu")),
				Entry("has string length in range", test.RandomStringFromRangeAndCharset(32, 32, test.CharsetLowercase+test.CharsetNumeric)),
				Entry("has string length out of range (upper)", "0123456789abcdefghijklmnopqrstuvw", image.ErrorValueStringAsIDNotValid("0123456789abcdefghijklmnopqrstuvw")),
				Entry("has uppercase characters", "0123456789ABCDEFGHIJKLMNOPQRSTUV", image.ErrorValueStringAsIDNotValid("0123456789ABCDEFGHIJKLMNOPQRSTUV")),
				Entry("has symbols", "0123456789!@#$%^abcdefghijklmnop", image.ErrorValueStringAsIDNotValid("0123456789!@#$%^abcdefghijklmnop")),
				Entry("has whitespace", "0123456789      abcdefghijklmnop", image.ErrorValueStringAsIDNotValid("0123456789      abcdefghijklmnop")),
			)
		})
	})

	Context("ContentID", func() {
		Context("NewContentID", func() {
			It("returns a string of 16 lowercase hexidecimal characters", func() {
				Expect(image.NewContentID()).To(MatchRegexp("^[0-9a-f]{16}$"))
			})

			It("returns different ContentIDs for each invocation", func() {
				Expect(image.NewContentID()).ToNot(Equal(image.NewContentID()))
			})
		})

		Context("IsValidContentID, ContentIDValidator, and ValidateContentID", func() {
			DescribeTable("return the expected results when the input",
				func(value string, expectedErrors ...error) {
					Expect(image.IsValidContentID(value)).To(Equal(len(expectedErrors) == 0))
					errorReporter := structureTest.NewErrorReporter()
					image.ContentIDValidator(value, errorReporter)
					errorsTest.ExpectEqual(errorReporter.Error(), expectedErrors...)
					errorsTest.ExpectEqual(image.ValidateContentID(value), expectedErrors...)
				},
				Entry("is an empty", "", structureValidator.ErrorValueEmpty()),
				Entry("has string length out of range (lower)", "0123456789abcde", image.ErrorValueStringAsContentIDNotValid("0123456789abcde")),
				Entry("has string length in range", test.RandomStringFromRangeAndCharset(16, 16, test.CharsetLowercase+test.CharsetNumeric)),
				Entry("has string length out of range (upper)", "0123456789abcdefg", image.ErrorValueStringAsContentIDNotValid("0123456789abcdefg")),
				Entry("has uppercase characters", "0123456789ABCDEF", image.ErrorValueStringAsContentIDNotValid("0123456789ABCDEF")),
				Entry("has symbols", "0123456789!@#$%^", image.ErrorValueStringAsContentIDNotValid("0123456789!@#$%^")),
				Entry("has whitespace", "0123456789      ", image.ErrorValueStringAsContentIDNotValid("0123456789      ")),
			)
		})
	})

	Context("RenditionsID", func() {
		Context("NewRenditionsID", func() {
			It("returns a string of 16 lowercase hexidecimal characters", func() {
				Expect(image.NewRenditionsID()).To(MatchRegexp("^[0-9a-f]{16}$"))
			})

			It("returns different RenditionsIDs for each invocation", func() {
				Expect(image.NewRenditionsID()).ToNot(Equal(image.NewRenditionsID()))
			})
		})

		Context("IsValidRenditionsID, RenditionsIDValidator, and ValidateRenditionsID", func() {
			DescribeTable("return the expected results when the input",
				func(value string, expectedErrors ...error) {
					Expect(image.IsValidRenditionsID(value)).To(Equal(len(expectedErrors) == 0))
					errorReporter := structureTest.NewErrorReporter()
					image.RenditionsIDValidator(value, errorReporter)
					errorsTest.ExpectEqual(errorReporter.Error(), expectedErrors...)
					errorsTest.ExpectEqual(image.ValidateRenditionsID(value), expectedErrors...)
				},
				Entry("is an empty", "", structureValidator.ErrorValueEmpty()),
				Entry("has string length out of range (lower)", "0123456789abcde", image.ErrorValueStringAsRenditionsIDNotValid("0123456789abcde")),
				Entry("has string length in range", test.RandomStringFromRangeAndCharset(16, 16, test.CharsetLowercase+test.CharsetNumeric)),
				Entry("has string length out of range (upper)", "0123456789abcdefg", image.ErrorValueStringAsRenditionsIDNotValid("0123456789abcdefg")),
				Entry("has uppercase characters", "0123456789ABCDEF", image.ErrorValueStringAsRenditionsIDNotValid("0123456789ABCDEF")),
				Entry("has symbols", "0123456789!@#$%^", image.ErrorValueStringAsRenditionsIDNotValid("0123456789!@#$%^")),
				Entry("has whitespace", "0123456789      ", image.ErrorValueStringAsRenditionsIDNotValid("0123456789      ")),
			)
		})
	})

	Context("ContentIntent", func() {
		Context("IsValidContentIntent, ContentIntentValidator, and ValidateContentIntent", func() {
			DescribeTable("return the expected results when the input",
				func(value string, expectedErrors ...error) {
					Expect(image.IsValidContentIntent(value)).To(Equal(len(expectedErrors) == 0))
					errorReporter := structureTest.NewErrorReporter()
					image.ContentIntentValidator(value, errorReporter)
					errorsTest.ExpectEqual(errorReporter.Error(), expectedErrors...)
					errorsTest.ExpectEqual(image.ValidateContentIntent(value), expectedErrors...)
				},
				Entry("is an empty", "", structureValidator.ErrorValueEmpty()),
				Entry("is invalid", "invalid", image.ErrorValueStringAsContentIntentNotValid("invalid")),
				Entry("is alternate", "alternate"),
				Entry("is original", "original"),
			)
		})
	})

	Context("MediaType", func() {
		Context("IsValidMediaType, MediaTypeValidator, and ValidateMediaType", func() {
			DescribeTable("return the expected results when the input",
				func(value string, expectedErrors ...error) {
					Expect(image.IsValidMediaType(value)).To(Equal(len(expectedErrors) == 0))
					errorReporter := structureTest.NewErrorReporter()
					image.MediaTypeValidator(value, errorReporter)
					errorsTest.ExpectEqual(errorReporter.Error(), expectedErrors...)
					errorsTest.ExpectEqual(image.ValidateMediaType(value), expectedErrors...)
				},
				Entry("is empty", "", structureValidator.ErrorValueEmpty()),
				Entry("is invalid", "/", net.ErrorValueStringAsMediaTypeNotValid("/")),
				Entry("is unsupported", "application/octet-stream", request.ErrorMediaTypeNotSupported("application/octet-stream")),
				Entry("is image/jpeg", "image/jpeg"),
				Entry("is image/png", "image/png"),
			)
		})

		Context("MediaTypeSupportsQuality", func() {
			DescribeTable("returns expected value for media type",
				func(value string, expectedResult bool) {
					Expect(image.MediaTypeSupportsQuality(value)).To(Equal(expectedResult))
				},
				Entry("is empty", "", false),
				Entry("is invalid", "/", false),
				Entry("is unsupported", "application/octet-stream", false),
				Entry("is image/jpeg", "image/jpeg", true),
				Entry("is image/png", "image/png", false),
			)
		})

		Context("MediaTypeSupportsTransparency", func() {
			DescribeTable("returns expected value for media type",
				func(value string, expectedResult bool) {
					Expect(image.MediaTypeSupportsTransparency(value)).To(Equal(expectedResult))
				},
				Entry("is empty", "", false),
				Entry("is invalid", "/", false),
				Entry("is unsupported", "application/octet-stream", false),
				Entry("is image/jpeg", "image/jpeg", false),
				Entry("is image/png", "image/png", true),
			)
		})

		Context("MediaTypeFromExtension", func() {
			DescribeTable("returns expected value for media type",
				func(value string, expectedResult string, expectedValid bool) {
					result, valid := image.MediaTypeFromExtension(value)
					Expect(valid).To(Equal(expectedValid))
					Expect(result).To(Equal(expectedResult))
				},
				Entry("is empty", "", "", false),
				Entry("is invalid", "/", "", false),
				Entry("is unsupported", "bin", "", false),
				Entry("is jpeg", "jpeg", "image/jpeg", true),
				Entry("is jpg", "jpg", "image/jpeg", true),
				Entry("is png", "png", "image/png", true),
			)
		})

		Context("ExtensionFromMediaType", func() {
			DescribeTable("returns expected value for media type",
				func(value string, expectedResult string, expectedValid bool) {
					result, valid := image.ExtensionFromMediaType(value)
					Expect(valid).To(Equal(expectedValid))
					Expect(result).To(Equal(expectedResult))
				},
				Entry("is empty", "", "", false),
				Entry("is invalid", "/", "", false),
				Entry("is unsupported", "application/octet-stream", "", false),
				Entry("is image/jpeg", "image/jpeg", "jpeg", true),
				Entry("is image/png", "image/png", "png", true),
			)
		})
	})

	Context("Mode", func() {
		Context("NormalizeMode", func() {
			DescribeTable("returns expected value for media type",
				func(value string, expectedResult string) {
					Expect(image.NormalizeMode(value)).To(Equal(expectedResult))
				},
				Entry("is empty", "", ""),
				Entry("is invalid", "/", "/"),
				Entry("is fill", "fill", "fill"),
				Entry("is fillDown", "fillDown", "fill"),
				Entry("is fit", "fit", "fit"),
				Entry("is fitDown", "fitDown", "fit"),
				Entry("is pad", "pad", "pad"),
				Entry("is padDown", "padDown", "pad"),
				Entry("is scale", "scale", "scale"),
				Entry("is scaleDown", "scaleDown", "scale"),
			)
		})
	})

	Context("Color", func() {
		Context("ParseColor", func() {
			DescribeTable("return the expected results when the input",
				func(value string, expectedResult *image.Color, expectedErrors ...error) {
					result, err := image.ParseColor(value)
					Expect(result).To(Equal(expectedResult))
					errorsTest.ExpectEqual(err, expectedErrors...)
				},
				Entry("is empty", "", nil, image.ErrorValueStringAsColorNotValid("")),
				Entry("is invalid", "AaFf012X", nil, image.ErrorValueStringAsColorNotValid("AaFf012X")),
				Entry("is invalid less than three bytes", "AaFf0", nil, image.ErrorValueStringAsColorNotValid("AaFf0")),
				Entry("is valid three bytes", "AaFf01", &image.Color{NRGBA: color.NRGBA{R: 0xAA, G: 0xFF, B: 0x01, A: 0xFF}}),
				Entry("is invalid between three and four bytes", "AaFf012", nil, image.ErrorValueStringAsColorNotValid("AaFf012")),
				Entry("is valid four bytes", "AaFf0123", &image.Color{NRGBA: color.NRGBA{R: 0xAA, G: 0xFF, B: 0x01, A: 0x23}}),
				Entry("is invalid more than four bytes", "AaFf01234", nil, image.ErrorValueStringAsColorNotValid("AaFf01234")),
				Entry("has prefix and is invalid", "0xAaFf012X", nil, image.ErrorValueStringAsColorNotValid("0xAaFf012X")),
				Entry("has prefix and is invalid less than three bytes", "0xAaFf0", nil, image.ErrorValueStringAsColorNotValid("0xAaFf0")),
				Entry("has prefix and is valid three bytes", "0xAaFf01", &image.Color{NRGBA: color.NRGBA{R: 0xAA, G: 0xFF, B: 0x01, A: 0xFF}}),
				Entry("has prefix and is invalid between three and four bytes", "0xAaFf012", nil, image.ErrorValueStringAsColorNotValid("0xAaFf012")),
				Entry("has prefix and is valid four bytes", "0xAaFf0123", &image.Color{NRGBA: color.NRGBA{R: 0xAA, G: 0xFF, B: 0x01, A: 0x23}}),
			)
		})

		Context("NewColor", func() {
			DescribeTable("return the expected results when the input",
				func(r int, g int, b int, a int, expectedResult *image.Color) {
					Expect(image.NewColor(uint8(r), uint8(g), uint8(b), uint8(a))).To(Equal(expectedResult))
				},
				Entry("is zero", 0x00, 0x00, 0x00, 0x00, &image.Color{NRGBA: color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x00}}),
				Entry("is black", 0x00, 0x00, 0x00, 0xFF, &image.Color{NRGBA: color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xFF}}),
				Entry("is white", 0xFF, 0xFF, 0xFF, 0xFF, &image.Color{NRGBA: color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}}),
				Entry("is any color", 0x12, 0x34, 0x56, 0x78, &image.Color{NRGBA: color.NRGBA{R: 0x12, G: 0x34, B: 0x56, A: 0x78}}),
			)
		})

		Context("String", func() {
			DescribeTable("return the expected results when the input",
				func(value *image.Color, expectedResult string) {
					Expect(value.String()).To(Equal(expectedResult))
				},
				Entry("is zero", &image.Color{NRGBA: color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x00}}, "00000000"),
				Entry("is black", &image.Color{NRGBA: color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xFF}}, "000000ff"),
				Entry("is white", &image.Color{NRGBA: color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}}, "ffffffff"),
				Entry("is any color", &image.Color{NRGBA: color.NRGBA{R: 0x12, G: 0x34, B: 0x56, A: 0x78}}, "12345678"),
			)
		})
	})
})
