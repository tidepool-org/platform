package timechange

import (
	"time"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	timeZone "github.com/tidepool-org/platform/time/zone"
)

const (
	InfoTimeFormat = "2006-01-02T15:04:05"
)

type Info struct {
	Time         *time.Time `json:"time,omitempty" bson:"time,omitempty"`
	TimeZoneName *string    `json:"timeZoneName,omitempty" bson:"timeZoneName,omitempty"`
}

func ParseInfo(parser structure.ObjectParser) *Info {
	if !parser.Exists() {
		return nil
	}
	datum := NewInfo()
	parser.Parse(datum)
	return datum
}

func NewInfo() *Info {
	return &Info{}
}

func (i *Info) Parse(parser structure.ObjectParser) {
	i.Time = parser.Time("time", InfoTimeFormat)
	i.TimeZoneName = parser.String("timeZoneName")
}

func (i *Info) Validate(validator structure.Validator) {
	validator.Time("time", i.Time).NotZero()
	validator.String("timeZoneName", i.TimeZoneName).Using(timeZone.NameValidator)
	if i.Time == nil && i.TimeZoneName == nil {
		validator.ReportError(structureValidator.ErrorValuesNotExistForAny("time", "timeZoneName"))
	}
}

func (i *Info) Normalize(normalizer data.Normalizer) {}
