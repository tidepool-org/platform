package prescription

import (
	"regexp"
	"time"

	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/user"
)

const (
	TrainingInPerson = "inPerson"
	TrainingInModule = "inModule"

	TherapySettingInitial              = "initial"
	TherapySettingTransferPumpSettings = "transferPumpSettings"

	SexMale        = "male"
	SexFemale      = "female"
	SexUndisclosed = "undisclosed"

	UnitKg = "kg"

	AccountTypePatient   = "patient"
	AccountTypeCaregiver = "caregiver"

	usPhoneNumberRegexString = "^\\d{10}|\\(\\d{3}\\) ?\\d{3}\\-\\d{4}$" // Matches 1234567890, (123)456-7890 or (123) 456-7890
	usPhoneNumberCountryCode = 1
)

var (
	usPhoneNumberRegex = regexp.MustCompile(usPhoneNumberRegexString)
)

type RevisionCreate struct {
	AccountType             string           `json:"accountType,omitempty"`
	CaregiverFirstName      string           `json:"caregiverFirstName,omitempty"`
	CaregiverLastName       string           `json:"caregiverLastName,omitempty"`
	FirstName               string           `json:"firstName,omitempty"`
	LastName                string           `json:"lastName,omitempty"`
	Birthday                string           `json:"birthday,omitempty"`
	MRN                     string           `json:"mrn,omitempty"`
	Email                   string           `json:"email,omitempty"`
	Sex                     string           `json:"sex,omitempty"`
	Weight                  *Weight          `json:"weight,omitempty"`
	YearOfDiagnosis         int              `json:"yearOfDiagnosis,omitempty"`
	PhoneNumber             *PhoneNumber     `json:"phoneNumber,omitempty"`
	InitialSettings         *InitialSettings `json:"initialSettings,omitempty"`
	Training                string           `json:"training,omitempty"`
	TherapySettings         string           `json:"therapySettings,omitempty"`
	PrescriberTermsAccepted bool             `json:"prescriberTermsAccepted,omitempty"`
	State                   string           `json:"state"`
}

func NewRevisionCreate() *RevisionCreate {
	return &RevisionCreate{}
}

func (r *RevisionCreate) Validate(validator structure.Validator) {
	if r.AccountType != "" {
		validator.String("accountType", &r.AccountType).OneOf(AccountTypes()...)
		if r.AccountType == AccountTypePatient {
			validator.String("caregiverFirstName", &r.CaregiverFirstName).Empty()
			validator.String("caregiverLastName", &r.CaregiverLastName).Empty()
		}
	}
	if r.Birthday != "" {
		validator.String("birthday", &r.Birthday).AsTime("2006-01-02").NotZero().BeforeNow(time.Second)
	}
	if r.Email != "" {
		validator.String("email", &r.Email).Email()
	}
	if r.Sex != "" {
		validator.String("sex", &r.Sex).OneOf(SexValues()...)
	}
	if r.Weight != nil {
		r.Weight.Validate(validator.WithReference("weight"))
	}
	if r.YearOfDiagnosis != 0 {
		validator.Int("yearOfDiagnosis", &r.YearOfDiagnosis).GreaterThan(1900)
	}
	if r.PhoneNumber != nil {
		r.PhoneNumber.Validate(validator.WithReference("phoneNumber"))
	}
	if r.InitialSettings != nil {
		r.InitialSettings.Validate(validator.WithReference("initialSettings"))
	}
	if r.Training != "" {
		validator.String("training", &r.Training).OneOf(Trainings()...)
	}
	if r.TherapySettings != "" {
		validator.String("therapySettings", &r.TherapySettings).OneOf(TherapySettings()...)
	}
	validator.String("state", &r.State).OneOf(RevisionStates()...)
	if r.State == StateSubmitted {
		r.ValidateAllRequired(validator)
	}
}

func (r *RevisionCreate) ValidateAllRequired(validator structure.Validator) {
	validator.String("accountType", &r.AccountType).NotEmpty()
	if r.AccountType == AccountTypeCaregiver {
		validator.String("caregiverFirstName", &r.CaregiverFirstName).NotEmpty()
		validator.String("caregiverLastName", &r.CaregiverLastName).NotEmpty()
	}
	validator.String("firstName", &r.FirstName).NotEmpty()
	validator.String("lastName", &r.LastName).NotEmpty()
	validator.String("birthday", &r.Birthday).NotEmpty()
	validator.String("email", &r.Email).NotEmpty()
	validator.String("sex", &r.Sex).NotEmpty()
	validator.Int("yearOfDiagnosis", &r.YearOfDiagnosis).GreaterThan(1900)
	validator.String("training", &r.Training).NotEmpty()
	validator.String("therapySettings", &r.TherapySettings).NotEmpty()
	validator.Bool("prescriberTermsAccepted", &r.PrescriberTermsAccepted).True()

	// if phoneNumber is nil validate will fail
	phoneValidator := validator.WithReference("phoneNumber")
	if r.PhoneNumber != nil {
		r.PhoneNumber.Validate(phoneValidator)
	}

	// if weight is nil validate will fail
	weightValidator := validator.WithReference("weight")
	if r.Weight != nil {
		r.Weight.ValidateAllRequired(weightValidator)
	}

	initialSettingsValidator := validator.WithReference("initialSettings")
	if r.InitialSettings != nil {
		r.InitialSettings.ValidateAllRequired(initialSettingsValidator)
	} else {
		initialSettingsValidator.ReportError(structureValidator.ErrorValueEmpty())
	}
}

type Signature struct {
	Value  string `json:"signature" bson:"signature"`
	UserID string `json:"signatureUserId" bson:"signatureUserId"`
	KeyID  string `json:"signatureKeyId" bson:"signatureKeyId"`
}

type Revision struct {
	RevisionID int         `json:"revisionId" bson:"revisionId"`
	Signature  *Signature  `json:"signature,omitempty" bson:"signature,omitempty"`
	Attributes *Attributes `json:"attributes" bson:"attributes"`
}

type Revisions []*Revision

func NewRevision(userID string, revisionID int, create *RevisionCreate) *Revision {
	now := time.Now()
	return &Revision{
		RevisionID: revisionID,
		Attributes: &Attributes{
			AccountType:             create.AccountType,
			CaregiverFirstName:      create.CaregiverFirstName,
			CaregiverLastName:       create.CaregiverLastName,
			FirstName:               create.FirstName,
			LastName:                create.LastName,
			Birthday:                create.Birthday,
			MRN:                     create.MRN,
			Email:                   create.Email,
			Sex:                     create.Sex,
			Weight:                  create.Weight,
			YearOfDiagnosis:         create.YearOfDiagnosis,
			PhoneNumber:             create.PhoneNumber,
			InitialSettings:         create.InitialSettings,
			Training:                create.Training,
			TherapySettings:         create.TherapySettings,
			PrescriberTermsAccepted: create.PrescriberTermsAccepted,
			State:                   create.State,
			CreatedTime:             now,
			CreatedUserID:           userID,
		},
	}
}

func (r *Revision) Validate(validator structure.Validator) {
	validator.Int("revisionId", &r.RevisionID).GreaterThanOrEqualTo(0)
	attributesValidator := validator.WithReference("attributes")
	if r.Attributes != nil {
		r.Attributes.Validate(attributesValidator)
	} else {
		attributesValidator.ReportError(structureValidator.ErrorValueEmpty())
	}
}

func (r *Revision) CalculateExpirationTime() *time.Time {
	if r.Attributes.State != StateSubmitted {
		return nil
	}

	expiration := time.Now().Add(MaximumExpirationTime)
	return &expiration
}

func (r *Revision) GetPrescriberUserID() string {
	if r.Attributes.State != StateSubmitted {
		return ""
	}

	return r.Attributes.CreatedUserID
}

type Attributes struct {
	AccountType             string           `json:"accountType,omitempty" bson:"accountType,omitempty"`
	CaregiverFirstName      string           `json:"caregiverFirstName,omitempty" bson:"caregiverFirstName,omitempty"`
	CaregiverLastName       string           `json:"caregiverLastName,omitempty" bson:"caregiverLastName,omitempty"`
	FirstName               string           `json:"firstName,omitempty" bson:"firstName,omitempty"`
	LastName                string           `json:"lastName,omitempty" bson:"lastName,omitempty"`
	Birthday                string           `json:"birthday,omitempty" bson:"birthday,omitempty"`
	MRN                     string           `json:"mrn,omitempty" bson:"mrn,omitempty"`
	Email                   string           `json:"email,omitempty" bson:"email,omitempty"`
	Sex                     string           `json:"sex,omitempty" bson:"sex,omitempty"`
	Weight                  *Weight          `json:"weight,omitempty" bson:"weight,omitempty"`
	YearOfDiagnosis         int              `json:"yearOfDiagnosis,omitempty" bson:"yearOfDiagnosis,omitempty"`
	PhoneNumber             *PhoneNumber     `json:"phoneNumber,omitempty" bson:"phoneNumber,omitempty"`
	InitialSettings         *InitialSettings `json:"initialSettings,omitempty" bson:"initialSettings,omitempty"`
	Training                string           `json:"training,omitempty" bson:"training,omitempty"`
	TherapySettings         string           `json:"therapySettings,omitempty" bson:"therapySettings,omitempty"`
	PrescriberTermsAccepted bool             `json:"prescriberTermsAccepted,omitempty" bson:"prescriberTermsAccepted,omitempty"`
	State                   string           `json:"state" bson:"state"`
	CreatedTime             time.Time        `json:"createdTime,omitempty" bson:"createdTime,omitempty"`
	CreatedUserID           string           `json:"createdUserId,omitempty" bson:"cratedUserId,omitempty"`
}

func (a *Attributes) Validate(validator structure.Validator) {
	if a.AccountType != "" {
		validator.String("accountType", &a.AccountType).OneOf(AccountTypes()...)
		if a.AccountType == AccountTypePatient {
			validator.String("caregiverFirstName", &a.CaregiverFirstName).Empty()
			validator.String("caregiverLastName", &a.CaregiverLastName).Empty()
		}
	}
	if a.Birthday != "" {
		validator.String("birthday", &a.Birthday).AsTime("2006-01-02").NotZero().BeforeNow(time.Second)
	}
	if a.Email != "" {
		validator.String("email", &a.Email).Email()
	}
	if a.Sex != "" {
		validator.String("sex", &a.Sex).OneOf(SexValues()...)
	}
	if a.YearOfDiagnosis != 0 {
		validator.Int("yearOfDiagnosis", &a.YearOfDiagnosis).GreaterThan(1900)
	}
	if a.PhoneNumber != nil {
		a.PhoneNumber.Validate(validator.WithReference("phoneNumber"))
	}
	if a.Training != "" {
		validator.String("training", &a.Training).OneOf(Trainings()...)
	}
	if a.TherapySettings != "" {
		validator.String("therapySettings", &a.TherapySettings).OneOf(TherapySettings()...)
	}
	if a.Weight != nil {
		a.Weight.Validate(validator.WithReference("weight"))
	}
	if a.InitialSettings != nil {
		a.InitialSettings.Validate(validator.WithReference("initialSettings"))
	}
	validator.String("state", &a.State).OneOf(RevisionStates()...)
	validator.Time("createdTime", &a.CreatedTime).BeforeNow(time.Second)
	validator.String("createdUserId", &a.CreatedUserID).Using(user.IDValidator)

	if a.State == StateSubmitted {
		a.ValidateAllRequired(validator)
	}
}

func (a *Attributes) ValidateAllRequired(validator structure.Validator) {
	validator.String("accountType", &a.AccountType).NotEmpty()
	if a.AccountType == AccountTypeCaregiver {
		validator.String("caregiverFirstName", &a.CaregiverFirstName).NotEmpty()
		validator.String("caregiverLastName", &a.CaregiverLastName).NotEmpty()
	}
	validator.String("firstName", &a.FirstName).NotEmpty()
	validator.String("lastName", &a.LastName).NotEmpty()
	validator.String("birthday", &a.Birthday).NotEmpty()
	validator.String("email", &a.Email).NotEmpty()
	validator.String("sex", &a.Sex).NotEmpty()
	validator.Int("yearOfDiagnosis", &a.YearOfDiagnosis).GreaterThan(1900)
	validator.String("training", &a.Training).NotEmpty()
	validator.String("therapySettings", &a.TherapySettings).NotEmpty()
	validator.Bool("prescriberTermsAccepted", &a.PrescriberTermsAccepted).True()

	// if phoneNumber is nil validate will fail
	phoneValidator := validator.WithReference("phoneNumber")
	if a.PhoneNumber != nil {
		a.PhoneNumber.Validate(phoneValidator)
	} else {
		phoneValidator.ReportError(structureValidator.ErrorValueEmpty())
	}

	weightValidator := validator.WithReference("weight")
	if a.Weight != nil {
		a.Weight.ValidateAllRequired(weightValidator)
	}

	initialSettingsValidator := validator.WithReference("initialSettings")
	if a.InitialSettings != nil {
		a.InitialSettings.ValidateAllRequired(initialSettingsValidator)
	} else {
		initialSettingsValidator.ReportError(structureValidator.ErrorValueEmpty())
	}
}

type Weight struct {
	Value *float64 `json:"value,omitempty" bson:"value,omitempty"`
	Units string   `json:"units,omitempty" bson:"units,omitempty"`
}

func (w *Weight) Validate(validator structure.Validator) {
	if w.Value != nil {
		validator.Float64("value", w.Value).GreaterThan(0)
	}
	if w.Units != "" {
		validator.String("units", &w.Units).EqualTo(UnitKg)
	}
}

func (w *Weight) ValidateAllRequired(validator structure.Validator) {
	validator.Float64("value", w.Value).GreaterThan(0)
	validator.String("units", &w.Units).NotEmpty()
}

type PhoneNumber struct {
	CountryCode int    `json:"countryCode,omitempty" bson:"value,omitempty"`
	Number      string `json:"number,omitempty" bson:"number,omitempty"`
}

func (p *PhoneNumber) Validate(validator structure.Validator) {
	validator.Int("countryCode", &p.CountryCode).EqualTo(usPhoneNumberCountryCode)
	validator.String("number", &p.Number).Matches(usPhoneNumberRegex)
}

func RevisionStates() []string {
	return []string{
		StateDraft,
		StatePending,
		StateSubmitted,
	}
}

func Trainings() []string {
	return []string{
		TrainingInModule,
		TrainingInPerson,
	}
}

func TherapySettings() []string {
	return []string{
		TherapySettingInitial,
		TherapySettingTransferPumpSettings,
	}
}

func SexValues() []string {
	return []string{
		SexMale,
		SexFemale,
		SexUndisclosed,
	}
}

func AccountTypes() []string {
	return []string{
		AccountTypePatient,
		AccountTypeCaregiver,
	}
}
