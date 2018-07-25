package association

import (
	"strconv"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/validate"
)

const (
	AssociationArrayLengthMaximum = 100
	ReasonLengthMaximum           = 1000
	TypeDatum                     = "datum"
	TypeURL                       = "url"
)

func Types() []string {
	return []string{
		TypeDatum,
		TypeURL,
	}
}

type Association struct {
	ID     *string `json:"id,omitempty" bson:"id,omitempty"`
	Reason *string `json:"reason,omitempty" bson:"reason,omitempty"`
	Type   *string `json:"type,omitempty" bson:"type,omitempty"`
	URL    *string `json:"url,omitempty" bson:"url,omitempty"`
}

func ParseAssociation(parser data.ObjectParser) *Association {
	if parser.Object() == nil {
		return nil
	}
	datum := NewAssociation()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewAssociation() *Association {
	return &Association{}
}

func (a *Association) Parse(parser data.ObjectParser) {
	a.ID = parser.ParseString("id")
	a.Reason = parser.ParseString("reason")
	a.Type = parser.ParseString("type")
	a.URL = parser.ParseString("url")
}

func (a *Association) Validate(validator structure.Validator) {
	if a.Type != nil {
		switch *a.Type {
		case TypeDatum:
			validator.String("id", a.ID).Exists().Using(id.Validate)
		case TypeURL:
			validator.String("id", a.ID).NotExists()
		}
	}
	validator.String("reason", a.Reason).NotEmpty().LengthLessThanOrEqualTo(ReasonLengthMaximum)
	validator.String("type", a.Type).Exists().OneOf(Types()...)
	if a.Type != nil {
		switch *a.Type {
		case TypeDatum:
			validator.String("url", a.URL).NotExists()
		case TypeURL:
			validator.String("url", a.URL).Exists().Using(validate.URL)
		}
	}
}

func (a *Association) Normalize(normalizer data.Normalizer) {}

type AssociationArray []*Association

func ParseAssociationArray(parser data.ArrayParser) *AssociationArray {
	if parser.Array() == nil {
		return nil
	}
	datum := NewAssociationArray()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewAssociationArray() *AssociationArray {
	return &AssociationArray{}
}

func (a *AssociationArray) Parse(parser data.ArrayParser) {
	for index := range *parser.Array() {
		*a = append(*a, ParseAssociation(parser.NewChildObjectParser(index)))
	}
}

func (a *AssociationArray) Validate(validator structure.Validator) {
	if length := len(*a); length == 0 {
		validator.ReportError(structureValidator.ErrorValueEmpty())
	} else if length > AssociationArrayLengthMaximum {
		validator.ReportError(structureValidator.ErrorLengthNotLessThanOrEqualTo(length, AssociationArrayLengthMaximum))
	}

	for index, datum := range *a {
		datumValidator := validator.WithReference(strconv.Itoa(index))
		if datum != nil {
			datum.Validate(datumValidator)
		} else {
			datumValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
	}
}

func (a *AssociationArray) Normalize(normalizer data.Normalizer) {
	for index, datum := range *a {
		if datum != nil {
			datum.Normalize(normalizer.WithReference(strconv.Itoa(index)))
		}
	}
}
