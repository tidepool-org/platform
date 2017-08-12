package v1

import (
	"net/http"

	"github.com/tidepool-org/platform/client"
	dataService "github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/service"
	userClient "github.com/tidepool-org/platform/user/client"
)

func DatasetsDelete(dataServiceContext dataService.Context) {
	datasetID := dataServiceContext.Request().PathParam("datasetid")
	if datasetID == "" {
		dataServiceContext.RespondWithError(ErrorDatasetIDMissing())
		return
	}

	dataset, err := dataServiceContext.DataStoreSession().GetDatasetByID(datasetID)
	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to get dataset by id", err)
		return
	}
	if dataset == nil {
		dataServiceContext.RespondWithError(ErrorDatasetIDNotFound(datasetID))
		return
	}

	targetUserID := dataset.UserID
	if targetUserID == "" {
		dataServiceContext.RespondWithInternalServerFailure("Unable to get user id from dataset")
		return
	}

	if !dataServiceContext.AuthDetails().IsServer() {
		authUserID := dataServiceContext.AuthDetails().UserID()

		var permissions userClient.Permissions
		permissions, err = dataServiceContext.UserClient().GetUserPermissions(dataServiceContext, authUserID, targetUserID)
		if err != nil {
			if client.IsUnauthorizedError(err) {
				dataServiceContext.RespondWithError(service.ErrorUnauthorized())
			} else {
				dataServiceContext.RespondWithInternalServerFailure("Unable to get user permissions", err)
			}
			return
		}
		if _, ok := permissions[userClient.OwnerPermission]; !ok {
			if _, ok = permissions[userClient.CustodianPermission]; !ok {
				if _, ok = permissions[userClient.UploadPermission]; !ok || authUserID != dataset.ByUser {
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
		deduplicator, newErr := dataServiceContext.DataDeduplicatorFactory().NewRegisteredDeduplicatorForDataset(dataServiceContext.Logger(), dataServiceContext.DataStoreSession(), dataset)
		if newErr != nil {
			dataServiceContext.RespondWithInternalServerFailure("Unable to create registered deduplicator for dataset", newErr)
			return
		}
		err = deduplicator.DeleteDataset()
	} else {
		err = dataServiceContext.DataStoreSession().DeleteDataset(dataset)
	}

	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to delete dataset", err)
		return
	}

	if err = dataServiceContext.MetricClient().RecordMetric(dataServiceContext, "datasets_delete"); err != nil {
		dataServiceContext.Logger().WithError(err).Error("Unable to record metric")
	}

	dataServiceContext.RespondWithStatusAndData(http.StatusOK, struct{}{})
}
