package prescription

import (
	"time"

	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/validate"

	"github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/device"
)

const (
	PrescriptionStateDraft     = "draft"
	PrescriptionStatePending   = "pending"
	PrescriptionStateSubmitted = "submitted"
	PrescriptionStateReviewed  = "reviewed"
	PrescriptionStateExpired   = "expired"
	PrescriptionStateActive    = "active"
	PrescriptionStateInactive  = "inactive"

	PrescriptionTrainingInPerson = "inPerson"
	PrescriptionTrainingInModule = "inModule"

	PrescriptionTherapySettingInitial              = "initial"
	PrescriptionTherapySettingTransferPumpSettings = "transferPumpSettings"
	PrescriptionTherapySettingCertifiedPumpTrainer = "certifiedPumpTrainer"

	PrescriptionLoopModeClosedLoop  = "closedLoop"
	PrescriptionLoopModeSuspendOnly = "suspendOnly"
)

type Prescription struct {
	ID              string     `json:"id" bson:"id" validate:"required"`
	PatientID       *string    `json:"patientId,omitempty" bson:"patientId,omitempty"`
	AccessCode      *string    `json:"accessCode,omitempty" bson:"-" validate:"alphanum,len=6,omitempty"`
	AccessCodeHash  string     `json:"-" bson:"accessCodeHash" validate:"hexadecimal,len=40"`
	State           string     `json:"state" bson:"state" validate:"oneof=draft pending submitted reviewed expired active inactive"`
	LatestRevision  *Revision  `json:"latestRevision,omitempty" bson:"latestRevision,omitempty" validate:"-"`
	RevisionHistory Revisions  `json:"-,omitempty" bson:"revisionHistory,omitempty" validate:"-"`
	ExpirationTime  time.Time  `json:"expirationTime" bson:"expirationTime" validate:"required"`
	CreatedTime     time.Time  `json:"createdTime" bson:"createdTime" validate:"required"`
	CreatedUserID   string     `json:"createdUserId" bson:"createdUserId" validate:"required"`
	DeletedTime     *time.Time `json:"deletedTime,omitempty" bson:"deletedTime,omitempty"`
	DeletedUserID   *string    `json:"deletedUserId,omitempty" bson:"deletedUserId,omitempty"`
}

type Prescriptions []*Prescription

func (p *Prescription) Validate(validator structure.Validator) {
	validate.StructWithLegacyErrorReporting(p, validator)
}

type Revision struct {
	RevisionID      int    `json:"-" bson:"revisionId"`
	SignatureUserID string `json:"signatureUserId" bson:"signatureUserId"`
	SignatureKeyID  string `json:"signatureKeyId" bson:"signatureKeyId"`
	Signature       string `json:"signature" bson:"signature"`
	Attributes      Attributes
}

type Revisions []*Revision

type Attributes struct {
	FirstName               *string          `json:"firstName,omitempty" bson:"firstName,omitempty"`
	LastName                *string          `json:"lastName,omitempty" bson:"lastName,omitempty"`
	Birthday                *string          `json:"birthday,omitempty" bson:"birthday,omitempty"`
	MRN                     *string          `json:"mrn,omitempty" bson:"mrn,omitempty"`
	Email                   *string          `json:"email,omitempty" bson:"email,omitempty"`
	Sex                     *string          `json:"sex,omitempty" bson:"sex,omitempty"`
	Weight                  *Weight          `json:"weight,omitempty" bson:"weight,omitempty"`
	YearOfDiagnosis         *string          `json:"yearOfDiagnosis,omitempty" bson:"yearOfDiagnosis,omitempty"`
	PhoneNumber             *string          `json:"phoneNumber,omitempty" bson:"phoneNumber,omitempty"`
	Address                 *Address         `json:"address,omitempty" bson:"address,omitempty"`
	InitialSettings         *InitialSettings `json:"initialSettings,omitempty" bson:"initialSettings,omitempty"`
	Training                *string          `json:"training,omitempty" bson:"training,omitempty"`
	TherapySettings         *string          `json:"therapySettings,omitempty" bson:"therapySettings,omitempty"`
	LoopMode                *string          `json:"loopMode,omitempty" bson:"loopMode,omitempty"`
	PrescriberTermsAccepted *bool            `json:"prescriberTermsAccepted,omitempty" bson:"prescriberTermsAccepted,omitempty"`
	State                   *string          `json:"state,omitempty" bson:"state,omitempty"`
	ModifiedTime            *time.Time       `json:"modifiedTime,omitempty" bson:"modifiedTime,omitempty"`
	ModifiedUserID          *string          `json:"modifiedUserId,omitempty" bson:"modifiedUserId,omitempty"`
}

type Weight struct {
	Value *float64 `json:"value,omitempty" bson:"value,omitempty"`
	Units string   `json:"units,omitempty" bson:"units,omitempty"`
}

type Address struct {
	Line1      *string `json:"line1,omitempty" bson:"line1,omitempty"`
	Line2      *string `json:"line2,omitempty" bson:"line2,omitempty"`
	City       *string `json:"city,omitempty" bson:"city,omitempty"`
	State      *string `json:"state,omitempty" bson:"state,omitempty"`
	PostalCode *string `json:"postalCode,omitempty" bson:"postalCode,omitempty"`
	Country    *string `json:"country,omitempty" bson:"country,omitempty"`
}

type InitialSettings struct {
	BasalRateSchedule          *pump.BasalRateStartArray          `json:"basalSchedule,omitempty" bson:"basalSchedule,omitempty"`
	BloodGlucoseTargetSchedule *pump.BloodGlucoseTargetStartArray `json:"bgTarget,omitempty" bson:"bgTarget,omitempty"`
	CarbohydrateRatioSchedule  *pump.CarbohydrateRatioStartArray  `json:"carbRatio,omitempty" bson:"carbRatio,omitempty"`
	InsulinSensitivitySchedule *pump.InsulinSensitivityStartArray `json:"insulinSensitivity,omitempty" bson:"insulinSensitivity,omitempty"`
	BasalRateMaximum           *pump.BasalRateMaximum             `json:"basalRateMaximum,omitempty" bson:"basalRateMaximum,omitempty"`
	BolusAmountMaximum         *pump.BolusAmountMaximum           `json:"bolusAmountMaximum,omitempty" bson:"bolusAmountMaximum,omitempty"`
	PumpType                   *device.Device                     `json:"pumpType" bson:"pumpType"`
	CGMType                    *device.Device                     `json:"cgmType" bson:"cgmType"`
	// TODO: Add Suspend threshold - Dependent on latest data model changes
	// TODO: Add Insulin model - Dependent on latest data model changes
	// TODO: Add Bolus schedule? Does not exist in current pump settings model.
}

func States() []string {
	return []string{
		PrescriptionStateDraft,
		PrescriptionStatePending,
		PrescriptionStateSubmitted,
		PrescriptionStateReviewed,
		PrescriptionStateExpired,
		PrescriptionStateActive,
		PrescriptionStateInactive,
	}
}

func Trainings() []string {
	return []string{
		PrescriptionTrainingInModule,
		PrescriptionTrainingInPerson,
	}
}

func TherapySettings() []string {
	return []string{
		PrescriptionTherapySettingInitial,
		PrescriptionTherapySettingTransferPumpSettings,
		PrescriptionTherapySettingCertifiedPumpTrainer,
	}
}

func LoopModes() []string {
	return []string{
		PrescriptionLoopModeClosedLoop,
		PrescriptionLoopModeSuspendOnly,
	}
}
