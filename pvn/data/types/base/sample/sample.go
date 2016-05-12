package sample

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

import "github.com/tidepool-org/platform/pvn/data"
import "github.com/tidepool-org/platform/pvn/data/types/base"

type Sample struct {
	base.Base
	SubType        *string                   `json:"subType,omitempty"`
	Boolean        *bool                     `json:"boolean,omitempty"`
	Integer        *int                      `json:"integer,omitempty"`
	Float          *float64                  `json:"float,omitempty"`
	String         *string                   `json:"string,omitempty"`
	StringArray    *[]string                 `json:"stringArray,omitempty"`
	Object         *map[string]interface{}   `json:"object,omitempty"`
	ObjectArray    *[]map[string]interface{} `json:"objectArray,omitempty"`
	Interface      *interface{}              `json:"interface,omitempty"`
	InterfaceArray *[]interface{}            `json:"interfaceArray,omitempty"`
	Time           *string                   `json:"time,omitempty"`
}

func Type() string {
	return "sample"
}

func New() *Sample {
	sampleType := Type()

	sample := &Sample{}
	sample.Type = &sampleType
	return sample
}

func (s *Sample) Parse(parser data.ObjectParser) {

	// NOTE: We would only parse the fields that we want from client input. Fields that
	// we add later (like uploadId in an upload record) we would explicitly NOT parse. This
	// way we have control of what input is actually used.

	s.Base.Parse(parser)
	s.Boolean = parser.ParseBoolean("boolean")
	s.Integer = parser.ParseInteger("integer")
	s.Float = parser.ParseFloat("float")
	s.String = parser.ParseString("string")
	s.StringArray = parser.ParseStringArray("stringArray")
	s.Object = parser.ParseObject("object")
	s.ObjectArray = parser.ParseObjectArray("objectArray")
	s.Interface = parser.ParseInterface("interface")
	s.InterfaceArray = parser.ParseInterfaceArray("interfaceArray")
	s.Time = parser.ParseString("time")
}

func (s *Sample) Validate(validator data.Validator) {

	// NOTE: We might eventually have multiple Validate methods, such as:
	//   ValidateAfterParsing - validating after parsing an upload
	//   ValidateBeforePersisting - validating before storing in the database
	//   ValidateBeforeResponse - validating before returning to client
	// We would also accomplish the same thing with an extra parameter on the Validate
	// function (eg. validation phase, or something like that)

	s.Base.Validate(validator)
	validator.ValidateString("type", s.Type).Exists().EqualTo(Type())
	validator.ValidateString("subType", s.SubType).Exists()
	validator.ValidateBoolean("boolean", s.Boolean).Exists().True()
	validator.ValidateInteger("integer", s.Integer).Exists().LessThan(10).GreaterThan(3).OneOf([]int{4, 7, 9})
	validator.ValidateFloat("float", s.Float).Exists().LessThan(6.7).GreaterThanOrEqualTo(0.01).OneOf([]float64{4.5, 6.7})
	validator.ValidateString("string", s.String).Exists().LengthInRange(2, 10).OneOf([]string{"aaa", "bbb", "asdfasdf"})
	validator.ValidateStringArray("stringArray", s.StringArray).Exists().LengthEqualTo(2).EachOneOf([]string{"bach", "blech"})
	validator.ValidateObject("object", s.Object).Exists()
	validator.ValidateObjectArray("objectArray", s.ObjectArray).Exists().LengthGreaterThanOrEqualTo(1)
	validator.ValidateInterface("interface", s.Interface).Exists()
	validator.ValidateInterfaceArray("interfaceArray", s.InterfaceArray).Exists().LengthNotEqualTo(5)
	validator.ValidateStringAsTime("time", s.Time, "2006-01-02T15:04:05Z07:00").Exists().BeforeNow()
}

func (s *Sample) Normalize(normalizer data.Normalizer) {
	s.Base.Normalize(normalizer)

	// NOTE: Typically, Normalize would be called ONLY if there are no validation errors, so
	// this is really just defensive coding

	if s.Boolean != nil && s.Integer != nil {

		// NOTE: Pretend this a unit conversion
		if *s.Boolean {
			*s.Boolean = false
			*s.Integer *= 2
		}
	}
}
