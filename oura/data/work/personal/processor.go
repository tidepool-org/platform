package personal

import (
	"context"
	"time"

	providerSessionWork "github.com/tidepool-org/platform/auth/providersession/work"
	"github.com/tidepool-org/platform/compress"
	"github.com/tidepool-org/platform/crypto"
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

const MetadataKeyPreviousHash = "previousHash"

type ProviderSessionMetadata = providerSessionWork.Metadata

type Metadata struct {
	ProviderSessionMetadata `bson:",inline"`
	PreviousHash            *string `json:"previousHash,omitempty" bson:"previousHash,omitempty"`
}

func (m *Metadata) Parse(parser structure.ObjectParser) {
	m.ProviderSessionMetadata.Parse(parser)
	m.PreviousHash = parser.String(MetadataKeyPreviousHash)
}

func (m *Metadata) Validate(validator structure.Validator) {
	if m.ProviderSessionID != nil {
		m.ProviderSessionMetadata.Validate(validator)
	} else {
		validator.WithReference(providerSessionWork.MetadataKeyProviderSessionID).ReportError(structureValidator.ErrorValueNotExists())
	}
	validator.String(MetadataKeyPreviousHash, m.PreviousHash).Using(crypto.Base64EncodedSHA256HashValidator)
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
	// If data type scope is not authorized, then skip
	if !oura.DataTypeInScopes(oura.DataTypePersonalInfo, p.ProviderSession().OAuthToken.Scope) {
		log.LoggerFromContext(p.Context()).Debug("skipping data type not authorized by scope")
		return nil
	}

	// Get personal info
	personalInfo, err := p.GetPersonalInfo(p.Context(), p.TokenSource())
	if err != nil {
		return p.Failing(errors.Wrap(err, "unable to get datum"))
	}

	// If hash matches previous, then skip
	hash, err := personalInfo.Hash()
	if err != nil {
		return p.Failed(errors.Wrap(err, "unable to compute hash of datum"))
	}
	if previousHash := p.Metadata().PreviousHash; previousHash != nil && hash == *previousHash {
		log.LoggerFromContext(p.Context()).Debug("skipping datum with matching hash")
		return nil
	}

	// Convert to datum
	datum, err := metadata.Encode(personalInfo)
	if err != nil {
		return p.Failed(errors.Wrap(err, "unable to encode datum"))
	}

	// Create data raw
	if result := p.createDataRaw(oura.DataTypePersonalInfo, &times.TimeRange{To: pointer.From(p.Now())}, oura.Data{datum}); result != nil {
		return result
	}

	// Update hash
	p.Metadata().PreviousHash = pointer.From(hash)
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
			TimeRangeMetadata: ouraData.TimeRangeMetadata{
				TimeRange: timeRange,
			},
		},
	); err != nil {
		return p.Failed(errors.Wrap(err, "unable to encode data raw metadata"))
	} else {
		p.ClearDataRaw()
		return p.CreateDataRawForDataSource(dataRawCreate, compress.JSONEncoderReader(&oura.DataMap{dataType: data})) // Store as map for later processing
	}
}
