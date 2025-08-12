package v1

import (
	"fmt"

	"github.com/tidepool-org/platform/auth"
	dataService "github.com/tidepool-org/platform/data/service"
	dataSource "github.com/tidepool-org/platform/data/source"
	"github.com/tidepool-org/platform/log"
	oauthProvider "github.com/tidepool-org/platform/oauth/provider"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service"
	twiistProvider "github.com/tidepool-org/platform/twiist/provider"
)

func NewTwiistDataCreateHandler(datasetDataCreate func(ctx dataService.Context)) func(ctx dataService.Context) {
	return func(dataServiceContext dataService.Context) {
		req := dataServiceContext.Request()
		lgr := log.LoggerFromContext(req.Context())

		tidepoolLinkID := req.PathParams["tidepoolLinkId"]
		if tidepoolLinkID == "" {
			dataServiceContext.RespondWithError(ErrorTidepoolLinkIDMissing())
			return
		}

		// Authorize the service account
		authDetails := request.GetAuthDetails(req.Context())
		if !authDetails.IsService() && !dataServiceContext.TwiistServiceAccountAuthorizer().IsServiceAccountAuthorized(authDetails.UserID()) {
			lgr.Debug("the subject is not authorized twiist service account")
			dataServiceContext.RespondWithError(service.ErrorUnauthorized())
			return
		}

		// Inject service auth details, because the twiist service account doesn't have direct sharing permissions
		// to upload data to linked accounts
		ctx := request.NewContextWithAuthDetails(req.Context(), request.NewAuthDetails(request.MethodServiceSecret, "", ""))
		req.Request = dataServiceContext.Request().Clone(ctx)

		// Find matching provider session
		providerSessionFilter := &auth.ProviderSessionFilter{
			Type:       pointer.FromString(oauthProvider.ProviderType),
			Name:       pointer.FromString(twiistProvider.ProviderName),
			ExternalID: pointer.FromString(tidepoolLinkID),
		}
		providerSessions, err := dataServiceContext.AuthClient().ListProviderSessions(ctx, providerSessionFilter, nil)
		if err != nil {
			lgr.WithError(err).Errorf("unable to fetch provider sessions for tidepool link id %s", tidepoolLinkID)
			dataServiceContext.RespondWithInternalServerFailure("unable to fetch provider sessions", err)
			return
		} else if length := len(providerSessions); length == 0 {
			lgr.Infof("no connected provider sessions found for tidepool link id %s", tidepoolLinkID)
			dataServiceContext.RespondWithError(ErrorTidepoolLinkIDNotFound())
			return
		} else if length > 1 {
			lgr.Warnf("multiple connected provider sessions found for tidepool link id %s", tidepoolLinkID)
		}
		providerSession := providerSessions[0]

		// Find matching data source
		dataSrcFilter := &dataSource.Filter{
			ProviderSessionID: pointer.FromAny([]string{providerSession.ID}),
		}
		dataSrcs, err := dataServiceContext.DataSourceClient().ListAll(ctx, dataSrcFilter, nil)
		if err != nil {
			lgr.WithError(err).Errorf("unable to fetch data sources for tidepool link id %s", tidepoolLinkID)
			dataServiceContext.RespondWithInternalServerFailure("unable to fetch data sources", err)
			return
		} else if length := len(dataSrcs); length == 0 {
			lgr.Warnf("no connected data sources found for tidepool link id %s", tidepoolLinkID)
			dataServiceContext.RespondWithError(ErrorTidepoolLinkIDNotFound())
			return
		} else if length > 1 {
			lgr.Warnf("multiple connected data sources found for tidepool link id %s", tidepoolLinkID)
		}
		dataSrc := dataSrcs[0]

		// Sanity check
		if *dataSrc.State != dataSource.StateConnected {
			lgr.Warnf("data source with id %s is not connected for tidepool link id %s", *dataSrc.ID, tidepoolLinkID)
			dataServiceContext.RespondWithError(ErrorTidepoolLinkIDNotFound())
			return
		}

		// Use last data set id
		dataSetID := dataSrc.LastDataSetID()
		if dataSetID == nil {
			lgr.Warnf("no data sets found for tidepool link id %q", tidepoolLinkID)
			dataServiceContext.RespondWithInternalServerFailure(fmt.Sprintf("data set id is missing in data source %q", *dataSrc.ID))
			return
		}

		// Inject the resolved data set id as a path parameter, so it can be used by DataSetsDataCreate
		req.PathParams["dataSetId"] = *dataSetID

		datasetDataCreate(dataServiceContext)
	}
}
