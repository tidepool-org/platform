package physical

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types"
)

const (
	ReportedIntensityHigh   = "high"
	ReportedIntensityMedium = "medium"
	ReportedIntensityLow    = "low"
)

type Physical struct {
	types.Base `bson:",inline"`

	ReportedIntensity *string   `json:"reportedIntensity,omitempty" bson:"reportedIntensity,omitempty"`
	Duration          *Duration `json:"duration,omitempty" bson:"duration,omitempty"`
}

func Type() string {
	return "physicalActivity"
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

	p.ReportedIntensity = nil
	p.Duration = nil
}

func (p *Physical) Parse(parser data.ObjectParser) error {
	parser.SetMeta(p.Meta())

	if err := p.Base.Parse(parser); err != nil {
		return err
	}

	p.ReportedIntensity = parser.ParseString("reportedIntensity")
	p.Duration = ParseDuration(parser.NewChildObjectParser("duration"))

	return nil
}

func (p *Physical) Validate(validator data.Validator) error {
	validator.SetMeta(p.Meta())

	if err := p.Base.Validate(validator); err != nil {
		return err
	}

	validator.ValidateString("type", &p.Type).EqualTo(Type())
	validator.ValidateString("reportedIntensity", p.ReportedIntensity).OneOf([]string{ReportedIntensityHigh, ReportedIntensityMedium, ReportedIntensityLow})
	if p.Duration != nil {
		p.Duration.Validate(validator.NewChildValidator("duration"))
	}

	return nil
}

func (p *Physical) Normalize(normalizer data.Normalizer) error {
	normalizer.SetMeta(p.Meta())

	if err := p.Base.Normalize(normalizer); err != nil {
		return err
	}

	if p.Duration != nil {
		p.Duration.Normalize(normalizer.NewChildNormalizer("duration"))
	}

	return nil
}
