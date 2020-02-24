package devicecheck

import (
	"time"

	"github.com/google/uuid"
)

const validateDeviceTokenPath = "/validate_device_token"

type validateDeviceTokenRequestBody struct {
	DeviceToken   string `json:"device_token"`
	TransactionID string `json:"transaction_id"`
	Timestamp     int64  `json:"timestamp"`
}

func (api api) validateDeviceToken(deviceToken, jwt string) (int, []byte, error) {
	b := validateDeviceTokenRequestBody{
		DeviceToken:   deviceToken,
		TransactionID: uuid.New().String(),
		Timestamp:     time.Now().UTC().UnixNano() / int64(time.Millisecond),
	}

	return api.do(jwt, validateDeviceTokenPath, b)
}

// ValidateDeviceToken validates a device for device token
func (client *Client) ValidateDeviceToken(deviceToken string) error {
	key, err := client.cred.key()
	if err != nil {
		return err
	}

	jwt, err := client.jwt.generate(key)
	if err != nil {
		return err
	}

	code, body, err := client.api.validateDeviceToken(deviceToken, jwt)
	if err != nil {
		return err
	}

	return newError(code, body)
}
