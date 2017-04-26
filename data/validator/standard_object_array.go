package validator

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/service"
)

type StandardObjectArray struct {
	context   data.Context
	reference interface{}
	value     *[]map[string]interface{}
}

func NewStandardObjectArray(context data.Context, reference interface{}, value *[]map[string]interface{}) *StandardObjectArray {
	if context == nil {
		return nil
	}

	return &StandardObjectArray{
		context:   context,
		reference: reference,
		value:     value,
	}
}

func (s *StandardObjectArray) Exists() data.ObjectArray {
	if s.value == nil {
		s.context.AppendError(s.reference, service.ErrorValueNotExists())
	}
	return s
}

func (s *StandardObjectArray) NotExists() data.ObjectArray {
	if s.value != nil {
		s.context.AppendError(s.reference, service.ErrorValueExists())
	}
	return s
}

func (s *StandardObjectArray) Empty() data.ObjectArray {
	if s.value != nil {
		if len(*s.value) != 0 {
			s.context.AppendError(s.reference, service.ErrorValueNotEmpty())
		}
	}
	return s
}

func (s *StandardObjectArray) NotEmpty() data.ObjectArray {
	if s.value != nil {
		if len(*s.value) == 0 {
			s.context.AppendError(s.reference, service.ErrorValueEmpty())
		}
	}
	return s
}

func (s *StandardObjectArray) LengthEqualTo(limit int) data.ObjectArray {
	if s.value != nil {
		if length := len(*s.value); length != limit {
			s.context.AppendError(s.reference, service.ErrorLengthNotEqualTo(length, limit))
		}
	}
	return s
}

func (s *StandardObjectArray) LengthNotEqualTo(limit int) data.ObjectArray {
	if s.value != nil {
		if length := len(*s.value); length == limit {
			s.context.AppendError(s.reference, service.ErrorLengthEqualTo(length, limit))
		}
	}
	return s
}

func (s *StandardObjectArray) LengthLessThan(limit int) data.ObjectArray {
	if s.value != nil {
		if length := len(*s.value); length >= limit {
			s.context.AppendError(s.reference, service.ErrorLengthNotLessThan(length, limit))
		}
	}
	return s
}

func (s *StandardObjectArray) LengthLessThanOrEqualTo(limit int) data.ObjectArray {
	if s.value != nil {
		if length := len(*s.value); length > limit {
			s.context.AppendError(s.reference, service.ErrorLengthNotLessThanOrEqualTo(length, limit))
		}
	}
	return s
}

func (s *StandardObjectArray) LengthGreaterThan(limit int) data.ObjectArray {
	if s.value != nil {
		if length := len(*s.value); length <= limit {
			s.context.AppendError(s.reference, service.ErrorLengthNotGreaterThan(length, limit))
		}
	}
	return s
}

func (s *StandardObjectArray) LengthGreaterThanOrEqualTo(limit int) data.ObjectArray {
	if s.value != nil {
		if length := len(*s.value); length < limit {
			s.context.AppendError(s.reference, service.ErrorLengthNotGreaterThanOrEqualTo(length, limit))
		}
	}
	return s
}

func (s *StandardObjectArray) LengthInRange(lowerLimit int, upperLimit int) data.ObjectArray {
	if s.value != nil {
		if length := len(*s.value); length < lowerLimit || length > upperLimit {
			s.context.AppendError(s.reference, service.ErrorLengthNotInRange(length, lowerLimit, upperLimit))
		}
	}
	return s
}
