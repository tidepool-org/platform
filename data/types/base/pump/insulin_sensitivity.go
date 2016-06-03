package pump

import "github.com/tidepool-org/platform/data"

type InsulinSensitivity struct {
	Amount *float64 `json:"amount,omitempty" bson:"amount,omitempty"`
	Start  *int     `json:"start,omitempty" bson:"start,omitempty"`
}

func NewInsulinSensitivity() *InsulinSensitivity {
	return &InsulinSensitivity{}
}

func (i *InsulinSensitivity) Parse(parser data.ObjectParser) {
	i.Amount = parser.ParseFloat("amount")
	i.Start = parser.ParseInteger("start")
}

func (i *InsulinSensitivity) Validate(validator data.Validator, units *string) {
	validator.ValidateFloatAsBloodGlucoseValue("amount", i.Amount).Exists().InRangeForUnits(units)
	validator.ValidateInteger("start", i.Start).Exists().InRange(0, 86400000)
}

func (i *InsulinSensitivity) Normalize(normalizer data.Normalizer, units *string) {
	i.Amount = normalizer.NormalizeBloodGlucose(units).Value(i.Amount)
}

func ParseInsulinSensitivity(parser data.ObjectParser) *InsulinSensitivity {
	var insulinSensitivity *InsulinSensitivity
	if parser.Object() != nil {
		insulinSensitivity = NewInsulinSensitivity()
		insulinSensitivity.Parse(parser)
		parser.ProcessNotParsed()
	}
	return insulinSensitivity
}

func ParseInsulinSensitivityArray(parser data.ArrayParser) *[]*InsulinSensitivity {
	var insulinSensitivityArray *[]*InsulinSensitivity
	if parser.Array() != nil {
		insulinSensitivityArray = &[]*InsulinSensitivity{}
		for index := range *parser.Array() {
			*insulinSensitivityArray = append(*insulinSensitivityArray, ParseInsulinSensitivity(parser.NewChildObjectParser(index)))
		}
		parser.ProcessNotParsed()
	}
	return insulinSensitivityArray
}
