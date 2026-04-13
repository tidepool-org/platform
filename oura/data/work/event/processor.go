package event

import (
	"context"
	"slices"
	"time"

	providerSessionWork "github.com/tidepool-org/platform/auth/providersession/work"
	"github.com/tidepool-org/platform/compress"
	dataRaw "github.com/tidepool-org/platform/data/raw"
	dataRawWork "github.com/tidepool-org/platform/data/raw/work"
	dataSourceWork "github.com/tidepool-org/platform/data/source/work"
	dataWork "github.com/tidepool-org/platform/data/work"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/metadata"
	"github.com/tidepool-org/platform/net"
	oauthWork "github.com/tidepool-org/platform/oauth/work"
	"github.com/tidepool-org/platform/oura"
	ouraDataWork "github.com/tidepool-org/platform/oura/data/work"
	ouraWebhook "github.com/tidepool-org/platform/oura/webhook"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

const (
	FailingRetryDuration       = time.Minute
	FailingRetryDurationJitter = 5 * time.Second
)

type (
	ProviderSessionMetadata = providerSessionWork.Metadata
	EventMetadata           = ouraWebhook.EventMetadata
)

type Metadata struct {
	ProviderSessionMetadata `bson:",inline"`
	EventMetadata           `bson:",inline"`
}

func (m *Metadata) Parse(parser structure.ObjectParser) {
	m.ProviderSessionMetadata.Parse(parser)
	m.EventMetadata.Parse(parser)
}

func (m *Metadata) Validate(validator structure.Validator) {
	m.ProviderSessionMetadata.Validate(validator)
	m.EventMetadata.Validate(validator)
}

type (
	ProviderSessionMixin           = providerSessionWork.MixinFromWork
	DataSourceMixin                = dataSourceWork.Mixin
	ProviderSessionDataSourceMixin = dataWork.ProviderSessionDataSourceMixin
	OAuthMixin                     = oauthWork.Mixin
	DataRawMixin                   = dataRawWork.MixinWithParsedMetadata[ouraDataWork.Metadata]
	DataSourceDataRawMixin         = dataWork.DataSourceDataRawMixin
)

type Processor struct {
	*workBase.Processor[Metadata]
	ProviderSessionMixin
	DataSourceMixin
	ProviderSessionDataSourceMixin
	OAuthMixin
	DataRawMixin
	DataSourceDataRawMixin
	OuraClient
	data map[string]any
	Now  func() time.Time
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

	processor, err := workBase.NewProcessor[Metadata](dependencies.Dependencies, processResultBuilder)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create processor")
	}
	providerSessionMixin, err := providerSessionWork.NewMixinFromWork(processor, dependencies.ProviderSessionClient, &processor.Metadata().ProviderSessionMetadata)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create provider session mixin")
	}
	dataSourceMixin, err := dataSourceWork.NewMixin(processor, dependencies.DataSourceClient)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create data source mixin")
	}
	providerSessionDataSourceMixin, err := dataWork.NewProviderSessionDataSourceMixin(processor, providerSessionMixin, dataSourceMixin)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create provider session data source mixin")
	}
	oauthMixin, err := oauthWork.NewMixin(processor, providerSessionMixin)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create oauth mixin")
	}
	dataRawMixin, err := dataRawWork.NewMixinWithParsedMetadata[ouraDataWork.Metadata](processor, dependencies.DataRawClient)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create data raw mixin")
	}
	dataSourceDataRawMixin, err := dataWork.NewDataSourceDataRawMixin(processor, dataSourceMixin, dataRawMixin)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create data source data raw mixin")
	}

	return &Processor{
		Processor:                      processor,
		ProviderSessionMixin:           providerSessionMixin,
		DataSourceMixin:                dataSourceMixin,
		ProviderSessionDataSourceMixin: providerSessionDataSourceMixin,
		OAuthMixin:                     oauthMixin,
		DataRawMixin:                   dataRawMixin,
		DataSourceDataRawMixin:         dataSourceDataRawMixin,
		OuraClient:                     dependencies.OuraClient,
		Now:                            time.Now,
	}, nil
}

func (p *Processor) Process(ctx context.Context, wrk *work.Work, processingUpdater work.ProcessingUpdater) *work.ProcessResult {
	return append(p.ProcessPipeline(ctx, wrk, processingUpdater),
		p.FetchProviderSessionFromWorkMetadata,
		p.FetchDataSourceFromProviderSession,
		p.FetchTokenSource,
		p.fetchEventData,
		p.createDataRaw,
	).Process(p.Delete)
}

func (p *Processor) fetchEventData() *work.ProcessResult {
	if p.DataSource().DataSetID == nil {
		return p.Failed(errors.New("data set id is missing"))
	}

	// Determine authorized scope
	scope := pointer.DefaultStringArray(p.ProviderSession().OAuthToken.Scope, nil)

	// Determine if data type is in scope, if not, then warn and done
	event := p.Metadata().Event
	if !slices.Contains(oura.DataTypesForScopes(scope), *event.DataType) {
		log.LoggerFromContext(p.Context()).WithFields(log.Fields{"event": event, "scope": scope}).Warn("event for data type not in scope")
		return p.Delete()
	}

	// If event type is create or update, then get latest datum
	var data map[string]any
	switch *p.Metadata().Event.EventType {
	case oura.EventTypeCreate, oura.EventTypeUpdate:
		if datum, err := p.GetDatum(p.Context(), *event.DataType, *event.ObjectID, p.TokenSource()); err != nil {
			return p.Failing(errors.Wrapf(err, "unable to get data for data with type %q and object id %q", *event.DataType, *event.ObjectID))
		} else {
			data = datum
		}
	case oura.EventTypeDelete:
		data = map[string]any{}
	}

	p.data = data
	return nil
}

func (p *Processor) createDataRaw() *work.ProcessResult {
	if dataRawCreate, err := metadata.WithMetadata(
		&dataRaw.Create{
			MediaType:      pointer.From(net.MediaTypeJSON),
			ArchivableTime: pointer.From(p.Now().UTC()),
		},
		&ouraDataWork.Metadata{
			Scope: p.ProviderSession().OAuthToken.Scope,
			EventMetadata: ouraDataWork.EventMetadata{
				Event: p.Metadata().Event,
			},
		},
	); err != nil {
		return p.Failing(errors.Wrap(err, "unable to encode data raw metadata"))
	} else {
		return p.CreateDataRawForDataSource(dataRawCreate, compress.JSONEncoderReader(p.data))
	}
}
