package client

import (
	"context"
	"time"

	"golang.org/x/oauth2"

	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/dexcom"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/oauth"
	oauthClient "github.com/tidepool-org/platform/oauth/client"
)

type Client struct {
	client *oauthClient.Client
}

func New(cfg *client.Config, prvdr oauth.Provider) (*Client, error) {
	clnt, err := oauthClient.New(cfg, prvdr)
	if err != nil {
		return nil, err
	}

	// Dexcom authorization server does not support HTTP Basic authentication
	oauth2.RegisterBrokenAuthHeaderProvider(cfg.Address)

	return &Client{
		client: clnt,
	}, nil
}

func (c *Client) GetCalibrations(ctx context.Context, startTime time.Time, endTime time.Time, tokenSource oauth.TokenSource) (*dexcom.CalibrationsResponse, error) {
	calibrationsResponse := &dexcom.CalibrationsResponse{}
	if err := c.sendDexcomRequest(ctx, startTime, endTime, "GET", c.client.ConstructURL("p", "v1", "users", "self", "calibrations"), calibrationsResponse, tokenSource); err != nil {
		return nil, errors.Wrap(err, "unable to get calibrations")
	}

	return calibrationsResponse, nil
}

func (c *Client) GetDevices(ctx context.Context, startTime time.Time, endTime time.Time, tokenSource oauth.TokenSource) (*dexcom.DevicesResponse, error) {
	devicesResponse := &dexcom.DevicesResponse{}
	if err := c.sendDexcomRequest(ctx, startTime, endTime, "GET", c.client.ConstructURL("p", "v1", "users", "self", "devices"), devicesResponse, tokenSource); err != nil {
		return nil, errors.Wrap(err, "unable to get devices")
	}

	return devicesResponse, nil
}

func (c *Client) GetEGVs(ctx context.Context, startTime time.Time, endTime time.Time, tokenSource oauth.TokenSource) (*dexcom.EGVsResponse, error) {
	egvsResponse := &dexcom.EGVsResponse{}
	if err := c.sendDexcomRequest(ctx, startTime, endTime, "GET", c.client.ConstructURL("p", "v1", "users", "self", "egvs"), egvsResponse, tokenSource); err != nil {
		return nil, errors.Wrap(err, "unable to get egvs")
	}

	return egvsResponse, nil
}

func (c *Client) GetEvents(ctx context.Context, startTime time.Time, endTime time.Time, tokenSource oauth.TokenSource) (*dexcom.EventsResponse, error) {
	eventsResponse := &dexcom.EventsResponse{}
	if err := c.sendDexcomRequest(ctx, startTime, endTime, "GET", c.client.ConstructURL("p", "v1", "users", "self", "events"), eventsResponse, tokenSource); err != nil {
		return nil, errors.Wrap(err, "unable to get events")
	}

	return eventsResponse, nil
}

func (c *Client) sendDexcomRequest(ctx context.Context, startTime time.Time, endTime time.Time, method string, url string, responseBody interface{}, tokenSource oauth.TokenSource) error {
	query := map[string]string{
		"startDate": startTime.Format(dexcom.DateTimeFormat),
		"endDate":   endTime.Format(dexcom.DateTimeFormat),
	}
	// TODO: Check for actual dexcom errors here!!!
	return c.client.SendOAuthRequest(ctx, method, c.client.AppendURLQuery(url, query), nil, nil, responseBody, tokenSource)
}
