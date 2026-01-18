package customerio

import (
	"context"
	"net/http"

	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/request"
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

type FindCustomersResponse struct {
	Identifiers []Identifiers `json:"identifiers"`
}

type entityRequest struct {
	Type        string            `json:"type"`
	Identifiers map[string]string `json:"identifiers"`
	Action      string            `json:"action"`
	Attributes  Attributes        `json:"attributes,omitempty"`
}

func (c *Client) GetCustomer(ctx context.Context, cid string, typ IDType) (*Customer, error) {
	ctx = log.NewContextWithLogger(ctx, c.logger)
	url := c.appClient.ConstructURL("v1", "customers", cid, "attributes")

	mutators := []request.RequestMutator{
		request.NewParameterMutator("id_type", string(typ)),
		c.appAPIAuthMutator(),
	}

	c.logger.WithField("cid", cid).WithField("url", url).Debug("fetching customer")

	var response customerResponse
	if err := c.appClient.RequestDataWithHTTPClient(ctx, http.MethodGet, url, mutators, nil, &response, nil, c.httpClient); err != nil {
		if request.IsErrorResourceNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return &Customer{
		Identifiers: response.Customer.Identifiers,
		Attributes:  response.Customer.Attributes,
	}, nil
}

func (c *Client) FindCustomers(ctx context.Context, filter map[string]any) (*FindCustomersResponse, error) {
	ctx = log.NewContextWithLogger(ctx, c.logger)
	url := c.appClient.ConstructURL("v1", "customers")

	mutators := []request.RequestMutator{
		c.appAPIAuthMutator(),
	}

	c.logger.WithField("url", url).WithField("filter", filter).Debug("finding customer")

	var response FindCustomersResponse
	if err := c.appClient.RequestDataWithHTTPClient(ctx, http.MethodPost, url, mutators, filter, &response, nil, c.httpClient); err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *Client) UpdateCustomer(ctx context.Context, customer Customer) error {
	ctx = log.NewContextWithLogger(ctx, c.logger)
	url := c.trackClient.ConstructURL("api", "v2", "entity")

	// Prepare the request body
	reqBody := entityRequest{
		Type:        "person",
		Identifiers: map[string]string{"cio_id": customer.CID},
		Action:      "identify",
		Attributes:  customer.Attributes,
	}
	reqBody.Attributes.Update = true

	mutators := []request.RequestMutator{
		c.trackAPIAuthMutator(),
	}

	if err := c.trackClient.RequestDataWithHTTPClient(ctx, http.MethodPost, url, mutators, reqBody, nil, nil, c.httpClient); err != nil {
		return err
	}

	return nil
}
