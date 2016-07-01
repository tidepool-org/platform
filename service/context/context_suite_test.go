package context_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"net/http"
	"testing"

	"github.com/ant0ine/go-json-rest/rest"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "service/context")
}

func NewTestRequest() *rest.Request {
	baseRequest, err := http.NewRequest("GET", "http://127.0.0.1/", nil)
	Expect(err).ToNot(HaveOccurred())
	Expect(baseRequest).ToNot(BeNil())
	return &rest.Request{
		Request:    baseRequest,
		PathParams: map[string]string{},
		Env:        map[string]interface{}{},
	}
}

type EncodeJSONOutput struct {
	ByteArray []byte
	Error     error
}

type TestResponseWriter struct {
	header            http.Header
	WriteJSONInputs   []interface{}
	WriteJSONOutputs  []error
	EncodeJSONInputs  []interface{}
	EncodeJSONOutputs []EncodeJSONOutput
	WriteHeaderInputs []int
}

func (t *TestResponseWriter) Header() http.Header {
	return t.header
}

func (t *TestResponseWriter) WriteJson(v interface{}) error {
	t.WriteJSONInputs = append(t.WriteJSONInputs, v)
	output := t.WriteJSONOutputs[0]
	t.WriteJSONOutputs = t.WriteJSONOutputs[1:]
	return output
}

func (t *TestResponseWriter) EncodeJson(v interface{}) ([]byte, error) {
	t.EncodeJSONInputs = append(t.EncodeJSONInputs, v)
	output := t.EncodeJSONOutputs[0]
	t.EncodeJSONOutputs = t.EncodeJSONOutputs[1:]
	return output.ByteArray, output.Error
}

func (t *TestResponseWriter) WriteHeader(code int) {
	t.WriteHeaderInputs = append(t.WriteHeaderInputs, code)
}

func NewTestResponseWriter() *TestResponseWriter {
	return &TestResponseWriter{
		header: http.Header{},
	}
}
