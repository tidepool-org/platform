package devicecheck

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
)

const queryTwoBitsPath = "/query_two_bits"

type queryTwoBitsRequestBody struct {
	DeviceToken   string `json:"device_token"`
	TransactionID string `json:"transaction_id"`
	Timestamp     int64  `json:"timestamp"`
}

// QueryTwoBitsResult provides a result of query-two-bits method
type QueryTwoBitsResult struct {
	Bit0           bool   `json:"bit0"`
	Bit1           bool   `json:"bit1"`
	LastUpdateTime dcTime `json:"last_update_time"`
}

type dcTime struct {
	time.Time
}

const timeFormat = "2006-01"

func (t dcTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Format(timeFormat))
}

func (t *dcTime) UnmarshalJSON(b []byte) (err error) {
	t.Time, err = time.Parse(timeFormat, strings.Trim(string(b), `"`))
	return
}

func (api api) queryTwoBits(deviceToken, jwt string) (int, []byte, error) {
	b := queryTwoBitsRequestBody{
		DeviceToken:   deviceToken,
		TransactionID: uuid.New().String(),
		Timestamp:     time.Now().UTC().UnixNano() / int64(time.Millisecond),
	}

	return api.do(jwt, queryTwoBitsPath, b)
}

// QueryTwoBits queries two bits for device token
func (client *Client) QueryTwoBits(deviceToken string, result *QueryTwoBitsResult) error {
	key, err := client.cred.key()
	if err != nil {
		return err
	}

	jwt, err := client.jwt.generate(key)
	if err != nil {
		return err
	}

	code, body, err := client.api.queryTwoBits(deviceToken, jwt)
	if err != nil {
		return err
	}

	if err := newError(code, body); err != nil {
		return err
	}

	return json.Unmarshal(body, result)
}
