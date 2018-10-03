package pump

import (
	"sort"
	"strconv"

	"github.com/tidepool-org/platform/data"
	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	InsulinSensitivityStartStartMaximum = 86400000
	InsulinSensitivityStartStartMinimum = 0
)

type InsulinSensitivityStart struct {
	Amount *float64 `json:"amount,omitempty" bson:"amount,omitempty"`
	Start  *int     `json:"start,omitempty" bson:"start,omitempty"`
}

func ParseInsulinSensitivityStart(parser data.ObjectParser) *InsulinSensitivityStart {
	if parser.Object() == nil {
		return nil
	}
	datum := NewInsulinSensitivityStart()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewInsulinSensitivityStart() *InsulinSensitivityStart {
	return &InsulinSensitivityStart{}
}

func (i *InsulinSensitivityStart) Parse(parser data.ObjectParser) {
	i.Amount = parser.ParseFloat("amount")
	i.Start = parser.ParseInteger("start")
}

func (i *InsulinSensitivityStart) Validate(validator structure.Validator, units *string, startMinimum *int) {
	validator.Float64("amount", i.Amount).Exists().InRange(dataBloodGlucose.ValueRangeForUnits(units))
	startValidator := validator.Int("start", i.Start).Exists()
	if startMinimum != nil {
		if *startMinimum == InsulinSensitivityStartStartMinimum {
			startValidator.EqualTo(InsulinSensitivityStartStartMinimum)
		} else {
			startValidator.InRange(*startMinimum, InsulinSensitivityStartStartMaximum)
		}
	} else {
		startValidator.InRange(InsulinSensitivityStartStartMinimum, InsulinSensitivityStartStartMaximum)
	}
}

func (i *InsulinSensitivityStart) Normalize(normalizer data.Normalizer, units *string) {
	if normalizer.Origin() == structure.OriginExternal {
		i.Amount = dataBloodGlucose.NormalizeValueForUnits(i.Amount, units)
	}
}

type InsulinSensitivityStartArray []*InsulinSensitivityStart

func ParseInsulinSensitivityStartArray(parser data.ArrayParser) *InsulinSensitivityStartArray {
	if parser.Array() == nil {
		return nil
	}
	datum := NewInsulinSensitivityStartArray()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewInsulinSensitivityStartArray() *InsulinSensitivityStartArray {
	return &InsulinSensitivityStartArray{}
}

func (i *InsulinSensitivityStartArray) Parse(parser data.ArrayParser) {
	for index := range *parser.Array() {
		*i = append(*i, ParseInsulinSensitivityStart(parser.NewChildObjectParser(index)))
	}
}

func (i *InsulinSensitivityStartArray) Validate(validator structure.Validator, units *string) {
	startMinimum := pointer.FromInt(InsulinSensitivityStartStartMinimum)
	for index, datum := range *i {
		datumValidator := validator.WithReference(strconv.Itoa(index))
		if datum != nil {
			datum.Validate(datumValidator, units, startMinimum)
			if index == 0 {
				startMinimum = pointer.FromInt(InsulinSensitivityStartStartMinimum + 1)
			} else if datum.Start != nil {
				startMinimum = pointer.FromInt(*datum.Start + 1)
			}
		} else {
			datumValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
}

func (i *InsulinSensitivityStartArray) Normalize(normalizer data.Normalizer, units *string) {
	for index, datum := range *i {
		if datum != nil {
			datum.Normalize(normalizer.WithReference(strconv.Itoa(index)), units)
		}
	}
}

func (i *InsulinSensitivityStartArray) First() *InsulinSensitivityStart {
	if len(*i) > 0 {
		return (*i)[0]
	}
	return nil
}

func (i *InsulinSensitivityStartArray) Last() *InsulinSensitivityStart {
	if length := len(*i); length > 0 {
		return (*i)[length-1]
	}
	return nil
}

type InsulinSensitivityStartArrayMap map[string]*InsulinSensitivityStartArray

func ParseInsulinSensitivityStartArrayMap(parser data.ObjectParser) *InsulinSensitivityStartArrayMap {
	if parser.Object() == nil {
		return nil
	}
	datum := NewInsulinSensitivityStartArrayMap()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewInsulinSensitivityStartArrayMap() *InsulinSensitivityStartArrayMap {
	return &InsulinSensitivityStartArrayMap{}
}

func (i *InsulinSensitivityStartArrayMap) Parse(parser data.ObjectParser) {
	for name := range *parser.Object() {
		i.Set(name, ParseInsulinSensitivityStartArray(parser.NewChildArrayParser(name)))
	}
}

func (i *InsulinSensitivityStartArrayMap) Validate(validator structure.Validator, units *string) {
	for _, name := range i.sortedNames() {
		datumArrayValidator := validator.WithReference(name)
		if datumArray := i.Get(name); datumArray != nil {
			datumArray.Validate(datumArrayValidator, units)
		} else {
			datumArrayValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
}

func (i *InsulinSensitivityStartArrayMap) Normalize(normalizer data.Normalizer, units *string) {
	for _, name := range i.sortedNames() {
		if datumArray := i.Get(name); datumArray != nil {
			datumArray.Normalize(normalizer.WithReference(name), units)
		}
	}
}

func (i *InsulinSensitivityStartArrayMap) Get(name string) *InsulinSensitivityStartArray {
	if datumArray, exists := (*i)[name]; exists {
		return datumArray
	}
	return nil
}

func (i *InsulinSensitivityStartArrayMap) Set(name string, datumArray *InsulinSensitivityStartArray) {
	(*i)[name] = datumArray
}

func (i *InsulinSensitivityStartArrayMap) sortedNames() []string {
	names := []string{}
	for name := range *i {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
