package test

import (
	"fmt"

	netTest "github.com/tidepool-org/platform/net/test"
	"github.com/tidepool-org/platform/test"
)

type Reporter struct {
	base   string
	commit string
}

func NewReporter() *Reporter {
	return &Reporter{
		base:   netTest.RandomSemanticVersion(),
		commit: test.RandomStringFromRangeAndCharset(40, 40, test.CharsetHexadecimalLowercase),
	}
}

func (r *Reporter) Base() string {
	return r.base
}

func (r *Reporter) ShortCommit() string {
	return r.commit[0:8]
}

func (r *Reporter) FullCommit() string {
	return r.commit
}

func (r *Reporter) Short() string {
	return fmt.Sprintf("%s+%s", r.Base(), r.ShortCommit())
}

func (r *Reporter) Long() string {
	return fmt.Sprintf("%s+%s", r.Base(), r.FullCommit())
}
