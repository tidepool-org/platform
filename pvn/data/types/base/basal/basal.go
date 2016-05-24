package basal

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

type Basal struct {
	base.Base `bson:",inline"`

	DeliveryType string `json:"deliveryType,omitempty" bson:"deliveryType,omitempty"`
}

func Type() string {
	return "basal"
}

func New(deliveryType string) (*Basal, error) {
	if deliveryType == "" {
		return nil, app.Error("basal", "delivery type is missing")
	}

	basalBase, err := base.New(Type())
	if err != nil {
		return nil, err
	}

	return &Basal{
		Base:         *basalBase,
		DeliveryType: deliveryType,
	}, nil
}
