package physical

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/structure"
)

const (
	ReportedIntensityHigh   = "high"
	ReportedIntensityLow    = "low"
	ReportedIntensityMedium = "medium"
)

func ReportedIntensities() []string {
	return []string{
		ReportedIntensityHigh,
		ReportedIntensityLow,
		ReportedIntensityMedium,
	}
}

type Physical struct {
	types.Base `bson:",inline"`

	Duration          *Duration `json:"duration,omitempty" bson:"duration,omitempty"`
	ReportedIntensity *string   `json:"reportedIntensity,omitempty" bson:"reportedIntensity,omitempty"`
}

func Type() string {
	return "physicalActivity" // TODO: Change to "activity/physical"
}

func NewDatum() data.Datum {
	return New()
}

func New() *Physical {
	return &Physical{}
}

func Init() *Physical {
	physical := New()
	physical.Init()
	return physical
}

func (p *Physical) Init() {
	p.Base.Init()
	p.Type = Type()

	p.Duration = nil
	p.ReportedIntensity = nil
}

func (p *Physical) Parse(parser data.ObjectParser) error {
	parser.SetMeta(p.Meta())

	if err := p.Base.Parse(parser); err != nil {
		return err
	}

	p.Duration = ParseDuration(parser.NewChildObjectParser("duration"))
	p.ReportedIntensity = parser.ParseString("reportedIntensity")

	return nil
}

func (p *Physical) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(p.Meta())
	}

	p.Base.Validate(validator)

	if p.Type != "" {
		validator.String("type", &p.Type).EqualTo(Type())
	}

	if p.Duration != nil {
		p.Duration.Validate(validator.WithReference("duration"))
	}
	validator.String("reportedIntensity", p.ReportedIntensity).OneOf(ReportedIntensities()...)
}

func (p *Physical) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(p.Meta())
	}

	p.Base.Normalize(normalizer)

	if p.Duration != nil {
		p.Duration.Normalize(normalizer.WithReference("duration"))
	}
}
