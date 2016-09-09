package v1

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

import (
	"net/http"

	"github.com/tidepool-org/platform/dataservices/service"
	commonService "github.com/tidepool-org/platform/service"
)

func UsersDataDelete(serviceContext service.Context) {
	targetUserID := serviceContext.Request().PathParam("userid")
	if targetUserID == "" {
		serviceContext.RespondWithError(ErrorUserIDMissing())
		return
	}

	if !serviceContext.AuthenticationDetails().IsServer() {
		serviceContext.RespondWithError(commonService.ErrorUnauthorized())
		return
	}

	if err := serviceContext.DataStoreSession().DeleteDataForUser(targetUserID); err != nil {
		serviceContext.RespondWithInternalServerFailure("Unable to delete data for user", err)
		return
	}

	if err := serviceContext.MetricServicesClient().RecordMetric(serviceContext, "users_data_delete"); err != nil {
		serviceContext.Logger().WithError(err).Error("Unable to record metric")
	}

	serviceContext.RespondWithStatusAndData(http.StatusOK, []struct{}{})
}
