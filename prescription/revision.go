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

	UnitKg  = "kg"
	UnitLbs = "lbs"

	AccountTypePatient   = "patient"
	AccountTypeCaregiver = "caregiver"

	usPhoneNumberRegexString = "^\\d{10}|\\(\\d{3}\\) ?\\d{3}\\-\\d{4}$" // Matches 1234567890, (123)456-7890 or (123) 456-7890
	usPhoneNumberCountryCode = 1
)

var (
	usPhoneNumberRegex = regexp.MustCompile(usPhoneNumberRegexString)
)

type RevisionCreate struct {
	DataAttributes `json:",inline"`
	CreatedUserID  string `json:"createdUserId"`
	RevisionHash   string `json:"revisionHash"`
	ClinicID       string `json:"-"`
	ClinicianID    string `json:"-"`
	IsPrescriber   bool   `json:"-"`
}

func NewRevisionCreate(clinicID, clinicianID string, isPrescriber bool) *RevisionCreate {
	return &RevisionCreate{
		ClinicID:     clinicID,
		ClinicianID:  clinicianID,
		IsPrescriber: isPrescriber,
	}
}

func (r *RevisionCreate) Validate(validator structure.Validator) {
	integrityAttributes := NewIntegrityAttributesFromRevisionCreate(*r)
	integrityHash := MustGenerateIntegrityHash(integrityAttributes)
	validator.String("revisionHash", &r.RevisionHash).Exists().EqualTo(integrityHash.Hash)
	validator.String("createdUserId", &r.CreatedUserID).Exists().EqualTo(r.ClinicianID)
	validator.String("clinicianId", &r.ClinicianID).Exists().NotEmpty().Using(user.IDValidator)
	validator.String("clinicId", &r.ClinicID).Exists().NotEmpty()
	r.DataAttributes.Validate(validator)
}

func (r *RevisionCreate) IsClinicianAuthorized() bool {
	if r.DataAttributes.State == StateSubmitted {
		// Only prescribers are authorized to put prescriptions in submitted state
		return r.IsPrescriber
	}
	return true
}

type Revision struct {
	RevisionID    int            `json:"revisionId" bson:"revisionId"`
	IntegrityHash *IntegrityHash `json:"integrityHash" bson:"integrityHash"`
	Attributes    *Attributes    `json:"attributes" bson:"attributes"`
}

type Revisions []*Revision

func NewRevision(revisionID int, create *RevisionCreate) *Revision {
	now := time.Now()
	return &Revision{
		RevisionID: revisionID,
		// Trust the integrity hash that's sent by the frontend, because it's already validated
		// in RevisionCreate validation. Just set the algorithm field for completeness.
		IntegrityHash: &IntegrityHash{
			Algorithm: algorithmJCSSha512,
			Hash:      create.RevisionHash,
		},
		Attributes: &Attributes{
			DataAttributes: DataAttributes{
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
				Calculator:              create.Calculator,
				Training:                create.Training,
				TherapySettings:         create.TherapySettings,
				PrescriberTermsAccepted: create.PrescriberTermsAccepted,
				State:                   create.State,
			},
			CreationAttributes: CreationAttributes{
				CreatedTime:   now,
				CreatedUserID: create.ClinicianID,
			},
		},
	}
}

func (r *Revision) Validate(validator structure.Validator) {
	validator.Int("revisionId", &r.RevisionID).GreaterThanOrEqualTo(0)
	integrityHashValidator := validator.WithReference("integrityHash")
	if r.IntegrityHash != nil {
		integrityHashValidator.String("hash", &r.IntegrityHash.Hash).Exists().NotEmpty()
		integrityHashValidator.String("algorithm", &r.IntegrityHash.Algorithm).Exists().EqualTo(algorithmJCSSha512)
	} else {
		integrityHashValidator.ReportError(structureValidator.ErrorValueEmpty())
	}
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

func (r *Revision) GetSubmittedTime() *time.Time {
	if r.Attributes.State != StateSubmitted {
		return nil
	}

	now := time.Now()
	return &now
}

type Attributes struct {
	DataAttributes     `json:",inline" bson:",inline"`
	CreationAttributes `json:",inline" bson:",inline"`
}

func (a *Attributes) Validate(validator structure.Validator) {
	a.DataAttributes.Validate(validator)
	a.CreationAttributes.Validate(validator)
}

type DataAttributes struct {
	AccountType             *string          `json:"accountType,omitempty" bson:"accountType"`
	CaregiverFirstName      *string          `json:"caregiverFirstName,omitempty" bson:"caregiverFirstName"`
	CaregiverLastName       *string          `json:"caregiverLastName,omitempty" bson:"caregiverLastName"`
	FirstName               *string          `json:"firstName,omitempty" bson:"firstName"`
	LastName                *string          `json:"lastName,omitempty" bson:"lastName"`
	Birthday                *string          `json:"birthday,omitempty" bson:"birthday"`
	MRN                     *string          `json:"mrn,omitempty" bson:"mrn"`
	Email                   *string          `json:"email,omitempty" bson:"email"`
	Sex                     *string          `json:"sex,omitempty" bson:"sex"`
	Weight                  *Weight          `json:"weight,omitempty" bson:"weight"`
	YearOfDiagnosis         *int             `json:"yearOfDiagnosis,omitempty" bson:"yearOfDiagnosis"`
	PhoneNumber             *PhoneNumber     `json:"phoneNumber,omitempty" bson:"phoneNumber"`
	InitialSettings         *InitialSettings `json:"initialSettings,omitempty" bson:"initialSettings"`
	Calculator              *Calculator      `json:"calculator,omitempty" bson:"calculator"`
	Training                *string          `json:"training,omitempty" bson:"training"`
	TherapySettings         *string          `json:"therapySettings,omitempty" bson:"therapySettings"`
	PrescriberTermsAccepted *bool            `json:"prescriberTermsAccepted,omitempty" bson:"prescriberTermsAccepted"`
	State                   string           `json:"state" bson:"state"`
}

func (d *DataAttributes) Validate(validator structure.Validator) {
	validator.String("accountType", d.AccountType).OneOf(AccountTypes()...)
	if d.AccountType != nil {
		if *d.AccountType == AccountTypePatient {
			validator.String("caregiverFirstName", d.CaregiverFirstName).Empty()
			validator.String("caregiverLastName", d.CaregiverLastName).Empty()
		}
	}
	validator.String("firstName", d.FirstName).NotEmpty()
	validator.String("lastName", d.LastName).NotEmpty()
	validator.String("birthday", d.Birthday).AsTime("2006-01-02").NotZero().BeforeNow(time.Second)
	validator.String("email", d.Email).Email()
	validator.String("sex", d.Sex).OneOf(SexValues()...)
	validator.Int("yearOfDiagnosis", d.YearOfDiagnosis).GreaterThan(1900)
	validator.String("training", d.Training).OneOf(Trainings()...)
	validator.String("therapySettings", d.TherapySettings).OneOf(TherapySettings()...)
	if d.PhoneNumber != nil {
		d.PhoneNumber.Validate(validator.WithReference("phoneNumber"))
	}
	if d.Weight != nil {
		d.Weight.Validate(validator.WithReference("weight"))
	}
	if d.InitialSettings != nil {
		d.InitialSettings.Validate(validator.WithReference("initialSettings"))
	}
	if d.Calculator != nil {
		d.Calculator.Validate(validator.WithReference("calculator"))
	}

	validator.String("state", &d.State).Exists().OneOf(RevisionStates()...)
	if d.State == StateSubmitted {
		d.ValidateSubmittedPrescription(validator)
	}
}

func (d *DataAttributes) ValidateSubmittedPrescription(validator structure.Validator) {
	validator.String("accountType", d.AccountType).Exists()
	if d.AccountType != nil {
		if *d.AccountType == AccountTypeCaregiver {
			validator.String("caregiverFirstName", d.CaregiverFirstName).NotEmpty().Exists()
			validator.String("caregiverLastName", d.CaregiverLastName).NotEmpty().Exists()
		}
	}
	validator.String("firstName", d.FirstName).Exists()
	validator.String("lastName", d.LastName).Exists()
	validator.String("birthday", d.Birthday).Exists()
	validator.String("email", d.Email).Exists()
	validator.String("sex", d.Sex).Exists()
	validator.String("therapySettings", d.TherapySettings).Exists()
	validator.Bool("prescriberTermsAccepted", d.PrescriberTermsAccepted).Exists().True()

	weightValidator := validator.WithReference("weight")
	if d.Weight != nil {
		d.Weight.ValidateSubmittedPrescription(weightValidator)
	}

	initialSettingsValidator := validator.WithReference("initialSettings")
	if d.InitialSettings != nil {
		d.InitialSettings.ValidateSubmittedPrescription(initialSettingsValidator)
	} else {
		initialSettingsValidator.ReportError(structureValidator.ErrorValueEmpty())
	}
}

type CreationAttributes struct {
	CreatedTime   time.Time `json:"createdTime,omitempty" bson:"createdTime"`
	CreatedUserID string    `json:"createdUserId,omitempty" bson:"createdUserId"`
}

func (c *CreationAttributes) Validate(validator structure.Validator) {
	validator.Time("createdTime", &c.CreatedTime).NotZero().BeforeNow(time.Second)
	validator.String("createdUserId", &c.CreatedUserID).Using(user.IDValidator)
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

func (w *Weight) ValidateSubmittedPrescription(validator structure.Validator) {
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
