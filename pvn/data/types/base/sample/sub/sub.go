package sub

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

import (
	"github.com/tidepool-org/platform/pvn/data"
	"github.com/tidepool-org/platform/pvn/data/types/base/sample"
	"github.com/tidepool-org/platform/pvn/data/types/common"
)

type Sub struct {
	sample.Sample
	InnerStruct      *common.InnerStruct    `json:"innerStruct,omitempty"`
	InnerStructArray *[]*common.InnerStruct `json:"innerStructArray,omitempty"`
}

func Type() string {
	return sample.Type()
}

func SubType() string {
	return "sub"
}

func New() *Sub {
	subType := Type()
	subSubType := SubType()

	sub := &Sub{}
	sub.Type = &subType
	sub.SubType = &subSubType
	return sub
}

func (s *Sub) Parse(parser data.ObjectParser) {
	s.Sample.Parse(parser)

	// NOTE: Anytime we have a CONTAINED object, we create a new CHILD normalizer with a
	// reference to the CONTAINED object

	s.InnerStruct = common.ParseInnerStruct(parser.NewChildObjectParser("innerStruct"))
	s.InnerStructArray = common.ParseInnerStructArray(parser.NewChildArrayParser("innerStructArray"))
}

func (s *Sub) Validate(validator data.Validator) {
	s.Sample.Validate(validator)
	validator.ValidateString("subType", s.SubType).Exists().EqualTo(SubType())

	// NOTE: Anytime we have a CONTAINED object, we create a new CHILD validator with a
	// reference to the CONTAINED object

	if s.InnerStruct != nil {
		s.InnerStruct.Validate(validator.NewChildValidator("innerStruct"))
	}

	if s.InnerStructArray != nil {
		innerStructArrayValidator := validator.NewChildValidator("innerStructArray")
		for index, innerStruct := range *s.InnerStructArray {
			if innerStruct != nil {
				innerStruct.Validate(innerStructArrayValidator.NewChildValidator(index))
			}
		}
	}
}

func (s *Sub) Normalize(normalizer data.Normalizer) {
	s.Sample.Normalize(normalizer)

	// NOTE: Anytime we have a CONTAINED object, we create a new CHILD normalizer with a
	// reference to the CONTAINED object

	// NOTE: If the InnerStruct were an embedded data type, we'd add the data to the
	// normalizer and remove it from here

	if s.InnerStruct != nil {
		s.InnerStruct.Normalize(normalizer.NewChildNormalizer("innerStruct"))
	}

	if s.InnerStructArray != nil {
		innerStructArrayNormalizer := normalizer.NewChildNormalizer("innerStructArray")
		for index, innerStruct := range *s.InnerStructArray {
			if innerStruct != nil {
				innerStruct.Normalize(innerStructArrayNormalizer.NewChildNormalizer(index))
			}
		}
	}
}
