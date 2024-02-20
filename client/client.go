package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"unicode/utf8"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/structure"
)

const (
	ResponseBodyLimit = 1 << 20
)

type Client struct {
	address             string
	userAgent           string
	responseErrorParser ResponseErrorParser
}

func New(cfg *Config) (*Client, error) {
	if cfg == nil {
		return nil, errors.New("config is missing")
	} else if err := cfg.Validate(); err != nil {
		return nil, errors.Wrap(err, "config is invalid")
	}

	return &Client{
		address:             cfg.Address,
		userAgent:           cfg.UserAgent,
		responseErrorParser: cfg.ResponseErrorParser,
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

func (c *Client) RequestStreamWithHTTPClient(ctx context.Context, method string, url string, mutators []request.RequestMutator, requestBody interface{}, inspectors []request.ResponseInspector, httpClient *http.Client) (io.ReadCloser, error) {
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

	for _, inspector := range inspectors {
		inspector.InspectResponse(res)
	}

	return c.handleResponse(ctx, res, req)
}

func (c *Client) RequestDataWithHTTPClient(ctx context.Context, method string, url string, mutators []request.RequestMutator, requestBody interface{}, responseBody interface{}, inspectors []request.ResponseInspector, httpClient *http.Client) error {
	headerInspector := request.NewHeadersInspector(log.LoggerFromContext(ctx))
	body, err := c.RequestStreamWithHTTPClient(ctx, method, url, mutators, requestBody, append(inspectors, headerInspector), httpClient)
	if err != nil {
		return err
	} else if body == nil {
		return nil
	}

	defer drainAndClose(body)

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

	if c.userAgent != "" {
		mutators = append(mutators, request.NewHeaderMutator("User-Agent", c.userAgent))
	}

	var body io.Reader
	if requestBody != nil {
		if valueOf := reflect.ValueOf(requestBody); valueOf.Kind() != reflect.Ptr || !valueOf.IsNil() {
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
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to create request to %s %s", method, url)
	}

	req = req.WithContext(ctx)

	for _, mutator := range mutators {
		if err = mutator.MutateRequest(req); err != nil {
			return nil, errors.Wrapf(err, "unable to mutate request to %s %s", method, url)
		}
	}

	// TODO: Prevents random EOF errors (I think due to the server closing Keep Alive connections automatically)
	// TODO: Would be better to retry the request with exponential fallback
	req.Close = true

	return req, nil
}

func (c *Client) handleResponse(ctx context.Context, res *http.Response, req *http.Request) (io.ReadCloser, error) {
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"method": req.Method, "url": req.URL.String()})

	if request.IsStatusCodeSuccess(res.StatusCode) {
		switch res.StatusCode {
		case http.StatusNoContent, http.StatusResetContent:
			drainAndClose(res.Body)
			return nil, nil
		default:
			return res.Body, nil
		}
	}

	defer drainAndClose(res.Body)

	bites, err := io.ReadAll(io.LimitReader(res.Body, ResponseBodyLimit))
	if err != nil {
		return nil, errors.Wrap(err, "unable to read response body")
	}
	res.Body = io.NopCloser(bytes.NewBuffer(bites))

	var responseErr error

	if len(bites) < ResponseBodyLimit {
		if c.responseErrorParser != nil {
			responseErr = c.responseErrorParser(ctx, res, req)
		}
	}

	if responseErr == nil {
		responseErr = errorFromStatusCode(res, req)
		logger = logger.WithField("responseBody", responseBodyFromBytes(bites))
	}

	logger = logger.WithError(responseErr)

	switch errors.Code(responseErr) {
	case request.ErrorCodeBadRequest:
		logger.Error("Bad request")
	case request.ErrorCodeTooManyRequests:
		logger.Error("Too many requests")
	case request.ErrorCodeUnexpectedResponse:
		logger.Error("Unexpected response")
	}

	return nil, responseErr
}

func errorFromStatusCode(res *http.Response, req *http.Request) error {
	switch res.StatusCode {
	case http.StatusBadRequest:
		return request.ErrorBadRequest()
	case http.StatusUnauthorized:
		return request.ErrorUnauthenticated()
	case http.StatusForbidden:
		return request.ErrorUnauthorized()
	case http.StatusNotFound:
		return request.ErrorResourceNotFound()
	case http.StatusRequestEntityTooLarge:
		return request.ErrorResourceTooLarge()
	case http.StatusTooManyRequests:
		return request.ErrorTooManyRequests()
	default:
		return request.ErrorUnexpectedResponse(res, req)
	}
}

func responseBodyFromBytes(bites []byte) interface{} {
	if utf8.Valid(bites) {
		return string(bites)
	}
	return bites
}

func drainAndClose(reader io.ReadCloser) {
	io.Copy(io.Discard, reader)
	reader.Close()
}
