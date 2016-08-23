package log_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"

	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/version"
)

var _ = Describe("Standard", func() {
	var versionReporter version.Reporter

	BeforeEach(func() {
		var err error
		versionReporter, err = version.NewReporter("0.0.0", "0000000", "0000000000000000000000000000000000000000")
		Expect(err).ToNot(HaveOccurred())
		Expect(versionReporter).ToNot(BeNil())
	})

	Context("NewStandard", func() {
		It("returns an error if version reporter is missing", func() {
			standard, err := log.NewStandard(nil, &log.Config{Level: "debug"})
			Expect(err).To(MatchError("log: version reporter is missing"))
			Expect(standard).To(BeNil())
		})

		It("returns an error if config is missing", func() {
			standard, err := log.NewStandard(versionReporter, nil)
			Expect(err).To(MatchError("log: config is missing"))
			Expect(standard).To(BeNil())
		})

		It("returns an error if config level is missing", func() {
			standard, err := log.NewStandard(versionReporter, &log.Config{})
			Expect(err).To(MatchError("log: config is invalid"))
			Expect(standard).To(BeNil())
		})

		It("returns an error if config level is invalid", func() {
			standard, err := log.NewStandard(versionReporter, &log.Config{Level: "invalid"})
			Expect(err).To(MatchError("log: config is invalid"))
			Expect(standard).To(BeNil())
		})

		It("returns successfully", func() {
			Expect(log.NewStandard(versionReporter, &log.Config{Level: "debug"})).ToNot(BeNil())
		})
	})

	Context("with new standard logger", func() {
		var standard *log.Standard

		BeforeEach(func() {
			var err error
			standard, err = log.NewStandard(versionReporter, &log.Config{Level: "fatal"})
			Expect(err).ToNot(HaveOccurred())
			Expect(standard).ToNot(BeNil())
		})

		Context("Debug", func() {
			It("works as expected", func() {
				standard.Debug("message")
			})
		})

		Context("Info", func() {
			It("works as expected", func() {
				standard.Info("message")
			})
		})

		Context("Warn", func() {
			It("works as expected", func() {
				standard.Warn("message")
			})
		})

		Context("Error", func() {
			It("works as expected", func() {
				standard.Error("message")
			})
		})

		Context("WithError", func() {
			It("returns a logger with an error", func() {
				Expect(standard.WithError(errors.New("test: error"))).ToNot(BeNil())
			})

			It("returns a logger with nil error", func() {
				Expect(standard.WithError(nil)).ToNot(BeNil())
			})
		})

		Context("WithField", func() {
			It("returns a logger with a field", func() {
				Expect(standard.WithField("field", 1)).ToNot(BeNil())
			})

			It("returns a logger with a field with empty key", func() {
				Expect(standard.WithField("", 1)).ToNot(BeNil())
			})

			It("returns a logger with a field with nil value", func() {
				Expect(standard.WithField("field", nil)).ToNot(BeNil())
			})
		})

		Context("WithFields", func() {
			It("returns a logger with fields", func() {
				Expect(standard.WithFields(log.Fields{"field": 1})).ToNot(BeNil())
			})

			It("returns a logger with fields", func() {
				Expect(standard.WithFields(nil)).ToNot(BeNil())
			})
		})
	})
})
