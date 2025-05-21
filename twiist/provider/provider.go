package provider

import (
	"context"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/data"
	dataDeduplicatorDeduplicator "github.com/tidepool-org/platform/data/deduplicator/deduplicator"
	dataSource "github.com/tidepool-org/platform/data/source"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/oauth"
	oauthProvider "github.com/tidepool-org/platform/oauth/provider"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/twiist"
)

const ProviderName = "twiist"

//go:generate mockgen -source=provider.go -destination=test/provider_mocks.go -package=test ProviderSessionClient
type ProviderSessionClient interface {
	UpdateProviderSession(ctx context.Context, id string, update *auth.ProviderSessionUpdate) (*auth.ProviderSession, error)
	DeleteProviderSession(ctx context.Context, id string) error
}

//go:generate mockgen -source=provider.go -destination=test/provider_mocks.go -package=test DataSourceClient
type DataSourceClient interface {
	List(ctx context.Context, userID string, filter *dataSource.Filter, pagination *page.Pagination) (dataSource.SourceArray, error)
	Create(ctx context.Context, userID string, create *dataSource.Create) (*dataSource.Source, error)
	Update(ctx context.Context, id string, condition *request.Condition, update *dataSource.Update) (*dataSource.Source, error)
}

//go:generate mockgen -source=provider.go -destination=test/provider_mocks.go -package=test DataSetClient
type DataSetClient interface {
	CreateUserDataSet(ctx context.Context, userID string, create *data.DataSetCreate) (*data.DataSet, error)
	GetDataSet(ctx context.Context, id string) (*data.DataSet, error)
}

type ProviderDependencies struct {
	ConfigReporter        config.Reporter
	ProviderSessionClient ProviderSessionClient
	DataSourceClient      DataSourceClient
	DataSetClient         DataSetClient
	JWKS                  jwk.Set
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
	if p.DataSetClient == nil {
		return errors.New("data set client is missing")
	}
	return nil
}

type Provider struct {
	*oauthProvider.Provider
	providerSessionClient ProviderSessionClient
	dataSourceClient      DataSourceClient
	dataSetClient         DataSetClient
}

// Compile time check for making sure Provider is a valid oauth.Provider
var _ oauth.Provider = &Provider{}

func NewProvider(providerDependencies ProviderDependencies) (*Provider, error) {
	if err := providerDependencies.Validate(); err != nil {
		return nil, err
	}

	prvdr, err := oauthProvider.NewProvider(ProviderName, providerDependencies.ConfigReporter.WithScopes(ProviderName), providerDependencies.JWKS)
	if err != nil {
		return nil, err
	}

	return &Provider{
		Provider:              prvdr,
		providerSessionClient: providerDependencies.ProviderSessionClient,
		dataSourceClient:      providerDependencies.DataSourceClient,
		dataSetClient:         providerDependencies.DataSetClient,
	}, nil
}

func (p *Provider) OnCreate(ctx context.Context, providerSession *auth.ProviderSession) error {
	if providerSession == nil {
		return errors.New("provider session is missing")
	}

	dataSrc, err := p.prepareDataSource(ctx, providerSession)
	if err != nil {
		return errors.Wrap(err, "unable to prepare data source")
	}

	if err = p.prepareDataSet(ctx, dataSrc); err != nil {
		return errors.Wrap(err, "unable to prepare data set")
	}

	if err = p.connectDataSource(ctx, providerSession, dataSrc); err != nil {
		return errors.Wrap(err, "unable to connect data source")
	}

	return nil
}

func (p *Provider) OnDelete(ctx context.Context, providerSession *auth.ProviderSession) error {
	if providerSession == nil {
		return errors.New("provider session is missing")
	}

	lgr := log.LoggerFromContext(ctx)

	dataSrcFilter := &dataSource.Filter{
		ProviderType:      pointer.FromStringArray([]string{p.Type()}),
		ProviderName:      pointer.FromStringArray([]string{p.Name()}),
		ProviderSessionID: pointer.FromStringArray([]string{providerSession.ID}),
	}
	dataSrcs, err := p.dataSourceClient.List(ctx, providerSession.UserID, dataSrcFilter, nil)
	if err != nil {
		return errors.Wrap(err, "unable to get data sources")
	}

	if count := len(dataSrcs); count > 1 {
		lgr.WithField("count", count).Warn("unexpected number of data sources found for provider session")
	}

	dataSrcUpdate := &dataSource.Update{
		State: pointer.FromString(dataSource.StateDisconnected),
	}
	for _, dataSrc := range dataSrcs {
		dataSrcCtx, dataSrcLgr := log.ContextAndLoggerWithField(ctx, "dataSourceId", dataSrc.ID)

		if _, err = p.dataSourceClient.Update(dataSrcCtx, *dataSrc.ID, nil, dataSrcUpdate); err != nil {
			dataSrcLgr.WithError(err).Error("unable to update data source while deleting provider session")
		}
	}

	// TODO: BACK-3652 - Allow unlinking of twiist account from Tidepool-initiated action

	return nil
}

func (p *Provider) SupportsUserInitiatedAccountUnlinking() bool {
	return false
}

func (p *Provider) prepareDataSource(ctx context.Context, providerSession *auth.ProviderSession) (*dataSource.Source, error) {
	lgr := log.LoggerFromContext(ctx)

	providerSessionExternalID, dataSrcExternalID, err := p.extractExternalIDsFromProviderSession(providerSession)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get external ids from provider session")
	}

	dataSrcFilter := &dataSource.Filter{
		ProviderType:       pointer.FromStringArray([]string{p.Type()}),
		ProviderName:       pointer.FromStringArray([]string{p.Name()}),
		ProviderExternalID: pointer.FromStringArray([]string{dataSrcExternalID}),
	}
	dataSrcs, err := p.dataSourceClient.List(ctx, providerSession.UserID, dataSrcFilter, nil)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get data sources")
	}

	var dataSrc *dataSource.Source
	if count := len(dataSrcs); count > 0 {
		if count > 1 {
			lgr.WithField("count", count).Error("user has multiple data sources for provider")
		}
		dataSrc = dataSrcs[0]
	}

	if dataSrc == nil {
		dataSrcCreate := &dataSource.Create{
			ProviderType:       pointer.FromString(p.Type()),
			ProviderName:       pointer.FromString(p.Name()),
			ProviderExternalID: pointer.FromString(dataSrcExternalID),
		}
		if dataSrc, err = p.dataSourceClient.Create(ctx, providerSession.UserID, dataSrcCreate); err != nil {
			return nil, errors.Wrap(err, "unable to create data source")
		}
	}

	// Unexpected association with provider session id, clean up
	if dataSrc.ProviderSessionID != nil {
		lgr.Warn("data source associated with existing provider session")

		if err := p.providerSessionClient.DeleteProviderSession(ctx, *dataSrc.ProviderSessionID); err != nil {
			lgr.WithError(err).Warn("failure deleting existing provider session")
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

	providerSession.ExternalID = pointer.FromString(providerSessionExternalID)

	return dataSrc, nil
}

func (p *Provider) prepareDataSet(ctx context.Context, dataSrc *dataSource.Source) error {
	dataSet, err := newDataSetEnsurer(p.dataSetClient).Ensure(ctx, *dataSrc)
	if err != nil {
		return err
	}

	dataSrc.AddDataSetID(*dataSet.ID)

	return nil
}

func (p *Provider) connectDataSource(ctx context.Context, providerSession *auth.ProviderSession, dataSrc *dataSource.Source) error {
	providerSessionUpdate := &auth.ProviderSessionUpdate{
		OAuthToken: providerSession.OAuthToken,
		ExternalID: providerSession.ExternalID,
	}
	if _, err := p.providerSessionClient.UpdateProviderSession(ctx, providerSession.ID, providerSessionUpdate); err != nil {
		return errors.Wrap(err, "unable to update provider session")
	}

	dataSrcUpdate := &dataSource.Update{
		DataSetIDs:         dataSrc.DataSetIDs,
		ProviderExternalID: dataSrc.ProviderExternalID,
		ProviderSessionID:  pointer.FromString(providerSession.ID),
		State:              pointer.FromString(dataSource.StateConnected),
	}
	if _, err := p.dataSourceClient.Update(ctx, *dataSrc.ID, nil, dataSrcUpdate); err != nil {
		return errors.Wrap(err, "unable to update data source")
	}

	return nil
}

// The ExternalID of the provider session is set to the Tidepool link id.  The Tidepool link id
// changes every time a user re-links their account.
//
// The ExternalID of the data source is set to the twiist user id. The twiist user id is
// the stable identifier for the user in twiist and is necessary to reconnect disconnected
// data sources.
func (p *Provider) extractExternalIDsFromProviderSession(providerSession *auth.ProviderSession) (string, string, error) {
	var claims Claims

	if providerSession.OAuthToken == nil {
		return "", "", nil
	} else if idToken := providerSession.OAuthToken.IDToken; idToken == nil {
		return "", "", nil
	} else if err := p.Provider.ParseToken(*idToken, &claims); err != nil {
		return "", "", errors.Wrap(err, "unable to parse id token")
	} else if !claims.VerifyAudience(p.ClientID(), true) {
		return "", "", errors.New("unable to verify id token audience claim")
	} else if claims.Subject == "" {
		return "", "", errors.New("subject is missing from claims from id token")
	} else if claims.TidepoolLinkID == "" {
		return "", "", errors.New("tidepool link id is missing from claims from id token")
	} else {
		return claims.TidepoolLinkID, claims.Subject, nil
	}
}

func newDataSetEnsurer(dataSetClient dataSource.DataSetEnsurerClient) *dataSource.DataSetEnsurer {
	return &dataSource.DataSetEnsurer{
		Client:  dataSetClient,
		Factory: dataSetCreateFactory{},
	}
}

type dataSetCreateFactory struct{}

func (d dataSetCreateFactory) NewDataSetCreate(dataSrc dataSource.Source) data.DataSetCreate {
	return data.DataSetCreate{
		Client: &data.DataSetClient{
			Name:    pointer.FromString(twiist.DataSetClientName),
			Version: pointer.FromString(twiist.DataSetClientVersion),
		},
		DataSetType: pointer.FromString(data.DataSetTypeContinuous),
		Deduplicator: &data.DeduplicatorDescriptor{
			Name: pointer.FromString(dataDeduplicatorDeduplicator.DataSetDeleteOriginOlderName),
		},
		DeviceManufacturers: pointer.FromStringArray(twiist.DeviceManufacturers),
		DeviceTags:          pointer.FromStringArray(twiist.DeviceTags),
		Time:                pointer.FromTime(time.Now()),
		TimeProcessing:      pointer.FromString(data.TimeProcessingNone),
	}
}
