package customerio

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/request"
)

type Client struct {
	appClient   *client.Client
	trackClient *client.Client
	config      Config
	logger      log.Logger
	httpClient  *http.Client
}

type Config struct {
	AppAPIBaseURL   string `envconfig:"TIDEPOOL_CUSTOMERIO_APP_API_BASE_URL" default:"https://api.customer.io"`
	AppAPIKey       string `envconfig:"TIDEPOOL_CUSTOMERIO_APP_API_KEY"`
	SiteID          string `envconfig:"TIDEPOOL_CUSTOMERIO_SITE_ID"`
	TrackAPIBaseURL string `envconfig:"TIDEPOOL_CUSTOMERIO_TRACK_API_BASE_URL" default:"https://track.customer.io"`
	TrackAPIKey     string `envconfig:"TIDEPOOL_CUSTOMERIO_TRACK_API_KEY"`
}

func NewClient(config Config, logger log.Logger) (*Client, error) {
	errorParser := newErrorResponseParser()

	appClient, err := client.NewWithErrorParser(&client.Config{
		Address: config.AppAPIBaseURL,
	}, errorParser)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create app API client")
	}

	trackClient, err := client.NewWithErrorParser(&client.Config{
		Address: config.TrackAPIBaseURL,
	}, errorParser)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create track API client")
	}

	return &Client{
		appClient:   appClient,
		trackClient: trackClient,
		config:      config,
		logger:      logger,
		httpClient:  http.DefaultClient,
	}, nil
}

// appAPIAuthMutator returns a request mutator for App API authentication (Bearer token)
func (c *Client) appAPIAuthMutator() *request.HeaderMutator {
	return request.NewHeaderMutator("Authorization", fmt.Sprintf("Bearer %s", c.config.AppAPIKey))
}

// trackAPIAuthMutator returns a request mutator for Track API authentication (Basic auth)
func (c *Client) trackAPIAuthMutator() *request.HeaderMutator {
	auth := c.config.SiteID + ":" + c.config.TrackAPIKey
	basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
	return request.NewHeaderMutator("Authorization", basicAuth)
}

// errorResponseParser implements client.ErrorResponseParser for Customer.io API errors
type errorResponseParser struct{}

func newErrorResponseParser() *errorResponseParser {
	return &errorResponseParser{}
}

func (p *errorResponseParser) ParseErrorResponse(ctx context.Context, res *http.Response, req *http.Request) error {
	var errResp errorResponse
	if err := json.NewDecoder(res.Body).Decode(&errResp); err != nil {
		return nil
	}

	if len(errResp.Errors) > 0 {
		return errors.Newf("API error (status %d): %s", res.StatusCode, errResp.Errors[0].Message)
	}

	return nil
}

type errorResponse struct {
	Errors []struct {
		Reason  string `json:"reason,omitempty"`
		Field   string `json:"field,omitempty"`
		Message string `json:"message,omitempty"`
	} `json:"errors,omitempty"`
}
