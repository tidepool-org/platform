package data

import (
	"github.com/tidepool-org/platform/oura"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/times"
)

const MetadataKeyDataType = "dataType"

type (
	EventMetadata     = oura.EventMetadata
	TimeRangeMetadata = times.TimeRangeMetadata
)

type Metadata struct {
	DataType          string `json:"dataType,omitempty" bson:"dataType,omitempty"`
	EventMetadata     `bson:",inline"`
	TimeRangeMetadata `bson:",inline"`
}

func (m *Metadata) Parse(parser structure.ObjectParser) {
	if ptr := parser.String(MetadataKeyDataType); ptr != nil {
		m.DataType = *ptr
	}
	m.EventMetadata.Parse(parser)
	m.TimeRangeMetadata.Parse(parser)
}

func (m *Metadata) Validate(validator structure.Validator) {
	validator.String(MetadataKeyDataType, &m.DataType).OneOf(oura.DataTypes()...)
	m.EventMetadata.Validate(validator)
	m.TimeRangeMetadata.Validate(validator)
	if (m.Event == nil) == (m.TimeRange == nil) {
		validator.ReportError(structureValidator.ErrorValuesNotExistForOne(oura.MetadataKeyEvent, times.MetadataKeyTimeRange))
	}
}
