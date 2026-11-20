package times

import (
	"encoding/json"
	"time"

	"github.com/gowebpki/jcs"

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

// Span returns the smallest time range covering both time ranges, including any gap between them. A nil From or To
// means no information, so it yields to the other time range. Equal times keep this time range's value. If the result
// has From after To, both are nil. Since nil also means no information, the grouping of chained calls can then change
// the result. For example, ({From: 3} ∪ {To: 2}) ∪ {From: 1} is {From: 1}, but {From: 3} ∪ ({To: 2} ∪ {From: 1}) is
// {From: 1, To: 2}, where ∪ is Span.
func (t TimeRange) Span(other TimeRange) TimeRange {
	span := TimeRange{}
	if t.From != nil && (other.From == nil || !other.From.Before(*t.From)) {
		span.From = pointer.Clone(t.From)
	} else {
		span.From = pointer.Clone(other.From)
	}
	if t.To != nil && (other.To == nil || !other.To.After(*t.To)) {
		span.To = pointer.Clone(t.To)
	} else {
		span.To = pointer.Clone(other.To)
	}
	if span.From != nil && span.To != nil && span.From.After(*span.To) {
		return TimeRange{}
	}
	return span
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
		return "", errors.Wrap(err, "unable to generate JSON for hash")
	} else if bites, err = jcs.Transform(bites); err != nil {
		return "", errors.Wrap(err, "unable to canonicalize JSON for hash")
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
