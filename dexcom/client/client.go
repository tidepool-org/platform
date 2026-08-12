package client

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/dexcom"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/oauth"
	oauthClient "github.com/tidepool-org/platform/oauth/client"
)

type Client struct {
	client *oauthClient.Client
}

func New(cfg *client.Config, tknSrcSrc oauth.TokenSourceSource) (*Client, error) {
	if cfg == nil {
		return nil, errors.New("config is missing")
	} else if err := cfg.Validate(); err != nil {
		return nil, errors.Wrap(err, "config is invalid")
	}

	if cfg.Timeout == 0 {
		cfg.Timeout = 1 * time.Minute
	}

	httpClient := &http.Client{
		Transport:     prometheusRequestMetricsRoundTripper,
		CheckRedirect: http.DefaultClient.CheckRedirect,
		Jar:           http.DefaultClient.Jar,
		Timeout:       http.DefaultClient.Timeout,
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
	return log.WarnIfDurationExceedsMaximum(ctx, requestDurationMaximum, url, func(ctx context.Context) error {
		return c.client.SendOAuthRequest(ctx, method, url, nil, nil, responseBody, nil, tokenSource)
	})
}

// Some Dexcom API responses include a "request-time" header with, supposedly, the internal duration of the request.
// This could be useful for debugging connection issues if compared against the calculated request duration.
// See client/prometheus.go for details on how the request duration is calculated and recorded. The format for
// this header is non-standard duration (e.g. "1234 ms") and the space needs to be removed for Golang to parse.

const RequestTimeHeaderName = "request-time"

type PrometheusRequestMetricsRoundTripper struct {
	*client.PrometheusRequestMetricsRoundTripper
	requestTimeHistogramVec *prometheus.HistogramVec
}

func NewPrometheusRequestMetricsRoundTripper(name string, help string) *PrometheusRequestMetricsRoundTripper {
	return &PrometheusRequestMetricsRoundTripper{
		PrometheusRequestMetricsRoundTripper: client.NewPrometheusRequestMetricsRoundTripper(name, help),
		requestTimeHistogramVec: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    fmt.Sprintf("%s_request_time_seconds", name),
				Help:    fmt.Sprintf("%s request time (seconds)", help),
				Buckets: client.DurationBucketsDefault,
			},
			client.PrometheusLabelNames(),
		),
	}
}

func (p *PrometheusRequestMetricsRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	res, err := p.PrometheusRequestMetricsRoundTripper.RoundTrip(req)

	if res != nil {
		if labels := p.Labels(req, res); labels != nil {
			if requestTimeHeader := strings.ReplaceAll(res.Header.Get(RequestTimeHeaderName), " ", ""); requestTimeHeader != "" {
				if requestTime, parseErr := time.ParseDuration(requestTimeHeader); parseErr == nil {
					p.requestTimeHistogramVec.With(*labels).Observe(requestTime.Seconds())
				}
			}
		}
	}

	return res, err
}

const requestDurationMaximum = 60 * time.Second

var prometheusRequestMetricsRoundTripper = NewPrometheusRequestMetricsRoundTripper("tidepool_dexcom_api", "Tidepool Dexcom API")
