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

// Enqueue creates work reporting one reason to signal a change to the data of a user to trigger the postprocessor.
// Reporting a change is a single insert, deduplicated by user and reason against the work not yet claimed, so
// that frequent changes coalesce into one work item rather than one each. A change reported while such work waits
// is discarded, not merged, so the reason is the only thing a change carries: anything else reported with it, such
// as which data set changed, would be lost. The work is deduplicated per reason rather than per user so that a
// reason reported while another waits is not discarded, which is why a caller with several reasons to report calls
// once per reason. Work that has been claimed releases its deduplication id, as it may no longer observe the change
// reported, and the work created in its place is merged into it during processing if it is still pending then.
func Enqueue(ctx context.Context, workClient work.Client, userID string, reason string) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if workClient == nil {
		return errors.New("work client is missing")
	}
	if userID == "" {
		return errors.New("user id is missing")
	}
	if !slices.Contains(Reasons(), reason) {
		return errors.New("reason is invalid")
	}

	create, err := newCreate(userID, reason)
	if err != nil {
		return err
	}

	if _, err = workClient.Create(ctx, create); err != nil {
		return errors.Wrap(err, "unable to create work")
	}

	return nil
}

func newCreate(userID string, reason string) (*work.Create, error) {
	create, err := metadata.WithMetadata(
		&work.Create{
			Type:                    Type,
			GroupID:                 pointer.FromString(IDFromUserID(userID)),
			DeduplicationID:         pointer.FromString(DeduplicationIDFromUserIDAndReason(userID, reason)),
			SerialID:                pointer.FromString(IDFromUserID(userID)),
			ProcessingAvailableTime: time.Now(),
			ProcessingTimeout:       int(ProcessingTimeout.Seconds()),
		},
		&Metadata{
			Metadata: userWork.Metadata{UserID: pointer.FromString(userID)},
			Reasons:  []string{reason},
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
