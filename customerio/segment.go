package customerio

import (
	"context"
	"net/http"

	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/request"
)

type Identifiers struct {
	Email string `json:"email"`
	ID    string `json:"id"`
	CID   string `json:"cio_id"`
}

type segmentMembershipResponse struct {
	Identifiers []Identifiers `json:"identifiers"`
	IDs         []string      `json:"ids"`
	Next        string        `json:"next,omitempty"`
}

func (c *Client) ListCustomersInSegment(ctx context.Context, segmentID string) ([]Identifiers, error) {
	ctx = log.NewContextWithLogger(ctx, c.logger)
	var allIdentifiers []Identifiers
	start := ""

	for {
		url := c.appClient.ConstructURL("v1", "segments", segmentID, "membership")

		mutators := []request.RequestMutator{
			c.appAPIAuthMutator(),
		}

		// Add pagination parameter if available
		if start != "" {
			mutators = append(mutators, request.NewParameterMutator("start", start))
		}

		var response segmentMembershipResponse
		if err := c.appClient.RequestDataWithHTTPClient(ctx, http.MethodGet, url, mutators, nil, &response, nil, c.httpClient); err != nil {
			return nil, err
		}

		allIdentifiers = append(allIdentifiers, response.Identifiers...)

		// Check if there are more pages
		if response.Next == "" {
			break
		}
		start = response.Next
	}

	return allIdentifiers, nil
}
