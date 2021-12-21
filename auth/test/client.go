package test

type Client struct {
	*ExternalAccessor
}

func NewClient() *Client {
	return &Client{
		ExternalAccessor: NewExternalAccessor(),
	}
}

func (c *Client) AssertOutputsEmpty() {
	c.ExternalAccessor.AssertOutputsEmpty()
}
