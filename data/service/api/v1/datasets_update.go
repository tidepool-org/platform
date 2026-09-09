package v1

import (
	"net/http"

	"github.com/tidepool-org/platform/data"
	dataService "github.com/tidepool-org/platform/data/service"
	dataWorkPostprocess "github.com/tidepool-org/platform/data/work/postprocess"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service"
)

func DataSetsUpdate(dataServiceContext dataService.Context) {
	req := dataServiceContext.Request()
	res := dataServiceContext.Response()
	ctx := req.Context()
	lgr := log.LoggerFromContext(ctx)

	dataSetID := req.PathParam("dataSetId")
	if dataSetID == "" {
		dataServiceContext.RespondWithError(ErrorDataSetIDMissing())
		return
	}

	dataSet, err := dataServiceContext.DataRepository().GetDataSet(ctx, dataSetID)
	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to get data set by id", err)
		return
	}
	if dataSet == nil {
		dataServiceContext.RespondWithError(ErrorDataSetIDNotFound(dataSetID))
		return
	}

	details := request.GetAuthDetails(ctx)
	if !details.IsService() {
		var permissions permission.Permissions
		permissions, err = dataServiceContext.PermissionClient().GetUserPermissions(ctx, details.UserID(), *dataSet.UserID)
		if err != nil {
			if request.IsErrorUnauthorized(err) {
				dataServiceContext.RespondWithError(service.ErrorUnauthorized())
			} else {
				dataServiceContext.RespondWithInternalServerFailure("Unable to get user permissions", err)
			}
			return
		}
		if _, ok := permissions[permission.Write]; !ok {
			dataServiceContext.RespondWithError(service.ErrorUnauthorized())
			return
		}
	}

	if (dataSet.State != nil && *dataSet.State == "closed") || (dataSet.DataState != nil && *dataSet.DataState == "closed") { // TODO: Deprecated DataState (after data migration)
		dataServiceContext.RespondWithError(ErrorDataSetClosed(dataSetID))
		return
	}

	update := data.NewDataSetUpdate()
	if dataSet.DataSetType != nil && *dataSet.DataSetType == data.DataSetTypeContinuous {
		if !details.IsService() {
			dataServiceContext.RespondWithError(service.ErrorUnauthorized())
			return
		}
		if err = request.DecodeRequestBody(req.Request, update); err != nil {
			request.MustNewResponder(res, req).Error(http.StatusBadRequest, err)
			return
		}
	} else {
		update.State = pointer.FromString(data.DataSetStateClosed)
	}

	dataSet, err = dataServiceContext.DataRepository().UpdateDataSet(ctx, dataSetID, update)
	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to update data set", err)
		return
	}

	if update.State != nil && *update.State == "closed" {
		deduplicator, getErr := dataServiceContext.DataDeduplicatorFactory().Get(ctx, dataSet)
		if getErr != nil {
			dataServiceContext.RespondWithInternalServerFailure("Unable to get deduplicator", getErr)
			return
		} else if deduplicator == nil {
			dataServiceContext.RespondWithInternalServerFailure("Deduplicator not found")
			return
		} else if err = deduplicator.Close(ctx, dataSet); err != nil {
			dataServiceContext.RespondWithInternalServerFailure("Unable to close", err)
			return
		}

		if err = dataWorkPostprocess.Enqueue(ctx, dataServiceContext.WorkClient(), *dataSet.UserID, dataWorkPostprocess.ReasonUploadCompleted); err != nil {
			lgr.WithError(err).Error("Unable to report upload completed")
		}
	}

	if err = dataServiceContext.MetricClient().RecordMetric(ctx, "data_sets_update"); err != nil {
		lgr.WithError(err).Error("Unable to record metric")
	}

	dataServiceContext.RespondWithStatusAndData(http.StatusOK, dataSet)
}
