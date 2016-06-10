package environment_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/environment"
)

var _ = Describe("Environment", func() {
	Context("NewReporter", func() {
		It("returns an error if name is missing", func() {
			reporter, err := environment.NewReporter("")
			Expect(err).To(MatchError("environment: name is missing"))
			Expect(reporter).To(BeNil())
		})

		It("returns a new object if name is specified", func() {
			Expect(environment.NewReporter("brownie")).ToNot(BeNil())
		})
	})

	Context("Reporter", func() {
		Context("Name", func() {
			It("returns the name", func() {
				reporter, err := environment.NewReporter("brownie")
				Expect(err).ToNot(HaveOccurred())
				Expect(reporter).ToNot(BeNil())
				Expect(reporter.Name()).To(Equal("brownie"))
			})
		})

		Context("IsLocal", func() {
			It("returns false if environment is not local", func() {
				reporter, err := environment.NewReporter("brownie")
				Expect(err).ToNot(HaveOccurred())
				Expect(reporter).ToNot(BeNil())
				Expect(reporter.IsLocal()).To(BeFalse())
			})

			It("returns true if environment is local", func() {
				reporter, err := environment.NewReporter("local")
				Expect(err).ToNot(HaveOccurred())
				Expect(reporter).ToNot(BeNil())
				Expect(reporter.IsLocal()).To(BeTrue())
			})
		})

		Context("IsTest", func() {
			It("returns false if environment is not test", func() {
				reporter, err := environment.NewReporter("brownie")
				Expect(err).ToNot(HaveOccurred())
				Expect(reporter).ToNot(BeNil())
				Expect(reporter.IsTest()).To(BeFalse())
			})

			It("returns true if environment is test", func() {
				reporter, err := environment.NewReporter("test")
				Expect(err).ToNot(HaveOccurred())
				Expect(reporter).ToNot(BeNil())
				Expect(reporter.IsTest()).To(BeTrue())
			})
		})

		Context("IsDeployed", func() {
			It("returns false if environment is local", func() {
				reporter, err := environment.NewReporter("local")
				Expect(err).ToNot(HaveOccurred())
				Expect(reporter).ToNot(BeNil())
				Expect(reporter.IsDeployed()).To(BeFalse())
			})

			It("returns false if environment is test", func() {
				reporter, err := environment.NewReporter("test")
				Expect(err).ToNot(HaveOccurred())
				Expect(reporter).ToNot(BeNil())
				Expect(reporter.IsDeployed()).To(BeFalse())
			})

			It("returns true if environment is not local nor test", func() {
				reporter, err := environment.NewReporter("brownie")
				Expect(err).ToNot(HaveOccurred())
				Expect(reporter).ToNot(BeNil())
				Expect(reporter.IsDeployed()).To(BeTrue())
			})
		})
	})
})
