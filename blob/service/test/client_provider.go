package test

import (
	blobStoreStructured "github.com/tidepool-org/platform/blob/store/structured"
	blobStoreUnstructured "github.com/tidepool-org/platform/blob/store/unstructured"
	"github.com/tidepool-org/platform/user"
)

type ClientProvider struct {
	BlobStructuredStoreInvocations   int
	BlobStructuredStoreStub          func() blobStoreStructured.Store
	BlobStructuredStoreOutputs       []blobStoreStructured.Store
	BlobStructuredStoreOutput        *blobStoreStructured.Store
	BlobUnstructuredStoreInvocations int
	BlobUnstructuredStoreStub        func() blobStoreUnstructured.Store
	BlobUnstructuredStoreOutputs     []blobStoreUnstructured.Store
	BlobUnstructuredStoreOutput      *blobStoreUnstructured.Store
	UserClientInvocations            int
	UserClientStub                   func() user.Client
	UserClientOutputs                []user.Client
	UserClientOutput                 *user.Client
}

func NewClientProvider() *ClientProvider {
	return &ClientProvider{}
}

func (c *ClientProvider) BlobStructuredStore() blobStoreStructured.Store {
	c.BlobStructuredStoreInvocations++
	if c.BlobStructuredStoreStub != nil {
		return c.BlobStructuredStoreStub()
	}
	if len(c.BlobStructuredStoreOutputs) > 0 {
		output := c.BlobStructuredStoreOutputs[0]
		c.BlobStructuredStoreOutputs = c.BlobStructuredStoreOutputs[1:]
		return output
	}
	if c.BlobStructuredStoreOutput != nil {
		return *c.BlobStructuredStoreOutput
	}
	panic("BlobStructuredStore has no output")
}

func (c *ClientProvider) BlobUnstructuredStore() blobStoreUnstructured.Store {
	c.BlobUnstructuredStoreInvocations++
	if c.BlobUnstructuredStoreStub != nil {
		return c.BlobUnstructuredStoreStub()
	}
	if len(c.BlobUnstructuredStoreOutputs) > 0 {
		output := c.BlobUnstructuredStoreOutputs[0]
		c.BlobUnstructuredStoreOutputs = c.BlobUnstructuredStoreOutputs[1:]
		return output
	}
	if c.BlobUnstructuredStoreOutput != nil {
		return *c.BlobUnstructuredStoreOutput
	}
	panic("BlobUnstructuredStore has no output")
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
	if len(c.BlobStructuredStoreOutputs) > 0 {
		panic("BlobStructuredStoreOutputs is not empty")
	}
	if len(c.BlobUnstructuredStoreOutputs) > 0 {
		panic("BlobUnstructuredStoreOutputs is not empty")
	}
	if len(c.UserClientOutputs) > 0 {
		panic("UserClientOutputs is not empty")
	}
}
