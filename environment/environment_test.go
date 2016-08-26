package environment_test

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/environment"
)

var _ = Describe("Environment", func() {
	Context("NewReporter", func() {
		It("returns an error if name is missing", func() {
			reporter, err := environment.NewReporter("", "brounie")
			Expect(err).To(MatchError("environment: name is missing"))
			Expect(reporter).To(BeNil())
		})

		It("returns successfully if prefix is missing", func() {
			Expect(environment.NewReporter("brownie", "")).ToNot(BeNil())
		})

		It("returns successfully", func() {
			Expect(environment.NewReporter("brownie", "brounie")).ToNot(BeNil())
		})
	})

	Context("Reporter", func() {
		Context("Name", func() {
			It("returns the name", func() {
				reporter, err := environment.NewReporter("brownie", "brounie")
				Expect(err).ToNot(HaveOccurred())
				Expect(reporter).ToNot(BeNil())
				Expect(reporter.Name()).To(Equal("brownie"))
			})
		})

		Context("IsLocal", func() {
			It("returns false if environment is not local", func() {
				reporter, err := environment.NewReporter("brownie", "brounie")
				Expect(err).ToNot(HaveOccurred())
				Expect(reporter).ToNot(BeNil())
				Expect(reporter.IsLocal()).To(BeFalse())
			})

			It("returns true if environment is local", func() {
				reporter, err := environment.NewReporter("local", "brounie")
				Expect(err).ToNot(HaveOccurred())
				Expect(reporter).ToNot(BeNil())
				Expect(reporter.IsLocal()).To(BeTrue())
			})
		})

		Context("IsTest", func() {
			It("returns false if environment is not test", func() {
				reporter, err := environment.NewReporter("brownie", "brounie")
				Expect(err).ToNot(HaveOccurred())
				Expect(reporter).ToNot(BeNil())
				Expect(reporter.IsTest()).To(BeFalse())
			})

			It("returns true if environment is test", func() {
				reporter, err := environment.NewReporter("test", "brounie")
				Expect(err).ToNot(HaveOccurred())
				Expect(reporter).ToNot(BeNil())
				Expect(reporter.IsTest()).To(BeTrue())
			})
		})

		Context("IsDeployed", func() {
			It("returns false if environment is local", func() {
				reporter, err := environment.NewReporter("local", "brounie")
				Expect(err).ToNot(HaveOccurred())
				Expect(reporter).ToNot(BeNil())
				Expect(reporter.IsDeployed()).To(BeFalse())
			})

			It("returns false if environment is test", func() {
				reporter, err := environment.NewReporter("test", "brounie")
				Expect(err).ToNot(HaveOccurred())
				Expect(reporter).ToNot(BeNil())
				Expect(reporter.IsDeployed()).To(BeFalse())
			})

			It("returns true if environment is not local nor test", func() {
				reporter, err := environment.NewReporter("brownie", "brounie")
				Expect(err).ToNot(HaveOccurred())
				Expect(reporter).ToNot(BeNil())
				Expect(reporter.IsDeployed()).To(BeTrue())
			})
		})

		Context("Prefix", func() {
			It("returns the prefix", func() {
				reporter, err := environment.NewReporter("brownie", "brounie")
				Expect(err).ToNot(HaveOccurred())
				Expect(reporter).ToNot(BeNil())
				Expect(reporter.Prefix()).To(Equal("brounie"))
			})

			It("returns the prefix even if missing", func() {
				reporter, err := environment.NewReporter("brownie", "")
				Expect(err).ToNot(HaveOccurred())
				Expect(reporter).ToNot(BeNil())
				Expect(reporter.Prefix()).To(Equal(""))
			})
		})

		Context("GetKey", func() {
			It("returns the environment key", func() {
				key := app.NewID()
				reporter, err := environment.NewReporter("brownie", "brounie")
				Expect(err).ToNot(HaveOccurred())
				Expect(reporter).ToNot(BeNil())
				Expect(reporter.GetKey(key)).To(Equal(fmt.Sprintf("brounie_%s", key)))
			})

			It("returns the environment key even with a missing prefix", func() {
				key := app.NewID()
				reporter, err := environment.NewReporter("brownie", "")
				Expect(err).ToNot(HaveOccurred())
				Expect(reporter).ToNot(BeNil())
				Expect(reporter.GetKey(key)).To(Equal(key))
			})
		})

		Context("GetValue", func() {
			It("returns the environment variable", func() {
				key := app.NewID()
				value := app.NewID()
				Expect(os.Setenv(fmt.Sprintf("brounie_%s", key), value)).To(Succeed())
				reporter, err := environment.NewReporter("brownie", "brounie")
				Expect(err).ToNot(HaveOccurred())
				Expect(reporter).ToNot(BeNil())
				Expect(reporter.GetValue(key)).To(Equal(value))
			})

			It("returns the environment variable even with a missing prefix", func() {
				key := app.NewID()
				value := app.NewID()
				Expect(os.Setenv(key, value)).To(Succeed())
				reporter, err := environment.NewReporter("brownie", "")
				Expect(err).ToNot(HaveOccurred())
				Expect(reporter).ToNot(BeNil())
				Expect(reporter.GetValue(key)).To(Equal(value))
			})
		})
	})

	Context("GetKey", func() {
		It("returns the environment key", func() {
			key := app.NewID()
			Expect(environment.GetKey(key, "brounie")).To(Equal(fmt.Sprintf("brounie_%s", key)))
		})

		It("returns the environment key even with a missing prefix", func() {
			key := app.NewID()
			Expect(environment.GetKey(key, "")).To(Equal(key))
		})
	})

	Context("GetValue", func() {
		It("returns the environment variable", func() {
			key := app.NewID()
			value := app.NewID()
			Expect(os.Setenv(fmt.Sprintf("brounie_%s", key), value)).To(Succeed())
			Expect(environment.GetValue(key, "brounie")).To(Equal(value))
		})

		It("returns the environment variable even with a missing prefix", func() {
			key := app.NewID()
			value := app.NewID()
			Expect(os.Setenv(key, value)).To(Succeed())
			Expect(environment.GetValue(key, "")).To(Equal(value))
		})
	})
})
