package customerio

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	IDTypeCIOID  IDType = "cio_id"
	IDTypeUserID IDType = "id"
)

type IDType string

type Customer struct {
	Identifiers `json:",inline"`
	Attributes  `json:",inline"`
}

type Attributes struct {
	Phase1 string `json:"phase1,omitempty"`
	UserID string `json:"user_id"`

	OuraSizingKitDiscountCode string `json:"oura_sizing_kit_discount_code,omitempty"`
	OuraRingDiscountCode      string `json:"oura_ring_discount_code,omitempty"`
	OuraParticipantID         string `json:"oura_participant_id,omitempty"`

	Update bool `json:"_update,omitempty"`
}

type customerResponse struct {
	Customer struct {
		ID          string
		Identifiers Identifiers `json:"identifiers"`
		Attributes  Attributes  `json:"attributes"`
	} `json:"customer"`
}

type entityRequest struct {
	Type        string            `json:"type"`
	Identifiers map[string]string `json:"identifiers"`
	Action      string            `json:"action"`
	Attributes  Attributes        `json:"attributes,omitempty"`
}

type errorResponse struct {
	Errors []struct {
		Reason  string `json:"reason,omitempty"`
		Field   string `json:"field,omitempty"`
		Message string `json:"message,omitempty"`
	} `json:"errors,omitempty"`
}

func (c *Client) GetCustomer(ctx context.Context, cid string, typ IDType) (*Customer, error) {
	url := fmt.Sprintf("%s/v1/customers/%s/attributes", c.appAPIBaseURL, cid)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameter for id_type if using cio_id
	q := req.URL.Query()
	q.Add("id_type", string(typ))
	req.URL.RawQuery = q.Encode()

	// Add authorization header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.appAPIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	} else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response customerResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &Customer{
		Identifiers: response.Customer.Identifiers,
		Attributes:  response.Customer.Attributes,
	}, nil
}

func (c *Client) UpdateCustomer(ctx context.Context, customer Customer) error {
	url := fmt.Sprintf("%s/v2/entity", c.trackAPIBaseURL)

	// Prepare the request body
	reqBody := entityRequest{
		Type:        "person",
		Identifiers: map[string]string{"cio_id": customer.CID},
		Action:      "identify",
		Attributes:  customer.Attributes,
	}
	reqBody.Attributes.Update = true

	jsonBody, err := json.Marshal(reqBody)
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
