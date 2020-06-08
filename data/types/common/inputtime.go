package common

import (
	"time"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
)

type InputTime struct {
	InputTime *string `json:"inputTime,omitempty" bson:"inputTime,omitempty"`
}

func NewInputTime() *InputTime {
	return &InputTime{}
}

func (i *InputTime) Parse(parser structure.ObjectParser) {
	i.InputTime = parser.String("inputTime")
}

func (i *InputTime) Validate(validator structure.Validator) {
	timeValidator := validator.String("inputTime", i.InputTime)
	timeValidator.AsTime(types.TimeFormat)
}

func (i *InputTime) Normalize(normalizer data.Normalizer) {
	if i.InputTime != nil && *i.InputTime != "" {
		parsedTime, err := time.Parse(types.TimeFormat, *i.InputTime)
		if err != nil {
			parsedTime, err = time.Parse(types.ParsingTimeFormat, *i.InputTime)
		}
		if err == nil {
			utcTimeString := parsedTime.UTC().Format(types.TimeFormat)
			// Time field is not well formatted in UTC
			if utcTimeString != *i.InputTime {
				i.InputTime = pointer.FromString(utcTimeString)
			}
		}
	}
}
