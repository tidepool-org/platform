package event

import (
	"context"
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
	ouraData "github.com/tidepool-org/platform/oura/data"
	ouraDataWork "github.com/tidepool-org/platform/oura/data/work"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

const (
	FailingRetryDuration       = 1 * time.Minute
	FailingRetryDurationJitter = 5 * time.Second
)

type (
	ProviderSessionMetadata = providerSessionWork.Metadata
	EventMetadata           = oura.EventMetadata
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
	if m.ProviderSessionID != nil {
		m.ProviderSessionMetadata.Validate(validator)
	} else {
		validator.WithReference(providerSessionWork.MetadataKeyProviderSessionID).ReportError(structureValidator.ErrorValueNotExists())
	}
	if m.Event != nil {
		m.EventMetadata.Validate(validator)
	} else {
		validator.WithReference(oura.MetadataKeyEvent).ReportError(structureValidator.ErrorValueNotExists())
	}
}

type (
	ProviderSessionMixin           = providerSessionWork.MixinFromWork
	DataSourceMixin                = dataSourceWork.Mixin
	ProviderSessionDataSourceMixin = dataWork.ProviderSessionDataSourceMixin
	OAuthMixin                     = oauthWork.Mixin
	DataRawMixin                   = dataRawWork.MixinWithParsedMetadata[ouraData.Metadata]
	DataSourceDataRawMixin         = dataWork.DataSourceDataRawMixin
	OuraClient                     = oura.Client
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
}

func NewProcessor(dependencies ouraDataWork.Dependencies) (*Processor, error) {
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
	dataRawMixin, err := dataRawWork.NewMixinWithParsedMetadata[ouraData.Metadata](processor, dependencies.DataRawClient)
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
	}, nil
}

func (p *Processor) Process(ctx context.Context, wrk *work.Work, processingUpdater work.ProcessingUpdater) *work.ProcessResult {
	return append(p.ProcessPipeline(ctx, wrk, processingUpdater),
		p.FetchProviderSessionFromWorkMetadata,
		p.FetchDataSourceFromProviderSession,
		p.EnsureDataSourceHasDataSetID,
		p.FetchTokenSource,
		p.fetchData,
	).Process(p.Delete)
}

func (p *Processor) fetchData() *work.ProcessResult {
	event := p.Metadata().Event

	// If data type scope is not authorized, then skip
	if !oura.DataTypeInScopes(*event.DataType, p.ProviderSession().OAuthToken.Scope) {
		log.LoggerFromContext(p.Context()).Info("skipping datum with data type not authorized by scope")
		return nil
	}

	// If event type is create or update, then get latest datum
	data := oura.Data{}
	switch *p.Metadata().Event.EventType {
	case oura.EventTypeCreate, oura.EventTypeUpdate:
		if datum, err := p.GetDatum(p.Context(), *event.DataType, *event.ObjectID, p.TokenSource()); err != nil {
			// Fail immediately if not found, likely means data was deleted, but mark as failed for now for future investigation, otherwise retry
			if request.IsErrorResourceNotFound(errors.Cause(err)) {
				return p.Failed(errors.Wrapf(err, "datum with data type %q and object id %q not found", *event.DataType, *event.ObjectID))
			} else {
				return p.Failing(errors.Wrapf(err, "unable to get datum with data type %q and object id %q", *event.DataType, *event.ObjectID))
			}
		} else {
			data = append(data, datum)
		}
	case oura.EventTypeDelete:
	}

	return p.createDataRaw(*event.DataType, event, data)
}

func (p *Processor) createDataRaw(dataType string, event *oura.Event, data oura.Data) *work.ProcessResult {
	if dataRawCreate, err := metadata.WithMetadata(
		&dataRaw.Create{
			MediaType:      pointer.From(net.MediaTypeJSON),
			ArchivableTime: pointer.From(p.Now()),
		},
		&ouraData.Metadata{
			DataType: dataType,
			EventMetadata: oura.EventMetadata{
				Event: event,
			},
		},
	); err != nil {
		return p.Failing(errors.Wrap(err, "unable to encode data raw metadata"))
	} else {
		return p.CreateDataRawForDataSource(dataRawCreate, compress.JSONEncoderReader(&oura.DataMap{dataType: data})) // Store as map for later processing
	}
}
