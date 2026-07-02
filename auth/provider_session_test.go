package auth_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/auth"
	authTest "github.com/tidepool-org/platform/auth/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureTest "github.com/tidepool-org/platform/structure/test"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
	"github.com/tidepool-org/platform/times"
	timesTest "github.com/tidepool-org/platform/times/test"
)

var _ = Describe("ProviderSession", func() {
	Context("ProviderSessionRefresh", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *auth.ProviderSessionRefresh)) {
				datum := authTest.RandomProviderSessionRefresh(test.AllowOptionals())
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, authTest.NewObjectFromProviderSessionRefresh(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, authTest.NewObjectFromProviderSessionRefresh(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *auth.ProviderSessionRefresh) {},
			),
			Entry("empty",
				func(datum *auth.ProviderSessionRefresh) {
					*datum = auth.ProviderSessionRefresh{}
				},
			),
			Entry("all",
				func(datum *auth.ProviderSessionRefresh) {
					datum.TimeRange = timesTest.RandomTimeRange(test.AllowOptionals())
				},
			),
		)

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]any, expectedDatum *auth.ProviderSessionRefresh), expectedErrors ...error) {
					expectedDatum := authTest.RandomProviderSessionRefresh(test.AllowOptionals())
					object := authTest.NewObjectFromProviderSessionRefresh(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					result := &auth.ProviderSessionRefresh{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(result), expectedErrors...)
					Expect(result).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]any, expectedDatum *auth.ProviderSessionRefresh) {},
				),
				Entry("empty",
					func(object map[string]any, expectedDatum *auth.ProviderSessionRefresh) {
						clear(object)
						*expectedDatum = auth.ProviderSessionRefresh{}
					},
				),
				Entry("multiple errors",
					func(object map[string]any, expectedDatum *auth.ProviderSessionRefresh) {
						object["timeRange"] = true
						expectedDatum.TimeRange = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/timeRange"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *auth.ProviderSessionRefresh), expectedErrors ...error) {
					datum := authTest.RandomProviderSessionRefresh(test.AllowOptionals())
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *auth.ProviderSessionRefresh) {},
				),
				Entry("time range missing",
					func(datum *auth.ProviderSessionRefresh) {
						datum.TimeRange = nil
					},
				),
				Entry("time range empty",
					func(datum *auth.ProviderSessionRefresh) {
						datum.TimeRange = &times.TimeRange{}
					},
				),
				Entry("time range invalid",
					func(datum *auth.ProviderSessionRefresh) {
						datum.TimeRange = &times.TimeRange{
							From: pointer.From(time.Time{}),
						}
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/timeRange/from"),
				),
				Entry("time range valid",
					func(datum *auth.ProviderSessionRefresh) {
						datum.TimeRange = timesTest.RandomTimeRange(test.AllowOptionals())
					},
				),
				Entry("multiple errors",
					func(datum *auth.ProviderSessionRefresh) {
						datum.TimeRange = &times.TimeRange{
							From: pointer.From(time.Time{}),
						}
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/timeRange/from"),
				),
			)
		})
	})

	Context("NewProviderSessionID", func() {
		It("returns a string of 32 lowercase hexadecimal characters", func() {
			Expect(auth.NewProviderSessionID()).To(MatchRegexp("^[0-9a-f]{32}$"))
		})

		It("returns different IDs for each invocation", func() {
			Expect(auth.NewProviderSessionID()).ToNot(Equal(auth.NewProviderSessionID()))
		})
	})

	Context("IsValidProviderSessionID, ProviderSessionIDValidator, and ValidateProviderSessionID", func() {
		DescribeTable("return the expected results when the input",
			func(value string, expectedErrors ...error) {
				Expect(auth.IsValidProviderSessionID(value)).To(Equal(len(expectedErrors) == 0))
				errorReporter := structureTest.NewErrorReporter()
				auth.ProviderSessionIDValidator(value, errorReporter)
				errorsTest.ExpectEqual(errorReporter.Error(), expectedErrors...)
				errorsTest.ExpectEqual(auth.ValidateProviderSessionID(value), expectedErrors...)
			},
			Entry("is empty", "", structureValidator.ErrorValueEmpty()),
			Entry("has string length out of range (lower)", "0123456789abcdef0123456789abcde", auth.ErrorValueStringAsProviderSessionIDNotValid("0123456789abcdef0123456789abcde")),
			Entry("has string length in range", test.RandomStringFromRangeAndCharset(32, 32, test.CharsetHexadecimalLowercase)),
			Entry("has string length out of range (upper)", "0123456789abcdef0123456789abcdef0", auth.ErrorValueStringAsProviderSessionIDNotValid("0123456789abcdef0123456789abcdef0")),
			Entry("has uppercase characters", "0123456789ABCDEF0123456789abcdef", auth.ErrorValueStringAsProviderSessionIDNotValid("0123456789ABCDEF0123456789abcdef")),
			Entry("has symbols", "0123456789$%^&*(0123456789abcdef", auth.ErrorValueStringAsProviderSessionIDNotValid("0123456789$%^&*(0123456789abcdef")),
			Entry("has whitespace", "0123456789      0123456789abcdef", auth.ErrorValueStringAsProviderSessionIDNotValid("0123456789      0123456789abcdef")),
		)
	})

	Context("Errors", func() {
		DescribeTable("have expected details when error",
			errorsTest.ExpectErrorDetails,
			Entry("is ErrorValueStringAsProviderSessionIDNotValid with empty string", auth.ErrorValueStringAsProviderSessionIDNotValid(""), "value-not-valid", "value is not valid", `value "" is not valid as provider session id`),
			Entry("is ErrorValueStringAsProviderSessionIDNotValid with non-empty string", auth.ErrorValueStringAsProviderSessionIDNotValid("0123456789abcdef0123456789abcdef"), "value-not-valid", "value is not valid", `value "0123456789abcdef0123456789abcdef" is not valid as provider session id`),
		)
	})
})
