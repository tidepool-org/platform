package prescription

import (
	"regexp"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/data/types/settings/pump"
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

	usPhoneNumberRegexString = "^\\d{10}|\\(\\d{3}\\) ?\\d{3}\\-\\d{4}$" // Matches 1234567890, (123)456-7890 or (123) 456-7890
	usPhoneNumberCountryCode = 1
)

var (
	usPhoneNumberRegex = regexp.MustCompile(usPhoneNumberRegexString)
)

type RevisionCreate struct {
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

type InitialSettings struct {
	BloodGlucoseUnits          string                             `json:"bloodGlucoseUnits,omitempty" bson:"bloodGlucoseUnits,omitempty"`
	BasalRateSchedule          *pump.BasalRateStartArray          `json:"basalRateSchedule,omitempty" bson:"basalRateSchedule,omitempty"`
	BloodGlucoseTargetSchedule *pump.BloodGlucoseTargetStartArray `json:"bloodGlucoseTargetSchedule,omitempty" bson:"bloodGlucoseTargetSchedule,omitempty"`
	CarbohydrateRatioSchedule  *pump.CarbohydrateRatioStartArray  `json:"carbohydrateRatioSchedule,omitempty" bson:"carbohydrateRatioSchedule,omitempty"`
	InsulinSensitivitySchedule *pump.InsulinSensitivityStartArray `json:"insulinSensitivitySchedule,omitempty" bson:"insulinSensitivitySchedule,omitempty"`
	BasalRateMaximum           *pump.BasalRateMaximum             `json:"basalRateMaximum,omitempty" bson:"basalRateMaximum,omitempty"`
	BolusAmountMaximum         *pump.BolusAmountMaximum           `json:"bolusAmountMaximum,omitempty" bson:"bolusAmountMaximum,omitempty"`
	PumpID                     *primitive.ObjectID                `json:"pumpId" bson:"pumpId"`
	CgmID                      *primitive.ObjectID                `json:"cgmId" bson:"cgmId"`
	// TODO: Add Suspend threshold - Dependent on latest data model changes
	// TODO: Add Insulin model - Dependent on latest data model changes
}

func (i *InitialSettings) Validate(validator structure.Validator) {
	validator.String("bloodGlucoseUnits", &i.BloodGlucoseUnits).OneOf(glucose.Units()...)
	if i.BasalRateSchedule != nil {
		i.BasalRateSchedule.Validate(validator.WithReference("basalSchedule"))
	}
	if i.BloodGlucoseTargetSchedule != nil {
		i.BloodGlucoseTargetSchedule.Validate(validator.WithReference("bloodGlucoseTargetSchedule"), &i.BloodGlucoseUnits)
	}
	if i.CarbohydrateRatioSchedule != nil {
		i.CarbohydrateRatioSchedule.Validate(validator.WithReference("carbohydrateRatioSchedule"))
	}
	if i.InsulinSensitivitySchedule != nil {
		i.InsulinSensitivitySchedule.Validate(validator.WithReference("insulinSensitivitySchedule"), &i.BloodGlucoseUnits)
	}
	if i.BasalRateMaximum != nil {
		i.BasalRateMaximum.Validate(validator.WithReference("basalRateMaximum"))
	}
	if i.BolusAmountMaximum != nil {
		i.BolusAmountMaximum.Validate(validator.WithReference("bolusAmountMaximum"))
	}
	if i.PumpID != nil {
		id := i.PumpID.Hex()
		validator.String("pumpId", &id).Hexadecimal().LengthEqualTo(24)
	}
	if i.CgmID != nil {
		id := i.CgmID.Hex()
		validator.String("cgmId", &id).Hexadecimal().LengthEqualTo(24)
	}
}

func (i *InitialSettings) ValidateAllRequired(validator structure.Validator) {
	if i.BasalRateSchedule == nil {
		validator.WithReference("basalSchedule").ReportError(structureValidator.ErrorValueEmpty())
	}
	if i.BloodGlucoseTargetSchedule == nil {
		validator.WithReference("bloodGlucoseTargetSchedule").ReportError(structureValidator.ErrorValueEmpty())
	}
	if i.CarbohydrateRatioSchedule == nil {
		validator.WithReference("carbohydrateRatioSchedule").ReportError(structureValidator.ErrorValueEmpty())
	}
	if i.InsulinSensitivitySchedule == nil {
		validator.WithReference("insulinSensitivitySchedule").ReportError(structureValidator.ErrorValueEmpty())
	}
	if i.BasalRateMaximum == nil {
		validator.WithReference("basalRateMaximum").ReportError(structureValidator.ErrorValueEmpty())
	}
	if i.BolusAmountMaximum == nil {
		validator.WithReference("bolusAmountMaximum").ReportError(structureValidator.ErrorValueEmpty())
	}
	if i.PumpID == nil {
		validator.WithReference("pumpId").ReportError(structureValidator.ErrorValueEmpty())
	}
	if i.CgmID == nil {
		validator.WithReference("cgmId").ReportError(structureValidator.ErrorValueEmpty())
	}
	// TODO: Validate Suspend Threshold and Insulin Type
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
