package request_test

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	cryptoTest "github.com/tidepool-org/platform/crypto/test"
	netTest "github.com/tidepool-org/platform/net/test"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/test"
	testHttp "github.com/tidepool-org/platform/test/http"
)

var _ = Describe("Parser", func() {
	Context("with header", func() {
		var header http.Header
		var key string

		BeforeEach(func() {
			header = testHttp.RandomHeader()
			key = testHttp.NewHeaderKey()
		})

		Context("ParseSingletonHeader", func() {
			var value string

			BeforeEach(func() {
				value = testHttp.NewHeaderValue()
			})

			It("returns nil if the header is nil", func() {
				header = nil
				result, err := request.ParseSingletonHeader(header, key)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("returns nil if the key is not present", func() {
				result, err := request.ParseSingletonHeader(header, key)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("returns nil if there are no values", func() {
				header[key] = []string{}
				result, err := request.ParseSingletonHeader(header, key)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("returns an error if there is more than one value", func() {
				header[key] = []string{testHttp.NewHeaderValue(), value, testHttp.NewHeaderValue()}
				result, err := request.ParseSingletonHeader(header, key)
				Expect(err).To(MatchError(fmt.Sprintf("header %q is invalid", key)))
				Expect(result).To(BeNil())
			})

			It("returns the value if there is exactly one", func() {
				header[key] = []string{value}
				result, err := request.ParseSingletonHeader(header, key)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).ToNot(BeNil())
				Expect(*result).To(Equal(value))
			})
		})

		Context("ParseDigestMD5Header", func() {
			var value string

			BeforeEach(func() {
				value = cryptoTest.RandomBase64EncodedMD5Hash()
			})

			It("returns nil if the header is nil", func() {
				header = nil
				result, err := request.ParseDigestMD5Header(header, key)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("returns nil if the key is not present", func() {
				result, err := request.ParseDigestMD5Header(header, key)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("returns nil if there are no values", func() {
				header[key] = []string{}
				result, err := request.ParseDigestMD5Header(header, key)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("returns an error if there is more than one value", func() {
				header[key] = []string{testHttp.NewHeaderValue(), fmt.Sprintf("MD5=%s", value), testHttp.NewHeaderValue()}
				result, err := request.ParseDigestMD5Header(header, key)
				Expect(err).To(MatchError(fmt.Sprintf("header %q is invalid", key)))
				Expect(result).To(BeNil())
			})

			It("returns an error if the value is not valid", func() {
				header[key] = []string{cryptoTest.RandomBase64EncodedMD5Hash()}
				result, err := request.ParseDigestMD5Header(header, key)
				Expect(err).To(MatchError(fmt.Sprintf("header %q is invalid", key)))
				Expect(result).To(BeNil())
			})

			It("returns an error if the value algorithm is not MD5", func() {
				header[key] = []string{fmt.Sprintf("SHA1=%s", value)}
				result, err := request.ParseDigestMD5Header(header, key)
				Expect(err).To(MatchError(fmt.Sprintf("header %q is invalid", key)))
				Expect(result).To(BeNil())
			})

			It("returns an error if the value digest is not valid", func() {
				header[key] = []string{fmt.Sprintf("MD5=%s", test.RandomStringFromRangeAndCharset(20, 20, test.CharsetAlphaNumeric))}
				result, err := request.ParseDigestMD5Header(header, key)
				Expect(err).To(MatchError(fmt.Sprintf("header %q is invalid", key)))
				Expect(result).To(BeNil())
			})

			It("returns the value if there is exactly one and it is valid", func() {
				header[key] = []string{fmt.Sprintf("MD5=%s", value)}
				result, err := request.ParseDigestMD5Header(header, key)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).ToNot(BeNil())
				Expect(*result).To(Equal(value))
			})
		})

		Context("ParseMediaTypeHeader", func() {
			var value string

			BeforeEach(func() {
				value = netTest.RandomMediaType()
			})

			It("returns nil if the header is nil", func() {
				header = nil
				result, err := request.ParseMediaTypeHeader(header, key)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("returns nil if the key is not present", func() {
				result, err := request.ParseMediaTypeHeader(header, key)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("returns nil if there are no values", func() {
				header[key] = []string{}
				result, err := request.ParseMediaTypeHeader(header, key)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("returns an error if there is more than one value", func() {
				header[key] = []string{netTest.RandomMediaType(), value, netTest.RandomMediaType()}
				result, err := request.ParseMediaTypeHeader(header, key)
				Expect(err).To(MatchError(fmt.Sprintf("header %q is invalid", key)))
				Expect(result).To(BeNil())
			})

			It("returns an error if the value is not valid", func() {
				header[key] = []string{"/"}
				result, err := request.ParseMediaTypeHeader(header, key)
				Expect(err).To(MatchError(fmt.Sprintf("header %q is invalid", key)))
				Expect(result).To(BeNil())
			})

			It("returns the value if there is exactly one and it is valid", func() {
				header[key] = []string{value}
				result, err := request.ParseMediaTypeHeader(header, key)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).ToNot(BeNil())
				Expect(*result).To(Equal(value))
			})
		})

		Context("ParseIntHeader", func() {
			var value int

			BeforeEach(func() {
				value = test.RandomInt()
			})

			It("returns nil if the header is nil", func() {
				header = nil
				result, err := request.ParseIntHeader(header, key)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("returns nil if the key is not present", func() {
				result, err := request.ParseIntHeader(header, key)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("returns nil if there are no values", func() {
				header[key] = []string{}
				result, err := request.ParseIntHeader(header, key)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("returns an error if there is more than one value", func() {
				header[key] = []string{strconv.Itoa(test.RandomInt()), strconv.Itoa(value), strconv.Itoa(test.RandomInt())}
				result, err := request.ParseIntHeader(header, key)
				Expect(err).To(MatchError(fmt.Sprintf("header %q is invalid", key)))
				Expect(result).To(BeNil())
			})

			It("returns an error if the value is not valid", func() {
				header[key] = []string{"abc"}
				result, err := request.ParseIntHeader(header, key)
				Expect(err).To(MatchError(fmt.Sprintf("header %q is invalid", key)))
				Expect(result).To(BeNil())
			})

			It("returns the value if there is exactly one and it is valid", func() {
				header[key] = []string{strconv.Itoa(value)}
				result, err := request.ParseIntHeader(header, key)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).ToNot(BeNil())
				Expect(*result).To(Equal(value))
			})
		})
		Context("ParseTimeHeader", func() {
			var value string
			var layout string

			BeforeEach(func() {
				layout = time.RFC3339
				value = time.Now().Format(layout)
			})

			It("returns nil if the header is nil", func() {
				header = nil
				result, err := request.ParseTimeHeader(header, key, layout)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("returns nil if the key is not present", func() {
				result, err := request.ParseTimeHeader(header, key, layout)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("returns nil if there are no values", func() {
				header[key] = []string{}
				result, err := request.ParseTimeHeader(header, key, layout)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("returns an error if there is more than one value", func() {
				header[key] = []string{time.Now().Add(1 * time.Hour).Format(layout), value}
				result, err := request.ParseTimeHeader(header, key, layout)
				Expect(err).To(MatchError(fmt.Sprintf("header %q is invalid", key)))
				Expect(result).To(BeNil())
			})

			It("returns an error if the value is not valid", func() {
				header[key] = []string{"abc"}
				result, err := request.ParseTimeHeader(header, key, layout)
				Expect(err).To(MatchError(fmt.Sprintf("header %q is invalid", key)))
				Expect(result).To(BeNil())
			})

			It("returns the value if there is exactly one and it is valid", func() {
				header[key] = []string{value}
				result, err := request.ParseTimeHeader(header, key, layout)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).ToNot(BeNil())
				Expect(result.Format(layout)).To(Equal(value))
			})
		})
	})
})
