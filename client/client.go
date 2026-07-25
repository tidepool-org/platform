package client

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/structure"
)

const (
	ResponseBodyLimit = 1 << 20
)

// If specified, allows a client or derived class to parse any response that has
// a non-200 status code. The function should parse the response and return a
// corresponding error. If the response body cannot be parsed for any reason,
// then it should nil to indicate that no error was parsed. In such a case, an
// generalized error representing the status code will be used.
type ErrorResponseParser interface {
	ParseErrorResponse(ctx context.Context, res *http.Response, req *http.Request) error
}

type Client struct {
	config              Config
	errorResponseParser ErrorResponseParser
}

func New(cfg *Config) (*Client, error) {
	return NewWithErrorParser(cfg, nil)
}

func NewWithErrorParser(cfg *Config, errorResponseParser ErrorResponseParser) (*Client, error) {
	if cfg == nil {
		return nil, errors.New("config is missing")
	} else if err := cfg.Validate(); err != nil {
		return nil, errors.Wrap(err, "config is invalid")
	}

	return &Client{
		config:              *cfg,
		errorResponseParser: errorResponseParser,
	}, nil
}

func (c *Client) ClientTimeout() time.Duration {
	return c.config.ClientTimeout
}

func (c *Client) ResponseTimeout() time.Duration {
	return c.config.ResponseTimeout
}

func (c *Client) ConstructURL(paths ...string) string {
	return ConstructURL(c.config.Address, paths...)
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

func (c *Client) RequestStreamWithHTTPClient(ctx context.Context, method string, url string, mutators []request.RequestMutator, requestBody any, inspectors []request.ResponseInspector, httpClient *http.Client) (io.ReadCloser, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}

	// Ensure the HTTP client. If no timeout, then use timeout from config.
	httpClient = pointer.From(pointer.Default(httpClient, *http.DefaultClient))
	if httpClient.Timeout == 0 {
		if clientTimeout := c.ClientTimeout(); clientTimeout > 0 {
			httpClient.Timeout = clientTimeout
		}
	}

	// The request must carry a cancelable context for the response timeout to reach the transport. A deadline on that
	// context cannot be used, though, as it would also abort any read of the returned response body, so instead cancel
	// via a timer that is stopped once the response headers arrive. The cause preserves the deadline exceeded error.
	ctx, cancel := context.WithCancelCause(ctx)

	req, err := c.createRequest(ctx, method, url, mutators, requestBody)
	if err != nil {
		cancel(nil)
		return nil, err
	}

	// Use timer to tracker timeout for response headers. If the timeout is reached, then cancel the context with a
	// cause of deadline exceeded.
	var timer *time.Timer
	if responseTimeout := c.ResponseTimeout(); responseTimeout > 0 {
		timer = time.AfterFunc(responseTimeout, func() { cancel(context.DeadlineExceeded) })
	}

	res, err := httpClient.Do(req)

	if timer != nil {
		timer.Stop()
	}

	if err != nil {
		cancel(nil)
		return nil, errors.Wrapf(err, "unable to perform request to %s %s", method, url)
	}

	for _, inspector := range inspectors {
		inspector.InspectResponse(res)
	}

	body, err := c.handleResponse(ctx, res, req)
	if body == nil {
		cancel(nil)
		return nil, err
	}

	return &ReadCloserWithCancelCause{
		ReadCloser:      body,
		CancelCauseFunc: cancel,
	}, err
}

func (c *Client) RequestDataWithHTTPClient(ctx context.Context, method string, url string, mutators []request.RequestMutator, requestBody any, responseBody any, inspectors []request.ResponseInspector, httpClient *http.Client) error {
	body, err := c.RequestStreamWithHTTPClient(ctx, method, url, mutators, requestBody, inspectors, httpClient)
	if err != nil {
		return err
	} else if body == nil {
		return nil
	}

	defer request.DrainAndClose(body)

	if responseBody == nil {
		return nil
	}

	return request.DecodeStream(ctx, structure.NewPointerSource(), body, responseBody)
}

func (c *Client) createRequest(ctx context.Context, method string, url string, mutators []request.RequestMutator, requestBody any) (*http.Request, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if method == "" {
		return nil, errors.New("method is missing")
	}
	if url == "" {
		return nil, errors.New("url is missing")
	}

	if c.config.UserAgent != "" {
		mutators = append(mutators, request.NewHeaderMutator("User-Agent", c.config.UserAgent))
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

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to create request to %s %s", method, url)
	}

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
			defer request.DrainAndClose(res.Body)
			return nil, nil
		default:
			return res.Body, nil
		}
	}

	defer request.DrainAndClose(res.Body)

	bites, err := io.ReadAll(io.LimitReader(res.Body, ResponseBodyLimit))
	if err != nil {
		return nil, errors.Wrap(err, "unable to read response body")
	}

	var responseErr error

	// If we fully consume the response body, then allow error response parser to parse
	if len(bites) < ResponseBodyLimit {
		if c.errorResponseParser != nil {
			res.Body = io.NopCloser(bytes.NewBuffer(bites))
			responseErr = c.errorResponseParser.ParseErrorResponse(ctx, res, req)
		}
	}

	// If no error yet, then generate generic error
	if responseErr == nil {
		res.Body = io.NopCloser(bytes.NewBuffer(bites))
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

func responseBodyFromBytes(bites []byte) any {
	if utf8.Valid(bites) {
		return string(bites)
	}
	return bites
}

type ReadCloserWithCancelCause struct {
	io.ReadCloser
	context.CancelCauseFunc
}

func (r *ReadCloserWithCancelCause) Close() error {
	defer r.CancelCauseFunc(nil)
	return r.ReadCloser.Close()
}

func NewSerializableErrorResponseParser() *SerializableErrorResponseParser {
	return &SerializableErrorResponseParser{}
}

type SerializableErrorResponseParser struct{}

func (s *SerializableErrorResponseParser) ParseErrorResponse(ctx context.Context, res *http.Response, req *http.Request) error {
	serializable := &errors.Serializable{}
	if err := json.NewDecoder(res.Body).Decode(serializable); err != nil {
		return nil
	}
	return serializable.Error
}

func ConstructURL(url string, paths ...string) string {
	return strings.TrimRight(url, "/") + ConstructPath(paths...)
}

func ConstructPath(paths ...string) string {
	segments := []string{}
	for _, path := range paths {
		segments = append(segments, url.PathEscape(strings.Trim(path, "/")))
	}
	return "/" + strings.Join(segments, "/")
}
