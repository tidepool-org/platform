package requests

import (
	"context"
	"fmt"
	"time"

	"github.com/tidepool-org/platform/clinics"
	dataSource "github.com/tidepool-org/platform/data/source"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/mailer"
	"github.com/tidepool-org/platform/metadata"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/user"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

const (
	Type              = "org.tidepool.processors.connections.requests"
	Quantity          = 1
	Frequency         = 1 * time.Minute
	ProcessingTimeout = 1 * time.Minute
)

type (
	ClinicClient     = clinics.Client
	DataSourceClient = dataSource.Client
	MailerClient     = mailer.Mailer
	UserClient       = user.Client
)

type Dependencies struct {
	workBase.Dependencies
	ClinicClient
	DataSourceClient
	MailerClient
	UserClient
}

func (d Dependencies) Validate() error {
	if err := d.Dependencies.Validate(); err != nil {
		return err
	}
	if d.ClinicClient == nil {
		return errors.New("clinic client is missing")
	}
	if d.DataSourceClient == nil {
		return errors.New("data source client is missing")
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

func AddWorkItem(ctx context.Context, client work.Client, workMetadata Metadata) error {
	if create, err := NewWorkCreate(workMetadata); err != nil {
		return errors.Wrap(err, "unable to create work create")
	} else if _, err = client.DeleteAllByGroupID(ctx, *create.GroupID); err != nil {
		return errors.Wrapf(err, "unable to delete existing group with id %q", *create.GroupID)
	} else if _, err := client.Create(ctx, create); err != nil {
		return err
	} else {
		return nil
	}
}

func NewWorkCreate(workMetadata Metadata) (*work.Create, error) {
	var processingAvailableTime time.Time
	if workMetadata.WhenToSend.IsZero() {
		processingAvailableTime = time.Now().Add(7 * 24 * time.Hour)
	} else {
		processingAvailableTime = workMetadata.WhenToSend
	}
	return metadata.WithMetadata(
		&work.Create{
			Type:                    Type,
			SerialID:                pointer.FromString(workMetadata.UserID),
			GroupID:                 pointer.FromString(NewGroupID(workMetadata.UserID, workMetadata.ProviderName)),
			ProcessingTimeout:       int(ProcessingTimeout.Seconds()),
			ProcessingAvailableTime: processingAvailableTime,
		},
		&workMetadata,
	)
}

// NewGroupID returns a string suitable for [work.Work.GroupID] for batch deletions.
func NewGroupID(userID string, providerName string) string {
	return fmt.Sprintf("%s:%s:%s", Type, userID, providerName)
}
