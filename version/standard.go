package version

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
	"fmt"

	"github.com/tidepool-org/platform/app"
)

func NewStandard(base string, shortCommit string, fullCommit string) *Standard {
	return &Standard{
		base:        app.FirstStringNotEmpty(base, "0.0.0"),
		shortCommit: app.FirstStringNotEmpty(shortCommit, "0000000"),
		fullCommit:  app.FirstStringNotEmpty(fullCommit, "0000000000000000000000000000000000000000"),
	}
}

type Standard struct {
	base        string
	shortCommit string
	fullCommit  string
}

func (v *Standard) Base() string {
	return v.base
}

func (v *Standard) ShortCommit() string {
	return v.shortCommit
}

func (v *Standard) FullCommit() string {
	return v.fullCommit
}

func (v *Standard) Short() string {
	return fmt.Sprintf("%s+%s", v.Base(), v.ShortCommit())
}

func (v *Standard) Long() string {
	return fmt.Sprintf("%s+%s", v.Base(), v.FullCommit())
}
