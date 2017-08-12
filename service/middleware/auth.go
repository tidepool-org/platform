package middleware

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/auth"
	authContext "github.com/tidepool-org/platform/auth/context"
	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/crypto"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/service"
)

type Auth struct {
	authClient auth.Client
}

func NewAuth(authClient auth.Client) (*Auth, error) {
	if authClient == nil {
		return nil, errors.New("middleware", "auth client is missing")
	}

	return &Auth{
		authClient: authClient,
	}, nil
}

func (a *Auth) MiddlewareFunc(handlerFunc rest.HandlerFunc) rest.HandlerFunc {
	return func(response rest.ResponseWriter, request *rest.Request) {
		if handlerFunc != nil && response != nil && request != nil {
			if token := request.Header.Get(auth.TidepoolAuthTokenHeaderName); token != "" {
				if oldLogger := service.GetRequestLogger(request); oldLogger != nil {
					defer service.SetRequestLogger(request, oldLogger)
					service.SetRequestLogger(request, oldLogger.WithField("authTokenHash", crypto.HashWithMD5(token)))
				}

				context, err := authContext.New(response, request, a.authClient)
				if err != nil {
					response.WriteHeader(http.StatusInternalServerError)
					return
				}

				newAuthDetails, err := context.AuthClient().ValidateToken(context, token)
				if err != nil {
					if !client.IsUnauthorizedError(err) {
						context.RespondWithInternalServerFailure("Unable to validate token", err, token)
						return
					}
				} else {
					oldAuthDetails := service.GetRequestAuthDetails(request)
					defer service.SetRequestAuthDetails(request, oldAuthDetails)
					service.SetRequestAuthDetails(request, newAuthDetails)
				}
			}

			handlerFunc(response, request)
		}
	}
}
