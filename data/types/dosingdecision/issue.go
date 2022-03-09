package dosingdecision

import (
	"strconv"

	"github.com/tidepool-org/platform/metadata"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	IssueArrayLengthMaximum = 100
	IssueIDLengthMaximum    = 100
)

type Issue struct {
	ID       *string            `json:"id,omitempty" bson:"id,omitempty"`
	Metadata *metadata.Metadata `json:"metadata,omitempty" bson:"metadata,omitempty"`
}

func ParseIssue(parser structure.ObjectParser) *Issue {
	if !parser.Exists() {
		return nil
	}
	datum := NewIssue()
	parser.Parse(datum)
	return datum
}

func NewIssue() *Issue {
	return &Issue{}
}

func (i *Issue) Parse(parser structure.ObjectParser) {
	i.ID = parser.String("id")
	i.Metadata = metadata.ParseMetadata(parser.WithReferenceObjectParser("metadata"))
}

func (i *Issue) Validate(validator structure.Validator) {
	validator.String("id", i.ID).Exists().NotEmpty().LengthLessThanOrEqualTo(IssueIDLengthMaximum)
	if i.Metadata != nil {
		i.Metadata.Validate(validator.WithReference("metadata"))
	}
}

type IssueArray []*Issue

func ParseIssueArray(parser structure.ArrayParser) *IssueArray {
	if !parser.Exists() {
		return nil
	}
	datum := NewIssueArray()
	parser.Parse(datum)
	return datum
}

func NewIssueArray() *IssueArray {
	return &IssueArray{}
}

func (i *IssueArray) Parse(parser structure.ArrayParser) {
	for _, reference := range parser.References() {
		*i = append(*i, ParseIssue(parser.WithReferenceObjectParser(reference)))
	}
}

func (i *IssueArray) Validate(validator structure.Validator) {
	if length := len(*i); length == 0 {
		validator.ReportError(structureValidator.ErrorValueEmpty())
	} else if length > IssueArrayLengthMaximum {
		validator.ReportError(structureValidator.ErrorLengthNotLessThanOrEqualTo(length, IssueArrayLengthMaximum))
	}
	for index, datum := range *i {
		if datumValidator := validator.WithReference(strconv.Itoa(index)); datum != nil {
			datum.Validate(datumValidator)
		} else {
			datumValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
}
