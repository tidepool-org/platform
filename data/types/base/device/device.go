package device

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
	"github.com/tidepool-org/platform/data/types/base"
)

type Device struct {
	base.Base `bson:",inline"`

	SubType string `json:"subType,omitempty" bson:"subType,omitempty"`
}

func Type() string {
	return "deviceEvent"
}

func New(subType string) (*Device, error) {
	if subType == "" {
		return nil, app.Error("basal", "sub type is missing")
	}

	deviceBase, err := base.New(Type())
	if err != nil {
		return nil, err
	}

	return &Device{
		Base:    *deviceBase,
		SubType: subType,
	}, nil
}
