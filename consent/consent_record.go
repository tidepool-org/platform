package consent

import (
	"context"
	"encoding/json"
	"time"

	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"

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

	AgeGroupUnderTwelve       AgeGroup = "<12"
	AgeGroupThirteenSeventeen AgeGroup = "13-17"
	AgeGroupOverEighteen      AgeGroup = ">18"

	BigDataDonationProjectOrganizationsADCES                 BigDataDonationProjectOrganization = "ADCES Foundation"
	BigDataDonationProjectOrganizationsBeyondType1           BigDataDonationProjectOrganization = "Beyond Type 1"
	BigDataDonationProjectOrganizationsChildrenWithDiabetes  BigDataDonationProjectOrganization = "Children With Diabetes"
	BigDataDonationProjectOrganizationsTheDiabetesLink       BigDataDonationProjectOrganization = "The Diabetes Link"
	BigDataDonationProjectOrganizationsDYF                   BigDataDonationProjectOrganization = "Diabetes Youth Families (DYF)"
	BigDataDonationProjectOrganizationsDiabetesSisters       BigDataDonationProjectOrganization = "DiabetesSisters"
	BigDataDonationProjectOrganizationsTheDiaTribeFoundation BigDataDonationProjectOrganization = "The diaTribe Foundation"
	BigDataDonationProjectOrganizationsBreakthroughT1D       BigDataDonationProjectOrganization = "Breakthrough T1D"
)

type RecordAccessor interface {
	GetConsentRecord(context.Context, string, string) (*Record, error)
	CreateConsentRecord(context.Context, string, *RecordCreate) (*Record, error)
	ListConsentRecords(context.Context, string, *RecordFilter, *page.Pagination) (*storeStructuredMongo.ListResult[Record], error)
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
	Type               string          `json:"type,omitempty" bson:"type"`
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
	validator.String("type", &c.Type).Exists().LengthInRange(TypeMinLength, TypeMaxLength)
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
	Type               string          `json:"type,omitempty" bson:"type"`
	Version            int             `json:"version,omitempty" bson:"version"`
}

func NewConsentRecordCreate() *RecordCreate {
	return &RecordCreate{
		CreatedTime: time.Now(),
	}
}

func (r *RecordCreate) Parse(parser structure.ObjectParser) {
	if ptr := parser.String("ageGroup"); ptr != nil {
		r.AgeGroup = AgeGroup(*ptr)
	}
	if ptr := parser.String("grantorType"); ptr != nil {
		r.GrantorType = *ptr
	}
	if metadataParser := parser.WithReferenceObjectParser("metadata"); metadataParser.Exists() {
		r.Metadata = &RecordMetadata{}
		r.Metadata.Parse(metadataParser)
	}
	if ptr := parser.String("ownerName"); ptr != nil {
		r.OwnerName = *ptr
	}
	r.ParentGuardianName = parser.String("parentGuardianName")
	if ptr := parser.String("type"); ptr != nil {
		r.Type = *ptr
	}
	if ptr := parser.Int("version"); ptr != nil {
		r.Version = *ptr
	}
}

func (r *RecordCreate) Validate(validator structure.Validator) {
	validator.String("ageGroup", structure.ValueAsString(&r.AgeGroup)).OneOf(structure.ValuesAsStringArray(AgeGroups())...)
	validator.Time("createdTime", &r.CreatedTime).NotZero().BeforeNow(time.Second)
	validator.String("grantorType", structure.ValueAsString(&r.GrantorType)).Exists().OneOf(structure.ValuesAsStringArray(GrantorTypes())...)
	r.Metadata.Validator(r.Type)(validator.WithReference("metadata"))
	validator.String("ownerName", &r.OwnerName).Exists().LengthInRange(1, 256)
	validator.String("parentGuardianName", r.ParentGuardianName).LengthInRange(1, 256)
	validator.String("type", &r.Type).Exists().LengthInRange(TypeMinLength, TypeMaxLength)
	validator.Int("version", &r.Version).Exists().GreaterThan(0)
}

func NewConsentRecordID() string {
	return id.Must(id.New(16))
}

type RecordFilter struct {
	Latest  *bool
	Status  *RecordStatus
	Type    *string
	Version *int
	ID      *string
}

func NewConsentRecordFilter() *RecordFilter {
	return &RecordFilter{
		Latest: pointer.FromBool(true),
	}
}

func (r *RecordFilter) Parse(parser structure.ObjectParser) {
	latest := parser.Bool("latest")
	if latest != nil {
		r.Latest = latest
	}
	r.Status = NewConsentRecordStatus(parser.String("status"))
	r.Type = parser.String("type")
	r.Version = parser.Int("version")
}

func (r *RecordFilter) Validate(validator structure.Validator) {
	validator.String("id", r.ID).NotEmpty()
	validator.Bool("latest", r.Latest).Exists()
	validator.String("status", structure.ValueAsString(r.Status)).OneOf(structure.ValuesAsStringArray(RecordStatuses())...)
	validator.String("type", r.Type).LengthInRange(TypeMinLength, TypeMaxLength)
	validator.Int("version", r.Version).GreaterThan(0)
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
		AgeGroupUnderTwelve,
		AgeGroupThirteenSeventeen,
		AgeGroupOverEighteen,
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

func (r *RecordMetadata) Parse(parser structure.ObjectParser) {
	if ptr := parser.StringArray("supportedOrganizations"); ptr != nil {
		r.SupportedOrganizations = make([]BigDataDonationProjectOrganization, len(*ptr))
		for i, v := range *ptr {
			r.SupportedOrganizations[i] = BigDataDonationProjectOrganization(v)
		}
	}
}

func (r *RecordMetadata) Validator(typ string) func(structure.Validator) {
	switch typ {
	case TypeBigDataDonationProject:
		return r.ValidateBigDataDonationProject
	default:
		return func(validator structure.Validator) {}
	}
}

func (r *RecordMetadata) ValidateBigDataDonationProject(validator structure.Validator) {
	validator.StringArray("supportedOrganizations", pointer.FromAny(structure.ValuesAsStringArray(r.SupportedOrganizations))).EachOneOf(structure.ValuesAsStringArray(BigDataDonationProjectOrganizations())...)
}

type BigDataDonationProjectOrganization string

func BigDataDonationProjectOrganizations() []BigDataDonationProjectOrganization {
	return []BigDataDonationProjectOrganization{
		BigDataDonationProjectOrganizationsADCES,
		BigDataDonationProjectOrganizationsBeyondType1,
		BigDataDonationProjectOrganizationsChildrenWithDiabetes,
		BigDataDonationProjectOrganizationsTheDiabetesLink,
		BigDataDonationProjectOrganizationsDYF,
		BigDataDonationProjectOrganizationsDiabetesSisters,
		BigDataDonationProjectOrganizationsTheDiaTribeFoundation,
		BigDataDonationProjectOrganizationsBreakthroughT1D,
	}
}

type RecordUpdate struct {
	raw    json.RawMessage
	record *Record

	Metadata *RecordMetadata `json:"metadata,omitempty" bson:"metadata"`
}

func NewConsentRecordUpdate(body []byte, record *Record) (*RecordUpdate, error) {
	raw := json.RawMessage{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	return &RecordUpdate{
		raw:    raw,
		record: record,
	}, nil
}

func (r *RecordUpdate) ApplyPatch() (*Record, error) {
	if err := json.Unmarshal(r.raw, r.record); err != nil {
		return nil, errors.Wrap(err, "unable to unmarshal patch")
	}

	return r.record, nil
}

func (r *RecordUpdate) Parse(parser structure.ObjectParser) {
	if metadataParser := parser.WithReferenceObjectParser("metadata"); metadataParser.Exists() {
		r.Metadata = &RecordMetadata{}
		r.Metadata.Parse(metadataParser)
	}
}

func (r *RecordUpdate) Validate(validator structure.Validator) {
	r.Metadata.Validator(r.record.Type)(validator.WithReference("metadata"))
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

func (r *RecordRevoke) Validate(validator structure.Validator) {
	validator.String("id", &r.ID).NotEmpty()
	validator.Time("revocationTime", &r.RevocationTime).NotZero()
}
