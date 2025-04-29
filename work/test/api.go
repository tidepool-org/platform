package test

import (
	"net/http"

	dataService "github.com/tidepool-org/platform/data/service"
	request "github.com/tidepool-org/platform/request"
	work "github.com/tidepool-org/platform/work"
)

func WorkRoutes() []dataService.Route {
	return []dataService.Route{
		dataService.Post("/v1/work", CreateWorkHandler),
		dataService.Delete("/v1/work/:groupId", DeleteWorkHandler),
	}
}

func DeleteWorkHandler(dataServiceContext dataService.Context) {
	ctx := dataServiceContext.Request().Context()
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()

	responder := request.MustNewResponder(res, req)

	groupID := req.PathParam("groupId")
	if groupID == "" {
		responder.Error(http.StatusBadRequest, request.ErrorParameterMissing("groupId"))
		return
	}

	_, err := dataServiceContext.WorkClient().DeleteAllByGroupID(ctx, groupID)

	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Empty(http.StatusNoContent)
}

func CreateWorkHandler(dataServiceContext dataService.Context) {
	ctx := dataServiceContext.Request().Context()
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()

	responder := request.MustNewResponder(res, req)

	create := &work.Create{}
	if err := request.DecodeRequestBody(req.Request, create); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	wc, err := NewLoadWorkCreate(create)
	if err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	wrk, err := dataServiceContext.WorkClient().Create(ctx, wc)

	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Data(http.StatusCreated, wrk)
}
