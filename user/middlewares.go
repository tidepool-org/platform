package user

import (
	"net/http"

	log "github.com/tidepool-org/platform/logger"

	"github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/ant0ine/go-json-rest/rest"
)

//Interface that all middleware components implement
type MiddleWare interface {
	MiddlewareFunc(h rest.HandlerFunc) rest.HandlerFunc
}

//Authorization middleware is used for validation of incoming tokens
type AuthorizationMiddleware struct {
	Client Client
}

func NewAuthorizationMiddleware(userClient Client) *AuthorizationMiddleware {
	return &AuthorizationMiddleware{Client: userClient}
}

//Valid - then we continue
//Invalid - then we return 401 (http.StatusUnauthorized)
func (mw *AuthorizationMiddleware) MiddlewareFunc(h rest.HandlerFunc) rest.HandlerFunc {

	return func(w rest.ResponseWriter, r *rest.Request) {

		token := r.Header.Get(x_tidepool_session_token)

		if tokenData := mw.Client.CheckToken(token); tokenData != nil {
			log.Logging.Info("token", token)
			h(w, r)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
}
