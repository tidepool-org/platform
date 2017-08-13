package v1

import (
	"net/http"

	"github.com/tidepool-org/platform/client"
	dataService "github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/service"
	userClient "github.com/tidepool-org/platform/user/client"
)

func DatasetsUpdate(dataServiceContext dataService.Context) {
	datasetID := dataServiceContext.Request().PathParam("datasetid")
	if datasetID == "" {
		dataServiceContext.RespondWithError(ErrorDatasetIDMissing())
		return
	}

	dataset, err := dataServiceContext.DataSession().GetDatasetByID(datasetID)
	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to get dataset by id", err)
		return
	}
	if dataset == nil {
		dataServiceContext.RespondWithError(ErrorDatasetIDNotFound(datasetID))
		return
	}

	if !dataServiceContext.AuthDetails().IsServer() {
		var permissions userClient.Permissions
		permissions, err = dataServiceContext.UserClient().GetUserPermissions(dataServiceContext, dataServiceContext.AuthDetails().UserID(), dataset.UserID)
		if err != nil {
			if client.IsUnauthorizedError(err) {
				dataServiceContext.RespondWithError(service.ErrorUnauthorized())
			} else {
				dataServiceContext.RespondWithInternalServerFailure("Unable to get user permissions", err)
			}
			return
		}
		if _, ok := permissions[userClient.UploadPermission]; !ok {
			dataServiceContext.RespondWithError(service.ErrorUnauthorized())
			return
		}
	}

	if dataset.State == "closed" || dataset.DataState == "closed" { // TODO: Deprecated DataState (after data migration)
		dataServiceContext.RespondWithError(ErrorDatasetClosed(datasetID))
		return
	}

	dataset.State = "closed"

	if err = dataServiceContext.DataSession().UpdateDataset(dataset); err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to update dataset", err)
		return
	}

	deduplicator, err := dataServiceContext.DataDeduplicatorFactory().NewRegisteredDeduplicatorForDataset(dataServiceContext.Logger(), dataServiceContext.DataSession(), dataset)
	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to create registered deduplicator for dataset", err)
		return
	}

	if err = deduplicator.DeduplicateDataset(); err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to deduplicate dataset", err)
		return
	}

	if err = dataServiceContext.MetricClient().RecordMetric(dataServiceContext, "datasets_update"); err != nil {
		dataServiceContext.Logger().WithError(err).Error("Unable to record metric")
	}

	dataServiceContext.RespondWithStatusAndData(http.StatusOK, dataset)
}
