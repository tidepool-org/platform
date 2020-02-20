package test

type Client struct {
	*ProviderSessionAccessor
	*RestrictedTokenAccessor
	*DeviceAuthorizationAccessor
	*ExternalAccessor
}

func NewClient() *Client {
	return &Client{
		ProviderSessionAccessor:     NewProviderSessionAccessor(),
		RestrictedTokenAccessor:     NewRestrictedTokenAccessor(),
		DeviceAuthorizationAccessor: NewDeviceAuthorizationAccessor(),
		ExternalAccessor:            NewExternalAccessor(),
	}
}

func (c *Client) AssertOutputsEmpty() {
	c.ProviderSessionAccessor.Expectations()
	c.RestrictedTokenAccessor.Expectations()
	c.DeviceAuthorizationAccessor.Expectations()
	c.ExternalAccessor.AssertOutputsEmpty()
}
