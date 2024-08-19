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
	BloodGlucoseTargetStartStartMaximum = 86400000
	BloodGlucoseTargetStartStartMinimum = 0
)

type BloodGlucoseTargetStart struct {
	dataBloodGlucose.Target `bson:",inline"`

	Start *int `json:"start,omitempty" bson:"start,omitempty"`
}

func ParseBloodGlucoseTargetStart(parser structure.ObjectParser) *BloodGlucoseTargetStart {
	if !parser.Exists() {
		return nil
	}
	datum := NewBloodGlucoseTargetStart()
	parser.Parse(datum)
	return datum
}

func NewBloodGlucoseTargetStart() *BloodGlucoseTargetStart {
	return &BloodGlucoseTargetStart{}
}

func (b *BloodGlucoseTargetStart) Parse(parser structure.ObjectParser) {
	b.Target.Parse(parser)

	b.Start = parser.Int("start")
}

func (b *BloodGlucoseTargetStart) Validate(validator structure.Validator, units *string, startMinimum *int) {
	b.Target.Validate(validator, units)

	startValidator := validator.Int("start", b.Start).Exists()
	if startMinimum != nil {
		if *startMinimum == BloodGlucoseTargetStartStartMinimum {
			startValidator.EqualTo(BloodGlucoseTargetStartStartMinimum)
		} else {
			startValidator.InRange(*startMinimum, BloodGlucoseTargetStartStartMaximum)
		}
	} else {
		startValidator.InRange(BloodGlucoseTargetStartStartMinimum, BloodGlucoseTargetStartStartMaximum)
	}
}

func (b *BloodGlucoseTargetStart) Normalize(normalizer data.Normalizer, units *string) {
	b.Target.Normalize(normalizer, units)
}

type BloodGlucoseTargetStartArray []*BloodGlucoseTargetStart

func ParseBloodGlucoseTargetStartArray(parser structure.ArrayParser) *BloodGlucoseTargetStartArray {
	if !parser.Exists() {
		return nil
	}
	datum := NewBloodGlucoseTargetStartArray()
	parser.Parse(datum)
	return datum
}

func NewBloodGlucoseTargetStartArray() *BloodGlucoseTargetStartArray {
	return &BloodGlucoseTargetStartArray{}
}

func (b *BloodGlucoseTargetStartArray) Parse(parser structure.ArrayParser) {
	for _, reference := range parser.References() {
		*b = append(*b, ParseBloodGlucoseTargetStart(parser.WithReferenceObjectParser(reference)))
	}
}

func (b *BloodGlucoseTargetStartArray) Validate(validator structure.Validator, units *string) {
	startMinimum := pointer.FromInt(BloodGlucoseTargetStartStartMinimum)
	for index, datum := range *b {
		if datumValidator := validator.WithReference(strconv.Itoa(index)); datum != nil {
			datum.Validate(datumValidator, units, startMinimum)
			if index == 0 {
				startMinimum = pointer.FromInt(BloodGlucoseTargetStartStartMinimum + 1)
			} else if datum.Start != nil {
				startMinimum = pointer.FromInt(*datum.Start + 1)
			}
		} else {
			datumValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
}

func (b *BloodGlucoseTargetStartArray) Normalize(normalizer data.Normalizer, units *string) {
	for index, datum := range *b {
		if datum != nil {
			datum.Normalize(normalizer.WithReference(strconv.Itoa(index)), units)
		}
	}
}

func (b *BloodGlucoseTargetStartArray) First() *BloodGlucoseTargetStart {
	if len(*b) > 0 {
		return (*b)[0]
	}
	return nil
}

func (b *BloodGlucoseTargetStartArray) Last() *BloodGlucoseTargetStart {
	if length := len(*b); length > 0 {
		return (*b)[length-1]
	}
	return nil
}

func (b *BloodGlucoseTargetStartArray) GetBounds() *dataBloodGlucose.Bounds {
	allBounds := make([]*dataBloodGlucose.Bounds, 0)
	for _, v := range *b {
		if b := v.GetBounds(); b != nil {
			allBounds = append(allBounds, b)
		}
	}

	if len(allBounds) == 0 {
		return nil
	}

	bounds := dataBloodGlucose.Bounds{
		Lower: allBounds[0].Lower,
		Upper: allBounds[0].Upper,
	}
	for _, b := range allBounds {
		if b.Lower < bounds.Lower {
			bounds.Lower = b.Lower
		}
		if b.Upper > bounds.Upper {
			bounds.Upper = b.Upper
		}
	}

	return &bounds
}

type BloodGlucoseTargetStartArrayMap map[string]*BloodGlucoseTargetStartArray

func ParseBloodGlucoseTargetStartArrayMap(parser structure.ObjectParser) *BloodGlucoseTargetStartArrayMap {
	if !parser.Exists() {
		return nil
	}
	datum := NewBloodGlucoseTargetStartArrayMap()
	parser.Parse(datum)
	return datum
}

func NewBloodGlucoseTargetStartArrayMap() *BloodGlucoseTargetStartArrayMap {
	return &BloodGlucoseTargetStartArrayMap{}
}

func (b *BloodGlucoseTargetStartArrayMap) Parse(parser structure.ObjectParser) {
	for _, reference := range parser.References() {
		b.Set(reference, ParseBloodGlucoseTargetStartArray(parser.WithReferenceArrayParser(reference)))
	}
}

func (b *BloodGlucoseTargetStartArrayMap) Validate(validator structure.Validator, units *string) {
	for _, name := range b.sortedNames() {
		datumArrayValidator := validator.WithReference(name)
		if datumArray := b.Get(name); datumArray != nil {
			datumArray.Validate(datumArrayValidator, units)
		} else {
			datumArrayValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
}

func (b *BloodGlucoseTargetStartArrayMap) Normalize(normalizer data.Normalizer, units *string) {
	for _, name := range b.sortedNames() {
		if datumArray := b.Get(name); datumArray != nil {
			datumArray.Normalize(normalizer.WithReference(name), units)
		}
	}
}

func (b *BloodGlucoseTargetStartArrayMap) Get(name string) *BloodGlucoseTargetStartArray {
	if datumArray, exists := (*b)[name]; exists {
		return datumArray
	}
	return nil
}

func (b *BloodGlucoseTargetStartArrayMap) Set(name string, datumArray *BloodGlucoseTargetStartArray) {
	(*b)[name] = datumArray
}

func (b *BloodGlucoseTargetStartArrayMap) sortedNames() []string {
	names := []string{}
	for name := range *b {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
