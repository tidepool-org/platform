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

func NewBaseCommit(base string, commit string) *BaseCommit {
	return &BaseCommit{
		base:   app.FirstStringNotEmpty(base, "0.0.0"),
		commit: app.FirstStringNotEmpty(commit, "0000000000000000000000000000000000000000"),
	}
}

type BaseCommit struct {
	base   string
	commit string
}

func (v *BaseCommit) Base() string {
	return v.base
}

func (v *BaseCommit) Commit() string {
	return v.commit
}

func (v *BaseCommit) ShortCommit() string {
	if len(v.commit) > 8 {
		return v.commit[:8]
	}
	return v.commit
}

func (v *BaseCommit) Short() string {
	return fmt.Sprintf("%s+%s", v.Base(), v.ShortCommit())
}

func (v *BaseCommit) Long() string {
	return fmt.Sprintf("%s+%s", v.Base(), v.Commit())
}
