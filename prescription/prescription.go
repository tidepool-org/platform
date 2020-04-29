package prescription

import (
	"context"
	"time"

	"github.com/tidepool-org/platform/pointer"

	"github.com/tidepool-org/platform/errors"

	"github.com/globalsign/mgo/bson"

	"github.com/tidepool-org/platform/page"

	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/user"

	"github.com/tidepool-org/platform/structure"
)

const (
	StateDraft     = "draft"
	StatePending   = "pending"
	StateSubmitted = "submitted"
	StateReviewed  = "reviewed"
	StateExpired   = "expired"
	StateActive    = "active"
	StateInactive  = "inactive"

	MaximumExpirationTime = time.Hour * 24 * 30 // 30 days
)

type Service interface {
	Accessor
}

type Accessor interface {
	CreatePrescription(ctx context.Context, userID string, create *RevisionCreate) (*Prescription, error)
	ListPrescriptions(ctx context.Context, filter *Filter, pagination *page.Pagination) (Prescriptions, error)
	DeletePrescription(ctx context.Context, clinicianID string, id string) (bool, error)
	AddRevision(ctx context.Context, usr *user.User, id string, create *RevisionCreate) (*Prescription, error)
	ClaimPrescription(ctx context.Context, usr *user.User, claim *Claim) (*Prescription, error)
	UpdatePrescriptionState(ctx context.Context, usr *user.User, id string, update *StateUpdate) (*Prescription, error)
}

type Prescription struct {
	ID               bson.ObjectId `json:"id" bson:"_id"`
	PatientID        string        `json:"patientId,omitempty" bson:"patientId,omitempty"`
	AccessCode       string        `json:"accessCode,omitempty" bson:"accessCode"`
	State            string        `json:"state" bson:"state"`
	LatestRevision   *Revision     `json:"latestRevision" bson:"latestRevision"`
	RevisionHistory  Revisions     `json:"-" bson:"revisionHistory"`
	ExpirationTime   *time.Time    `json:"expirationTime" bson:"expirationTime"`
	PrescriberUserID string        `json:"prescriberUserId,omitempty" bson:"prescriberUserId,omitempty"`
	CreatedTime      time.Time     `json:"createdTime" bson:"createdTime"`
	CreatedUserID    string        `json:"createdUserId" bson:"createdUserId"`
	DeletedTime      *time.Time    `json:"deletedTime,omitempty" bson:"deletedTime,omitempty"`
	DeletedUserID    string        `json:"deletedUserId,omitempty" bson:"deletedUserId,omitempty"`
	ModifiedTime     time.Time     `json:"modifiedTime,omitempty" bson:"modifiedTime,omitempty"`
	ModifiedUserID   string        `json:"modifiedUserId" bson:"modifiedUserId"`
}

func NewPrescription(userID string, revisionCreate *RevisionCreate) *Prescription {
	now := time.Now()
	accessCode := GenerateAccessCode()
	revision := NewRevision(userID, 0, revisionCreate)
	revisionHistory := []*Revision{revision}
	prescription := &Prescription{
		ID:               bson.NewObjectId(),
		AccessCode:       accessCode,
		State:            revisionCreate.State,
		LatestRevision:   revision,
		RevisionHistory:  revisionHistory,
		ExpirationTime:   revision.CalculateExpirationTime(),
		CreatedTime:      now,
		CreatedUserID:    userID,
		PrescriberUserID: revision.GetPrescriberUserID(),
		ModifiedTime:     now,
		ModifiedUserID:   userID,
	}

	return prescription
}

type Prescriptions []*Prescription

func (p *Prescription) Validate(validator structure.Validator) {
	id := p.ID.Hex()
	validator.String("id", &id).Hexadecimal().LengthEqualTo(24)

	if p.PatientID != "" {
		validator.String("patientId", &p.PatientID).Using(user.IDValidator)
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
		StateReviewed,
		StateExpired,
		StateActive,
		StateInactive,
	}
}

func StatesVisibleToPatients() []string {
	return []string{
		StateSubmitted,
		StateReviewed,
		StateActive,
		StateInactive,
	}
}

func validPatientStateTransitions() map[string][]string {
	return map[string][]string{
		StateSubmitted: {StateReviewed},
		StateReviewed:  {StateActive},
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

func ValidStateTransitions(usr *user.User, state string) []string {
	if usr == nil {
		return []string{}
	}

	transitions := stateTransitionsForUser(usr)
	valid, ok := transitions[state]
	if !ok {
		return []string{}
	}

	return valid
}

type Filter struct {
	currentUser       *user.User
	clinicianID       string
	PatientID         string
	PatientEmail      string
	State             string
	ID                string
	CreatedTimeStart  *time.Time
	CreatedTimeEnd    *time.Time
	ModifiedTimeStart *time.Time
	ModifiedTimeEnd   *time.Time
}

func NewFilter(currentUser *user.User) (*Filter, error) {
	if currentUser == nil {
		return nil, errors.New("current user is missing")
	}

	f := &Filter{
		currentUser: currentUser,
	}

	if currentUser.HasRole(user.RoleClinic) {
		f.clinicianID = *currentUser.UserID
	} else {
		f.PatientID = *currentUser.UserID
	}

	return f, nil
}

func (f *Filter) GetClinicianID() string {
	return f.clinicianID
}

func (f *Filter) Validate(validator structure.Validator) {
	if f.ID != "" {
		validator.String("id", &f.ID).Hexadecimal().LengthEqualTo(24)
	}
	if f.currentUser.HasRole(user.RoleClinic) {
		validator.String("clinicianId", &f.clinicianID).NotEmpty().EqualTo(*f.currentUser.UserID)
		if f.State != "" {
			validator.String("state", &f.State).OneOf(States()...)
		}
		if f.PatientID != "" {
			validator.String("patientId", &f.PatientID).Using(user.IDValidator)
		}
		if f.PatientEmail != "" {
			validator.String("patientEmail", &f.PatientEmail).Email()
		}
	} else {
		validator.String("patientId", &f.PatientID).NotEmpty().EqualTo(*f.currentUser.UserID)
		if f.State != "" {
			validator.String("state", &f.State).OneOf(StatesVisibleToPatients()...)
		}
		validator.String("patientEmail", &f.PatientEmail).Empty()
	}
}

func (f *Filter) Parse(parser structure.ObjectParser) {
	if ptr := parser.String("id"); ptr != nil {
		f.ID = *ptr
	}
	if f.currentUser.HasRole(user.RoleClinic) {
		if ptr := parser.String("patientId"); ptr != nil {
			f.PatientID = *ptr
		}
		if ptr := parser.String("patientEmail"); ptr != nil {
			f.PatientEmail = *ptr
		}
	}
	if ptr := parser.String("state"); ptr != nil {
		f.State = *ptr
	}
	if ptr := parser.Time("createdTimeStart", time.RFC3339Nano); ptr != nil {
		f.CreatedTimeStart = ptr
	}
	if ptr := parser.Time("createdTimeEnd", time.RFC3339Nano); ptr != nil {
		f.CreatedTimeEnd = ptr
	}
	if ptr := parser.Time("modifiedTimeStart", time.RFC3339Nano); ptr != nil {
		f.ModifiedTimeStart = ptr
	}
	if ptr := parser.Time("modifiedTimeEnd", time.RFC3339Nano); ptr != nil {
		f.ModifiedTimeEnd = ptr
	}
}

type Update struct {
	prescription     *Prescription
	usr              *user.User
	Revision         *Revision
	State            string
	PrescriberUserID string
	PatientID        string
	ExpirationTime   *time.Time
	ModifiedTime     time.Time
	ModifiedUserID   string
}

func NewPrescriptionAddRevisionUpdate(usr *user.User, prescription *Prescription, create *RevisionCreate) *Update {
	revisionID := prescription.LatestRevision.RevisionID + 1
	revision := NewRevision(*usr.UserID, revisionID, create)
	update := &Update{
		usr:              usr,
		prescription:     prescription,
		Revision:         revision,
		State:            create.State,
		PrescriberUserID: revision.GetPrescriberUserID(),
		ExpirationTime:   revision.CalculateExpirationTime(),
		ModifiedUserID:   *usr.UserID,
		ModifiedTime:     time.Now(),
	}

	return update
}

func NewPrescriptionClaimUpdate(usr *user.User, prescription *Prescription) *Update {
	return &Update{
		usr:            usr,
		prescription:   prescription,
		State:          StateReviewed,
		PatientID:      *usr.UserID,
		ModifiedUserID: *usr.UserID,
		ModifiedTime:   time.Now(),
	}
}

func NewPrescriptionStateUpdate(usr *user.User, prescription *Prescription, update *StateUpdate) *Update {
	return &Update{
		usr:            usr,
		prescription:   prescription,
		State:          update.State,
		ModifiedUserID: *usr.UserID,
		ModifiedTime:   time.Now(),
	}
}

func (u *Update) GetUpdatedAccessCode() *string {
	if u.State != StateReviewed {
		return nil
	}

	// Remove the access code when the user reviews the prescription
	return pointer.FromString("")
}

func (u *Update) GetCurrentUserID() string {
	return *u.usr.UserID
}

func (u *Update) GetPrescriptionID() bson.ObjectId {
	return u.prescription.ID
}

func (u *Update) Validate(validator structure.Validator) {
	if u.usr == nil {
		validator.WithReference("user").ReportError(structureValidator.ErrorValueEmpty())
		return
	}
	if u.prescription == nil {
		validator.WithReference("prescription").ReportError(structureValidator.ErrorValueEmpty())
		return
	}

	validator.String("state", &u.State).OneOf(ValidStateTransitions(u.usr, u.prescription.State)...)

	if u.usr.HasRole(user.RoleClinic) {
		u.validateForClinician(validator)
	} else {
		u.validateForPatient(validator)
	}
}

func (u *Update) validateForClinician(validator structure.Validator) {
	if u.Revision != nil {
		u.Revision.Validate(validator.WithReference("revision"))
	} else {
		validator.WithReference("revision").ReportError(structureValidator.ErrorValueEmpty())
	}
	if u.PrescriberUserID != "" {
		validator.String("prescriberUserId", &u.PrescriberUserID).EqualTo(*u.usr.UserID)
	}
	if u.PatientID != "" {
		validator.String("patientId", &u.PatientID).Using(user.IDValidator)
	}
}

func (u *Update) validateForPatient(validator structure.Validator) {
	if u.Revision != nil {
		validator.WithReference("revision").ReportError(structureValidator.ErrorValueExists())
	}
	if u.PrescriberUserID != "" {
		validator.String("prescriberUserId", &u.PrescriberUserID).Empty()
	}
	if u.PatientID != "" {
		validator.String("patientId", &u.PatientID).EqualTo(*u.usr.UserID)
	}
}

type Claim struct {
	AccessCode string `json:"accessCode"`
}

func NewPrescriptionClaim() *Claim {
	return &Claim{}
}

func (p *Claim) Validate(validator structure.Validator) {
	validator.String("accessCode", &p.AccessCode).NotEmpty()
}

type StateUpdate struct {
	State string `json:"state"`
}

func NewStateUpdate() *StateUpdate {
	return &StateUpdate{}
}

func (s *StateUpdate) Validate(validator structure.Validator) {
	validator.String("status", &s.State).OneOf(StatesVisibleToPatients()...)
}
