package client

import (
	"context"
	"net/url"
	"strings"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/platform"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/version"
)

type Client struct {
	client          *platform.Client
	name            string
	versionReporter version.Reporter
}

func New(cfg *platform.Config, name string, versionReporter version.Reporter) (*Client, error) {
	if cfg == nil {
		return nil, errors.New("config is missing")
	}
	if name == "" {
		return nil, errors.New("name is missing")
	}
	if versionReporter == nil {
		return nil, errors.New("version reporter is missing")
	}

	clnt, err := platform.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return &Client{
		client:          clnt,
		name:            name,
		versionReporter: versionReporter,
	}, nil
}

func (c *Client) RecordMetric(ctx context.Context, metric string, data ...map[string]string) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if metric == "" {
		return errors.New("metric is missing")
	}

	data = append(data, map[string]string{"sourceVersion": c.versionReporter.Base()})

	var requestURL string
	if details := request.DetailsFromContext(ctx); details.IsService() {
		requestURL = c.client.ConstructURL("metrics", "server", c.name, metric)
	} else {
		requestURL = c.client.ConstructURL("metrics", "thisuser", metric)
	}

	var parameters []string
	for _, datum := range data {
		for key, value := range datum {
			if key != "" {
				parameters = append(parameters, url.QueryEscape(key)+"="+url.QueryEscape(value))
			}
		}
	}

	log.LoggerFromContext(ctx).WithFields(log.Fields{"metric": metric, "data": data}).Debug("Recording metric")

	return c.client.SendRequestAsUser(ctx, "GET", requestURL+"?"+strings.Join(parameters, "&"), nil, nil, nil)
}
