package provider

import (
	"context"
	"net/url"
	"slices"

	"github.com/tidepool-org/platform/auth"
	authProviderSession "github.com/tidepool-org/platform/auth/providersession"
	customerioWork "github.com/tidepool-org/platform/customerio/work/event"
	dataSource "github.com/tidepool-org/platform/data/source"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/oauth"
	oauthProviderClient "github.com/tidepool-org/platform/oauth/provider/client"
	"github.com/tidepool-org/platform/oura"
	ouraClient "github.com/tidepool-org/platform/oura/client"
	ouraWorkData "github.com/tidepool-org/platform/oura/work/data"
	ouraWorkDataSetup "github.com/tidepool-org/platform/oura/work/data/setup"
	ouraWorkUsersRevoke "github.com/tidepool-org/platform/oura/work/users/revoke"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/work"
)

type Dependencies struct {
	Config                Config
	ProviderSessionClient authProviderSession.Client
	DataSourceClient      dataSource.Client
	WorkClient            work.Client
}

func (d Dependencies) Validate() error {
	if err := d.Config.Validate(); err != nil {
		return errors.Wrap(err, "config is invalid")
	}
	if d.ProviderSessionClient == nil {
		return errors.New("provider session client is missing")
	}
	if d.DataSourceClient == nil {
		return errors.New("data source client is missing")
	}
	if d.WorkClient == nil {
		return errors.New("work client is missing")
	}
	return nil
}

type Provider struct {
	*oauthProviderClient.Provider
	providerSessionClient authProviderSession.Client
	dataSourceClient      dataSource.Client
	workClient            work.Client
	acceptURL             *string
	partnerURL            *url.URL
	partnerSecret         string
	client                *ouraClient.Client
}

// Compile time check for making sure Provider is a valid oauth.Provider
var _ oauth.Provider = &Provider{}

func New(dependencies Dependencies) (*Provider, error) {
	if err := dependencies.Validate(); err != nil {
		return nil, errors.Wrap(err, "dependencies is invalid")
	}

	oauthProviderClient, err := oauthProviderClient.New(oura.ProviderName, dependencies.Config.Config, nil)
	if err != nil {
		return nil, err
	}

	partnerURL, err := url.Parse(dependencies.Config.PartnerURL)
	if err != nil {
		return nil, errors.Wrap(err, "partner url is invalid")
	}

	provider := &Provider{
		Provider:              oauthProviderClient,
		providerSessionClient: dependencies.ProviderSessionClient,
		dataSourceClient:      dependencies.DataSourceClient,
		workClient:            dependencies.WorkClient,
		acceptURL:             dependencies.Config.Provider.AcceptURL,
		partnerURL:            partnerURL,
		partnerSecret:         dependencies.Config.PartnerSecret,
	}

	if provider.client, err = ouraClient.NewWithClient(oauthProviderClient.Client, provider); err != nil {
		return nil, err
	}

	return provider, nil
}

func (p *Provider) OnCreate(ctx context.Context, providerSession *auth.ProviderSession) error {
	dataSrc, err := p.prepareDataSourceForProviderSession(ctx, providerSession)
	if err != nil {
		return errors.Wrap(err, "unable to prepare data source")
	}
	dataSrc, err = p.connectDataSourceToProviderSession(ctx, providerSession, dataSrc)
	if err != nil {
		return errors.Wrap(err, "unable to connect data source")
	}
	if err = p.createDataSetupWork(ctx, dataSrc); err != nil {
		return errors.Wrap(err, "unable to create data setup work")
	}
	if err = p.createDataSourceStateChangeEventWork(ctx, dataSrc); err != nil {
		return errors.Wrap(err, "unable to create data source state change event work")
	}
	return nil
}

func (p *Provider) OnDelete(ctx context.Context, providerSession *auth.ProviderSession) error {
	if err := p.disconnectDataSourcesFromProviderSession(ctx, providerSession); err != nil {
		return errors.Wrap(err, "unable to disconnect data sources")
	}
	if err := p.createUserRevokeWork(ctx, providerSession); err != nil {
		return errors.Wrap(err, "unable to create user revoke work")
	}
	return nil
}

// FUTURE: Remove this function to allow all users to authorize
func (p *Provider) AllowUserInitiatedAction(ctx context.Context, userID string, action string) (bool, error) {
	switch action {
	case oauth.ActionAuthorize:
		dataSrcFilter := &dataSource.Filter{
			ProviderType: pointer.FromStringArray([]string{p.Type()}),
			ProviderName: pointer.FromStringArray([]string{p.Name()}),
		}
		dataSrcs, err := p.dataSourceClient.List(ctx, userID, dataSrcFilter, page.NewPaginationMinimum())
		if err != nil {
			return false, errors.Wrap(err, "unable to get data sources")
		}
		return len(dataSrcs) > 0, nil
	default:
		return p.Provider.AllowUserInitiatedAction(ctx, userID, action)
	}
}

func (p *Provider) UserActionAcceptURL(ctx context.Context, userID string, action string) (*string, error) {
	switch action {
	case oauth.ActionAuthorize:
		return p.acceptURL, nil
	default:
		return p.Provider.UserActionAcceptURL(ctx, userID, action)
	}
}

func (p *Provider) PartnerURL() *url.URL {
	return p.partnerURL
}

func (p *Provider) PartnerSecret() string {
	return p.partnerSecret
}

func (p *Provider) Client() *ouraClient.Client {
	return p.client
}

func (p *Provider) prepareDataSourceForProviderSession(ctx context.Context, providerSession *auth.ProviderSession) (*dataSource.Source, error) {
	lgr := log.LoggerFromContext(ctx)

	// Get all data sources
	dataSrcFilter := &dataSource.Filter{
		ProviderType: pointer.FromStringArray([]string{p.Type()}),
		ProviderName: pointer.FromStringArray([]string{p.Name()}),
	}
	dataSrcs, err := page.Collect(func(pagination page.Pagination) ([]*dataSource.Source, error) {
		return p.dataSourceClient.List(ctx, providerSession.UserID, dataSrcFilter, &pagination)
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to list data sources")
	}

	// FUTURE: Remove this block to allow all users to connect
	if len(dataSrcs) == 0 {
		return nil, errors.New("data source does not exist")
	}

	// Only consider data sources without existing provider external id association
	dataSrcs = slices.DeleteFunc(dataSrcs, func(ds *dataSource.Source) bool { return ds.ProviderExternalID != nil })

	var dataSrc *dataSource.Source
	if len(dataSrcs) > 0 {
		dataSrc = dataSrcs[0]
	} else {
		dataSrcCreate := &dataSource.Create{
			ProviderType: pointer.FromString(p.Type()),
			ProviderName: pointer.FromString(p.Name()),
		}
		if dataSrc, err = p.dataSourceClient.Create(ctx, providerSession.UserID, dataSrcCreate); err != nil {
			return nil, errors.Wrap(err, "unable to create data source")
		}
	}

	// Unexpected association with provider session id, clean up
	if dataSrc.ProviderSessionID != nil {
		lgr.Warn("data source associated with existing provider session")

		if err := p.providerSessionClient.DeleteProviderSession(ctx, *dataSrc.ProviderSessionID); err != nil {
			lgr.WithError(err).Error("unable to delete existing provider session")
		}
	}

	// Unexpected state for data source, cleanup
	if *dataSrc.State != dataSource.StateDisconnected {
		lgr.WithField("state", dataSrc.State).Warn("data source in unexpected state")

		srcUpdate := &dataSource.Update{
			State: pointer.FromString(dataSource.StateDisconnected),
		}
		if _, err = p.dataSourceClient.Update(ctx, *dataSrc.ID, nil, srcUpdate); err != nil {
			return nil, errors.Wrap(err, "unable to update data source")
		}
	}

	return dataSrc, nil
}

func (p *Provider) connectDataSourceToProviderSession(ctx context.Context, providerSession *auth.ProviderSession, dataSrc *dataSource.Source) (*dataSource.Source, error) {
	providerSessionUpdate := &auth.ProviderSessionUpdate{
		OAuthToken: providerSession.OAuthToken,
		ExternalID: providerSession.ExternalID,
	}
	if _, err := p.providerSessionClient.UpdateProviderSession(ctx, providerSession.ID, providerSessionUpdate); err != nil {
		return nil, errors.Wrap(err, "unable to update provider session")
	}

	dataSrcUpdate := &dataSource.Update{
		ProviderSessionID:  pointer.FromString(providerSession.ID),
		ProviderExternalID: dataSrc.ProviderExternalID,
		State:              pointer.FromString(dataSource.StateConnected),
		DataSetIDs:         dataSrc.DataSetIDs,
	}
	dataSrc, err := p.dataSourceClient.Update(ctx, *dataSrc.ID, nil, dataSrcUpdate)
	if err != nil {
		return nil, errors.Wrap(err, "unable to update data source")
	}

	return dataSrc, nil
}

func (p *Provider) disconnectDataSourcesFromProviderSession(ctx context.Context, providerSession *auth.ProviderSession) error {
	if providerSession == nil {
		return errors.New("provider session is missing")
	}

	ctx, lgr := log.ContextAndLoggerWithField(ctx, "providerSessionId", providerSession.ID)

	dataSrcFilter := &dataSource.Filter{
		ProviderType:      pointer.FromStringArray([]string{providerSession.Type}),
		ProviderName:      pointer.FromStringArray([]string{providerSession.Name}),
		ProviderSessionID: pointer.FromStringArray([]string{providerSession.ID}),
	}
	dataSrcs, err := page.Collect(func(pagination page.Pagination) ([]*dataSource.Source, error) {
		return p.dataSourceClient.List(ctx, providerSession.UserID, dataSrcFilter, &pagination)
	})
	if err != nil {
		return errors.Wrap(err, "unable to get data sources")
	}

	if count := len(dataSrcs); count > 1 {
		lgr.WithField("count", count).Error("unexpected number of data sources found for provider session")
	}

	return p.disconnectDataSources(ctx, dataSrcs)
}

func (p *Provider) disconnectDataSources(ctx context.Context, dataSrcs dataSource.SourceArray) error {
	for _, dataSrc := range dataSrcs {
		if err := p.disconnectDataSource(ctx, dataSrc); err != nil {
			log.LoggerFromContext(ctx).WithField("dataSourceId", *dataSrc.ID).Error("unable to disconnect data source")
		}
	}
	return nil
}

func (p *Provider) disconnectDataSource(ctx context.Context, dataSrc *dataSource.Source) error {
	if dataSrc == nil {
		return errors.New("data source is missing")
	}

	ctx, lgr := log.ContextAndLoggerWithField(ctx, "dataSourceId", *dataSrc.ID)

	if _, err := p.dataSourceClient.Update(ctx, *dataSrc.ID, nil, &dataSource.Update{State: pointer.FromString(dataSource.StateDisconnected)}); err != nil {
		return errors.Wrap(err, "unable to update data source")
	}

	count, err := p.workClient.DeleteAllByGroupID(ctx, ouraWorkData.GroupIDFromDataSourceID(*dataSrc.ID))
	if err != nil {
		return errors.Wrap(err, "unable to delete all work")
	}

	lgr.WithField("count", count).Debug("deleted work associated with data source")
	return nil
}

func (p *Provider) createDataSetupWork(ctx context.Context, dataSrc *dataSource.Source) error {
	workCreate, err := ouraWorkDataSetup.NewWorkCreate(dataSrc)
	if err != nil {
		return errors.Wrap(err, "unable to create data setup work create")
	}
	_, err = p.workClient.Create(ctx, workCreate)
	return err
}

func (p *Provider) createDataSourceStateChangeEventWork(ctx context.Context, dataSrc *dataSource.Source) error {
	workCreate, err := customerioWork.NewDataSourceStateChangedEventWorkCreate(dataSrc)
	if err != nil {
		return errors.Wrap(err, "unable to create customer.io data source state changed event work create")
	}
	_, err = p.workClient.Create(ctx, workCreate)
	return err
}

func (p *Provider) createUserRevokeWork(ctx context.Context, providerSession *auth.ProviderSession) error {
	workCreate, err := ouraWorkUsersRevoke.NewWorkCreate(providerSession)
	if err != nil {
		return errors.Wrap(err, "unable to create user revoke work create")
	}
	_, err = p.workClient.Create(ctx, workCreate)
	return err
}
