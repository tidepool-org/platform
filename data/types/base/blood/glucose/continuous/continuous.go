package continuous

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/base/blood/glucose"
)

type Continuous struct {
	glucose.Glucose `bson:",inline"`
}

func Type() string {
	return "cbg"
}

func NewDatum() data.Datum {
	return New()
}

func New() *Continuous {
	return &Continuous{}
}

func Init() *Continuous {
	continuous := New()
	continuous.Init()
	return continuous
}

func (c *Continuous) Init() {
	c.Glucose.Init()
	c.Type = Type()
}
