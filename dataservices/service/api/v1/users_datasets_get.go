package v1

import (
	"net/http"
	"strconv"

	"github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/dataservices/service"
	commonService "github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/userservices/client"
)

const (
	ParameterFilterDeleted  = "deleted"
	ParameterPaginationPage = "page"
	ParameterPaginationSize = "size"
)

func UsersDatasetsGet(serviceContext service.Context) {
	targetUserID := serviceContext.Request().PathParam("userid")
	if targetUserID == "" {
		serviceContext.RespondWithError(ErrorUserIDMissing())
		return
	}

	if !serviceContext.AuthenticationDetails().IsServer() {
		permissions, err := serviceContext.UserServicesClient().GetUserPermissions(serviceContext, serviceContext.AuthenticationDetails().UserID(), targetUserID)
		if err != nil {
			if client.IsUnauthorizedError(err) {
				serviceContext.RespondWithError(commonService.ErrorUnauthorized())
			} else {
				serviceContext.RespondWithInternalServerFailure("Unable to get user permissions", err)
			}
			return
		}
		if _, ok := permissions[client.ViewPermission]; !ok {
			serviceContext.RespondWithError(commonService.ErrorUnauthorized())
			return
		}
	}

	filter := store.NewFilter()
	pagination := store.NewPagination()

	// TODO: Consider refactoring query string parsing into separate function/interface/package

	var errors []*commonService.Error
	for key, values := range serviceContext.Request().URL.Query() {
		for _, value := range values {
			switch key {
			case ParameterFilterDeleted:
				if parsedValue, err := strconv.ParseBool(value); err != nil {
					errors = append(errors, commonService.ErrorTypeNotBoolean(value).WithSourceParameter(ParameterFilterDeleted))
				} else {
					filter.Deleted = parsedValue
				}
			case ParameterPaginationPage:
				if parsedValue, err := strconv.Atoi(value); err != nil {
					errors = append(errors, commonService.ErrorTypeNotInteger(value).WithSourceParameter(ParameterPaginationPage))
				} else if parsedValue < store.PaginationPageMinimum {
					errors = append(errors, commonService.ErrorValueNotGreaterThanOrEqualTo(parsedValue, store.PaginationPageMinimum).WithSourceParameter(ParameterPaginationPage))
				} else {
					pagination.Page = parsedValue
				}
			case ParameterPaginationSize:
				if parsedValue, err := strconv.Atoi(value); err != nil {
					errors = append(errors, commonService.ErrorTypeNotInteger(value).WithSourceParameter(ParameterPaginationSize))
				} else if parsedValue < store.PaginationSizeMinimum || parsedValue > store.PaginationSizeMaximum {
					errors = append(errors, commonService.ErrorValueNotInRange(parsedValue, store.PaginationSizeMinimum, store.PaginationSizeMaximum).WithSourceParameter(ParameterPaginationSize))
				} else {
					pagination.Size = parsedValue
				}
			}
		}
	}
	if len(errors) > 0 {
		serviceContext.RespondWithStatusAndErrors(http.StatusBadRequest, errors)
		return
	}

	datasets, err := serviceContext.DataStoreSession().GetDatasetsForUserByID(targetUserID, filter, pagination)
	if err != nil {
		serviceContext.RespondWithInternalServerFailure("Unable to get datasets for user", err)
		return
	}

	serviceContext.RespondWithStatusAndData(http.StatusOK, datasets)
}
