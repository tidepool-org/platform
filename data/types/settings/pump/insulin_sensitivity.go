package pump

import (
	"strconv"

	"github.com/tidepool-org/platform/data"
	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	InsulinSensitivityStartMaximum = 86400000
	InsulinSensitivityStartMinimum = 0
)

type InsulinSensitivity struct {
	Amount *float64 `json:"amount,omitempty" bson:"amount,omitempty"`
	Start  *int     `json:"start,omitempty" bson:"start,omitempty"`
}

func ParseInsulinSensitivity(parser data.ObjectParser) *InsulinSensitivity {
	if parser.Object() == nil {
		return nil
	}
	insulinSensitivity := NewInsulinSensitivity()
	insulinSensitivity.Parse(parser)
	parser.ProcessNotParsed()
	return insulinSensitivity
}

func NewInsulinSensitivity() *InsulinSensitivity {
	return &InsulinSensitivity{}
}

func (i *InsulinSensitivity) Parse(parser data.ObjectParser) {
	i.Amount = parser.ParseFloat("amount")
	i.Start = parser.ParseInteger("start")
}

func (i *InsulinSensitivity) Validate(validator structure.Validator, units *string) {
	validator.Float64("amount", i.Amount).Exists().InRange(dataBloodGlucose.ValueRangeForUnits(units))
	validator.Int("start", i.Start).Exists().InRange(InsulinSensitivityStartMinimum, InsulinSensitivityStartMaximum)
}

func (i *InsulinSensitivity) Normalize(normalizer data.Normalizer, units *string) {
	if normalizer.Origin() == structure.OriginExternal {
		i.Amount = dataBloodGlucose.NormalizeValueForUnits(i.Amount, units)
	}
}

// TODO: Can/should we validate that each Start in the array is greater than the previous Start?

type InsulinSensitivityArray []*InsulinSensitivity

func ParseInsulinSensitivityArray(parser data.ArrayParser) *InsulinSensitivityArray {
	if parser.Array() == nil {
		return nil
	}
	insulinSensitivityArray := NewInsulinSensitivityArray()
	insulinSensitivityArray.Parse(parser)
	parser.ProcessNotParsed()
	return insulinSensitivityArray
}

func NewInsulinSensitivityArray() *InsulinSensitivityArray {
	return &InsulinSensitivityArray{}
}

func (i *InsulinSensitivityArray) Parse(parser data.ArrayParser) {
	for index := range *parser.Array() {
		*i = append(*i, ParseInsulinSensitivity(parser.NewChildObjectParser(index)))
	}
}

func (i *InsulinSensitivityArray) Validate(validator structure.Validator, units *string) {
	for index, insulinSensitivity := range *i {
		insulinSensitivityValidator := validator.WithReference(strconv.Itoa(index))
		if insulinSensitivity != nil {
			insulinSensitivity.Validate(insulinSensitivityValidator, units)
		} else {
			insulinSensitivityValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
}

func (i *InsulinSensitivityArray) Normalize(normalizer data.Normalizer, units *string) {
	for index, insulinSensitivity := range *i {
		if insulinSensitivity != nil {
			insulinSensitivity.Normalize(normalizer.WithReference(strconv.Itoa(index)), units)
		}
	}
}
