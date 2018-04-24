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

func (s *StringArray) EachNotEmpty() structure.StringArray {
	if s.value != nil {
		validator := NewValidator(s.base)
		for index, value := range *s.value {
			validator.String(strconv.Itoa(index), &value).NotEmpty()
		}
	}
	return s
}

func (s *StringArray) EachOneOf(allowedValues ...string) structure.StringArray {
	if s.value != nil {
		validator := NewValidator(s.base)
		for index, value := range *s.value {
			validator.String(strconv.Itoa(index), &value).OneOf(allowedValues...)
		}
	}
	return s
}

func (s *StringArray) EachNotOneOf(disallowedValues ...string) structure.StringArray {
	if s.value != nil {
		validator := NewValidator(s.base)
		for index, value := range *s.value {
			validator.String(strconv.Itoa(index), &value).NotOneOf(disallowedValues...)
		}
	}
	return s
}

func (s *StringArray) EachMatches(expression *regexp.Regexp) structure.StringArray {
	if s.value != nil {
		validator := NewValidator(s.base)
		for index, value := range *s.value {
			validator.String(strconv.Itoa(index), &value).Matches(expression)
		}
	}
	return s
}

func (s *StringArray) EachNotMatches(expression *regexp.Regexp) structure.StringArray {
	if s.value != nil {
		validator := NewValidator(s.base)
		for index, value := range *s.value {
			validator.String(strconv.Itoa(index), &value).NotMatches(expression)
		}
	}
	return s
}

func (s *StringArray) Using(using func(value []string, errorReporter structure.ErrorReporter)) structure.StringArray {
	if s.value != nil {
		if using != nil {
			using(*s.value, s.base)
		}
	}
	return s
}
