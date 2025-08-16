package v1

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/request"
)

// requireCustodian aborts with an error if the user associated w/ the
// request doesn't have custodian access to the user with the id defined in the
// url param targetParamUserID.
//
// This mimics the logic of amoeba's requireCustodian access. This means a
// user has access to the target user if any of the following is true:
//   - The is a service call (AuthDetails.IsService() == true)
//   - The requester and target are the same - AuthDetails.UserID == targetParamUserID
//   - The requester has explicit permissions to access targetParamUserID
func (r *Router) requireCustodian(targetParamUserID string, handlerFunc rest.HandlerFunc) rest.HandlerFunc {
	return func(res rest.ResponseWriter, req *rest.Request) {
		if handlerFunc != nil && res != nil && req != nil {
			targetUserID := req.PathParam(targetParamUserID)
			responder := request.MustNewResponder(res, req)
			ctx := req.Context()
			details := request.GetAuthDetails(ctx)
			if details == nil {
				request.MustNewResponder(res, req).Error(http.StatusUnauthorized, request.ErrorUnauthenticated())
				return
			}
			if details.IsService() || details.UserID() == targetUserID {
				handlerFunc(res, req)
				return
			}
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
}

// requireMembership proceeds if the user with the id specified in the URL
// paramter targetParamUserID has some association with the user in the current
// request - the "requester". This mimics amoeba's requireMembership function.
//
// This proceeds if any of the following are true:
//   - The is a service call (AuthDetails.IsService() == true)
//   - The requester and target are the same - AuthDetails.UserID == targetParamUserID
//   - The requester has any permissions to targetParamUserID OR targetParamUserID has permissions to the requester.
func (r *Router) requireMembership(targetParamUserID string, handlerFunc rest.HandlerFunc) rest.HandlerFunc {
	return func(res rest.ResponseWriter, req *rest.Request) {
		if handlerFunc != nil && res != nil && req != nil {
			targetUserID := req.PathParam(targetParamUserID)
			responder := request.MustNewResponder(res, req)
			ctx := req.Context()
			details := request.GetAuthDetails(ctx)
			if details == nil {
				request.MustNewResponder(res, req).Error(http.StatusUnauthorized, request.ErrorUnauthenticated())
				return
			}
			if details.IsService() || details.UserID() == targetUserID {
				handlerFunc(res, req)
				return
			}
			hasPerms, err := r.PermissionsClient().HasMembershipRelationship(ctx, details.UserID(), targetUserID)
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
}

// requireWriteAccess aborts with an error if the request isn't a server request
// or the authenticated user doesn't have access to the user id in the url param,
// targetParamUserID
func (r *Router) requireWriteAccess(targetParamUserID string, handlerFunc rest.HandlerFunc) rest.HandlerFunc {
	return func(res rest.ResponseWriter, req *rest.Request) {
		if handlerFunc != nil && res != nil && req != nil {
			targetUserID := req.PathParam(targetParamUserID)
			responder := request.MustNewResponder(res, req)
			ctx := req.Context()
			details := request.GetAuthDetails(ctx)
			if details == nil {
				responder.Empty(http.StatusUnauthorized)
				return
			}
			if details.IsService() {
				handlerFunc(res, req)
				return
			}
			hasPerms, err := r.PermissionsClient().HasWritePermissions(ctx, details.UserID(), targetUserID)
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
}
