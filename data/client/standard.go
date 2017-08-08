package client

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/service"
)

type Standard struct {
	config     *Config
	httpClient *http.Client
}

const TidepoolAuthenticationTokenHeaderName = "X-Tidepool-Session-Token"

func NewStandard(config *Config) (*Standard, error) {
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
		config:     config,
		httpClient: httpClient,
	}, nil
}

func (s *Standard) DestroyDataForUserByID(context Context, userID string) error {
	if context == nil {
		return errors.New("client", "context is missing")
	}
	if userID == "" {
		return errors.New("client", "user id is missing")
	}

	context.Logger().WithField("userId", userID).Debug("Deleting data for user")

	return s.sendRequest(context, "DELETE", s.buildURL("dataservices", "v1", "users", userID, "data")) // TODO: Fix url
}

func (s *Standard) sendRequest(context Context, requestMethod string, requestURL string) error {
	request, err := http.NewRequest(requestMethod, requestURL, nil)
	if err != nil {
		return errors.Wrapf(err, "client", "unable to create new request for %s %s", requestMethod, requestURL)
	}

	if err = service.CopyRequestTrace(context.Request(), request); err != nil {
		return errors.Wrapf(err, "client", "unable to copy request trace")
	}

	serverToken, err := context.UserClient().ServerToken()
	if err != nil {
		return err
	}

	request.Header.Add(TidepoolAuthenticationTokenHeaderName, serverToken)

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
