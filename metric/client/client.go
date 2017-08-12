package client

import (
	"net/url"
	"strings"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/version"
)

type Client interface {
	RecordMetric(context auth.Context, name string, data ...map[string]string) error
}

type clientImpl struct {
	client          *client.Client
	name            string
	versionReporter version.Reporter
}

func NewClient(config *client.Config, name string, versionReporter version.Reporter) (Client, error) {
	if config == nil {
		return nil, errors.New("client", "config is missing")
	}
	if name == "" {
		return nil, errors.New("client", "name is missing")
	}
	if versionReporter == nil {
		return nil, errors.New("client", "version reporter is missing")
	}

	clnt, err := client.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &clientImpl{
		client:          clnt,
		name:            name,
		versionReporter: versionReporter,
	}, nil
}

func (c *clientImpl) RecordMetric(context auth.Context, metric string, data ...map[string]string) error {
	if context == nil {
		return errors.New("client", "context is missing")
	}
	if metric == "" {
		return errors.New("client", "metric is missing")
	}

	data = append(data, map[string]string{"sourceVersion": c.versionReporter.Base()})

	var requestURL string
	if context.AuthDetails().IsServer() {
		requestURL = c.client.BuildURL("metrics", "server", c.name, metric)
	} else {
		requestURL = c.client.BuildURL("metrics", "thisuser", metric)
	}

	var parameters []string
	for _, datum := range data {
		for key, value := range datum {
			if key != "" {
				parameters = append(parameters, url.QueryEscape(key)+"="+url.QueryEscape(value))
			}
		}
	}

	context.Logger().WithFields(log.Fields{"metric": metric, "data": data}).Debug("Recording metric")

	return c.client.SendRequestWithAuthToken(context, "GET", requestURL+"?"+strings.Join(parameters, "&"), nil, nil)
}
