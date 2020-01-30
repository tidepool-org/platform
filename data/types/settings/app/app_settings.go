package app_settings

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

const (
	MinNameLength = 1
	MaxNameLength = 1000

	MinVersionLength = 1
	MaxVersionLength = 1000
)

type AppSettings struct {
	Name           *string `json:"name,omitempty" bson:"name,omitempty"`
	LoopAppVersion *string `json:"loopAppVersion,omitempty" bson:"loopAppVersion,omitempty"`
}

func ParseAppSettings(parser structure.ObjectParser) *AppSettings {
	if !parser.Exists() {
		return nil
	}
	datum := NewAppSettings()
	parser.Parse(datum)
	return datum
}

func NewAppSettings() *AppSettings {
	return &AppSettings{}
}

func (f *AppSettings) Parse(parser structure.ObjectParser) {
	f.Name = parser.String("name")
	f.LoopAppVersion = parser.String("loopAppVersion")
}

func (f *AppSettings) Validate(validator structure.Validator) {
	validator.String("name", f.Name).Exists().LengthInRange(MinNameLength, MaxNameLength)
	validator.String("loopAppVersion", f.LoopAppVersion).Exists().LengthInRange(MinVersionLength, MaxVersionLength)
}

func (f *AppSettings) Normalize(normalizer data.Normalizer) {}
