package test

import (
	"context"

	"github.com/onsi/gomega"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/page"
)

type ListUserProviderSessionsInput struct {
	Context    context.Context
	UserID     string
	Filter     *auth.ProviderSessionFilter
	Pagination *page.Pagination
}

type ListUserProviderSessionsOutput struct {
	ProviderSessions auth.ProviderSessions
	Error            error
}

type CreateUserProviderSessionInput struct {
	Context context.Context
	UserID  string
	Create  *auth.ProviderSessionCreate
}

type CreateUserProviderSessionOutput struct {
	ProviderSession *auth.ProviderSession
	Error           error
}

type DeleteAllProviderSessionsInput struct {
	Context context.Context
	UserID  string
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

type ProviderSessionAccessor struct {
	ListUserProviderSessionsInvocations  int
	ListUserProviderSessionsInputs       []ListUserProviderSessionsInput
	ListUserProviderSessionsOutputs      []ListUserProviderSessionsOutput
	CreateUserProviderSessionInvocations int
	CreateUserProviderSessionInputs      []CreateUserProviderSessionInput
	CreateUserProviderSessionOutputs     []CreateUserProviderSessionOutput
	DeleteAllProviderSessionsInvocations int
	DeleteAllProviderSessionsInputs      []DeleteAllProviderSessionsInput
	DeleteAllProviderSessionsOutputs     []error
	GetProviderSessionInvocations        int
	GetProviderSessionInputs             []GetProviderSessionInput
	GetProviderSessionOutputs            []GetProviderSessionOutput
	UpdateProviderSessionInvocations     int
	UpdateProviderSessionInputs          []UpdateProviderSessionInput
	UpdateProviderSessionOutputs         []UpdateProviderSessionOutput
	DeleteProviderSessionInvocations     int
	DeleteProviderSessionInputs          []DeleteProviderSessionInput
	DeleteProviderSessionOutputs         []error
}

func NewProviderSessionAccessor() *ProviderSessionAccessor {
	return &ProviderSessionAccessor{}
}

func (p *ProviderSessionAccessor) ListUserProviderSessions(ctx context.Context, userID string, filter *auth.ProviderSessionFilter, pagination *page.Pagination) (auth.ProviderSessions, error) {
	p.ListUserProviderSessionsInvocations++

	p.ListUserProviderSessionsInputs = append(p.ListUserProviderSessionsInputs, ListUserProviderSessionsInput{Context: ctx, UserID: userID, Filter: filter, Pagination: pagination})

	gomega.Expect(p.ListUserProviderSessionsOutputs).ToNot(gomega.BeEmpty())

	output := p.ListUserProviderSessionsOutputs[0]
	p.ListUserProviderSessionsOutputs = p.ListUserProviderSessionsOutputs[1:]
	return output.ProviderSessions, output.Error
}

func (p *ProviderSessionAccessor) CreateUserProviderSession(ctx context.Context, userID string, create *auth.ProviderSessionCreate) (*auth.ProviderSession, error) {
	p.CreateUserProviderSessionInvocations++

	p.CreateUserProviderSessionInputs = append(p.CreateUserProviderSessionInputs, CreateUserProviderSessionInput{Context: ctx, UserID: userID, Create: create})

	gomega.Expect(p.CreateUserProviderSessionOutputs).ToNot(gomega.BeEmpty())

	output := p.CreateUserProviderSessionOutputs[0]
	p.CreateUserProviderSessionOutputs = p.CreateUserProviderSessionOutputs[1:]
	return output.ProviderSession, output.Error
}

func (p *ProviderSessionAccessor) DeleteAllProviderSessions(ctx context.Context, userID string) error {
	p.DeleteAllProviderSessionsInvocations++

	p.DeleteAllProviderSessionsInputs = append(p.DeleteAllProviderSessionsInputs, DeleteAllProviderSessionsInput{Context: ctx, UserID: userID})

	gomega.Expect(p.DeleteAllProviderSessionsOutputs).ToNot(gomega.BeEmpty())

	output := p.DeleteAllProviderSessionsOutputs[0]
	p.DeleteAllProviderSessionsOutputs = p.DeleteAllProviderSessionsOutputs[1:]
	return output
}

func (p *ProviderSessionAccessor) GetProviderSession(ctx context.Context, id string) (*auth.ProviderSession, error) {
	p.GetProviderSessionInvocations++

	p.GetProviderSessionInputs = append(p.GetProviderSessionInputs, GetProviderSessionInput{Context: ctx, ID: id})

	gomega.Expect(p.GetProviderSessionOutputs).ToNot(gomega.BeEmpty())

	output := p.GetProviderSessionOutputs[0]
	p.GetProviderSessionOutputs = p.GetProviderSessionOutputs[1:]
	return output.ProviderSession, output.Error
}

func (p *ProviderSessionAccessor) UpdateProviderSession(ctx context.Context, id string, update *auth.ProviderSessionUpdate) (*auth.ProviderSession, error) {
	p.UpdateProviderSessionInvocations++

	p.UpdateProviderSessionInputs = append(p.UpdateProviderSessionInputs, UpdateProviderSessionInput{Context: ctx, ID: id, Update: update})

	gomega.Expect(p.UpdateProviderSessionOutputs).ToNot(gomega.BeEmpty())

	output := p.UpdateProviderSessionOutputs[0]
	p.UpdateProviderSessionOutputs = p.UpdateProviderSessionOutputs[1:]
	return output.ProviderSession, output.Error
}

func (p *ProviderSessionAccessor) DeleteProviderSession(ctx context.Context, id string) error {
	p.DeleteProviderSessionInvocations++

	p.DeleteProviderSessionInputs = append(p.DeleteProviderSessionInputs, DeleteProviderSessionInput{Context: ctx, ID: id})

	gomega.Expect(p.DeleteProviderSessionOutputs).ToNot(gomega.BeEmpty())

	output := p.DeleteProviderSessionOutputs[0]
	p.DeleteProviderSessionOutputs = p.DeleteProviderSessionOutputs[1:]
	return output
}

func (p *ProviderSessionAccessor) Expectations() {
	gomega.Expect(p.ListUserProviderSessionsOutputs).To(gomega.BeEmpty())
	gomega.Expect(p.CreateUserProviderSessionOutputs).To(gomega.BeEmpty())
	gomega.Expect(p.DeleteAllProviderSessionsOutputs).To(gomega.BeEmpty())
	gomega.Expect(p.GetProviderSessionOutputs).To(gomega.BeEmpty())
	gomega.Expect(p.UpdateProviderSessionOutputs).To(gomega.BeEmpty())
	gomega.Expect(p.DeleteProviderSessionOutputs).To(gomega.BeEmpty())
}
