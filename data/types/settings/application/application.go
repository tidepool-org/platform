package application

import (
	"github.com/tidepool-org/platform/data"
	dataTypes "github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/structure"
)

const (
	Type = "applicationSettings"

	NameLengthMaximum    = 1000
	NameLengthMinimum    = 1
	VersionLengthMaximum = 1000
	VersionLengthMinimum = 1
)

type Application struct {
	dataTypes.Base `bson:",inline"`

	Name    *string `json:"name,omitempty" bson:"name,omitempty"`
	Version *string `json:"version,omitempty" bson:"version,omitempty"`
}

func New() *Application {
	return &Application{
		Base: dataTypes.New(Type),
	}
}

func (a *Application) Parse(parser structure.ObjectParser) {
	if !parser.HasMeta() {
		parser = parser.WithMeta(a.Meta())
	}

	a.Base.Parse(parser)

	a.Name = parser.String("name")
	a.Version = parser.String("version")
}

func (a *Application) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(a.Meta())
	}

	a.Base.Validate(validator)

	if a.Type != "" {
		validator.String("type", &a.Type).EqualTo(Type)
	}

	validator.String("name", a.Name).Exists().LengthInRange(NameLengthMinimum, NameLengthMaximum)
	validator.String("version", a.Version).Exists().LengthInRange(VersionLengthMinimum, VersionLengthMaximum)
}

func (a *Application) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(a.Meta())
	}

	a.Base.Normalize(normalizer)
}
