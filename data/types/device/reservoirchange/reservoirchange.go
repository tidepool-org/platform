package reservoirchange

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/device"
	dataTypesDeviceStatus "github.com/tidepool-org/platform/data/types/device/status"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	SubType = "reservoirChange" // TODO: Rename Type to "device/reservoirChange"; remove SubType
)

type ReservoirChange struct {
	device.Device `bson:",inline"`

	Status   *data.Datum `json:"-" bson:"-"`
	StatusID *string     `json:"status,omitempty" bson:"status,omitempty"`
}

func New() *ReservoirChange {
	return &ReservoirChange{
		Device: device.New(SubType),
	}
}

func (r *ReservoirChange) Parse(parser data.ObjectParser) error {
	if err := r.Device.Parse(parser); err != nil {
		return err
	}

	r.Status = dataTypesDeviceStatus.ParseStatusDatum(parser.NewChildObjectParser("status"))

	return nil
}

func (r *ReservoirChange) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(r.Meta())
	}

	r.Device.Validate(validator)

	if r.SubType != "" {
		validator.String("subType", &r.SubType).EqualTo(SubType)
	}

	if validator.Origin() == structure.OriginExternal {
		if r.Status != nil {
			(*r.Status).Validate(validator.WithReference("status"))
		}
		validator.String("statusId", r.StatusID).NotExists()
	} else {
		if r.Status != nil {
			validator.WithReference("status").ReportError(structureValidator.ErrorValueExists())
		}
		validator.String("statusId", r.StatusID).Using(data.IDValidator)
	}
}

func (r *ReservoirChange) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(r.Meta())
	}

	r.Device.Normalize(normalizer)

	if r.Status != nil {
		(*r.Status).Normalize(normalizer.WithReference("status"))
	}

	if normalizer.Origin() == structure.OriginExternal {
		if r.Status != nil {
			normalizer.AddData(*r.Status)
			switch status := (*r.Status).(type) {
			case *dataTypesDeviceStatus.Status:
				r.StatusID = status.ID
			}
			r.Status = nil
		}
	}
}
