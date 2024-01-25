package devicecheck

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const validateDeviceTokenPath = "/validate_device_token"

type validateDeviceTokenRequestBody struct {
	DeviceToken   string `json:"device_token"`
	TransactionID string `json:"transaction_id"`
	Timestamp     int64  `json:"timestamp"`
}

// ValidateDeviceToken validates a device for device token.
func (client *Client) ValidateDeviceToken(deviceToken string) error {
	key, err := client.cred.key()
	if err != nil {
		return fmt.Errorf("devicecheck: failed to create key: %w", err)
	}

	jwt, err := client.jwt.generate(key)
	if err != nil {
		return fmt.Errorf("devicecheck: failed to generate jwt: %w", err)
	}

	body := validateDeviceTokenRequestBody{
		DeviceToken:   deviceToken,
		TransactionID: uuid.New().String(),
		Timestamp:     time.Now().UTC().UnixNano() / int64(time.Millisecond),
	}

	code, respBody, err := client.api.do(jwt, validateDeviceTokenPath, body)
	if err != nil {
		return fmt.Errorf("devicecheck: failed to validate device token: %w", err)
	}

	if code != http.StatusOK {
		return fmt.Errorf("devicecheck: %w", newError(code, respBody))
	}

	return nil
}
