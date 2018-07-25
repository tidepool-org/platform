package test

import (
	"github.com/tidepool-org/platform/auth"
	blobStoreStructured "github.com/tidepool-org/platform/blob/store/structured"
	blobStoreUnstructured "github.com/tidepool-org/platform/blob/store/unstructured"
)

type ClientProvider struct {
	AuthClientInvocations            int
	AuthClientStub                   func() auth.Client
	AuthClientOutputs                []auth.Client
	AuthClientOutput                 *auth.Client
	BlobStructuredStoreInvocations   int
	BlobStructuredStoreStub          func() blobStoreStructured.Store
	BlobStructuredStoreOutputs       []blobStoreStructured.Store
	BlobStructuredStoreOutput        *blobStoreStructured.Store
	BlobUnstructuredStoreInvocations int
	BlobUnstructuredStoreStub        func() blobStoreUnstructured.Store
	BlobUnstructuredStoreOutputs     []blobStoreUnstructured.Store
	BlobUnstructuredStoreOutput      *blobStoreUnstructured.Store
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

func (c *ClientProvider) AssertOutputsEmpty() {
	if len(c.AuthClientOutputs) > 0 {
		panic("AuthClientOutputs is not empty")
	}
	if len(c.BlobStructuredStoreOutputs) > 0 {
		panic("BlobStructuredStoreOutputs is not empty")
	}
	if len(c.BlobUnstructuredStoreOutputs) > 0 {
		panic("BlobUnstructuredStoreOutputs is not empty")
	}
}
