package v1

import (
	stdErrs "errors"
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service/api"
	"github.com/tidepool-org/platform/user"
)

func (r *Router) ProfileRoutes() []*rest.Route {
	return []*rest.Route{
		rest.Get("/v1/profiles/:userId", api.RequireUser(r.GetProfile)),
		rest.Put("/v1/profiles/:userId", api.RequireUser(r.UpdateProfile)),
	}
}

func (r *Router) GetProfile(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)
	ctx := req.Context()
	details := request.GetAuthDetails(ctx)
	userID := req.PathParam("userId")
	if !details.IsService() && details.UserID() != userID {
		responder.Empty(http.StatusNotFound)
		return
	}

	user, err := r.UserAccessor().FindUserById(ctx, userID)
	if err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}

	responder.Data(http.StatusOK, user.Profile)
}

func (r *Router) UpdateProfile(res rest.ResponseWriter, req *rest.Request) {
	responder := request.MustNewResponder(res, req)
	ctx := req.Context()
	details := request.GetAuthDetails(ctx)
	userID := req.PathParam("userId")
	if !details.IsService() && details.UserID() != userID {
		responder.Empty(http.StatusNotFound)
		return
	}

	profile := &user.UserProfile{}
	if err := request.DecodeRequestBody(req.Request, profile); err != nil {
		responder.Error(http.StatusBadRequest, err)
		return
	}
	err := r.UserAccessor().UpdateUserProfile(ctx, userID, profile)
	if stdErrs.Is(err, user.ErrUserNotFound) {
		responder.Empty(http.StatusNotFound)
		return
	}
	if err != nil {
		responder.InternalServerError(err)
		return
	}
	responder.Empty(http.StatusOK)
}
