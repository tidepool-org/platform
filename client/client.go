package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/structure"
)

type Client struct {
	address   string
	userAgent string
}

func New(cfg *Config) (*Client, error) {
	if cfg == nil {
		return nil, errors.New("config is missing")
	} else if err := cfg.Validate(); err != nil {
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
		segments = append(segments, url.PathEscape(strings.Trim(path, "/")))
	}
	return fmt.Sprintf("%s/%s", strings.TrimRight(c.address, "/"), strings.Join(segments, "/"))
}

func (c *Client) AppendURLQuery(urlString string, query map[string]string) string {
	values := url.Values{}
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

func (c *Client) RequestStreamWithHTTPClient(ctx context.Context, method string, url string, mutators []request.RequestMutator, requestBody interface{}, httpClient *http.Client) (io.ReadCloser, error) {
	if httpClient == nil {
		return nil, errors.New("http client is missing")
	}

	req, err := c.createRequest(ctx, method, url, mutators, requestBody)
	if err != nil {
		return nil, err
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to perform request to %s %s", method, url)
	}

	return c.handleResponse(ctx, res, req)
}

func (c *Client) RequestDataWithHTTPClient(ctx context.Context, method string, url string, mutators []request.RequestMutator, requestBody interface{}, responseBody interface{}, httpClient *http.Client) error {
	body, err := c.RequestStreamWithHTTPClient(ctx, method, url, mutators, requestBody, httpClient)
	if err != nil {
		return err
	} else if body == nil {
		return nil
	}

	defer body.Close()

	if responseBody == nil {
		return nil
	}

	return request.DecodeObject(structure.NewPointerSource(), body, responseBody)
}

func (c *Client) createRequest(ctx context.Context, method string, url string, mutators []request.RequestMutator, requestBody interface{}) (*http.Request, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if method == "" {
		return nil, errors.New("method is missing")
	}
	if url == "" {
		return nil, errors.New("url is missing")
	}

	mutators = append(mutators, request.NewHeaderMutator("User-Agent", c.userAgent))

	var body io.Reader
	if requestBody != nil {
		if reader, ok := requestBody.(io.Reader); ok {
			body = reader
		} else {
			buffer := &bytes.Buffer{}
			if err := json.NewEncoder(buffer).Encode(requestBody); err != nil {
				return nil, errors.Wrapf(err, "unable to serialize request to %s %s", method, url)
			}
			body = buffer
			mutators = append(mutators, request.NewHeaderMutator("Content-Type", "application/json; charset=utf-8"))
		}
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to create request to %s %s", method, url)
	}

	req = req.WithContext(ctx)

	for _, mutator := range mutators {
		if mutator != nil {
			if err = mutator.MutateRequest(req); err != nil {
				return nil, errors.Wrapf(err, "unable to mutate request to %s %s", method, url)
			}
		}
	}

	// TODO: Prevents random EOF errors (I think due to the server closing Keep Alive connections automatically)
	// TODO: Would be better to retry the request with exponential fallback
	req.Close = true

	return req, nil
}

func (c *Client) handleResponse(ctx context.Context, res *http.Response, req *http.Request) (io.ReadCloser, error) {
	if request.IsStatusCodeSuccess(res.StatusCode) {
		switch res.StatusCode {
		case http.StatusNoContent, http.StatusResetContent:
			res.Body.Close()
			return nil, nil
		default:
			return res.Body, nil
		}
	}

	defer res.Body.Close()

	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"method": req.Method, "url": req.URL.String()})

	buffer := bytes.Buffer{}
	if _, err := buffer.ReadFrom(io.LimitReader(res.Body, 1<<16)); err == nil {
		logger = logger.WithField("responseBody", buffer.String())
	}

	var err error
	switch res.StatusCode {
	case http.StatusBadRequest:
		logger.Warn("Bad request")
		err = request.ErrorBadRequest()
	case http.StatusUnauthorized:
		err = request.ErrorUnauthenticated()
	case http.StatusForbidden:
		err = request.ErrorUnauthorized()
	case http.StatusNotFound:
		err = request.ErrorResourceNotFound()
	case http.StatusTooManyRequests:
		logger.Warn("Too many requests")
		err = request.ErrorTooManyRequests()
	default:
		logger.Warn("Unexpected response")
		err = request.ErrorUnexpectedResponse(res, req)
	}
	return nil, err
}
