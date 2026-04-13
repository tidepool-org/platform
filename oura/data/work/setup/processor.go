package setup

import (
	"context"
	"time"

	"github.com/tidepool-org/platform/auth"
	providerSessionWork "github.com/tidepool-org/platform/auth/providersession/work"
	customerioWork "github.com/tidepool-org/platform/customerio/work/event"
	"github.com/tidepool-org/platform/data"
	dataDeduplicatorDeduplicator "github.com/tidepool-org/platform/data/deduplicator/deduplicator"
	dataSetWork "github.com/tidepool-org/platform/data/set/work"
	dataSource "github.com/tidepool-org/platform/data/source"
	dataSourceWork "github.com/tidepool-org/platform/data/source/work"
	dataWork "github.com/tidepool-org/platform/data/work"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/oauth"
	oauthWork "github.com/tidepool-org/platform/oauth/work"
	"github.com/tidepool-org/platform/oura"
	ouraDataWorkHistoric "github.com/tidepool-org/platform/oura/data/work/historic"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/times"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

const (
	FailingRetryDuration       = time.Minute
	FailingRetryDurationJitter = 5 * time.Second
)

type Metadata = providerSessionWork.Metadata

type (
	ProviderSessionMixin           = providerSessionWork.MixinFromWork
	DataSourceMixin                = dataSourceWork.Mixin
	DataSourceReplacerMixin        = dataWork.DataSourceReplacerMixin
	ProviderSessionDataSourceMixin = dataWork.ProviderSessionDataSourceMixin
	DataSetMixin                   = dataSetWork.Mixin
	DataSourceDataSetMixin         = dataWork.DataSourceDataSetMixin
	OAuthMixin                     = oauthWork.Mixin
)

type Processor struct {
	*workBase.Processor[Metadata]
	ProviderSessionMixin
	DataSourceMixin
	DataSourceReplacerMixin
	ProviderSessionDataSourceMixin
	OAuthMixin
	DataSetMixin
	DataSourceDataSetMixin
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

	processor, err := workBase.NewProcessor[Metadata](dependencies.Dependencies, processResultBuilder)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create processor")
	}
	providerSessionMixin, err := providerSessionWork.NewMixinFromWork(processor, dependencies.ProviderSessionClient, processor.Metadata())
	if err != nil {
		return nil, errors.Wrap(err, "unable to create provider session mixin")
	}
	dataSourceMixin, err := dataSourceWork.NewMixin(processor, dependencies.DataSourceClient)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create data source mixin")
	}
	dataSourceReplacerMixin, err := dataWork.NewDataSourceReplacerMixin(processor, dataSourceMixin, dependencies.ProviderSessionClient)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create data source replacer mixin")
	}
	providerSessionDataSourceMixin, err := dataWork.NewProviderSessionDataSourceMixin(processor, providerSessionMixin, dataSourceMixin)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create provider session data source mixin")
	}
	oauthMixin, err := oauthWork.NewMixin(processor, providerSessionMixin)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create oauth mixin")
	}
	dataSetMixin, err := dataSetWork.NewMixin(processor, dependencies.DataSetClient)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create data set mixin")
	}
	dataSourceDataSetMixin, err := dataWork.NewDataSourceDataSetMixin(processor, dataSourceMixin, dataSetMixin)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create data source data set mixin")
	}

	return &Processor{
		Processor:                      processor,
		ProviderSessionMixin:           providerSessionMixin,
		DataSourceMixin:                dataSourceMixin,
		DataSourceReplacerMixin:        dataSourceReplacerMixin,
		ProviderSessionDataSourceMixin: providerSessionDataSourceMixin,
		OAuthMixin:                     oauthMixin,
		DataSetMixin:                   dataSetMixin,
		DataSourceDataSetMixin:         dataSourceDataSetMixin,
		OuraClient:                     dependencies.OuraClient,
	}, nil
}

func (p *Processor) Process(ctx context.Context, wrk *work.Work, processingUpdater work.ProcessingUpdater) *work.ProcessResult {
	return append(p.ProcessPipeline(ctx, wrk, processingUpdater),
		p.FetchProviderSessionFromWorkMetadata,
		p.FetchDataSourceFromProviderSession,
		p.FetchTokenSource,
		p.updateDataSourceProviderExternalID,
		p.updateProviderSessionExternalID,
		p.ensureDataSetForDataSource,
		p.createDataSourceStateChangeEventWork,
		p.createDataHistoricWork,
	).Process(p.Delete)
}

func (p *Processor) updateDataSourceProviderExternalID() *work.ProcessResult {
	if p.DataSource().ProviderExternalID != nil {
		return nil
	}

	// Get the user personal info that has the oura external id
	personalInfo, err := p.GetPersonalInfo(p.Context(), p.TokenSource())
	if err != nil {
		return p.Failing(errors.Wrap(err, "unable to get user personal info"))
	}

	// Get all data sources
	dataSourceFilter := &dataSource.Filter{
		ProviderType:       pointer.From(oauth.ProviderType),
		ProviderName:       pointer.From(oura.ProviderName),
		ProviderExternalID: pointer.From(*personalInfo.ID),
	}
	dataSources, err := page.Collect(func(pagination page.Pagination) (dataSource.SourceArray, error) {
		return p.DataSourceClient().List(p.Context(), p.ProviderSession().UserID, dataSourceFilter, &pagination)
	})
	if err != nil {
		return p.Failing(errors.Wrap(err, "unable to list data sources"))
	}

	// If at least one data source, then replace the current data source, otherwise just update current with external id
	if count := len(dataSources); count > 0 {
		if count > 1 {
			log.LoggerFromContext(p.Context()).WithField("count", count).Error("unexpected number of data sources found for provider external id")
		}
		return p.ReplaceDataSource(dataSources[0])
	} else {
		return p.UpdateDataSource(&dataSource.Update{ProviderExternalID: personalInfo.ID})
	}
}

func (p *Processor) updateProviderSessionExternalID() *work.ProcessResult {
	if p.ProviderSession().ExternalID != nil {
		return nil
	}

	// Update current with external id
	providerSessionUpdate := &auth.ProviderSessionUpdate{
		OAuthToken: p.ProviderSession().OAuthToken,
		ExternalID: p.DataSource().ProviderExternalID,
	}
	return p.UpdateProviderSession(providerSessionUpdate)
}

func (p *Processor) ensureDataSetForDataSource() *work.ProcessResult {
	if p.DataSource().DataSetID == nil {
		return p.CreateDataSetForDataSource(NewDataSetCreate())
	} else {
		return p.FetchDataSetFromDataSource()
	}
}

func (p *Processor) createDataSourceStateChangeEventWork() *work.ProcessResult {
	if workCreate, err := customerioWork.NewDataSourceStateChangedEventWorkCreate(p.DataSource()); err != nil {
		return p.Failed(errors.Wrap(err, "unable to create data source state changed event work create"))
	} else if _, err = p.WorkClient().Create(p.Context(), workCreate); err != nil {
		return p.Failing(errors.Wrap(err, "unable to create data source state changed event work"))
	}

	log.LoggerFromContext(p.Context()).Debug("created data source state changed event work")
	return nil
}

func (p *Processor) createDataHistoricWork() *work.ProcessResult {
	if workCreate, err := ouraDataWorkHistoric.NewWorkCreate(p.ProviderSession().ID, times.TimeRange{From: p.DataSource().LatestDataTime}); err != nil {
		return p.Failed(errors.Wrap(err, "unable to create data historic work create"))
	} else if _, err = p.WorkClient().Create(p.Context(), workCreate); err != nil {
		return p.Failing(errors.Wrap(err, "unable to create data historic work"))
	}

	log.LoggerFromContext(p.Context()).Debug("created data historic work")
	return nil
}

func NewDataSetCreate() *data.DataSetCreate {
	return &data.DataSetCreate{
		Client: &data.DataSetClient{
			Name:    pointer.From(oura.DataSetClientName),
			Version: pointer.From(oura.DataSetClientVersion),
		},
		DataSetType: pointer.From(data.DataSetTypeContinuous),
		Deduplicator: &data.DeduplicatorDescriptor{
			Name:    pointer.From(dataDeduplicatorDeduplicator.DataSetDeleteOriginName),
			Version: pointer.From(dataDeduplicatorDeduplicator.DataSetDeleteOriginVersion),
		},
		DeviceManufacturers: pointer.From(oura.DeviceManufacturers),
		DeviceTags:          pointer.From(oura.DeviceTags),
		TimeProcessing:      pointer.From(data.TimeProcessingNone),
	}
}
