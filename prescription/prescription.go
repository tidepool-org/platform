package prescription

import (
	"github.com/tidepool-org/platform/data/types/settings/pump"
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
	PrescriptionTherapySettingCPT                  = "cpt" // TODO: Certified Personal Trainer?

	PrescriptionLoopModeClosedLoop  = "closedLoop"
	PrescriptionLoopModeSuspendOnly = "suspendOnly"
)


type Prescription struct {
	ID                string           `json:"id" bson:"id"`
	AccessCode        string           `json:"accessCode,omitempty" bson:"accessCode,omitempty"`
	FirstName         *string          `json:"firstName,omitempty" bson:"firstName,omitempty"`
	LastName          *string          `json:"lastName,omitempty" bson:"firstName,omitempty"`
	Birthday          *string          `json:"birthday,omitempty" bson:"birthday,omitempty"`
	MRN               *string          `json:"mrn,omitempty" bson:"mrn,omitempty"`
	Email             *string          `json:"email,omitempty" bson:"email,omitempty"`
	Sex               *string          `json:"sex,omitempty" bson:"sex,omitempty"`
	Weight            *Weight          `json:"weight,omitempty" bson:"weight,omitempty"`
	YearOfDiagnosis   *string          `json:"yearOfDiagnosis,omitempty" bson:"yearOfDiagnosis,omitempty"`
	PhoneNumber       *string          `json:"phoneNumber,omitempty" bson:"phoneNumber,omitempty"`
	Address           *Address         `json:"address,omitempty" bson:"address,omitempty"`
	InitialSettings   *InitialSettings `json:"initialSettings,omitempty" bson:"initialSettings,omitempty"`
	Training          *string          `json:"training,omitempty" bson:"training,omitempty"`
	TherapySettings   *string          `json:"therapySettings,omitempty" bson:"therapySettings,omitempty"`
	LoopMode          *string          `json:"loopMode,omitempty" bson:"loopMode,omitempty"`
	AcknowledgedTerms *bool            `json:"acknowledgedTerms,omitempty" bson:"acknowledgedTerms,omitempty"`
	State             *string          `json:"state,omitempty" bson:"state,omitempty"`
	CreatedTime       *string          `json:"createdTime,omitempty" bson:"createdTime,omitempty"`
	CreatedUserID     *string          `json:"createdUserId,omitempty" bson:"createdUserId,omitempty"`
	ModifiedTime      *string          `json:"modifiedTime,omitempty" bson:"modifiedTime,omitempty"`
	ModifiedUserID    *string          `json:"modifiedUserId,omitempty" bson:"modifiedUserId,omitempty"`
	DeletedTime       *string          `json:"deletedTime,omitempty" bson:"deletedTime,omitempty"`
	DeletedUserID     *string          `json:"deletedUserId,omitempty" bson:"deletedUserId,omitempty"`
	Revision          *int             `json:"revision,omitempty" bson:"revision"`
}

type Weight struct {
	Value *float64 `json:"value,omitempty" bson:"value,omitempty"`
	Units *string  `json:"units,omitempty" bson:"units,omitempty"`
}

type Address struct {
	Line1      *string `json:"line1,omitempty" bson:"line1,omitempty"`
	Line2      *string `json:"line2,omitempty" bson:"line2,omitempty"`
	City       *string `json:"city,omitempty" bson:"city,omitempty"`
	State      *string `json:"state,omitempty" bson:"state,omitempty"`
	PostalCode *string `json:"postalCode,omitempty" bson:"postalCode,omitempty"`
}

type InitialSettings struct {
	BasalRateSchedule          *pump.BasalRateStartArray          `json:"basalSchedule,omitempty" bson:"basalSchedule,omitempty"`
	BloodGlucoseTargetSchedule *pump.BloodGlucoseTargetStartArray `json:"bgTarget,omitempty" bson:"bgTarget,omitempty"`
	CarbohydrateRatioSchedule  *pump.CarbohydrateRatioStartArray  `json:"carbRatio,omitempty" bson:"carbRatio,omitempty"`
	InsulinSensitivitySchedule *pump.InsulinSensitivityStartArray `json:"insulinSensitivity,omitempty" bson:"insulinSensitivity,omitempty"`
	BasalRateMaximum           *pump.BasalRateMaximum             `json:"basalRateMaximum,omitempty" bson:"basalRateMaximum,omitempty"`
	BolusAmountMaximum         *pump.BolusAmountMaximum           `json:"bolusAmountMaximum,omitempty" bson:"bolusAmountMaximum,omitempty"`
	PumpType                   *PumpType                          `json:"pumpType" bson:"pumpType"`
	CGMType                    *CGMType                           `json:"cgmType" bson:"cgmType"`
	// TODO: Add Suspend threshold - Dependent on latest data model changes
	// TODO: Add Insulin model - Dependent on latest data model changes
	// TODO: Add Bolus schedule? Does not exist in current pump settings model.
}

type PumpType struct {
	Manufacturers *[]string `json:"manufacturers,omitempty" bson:"manufacturers,omitempty"`
	Model         *string   `json:"model,omitempty" bson:"model,omitempty"`
	SerialNumber  *string   `json:"serialNumber,omitempty" bson:"serialNumber,omitempty"`
}

type CGMType struct {
	Manufacturers *[]string `json:"manufacturers,omitempty" bson:"manufacturers,omitempty"`
	Model         *string   `json:"model,omitempty" bson:"model,omitempty"`
	SerialNumber  *string   `json:"serialNumber,omitempty" bson:"serialNumber,omitempty"`
}

func PrescriptionStates() []string {
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

func PrescriptionTrainings() []string {
	return []string{
		PrescriptionTrainingInModule,
		PrescriptionTrainingInPerson,
	}
}

func PrescriptionTherapySettings() []string {
	return []string{
		PrescriptionTherapySettingInitial,
		PrescriptionTherapySettingTransferPumpSettings,
		PrescriptionTherapySettingCPT,
	}
}

func LoopModes() []string {
	return []string{
		PrescriptionLoopModeClosedLoop,
		PrescriptionLoopModeSuspendOnly,
	}
}
