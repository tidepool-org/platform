package v1

import (
	"net/http"

	dataService "github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/user"
)

func DatasetsDelete(dataServiceContext dataService.Context) {
	ctx := dataServiceContext.Request().Context()
	lgr := log.LoggerFromContext(ctx)

	datasetID := dataServiceContext.Request().PathParam("dataSetId")
	if datasetID == "" {
		dataServiceContext.RespondWithError(ErrorDatasetIDMissing())
		return
	}

	dataset, err := dataServiceContext.DataSession().GetDatasetByID(ctx, datasetID)
	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to get dataset by id", err)
		return
	}
	if dataset == nil {
		dataServiceContext.RespondWithError(ErrorDatasetIDNotFound(datasetID))
		return
	}

	targetUserID := dataset.UserID
	if targetUserID == nil || *targetUserID == "" {
		dataServiceContext.RespondWithInternalServerFailure("Unable to get user id from dataset")
		return
	}

	if details := request.DetailsFromContext(ctx); !details.IsService() {
		authUserID := details.UserID()

		var permissions user.Permissions
		permissions, err = dataServiceContext.UserClient().GetUserPermissions(ctx, authUserID, *targetUserID)
		if err != nil {
			if request.IsErrorUnauthorized(err) {
				dataServiceContext.RespondWithError(service.ErrorUnauthorized())
			} else {
				dataServiceContext.RespondWithInternalServerFailure("Unable to get user permissions", err)
			}
			return
		}
		if _, ok := permissions[user.OwnerPermission]; !ok {
			if _, ok = permissions[user.CustodianPermission]; !ok {
				if _, ok = permissions[user.UploadPermission]; !ok || dataset.ByUser == nil || authUserID != *dataset.ByUser {
					dataServiceContext.RespondWithError(service.ErrorUnauthorized())
					return
				}
			}
		}
	}

	registered, err := dataServiceContext.DataDeduplicatorFactory().IsRegisteredWithDataset(dataset)
	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to check if registered with dataset", err)
		return
	}

	if registered {
		deduplicator, newErr := dataServiceContext.DataDeduplicatorFactory().NewRegisteredDeduplicatorForDataset(lgr, dataServiceContext.DataSession(), dataset)
		if newErr != nil {
			dataServiceContext.RespondWithInternalServerFailure("Unable to create registered deduplicator for dataset", newErr)
			return
		}
		err = deduplicator.DeleteDataset(ctx)
	} else {
		err = dataServiceContext.DataSession().DeleteDataset(ctx, dataset)
	}

	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to delete dataset", err)
		return
	}

	if err = dataServiceContext.MetricClient().RecordMetric(ctx, "datasets_delete"); err != nil {
		lgr.WithError(err).Error("Unable to record metric")
	}

	dataServiceContext.RespondWithStatusAndData(http.StatusOK, struct{}{})
}
