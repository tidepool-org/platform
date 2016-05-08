package common

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

// NOTE: How InnerStruct gets parsed, validated, and normalized is up to the enclosing Datum.
// Since InnerStruct is not a Datum itself, but just a structure WITHIN a Datum, it does not
// need to implement Parse, Validate, and Normalize. However, I think it makes sense to mirror
// this mechanism as it makes it easy to understand.

// NOTE: InnerStruct is, for this example, in the common area and so could be used within
// multiple Datum types. However, if InnerStruct was specific to one and only one Datum type, then
// this file should be in the package for the Datum type (for example, datum/base/sample/sub) and
// NOT in the common area.

import (
	"github.com/tidepool-org/platform/pvn/data"
	"github.com/tidepool-org/platform/pvn/data/types/base/sample"
)

type InnerStruct struct {
	One  *string   `json:"one,omitempty"`
	Twos *[]string `json:"twos,omitempty"`
}

func NewInnerStruct() *InnerStruct {
	return &InnerStruct{}
}

func (i *InnerStruct) Parse(parser data.ObjectParser) {
	i.One = parser.ParseString("one")
	i.Twos = parser.ParseStringArray("twos")
}

func (i *InnerStruct) Validate(validator data.Validator) {
	validator.ValidateString("one", i.One).Exists()
	validator.ValidateStringArray("twos", i.Twos).Exists()
}

func (i *InnerStruct) Normalize(normalizer data.Normalizer) {

	// NOTE: As an example, this just creates a new Sample to normalize
	datum := sample.New()
	datum.String = i.One
	datum.StringArray = i.Twos

	normalizer.AddData(datum)
}

func ParseInnerStruct(parser data.ObjectParser) *InnerStruct {
	var innerStruct *InnerStruct
	if parser.Object() != nil {
		innerStruct = NewInnerStruct()
		innerStruct.Parse(parser)
	}
	return innerStruct
}

func ParseInnerStructArray(parser data.ArrayParser) *[]*InnerStruct {
	var innerStructArray *[]*InnerStruct
	if parser.Array() != nil {
		innerStructArray = &[]*InnerStruct{}
		for index := range *parser.Array() {
			*innerStructArray = append(*innerStructArray, ParseInnerStruct(parser.NewChildObjectParser(index)))
		}
	}
	return innerStructArray
}
