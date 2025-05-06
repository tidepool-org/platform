package service

import (
	"fmt"
	"net/http"
	"time"

	dataService "github.com/tidepool-org/platform/data/service"
	request "github.com/tidepool-org/platform/request"
	workLoad "github.com/tidepool-org/platform/work/load"
)

func LoadTestRoutes() []dataService.Route {
	return []dataService.Route{
		dataService.Get("/v1/work/load/ok", okHandler),
		dataService.Post("/v1/work/load", createHandler),
		dataService.Delete("/v1/work/load/:groupId", deleteHandler),
	}
}

func okHandler(dataServiceContext dataService.Context) {
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()
	responder := request.MustNewResponder(res, req)
	responder.Data(http.StatusOK, map[string]any{"status": "All good ready for work load"})
}

func deleteHandler(dataServiceContext dataService.Context) {
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

func createHandler(dataServiceContext dataService.Context) {
	ctx := dataServiceContext.Request().Context()
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()

	start := time.Now()

	responder := request.MustNewResponder(res, req)

	items := []workLoad.LoadItem{}
	if err := request.DecodeRequestBody(req.Request, &items); err != nil {
		responder.Error(http.StatusBadRequest, fmt.Errorf("error decodeing request body %s", err.Error()))
		return
	}

	created := []any{}

	for _, item := range items {
		item.Create.ProcessingAvailableTime = start.Add(time.Second * time.Duration(item.SecondsOffsetFromStart))

		wc, err := workLoad.NewLoadWorkCreate(item.Create)
		if err != nil {
			responder.Error(http.StatusBadRequest, fmt.Errorf("error creating work.Create %s", err.Error()))
			return
		}
		if wrk, err := dataServiceContext.WorkClient().Create(ctx, wc); err != nil {
			responder.Error(http.StatusBadRequest, fmt.Errorf("error creating work.Work %s", err.Error()))
			return
		} else if wrk != nil {
			created = append(created, wrk)
		}
	}

	responder.Data(http.StatusCreated, created)
}
