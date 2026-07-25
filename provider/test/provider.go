package test

import (
	"context"

	"github.com/onsi/gomega"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/provider"
)

var _ provider.Provider = &Provider{}

type OnCreateInput struct {
	Context         context.Context
	ProviderSession *auth.ProviderSession
}

type OnDeleteInput struct {
	Context         context.Context
	ProviderSession *auth.ProviderSession
}

type OnRefreshInput struct {
	Context         context.Context
	ProviderSession *auth.ProviderSession
	Refresh         *auth.ProviderSessionRefresh
}

type Provider struct {
	typ                  string
	name                 string
	OnCreateInvocations  int
	OnCreateInputs       []OnCreateInput
	OnCreateOutputs      []error
	OnDeleteInvocations  int
	OnDeleteInputs       []OnDeleteInput
	OnDeleteOutputs      []error
	OnRefreshInvocations int
	OnRefreshInputs      []OnRefreshInput
	OnRefreshOutputs     []error
}

func NewProvider(typ string, name string) *Provider {
	return &Provider{
		typ:  typ,
		name: name,
	}
}

func (p *Provider) Type() string {
	return p.typ
}

func (p *Provider) Name() string {
	return p.name
}

func (p *Provider) OnCreate(ctx context.Context, providerSession *auth.ProviderSession) error {
	p.OnCreateInvocations++

	p.OnCreateInputs = append(p.OnCreateInputs, OnCreateInput{Context: ctx, ProviderSession: providerSession})

	gomega.Expect(p.OnCreateOutputs).ToNot(gomega.BeEmpty())

	output := p.OnCreateOutputs[0]
	p.OnCreateOutputs = p.OnCreateOutputs[1:]
	return output
}

func (p *Provider) OnDelete(ctx context.Context, providerSession *auth.ProviderSession) error {
	p.OnDeleteInvocations++

	p.OnDeleteInputs = append(p.OnDeleteInputs, OnDeleteInput{Context: ctx, ProviderSession: providerSession})

	gomega.Expect(p.OnDeleteOutputs).ToNot(gomega.BeEmpty())

	output := p.OnDeleteOutputs[0]
	p.OnDeleteOutputs = p.OnDeleteOutputs[1:]
	return output
}

func (p *Provider) OnRefresh(ctx context.Context, providerSession *auth.ProviderSession, refresh *auth.ProviderSessionRefresh) error {
	p.OnRefreshInvocations++

	p.OnRefreshInputs = append(p.OnRefreshInputs, OnRefreshInput{Context: ctx, ProviderSession: providerSession, Refresh: refresh})

	gomega.Expect(p.OnRefreshOutputs).ToNot(gomega.BeEmpty())

	output := p.OnRefreshOutputs[0]
	p.OnRefreshOutputs = p.OnRefreshOutputs[1:]
	return output
}

func (p *Provider) Expectations() {
	gomega.Expect(p.OnCreateOutputs).To(gomega.BeEmpty())
	gomega.Expect(p.OnDeleteOutputs).To(gomega.BeEmpty())
	gomega.Expect(p.OnRefreshOutputs).To(gomega.BeEmpty())
}
