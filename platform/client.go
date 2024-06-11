package platform

import (
	"context"
	"io"
	"net/http"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/request"
)

type AuthorizeAs int

const (
	AuthorizeAsService AuthorizeAs = iota
	AuthorizeAsUser
)

type Client struct {
	*client.Client
	authorizeAs   AuthorizeAs
	serviceSecret string
	httpClient    *http.Client
}

// Modern platform-based services encode all errors with detailed information in the
// response body. Deserialize those errors and pass onward.
func NewClient(cfg *Config, authorizeAs AuthorizeAs) (*Client, error) {
	return newClient(cfg, authorizeAs, client.NewSerializableErrorResponseParser())
}

// Legacy services (shoreline, highwater, gatekeeper, etc.) do not serialize their errors
// in the response body in a manner that platform can parse.
func NewLegacyClient(cfg *Config, authorizeAs AuthorizeAs) (*Client, error) {
	return newClient(cfg, authorizeAs, nil)
}

func newClient(cfg *Config, authorizeAs AuthorizeAs, errorResponseParser client.ErrorResponseParser) (*Client, error) {
	if cfg == nil {
		return nil, errors.New("config is missing")
	} else if err := cfg.Validate(); err != nil {
		return nil, errors.Wrap(err, "config is invalid")
	}
	if authorizeAs != AuthorizeAsService && authorizeAs != AuthorizeAsUser {
		return nil, errors.New("authorize as is invalid")
	}

	clnt, err := client.NewWithErrorParser(cfg.Config, errorResponseParser)
	if err != nil {
		return nil, err
	}

	// FUTURE: Use once all services support service secret
	// if authorizeAs == AuthorizeAsService {
	// 	if cfg.ServiceSecret == "" {
	// 		return errors.New("service secret is missing")
	// 	}
	// }

	return &Client{
		Client:        clnt,
		authorizeAs:   authorizeAs,
		serviceSecret: cfg.ServiceSecret,
		httpClient:    &http.Client{},
	}, nil
}

func (c *Client) IsAuthorizeAsService() bool {
	return c.authorizeAs == AuthorizeAsService
}

func (c *Client) Mutators(ctx context.Context) ([]request.RequestMutator, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}

	var authorizationMutator request.RequestMutator
	if c.IsAuthorizeAsService() {
		if c.serviceSecret != "" {
			authorizationMutator = NewServiceSecretHeaderMutator(c.serviceSecret)
		} else if serverSessionTokenProvider := auth.ServerSessionTokenProviderFromContext(ctx); serverSessionTokenProvider != nil {
			if serverSessionToken, err := serverSessionTokenProvider.ServerSessionToken(); err != nil {
				return nil, err
			} else {
				authorizationMutator = NewSessionTokenHeaderMutator(serverSessionToken)
			}
		} else {
			return nil, errors.New("service secret is missing")
		}
	} else {
		details := request.GetAuthDetails(ctx)
		if details == nil {
			return nil, errors.New("details is missing")
		}
		authorizationMutator = NewSessionTokenHeaderMutator(details.Token())
	}
	return []request.RequestMutator{authorizationMutator, NewTraceMutator(ctx)}, nil
}

func (c *Client) HTTPClient() *http.Client {
	return c.httpClient
}

func (c *Client) RequestStream(ctx context.Context, method string, url string, mutators []request.RequestMutator, requestBody interface{}, inspectors ...request.ResponseInspector) (io.ReadCloser, error) {
	clientMutators, err := c.Mutators(ctx)
	if err != nil {
		return nil, err
	}

	return c.RequestStreamWithHTTPClient(ctx, method, url, append(mutators, clientMutators...), requestBody, inspectors, c.HTTPClient())
}

func (c *Client) RequestData(ctx context.Context, method string, url string, mutators []request.RequestMutator, requestBody interface{}, responseBody interface{}, inspectors ...request.ResponseInspector) error {
	clientMutators, err := c.Mutators(ctx)
	if err != nil {
		return err
	}

	return c.RequestDataWithHTTPClient(ctx, method, url, append(mutators, clientMutators...), requestBody, responseBody, inspectors, c.HTTPClient())
}
