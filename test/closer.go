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

func ErrorCloser(reader io.Reader, err error) io.ReadCloser {
	readerWithErrorCloser := ReaderWithErrorCloser{Reader: reader, err: err}
	if _, ok := reader.(io.WriterTo); ok {
		return WriterToReaderWithErrorCloser{ReaderWithErrorCloser: readerWithErrorCloser}
	}
	return readerWithErrorCloser
}

type ReaderWithErrorCloser struct {
	io.Reader
	err error
}

func (r ReaderWithErrorCloser) Close() error {
	return r.err
}

type WriterToReaderWithErrorCloser struct {
	ReaderWithErrorCloser
}

func (w WriterToReaderWithErrorCloser) WriteTo(writer io.Writer) (n int64, err error) {
	return w.Reader.(io.WriterTo).WriteTo(writer)
}
