package data

import (
	"regexp"
	"strconv"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	DeleteOriginIDLengthMaximum = 100
)

type DeleteOrigin struct {
	ID *string `json:"id,omitempty"`
}

func ParseDeleteOrigin(parser structure.ObjectParser) *DeleteOrigin {
	if !parser.Exists() {
		return nil
	}
	datum := NewDeleteOrigin()
	parser.Parse(datum)
	return datum
}

func NewDeleteOrigin() *DeleteOrigin {
	return &DeleteOrigin{}
}

func (d *DeleteOrigin) Parse(parser structure.ObjectParser) {
	d.ID = parser.String("id")
}

func (d *DeleteOrigin) Validate(validator structure.Validator) {
	validator.String("id", d.ID).NotEmpty().LengthLessThanOrEqualTo(DeleteOriginIDLengthMaximum)
}

type Delete struct {
	ID     *string       `json:"id,omitempty"`
	Origin *DeleteOrigin `json:"origin,omitempty"`
}

func ParseDelete(parser structure.ObjectParser) *Delete {
	if !parser.Exists() {
		return nil
	}
	datum := NewDelete()
	parser.Parse(datum)
	return datum
}

func NewDelete() *Delete {
	return &Delete{}
}

func (d *Delete) Parse(parser structure.ObjectParser) {
	d.ID = parser.String("id")
	d.Origin = ParseDeleteOrigin(parser.WithReferenceObjectParser("origin"))
}

func (d *Delete) Validate(validator structure.Validator) {
	validator.String("id", d.ID).Using(IDValidator)
	if d.Origin != nil {
		d.Origin.Validate(validator.WithReference("origin"))
	}
}

type Deletes []*Delete

func ParseDeletes(parser structure.ArrayParser) *Deletes {
	if !parser.Exists() {
		return nil
	}
	datum := NewDeletes()
	parser.Parse(datum)
	return datum
}

func NewDeletes() *Deletes {
	return &Deletes{}
}

func (d *Deletes) Parse(parser structure.ArrayParser) {
	for _, reference := range parser.References() {
		*d = append(*d, ParseDelete(parser.WithReferenceObjectParser(reference)))
	}
}

func (d *Deletes) Validate(validator structure.Validator) {
	for index, delete := range *d {
		delete.Validate(validator.WithReference(strconv.Itoa(index)))
	}
}

func NewID() string {
	return id.Must(id.New(16))
}

func IsValidID(value string) bool {
	return ValidateID(value) == nil
}

func IDValidator(value string, errorReporter structure.ErrorReporter) {
	errorReporter.ReportError(ValidateID(value))
}

func ValidateID(value string) error {
	if value == "" {
		return structureValidator.ErrorValueEmpty()
	} else if !setIDExpression.MatchString(value) {
		return ErrorValueStringAsIDNotValid(value)
	}
	return nil
}

func ErrorValueStringAsIDNotValid(value string) error {
	return errors.Preparedf(structureValidator.ErrorCodeValueNotValid, "value is not valid", "value %q is not valid as data id", value)
}

var idExpression = regexp.MustCompile("^[0-9a-z]{32}$")
