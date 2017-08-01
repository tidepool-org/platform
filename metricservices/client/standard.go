package client

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/version"
)

type Standard struct {
	versionReporter version.Reporter
	name            string
	config          *Config
	httpClient      *http.Client
}

const TidepoolAuthenticationTokenHeaderName = "X-Tidepool-Session-Token"

func NewStandard(versionReporter version.Reporter, name string, config *Config) (*Standard, error) {
	if versionReporter == nil {
		return nil, errors.New("client", "version reporter is missing")
	}
	if name == "" {
		return nil, errors.New("client", "name is missing")
	}
	if config == nil {
		return nil, errors.New("client", "config is missing")
	}

	if err := config.Validate(); err != nil {
		return nil, errors.Wrap(err, "client", "config is invalid")
	}

	httpClient := &http.Client{
		Timeout: config.Timeout,
	}

	return &Standard{
		versionReporter: versionReporter,
		name:            name,
		config:          config,
		httpClient:      httpClient,
	}, nil
}

func (s *Standard) RecordMetric(context Context, metric string, data ...map[string]string) error {
	if context == nil {
		return errors.New("client", "context is missing")
	}
	if metric == "" {
		return errors.New("client", "metric is missing")
	}

	data = append(data, map[string]string{"sourceVersion": s.versionReporter.Base()})

	var requestURL string
	if context.AuthenticationDetails().IsServer() {
		requestURL = s.buildURL("metrics", "server", s.name, metric)
	} else {
		requestURL = s.buildURL("metrics", "thisuser", metric)
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

	return s.sendRequest(context, "GET", requestURL+"?"+strings.Join(parameters, "&"))
}

func (s *Standard) sendRequest(context Context, requestMethod string, requestURL string) error {
	request, err := http.NewRequest(requestMethod, requestURL, nil)
	if err != nil {
		return errors.Wrapf(err, "client", "unable to create new request for %s %s", requestMethod, requestURL)
	}

	if err = service.CopyRequestTrace(context.Request(), request); err != nil {
		return errors.Wrapf(err, "client", "unable to copy request trace")
	}

	request.Header.Add(TidepoolAuthenticationTokenHeaderName, context.AuthenticationDetails().Token())

	response, err := s.httpClient.Do(request)
	if err != nil {
		return errors.Wrapf(err, "client", "unable to perform request %s %s", requestMethod, requestURL)
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusUnauthorized:
		return errors.New("client", "unauthorized")
	default:
		return errors.New("client", fmt.Sprintf("unexpected response status code %d from %s %s", response.StatusCode, requestMethod, requestURL))
	}
}

func (s *Standard) buildURL(paths ...string) string {
	return strings.Join(append([]string{s.config.Address}, paths...), "/")
}
