package reservoirchange

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/device"
	"github.com/tidepool-org/platform/data/types/device/status"
	"github.com/tidepool-org/platform/service"
)

type ReservoirChange struct {
	device.Device `bson:",inline"`

	StatusID *string `json:"status,omitempty" bson:"status,omitempty"`

	// Embedded status
	status *data.Datum
}

func SubType() string {
	return "reservoirChange"
}

func NewDatum() data.Datum {
	return New()
}

func New() *ReservoirChange {
	return &ReservoirChange{}
}

func Init() *ReservoirChange {
	reservoirChange := New()
	reservoirChange.Init()
	return reservoirChange
}

func (r *ReservoirChange) Init() {
	r.Device.Init()
	r.SubType = SubType()

	r.StatusID = nil

	r.status = nil
}

func (r *ReservoirChange) Parse(parser data.ObjectParser) error {
	if err := r.Device.Parse(parser); err != nil {
		return err
	}

	// TODO: This is a bit hacky to ensure we only parse true status data. Is there a better way?

	if statusParser := parser.NewChildObjectParser("status"); statusParser.Object() != nil {
		if statusType := statusParser.ParseString("type"); statusType == nil {
			statusParser.AppendError("type", service.ErrorValueNotExists())
		} else if *statusType != device.Type() {
			statusParser.AppendError("type", service.ErrorValueStringNotOneOf(*statusType, []string{device.Type()}))
		} else if statusSubType := statusParser.ParseString("subType"); statusSubType == nil {
			statusParser.AppendError("subType", service.ErrorValueNotExists())
		} else if *statusSubType != status.SubType() {
			statusParser.AppendError("subType", service.ErrorValueStringNotOneOf(*statusSubType, []string{status.SubType()}))
		} else {
			r.status = parser.ParseDatum("status")
		}
	}

	return nil
}

func (r *ReservoirChange) Validate(validator data.Validator) error {
	if err := r.Device.Validate(validator); err != nil {
		return err
	}

	validator.ValidateString("subType", &r.SubType).EqualTo(SubType())

	if r.status != nil {
		(*r.status).Validate(validator.NewChildValidator("status"))
	}

	return nil
}

func (r *ReservoirChange) Normalize(normalizer data.Normalizer) error {
	if err := r.Device.Normalize(normalizer); err != nil {
		return err
	}

	if r.status != nil {
		if err := (*r.status).Normalize(normalizer.NewChildNormalizer("status")); err != nil {
			return err
		}

		r.StatusID = &(*r.status).(*status.Status).ID

		normalizer.AppendDatum(*r.status)
		r.status = nil
	}

	return nil
}
