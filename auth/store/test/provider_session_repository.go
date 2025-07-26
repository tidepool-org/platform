package test

import (
	"context"

	"github.com/onsi/gomega"

	"github.com/tidepool-org/platform/auth"
	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/page"
)

type ListAllProviderSessionsInput struct {
	Context    context.Context
	Filter     auth.ProviderSessionFilter
	Pagination page.Pagination
}

type ListAllProviderSessionsOutput struct {
	ProviderSessions auth.ProviderSessions
	Error            error
}

type ProviderSessionRepository struct {
	*authTest.ProviderSessionAccessor
	ListAllProviderSessionsInvocations int
	ListAllProviderSessionsInputs      []ListAllProviderSessionsInput
	ListAllProviderSessionsOutputs     []ListAllProviderSessionsOutput
}

func NewProviderSessionRepository() *ProviderSessionRepository {
	return &ProviderSessionRepository{
		ProviderSessionAccessor: authTest.NewProviderSessionAccessor(),
	}
}

func (p *ProviderSessionRepository) ListAllProviderSessions(ctx context.Context, filter auth.ProviderSessionFilter, pagination page.Pagination) (auth.ProviderSessions, error) {
	p.ListAllProviderSessionsInvocations++

	p.ListAllProviderSessionsInputs = append(p.ListAllProviderSessionsInputs, ListAllProviderSessionsInput{Context: ctx, Filter: filter, Pagination: pagination})

	gomega.Expect(p.ListAllProviderSessionsOutputs).ToNot(gomega.BeEmpty())

	output := p.ListAllProviderSessionsOutputs[0]
	p.ListAllProviderSessionsOutputs = p.ListAllProviderSessionsOutputs[1:]
	return output.ProviderSessions, output.Error
}

func (p *ProviderSessionRepository) Expectations() {
	p.ProviderSessionAccessor.Expectations()
	gomega.Expect(p.ListAllProviderSessionsOutputs).To(gomega.BeEmpty())
}
