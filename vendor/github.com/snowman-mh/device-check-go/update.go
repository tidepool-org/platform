package devicecheck

import (
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

func (api api) updateTwoBits(deviceToken, jwt string, bit0, bit1 bool) (int, []byte, error) {
	b := updateTwoBitsRequestBody{
		DeviceToken:   deviceToken,
		TransactionID: uuid.New().String(),
		Timestamp:     time.Now().UTC().UnixNano() / int64(time.Millisecond),
		Bit0:          bit0,
		Bit1:          bit1,
	}

	return api.do(jwt, updateTwoBitsPath, b)
}

// UpdateTwoBits updates two bits for device token
func (client *Client) UpdateTwoBits(deviceToken string, bit0, bit1 bool) error {
	key, err := client.cred.key()
	if err != nil {
		return err
	}

	jwt, err := client.jwt.generate(key)
	if err != nil {
		return err
	}

	code, body, err := client.api.updateTwoBits(deviceToken, jwt, bit0, bit1)
	if err != nil {
		return err
	}

	return newError(code, body)
}
