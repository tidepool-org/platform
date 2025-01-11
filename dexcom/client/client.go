package client

import (
	"context"
	"time"

	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/dexcom"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/oauth"
	oauthClient "github.com/tidepool-org/platform/oauth/client"
	"github.com/tidepool-org/platform/request"
)

type Client struct {
	client *oauthClient.Client
}

func New(cfg *client.Config, tknSrcSrc oauth.TokenSourceSource) (*Client, error) {
	baseClient, err := client.New(cfg)
	if err != nil {
		return nil, err
	}

	clnt, err := oauthClient.New(baseClient, tknSrcSrc)
	if err != nil {
		return nil, err
	}

	return &Client{
		client: clnt,
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

func (c *Client) sendDexcomRequestWithDataRange(ctx context.Context, startTime time.Time, endTime time.Time, method string, url string, responseBody interface{}, tokenSource oauth.TokenSource) error {
	url = c.client.AppendURLQuery(url, map[string]string{
		"startDate": startTime.UTC().Format(dexcom.DateRangeTimeFormat),
		"endDate":   endTime.UTC().Format(dexcom.DateRangeTimeFormat),
	})
	return c.sendDexcomRequest(ctx, method, url, responseBody, tokenSource)
}

func (c *Client) sendDexcomRequest(ctx context.Context, method string, url string, responseBody interface{}, tokenSource oauth.TokenSource) error {
	startTime := time.Now()

	err := c.client.SendOAuthRequest(ctx, method, url, nil, nil, responseBody, []request.ResponseInspector{prometheusCodePathResponseInspector}, tokenSource)

	if requestDuration := time.Since(startTime); requestDuration > requestDurationMaximum {
		log.LoggerFromContext(ctx).WithField("requestDuration", requestDuration.Truncate(time.Millisecond).Seconds()).Warn("Request duration exceeds maximum")
	}

	return err
}

const requestDurationMaximum = 30 * time.Second

var prometheusCodePathResponseInspector = request.NewPrometheusCodePathResponseInspector("tidepool_dexcom_api_client_requests", "Dexcom API client requests")
