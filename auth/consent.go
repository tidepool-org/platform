package auth

import (
	"context"
	"encoding/json"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"time"
)

const (
	ConsentTypeBigDataDonationProject ConsentType = "big_data_donation_project"

	ConsentContentTypeMarkdown ConsentContentType = "markdown"

	ConsentRecordStatusActive  ConsentRecordStatus = "active"
	ConsentRecordStatusRevoked ConsentRecordStatus = "revoked"

	ConsentGrantorTypeOwner          ConsentRecordGrantorType = "owner"
	ConsentGrantorTypeParentGuardian ConsentRecordGrantorType = "parent/guardian"

	ConsentRecordAgeGroupUnderTwelve       ConsentRecordAgeGroup = "<12"
	ConsentRecordAgeGroupThirteenSeventeen ConsentRecordAgeGroup = "13-17"
	ConsentRecordAgeGroupOverEighteen      ConsentRecordAgeGroup = ">18"

	BigDataDonationProjectOrganizationsADCES BigDataDonationProjectOrganization = "adces"
)

type ConsentAccessor interface {
	ListConsents(context.Context, *ConsentFilter, *page.Pagination) (Consents, error)
	EnsureConsent(context.Context, Consent) error
}

type Consents []Consent
type Consent struct {
	Type        ConsentType        `json:"type" bson:"type"`
	Version     int                `json:"version" bson:"version"`
	Content     string             `json:"content" bson:"content"`
	ContentType ConsentContentType `json:"contentType" bson:"contentType"`
	CreatedTime time.Time          `json:"createdTime" bson:"createdTime"`
}

func NewConsent() *Consent {
	return &Consent{
		CreatedTime: time.Now(),
	}
}

func (p *Consent) Validate(validator structure.Validator) {
	validator.String("type", structure.ValueAsString(&p.Type)).OneOf(structure.ValuesAsStringArray(ConsentTypes())...)
	validator.Int("version", &p.Version).GreaterThan(0)
	validator.String("content", &p.Content).NotEmpty()
	validator.String("contentType", structure.ValueAsString(&p.ContentType)).OneOf(structure.ValuesAsStringArray(ConsentContentTypes())...)
}

type ConsentType string

func NewConsentType(value *string) *ConsentType {
	if value == nil {
		return nil
	}
	return pointer.FromAny(ConsentType(*value))
}

func ConsentTypes() []ConsentType {
	return []ConsentType{
		ConsentTypeBigDataDonationProject,
	}
}

type ConsentContentType string

func ConsentContentTypes() []ConsentContentType {
	return []ConsentContentType{
		ConsentContentTypeMarkdown,
	}
}

type ConsentFilter struct {
	Type    *ConsentType `json:"type,omitempty"`
	Version *int         `json:"version,omitempty"`
}

func NewConsentFilter() *ConsentFilter {
	return &ConsentFilter{}
}

func (p *ConsentFilter) Parse(parser structure.ObjectParser) {
	p.Type = NewConsentType(parser.String("type"))
	p.Version = parser.Int("version")
}

func (p *ConsentFilter) Validate(validator structure.Validator) {
	validator.String("type", structure.ValueAsString(p.Type)).OneOf(structure.ValuesAsStringArray(ConsentTypes())...)
	validator.Int("version", p.Version).GreaterThan(0)
}

type ConsentRecordAccessor interface {
	CreateConsentRecord(context.Context, string, *ConsentRecordCreate) (*ConsentRecord, error)
	ListConsentRecords(context.Context, *ConsentRecordFilter, *page.Pagination) (ConsentRecords, error)
	RevokeConsentRecord(context.Context, string) error
	UpdateConsentRecord(context.Context, *ConsentRecord) (*ConsentRecord, error)
}

type ConsentRecords []ConsentRecord
type ConsentRecord struct {
	ID                 string                   `json:"id,omitempty" bson:"id"`
	UserID             string                   `json:"userId" bson:"userId"`
	Status             ConsentRecordStatus      `json:"status,omitempty" bson:"status"`
	AgeGroup           ConsentRecordAgeGroup    `json:"ageGroup,omitempty" bson:"ageGroup"`
	OwnerName          string                   `json:"ownerName,omitempty" bson:"ownerName"`
	ParentGuardianName *string                  `json:"parentGuardianName,omitempty" bson:"parentGuardianName"`
	GrantorType        ConsentRecordGrantorType `json:"grantorType,omitempty" bson:"grantorType"`
	Type               ConsentType              `json:"type,omitempty" bson:"type"`
	Version            int                      `json:"version,omitempty" bson:"version"`
	Metadata           *ConsentRecordMetadata   `json:"metadata,omitempty" bson:"metadata"`
	GrantTime          time.Time                `json:"grantTime" bson:"grantTime"`
	RevocationTime     *time.Time               `json:"revocationTime" bson:"revocationTime"`
	CreatedTime        time.Time                `json:"createdTime" bson:"createdTime"`
	ModifiedTime       time.Time                `json:"modifiedTime" bson:"modifiedTime"`
}

func NewConsentRecord(ctx context.Context, userID string, create *ConsentRecordCreate) (*ConsentRecord, error) {
	if userID == "" {
		return nil, errors.New("user id is missing")
	}
	if create == nil {
		return nil, errors.New("create is missing")
	}
	if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(create); err != nil {
		return nil, errors.Wrap(err, "create is invalid")
	}

	now := time.Now()
	return &ConsentRecord{
		ID:                 NewConsentRecordID(),
		UserID:             userID,
		Status:             ConsentRecordStatusActive,
		AgeGroup:           create.AgeGroup,
		OwnerName:          create.OwnerName,
		ParentGuardianName: create.ParentGuardianName,
		GrantorType:        create.GrantorType,
		Type:               create.Type,
		Version:            create.Version,
		Metadata:           create.Metadata,
		GrantTime:          now,
		CreatedTime:        now,
		ModifiedTime:       now,
	}, nil
}

func (c *ConsentRecord) Validate(validator structure.Validator) {
	validator.String("id", &c.ID).NotEmpty()
	validator.String("userId", &c.UserID).Exists().Using(UserIDValidator)
	validator.String("status", structure.ValueAsString(&c.Status)).OneOf(structure.ValuesAsStringArray(ConsentRecordStatuses())...)
	validator.String("ageGroup", structure.ValueAsString(&c.AgeGroup)).OneOf(structure.ValuesAsStringArray(ConsentRecordAgeGroups())...)
	validator.String("ownerName", &c.OwnerName).Exists().LengthInRange(1, 256)
	validator.String("parentGuardianName", c.ParentGuardianName).LengthInRange(1, 256)
	validator.String("grantorType", structure.ValueAsString(&c.GrantorType)).Exists().OneOf(structure.ValuesAsStringArray(ConsentRecordGrantorTypes())...)
	validator.String("type", structure.ValueAsString(&c.Type)).Exists().OneOf(structure.ValuesAsStringArray(ConsentTypes())...)
	validator.Int("version", &c.Version).Exists().GreaterThan(0)
	c.Metadata.Validator(c.Type)(validator.WithReference("metadata"))
	validator.Time("grantTime", &c.GrantTime).Exists().NotZero().BeforeNow(time.Second)
	validator.Time("revocationTime", c.RevocationTime).BeforeNow(time.Second)
	validator.Time("createdTime", &c.CreatedTime).Exists().BeforeNow(time.Second)
	validator.Time("modifiedTime", &c.ModifiedTime).Exists().BeforeNow(time.Second)
}

type ConsentRecordCreate struct {
	AgeGroup           ConsentRecordAgeGroup    `json:"ageGroup,omitempty" bson:"ageGroup"`
	OwnerName          string                   `json:"ownerName,omitempty" bson:"ownerName"`
	ParentGuardianName *string                  `json:"parentGuardianName,omitempty" bson:"parentGuardianName"`
	GrantorType        ConsentRecordGrantorType `json:"grantorType,omitempty" bson:"grantorType"`
	Type               ConsentType              `json:"type,omitempty" bson:"type"`
	Version            int                      `json:"version,omitempty" bson:"version"`
	Metadata           *ConsentRecordMetadata   `json:"metadata,omitempty" bson:"metadata"`
}

func NewConsentRecordCreate(ctx context.Context) *ConsentRecordCreate {
	return &ConsentRecordCreate{}
}

func (c *ConsentRecordCreate) Validate(validator structure.Validator) {
	validator.String("ageGroup", structure.ValueAsString(&c.AgeGroup)).OneOf(structure.ValuesAsStringArray(ConsentRecordAgeGroups())...)
	validator.String("ownerName", &c.OwnerName).Exists().LengthInRange(1, 256)
	validator.String("parentGuardianName", c.ParentGuardianName).LengthInRange(1, 256)
	validator.String("grantorType", structure.ValueAsString(&c.GrantorType)).Exists().OneOf(structure.ValuesAsStringArray(ConsentRecordGrantorTypes())...)
	validator.String("type", structure.ValueAsString(&c.Type)).Exists().OneOf(structure.ValuesAsStringArray(ConsentTypes())...)
	validator.Int("version", &c.Version).Exists().GreaterThan(0)
	c.Metadata.Validator(c.Type)(validator.WithReference("metadata"))
}

func NewConsentRecordID() string {
	return id.Must(id.New(16))
}

type ConsentRecordFilter struct {
	Latest  *bool
	Status  *ConsentRecordStatus
	Type    *ConsentType
	Version *int
	ID      *string
}

func NewConsentRecordFilter() *ConsentRecordFilter {
	return &ConsentRecordFilter{
		Latest: pointer.FromBool(true),
	}
}

func (p *ConsentRecordFilter) Parse(parser structure.ObjectParser) {
	latest := parser.Bool("latest")
	if latest != nil {
		p.Latest = latest
	}
	p.Status = NewConsentRecordStatus(parser.String("status"))
	p.Type = NewConsentType(parser.String("type"))
	p.Version = parser.Int("version")
}

func (p *ConsentRecordFilter) Validate(validator structure.Validator) {
	validator.String("id", p.ID).NotEmpty()
	validator.Bool("latest", p.Latest).Exists()
	validator.String("status", structure.ValueAsString(p.Status)).OneOf(structure.ValuesAsStringArray(ConsentRecordStatuses())...)
	validator.String("type", structure.ValueAsString(p.Type)).OneOf(structure.ValuesAsStringArray(ConsentTypes())...)
	validator.Int("version", p.Version).GreaterThan(0)
}

type ConsentRecordStatus string

func NewConsentRecordStatus(value *string) *ConsentRecordStatus {
	if value == nil {
		return nil
	}
	return pointer.FromAny(ConsentRecordStatus(*value))
}

func ConsentRecordStatuses() []ConsentRecordStatus {
	return []ConsentRecordStatus{
		ConsentRecordStatusActive,
		ConsentRecordStatusRevoked,
	}
}

type ConsentRecordAgeGroup string

func ConsentRecordAgeGroups() []ConsentRecordAgeGroup {
	return []ConsentRecordAgeGroup{
		ConsentRecordAgeGroupUnderTwelve,
		ConsentRecordAgeGroupThirteenSeventeen,
		ConsentRecordAgeGroupOverEighteen,
	}
}

type ConsentRecordGrantorType = string

func ConsentRecordGrantorTypes() []ConsentRecordGrantorType {
	return []ConsentRecordGrantorType{
		ConsentGrantorTypeOwner,
		ConsentGrantorTypeParentGuardian,
	}
}

type ConsentRecordMetadata struct {
	SupportedOrganizations []BigDataDonationProjectOrganization `json:"supportedOrganizations" bson:"supportedOrganizations"`
}

func (c *ConsentRecordMetadata) Validator(typ ConsentType) func(structure.Validator) {
	switch typ {
	case ConsentTypeBigDataDonationProject:
		return c.ValidateBigDataDonationProject
	default:
		return func(validator structure.Validator) {}
	}
}

func (c *ConsentRecordMetadata) ValidateBigDataDonationProject(validator structure.Validator) {
	validator.StringArray("supportedOrganizations", pointer.FromAny(structure.ValuesAsStringArray(c.SupportedOrganizations))).EachOneOf(structure.ValuesAsStringArray(BigDataDonationProjectOrganizations())...)
}

type BigDataDonationProjectOrganization string

func BigDataDonationProjectOrganizations() []BigDataDonationProjectOrganization {
	return []BigDataDonationProjectOrganization{
		BigDataDonationProjectOrganizationsADCES,
	}
}

type ConsentRecordUpdate struct {
	raw json.RawMessage

	Metadata *ConsentRecordMetadata `json:"metadata,omitempty" bson:"metadata"`
}

func NewConsentRecordUpdate(raw json.RawMessage) *ConsentRecordUpdate {
	return &ConsentRecordUpdate{
		raw: raw,
	}
}

func (c *ConsentRecordUpdate) ApplyPatch(ctx context.Context, record *ConsentRecord) error {
	validator := structureValidator.New(log.LoggerFromContext(ctx))

	c.Validate(record, validator)
	if validator.HasError() {
		return validator.Error()
	}

	if err := json.Unmarshal(c.raw, record); err != nil {
		return errors.Wrap(err, "unable to unmarshal patch")
	}

	return nil
}

func (c *ConsentRecordUpdate) Validate(record *ConsentRecord, validator structure.Validator) {
	c.Metadata.Validator(record.Type)(validator.WithReference("metadata"))
}
