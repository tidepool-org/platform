package test

import (
	dataSourceStoreStructured "github.com/tidepool-org/platform/data/source/store/structured"
	"github.com/tidepool-org/platform/user"
)

type ClientProvider struct {
	DataSourceStructuredStoreInvocations int
	DataSourceStructuredStoreStub        func() dataSourceStoreStructured.Store
	DataSourceStructuredStoreOutputs     []dataSourceStoreStructured.Store
	DataSourceStructuredStoreOutput      *dataSourceStoreStructured.Store
	UserClientInvocations                int
	UserClientStub                       func() user.Client
	UserClientOutputs                    []user.Client
	UserClientOutput                     *user.Client
}

func NewClientProvider() *ClientProvider {
	return &ClientProvider{}
}

func (c *ClientProvider) DataSourceStructuredStore() dataSourceStoreStructured.Store {
	c.DataSourceStructuredStoreInvocations++
	if c.DataSourceStructuredStoreStub != nil {
		return c.DataSourceStructuredStoreStub()
	}
	if len(c.DataSourceStructuredStoreOutputs) > 0 {
		output := c.DataSourceStructuredStoreOutputs[0]
		c.DataSourceStructuredStoreOutputs = c.DataSourceStructuredStoreOutputs[1:]
		return output
	}
	if c.DataSourceStructuredStoreOutput != nil {
		return *c.DataSourceStructuredStoreOutput
	}
	panic("DataSourceStructuredStore has no output")
}

func (c *ClientProvider) UserClient() user.Client {
	c.UserClientInvocations++
	if c.UserClientStub != nil {
		return c.UserClientStub()
	}
	if len(c.UserClientOutputs) > 0 {
		output := c.UserClientOutputs[0]
		c.UserClientOutputs = c.UserClientOutputs[1:]
		return output
	}
	if c.UserClientOutput != nil {
		return *c.UserClientOutput
	}
	panic("UserClient has no output")
}

func (c *ClientProvider) AssertOutputsEmpty() {
	if len(c.DataSourceStructuredStoreOutputs) > 0 {
		panic("DataSourceStructuredStoreOutputs is not empty")
	}
	if len(c.UserClientOutputs) > 0 {
		panic("UserClientOutputs is not empty")
	}
}
