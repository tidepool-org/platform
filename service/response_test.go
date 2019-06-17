package service_test

import (
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/service"
)

type TestResponseWriter struct {
	HeaderImpl http.Header
}

func NewTestResponseWriter() *TestResponseWriter {
	return &TestResponseWriter{
		HeaderImpl: http.Header{},
	}
}

func (t *TestResponseWriter) Header() http.Header {
	return t.HeaderImpl
}

func (t *TestResponseWriter) WriteJson(v interface{}) error {
	panic("Unexpected invocation of WriteJson on TestResponseWriter")
}

func (t *TestResponseWriter) EncodeJson(v interface{}) ([]byte, error) {
	panic("Unexpected invocation of EncodeJson on TestResponseWriter")
}

func (t *TestResponseWriter) WriteHeader(code int) {
	panic("Unexpected invocation of WriteHeader on TestResponseWriter")
}

var _ = Describe("Response", func() {
	Context("with response", func() {
		var responseWriter *TestResponseWriter

		BeforeEach(func() {
			responseWriter = NewTestResponseWriter()
			Expect(responseWriter).ToNot(BeNil())
		})

		Context("AddDateHeader", func() {
			It("adds a date header", func() {
				service.AddDateHeader(responseWriter)
				Expect(responseWriter.HeaderImpl).To(HaveKey("Date"))
				date, err := time.Parse(time.RFC1123, responseWriter.HeaderImpl.Get("Date"))
				Expect(err).ToNot(HaveOccurred())
				Expect(date).To(BeTemporally("~", time.Now(), time.Second))
			})
		})
	})
})
