package revoke

import (
	"context"
	"time"

	"github.com/tidepool-org/platform/auth"
	providerSession "github.com/tidepool-org/platform/auth/providersession"
	providerSessionWork "github.com/tidepool-org/platform/auth/providersession/work"
	"github.com/tidepool-org/platform/customerio"
	customerioWork "github.com/tidepool-org/platform/customerio/work/event"
	dataSource "github.com/tidepool-org/platform/data/source"
	dataSourceWork "github.com/tidepool-org/platform/data/source/work"
	dataWork "github.com/tidepool-org/platform/data/work"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/oauth"
	oauthWork "github.com/tidepool-org/platform/oauth/work"
	"github.com/tidepool-org/platform/oura"
	ouraWork "github.com/tidepool-org/platform/oura/work"
	ouraWorkData "github.com/tidepool-org/platform/oura/work/data"
	ouraWorkDataHistoric "github.com/tidepool-org/platform/oura/work/data/historic"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

const (
	Type      = "org.tidepool.oura.work.data.setup"
	Quantity  = 1
	Frequency = 5 * time.Second

	FailingRetryDuration       = 1 * time.Minute
	FailingRetryDurationJitter = 5 * time.Second

	ProcessingTimeout = 60 // Seconds
)

type Dependencies struct {
	ProviderSessionClient providerSession.Client
	DataSourceClient      dataSource.Client
	WorkClient            work.Client
	Client                ouraWork.Client
}

func (d Dependencies) Validate() error {
	if d.ProviderSessionClient == nil {
		return errors.New("provider session client is missing")
	}
	if d.DataSourceClient == nil {
		return errors.New("data source client is missing")
	}
	if d.WorkClient == nil {
		return errors.New("work client is missing")
	}
	if d.Client == nil {
		return errors.New("client is missing")
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

type (
	ProviderSessionMixin           = providerSessionWork.Mixin
	OAuthMixin                     = oauthWork.Mixin
	DataSourceMixin                = dataSourceWork.Mixin
	DataSourceProviderSessionMixin = dataWork.DataSourceProviderSessionMixin
)

type Processor struct {
	*workBase.Processor
	*ProviderSessionMixin
	*OAuthMixin
	*DataSourceMixin
	*DataSourceProviderSessionMixin
	DataSourceClient dataSource.Client
	WorkClient       work.Client
	Client           ouraWork.Client
	CustomerIOClient *customerio.Client
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

	processor, err := workBase.NewProcessor(processResultBuilder)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create processor")
	}

	providerSessionMixin, err := providerSessionWork.NewMixin(processor, dependencies.ProviderSessionClient)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create provider session mixin")
	}
	oauthMixin, err := oauthWork.NewMixin(processor, providerSessionMixin)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create oauth mixin")
	}
	dataSourceMixin, err := dataSourceWork.NewMixin(processor, dependencies.DataSourceClient)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create data source mixin")
	}
	dataSourceProviderSessionMixin, err := dataWork.NewDataSourceProviderSessionMixin(processor, dataSourceMixin, providerSessionMixin)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create data source provider session mixin")
	}

	return &Processor{
		Processor:                      processor,
		ProviderSessionMixin:           providerSessionMixin,
		OAuthMixin:                     oauthMixin,
		DataSourceMixin:                dataSourceMixin,
		DataSourceProviderSessionMixin: dataSourceProviderSessionMixin,
		DataSourceClient:               dependencies.DataSourceClient,
		WorkClient:                     dependencies.WorkClient,
		Client:                         dependencies.Client,
	}, nil
}

func (p *Processor) Process(ctx context.Context, wrk *work.Work, processingUpdater work.ProcessingUpdater) *work.ProcessResult {
	return work.ProcessPipeline{
		p.ProcessPipelineFunc(ctx, wrk, processingUpdater),
		p.DataSourceMixin.FetchDataSourceFromMetadata,
		p.FetchProviderSessionFromDataSource,
		p.OAuthMixin.FetchTokenSource,
		p.updateDataSourceProviderExternalID,
		p.updateProviderSessionExternalID,
		p.createDataHistoricWork,
		p.createDataSourceStateChangedEventWork,
		p.Delete,
	}.Process()
}

func (p *Processor) createDataSourceStateChangedEventWork() *work.ProcessResult {
	if workCreate, err := customerioWork.NewDataSourceStateChangedEventWorkCreate(p.DataSource); err != nil {
		return p.Failed(errors.Wrap(err, "unable to create work create"))
	} else if _, err = p.WorkClient.Create(p.Context(), workCreate); err != nil {
		return p.Failing(errors.Wrap(err, "unable to customer.io data source state changed event work"))
	} else {
		return nil
	}
}

func (p *Processor) updateDataSourceProviderExternalID() *work.ProcessResult {
	if p.DataSource.ProviderExternalID != nil {
		return nil
	}

	// Get the user personal info that has the oura external id
	personalInfo, err := p.Client.GetPersonalInfo(p.Context(), p.TokenSource())
	if err != nil {
		return p.Failing(errors.Wrap(err, "unable to get user personal info"))
	}

	// Get all data sources
	dataSrcFilter := &dataSource.Filter{
		ProviderType:       pointer.FromStringArray([]string{oauth.ProviderType}),
		ProviderName:       pointer.FromStringArray([]string{oura.ProviderName}),
		ProviderExternalID: pointer.FromStringArray([]string{*personalInfo.ID}),
	}
	dataSrcs, err := page.Collect(func(pagination page.Pagination) ([]*dataSource.Source, error) {
		return p.DataSourceClient.List(p.Context(), p.ProviderSession.UserID, dataSrcFilter, &pagination)
	})
	if err != nil {
		return p.Failing(errors.Wrap(err, "unable to list data sources"))
	}

	// If at least one data source, then replace the current data source, otherwise just update current with external id
	if count := len(dataSrcs); count > 0 {
		if count > 1 {
			p.Logger().WithField("count", count).Error("unexpected number of data sources found for provider external id")
		}
		return p.ReplaceDataSource(dataSrcs[0])
	} else {
		dataSrcUpdate := dataSource.Update{
			ProviderExternalID: personalInfo.ID,
		}
		return p.UpdateDataSource(dataSrcUpdate)
	}
}

func (p *Processor) updateProviderSessionExternalID() *work.ProcessResult {
	if p.ProviderSession.ExternalID != nil {
		return nil
	}

	// Update current with external id
	providerSessionUpdate := auth.ProviderSessionUpdate{
		OAuthToken: p.ProviderSession.OAuthToken,
		ExternalID: p.DataSource.ProviderExternalID,
	}
	return p.UpdateProviderSession(providerSessionUpdate)
}

func (p *Processor) createDataHistoricWork() *work.ProcessResult {
	if workCreate, err := ouraWorkDataHistoric.NewWorkCreate(p.DataSource, dataWork.TimeRange{From: p.DataSource.LatestDataTime}); err != nil {
		return p.Failed(errors.Wrap(err, "unable to create data historic work create"))
	} else if _, err = p.WorkClient.Create(p.Context(), workCreate); err != nil {
		return p.Failing(errors.Wrap(err, "unable to create data historic work"))
	} else {
		return nil
	}
}

func NewWorkCreate(dataSrc *dataSource.Source) (*work.Create, error) {
	if dataSrc == nil {
		return nil, errors.New("data source is missing")
	}
	return &work.Create{
		Type:              Type,
		GroupID:           pointer.FromString(ouraWorkData.GroupIDFromDataSourceID(*dataSrc.ID)),
		DeduplicationID:   pointer.FromString(*dataSrc.ID),
		SerialID:          pointer.FromString(ouraWorkData.SerialIDFromDataSourceID(*dataSrc.ID)),
		ProcessingTimeout: ProcessingTimeout,
		Metadata: map[string]any{
			dataSourceWork.MetadataKeyID: *dataSrc.ID,
		},
	}, nil
}
