package test

import (
	"github.com/tidepool-org/platform/auth"
	dataSourceStoreStructured "github.com/tidepool-org/platform/data/source/store/structured"
)

type ClientProvider struct {
	AuthClientInvocations                int
	AuthClientStub                       func() auth.Client
	AuthClientOutputs                    []auth.Client
	AuthClientOutput                     *auth.Client
	DataSourceStructuredStoreInvocations int
	DataSourceStructuredStoreStub        func() dataSourceStoreStructured.Store
	DataSourceStructuredStoreOutputs     []dataSourceStoreStructured.Store
	DataSourceStructuredStoreOutput      *dataSourceStoreStructured.Store
}

func NewClientProvider() *ClientProvider {
	return &ClientProvider{}
}

func (c *ClientProvider) AuthClient() auth.Client {
	c.AuthClientInvocations++
	if c.AuthClientStub != nil {
		return c.AuthClientStub()
	}
	if len(c.AuthClientOutputs) > 0 {
		output := c.AuthClientOutputs[0]
		c.AuthClientOutputs = c.AuthClientOutputs[1:]
		return output
	}
	if c.AuthClientOutput != nil {
		return *c.AuthClientOutput
	}
	panic("AuthClient has no output")
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

func (c *ClientProvider) AssertOutputsEmpty() {
	if len(c.AuthClientOutputs) > 0 {
		panic("AuthClientOutputs is not empty")
	}
	if len(c.DataSourceStructuredStoreOutputs) > 0 {
		panic("DataSourceStructuredStoreOutputs is not empty")
	}
}
