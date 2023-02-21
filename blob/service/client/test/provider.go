package test

import (
	"github.com/tidepool-org/platform/auth"
	blobStoreStructured "github.com/tidepool-org/platform/blob/store/structured"
	blobStoreUnstructured "github.com/tidepool-org/platform/blob/store/unstructured"
)

type Provider struct {
	AuthClientInvocations                     int
	AuthClientStub                            func() auth.Client
	AuthClientOutputs                         []auth.Client
	AuthClientOutput                          *auth.Client
	BlobStructuredStoreInvocations            int
	BlobStructuredStoreStub                   func() blobStoreStructured.Store
	BlobStructuredStoreOutputs                []blobStoreStructured.Store
	BlobStructuredStoreOutput                 *blobStoreStructured.Store
	BlobUnstructuredStoreInvocations          int
	BlobUnstructuredStoreStub                 func() blobStoreUnstructured.Store
	BlobUnstructuredStoreOutputs              []blobStoreUnstructured.Store
	BlobUnstructuredStoreOutput               *blobStoreUnstructured.Store
	DeviceLogBlobUnstructuredStoreInvocations int
	DeviceLogBlobUnstructuredStoreStub        func() blobStoreUnstructured.Store
	DeviceLogBlobUnstructuredStoreOutputs     []blobStoreUnstructured.Store
	DeviceLogBlobUnstructuredStoreOutput      *blobStoreUnstructured.Store
}

func NewProvider() *Provider {
	return &Provider{}
}

func (p *Provider) AuthClient() auth.Client {
	p.AuthClientInvocations++
	if p.AuthClientStub != nil {
		return p.AuthClientStub()
	}
	if len(p.AuthClientOutputs) > 0 {
		output := p.AuthClientOutputs[0]
		p.AuthClientOutputs = p.AuthClientOutputs[1:]
		return output
	}
	if p.AuthClientOutput != nil {
		return *p.AuthClientOutput
	}
	panic("AuthClient has no output")
}

func (p *Provider) BlobStructuredStore() blobStoreStructured.Store {
	p.BlobStructuredStoreInvocations++
	if p.BlobStructuredStoreStub != nil {
		return p.BlobStructuredStoreStub()
	}
	if len(p.BlobStructuredStoreOutputs) > 0 {
		output := p.BlobStructuredStoreOutputs[0]
		p.BlobStructuredStoreOutputs = p.BlobStructuredStoreOutputs[1:]
		return output
	}
	if p.BlobStructuredStoreOutput != nil {
		return *p.BlobStructuredStoreOutput
	}
	panic("BlobStructuredStore has no output")
}

func (p *Provider) BlobUnstructuredStore() blobStoreUnstructured.Store {
	p.BlobUnstructuredStoreInvocations++
	if p.BlobUnstructuredStoreStub != nil {
		return p.BlobUnstructuredStoreStub()
	}
	if len(p.BlobUnstructuredStoreOutputs) > 0 {
		output := p.BlobUnstructuredStoreOutputs[0]
		p.BlobUnstructuredStoreOutputs = p.BlobUnstructuredStoreOutputs[1:]
		return output
	}
	if p.BlobUnstructuredStoreOutput != nil {
		return *p.BlobUnstructuredStoreOutput
	}
	panic("BlobUnstructuredStore has no output")
}

func (p *Provider) DeviceLogsUnstructuredStore() blobStoreUnstructured.Store {
	p.DeviceLogBlobUnstructuredStoreInvocations++
	if p.DeviceLogBlobUnstructuredStoreStub != nil {
		return p.DeviceLogBlobUnstructuredStoreStub()
	}
	if len(p.DeviceLogBlobUnstructuredStoreOutputs) > 0 {
		output := p.DeviceLogBlobUnstructuredStoreOutputs[0]
		p.DeviceLogBlobUnstructuredStoreOutputs = p.DeviceLogBlobUnstructuredStoreOutputs[1:]
		return output
	}
	if p.DeviceLogBlobUnstructuredStoreOutput != nil {
		return *p.DeviceLogBlobUnstructuredStoreOutput
	}
	panic("DeviceLogsUnstructuredStore has no output")
}

func (p *Provider) AssertOutputsEmpty() {
	if len(p.AuthClientOutputs) > 0 {
		panic("AuthClientOutputs is not empty")
	}
	if len(p.BlobStructuredStoreOutputs) > 0 {
		panic("BlobStructuredStoreOutputs is not empty")
	}
	if len(p.BlobUnstructuredStoreOutputs) > 0 {
		panic("BlobUnstructuredStoreOutputs is not empty")
	}
}
