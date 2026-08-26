package client

import (
	"context"
	"net/http"
	"time"

	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/dexcom"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/oauth"
	oauthClient "github.com/tidepool-org/platform/oauth/client"
	"github.com/tidepool-org/platform/request"
)

const (
	RetrierRetries = 4
	RetrierDelay   = 2 * time.Second
	RetrierJitter  = 0.1
)

var Retrier = request.NewRetrier(RetrierRetries, RetrierDelay, RetrierJitter)

type Client struct {
	client  *oauthClient.Client
	retrier request.Retrier
}

func New(cfg *client.Config, httpClient *http.Client, tknSrcSrc oauth.TokenSourceSource, retrier request.Retrier) (*Client, error) {
	if cfg == nil {
		return nil, errors.New("config is missing")
	} else if err := cfg.Validate(); err != nil {
		return nil, errors.Wrap(err, "config is invalid")
	}
	if tknSrcSrc == nil {
		return nil, errors.New("token source source is missing")
	}
	if retrier == nil {
		return nil, errors.New("retrier is missing")
	}

	baseClient, err := client.New(cfg)
	if err != nil {
		return nil, err
	}

	clnt, err := oauthClient.NewWithClient(baseClient, httpClient, tknSrcSrc)
	if err != nil {
		return nil, err
	}

	return &Client{
		client:  clnt,
		retrier: retrier,
	}, nil
}

func (c *Client) GetDataRange(ctx context.Context, lastSyncTime *time.Time, tokenSource oauth.TokenSource) (*dexcom.DataRangesResponse, error) {
	dataRangeResponse := &dexcom.DataRangesResponse{}
	paths := []string{"v3", "users", "self", "dataRange"}

	url := c.client.ConstructURL(paths...)
	if lastSyncTime != nil {
		url = c.client.AppendURLQuery(url, map[string]string{
			"lastSyncTime": lastSyncTime.UTC().Format(time.RFC3339), // NOTE: Explicitly not normal Dexcom time format (Dexcom API requires timezone offset)
		})
	}

	if err := c.sendDexcomRequest(ctx, "GET", url, dataRangeResponse, tokenSource); err != nil {
		return nil, errors.Wrap(err, "unable to get data range")
	}

	return dataRangeResponse, nil
}

func (c *Client) GetAlerts(ctx context.Context, startTime time.Time, endTime time.Time, tokenSource oauth.TokenSource) (*dexcom.AlertsResponse, error) {
	alertsResponse := &dexcom.AlertsResponse{}
	paths := []string{"v3", "users", "self", "alerts"}

	if err := c.sendDexcomRequestWithDataRange(ctx, startTime, endTime, "GET", c.client.ConstructURL(paths...), alertsResponse, tokenSource); err != nil {
		return nil, errors.Wrap(err, "unable to get alerts")
	}

	return alertsResponse, nil
}

func (c *Client) GetCalibrations(ctx context.Context, startTime time.Time, endTime time.Time, tokenSource oauth.TokenSource) (*dexcom.CalibrationsResponse, error) {
	calibrationsResponse := &dexcom.CalibrationsResponse{}
	paths := []string{"v3", "users", "self", "calibrations"}

	if err := c.sendDexcomRequestWithDataRange(ctx, startTime, endTime, "GET", c.client.ConstructURL(paths...), calibrationsResponse, tokenSource); err != nil {
		return nil, errors.Wrap(err, "unable to get calibrations")
	}

	return calibrationsResponse, nil
}

func (c *Client) GetDevices(ctx context.Context, startTime time.Time, endTime time.Time, tokenSource oauth.TokenSource) (*dexcom.DevicesResponse, error) {
	devicesResponse := &dexcom.DevicesResponse{}
	paths := []string{"v3", "users", "self", "devices"}

	if err := c.sendDexcomRequestWithDataRange(ctx, startTime, endTime, "GET", c.client.ConstructURL(paths...), devicesResponse, tokenSource); err != nil {
		return nil, errors.Wrap(err, "unable to get devices")
	}

	return devicesResponse, nil
}

func (c *Client) GetEGVs(ctx context.Context, startTime time.Time, endTime time.Time, tokenSource oauth.TokenSource) (*dexcom.EGVsResponse, error) {
	egvsResponse := &dexcom.EGVsResponse{}
	paths := []string{"v3", "users", "self", "egvs"}

	if err := c.sendDexcomRequestWithDataRange(ctx, startTime, endTime, "GET", c.client.ConstructURL(paths...), egvsResponse, tokenSource); err != nil {
		return nil, errors.Wrap(err, "unable to get egvs")
	}

	return egvsResponse, nil
}

func (c *Client) GetEvents(ctx context.Context, startTime time.Time, endTime time.Time, tokenSource oauth.TokenSource) (*dexcom.EventsResponse, error) {
	eventsResponse := &dexcom.EventsResponse{}
	paths := []string{"v3", "users", "self", "events"}

	if err := c.sendDexcomRequestWithDataRange(ctx, startTime, endTime, "GET", c.client.ConstructURL(paths...), eventsResponse, tokenSource); err != nil {
		return nil, errors.Wrap(err, "unable to get events")
	}
	return eventsResponse, nil
}

func (c *Client) sendDexcomRequestWithDataRange(ctx context.Context, startTime time.Time, endTime time.Time, method string, url string, responseBody any, tokenSource oauth.TokenSource) error {
	url = c.client.AppendURLQuery(url, map[string]string{
		"startDate": startTime.UTC().Format(dexcom.DateRangeTimeFormat),
		"endDate":   endTime.UTC().Format(dexcom.DateRangeTimeFormat),
	})
	return c.sendDexcomRequest(ctx, method, url, responseBody, tokenSource)
}

func (c *Client) sendDexcomRequest(ctx context.Context, method string, url string, responseBody any, tokenSource oauth.TokenSource) error {
	if tokenSource == nil {
		return errors.New("token source is missing")
	}

	_, err := request.RetryFailure(ctx, c.retrier, func(ctx context.Context) (bool, error) {
		statusCodeInspector := request.NewStatusCodeInspector()
		if err := log.WarnIfDurationExceedsMaximum(ctx, requestDurationMaximum, url, func(ctx context.Context) error {
			return c.client.SendOAuthRequest(ctx, method, url, nil, nil, responseBody, []request.ResponseInspector{statusCodeInspector}, tokenSource)
		}); err != nil {
			return !isTransientFailure(err, statusCodeInspector.StatusCode), err
		} else {
			return true, nil
		}
	})

	return err
}

// isTransientFailure reports whether an identical retry could plausibly succeed. Authentication failures are excluded
// deliberately: Dexcom rotates its single-use refresh token on every refresh, so retrying one here would consume and
// re-persist that token repeatedly, and TaskRunner.handleDexcomClientError already retries them at task scope.
func isTransientFailure(err error, statusCode int) bool {
	switch {
	case request.IsErrorUnauthenticated(errors.LastCause(err)):
		return false
	case statusCode == 0: // No response, so a transport failure, timeout or cancellation
		return true
	default:
		return statusCode == http.StatusTooManyRequests || statusCode >= http.StatusInternalServerError
	}
}

const requestDurationMaximum = 30 * time.Second
