package test

import (
	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/blob"
)

type Provider struct {
	AuthClientInvocations int
	AuthClientStub        func() auth.Client
	AuthClientOutputs     []auth.Client
	AuthClientOutput      *auth.Client
	BlobClientInvocations int
	BlobClientStub        func() blob.Client
	BlobClientOutputs     []blob.Client
	BlobClientOutput      *blob.Client
}

func NewProvider() *Provider {
	return &Provider{}
}

func (p *Provider) BlobClient() blob.Client {
	p.BlobClientInvocations++
	if p.BlobClientStub != nil {
		return p.BlobClientStub()
	}
	if len(p.BlobClientOutputs) > 0 {
		output := p.BlobClientOutputs[0]
		p.BlobClientOutputs = p.BlobClientOutputs[1:]
		return output
	}
	if p.BlobClientOutput != nil {
		return *p.BlobClientOutput
	}
	panic("BlobClient has no output")
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

func (p *Provider) AssertOutputsEmpty() {
	if len(p.AuthClientOutputs) > 0 {
		panic("AuthClientOutputs is not empty")
	}
	if len(p.BlobClientOutputs) > 0 {
		panic("BlobClientOutputs is not empty")
	}
}
