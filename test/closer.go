package test

import "io"

type Closer struct {
	CloseInvocations int
	CloseStub        func() error
	CloseOutputs     []error
	CloseOutput      *error
}

func NewCloser() *Closer {
	return &Closer{}
}

func (c *Closer) Close() error {
	c.CloseInvocations++
	if c.CloseStub != nil {
		return c.CloseStub()
	}
	if len(c.CloseOutputs) > 0 {
		output := c.CloseOutputs[0]
		c.CloseOutputs = c.CloseOutputs[1:]
		return output
	}
	if c.CloseOutput != nil {
		return *c.CloseOutput
	}
	panic("Close has no output")
}

func (c *Closer) AssertOutputsEmpty() {
	if len(c.CloseOutputs) > 0 {
		panic("CloseOutputs is not empty")
	}
}

func ErrorReader(err error) io.Reader {
	return ReaderWithError{err: err}
}

type ReaderWithError struct {
	err error
}

func (r ReaderWithError) Read(bites []byte) (int, error) {
	return 0, r.err
}

func ErrorReadCloser(reader io.Reader, err error) io.ReadCloser {
	readerWithErrorCloser := ReadCloserWithError{Reader: reader, err: err}
	if _, ok := reader.(io.WriterTo); ok {
		return WriterToReadCloserWithError{ReadCloserWithError: readerWithErrorCloser}
	}
	return readerWithErrorCloser
}

type ReadCloserWithError struct {
	io.Reader
	err error
}

func (r ReadCloserWithError) Close() error {
	return r.err
}

type WriterToReadCloserWithError struct {
	ReadCloserWithError
}

func (w WriterToReadCloserWithError) WriteTo(writer io.Writer) (int64, error) {
	if writerTo, ok := w.Reader.(io.WriterTo); !ok {
		panic("Reader does not implement io.WriterTo")
	} else {
		return writerTo.WriteTo(writer)
	}
}
