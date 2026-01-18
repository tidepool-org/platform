package work

import (
	"context"

	"github.com/tidepool-org/platform/auth"
	providerSession "github.com/tidepool-org/platform/auth/providersession"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

const MetadataKeyID = "providerSessionId"

type Mixin struct {
	*workBase.Processor
	Client          providerSession.Client
	ProviderSession *auth.ProviderSession
}

func NewMixin(processor *workBase.Processor, client providerSession.Client) (*Mixin, error) {
	if processor == nil {
		return nil, errors.New("processor is missing")
	}
	if client == nil {
		return nil, errors.New("client is missing")
	}
	return &Mixin{
		Processor: processor,
		Client:    client,
	}, nil
}

func (m *Mixin) ProviderSessionIDFromMetadata() (*string, error) {
	parser := m.MetadataParser()
	providerSessionID := parser.String(MetadataKeyID)
	if err := parser.Error(); err != nil {
		return nil, errors.Wrap(err, "unable to parse provider session id from metadata")
	}
	return providerSessionID, nil
}

func (m *Mixin) FetchProviderSessionFromMetadata() *work.ProcessResult {
	providerSessionID, err := m.ProviderSessionIDFromMetadata()
	if err != nil || providerSessionID == nil {
		return m.Failed(errors.Wrap(err, "unable to get provider session id from metadata"))
	}
	return m.FetchProviderSession(*providerSessionID)
}

func (m *Mixin) FetchProviderSession(providerSessionID string) *work.ProcessResult {
	providerSession, err := m.Client.GetProviderSession(m.Context(), providerSessionID)
	if err != nil {
		return m.Failing(errors.Wrap(err, "unable to fetch provider session"))
	} else if providerSession == nil {
		return m.Failed(errors.New("provider session is missing"))
	}
	m.ProviderSession = providerSession

	m.AddFieldToContext("providerSession", log.Fields{"id": m.ProviderSession.ID, "userId": m.ProviderSession.UserID})

	return nil
}

func (m *Mixin) UpdateProviderSession(providerSessionUpdate auth.ProviderSessionUpdate) *work.ProcessResult {
	if m.ProviderSession == nil {
		return m.Failed(errors.New("provider session is missing"))
	}

	providerSession, err := m.Client.UpdateProviderSession(context.WithoutCancel(m.Context()), m.ProviderSession.ID, &providerSessionUpdate)
	if err != nil {
		return m.Failing(errors.Wrap(err, "unable to update provider session"))
	} else if providerSession == nil {
		return m.Failed(errors.New("provider session is missing"))
	}
	m.ProviderSession = providerSession

	return nil
}
