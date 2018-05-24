package v1

import (
	"net/http"

	"github.com/tidepool-org/platform/data"
	dataService "github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/user"
)

func DatasetsUpdate(dataServiceContext dataService.Context) {
	req := dataServiceContext.Request()
	res := dataServiceContext.Response()
	ctx := req.Context()
	lgr := log.LoggerFromContext(ctx)

	datasetID := req.PathParam("dataSetId")
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

	details := request.DetailsFromContext(ctx)
	if !details.IsService() {
		var permissions user.Permissions
		permissions, err = dataServiceContext.UserClient().GetUserPermissions(ctx, details.UserID(), *dataset.UserID)
		if err != nil {
			if errors.Code(err) == request.ErrorCodeUnauthorized {
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

	if (dataset.State != nil && *dataset.State == "closed") || (dataset.DataState != nil && *dataset.DataState == "closed") { // TODO: Deprecated DataState (after data migration)
		dataServiceContext.RespondWithError(ErrorDatasetClosed(datasetID))
		return
	}

	update := data.NewDataSetUpdate()
	if dataset.DataSetType != nil && *dataset.DataSetType == upload.DataSetTypeContinuous {
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

	dataset, err = dataServiceContext.DataSession().UpdateDataSet(ctx, datasetID, update)
	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to update dataset", err)
		return
	}

	deduplicator, err := dataServiceContext.DataDeduplicatorFactory().NewRegisteredDeduplicatorForDataset(lgr, dataServiceContext.DataSession(), dataset)
	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to create registered deduplicator for dataset", err)
		return
	}

	if err = deduplicator.DeduplicateDataset(ctx); err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to deduplicate dataset", err)
		return
	}

	if err = dataServiceContext.MetricClient().RecordMetric(ctx, "datasets_update"); err != nil {
		lgr.WithError(err).Error("Unable to record metric")
	}

	dataServiceContext.RespondWithStatusAndData(http.StatusOK, dataset)
}
