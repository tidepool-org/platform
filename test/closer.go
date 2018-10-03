package test

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
