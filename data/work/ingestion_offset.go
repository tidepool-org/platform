package work

import "github.com/tidepool-org/platform/structure"

const MetadataKeyIngestionOffset = "ingestionOffset"

type IngestionOffsetMetadata struct {
	IngestionOffset *int `json:"ingestionOffset,omitempty" bson:"ingestionOffset,omitempty"`
}

func (i *IngestionOffsetMetadata) Parse(parser structure.ObjectParser) {
	i.IngestionOffset = parser.Int(MetadataKeyIngestionOffset)
}

func (i *IngestionOffsetMetadata) Validate(validator structure.Validator) {
	validator.Int(MetadataKeyIngestionOffset, i.IngestionOffset).GreaterThanOrEqualTo(0)
}
