package work

import (
	"fmt"

	ouraWebhook "github.com/tidepool-org/platform/oura/webhook"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/times"
)

const (
	Domain = "org.tidepool.oura.data"

	MetadataKeyScope = "scope"
)

type (
	TimeRangeMetadata = times.TimeRangeMetadata
	EventMetadata     = ouraWebhook.EventMetadata
)

type Metadata struct {
	Scope             *[]string `json:"scope,omitempty" bson:"scope,omitempty"`
	TimeRangeMetadata `bson:",inline"`
	EventMetadata     `bson:",inline"`
}

func (m *Metadata) Parse(parser structure.ObjectParser) {
	m.Scope = parser.StringArray(MetadataKeyScope)
	m.TimeRangeMetadata.Parse(parser)
	m.EventMetadata.Parse(parser)
}

func (m *Metadata) Validate(validator structure.Validator) {
	validator.StringArray(MetadataKeyScope, m.Scope).NotEmpty()
	m.TimeRangeMetadata.Validate(validator)
	m.EventMetadata.Validate(validator)
	if (m.TimeRange == nil) == (m.Event == nil) {
		validator.ReportError(structureValidator.ErrorValuesNotExistForOne(ouraWebhook.MetadataKeyEvent, times.MetadataKeyTimeRange))
	}
}

func SerialIDFromProviderSessionID(providerSessionID string) string {
	return fmt.Sprintf("%s:%s", Domain, providerSessionID)
}
