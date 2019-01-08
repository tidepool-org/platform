package pump

import (
	"sort"
	"strconv"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	BasalRateStartRateMaximum  = 100.0
	BasalRateStartRateMinimum  = 0.0
	BasalRateStartStartMaximum = 86400000
	BasalRateStartStartMinimum = 0
)

type BasalRateStart struct {
	Rate  *float64 `json:"rate,omitempty" bson:"rate,omitempty"`
	Start *int     `json:"start,omitempty" bson:"start,omitempty"`
}

func ParseBasalRateStart(parser structure.ObjectParser) *BasalRateStart {
	if !parser.Exists() {
		return nil
	}
	datum := NewBasalRateStart()
	parser.Parse(datum)
	return datum
}

func NewBasalRateStart() *BasalRateStart {
	return &BasalRateStart{}
}

func (b *BasalRateStart) Parse(parser structure.ObjectParser) {
	b.Rate = parser.Float64("rate")
	b.Start = parser.Int("start")
}

func (b *BasalRateStart) Validate(validator structure.Validator, startMinimum *int) {
	validator.Float64("rate", b.Rate).Exists().InRange(BasalRateStartRateMinimum, BasalRateStartRateMaximum)
	startValidator := validator.Int("start", b.Start).Exists()
	if startMinimum != nil {
		if *startMinimum == BasalRateStartStartMinimum {
			startValidator.EqualTo(BasalRateStartStartMinimum)
		} else {
			startValidator.InRange(*startMinimum, BasalRateStartStartMaximum)
		}
	} else {
		startValidator.InRange(BasalRateStartStartMinimum, BasalRateStartStartMaximum)
	}
}

func (b *BasalRateStart) Normalize(normalizer data.Normalizer) {}

type BasalRateStartArray []*BasalRateStart

func ParseBasalRateStartArray(parser structure.ArrayParser) *BasalRateStartArray {
	if !parser.Exists() {
		return nil
	}
	datum := NewBasalRateStartArray()
	parser.Parse(datum)
	return datum
}

func NewBasalRateStartArray() *BasalRateStartArray {
	return &BasalRateStartArray{}
}

func (b *BasalRateStartArray) Parse(parser structure.ArrayParser) {
	for _, reference := range parser.References() {
		*b = append(*b, ParseBasalRateStart(parser.WithReferenceObjectParser(reference)))
	}
}

func (b *BasalRateStartArray) Validate(validator structure.Validator) {
	startMinimum := pointer.FromInt(BasalRateStartStartMinimum)
	for index, datum := range *b {
		if datumValidator := validator.WithReference(strconv.Itoa(index)); datum != nil {
			datum.Validate(datumValidator, startMinimum)
			if index == 0 {
				startMinimum = pointer.FromInt(BasalRateStartStartMinimum + 1)
			} else if datum.Start != nil {
				startMinimum = pointer.FromInt(*datum.Start + 1)
			}
		} else {
			datumValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
}

func (b *BasalRateStartArray) Normalize(normalizer data.Normalizer) {
	for index, datum := range *b {
		if datum != nil {
			datum.Normalize(normalizer.WithReference(strconv.Itoa(index)))
		}
	}
}

func (b *BasalRateStartArray) First() *BasalRateStart {
	if len(*b) > 0 {
		return (*b)[0]
	}
	return nil
}

func (b *BasalRateStartArray) Last() *BasalRateStart {
	if length := len(*b); length > 0 {
		return (*b)[length-1]
	}
	return nil
}

type BasalRateStartArrayMap map[string]*BasalRateStartArray

func ParseBasalRateStartArrayMap(parser structure.ObjectParser) *BasalRateStartArrayMap {
	if !parser.Exists() {
		return nil
	}
	datum := NewBasalRateStartArrayMap()
	parser.Parse(datum)
	return datum
}

func NewBasalRateStartArrayMap() *BasalRateStartArrayMap {
	return &BasalRateStartArrayMap{}
}

func (b *BasalRateStartArrayMap) Parse(parser structure.ObjectParser) {
	for _, reference := range parser.References() {
		b.Set(reference, ParseBasalRateStartArray(parser.WithReferenceArrayParser(reference)))
	}
}

func (b *BasalRateStartArrayMap) Validate(validator structure.Validator) {
	for _, name := range b.sortedNames() {
		datumArrayValidator := validator.WithReference(name)
		if datumArray := b.Get(name); datumArray != nil {
			datumArray.Validate(datumArrayValidator)
		} else {
			datumArrayValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
}

func (b *BasalRateStartArrayMap) Normalize(normalizer data.Normalizer) {
	for _, name := range b.sortedNames() {
		if datumArray := b.Get(name); datumArray != nil {
			datumArray.Normalize(normalizer.WithReference(name))
		}
	}
}

func (b *BasalRateStartArrayMap) Get(name string) *BasalRateStartArray {
	if datumArray, exists := (*b)[name]; exists {
		return datumArray
	}
	return nil
}

func (b *BasalRateStartArrayMap) Set(name string, datumArray *BasalRateStartArray) {
	(*b)[name] = datumArray
}

func (b *BasalRateStartArrayMap) sortedNames() []string {
	names := []string{}
	for name := range *b {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
