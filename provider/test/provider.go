package test

import (
	"context"

	"github.com/onsi/gomega"

	"github.com/tidepool-org/platform/auth"
)

type OnCreateInput struct {
	Context         context.Context
	UserID          string
	ProviderSession *auth.ProviderSession
}

type OnDeleteInput struct {
	Context         context.Context
	UserID          string
	ProviderSession *auth.ProviderSession
}

type Provider struct {
	Type                string
	Name                string
	OnCreateInvocations int
	OnCreateInputs      []OnCreateInput
	OnCreateOutputs     []error
	OnDeleteInvocations int
	OnDeleteInputs      []OnDeleteInput
	OnDeleteOutputs     []error
}

func NewProvider(typ string, name string) *Provider {
	return &Provider{
		Type: typ,
		Name: name,
	}
}

func (p *Provider) OnCreate(ctx context.Context, userID string, providerSession *auth.ProviderSession) error {
	p.OnCreateInvocations++

	p.OnCreateInputs = append(p.OnCreateInputs, OnCreateInput{Context: ctx, UserID: userID, ProviderSession: providerSession})

	gomega.Expect(p.OnCreateOutputs).ToNot(gomega.BeEmpty())

	output := p.OnCreateOutputs[0]
	p.OnCreateOutputs = p.OnCreateOutputs[1:]
	return output
}

func (p *Provider) OnDelete(ctx context.Context, userID string, providerSession *auth.ProviderSession) error {
	p.OnDeleteInvocations++

	p.OnDeleteInputs = append(p.OnDeleteInputs, OnDeleteInput{Context: ctx, UserID: userID, ProviderSession: providerSession})

	gomega.Expect(p.OnDeleteOutputs).ToNot(gomega.BeEmpty())

	output := p.OnDeleteOutputs[0]
	p.OnDeleteOutputs = p.OnDeleteOutputs[1:]
	return output
}

func (p *Provider) Expectations() {
	gomega.Expect(p.OnCreateOutputs).To(gomega.BeEmpty())
	gomega.Expect(p.OnDeleteOutputs).To(gomega.BeEmpty())
}
