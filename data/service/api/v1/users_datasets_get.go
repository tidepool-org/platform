package v1

import (
	"net/http"
	"strconv"

	dataService "github.com/tidepool-org/platform/data/service"
	dataStoreDEPRECATED "github.com/tidepool-org/platform/data/storeDEPRECATED"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/user"
)

const (
	ParameterFilterDeleted  = "deleted"
	ParameterPaginationPage = "page"
	ParameterPaginationSize = "size"
)

func UsersDatasetsGet(dataServiceContext dataService.Context) {
	ctx := dataServiceContext.Request().Context()

	targetUserID := dataServiceContext.Request().PathParam("userId")
	if targetUserID == "" {
		dataServiceContext.RespondWithError(ErrorUserIDMissing())
		return
	}

	if details := request.DetailsFromContext(ctx); !details.IsService() {
		permissions, err := dataServiceContext.UserClient().GetUserPermissions(ctx, details.UserID(), targetUserID)
		if err != nil {
			if errors.Code(err) == request.ErrorCodeUnauthorized {
				dataServiceContext.RespondWithError(service.ErrorUnauthorized())
			} else {
				dataServiceContext.RespondWithInternalServerFailure("Unable to get user permissions", err)
			}
			return
		}
		if _, ok := permissions[user.ViewPermission]; !ok {
			dataServiceContext.RespondWithError(service.ErrorUnauthorized())
			return
		}
	}

	filter := dataStoreDEPRECATED.NewFilter()
	pagination := page.NewPagination()

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
				} else if parsedValue < page.PaginationPageMinimum {
					errors = append(errors, service.ErrorValueNotGreaterThanOrEqualTo(parsedValue, page.PaginationPageMinimum).WithSourceParameter(ParameterPaginationPage))
				} else {
					pagination.Page = parsedValue
				}
			case ParameterPaginationSize:
				if parsedValue, err := strconv.Atoi(value); err != nil {
					errors = append(errors, service.ErrorTypeNotInteger(value).WithSourceParameter(ParameterPaginationSize))
				} else if parsedValue < page.PaginationSizeMinimum || parsedValue > page.PaginationSizeMaximum {
					errors = append(errors, service.ErrorValueNotInRange(parsedValue, page.PaginationSizeMinimum, page.PaginationSizeMaximum).WithSourceParameter(ParameterPaginationSize))
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

	datasets, err := dataServiceContext.DataSession().GetDatasetsForUserByID(ctx, targetUserID, filter, pagination)
	if err != nil {
		dataServiceContext.RespondWithInternalServerFailure("Unable to get datasets for user", err)
		return
	}

	dataServiceContext.RespondWithStatusAndData(http.StatusOK, datasets)
}
