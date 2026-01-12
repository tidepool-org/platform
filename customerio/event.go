package customerio

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/tidepool-org/platform/errors"
)

type Event struct {
	Name string `json:"name"`
	ID   string `json:"id"`
	Data any    `json:"data"`
}

func (c *Client) SendEvent(ctx context.Context, cid string, event Event) error {
	url := fmt.Sprintf("%s/api/v1/customers/%s/events", c.config.TrackAPIBaseURL, cid)

	jsonBody, err := json.Marshal(event)
	if err != nil {
		return errors.Wrap(err, "failed to marshal request body")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return errors.Wrap(err, "failed to create request")
	}

	// Add the authorization header (Basic Auth for Track API)
	req.SetBasicAuth(c.config.SiteID, c.config.TrackAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to execute request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp errorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil && len(errResp.Errors) > 0 {
			return errors.Newf("API error (status %d): %s", resp.StatusCode, errResp.Errors[0].Message)
		}
		return errors.Newf("unexpected status code: %s", resp.Status)
	}

	return nil
}
