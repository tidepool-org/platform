package prescription

import (
	"context"
	"time"

	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/user"

	"github.com/tidepool-org/platform/id"

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

type Client interface {
	Accessor
}

type Accessor interface {
	CreatePrescription(ctx context.Context, userID string, create *RevisionCreate) (*Prescription, error)
}

type Prescription struct {
	ID               string     `json:"id" bson:"id"`
	PatientID        string     `json:"patientId,omitempty" bson:"patientId,omitempty"`
	AccessCode       string     `json:"accessCode,omitempty" bson:"accessCode"`
	State            string     `json:"state" bson:"state"`
	LatestRevision   *Revision  `json:"latestRevision" bson:"latestRevision"`
	RevisionHistory  Revisions  `json:"-" bson:"revisionHistory"`
	ExpirationTime   *time.Time `json:"expirationTime" bson:"expirationTime"`
	PrescriberUserID string     `json:"prescriberUserId,omitempty" bson:"prescriberUserId,omitempty"`
	CreatedTime      time.Time  `json:"createdTime" bson:"createdTime"`
	CreatedUserID    string     `json:"createdUserId" bson:"createdUserId"`
	DeletedTime      *time.Time `json:"deletedTime,omitempty" bson:"deletedTime,omitempty"`
	DeletedUserID    string     `json:"deletedUserId,omitempty" bson:"deletedUserId,omitempty"`
}

func NewPrescriptionID() string {
	return id.Must(id.New(8))
}

func NewPrescription(ctx context.Context, userID string, revisionCreate *RevisionCreate) (*Prescription, error) {
	now := time.Now()
	accessCode := GenerateAccessCode()
	revision := NewRevision(ctx, userID, 0, revisionCreate)
	revisionHistory := []*Revision{revision}
	prescription := &Prescription{
		ID:              NewPrescriptionID(),
		AccessCode:      accessCode,
		State:           revisionCreate.State,
		LatestRevision:  revision,
		RevisionHistory: revisionHistory,
		ExpirationTime:  revision.CalculateExpirationTime(),
		CreatedTime:     now,
		CreatedUserID:   userID,
	}

	return prescription, nil
}

type Prescriptions []*Prescription

func (p *Prescription) Validate(validator structure.Validator) {
	validator.String("id", &p.ID).LengthEqualTo(16).Hexadecimal()

	if p.PatientID != "" {
		validator.String("patientId", &p.PatientID).Using(user.IDValidator)
	}

	validator.String("accessCode", &p.AccessCode).LengthEqualTo(6).Alphanumeric()

	validator.String("state", &p.State).OneOf(States()...)

	if p.LatestRevision != nil {
		p.Validate(validator.WithReference("latestRevision"))
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

	if p.CreatedUserID != "" {
		validator.String("createdUserId", &p.CreatedUserID).Using(user.IDValidator)
	}

	if p.DeletedTime != nil {
		validator.Time("deletedTime", p.DeletedTime).NotZero()
	}

	if p.DeletedUserID != "" {
		validator.String("deletedUserId", &p.DeletedUserID).Using(user.IDValidator)
	}
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
