package workTestLoadService

import (
	"fmt"
	"net/http"
	"time"

	dataService "github.com/tidepool-org/platform/data/service"
	"github.com/tidepool-org/platform/request"
	workTestLoad "github.com/tidepool-org/platform/work/test/load"
)

func LoadTestRoutes() []dataService.Route {
	return []dataService.Route{
		dataService.Get("/v1/work/load/status", statusHandler),
		dataService.Post("/v1/work/load", createHandler),
		dataService.Delete("/v1/work/load/:groupId", deleteHandler),
	}
}

func statusHandler(dataServiceContext dataService.Context) {
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()
	responder := request.MustNewResponder(res, req)
	responder.Data(http.StatusOK, map[string]any{"status": "All good ready for load test of work"})
}

func deleteHandler(dataServiceContext dataService.Context) {
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()
	ctx := req.Context()

	responder := request.MustNewResponder(res, req)

	groupID := req.PathParam("groupId")
	if groupID == "" {
		responder.Error(http.StatusBadRequest, request.ErrorParameterMissing("groupId"))
		return
	}

	if _, err := dataServiceContext.WorkClient().DeleteAllByGroupID(ctx, groupID); err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Empty(http.StatusOK)
}

func createHandler(dataServiceContext dataService.Context) {
	res := dataServiceContext.Response()
	req := dataServiceContext.Request()
	ctx := req.Context()

	start := time.Now()

	responder := request.MustNewResponder(res, req)

	items := []workTestLoad.LoadItem{}
	if err := request.DecodeRequestBody(req.Request, &items); err != nil {
		responder.Error(http.StatusBadRequest, fmt.Errorf("error decoding work load test request body %w", err))
		return
	}

	created := []any{}
	errored := []any{}

	for _, item := range items {
		item.Create.ProcessingAvailableTime = start.Add(time.Millisecond * time.Duration(item.OffsetMilliseconds))
		if wrk, err := dataServiceContext.WorkClient().Create(ctx, item.Create); err != nil {
			errored = append(errored, err)
			continue
		} else {
			created = append(created, wrk)
		}
	}

	responder.Data(http.StatusCreated, map[string]any{
		"created": created,
		"errored": errored,
	})
}
