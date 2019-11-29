package continuous_test

import (
	. "github.com/onsi/ginkgo"

	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
)

func CloneTrend(datum *continuous.Trend) *continuous.Trend {
	if datum == nil {
		return nil
	}
	clone := continuous.NewTrend()
	return clone
}

var _ = Describe("Trend", func() {
})
