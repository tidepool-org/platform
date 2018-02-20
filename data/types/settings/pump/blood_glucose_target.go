package pump

import (
	"strconv"

	"github.com/tidepool-org/platform/data"
	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	BloodGlucoseTargetStartMaximum = 86400000
	BloodGlucoseTargetStartMinimum = 0
)

type BloodGlucoseTarget struct {
	dataBloodGlucose.Target `bson:",inline"`

	Start *int `json:"start,omitempty" bson:"start,omitempty"`
}

func ParseBloodGlucoseTarget(parser data.ObjectParser) *BloodGlucoseTarget {
	if parser.Object() == nil {
		return nil
	}
	bloodGlucoseTarget := NewBloodGlucoseTarget()
	bloodGlucoseTarget.Parse(parser)
	parser.ProcessNotParsed()
	return bloodGlucoseTarget
}

func NewBloodGlucoseTarget() *BloodGlucoseTarget {
	return &BloodGlucoseTarget{}
}

func (b *BloodGlucoseTarget) Parse(parser data.ObjectParser) {
	b.Target.Parse(parser)

	b.Start = parser.ParseInteger("start")
}

func (b *BloodGlucoseTarget) Validate(validator structure.Validator, units *string) {
	b.Target.Validate(validator, units)

	validator.Int("start", b.Start).Exists().InRange(BloodGlucoseTargetStartMinimum, BloodGlucoseTargetStartMaximum)
}

func (b *BloodGlucoseTarget) Normalize(normalizer data.Normalizer, units *string) {
	b.Target.Normalize(normalizer, units)
}

// TODO: Can/should we validate that each Start in the array is greater than the previous Start?

type BloodGlucoseTargetArray []*BloodGlucoseTarget

func ParseBloodGlucoseTargetArray(parser data.ArrayParser) *BloodGlucoseTargetArray {
	if parser.Array() == nil {
		return nil
	}
	bloodGlucoseTargetArray := NewBloodGlucoseTargetArray()
	bloodGlucoseTargetArray.Parse(parser)
	parser.ProcessNotParsed()
	return bloodGlucoseTargetArray
}

func NewBloodGlucoseTargetArray() *BloodGlucoseTargetArray {
	return &BloodGlucoseTargetArray{}
}

func (b *BloodGlucoseTargetArray) Parse(parser data.ArrayParser) {
	for index := range *parser.Array() {
		*b = append(*b, ParseBloodGlucoseTarget(parser.NewChildObjectParser(index)))
	}
}

func (b *BloodGlucoseTargetArray) Validate(validator structure.Validator, units *string) {
	for index, bloodGlucoseTarget := range *b {
		bloodGlucoseTargetValidator := validator.WithReference(strconv.Itoa(index))
		if bloodGlucoseTarget != nil {
			bloodGlucoseTarget.Validate(bloodGlucoseTargetValidator, units)
		} else {
			bloodGlucoseTargetValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
}

func (b *BloodGlucoseTargetArray) Normalize(normalizer data.Normalizer, units *string) {
	for index, bloodGlucoseTarget := range *b {
		if bloodGlucoseTarget != nil {
			bloodGlucoseTarget.Normalize(normalizer.WithReference(strconv.Itoa(index)), units)
		}
	}
}
