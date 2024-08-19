package devicecheck

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const updateTwoBitsPath = "/update_two_bits"

type updateTwoBitsRequestBody struct {
	DeviceToken   string `json:"device_token"`
	TransactionID string `json:"transaction_id"`
	Timestamp     int64  `json:"timestamp"`
	Bit0          bool   `json:"bit0"`
	Bit1          bool   `json:"bit1"`
}

// UpdateTwoBits updates two bits for device token.
func (client *Client) UpdateTwoBits(deviceToken string, bit0, bit1 bool) error {
	key, err := client.cred.key()
	if err != nil {
		return fmt.Errorf("devicecheck: failed to create key: %w", err)
	}

	jwt, err := client.jwt.generate(key)
	if err != nil {
		return fmt.Errorf("devicecheck: failed to generate jwt: %w", err)
	}

	body := updateTwoBitsRequestBody{
		DeviceToken:   deviceToken,
		TransactionID: uuid.New().String(),
		Timestamp:     time.Now().UTC().UnixNano() / int64(time.Millisecond),
		Bit0:          bit0,
		Bit1:          bit1,
	}

	code, respBody, err := client.api.do(jwt, updateTwoBitsPath, body)
	if err != nil {
		return fmt.Errorf("devicecheck: failed to update two bits: %w", err)
	}

	if code != http.StatusOK {
		return fmt.Errorf("devicecheck: %w", newError(code, respBody))
	}

	return nil
}
