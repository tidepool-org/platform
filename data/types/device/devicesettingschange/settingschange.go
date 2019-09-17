package devicesettingschange

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
)

type SettingsChange struct {
	From *string `json:"from,omitempty" bson:"from,omitempty"`
	To   *string `json:"to,omitempty" bson:"to,omitempty"`
}

func ParseSettingsChange(parser structure.ObjectParser) *SettingsChange {
	if !parser.Exists() {
		return nil
	}
	datum := NewSettingsChange()
	parser.Parse(datum)
	return datum
}

func NewSettingsChange() *SettingsChange {
	return &SettingsChange{}
}

func (a *SettingsChange) Parse(parser structure.ObjectParser) {
	a.From = parser.String("from")
	a.To = parser.String("to")
}

func (a *SettingsChange) Validate(validator structure.Validator) {
}

func (a *SettingsChange) Normalize(normalizer data.Normalizer) {}
