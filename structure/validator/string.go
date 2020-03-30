package validator

import (
	"regexp"
	"time"

	"github.com/google/uuid"

	"github.com/tidepool-org/platform/structure"
	structureBase "github.com/tidepool-org/platform/structure/base"
)

const (
	// Email regexp from https://github.com/go-playground/validator/
	emailRegexString        = "^(?:(?:(?:(?:[a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+(?:\\.([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+)*)|(?:(?:\\x22)(?:(?:(?:(?:\\x20|\\x09)*(?:\\x0d\\x0a))?(?:\\x20|\\x09)+)?(?:(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x7f]|\\x21|[\\x23-\\x5b]|[\\x5d-\\x7e]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[\\x01-\\x09\\x0b\\x0c\\x0d-\\x7f]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}]))))*(?:(?:(?:\\x20|\\x09)*(?:\\x0d\\x0a))?(\\x20|\\x09)+)?(?:\\x22))))@(?:(?:(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])(?:[a-zA-Z]|\\d|-|\\.|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.)+(?:(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])(?:[a-zA-Z]|\\d|-|\\.|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.?$"
	hexadecimalRegexString  = "^[A-Fa-f0-9]+$"
	alphanumericRegexString = "^[A-Za-z0-9]+$"
)

var (
	emailRegex        = regexp.MustCompile(emailRegexString)
	hexadecimalRegex  = regexp.MustCompile(hexadecimalRegexString)
	alphanumericRegex = regexp.MustCompile(alphanumericRegexString)
)

type String struct {
	base  *structureBase.Base
	value *string
}

func NewString(base *structureBase.Base, value *string) *String {
	return &String{
		base:  base,
		value: value,
	}
}

func (s *String) Exists() structure.String {
	if s.value == nil {
		s.base.ReportError(ErrorValueNotExists())
	}
	return s
}

func (s *String) NotExists() structure.String {
	if s.value != nil {
		s.base.ReportError(ErrorValueExists())
	}
	return s
}

func (s *String) Empty() structure.String {
	if s.value != nil {
		if len([]rune(*s.value)) > 0 {
			s.base.ReportError(ErrorValueNotEmpty())
		}
	}
	return s
}

func (s *String) NotEmpty() structure.String {
	if s.value != nil {
		if len([]rune(*s.value)) == 0 {
			s.base.ReportError(ErrorValueEmpty())
		}
	}
	return s
}

func (s *String) EqualTo(value string) structure.String {
	if s.value != nil {
		if *s.value != value {
			s.base.ReportError(ErrorValueNotEqualTo(*s.value, value))
		}
	}
	return s
}

func (s *String) NotEqualTo(value string) structure.String {
	if s.value != nil {
		if *s.value == value {
			s.base.ReportError(ErrorValueEqualTo(*s.value, value))
		}
	}
	return s
}

func (s *String) LengthEqualTo(limit int) structure.String {
	if s.value != nil {
		if length := len([]rune(*s.value)); length != limit {
			s.base.ReportError(ErrorLengthNotEqualTo(length, limit))
		}
	}
	return s
}

func (s *String) LengthNotEqualTo(limit int) structure.String {
	if s.value != nil {
		if length := len([]rune(*s.value)); length == limit {
			s.base.ReportError(ErrorLengthEqualTo(length, limit))
		}
	}
	return s
}

func (s *String) LengthLessThan(limit int) structure.String {
	if s.value != nil {
		if length := len([]rune(*s.value)); length >= limit {
			s.base.ReportError(ErrorLengthNotLessThan(length, limit))
		}
	}
	return s
}

func (s *String) LengthLessThanOrEqualTo(limit int) structure.String {
	if s.value != nil {
		if length := len([]rune(*s.value)); length > limit {
			s.base.ReportError(ErrorLengthNotLessThanOrEqualTo(length, limit))
		}
	}
	return s
}

func (s *String) LengthGreaterThan(limit int) structure.String {
	if s.value != nil {
		if length := len([]rune(*s.value)); length <= limit {
			s.base.ReportError(ErrorLengthNotGreaterThan(length, limit))
		}
	}
	return s
}

func (s *String) LengthGreaterThanOrEqualTo(limit int) structure.String {
	if s.value != nil {
		if length := len([]rune(*s.value)); length < limit {
			s.base.ReportError(ErrorLengthNotGreaterThanOrEqualTo(length, limit))
		}
	}
	return s
}

func (s *String) LengthInRange(lowerLimit int, upperLimit int) structure.String {
	if s.value != nil {
		if length := len([]rune(*s.value)); length < lowerLimit || length > upperLimit {
			s.base.ReportError(ErrorLengthNotInRange(length, lowerLimit, upperLimit))
		}
	}
	return s
}

func (s *String) OneOf(allowedValues ...string) structure.String {
	if s.value != nil {
		for _, allowedValue := range allowedValues {
			if allowedValue == *s.value {
				return s
			}
		}
		s.base.ReportError(ErrorValueStringNotOneOf(*s.value, allowedValues))
	}
	return s
}

func (s *String) NotOneOf(disallowedValues ...string) structure.String {
	if s.value != nil {
		for _, disallowedValue := range disallowedValues {
			if disallowedValue == *s.value {
				s.base.ReportError(ErrorValueStringOneOf(*s.value, disallowedValues))
				return s
			}
		}
	}
	return s
}

func (s *String) Matches(expression *regexp.Regexp) structure.String {
	if s.value != nil {
		if expression == nil || !expression.MatchString(*s.value) {
			s.base.ReportError(ErrorValueStringNotMatches(*s.value, expression))
		}
	}
	return s
}

func (s *String) NotMatches(expression *regexp.Regexp) structure.String {
	if s.value != nil {
		if expression == nil || expression.MatchString(*s.value) {
			s.base.ReportError(ErrorValueStringMatches(*s.value, expression))
		}
	}
	return s
}

func (s *String) Using(usingFunc structure.StringUsingFunc) structure.String {
	if s.value != nil {
		if usingFunc != nil {
			usingFunc(*s.value, s.base)
		}
	}
	return s
}

func (s *String) AsTime(layout string) structure.Time {
	var valueAsTime *time.Time

	if s.value != nil {
		if parsed, err := time.Parse(layout, *s.value); err != nil {
			s.base.ReportError(ErrorValueStringAsTimeNotValid(*s.value, layout))
		} else {
			valueAsTime = &parsed
		}
	}

	return NewTime(s.base, valueAsTime)
}

func (s *String) Email() structure.String {
	return s.Matches(emailRegex)
}

func (s *String) Alphanumeric() structure.String {
	return s.Matches(alphanumericRegex)
}

func (s *String) Hexadecimal() structure.String {
	return s.Matches(hexadecimalRegex)
}

func (s *String) UUID() structure.String {
	if s.value != nil {
		v := *s.value
		if _, err := uuid.Parse(v); err != nil {
			s.base.ReportError(ErrorValueStringNotValidUUID(v))
		}
	}

	return s
}
