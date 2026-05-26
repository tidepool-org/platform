package times

import (
	"encoding/json"
	"time"

	"github.com/tidepool-org/platform/crypto"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
)

type TimeRange struct {
	From *time.Time `json:"from,omitempty" bson:"from,omitempty"`
	To   *time.Time `json:"to,omitempty" bson:"to,omitempty"`
}

func ParseTimeRange(parser structure.ObjectParser) *TimeRange {
	if !parser.Exists() {
		return nil
	}
	datum := &TimeRange{}
	datum.Parse(parser)
	return datum
}

func (t *TimeRange) Parse(parser structure.ObjectParser) {
	t.From = parser.Time("from", time.RFC3339Nano)
	t.To = parser.Time("to", time.RFC3339Nano)
}

func (t *TimeRange) Validate(validator structure.Validator) {
	validator.Time("from", t.From).NotZero()
	validator.Time("to", t.To).NotZero().After(pointer.Default(t.From, time.Time{}))
}

func (t TimeRange) InLocation(location *time.Location) TimeRange {
	if location == nil {
		return t
	}
	inLocation := TimeRange{}
	if t.From != nil {
		inLocation.From = pointer.From(t.From.In(location))
	}
	if t.To != nil {
		inLocation.To = pointer.From(t.To.In(location))
	}
	return inLocation
}

func (t TimeRange) Clamped(minimum time.Time, maximum time.Time) TimeRange {
	clamped := TimeRange{}
	if t.From != nil {
		clamped.From = pointer.From(Clamp(*t.From, minimum, maximum))
	}
	if t.To != nil {
		clamped.To = pointer.From(Clamp(*t.To, minimum, maximum))
	}
	return clamped
}

func (t TimeRange) Truncated(duration time.Duration) TimeRange {
	truncated := TimeRange{}
	if t.From != nil {
		truncated.From = pointer.From(t.From.Truncate(duration))
	}
	if t.To != nil {
		truncated.To = pointer.From(t.To.Truncate(duration))
	}
	return truncated
}

func (t TimeRange) Date() TimeRange {
	date := TimeRange{}
	if t.From != nil {
		date.From = pointer.From(time.Date(t.From.Year(), t.From.Month(), t.From.Day(), 0, 0, 0, 0, t.From.Location()))
	}
	if t.To != nil {
		date.To = pointer.From(time.Date(t.To.Year(), t.To.Month(), t.To.Day(), 0, 0, 0, 0, t.To.Location()))
	}
	return date
}

func (t TimeRange) Hash() (string, error) {
	if bites, err := json.Marshal(t); err != nil {
		return "", errors.Wrap(err, "unable to generate hash")
	} else {
		return crypto.Base64EncodedSHA256Hash(bites), nil
	}
}

const MetadataKeyTimeRange = "timeRange"

type TimeRangeMetadata struct {
	TimeRange *TimeRange `json:"timeRange,omitempty" bson:"timeRange,omitempty"`
}

func (t *TimeRangeMetadata) Parse(parser structure.ObjectParser) {
	t.TimeRange = ParseTimeRange(parser.WithReferenceObjectParser(MetadataKeyTimeRange))
}

func (t *TimeRangeMetadata) Validate(validator structure.Validator) {
	if t.TimeRange != nil {
		t.TimeRange.Validate(validator.WithReference(MetadataKeyTimeRange))
	}
}
