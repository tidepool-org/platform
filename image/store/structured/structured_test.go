package structured_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/crypto"
	cryptoTest "github.com/tidepool-org/platform/crypto/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/image"
	imageStoreStructured "github.com/tidepool-org/platform/image/store/structured"
	imageStoreStructuredTest "github.com/tidepool-org/platform/image/store/structured/test"
	imageTest "github.com/tidepool-org/platform/image/test"
	"github.com/tidepool-org/platform/net"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("Structured", func() {
	Context("NewUpdate", func() {
		It("returns successfully with default values", func() {
			Expect(imageStoreStructured.NewUpdate()).To(Equal(&imageStoreStructured.Update{}))
		})
	})

	Context("Validate", func() {
		DescribeTable("validates the datum",
			func(mutator func(datum *imageStoreStructured.Update), expectedErrors ...error) {
				datum := imageStoreStructuredTest.RandomUpdate()
				mutator(datum)
				errorsTest.ExpectEqual(structureValidator.New().Validate(datum), expectedErrors...)
			},
			Entry("succeeds",
				func(datum *imageStoreStructured.Update) {},
			),
			Entry("metadata missing",
				func(datum *imageStoreStructured.Update) { datum.Metadata = nil },
			),
			Entry("metadata invalid",
				func(datum *imageStoreStructured.Update) { datum.Metadata.Name = pointer.FromString("") },
				errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/metadata/name"),
			),
			Entry("metadata valid",
				func(datum *imageStoreStructured.Update) { datum.Metadata = imageTest.RandomMetadata() },
			),
			Entry("content id missing; content intent, content attributes, renditions id, and rendition missing",
				func(datum *imageStoreStructured.Update) {
					datum.ContentID = nil
					datum.ContentIntent = nil
					datum.ContentAttributes = nil
					datum.RenditionsID = nil
					datum.Rendition = nil
				},
			),
			Entry("content id missing; content intent missing",
				func(datum *imageStoreStructured.Update) {
					datum.ContentID = nil
					datum.ContentIntent = nil
					datum.ContentAttributes = imageStoreStructuredTest.RandomContentAttributes()
					datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
					datum.Rendition = pointer.FromString(imageTest.RandomRenditionString())
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentAttributes"),
			),
			Entry("content id missing; content intent invalid",
				func(datum *imageStoreStructured.Update) {
					datum.ContentID = nil
					datum.ContentIntent = pointer.FromString("")
					datum.ContentAttributes = imageStoreStructuredTest.RandomContentAttributes()
					datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
					datum.Rendition = pointer.FromString(imageTest.RandomRenditionString())
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentIntent"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentAttributes"),
			),
			Entry("content id missing; content attributes missing",
				func(datum *imageStoreStructured.Update) {
					datum.ContentID = nil
					datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
					datum.ContentAttributes = nil
					datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
					datum.Rendition = pointer.FromString(imageTest.RandomRenditionString())
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentIntent"),
			),
			Entry("content id missing; content attributes empty",
				func(datum *imageStoreStructured.Update) {
					datum.ContentID = nil
					datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
					datum.ContentAttributes = imageStoreStructured.NewContentAttributes()
					datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
					datum.Rendition = pointer.FromString(imageTest.RandomRenditionString())
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentIntent"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentAttributes"),
			),
			Entry("content id missing; content attributes invalid",
				func(datum *imageStoreStructured.Update) {
					datum.ContentID = nil
					datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
					datum.ContentAttributes = imageStoreStructuredTest.RandomContentAttributes()
					datum.ContentAttributes.DigestMD5 = pointer.FromString("")
					datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
					datum.Rendition = pointer.FromString(imageTest.RandomRenditionString())
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentIntent"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentAttributes"),
			),
			Entry("content id missing; renditions id missing",
				func(datum *imageStoreStructured.Update) {
					datum.ContentID = nil
					datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
					datum.ContentAttributes = imageStoreStructuredTest.RandomContentAttributes()
					datum.RenditionsID = nil
					datum.Rendition = pointer.FromString(imageTest.RandomRenditionString())
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentIntent"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentAttributes"),
			),
			Entry("content id missing; renditions id invalid",
				func(datum *imageStoreStructured.Update) {
					datum.ContentID = nil
					datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
					datum.ContentAttributes = imageStoreStructuredTest.RandomContentAttributes()
					datum.RenditionsID = pointer.FromString("")
					datum.Rendition = pointer.FromString(imageTest.RandomRenditionString())
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentIntent"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentAttributes"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/renditionsId"),
			),
			Entry("content id missing; rendition missing",
				func(datum *imageStoreStructured.Update) {
					datum.ContentID = nil
					datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
					datum.ContentAttributes = imageStoreStructuredTest.RandomContentAttributes()
					datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
					datum.Rendition = nil
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentIntent"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentAttributes"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rendition"),
			),
			Entry("content id missing; rendition invalid",
				func(datum *imageStoreStructured.Update) {
					datum.ContentID = nil
					datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
					datum.ContentAttributes = imageStoreStructuredTest.RandomContentAttributes()
					datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
					datum.Rendition = pointer.FromString("")
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentIntent"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentAttributes"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/rendition"),
			),
			Entry("content id missing; content id, content intent, content attributes, renditions id, and rendition valid",
				func(datum *imageStoreStructured.Update) {
					datum.ContentID = nil
					datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
					datum.ContentAttributes = imageStoreStructuredTest.RandomContentAttributes()
					datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
					datum.Rendition = pointer.FromString(imageTest.RandomRenditionString())
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentIntent"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/contentAttributes"),
			),
			Entry("content id invalid; content intent, content attributes, renditions id, and rendition missing",
				func(datum *imageStoreStructured.Update) {
					datum.ContentID = pointer.FromString("")
					datum.ContentIntent = nil
					datum.ContentAttributes = nil
					datum.RenditionsID = nil
					datum.Rendition = nil
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/contentId"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/contentIntent"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/contentAttributes"),
			),
			Entry("content id invalid; content intent missing",
				func(datum *imageStoreStructured.Update) {
					datum.ContentID = pointer.FromString("")
					datum.ContentIntent = nil
					datum.ContentAttributes = imageStoreStructuredTest.RandomContentAttributes()
					datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
					datum.Rendition = pointer.FromString(imageTest.RandomRenditionString())
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/contentId"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/contentIntent"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/renditionsId"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/rendition"),
			),
			Entry("content id invalid; content intent invalid",
				func(datum *imageStoreStructured.Update) {
					datum.ContentID = pointer.FromString("")
					datum.ContentIntent = pointer.FromString("")
					datum.ContentAttributes = imageStoreStructuredTest.RandomContentAttributes()
					datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
					datum.Rendition = pointer.FromString(imageTest.RandomRenditionString())
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/contentId"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", image.ContentIntents()), "/contentIntent"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/renditionsId"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/rendition"),
			),
			Entry("content id invalid; content attributes missing",
				func(datum *imageStoreStructured.Update) {
					datum.ContentID = pointer.FromString("")
					datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
					datum.ContentAttributes = nil
					datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
					datum.Rendition = pointer.FromString(imageTest.RandomRenditionString())
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/contentId"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/contentAttributes"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/renditionsId"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/rendition"),
			),
			Entry("content id invalid; content attributes empty",
				func(datum *imageStoreStructured.Update) {
					datum.ContentID = pointer.FromString("")
					datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
					datum.ContentAttributes = imageStoreStructured.NewContentAttributes()
					datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
					datum.Rendition = pointer.FromString(imageTest.RandomRenditionString())
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/contentId"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/contentAttributes/digestMD5"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/contentAttributes/mediaType"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/contentAttributes/width"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/contentAttributes/height"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/contentAttributes/size"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/renditionsId"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/rendition"),
			),
			Entry("content id invalid; content attributes invalid",
				func(datum *imageStoreStructured.Update) {
					datum.ContentID = pointer.FromString("")
					datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
					datum.ContentAttributes = imageStoreStructuredTest.RandomContentAttributes()
					datum.ContentAttributes.DigestMD5 = pointer.FromString("")
					datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
					datum.Rendition = pointer.FromString(imageTest.RandomRenditionString())
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/contentId"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/contentAttributes/digestMD5"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/renditionsId"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/rendition"),
			),
			Entry("content id invalid; renditions id missing",
				func(datum *imageStoreStructured.Update) {
					datum.ContentID = pointer.FromString("")
					datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
					datum.ContentAttributes = imageStoreStructuredTest.RandomContentAttributes()
					datum.RenditionsID = nil
					datum.Rendition = pointer.FromString(imageTest.RandomRenditionString())
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/contentId"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/rendition"),
			),
			Entry("content id invalid; renditions id invalid",
				func(datum *imageStoreStructured.Update) {
					datum.ContentID = pointer.FromString("")
					datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
					datum.ContentAttributes = imageStoreStructuredTest.RandomContentAttributes()
					datum.RenditionsID = pointer.FromString("")
					datum.Rendition = pointer.FromString(imageTest.RandomRenditionString())
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/contentId"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/renditionsId"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/rendition"),
			),
			Entry("content id invalid; rendition missing",
				func(datum *imageStoreStructured.Update) {
					datum.ContentID = pointer.FromString("")
					datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
					datum.ContentAttributes = imageStoreStructuredTest.RandomContentAttributes()
					datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
					datum.Rendition = nil
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/contentId"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/renditionsId"),
			),
			Entry("content id invalid; rendition invalid",
				func(datum *imageStoreStructured.Update) {
					datum.ContentID = pointer.FromString("")
					datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
					datum.ContentAttributes = imageStoreStructuredTest.RandomContentAttributes()
					datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
					datum.Rendition = pointer.FromString("")
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/contentId"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/renditionsId"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/rendition"),
			),
			Entry("content id invalid; content id, content intent, content attributes, renditions id, and rendition valid",
				func(datum *imageStoreStructured.Update) {
					datum.ContentID = pointer.FromString("")
					datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
					datum.ContentAttributes = imageStoreStructuredTest.RandomContentAttributes()
					datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
					datum.Rendition = pointer.FromString(imageTest.RandomRenditionString())
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/contentId"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/renditionsId"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/rendition"),
			),
			Entry("content id valid; content intent, content attributes, renditions id, and rendition missing",
				func(datum *imageStoreStructured.Update) {
					datum.ContentID = pointer.FromString(imageTest.RandomContentID())
					datum.ContentIntent = nil
					datum.ContentAttributes = nil
					datum.RenditionsID = nil
					datum.Rendition = nil
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/contentIntent"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/contentAttributes"),
			),
			Entry("content id valid; content intent missing",
				func(datum *imageStoreStructured.Update) {
					datum.ContentID = pointer.FromString(imageTest.RandomContentID())
					datum.ContentIntent = nil
					datum.ContentAttributes = imageStoreStructuredTest.RandomContentAttributes()
					datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
					datum.Rendition = pointer.FromString(imageTest.RandomRenditionString())
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/contentIntent"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/renditionsId"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/rendition"),
			),
			Entry("content id valid; content intent invalid",
				func(datum *imageStoreStructured.Update) {
					datum.ContentID = pointer.FromString(imageTest.RandomContentID())
					datum.ContentIntent = pointer.FromString("")
					datum.ContentAttributes = imageStoreStructuredTest.RandomContentAttributes()
					datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
					datum.Rendition = pointer.FromString(imageTest.RandomRenditionString())
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", image.ContentIntents()), "/contentIntent"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/renditionsId"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/rendition"),
			),
			Entry("content id valid; content attributes missing",
				func(datum *imageStoreStructured.Update) {
					datum.ContentID = pointer.FromString(imageTest.RandomContentID())
					datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
					datum.ContentAttributes = nil
					datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
					datum.Rendition = pointer.FromString(imageTest.RandomRenditionString())
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/contentAttributes"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/renditionsId"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/rendition"),
			),
			Entry("content id valid; content attributes empty",
				func(datum *imageStoreStructured.Update) {
					datum.ContentID = pointer.FromString(imageTest.RandomContentID())
					datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
					datum.ContentAttributes = imageStoreStructured.NewContentAttributes()
					datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
					datum.Rendition = pointer.FromString(imageTest.RandomRenditionString())
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/contentAttributes/digestMD5"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/contentAttributes/mediaType"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/contentAttributes/width"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/contentAttributes/height"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/contentAttributes/size"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/renditionsId"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/rendition"),
			),
			Entry("content id valid; content attributes invalid",
				func(datum *imageStoreStructured.Update) {
					datum.ContentID = pointer.FromString(imageTest.RandomContentID())
					datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
					datum.ContentAttributes = imageStoreStructuredTest.RandomContentAttributes()
					datum.ContentAttributes.DigestMD5 = pointer.FromString("")
					datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
					datum.Rendition = pointer.FromString(imageTest.RandomRenditionString())
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/contentAttributes/digestMD5"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/renditionsId"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/rendition"),
			),
			Entry("content id valid; renditions id missing",
				func(datum *imageStoreStructured.Update) {
					datum.ContentID = pointer.FromString(imageTest.RandomContentID())
					datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
					datum.ContentAttributes = imageStoreStructuredTest.RandomContentAttributes()
					datum.RenditionsID = nil
					datum.Rendition = pointer.FromString(imageTest.RandomRenditionString())
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/rendition"),
			),
			Entry("content id valid; renditions id invalid",
				func(datum *imageStoreStructured.Update) {
					datum.ContentID = pointer.FromString(imageTest.RandomContentID())
					datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
					datum.ContentAttributes = imageStoreStructuredTest.RandomContentAttributes()
					datum.RenditionsID = pointer.FromString("")
					datum.Rendition = pointer.FromString(imageTest.RandomRenditionString())
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/renditionsId"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/rendition"),
			),
			Entry("content id valid; rendition missing",
				func(datum *imageStoreStructured.Update) {
					datum.ContentID = pointer.FromString(imageTest.RandomContentID())
					datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
					datum.ContentAttributes = imageStoreStructuredTest.RandomContentAttributes()
					datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
					datum.Rendition = nil
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/renditionsId"),
			),
			Entry("content id valid; rendition invalid",
				func(datum *imageStoreStructured.Update) {
					datum.ContentID = pointer.FromString(imageTest.RandomContentID())
					datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
					datum.ContentAttributes = imageStoreStructuredTest.RandomContentAttributes()
					datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
					datum.Rendition = pointer.FromString("")
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/renditionsId"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/rendition"),
			),
			Entry("content id valid; content id, content intent, content attributes, renditions id, and rendition valid",
				func(datum *imageStoreStructured.Update) {
					datum.ContentID = pointer.FromString(imageTest.RandomContentID())
					datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
					datum.ContentAttributes = imageStoreStructuredTest.RandomContentAttributes()
					datum.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
					datum.Rendition = pointer.FromString(imageTest.RandomRenditionString())
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/renditionsId"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/rendition"),
			),
			Entry("multiple errors",
				func(datum *imageStoreStructured.Update) {
					datum.Metadata.Name = pointer.FromString("")
					datum.ContentID = pointer.FromString("")
					datum.ContentIntent = pointer.FromString("")
					datum.ContentAttributes = nil
					datum.RenditionsID = pointer.FromString("")
					datum.Rendition = pointer.FromString("")
				},
				errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/metadata/name"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/contentId"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", image.ContentIntents()), "/contentIntent"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/contentAttributes"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/renditionsId"),
				errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/rendition"),
			),
		)

		Context("IsEmpty", func() {
			var datum *imageStoreStructured.Update

			BeforeEach(func() {
				datum = imageStoreStructured.NewUpdate()
			})

			It("returns true when no fields are specified", func() {
				Expect(datum.IsEmpty()).To(BeTrue())
			})

			It("returns true when the metadata field is specified and empty", func() {
				datum.Metadata = image.NewMetadata()
				Expect(datum.IsEmpty()).To(BeTrue())
			})

			It("returns false when the metadata field is specified and not empty", func() {
				datum.Metadata = imageTest.RandomMetadata()
				Expect(datum.IsEmpty()).To(BeFalse())
			})

			It("returns false when the content intent field is specified", func() {
				datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
				Expect(datum.IsEmpty()).To(BeFalse())
			})

			It("returns false when the content attributes field is specified", func() {
				datum.ContentAttributes = imageStoreStructuredTest.RandomContentAttributes()
				Expect(datum.IsEmpty()).To(BeFalse())
			})

			It("returns false when the rendition field is specified", func() {
				datum.Rendition = pointer.FromString(imageTest.RandomRenditionString())
				Expect(datum.IsEmpty()).To(BeFalse())
			})

			It("returns false when multiple fields are specified", func() {
				datum.Metadata = imageTest.RandomMetadata()
				datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
				datum.ContentAttributes = imageStoreStructuredTest.RandomContentAttributes()
				datum.Rendition = pointer.FromString(imageTest.RandomRenditionString())
				Expect(datum.IsEmpty()).To(BeFalse())
			})
		})
	})

	Context("ContentAttributes", func() {
		Context("NewContentAttributes", func() {
			It("returns successfully with default values", func() {
				Expect(imageStoreStructured.NewContentAttributes()).To(Equal(&imageStoreStructured.ContentAttributes{}))
			})
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *imageStoreStructured.ContentAttributes), expectedErrors ...error) {
					datum := imageStoreStructuredTest.RandomContentAttributes()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New().Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *imageStoreStructured.ContentAttributes) {},
				),
				Entry("digest MD5 missing",
					func(datum *imageStoreStructured.ContentAttributes) { datum.DigestMD5 = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/digestMD5"),
				),
				Entry("digest MD5 empty",
					func(datum *imageStoreStructured.ContentAttributes) { datum.DigestMD5 = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/digestMD5"),
				),
				Entry("digest MD5 invalid",
					func(datum *imageStoreStructured.ContentAttributes) { datum.DigestMD5 = pointer.FromString("#") },
					errorsTest.WithPointerSource(crypto.ErrorValueStringAsBase64EncodedMD5HashNotValid("#"), "/digestMD5"),
				),
				Entry("digest MD5 valid",
					func(datum *imageStoreStructured.ContentAttributes) {
						datum.DigestMD5 = pointer.FromString(cryptoTest.RandomBase64EncodedMD5Hash())
					},
				),
				Entry("media type missing",
					func(datum *imageStoreStructured.ContentAttributes) { datum.MediaType = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/mediaType"),
				),
				Entry("media type empty",
					func(datum *imageStoreStructured.ContentAttributes) { datum.MediaType = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/mediaType"),
				),
				Entry("media type invalid",
					func(datum *imageStoreStructured.ContentAttributes) { datum.MediaType = pointer.FromString("/") },
					errorsTest.WithPointerSource(net.ErrorValueStringAsMediaTypeNotValid("/"), "/mediaType"),
				),
				Entry("media type unsupported",
					func(datum *imageStoreStructured.ContentAttributes) {
						datum.MediaType = pointer.FromString("application/octet-stream")
					},
					errorsTest.WithPointerSource(request.ErrorMediaTypeNotSupported("application/octet-stream"), "/mediaType"),
				),
				Entry("media type valid",
					func(datum *imageStoreStructured.ContentAttributes) {
						datum.MediaType = pointer.FromString(imageTest.RandomMediaType())
					},
				),
				Entry("width missing",
					func(datum *imageStoreStructured.ContentAttributes) { datum.Width = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/width"),
				),
				Entry("width out of range (lower)",
					func(datum *imageStoreStructured.ContentAttributes) { datum.Width = pointer.FromInt(0) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotGreaterThan(0, 0), "/width"),
				),
				Entry("width in range (lower)",
					func(datum *imageStoreStructured.ContentAttributes) { datum.Width = pointer.FromInt(1) },
				),
				Entry("height missing",
					func(datum *imageStoreStructured.ContentAttributes) { datum.Height = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/height"),
				),
				Entry("height out of range (lower)",
					func(datum *imageStoreStructured.ContentAttributes) { datum.Height = pointer.FromInt(0) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotGreaterThan(0, 0), "/height"),
				),
				Entry("height in range (lower)",
					func(datum *imageStoreStructured.ContentAttributes) { datum.Height = pointer.FromInt(1) },
				),
				Entry("size missing",
					func(datum *imageStoreStructured.ContentAttributes) { datum.Size = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/size"),
				),
				Entry("size out of range (lower)",
					func(datum *imageStoreStructured.ContentAttributes) { datum.Size = pointer.FromInt(0) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotGreaterThan(0, 0), "/size"),
				),
				Entry("size in range (lower)",
					func(datum *imageStoreStructured.ContentAttributes) { datum.Size = pointer.FromInt(1) },
				),
				Entry("multiple errors",
					func(datum *imageStoreStructured.ContentAttributes) {
						datum.DigestMD5 = pointer.FromString("")
						datum.MediaType = nil
						datum.Width = pointer.FromInt(0)
						datum.Height = pointer.FromInt(0)
						datum.Size = pointer.FromInt(0)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/digestMD5"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/mediaType"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotGreaterThan(0, 0), "/width"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotGreaterThan(0, 0), "/height"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotGreaterThan(0, 0), "/size"),
				),
			)
		})
	})
})
