package v1

import (
	"fmt"

	"github.com/tidepool-org/platform/log"

	dataService "github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/data/source"
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
		if !authDetails.IsService() && !dataServiceContext.TwiistServiceAccountAuthorizer().IsAuthorized(authDetails.UserID()) {
			lgr.Debugf("the subject is not authorized twiist service account")
			dataServiceContext.RespondWithError(service.ErrorUnauthorized())
			return
		}

		// Inject service auth details, because the twiist service account doesn't have direct sharing permissions
		// to upload data to linked accounts
		ctx := request.NewContextWithAuthDetails(req.Context(), request.NewAuthDetails(request.MethodServiceSecret, "", ""))
		req.Request = dataServiceContext.Request().Clone(ctx)

		filter := source.NewFilter()
		filter.ProviderName = pointer.FromAny([]string{twiistProvider.ProviderName})
		filter.ProviderExternalID = pointer.FromAny([]string{tidepoolLinkID})
		filter.State = pointer.FromAny([]string{source.StateConnected})

		dataSources, err := dataServiceContext.DataSourceClient().List(ctx, filter, nil)
		if err != nil {
			lgr.WithError(err).Warnf("unable to fetch data source for tidepool link id %s", tidepoolLinkID)
			dataServiceContext.RespondWithInternalServerFailure("unable to fetch data sources", err)
			return
		}
		if len(dataSources) == 0 {
			lgr.WithError(err).Warnf("no connected data source found for tidepool link id %s", tidepoolLinkID)
			dataServiceContext.RespondWithError(ErrorTidepoolLinkIDNotFound())
			return
		}

		var dataSetID string
		dataSource := dataSources[0]
		if dataSource.DataSetIDs != nil || len(*dataSource.DataSetIDs) > 0 {
			dataSetID = (*dataSource.DataSetIDs)[len(*dataSource.DataSetIDs)-1]
		}
		if dataSetID == "" {
			lgr.WithError(err).Warnf("no data sets found for tidepool link id %s", tidepoolLinkID)
			dataServiceContext.RespondWithInternalServerFailure(fmt.Sprintf("data set id is missing in data source %s", *dataSource.ID), err)
			return
		}

		// Inject the resolved data set id as a path parameter, so it can be used by DataSetsDataCreate
		req.PathParams["dataSetId"] = dataSetID

		datasetDataCreate(dataServiceContext)
	}
}
