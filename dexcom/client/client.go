package client

import (
	"context"
	"time"

	"golang.org/x/oauth2"

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

	// NOTE: Dexcom authorization server does not support HTTP Basic authentication
	oauth2.RegisterBrokenAuthHeaderProvider(cfg.Address)

	isSandboxData := false
	if cfg != nil && cfg.Address == "https://sandbox-api.dexcom.com" {
		isSandboxData = true
	}

	return &Client{
		client:        clnt,
		isSandboxData: isSandboxData,
	}, nil
}

func (c *Client) GetCalibrations(ctx context.Context, startTime time.Time, endTime time.Time, tokenSource oauth.TokenSource) (*dexcom.CalibrationsResponse, error) {
	calibrationsResponse := &dexcom.CalibrationsResponse{}
	paths := []string{"v3", "users", "self", "calibrations"}
	if c.isSandboxData {
		paths = paths[1:]
	}

	if err := c.sendDexcomRequest(ctx, startTime, endTime, "GET", c.client.ConstructURL(paths...), calibrationsResponse, tokenSource); err != nil {
		return nil, errors.Wrap(err, "unable to get calibrations")
	}

	return calibrationsResponse, nil
}

func (c *Client) GetDevices(ctx context.Context, startTime time.Time, endTime time.Time, tokenSource oauth.TokenSource) (*dexcom.DevicesResponse, error) {
	devicesResponse := &dexcom.DevicesResponse{IsSandboxData: c.isSandboxData}
	paths := []string{"v3", "users", "self", "devices"}
	if c.isSandboxData {
		paths = paths[1:]
	}

	if err := c.sendDexcomRequest(ctx, startTime, endTime, "GET", c.client.ConstructURL(paths...), devicesResponse, tokenSource); err != nil {
		return nil, errors.Wrap(err, "unable to get devices")
	}

	return devicesResponse, nil
}

func (c *Client) GetEGVs(ctx context.Context, startTime time.Time, endTime time.Time, tokenSource oauth.TokenSource) (*dexcom.EGVsResponse, error) {
	egvsResponse := &dexcom.EGVsResponse{}
	paths := []string{"v3", "users", "self", "egvs"}
	if c.isSandboxData {
		paths = paths[1:]
	}

	if err := c.sendDexcomRequest(ctx, startTime, endTime, "GET", c.client.ConstructURL(paths...), egvsResponse, tokenSource); err != nil {
		return nil, errors.Wrap(err, "unable to get egvs")
	}

	return egvsResponse, nil
}

func (c *Client) GetEvents(ctx context.Context, startTime time.Time, endTime time.Time, tokenSource oauth.TokenSource) (*dexcom.EventsResponse, error) {
	eventsResponse := &dexcom.EventsResponse{}
	paths := []string{"v3", "users", "self", "events"}
	if c.isSandboxData {
		paths = paths[1:]
	}

	if err := c.sendDexcomRequest(ctx, startTime, endTime, "GET", c.client.ConstructURL(paths...), eventsResponse, tokenSource); err != nil {
		return nil, errors.Wrap(err, "unable to get events")
	}

	return eventsResponse, nil
}

func (c *Client) sendDexcomRequest(ctx context.Context, startTime time.Time, endTime time.Time, method string, url string, responseBody interface{}, tokenSource oauth.TokenSource) error {
	now := time.Now()

	url = c.client.AppendURLQuery(url, map[string]string{
		"startDate": startTime.UTC().Format(dexcom.TimeFormat),
		"endDate":   endTime.UTC().Format(dexcom.TimeFormat),
	})

	err := c.client.SendOAuthRequest(ctx, method, url, nil, nil, responseBody, tokenSource)
	if oauth.IsAccessTokenError(err) {
		tokenSource.ExpireToken()
		err = c.client.SendOAuthRequest(ctx, method, url, nil, nil, responseBody, tokenSource)
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
