package consent

import (
	"context"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"

	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/structure"
)

const (
	TypeMinLength = 0
	TypeMaxLength = 64

	TypeBigDataDonationProject = "big_data_donation_project"

	ContentTypeMarkdown ContentType = "markdown"
)

//go:generate mockgen -source=consent.go -destination=test/service_mocks.go -package=test Service
type Service interface {
	ConsentAccessor
	RecordAccessor
}

type ConsentAccessor interface {
	ListConsents(context.Context, *Filter, *page.Pagination) (*storeStructuredMongo.ListResult[Consent], error)
	EnsureConsent(context.Context, *Consent) error
}

type Consents []Consent
type Consent struct {
	Type        string      `json:"type" bson:"type"`
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
	validator.String("type", &p.Type).LengthInRange(TypeMinLength, TypeMaxLength)
	validator.Int("version", &p.Version).GreaterThan(0)
	validator.String("content", &p.Content).NotEmpty()
	validator.String("contentType", structure.ValueAsString(&p.ContentType)).OneOf(structure.ValuesAsStringArray(ContentTypes())...)
}

type ContentType string

func ContentTypes() []ContentType {
	return []ContentType{
		ContentTypeMarkdown,
	}
}

type Filter struct {
	Type    *string `json:"type,omitempty"`
	Version *int    `json:"version,omitempty"`
	Latest  *bool   `json:"latest,omitempty"`
}

func NewConsentFilter() *Filter {
	return &Filter{}
}

func (p *Filter) Parse(parser structure.ObjectParser) {
	p.Type = parser.String("type")
	p.Version = parser.Int("version")
	p.Latest = parser.Bool("latest")
}

func (p *Filter) Validate(validator structure.Validator) {
	typeValidator := validator.String("type", p.Type).LengthInRange(TypeMinLength, TypeMaxLength)
	versionValidator := validator.Int("version", p.Version).GreaterThan(0)
	if p.Latest != nil {
		typeValidator.Exists()
		versionValidator.NotExists()
	}
}

func PrettifyType(typ string) string {
	return cases.Title(language.English, cases.Compact).String(strings.ReplaceAll(typ, "_", " "))
}
