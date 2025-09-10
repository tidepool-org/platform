package consent

import (
	"context"
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

	AgeGroupUnderThirteen     AgeGroup = "<13"
	AgeGroupThirteenSeventeen AgeGroup = "13-17"
	AgeGroupEighteenOrOver    AgeGroup = ">=18"

	BigDataDonationProjectOrganizationsADCES                 BigDataDonationProjectOrganization = "ADCES Foundation"
	BigDataDonationProjectOrganizationsBeyondType1           BigDataDonationProjectOrganization = "Beyond Type 1"
	BigDataDonationProjectOrganizationsChildrenWithDiabetes  BigDataDonationProjectOrganization = "Children With Diabetes"
	BigDataDonationProjectOrganizationsTheDiabetesLink       BigDataDonationProjectOrganization = "The Diabetes Link"
	BigDataDonationProjectOrganizationsDYF                   BigDataDonationProjectOrganization = "Diabetes Youth Families (DYF)"
	BigDataDonationProjectOrganizationsDiabetesSisters       BigDataDonationProjectOrganization = "DiabetesSisters"
	BigDataDonationProjectOrganizationsTheDiaTribeFoundation BigDataDonationProjectOrganization = "The diaTribe Foundation"
	BigDataDonationProjectOrganizationsBreakthroughT1D       BigDataDonationProjectOrganization = "Breakthrough T1D"
	BigDataDonationProjectOrganizationsNightscoutFoundation  BigDataDonationProjectOrganization = "Nightscout Foundation"
	BigDataDonationProjectOrganizationsT1DExchange           BigDataDonationProjectOrganization = "T1D Exchange"
)

type RecordAccessor interface {
	GetConsentRecord(context.Context, string, string) (*Record, error)
	CreateConsentRecord(context.Context, string, *RecordCreate) (*Record, error)
	ListConsentRecords(context.Context, string, *RecordFilter, *page.Pagination) (*storeStructuredMongo.ListResult[Record], error)
	RevokeConsentRecord(context.Context, string, *RecordRevoke) error
	UpdateConsentRecord(context.Context, *Record) (*Record, error)
}

type Record struct {
	ID                 string          `json:"id" bson:"id"`
	UserID             string          `json:"userId" bson:"userId"`
	Status             RecordStatus    `json:"status" bson:"status"`
	AgeGroup           AgeGroup        `json:"ageGroup" bson:"ageGroup"`
	OwnerName          string          `json:"ownerName" bson:"ownerName"`
	ParentGuardianName *string         `json:"parentGuardianName,omitempty" bson:"parentGuardianName,omitempty"`
	GrantorType        GrantorType     `json:"grantorType" bson:"grantorType"`
	Type               string          `json:"type" bson:"type"`
	Version            int             `json:"version" bson:"version"`
	Metadata           *RecordMetadata `json:"metadata" bson:"metadata"`
	GrantTime          time.Time       `json:"grantTime" bson:"grantTime"`
	RevocationTime     *time.Time      `json:"revocationTime,omitempty" bson:"revocationTime,omitempty"`
	CreatedTime        time.Time       `json:"createdTime" bson:"createdTime"`
	ModifiedTime       time.Time       `json:"modifiedTime" bson:"modifiedTime"`
}

func NewRecord(ctx context.Context, userID string, create *RecordCreate) (*Record, error) {
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
		ID:                 NewRecordID(),
		UserID:             userID,
		Status:             RecordStatusActive,
		AgeGroup:           create.AgeGroup,
		OwnerName:          create.OwnerName,
		ParentGuardianName: create.ParentGuardianName,
		GrantorType:        create.GrantorType,
		Type:               create.Type,
		Version:            create.Version,
		Metadata:           create.Metadata,
		GrantTime:          create.GrantTime,
		CreatedTime:        create.CreatedTime,
		ModifiedTime:       time.Now(),
	}, nil
}

func (r *Record) Validate(validator structure.Validator) {
	validator.String("id", &r.ID).NotEmpty()
	validator.String("userId", &r.UserID).Exists().Using(auth.UserIDValidator)
	validator.String("status", structure.ValueAsString(&r.Status)).OneOf(structure.ValuesAsStringArray(RecordStatuses())...)
	validator.String("ageGroup", structure.ValueAsString(&r.AgeGroup)).OneOf(structure.ValuesAsStringArray(AgeGroups())...)
	validator.String("ownerName", &r.OwnerName).Exists().LengthInRange(1, 256)
	validator.String("grantorType", structure.ValueAsString(&r.GrantorType)).Exists().OneOf(structure.ValuesAsStringArray(GrantorTypes())...)
	validator.String("type", &r.Type).Exists().LengthInRange(TypeMinLength, TypeMaxLength)
	validator.Int("version", &r.Version).Exists().GreaterThanOrEqualTo(0)

	validator.Time("grantTime", &r.GrantTime).Exists().NotZero().BeforeNow(time.Second)
	validator.Time("revocationTime", r.RevocationTime).BeforeNow(time.Second)
	validator.Time("createdTime", &r.CreatedTime).Exists().BeforeNow(time.Second)
	validator.Time("modifiedTime", &r.ModifiedTime).Exists().BeforeNow(time.Second)

	r.Metadata.Validator(r.Type)(validator.WithReference("metadata"))

	parentGuardianNameValidator := validator.String("parentGuardianName", r.ParentGuardianName).LengthInRange(1, 256)
	grantorTypeValidator := validator.String("grantorType", structure.ValueAsString(&r.GrantorType)).Exists()

	if r.AgeGroup != AgeGroupEighteenOrOver {
		parentGuardianNameValidator.Exists()
		grantorTypeValidator.EqualTo(GrantorTypeParentGuardian)
	} else {
		parentGuardianNameValidator.NotExists()
	}
}

func (r *Record) ToUpdate() *RecordUpdate {
	update := RecordUpdate(*r)
	return &update
}

type RecordCreate struct {
	AgeGroup           AgeGroup        `json:"ageGroup" bson:"ageGroup"`
	CreatedTime        time.Time       `bson:"createdTime"`
	GrantTime          time.Time       `bson:"grantTime"`
	GrantorType        GrantorType     `json:"grantorType" bson:"grantorType"`
	Metadata           *RecordMetadata `json:"metadata" bson:"metadata"`
	OwnerName          string          `json:"ownerName" bson:"ownerName"`
	ParentGuardianName *string         `json:"parentGuardianName,omitempty" bson:"parentGuardianName"`
	Type               string          `json:"type" bson:"type"`
	Version            int             `json:"version" bson:"version"`
}

func NewRecordCreate() *RecordCreate {
	now := time.Now()
	return &RecordCreate{
		CreatedTime: now,
		GrantTime:   now,
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
	validator.Int("version", &r.Version).Exists().GreaterThanOrEqualTo(0)

	parentGuardianNameValidator := validator.String("parentGuardianName", r.ParentGuardianName).LengthInRange(1, 256)
	grantorTypeValidator := validator.String("grantorType", structure.ValueAsString(&r.GrantorType)).Exists()

	if r.AgeGroup == AgeGroupEighteenOrOver {
		parentGuardianNameValidator.NotExists()
		grantorTypeValidator.EqualTo(GrantorTypeOwner)
	} else {
		parentGuardianNameValidator.Exists()
		grantorTypeValidator.EqualTo(GrantorTypeParentGuardian)
	}
}

func NewRecordID() string {
	return id.Must(id.New(16))
}

type RecordFilter struct {
	Latest  *bool
	Status  *RecordStatus
	Type    *string
	Version *int
	ID      *string
}

func NewRecordFilter() *RecordFilter {
	return &RecordFilter{
		Latest: pointer.FromBool(true),
	}
}

func (r *RecordFilter) Parse(parser structure.ObjectParser) {
	latest := parser.Bool("latest")
	if latest != nil {
		r.Latest = latest
	}
	r.Status = NewRecordStatus(parser.String("status"))
	r.Type = parser.String("type")
	r.Version = parser.Int("version")
}

func (r *RecordFilter) Validate(validator structure.Validator) {
	validator.String("id", r.ID).NotEmpty()
	validator.Bool("latest", r.Latest).Exists()
	validator.String("status", structure.ValueAsString(r.Status)).OneOf(structure.ValuesAsStringArray(RecordStatuses())...)
	validator.String("type", r.Type).LengthInRange(TypeMinLength, TypeMaxLength)
	validator.Int("version", r.Version).GreaterThanOrEqualTo(0)
}

type RecordStatus string

func NewRecordStatus(value *string) *RecordStatus {
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
		AgeGroupUnderThirteen,
		AgeGroupThirteenSeventeen,
		AgeGroupEighteenOrOver,
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

func (r *RecordMetadata) Validate(validator structure.Validator) {
	if r == nil {
		return
	}

	validator.StringArray("supportedOrganizations", pointer.FromAny(structure.ValuesAsStringArray(r.SupportedOrganizations))).Empty()
}

func (r *RecordMetadata) Parse(parser structure.ObjectParser) {
	if ptr := parser.StringArray("supportedOrganizations"); ptr != nil {
		r.SupportedOrganizations = make([]BigDataDonationProjectOrganization, len(*ptr))
		for i, v := range *ptr {
			r.SupportedOrganizations[i] = BigDataDonationProjectOrganization(v)
		}
	} else {
		r.SupportedOrganizations = nil
	}
}

func (r *RecordMetadata) Validator(typ string) func(structure.Validator) {
	if r != nil && typ == TypeBigDataDonationProject {
		return r.ValidateBigDataDonationProject
	}
	return r.Validate
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
		BigDataDonationProjectOrganizationsNightscoutFoundation,
		BigDataDonationProjectOrganizationsT1DExchange,
	}
}

// RecordUpdate extends Record to allow only specific fields to be modified
type RecordUpdate Record

func (r *RecordUpdate) Parse(parser structure.ObjectParser) {
	if metadataParser := parser.WithReferenceObjectParser("metadata"); metadataParser.Exists() {
		if r.Metadata == nil {
			r.Metadata = &RecordMetadata{}
		}
		r.Metadata.Parse(metadataParser)
	} else {
		r.Metadata = nil
	}
}

func (r *RecordUpdate) Validate(validator structure.Validator) {
	r.Metadata.Validator(r.Type)(validator.WithReference("metadata"))
}

func (r *RecordUpdate) ToRecord() *Record {
	record := Record(*r)
	return &record
}

type RecordRevoke struct {
	ID             string
	RevocationTime time.Time
}

func NewRecordRevoke() *RecordRevoke {
	return &RecordRevoke{
		RevocationTime: time.Now(),
	}
}

func (r *RecordRevoke) Validate(validator structure.Validator) {
	validator.String("id", &r.ID).NotEmpty()
	validator.Time("revocationTime", &r.RevocationTime).NotZero()
}
