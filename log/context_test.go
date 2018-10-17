package log_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/log"
	logNull "github.com/tidepool-org/platform/log/null"
)

var _ = Describe("Context", func() {
	Context("NewContextWithLogger", func() {
		It("returns a new context", func() {
			ctx := context.Background()
			result := log.NewContextWithLogger(ctx, logNull.NewLogger())
			Expect(result).ToNot(BeNil())
			Expect(result).ToNot(Equal(ctx))
		})
	})

	Context("LoggerFromContext", func() {
		It("returns nil if context is nil", func() {
			Expect(log.LoggerFromContext(nil)).To(BeNil())
		})

		It("returns nil if logger is not in context", func() {
			Expect(log.LoggerFromContext(context.Background())).To(BeNil())
		})

		It("returns logger", func() {
			logger := logNull.NewLogger()
			Expect(log.LoggerFromContext(log.NewContextWithLogger(context.Background(), logger))).To(BeIdenticalTo(logger))
		})
	})
})
