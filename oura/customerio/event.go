package customerio

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Event struct {
	Name string `json:"name"`
	ID   string `json:"id"`
	Data any    `json:"data"`
}

func (c *Client) SendEvent(ctx context.Context, cid string, event Event) error {
	url := fmt.Sprintf("%s/v1/customers/%s/events", c.trackAPIBaseURL, cid)

	jsonBody, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add the authorization header (Basic Auth for Track API)
	req.SetBasicAuth(c.siteID, c.trackAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp errorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil && len(errResp.Errors) > 0 {
			return fmt.Errorf("API error (status %d): %s", resp.StatusCode, errResp.Errors[0].Message)
		}
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
