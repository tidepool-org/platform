package v1

import (
	"net/http"

	"github.com/tidepool-org/platform/data"
	dataService "github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/user"
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

	dataSet, err := dataServiceContext.DataSession().GetDataSetByID(ctx, dataSetID)
	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to get data set by id", err)
		return
	}
	if dataSet == nil {
		dataServiceContext.RespondWithError(ErrorDataSetIDNotFound(dataSetID))
		return
	}

	details := request.DetailsFromContext(ctx)
	if !details.IsService() {
		var permissions user.Permissions
		permissions, err = dataServiceContext.UserClient().GetUserPermissions(ctx, details.UserID(), *dataSet.UserID)
		if err != nil {
			if request.IsErrorUnauthorized(err) {
				dataServiceContext.RespondWithError(service.ErrorUnauthorized())
			} else {
				dataServiceContext.RespondWithInternalServerFailure("Unable to get user permissions", err)
			}
			return
		}
		if _, ok := permissions[user.UploadPermission]; !ok {
			dataServiceContext.RespondWithError(service.ErrorUnauthorized())
			return
		}
	}

	if (dataSet.State != nil && *dataSet.State == "closed") || (dataSet.DataState != nil && *dataSet.DataState == "closed") { // TODO: Deprecated DataState (after data migration)
		dataServiceContext.RespondWithError(ErrorDataSetClosed(dataSetID))
		return
	}

	update := data.NewDataSetUpdate()
	if dataSet.DataSetType != nil && *dataSet.DataSetType == upload.DataSetTypeContinuous {
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

	dataSet, err = dataServiceContext.DataSession().UpdateDataSet(ctx, dataSetID, update)
	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to update data set", err)
		return
	}

	deduplicator, err := dataServiceContext.DataDeduplicatorFactory().NewRegisteredDeduplicatorForDataSet(lgr, dataServiceContext.DataSession(), dataSet)
	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to create registered deduplicator for data set", err)
		return
	}

	if err = deduplicator.DeduplicateDataSet(ctx); err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to deduplicate data set", err)
		return
	}

	if err = dataServiceContext.MetricClient().RecordMetric(ctx, "data_sets_update"); err != nil {
		lgr.WithError(err).Error("Unable to record metric")
	}

	dataServiceContext.RespondWithStatusAndData(http.StatusOK, dataSet)
}
