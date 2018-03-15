package validate_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"encoding/json"
	"fmt"

	"github.com/tidepool-org/platform/errors"
	testErrors "github.com/tidepool-org/platform/errors/test"
	testStructure "github.com/tidepool-org/platform/structure/test"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
	testHTTP "github.com/tidepool-org/platform/test/http"
	"github.com/tidepool-org/platform/validate"
)

var _ = Describe("Valdate", func() {
	Context("ReverseDomain", func() {
		DescribeTable("validates the reverse domain",
			func(value string, expectedErrors ...error) {
				errorReporter := testStructure.NewErrorReporter()
				Expect(errorReporter).ToNot(BeNil())
				validate.ReverseDomain(value, errorReporter)
				testErrors.ExpectEqual(errorReporter.Error(), expectedErrors...)
			},
			Entry("succeeds with single domain", "top.one"),
			Entry("succeeds with multiple domains", "top.one.two.three"),
			Entry("empty", "", structureValidator.ErrorValueEmpty()),
			Entry("tld; out of range (lower)", "a.one.two.three", validate.ErrorValueStringAsReverseDomainNotValid("a.one.two.three")),
			Entry("tld; in range (lower)", "ab.one.two.three"),
			Entry("tld; in range (upper)", "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijk.one.two.three"),
			Entry("tld; out of range (upper)", "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijkl.one.two.three", validate.ErrorValueStringAsReverseDomainNotValid("abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijkl.one.two.three")),
			Entry("tld; invalid character; uppercase", "aBc.one.two.three", validate.ErrorValueStringAsReverseDomainNotValid("aBc.one.two.three")),
			Entry("tld; invalid character; -", "a-c.one.two.three", validate.ErrorValueStringAsReverseDomainNotValid("a-c.one.two.three")),
			Entry("tld; invalid character; _", "a_c.one.two.three", validate.ErrorValueStringAsReverseDomainNotValid("a_c.one.two.three")),
			Entry("tld; invalid character; 0", "a0c.one.two.three", validate.ErrorValueStringAsReverseDomainNotValid("a0c.one.two.three")),
			Entry("tld; only", "abc", validate.ErrorValueStringAsReverseDomainNotValid("abc")),
			Entry("tld; trailing dot", "abc.", validate.ErrorValueStringAsReverseDomainNotValid("abc.")),
			Entry("single domain; out of range (lower)", "org..two", validate.ErrorValueStringAsReverseDomainNotValid("org..two")),
			Entry("single domain; in range (lower)", "org.a"),
			Entry("single domain; in range (lower); multiple", "org.ab"),
			Entry("single domain; in range (upper)", "org.abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz-1234567890"),
			Entry("single domain; out of range (upper)", "org.abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz-1234567890a", validate.ErrorValueStringAsReverseDomainNotValid("org.abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz-1234567890a")),
			Entry("single domain; invalid character; _", "org.a_c", validate.ErrorValueStringAsReverseDomainNotValid("org.a_c")),
			Entry("single domain; starts with dash", "org.-bc", validate.ErrorValueStringAsReverseDomainNotValid("org.-bc")),
			Entry("single domain; ends with dash", "org.ab-", validate.ErrorValueStringAsReverseDomainNotValid("org.ab-")),
			Entry("single domain; ends with dot", "org.abc.", validate.ErrorValueStringAsReverseDomainNotValid("org.abc.")),
			Entry("multiple domains; out of range (lower)", "org.one..three", validate.ErrorValueStringAsReverseDomainNotValid("org.one..three")),
			Entry("multiple domains; in range (lower)", "org.one.a"),
			Entry("multiple domains; in range (lower); multiple", "org.one.ab"),
			Entry("multiple domains; in range (upper)", "org.one.abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz-1234567890"),
			Entry("multiple domains; out of range (upper)", "org.one.abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz-1234567890a", validate.ErrorValueStringAsReverseDomainNotValid("org.one.abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz-1234567890a")),
			Entry("multiple domains; invalid character; _", "org.one.a_c", validate.ErrorValueStringAsReverseDomainNotValid("org.one.a_c")),
			Entry("multiple domains; starts with dash", "org.one.-bc", validate.ErrorValueStringAsReverseDomainNotValid("org.one.-bc")),
			Entry("multiple domains; ends with dash", "org.one.ab-", validate.ErrorValueStringAsReverseDomainNotValid("org.one.ab-")),
			Entry("multiple domains; ends with dot", "org.one.abc.", validate.ErrorValueStringAsReverseDomainNotValid("org.one.abc.")),
			Entry("length in range (upper)", "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijk.abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz-0123456789.abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz-0123456789.abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz-01234567"),
			Entry("length out of range (upper)", "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijk.abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz-0123456789.abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz-0123456789.abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz-012345678", structureValidator.ErrorLengthNotLessThanOrEqualTo(254, 253)),
		)
	})

	Context("SemanticVersion", func() {
		DescribeTable("validates the semantic version",
			func(value string, expectedErrors ...error) {
				errorReporter := testStructure.NewErrorReporter()
				Expect(errorReporter).ToNot(BeNil())
				validate.SemanticVersion(value, errorReporter)
				testErrors.ExpectEqual(errorReporter.Error(), expectedErrors...)
			},
			Entry("succeeds", "1.2.3"),
			Entry("empty", "", structureValidator.ErrorValueEmpty()),
			Entry("build missing", "1.2", validate.ErrorValueStringAsSemanticVersionNotValid("1.2")),
			Entry("minor and build missing", "1", validate.ErrorValueStringAsSemanticVersionNotValid("1")),
			Entry("v prefix", "v1.2.3", validate.ErrorValueStringAsSemanticVersionNotValid("v1.2.3")),
			Entry("length in range (upper)", "1.2.3-abcdefghijklmnopqrstuvwxyz-abcdefghijklmnopqrstuvwxyz-abcdefghijklmnopqrstuvwxyz-abcdefghijklm"),
			Entry("length out of range (upper)", "1.2.3-abcdefghijklmnopqrstuvwxyz-abcdefghijklmnopqrstuvwxyz-abcdefghijklmnopqrstuvwxyz-abcdefghijklmn", structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100)),
		)
	})

	Context("URL", func() {
		DescribeTable("validates the URL",
			func(value string, expectedErrors ...error) {
				errorReporter := testStructure.NewErrorReporter()
				Expect(errorReporter).ToNot(BeNil())
				validate.URL(value, errorReporter)
				testErrors.ExpectEqual(errorReporter.Error(), expectedErrors...)
			},
			Entry("succeeds", "http://test.org"),
			Entry("empty", "", structureValidator.ErrorValueEmpty()),
			Entry("not parsable", "http:::", validate.ErrorValueStringAsURLNotValid("http:::")),
			Entry("relative", "/relative/path", validate.ErrorValueStringAsURLNotValid("/relative/path")),
			Entry("host missing", "http:///nohost", validate.ErrorValueStringAsURLNotValid("http:///nohost")),
			Entry("length in range (upper)", "http://"+test.NewString(1993, testHTTP.CharsetPath)),
			Entry("length out of range (upper)", "http://"+test.NewString(1994, testHTTP.CharsetPath), structureValidator.ErrorLengthNotLessThanOrEqualTo(2001, 2000)),
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
			Entry("is ErrorValueStringAsReverseDomainNotValid", validate.ErrorValueStringAsReverseDomainNotValid("abc"), "value-not-valid", "value is not valid", `value "abc" is not valid as reverse domain`),
			Entry("is ErrorValueStringAsSemanticVersionNotValid", validate.ErrorValueStringAsSemanticVersionNotValid("abc"), "value-not-valid", "value is not valid", `value "abc" is not valid as semantic version`),
			Entry("is ErrorValueStringAsURLNotValid", validate.ErrorValueStringAsURLNotValid("abc"), "value-not-valid", "value is not valid", `value "abc" is not valid as url`),
		)
	})
})
