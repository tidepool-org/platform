package work

import (
	"strings"
	"time"

	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
)

const (
	TimeRangeFormat = time.RFC3339Nano
)

type TimeRange struct {
	From *time.Time `json:"from,omitempty" bson:"from,omitempty"`
	To   *time.Time `json:"to,omitempty" bson:"to,omitempty"`
}

func ParseTimeRange(parser structure.ObjectParser) *TimeRange {
	if !parser.Exists() {
		return nil
	}
	datum := NewTimeRange()
	parser.Parse(datum)
	return datum
}

func NewTimeRange() *TimeRange {
	return &TimeRange{}
}

func (t *TimeRange) Parse(parser structure.ObjectParser) {
	t.From = parser.Time("from", TimeRangeFormat)
	t.To = parser.Time("to", TimeRangeFormat)
}

func (t *TimeRange) Validate(validator structure.Validator) {
	validator.Time("from", t.From).NotZero()
	validator.Time("to", t.To).NotZero().After(pointer.DefaultTime(t.From, time.Time{}))
}

func (t TimeRange) Truncate(duration time.Duration) TimeRange {
	truncated := TimeRange{}
	if t.From != nil {
		truncated.From = pointer.FromTime(t.From.Truncate(duration))
	}
	if t.To != nil {
		truncated.To = pointer.FromTime(t.To.Truncate(duration))
	}
	return truncated
}

func (t *TimeRange) String() string {
	var parts []string
	if t.From != nil {
		parts = append(parts, t.From.Format(TimeRangeFormat))
	}
	parts = append(parts, "-")
	if t.To != nil {
		parts = append(parts, t.To.Format(TimeRangeFormat))
	}
	return strings.Join(parts, "")
}
