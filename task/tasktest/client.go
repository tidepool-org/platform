package tasktest

type Client struct {
	*TaskAccessor
}

func NewClient() *Client {
	return &Client{
		TaskAccessor: NewTaskAccessor(),
	}
}

func (c *Client) Expectations() {
	c.TaskAccessor.Expectations()
}
