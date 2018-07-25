package test

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) AssertOutputsEmpty() {}
