package test

type Client struct {
	*ProviderSessionAccessor
	*RestrictedTokenAccessor
	*ExternalAccessor
}

func NewClient() *Client {
	return &Client{
		ProviderSessionAccessor: NewProviderSessionAccessor(),
		RestrictedTokenAccessor: NewRestrictedTokenAccessor(),
		ExternalAccessor:        NewExternalAccessor(),
	}
}

func (c *Client) Expectations() {
	c.ProviderSessionAccessor.Expectations()
	c.RestrictedTokenAccessor.Expectations()
	c.ExternalAccessor.Expectations()
}
