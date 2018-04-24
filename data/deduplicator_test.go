package data_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"encoding/json"
	"fmt"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/errors"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/id"
	testStructure "github.com/tidepool-org/platform/structure/test"
)

var _ = Describe("Deduplicator", func() {
	Context("DeduplicatorDescriptor", func() {
		Context("NewDeduplicatorDescriptor", func() {
			It("returns a new deduplicator descriptor", func() {
				Expect(data.NewDeduplicatorDescriptor()).ToNot(BeNil())
			})
		})

		Context("with a new deduplicator descriptor", func() {
			var testDeduplicatorDescriptor *data.DeduplicatorDescriptor
			var testName string
			var testVersion string

			BeforeEach(func() {
				testDeduplicatorDescriptor = data.NewDeduplicatorDescriptor()
				Expect(testDeduplicatorDescriptor).ToNot(BeNil())
				testName = id.New()
				testVersion = "1.2.3"
				testDeduplicatorDescriptor.Name = testName
				testDeduplicatorDescriptor.Version = testVersion
			})

			Context("IsRegisteredWithAnyDeduplicator", func() {
				It("returns false if the deduplicator descriptor name is missing", func() {
					testDeduplicatorDescriptor.Name = ""
					Expect(testDeduplicatorDescriptor.IsRegisteredWithAnyDeduplicator()).To(BeFalse())
				})

				It("returns true if the deduplicator descriptor name is present", func() {
					Expect(testDeduplicatorDescriptor.IsRegisteredWithAnyDeduplicator()).To(BeTrue())
				})
			})

			Context("IsRegisteredWithNamedDeduplicator", func() {
				It("returns false if the deduplicator descriptor name is missing", func() {
					testDeduplicatorDescriptor.Name = ""
					Expect(testDeduplicatorDescriptor.IsRegisteredWithNamedDeduplicator(testName)).To(BeFalse())
				})

				It("returns true if the deduplicator descriptor name is present, but does not match", func() {
					Expect(testDeduplicatorDescriptor.IsRegisteredWithNamedDeduplicator(id.New())).To(BeFalse())
				})

				It("returns true if the deduplicator descriptor name is present and matches", func() {
					Expect(testDeduplicatorDescriptor.IsRegisteredWithNamedDeduplicator(testName)).To(BeTrue())
				})
			})

			Context("RegisterWithDeduplicator", func() {
				var testDeduplicator *test.Deduplicator

				BeforeEach(func() {
					testDeduplicator = test.NewDeduplicator()
				})

				AfterEach(func() {
					testDeduplicator.Expectations()
				})

				It("returns error if the deduplicator descriptor already has a name", func() {
					err := testDeduplicatorDescriptor.RegisterWithDeduplicator(testDeduplicator)
					Expect(err).To(MatchError(fmt.Sprintf("deduplicator descriptor already registered with %q", testName)))
				})

				It("returns successfully if the deduplicator descriptor does not already have a name", func() {
					testDeduplicatorDescriptor.Name = ""
					testDeduplicatorDescriptor.Version = ""
					testDeduplicator.NameOutputs = []string{testName}
					testDeduplicator.VersionOutputs = []string{testVersion}
					err := testDeduplicatorDescriptor.RegisterWithDeduplicator(testDeduplicator)
					Expect(err).ToNot(HaveOccurred())
					Expect(testDeduplicatorDescriptor.Name).To(Equal(testName))
					Expect(testDeduplicatorDescriptor.Version).To(Equal(testVersion))
				})
			})
		})
	})

	Context("ValidateReverseDomain", func() {
		DescribeTable("validates the reverse domain",
			func(value string, expectedErrors ...error) {
				errorReporter := testStructure.NewErrorReporter()
				Expect(errorReporter).ToNot(BeNil())
				data.ValidateReverseDomain(value, errorReporter)
				testErrors.ExpectEqual(errorReporter.Error(), expectedErrors...)
			},
			Entry("succeeds with single domain", "top.one"),
			Entry("succeeds with multiple domains", "top.one.two.three"),
			Entry("tld; out of range (lower)", "a.one.two.three", data.ErrorValueStringAsReverseDomainNotValid("a.one.two.three")),
			Entry("tld; in range (lower)", "ab.one.two.three"),
			Entry("tld; in range (upper)", "abcdef.one.two.three"),
			Entry("tld; out of range (upper)", "abcdefg.one.two.three", data.ErrorValueStringAsReverseDomainNotValid("abcdefg.one.two.three")),
			Entry("tld; invalid character; -", "a-c.one.two.three", data.ErrorValueStringAsReverseDomainNotValid("a-c.one.two.three")),
			Entry("tld; invalid character; _", "a_c.one.two.three", data.ErrorValueStringAsReverseDomainNotValid("a_c.one.two.three")),
			Entry("tld; invalid character; 0", "a0c.one.two.three", data.ErrorValueStringAsReverseDomainNotValid("a0c.one.two.three")),
			Entry("tld; only", "abc", data.ErrorValueStringAsReverseDomainNotValid("abc")),
			Entry("tld; trailing dot", "abc.", data.ErrorValueStringAsReverseDomainNotValid("abc.")),
			Entry("single domain; out of range (lower)", "org..two", data.ErrorValueStringAsReverseDomainNotValid("org..two")),
			Entry("single domain; in range (lower)", "org.a"),
			Entry("single domain; in range (lower); multiple", "org.ab"),
			Entry("single domain; in range (upper)", "org.abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-1234567890"),
			Entry("single domain; out of range (upper)", "org.abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-1234567890a", data.ErrorValueStringAsReverseDomainNotValid("org.abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-1234567890a")),
			Entry("single domain; invalid character; _", "org.a_c", data.ErrorValueStringAsReverseDomainNotValid("org.a_c")),
			Entry("single domain; starts with dash", "org.-bc", data.ErrorValueStringAsReverseDomainNotValid("org.-bc")),
			Entry("single domain; ends with dash", "org.ab-", data.ErrorValueStringAsReverseDomainNotValid("org.ab-")),
			Entry("single domain; ends with dot", "org.abc.", data.ErrorValueStringAsReverseDomainNotValid("org.abc.")),
			Entry("multiple domains; out of range (lower)", "org.one..three", data.ErrorValueStringAsReverseDomainNotValid("org.one..three")),
			Entry("multiple domains; in range (lower)", "org.one.a"),
			Entry("multiple domains; in range (lower); multiple", "org.one.ab"),
			Entry("multiple domains; in range (upper)", "org.one.abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-1234567890"),
			Entry("multiple domains; out of range (upper)", "org.one.abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-1234567890a", data.ErrorValueStringAsReverseDomainNotValid("org.one.abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-1234567890a")),
			Entry("multiple domains; invalid character; _", "org.one.a_c", data.ErrorValueStringAsReverseDomainNotValid("org.one.a_c")),
			Entry("multiple domains; starts with dash", "org.one.-bc", data.ErrorValueStringAsReverseDomainNotValid("org.one.-bc")),
			Entry("multiple domains; ends with dash", "org.one.ab-", data.ErrorValueStringAsReverseDomainNotValid("org.one.ab-")),
			Entry("multiple domains; ends with dot", "org.one.abc.", data.ErrorValueStringAsReverseDomainNotValid("org.one.abc.")),
		)
	})

	Context("ValidateSemanticVersion", func() {
		DescribeTable("validates the semantic version",
			func(value string, expectedErrors ...error) {
				errorReporter := testStructure.NewErrorReporter()
				Expect(errorReporter).ToNot(BeNil())
				data.ValidateSemanticVersion(value, errorReporter)
				testErrors.ExpectEqual(errorReporter.Error(), expectedErrors...)
			},
			Entry("succeeds", "1.2.3"),
			Entry("build missing", "1.2", data.ErrorValueStringAsSemanticVersionNotValid("1.2")),
			Entry("minor and build missing", "1", data.ErrorValueStringAsSemanticVersionNotValid("1")),
			Entry("v prefix", "v1.2.3", data.ErrorValueStringAsSemanticVersionNotValid("v1.2.3")),
		)
	})

	Context("Errors", func() {
		DescribeTable("all errors",
			func(err error, code string, title string, detail string) {
				Expect(err).ToNot(BeNil())
				Expect(errors.Code(err)).To(Equal(code))
				Expect(errors.Cause(err)).To(Equal(err))
				bytes, bytesErr := json.Marshal(errors.Sanitize(err))
				Expect(bytesErr).ToNot(HaveOccurred())
				Expect(bytes).To(MatchJSON(fmt.Sprintf(`{"code": %q, "title": %q, "detail": %q}`, code, title, detail)))
			},
			Entry("is ErrorValueStringAsReverseDomainNotValid", data.ErrorValueStringAsReverseDomainNotValid("abc"), "value-not-valid", "value is not valid", `value "abc" is not valid as reverse domain`),
			Entry("is ErrorValueStringAsSemanticVersionNotValid", data.ErrorValueStringAsSemanticVersionNotValid("abc"), "value-not-valid", "value is not valid", `value "abc" is not valid as semantic version`),
		)
	})
})
