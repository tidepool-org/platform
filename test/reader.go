package test

type ReadOutput struct {
	BytesRead int
	Error     error
}

type Reader struct {
	ReadInvocations int
	ReadInputs      [][]byte
	ReadStub        func(bytes []byte) (int, error)
	ReadOutputs     []ReadOutput
	ReadOutput      *ReadOutput
}

func NewReader() *Reader {
	return &Reader{}
}

func (r *Reader) Read(bytes []byte) (int, error) {
	r.ReadInvocations++
	r.ReadInputs = append(r.ReadInputs, bytes)
	if r.ReadStub != nil {
		return r.ReadStub(bytes)
	}
	if len(r.ReadOutputs) > 0 {
		output := r.ReadOutputs[0]
		r.ReadOutputs = r.ReadOutputs[1:]
		return output.BytesRead, output.Error
	}
	if r.ReadOutput != nil {
		return r.ReadOutput.BytesRead, r.ReadOutput.Error
	}
	panic("Read has no output")
}

func (r *Reader) AssertOutputsEmpty() {
	if len(r.ReadOutputs) > 0 {
		panic("ReadOutputs is not empty")
	}
}
