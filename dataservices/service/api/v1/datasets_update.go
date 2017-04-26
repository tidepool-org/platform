package v1

import (
	"net/http"

	"github.com/tidepool-org/platform/dataservices/service"
	commonService "github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/userservices/client"
)

func DatasetsUpdate(serviceContext service.Context) {
	datasetID := serviceContext.Request().PathParam("datasetid")
	if datasetID == "" {
		serviceContext.RespondWithError(ErrorDatasetIDMissing())
		return
	}

	dataset, err := serviceContext.DataStoreSession().GetDatasetByID(datasetID)
	if err != nil {
		serviceContext.RespondWithInternalServerFailure("Unable to get dataset by id", err)
		return
	}
	if dataset == nil {
		serviceContext.RespondWithError(ErrorDatasetIDNotFound(datasetID))
		return
	}

	if !serviceContext.AuthenticationDetails().IsServer() {
		var permissions client.Permissions
		permissions, err = serviceContext.UserServicesClient().GetUserPermissions(serviceContext, serviceContext.AuthenticationDetails().UserID(), dataset.UserID)
		if err != nil {
			if client.IsUnauthorizedError(err) {
				serviceContext.RespondWithError(commonService.ErrorUnauthorized())
			} else {
				serviceContext.RespondWithInternalServerFailure("Unable to get user permissions", err)
			}
			return
		}
		if _, ok := permissions[client.UploadPermission]; !ok {
			serviceContext.RespondWithError(commonService.ErrorUnauthorized())
			return
		}
	}

	if dataset.State == "closed" || dataset.DataState == "closed" { // TODO: Deprecated DataState (after data migration)
		serviceContext.RespondWithError(ErrorDatasetClosed(datasetID))
		return
	}

	dataset.State = "closed"

	if err = serviceContext.DataStoreSession().UpdateDataset(dataset); err != nil {
		serviceContext.RespondWithInternalServerFailure("Unable to update dataset", err)
		return
	}

	deduplicator, err := serviceContext.DataDeduplicatorFactory().NewRegisteredDeduplicatorForDataset(serviceContext.Logger(), serviceContext.DataStoreSession(), dataset)
	if err != nil {
		serviceContext.RespondWithInternalServerFailure("Unable to create registered deduplicator for dataset", err)
		return
	}

	if err = deduplicator.DeduplicateDataset(); err != nil {
		serviceContext.RespondWithInternalServerFailure("Unable to deduplicate dataset", err)
		return
	}

	if err = serviceContext.MetricServicesClient().RecordMetric(serviceContext, "datasets_update"); err != nil {
		serviceContext.Logger().WithError(err).Error("Unable to record metric")
	}

	serviceContext.RespondWithStatusAndData(http.StatusOK, dataset)
}
