package revoke

import (
	"context"
	"time"

	"github.com/tidepool-org/platform/errors"
	oauthWork "github.com/tidepool-org/platform/oauth/work"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

const (
	FailingRetryDuration       = time.Minute
	FailingRetryDurationJitter = 5 * time.Second
)

type Processor struct {
	*workBase.Processor[oauthWork.TokenMetadata]
	OuraClient
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

	processor, err := workBase.NewProcessor[oauthWork.TokenMetadata](dependencies.Dependencies, processResultBuilder)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create processor")
	}

	return &Processor{
		Processor:  processor,
		OuraClient: dependencies.OuraClient,
	}, nil
}

func (p *Processor) Process(ctx context.Context, wrk *work.Work, processingUpdater work.ProcessingUpdater) *work.ProcessResult {
	return append(p.ProcessPipeline(ctx, wrk, processingUpdater),
		p.revokeOAuthToken,
	).Process(p.Delete)
}

func (p *Processor) revokeOAuthToken() *work.ProcessResult {
	if err := p.RevokeOAuthToken(p.Context(), p.Metadata().OAuthToken); err != nil {
		return p.Failing(errors.Wrap(err, "unable to revoke oauth token"))
	} else {
		return nil
	}
}
