package v1

import (
	"net/http"

	dataService "github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service/api"
	"github.com/tidepool-org/platform/work"
)

func WorkRoutes() []dataService.Route {
	return []dataService.Route{
		dataService.Post("/v1/work", CreateWork, api.RequireServer),
		dataService.Get("/v1/work/:id", GetWork, api.RequireServer),
	}
}

func CreateWork(dataServiceContext dataService.Context) {
	req := dataServiceContext.Request()
	ctx := req.Context()
	responder := request.MustNewResponder(dataServiceContext.Response(), req)

	create := &work.Create{}
	if err := request.DecodeRequestBody(req.Request, create); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	wrk, err := dataServiceContext.WorkClient().Create(ctx, create)
	if err != nil {
		responder.InternalServerError(err)
		return
	}
	if wrk == nil { // deduplicated: an equivalent item already waits
		responder.Empty(http.StatusNoContent)
		return
	}

	responder.Data(http.StatusCreated, wrk)
}

func GetWork(dataServiceContext dataService.Context) {
	req := dataServiceContext.Request()
	ctx := req.Context()
	responder := request.MustNewResponder(dataServiceContext.Response(), req)

	id := req.PathParam("id")
	if id == "" {
		responder.Error(http.StatusBadRequest, request.ErrorParameterMissing("id"))
		return
	}

	wrk, err := dataServiceContext.WorkClient().Get(ctx, id, nil)
	if err != nil {
		responder.InternalServerError(err)
		return
	}
	if wrk == nil {
		responder.Error(http.StatusNotFound, request.ErrorResourceNotFoundWithID(id))
		return
	}

	responder.Data(http.StatusOK, wrk)
}
