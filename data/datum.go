package data

import (
	"regexp"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

type Datum interface {
	Init()

	Meta() interface{}

	Parse(parser ObjectParser) error
	Validate(validator structure.Validator)
	Normalize(normalizer Normalizer)

	IdentityFields() ([]string, error)

	GetPayload() *Blob

	SetUserID(userID *string)
	SetDatasetID(datasetID *string)
	SetActive(active bool)
	SetDeviceID(deviceID *string)
	SetCreatedTime(createdTime *string)
	SetCreatedUserID(createdUserID *string)
	SetModifiedTime(modifiedTime *string)
	SetModifiedUserID(modifiedUserID *string)
	SetDeletedTime(deletedTime *string)
	SetDeletedUserID(deletedUserID *string)

	DeduplicatorDescriptor() *DeduplicatorDescriptor
	SetDeduplicatorDescriptor(deduplicatorDescriptor *DeduplicatorDescriptor)
}

func DatumAsPointer(datum Datum) *Datum {
	return &datum
}

var dataSetIDExpression = regexp.MustCompile("(upid_[0-9a-f]{12}|upid_[0-9a-f]{32}|[0-9a-f]{32})") // TODO: Want just "[0-9a-f]{32}"

func ValidateDataSetID(value string, errorReporter structure.ErrorReporter) {
	if value == "" {
		errorReporter.ReportError(structureValidator.ErrorValueEmpty())
	} else if !dataSetIDExpression.MatchString(value) {
		errorReporter.ReportError(ErrorValueStringAsDataSetIDNotValid(value))
	}
}

var userIDExpression = regexp.MustCompile("[0-9a-f]{10}")

func ValidateUserID(value string, errorReporter structure.ErrorReporter) {
	if value == "" {
		errorReporter.ReportError(structureValidator.ErrorValueEmpty())
	} else if !userIDExpression.MatchString(value) {
		errorReporter.ReportError(ErrorValueStringAsUserIDNotValid(value))
	}
}

func ErrorValueStringAsDataSetIDNotValid(value string) error {
	return errors.Preparedf(structureValidator.ErrorCodeValueNotValid, "value is not valid", "value %q is not valid as data set id", value)
}

func ErrorValueStringAsUserIDNotValid(value string) error {
	return errors.Preparedf(structureValidator.ErrorCodeValueNotValid, "value is not valid", "value %q is not valid as user id", value)
}
