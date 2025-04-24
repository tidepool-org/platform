package log_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
)

var _ = Describe("Context", func() {
	Context("with context and logger", func() {
		var ctx context.Context
		var logger log.Logger

		BeforeEach(func() {
			ctx = context.Background()
			logger = logTest.NewLogger()
		})

		Context("NewContextWithLogger", func() {
			It("returns new context", func() {
				result := log.NewContextWithLogger(ctx, logger)
				Expect(result).ToNot(BeNil())
				Expect(result).ToNot(Equal(ctx))
				Expect(log.LoggerFromContext(result)).To(BeIdenticalTo(logger))
			})
		})

		Context("LoggerFromContext", func() {
			It("returns nil if context is nil", func() {
				ctx = nil
				Expect(log.LoggerFromContext(ctx)).To(BeNil())
			})

			It("returns nil if logger is not in context", func() {
				Expect(log.LoggerFromContext(ctx)).To(BeNil())
			})

			It("returns logger", func() {
				Expect(log.LoggerFromContext(log.NewContextWithLogger(ctx, logger))).To(BeIdenticalTo(logger))
			})
		})

		Context("with field", func() {
			var key string
			var value interface{}

			BeforeEach(func() {
				key = logTest.RandomKey()
				value = logTest.RandomValue()
			})

			Context("ContextWithField", func() {
				It("returns nil if context is nil", func() {
					ctx = nil
					Expect(log.ContextWithField(ctx, key, value)).To(BeNil())
				})

				It("returns context if logger is not in context", func() {
					Expect(log.ContextWithField(ctx, key, value)).To(Equal(ctx))
				})

				It("returns new context with logger with field", func() {
					result := log.ContextWithField(log.NewContextWithLogger(ctx, logger), key, value)
					Expect(result).ToNot(BeNil())
					Expect(result).ToNot(Equal(ctx))
					Expect(log.LoggerFromContext(result)).To(Equal(logger.WithField(key, value)))
				})
			})

			Context("ContextAndLoggerWithField", func() {
				It("returns nil context and logger if context is nil", func() {
					ctx = nil
					resultContext, resultLogger := log.ContextAndLoggerWithField(ctx, key, value)
					Expect(resultContext).To(BeNil())
					Expect(resultLogger).To(BeNil())
				})

				It("returns nil logger if logger is not in context", func() {
					resultContext, resultLogger := log.ContextAndLoggerWithField(ctx, key, value)
					Expect(resultContext).To(Equal(ctx))
					Expect(resultLogger).To(BeNil())
				})

				It("returns new content with logger with field and logger with field", func() {
					resultContext, resultLogger := log.ContextAndLoggerWithField(log.NewContextWithLogger(ctx, logger), key, value)
					Expect(resultContext).ToNot(BeNil())
					Expect(resultContext).ToNot(Equal(ctx))
					Expect(resultLogger).To(Equal(logger.WithField(key, value)))
					Expect(log.LoggerFromContext(resultContext)).To(Equal(logger.WithField(key, value)))
				})
			})
		})

		Context("with fields", func() {
			var fields log.Fields

			BeforeEach(func() {
				fields = logTest.RandomFields()
			})

			Context("ContextWithFields", func() {
				It("returns nil if context is nil", func() {
					ctx = nil
					Expect(log.ContextWithFields(ctx, fields)).To(BeNil())
				})

				It("returns context if logger is not in context", func() {
					Expect(log.ContextWithFields(ctx, fields)).To(Equal(ctx))
				})

				It("returns new context with logger with fields", func() {
					result := log.ContextWithFields(log.NewContextWithLogger(ctx, logger), fields)
					Expect(result).ToNot(BeNil())
					Expect(result).ToNot(Equal(ctx))
					Expect(log.LoggerFromContext(result)).To(Equal(logger.WithFields(fields)))
				})
			})

			Context("ContextAndLoggerWithFields", func() {
				It("returns nil context and logger if context is nil", func() {
					ctx = nil
					resultContext, resultLogger := log.ContextAndLoggerWithFields(ctx, fields)
					Expect(resultContext).To(BeNil())
					Expect(resultLogger).To(BeNil())
				})

				It("returns nil logger if logger is not in context", func() {
					resultContext, resultLogger := log.ContextAndLoggerWithFields(ctx, fields)
					Expect(resultContext).To(Equal(ctx))
					Expect(resultLogger).To(BeNil())
				})

				It("returns new content with logger with fields and logger with fields", func() {
					resultContext, resultLogger := log.ContextAndLoggerWithFields(log.NewContextWithLogger(ctx, logger), fields)
					Expect(resultContext).ToNot(BeNil())
					Expect(resultContext).ToNot(Equal(ctx))
					Expect(resultLogger).To(Equal(logger.WithFields(fields)))
					Expect(log.LoggerFromContext(resultContext)).To(Equal(logger.WithFields(fields)))
				})
			})
		})
	})
})
