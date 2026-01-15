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

//go:generate mockgen -source=processor.go -destination=test/processor_mocks.go -package=test Client
type Client interface {
	GetProviderSession(ctx context.Context, id string) (*auth.ProviderSession, error)
	UpdateProviderSession(ctx context.Context, id string, update *auth.ProviderSessionUpdate) (*auth.ProviderSession, error)
	DeleteProviderSession(ctx context.Context, id string) error
}

type Processor struct {
	*workBase.Processor
	Client          Client
	ProviderSession *auth.ProviderSession
}

func NewProcessor(processor *workBase.Processor, client Client) (*Processor, error) {
	if processor == nil {
		return nil, errors.New("processor is missing")
	}
	if client == nil {
		return nil, errors.New("client is missing")
	}
	return &Processor{
		Processor: processor,
		Client:    client,
	}, nil
}

func (p *Processor) ProviderSessionIDFromMetadata() (*string, error) {
	parser := p.MetadataParser()
	providerSessionID := parser.String(MetadataKeyID)
	if err := parser.Error(); err != nil {
		return nil, errors.Wrap(err, "unable to parse provider session id from metadata")
	}
	return providerSessionID, nil
}

func (p *Processor) FetchProviderSessionFromMetadata() *work.ProcessResult {
	providerSessionID, err := p.ProviderSessionIDFromMetadata()
	if err != nil || providerSessionID == nil {
		return p.Failed(errors.Wrap(err, "unable to get provider session id from metadata"))
	}
	return p.FetchProviderSession(*providerSessionID)
}

func (p *Processor) FetchProviderSession(providerSessionID string) *work.ProcessResult {
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

func (p *Processor) UpdateProviderSession(providerSessionUpdate auth.ProviderSessionUpdate) *work.ProcessResult {
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
