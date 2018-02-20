package pump

import (
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

// TODO: Can we have multiple schedules with the same name?
// TODO: Can we have an empty name (i.e. "")?

type BasalScheduleNameArrayEntry struct {
	Name  string
	Array *BasalScheduleArray
}

type BasalScheduleArrayMap struct {
	Entries []BasalScheduleNameArrayEntry
}

func ParseBasalScheduleArrayMap(parser data.ObjectParser) *BasalScheduleArrayMap {
	if parser.Object() == nil {
		return nil
	}
	basalScheduleArrayMap := NewBasalScheduleArrayMap()
	basalScheduleArrayMap.Parse(parser)
	parser.ProcessNotParsed()
	return basalScheduleArrayMap
}

func NewBasalScheduleArrayMap() *BasalScheduleArrayMap {
	return &BasalScheduleArrayMap{}
}

func (b *BasalScheduleArrayMap) Parse(parser data.ObjectParser) {
	for key := range *parser.Object() {
		b.Set(key, ParseBasalScheduleArray(parser.NewChildArrayParser(key)))
	}
}

func (b *BasalScheduleArrayMap) Validate(validator structure.Validator) {
	for _, entry := range b.Entries {
		arrayValidator := validator.WithReference(entry.Name)
		if entry.Array != nil {
			entry.Array.Validate(arrayValidator)
		} else {
			arrayValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
}

func (b *BasalScheduleArrayMap) Normalize(normalizer data.Normalizer) {
	for _, entry := range b.Entries {
		if entry.Array != nil {
			entry.Array.Normalize(normalizer.WithReference(entry.Name))
		}
	}
}

func (b *BasalScheduleArrayMap) Get(name string) *BasalScheduleArray {
	if index := b.find(name); index != -1 {
		return b.Entries[index].Array
	}
	return nil
}

func (b *BasalScheduleArrayMap) Set(name string, basalScheduleArray *BasalScheduleArray) {
	if index := b.find(name); index != -1 {
		b.Entries = append(b.Entries[:index], b.Entries[index+1:]...)
	}
	b.Entries = append(b.Entries, BasalScheduleNameArrayEntry{Name: name, Array: basalScheduleArray})
}

func (b *BasalScheduleArrayMap) find(name string) int {
	for index, entry := range b.Entries {
		if entry.Name == name {
			return index
		}
	}
	return -1
}
