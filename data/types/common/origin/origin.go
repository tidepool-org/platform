package origin

import (
	"time"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	IDLengthMaximum      = 100
	NameLengthMaximum    = 100
	TimeFormat           = time.RFC3339
	TypeDevice           = "device"
	TypeManual           = "manual"
	TypeService          = "service"
	VersionLengthMaximum = 100
)

func Types() []string {
	return []string{
		TypeDevice,
		TypeManual,
		TypeService,
	}
}

type Origin struct {
	ID      *string    `json:"id,omitempty" bson:"id,omitempty"`
	Name    *string    `json:"name,omitempty" bson:"name,omitempty"`
	Payload *data.Blob `json:"payload,omitempty" bson:"payload,omitempty"`
	Time    *time.Time `json:"time,omitempty" bson:"time,omitempty"`
	Type    *string    `json:"type,omitempty" bson:"type,omitempty"`
	Version *string    `json:"version,omitempty" bson:"version,omitempty"`
}

func ParseOrigin(parser data.ObjectParser) *Origin {
	if parser.Object() == nil {
		return nil
	}
	datum := NewOrigin()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewOrigin() *Origin {
	return &Origin{}
}

func (o *Origin) Parse(parser data.ObjectParser) {
	o.ID = parser.ParseString("id")
	o.Name = parser.ParseString("name")
	o.Payload = data.ParseBlob(parser.NewChildObjectParser("payload"))
	o.Time = parser.ParseTime("time", TimeFormat)
	o.Type = parser.ParseString("type")
	o.Version = parser.ParseString("version")
}

func (o *Origin) Validate(validator structure.Validator) {
	validator.String("id", o.ID).NotEmpty().LengthLessThanOrEqualTo(IDLengthMaximum)
	validator.String("name", o.Name).NotEmpty().LengthLessThanOrEqualTo(NameLengthMaximum)
	if o.Payload != nil {
		o.Payload.Validate(validator.WithReference("payload"))
	}
	validator.Time("time", o.Time).NotZero()
	validator.String("type", o.Type).OneOf(Types()...)
	validator.String("version", o.Version).NotEmpty().LengthLessThanOrEqualTo(VersionLengthMaximum)
}

func (o *Origin) Normalize(normalizer data.Normalizer) {
	if o.Payload != nil {
		o.Payload.Normalize(normalizer.WithReference("payload"))
	}
}
