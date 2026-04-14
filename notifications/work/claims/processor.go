package claims

import (
	"context"
	"fmt"
	"time"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/notifications"
	"github.com/tidepool-org/platform/notifications/history"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/work"
)

const (
	Type                     = "org.tidepool.user.notification.account.claim"
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
func NewGroupID(userID string) string {
	return fmt.Sprintf("%s:%s", Type, userID)
}

type Metadata struct {
	ClinicID   string    `json:"clinicId,omitempty"`
	UserID     string    `json:"userId,omitempty"`
	WhenToSend time.Time `json:"whenToSend,omitzero"`
}

func (d *Metadata) Parse(parser structure.ObjectParser) {
	d.ClinicID = pointer.ToString(parser.String("clinicId"))
	d.UserID = pointer.ToString(parser.String("userId"))
	d.WhenToSend = pointer.ToTime(parser.Time("whenToSend", time.RFC3339Nano))
}

func (d *Metadata) Validate(validator structure.Validator) {
	validator.String("clinicId", &d.ClinicID).NotEmpty()
	validator.String("userId", &d.UserID).NotEmpty()
}

func AddWorkItem(ctx context.Context, client work.Client, recorder history.Recorder, metadata Metadata) error {
	whenToSend := metadata.WhenToSend
	if whenToSend.IsZero() {
		whenToSend = time.Now().Add(time.Hour * 24 * 7)
	}
	create := newWorkCreate(whenToSend, metadata)
	groupID := pointer.DefaultString(create.GroupID, "")
	if groupID != "" {
		// Delete any other work items with the same group id because if a new reminder is added, any older ones would be too early since the last reminder of the same group id.
		if _, err := client.DeleteAllByGroupID(ctx, groupID); err != nil {
			return errors.Wrapf(err, `unable to delete existing groups by id "%s"`, groupID)
		}
	}
	wrk, err := client.Create(ctx, create)
	if err != nil {
		return err
	}
	entry := history.Entry{
		Metadata:      wrk.Metadata,
		ProcessorType: Type,
		EventType:     history.NotificationQueued,
		GroupID:       groupID,
		UserID:        metadata.UserID,
	}
	if err := recorder.Create(ctx, entry); err != nil {
		return err
	}
	return nil
}

func newWorkCreate(notBefore time.Time, metadata Metadata) *work.Create {
	return &work.Create{
		Type:                    Type,
		SerialID:                pointer.FromString(metadata.UserID),
		GroupID:                 pointer.FromString(NewGroupID(metadata.UserID)),
		ProcessingTimeout:       processingTimeoutSeconds,
		ProcessingAvailableTime: notBefore,
		Metadata:                fromClaimAccountData(metadata),
	}
}

type processor struct {
	dependencies notifications.Dependencies
}

func NewProcessor(dependencies notifications.Dependencies) *processor {
	return &processor{
		dependencies: dependencies,
	}
}

func (p *processor) Type() string {
	return Type
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
		return notifications.NewFailingResult(err, wrk)
	}

	patient, err := p.dependencies.Clinics.GetPatient(ctx, data.ClinicID, data.UserID)
	if err != nil {
		return notifications.NewFailingResult(err, wrk)
	}
	if patient == nil {
		return notifications.NewFailingResult(errors.Newf(`unable to find patient with userId "%v"`, data.UserID), wrk)
	}
	if pointer.ToString(patient.Email) == "" {
		return notifications.NewFailingResult(errors.Newf(`unable to find email for patient with userId "%v"`, data.UserID), wrk)
	}
	// If user already claimed they will no longer have the custodian field set
	if patient != nil && (patient.Permissions == nil || patient.Permissions.Custodian == nil) {
		entry := history.Entry{
			Metadata:      wrk.Metadata,
			ProcessorType: Type,
			Email:         pointer.DefaultString(patient.Email, ""),
			GroupID:       pointer.DefaultString(wrk.GroupID, ""),
			EventType:     history.NotificationConditionsExpired,
			UserID:        data.UserID,
		}
		if err := p.dependencies.Recorder.Create(ctx, entry); err != nil {
			if lgr := log.LoggerFromContext(ctx); lgr != nil {
				lgr.WithFields(wrk.Metadata).Warn("unable to to record notification conditions expired event.")
			}
		}

		return *work.NewProcessResultDelete()
	}

	entry := history.Entry{
		Metadata:      wrk.Metadata,
		ProcessorType: Type,
		Email:         pointer.DefaultString(patient.Email, ""),
		GroupID:       pointer.DefaultString(wrk.GroupID, ""),
		EventType:     history.NotificationAttempted,
		UserID:        data.UserID,
	}
	if err := p.dependencies.Recorder.Create(ctx, entry); err != nil {
		if lgr := log.LoggerFromContext(ctx); lgr != nil {
			lgr.WithFields(wrk.Metadata).Warn("unable to to record notification attempted event.")
		}
	}

	if _, err := p.dependencies.Confirmation.ResendAccountSignupWithResponse(ctx, *patient.Email); err != nil {
		entry := history.Entry{
			Metadata:      wrk.Metadata,
			ProcessorType: Type,
			Email:         pointer.DefaultString(patient.Email, ""),
			GroupID:       pointer.DefaultString(wrk.GroupID, ""),
			EventType:     history.NotificationEmailError,
			UserID:        data.UserID,
			Error:         errors.Wrap(err, "unable to resend account signup"),
		}
		if err := p.dependencies.Recorder.Create(ctx, entry); err != nil {
			if lgr := log.LoggerFromContext(ctx); lgr != nil {
				lgr.WithFields(wrk.Metadata).Warn("unable to to record notification email send error event.")
			}
		}

		return notifications.NewFailingResult(errors.Newf(`unable to resend account signup email`), wrk)
	}
	entry = history.Entry{
		Metadata:      wrk.Metadata,
		ProcessorType: Type,
		Email:         pointer.DefaultString(patient.Email, ""),
		GroupID:       pointer.DefaultString(wrk.GroupID, ""),
		EventType:     history.NotificationEmailSent,
		UserID:        data.UserID,
	}
	if err := p.dependencies.Recorder.Create(ctx, entry); err != nil {
		if lgr := log.LoggerFromContext(ctx); lgr != nil {
			lgr.WithFields(wrk.Metadata).Warn("unable to to record notification email sent event.")
		}
	}

	return *work.NewProcessResultDelete()
}

func toClaimAccountData(wrk *work.Work) (*Metadata, error) {
	wrk.EnsureMetadata()
	var data Metadata
	if userID, ok := wrk.Metadata["userId"].(string); ok {
		data.UserID = userID
	} else {
		return nil, errors.Newf(`expected field "userId" to exist and be a string, received %T`, wrk.Metadata["userId"])
	}
	if clinicID, ok := wrk.Metadata["clinicId"].(string); ok {
		data.ClinicID = clinicID
	} else {
		return nil, errors.Newf(`expected field "clinicId" to exist and be a string, received %T`, wrk.Metadata["clinicId"])
	}
	return &data, nil
}

func fromClaimAccountData(data Metadata) map[string]any {
	return map[string]any{
		"userId":   data.UserID,
		"clinicId": data.ClinicID,
	}
}
