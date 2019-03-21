package test

import (
	"context"

	"github.com/onsi/gomega"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/page"
)

type ListUserRestrictedTokensInput struct {
	Context    context.Context
	UserID     string
	Filter     *auth.RestrictedTokenFilter
	Pagination *page.Pagination
}

type ListUserRestrictedTokensOutput struct {
	RestrictedTokens auth.RestrictedTokens
	Error            error
}

type CreateUserRestrictedTokenInput struct {
	Context context.Context
	UserID  string
	Create  *auth.RestrictedTokenCreate
}

type CreateUserRestrictedTokenOutput struct {
	RestrictedToken *auth.RestrictedToken
	Error           error
}

type DeleteAllRestrictedTokensInput struct {
	Context context.Context
	UserID  string
}

type GetRestrictedTokenInput struct {
	Context context.Context
	ID      string
}

type GetRestrictedTokenOutput struct {
	RestrictedToken *auth.RestrictedToken
	Error           error
}

type UpdateRestrictedTokenInput struct {
	Context context.Context
	ID      string
	Update  *auth.RestrictedTokenUpdate
}

type UpdateRestrictedTokenOutput struct {
	RestrictedToken *auth.RestrictedToken
	Error           error
}

type DeleteRestrictedTokenInput struct {
	Context context.Context
	ID      string
}

type RestrictedTokenAccessor struct {
	ListUserRestrictedTokensInvocations  int
	ListUserRestrictedTokensInputs       []ListUserRestrictedTokensInput
	ListUserRestrictedTokensOutputs      []ListUserRestrictedTokensOutput
	CreateUserRestrictedTokenInvocations int
	CreateUserRestrictedTokenInputs      []CreateUserRestrictedTokenInput
	CreateUserRestrictedTokenOutputs     []CreateUserRestrictedTokenOutput
	DeleteAllRestrictedTokensInvocations int
	DeleteAllRestrictedTokensInputs      []DeleteAllRestrictedTokensInput
	DeleteAllRestrictedTokensOutputs     []error
	GetRestrictedTokenInvocations        int
	GetRestrictedTokenInputs             []GetRestrictedTokenInput
	GetRestrictedTokenOutputs            []GetRestrictedTokenOutput
	UpdateRestrictedTokenInvocations     int
	UpdateRestrictedTokenInputs          []UpdateRestrictedTokenInput
	UpdateRestrictedTokenOutputs         []UpdateRestrictedTokenOutput
	DeleteRestrictedTokenInvocations     int
	DeleteRestrictedTokenInputs          []DeleteRestrictedTokenInput
	DeleteRestrictedTokenOutputs         []error
}

func NewRestrictedTokenAccessor() *RestrictedTokenAccessor {
	return &RestrictedTokenAccessor{}
}

func (r *RestrictedTokenAccessor) ListUserRestrictedTokens(ctx context.Context, userID string, filter *auth.RestrictedTokenFilter, pagination *page.Pagination) (auth.RestrictedTokens, error) {
	r.ListUserRestrictedTokensInvocations++

	r.ListUserRestrictedTokensInputs = append(r.ListUserRestrictedTokensInputs, ListUserRestrictedTokensInput{Context: ctx, UserID: userID, Filter: filter, Pagination: pagination})

	gomega.Expect(r.ListUserRestrictedTokensOutputs).ToNot(gomega.BeEmpty())

	output := r.ListUserRestrictedTokensOutputs[0]
	r.ListUserRestrictedTokensOutputs = r.ListUserRestrictedTokensOutputs[1:]
	return output.RestrictedTokens, output.Error
}

func (r *RestrictedTokenAccessor) CreateUserRestrictedToken(ctx context.Context, userID string, create *auth.RestrictedTokenCreate) (*auth.RestrictedToken, error) {
	r.CreateUserRestrictedTokenInvocations++

	r.CreateUserRestrictedTokenInputs = append(r.CreateUserRestrictedTokenInputs, CreateUserRestrictedTokenInput{Context: ctx, UserID: userID, Create: create})

	gomega.Expect(r.CreateUserRestrictedTokenOutputs).ToNot(gomega.BeEmpty())

	output := r.CreateUserRestrictedTokenOutputs[0]
	r.CreateUserRestrictedTokenOutputs = r.CreateUserRestrictedTokenOutputs[1:]
	return output.RestrictedToken, output.Error
}

func (r *RestrictedTokenAccessor) DeleteAllRestrictedTokens(ctx context.Context, userID string) error {
	r.DeleteAllRestrictedTokensInvocations++

	r.DeleteAllRestrictedTokensInputs = append(r.DeleteAllRestrictedTokensInputs, DeleteAllRestrictedTokensInput{Context: ctx, UserID: userID})

	gomega.Expect(r.DeleteAllRestrictedTokensOutputs).ToNot(gomega.BeEmpty())

	output := r.DeleteAllRestrictedTokensOutputs[0]
	r.DeleteAllRestrictedTokensOutputs = r.DeleteAllRestrictedTokensOutputs[1:]
	return output
}

func (r *RestrictedTokenAccessor) GetRestrictedToken(ctx context.Context, id string) (*auth.RestrictedToken, error) {
	r.GetRestrictedTokenInvocations++

	r.GetRestrictedTokenInputs = append(r.GetRestrictedTokenInputs, GetRestrictedTokenInput{Context: ctx, ID: id})

	gomega.Expect(r.GetRestrictedTokenOutputs).ToNot(gomega.BeEmpty())

	output := r.GetRestrictedTokenOutputs[0]
	r.GetRestrictedTokenOutputs = r.GetRestrictedTokenOutputs[1:]
	return output.RestrictedToken, output.Error
}

func (r *RestrictedTokenAccessor) UpdateRestrictedToken(ctx context.Context, id string, update *auth.RestrictedTokenUpdate) (*auth.RestrictedToken, error) {
	r.UpdateRestrictedTokenInvocations++

	r.UpdateRestrictedTokenInputs = append(r.UpdateRestrictedTokenInputs, UpdateRestrictedTokenInput{Context: ctx, ID: id, Update: update})

	gomega.Expect(r.UpdateRestrictedTokenOutputs).ToNot(gomega.BeEmpty())

	output := r.UpdateRestrictedTokenOutputs[0]
	r.UpdateRestrictedTokenOutputs = r.UpdateRestrictedTokenOutputs[1:]
	return output.RestrictedToken, output.Error
}

func (r *RestrictedTokenAccessor) DeleteRestrictedToken(ctx context.Context, id string) error {
	r.DeleteRestrictedTokenInvocations++

	r.DeleteRestrictedTokenInputs = append(r.DeleteRestrictedTokenInputs, DeleteRestrictedTokenInput{Context: ctx, ID: id})

	gomega.Expect(r.DeleteRestrictedTokenOutputs).ToNot(gomega.BeEmpty())

	output := r.DeleteRestrictedTokenOutputs[0]
	r.DeleteRestrictedTokenOutputs = r.DeleteRestrictedTokenOutputs[1:]
	return output
}

func (r *RestrictedTokenAccessor) Expectations() {
	gomega.Expect(r.ListUserRestrictedTokensOutputs).To(gomega.BeEmpty())
	gomega.Expect(r.CreateUserRestrictedTokenOutputs).To(gomega.BeEmpty())
	gomega.Expect(r.DeleteAllRestrictedTokensOutputs).To(gomega.BeEmpty())
	gomega.Expect(r.GetRestrictedTokenOutputs).To(gomega.BeEmpty())
	gomega.Expect(r.UpdateRestrictedTokenOutputs).To(gomega.BeEmpty())
	gomega.Expect(r.DeleteRestrictedTokenOutputs).To(gomega.BeEmpty())
}
