package bolus

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

import (
	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/pvn/data/types/base"
)

type Bolus struct {
	base.Base `bson:",inline"`

	SubType string `json:"subType,omitempty" bson:"subType,omitempty"`
}

func Type() string {
	return "bolus"
}

func New(subType string) (*Bolus, error) {
	if subType == "" {
		return nil, app.Error("basal", "sub type is missing")
	}

	bolusBase, err := base.New(Type())
	if err != nil {
		return nil, err
	}

	return &Bolus{
		Base:    *bolusBase,
		SubType: subType,
	}, nil
}
