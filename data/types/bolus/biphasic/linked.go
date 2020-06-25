package biphasic

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/bolus/normal"
	"github.com/tidepool-org/platform/data/types/common"
	"github.com/tidepool-org/platform/structure"
)

type LinkedBolus struct {
	Normal *float64 `json:"normal,omitempty" bson:"normal,omitempty"`
	// will be using commontypes.Duration
	Duration *common.Duration `json:"duration,omitempty" bson:"duration,omitempty"`
}

func ParseLinkedBolus(parser structure.ObjectParser) *LinkedBolus {
	if !parser.Exists() {
		return nil
	}
	datum := NewLinkedBolus()
	parser.Parse(datum)
	return datum
}

func NewLinkedBolus() *LinkedBolus {
	return &LinkedBolus{}
}

func (l *LinkedBolus) Parse(parser structure.ObjectParser) {
	l.Normal = parser.Float64("normal")
	l.Duration = common.ParseDuration(parser.WithReferenceObjectParser("duration"))
}

func (l *LinkedBolus) Validate(validator structure.Validator) {
	validator.Float64("normal", l.Normal).Exists().InRange(normal.NormalMinimum, normal.NormalMaximum)

	if l.Duration != nil {
		l.Duration.Validate(validator)
	}
}

func (l *LinkedBolus) Normalize(normalizer data.Normalizer) {
	if l.Duration != nil {
		l.Duration.Normalize(normalizer)
	}
}
