package customerio

import (
	"bytes"
	"context"
	"crypto/sha1"
	"net/http"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/tidepool-org/platform/errors"

	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/request"
)

type Event struct {
	Name string    `json:"name"`
	ID   ulid.ULID `json:"id,omitzero"`
	Data any       `json:"data"`
}

// SetDeduplicationID generates ULID that's used for deduplication and using the provided time and the first 10 bytes of the sha1 hashed deduplication ID
// Returns an error if time is before epoch or deduplicationID is empty
func (e *Event) SetDeduplicationID(time time.Time, deduplicationID string) (err error) {
	e.ID, err = CreateUlid(time, deduplicationID)
	return
}

func (c *Client) SendEvent(ctx context.Context, userID string, event *Event) error {
	if event == nil {
		return errors.New("event is missing")
	}

	ctx = log.NewContextWithLogger(ctx, c.logger)
	url := c.trackClient.ConstructURL("api", "v1", "customers", userID, "events")

	mutators := []request.RequestMutator{
		c.trackAPIAuthMutator(),
	}

	if err := c.trackClient.RequestDataWithHTTPClient(ctx, http.MethodPost, url, mutators, event, nil, nil, c.httpClient); err != nil {
		return err
	}

	return nil
}

func CreateUlid(time time.Time, id string) (ulid.ULID, error) {
	if id == "" {
		return ulid.Zero, errors.New("id must not be empty")
	}

	ms := time.UnixMilli()
	if ms < 0 {
		return ulid.Zero, errors.New("time must be after epoch")
	}

	hash := sha1.Sum([]byte(id))
	deduplicationID, err := ulid.New(uint64(ms), bytes.NewReader(hash[:10]))
	if err != nil {
		return ulid.Zero, errors.Wrap(err, "unable to create event id")
	}
	return deduplicationID, nil
}
