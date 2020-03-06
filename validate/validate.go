package validate

import (
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/structure"
)

var (
	v                   *validator.Validate
	universalTranslator *ut.UniversalTranslator
)

func Initialize() error {
	v = validator.New()
	englishLocale := en.New()
	universalTranslator = ut.New(englishLocale)
	englishTranslator := universalTranslator.GetFallback()
	return enTranslations.RegisterDefaultTranslations(v, englishTranslator)
}

func StructWithLegacyErrorReporting(s interface{}, legacyValidator structure.Validator) {
	if err := v.Struct(s); err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			panic(err)
		}
		ReportErrorsToStructValidator(err.(validator.ValidationErrors), legacyValidator)
	}
}

func ReportErrorsToStructValidator(errors validator.ValidationErrors, legacyValidator structure.Validator) {
	for _, err := range errors {
		ReportErrorToStructValidator(err, legacyValidator)
	}
}

func ReportErrorToStructValidator(e validator.FieldError, legacyValidator structure.Validator) {
	codeAndTitle := GetErrorCodeAndTitle(e)
	message := e.Translate(universalTranslator.GetFallback())
	err := errors.Prepared(codeAndTitle.ErrorCode, codeAndTitle.ErrorTitle, message)
	legacyValidator.WithReference(e.Field()).ReportError(err)
}
