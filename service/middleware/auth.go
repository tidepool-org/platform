package middleware

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/mdblp/go-json-rest/rest"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/crypto"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service"
)

type Auth struct {
	serviceSecret  string
	authClient     auth.Client
	tokenValidator *validator.Validator
}

type OAuthCustomClaims struct {
	Scope    string   `json:"scope"`
	Roles    []string `json:"http://your-loops.com/roles"`
	IsServer bool     `json:"isServer"`
}

func (c OAuthCustomClaims) Validate(ctx context.Context) error {
	if len(c.Roles) == 0 {
		return errors.New("Roles not set in the access token")
	}
	return nil
}

func NewAuth(serviceSecret string, authClient auth.Client) (*Auth, error) {
	if serviceSecret == "" {
		return nil, errors.New("service secret is missing")
	}
	if authClient == nil {
		return nil, errors.New("auth client is missing")
	}
	validator, err := setupAuth0()
	if err != nil {
		return nil, err
	}

	return &Auth{
		serviceSecret:  serviceSecret,
		authClient:     authClient,
		tokenValidator: validator,
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

	return request.NewDetails(request.MethodServiceSecret, "", "", "server"), nil
}

func (a *Auth) authenticateAccessToken(req *rest.Request) (request.Details, error) {
	lgr := log.LoggerFromContext(req.Context())
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

	//Validate against auth0
	var parsedToken *validator.ValidatedClaims
	if t, err := a.tokenValidator.ValidateToken(req.Context(), parts[1]); err != nil {
		lgr.Error("Error decoding bearer token")
		return nil, request.ErrorUnauthorized()
	} else {
		parsedToken = t.(*validator.ValidatedClaims)
	}
	uid := strings.Split(parsedToken.RegisteredClaims.Subject, "|")[1]
	customClaims := parsedToken.CustomClaims.(*OAuthCustomClaims)
	return request.NewDetails(request.MethodAccessToken, uid, parts[1], customClaims.Roles[0]), nil
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

func setupAuth0() (*validator.Validator, error) {
	//target audience is used to verify the token was issued for a specific domain or url.
	//by default it will be empty but we would (in the future) use this to authorize or deny access to some urls
	targetAudience := []string{}
	if value, present := os.LookupEnv("AUTH0_AUDIENCE"); present {
		targetAudience = []string{value}
	}
	issuerURL, err := url.Parse(os.Getenv("AUTH0_URL") + "/")
	if err != nil {
		return nil, errors.New("Failed to parse the issuer url: " + err.Error())
	}
	var keyProvider *jwks.CachingProvider
	// Use a custom CA cert if it's provided
	if os.Getenv("SSL_CUSTOM_CA_KEY") != "" {
		keyProvider = jwks.NewCachingProvider(issuerURL, 5*time.Minute, WithCustomCA(os.Getenv("SSL_CUSTOM_CA_KEY")))
	} else {
		keyProvider = jwks.NewCachingProvider(issuerURL, 5*time.Minute)
	}
	jwtValidator, err := validator.New(
		keyProvider.KeyFunc,
		validator.RS256,
		issuerURL.String(),
		targetAudience,
		validator.WithCustomClaims(
			func() validator.CustomClaims {
				return &OAuthCustomClaims{}
			},
		),
		validator.WithAllowedClockSkew(time.Minute),
	)
	if err != nil {
		return nil, errors.New("Failed to set up the jwt validator: " + err.Error())
	}

	return jwtValidator, nil
}

// WithCustomCa is a Provider Option for our jwks CachingProvider
// It is used to specify a local CA cert, usefull when using a local OAuth server which use a self-signed cert
func WithCustomCA(pem string) jwks.ProviderOption {
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM([]byte(pem))

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs: certPool,
		},
	}

	return func(p *jwks.Provider) {
		p.Client.Transport = tr
	}
}
