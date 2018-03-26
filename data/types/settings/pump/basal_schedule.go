package pump

import (
	"sort"
	"strconv"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	BasalScheduleRateMaximum  = 100.0
	BasalScheduleRateMinimum  = 0.0
	BasalScheduleStartMaximum = 86400000
	BasalScheduleStartMinimum = 0
)

type BasalSchedule struct {
	Rate  *float64 `json:"rate,omitempty" bson:"rate,omitempty"`
	Start *int     `json:"start,omitempty" bson:"start,omitempty"`
}

func ParseBasalSchedule(parser data.ObjectParser) *BasalSchedule {
	if parser.Object() == nil {
		return nil
	}
	basalSchedule := NewBasalSchedule()
	basalSchedule.Parse(parser)
	parser.ProcessNotParsed()
	return basalSchedule
}

func NewBasalSchedule() *BasalSchedule {
	return &BasalSchedule{}
}

func (b *BasalSchedule) Parse(parser data.ObjectParser) {
	b.Rate = parser.ParseFloat("rate")
	b.Start = parser.ParseInteger("start")
}

func (b *BasalSchedule) Validate(validator structure.Validator) {
	validator.Float64("rate", b.Rate).Exists().InRange(BasalScheduleRateMinimum, BasalScheduleRateMaximum)
	validator.Int("start", b.Start).Exists().InRange(BasalScheduleStartMinimum, BasalScheduleStartMaximum)
}

func (b *BasalSchedule) Normalize(normalizer data.Normalizer) {}

// TODO: Can/should we validate that each Start in the array is greater than the previous Start?

type BasalScheduleArray []*BasalSchedule

func ParseBasalScheduleArray(parser data.ArrayParser) *BasalScheduleArray {
	if parser.Array() == nil {
		return nil
	}
	basalScheduleArray := NewBasalScheduleArray()
	basalScheduleArray.Parse(parser)
	parser.ProcessNotParsed()
	return basalScheduleArray
}

func NewBasalScheduleArray() *BasalScheduleArray {
	return &BasalScheduleArray{}
}

func (b *BasalScheduleArray) Parse(parser data.ArrayParser) {
	for index := range *parser.Array() {
		*b = append(*b, ParseBasalSchedule(parser.NewChildObjectParser(index)))
	}
}

func (b *BasalScheduleArray) Validate(validator structure.Validator) {
	for index, basalSchedule := range *b {
		basalScheduleValidator := validator.WithReference(strconv.Itoa(index))
		if basalSchedule != nil {
			basalSchedule.Validate(basalScheduleValidator)
		} else {
			basalScheduleValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
}

func (b *BasalScheduleArray) Normalize(normalizer data.Normalizer) {
	for index, basalSchedule := range *b {
		if basalSchedule != nil {
			basalSchedule.Normalize(normalizer.WithReference(strconv.Itoa(index)))
		}
	}
}

type BasalScheduleArrayMap map[string]*BasalScheduleArray

func ParseBasalScheduleArrayMap(parser data.ObjectParser) *BasalScheduleArrayMap {
	if parser.Object() == nil {
		return nil
	}
	datum := NewBasalScheduleArrayMap()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewBasalScheduleArrayMap() *BasalScheduleArrayMap {
	return &BasalScheduleArrayMap{}
}

func (b *BasalScheduleArrayMap) Parse(parser data.ObjectParser) {
	for name := range *parser.Object() {
		b.Set(name, ParseBasalScheduleArray(parser.NewChildArrayParser(name)))
	}
}

func (b *BasalScheduleArrayMap) Validate(validator structure.Validator) {
	for _, name := range b.sortedNames() {
		arrayValidator := validator.WithReference(name)
		if array := b.Get(name); array != nil {
			array.Validate(arrayValidator)
		} else {
			arrayValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
}

func (b *BasalScheduleArrayMap) Normalize(normalizer data.Normalizer) {
	for _, name := range b.sortedNames() {
		if array := b.Get(name); array != nil {
			array.Normalize(normalizer.WithReference(name))
		}
	}
}

func (b *BasalScheduleArrayMap) Get(name string) *BasalScheduleArray {
	if array, exists := (*b)[name]; exists {
		return array
	}
	return nil
}

func (b *BasalScheduleArrayMap) Set(name string, array *BasalScheduleArray) {
	(*b)[name] = array
}

func (b *BasalScheduleArrayMap) sortedNames() []string {
	names := []string{}
	for name := range *b {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
