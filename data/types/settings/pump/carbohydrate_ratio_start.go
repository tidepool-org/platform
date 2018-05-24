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

func ParseCarbohydrateRatioStart(parser data.ObjectParser) *CarbohydrateRatioStart {
	if parser.Object() == nil {
		return nil
	}
	datum := NewCarbohydrateRatioStart()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewCarbohydrateRatioStart() *CarbohydrateRatioStart {
	return &CarbohydrateRatioStart{}
}

func (c *CarbohydrateRatioStart) Parse(parser data.ObjectParser) {
	c.Amount = parser.ParseFloat("amount")
	c.Start = parser.ParseInteger("start")
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

func ParseCarbohydrateRatioStartArray(parser data.ArrayParser) *CarbohydrateRatioStartArray {
	if parser.Array() == nil {
		return nil
	}
	datum := NewCarbohydrateRatioStartArray()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewCarbohydrateRatioStartArray() *CarbohydrateRatioStartArray {
	return &CarbohydrateRatioStartArray{}
}

func (c *CarbohydrateRatioStartArray) Parse(parser data.ArrayParser) {
	for index := range *parser.Array() {
		*c = append(*c, ParseCarbohydrateRatioStart(parser.NewChildObjectParser(index)))
	}
}

func (c *CarbohydrateRatioStartArray) Validate(validator structure.Validator) {
	startMinimum := pointer.FromInt(CarbohydrateRatioStartStartMinimum)
	for index, datum := range *c {
		datumValidator := validator.WithReference(strconv.Itoa(index))
		if datum != nil {
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

func ParseCarbohydrateRatioStartArrayMap(parser data.ObjectParser) *CarbohydrateRatioStartArrayMap {
	if parser.Object() == nil {
		return nil
	}
	datum := NewCarbohydrateRatioStartArrayMap()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewCarbohydrateRatioStartArrayMap() *CarbohydrateRatioStartArrayMap {
	return &CarbohydrateRatioStartArrayMap{}
}

func (c *CarbohydrateRatioStartArrayMap) Parse(parser data.ObjectParser) {
	for name := range *parser.Object() {
		c.Set(name, ParseCarbohydrateRatioStartArray(parser.NewChildArrayParser(name)))
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
