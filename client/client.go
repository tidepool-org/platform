package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	netURL "net/url"
	"strings"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/request"
)

type Client struct {
	address   string
	userAgent string
}

func New(cfg *Config) (*Client, error) {
	if cfg == nil {
		return nil, errors.New("config is missing")
	}

	if err := cfg.Validate(); err != nil {
		return nil, errors.Wrap(err, "config is invalid")
	}

	return &Client{
		address:   cfg.Address,
		userAgent: cfg.UserAgent,
	}, nil
}

func (c *Client) ConstructURL(paths ...string) string {
	segments := []string{}
	for _, path := range paths {
		segments = append(segments, netURL.PathEscape(strings.Trim(path, "/")))
	}
	return fmt.Sprintf("%s/%s", strings.TrimRight(c.address, "/"), strings.Join(segments, "/"))
}

func (c *Client) AppendURLQuery(urlString string, query map[string]string) string {
	values := netURL.Values{}
	for k, v := range query {
		values.Add(k, v)
	}

	queryString := values.Encode()
	if queryString != "" {
		if strings.Contains(urlString, "?") {
			urlString += "&"
		} else {
			urlString += "?"
		}
		urlString += queryString
	}

	return urlString
}

func (c *Client) SendRequest(ctx context.Context, method string, url string, mutators []request.Mutator, requestBody interface{}, responseBody interface{}, httpClient *http.Client) error {
	if httpClient == nil {
		return errors.New("http client is missing")
	}

	req, err := c.buildRequest(ctx, method, url, mutators, requestBody, responseBody)
	if err != nil {
		return err
	}

	// TODO: Prevents random EOF errors (I think due to the server closing Keep Alive connections automatically)
	// TODO: Would be better to retry the request with exponential fallback
	req.Close = true

	res, err := httpClient.Do(req)
	if err != nil {
		return errors.Wrapf(err, "unable to perform request %s %s", method, url)
	}
	if res.Body != nil {
		defer res.Body.Close()
	}

	switch res.StatusCode {
	case http.StatusOK, http.StatusCreated:
		return c.decodeResponseBody(res, responseBody)
	case http.StatusNoContent:
		return nil
	case http.StatusBadRequest:
		return c.handleBadRequest(res, req)
	case http.StatusUnauthorized:
		return request.ErrorUnauthenticated()
	case http.StatusForbidden:
		return request.ErrorUnauthorized()
	case http.StatusNotFound:
		return request.ErrorResourceNotFound()
	case http.StatusTooManyRequests:
		return request.ErrorTooManyRequests()
	}

	return request.ErrorUnexpectedResponse(res, req)
}

func (c *Client) buildRequest(ctx context.Context, method string, url string, mutators []request.Mutator, requestBody interface{}, responseBody interface{}) (*http.Request, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if method == "" {
		return nil, errors.New("method is missing")
	}
	if url == "" {
		return nil, errors.New("url is missing")
	}

	body, err := c.encodeRequestBody(requestBody)
	if err != nil {
		return nil, errors.Wrapf(err, "error encoding JSON request to %s %s", method, url)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to create new request for %s %s", method, url)
	}

	req = req.WithContext(ctx)

	for _, mutator := range mutators {
		if mutator != nil {
			if err = mutator.Mutate(req); err != nil {
				return nil, errors.Wrapf(err, "unable to mutate request")
			}
		}
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("User-Agent", c.userAgent)

	return req, nil
}

func (c *Client) encodeRequestBody(object interface{}) (io.Reader, error) {
	if object == nil {
		return nil, nil
	}

	buffer := &bytes.Buffer{}
	if err := json.NewEncoder(buffer).Encode(object); err != nil {
		return nil, err
	}

	return buffer, nil
}

func (c *Client) decodeResponseBody(res *http.Response, object interface{}) error {
	if object == nil {
		return nil
	}

	return request.DecodeResponseBody(res, object)
}

func (c *Client) handleBadRequest(res *http.Response, req *http.Request) error {
	if logger := log.LoggerFromContext(req.Context()); logger != nil {
		if res.Body != nil {
			buffer := bytes.Buffer{}
			if _, err := buffer.ReadFrom(io.LimitReader(res.Body, 1<<16)); err == nil {
				logger = logger.WithField("responseBody", buffer.String())
			}
		}
		logger.WithFields(log.Fields{"method": req.Method, "url": req.URL.String()}).Warn("bad request")
	}
	return request.ErrorBadRequest()
}
