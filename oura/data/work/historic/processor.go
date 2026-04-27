package historic

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
	ouraDataWork "github.com/tidepool-org/platform/oura/data/work"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/times"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

const (
	FailingRetryDuration       = 1 * time.Minute
	FailingRetryDurationJitter = 5 * time.Second
)

var LaunchDate = time.Date(2015, time.March, 1, 0, 0, 0, 0, time.UTC) // 2015-03-01

type (
	ProviderSessionMetadata = providerSessionWork.Metadata
	TimeRangeMetadata       = times.TimeRangeMetadata
)

type Metadata struct {
	ProviderSessionMetadata `bson:",inline"`
	TimeRangeMetadata       `bson:",inline"`
}

func (m *Metadata) Parse(parser structure.ObjectParser) {
	m.ProviderSessionMetadata.Parse(parser)
	m.TimeRangeMetadata.Parse(parser)
}

func (m *Metadata) Validate(validator structure.Validator) {
	m.ProviderSessionMetadata.Validate(validator)
	m.TimeRangeMetadata.Validate(validator)
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
	Now       func() time.Time
	timeRange times.TimeRange
	data      map[string]any
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
		p.prepareTimeRange,
		p.fetchHistoricData,
		p.createDataRaw,
	).Process(p.Delete)
}

func (p *Processor) prepareTimeRange() *work.ProcessResult {
	p.timeRange = NormalizeTimeRange(p.Metadata().TimeRange, LaunchDate, p.Now().UTC())

	log.LoggerFromContext(p.Context()).WithField("timeRange", log.Fields{"initial": p.Metadata().TimeRange, "final": p.timeRange}).Debug("prepared time range")
	return nil
}

func (p *Processor) fetchHistoricData() *work.ProcessResult {
	if p.DataSource().DataSetID == nil {
		return p.Failed(errors.New("data set id is missing"))
	}

	// Determine authorized scope
	scope := pointer.DefaultStringArray(p.ProviderSession().OAuthToken.Scope, nil)

	// Determine available data types, if not, then warn and done
	dataTypes := oura.DataTypesForScopes(scope)
	if len(dataTypes) == 0 {
		log.LoggerFromContext(p.Context()).WithField("scope", scope).Warn("no data types in scope for historic data")
		return p.Delete()
	}

	// Fetch historic data for available data types
	data := map[string]any{}
	for _, dataType := range dataTypes {
		var dataTypeData []any

		// Loop through all "pages" of the data, per documentation, but in practice we have not seen any pagination
		var pagination oura.Pagination
		for {
			if dataResponse, err := p.GetData(p.Context(), dataType, &p.timeRange, &pagination, p.TokenSource()); err != nil {
				return p.Failing(errors.Wrapf(err, "unable to get data for data type %q", dataType))
			} else if dataResponse == nil {
				return p.Failing(errors.Newf("data response for data type %q is missing", dataType))
			} else {
				dataTypeData = append(dataTypeData, dataResponse.Data...)
				pagination = dataResponse.Pagination
			}
			if !pagination.HasNext() {
				break
			}
		}

		// Add to historic data
		data[dataType] = dataTypeData
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
			TimeRangeMetadata: ouraDataWork.TimeRangeMetadata{
				TimeRange: pointer.From(p.timeRange),
			},
		},
	); err != nil {
		return p.Failing(errors.Wrap(err, "unable to encode data raw metadata"))
	} else {
		return p.CreateDataRawForDataSource(dataRawCreate, compress.JSONEncoderReader(p.data))
	}
}

// Limit time range to between original launch date (2015-03-01) and now, UTC, date-only
func NormalizeTimeRange(timeRange *times.TimeRange, minimum time.Time, maximum time.Time) times.TimeRange {
	normalized := pointer.Default(timeRange, times.TimeRange{})
	normalized.From = pointer.DefaultPointer(normalized.From, pointer.From(minimum))
	normalized.To = pointer.DefaultPointer(normalized.To, pointer.From(maximum))
	return normalized.Clamped(minimum, maximum).InLocation(time.UTC).Date()
}
