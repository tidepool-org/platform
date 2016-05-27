package pump

import "github.com/tidepool-org/platform/data"

type BasalSchedule struct {
	Rate  *float64 `json:"rate,omitempty" bson:"rate,omitempty"`
	Start *int     `json:"start,omitempty" bson:"start,omitempty"`
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

func parseScheduleItem(parser data.ObjectParser) *BasalSchedule {
	basalSchedule := &BasalSchedule{}
	if parser.Object() != nil {
		basalSchedule = NewBasalSchedule()
		basalSchedule.Parse(parser)
	}
	return basalSchedule
}

func ParseBasalScheduleArray(parser data.ArrayParser) *[]*BasalSchedule {
	basalScheduleArray := &[]*BasalSchedule{}
	if parser.Array() != nil {
		for index := range *parser.Array() {
			*basalScheduleArray = append(*basalScheduleArray, parseScheduleItem(parser.NewChildObjectParser(index)))
		}
	}
	return basalScheduleArray
}

func ParseBasalSchedulesMap(parser data.ObjectParser) *map[string]*[]*BasalSchedule {
	basalScheduleMap := map[string]*[]*BasalSchedule{}
	if parser.Object() != nil {
		for key := range *parser.Object() {
			basalScheduleMap[key] = ParseBasalScheduleArray(parser.NewChildArrayParser(key))
		}
	}
	return &basalScheduleMap
}
