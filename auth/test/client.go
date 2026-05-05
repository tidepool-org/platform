package test

type Client struct {
	*ProviderSessionClient
	*RestrictedTokenAccessor
	*ExternalAccessor
}

func NewClient() *Client {
	return &Client{
		ProviderSessionClient:   NewProviderSessionClient(),
		RestrictedTokenAccessor: NewRestrictedTokenAccessor(),
		ExternalAccessor:        NewExternalAccessor(),
	}
}

func (c *Client) AssertOutputsEmpty() {
	c.ProviderSessionClient.Expectations()
	c.RestrictedTokenAccessor.Expectations()
	c.ExternalAccessor.AssertOutputsEmpty()
}
