package test

import (
	"bytes"
	"io"
)

func RandomReader() io.Reader {
	return bytes.NewReader(RandomBytes())
}

func RandomReadCloser() io.ReadCloser {
	return io.NopCloser(RandomReader())
}

func ZeroReader() io.Reader {
	return &ReaderZero{}
}

type ReaderZero struct{}

func (z *ReaderZero) Read(bites []byte) (int, error) {
	clear(bites)
	return len(bites), nil
}

func RandReader() io.Reader {
	return &ReaderRand{}
}

type ReaderRand struct{}

func (z *ReaderRand) Read(bites []byte) (int, error) {
	length := len(bites)
	copy(bites, RandomBytesFromRange(length, length))
	return length, nil
}

type ReadOutput struct {
	BytesRead int
	Error     error
}

type Reader struct {
	ReadInvocations int
	ReadInputs      [][]byte
	ReadStub        func(bites []byte) (int, error)
	ReadOutputs     []ReadOutput
	ReadOutput      *ReadOutput
}

func NewReader() *Reader {
	return &Reader{}
}

func (r *Reader) Read(bites []byte) (int, error) {
	r.ReadInvocations++
	r.ReadInputs = append(r.ReadInputs, bites)
	if r.ReadStub != nil {
		return r.ReadStub(bites)
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
