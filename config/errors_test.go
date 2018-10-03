package config_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	"github.com/tidepool-org/platform/config"
	errorsTest "github.com/tidepool-org/platform/errors/test"
)

var _ = Describe("Errors", func() {
	DescribeTable("have expected details when error",
		errorsTest.ExpectErrorDetails,
		Entry("is ErrorKeyNotFound", config.ErrorKeyNotFound("TEST"), "key-not-found", "key not found", "key \"TEST\" not found"),
	)
})
