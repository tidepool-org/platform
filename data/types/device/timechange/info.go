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

func ParseInfo(parser data.ObjectParser) *Info {
	if parser.Object() == nil {
		return nil
	}
	datum := NewInfo()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewInfo() *Info {
	return &Info{}
}

func (i *Info) Parse(parser data.ObjectParser) {
	i.Time = parser.ParseTime("time", InfoTimeFormat)
	i.TimeZoneName = parser.ParseString("timeZoneName")
}

func (i *Info) Validate(validator structure.Validator) {
	validator.Time("time", i.Time).NotZero()
	validator.String("timeZoneName", i.TimeZoneName).Using(timeZone.NameValidator)
	if i.Time == nil && i.TimeZoneName == nil {
		validator.ReportError(structureValidator.ErrorValuesNotExistForAny("time", "timeZoneName"))
	}
}

func (i *Info) Normalize(normalizer data.Normalizer) {}
