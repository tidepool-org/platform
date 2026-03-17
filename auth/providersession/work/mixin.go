package work

import (
	"context"

	"github.com/tidepool-org/platform/auth"
	providerSession "github.com/tidepool-org/platform/auth/providersession"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/work"
)

const MetadataKeyProviderSessionID = "providerSessionId"

type Metadata struct {
	ProviderSessionID *string `json:"providerSessionId,omitempty" bson:"providerSessionId,omitempty"`
}

func (m *Metadata) Parse(parser structure.ObjectParser) {
	m.ProviderSessionID = parser.String(MetadataKeyProviderSessionID)
}

func (m *Metadata) Validate(validator structure.Validator) {
	validator.String(MetadataKeyProviderSessionID, m.ProviderSessionID).Using(auth.ProviderSessionIDValidator)
}

//go:generate mockgen -source=mixin.go -destination=test/mixin_mocks.go -package=test Mixin
type Mixin interface {
	ProviderSessionClient() providerSession.Client

	HasProviderSession() bool
	ProviderSession() *auth.ProviderSession
	SetProviderSession(providerSession *auth.ProviderSession) *work.ProcessResult

	FetchProviderSession(providerSessionID string) *work.ProcessResult
	UpdateProviderSession(providerSessionUpdate *auth.ProviderSessionUpdate) *work.ProcessResult

	AddProviderSessionToContext()
}

//go:generate mockgen -source=mixin.go -destination=test/mixin_mocks.go -package=test MixinFromWork
type MixinFromWork interface {
	Mixin

	HasWorkMetadata() bool

	FetchProviderSessionFromWorkMetadata() *work.ProcessResult
	UpdateWorkMetadataFromProviderSession() *work.ProcessResult
}

func NewMixin(provider work.Provider, providerSessionClient providerSession.Client) (Mixin, error) {
	if provider == nil {
		return nil, errors.New("provider is missing")
	}
	if providerSessionClient == nil {
		return nil, errors.New("provider session client is missing")
	}
	return &mixin{
		Provider:              provider,
		providerSessionClient: providerSessionClient,
	}, nil
}

func NewMixinFromWork(provider work.Provider, providerSessionClient providerSession.Client, workMetadata *Metadata) (MixinFromWork, error) {
	if provider == nil {
		return nil, errors.New("provider is missing")
	}
	if providerSessionClient == nil {
		return nil, errors.New("provider session client is missing")
	}
	if workMetadata == nil {
		return nil, errors.New("work metadata is missing")
	}
	return &mixin{
		Provider:              provider,
		providerSessionClient: providerSessionClient,
		workMetadata:          workMetadata,
	}, nil
}

type mixin struct {
	work.Provider
	providerSessionClient providerSession.Client
	providerSession       *auth.ProviderSession
	workMetadata          *Metadata
}

func (m *mixin) ProviderSessionClient() providerSession.Client {
	return m.providerSessionClient
}

func (m *mixin) HasProviderSession() bool {
	return m.providerSession != nil
}

func (m *mixin) ProviderSession() *auth.ProviderSession {
	return m.providerSession
}

func (m *mixin) SetProviderSession(providerSession *auth.ProviderSession) *work.ProcessResult {
	m.providerSession = providerSession
	m.AddProviderSessionToContext()
	return nil
}

func (m *mixin) FetchProviderSession(providerSessionID string) *work.ProcessResult {
	if providerSession, err := m.providerSessionClient.GetProviderSession(m.Context(), providerSessionID); err != nil {
		return m.Failing(errors.Wrap(err, "unable to get provider session"))
	} else if providerSession == nil {
		return m.Failed(errors.New("provider session is missing"))
	} else {
		return m.SetProviderSession(providerSession)
	}
}

func (m *mixin) UpdateProviderSession(providerSessionUpdate *auth.ProviderSessionUpdate) *work.ProcessResult {
	if providerSessionUpdate == nil {
		return m.Failed(errors.New("provider session update is missing"))
	}
	if m.providerSession == nil {
		return m.Failed(errors.New("provider session is missing"))
	}

	if providerSession, err := m.providerSessionClient.UpdateProviderSession(context.WithoutCancel(m.Context()), m.providerSession.ID, providerSessionUpdate); err != nil {
		return m.Failing(errors.Wrap(err, "unable to update provider session"))
	} else if providerSession == nil {
		return m.Failed(errors.New("provider session is missing"))
	} else {
		return m.SetProviderSession(providerSession)
	}
}

func (m *mixin) HasWorkMetadata() bool {
	return m.workMetadata != nil
}

func (m *mixin) FetchProviderSessionFromWorkMetadata() *work.ProcessResult {
	if m.workMetadata == nil {
		return m.Failed(errors.New("work metadata is missing"))
	} else if m.workMetadata.ProviderSessionID == nil {
		return m.Failed(errors.New("work metadata provider session id is missing"))
	} else {
		return m.FetchProviderSession(*m.workMetadata.ProviderSessionID)
	}
}

func (m *mixin) UpdateWorkMetadataFromProviderSession() *work.ProcessResult {
	if m.providerSession == nil {
		return m.Failed(errors.New("provider session is missing"))
	} else if m.workMetadata == nil {
		return m.Failed(errors.New("work metadata is missing"))
	}
	m.workMetadata.ProviderSessionID = &m.providerSession.ID
	return nil
}

func (m *mixin) AddProviderSessionToContext() {
	m.AddFieldToContext("providerSession", providerSessionToFields(m.providerSession))
}

func providerSessionToFields(providerSession *auth.ProviderSession) log.Fields {
	if providerSession == nil {
		return nil
	}
	return log.Fields{
		"id":         providerSession.ID,
		"userId":     providerSession.UserID,
		"type":       providerSession.Type,
		"name":       providerSession.Name,
		"externalId": providerSession.ExternalID,
	}
}
