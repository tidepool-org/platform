package provider

import (
	"context"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"

	"github.com/tidepool-org/platform/data"
	dataClient "github.com/tidepool-org/platform/data/client"
	dataDeduplicatorDeduplicator "github.com/tidepool-org/platform/data/deduplicator/deduplicator"
	dataTypesUpload "github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/twiist"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/config"
	dataSource "github.com/tidepool-org/platform/data/source"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	oauthProvider "github.com/tidepool-org/platform/oauth/provider"
	"github.com/tidepool-org/platform/pointer"
)

const ProviderName = "twiist"
const MetadataKeyExternalSubjectID = "externalSubjectID"

type Provider struct {
	*oauthProvider.Provider
	dataClient       dataClient.Client
	dataSourceClient dataSource.Client
}

func New(configReporter config.Reporter, dataClient dataClient.Client, dataSourceClient dataSource.Client, jwks jwk.Set) (*Provider, error) {
	if configReporter == nil {
		return nil, errors.New("config reporter is missing")
	}
	if dataClient == nil {
		return nil, errors.New("data client is missing")
	}
	if dataSourceClient == nil {
		return nil, errors.New("data source client is missing")
	}
	if jwks == nil {
		return nil, errors.New("jwks is missing")
	}

	prvdr, err := oauthProvider.NewProvider(ProviderName, configReporter.WithScopes(ProviderName), jwks)
	if err != nil {
		return nil, err
	}

	return &Provider{
		Provider:         prvdr,
		dataClient:       dataClient,
		dataSourceClient: dataSourceClient,
	}, nil
}

func (p *Provider) BeforeCreate(ctx context.Context, _ string, create *auth.ProviderSessionCreate) error {
	if create == nil {
		return errors.New("create is missing")
	}

	claims, err := p.getClaimsFromIDToken(ctx, create.OAuthToken)
	if err != nil {
		return errors.Wrap(err, "unable to get claims from id token")
	}
	if claims.TidepoolLinkID == "" {
		return errors.New("tidepool_link_id was not found in id_token")
	}

	create.ExternalID = &claims.TidepoolLinkID
	return nil
}

func (p *Provider) OnCreate(ctx context.Context, userID string, providerSession *auth.ProviderSession) error {
	if userID == "" {
		return errors.New("user id is missing")
	}
	if providerSession == nil {
		return errors.New("provider session is missing")
	}

	source, err := p.prepareDataSource(ctx, userID, providerSession)
	if err != nil {
		return err
	}
	dataSet, err := p.prepareDataSet(ctx, source)
	if err != nil {
		return err
	}

	err = p.connectDataSource(ctx, source, providerSession, dataSet)
	if err != nil {
		return errors.Wrap(err, "unable to connect data source")
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

	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"userId": userID, "providerSessionId": providerSession.ID})

	filter := dataSource.NewFilter()
	filter.ProviderType = pointer.FromStringArray([]string{p.Type()})
	filter.ProviderName = pointer.FromStringArray([]string{p.Name()})
	filter.State = pointer.FromStringArray([]string{dataSource.StateConnected})

	update := dataSource.NewUpdate()
	update.State = pointer.FromString(dataSource.StateDisconnected)

	dataSources, err := p.dataSourceClient.List(ctx, userID, filter, nil)
	if err != nil {
		logger.WithError(err).Error("Unable to fetch list data sources while deleting provider session")
		return err
	}
	for _, source := range dataSources {
		if source == nil || source.ID == nil {
			continue
		}
		_, err = p.dataSourceClient.Update(ctx, *source.ID, nil, update)
		if err != nil {
			logger.WithError(err).WithField("dataSourceId", *source.ID).Error("Unable to update data source while deleting provider session")
		}
	}

	return nil
}

func (p *Provider) prepareDataSource(ctx context.Context, userID string, providerSession *auth.ProviderSession) (*dataSource.Source, error) {
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"userId": userID, "type": p.Type(), "name": p.Name()})

	filter := dataSource.NewFilter()
	filter.ProviderType = pointer.FromStringArray([]string{p.Type()})
	filter.ProviderName = pointer.FromStringArray([]string{p.Name()})
	sources, err := p.dataSourceClient.List(ctx, userID, filter, nil)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch data sources")
	}

	var source *dataSource.Source
	if count := len(sources); count > 0 {
		if count > 1 {
			logger.WithField("count", count).Warn("unexpected number of data sources found")
		}

		for _, source := range sources {
			if *source.State != dataSource.StateDisconnected {
				logger.WithFields(log.Fields{"id": source.ID, "state": source.State}).Warn("data source in unexpected state")

				update := dataSource.NewUpdate()
				update.State = pointer.FromString(dataSource.StateDisconnected)

				_, err = p.dataSourceClient.Update(ctx, *source.ID, nil, update)
				if err != nil {
					return nil, errors.Wrap(err, "unable to update data source")
				}
			}
		}

		source = sources[0]
	} else {
		create := dataSource.NewCreate()
		create.ProviderType = pointer.FromString(p.Type())
		create.ProviderName = pointer.FromString(p.Name())
		create.ProviderExternalID = providerSession.ExternalID

		source, err = p.dataSourceClient.Create(ctx, userID, create)
		if err != nil {
			return nil, errors.Wrap(err, "unable to create data source")
		}
	}

	return source, nil
}

func (p *Provider) prepareDataSet(ctx context.Context, source *dataSource.Source) (*data.DataSet, error) {
	dataSet, err := p.findDataSet(ctx, source)
	if err != nil {
		return nil, err
	}
	if dataSet == nil {
		dataSet, err = p.createDataSet(ctx, source)
		if err != nil {
			return nil, errors.Wrap(err, "unable to create data source")
		}
	}

	return dataSet, nil
}

func (p *Provider) connectDataSource(ctx context.Context, source *dataSource.Source, providerSession *auth.ProviderSession, dataSet *data.DataSet) error {
	var dataSetIDs []string
	if source.DataSetIDs != nil {
		dataSetIDs = *source.DataSetIDs
	}

	var exists bool
	for _, dataSetID := range dataSetIDs {
		if dataSetID == *dataSet.ID {
			exists = true
			break
		}
	}

	// Only append the Data Set ID if it's unique
	if !exists {
		dataSetIDs = append(dataSetIDs, *dataSet.ID)
	}

	update := dataSource.Update{
		DataSetIDs:         pointer.FromStringArray(dataSetIDs),
		ProviderExternalID: providerSession.ExternalID,
		ProviderSessionID:  pointer.FromString(providerSession.ID),
		State:              pointer.FromString(dataSource.StateConnected),
	}

	// The external id of the data source and provider session is set to the Tidepool Link ID.
	// The Tidepool Link ID changes every time a user re-links their account. We also need to
	// keep track of the twiist user id which is different from the Tidepool Link ID, in case
	// we need to support backfilling of historical data and a user has more than one twiist
	// accounts. The twiist user id can be found in the id token.
	claims, err := p.getClaimsFromIDToken(ctx, providerSession.OAuthToken)
	if err != nil {
		return errors.Wrap(err, "unable to get claims from id token")
	}
	if claims.Subject != "" {
		update.Metadata = source.Metadata
		if update.Metadata == nil {
			update.Metadata = make(map[string]any)
		}
		update.Metadata[MetadataKeyExternalSubjectID] = claims.Subject
	}

	_, err = p.dataSourceClient.Update(ctx, *source.ID, nil, &update)
	if err != nil {
		return errors.Wrap(err, "unable to update source with data set id")
	}

	return nil
}

func (p *Provider) findDataSet(ctx context.Context, source *dataSource.Source) (*data.DataSet, error) {
	if source.DataSetIDs != nil {
		for index := len(*source.DataSetIDs) - 1; index >= 0; index-- {
			if dataSet, err := p.dataClient.GetDataSet(ctx, (*source.DataSetIDs)[index]); err != nil {
				return nil, errors.Wrap(err, "unable to get data set")
			} else if dataSet != nil {
				return dataSet, nil
			}
		}
	}
	return nil, nil
}

func (p *Provider) createDataSet(ctx context.Context, source *dataSource.Source) (*data.DataSet, error) {
	dataSetCreate := data.NewDataSetCreate()
	dataSetCreate.Client = &data.DataSetClient{
		Name:    pointer.FromString(twiist.DataSetClientName),
		Version: pointer.FromString(twiist.DataSetClientVersion),
	}
	dataSetCreate.DataSetType = pointer.FromString(data.DataSetTypeContinuous)
	dataSetCreate.Deduplicator = data.NewDeduplicatorDescriptor()
	dataSetCreate.Deduplicator.Name = pointer.FromString(dataDeduplicatorDeduplicator.DataSetDeleteOriginName)
	dataSetCreate.Deduplicator.Version = pointer.FromString("1.0.0")
	dataSetCreate.DeviceManufacturers = pointer.FromStringArray(twiist.DeviceManufacturers)
	dataSetCreate.DeviceTags = pointer.FromStringArray([]string{data.DeviceTagCGM, data.DeviceTagBGM, data.DeviceTagInsulinPump})
	dataSetCreate.Time = pointer.FromTime(time.Now())
	dataSetCreate.TimeProcessing = pointer.FromString(dataTypesUpload.TimeProcessingNone)

	dataSet, err := p.dataClient.CreateUserDataSet(ctx, *source.UserID, dataSetCreate)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create data set")
	}

	return dataSet, nil
}

func (p *Provider) getClaimsFromIDToken(ctx context.Context, token *auth.OAuthToken) (*Claims, error) {
	if token == nil {
		return nil, errors.New("oauth token is missing")
	}
	if token.IDToken == nil {
		return nil, errors.New("id token is missing")
	}
	claims := &Claims{}
	err := p.Provider.ParseIDToken(ctx, *token.IDToken, claims)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse id_token")
	}

	return claims, nil
}
