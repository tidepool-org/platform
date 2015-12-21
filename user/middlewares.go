package user

import (
	"net/http"

	"github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/ant0ine/go-json-rest/rest"
)

type AuthorizationMiddleware struct {
	Client Client
}

func (mw *AuthorizationMiddleware) MiddlewareFunc(h rest.HandlerFunc) rest.HandlerFunc {

	return func(w rest.ResponseWriter, r *rest.Request) {

		token := r.Header.Get(x_tidepool_session_token)

		if tokenData := mw.Client.CheckToken(token); tokenData != nil {
			h(w, r)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
}
