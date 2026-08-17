package test

import (
	"context"

	"github.com/onsi/gomega"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/page"
)

type CreateProviderSessionInput struct {
	Context context.Context
	Create  *auth.ProviderSessionCreate
}

type CreateProviderSessionOutput struct {
	ProviderSession *auth.ProviderSession
	Error           error
}

type DeleteProviderSessionsInput struct {
	Context context.Context
	Filter  *auth.ProviderSessionFilter
}

type ListProviderSessionsInput struct {
	Context    context.Context
	Filter     *auth.ProviderSessionFilter
	Pagination *page.Pagination
}

type ListProviderSessionsOutput struct {
	ProviderSessions auth.ProviderSessions
	Error            error
}

type GetProviderSessionInput struct {
	Context context.Context
	ID      string
}

type GetProviderSessionOutput struct {
	ProviderSession *auth.ProviderSession
	Error           error
}

type UpdateProviderSessionInput struct {
	Context context.Context
	ID      string
	Update  *auth.ProviderSessionUpdate
}

type UpdateProviderSessionOutput struct {
	ProviderSession *auth.ProviderSession
	Error           error
}

type DeleteProviderSessionInput struct {
	Context context.Context
	ID      string
}

type ProviderSessionClient struct {
	CreateProviderSessionInvocations  int
	CreateProviderSessionInputs       []CreateProviderSessionInput
	CreateProviderSessionOutputs      []CreateProviderSessionOutput
	DeleteProviderSessionsInvocations int
	DeleteProviderSessionsInputs      []DeleteProviderSessionsInput
	DeleteProviderSessionsOutputs     []error
	ListProviderSessionsInvocations   int
	ListProviderSessionsInputs        []ListProviderSessionsInput
	ListProviderSessionsOutputs       []ListProviderSessionsOutput
	GetProviderSessionInvocations     int
	GetProviderSessionInputs          []GetProviderSessionInput
	GetProviderSessionOutputs         []GetProviderSessionOutput
	UpdateProviderSessionInvocations  int
	UpdateProviderSessionInputs       []UpdateProviderSessionInput
	UpdateProviderSessionOutputs      []UpdateProviderSessionOutput
	DeleteProviderSessionInvocations  int
	DeleteProviderSessionInputs       []DeleteProviderSessionInput
	DeleteProviderSessionOutputs      []error
}

func NewProviderSessionClient() *ProviderSessionClient {
	return &ProviderSessionClient{}
}

func (p *ProviderSessionClient) CreateProviderSession(ctx context.Context, create *auth.ProviderSessionCreate) (*auth.ProviderSession, error) {
	p.CreateProviderSessionInvocations++

	p.CreateProviderSessionInputs = append(p.CreateProviderSessionInputs, CreateProviderSessionInput{Context: ctx, Create: create})

	gomega.Expect(p.CreateProviderSessionOutputs).ToNot(gomega.BeEmpty())

	output := p.CreateProviderSessionOutputs[0]
	p.CreateProviderSessionOutputs = p.CreateProviderSessionOutputs[1:]
	return output.ProviderSession, output.Error
}

func (p *ProviderSessionClient) DeleteProviderSessions(ctx context.Context, filter *auth.ProviderSessionFilter) error {
	p.DeleteProviderSessionsInvocations++

	p.DeleteProviderSessionsInputs = append(p.DeleteProviderSessionsInputs, DeleteProviderSessionsInput{Context: ctx, Filter: filter})

	gomega.Expect(p.DeleteProviderSessionsOutputs).ToNot(gomega.BeEmpty())

	output := p.DeleteProviderSessionsOutputs[0]
	p.DeleteProviderSessionsOutputs = p.DeleteProviderSessionsOutputs[1:]
	return output
}

func (p *ProviderSessionClient) ListProviderSessions(ctx context.Context, filter *auth.ProviderSessionFilter, pagination *page.Pagination) (auth.ProviderSessions, error) {
	p.ListProviderSessionsInvocations++

	p.ListProviderSessionsInputs = append(p.ListProviderSessionsInputs, ListProviderSessionsInput{Context: ctx, Filter: filter, Pagination: pagination})

	gomega.Expect(p.ListProviderSessionsOutputs).ToNot(gomega.BeEmpty())

	output := p.ListProviderSessionsOutputs[0]
	p.ListProviderSessionsOutputs = p.ListProviderSessionsOutputs[1:]
	return output.ProviderSessions, output.Error
}

func (p *ProviderSessionClient) GetProviderSession(ctx context.Context, id string) (*auth.ProviderSession, error) {
	p.GetProviderSessionInvocations++

	p.GetProviderSessionInputs = append(p.GetProviderSessionInputs, GetProviderSessionInput{Context: ctx, ID: id})

	gomega.Expect(p.GetProviderSessionOutputs).ToNot(gomega.BeEmpty())

	output := p.GetProviderSessionOutputs[0]
	p.GetProviderSessionOutputs = p.GetProviderSessionOutputs[1:]
	return output.ProviderSession, output.Error
}

func (p *ProviderSessionClient) UpdateProviderSession(ctx context.Context, id string, update *auth.ProviderSessionUpdate) (*auth.ProviderSession, error) {
	p.UpdateProviderSessionInvocations++

	p.UpdateProviderSessionInputs = append(p.UpdateProviderSessionInputs, UpdateProviderSessionInput{Context: ctx, ID: id, Update: update})

	gomega.Expect(p.UpdateProviderSessionOutputs).ToNot(gomega.BeEmpty())

	output := p.UpdateProviderSessionOutputs[0]
	p.UpdateProviderSessionOutputs = p.UpdateProviderSessionOutputs[1:]
	return output.ProviderSession, output.Error
}

func (p *ProviderSessionClient) DeleteProviderSession(ctx context.Context, id string) error {
	p.DeleteProviderSessionInvocations++

	p.DeleteProviderSessionInputs = append(p.DeleteProviderSessionInputs, DeleteProviderSessionInput{Context: ctx, ID: id})

	gomega.Expect(p.DeleteProviderSessionOutputs).ToNot(gomega.BeEmpty())

	output := p.DeleteProviderSessionOutputs[0]
	p.DeleteProviderSessionOutputs = p.DeleteProviderSessionOutputs[1:]
	return output
}

func (p *ProviderSessionClient) Expectations() {
	gomega.Expect(p.CreateProviderSessionOutputs).To(gomega.BeEmpty())
	gomega.Expect(p.DeleteProviderSessionsOutputs).To(gomega.BeEmpty())
	gomega.Expect(p.ListProviderSessionsOutputs).To(gomega.BeEmpty())
	gomega.Expect(p.GetProviderSessionOutputs).To(gomega.BeEmpty())
	gomega.Expect(p.UpdateProviderSessionOutputs).To(gomega.BeEmpty())
	gomega.Expect(p.DeleteProviderSessionOutputs).To(gomega.BeEmpty())
}
