package postprocess

import (
	"context"
	"slices"
	"time"

	mapset "github.com/deckarep/golang-set/v2"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/metadata"
	"github.com/tidepool-org/platform/pointer"
	userWork "github.com/tidepool-org/platform/user/work"
	"github.com/tidepool-org/platform/work"
)

// Enqueue creates a work item to signal a change to the data of a user to trigger the postprocessor.
// Work is created for every change reported, rather than merged into the work already pending for the user,
// so that reporting a change is a single insert when data is uploaded. If there are multiple pending work items
// for the same user, they are merged during processing.
func Enqueue(ctx context.Context, workClient work.Client, userID string, reasons ...string) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if workClient == nil {
		return errors.New("work client is missing")
	}
	if userID == "" {
		return errors.New("user id is missing")
	}
	if len(reasons) == 0 {
		return errors.New("reasons is missing")
	}

	create, err := newCreate(userID, reasons)
	if err != nil {
		return err
	}

	if _, err = workClient.Create(ctx, create); err != nil {
		return errors.Wrap(err, "unable to create work")
	}

	return nil
}

func newCreate(userID string, reasons []string) (*work.Create, error) {
	create, err := metadata.WithMetadata(
		&work.Create{
			Type:                    Type,
			GroupID:                 pointer.FromString(IDFromUserID(userID)),
			SerialID:                pointer.FromString(IDFromUserID(userID)),
			ProcessingAvailableTime: time.Now(),
			ProcessingTimeout:       int(ProcessingTimeout.Seconds()),
		},
		&Metadata{
			Metadata: userWork.Metadata{UserID: pointer.FromString(userID)},
			Reasons:  normalizeReasons(reasons),
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create work create")
	}
	return create, nil
}

// normalizeReasons reports the reasons given, combined, without duplicates, and ordered
func normalizeReasons(reasons ...[]string) []string {
	union := mapset.NewSet[string]()
	for _, each := range reasons {
		union.Append(each...)
	}
	return slices.Sorted(slices.Values(union.ToSlice()))
}
