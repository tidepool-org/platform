package v1

import (
	"net/http"
	"strconv"

	dataService "github.com/tidepool-org/platform/data/service"
	dataStore "github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/service"
	userClient "github.com/tidepool-org/platform/user/client"
)

const (
	ParameterFilterDeleted  = "deleted"
	ParameterPaginationPage = "page"
	ParameterPaginationSize = "size"
)

func UsersDatasetsGet(dataServiceContext dataService.Context) {
	targetUserID := dataServiceContext.Request().PathParam("userid")
	if targetUserID == "" {
		dataServiceContext.RespondWithError(ErrorUserIDMissing())
		return
	}

	if !dataServiceContext.AuthenticationDetails().IsServer() {
		permissions, err := dataServiceContext.UserClient().GetUserPermissions(dataServiceContext, dataServiceContext.AuthenticationDetails().UserID(), targetUserID)
		if err != nil {
			if userClient.IsUnauthorizedError(err) {
				dataServiceContext.RespondWithError(service.ErrorUnauthorized())
			} else {
				dataServiceContext.RespondWithInternalServerFailure("Unable to get user permissions", err)
			}
			return
		}
		if _, ok := permissions[userClient.ViewPermission]; !ok {
			dataServiceContext.RespondWithError(service.ErrorUnauthorized())
			return
		}
	}

	filter := dataStore.NewFilter()
	pagination := dataStore.NewPagination()

	// TODO: Consider refactoring query string parsing into separate function/interface/package

	var errors []*service.Error
	for key, values := range dataServiceContext.Request().URL.Query() {
		for _, value := range values {
			switch key {
			case ParameterFilterDeleted:
				if parsedValue, err := strconv.ParseBool(value); err != nil {
					errors = append(errors, service.ErrorTypeNotBoolean(value).WithSourceParameter(ParameterFilterDeleted))
				} else {
					filter.Deleted = parsedValue
				}
			case ParameterPaginationPage:
				if parsedValue, err := strconv.Atoi(value); err != nil {
					errors = append(errors, service.ErrorTypeNotInteger(value).WithSourceParameter(ParameterPaginationPage))
				} else if parsedValue < dataStore.PaginationPageMinimum {
					errors = append(errors, service.ErrorValueNotGreaterThanOrEqualTo(parsedValue, dataStore.PaginationPageMinimum).WithSourceParameter(ParameterPaginationPage))
				} else {
					pagination.Page = parsedValue
				}
			case ParameterPaginationSize:
				if parsedValue, err := strconv.Atoi(value); err != nil {
					errors = append(errors, service.ErrorTypeNotInteger(value).WithSourceParameter(ParameterPaginationSize))
				} else if parsedValue < dataStore.PaginationSizeMinimum || parsedValue > dataStore.PaginationSizeMaximum {
					errors = append(errors, service.ErrorValueNotInRange(parsedValue, dataStore.PaginationSizeMinimum, dataStore.PaginationSizeMaximum).WithSourceParameter(ParameterPaginationSize))
				} else {
					pagination.Size = parsedValue
				}
			}
		}
	}
	if len(errors) > 0 {
		dataServiceContext.RespondWithStatusAndErrors(http.StatusBadRequest, errors)
		return
	}

	datasets, err := dataServiceContext.DataStoreSession().GetDatasetsForUserByID(targetUserID, filter, pagination)
	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to get datasets for user", err)
		return
	}

	dataServiceContext.RespondWithStatusAndData(http.StatusOK, datasets)
}
