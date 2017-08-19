package client

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/service"
)

type Client struct {
	httpClient *http.Client
	address    string
}

func NewClient(config *Config) (*Client, error) {
	if config == nil {
		return nil, errors.New("client", "config is missing")
	}

	if err := config.Validate(); err != nil {
		return nil, errors.Wrap(err, "client", "config is invalid")
	}

	httpClient := &http.Client{
		Timeout: config.Timeout,
	}

	return &Client{
		httpClient: httpClient,
		address:    config.Address,
	}, nil
}

func (c *Client) HTTPClient() *http.Client {
	return c.httpClient
}

func (c *Client) BuildURL(paths ...string) string {
	parts := []string{c.address}
	for _, path := range paths {
		parts = append(parts, url.PathEscape(path))
	}
	return strings.Join(parts, "/")
}

func (c *Client) SendRequestWithAuthToken(context auth.Context, method string, url string, requestObject interface{}, responseObject interface{}) error {
	if context == nil {
		return errors.New("client", "context is missing")
	}

	return c.sendRequest(context, method, url, context.AuthDetails().Token(), requestObject, responseObject)
}

func (c *Client) SendRequestWithServerToken(context auth.Context, method string, url string, requestObject interface{}, responseObject interface{}) error {
	if context == nil {
		return errors.New("client", "context is missing")
	}

	token, err := context.AuthClient().ServerToken()
	if err != nil {
		return err
	}

	return c.sendRequest(context, method, url, token, requestObject, responseObject)
}

func (c *Client) sendRequest(context auth.Context, method string, url string, token string, requestObject interface{}, responseObject interface{}) error {
	request, err := c.buildRequest(context, method, url, token, requestObject, responseObject)
	if err != nil {
		return err
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return errors.Wrapf(err, "client", "unable to perform request %s %s", method, url)
	}
	if response.Body != nil {
		defer response.Body.Close()
	}

	switch response.StatusCode {
	case http.StatusOK, http.StatusCreated:
		if err = c.decodeResponseObject(response.Body, responseObject); err != nil {
			return errors.Wrapf(err, "client", "error decoding JSON response from %s %s", method, url)
		}
		return nil
	case http.StatusUnauthorized:
		return NewUnauthorizedError()
	}

	return NewUnexpectedResponseError(response, request)
}

func (c *Client) buildRequest(context auth.Context, method string, url string, token string, requestObject interface{}, responseObject interface{}) (*http.Request, error) {
	if context == nil {
		return nil, errors.New("client", "context is missing")
	}
	if method == "" {
		return nil, errors.New("client", "method is missing")
	}
	if url == "" {
		return nil, errors.New("client", "url is missing")
	}
	if token == "" {
		return nil, errors.New("client", "token is missing")
	}

	requestBody, err := c.encodeRequestObject(requestObject)
	if err != nil {
		return nil, errors.Wrapf(err, "client", "error encoding JSON request to %s %s", method, url)
	}

	request, err := http.NewRequest(method, url, requestBody)
	if err != nil {
		return nil, errors.Wrapf(err, "client", "unable to create new request for %s %s", method, url)
	}

	if err = service.CopyRequestTrace(context.Request(), request); err != nil {
		return nil, errors.Wrapf(err, "client", "unable to copy request trace")
	}

	request.Header.Add(auth.TidepoolAuthTokenHeaderName, token)

	return request, nil
}

func (c *Client) encodeRequestObject(requestObject interface{}) (io.Reader, error) {
	if requestObject == nil {
		return nil, nil
	}

	buffer := &bytes.Buffer{}
	if err := json.NewEncoder(buffer).Encode(requestObject); err != nil {
		return nil, err
	}

	return buffer, nil
}

func (c *Client) decodeResponseObject(responseBody io.Reader, responseObject interface{}) error {
	if responseObject == nil {
		return nil
	}
	if responseBody == nil {
		return errors.New("client", "response body is empty")
	}

	return json.NewDecoder(responseBody).Decode(responseObject)
}
