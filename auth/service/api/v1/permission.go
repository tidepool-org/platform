package v1

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service/api"
)

// requireUserHasCustodian aborts with an error if a a request isn't
// authenticated as a user and the user does not have custodian access to the
// user with the id defined in the url param targetParamUserID
func (r *Router) requireUserHasCustodian(targetParamUserID string, handlerFunc rest.HandlerFunc) rest.HandlerFunc {
	fn := func(res rest.ResponseWriter, req *rest.Request) {
		if handlerFunc != nil && res != nil && req != nil {
			targetUserID := req.PathParam(targetParamUserID)
			responder := request.MustNewResponder(res, req)
			ctx := req.Context()
			details := request.GetAuthDetails(ctx)
			hasPerms, err := r.PermissionsClient().HasCustodianPermissions(ctx, details.UserID(), targetUserID)
			if err != nil {
				responder.InternalServerError(err)
				return
			}
			if !hasPerms {
				responder.Empty(http.StatusForbidden)
				return
			}
			handlerFunc(res, req)
		}
	}
	return api.RequireUser(fn)
}
