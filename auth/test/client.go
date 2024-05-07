package test

type Client struct {
	*ProviderSessionAccessor
	*RestrictedTokenAccessor
	*ExternalAccessor
	*DeviceTokensClient
}

func NewClient() *Client {
	return &Client{
		ProviderSessionAccessor: NewProviderSessionAccessor(),
		RestrictedTokenAccessor: NewRestrictedTokenAccessor(),
		ExternalAccessor:        NewExternalAccessor(),
		DeviceTokensClient:      NewDeviceTokensClient(),
	}
}

func (c *Client) AssertOutputsEmpty() {
	c.ProviderSessionAccessor.Expectations()
	c.RestrictedTokenAccessor.Expectations()
	c.ExternalAccessor.AssertOutputsEmpty()
}
