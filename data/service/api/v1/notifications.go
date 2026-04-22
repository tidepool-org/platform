package v1

import (
	"net/http"

	dataService "github.com/tidepool-org/platform/data/service"
	notificationsWorkClaims "github.com/tidepool-org/platform/notifications/work/claims"
	notificationsWorkConnectionsIssues "github.com/tidepool-org/platform/notifications/work/connections/issues"
	notificationsWorkConnectionsRequests "github.com/tidepool-org/platform/notifications/work/connections/requests"
	"github.com/tidepool-org/platform/request"
	serviceApi "github.com/tidepool-org/platform/service/api"
)

func NotificationsRoutes() []dataService.Route {
	return []dataService.Route{
		dataService.Post("/v1/notifications/account/claims", QueueClaimAccountNotification, serviceApi.RequireServer),
		dataService.Post("/v1/notifications/account/connections", QueueConnectAccountNotification, serviceApi.RequireServer),
		dataService.Post("/v1/notifications/device/issues", SendDeviceIssuesNotification, serviceApi.RequireServer),
	}
}

func QueueClaimAccountNotification(dataServiceContext dataService.Context) {
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()

	responder := request.MustNewResponder(res, req)

	var metadata notificationsWorkClaims.Metadata
	if err := request.DecodeRequestBody(req.Request, &metadata); err != nil {
		request.MustNewResponder(res, req).Error(http.StatusBadRequest, err)
		return
	}

	if err := notificationsWorkClaims.AddWorkItem(req.Context(), dataServiceContext.WorkClient(), metadata); err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Empty(http.StatusCreated)
}

func QueueConnectAccountNotification(dataServiceContext dataService.Context) {
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()

	responder := request.MustNewResponder(res, req)

	var metadata notificationsWorkConnectionsRequests.Metadata
	if err := request.DecodeRequestBody(req.Request, &metadata); err != nil {
		request.MustNewResponder(res, req).Error(http.StatusBadRequest, err)
		return
	}

	if err := notificationsWorkConnectionsRequests.AddWorkItem(req.Context(), dataServiceContext.WorkClient(), metadata); err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Empty(http.StatusCreated)
}

func SendDeviceIssuesNotification(dataServiceContext dataService.Context) {
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()

	responder := request.MustNewResponder(res, req)

	var metadata notificationsWorkConnectionsIssues.Metadata
	if err := request.DecodeRequestBody(req.Request, &metadata); err != nil {
		request.MustNewResponder(res, req).Error(http.StatusBadRequest, err)
		return
	}

	if err := notificationsWorkConnectionsIssues.AddWorkItem(req.Context(), dataServiceContext.WorkClient(), metadata); err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Empty(http.StatusCreated)
}
