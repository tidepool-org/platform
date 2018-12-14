package image

import (
	"github.com/tidepool-org/platform/errors"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	ErrorCodeImageContentIntentUnexpected = "image-content-intent-unexpected"
	ErrorCodeImageMalformed               = "image-malformed"
)

func ErrorValueStringAsIDNotValid(value string) error {
	return errors.Preparedf(structureValidator.ErrorCodeValueNotValid, "value is not valid", "value %q is not valid as image id", value)
}

func ErrorValueStringAsContentIDNotValid(value string) error {
	return errors.Preparedf(structureValidator.ErrorCodeValueNotValid, "value is not valid", "value %q is not valid as image content id", value)
}

func ErrorValueStringAsRenditionsIDNotValid(value string) error {
	return errors.Preparedf(structureValidator.ErrorCodeValueNotValid, "value is not valid", "value %q is not valid as image renditions id", value)
}

func ErrorValueStringAsContentIntentNotValid(value string) error {
	return errors.Preparedf(structureValidator.ErrorCodeValueNotValid, "value is not valid", "value %q is not valid as image content intent", value)
}

func ErrorImageContentIntentUnexpected(value string) error {
	return errors.Preparedf(ErrorCodeImageContentIntentUnexpected, "image content intent unexpected", "image content intent %q unexpected", value)
}

func ErrorValueRenditionNotParsable(value string) error {
	return errors.Preparedf(structureParser.ErrorCodeValueNotParsable, "value is not a parsable rendition", "value %q is not a parsable rendition", value)
}

func ErrorValueStringAsColorNotValid(value string) error {
	return errors.Preparedf(structureValidator.ErrorCodeValueNotValid, "value is not valid", "value %q is not valid as color", value)
}

func ErrorImageMalformed(reason string) error {
	return errors.Prepared(ErrorCodeImageMalformed, "image is malformed", reason)
}
