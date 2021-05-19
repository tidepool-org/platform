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
}

func NewRevisionCreate() *RevisionCreate {
	return &RevisionCreate{}
}

func (r *RevisionCreate) Validate(validator structure.Validator) {
	r.DataAttributes.Validate(validator)
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
				CreatedUserID: userID,
			},
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
	DataAttributes     `json:",inline" bson:",inline"`
	CreationAttributes `json:",inline" bson:",inline"`
}

func (a *Attributes) Validate(validator structure.Validator) {
	a.DataAttributes.Validate(validator)
	a.CreationAttributes.Validate(validator)
}

type DataAttributes struct {
	AccountType             string           `json:"accountType,omitempty" bson:"accountType"`
	CaregiverFirstName      string           `json:"caregiverFirstName,omitempty" bson:"caregiverFirstName"`
	CaregiverLastName       string           `json:"caregiverLastName,omitempty" bson:"caregiverLastName"`
	FirstName               string           `json:"firstName,omitempty" bson:"firstName"`
	LastName                string           `json:"lastName,omitempty" bson:"lastName"`
	Birthday                string           `json:"birthday,omitempty" bson:"birthday"`
	MRN                     string           `json:"mrn,omitempty" bson:"mrn"`
	Email                   string           `json:"email,omitempty" bson:"email"`
	Sex                     string           `json:"sex,omitempty" bson:"sex"`
	Weight                  *Weight          `json:"weight,omitempty" bson:"weight"`
	YearOfDiagnosis         int              `json:"yearOfDiagnosis,omitempty" bson:"yearOfDiagnosis"`
	PhoneNumber             *PhoneNumber     `json:"phoneNumber,omitempty" bson:"phoneNumber"`
	InitialSettings         *InitialSettings `json:"initialSettings,omitempty" bson:"initialSettings"`
	Calculator              *Calculator      `json:"calculator,omitempty" bson:"calculator"`
	Training                string           `json:"training,omitempty" bson:"training"`
	TherapySettings         string           `json:"therapySettings,omitempty" bson:"therapySettings"`
	PrescriberTermsAccepted bool             `json:"prescriberTermsAccepted,omitempty" bson:"prescriberTermsAccepted"`
	State                   string           `json:"state" bson:"state"`
}

func (d *DataAttributes) Validate(validator structure.Validator) {
	if d.AccountType != "" {
		validator.String("accountType", &d.AccountType).OneOf(AccountTypes()...)
		if d.AccountType == AccountTypePatient {
			validator.String("caregiverFirstName", &d.CaregiverFirstName).Empty()
			validator.String("caregiverLastName", &d.CaregiverLastName).Empty()
		}
	}
	if d.Birthday != "" {
		validator.String("birthday", &d.Birthday).AsTime("2006-01-02").NotZero().BeforeNow(time.Second)
	}
	if d.Email != "" {
		validator.String("email", &d.Email).Email()
	}
	if d.Sex != "" {
		validator.String("sex", &d.Sex).OneOf(SexValues()...)
	}
	if d.YearOfDiagnosis != 0 {
		validator.Int("yearOfDiagnosis", &d.YearOfDiagnosis).GreaterThan(1900)
	}
	if d.PhoneNumber != nil {
		d.PhoneNumber.Validate(validator.WithReference("phoneNumber"))
	}
	if d.Training != "" {
		validator.String("training", &d.Training).OneOf(Trainings()...)
	}
	if d.TherapySettings != "" {
		validator.String("therapySettings", &d.TherapySettings).OneOf(TherapySettings()...)
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
	validator.String("state", &d.State).OneOf(RevisionStates()...)

	if d.State == StateSubmitted {
		d.ValidateSubmittedPrescription(validator)
	}
}

func (d *DataAttributes) ValidateSubmittedPrescription(validator structure.Validator) {
	validator.String("accountType", &d.AccountType).NotEmpty()
	if d.AccountType == AccountTypeCaregiver {
		validator.String("caregiverFirstName", &d.CaregiverFirstName).NotEmpty()
		validator.String("caregiverLastName", &d.CaregiverLastName).NotEmpty()
	}
	validator.String("firstName", &d.FirstName).NotEmpty()
	validator.String("lastName", &d.LastName).NotEmpty()
	validator.String("birthday", &d.Birthday).NotEmpty()
	validator.String("email", &d.Email).NotEmpty()
	validator.String("sex", &d.Sex).NotEmpty()
	validator.Int("yearOfDiagnosis", &d.YearOfDiagnosis).GreaterThan(1900)
	validator.String("training", &d.Training).NotEmpty()
	validator.String("therapySettings", &d.TherapySettings).NotEmpty()
	validator.Bool("prescriberTermsAccepted", &d.PrescriberTermsAccepted).True()

	// if phoneNumber is nil validate will fail
	phoneValidator := validator.WithReference("phoneNumber")
	if d.PhoneNumber != nil {
		d.PhoneNumber.Validate(phoneValidator)
	} else {
		phoneValidator.ReportError(structureValidator.ErrorValueEmpty())
	}

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
	CreatedUserID string    `json:"createdUserId,omitempty" bson:"cratedUserId"`
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
