package reservoirchange

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/device"
	"github.com/tidepool-org/platform/data/types/device/status"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	SubType = "reservoirChange" // TODO: Rename Type to "device/reservoirChange"; remove SubType
)

type ReservoirChange struct {
	device.Device `bson:",inline"`

	Status   *status.Status `json:"-" bson:"-"`
	StatusID *string        `json:"status,omitempty" bson:"status,omitempty"`
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
	r.SubType = SubType

	r.Status = nil
	r.StatusID = nil
}

func (r *ReservoirChange) Parse(parser data.ObjectParser) error {
	if err := r.Device.Parse(parser); err != nil {
		return err
	}

	// TODO: This is a bit hacky to ensure we only parse true status data. Is there a better way?

	if statusParser := parser.NewChildObjectParser("status"); statusParser.Object() != nil {
		if statusType := statusParser.ParseString("type"); statusType == nil {
			statusParser.AppendError("type", service.ErrorValueNotExists())
		} else if *statusType != device.Type {
			statusParser.AppendError("type", service.ErrorValueStringNotOneOf(*statusType, []string{device.Type}))
		} else if statusSubType := statusParser.ParseString("subType"); statusSubType == nil {
			statusParser.AppendError("subType", service.ErrorValueNotExists())
		} else if *statusSubType != status.SubType {
			statusParser.AppendError("subType", service.ErrorValueStringNotOneOf(*statusSubType, []string{status.SubType}))
		} else if datum := parser.ParseDatum("status"); datum != nil {
			r.Status = (*datum).(*status.Status)
		}
	}

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
			r.Status.Validate(validator.WithReference("status"))
		}
		validator.String("statusId", r.StatusID).NotExists()
	} else {
		if r.Status != nil {
			validator.WithReference("status").ReportError(structureValidator.ErrorValueExists())
		}
		validator.String("statusId", r.StatusID).Using(id.Validate)
	}
}

func (r *ReservoirChange) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(r.Meta())
	}

	r.Device.Normalize(normalizer)

	if r.Status != nil {
		r.Status.Normalize(normalizer.WithReference("status"))
	}

	if normalizer.Origin() == structure.OriginExternal {
		if r.Status != nil {
			normalizer.AddData(r.Status)
			r.StatusID = r.Status.ID
			r.Status = nil
		}
	}
}
