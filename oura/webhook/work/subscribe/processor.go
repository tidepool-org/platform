package subscribe

import (
	"context"
	"time"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/oura"
	ouraWebhook "github.com/tidepool-org/platform/oura/webhook"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

const (
	PendingAvailableDuration      = 24 * time.Hour
	FailingRetryDuration          = 10 * time.Minute
	FailingRetryDurationJitter    = time.Minute
	ExpirationTimeDurationMinimum = 7 * 24 * time.Hour // Normal expiration seems like 90 days, but this will refresh every 7
)

type Processor struct {
	*workBase.ProcessorWithoutMetadata
	OuraClient
}

func NewProcessor(dependencies Dependencies) (*Processor, error) {
	if err := dependencies.Validate(); err != nil {
		return nil, errors.Wrap(err, "dependencies is invalid")
	}

	processResultBuilder := &workBase.ProcessResultBuilder{
		ProcessResultPendingBuilder: &workBase.ConstantProcessResultPendingBuilder{
			Duration: PendingAvailableDuration,
		},
		ProcessResultFailingBuilder: &workBase.ExponentialProcessResultFailingBuilder{
			Duration:       FailingRetryDuration,
			DurationJitter: FailingRetryDurationJitter,
		},
	}

	processorWithoutMetadata, err := workBase.NewProcessorWithoutMetadata(dependencies.Dependencies, processResultBuilder)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create processor")
	}

	return &Processor{
		ProcessorWithoutMetadata: processorWithoutMetadata,
		OuraClient:               dependencies.OuraClient,
	}, nil
}

func (p *Processor) Process(ctx context.Context, wrk *work.Work, processingUpdater work.ProcessingUpdater) *work.ProcessResult {
	return append(p.ProcessPipeline(ctx, wrk, processingUpdater),
		p.synchronizeSubscriptions,
	).Process(p.Pending)
}

func (p *Processor) synchronizeSubscriptions() *work.ProcessResult {
	existingSubscriptions, err := p.ListSubscriptions(p.Context())
	if err != nil {
		return p.Failing(errors.Wrap(err, "unable to list existing subscriptions"))
	}

	callbackURL := p.PartnerURL() + ouraWebhook.EventPath
	verificationToken := p.PartnerSecret()

	for _, dataType := range ouraWebhook.DataTypes() {
		for _, eventType := range oura.EventTypes() {

			// If an existing subscription exists for this event type and data type, either update it or renew it
			if existingSubscription := existingSubscriptions.Get(dataType, eventType); existingSubscription != nil && existingSubscription.ID != nil {
				if existingSubscription.CallbackURL == nil || *existingSubscription.CallbackURL != callbackURL {
					updateSubscription := &oura.UpdateSubscription{
						CallbackURL:       pointer.FromString(callbackURL),
						VerificationToken: pointer.FromString(verificationToken),
						DataType:          existingSubscription.DataType,
						EventType:         existingSubscription.EventType,
					}
					if _, err := p.UpdateSubscription(p.Context(), *existingSubscription.ID, updateSubscription); err != nil {
						return p.Failing(errors.Wrapf(err, "unable to update existing subscription with id %q, data type %q, and event type %q", *existingSubscription.ID, dataType, eventType))
					}
				} else if existingSubscription.ExpirationTime != nil {
					if expirationTime, err := time.Parse(oura.SubscriptionExpirationTimeFormat, *existingSubscription.ExpirationTime); err != nil {
						return p.Failing(errors.Wrapf(err, "unable to parse expiration time of existing subscription with id %q, data type %q, and event type %q", *existingSubscription.ID, dataType, eventType))
					} else if time.Until(expirationTime) < ExpirationTimeDurationMinimum {
						if _, err := p.RenewSubscription(p.Context(), *existingSubscription.ID); err != nil {
							return p.Failing(errors.Wrapf(err, "unable to renew existing subscription with id %q, data type %q, and event type %q", *existingSubscription.ID, dataType, eventType))
						}
					}
				}
			} else {
				createSubscription := &oura.CreateSubscription{
					CallbackURL:       pointer.FromString(callbackURL),
					VerificationToken: pointer.FromString(verificationToken),
					DataType:          pointer.FromString(dataType),
					EventType:         pointer.FromString(eventType),
				}
				if _, err := p.CreateSubscription(p.Context(), createSubscription); err != nil {
					return p.Failing(errors.Wrapf(err, "unable to create new subscription with data type %q and event type %q", dataType, eventType))
				}
			}
		}
	}

	return nil
}
