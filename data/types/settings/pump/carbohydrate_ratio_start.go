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
	CarbohydrateRatioStartAmountMaximum = 250.0
	CarbohydrateRatioStartAmountMinimum = 0.0
	CarbohydrateRatioStartStartMaximum  = 86400000
	CarbohydrateRatioStartStartMinimum  = 0
)

type CarbohydrateRatioStart struct {
	Amount *float64 `json:"amount,omitempty" bson:"amount,omitempty"`
	Start  *int     `json:"start,omitempty" bson:"start,omitempty"`
}

func ParseCarbohydrateRatioStart(parser structure.ObjectParser) *CarbohydrateRatioStart {
	if !parser.Exists() {
		return nil
	}
	datum := NewCarbohydrateRatioStart()
	parser.Parse(datum)
	return datum
}

func NewCarbohydrateRatioStart() *CarbohydrateRatioStart {
	return &CarbohydrateRatioStart{}
}

func (c *CarbohydrateRatioStart) Parse(parser structure.ObjectParser) {
	c.Amount = parser.Float64("amount")
	c.Start = parser.Int("start")
}

func (c *CarbohydrateRatioStart) Validate(validator structure.Validator, startMinimum *int) {
	validator.Float64("amount", c.Amount).Exists().InRange(CarbohydrateRatioStartAmountMinimum, CarbohydrateRatioStartAmountMaximum)
	startValidator := validator.Int("start", c.Start).Exists()
	if startMinimum != nil {
		if *startMinimum == CarbohydrateRatioStartStartMinimum {
			startValidator.EqualTo(CarbohydrateRatioStartStartMinimum)
		} else {
			startValidator.InRange(*startMinimum, CarbohydrateRatioStartStartMaximum)
		}
	} else {
		startValidator.InRange(CarbohydrateRatioStartStartMinimum, CarbohydrateRatioStartStartMaximum)
	}
}

func (c *CarbohydrateRatioStart) Normalize(normalizer data.Normalizer) {}

type CarbohydrateRatioStartArray []*CarbohydrateRatioStart

func ParseCarbohydrateRatioStartArray(parser structure.ArrayParser) *CarbohydrateRatioStartArray {
	if !parser.Exists() {
		return nil
	}
	datum := NewCarbohydrateRatioStartArray()
	parser.Parse(datum)
	return datum
}

func NewCarbohydrateRatioStartArray() *CarbohydrateRatioStartArray {
	return &CarbohydrateRatioStartArray{}
}

func (c *CarbohydrateRatioStartArray) Parse(parser structure.ArrayParser) {
	for _, reference := range parser.References() {
		*c = append(*c, ParseCarbohydrateRatioStart(parser.WithReferenceObjectParser(reference)))
	}
}

func (c *CarbohydrateRatioStartArray) Validate(validator structure.Validator) {
	startMinimum := pointer.FromInt(CarbohydrateRatioStartStartMinimum)
	for index, datum := range *c {
		if datumValidator := validator.WithReference(strconv.Itoa(index)); datum != nil {
			datum.Validate(datumValidator, startMinimum)
			if index == 0 {
				startMinimum = pointer.FromInt(CarbohydrateRatioStartStartMinimum + 1)
			} else if datum.Start != nil {
				startMinimum = pointer.FromInt(*datum.Start + 1)
			}
		} else {
			datumValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
}

func (c *CarbohydrateRatioStartArray) Normalize(normalizer data.Normalizer) {
	for index, datum := range *c {
		if datum != nil {
			datum.Normalize(normalizer.WithReference(strconv.Itoa(index)))
		}
	}
}

func (c *CarbohydrateRatioStartArray) First() *CarbohydrateRatioStart {
	if len(*c) > 0 {
		return (*c)[0]
	}
	return nil
}

func (c *CarbohydrateRatioStartArray) Last() *CarbohydrateRatioStart {
	if length := len(*c); length > 0 {
		return (*c)[length-1]
	}
	return nil
}

type CarbohydrateRatioStartArrayMap map[string]*CarbohydrateRatioStartArray

func ParseCarbohydrateRatioStartArrayMap(parser structure.ObjectParser) *CarbohydrateRatioStartArrayMap {
	if !parser.Exists() {
		return nil
	}
	datum := NewCarbohydrateRatioStartArrayMap()
	parser.Parse(datum)
	return datum
}

func NewCarbohydrateRatioStartArrayMap() *CarbohydrateRatioStartArrayMap {
	return &CarbohydrateRatioStartArrayMap{}
}

func (c *CarbohydrateRatioStartArrayMap) Parse(parser structure.ObjectParser) {
	for _, reference := range parser.References() {
		c.Set(reference, ParseCarbohydrateRatioStartArray(parser.WithReferenceArrayParser(reference)))
	}
}

func (c *CarbohydrateRatioStartArrayMap) Validate(validator structure.Validator) {
	for _, name := range c.sortedNames() {
		datumArrayValidator := validator.WithReference(name)
		if datumArray := c.Get(name); datumArray != nil {
			datumArray.Validate(datumArrayValidator)
		} else {
			datumArrayValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
}

func (c *CarbohydrateRatioStartArrayMap) Normalize(normalizer data.Normalizer) {
	for _, name := range c.sortedNames() {
		if datumArray := c.Get(name); datumArray != nil {
			datumArray.Normalize(normalizer.WithReference(name))
		}
	}
}

func (c *CarbohydrateRatioStartArrayMap) Get(name string) *CarbohydrateRatioStartArray {
	if datumArray, exists := (*c)[name]; exists {
		return datumArray
	}
	return nil
}

func (c *CarbohydrateRatioStartArrayMap) Set(name string, datumArray *CarbohydrateRatioStartArray) {
	(*c)[name] = datumArray
}

func (c *CarbohydrateRatioStartArrayMap) sortedNames() []string {
	names := []string{}
	for name := range *c {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
