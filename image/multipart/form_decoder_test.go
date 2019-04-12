package multipart_test

import (
	"bytes"
	"fmt"
	"io"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	imageMultipart "github.com/tidepool-org/platform/image/multipart"
	netTest "github.com/tidepool-org/platform/net/test"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("FormDecoder", func() {
	Context("NewFormDecoder", func() {
		It("returns successfully", func() {
			Expect(imageMultipart.NewFormDecoder()).ToNot(BeNil())
		})
	})

	Context("with new form decoder", func() {
		var formDecoder *imageMultipart.FormDecoderImpl
		var reader io.Reader
		var contentType string

		BeforeEach(func() {
			formDecoder = imageMultipart.NewFormDecoder()
			Expect(formDecoder).ToNot(BeNil())
			reader = &bytes.Buffer{}
			contentType = fmt.Sprintf("multipart/form; boundary=%s", test.RandomStringFromRangeAndCharset(16, 128, test.CharsetAlphaNumeric))
		})

		It("returns an error when the reader is missing", func() {
			reader = nil
			metadata, contentIntent, content, err := formDecoder.DecodeForm(reader, contentType)
			errorsTest.ExpectEqual(err, errors.New("reader is missing"))
			Expect(metadata).To(BeNil())
			Expect(contentIntent).To(BeEmpty())
			Expect(content).To(BeNil())
		})

		It("returns an error when the content type is missing", func() {
			contentType = ""
			metadata, contentIntent, content, err := formDecoder.DecodeForm(reader, contentType)
			errorsTest.ExpectEqual(err, errors.New("content type is missing"))
			Expect(metadata).To(BeNil())
			Expect(contentIntent).To(BeEmpty())
			Expect(content).To(BeNil())
		})

		It("returns an error when the content type is invalid", func() {
			contentType = "/"
			metadata, contentIntent, content, err := formDecoder.DecodeForm(reader, contentType)
			errorsTest.ExpectEqual(err, errors.New("content type is invalid"))
			Expect(metadata).To(BeNil())
			Expect(contentIntent).To(BeEmpty())
			Expect(content).To(BeNil())
		})

		It("returns an error when the content type is not supported", func() {
			contentType = netTest.RandomMediaType()
			metadata, contentIntent, content, err := formDecoder.DecodeForm(reader, contentType)
			errorsTest.ExpectEqual(err, errors.New("content type is not supported"))
			Expect(metadata).To(BeNil())
			Expect(contentIntent).To(BeEmpty())
			Expect(content).To(BeNil())
		})

		It("returns an error when the boundary is missing", func() {
			contentType = "multipart/form"
			metadata, contentIntent, content, err := formDecoder.DecodeForm(reader, contentType)
			errorsTest.ExpectEqual(err, errors.New("boundary is missing"))
			Expect(metadata).To(BeNil())
			Expect(contentIntent).To(BeEmpty())
			Expect(content).To(BeNil())
		})
	})
})
