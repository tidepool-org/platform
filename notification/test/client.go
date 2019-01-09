package test

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) UnusedOutputsCount() int {
	return 0
}
