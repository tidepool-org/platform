package provider

import (
	"context"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/config"
	dataSource "github.com/tidepool-org/platform/data/source"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	oauthProvider "github.com/tidepool-org/platform/oauth/provider"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
)

const ProviderName = "twiist"

//go:generate mockgen -source=provider.go -destination=test/provider.go -package test ProviderSessionClient
type ProviderSessionClient interface {
	DeleteProviderSession(ctx context.Context, id string) error
}

//go:generate mockgen -source=provider.go -destination=test/provider.go -package test DataSourceClient
type DataSourceClient interface {
	List(ctx context.Context, userID string, filter *dataSource.Filter, pagination *page.Pagination) (dataSource.SourceArray, error)
	Create(ctx context.Context, userID string, create *dataSource.Create) (*dataSource.Source, error)
	Update(ctx context.Context, id string, condition *request.Condition, update *dataSource.Update) (*dataSource.Source, error)
}

type ProviderDependencies struct {
	ConfigReporter        config.Reporter
	ProviderSessionClient ProviderSessionClient
	DataSourceClient      DataSourceClient
}

func (p ProviderDependencies) Validate() error {
	if p.ConfigReporter == nil {
		return errors.New("config reporter is missing")
	}
	if p.ProviderSessionClient == nil {
		return errors.New("provider session client is missing")
	}
	if p.DataSourceClient == nil {
		return errors.New("data source client is missing")
	}
	return nil
}

type Provider struct {
	*oauthProvider.Provider
	providerSessionClient ProviderSessionClient
	dataSourceClient      DataSourceClient
}

func NewProvider(providerDependencies ProviderDependencies) (*Provider, error) {
	if err := providerDependencies.Validate(); err != nil {
		return nil, err
	}

	prvdr, err := oauthProvider.NewProvider(ProviderName, providerDependencies.ConfigReporter.WithScopes(ProviderName))
	if err != nil {
		return nil, err
	}

	return &Provider{
		Provider:              prvdr,
		providerSessionClient: providerDependencies.ProviderSessionClient,
		dataSourceClient:      providerDependencies.DataSourceClient,
	}, nil
}

func (p *Provider) OnCreate(ctx context.Context, userID string, providerSession *auth.ProviderSession) error {
	if userID == "" {
		return errors.New("user id is missing")
	}
	if providerSession == nil {
		return errors.New("provider session is missing")
	}

	lgr := log.LoggerFromContext(ctx)

	srcFilter := &dataSource.Filter{
		ProviderType: pointer.FromStringArray([]string{p.Type()}),
		ProviderName: pointer.FromStringArray([]string{p.Name()}),
	}
	srcs, err := p.dataSourceClient.List(ctx, userID, srcFilter, nil)
	if err != nil {
		return errors.Wrap(err, "unable to get data sources")
	}

	var src *dataSource.Source
	if count := len(srcs); count > 0 {
		if count > 1 {
			lgr.WithField("count", count).Error("user has multiple data sources for provider")
		}
		src = srcs[0]
	}

	if src == nil {
		srcCreate := &dataSource.Create{
			ProviderType: pointer.FromString(p.Type()),
			ProviderName: pointer.FromString(p.Name()),
		}
		if src, err = p.dataSourceClient.Create(ctx, userID, srcCreate); err != nil {
			return errors.Wrap(err, "unable to create data source")
		}
	}

	ctx, lgr = log.ContextAndLoggerWithField(ctx, "dataSourceId", *src.ID)

	// Unexpected association with provider session id, clean up
	if src.ProviderSessionID != nil {
		lgr.Warn("data source associated with existing provider session")

		if err := p.providerSessionClient.DeleteProviderSession(ctx, *src.ProviderSessionID); err != nil {
			lgr.WithError(err).Warn("failure deleting existing provider session")
		}
	}

	// Unexpected state for data source, cleanup
	if *src.State != dataSource.StateDisconnected {
		lgr.WithField("state", src.State).Warn("data source in unexpected state")

		srcUpdate := &dataSource.Update{
			State: pointer.FromString(dataSource.StateDisconnected),
		}
		if _, err = p.dataSourceClient.Update(ctx, *src.ID, nil, srcUpdate); err != nil {
			return errors.Wrap(err, "unable to update data source")
		}
	}

	srcUpdate := &dataSource.Update{
		ProviderSessionID: pointer.FromString(providerSession.ID),
		State:             pointer.FromString(dataSource.StateConnected),
	}
	if _, err = p.dataSourceClient.Update(ctx, *src.ID, nil, srcUpdate); err != nil {
		return errors.Wrap(err, "unable to update data source")
	}

	return nil
}

func (p *Provider) OnDelete(ctx context.Context, userID string, providerSession *auth.ProviderSession) error {
	if userID == "" {
		return errors.New("user id is missing")
	}
	if providerSession == nil {
		return errors.New("provider session is missing")
	}

	lgr := log.LoggerFromContext(ctx)

	srcFilter := &dataSource.Filter{
		ProviderType:      pointer.FromStringArray([]string{p.Type()}),
		ProviderName:      pointer.FromStringArray([]string{p.Name()}),
		ProviderSessionID: pointer.FromStringArray([]string{providerSession.ID}),
	}
	srcs, err := p.dataSourceClient.List(ctx, userID, srcFilter, nil)
	if err != nil {
		return errors.Wrap(err, "unable to get data sources")
	}

	if count := len(srcs); count > 1 {
		lgr.WithField("count", count).Warn("unexpected number of data sources found for provider session")
	}

	srcUpdate := &dataSource.Update{
		State: pointer.FromString(dataSource.StateDisconnected),
	}
	for _, src := range srcs {
		srcCtx, srcLgr := log.ContextAndLoggerWithField(ctx, "dataSourceId", *src.ID) // TODO: Update both context and logger

		if _, err = p.dataSourceClient.Update(srcCtx, *src.ID, nil, srcUpdate); err != nil {
			srcLgr.WithError(err).Error("unable to update data source while deleting provider session")
		}
	}

	return nil
}
