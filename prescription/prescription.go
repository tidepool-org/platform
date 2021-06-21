package prescription

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/tidepool-org/platform/pointer"

	"github.com/tidepool-org/platform/errors"

	"github.com/tidepool-org/platform/page"

	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/user"

	"github.com/tidepool-org/platform/structure"
)

const (
	StateDraft     = "draft"
	StatePending   = "pending"
	StateSubmitted = "submitted"
	StateClaimed   = "claimed"
	StateExpired   = "expired"
	StateActive    = "active"
	StateInactive  = "inactive"

	MaximumExpirationTime = time.Hour * 24 * 90 // 90 days
)

type Service interface {
	Accessor
}

type Accessor interface {
	CreatePrescription(ctx context.Context, create *RevisionCreate) (*Prescription, error)
	ListPrescriptions(ctx context.Context, filter *Filter, pagination *page.Pagination) (Prescriptions, error)
	DeletePrescription(ctx context.Context, clinicID, prescriptionID, clinicianID string) (bool, error)
	AddRevision(ctx context.Context, prescriptionID string, create *RevisionCreate) (*Prescription, error)
	ClaimPrescription(ctx context.Context, claim *Claim) (*Prescription, error)
	UpdatePrescriptionState(ctx context.Context, prescriptionID string, update *StateUpdate) (*Prescription, error)
}

type Prescription struct {
	ID               primitive.ObjectID `json:"id" bson:"_id"`
	PatientUserID    string             `json:"patientUserId,omitempty" bson:"patientUserId,omitempty"`
	AccessCode       string             `json:"accessCode,omitempty" bson:"accessCode"`
	State            string             `json:"state" bson:"state"`
	LatestRevision   *Revision          `json:"latestRevision" bson:"latestRevision"`
	RevisionHistory  Revisions          `json:"-" bson:"revisionHistory"`
	ExpirationTime   *time.Time         `json:"expirationTime" bson:"expirationTime"`
	PrescriberUserID string             `json:"prescriberUserId,omitempty" bson:"prescriberUserId,omitempty"`
	ClinicID         string             `json:"clinicId" bson:"clinicId"`
	CreatedTime      time.Time          `json:"createdTime" bson:"createdTime"`
	CreatedUserID    string             `json:"createdUserId" bson:"createdUserId"`
	DeletedTime      *time.Time         `json:"deletedTime,omitempty" bson:"deletedTime,omitempty"`
	DeletedUserID    string             `json:"deletedUserId,omitempty" bson:"deletedUserId,omitempty"`
	ModifiedTime     time.Time          `json:"modifiedTime,omitempty" bson:"modifiedTime,omitempty"`
	ModifiedUserID   string             `json:"modifiedUserId" bson:"modifiedUserId"`
	SubmittedTime    *time.Time         `json:"submittedTime,omitempty" bson:"submittedTime,omitempty"`
}

func NewPrescription(revisionCreate *RevisionCreate) *Prescription {
	now := time.Now()
	accessCode := GenerateAccessCode()
	revision := NewRevision(0, revisionCreate)
	revisionHistory := []*Revision{revision}
	prescription := &Prescription{
		ID:               primitive.NewObjectID(),
		AccessCode:       accessCode,
		State:            revisionCreate.State,
		LatestRevision:   revision,
		RevisionHistory:  revisionHistory,
		ExpirationTime:   revision.CalculateExpirationTime(),
		CreatedTime:      now,
		CreatedUserID:    revisionCreate.ClinicianID,
		ClinicID:         revisionCreate.ClinicID,
		PrescriberUserID: revision.GetPrescriberUserID(),
		ModifiedTime:     now,
		ModifiedUserID:   revisionCreate.ClinicianID,
		SubmittedTime:    revision.GetSubmittedTime(),
	}

	return prescription
}

type Prescriptions []*Prescription

func (p *Prescription) Validate(validator structure.Validator) {
	id := p.ID.Hex()
	validator.String("id", &id).Hexadecimal().LengthEqualTo(24)

	if p.PatientUserID != "" {
		validator.String("patientUserId", &p.PatientUserID).Using(user.IDValidator)
	}

	validator.String("accessCode", &p.AccessCode).LengthEqualTo(6).Alphanumeric()

	validator.String("state", &p.State).OneOf(States()...)

	if p.LatestRevision != nil {
		p.LatestRevision.Validate(validator.WithReference("latestRevision"))
	} else {
		validator.WithReference("latestRevision").ReportError(structureValidator.ErrorValueEmpty())
	}

	if p.ExpirationTime != nil {
		validator.Time("expirationTime", p.ExpirationTime).NotZero()
	}

	if p.PrescriberUserID != "" {
		validator.String("prescriberId", &p.PrescriberUserID).Using(user.IDValidator)
	}

	validator.Time("createdTime", &p.CreatedTime).NotZero()
	validator.String("createdUserId", &p.CreatedUserID).NotEmpty().Using(user.IDValidator)

	if p.DeletedTime != nil {
		validator.Time("deletedTime", p.DeletedTime).NotZero()
	}
	if p.DeletedUserID != "" {
		validator.String("deletedUserId", &p.DeletedUserID).Using(user.IDValidator)
	}

	validator.Time("modifiedTime", &p.ModifiedTime).NotZero()
	validator.String("modifiedUserId", &p.ModifiedUserID).NotEmpty().Using(user.IDValidator)
}

func States() []string {
	return []string{
		StateDraft,
		StatePending,
		StateSubmitted,
		StateClaimed,
		StateExpired,
		StateActive,
		StateInactive,
	}
}

func StatesVisibleToPatients() []string {
	return []string{
		StateSubmitted,
		StateClaimed,
		StateActive,
		StateInactive,
	}
}

func validPatientStateTransitions() map[string][]string {
	return map[string][]string{
		StateSubmitted: {StateClaimed},
		StateClaimed:   {StateActive},
	}
}

func validClinicianStateTransitions() map[string][]string {
	return map[string][]string{
		StateDraft:   {StateDraft, StatePending, StateSubmitted},
		StatePending: {StatePending, StateSubmitted},
	}
}

func stateTransitionsForUser(usr *user.User) map[string][]string {
	if usr.HasRole(user.RoleClinic) {
		return validClinicianStateTransitions()
	}

	return validPatientStateTransitions()
}

func ValidStateTransitions(transitions map[string][]string, state string) []string {
	valid, ok := transitions[state]
	if !ok {
		return []string{}
	}

	return valid
}

type Filter struct {
	ClinicID       string
	PatientUserID  string
	PatientEmail   string
	State          string
	ID             string
	CreatedAfter   *time.Time
	CreatedBefore  *time.Time
	ModifiedAfter  *time.Time
	ModifiedBefore *time.Time
}

func NewClinicFilter(clinicID string) (*Filter, error) {
	if clinicID == "" {
		return nil, errors.New("clinic id is missing")
	}

	return &Filter{
		ClinicID: clinicID,
	}, nil
}

func NewPatientFilter(userID string) (*Filter, error) {
	if userID == "" {
		return nil, errors.New("user id is missing")
	}

	return &Filter{
		PatientUserID: userID,
	}, nil
}

func (f *Filter) Validate(validator structure.Validator) {
	if f.ID != "" {
		validator.String("id", &f.ID).Hexadecimal().LengthEqualTo(24)
	}
	if f.ClinicID != "" {
		if f.State != "" {
			validator.String("state", &f.State).OneOf(States()...)
		}
		if f.PatientUserID != "" {
			validator.String("patientUserId", &f.PatientUserID).Using(user.IDValidator)
		}
		if f.PatientEmail != "" {
			validator.String("patientEmail", &f.PatientEmail).Email()
		}
	} else {
		validator.String("patientUserId", &f.PatientUserID).Using(user.IDValidator)
		validator.String("patientEmail", &f.PatientEmail).Empty()
		if f.State != "" {
			validator.String("state", &f.State).OneOf(StatesVisibleToPatients()...)
		}
	}
}

func (f *Filter) Parse(parser structure.ObjectParser) {
	if ptr := parser.String("id"); ptr != nil {
		f.ID = *ptr
	}
	if f.ClinicID != "" {
		if ptr := parser.String("patientUserId"); ptr != nil {
			f.PatientUserID = *ptr
		}
		if ptr := parser.String("patientEmail"); ptr != nil {
			f.PatientEmail = *ptr
		}
	}
	if ptr := parser.String("state"); ptr != nil {
		f.State = *ptr
	}
	if ptr := parser.Time("createdAfter", time.RFC3339Nano); ptr != nil {
		f.CreatedAfter = ptr
	}
	if ptr := parser.Time("createdBefore", time.RFC3339Nano); ptr != nil {
		f.CreatedBefore = ptr
	}
	if ptr := parser.Time("modifiedAfter", time.RFC3339Nano); ptr != nil {
		f.ModifiedAfter = ptr
	}
	if ptr := parser.Time("modifiedBefore", time.RFC3339Nano); ptr != nil {
		f.ModifiedBefore = ptr
	}
}

type Update struct {
	prescription     *Prescription
	Revision         *Revision
	State            string
	PrescriberUserID string
	PatientUserID    string
	ExpirationTime   *time.Time
	ModifiedTime     time.Time
	ModifiedUserID   string
	SubmittedTime    *time.Time
}

func NewPrescriptionAddRevisionUpdate(prescription *Prescription, create *RevisionCreate) *Update {
	revisionID := prescription.LatestRevision.RevisionID + 1
	revision := NewRevision(revisionID, create)
	update := &Update{
		prescription:     prescription,
		Revision:         revision,
		State:            create.State,
		PrescriberUserID: revision.GetPrescriberUserID(),
		ExpirationTime:   revision.CalculateExpirationTime(),
		ModifiedUserID:   create.ClinicianID,
		ModifiedTime:     revision.Attributes.CreatedTime,
		SubmittedTime:    revision.GetSubmittedTime(),
	}

	return update
}

func NewPrescriptionClaimUpdate(patientID string, prescription *Prescription) *Update {
	return &Update{
		prescription:   prescription,
		State:          StateClaimed,
		PatientUserID:  patientID,
		ModifiedUserID: patientID,
		ModifiedTime:   time.Now(),
	}
}

func NewPrescriptionStateUpdate(prescription *Prescription, update *StateUpdate) *Update {
	return &Update{
		prescription:   prescription,
		State:          update.State,
		ModifiedUserID: update.PatientID,
		PatientUserID:  update.PatientID,
		ModifiedTime:   time.Now(),
	}
}

func (u *Update) GetUpdatedAccessCode() *string {
	if u.State != StateClaimed {
		return nil
	}

	// Remove the access code when the user claims the prescription
	return pointer.FromString("")
}

func (u *Update) GetPrescriptionID() primitive.ObjectID {
	return u.prescription.ID
}

func (u *Update) Validate(validator structure.Validator) {
	if u.prescription == nil {
		validator.WithReference("prescription").ReportError(structureValidator.ErrorValueEmpty())
		return
	}
	if u.PatientUserID != "" {
		u.validateForPatient(validator)
	} else {
		u.validateForClinician(validator)
	}
}

func (u *Update) validateForClinician(validator structure.Validator) {
	stateTransitions := ValidStateTransitions(validClinicianStateTransitions(), u.prescription.State)
	validator.String("state", &u.State).OneOf(stateTransitions...)
	validator.String("patientUserId", &u.PatientUserID).Empty()

	if u.Revision != nil {
		u.Revision.Validate(validator.WithReference("revision"))
	} else {
		validator.WithReference("revision").ReportError(structureValidator.ErrorValueEmpty())
	}
}

func (u *Update) validateForPatient(validator structure.Validator) {
	stateTransitions := ValidStateTransitions(validPatientStateTransitions(), u.prescription.State)
	validator.String("state", &u.State).OneOf(stateTransitions...)
	validator.String("prescriberUserId", &u.PrescriberUserID).Empty()

	if u.Revision != nil {
		validator.WithReference("revision").ReportError(structureValidator.ErrorValueExists())
	}
	if u.PatientUserID != "" {
		validator.String("patientUserId", &u.PatientUserID).Using(user.IDValidator)
	}
}

type Claim struct {
	PatientID  string `json:"-"`
	AccessCode string `json:"accessCode"`
	Birthday   string `json:"birthday"`
}

func NewPrescriptionClaim(patientID string) *Claim {
	return &Claim{
		PatientID: patientID,
	}
}

func (p *Claim) Validate(validator structure.Validator) {
	validator.String("patientId", &p.PatientID).Exists().NotEmpty().Using(user.IDValidator)
	validator.String("accessCode", &p.AccessCode).NotEmpty()
	validator.String("birthday", &p.Birthday).NotEmpty().AsTime("2006-01-02")
}

type StateUpdate struct {
	PatientID string `json:"-"`
	State     string `json:"state"`
}

func NewStateUpdate(patientID string) *StateUpdate {
	return &StateUpdate{
		PatientID: patientID,
	}
}

func (s *StateUpdate) Validate(validator structure.Validator) {
	validator.String("patientId", &s.PatientID).Exists().NotEmpty().Using(user.IDValidator)
	validator.String("status", &s.State).OneOf(StatesVisibleToPatients()...)
}
