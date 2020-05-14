package test

import (
	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/blob"
	confirmationStore "github.com/tidepool-org/platform/confirmation/store"
	dataClient "github.com/tidepool-org/platform/data/client"
	dataSource "github.com/tidepool-org/platform/data/source"
	"github.com/tidepool-org/platform/image"
	messageStore "github.com/tidepool-org/platform/message/store"
	"github.com/tidepool-org/platform/permission"
	permissionStore "github.com/tidepool-org/platform/permission/store"
	profileStoreStructured "github.com/tidepool-org/platform/profile/store/structured"
	sessionStore "github.com/tidepool-org/platform/session/store"
	userServiceClient "github.com/tidepool-org/platform/user/service/client"
	userStoreStructured "github.com/tidepool-org/platform/user/store/structured"
)

type Provider struct {
	AuthClientInvocations          int
	AuthClientStub                 func() auth.Client
	AuthClientOutputs              []auth.Client
	AuthClientOutput               *auth.Client
	BlobClientInvocations          int
	BlobClientStub                 func() blob.Client
	BlobClientOutputs              []blob.Client
	BlobClientOutput               *blob.Client
	DataClientInvocations          int
	DataClientStub                 func() dataClient.Client
	DataClientOutputs              []dataClient.Client
	DataClientOutput               *dataClient.Client
	DataSourceClientInvocations    int
	DataSourceClientStub           func() dataSource.Client
	DataSourceClientOutputs        []dataSource.Client
	DataSourceClientOutput         *dataSource.Client
	ImageClientInvocations         int
	ImageClientStub                func() image.Client
	ImageClientOutputs             []image.Client
	ImageClientOutput              *image.Client
	PermissionClientInvocations    int
	PermissionClientStub           func() permission.Client
	PermissionClientOutputs        []permission.Client
	PermissionClientOutput         *permission.Client
	ConfirmationStoreInvocations   int
	ConfirmationStoreStub          func() confirmationStore.Store
	ConfirmationStoreOutputs       []confirmationStore.Store
	ConfirmationStoreOutput        *confirmationStore.Store
	MessageStoreInvocations        int
	MessageStoreStub               func() messageStore.Store
	MessageStoreOutputs            []messageStore.Store
	MessageStoreOutput             *messageStore.Store
	PermissionStoreInvocations     int
	PermissionStoreStub            func() permissionStore.Store
	PermissionStoreOutputs         []permissionStore.Store
	PermissionStoreOutput          *permissionStore.Store
	ProfileStoreInvocations        int
	ProfileStoreStub               func() profileStoreStructured.Store
	ProfileStoreOutputs            []profileStoreStructured.Store
	ProfileStoreOutput             *profileStoreStructured.Store
	SessionStoreInvocations        int
	SessionStoreStub               func() sessionStore.Store
	SessionStoreOutputs            []sessionStore.Store
	SessionStoreOutput             *sessionStore.Store
	UserStructuredStoreInvocations int
	UserStructuredStoreStub        func() userStoreStructured.Store
	UserStructuredStoreOutputs     []userStoreStructured.Store
	UserStructuredStoreOutput      *userStoreStructured.Store
	PasswordHasherInvocations      int
	PasswordHasherStub             func() userServiceClient.PasswordHasher
	PasswordHasherOutputs          []userServiceClient.PasswordHasher
	PasswordHasherOutput           *userServiceClient.PasswordHasher
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

func (p *Provider) DataClient() dataClient.Client {
	p.DataClientInvocations++
	if p.DataClientStub != nil {
		return p.DataClientStub()
	}
	if len(p.DataClientOutputs) > 0 {
		output := p.DataClientOutputs[0]
		p.DataClientOutputs = p.DataClientOutputs[1:]
		return output
	}
	if p.DataClientOutput != nil {
		return *p.DataClientOutput
	}
	panic("DataClient has no output")
}

func (p *Provider) DataSourceClient() dataSource.Client {
	p.DataSourceClientInvocations++
	if p.DataSourceClientStub != nil {
		return p.DataSourceClientStub()
	}
	if len(p.DataSourceClientOutputs) > 0 {
		output := p.DataSourceClientOutputs[0]
		p.DataSourceClientOutputs = p.DataSourceClientOutputs[1:]
		return output
	}
	if p.DataSourceClientOutput != nil {
		return *p.DataSourceClientOutput
	}
	panic("DataSourceClient has no output")
}

func (p *Provider) ImageClient() image.Client {
	p.ImageClientInvocations++
	if p.ImageClientStub != nil {
		return p.ImageClientStub()
	}
	if len(p.ImageClientOutputs) > 0 {
		output := p.ImageClientOutputs[0]
		p.ImageClientOutputs = p.ImageClientOutputs[1:]
		return output
	}
	if p.ImageClientOutput != nil {
		return *p.ImageClientOutput
	}
	panic("ImageClient has no output")
}

func (p *Provider) PermissionClient() permission.Client {
	p.PermissionClientInvocations++
	if p.PermissionClientStub != nil {
		return p.PermissionClientStub()
	}
	if len(p.PermissionClientOutputs) > 0 {
		output := p.PermissionClientOutputs[0]
		p.PermissionClientOutputs = p.PermissionClientOutputs[1:]
		return output
	}
	if p.PermissionClientOutput != nil {
		return *p.PermissionClientOutput
	}
	panic("PermissionClient has no output")
}

func (p *Provider) ConfirmationStore() confirmationStore.Store {
	p.ConfirmationStoreInvocations++
	if p.ConfirmationStoreStub != nil {
		return p.ConfirmationStoreStub()
	}
	if len(p.ConfirmationStoreOutputs) > 0 {
		output := p.ConfirmationStoreOutputs[0]
		p.ConfirmationStoreOutputs = p.ConfirmationStoreOutputs[1:]
		return output
	}
	if p.ConfirmationStoreOutput != nil {
		return *p.ConfirmationStoreOutput
	}
	panic("ConfirmationStore has no output")
}

func (p *Provider) MessageStore() messageStore.Store {
	p.MessageStoreInvocations++
	if p.MessageStoreStub != nil {
		return p.MessageStoreStub()
	}
	if len(p.MessageStoreOutputs) > 0 {
		output := p.MessageStoreOutputs[0]
		p.MessageStoreOutputs = p.MessageStoreOutputs[1:]
		return output
	}
	if p.MessageStoreOutput != nil {
		return *p.MessageStoreOutput
	}
	panic("MessageStore has no output")
}

func (p *Provider) PermissionStore() permissionStore.Store {
	p.PermissionStoreInvocations++
	if p.PermissionStoreStub != nil {
		return p.PermissionStoreStub()
	}
	if len(p.PermissionStoreOutputs) > 0 {
		output := p.PermissionStoreOutputs[0]
		p.PermissionStoreOutputs = p.PermissionStoreOutputs[1:]
		return output
	}
	if p.PermissionStoreOutput != nil {
		return *p.PermissionStoreOutput
	}
	panic("PermissionStore has no output")
}

func (p *Provider) ProfileStore() profileStoreStructured.Store {
	p.ProfileStoreInvocations++
	if p.ProfileStoreStub != nil {
		return p.ProfileStoreStub()
	}
	if len(p.ProfileStoreOutputs) > 0 {
		output := p.ProfileStoreOutputs[0]
		p.ProfileStoreOutputs = p.ProfileStoreOutputs[1:]
		return output
	}
	if p.ProfileStoreOutput != nil {
		return *p.ProfileStoreOutput
	}
	panic("ProfileStore has no output")
}

func (p *Provider) SessionStore() sessionStore.Store {
	p.SessionStoreInvocations++
	if p.SessionStoreStub != nil {
		return p.SessionStoreStub()
	}
	if len(p.SessionStoreOutputs) > 0 {
		output := p.SessionStoreOutputs[0]
		p.SessionStoreOutputs = p.SessionStoreOutputs[1:]
		return output
	}
	if p.SessionStoreOutput != nil {
		return *p.SessionStoreOutput
	}
	panic("SessionStore has no output")
}

func (p *Provider) UserStructuredStore() userStoreStructured.Store {
	p.UserStructuredStoreInvocations++
	if p.UserStructuredStoreStub != nil {
		return p.UserStructuredStoreStub()
	}
	if len(p.UserStructuredStoreOutputs) > 0 {
		output := p.UserStructuredStoreOutputs[0]
		p.UserStructuredStoreOutputs = p.UserStructuredStoreOutputs[1:]
		return output
	}
	if p.UserStructuredStoreOutput != nil {
		return *p.UserStructuredStoreOutput
	}
	panic("UserStructuredStore has no output")
}

func (p *Provider) PasswordHasher() userServiceClient.PasswordHasher {
	p.PasswordHasherInvocations++
	if p.PasswordHasherStub != nil {
		return p.PasswordHasherStub()
	}
	if len(p.PasswordHasherOutputs) > 0 {
		output := p.PasswordHasherOutputs[0]
		p.PasswordHasherOutputs = p.PasswordHasherOutputs[1:]
		return output
	}
	if p.PasswordHasherOutput != nil {
		return *p.PasswordHasherOutput
	}
	panic("PasswordHasher has no output")
}

func (p *Provider) AssertOutputsEmpty() {
	if len(p.AuthClientOutputs) > 0 {
		panic("AuthClientOutputs is not empty")
	}
	if len(p.BlobClientOutputs) > 0 {
		panic("BlobClientOutputs is not empty")
	}
	if len(p.DataClientOutputs) > 0 {
		panic("DataClientOutputs is not empty")
	}
	if len(p.DataSourceClientOutputs) > 0 {
		panic("DataSourceClientOutputs is not empty")
	}
	if len(p.ImageClientOutputs) > 0 {
		panic("ImageClientOutputs is not empty")
	}
	if len(p.PermissionClientOutputs) > 0 {
		panic("PermissionClientOutputs is not empty")
	}
	if len(p.ConfirmationStoreOutputs) > 0 {
		panic("ConfirmationStoreOutputs is not empty")
	}
	if len(p.MessageStoreOutputs) > 0 {
		panic("MessageStoreOutputs is not empty")
	}
	if len(p.PermissionStoreOutputs) > 0 {
		panic("PermissionStoreOutputs is not empty")
	}
	if len(p.ProfileStoreOutputs) > 0 {
		panic("ProfileStoreOutputs is not empty")
	}
	if len(p.SessionStoreOutputs) > 0 {
		panic("SessionStoreOutputs is not empty")
	}
	if len(p.UserStructuredStoreOutputs) > 0 {
		panic("UserStructuredStoreOutputs is not empty")
	}
	if len(p.PasswordHasherOutputs) > 0 {
		panic("PasswordHasherOutputs is not empty")
	}
}
