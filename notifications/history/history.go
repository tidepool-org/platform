package history

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/structure"
)

const (
	NotificationQueued            = "queued"
	NotificationGeneralError      = "error"
	NotificationConditionsExpired = "conditions-expired"
	NotificationAttempted         = "email-attempted"
	NotificationEmailError        = "email-error"
	NotificationEmailSent         = "email-sent"
	NotificationOther             = "other"
	NotificationZero              = ""
)

var (
	notificationEventTypes = []string{
		NotificationQueued,
		NotificationGeneralError,
		NotificationConditionsExpired,
		NotificationAttempted,
		NotificationEmailError,
		NotificationEmailSent,
		NotificationOther,
		NotificationZero,
	}
)

type Recorder interface {
	Create(ctx context.Context, entry Entry) error
	List(ctx context.Context, filter Filter, pagination *page.Pagination) ([]Entry, error)
}

type Filter struct {
	ProcessorType string
	EventType     string
	GroupID       string
	UserID        string
	DataSourceID  string
}

func (f *Filter) Validate(validator structure.Validator) {
	validator.String("processorType", &f.ProcessorType).NotEmpty()
	eventType := string(f.EventType)
	validator.String("eventType", &eventType).OneOf(notificationEventTypes...)
	validator.String("userID", &f.UserID).NotEmpty()
}

type Entry struct {
	ProcessorType string    `bson:"processorType,omitempty"`
	EventType     string    `bson:"eventType,omitempty"`
	CreatedTime   time.Time `bson:"createdTime"`
	DataSourceID  string    `bson:"dataSourceId,omitempty"`
	Email         string    `bson:"email,omitempty"`
	GroupID       string    `bson:"groupId,omitempty"`
	Metadata      bson.M    `bson:"metadata,omitempty"`
	UserID        string    `bson:"userId,omitempty"`
}

func (l *Entry) Validate(validator structure.Validator) {
	validator.String("processorType", &l.ProcessorType).NotEmpty()
	validator.String("eventType", (*string)(&l.EventType)).NotEmpty()
	validator.String("userId", &l.UserID).NotEmpty()
	validator.Time("createdTime", &l.CreatedTime).NotZero()
}
