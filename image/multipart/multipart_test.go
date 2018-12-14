package multipart_test

import (
	"bytes"
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/image"
	imageMultipart "github.com/tidepool-org/platform/image/multipart"
	imageTest "github.com/tidepool-org/platform/image/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

type Components struct {
	Metadata      *image.Metadata
	ContentIntent string
	Content       *image.Content
}

var _ = Describe("Multipart", func() {
	var formDecoder *imageMultipart.FormDecoderImpl
	var formEncoder *imageMultipart.FormEncoderImpl

	BeforeEach(func() {
		formEncoder = imageMultipart.NewFormEncoder()
		Expect(formEncoder).ToNot(BeNil())
		formDecoder = imageMultipart.NewFormDecoder()
		Expect(formDecoder).ToNot(BeNil())
	})

	DescribeTable("returns the expected results and errors",
		func(mutator func(components *Components), expectedErrors ...error) {
			body := imageTest.RandomContentBytes()
			content := imageTest.RandomContent()
			content.Body = ioutil.NopCloser(bytes.NewReader(body))
			components := &Components{
				Metadata:      imageTest.RandomMetadata(),
				ContentIntent: imageTest.RandomContentIntent(),
				Content:       content,
			}
			mutator(components)
			reader, contentType := formEncoder.EncodeForm(components.Metadata, components.ContentIntent, components.Content)
			Expect(reader).ToNot(BeNil())
			defer reader.Close()
			Expect(contentType).ToNot(BeEmpty())
			resultMetadata, resultContentIntent, resultContent, err := formDecoder.DecodeForm(reader, contentType)
			errorsTest.ExpectEqual(err, expectedErrors...)
			if err != nil {
				Expect(resultMetadata).To(BeNil())
				Expect(resultContentIntent).To(BeEmpty())
				Expect(resultContent).To(BeNil())
			} else {
				Expect(resultContent).ToNot(BeNil())
				Expect(resultContent.Body).ToNot(BeNil())
				defer resultContent.Body.Close()
				resultBody, resultBodyErr := ioutil.ReadAll(resultContent.Body)
				Expect(resultBodyErr).ToNot(HaveOccurred())
				Expect(resultBody).To(Equal(body))
				Expect(resultContent.DigestMD5).To(Equal(components.Content.DigestMD5))
				Expect(resultContent.MediaType).To(Equal(components.Content.MediaType))
				Expect(resultContentIntent).To(Equal(components.ContentIntent))
				if components.Metadata == nil {
					Expect(resultMetadata).To(Equal(image.NewMetadata()))
				} else {
					Expect(resultMetadata).To(Equal(components.Metadata))
				}
			}
		},
		Entry("succeeds",
			func(components *Components) {},
		),
		Entry("metadata missing",
			func(components *Components) {
				components.Metadata = nil
			},
		),
		Entry("metadata name invalid",
			func(components *Components) {
				components.Metadata.Name = pointer.FromString("")
			},
			errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/metadata/0/name"),
		),
		Entry("content intent missing",
			func(components *Components) {
				components.ContentIntent = ""
			},
			errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/contentIntent"),
		),
		Entry("content intent invalid",
			func(components *Components) {
				components.ContentIntent = "invalid"
			},
			errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", image.ContentIntents()), "/contentIntent/0"),
		),
		Entry("content intent alternate",
			func(components *Components) {
				components.ContentIntent = "alternate"
			},
		),
		Entry("content intent original",
			func(components *Components) {
				components.ContentIntent = "original"
			},
		),
		Entry("content missing",
			func(components *Components) {
				components.Content = nil
			},
			errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/content"),
		),
		Entry("content body missing",
			func(components *Components) {
				components.Content.Body = nil
			},
			errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/content"),
		),
		Entry("content digest missing",
			func(components *Components) {
				components.Content.DigestMD5 = nil
			},
		),
		Entry("content digest invalid",
			func(components *Components) {
				components.Content.DigestMD5 = pointer.FromString("invalid")
			},
			errorsTest.WithPointerSource(request.ErrorHeaderInvalid("Digest"), "/content/0"),
		),
		Entry("content media type missing",
			func(components *Components) {
				components.Content.MediaType = nil
			},
			errorsTest.WithPointerSource(request.ErrorHeaderMissing("Content-Type"), "/content/0"),
		),
		Entry("content media type invalid",
			func(components *Components) {
				components.Content.MediaType = pointer.FromString("/")
			},
			errorsTest.WithPointerSource(request.ErrorHeaderInvalid("Content-Type"), "/content/0"),
		),
		Entry("content media type not supported",
			func(components *Components) {
				components.Content.MediaType = pointer.FromString("application/octet-stream")
			},
			errorsTest.WithPointerSource(request.ErrorMediaTypeNotSupported("application/octet-stream"), "/content/0"),
		),
		Entry("content media type image/jpeg",
			func(components *Components) {
				components.Content.MediaType = pointer.FromString("image/jpeg")
			},
		),
		Entry("content media type image/png",
			func(components *Components) {
				components.Content.MediaType = pointer.FromString("image/png")
			},
		),
	)
})
