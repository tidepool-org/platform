package consent

import (
	"context"
	"time"

	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
)

const (
	TypeBigDataDonationProject Type = "big_data_donation_project"

	ContentTypeMarkdown ContentType = "markdown"
)

type Service interface {
	ConsentAccessor
	RecordAccessor
}

type ConsentAccessor interface {
	ListConsents(context.Context, *Filter, *page.Pagination) (Consents, error)
	EnsureConsent(context.Context, *Consent) error
}

type Consents []Consent
type Consent struct {
	Type        Type        `json:"type" bson:"type"`
	Version     int         `json:"version" bson:"version"`
	Content     string      `json:"content" bson:"content"`
	ContentType ContentType `json:"contentType" bson:"contentType"`
	CreatedTime time.Time   `json:"createdTime" bson:"createdTime"`
}

func NewConsent() *Consent {
	return &Consent{
		CreatedTime: time.Now(),
	}
}

func (p *Consent) Validate(validator structure.Validator) {
	validator.String("type", structure.ValueAsString(&p.Type)).OneOf(structure.ValuesAsStringArray(Types())...)
	validator.Int("version", &p.Version).GreaterThan(0)
	validator.String("content", &p.Content).NotEmpty()
	validator.String("contentType", structure.ValueAsString(&p.ContentType)).OneOf(structure.ValuesAsStringArray(ContentTypes())...)
}

type Type string

func NewConsentType(value *string) *Type {
	if value == nil {
		return nil
	}
	return pointer.FromAny(Type(*value))
}

func Types() []Type {
	return []Type{
		TypeBigDataDonationProject,
	}
}

type ContentType string

func ContentTypes() []ContentType {
	return []ContentType{
		ContentTypeMarkdown,
	}
}

type Filter struct {
	Type    *Type `json:"type,omitempty"`
	Version *int  `json:"version,omitempty"`
	Latest  *bool `json:"latest,omitempty"`
}

func NewConsentFilter() *Filter {
	return &Filter{}
}

func (p *Filter) Parse(parser structure.ObjectParser) {
	p.Type = NewConsentType(parser.String("type"))
	p.Version = parser.Int("version")
	p.Latest = parser.Bool("latest")
}

func (p *Filter) Validate(validator structure.Validator) {
	typeValidator := validator.String("type", structure.ValueAsString(p.Type)).OneOf(structure.ValuesAsStringArray(Types())...)
	versionValidator := validator.Int("version", p.Version).GreaterThan(0)
	if p.Latest != nil {
		typeValidator.Exists()
		versionValidator.NotExists()
	}
}
