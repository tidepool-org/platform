package v1

import (
	"net/http"

	"github.com/mdblp/go-json-rest/rest"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/user"
)

type Provider interface {
	UserClient() user.Client
}

type Router struct {
	provider Provider
}

func NewRouter(provider Provider) (*Router, error) {
	if provider == nil {
		return nil, errors.New("provider is missing")
	}

	return &Router{
		provider: provider,
	}, nil
}

func (r *Router) Routes() []*rest.Route {
	return []*rest.Route{
		rest.Get("/v1/users/:id", r.Get),
		rest.Delete("/v1/users/:id", r.Delete),
	}
}

func (r *Router) Get(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	id, err := request.DecodeRequestPathParameter(req, "id", user.IsValidID)
	if err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	result, err := r.provider.UserClient().Get(req.Context(), id)
	if responder.RespondIfError(err) {
		return
	} else if result == nil {
		responder.Error(http.StatusNotFound, request.ErrorResourceNotFoundWithID(id))
		return
	}

	responder.Data(http.StatusOK, result)
}

func (r *Router) Delete(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)

	id, err := request.DecodeRequestPathParameter(req, "id", user.IsValidID)
	if err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	condition := request.NewCondition()
	if err = request.DecodeRequestQuery(req.Request, condition); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	deleet := user.NewDelete()
	if err = request.DecodeRequestBody(req.Request, deleet); err != nil {
		if errors.Code(err) == request.ErrorCodeJSONNotFound {
			deleet = nil
		} else {
			responder.Error(http.StatusBadRequest, err)
			return
		}
	}

	deleted, err := r.provider.UserClient().Delete(req.Context(), id, deleet, condition)
	if responder.RespondIfError(err) {
		return
	} else if !deleted {
		responder.Error(http.StatusNotFound, request.ErrorResourceNotFoundWithIDAndOptionalRevision(id, condition.Revision))
		return
	}

	responder.Empty(http.StatusNoContent)
}
