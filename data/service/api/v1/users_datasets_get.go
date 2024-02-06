package v1

import (
	"net/http"

	"github.com/tidepool-org/platform/data"
	dataService "github.com/tidepool-org/platform/data/service"
	dataStore "github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service"
)

// UsersDataSetsGet godoc
// @Summary Get data sets
// @Description Get data sets
// @Description Caller must be a service, the owner, or have the authorizations to do it in behalf of the user.
// @ID platform-data-api-UsersDataSetsGet
// @Produce json
// @Param userId path string true "user ID"
// @Param page query int false "When using pagination, page number" default(0)
// @Param size query int false "When using pagination, number of elements by page, 1<size<1000" minimum(1) maximum(1000) default(100)
// @Param deleted query bool false "True to return the deleted datasets"
// @Param deviceId  query string false "Filter on the deviceId"
// @Param state query string false "Filter of the state: open or closed"
// @Param dataSetType query string false "Filter of the type: continuous or normal"
// @Security TidepoolSessionToken
// @Security TidepoolServiceSecret
// @Security TidepoolAuthorization
// @Security TidepoolRestrictedToken
// @Success 200 {array} upload.Upload "Operation is a success"
// @Failure 400 {object} service.Error "User id is missing or JSON body is malformed"
// @Failure 403 {object} service.Error "Forbiden: caller is not authorized"
// @Failure 500 {object} service.Error "Unable to perform the operation"
// @Router /v1/users/:userId/datasets [get]
func UsersDataSetsGet(dataServiceContext dataService.Context) {
	req := dataServiceContext.Request()

	targetUserID := req.PathParam("userId")
	if targetUserID == "" {
		dataServiceContext.RespondWithError(ErrorUserIDMissing())
		return
	}

	permissions, err := dataServiceContext.PermissionClient().GetUserPermissions(req, targetUserID)
	if err != nil {
		if request.IsErrorUnauthorized(err) {
			dataServiceContext.RespondWithError(service.ErrorUnauthorized())
		} else {
			dataServiceContext.RespondWithInternalServerFailure("Unable to get user permissions", err)
		}
		return
	}
	if !permissions {
		dataServiceContext.RespondWithError(service.ErrorUnauthorized())
		return
	}

	filter := dataStore.NewFilter()
	pagination := page.NewPagination()
	if err := request.DecodeRequestQuery(req.Request, filter, pagination); err != nil {
		request.MustNewResponder(dataServiceContext.Response(), req).Error(http.StatusBadRequest, err)
		return
	}

	if dataServiceContext.IsUploadIdUsed() {
		ctx := req.Context()
		dataSets, err := dataServiceContext.DataRepository().GetDataSetsForUserByID(ctx, targetUserID, filter, pagination)
		if err != nil {
			dataServiceContext.RespondWithInternalServerFailure("Unable to get data sets for user", err)
			return
		}
		dataServiceContext.RespondWithStatusAndData(http.StatusOK, dataSets)
	} else {
		// fake one
		id := pointer.FromString(data.NewID())
		dataset := &upload.Upload{
			Base: types.Base{
				Active:          true,
				ID:              id,
				UserID:          pointer.FromString(targetUserID),
				VersionInternal: 3,
				Type:            "upload",
				UploadID:        id,
			},
			Client: &upload.Client{
				Name:    pointer.FromString("api.your-loops.com"),
				Version: pointer.FromString("1.0.0"),
			},
			DeviceManufacturers: &[]string{"Diabeloop"},
			DataSetType:         pointer.FromString("continuous"),
			DeviceModel:         pointer.FromString("DBLG1"),
			DeviceTags:          &[]string{"cgm", "insulin-pump"},
			DataState:           pointer.FromString("open"),
			State:               pointer.FromString("open"),
		}
		var dataSets []*upload.Upload
		dataSets = append(dataSets, dataset)
		dataServiceContext.RespondWithStatusAndData(http.StatusOK, dataSets)
	}
}
