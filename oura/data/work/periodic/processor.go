package periodic

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
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/times"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

const (
	PendingAvailableDuration   = 12 * time.Hour // Data returned is for previous 24 hours, using 12 hours ensures we do not miss data
	FailingRetryDuration       = 1 * time.Minute
	FailingRetryDurationJitter = 5 * time.Second
)

func DataTypes() []string {
	return []string{
		oura.DataTypeHeartRate,
		oura.DataTypeRingBatteryLevel,
	}
}

const (
	MetadataKeyDataTypeStartTimes = "dataTypeStartTimes"
	MetadataKeyDataTypeNextTokens = "dataTypeNextTokens"
)

type ProviderSessionMetadata = providerSessionWork.Metadata

type Metadata struct {
	ProviderSessionMetadata `bson:",inline"`
	DataTypeStartTimes      *dataWork.StringTimeMap   `json:"dataTypeStartTimes,omitempty" bson:"dataTypeStartTimes,omitempty"`
	DataTypeNextTokens      *dataWork.StringStringMap `json:"dataTypeNextTokens,omitempty" bson:"dataTypeNextTokens,omitempty"`
}

func (m *Metadata) Parse(parser structure.ObjectParser) {
	m.ProviderSessionMetadata.Parse(parser)
	m.DataTypeStartTimes = dataWork.ParseStringTimeMap(parser.WithReferenceObjectParser(MetadataKeyDataTypeStartTimes))
	m.DataTypeNextTokens = dataWork.ParseStringStringMap(parser.WithReferenceObjectParser(MetadataKeyDataTypeNextTokens))
}

func (m *Metadata) Validate(validator structure.Validator) {
	if m.ProviderSessionID != nil {
		m.ProviderSessionMetadata.Validate(validator)
	} else {
		validator.WithReference(providerSessionWork.MetadataKeyProviderSessionID).ReportError(structureValidator.ErrorValueNotExists())
	}
	if dataTypeStartTimesValidator := validator.WithReference(MetadataKeyDataTypeStartTimes); m.DataTypeStartTimes != nil {
		m.DataTypeStartTimes.Validate(dataTypeStartTimesValidator)
		for _, reference := range m.DataTypeStartTimes.SortedKeys() {
			dataTypeStartTimesValidator.WithReference(reference).String(structure.ReferenceSelf, &reference).OneOf(DataTypes()...)
			dataTypeStartTimesValidator.Time(reference, (*m.DataTypeStartTimes)[reference]).NotZero()
		}
	}
	if dataTypeNextTokensValidator := validator.WithReference(MetadataKeyDataTypeNextTokens); m.DataTypeNextTokens != nil {
		m.DataTypeNextTokens.Validate(dataTypeNextTokensValidator)
		for _, reference := range m.DataTypeNextTokens.SortedKeys() {
			dataTypeNextTokensValidator.WithReference(reference).String(structure.ReferenceSelf, &reference).OneOf(DataTypes()...)
			dataTypeNextTokensValidator.String(reference, (*m.DataTypeNextTokens)[reference]).NotEmpty()
		}
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
	).Process(p.Pending)
}

func (p *Processor) fetchData() *work.ProcessResult {
	// If metadata not specified, then create
	dataTypeStartTimes := p.Metadata().DataTypeStartTimes
	if dataTypeStartTimes == nil {
		dataTypeStartTimes = &dataWork.StringTimeMap{}
	}
	dataTypeNextTokens := p.Metadata().DataTypeNextTokens
	if dataTypeNextTokens == nil {
		dataTypeNextTokens = &dataWork.StringStringMap{}
	}

	// Fetch all data for available data types
	for _, dataType := range DataTypes() {
		ctx, lgr := log.ContextAndLoggerWithField(p.Context(), "dataType", dataType)

		// If data type scope is not authorized, then skip
		if !oura.DataTypeInScopes(dataType, p.ProviderSession().OAuthToken.Scope) {
			lgr.Debug("skipping data type not authorized by scope")
			continue
		}

		// Initial start time
		startTime := (*dataTypeStartTimes)[dataType]
		timeRange := times.TimeRange{
			From: startTime,
		}

		// Use next token for pagination
		pagination := oura.Pagination{
			NextToken: (*dataTypeNextTokens)[dataType],
		}

		for {
			// Fetch page of data for data type
			dataResponse, err := p.GetData(ctx, dataType, &timeRange, &pagination, p.TokenSource())
			if err != nil {
				return p.Failing(errors.Wrapf(err, "unable to get data for data type %q", dataType))
			} else if dataResponse == nil {
				return p.Failing(errors.Newf("data response for data type %q is missing", dataType))
			}

			// Adjust start time, if necessary
			if timeMaximum := dataResponse.Data.TimeMaximum(); timeMaximum != nil && (startTime == nil || timeMaximum.After(*startTime)) {
				startTime = timeMaximum
			}

			// Persist data
			if len(dataResponse.Data) > 0 {
				if result := p.createDataRaw(dataType, &timeRange, dataResponse.Data); result != nil {
					return result
				}
			}

			// If done, then break
			if pagination = dataResponse.Pagination; !pagination.HasNext() {
				break
			}

			// Update next token
			(*dataTypeNextTokens)[dataType] = dataResponse.NextToken

			// Update work metadata with start times and next tokens
			p.Metadata().DataTypeStartTimes = dataTypeStartTimes
			p.Metadata().DataTypeNextTokens = dataTypeNextTokens
			if result := p.ProcessingUpdate(); result != nil {
				return result
			}
		}

		// Update start time and next token
		if startTime != nil {
			(*dataTypeStartTimes)[dataType] = startTime
		}
		delete(*dataTypeNextTokens, dataType)

		// Update work metadata with start times and next tokens
		p.Metadata().DataTypeStartTimes = dataTypeStartTimes
		p.Metadata().DataTypeNextTokens = dataTypeNextTokens
		if result := p.ProcessingUpdate(); result != nil {
			return result
		}
	}

	p.Metadata().DataTypeNextTokens = dataTypeNextTokens
	return nil
}

func (p *Processor) createDataRaw(dataType string, timeRange *times.TimeRange, data oura.Data) *work.ProcessResult {
	if dataRawCreate, err := metadata.WithMetadata(
		&dataRaw.Create{
			MediaType:      pointer.From(net.MediaTypeJSON),
			ArchivableTime: pointer.From(p.Now()),
		},
		&ouraData.Metadata{
			DataType: dataType,
			TimeRangeMetadata: times.TimeRangeMetadata{
				TimeRange: timeRange,
			},
		},
	); err != nil {
		return p.Failing(errors.Wrap(err, "unable to encode data raw metadata"))
	} else {
		p.ClearDataRaw()
		return p.CreateDataRawForDataSource(dataRawCreate, compress.JSONEncoderReader(&oura.DataMap{dataType: data})) // Store as map for later processing
	}
}
