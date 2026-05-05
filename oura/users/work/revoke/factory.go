package revoke

import (
	"time"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/metadata"
	"github.com/tidepool-org/platform/oura"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

const (
	Type              = "org.tidepool.oura.users.revoke"
	Quantity          = 1
	Frequency         = 5 * time.Second
	ProcessingTimeout = 3 * time.Minute
)

type OuraClient = oura.Client

type Dependencies struct {
	workBase.Dependencies
	OuraClient
}

func (d Dependencies) Validate() error {
	if err := d.Dependencies.Validate(); err != nil {
		return err
	}
	if d.OuraClient == nil {
		return errors.New("oura client is missing")
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

func NewWorkCreate(providerSessionID string, oauthToken *auth.OAuthToken) (*work.Create, error) {
	if providerSessionID == "" {
		return nil, errors.New("provider session id is missing")
	}
	if oauthToken == nil {
		return nil, errors.New("oauth token is missing")
	}

	return metadata.WithMetadata(
		&work.Create{
			Type:              Type,
			DeduplicationID:   pointer.From(providerSessionID),
			ProcessingTimeout: int(ProcessingTimeout.Seconds()),
		},
		&Metadata{
			OAuthToken: oauthToken,
		},
	)
}
