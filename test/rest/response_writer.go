package rest

import (
	"bufio"
	"net"
	"net/http"
)

type EncodeJsonOutput struct {
	Bytes []byte
	Error error
}

type WriteOutput struct {
	BytesWritten int
	Error        error
}

type HijackOutput struct {
	Connection net.Conn
	ReadWriter *bufio.ReadWriter
	Error      error
}

type ResponseWriter struct {
	HeaderInvocations      int
	HeaderStub             func() http.Header
	HeaderOutputs          []http.Header
	HeaderOutput           *http.Header
	WriteJsonInvocations   int
	WriteJsonInputs        []interface{}
	WriteJsonStub          func(object interface{}) error
	WriteJsonOutputs       []error
	WriteJsonOutput        *error
	EncodeJsonInvocations  int
	EncodeJsonInputs       []interface{}
	EncodeJsonStub         func(object interface{}) ([]byte, error)
	EncodeJsonOutputs      []EncodeJsonOutput
	EncodeJsonOutput       *EncodeJsonOutput
	WriteHeaderInvocations int
	WriteHeaderInputs      []int
	WriteHeaderStub        func(statusCode int)
	WriteInvocations       int
	WriteInputs            [][]byte
	WriteStub              func(bytes []byte) (int, error)
	WriteOutputs           []WriteOutput
	WriteOutput            *WriteOutput
	FlushInvocations       int
	FlushStub              func()
	CloseNotifyInvocations int
	CloseNotifyStub        func() <-chan bool
	CloseNotifyOutputs     []<-chan bool
	CloseNotifyOutput      *<-chan bool
	HijackInvocations      int
	HijackStub             func() (net.Conn, *bufio.ReadWriter, error)
	HijackOutputs          []HijackOutput
	HijackOutput           *HijackOutput
}

func NewResponseWriter() *ResponseWriter {
	return &ResponseWriter{}
}

func (r *ResponseWriter) Header() http.Header {
	r.HeaderInvocations++
	if r.HeaderStub != nil {
		return r.HeaderStub()
	}
	if len(r.HeaderOutputs) > 0 {
		output := r.HeaderOutputs[0]
		r.HeaderOutputs = r.HeaderOutputs[1:]
		return output
	}
	if r.HeaderOutput != nil {
		return *r.HeaderOutput
	}
	panic("Header has no output")
}

func (r *ResponseWriter) WriteJson(object interface{}) error {
	r.WriteJsonInvocations++
	r.WriteJsonInputs = append(r.WriteJsonInputs, object)
	if r.WriteJsonStub != nil {
		return r.WriteJsonStub(object)
	}
	if len(r.WriteJsonOutputs) > 0 {
		output := r.WriteJsonOutputs[0]
		r.WriteJsonOutputs = r.WriteJsonOutputs[1:]
		return output
	}
	if r.WriteJsonOutput != nil {
		return *r.WriteJsonOutput
	}
	panic("WriteJson has no output")
}

func (r *ResponseWriter) EncodeJson(object interface{}) ([]byte, error) {
	r.EncodeJsonInvocations++
	r.EncodeJsonInputs = append(r.EncodeJsonInputs, object)
	if r.EncodeJsonStub != nil {
		return r.EncodeJsonStub(object)
	}
	if len(r.EncodeJsonOutputs) > 0 {
		output := r.EncodeJsonOutputs[0]
		r.EncodeJsonOutputs = r.EncodeJsonOutputs[1:]
		return output.Bytes, output.Error
	}
	if r.EncodeJsonOutput != nil {
		return r.EncodeJsonOutput.Bytes, r.EncodeJsonOutput.Error
	}
	panic("EncodeJson has no output")
}

func (r *ResponseWriter) WriteHeader(statusCode int) {
	r.WriteHeaderInvocations++
	r.WriteHeaderInputs = append(r.WriteHeaderInputs, statusCode)
	if r.WriteHeaderStub != nil {
		r.WriteHeaderStub(statusCode)
	}
}

func (r *ResponseWriter) Write(bytes []byte) (int, error) {
	r.WriteInvocations++
	r.WriteInputs = append(r.WriteInputs, bytes)
	if r.WriteStub != nil {
		return r.WriteStub(bytes)
	}
	if len(r.WriteOutputs) > 0 {
		output := r.WriteOutputs[0]
		r.WriteOutputs = r.WriteOutputs[1:]
		return output.BytesWritten, output.Error
	}
	if r.WriteOutput != nil {
		return r.WriteOutput.BytesWritten, r.WriteOutput.Error
	}
	panic("Write has no output")
}

func (r *ResponseWriter) Flush() {
	r.FlushInvocations++
	if r.FlushStub != nil {
		r.FlushStub()
	}
}

func (r *ResponseWriter) CloseNotify() <-chan bool {
	r.CloseNotifyInvocations++
	if r.CloseNotifyStub != nil {
		return r.CloseNotifyStub()
	}
	if len(r.CloseNotifyOutputs) > 0 {
		output := r.CloseNotifyOutputs[0]
		r.CloseNotifyOutputs = r.CloseNotifyOutputs[1:]
		return output
	}
	if r.CloseNotifyOutput != nil {
		return *r.CloseNotifyOutput
	}
	panic("CloseNotify has no output")
}

func (r *ResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	r.HijackInvocations++
	if r.HijackStub != nil {
		return r.HijackStub()
	}
	if len(r.HijackOutputs) > 0 {
		output := r.HijackOutputs[0]
		r.HijackOutputs = r.HijackOutputs[1:]
		return output.Connection, output.ReadWriter, output.Error
	}
	if r.HijackOutput != nil {
		return r.HijackOutput.Connection, r.HijackOutput.ReadWriter, r.HijackOutput.Error
	}
	panic("Hijack has no output")
}

func (r *ResponseWriter) AssertOutputsEmpty() {
	if len(r.HeaderOutputs) > 0 {
		panic("HeaderOutputs is not empty")
	}
	if len(r.WriteJsonOutputs) > 0 {
		panic("WriteJsonOutputs is not empty")
	}
	if len(r.EncodeJsonOutputs) > 0 {
		panic("EncodeJsonOutputs is not empty")
	}
	if len(r.WriteOutputs) > 0 {
		panic("WriteOutputs is not empty")
	}
	if len(r.CloseNotifyOutputs) > 0 {
		panic("CloseNotifyOutputs is not empty")
	}
	if len(r.HijackOutputs) > 0 {
		panic("HijackOutputs  is not empty")
	}
}
