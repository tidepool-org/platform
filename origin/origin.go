package origin

import (
	"time"

	"github.com/tidepool-org/platform/metadata"
	"github.com/tidepool-org/platform/structure"
)

const (
	IDLengthMaximum      = 100
	NameLengthMaximum    = 100
	TimeFormat           = time.RFC3339Nano
	TypeApplication      = "application"
	TypeDevice           = "device"
	TypeManual           = "manual"
	TypeService          = "service"
	VersionLengthMaximum = 100
)

func Types() []string {
	return []string{
		TypeApplication,
		TypeDevice,
		TypeManual,
		TypeService,
	}
}

type Origin struct {
	ID      *string            `json:"id,omitempty" bson:"id,omitempty"`
	Name    *string            `json:"name,omitempty" bson:"name,omitempty"`
	Payload *metadata.Metadata `json:"payload,omitempty" bson:"payload,omitempty"`
	Time    *string            `json:"time,omitempty" bson:"time,omitempty"`
	Type    *string            `json:"type,omitempty" bson:"type,omitempty"`
	Version *string            `json:"version,omitempty" bson:"version,omitempty"`
}

func ParseOrigin(parser structure.ObjectParser) *Origin {
	if !parser.Exists() {
		return nil
	}
	datum := NewOrigin()
	parser.Parse(datum)
	return datum
}

func NewOrigin() *Origin {
	return &Origin{}
}

func (o *Origin) Parse(parser structure.ObjectParser) {
	o.ID = parser.String("id")
	o.Name = parser.String("name")
	o.Payload = metadata.ParseMetadata(parser.WithReferenceObjectParser("payload"))
	o.Time = parser.String("time")
	o.Type = parser.String("type")
	o.Version = parser.String("version")
}

func (o *Origin) Validate(validator structure.Validator) {
	validator.String("id", o.ID).NotEmpty().LengthLessThanOrEqualTo(IDLengthMaximum)
	validator.String("name", o.Name).NotEmpty().LengthLessThanOrEqualTo(NameLengthMaximum)
	if o.Payload != nil {
		o.Payload.Validate(validator.WithReference("payload"))
	}
	validator.String("time", o.Time).AsTime(TimeFormat).NotZero()
	validator.String("type", o.Type).OneOf(Types()...)
	validator.String("version", o.Version).NotEmpty().LengthLessThanOrEqualTo(VersionLengthMaximum)
}
