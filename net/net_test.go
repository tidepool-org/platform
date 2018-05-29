package net_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/net"
	structureTest "github.com/tidepool-org/platform/structure/test"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
	testHttp "github.com/tidepool-org/platform/test/http"
)

var _ = Describe("Validate", func() {
	Context("IsValidMediaType, MediaTypeValidator, and ValidateMediaType", func() {
		DescribeTable("return the expected results when the input",
			func(value string, expectedErrors ...error) {
				Expect(net.IsValidMediaType(value)).To(Equal(len(expectedErrors) == 0))
				errorReporter := structureTest.NewErrorReporter()
				net.MediaTypeValidator(value, errorReporter)
				errorsTest.ExpectEqual(errorReporter.Error(), expectedErrors...)
				errorsTest.ExpectEqual(net.ValidateMediaType(value), expectedErrors...)
			},
			Entry("is empty", "", structureValidator.ErrorValueEmpty()),
			Entry("has valid media type", "text/plain"),
			Entry("has valid media type with whitespace", "  text/plain  "),
			Entry("has invalid media type with only slash", "/", net.ErrorValueStringAsMediaTypeNotValid("/")),
			Entry("has invalid media type with only type", "text/", net.ErrorValueStringAsMediaTypeNotValid("text/")),
			Entry("has invalid media type with only subtype", "/plain", net.ErrorValueStringAsMediaTypeNotValid("/plain")),
			Entry("has valid media type with parameter", "text/plain; x=y"),
			Entry("has valid media type with multiple parameters", "text/plain; x=y; y=z"),
			Entry("has valid media type with parameter with whitespace", " text/plain  ; x = y ; y = z "),
			Entry("has valid media type with parameter with trailing semicolon", "text/plain; x=y; y=z; "),
			Entry("has valid media type with duplicate parameter", "text/plain; x=y; x=y", net.ErrorValueStringAsMediaTypeNotValid("text/plain; x=y; x=y")),
			Entry("has valid media type with invalid parameter", "text/plain; x", net.ErrorValueStringAsMediaTypeNotValid("text/plain; x")),
			Entry("has length in range (upper)", "text/plain; x="+test.NewString(242, test.CharsetAlphaNumeric)),
			Entry("has length out of range (upper)", "text/plain; x="+test.NewString(243, test.CharsetAlphaNumeric), structureValidator.ErrorLengthNotLessThanOrEqualTo(257, 256)),
		)
	})

	Context("NormalizeMediaType", func() {
		DescribeTable("returns the expected results when",
			func(value string, expectedResult string, expectedOk bool) {
				result, ok := net.NormalizeMediaType(value)
				Expect(ok).To(Equal(expectedOk))
				Expect(result).To(Equal(expectedResult))
			},
			Entry("is empty", "", "", false),
			Entry("has valid media type", "text/plain", "text/plain", true),
			Entry("has valid media type with uppercase", "TEXT/PLAIN", "text/plain", true),
			Entry("has valid media type with whitespace", "  text/plain  ", "text/plain", true),
			Entry("has invalid media type with only slash", "/", "", false),
			Entry("has invalid media type with only type", "text/", "", false),
			Entry("has invalid media type with only subtype", "/plain", "", false),
			Entry("has valid media type with parameter", "text/plain; x=y", "text/plain; x=y", true),
			Entry("has valid media type with parameter with key uppercase", "text/plain; X=y", "text/plain; x=y", true),
			Entry("has valid media type with multiple parameters", "text/plain; X=Y; Y=Z", "text/plain; x=Y; y=Z", true),
			Entry("has valid media type with parameter with whitespace", " text/plain  ; X = y ; y = Z ", "text/plain; x=y; y=Z", true),
			Entry("has valid media type with parameter with trailing semicolon", "text/plain; x=y; y=z; ", "text/plain; x=y; y=z", true),
			Entry("has valid media type with duplicate parameter", "text/plain; x=y; X=y", "", false),
			Entry("has valid media type with invalid parameter", "text/plain; x", "", false),
		)
	})

	Context("IsValidReverseDomain, ReverseDomainValidator, and ValidateReverseDomain", func() {
		DescribeTable("return the expected results when the input",
			func(value string, expectedErrors ...error) {
				Expect(net.IsValidReverseDomain(value)).To(Equal(len(expectedErrors) == 0))
				errorReporter := structureTest.NewErrorReporter()
				net.ReverseDomainValidator(value, errorReporter)
				errorsTest.ExpectEqual(errorReporter.Error(), expectedErrors...)
				errorsTest.ExpectEqual(net.ValidateReverseDomain(value), expectedErrors...)
			},
			Entry("is valid with single domain", "top.one"),
			Entry("is valid with multiple domains", "top.one.two.three"),
			Entry("is empty", "", structureValidator.ErrorValueEmpty()),
			Entry("has tld; out of range (lower)", "a.one.two.three", net.ErrorValueStringAsReverseDomainNotValid("a.one.two.three")),
			Entry("has tld; in range (lower)", "ab.one.two.three"),
			Entry("has tld; in range (upper)", "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijk.one.two.three"),
			Entry("has tld; out of range (upper)", "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijkl.one.two.three", net.ErrorValueStringAsReverseDomainNotValid("abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijkl.one.two.three")),
			Entry("has tld; invalid character; uppercase", "aBc.one.two.three", net.ErrorValueStringAsReverseDomainNotValid("aBc.one.two.three")),
			Entry("has tld; invalid character; -", "a-c.one.two.three", net.ErrorValueStringAsReverseDomainNotValid("a-c.one.two.three")),
			Entry("has tld; invalid character; _", "a_c.one.two.three", net.ErrorValueStringAsReverseDomainNotValid("a_c.one.two.three")),
			Entry("has tld; invalid character; 0", "a0c.one.two.three", net.ErrorValueStringAsReverseDomainNotValid("a0c.one.two.three")),
			Entry("has tld; only", "abc", net.ErrorValueStringAsReverseDomainNotValid("abc")),
			Entry("has tld; trailing dot", "abc.", net.ErrorValueStringAsReverseDomainNotValid("abc.")),
			Entry("has single domain; out of range (lower)", "org..two", net.ErrorValueStringAsReverseDomainNotValid("org..two")),
			Entry("has single domain; in range (lower)", "org.a"),
			Entry("has single domain; in range (lower); multiple", "org.ab"),
			Entry("has single domain; in range (upper)", "org.abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz-1234567890"),
			Entry("has single domain; out of range (upper)", "org.abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz-1234567890a", net.ErrorValueStringAsReverseDomainNotValid("org.abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz-1234567890a")),
			Entry("has single domain; invalid character; _", "org.a_c", net.ErrorValueStringAsReverseDomainNotValid("org.a_c")),
			Entry("has single domain; starts with dash", "org.-bc", net.ErrorValueStringAsReverseDomainNotValid("org.-bc")),
			Entry("has single domain; ends with dash", "org.ab-", net.ErrorValueStringAsReverseDomainNotValid("org.ab-")),
			Entry("has single domain; ends with dot", "org.abc.", net.ErrorValueStringAsReverseDomainNotValid("org.abc.")),
			Entry("has multiple domains; out of range (lower)", "org.one..three", net.ErrorValueStringAsReverseDomainNotValid("org.one..three")),
			Entry("has multiple domains; in range (lower)", "org.one.a"),
			Entry("has multiple domains; in range (lower); multiple", "org.one.ab"),
			Entry("has multiple domains; in range (upper)", "org.one.abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz-1234567890"),
			Entry("has multiple domains; out of range (upper)", "org.one.abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz-1234567890a", net.ErrorValueStringAsReverseDomainNotValid("org.one.abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz-1234567890a")),
			Entry("has multiple domains; invalid character; _", "org.one.a_c", net.ErrorValueStringAsReverseDomainNotValid("org.one.a_c")),
			Entry("has multiple domains; starts with dash", "org.one.-bc", net.ErrorValueStringAsReverseDomainNotValid("org.one.-bc")),
			Entry("has multiple domains; ends with dash", "org.one.ab-", net.ErrorValueStringAsReverseDomainNotValid("org.one.ab-")),
			Entry("has multiple domains; ends with dot", "org.one.abc.", net.ErrorValueStringAsReverseDomainNotValid("org.one.abc.")),
			Entry("has length in range (upper)", "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijk.abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz-0123456789.abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz-0123456789.abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz-01234567"),
			Entry("has length out of range (upper)", "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijk.abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz-0123456789.abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz-0123456789.abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz-012345678", structureValidator.ErrorLengthNotLessThanOrEqualTo(254, 253)),
		)
	})

	Context("IsValidSemanticVersion, SemanticVersionValidator, and ValidateSemanticVersion", func() {
		DescribeTable("return the expected results when the input",
			func(value string, expectedErrors ...error) {
				Expect(net.IsValidSemanticVersion(value)).To(Equal(len(expectedErrors) == 0))
				errorReporter := structureTest.NewErrorReporter()
				net.SemanticVersionValidator(value, errorReporter)
				errorsTest.ExpectEqual(errorReporter.Error(), expectedErrors...)
				errorsTest.ExpectEqual(net.ValidateSemanticVersion(value), expectedErrors...)
			},
			Entry("is valid", "1.2.3"),
			Entry("is empty", "", structureValidator.ErrorValueEmpty()),
			Entry("has build missing", "1.2", net.ErrorValueStringAsSemanticVersionNotValid("1.2")),
			Entry("has minor and build missing", "1", net.ErrorValueStringAsSemanticVersionNotValid("1")),
			Entry("has v prefix", "v1.2.3", net.ErrorValueStringAsSemanticVersionNotValid("v1.2.3")),
			Entry("has length in range (upper)", "1.2.3-"+test.NewString(250, test.CharsetAlphaNumeric)),
			Entry("has length out of range (upper)", "1.2.3-"+test.NewString(251, test.CharsetAlphaNumeric), structureValidator.ErrorLengthNotLessThanOrEqualTo(257, 256)),
		)
	})

	Context("IsValidURL, URLValidator, and ValidateURL", func() {
		DescribeTable("return the expected results when the input",
			func(value string, expectedErrors ...error) {
				Expect(net.IsValidURL(value)).To(Equal(len(expectedErrors) == 0))
				errorReporter := structureTest.NewErrorReporter()
				net.URLValidator(value, errorReporter)
				errorsTest.ExpectEqual(errorReporter.Error(), expectedErrors...)
				errorsTest.ExpectEqual(net.ValidateURL(value), expectedErrors...)
			},
			Entry("is valid", "http://test.org"),
			Entry("is empty", "", structureValidator.ErrorValueEmpty()),
			Entry("is not parsable", "http:::", net.ErrorValueStringAsURLNotValid("http:::")),
			Entry("is relative", "/relative/path", net.ErrorValueStringAsURLNotValid("/relative/path")),
			Entry("has host missing", "http:///nohost", net.ErrorValueStringAsURLNotValid("http:///nohost")),
			Entry("has length in range (upper)", "http://"+test.NewString(2040, testHttp.CharsetPath)),
			Entry("has length out of range (upper)", "http://"+test.NewString(2041, testHttp.CharsetPath), structureValidator.ErrorLengthNotLessThanOrEqualTo(2048, 2047)),
		)
	})

	Context("Errors", func() {
		DescribeTable("have expected details when error",
			errorsTest.ExpectErrorDetails,
			Entry("is ErrorValueStringAsMediaTypeNotValid with empty string", net.ErrorValueStringAsMediaTypeNotValid(""), "value-not-valid", "value is not valid", `value "" is not valid as media type`),
			Entry("is ErrorValueStringAsMediaTypeNotValid with non-empty string", net.ErrorValueStringAsMediaTypeNotValid("text/plain"), "value-not-valid", "value is not valid", `value "text/plain" is not valid as media type`),
			Entry("is ErrorValueStringAsReverseDomainNotValid with empty string", net.ErrorValueStringAsReverseDomainNotValid(""), "value-not-valid", "value is not valid", `value "" is not valid as reverse domain`),
			Entry("is ErrorValueStringAsReverseDomainNotValid with non-empty string", net.ErrorValueStringAsReverseDomainNotValid("top.one.two.three"), "value-not-valid", "value is not valid", `value "top.one.two.three" is not valid as reverse domain`),
			Entry("is ErrorValueStringAsSemanticVersionNotValid with empty string", net.ErrorValueStringAsSemanticVersionNotValid(""), "value-not-valid", "value is not valid", `value "" is not valid as semantic version`),
			Entry("is ErrorValueStringAsSemanticVersionNotValid with non-empty string", net.ErrorValueStringAsSemanticVersionNotValid("1.2.3"), "value-not-valid", "value is not valid", `value "1.2.3" is not valid as semantic version`),
			Entry("is ErrorValueStringAsURLNotValid with empty string", net.ErrorValueStringAsURLNotValid(""), "value-not-valid", "value is not valid", `value "" is not valid as url`),
			Entry("is ErrorValueStringAsURLNotValid with non-empty string", net.ErrorValueStringAsURLNotValid("http://test.org"), "value-not-valid", "value is not valid", `value "http://test.org" is not valid as url`),
		)
	})
})
