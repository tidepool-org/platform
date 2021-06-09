package devicecheck

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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

// QueryTwoBitsResult provides a result of query-two-bits method.
type QueryTwoBitsResult struct {
	Bit0           bool `json:"bit0"`
	Bit1           bool `json:"bit1"`
	LastUpdateTime Time `json:"last_update_time"`
}

type Time struct {
	time.Time
}

const timeFormat = "2006-01"

func (t Time) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Format(timeFormat))
}

func (t *Time) UnmarshalJSON(b []byte) error {
	tm, err := time.Parse(timeFormat, strings.Trim(string(b), `"`))
	if err != nil {
		return fmt.Errorf("time: %w", err)
	}

	t.Time = tm

	return nil
}

// QueryTwoBits queries two bits for device token. Returns ErrBitStateNotFound if the bits have not been set.
func (client *Client) QueryTwoBits(deviceToken string, result *QueryTwoBitsResult) error {
	key, err := client.cred.key()
	if err != nil {
		return fmt.Errorf("devicecheck: failed to create key: %w", err)
	}

	jwt, err := client.jwt.generate(key)
	if err != nil {
		return fmt.Errorf("devicecheck: failed to generate jwt: %w", err)
	}

	body := queryTwoBitsRequestBody{
		DeviceToken:   deviceToken,
		TransactionID: uuid.New().String(),
		Timestamp:     time.Now().UTC().UnixNano() / int64(time.Millisecond),
	}

	resp, err := client.api.do(jwt, queryTwoBitsPath, body)
	if err != nil {
		return fmt.Errorf("devicecheck: failed to query two bits: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("devicecheck: failed to read response body: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		if isErrBitStateNotFound(string(respBody)) {
			return fmt.Errorf("devicecheck: %w", ErrBitStateNotFound)
		}

		return json.NewDecoder(bytes.NewReader(respBody)).Decode(result)
	default:
		return fmt.Errorf("devicecheck: %w", newError(resp.StatusCode, string(respBody)))
	}
}
