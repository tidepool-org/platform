package claims

import (
	"context"
	"fmt"
	"time"

	confirmationClient "github.com/tidepool-org/hydrophone/client"

	"github.com/tidepool-org/platform/clinics"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/metadata"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

const (
	Type              = "org.tidepool.processors.users.claims"
	Quantity          = 1
	Frequency         = 1 * time.Minute
	ProcessingTimeout = 1 * time.Minute
)

type (
	ClinicClient       = clinics.Client
	ConfirmationClient = confirmationClient.ClientWithResponsesInterface
)

type Dependencies struct {
	workBase.Dependencies
	ClinicClient
	ConfirmationClient
}

func (d Dependencies) Validate() error {
	if err := d.Dependencies.Validate(); err != nil {
		return err
	}
	if d.ClinicClient == nil {
		return errors.New("clinics is missing")
	}
	if d.ConfirmationClient == nil {
		return errors.New("confirmation client is missing")
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
			GroupID:                 pointer.FromString(NewGroupID(workMetadata.UserID)),
			ProcessingTimeout:       int(ProcessingTimeout.Seconds()),
			ProcessingAvailableTime: processingAvailableTime,
		},
		&workMetadata,
	)
}

// NewGroupID returns a string suitable for [work.Work.GroupID] that is meant
// to group related claim account notifications together so they can all be
// deleted if the condition to send them is no longer active. For example, if a
// user has already claimed their account but there is a pending notification
// that hasn't been processed yet the processor should delete all work items
// of the same group id when it is time to process the item.
func NewGroupID(userID string) string {
	return fmt.Sprintf("%s:%s", Type, userID)
}
