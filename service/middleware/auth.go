package middleware

import (
	"strings"

	"github.com/mdblp/go-json-rest/rest"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/crypto"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service"
)

type Auth struct {
	serviceSecret string
	authClient    auth.Client
}

func NewAuth(serviceSecret string, authClient auth.Client) (*Auth, error) {
	if serviceSecret == "" {
		return nil, errors.New("service secret is missing")
	}
	if authClient == nil {
		return nil, errors.New("auth client is missing")
	}

	return &Auth{
		serviceSecret: serviceSecret,
		authClient:    authClient,
	}, nil
}

func (a *Auth) MiddlewareFunc(handlerFunc rest.HandlerFunc) rest.HandlerFunc {
	return func(res rest.ResponseWriter, req *rest.Request) {
		if handlerFunc != nil && res != nil && req != nil {
			oldRequest := req.Request
			defer func() {
				req.Request = oldRequest
			}()

			lgr := log.LoggerFromContext(req.Context())

			if serverSessionToken, err := a.authClient.ServerSessionToken(); err == nil {
				req.Request = req.WithContext(auth.NewContextWithServerSessionToken(req.Context(), serverSessionToken))
			} else {
				lgr.WithError(err).Warn("Unable to obtain server session token in auth middleware")
			}

			if details, err := a.authenticate(req); err != nil {
				// TODO: Sleep exponential fallback based upon IP and occurrences in period
				request.MustNewResponder(res, req).Error(request.StatusCodeForError(err), err)
				return
			} else if details != nil {
				// DEPRECATED - old context mechanism
				oldDetails := service.GetRequestAuthDetails(req)
				defer service.SetRequestAuthDetails(req, oldDetails)
				service.SetRequestAuthDetails(req, details)
				if details.HasToken() {
					if reqLgr := service.GetRequestLogger(req); reqLgr != nil {
						defer service.SetRequestLogger(req, reqLgr)
						service.SetRequestLogger(req, reqLgr.WithField("tokenHash", crypto.HexEncodedMD5Hash(details.Token())))
					}
				}

				req.Request = req.WithContext(request.NewContextWithDetails(req.Context(), details))
				if details.HasToken() {
					req.Request = req.WithContext(log.NewContextWithLogger(req.Context(), lgr.WithField("tokenHash", crypto.HexEncodedMD5Hash(details.Token()))))
				}
			}

			handlerFunc(res, req)
		}
	}
}

func (a *Auth) authenticate(req *rest.Request) (request.Details, error) {
	details, err := a.authenticateServiceSecret(req)
	if err != nil || details != nil {
		return details, err
	}

	details, err = a.authenticateAccessToken(req)
	if err != nil || details != nil {
		return details, err
	}

	return a.authenticateSessionToken(req)
}

func (a *Auth) authenticateServiceSecret(req *rest.Request) (request.Details, error) {
	values, found := req.Header[auth.TidepoolServiceSecretHeaderKey]
	if !found {
		return nil, nil
	} else if len(values) != 1 {
		return nil, request.ErrorUnauthorized()
	}

	if values[0] != a.serviceSecret {
		return nil, request.ErrorUnauthorized()
	}

	return request.NewDetails(request.MethodServiceSecret, "", ""), nil
}

func (a *Auth) authenticateAccessToken(req *rest.Request) (request.Details, error) {
	values, found := req.Header[auth.TidepoolAuthorizationHeaderKey]
	if !found {
		return nil, nil
	} else if len(values) != 1 {
		return nil, request.ErrorUnauthorized()
	}

	parts := strings.SplitN(values[0], " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return nil, request.ErrorUnauthorized()
	}

	details, err := a.authClient.ValidateSessionToken(req.Context(), parts[1])
	if err != nil {
		return nil, nil
	}

	return request.NewDetails(request.MethodAccessToken, details.UserID(), details.Token()), nil
}

func (a *Auth) authenticateSessionToken(req *rest.Request) (request.Details, error) {
	values, found := req.Header[auth.TidepoolSessionTokenHeaderKey]
	if !found {
		return nil, nil
	} else if len(values) != 1 {
		return nil, request.ErrorUnauthorized()
	}

	details, err := a.authClient.ValidateSessionToken(req.Context(), values[0])
	if err != nil {
		return nil, nil
	}

	return details, nil
}
