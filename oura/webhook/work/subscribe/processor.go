package subscribe

import (
	"context"
	"slices"
	"time"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/oura"
	ouraWebhook "github.com/tidepool-org/platform/oura/webhook"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

const (
	PendingAvailableDuration      = 24 * time.Hour
	FailingRetryDuration          = 1 * time.Minute
	FailingRetryDurationJitter    = 10 * time.Second
	ExpirationTimeDurationMinimum = 7 * 24 * time.Hour // Normal expiration seems like 90 days, but this will refresh every 7

	OverrideDisabled = "disabled" // Delete all existing subscriptions; data loss while subscriptions disabled, use with caution
	OverrideRenew    = "renew"    // DEBUG: Renew (expiration time) all existing subscriptions, create any missing subscriptions; if error, then may not renew all subscriptions
	OverrideReset    = "reset"    // Delete all existing subscriptions, then create all required subscriptions; potential data loss between delete and create; use with caution
	OverrideUpdate   = "update"   // DEBUG: Update (callback URL) all existing subscriptions, create any missing subscriptions; if error, then may not update all subscriptions
)

func Overrides() []string {
	return []string{
		OverrideDisabled,
		OverrideRenew,
		OverrideReset,
		OverrideUpdate,
	}
}

const MetadataKeyOverride = "override"

type Metadata struct {
	Override *string `json:"override,omitempty" bson:"override,omitempty"`
}

func (m *Metadata) Parse(parser structure.ObjectParser) {
	m.Override = parser.String(MetadataKeyOverride)
}

func (m *Metadata) Validate(validator structure.Validator) {
	validator.String(MetadataKeyOverride, m.Override).OneOf(Overrides()...)
}

type Processor struct {
	*workBase.Processor[Metadata]
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

	processor, err := workBase.NewProcessor[Metadata](dependencies.Dependencies, processResultBuilder)
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
		p.synchronizeSubscriptions,
	).Process(p.Pending)
}

func (p *Processor) synchronizeSubscriptions() *work.ProcessResult {
	existingSubscriptions, err := p.ListSubscriptions(p.Context())
	if err != nil {
		return p.Failing(errors.Wrap(err, "unable to list existing subscriptions"))
	}

	// Process any override
	override := pointer.Default(p.Metadata().Override, "")
	switch override {
	case OverrideDisabled:
		if result := p.deleteSubscriptions(existingSubscriptions); result != nil {
			return result
		}
		return nil // Done, leave override as-is
	case OverrideReset:
		if result := p.deleteSubscriptions(existingSubscriptions); result != nil {
			return result
		}
		existingSubscriptions = oura.Subscriptions{}
		p.Metadata().Override = nil // Clear early after deleting
	}

	callbackURL := p.PartnerURL() + ouraWebhook.EventPath
	verificationToken := p.PartnerSecret()

	for _, dataType := range ouraWebhook.DataTypes() {
		for _, eventType := range oura.EventTypes() {
			// If an existing subscription exists for this event type and data type, then ensure correct, otherwise create one
			if existingSubscription := existingSubscriptions.Get(dataType, eventType); existingSubscription != nil && existingSubscription.ID != nil {
				subscription := existingSubscription
				subscriptionID := *subscription.ID

				// If the callback url is missing or incorrect, then update
				if (override == OverrideUpdate) || subscription.CallbackURL == nil || *subscription.CallbackURL != callbackURL {
					updateSubscription := &oura.UpdateSubscription{
						CallbackURL:       pointer.From(callbackURL),
						VerificationToken: pointer.From(verificationToken),
						DataType:          subscription.DataType,
						EventType:         subscription.EventType,
					}
					if subscription, err = p.UpdateSubscription(p.Context(), subscriptionID, updateSubscription); err != nil {
						return p.Failing(errors.Wrapf(err, "unable to update existing subscription with id %q, data type %q, and event type %q", subscriptionID, dataType, eventType))
					} else if subscription == nil {
						return p.Failed(errors.Newf("updated subscription is missing with id %q, data type %q, and event type %q", subscriptionID, dataType, eventType))
					} else {
						log.LoggerFromContext(p.Context()).WithField("subscription", subscription).Info("updated subscription")
					}
				}

				// If the subscription is nearing expiration, then renew
				if subscription.ExpirationTime != nil {
					if expirationTime, err := time.ParseInLocation(oura.SubscriptionExpirationTimeFormat, *subscription.ExpirationTime, time.UTC); err != nil {
						return p.Failing(errors.Wrapf(err, "unable to parse expiration time of existing subscription with id %q, data type %q, and event type %q", subscriptionID, dataType, eventType))
					} else if (override == OverrideRenew) || time.Until(expirationTime) < ExpirationTimeDurationMinimum {
						if subscription, err = p.RenewSubscription(p.Context(), subscriptionID); err != nil {
							return p.Failing(errors.Wrapf(err, "unable to renew existing subscription with id %q, data type %q, and event type %q", subscriptionID, dataType, eventType))
						} else if subscription == nil {
							return p.Failed(errors.Newf("renewed subscription is missing with id %q, data type %q, and event type %q", subscriptionID, dataType, eventType))
						} else {
							log.LoggerFromContext(p.Context()).WithField("subscription", subscription).Info("renewed subscription")
						}
					}
				}

				// Remove from list (remaining subscriptions will be deleted below)
				existingSubscriptions = slices.DeleteFunc(existingSubscriptions, func(s *oura.Subscription) bool { return s == existingSubscription })
			} else {

				// Create subscription
				createSubscription := &oura.CreateSubscription{
					CallbackURL:       pointer.From(callbackURL),
					VerificationToken: pointer.From(verificationToken),
					DataType:          pointer.From(dataType),
					EventType:         pointer.From(eventType),
				}
				if subscription, err := p.CreateSubscription(p.Context(), createSubscription); err != nil {
					return p.Failing(errors.Wrapf(err, "unable to create new subscription with data type %q and event type %q", dataType, eventType))
				} else if subscription == nil {
					return p.Failed(errors.Newf("created subscription is missing with data type %q and event type %q", dataType, eventType))
				} else {
					log.LoggerFromContext(p.Context()).WithField("subscription", subscription).Info("created subscription")
				}
			}
		}
	}

	// Delete any remaining (undesired) subscriptions
	if result := p.deleteSubscriptions(existingSubscriptions); result != nil {
		return result
	}

	// Clear any override
	p.Metadata().Override = nil

	return nil
}

func (p *Processor) deleteSubscriptions(subscriptions oura.Subscriptions) *work.ProcessResult {
	for _, subscription := range subscriptions {
		if subscription.ID != nil {
			if err := p.DeleteSubscription(p.Context(), *subscription.ID); err != nil {
				return p.Failing(errors.Wrapf(err, "unable to delete existing subscription with id %q", *subscription.ID))
			} else {
				log.LoggerFromContext(p.Context()).WithField("subscription", subscription).Info("deleted subscription")
			}
		}
	}
	return nil
}
