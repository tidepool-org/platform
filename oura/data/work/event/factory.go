package event

import (
	"fmt"
	"time"

	providerSession "github.com/tidepool-org/platform/auth/providersession"
	"github.com/tidepool-org/platform/crypto"
	dataRaw "github.com/tidepool-org/platform/data/raw"
	dataSource "github.com/tidepool-org/platform/data/source"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/metadata"
	"github.com/tidepool-org/platform/oura"
	ouraDataWork "github.com/tidepool-org/platform/oura/data/work"
	ouraWebhook "github.com/tidepool-org/platform/oura/webhook"
	ouraWork "github.com/tidepool-org/platform/oura/work"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

const (
	Type              = "org.tidepool.oura.data.event"
	Quantity          = 4
	Frequency         = 5 * time.Second
	ProcessingTimeout = 3 * time.Minute
)

type (
	ProviderSessionClient = providerSession.Client
	DataSourceClient      = dataSource.Client
	DataRawClient         = dataRaw.Client
	OuraClient            = oura.Client
)

type Dependencies struct {
	workBase.Dependencies
	ProviderSessionClient
	DataSourceClient
	DataRawClient
	OuraClient
}

func (d Dependencies) Validate() error {
	if err := d.Dependencies.Validate(); err != nil {
		return err
	}
	if d.ProviderSessionClient == nil {
		return errors.New("provider session client is missing")
	}
	if d.DataSourceClient == nil {
		return errors.New("data source client is missing")
	}
	if d.DataRawClient == nil {
		return errors.New("data raw client is missing")
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

func NewWorkCreate(providerSessionID string, event *ouraWebhook.Event) (*work.Create, error) {
	if providerSessionID == "" {
		return nil, errors.New("provider session id is missing")
	}
	if event == nil {
		return nil, errors.New("event is missing")
	}

	return metadata.WithMetadata(
		&work.Create{
			Type:              Type,
			GroupID:           pointer.From(ouraWork.GroupIDFromProviderSessionID(providerSessionID)),
			DeduplicationID:   pointer.From(crypto.HexEncodedSHA256Hash(fmt.Sprintf("%s:%s", providerSessionID, event.String()))),
			SerialID:          pointer.From(ouraDataWork.SerialIDFromProviderSessionID(providerSessionID)),
			ProcessingTimeout: int(ProcessingTimeout.Seconds()),
		},
		&Metadata{
			ProviderSessionMetadata: ProviderSessionMetadata{
				ProviderSessionID: pointer.From(providerSessionID),
			},
			EventMetadata: EventMetadata{
				Event: event,
			},
		},
	)
}
