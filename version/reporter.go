package version

import (
	"fmt"

	"github.com/tidepool-org/platform/errors"
)

func NewReporter(base string, shortCommit string, fullCommit string) (Reporter, error) {
	if base == "" {
		return nil, errors.New("base is missing")
	}
	if shortCommit == "" {
		return nil, errors.New("short commit is missing")
	}
	if fullCommit == "" {
		return nil, errors.New("full commit is missing")
	}

	return &reporter{
		base:        base,
		shortCommit: shortCommit,
		fullCommit:  fullCommit,
	}, nil
}

type reporter struct {
	base        string
	shortCommit string
	fullCommit  string
}

func (r *reporter) Base() string {
	return r.base
}

func (r *reporter) ShortCommit() string {
	return r.shortCommit
}

func (r *reporter) FullCommit() string {
	return r.fullCommit
}

func (r *reporter) Short() string {
	return fmt.Sprintf("%s+%s", r.Base(), r.ShortCommit())
}

func (r *reporter) Long() string {
	return fmt.Sprintf("%s+%s", r.Base(), r.FullCommit())
}
