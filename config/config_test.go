package config_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/config"
)

var _ = Describe("Config", func() {
	Context("SplitTrimCompact", func() {
		DescribeTable("returns expected strings when",
			func(sourceString string, expectedStrings []string) {
				Expect(config.SplitTrimCompact(sourceString)).To(Equal(expectedStrings))
			},
			Entry("has empty string with no separator", "", []string{}),
			Entry("has whitespace-only string with no separator", "   ", []string{}),
			Entry("has string with only separators", ",,,", []string{}),
			Entry("has whitespace-only string with separators", "  ,,   ,, ", []string{}),
			Entry("has non-whitespace string with no separator", "alpha", []string{"alpha"}),
			Entry("has string with whitespace no separator", "  alpha  ", []string{"alpha"}),
			Entry("has string with separators", "alpha,beta,charlie", []string{"alpha", "beta", "charlie"}),
			Entry("has string with whitespace and separators", "  alpha   ,  beta, charlie    ", []string{"alpha", "beta", "charlie"}),
		)
	})
})
