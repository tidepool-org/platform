package user

import (
	"net/http"

	"github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/ant0ine/go-json-rest/rest"
)

//ChainedMiddleware used for join Middleware function calls
type ChainedMiddleware func(rest.HandlerFunc) rest.HandlerFunc

//AuthorizationMiddleware is used for validation of incoming tokens
type AuthorizationMiddleware struct {
	Client Client
}

//NewAuthorizationMiddleware creates an initialised AuthorizationMiddleware
func NewAuthorizationMiddleware(userClient Client) *AuthorizationMiddleware {
	return &AuthorizationMiddleware{Client: userClient}
}

//ValidateToken returns if valid or sends http.StatusUnauthorized if invalid
func (mw *AuthorizationMiddleware) ValidateToken(h rest.HandlerFunc) rest.HandlerFunc {

	return func(w rest.ResponseWriter, r *rest.Request) {

		token := r.Header.Get(xTidepoolSessionToken)
		userid := r.PathParam("userid")

		if tokenData := mw.Client.CheckToken(token); tokenData != nil {
			if tokenData.IsServer || tokenData.UserID == userid {
				h(w, r)
				return
			}
			log.Info("id's don't match and not server token", tokenData.UserID, userid)
		}
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
}

//PermissonsMiddleware middleware is used for getting user permissons
type PermissonsMiddleware struct {
	Client Client
}

//PERMISSIONS constant for accessing permissions that are attached to request.Env
const PERMISSIONS = "PERMISSIONS"

//NewPermissonsMiddleware creates initialised PermissonsMiddleware
func NewPermissonsMiddleware(userClient Client) *PermissonsMiddleware {
	return &PermissonsMiddleware{Client: userClient}
}

//GetPermissons attach's permissons if they exist
//http.StatusInternalServerError if there is an error getting the user permissons
func (mw *PermissonsMiddleware) GetPermissons(h rest.HandlerFunc) rest.HandlerFunc {

	return func(w rest.ResponseWriter, r *rest.Request) {

		token := r.Header.Get(xTidepoolSessionToken)
		userid := r.PathParam("userid")

		permissions, err := mw.Client.GetUserPermissons(userid, token)
		if err != nil {
			rest.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		r.Env[PERMISSIONS] = permissions
		h(w, r)
		return

	}
}
