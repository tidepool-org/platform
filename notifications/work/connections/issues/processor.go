package issues

import (
	"context"
	"fmt"
	"time"

	"github.com/tidepool-org/go-common/events"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/notifications"
	"github.com/tidepool-org/platform/notifications/history"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/work"
)

const (
	Type                     = "org.tidepool.user.notification.connection.issue"
	quantity                 = 2
	frequency                = time.Minute
	processingTimeoutSeconds = 60
)

// NewGroupID returns a string suitable for [work.Work.GroupID] for batch deletions.
func NewGroupID(dataSourceID string) string {
	return fmt.Sprintf("%s:%s", Type, dataSourceID)
}

type processor struct {
	dependencies notifications.Dependencies
}

type Metadata struct {
	DataSourceState   string `json:"dataSourceState,omitempty"`
	DataSourceID      string `json:"dataSourceId,omitempty"`
	EmailTemplate     string `json:"emailTemplate,omitempty"`
	FullName          string `json:"fullName,omitempty"`
	ProviderName      string `json:"providerName,omitempty"`
	RestrictedTokenID string `json:"restrictedTokenId,omitempty"`
	UserID            string `json:"userId,omitempty"`
}

func (d *Metadata) Parse(parser structure.ObjectParser) {
	d.DataSourceState = pointer.ToString(parser.String("dataSourceState"))
	d.DataSourceID = pointer.ToString(parser.String("dataSourceId"))
	d.EmailTemplate = pointer.ToString(parser.String("emailTemplate"))
	d.FullName = pointer.ToString(parser.String("fullName"))
	d.ProviderName = pointer.ToString(parser.String("providerName"))
	d.RestrictedTokenID = pointer.ToString(parser.String("restrictedTokenId"))
	d.UserID = pointer.ToString(parser.String("userId"))
}

func (d *Metadata) Validate(validator structure.Validator) {
	validator.String("dataSourceState", &d.DataSourceState).NotEmpty()
	validator.String("dataSourceId", &d.DataSourceID).NotEmpty()
	validator.String("emailTemplate", &d.EmailTemplate).NotEmpty()
	validator.String("fullName", &d.FullName).NotEmpty()
	validator.String("providerName", &d.ProviderName).NotEmpty()
	validator.String("userId", &d.UserID).NotEmpty()
}

func AddWorkItem(ctx context.Context, client work.Client, recorder history.Recorder, metadata Metadata) error {
	create := newWorkCreate(metadata)
	if _, err := client.Create(ctx, create); err != nil {
		return err
	}
	groupID := pointer.DefaultString(create.GroupID, "")
	entry := history.Entry{
		Metadata:      metadata,
		ProcessorType: Type,
		EventType:     history.NotificationQueued,
		GroupID:       groupID,
		DataSourceID:  metadata.DataSourceID,
		UserID:        metadata.UserID,
	}
	if err := recorder.Create(ctx, entry); err != nil {
		return err
	}
	return nil
}

func newWorkCreate(metadata Metadata) *work.Create {
	return &work.Create{
		Type:              Type,
		SerialID:          pointer.FromString(metadata.UserID),
		GroupID:           pointer.FromString(NewGroupID(metadata.DataSourceID)),
		ProcessingTimeout: processingTimeoutSeconds,
		Metadata:          fromMetadata(metadata),
	}
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
	data, err := toMetadata(wrk)
	if err != nil {
		return notifications.NewFailingResult(err, wrk)
	}

	user, err := p.dependencies.Users.Get(ctx, data.UserID)
	if err != nil {
		return notifications.NewFailingResult(err, wrk)
	}
	if user == nil || user.Username == nil {
		entry := history.Entry{
			Metadata:      wrk.Metadata,
			ProcessorType: Type,
			EventType:     history.NotificationGeneralError,
			GroupID:       pointer.DefaultString(wrk.GroupID, ""),
			DataSourceID:  data.DataSourceID,
			UserID:        data.UserID,
			Error:         errors.New("user not found"),
		}
		if err := p.dependencies.Recorder.Create(ctx, entry); err != nil {
			if lgr := log.LoggerFromContext(ctx); lgr != nil {
				lgr.WithFields(wrk.Metadata).Warn("unable to to record notification error event.")
			}
		}

		return notifications.NewFailingResult(errors.Newf(`unable to find user for userId "%s"`, data.UserID), wrk)
	}

	emailVars := map[string]string{
		"RestrictedTokenId": data.RestrictedTokenID,
		"FullName":          data.FullName,
		"ProviderName":      data.ProviderName,
	}
	templateEvent := events.SendEmailTemplateEvent{
		Recipient: *user.Username,
		Template:  data.EmailTemplate,
		Variables: emailVars,
	}
	entry := history.Entry{
		Metadata:      wrk.Metadata,
		ProcessorType: Type,
		GroupID:       pointer.DefaultString(wrk.GroupID, ""),
		DataSourceID:  data.DataSourceID,
		EventType:     history.NotificationAttempted,
		UserID:        data.UserID,
	}
	if err := p.dependencies.Recorder.Create(ctx, entry); err != nil {
		if lgr := log.LoggerFromContext(ctx); lgr != nil {
			lgr.WithFields(wrk.Metadata).Warn("unable to to record notification email attempted event.")
		}
	}

	if err := p.dependencies.Mailer.SendEmailTemplate(ctx, templateEvent); err != nil {
		entry := history.Entry{
			Metadata:      wrk.Metadata,
			ProcessorType: Type,
			GroupID:       pointer.DefaultString(wrk.GroupID, ""),
			DataSourceID:  data.DataSourceID,
			EventType:     history.NotificationEmailError,
			UserID:        data.UserID,
			Error:         errors.Wrap(err, "unable to send email template for device issues"),
		}
		if err := p.dependencies.Recorder.Create(ctx, entry); err != nil {
			if lgr := log.LoggerFromContext(ctx); lgr != nil {
				lgr.WithFields(wrk.Metadata).Warn("unable to to record notification send email error event.")
			}
		}

		return notifications.NewFailingResult(err, wrk)
	}
	entry = history.Entry{
		Metadata:      wrk.Metadata,
		ProcessorType: Type,
		Email:         pointer.DefaultString(user.Username, ""),
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

func toMetadata(wrk *work.Work) (*Metadata, error) {
	wrk.EnsureMetadata()
	var data Metadata
	if userID, ok := wrk.Metadata["userId"].(string); ok {
		data.UserID = userID
	} else {
		return nil, errors.Newf(`expected field "userId" to exist and be a string, received %T`, wrk.Metadata["userId"])
	}
	if providerName, ok := wrk.Metadata["providerName"].(string); ok {
		data.ProviderName = providerName
	} else {
		return nil, errors.Newf(`expected field "providerName" to exist and be a string, received %T`, wrk.Metadata["providerName"])
	}
	if dataSourceState, ok := wrk.Metadata["dataSourceState"].(string); ok {
		data.DataSourceState = dataSourceState
	} else {
		return nil, errors.Newf(`expected field "dataSourceState" to exist and be a string, received %T`, wrk.Metadata["dataSourceState"])
	}
	if fullName, ok := wrk.Metadata["fullName"].(string); ok {
		data.FullName = fullName
	} else {
		return nil, errors.Newf(`expected field "fullName" to exist and be a string, received %T`, wrk.Metadata["fullName"])
	}
	if restrictedTokenID, ok := wrk.Metadata["restrictedTokenId"].(string); ok {
		data.RestrictedTokenID = restrictedTokenID
	} else {
		return nil, errors.Newf(`expected field "restrictedTokenId" to exist and be a string, received %T`, wrk.Metadata["restrictedTokenId"])
	}
	if emailTemplate, ok := wrk.Metadata["emailTemplate"].(string); ok {
		data.EmailTemplate = emailTemplate
	} else {
		return nil, errors.Newf(`expected field "emailTemplate" to exist and be a string, received %T`, wrk.Metadata["emailTemplate"])
	}
	return &data, nil
}

func fromMetadata(data Metadata) map[string]any {
	return map[string]any{
		"userId":            data.UserID,
		"providerName":      data.ProviderName,
		"dataSourceState":   data.DataSourceState,
		"fullName":          data.FullName,
		"restrictedTokenId": data.RestrictedTokenID,
		"emailTemplate":     data.EmailTemplate,
	}
}
