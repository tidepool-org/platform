package v1

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service/api"
	"github.com/tidepool-org/platform/task"
	"github.com/tidepool-org/platform/task/service"
)

type Router struct {
	service.Service
}

func NewRouter(svc service.Service) (*Router, error) {
	if svc == nil {
		return nil, errors.New("service is missing")
	}

	return &Router{
		Service: svc,
	}, nil
}

func (r *Router) Routes() []*rest.Route {
	return []*rest.Route{
		rest.Get("/v1/tasks", api.RequireServer(r.ListTasks)),
		rest.Post("/v1/tasks", api.RequireServer(r.CreateTask)),
		rest.Get("/v1/tasks/:id", api.RequireServer(r.GetTask)),
		rest.Put("/v1/tasks/:id", api.RequireServer(r.UpdateTask)),
		rest.Delete("/v1/tasks/:id", api.RequireServer(r.DeleteTask)),
	}
}

func (r *Router) ListTasks(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	filter := task.NewTaskFilter()
	pagination := page.NewPagination()
	if err := request.DecodeRequestQuery(req.Request, filter, pagination); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	tsks, err := r.TaskClient().ListTasks(req.Context(), filter, pagination)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Data(http.StatusOK, tsks)
}

func (r *Router) CreateTask(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	create := task.NewTaskCreate()
	if err := request.DecodeRequestBody(req.Request, create); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	tsk, err := r.TaskClient().CreateTask(req.Context(), create)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Data(http.StatusCreated, tsk)
}

func (r *Router) GetTask(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	id := req.PathParam("id")
	if id == "" {
		responder.Error(http.StatusBadRequest, request.ErrorParameterMissing("id"))
		return
	}

	tsk, err := r.TaskClient().GetTask(req.Context(), id)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	} else if tsk == nil {
		responder.Error(http.StatusNotFound, request.ErrorResourceNotFoundWithID(id))
		return
	}

	responder.Data(http.StatusOK, tsk)
}

func (r *Router) UpdateTask(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	id := req.PathParam("id")
	if id == "" {
		responder.Error(http.StatusBadRequest, request.ErrorParameterMissing("id"))
		return
	}

	update := task.NewTaskUpdate()
	if err := request.DecodeRequestBody(req.Request, update); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	tsk, err := r.TaskClient().UpdateTask(req.Context(), id, update)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	} else if tsk == nil {
		responder.Error(http.StatusNotFound, request.ErrorResourceNotFoundWithID(id))
		return
	}

	responder.Data(http.StatusOK, tsk)
}

func (r *Router) DeleteTask(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	id := req.PathParam("id")
	if id == "" {
		responder.Error(http.StatusBadRequest, request.ErrorParameterMissing("id"))
		return
	}

	err := r.TaskClient().DeleteTask(req.Context(), id)
	if err != nil {
		responder.Error(http.StatusInternalServerError, err)
		return
	}

	responder.Empty(http.StatusOK)
}
