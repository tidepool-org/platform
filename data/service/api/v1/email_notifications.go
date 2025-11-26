package v1

import (
	"net/http"
	"time"

	dataService "github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/request"
	serviceApi "github.com/tidepool-org/platform/service/api"

	"github.com/tidepool-org/platform/conditionalnotifications/claimaccount"
	"github.com/tidepool-org/platform/conditionalnotifications/connectaccount"
	"github.com/tidepool-org/platform/conditionalnotifications/connectionissues"
)

func NotificationRoutes() []dataService.Route {
	return []dataService.Route{
		dataService.Post("/v1/notifications/account/claims", queueClaimAccountNotification, serviceApi.RequireServer),
		dataService.Post("/v1/notifications/account/connections", queueConnectAccountNotification, serviceApi.RequireServer),
		dataService.Post("/v1/notifications/device/issues", sendDeviceIssuesNotification, serviceApi.RequireServer),
	}
}

func queueClaimAccountNotification(dataServiceContext dataService.Context) {
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()

	responder := request.MustNewResponder(res, req)

	var data claimaccount.Metadata
	if err := request.DecodeRequestBody(req.Request, &data); err != nil {
		request.MustNewResponder(res, req).Error(http.StatusBadRequest, err)
		return
	}

	notBefore := data.WhenToSend
	if notBefore.IsZero() {
		notBefore = time.Now().Add(time.Hour * 24 * 7)
	}

	createDetails := claimaccount.NewWorkCreate(notBefore, data)
	_, err := dataServiceContext.WorkClient().Create(req.Context(), createDetails)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Empty(http.StatusCreated)
}

func queueConnectAccountNotification(dataServiceContext dataService.Context) {
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()

	responder := request.MustNewResponder(res, req)

	var data connectaccount.Metadata
	if err := request.DecodeRequestBody(req.Request, &data); err != nil {
		request.MustNewResponder(res, req).Error(http.StatusBadRequest, err)
		return
	}

	createDetails := connectaccount.NewWorkCreate(time.Now().Add(time.Hour*24*7), data)
	_, err := dataServiceContext.WorkClient().Create(req.Context(), createDetails)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Empty(http.StatusCreated)
}

func sendDeviceIssuesNotification(dataServiceContext dataService.Context) {
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()

	responder := request.MustNewResponder(res, req)

	var data connectionissues.Metadata
	if err := request.DecodeRequestBody(req.Request, &data); err != nil {
		request.MustNewResponder(res, req).Error(http.StatusBadRequest, err)
		return
	}

	createDetails := connectionissues.NewWorkCreate(data)
	_, err := dataServiceContext.WorkClient().Create(req.Context(), createDetails)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Empty(http.StatusCreated)
}
