package pump

import "github.com/tidepool-org/platform/pvn/data"

type BasalSchedule struct {
	Rate  *float64 `json:"rate" bson:"rate"`
	Start *int     `json:"start" bson:"start"`
}

func NewBasalSchedule() *BasalSchedule {
	return &BasalSchedule{}
}

func (b *BasalSchedule) Parse(parser data.ObjectParser) {
	b.Rate = parser.ParseFloat("rate")
	b.Start = parser.ParseInteger("start")
}

func (b *BasalSchedule) Validate(validator data.Validator) {
	validator.ValidateFloat("rate", b.Rate).Exists().InRange(0.0, 20.0)
	validator.ValidateInteger("start", b.Start).Exists().InRange(0, 86400000)
}

func (b *BasalSchedule) Normalize(normalizer data.Normalizer) {
}

func ParseBasalSchedule(parser data.ObjectParser) *BasalSchedule {
	var basalSchedule *BasalSchedule
	if parser.Object() != nil {
		basalSchedule = NewBasalSchedule()
		basalSchedule.Parse(parser)
	}
	return basalSchedule
}

func ParseBasalScheduleArray(parser data.ArrayParser) *map[string][]*BasalSchedule {
	var basalScheduleArray *map[string][]*BasalSchedule
	if parser.Array() != nil {
		basalScheduleArray = &map[string][]*BasalSchedule{}

		// for index := range *parser.Array() {
		// 	*basalScheduleArray[key] = ParseBasalSchedule(parser.NewChildObjectParser(key))
		// }
	}
	return basalScheduleArray
}
