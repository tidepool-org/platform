package user

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"
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

//MetadataMiddleware middleware is used for getting user permissons
type MetadataMiddleware struct {
	Client Client
}

//PERMISSIONS constant for accessing users permissions that are attached to request.Env
const PERMISSIONS = "permissons"

//GROUPID constant for accessing users groupID that are attached to request.Env
const GROUPID = "groupID"

//NewMetadataMiddleware creates initialised MetadataMiddleware
func NewMetadataMiddleware(userClient Client) *MetadataMiddleware {
	return &MetadataMiddleware{Client: userClient}
}

//GetPermissons attach's permissons if they exist
//http.StatusInternalServerError if there is an error getting the user permissons
func (mw *MetadataMiddleware) GetPermissons(h rest.HandlerFunc) rest.HandlerFunc {

	return func(w rest.ResponseWriter, r *rest.Request) {

		userid := r.PathParam("userid")

		permissions, err := mw.Client.GetUserPermissons(userid)
		if err != nil {
			rest.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		r.Env[PERMISSIONS] = permissions
		h(w, r)
		return

	}
}

//GetGroupID attach's the users groupId
//http.StatusInternalServerError if there is an error getting the groupID
//http.StatusBadRequest if there is no groupID found for the given userID
func (mw *MetadataMiddleware) GetGroupID(h rest.HandlerFunc) rest.HandlerFunc {

	return func(w rest.ResponseWriter, r *rest.Request) {

		userid := r.PathParam("userid")

		groupID, err := mw.Client.GetUserGroupID(userid)
		if err != nil {
			rest.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if groupID == "" {
			rest.Error(w, "no groupID found for user", http.StatusBadRequest)
			return
		}

		r.Env[GROUPID] = groupID
		h(w, r)
		return

	}
}
