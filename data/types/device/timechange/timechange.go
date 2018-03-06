package timechange

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/device"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	SubType = "timeChange" // TODO: Rename Type to "device/timeChange"; remove SubType
)

type TimeChange struct {
	device.Device `bson:",inline"`

	Change *Change `json:"change,omitempty" bson:"change,omitempty"`
}

func NewDatum() data.Datum {
	return New()
}

func New() *TimeChange {
	return &TimeChange{}
}

func Init() *TimeChange {
	timeChange := New()
	timeChange.Init()
	return timeChange
}

func (t *TimeChange) Init() {
	t.Device.Init()
	t.SubType = SubType

	t.Change = nil
}

func (t *TimeChange) Parse(parser data.ObjectParser) error {
	if err := t.Device.Parse(parser); err != nil {
		return err
	}

	t.Change = ParseChange(parser.NewChildObjectParser("change"))

	return nil
}

func (t *TimeChange) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(t.Meta())
	}

	t.Device.Validate(validator)

	if t.SubType != "" {
		validator.String("subType", &t.SubType).EqualTo(SubType)
	}

	changeValidator := validator.WithReference("change")
	if t.Change != nil {
		t.Change.Validate(changeValidator)
	} else {
		changeValidator.ReportError(structureValidator.ErrorValueNotExists())
	}
}

func (t *TimeChange) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(t.Meta())
	}

	t.Device.Normalize(normalizer)

	if t.Change != nil {
		t.Change.Normalize(normalizer.WithReference("change"))
	}
}
