package client

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/service"
)

type Standard struct {
	config     *Config
	httpClient *http.Client
}

const TidepoolAuthenticationTokenHeaderName = "X-Tidepool-Session-Token"

func NewStandard(config *Config) (*Standard, error) {
	if config == nil {
		return nil, app.Error("client", "config is missing")
	}

	config = config.Clone()
	if err := config.Validate(); err != nil {
		return nil, app.ExtError(err, "client", "config is invalid")
	}

	httpClient := &http.Client{
		Timeout: time.Duration(config.RequestTimeout) * time.Second,
	}

	return &Standard{
		config:     config,
		httpClient: httpClient,
	}, nil
}

func (s *Standard) DestroyDataForUserByID(context Context, userID string) error {
	if context == nil {
		return app.Error("client", "context is missing")
	}
	if userID == "" {
		return app.Error("client", "user id is missing")
	}

	context.Logger().WithField("user-id", userID).Debug("Deleting data for user")

	return s.sendRequest(context, "DELETE", s.buildURL("dataservices", "v1", "users", userID, "data"))
}

func (s *Standard) sendRequest(context Context, requestMethod string, requestURL string) error {
	request, err := http.NewRequest(requestMethod, requestURL, nil)
	if err != nil {
		return app.ExtErrorf(err, "client", "unable to create new request for %s %s", requestMethod, requestURL)
	}

	if err = service.CopyRequestTrace(context.Request(), request); err != nil {
		return app.ExtErrorf(err, "client", "unable to copy request trace")
	}

	serverToken, err := context.UserServicesClient().ServerToken()
	if err != nil {
		return err
	}

	request.Header.Add(TidepoolAuthenticationTokenHeaderName, serverToken)

	response, err := s.httpClient.Do(request)
	if err != nil {
		return app.ExtErrorf(err, "client", "unable to perform request %s %s", requestMethod, requestURL)
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusUnauthorized:
		return app.Error("client", "unauthorized")
	default:
		return app.Error("client", fmt.Sprintf("unexpected response status code %d from %s %s", response.StatusCode, requestMethod, requestURL))
	}
}

func (s *Standard) buildURL(paths ...string) string {
	return strings.Join(append([]string{s.config.Address}, paths...), "/")
}
