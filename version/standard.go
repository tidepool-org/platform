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

func NewStandard(base string, commit string) *Standard {
	return &Standard{
		base:   app.FirstStringNotEmpty(base, "0.0.0"),
		commit: app.FirstStringNotEmpty(commit, "0000000000000000000000000000000000000000"),
	}
}

type Standard struct {
	base   string
	commit string
}

func (v *Standard) Base() string {
	return v.base
}

func (v *Standard) Commit() string {
	return v.commit
}

func (v *Standard) ShortCommit() string {
	if len(v.commit) > 8 {
		return v.commit[:8]
	}
	return v.commit
}

func (v *Standard) Short() string {
	return fmt.Sprintf("%s+%s", v.Base(), v.ShortCommit())
}

func (v *Standard) Long() string {
	return fmt.Sprintf("%s+%s", v.Base(), v.Commit())
}
