package application_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"
	"os"
	"strings"

	"github.com/tidepool-org/platform/application"
	netTest "github.com/tidepool-org/platform/net/test"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Provider", func() {
	var prefix string
	var name string
	var scopes []string

	BeforeEach(func() {
		application.VersionBase = netTest.RandomSemanticVersion()
		application.VersionFullCommit = test.RandomStringFromRangeAndCharset(40, 40, test.CharsetHexidecimalLowercase)
		application.VersionShortCommit = application.VersionFullCommit[0:8]
		prefix = test.RandomStringFromRangeAndCharset(4, 8, test.CharsetUppercase)
		name = test.RandomStringFromRangeAndCharset(4, 8, test.CharsetAlphaNumeric)
		scopes = test.RandomStringArrayFromRangeAndCharset(0, 2, test.CharsetAlphaNumeric)
		os.Setenv(fmt.Sprintf("%s_LOGGER_LEVEL", prefix), "error")
		os.Args[0] = name
	})

	Context("NewProvider", func() {
		It("returns an error when the prefix is missing", func() {
			prefix = ""
			provider, err := application.NewProvider(prefix, scopes...)
			Expect(err).To(MatchError("prefix is missing"))
			Expect(provider).To(BeNil())
		})

		It("returns an error when the version base is missing", func() {
			application.VersionBase = ""
			provider, err := application.NewProvider(prefix, scopes...)
			Expect(err).To(MatchError("unable to create version reporter; base is missing"))
			Expect(provider).To(BeNil())
		})

		It("returns an error when the prefix is invalid", func() {
			prefix = "#invalid#"
			provider, err := application.NewProvider(prefix, scopes...)
			Expect(err).To(MatchError("unable to create config reporter; prefix is invalid"))
			Expect(provider).To(BeNil())
		})

		It("returns an error when the logger level is invalid", func() {
			os.Setenv(fmt.Sprintf("%s_LOGGER_LEVEL", prefix), "invalid")
			provider, err := application.NewProvider(prefix, scopes...)
			Expect(err).To(MatchError("unable to create logger; level not found"))
			Expect(provider).To(BeNil())
		})

		It("return successfully", func() {
			provider, err := application.NewProvider(prefix, scopes...)
			Expect(err).ToNot(HaveOccurred())
			Expect(provider).ToNot(BeNil())
		})
	})

	Context("with new provider", func() {
		var provider *application.ProviderImpl

		JustBeforeEach(func() {
			var err error
			provider, err = application.NewProvider(prefix, scopes...)
			Expect(err).ToNot(HaveOccurred())
			Expect(provider).ToNot(BeNil())
		})

		Context("VersionReporter", func() {
			It("returns successfully", func() {
				Expect(provider.VersionReporter()).ToNot(BeNil())
			})

			It("returns expected versions", func() {
				versionReporter := provider.VersionReporter()
				Expect(versionReporter).ToNot(BeNil())
				Expect(versionReporter.Short()).To(Equal(fmt.Sprintf("%s+%s", application.VersionBase, application.VersionShortCommit)))
				Expect(versionReporter.Long()).To(Equal(fmt.Sprintf("%s+%s", application.VersionBase, application.VersionFullCommit)))
			})
		})

		Context("ConfigReporter", func() {
			It("returns successfully", func() {
				Expect(provider.ConfigReporter()).ToNot(BeNil())
			})

			It("returns expected config", func() {
				configReporter := provider.ConfigReporter()
				Expect(configReporter).ToNot(BeNil())
				Expect(configReporter.WithScopes("logger").Get("level")).To(Equal("error"))
			})
		})

		Context("Logger", func() {
			It("returns successfully", func() {
				Expect(provider.Logger()).ToNot(BeNil())
			})
		})

		Context("Prefix", func() {
			It("returns successfully", func() {
				Expect(provider.Prefix()).To(Equal(prefix))
			})
		})

		Context("Name", func() {
			It("returns successfully", func() {
				Expect(provider.Name()).To(Equal(name))
			})
		})

		Context("UserAgent", func() {
			It("returns successfully", func() {
				Expect(provider.UserAgent()).To(Equal(fmt.Sprintf("%s-%s/%s", strings.Title(strings.ToLower(prefix)), strings.Title(strings.ToLower(name)), application.VersionBase)))
			})
		})

		When("running in a debugger", func() {
			var debugName string

			BeforeEach(func() {
				debugName = test.RandomStringFromRangeAndCharset(4, 8, test.CharsetAlphaNumeric)
				os.Setenv(fmt.Sprintf("%s_DEBUG_NAME", prefix), debugName)
				os.Args[0] = "debug"
			})

			It("uses the debug name from the environment", func() {
				Expect(provider.Name()).To(Equal(debugName))
			})
		})
	})
})
