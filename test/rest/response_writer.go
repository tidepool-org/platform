package rest

import (
	"net/http"

	"github.com/tidepool-org/platform/id"
)

type EncodeJsonOutput struct {
	Bytes []byte
	Error error
}

type ResponseWriter struct {
	ID                     string
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
		ID:         id.New(),
		HeaderImpl: http.Header{},
	}
}

func (r *ResponseWriter) Header() http.Header {
	return r.HeaderImpl
}

func (r *ResponseWriter) WriteJson(object interface{}) error {
	r.WriteJsonInvocations++

	r.WriteJsonInputs = append(r.WriteJsonInputs, object)

	if len(r.WriteJsonOutputs) == 0 {
		panic("Unexpected invocation of WriteJson on ResponseWriter")
	}

	output := r.WriteJsonOutputs[0]
	r.WriteJsonOutputs = r.WriteJsonOutputs[1:]
	return output
}

func (r *ResponseWriter) EncodeJson(object interface{}) ([]byte, error) {
	r.EncodeJsonInvocations++

	r.EncodeJsonInputs = append(r.EncodeJsonInputs, object)

	if len(r.EncodeJsonOutputs) == 0 {
		panic("Unexpected invocation of EncodeJson on ResponseWriter")
	}

	output := r.EncodeJsonOutputs[0]
	r.EncodeJsonOutputs = r.EncodeJsonOutputs[1:]
	return output.Bytes, output.Error
}

func (r *ResponseWriter) WriteHeader(code int) {
	r.WriteHeaderInvocations++

	r.WriteHeaderInputs = append(r.WriteHeaderInputs, code)
}

func (r *ResponseWriter) UnusedOutputsCount() int {
	return len(r.WriteJsonOutputs) +
		len(r.EncodeJsonOutputs)
}
