package claims

import (
	"context"
	"fmt"
	"time"

	"github.com/tidepool-org/platform/notifications"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/work"
)

const (
	processorType            = "org.tidepool.processors.users.claims"
	quantity                 = 2
	frequency                = time.Minute
	processingTimeoutSeconds = 60
)

// NewGroupID returns a string suitable for [work.Work.GroupID] that is meant
// to group related claim account notifications together so they can all be
// deleted if the condition to send them is no longer active. For example, if a
// user has already claimed their account but there is a pending notification
// that hasn't been processed yetm the processor should delete all work items
// of the same group id when it is time to process the item.
func NewGroupID(userId string) string {
	return fmt.Sprintf("%s:%s", processorType, userId)
}

type Metadata struct {
	ClinicID   string    `json:"clinicId,omitempty"`
	UserId     string    `json:"userId,omitempty"`
	WhenToSend time.Time `json:"whenToSend,omitzero"`
}

func (d *Metadata) Parse(parser structure.ObjectParser) {
	d.ClinicID = pointer.ToString(parser.String("clinicId"))
	d.UserId = pointer.ToString(parser.String("userId"))
	d.WhenToSend = pointer.ToTime(parser.Time("whenToSend", time.RFC3339Nano))
}

func (d *Metadata) Validate(validator structure.Validator) {
	validator.String("clinicId", &d.ClinicID).NotEmpty()
	validator.String("userId", &d.UserId).NotEmpty()
}

func AddWorkItem(ctx context.Context, client work.Client, metadata Metadata) error {
	whenToSend := metadata.WhenToSend
	if whenToSend.IsZero() {
		whenToSend = time.Now().Add(time.Hour * 24 * 7)
	}
	create := newWorkCreate(whenToSend, metadata)
	if groupID := pointer.DefaultString(create.GroupID, ""); groupID != "" {
		// Delete any other work items with the same group id because if a new reminder is added, any older ones would be too early since the last reminder of the same group id.
		if _, err := client.DeleteAllByGroupID(ctx, groupID); err != nil {
			return fmt.Errorf(`unable to delete existing groups by id "%s": %w`, groupID, err)
		}
	}
	if _, err := client.Create(ctx, create); err != nil {
		return err
	}
	return nil
}

func newWorkCreate(notBefore time.Time, metadata Metadata) *work.Create {
	return &work.Create{
		Type:                    processorType,
		SerialID:                pointer.FromString(metadata.UserId),
		GroupID:                 pointer.FromString(NewGroupID(metadata.UserId)),
		ProcessingTimeout:       processingTimeoutSeconds,
		ProcessingAvailableTime: notBefore,
		Metadata:                fromClaimAccountData(metadata),
	}
}

type processor struct {
	dependencies conditionalnotifications.Dependencies
}

func NewProcessor(dependencies conditionalnotifications.Dependencies) *processor {
	return &processor{
		dependencies: dependencies,
	}
}

func (p *processor) Type() string {
	return processorType
}

func (p *processor) Quantity() int {
	return quantity
}

func (p *processor) Frequency() time.Duration {
	return frequency
}

func (p *processor) Process(ctx context.Context, wrk *work.Work, updater work.ProcessingUpdater) work.ProcessResult {
	data, err := toClaimAccountData(wrk)
	if err != nil {
		return conditionalnotifications.NewFailingResult(err, wrk)
	}

	patient, err := p.dependencies.Clinics.GetPatient(ctx, data.ClinicID, data.UserId)
	if err != nil {
		return conditionalnotifications.NewFailingResult(err, wrk)
	}
	if patient == nil {
		return conditionalnotifications.NewFailingResult(fmt.Errorf(`unable to find patient with userId "%v"`, data.UserId), wrk)
	}
	if pointer.ToString(patient.Email) == "" {
		return conditionalnotifications.NewFailingResult(fmt.Errorf(`unable to find email for patient with userId "%v"`, data.UserId), wrk)
	}
	// If user already claimed they will no longer have the custodian field set
	if patient != nil && (patient.Permissions == nil || patient.Permissions.Custodian == nil) {
		return *work.NewProcessResultDelete()
	}

	if _, err := p.dependencies.Confirmation.ResendAccountSignupWithResponse(ctx, *patient.Email); err != nil {
		return conditionalnotifications.NewFailingResult(fmt.Errorf(`unable to resend account signup email: %w`, err), wrk)
	}
	return *work.NewProcessResultDelete()
}

func toClaimAccountData(wrk *work.Work) (*Metadata, error) {
	wrk.EnsureMetadata()
	var data Metadata
	if userId, ok := wrk.Metadata["userId"].(string); ok {
		data.UserId = userId
	} else {
		return nil, fmt.Errorf(`expected field "userId" to exist and be a string, received %T`, wrk.Metadata["userId"])
	}
	if clinicId, ok := wrk.Metadata["clinicId"].(string); ok {
		data.ClinicID = clinicId
	} else {
		return nil, fmt.Errorf(`expected field "clinicId" to exist and be a string, received %T`, wrk.Metadata["clinicId"])
	}
	return &data, nil
}

func fromClaimAccountData(data Metadata) map[string]any {
	return map[string]any{
		"userId":   data.UserId,
		"clinicId": data.ClinicID,
	}
}
