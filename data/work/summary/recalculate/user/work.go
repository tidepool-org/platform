// Package user implements the work that recalculates the summaries of a single user.
package user

import (
	"fmt"
	"time"

	dataWorkPostprocess "github.com/tidepool-org/platform/data/work/postprocess"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/metadata"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	summaryTypes "github.com/tidepool-org/platform/summary/types"
	userWork "github.com/tidepool-org/platform/user/work"
	"github.com/tidepool-org/platform/work"
)

const (
	Type = "org.tidepool.data.summary.recalculate.user"

	ProcessingTimeout = 5 * time.Minute
)

const (
	MetadataKeyPendingSummaryUpdates = "pendingSummaryUpdates"
	MetadataKeyPendingSummaryDeletes = "pendingSummaryDeletes"
)

type Metadata struct {
	userWork.Metadata `bson:",inline"`

	// Summary changes already calculated by this work item but not yet reported to the
	// clinic service.
	PendingSummaryUpdates []string `json:"pendingSummaryUpdates,omitempty" bson:"pendingSummaryUpdates,omitempty"`
	PendingSummaryDeletes []string `json:"pendingSummaryDeletes,omitempty" bson:"pendingSummaryDeletes,omitempty"`
}

func (m *Metadata) Parse(parser structure.ObjectParser) {
	m.Metadata.Parse(parser)
	if ptr := parser.StringArray(MetadataKeyPendingSummaryUpdates); ptr != nil {
		m.PendingSummaryUpdates = *ptr
	}
	if ptr := parser.StringArray(MetadataKeyPendingSummaryDeletes); ptr != nil {
		m.PendingSummaryDeletes = *ptr
	}
}

func (m *Metadata) Validate(validator structure.Validator) {
	m.Metadata.Validate(validator)
	validator.StringArray(MetadataKeyPendingSummaryUpdates, &m.PendingSummaryUpdates).
		EachOneOf(summaryTypes.SummaryTypeCGM, summaryTypes.SummaryTypeBGM).EachUnique()
	validator.StringArray(MetadataKeyPendingSummaryDeletes, &m.PendingSummaryDeletes).
		EachNotEmpty().EachUnique()
}

// DeduplicationID returns the deduplication id of the work that recalculates the summaries
// of the user for the recalculation with the given id. A user has at most one such work for
// each recalculation, so requesting it again while it's pending creates no more work.
func DeduplicationID(recalculationID string, userID string) string {
	return fmt.Sprintf("%s:%s", recalculationID, userID)
}

// NewWorkCreate returns the work to recalculate the summaries of the user, for the
// recalculation with the given id.
//
// The work has the serial id of the postprocess work of the user, so that the summaries of
// a user aren't recalculated while they're being updated for a change to the user's data.
func NewWorkCreate(recalculationID string, userID string) (*work.Create, error) {
	if recalculationID == "" {
		return nil, errors.New("recalculation id is missing")
	}
	if userID == "" {
		return nil, errors.New("user id is missing")
	}

	create, err := metadata.WithMetadata(
		&work.Create{
			Type:                    Type,
			DeduplicationID:         pointer.From(DeduplicationID(recalculationID, userID)),
			SerialID:                pointer.From(dataWorkPostprocess.IDFromUserID(userID)),
			ProcessingAvailableTime: time.Now(),
			ProcessingTimeout:       int(ProcessingTimeout.Seconds()),
		},
		&Metadata{
			Metadata: userWork.Metadata{UserID: pointer.From(userID)},
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create work create")
	}
	return create, nil
}
