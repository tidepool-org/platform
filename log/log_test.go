package log_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/version"
)

var _ = Describe("Log", func() {
	var versionReporter version.Reporter

	BeforeEach(func() {
		var err error
		versionReporter, err = version.NewReporter("0.0.0", "0000000", "0000000000000000000000000000000000000000")
		Expect(err).To(Succeed())
		Expect(versionReporter).ToNot(BeNil())
	})

	Context("NewLogger", func() {
		It("returns an error if config is missing", func() {
			reporter, err := log.NewLogger(nil, versionReporter)
			Expect(err).To(MatchError("log: config is missing"))
			Expect(reporter).To(BeNil())
		})

		It("returns an error if version reporter is missing", func() {
			reporter, err := log.NewLogger(&log.Config{Level: "debug"}, nil)
			Expect(err).To(MatchError("log: version reporter is missing"))
			Expect(reporter).To(BeNil())
		})

		It("returns an error if config level is missing", func() {
			reporter, err := log.NewLogger(&log.Config{}, versionReporter)
			Expect(err).To(MatchError("log: config is invalid"))
			Expect(reporter).To(BeNil())
		})

		It("returns an error if config level is invalid", func() {
			reporter, err := log.NewLogger(&log.Config{Level: "invalid"}, versionReporter)
			Expect(err).To(MatchError("log: config is invalid"))
			Expect(reporter).To(BeNil())
		})

		It("returns successfully", func() {
			reporter, err := log.NewLogger(&log.Config{Level: "debug"}, versionReporter)
			Expect(err).To(Succeed())
			Expect(reporter).ToNot(BeNil())
		})
	})

	Context("Logger", func() {
		var logger log.Logger

		BeforeEach(func() {
			var err error
			logger, err = log.NewLogger(&log.Config{Level: "debug"}, versionReporter)
			Expect(err).To(Succeed())
			Expect(logger).ToNot(BeNil())
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
