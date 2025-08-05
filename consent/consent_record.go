package consent

import (
	"context"
	"encoding/json"
	"time"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	RecordStatusActive  RecordStatus = "active"
	RecordStatusRevoked RecordStatus = "revoked"

	GrantorTypeOwner          GrantorType = "owner"
	GrantorTypeParentGuardian GrantorType = "parent/guardian"

	ConsentRecordAgeGroupUnderTwelve       AgeGroup = "<12"
	ConsentRecordAgeGroupThirteenSeventeen AgeGroup = "13-17"
	ConsentRecordAgeGroupOverEighteen      AgeGroup = ">18"

	BigDataDonationProjectOrganizationsADCES BigDataDonationProjectOrganization = "adces"
)

type RecordAccessor interface {
	GetConsentRecord(context.Context, string, string) (*Record, error)
	CreateConsentRecord(context.Context, string, *RecordCreate) (*Record, error)
	ListConsentRecords(context.Context, string, *RecordFilter, *page.Pagination) (Records, error)
	RevokeConsentRecord(context.Context, string, *RecordRevoke) error
	UpdateConsentRecord(context.Context, *Record) (*Record, error)
}

type Records []Record
type Record struct {
	ID                 string          `json:"id,omitempty" bson:"id"`
	UserID             string          `json:"userId" bson:"userId"`
	Status             RecordStatus    `json:"status,omitempty" bson:"status"`
	AgeGroup           AgeGroup        `json:"ageGroup,omitempty" bson:"ageGroup"`
	OwnerName          string          `json:"ownerName,omitempty" bson:"ownerName"`
	ParentGuardianName *string         `json:"parentGuardianName,omitempty" bson:"parentGuardianName"`
	GrantorType        GrantorType     `json:"grantorType,omitempty" bson:"grantorType"`
	Type               Type            `json:"type,omitempty" bson:"type"`
	Version            int             `json:"version,omitempty" bson:"version"`
	Metadata           *RecordMetadata `json:"metadata,omitempty" bson:"metadata"`
	GrantTime          time.Time       `json:"grantTime" bson:"grantTime"`
	RevocationTime     *time.Time      `json:"revocationTime" bson:"revocationTime"`
	CreatedTime        time.Time       `json:"createdTime" bson:"createdTime"`
	ModifiedTime       time.Time       `json:"modifiedTime" bson:"modifiedTime"`
}

func NewConsentRecord(ctx context.Context, userID string, create *RecordCreate) (*Record, error) {
	if userID == "" {
		return nil, errors.New("user id is missing")
	}
	if create == nil {
		return nil, errors.New("create is missing")
	}
	if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(create); err != nil {
		return nil, errors.Wrap(err, "create is invalid")
	}

	return &Record{
		ID:                 NewConsentRecordID(),
		UserID:             userID,
		Status:             RecordStatusActive,
		AgeGroup:           create.AgeGroup,
		OwnerName:          create.OwnerName,
		ParentGuardianName: create.ParentGuardianName,
		GrantorType:        create.GrantorType,
		Type:               create.Type,
		Version:            create.Version,
		Metadata:           create.Metadata,
		GrantTime:          create.CreatedTime,
		CreatedTime:        create.CreatedTime,
		ModifiedTime:       time.Now(),
	}, nil
}

func (c *Record) Validate(validator structure.Validator) {
	validator.String("id", &c.ID).NotEmpty()
	validator.String("userId", &c.UserID).Exists().Using(auth.UserIDValidator)
	validator.String("status", structure.ValueAsString(&c.Status)).OneOf(structure.ValuesAsStringArray(RecordStatuses())...)
	validator.String("ageGroup", structure.ValueAsString(&c.AgeGroup)).OneOf(structure.ValuesAsStringArray(AgeGroups())...)
	validator.String("ownerName", &c.OwnerName).Exists().LengthInRange(1, 256)
	validator.String("parentGuardianName", c.ParentGuardianName).LengthInRange(1, 256)
	validator.String("grantorType", structure.ValueAsString(&c.GrantorType)).Exists().OneOf(structure.ValuesAsStringArray(GrantorTypes())...)
	validator.String("type", structure.ValueAsString(&c.Type)).Exists().OneOf(structure.ValuesAsStringArray(Types())...)
	validator.Int("version", &c.Version).Exists().GreaterThan(0)
	c.Metadata.Validator(c.Type)(validator.WithReference("metadata"))
	validator.Time("grantTime", &c.GrantTime).Exists().NotZero().BeforeNow(time.Second)
	validator.Time("revocationTime", c.RevocationTime).BeforeNow(time.Second)
	validator.Time("createdTime", &c.CreatedTime).Exists().BeforeNow(time.Second)
	validator.Time("modifiedTime", &c.ModifiedTime).Exists().BeforeNow(time.Second)
}

type RecordCreate struct {
	AgeGroup           AgeGroup        `json:"ageGroup,omitempty" bson:"ageGroup"`
	CreatedTime        time.Time       `bson:"createdTime"`
	GrantorType        GrantorType     `json:"grantorType,omitempty" bson:"grantorType"`
	Metadata           *RecordMetadata `json:"metadata,omitempty" bson:"metadata"`
	OwnerName          string          `json:"ownerName,omitempty" bson:"ownerName"`
	ParentGuardianName *string         `json:"parentGuardianName,omitempty" bson:"parentGuardianName"`
	Type               Type            `json:"type,omitempty" bson:"type"`
	Version            int             `json:"version,omitempty" bson:"version"`
}

func NewConsentRecordCreate() *RecordCreate {
	return &RecordCreate{
		CreatedTime: time.Now(),
	}
}

func (c *RecordCreate) Validate(validator structure.Validator) {
	validator.String("ageGroup", structure.ValueAsString(&c.AgeGroup)).OneOf(structure.ValuesAsStringArray(AgeGroups())...)
	validator.Time("createdTime", &c.CreatedTime).NotZero().BeforeNow(time.Second)
	validator.String("grantorType", structure.ValueAsString(&c.GrantorType)).Exists().OneOf(structure.ValuesAsStringArray(GrantorTypes())...)
	c.Metadata.Validator(c.Type)(validator.WithReference("metadata"))
	validator.String("ownerName", &c.OwnerName).Exists().LengthInRange(1, 256)
	validator.String("parentGuardianName", c.ParentGuardianName).LengthInRange(1, 256)
	validator.String("type", structure.ValueAsString(&c.Type)).Exists().OneOf(structure.ValuesAsStringArray(Types())...)
	validator.Int("version", &c.Version).Exists().GreaterThan(0)
}

func NewConsentRecordID() string {
	return id.Must(id.New(16))
}

type RecordFilter struct {
	Latest  *bool
	Status  *RecordStatus
	Type    *Type
	Version *int
	ID      *string
}

func NewConsentRecordFilter() *RecordFilter {
	return &RecordFilter{
		Latest: pointer.FromBool(true),
	}
}

func (p *RecordFilter) Parse(parser structure.ObjectParser) {
	latest := parser.Bool("latest")
	if latest != nil {
		p.Latest = latest
	}
	p.Status = NewConsentRecordStatus(parser.String("status"))
	p.Type = NewConsentType(parser.String("type"))
	p.Version = parser.Int("version")
}

func (p *RecordFilter) Validate(validator structure.Validator) {
	validator.String("id", p.ID).NotEmpty()
	validator.Bool("latest", p.Latest).Exists()
	validator.String("status", structure.ValueAsString(p.Status)).OneOf(structure.ValuesAsStringArray(RecordStatuses())...)
	validator.String("type", structure.ValueAsString(p.Type)).OneOf(structure.ValuesAsStringArray(Types())...)
	validator.Int("version", p.Version).GreaterThan(0)
}

type RecordStatus string

func NewConsentRecordStatus(value *string) *RecordStatus {
	if value == nil {
		return nil
	}
	return pointer.FromAny(RecordStatus(*value))
}

func RecordStatuses() []RecordStatus {
	return []RecordStatus{
		RecordStatusActive,
		RecordStatusRevoked,
	}
}

type AgeGroup string

func AgeGroups() []AgeGroup {
	return []AgeGroup{
		ConsentRecordAgeGroupUnderTwelve,
		ConsentRecordAgeGroupThirteenSeventeen,
		ConsentRecordAgeGroupOverEighteen,
	}
}

type GrantorType = string

func GrantorTypes() []GrantorType {
	return []GrantorType{
		GrantorTypeOwner,
		GrantorTypeParentGuardian,
	}
}

type RecordMetadata struct {
	SupportedOrganizations []BigDataDonationProjectOrganization `json:"supportedOrganizations" bson:"supportedOrganizations"`
}

func (c *RecordMetadata) Validator(typ Type) func(structure.Validator) {
	switch typ {
	case TypeBigDataDonationProject:
		return c.ValidateBigDataDonationProject
	default:
		return func(validator structure.Validator) {}
	}
}

func (c *RecordMetadata) ValidateBigDataDonationProject(validator structure.Validator) {
	validator.StringArray("supportedOrganizations", pointer.FromAny(structure.ValuesAsStringArray(c.SupportedOrganizations))).EachOneOf(structure.ValuesAsStringArray(BigDataDonationProjectOrganizations())...)
}

type BigDataDonationProjectOrganization string

func BigDataDonationProjectOrganizations() []BigDataDonationProjectOrganization {
	return []BigDataDonationProjectOrganization{
		BigDataDonationProjectOrganizationsADCES,
	}
}

type RecordUpdate struct {
	raw json.RawMessage

	Metadata *RecordMetadata `json:"metadata,omitempty" bson:"metadata"`
}

func NewConsentRecordUpdate(body []byte) (*RecordUpdate, error) {
	raw := json.RawMessage{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	return &RecordUpdate{
		raw: raw,
	}, nil
}

func (c *RecordUpdate) ApplyPatch(ctx context.Context, record *Record) error {
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

func (c *RecordUpdate) Validate(record *Record, validator structure.Validator) {
	c.Metadata.Validator(record.Type)(validator.WithReference("metadata"))
}

type RecordRevoke struct {
	ID             string
	RevocationTime time.Time
}

func NewConsentRecordRevoke() *RecordRevoke {
	return &RecordRevoke{
		RevocationTime: time.Now(),
	}
}

func (c *RecordRevoke) Validate(validator structure.Validator) {
	validator.String("id", &c.ID).NotEmpty()
	validator.Time("revocationTime", &c.RevocationTime).NotZero()
}
