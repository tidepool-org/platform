package issues

import (
	"context"
	"fmt"
	"time"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/mailer"
	"github.com/tidepool-org/platform/metadata"
	notificationsHistory "github.com/tidepool-org/platform/notifications/history"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/user"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

const (
	Type              = "org.tidepool.user.notification.connection.issue"
	Quantity          = 1
	Frequency         = 1 * time.Minute
	ProcessingTimeout = 1 * time.Minute
)

type (
	HistoryRecorder = notificationsHistory.Recorder
	MailerClient    = mailer.Client
	UserClient      = user.Client
)

type Dependencies struct {
	workBase.Dependencies
	HistoryRecorder
	MailerClient
	UserClient
}

func (d Dependencies) Validate() error {
	if err := d.Dependencies.Validate(); err != nil {
		return err
	}
	if d.HistoryRecorder == nil {
		return errors.New("history recorder is missing")
	}
	if d.MailerClient == nil {
		return errors.New("mailer client is missing")
	}
	if d.UserClient == nil {
		return errors.New("user client is missing")
	}
	return nil
}

func NewProcessorFactory(dependencies Dependencies) (*workBase.ProcessorFactory, error) {
	if err := dependencies.Validate(); err != nil {
		return nil, errors.Wrap(err, "dependencies is invalid")
	}
	processorFactory := func() (work.Processor, error) { return NewProcessor(dependencies) }
	return workBase.NewProcessorFactory(Type, Quantity, Frequency, processorFactory)
}

func AddWorkItem(ctx context.Context, workMetadata Metadata, workClient work.Client, historyRecorder notificationsHistory.Recorder) error {
	if workClient == nil {
		return errors.New("work client is missing")
	} else if historyRecorder == nil {
		return errors.New("history recorder is missing")
	} else if create, err := NewWorkCreate(workMetadata); err != nil {
		return errors.Wrap(err, "unable to create work create")
	} else if wrk, err := workClient.Create(ctx, create); err != nil {
		return errors.Wrap(err, "unable to create work")
	} else {
		entry := notificationsHistory.Entry{
			EventType:     notificationsHistory.NotificationQueued,
			ProcessorType: Type,
			GroupID:       *wrk.GroupID,
			UserID:        workMetadata.UserID,
			DataSourceID:  workMetadata.DataSourceID,
			Metadata:      wrk.Metadata,
		}
		return historyRecorder.Create(ctx, entry)
	}
}

func NewWorkCreate(workMetadata Metadata) (*work.Create, error) {
	return metadata.WithMetadata(
		&work.Create{
			Type:              Type,
			SerialID:          pointer.FromString(workMetadata.UserID),
			GroupID:           pointer.FromString(NewGroupID(workMetadata.DataSourceID)),
			ProcessingTimeout: int(ProcessingTimeout.Seconds()),
		},
		&workMetadata,
	)
}

// NewGroupID returns a string suitable for [work.Work.GroupID] for batch deletions.
func NewGroupID(dataSourceID string) string {
	return fmt.Sprintf("%s:%s", Type, dataSourceID)
}
