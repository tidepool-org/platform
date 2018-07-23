package validator

import (
	"regexp"
	"strconv"

	"github.com/tidepool-org/platform/structure"
	structureBase "github.com/tidepool-org/platform/structure/base"
)

type StringArray struct {
	base  *structureBase.Base
	value *[]string
}

func NewStringArray(base *structureBase.Base, value *[]string) *StringArray {
	return &StringArray{
		base:  base,
		value: value,
	}
}

func (s *StringArray) Exists() structure.StringArray {
	if s.value == nil {
		s.base.ReportError(ErrorValueNotExists())
	}
	return s
}

func (s *StringArray) NotExists() structure.StringArray {
	if s.value != nil {
		s.base.ReportError(ErrorValueExists())
	}
	return s
}

func (s *StringArray) Empty() structure.StringArray {
	if s.value != nil {
		if len(*s.value) != 0 {
			s.base.ReportError(ErrorValueNotEmpty())
		}
	}
	return s
}

func (s *StringArray) NotEmpty() structure.StringArray {
	if s.value != nil {
		if len(*s.value) == 0 {
			s.base.ReportError(ErrorValueEmpty())
		}
	}
	return s
}

func (s *StringArray) LengthEqualTo(limit int) structure.StringArray {
	if s.value != nil {
		if length := len(*s.value); length != limit {
			s.base.ReportError(ErrorLengthNotEqualTo(length, limit))
		}
	}
	return s
}

func (s *StringArray) LengthNotEqualTo(limit int) structure.StringArray {
	if s.value != nil {
		if length := len(*s.value); length == limit {
			s.base.ReportError(ErrorLengthEqualTo(length, limit))
		}
	}
	return s
}

func (s *StringArray) LengthLessThan(limit int) structure.StringArray {
	if s.value != nil {
		if length := len(*s.value); length >= limit {
			s.base.ReportError(ErrorLengthNotLessThan(length, limit))
		}
	}
	return s
}

func (s *StringArray) LengthLessThanOrEqualTo(limit int) structure.StringArray {
	if s.value != nil {
		if length := len(*s.value); length > limit {
			s.base.ReportError(ErrorLengthNotLessThanOrEqualTo(length, limit))
		}
	}
	return s
}

func (s *StringArray) LengthGreaterThan(limit int) structure.StringArray {
	if s.value != nil {
		if length := len(*s.value); length <= limit {
			s.base.ReportError(ErrorLengthNotGreaterThan(length, limit))
		}
	}
	return s
}

func (s *StringArray) LengthGreaterThanOrEqualTo(limit int) structure.StringArray {
	if s.value != nil {
		if length := len(*s.value); length < limit {
			s.base.ReportError(ErrorLengthNotGreaterThanOrEqualTo(length, limit))
		}
	}
	return s
}

func (s *StringArray) LengthInRange(lowerLimit int, upperLimit int) structure.StringArray {
	if s.value != nil {
		if length := len(*s.value); length < lowerLimit || length > upperLimit {
			s.base.ReportError(ErrorLengthNotInRange(length, lowerLimit, upperLimit))
		}
	}
	return s
}

func (s *StringArray) Each(eachFunc structure.StringArrayEachFunc) structure.StringArray {
	if s.value != nil {
		if eachFunc != nil {
			validator := NewValidator(s.base)
			for index, value := range *s.value {
				eachFunc(validator.String(strconv.Itoa(index), &value))
			}
		}
	}
	return s
}

func (s *StringArray) EachNotEmpty() structure.StringArray {
	return s.Each(func(stringValidator structure.String) { stringValidator.NotEmpty() })
}

func (s *StringArray) EachOneOf(allowedValues ...string) structure.StringArray {
	return s.Each(func(stringValidator structure.String) { stringValidator.OneOf(allowedValues...) })
}

func (s *StringArray) EachNotOneOf(disallowedValues ...string) structure.StringArray {
	return s.Each(func(stringValidator structure.String) { stringValidator.NotOneOf(disallowedValues...) })
}

func (s *StringArray) EachMatches(expression *regexp.Regexp) structure.StringArray {
	return s.Each(func(stringValidator structure.String) { stringValidator.Matches(expression) })
}

func (s *StringArray) EachNotMatches(expression *regexp.Regexp) structure.StringArray {
	return s.Each(func(stringValidator structure.String) { stringValidator.NotMatches(expression) })
}

func (s *StringArray) EachUsing(eachUsingFunc structure.StringArrayEachUsingFunc) structure.StringArray {
	if s.value != nil {
		if eachUsingFunc != nil {
			validator := NewValidator(s.base)
			for index, value := range *s.value {
				eachUsingFunc(value, validator.WithReference(strconv.Itoa(index)))
			}
		}
	}
	return s
}

func (s *StringArray) EachUnique() structure.StringArray {
	if s.value != nil {
		values := map[string]bool{}
		for index, value := range *s.value {
			if _, found := values[value]; found {
				s.base.WithReference(strconv.Itoa(index)).ReportError(ErrorValueDuplicate())
			} else {
				values[value] = true
			}
		}
	}
	return s
}

func (s *StringArray) Using(usingFunc structure.StringArrayUsingFunc) structure.StringArray {
	if s.value != nil {
		if usingFunc != nil {
			usingFunc(*s.value, s.base)
		}
	}
	return s
}
