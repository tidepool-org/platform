package rest

import (
	"net/http"

	"github.com/onsi/gomega"

	"github.com/tidepool-org/platform/test"
)

type EncodeJsonOutput struct {
	Bytes []byte
	Error error
}

type ResponseWriter struct {
	*test.Mock
	HeaderImpl             http.Header
	WriteJsonInvocations   int
	WriteJsonInputs        []interface{}
	WriteJsonOutputs       []error
	EncodeJsonInvocations  int
	EncodeJsonInputs       []interface{}
	EncodeJsonOutputs      []EncodeJsonOutput
	WriteHeaderInvocations int
	WriteHeaderInputs      []int
}

func NewResponseWriter() *ResponseWriter {
	return &ResponseWriter{
		Mock:       test.NewMock(),
		HeaderImpl: http.Header{},
	}
}

func (r *ResponseWriter) Header() http.Header {
	return r.HeaderImpl
}

func (r *ResponseWriter) WriteJson(object interface{}) error {
	r.WriteJsonInvocations++

	r.WriteJsonInputs = append(r.WriteJsonInputs, object)

	gomega.Expect(r.WriteJsonOutputs).ToNot(gomega.BeEmpty())

	output := r.WriteJsonOutputs[0]
	r.WriteJsonOutputs = r.WriteJsonOutputs[1:]
	return output
}

func (r *ResponseWriter) EncodeJson(object interface{}) ([]byte, error) {
	r.EncodeJsonInvocations++

	r.EncodeJsonInputs = append(r.EncodeJsonInputs, object)

	gomega.Expect(r.EncodeJsonOutputs).ToNot(gomega.BeEmpty())

	output := r.EncodeJsonOutputs[0]
	r.EncodeJsonOutputs = r.EncodeJsonOutputs[1:]
	return output.Bytes, output.Error
}

func (r *ResponseWriter) WriteHeader(code int) {
	r.WriteHeaderInvocations++

	r.WriteHeaderInputs = append(r.WriteHeaderInputs, code)
}

func (r *ResponseWriter) Expectations() {
	r.Mock.Expectations()
	gomega.Expect(r.WriteJsonOutputs).To(gomega.BeEmpty())
	gomega.Expect(r.EncodeJsonOutputs).To(gomega.BeEmpty())
}
