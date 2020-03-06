package validate

import (
	"github.com/go-playground/validator/v10"

	legacyValidator "github.com/tidepool-org/platform/structure/validator"
)

type ErrorCodeAndTitle struct {
	ErrorCode  string
	ErrorTitle string
}

var (
	errorCodesAndTitles = map[string]ErrorCodeAndTitle{
		"required": {
			ErrorCode:  legacyValidator.ErrorCodeValueNotExists,
			ErrorTitle: "value does not exist",
		},
		"isdefault": {
			ErrorCode:  legacyValidator.ErrorCodeValueExists,
			ErrorTitle: "value exists",
		},
		"len": {
			ErrorCode:  legacyValidator.ErrorCodeLengthOutOfRange,
			ErrorTitle: "length is out of range",
		},
		"min": {
			ErrorCode:  legacyValidator.ErrorCodeValueOutOfRange,
			ErrorTitle: "value is out of range",
		},
		"max": {
			ErrorCode:  legacyValidator.ErrorCodeValueOutOfRange,
			ErrorTitle: "value is out of range",
		},
		"eq": {
			ErrorCode:  legacyValidator.ErrorCodeValueOutOfRange,
			ErrorTitle: "value is out of range",
		},
		"ne": {
			ErrorCode:  legacyValidator.ErrorCodeValueOutOfRange,
			ErrorTitle: "value is out of range",
		},
		"lt": {
			ErrorCode:  legacyValidator.ErrorCodeValueOutOfRange,
			ErrorTitle: "value is out of range",
		},
		"lte": {
			ErrorCode:  legacyValidator.ErrorCodeValueOutOfRange,
			ErrorTitle: "value is out of range",
		},
		"gt": {
			ErrorCode:  legacyValidator.ErrorCodeValueOutOfRange,
			ErrorTitle: "value is out of range",
		},
		"gte": {
			ErrorCode:  legacyValidator.ErrorCodeValueOutOfRange,
			ErrorTitle: "value is out of range",
		},
	}

	defaultErrorCodeAndTitle = ErrorCodeAndTitle{
		ErrorCode:  legacyValidator.ErrorCodeValueNotValid,
		ErrorTitle: "value is not valid",
	}
)

func GetErrorCodeAndTitle(fieldError validator.FieldError) ErrorCodeAndTitle {
	if errorCode, ok := errorCodesAndTitles[fieldError.Tag()]; ok {
		return errorCode
	}

	return defaultErrorCodeAndTitle
}
