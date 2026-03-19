package jotform

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/request"
)

type Client interface {
	ListFormSubmissions(ctx context.Context, formID string, filter *SubmissionFilter) (*FormSubmissionsResponse, error)
	GetSubmission(ctx context.Context, submissionID string) (*SubmissionResponse, error)
}

type SubmissionFilter struct {
	IDGreaterThan string
	Limit         int
}

type FormSubmissionsResponse struct {
	ResponseCode int       `json:"responseCode"`
	Message      string    `json:"message"`
	Content      []Content `json:"content"`
}

type defaultClient struct {
	client     *client.Client
	config     Config
	httpClient *http.Client
}

func NewClient(config Config) (Client, error) {
	c, err := client.NewWithErrorParser(&client.Config{
		Address: config.BaseURL,
	}, &errorResponseParser{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create defaultClient")
	}

	return &defaultClient{
		client:     c,
		config:     config,
		httpClient: http.DefaultClient,
	}, nil
}

// errorResponseParser implements client.ErrorResponseParser for Jotform API errors
type errorResponseParser struct{}

func (p *errorResponseParser) ParseErrorResponse(ctx context.Context, res *http.Response, req *http.Request) error {
	var errResp struct {
		ResponseCode int    `json:"responseCode"`
		Message      string `json:"message"`
	}
	if err := json.NewDecoder(res.Body).Decode(&errResp); err != nil {
		return nil
	}

	if errResp.Message != "" {
		return errors.Newf("Jotform API error (status %d): %s", res.StatusCode, errResp.Message)
	}

	return nil
}

// ListFormSubmissions fetches submissions for a form with optional filtering
func (c *defaultClient) ListFormSubmissions(ctx context.Context, formID string, filter *SubmissionFilter) (*FormSubmissionsResponse, error) {
	if formID == "" {
		return nil, errors.New("form ID is required")
	}

	url := c.client.ConstructURL("v1", "form", formID, "submissions")

	query := make(map[string]string)
	if filter != nil {
		query["limit"] = strconv.Itoa(filter.Limit)
		query["filter"] = fmt.Sprintf(`{"id:gt":"%s"}`, filter.IDGreaterThan)
	}
	query["orderby"] = "-id"

	url = c.client.AppendURLQuery(url, query)

	var response FormSubmissionsResponse
	if err := c.client.RequestDataWithHTTPClient(ctx, http.MethodGet, url, c.authMutators(), nil, &response, nil, c.httpClient); err != nil {
		return nil, err
	}

	if response.ResponseCode != http.StatusOK {
		return nil, errors.Newf("unexpected response code: %d", response.ResponseCode)
	}

	return &response, nil
}

// GetSubmission fetches a single submission by ID
func (c *defaultClient) GetSubmission(ctx context.Context, submissionID string) (*SubmissionResponse, error) {
	url := c.client.ConstructURL("v1", "submission", submissionID)

	var response SubmissionResponse
	if err := c.client.RequestDataWithHTTPClient(ctx, http.MethodGet, url, c.authMutators(), nil, &response, nil, c.httpClient); err != nil {
		return nil, err
	}

	// Sometimes the jotform webhook returns a 200 http response with a non-200 response code in the body
	if response.ResponseCode != http.StatusOK {
		return nil, errors.Newf("unexpected response code: %d", response.ResponseCode)
	}

	return &response, nil
}

func (c *defaultClient) authMutators() []request.RequestMutator{
	mutators := []request.RequestMutator{
		request.NewHeaderMutator("APIKEY", c.config.APIKey),
	}
	if c.config.TeamID != "" {
		mutators = append(mutators, request.NewHeaderMutator("jf-team-id", c.config.TeamID))
	}
	return mutators
}