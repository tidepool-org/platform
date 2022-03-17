package physical

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/common"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	Type = "physicalActivity" // TODO: Change to "activity/physical"

	ActivityTypeAmericanFootball              = "americanFootball"
	ActivityTypeArchery                       = "archery"
	ActivityTypeAustralianFootball            = "australianFootball"
	ActivityTypeBadminton                     = "badminton"
	ActivityTypeBarre                         = "barre"
	ActivityTypeBaseball                      = "baseball"
	ActivityTypeBasketball                    = "basketball"
	ActivityTypeBowling                       = "bowling"
	ActivityTypeBoxing                        = "boxing"
	ActivityTypeClimbing                      = "climbing"
	ActivityTypeCoreTraining                  = "coreTraining"
	ActivityTypeCricket                       = "cricket"
	ActivityTypeCrossCountrySkiing            = "crossCountrySkiing"
	ActivityTypeCrossTraining                 = "crossTraining"
	ActivityTypeCurling                       = "curling"
	ActivityTypeCycling                       = "cycling"
	ActivityTypeDance                         = "dance"
	ActivityTypeDanceInspiredTraining         = "danceInspiredTraining"
	ActivityTypeDownhillSkiing                = "downhillSkiing"
	ActivityTypeElliptical                    = "elliptical"
	ActivityTypeEquestrianSports              = "equestrianSports"
	ActivityTypeFencing                       = "fencing"
	ActivityTypeFishing                       = "fishing"
	ActivityTypeFlexibility                   = "flexibility"
	ActivityTypeFunctionalStrengthTraining    = "functionalStrengthTraining"
	ActivityTypeGolf                          = "golf"
	ActivityTypeGymnastics                    = "gymnastics"
	ActivityTypeHandball                      = "handball"
	ActivityTypeHandCycling                   = "handCycling"
	ActivityTypeHighIntensityIntervalTraining = "highIntensityIntervalTraining"
	ActivityTypeHiking                        = "hiking"
	ActivityTypeHockey                        = "hockey"
	ActivityTypeHunting                       = "hunting"
	ActivityTypeJumpRope                      = "jumpRope"
	ActivityTypeKickboxing                    = "kickboxing"
	ActivityTypeLacrosse                      = "lacrosse"
	ActivityTypeMartialArts                   = "martialArts"
	ActivityTypeMindAndBody                   = "mindAndBody"
	ActivityTypeMixedCardio                   = "mixedCardio"
	ActivityTypeMixedMetabolicCardioTraining  = "mixedMetabolicCardioTraining"
	ActivityTypeOther                         = "other"
	ActivityTypeOtherLengthMaximum            = 100
	ActivityTypePaddleSports                  = "paddleSports"
	ActivityTypePilates                       = "pilates"
	ActivityTypePlay                          = "play"
	ActivityTypePreparationAndRecovery        = "preparationAndRecovery"
	ActivityTypeRacquetball                   = "racquetball"
	ActivityTypeRowing                        = "rowing"
	ActivityTypeRugby                         = "rugby"
	ActivityTypeRunning                       = "running"
	ActivityTypeSailing                       = "sailing"
	ActivityTypeSkatingSports                 = "skatingSports"
	ActivityTypeSnowboarding                  = "snowboarding"
	ActivityTypeSnowSports                    = "snowSports"
	ActivityTypeSoccer                        = "soccer"
	ActivityTypeSoftball                      = "softball"
	ActivityTypeSquash                        = "squash"
	ActivityTypeStairClimbing                 = "stairClimbing"
	ActivityTypeStairs                        = "stairs"
	ActivityTypeStepTraining                  = "stepTraining"
	ActivityTypeSurfingSports                 = "surfingSports"
	ActivityTypeSwimming                      = "swimming"
	ActivityTypeTableTennis                   = "tableTennis"
	ActivityTypeTaiChi                        = "taiChi"
	ActivityTypeTennis                        = "tennis"
	ActivityTypeTrackAndField                 = "trackAndField"
	ActivityTypeTraditionalStrengthTraining   = "traditionalStrengthTraining"
	ActivityTypeVolleyball                    = "volleyball"
	ActivityTypeWalking                       = "walking"
	ActivityTypeWaterFitness                  = "waterFitness"
	ActivityTypeWaterPolo                     = "waterPolo"
	ActivityTypeWaterSports                   = "waterSports"
	ActivityTypeWheelchairRunPace             = "wheelchairRunPace"
	ActivityTypeWheelchairWalkPace            = "wheelchairWalkPace"
	ActivityTypeWrestling                     = "wrestling"
	ActivityTypeYoga                          = "yoga"
	NameLengthMaximum                         = 100
	ReportedIntensityHigh                     = "high"
	ReportedIntensityLow                      = "low"
	ReportedIntensityMedium                   = "medium"
)

// Note: ActivityTypes from Apple HealthKit HKWorkoutActivityType

func ActivityTypes() []string {
	return []string{
		ActivityTypeAmericanFootball,
		ActivityTypeArchery,
		ActivityTypeAustralianFootball,
		ActivityTypeBadminton,
		ActivityTypeBarre,
		ActivityTypeBaseball,
		ActivityTypeBasketball,
		ActivityTypeBowling,
		ActivityTypeBoxing,
		ActivityTypeClimbing,
		ActivityTypeCoreTraining,
		ActivityTypeCricket,
		ActivityTypeCrossCountrySkiing,
		ActivityTypeCrossTraining,
		ActivityTypeCurling,
		ActivityTypeCycling,
		ActivityTypeDance,
		ActivityTypeDanceInspiredTraining,
		ActivityTypeDownhillSkiing,
		ActivityTypeElliptical,
		ActivityTypeEquestrianSports,
		ActivityTypeFencing,
		ActivityTypeFishing,
		ActivityTypeFlexibility,
		ActivityTypeFunctionalStrengthTraining,
		ActivityTypeGolf,
		ActivityTypeGymnastics,
		ActivityTypeHandball,
		ActivityTypeHandCycling,
		ActivityTypeHighIntensityIntervalTraining,
		ActivityTypeHiking,
		ActivityTypeHockey,
		ActivityTypeHunting,
		ActivityTypeJumpRope,
		ActivityTypeKickboxing,
		ActivityTypeLacrosse,
		ActivityTypeMartialArts,
		ActivityTypeMindAndBody,
		ActivityTypeMixedCardio,
		ActivityTypeMixedMetabolicCardioTraining,
		ActivityTypeOther,
		ActivityTypePaddleSports,
		ActivityTypePilates,
		ActivityTypePlay,
		ActivityTypePreparationAndRecovery,
		ActivityTypeRacquetball,
		ActivityTypeRowing,
		ActivityTypeRugby,
		ActivityTypeRunning,
		ActivityTypeSailing,
		ActivityTypeSkatingSports,
		ActivityTypeSnowboarding,
		ActivityTypeSnowSports,
		ActivityTypeSoccer,
		ActivityTypeSoftball,
		ActivityTypeSquash,
		ActivityTypeStairClimbing,
		ActivityTypeStairs,
		ActivityTypeStepTraining,
		ActivityTypeSurfingSports,
		ActivityTypeSwimming,
		ActivityTypeTableTennis,
		ActivityTypeTaiChi,
		ActivityTypeTennis,
		ActivityTypeTrackAndField,
		ActivityTypeTraditionalStrengthTraining,
		ActivityTypeVolleyball,
		ActivityTypeWalking,
		ActivityTypeWaterFitness,
		ActivityTypeWaterPolo,
		ActivityTypeWaterSports,
		ActivityTypeWheelchairRunPace,
		ActivityTypeWheelchairWalkPace,
		ActivityTypeWrestling,
		ActivityTypeYoga,
	}
}

func ReportedIntensities() []string {
	return []string{
		ReportedIntensityHigh,
		ReportedIntensityLow,
		ReportedIntensityMedium,
	}
}

type Physical struct {
	types.Base `bson:",inline"`

	ActivityType      *string           `json:"activityType,omitempty" bson:"activityType,omitempty"`
	ActivityTypeOther *string           `json:"activityTypeOther,omitempty" bson:"activityTypeOther,omitempty"`
	Aggregate         *bool             `json:"aggregate,omitempty" bson:"aggregate,omitempty"`
	Distance          *Distance         `json:"distance,omitempty" bson:"distance,omitempty"`
	Duration          *common.Duration  `json:"duration,omitempty" bson:"duration,omitempty"`
	ElevationChange   *ElevationChange  `json:"elevationChange,omitempty" bson:"elevationChange,omitempty"`
	Energy            *Energy           `json:"energy,omitempty" bson:"energy,omitempty"`
	Flight            *Flight           `json:"flight,omitempty" bson:"flight,omitempty"`
	Lap               *Lap              `json:"lap,omitempty" bson:"lap,omitempty"`
	Name              *string           `json:"name,omitempty" bson:"name,omitempty"`
	ReportedIntensity *string           `json:"reportedIntensity,omitempty" bson:"reportedIntensity,omitempty"`
	Step              *Step             `json:"step,omitempty" bson:"step,omitempty"`
	InputTime         *common.InputTime `bson:",inline"`
}

func New() *Physical {
	return &Physical{
		Base:      types.New(Type),
		InputTime: common.NewInputTime(),
	}
}

func (p *Physical) Parse(parser structure.ObjectParser) {
	if !parser.HasMeta() {
		parser = parser.WithMeta(p.Meta())
	}

	p.Base.Parse(parser)

	p.ActivityType = parser.String("activityType")
	p.ActivityTypeOther = parser.String("activityTypeOther")
	p.Aggregate = parser.Bool("aggregate")
	p.Distance = ParseDistance(parser.WithReferenceObjectParser("distance"))
	p.Duration = common.ParseDuration(parser.WithReferenceObjectParser("duration"))
	p.ElevationChange = ParseElevationChange(parser.WithReferenceObjectParser("elevationChange"))
	p.Energy = ParseEnergy(parser.WithReferenceObjectParser("energy"))
	p.Flight = ParseFlight(parser.WithReferenceObjectParser("flight"))
	p.Lap = ParseLap(parser.WithReferenceObjectParser("lap"))
	p.Name = parser.String("name")
	p.ReportedIntensity = parser.String("reportedIntensity")
	p.Step = ParseStep(parser.WithReferenceObjectParser("step"))
	p.InputTime.Parse(parser)
}

func (p *Physical) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(p.Meta())
	}

	p.Base.Validate(validator)

	if p.Type != "" {
		validator.String("type", &p.Type).EqualTo(Type)
	}

	validator.String("activityType", p.ActivityType).OneOf(ActivityTypes()...)
	if p.ActivityType != nil && *p.ActivityType == ActivityTypeOther {
		validator.String("activityTypeOther", p.ActivityTypeOther).NotEmpty().LengthLessThanOrEqualTo(ActivityTypeOtherLengthMaximum)
	} else {
		validator.String("activityTypeOther", p.ActivityTypeOther).NotExists()
	}
	if p.Distance != nil {
		p.Distance.Validate(validator.WithReference("distance"))
	}
	if p.Duration != nil {
		p.Duration.Validate(validator.WithReference("duration"))
	}
	if p.ElevationChange != nil {
		p.ElevationChange.Validate(validator.WithReference("elevationChange"))
	}
	if p.Energy != nil {
		p.Energy.Validate(validator.WithReference("energy"))
	}
	if p.Flight != nil {
		p.Flight.Validate(validator.WithReference("flight"))
	}
	if p.Lap != nil {
		p.Lap.Validate(validator.WithReference("lap"))
	}
	validator.String("name", p.Name).NotEmpty().LengthLessThanOrEqualTo(NameLengthMaximum)
	validator.String("reportedIntensity", p.ReportedIntensity).OneOf(ReportedIntensities()...)
	if p.Step != nil {
		p.Step.Validate(validator.WithReference("step"))
	}
	validator.String("guid", p.GUID)
	if p.GUID != nil && p.Duration == nil {
		validator.WithReference("duration").ReportError(structureValidator.ErrorValueNotExists())
	}
	p.InputTime.Validate(validator)
}

// IsValid returns true if there is no error in the validator
func (p *Physical) IsValid(validator structure.Validator) bool {
	return !(validator.HasError())
}

func (p *Physical) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(p.Meta())
	}

	p.Base.Normalize(normalizer)

	if p.Distance != nil {
		p.Distance.Normalize(normalizer.WithReference("distance"))
	}
	if p.Duration != nil {
		p.Duration.Normalize(normalizer.WithReference("duration"))
	}
	if p.ElevationChange != nil {
		p.ElevationChange.Normalize(normalizer.WithReference("elevationChange"))
	}
	if p.Energy != nil {
		p.Energy.Normalize(normalizer.WithReference("energy"))
	}
	if p.Flight != nil {
		p.Flight.Normalize(normalizer.WithReference("flight"))
	}
	if p.Lap != nil {
		p.Lap.Normalize(normalizer.WithReference("lap"))
	}
	if p.Step != nil {
		p.Step.Normalize(normalizer.WithReference("step"))
	}
	p.InputTime.Normalize(normalizer.WithReference("inputTime"))
}
