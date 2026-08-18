package postprocess

import (
	"fmt"
	"time"

	mapset "github.com/deckarep/golang-set/v2"

	"github.com/tidepool-org/platform/structure"
	userWork "github.com/tidepool-org/platform/user/work"
)

const (
	Type = "org.tidepool.data.upload.postprocess"

	ProcessingTimeout = 5 * time.Minute
)

const (
	ReasonDataAdded = "DATA_ADDED"

	// ReasonUploadCompleted reports a data set was closed, or jellyfish uploaded a partial batch
	ReasonUploadCompleted = "UPLOAD_COMPLETED"

	// ReasonLegacyDataAdded reports jellyfish uploaded a full batch
	ReasonLegacyDataAdded = "LEGACY_DATA_ADDED"

	ReasonSchemaMigration = "SCHEMA_MIGRATION"
)

const (
	MetadataKeyReasons = "reasons"
)

func Reasons() []string {
	return []string{
		ReasonDataAdded,
		ReasonUploadCompleted,
		ReasonLegacyDataAdded,
		ReasonSchemaMigration,
	}
}

// Data added is intentionally absent to prevent continuous data uploads from constantly triggering syncs
var ehrSyncReasons = mapset.NewSet(
	ReasonUploadCompleted,
	ReasonLegacyDataAdded,
)

func TriggersEHRSync(reasons []string) bool {
	return ehrSyncReasons.ContainsAny(reasons...)
}

// Defer summary recalculation if jellyfish uploaded a full batch
var deferrableReasons = mapset.NewSet(
	ReasonLegacyDataAdded,
)

// shouldDefer reports whether reasons contains only deferrable reasons
func shouldDefer(reasons []string) bool {
	if len(reasons) == 0 {
		return false
	}
	return mapset.NewSet(reasons...).IsSubset(deferrableReasons)
}

// IDFromUserID returns both the serial id, which prevents the work of a user being processed
// concurrently, and the group id, which is the scope the work of a user is coalesced within. Both
// are the user, and neither is the data set, so that changes to any number of data sets, interleaved
// in any order, yield a single stream of work for the user.
func IDFromUserID(userID string) string {
	return fmt.Sprintf("%s:%s", Type, userID)
}

type Metadata struct {
	userWork.Metadata `bson:",inline"`
	Reasons           []string `json:"reasons,omitempty" bson:"reasons,omitempty"`
}

func (m *Metadata) Parse(parser structure.ObjectParser) {
	m.Metadata.Parse(parser)
	if ptr := parser.StringArray(MetadataKeyReasons); ptr != nil {
		m.Reasons = *ptr
	}
}

func (m *Metadata) Validate(validator structure.Validator) {
	m.Metadata.Validate(validator)
	validator.StringArray(MetadataKeyReasons, &m.Reasons).NotEmpty().EachOneOf(Reasons()...).EachUnique()
}
