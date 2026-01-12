package v1

import (
	"net/http"

	dataService "github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/request"
	serviceApi "github.com/tidepool-org/platform/service/api"

	"github.com/tidepool-org/platform/notifications/work/claims"
	connissues "github.com/tidepool-org/platform/notifications/work/connections/issues"
	connrequests "github.com/tidepool-org/platform/notifications/work/connections/requests"
)

func NotificationsRoutes() []dataService.Route {
	return []dataService.Route{
		dataService.Post("/v1/notifications/account/claims", QueueClaimAccountNotification, serviceApi.RequireServer),
		dataService.Post("/v1/notifications/account/connections", QueueConnectAccountNotification, serviceApi.RequireServer),
		dataService.Post("/v1/notifications/device/issues", SendDeviceIssuesNotification, serviceApi.RequireServer),
	}
}

//go:generate mockgen -source=../../context.go -destination=mocks/data_service_context_mock.go -package=mocks Context

func QueueClaimAccountNotification(dataServiceContext dataService.Context) {
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()

	responder := request.MustNewResponder(res, req)

	var data claims.Metadata
	if err := request.DecodeRequestBody(req.Request, &data); err != nil {
		request.MustNewResponder(res, req).Error(http.StatusBadRequest, err)
		return
	}

	if err := claims.AddWorkItem(req.Context(), dataServiceContext.WorkClient(), data); err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Empty(http.StatusCreated)
}

func QueueConnectAccountNotification(dataServiceContext dataService.Context) {
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()

	responder := request.MustNewResponder(res, req)

	var data connrequests.Metadata
	if err := request.DecodeRequestBody(req.Request, &data); err != nil {
		request.MustNewResponder(res, req).Error(http.StatusBadRequest, err)
		return
	}

	if err := connrequests.AddWorkItem(req.Context(), dataServiceContext.WorkClient(), data); err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Empty(http.StatusCreated)
}

func SendDeviceIssuesNotification(dataServiceContext dataService.Context) {
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()

	responder := request.MustNewResponder(res, req)

	var data connissues.Metadata
	if err := request.DecodeRequestBody(req.Request, &data); err != nil {
		request.MustNewResponder(res, req).Error(http.StatusBadRequest, err)
		return
	}

	if err := connissues.AddWorkItem(req.Context(), dataServiceContext.WorkClient(), data); err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Empty(http.StatusCreated)
}
