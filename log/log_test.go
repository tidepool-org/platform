package log_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"

	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/version"
)

var _ = Describe("Log", func() {
	var versionReporter version.Reporter

	BeforeEach(func() {
		var err error
		versionReporter, err = version.NewReporter("0.0.0", "0000000", "0000000000000000000000000000000000000000")
		Expect(err).ToNot(HaveOccurred())
		Expect(versionReporter).ToNot(BeNil())
	})

	Context("NewLogger", func() {
		It("returns an error if version reporter is missing", func() {
			reporter, err := log.NewLogger(nil, &log.Config{Level: "debug"})
			Expect(err).To(MatchError("log: version reporter is missing"))
			Expect(reporter).To(BeNil())
		})

		It("returns an error if config is missing", func() {
			reporter, err := log.NewLogger(versionReporter, nil)
			Expect(err).To(MatchError("log: config is missing"))
			Expect(reporter).To(BeNil())
		})

		It("returns an error if config level is missing", func() {
			reporter, err := log.NewLogger(versionReporter, &log.Config{})
			Expect(err).To(MatchError("log: config is invalid"))
			Expect(reporter).To(BeNil())
		})

		It("returns an error if config level is invalid", func() {
			reporter, err := log.NewLogger(versionReporter, &log.Config{Level: "invalid"})
			Expect(err).To(MatchError("log: config is invalid"))
			Expect(reporter).To(BeNil())
		})

		It("returns successfully", func() {
			Expect(log.NewLogger(versionReporter, &log.Config{Level: "debug"})).ToNot(BeNil())
		})
	})

	Context("Logger", func() {
		var logger log.Logger

		BeforeEach(func() {
			var err error
			logger, err = log.NewLogger(versionReporter, &log.Config{Level: "fatal"})
			Expect(err).ToNot(HaveOccurred())
			Expect(logger).ToNot(BeNil())
		})

		Context("Debug", func() {
			It("works as expected", func() {
				logger.Debug("message")
			})
		})

		Context("Info", func() {
			It("works as expected", func() {
				logger.Info("message")
			})
		})

		Context("Warn", func() {
			It("works as expected", func() {
				logger.Warn("message")
			})
		})

		Context("Error", func() {
			It("works as expected", func() {
				logger.Error("message")
			})
		})

		Context("WithError", func() {
			It("returns a logger with an error", func() {
				Expect(logger.WithError(errors.New("test: error"))).ToNot(BeNil())
			})

			It("returns a logger with nil error", func() {
				Expect(logger.WithError(nil)).ToNot(BeNil())
			})
		})

		Context("WithField", func() {
			It("returns a logger with a field", func() {
				Expect(logger.WithField("field", 1)).ToNot(BeNil())
			})

			It("returns a logger with a field with empty key", func() {
				Expect(logger.WithField("", 1)).ToNot(BeNil())
			})

			It("returns a logger with a field with nil value", func() {
				Expect(logger.WithField("field", nil)).ToNot(BeNil())
			})
		})

		Context("WithFields", func() {
			It("returns a logger with fields", func() {
				Expect(logger.WithFields(log.Fields{"field": 1})).ToNot(BeNil())
			})

			It("returns a logger with fields", func() {
				Expect(logger.WithFields(nil)).ToNot(BeNil())
			})
		})
	})
})
