package dosingdecision

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/structure"
)

const (
	Type     = "dosingDecision"
	Aid      = "aid"
	Cgm      = "cgm"
	Pump     = "pump"
	SmartPen = "smartpen"
)

func DeviceTypes() []string {
	return []string{Aid, Cgm, Pump, SmartPen}
}

type DosingDecision struct {
	types.Base `bson:",inline"`

	CarbsOnBoard               *CarbsOnBoard                      `json:"carbsOnBoard,omitempty" bson:"carbsOnBoard,omitempty"`
	Device                     *Device                            `json:"device,omitempty" bson:"device,omitempty"`
	GlucoseTargetRangeSchedule *pump.BloodGlucoseTargetStartArray `json:"glucoseTargetRangeSchedule,omitempty" bson:"glucoseTargetRangeSchedule,omitempty"`
	RecommendedBasal           *RecommendedBasal                  `json:"recommendedBasal,omitempty" bson:"recommendedBasal,omitempty"`
	Units                      *pump.Units                        `json:"units,omitempty" bson:"units,omitempty"`
}

func New() *DosingDecision {
	return &DosingDecision{
		Base: types.New(Type),
	}
}

func ParseDosingDecision(parser structure.ObjectParser) *DosingDecision {
	if !parser.Exists() {
		return nil
	}
	datum := NewDosingDecision()
	parser.Parse(datum)
	return datum
}

func NewDosingDecision() *DosingDecision {
	return &DosingDecision{}
}

func (a *DosingDecision) Parse(parser structure.ObjectParser) {
	if !parser.HasMeta() {
		parser = parser.WithMeta(a.Meta())
	}

	a.Base.Parse(parser)

	a.Device = ParseDevice(parser.WithReferenceObjectParser("device"))
	a.CarbsOnBoard = ParseCarbsOnBoard(parser.WithReferenceObjectParser("carbsOnBoard"))
	a.RecommendedBasal = ParseRecommendedBasal(parser.WithReferenceObjectParser("recommendedBasal"))
	a.GlucoseTargetRangeSchedule = pump.ParseBloodGlucoseTargetStartArray(parser.WithReferenceArrayParser("glucoseTargetRangeSchedule"))
	a.Units = pump.ParseUnits(parser.WithReferenceObjectParser("units"))

}

func (a *DosingDecision) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(a.Meta())
	}

	a.Base.Validate(validator)

	if a.Type != "" {
		validator.String("type", &a.Type).EqualTo(Type)
	}

	var unitsBloodGlucose *string
	if a.Units != nil {
		unitsBloodGlucose = a.Units.BloodGlucose
	}

	if a.CarbsOnBoard != nil {
		a.CarbsOnBoard.Validate(validator.WithReference(("carbsOnBoard")))
	}
	if a.RecommendedBasal != nil {
		a.RecommendedBasal.Validate(validator.WithReference(("recommendedBasal")))
	}
	if a.GlucoseTargetRangeSchedule != nil {
		a.GlucoseTargetRangeSchedule.Validate(validator.WithReference(("pumpManagerStatus")), unitsBloodGlucose)
	}
	if a.Units != nil {
		a.Units.Validate(validator.WithReference("units"))
	}
	if a.Device != nil {
		a.Device.Validate(validator.WithReference("device"))
	}
}

func (a *DosingDecision) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(a.Meta())
	}

	a.Base.Normalize(normalizer)

	var unitsBloodGlucose *string
	if a.Units != nil {
		unitsBloodGlucose = a.Units.BloodGlucose
	}
	a.CarbsOnBoard.Normalize(normalizer)
	a.RecommendedBasal.Normalize(normalizer)
	a.GlucoseTargetRangeSchedule.Normalize(normalizer, unitsBloodGlucose)
	a.Units.Normalize(normalizer)
}
