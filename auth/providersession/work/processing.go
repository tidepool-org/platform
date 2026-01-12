package work

import (
	"context"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

const MetadataKeyID = "providerSessionId"

//go:generate mockgen -source=processing.go -destination=test/processing_mocks.go -package=test Client
type Client interface {
	GetProviderSession(ctx context.Context, id string) (*auth.ProviderSession, error)
	UpdateProviderSession(ctx context.Context, id string, update *auth.ProviderSessionUpdate) (*auth.ProviderSession, error)
	DeleteProviderSession(ctx context.Context, id string) error
}

type Processing struct {
	*workBase.Processing
	Client          Client
	ProviderSession *auth.ProviderSession
}

func NewProcessing(processing *workBase.Processing, client Client) (*Processing, error) {
	if processing == nil {
		return nil, errors.New("processing is missing")
	}
	if client == nil {
		return nil, errors.New("client is missing")
	}
	return &Processing{
		Processing: processing,
		Client:     client,
	}, nil
}

func (p *Processing) ProviderSessionIDFromMetadata() (*string, error) {
	parser := p.MetadataParser()
	providerSessionID := parser.String(MetadataKeyID)
	if err := parser.Error(); err != nil {
		return nil, errors.Wrap(err, "unable to parse provider session id from metadata")
	}
	return providerSessionID, nil
}

func (p *Processing) FetchProviderSessionFromMetadata() *work.ProcessResult {
	providerSessionID, err := p.ProviderSessionIDFromMetadata()
	if err != nil || providerSessionID == nil {
		return p.Failed(errors.Wrap(err, "unable to get provider session id from metadata"))
	}
	return p.FetchProviderSession(*providerSessionID)
}

func (p *Processing) FetchProviderSession(providerSessionID string) *work.ProcessResult {
	providerSession, err := p.Client.GetProviderSession(p.Context(), providerSessionID)
	if err != nil {
		return p.Failing(errors.Wrap(err, "unable to fetch provider session"))
	} else if providerSession == nil {
		return p.Failed(errors.New("provider session is missing"))
	}
	p.ProviderSession = providerSession

	p.ContextWithField("providerSession", log.Fields{"id": p.ProviderSession.ID, "userId": p.ProviderSession.UserID})

	return nil
}

func (p *Processing) UpdateProviderSession(providerSessionUpdate auth.ProviderSessionUpdate) *work.ProcessResult {
	if p.ProviderSession == nil {
		return p.Failed(errors.New("provider session is missing"))
	}

	providerSession, err := p.Client.UpdateProviderSession(context.WithoutCancel(p.Context()), p.ProviderSession.ID, &providerSessionUpdate)
	if err != nil {
		return p.Failing(errors.Wrap(err, "unable to update provider session"))
	} else if providerSession == nil {
		return p.Failed(errors.New("provider session is missing"))
	}
	p.ProviderSession = providerSession

	return nil
}
