package provider

import (
	"context"
	"slices"

	"github.com/tidepool-org/platform/auth"
	authProviderSession "github.com/tidepool-org/platform/auth/providersession"
	dataSource "github.com/tidepool-org/platform/data/source"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/oauth"
	oauthProviderClient "github.com/tidepool-org/platform/oauth/provider/client"
	"github.com/tidepool-org/platform/oura"
	ouraClient "github.com/tidepool-org/platform/oura/client"
	ouraDataWorkSetup "github.com/tidepool-org/platform/oura/data/work/setup"
	ouraUsersWorkRevoke "github.com/tidepool-org/platform/oura/users/work/revoke"
	ouraWork "github.com/tidepool-org/platform/oura/work"
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
	partnerURL            string
	partnerSecret         string
	client                *ouraClient.Client
}

// Compile time check for making sure Provider is a valid oauth.Provider
var _ oauth.Provider = &Provider{}

func New(dependencies Dependencies) (*Provider, error) {
	if err := dependencies.Validate(); err != nil {
		return nil, errors.Wrap(err, "dependencies is invalid")
	}

	oauthProviderClient, err := oauthProviderClient.NewWithErrorParser(oura.ProviderName, dependencies.Config.Config, nil, &ouraClient.ErrorResponseParser{})
	if err != nil {
		return nil, err
	}

	provider := &Provider{
		Provider:              oauthProviderClient,
		providerSessionClient: dependencies.ProviderSessionClient,
		dataSourceClient:      dependencies.DataSourceClient,
		workClient:            dependencies.WorkClient,
		acceptURL:             dependencies.Config.ProviderConfig.AcceptURL,
		partnerURL:            dependencies.Config.PartnerURL,
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
	_, err = p.connectDataSourceToProviderSession(ctx, providerSession, dataSrc)
	if err != nil {
		return errors.Wrap(err, "unable to connect data source")
	}
	if err = p.createDataSetupWork(ctx, providerSession); err != nil {
		return errors.Wrap(err, "unable to create data setup work")
	}
	return nil
}

func (p *Provider) OnDelete(ctx context.Context, providerSession *auth.ProviderSession) error {
	if err := p.disconnectDataSourceFromProviderSession(ctx, providerSession); err != nil {
		return errors.Wrap(err, "unable to disconnect data source")
	}
	if err := p.deleteWorkForProviderSession(ctx, providerSession); err != nil {
		return errors.Wrap(err, "unable to delete work for provider session")
	}
	if err := p.createUsersRevokeWork(ctx, providerSession); err != nil {
		return errors.Wrap(err, "unable to create users revoke work")
	}
	return nil
}

// FUTURE: Remove this function to allow all users to authorize
func (p *Provider) AllowUserInitiatedAction(ctx context.Context, userID string, action string) (bool, error) {
	switch action {
	case oauth.ActionAuthorize:
		dataSrcFilter := &dataSource.Filter{
			ProviderType: pointer.From(p.Type()),
			ProviderName: pointer.From(p.Name()),
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

func (p *Provider) PartnerURL() string {
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
		ProviderType: pointer.From(p.Type()),
		ProviderName: pointer.From(p.Name()),
	}
	dataSrcs, err := page.Collect(func(pagination page.Pagination) (dataSource.SourceArray, error) {
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
			ProviderType: p.Type(),
			ProviderName: p.Name(),
		}
		if dataSrc, err = p.dataSourceClient.Create(ctx, providerSession.UserID, dataSrcCreate); err != nil {
			return nil, errors.Wrap(err, "unable to create data source")
		}
	}

	// Unexpected association with provider session id, clean up
	if dataSrc.ProviderSessionID != nil {
		lgr.Warn("data source associated with existing provider session")

		if err = p.providerSessionClient.DeleteProviderSession(ctx, *dataSrc.ProviderSessionID); err != nil {
			return nil, errors.Wrap(err, "unable to delete existing provider session")
		}
		if dataSrc, err = p.dataSourceClient.Get(ctx, dataSrc.ID); err != nil {
			return nil, errors.Wrap(err, "unable to get data source")
		}
	}

	// Unexpected state for data source, clean up
	if dataSrc.State != dataSource.StateDisconnected {
		lgr.WithField("state", dataSrc.State).Warn("data source in unexpected state")

		srcUpdate := &dataSource.Update{
			State: pointer.From(dataSource.StateDisconnected),
		}
		if dataSrc, err = p.dataSourceClient.Update(ctx, dataSrc.ID, nil, srcUpdate); err != nil {
			return nil, errors.Wrap(err, "unable to update data source")
		}
	}

	return dataSrc, nil
}

func (p *Provider) connectDataSourceToProviderSession(ctx context.Context, providerSession *auth.ProviderSession, dataSrc *dataSource.Source) (*dataSource.Source, error) {
	var err error

	providerSessionUpdate := &auth.ProviderSessionUpdate{
		OAuthToken: providerSession.OAuthToken,
		ExternalID: providerSession.ExternalID,
	}
	if providerSession, err = p.providerSessionClient.UpdateProviderSession(ctx, providerSession.ID, providerSessionUpdate); err != nil {
		return nil, errors.Wrap(err, "unable to update provider session")
	}

	dataSrcUpdate := &dataSource.Update{
		ProviderSessionID: pointer.From(providerSession.ID),
		State:             pointer.From(dataSource.StateConnected),
	}
	dataSrc, err = p.dataSourceClient.Update(ctx, dataSrc.ID, nil, dataSrcUpdate)
	if err != nil {
		return nil, errors.Wrap(err, "unable to update data source")
	}

	return dataSrc, nil
}

func (p *Provider) disconnectDataSourceFromProviderSession(ctx context.Context, providerSession *auth.ProviderSession) error {
	if providerSession == nil {
		return errors.New("provider session is missing")
	}

	ctx = log.ContextWithField(ctx, "providerSessionId", providerSession.ID)

	dataSrc, err := p.dataSourceClient.GetFromProviderSession(ctx, providerSession.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get data source from provider session")
	}

	ctx, lgr := log.ContextAndLoggerWithField(ctx, "dataSourceId", dataSrc.ID)

	dataSrcUpdate := &dataSource.Update{
		State: pointer.From(dataSource.StateDisconnected),
	}
	if _, err := p.dataSourceClient.Update(ctx, dataSrc.ID, nil, dataSrcUpdate); err != nil {
		return errors.Wrap(err, "unable to update data source")
	}

	lgr.Debug("disconnected data source from provider session")
	return nil
}

func (p *Provider) deleteWorkForProviderSession(ctx context.Context, providerSession *auth.ProviderSession) error {
	if providerSession == nil {
		return errors.New("provider session is missing")
	}

	count, err := p.workClient.DeleteAllByGroupID(ctx, ouraWork.GroupIDFromProviderSessionID(providerSession.ID))
	if err != nil {
		return errors.Wrap(err, "unable to delete all work by group id")
	}

	log.LoggerFromContext(ctx).WithField("count", count).Debug("deleted work for provider session")
	return nil
}

func (p *Provider) createDataSetupWork(ctx context.Context, providerSession *auth.ProviderSession) error {
	if workCreate, err := ouraDataWorkSetup.NewWorkCreate(providerSession.ID); err != nil {
		return errors.Wrap(err, "unable to create data setup work create")
	} else if _, err = p.workClient.Create(ctx, workCreate); err != nil {
		return err
	}

	log.LoggerFromContext(ctx).Debug("created data setup work create")
	return nil
}

func (p *Provider) createUsersRevokeWork(ctx context.Context, providerSession *auth.ProviderSession) error {
	if workCreate, err := ouraUsersWorkRevoke.NewWorkCreate(providerSession.ID, providerSession.OAuthToken); err != nil {
		return errors.Wrap(err, "unable to create users revoke work create")
	} else if _, err = p.workClient.Create(ctx, workCreate); err != nil {
		return err
	}

	log.LoggerFromContext(ctx).Debug("created users revoke work")
	return nil
}
