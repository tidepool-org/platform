package client

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/dexcom"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/oauth"
	oauthClient "github.com/tidepool-org/platform/oauth/client"
	"github.com/tidepool-org/platform/request"
)

type Client struct {
	client        *oauthClient.Client
	isSandboxData bool
}

func New(cfg *client.Config, tknSrcSrc oauth.TokenSourceSource) (*Client, error) {
	clnt, err := oauthClient.New(cfg, tknSrcSrc)
	if err != nil {
		return nil, err
	}

	isSandboxData := false
	if cfg != nil && cfg.Address == "https://sandbox-api.dexcom.com" {
		isSandboxData = true
	}

	return &Client{
		client:        clnt,
		isSandboxData: isSandboxData,
	}, nil
}

func (c *Client) GetAlerts(ctx context.Context, startTime time.Time, endTime time.Time, tokenSource oauth.TokenSource) (*dexcom.AlertsResponse, error) {
	alertsResponse := &dexcom.AlertsResponse{}
	paths := []string{"v3", "users", "self", "alerts"}

	if err := c.sendDexcomRequest(ctx, startTime, endTime, "GET", c.client.ConstructURL(paths...), alertsResponse, tokenSource); err != nil {
		return nil, errors.Wrap(err, "unable to get alerts")
	}

	return alertsResponse, nil
}

func (c *Client) GetCalibrations(ctx context.Context, startTime time.Time, endTime time.Time, tokenSource oauth.TokenSource) (*dexcom.CalibrationsResponse, error) {
	calibrationsResponse := &dexcom.CalibrationsResponse{}
	paths := []string{"v3", "users", "self", "calibrations"}

	if err := c.sendDexcomRequest(ctx, startTime, endTime, "GET", c.client.ConstructURL(paths...), calibrationsResponse, tokenSource); err != nil {
		return nil, errors.Wrap(err, "unable to get calibrations")
	}

	return calibrationsResponse, nil
}

func (c *Client) GetDevices(ctx context.Context, startTime time.Time, endTime time.Time, tokenSource oauth.TokenSource) (*dexcom.DevicesResponse, error) {
	devicesResponse := &dexcom.DevicesResponse{IsSandboxData: c.isSandboxData}
	paths := []string{"v3", "users", "self", "devices"}

	if err := c.sendDexcomRequest(ctx, startTime, endTime, "GET", c.client.ConstructURL(paths...), devicesResponse, tokenSource); err != nil {
		return nil, errors.Wrap(err, "unable to get devices")
	}

	return devicesResponse, nil
}

func (c *Client) GetEGVs(ctx context.Context, startTime time.Time, endTime time.Time, tokenSource oauth.TokenSource) (*dexcom.EGVsResponse, error) {
	egvsResponse := &dexcom.EGVsResponse{}
	paths := []string{"v3", "users", "self", "egvs"}

	if err := c.sendDexcomRequest(ctx, startTime, endTime, "GET", c.client.ConstructURL(paths...), egvsResponse, tokenSource); err != nil {
		return nil, errors.Wrap(err, "unable to get egvs")
	}

	return egvsResponse, nil
}

func (c *Client) GetEvents(ctx context.Context, startTime time.Time, endTime time.Time, tokenSource oauth.TokenSource) (*dexcom.EventsResponse, error) {
	eventsResponse := &dexcom.EventsResponse{}
	paths := []string{"v3", "users", "self", "events"}
	if err := c.sendDexcomRequest(ctx, startTime, endTime, "GET", c.client.ConstructURL(paths...), eventsResponse, tokenSource); err != nil {
		return nil, errors.Wrap(err, "unable to get events")
	}
	return eventsResponse, nil
}

func (c *Client) sendDexcomRequest(ctx context.Context, startTime time.Time, endTime time.Time, method string, url string, responseBody interface{}, tokenSource oauth.TokenSource) error {
	now := time.Now()

	url = c.client.AppendURLQuery(url, map[string]string{
		"startDate": startTime.UTC().Format(dexcom.DateRangeTimeFormat),
		"endDate":   endTime.UTC().Format(dexcom.DateRangeTimeFormat),
	})

	err := c.sendRequest(ctx, method, url, nil, nil, responseBody, tokenSource)
	if oauth.IsAccessTokenError(err) {
		tokenSource.ExpireToken()
		err = c.sendRequest(ctx, method, url, nil, nil, responseBody, tokenSource)
	}
	if oauth.IsRefreshTokenError(err) {
		err = errors.Wrap(request.ErrorUnauthenticated(), err.Error())
	}

	if requestDuration := time.Since(now); requestDuration > requestDurationMaximum {
		log.LoggerFromContext(ctx).WithField("requestDuration", requestDuration.Truncate(time.Millisecond).Seconds()).Warn("Request duration exceeds maximum")
	}

	return err
}

const requestDurationMaximum = 15 * time.Second

// sendRequest adds instrumentation before calling oauth.Client.SendOAuthRequest.
func (c *Client) sendRequest(ctx context.Context, method, url string, mutators []request.RequestMutator,
	requestBody any, responseBody any, httpClientSource oauth.HTTPClientSource) error {

	var inspectors = []request.ResponseInspector{
		&promDexcomInstrumentor{},
	}
	return c.client.SendOAuthRequest(ctx, method, url, mutators, requestBody, responseBody, inspectors, httpClientSource)
}

type promDexcomInstrumentor struct{}

// InspectResponse implements request.ResponseInspector.
func (i *promDexcomInstrumentor) InspectResponse(r *http.Response) {
	labels := prometheus.Labels{
		"code": strconv.Itoa(r.StatusCode),
		"path": r.Request.URL.Path,
	}
	promDexcomCounter.With(labels).Inc()
}

// promDexcomCounter instruments the Dexcom API paths and status codes called.
var promDexcomCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "tidepool_dexcom_api_client_requests",
	Help: "Dexcom API client requests",
}, []string{"code", "path"})
