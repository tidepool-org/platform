package test

import (
	"context"
	"github.com/onsi/gomega"
	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/page"
)

type GetUserDeviceAuthorizationInput struct {
	ctx    context.Context
	userID string
	id     string
}

type GetUserDeviceAuthorizationOutput struct {
	authorization *auth.DeviceAuthorization
	err           error
}

type ListUserDeviceAuthorizationsInput struct {
	ctx        context.Context
	userID     string
	pagination *page.Pagination
}

type ListUserDeviceAuthorizationsOutput struct {
	authorizations auth.DeviceAuthorizations
	err            error
}

type GetDeviceAuthorizationByTokenInput struct {
	ctx   context.Context
	token string
}

type GetDeviceAuthorizationByTokenOutput struct {
	authorization *auth.DeviceAuthorization
	err           error
}

type CreateUserDeviceAuthorizationInput struct {
	ctx    context.Context
	userID string
	create *auth.DeviceAuthorizationCreate
}

type CreateUserDeviceAuthorizationOutput struct {
	authorization *auth.DeviceAuthorization
	err           error
}

type UpdateDeviceAuthorizationInput struct {
	ctx    context.Context
	id     string
	update *auth.DeviceAuthorizationUpdate
}

type UpdateDeviceAuthorizationOutput struct {
	authorization *auth.DeviceAuthorization
	err           error
}

type DeviceAuthorizationAccessor struct {
	GetUserDeviceAuthorizationInvocations    int
	GetUserDeviceAuthorizationInputs         []GetUserDeviceAuthorizationInput
	GetUserDeviceAuthorizationOutputs        []GetUserDeviceAuthorizationOutput
	ListUserDeviceAuthorizationsInvocations  int
	ListUserDeviceAuthorizationsInputs       []ListUserDeviceAuthorizationsInput
	ListUserDeviceAuthorizationsOutputs      []ListUserDeviceAuthorizationsOutput
	GetDeviceAuthorizationByTokenInvocations int
	GetDeviceAuthorizationByTokenInputs      []GetDeviceAuthorizationByTokenInput
	GetDeviceAuthorizationByTokenOutputs     []GetDeviceAuthorizationByTokenOutput
	CreateUserDeviceAuthorizationInvocations int
	CreateUserDeviceAuthorizationInputs      []CreateUserDeviceAuthorizationInput
	CreateUserDeviceAuthorizationOutputs     []CreateUserDeviceAuthorizationOutput
	UpdateDeviceAuthorizationInvocations     int
	UpdateDeviceAuthorizationInputs          []UpdateDeviceAuthorizationInput
	UpdateDeviceAuthorizationOutputs         []UpdateDeviceAuthorizationOutput
}

func NewDeviceAuthorizationAccessor() *DeviceAuthorizationAccessor {
	return &DeviceAuthorizationAccessor{}
}

func (d *DeviceAuthorizationAccessor) GetUserDeviceAuthorization(ctx context.Context, userID string, id string) (*auth.DeviceAuthorization, error) {
	d.GetUserDeviceAuthorizationInvocations++

	d.GetUserDeviceAuthorizationInputs = append(d.GetUserDeviceAuthorizationInputs, GetUserDeviceAuthorizationInput{ctx, userID, id})

	gomega.Expect(d.GetUserDeviceAuthorizationOutputs).ToNot(gomega.BeEmpty())

	output := d.GetUserDeviceAuthorizationOutputs[0]
	d.GetUserDeviceAuthorizationOutputs = d.GetUserDeviceAuthorizationOutputs[1:]
	return output.authorization, output.err
}

func (d *DeviceAuthorizationAccessor) ListUserDeviceAuthorizations(ctx context.Context, userID string, pagination *page.Pagination) (auth.DeviceAuthorizations, error) {
	d.ListUserDeviceAuthorizationsInvocations++

	d.ListUserDeviceAuthorizationsInputs = append(d.ListUserDeviceAuthorizationsInputs, ListUserDeviceAuthorizationsInput{ctx, userID, pagination})

	gomega.Expect(d.ListUserDeviceAuthorizationsOutputs).ToNot(gomega.BeEmpty())

	output := d.ListUserDeviceAuthorizationsOutputs[0]
	d.ListUserDeviceAuthorizationsOutputs = d.ListUserDeviceAuthorizationsOutputs[1:]
	return output.authorizations, output.err
}

func (d *DeviceAuthorizationAccessor) GetDeviceAuthorizationByToken(ctx context.Context, token string) (*auth.DeviceAuthorization, error) {
	d.GetDeviceAuthorizationByTokenInvocations++

	d.GetDeviceAuthorizationByTokenInputs = append(d.GetDeviceAuthorizationByTokenInputs, GetDeviceAuthorizationByTokenInput{ctx, token})

	gomega.Expect(d.GetDeviceAuthorizationByTokenOutputs).ToNot(gomega.BeEmpty())

	output := d.GetDeviceAuthorizationByTokenOutputs[0]
	d.GetDeviceAuthorizationByTokenOutputs = d.GetDeviceAuthorizationByTokenOutputs[1:]
	return output.authorization, output.err
}

func (d *DeviceAuthorizationAccessor) CreateUserDeviceAuthorization(ctx context.Context, userID string, create *auth.DeviceAuthorizationCreate) (*auth.DeviceAuthorization, error) {
	d.CreateUserDeviceAuthorizationInvocations++

	d.CreateUserDeviceAuthorizationInputs = append(d.CreateUserDeviceAuthorizationInputs, CreateUserDeviceAuthorizationInput{ctx, userID, create})

	gomega.Expect(d.CreateUserDeviceAuthorizationOutputs).ToNot(gomega.BeEmpty())

	output := d.CreateUserDeviceAuthorizationOutputs[0]
	d.CreateUserDeviceAuthorizationOutputs = d.CreateUserDeviceAuthorizationOutputs[1:]
	return output.authorization, output.err
}

func (d *DeviceAuthorizationAccessor) UpdateDeviceAuthorization(ctx context.Context, id string, update *auth.DeviceAuthorizationUpdate) (*auth.DeviceAuthorization, error) {
	d.UpdateDeviceAuthorizationInvocations++

	d.UpdateDeviceAuthorizationInputs = append(d.UpdateDeviceAuthorizationInputs, UpdateDeviceAuthorizationInput{ctx, id, update})

	gomega.Expect(d.UpdateDeviceAuthorizationOutputs).ToNot(gomega.BeEmpty())

	output := d.UpdateDeviceAuthorizationOutputs[0]
	d.UpdateDeviceAuthorizationOutputs = d.UpdateDeviceAuthorizationOutputs[1:]
	return output.authorization, output.err
}

func (d *DeviceAuthorizationAccessor) Expectations() {
	gomega.Expect(d.GetUserDeviceAuthorizationOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.ListUserDeviceAuthorizationsOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.GetDeviceAuthorizationByTokenOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.CreateUserDeviceAuthorizationOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.UpdateDeviceAuthorizationOutputs).To(gomega.BeEmpty())
}
