package customerio

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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
	var allIdentifiers []Identifiers
	start := ""

	for {
		url := fmt.Sprintf("%s/v1/segments/%s/membership", c.appAPIBaseURL, segmentID)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Add pagination parameter if available
		if start != "" {
			q := req.URL.Query()
			q.Add("start", start)
			req.URL.RawQuery = q.Encode()
		}

		// Add authorization header
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.appAPIKey))
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to execute request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		var response segmentMembershipResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
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
