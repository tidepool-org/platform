package test

import (
	"github.com/tidepool-org/platform/auth"
	dataSourceStoreStructured "github.com/tidepool-org/platform/data/source/store/structured"
)

type Provider struct {
	AuthClientInvocations                int
	AuthClientStub                       func() auth.Client
	AuthClientOutputs                    []auth.Client
	AuthClientOutput                     *auth.Client
	DataSourceStructuredStoreInvocations int
	DataSourceStructuredStoreStub        func() dataSourceStoreStructured.Store
	DataSourceStructuredStoreOutputs     []dataSourceStoreStructured.Store
	DataSourceStructuredStoreOutput      *dataSourceStoreStructured.Store
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

func (p *Provider) DataSourceStructuredStore() dataSourceStoreStructured.Store {
	p.DataSourceStructuredStoreInvocations++
	if p.DataSourceStructuredStoreStub != nil {
		return p.DataSourceStructuredStoreStub()
	}
	if len(p.DataSourceStructuredStoreOutputs) > 0 {
		output := p.DataSourceStructuredStoreOutputs[0]
		p.DataSourceStructuredStoreOutputs = p.DataSourceStructuredStoreOutputs[1:]
		return output
	}
	if p.DataSourceStructuredStoreOutput != nil {
		return *p.DataSourceStructuredStoreOutput
	}
	panic("DataSourceStructuredStore has no output")
}

func (p *Provider) AssertOutputsEmpty() {
	if len(p.AuthClientOutputs) > 0 {
		panic("AuthClientOutputs is not empty")
	}
	if len(p.DataSourceStructuredStoreOutputs) > 0 {
		panic("DataSourceStructuredStoreOutputs is not empty")
	}
}
