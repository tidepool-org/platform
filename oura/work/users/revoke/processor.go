package revoke

import (
	"context"
	"time"

	"github.com/tidepool-org/platform/auth"
	providerSessionWork "github.com/tidepool-org/platform/auth/providersession/work"
	"github.com/tidepool-org/platform/errors"
	oauthWork "github.com/tidepool-org/platform/oauth/work"
	ouraWork "github.com/tidepool-org/platform/oura/work"
	ouraWorkUsers "github.com/tidepool-org/platform/oura/work/users"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

const (
	Type      = "org.tidepool.oura.work.users.revoke"
	Quantity  = 1
	Frequency = 5 * time.Second

	FailingRetryDuration       = 1 * time.Minute
	FailingRetryDurationJitter = 5 * time.Second

	ProcessingTimeout = 60 // Seconds
)

type Dependencies struct {
	workBase.Dependencies
	Client ouraWork.Client
}

func (d Dependencies) Validate() error {
	if err := d.Dependencies.Validate(); err != nil {
		return errors.Wrap(err, "dependencies is invalid")
	}
	if d.Client == nil {
		return errors.New("client is missing")
	}
	return nil
}

func NewProcessorFactory(dependencies Dependencies) (*workBase.ProcessorFactory, error) {
	if err := dependencies.Validate(); err != nil {
		return nil, errors.Wrap(err, "dependencies is invalid")
	}
	processorFactory := func() (work.Processor, error) { return NewProcessor(dependencies) }
	return workBase.NewProcessorFactory(Type, Quantity, Frequency, processorFactory)
}

type Processor struct {
	*workBase.Processor
	*oauthWork.OAuthTokenMixin
	Client ouraWork.Client
}

func NewProcessor(dependencies Dependencies) (*Processor, error) {
	if err := dependencies.Validate(); err != nil {
		return nil, errors.Wrap(err, "dependencies is invalid")
	}

	processResultBuilder := &workBase.ProcessResultBuilder{
		ProcessResultFailingBuilder: &workBase.ExponentialProcessResultFailingBuilder{
			Duration:       FailingRetryDuration,
			DurationJitter: FailingRetryDurationJitter,
		},
	}

	processor, err := workBase.NewProcessor(dependencies.Dependencies, processResultBuilder)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create processor")
	}

	oauthTokenMixin, err := oauthWork.NewOAuthTokenMixin(processor)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create oauth token mixin")
	}

	return &Processor{
		Processor:       processor,
		OAuthTokenMixin: oauthTokenMixin,
		Client:          dependencies.Client,
	}, nil
}

func (p *Processor) Process(ctx context.Context, wrk *work.Work, processingUpdater work.ProcessingUpdater) *work.ProcessResult {
	return append(p.ProcessPipeline(ctx, wrk, processingUpdater),
		p.revokeOAuthToken,
	).Process(p.Delete)
}

func (p *Processor) revokeOAuthToken() *work.ProcessResult {
	if oauthToken, err := p.OAuthTokenFromMetadata(); err != nil {
		return p.Failed(errors.Wrap(err, "unable to get oauth token from metadata"))
	} else if err := p.Client.RevokeOAuthToken(p.Context(), oauthToken); err != nil {
		return p.Failing(err)
	} else {
		return nil
	}
}

func NewWorkCreate(providerSession *auth.ProviderSession) (*work.Create, error) {
	if providerSession == nil {
		return nil, errors.New("provider session is missing")
	}
	return &work.Create{
		Type:              Type,
		GroupID:           pointer.FromString(ouraWorkUsers.GroupIDFromProviderSessionID(providerSession.ID)),
		DeduplicationID:   pointer.FromString(providerSession.ID),
		SerialID:          pointer.FromString(ouraWorkUsers.SerialIDFromProviderSessionID(providerSession.ID)),
		ProcessingTimeout: ProcessingTimeout,
		Metadata: map[string]any{
			providerSessionWork.MetadataKeyID: providerSession.ID,
			oauthWork.MetadataKeyOAuthToken:   providerSession.OAuthToken,
		},
	}, nil
}
