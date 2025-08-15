package data_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	structureTest "github.com/tidepool-org/platform/structure/test"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Data", func() {
	Context("SelectorDeduplicator", func() {
		Context("Matches", func() {
			hash := pointer.FromString(data.NewID())

			DescribeTable("return the expected results when the selector deduplicator",
				func(deduplicator *data.SelectorDeduplicator, otherDeduplicator *data.SelectorDeduplicator, expectedResult bool) {
					Expect(deduplicator.Matches(otherDeduplicator)).To(Equal(expectedResult))
				},
				Entry("both are nil", nil, nil, false),
				Entry("deduplicator is nil", nil, &data.SelectorDeduplicator{}, false),
				Entry("other deduplicator is nil", &data.SelectorDeduplicator{}, nil, false),
				Entry("both hashes are nil", &data.SelectorDeduplicator{}, &data.SelectorDeduplicator{}, true),
				Entry("deduplicator hash is nil", &data.SelectorDeduplicator{}, &data.SelectorDeduplicator{Hash: hash}, true),
				Entry("other deduplicator hash is nil", &data.SelectorDeduplicator{Hash: hash}, &data.SelectorDeduplicator{}, false),
				Entry("hash mismatch", &data.SelectorDeduplicator{Hash: hash}, &data.SelectorDeduplicator{Hash: pointer.FromString("mismatch")}, false),
				Entry("hash matches", &data.SelectorDeduplicator{Hash: hash}, &data.SelectorDeduplicator{Hash: hash}, true),
			)
		})

		Context("NewerThan", func() {
			DescribeTable("return the expected results when the selector deduplicator",
				func(deduplicator *data.SelectorDeduplicator, otherDeduplicator *data.SelectorDeduplicator, expectedResult bool) {
					Expect(deduplicator.NewerThan(otherDeduplicator)).To(Equal(expectedResult))
				},
				Entry("both are nil", nil, nil, false),
				Entry("deduplicator is nil", nil, &data.SelectorDeduplicator{}, false),
				Entry("other deduplicator is nil", &data.SelectorDeduplicator{}, nil, false),
				Entry("both are not nil", &data.SelectorDeduplicator{}, &data.SelectorDeduplicator{}, true),
			)
		})
	})

	Context("SelectorOrigin", func() {
		Context("Matches", func() {
			id := pointer.FromString(data.NewID())

			DescribeTable("return the expected results when the selector origin",
				func(origin *data.SelectorOrigin, otherOrigin *data.SelectorOrigin, expectedResult bool) {
					Expect(origin.Matches(otherOrigin)).To(Equal(expectedResult))
				},
				Entry("both are nil", nil, nil, false),
				Entry("origin is nil", nil, &data.SelectorOrigin{}, false),
				Entry("other origin is nil", &data.SelectorOrigin{}, nil, false),
				Entry("id and time are nil", &data.SelectorOrigin{}, &data.SelectorOrigin{}, true),
				Entry("origin id is nil", &data.SelectorOrigin{}, &data.SelectorOrigin{ID: id}, true),
				Entry("other origin id is nil", &data.SelectorOrigin{ID: id}, &data.SelectorOrigin{}, false),
				Entry("id mismatch", &data.SelectorOrigin{ID: id}, &data.SelectorOrigin{ID: pointer.FromString("mismatch")}, false),
				Entry("id includes", &data.SelectorOrigin{ID: id}, &data.SelectorOrigin{ID: id}, true),
			)
		})

		Context("NewerThan", func() {
			now := time.Now()
			tm := pointer.FromString(now.Format(time.RFC3339Nano))

			DescribeTable("return the expected results when the selector origin",
				func(origin *data.SelectorOrigin, otherOrigin *data.SelectorOrigin, expectedResult bool) {
					Expect(origin.NewerThan(otherOrigin)).To(Equal(expectedResult))
				},
				Entry("both are nil", nil, nil, false),
				Entry("origin is nil", nil, &data.SelectorOrigin{}, false),
				Entry("other origin is nil", &data.SelectorOrigin{}, nil, false),
				Entry("both times are nil", &data.SelectorOrigin{}, &data.SelectorOrigin{}, true),
				Entry("origin time is nil", &data.SelectorOrigin{}, &data.SelectorOrigin{Time: tm}, true),
				Entry("other origin time is nil", &data.SelectorOrigin{Time: tm}, &data.SelectorOrigin{}, false),
				Entry("time earlier", &data.SelectorOrigin{Time: tm}, &data.SelectorOrigin{Time: pointer.FromString(now.Add(time.Hour).Format(time.RFC3339Nano))}, false),
				Entry("time same", &data.SelectorOrigin{Time: tm}, &data.SelectorOrigin{Time: tm}, true),
				Entry("time same in different time zone", &data.SelectorOrigin{Time: tm}, &data.SelectorOrigin{Time: pointer.FromString(now.In(time.FixedZone("Etc/GMT-1", int(-time.Hour.Seconds()))).Format(time.RFC3339Nano))}, true),
				Entry("time later", &data.SelectorOrigin{Time: tm}, &data.SelectorOrigin{Time: pointer.FromString(now.Add(-time.Hour).Format(time.RFC3339Nano))}, true),
			)
		})
	})

	Context("Selector", func() {
		Context("Matches", func() {
			id := pointer.FromString(data.NewID())
			deduplicatorHash := pointer.FromString(data.NewID())
			originID := pointer.FromString(data.NewID())

			DescribeTable("return the expected results when the selector",
				func(selector *data.Selector, otherSector *data.Selector, expectedResult bool) {
					Expect(selector.Matches(otherSector)).To(Equal(expectedResult))
				},
				Entry("both are nil", nil, nil, false),
				Entry("selector is nil", nil, &data.Selector{}, false),
				Entry("other selector is nil", &data.Selector{}, nil, false),
				Entry("id, deduplicator, and origin are nil", &data.Selector{}, &data.Selector{}, true),
				Entry("selector id is nil", &data.Selector{}, &data.Selector{ID: id}, true),
				Entry("other selector id is nil", &data.Selector{ID: id}, &data.Selector{}, false),
				Entry("id mismatch", &data.Selector{ID: id}, &data.Selector{ID: pointer.FromString("mismatch")}, false),
				Entry("id includes", &data.Selector{ID: id}, &data.Selector{ID: id}, true),
				Entry("selector deduplicator is nil", &data.Selector{ID: id}, &data.Selector{ID: id, Deduplicator: &data.SelectorDeduplicator{Hash: deduplicatorHash}}, true),
				Entry("other selector deduplicator is nil", &data.Selector{ID: id, Deduplicator: &data.SelectorDeduplicator{Hash: deduplicatorHash}}, &data.Selector{ID: id}, false),
				Entry("deduplicator id mismatch", &data.Selector{ID: id, Deduplicator: &data.SelectorDeduplicator{Hash: deduplicatorHash}}, &data.Selector{ID: id, Deduplicator: &data.SelectorDeduplicator{Hash: pointer.FromString("mismatch")}}, false),
				Entry("deduplicator match", &data.Selector{ID: id, Deduplicator: &data.SelectorDeduplicator{Hash: deduplicatorHash}}, &data.Selector{ID: id, Deduplicator: &data.SelectorDeduplicator{Hash: deduplicatorHash}}, true),
				Entry("selector origin is nil", &data.Selector{ID: id}, &data.Selector{ID: id, Origin: &data.SelectorOrigin{ID: originID}}, true),
				Entry("other selector origin is nil", &data.Selector{ID: id, Origin: &data.SelectorOrigin{ID: originID}}, &data.Selector{ID: id}, false),
				Entry("origin id mismatch", &data.Selector{ID: id, Origin: &data.SelectorOrigin{ID: originID}}, &data.Selector{ID: id, Origin: &data.SelectorOrigin{ID: pointer.FromString("mismatch")}}, false),
				Entry("origin match", &data.Selector{ID: id, Origin: &data.SelectorOrigin{ID: originID}}, &data.Selector{ID: id, Origin: &data.SelectorOrigin{ID: originID}}, true),
			)
		})

		Context("NewerThan", func() {
			now := time.Now()
			tm := pointer.FromTime(now)
			tmString := pointer.FromString(tm.Format(time.RFC3339Nano))

			DescribeTable("return the expected results when the selector",
				func(deduplicator *data.Selector, otherDeduplicator *data.Selector, expectedResult bool) {
					Expect(deduplicator.NewerThan(otherDeduplicator)).To(Equal(expectedResult))
				},
				Entry("both are nil", nil, nil, false),
				Entry("selector is nil", nil, &data.Selector{}, false),
				Entry("other selector is nil", &data.Selector{}, nil, false),
				Entry("time, deduplicator, and origin are nil", &data.Selector{}, &data.Selector{}, true),
				Entry("selector time is nil", &data.Selector{}, &data.Selector{Time: tm}, true),
				Entry("other selector time is nil", &data.Selector{Time: tm}, &data.Selector{}, false),
				Entry("time earlier", &data.Selector{Time: tm}, &data.Selector{Time: pointer.FromTime(now.Add(time.Hour))}, false),
				Entry("time same", &data.Selector{Time: tm}, &data.Selector{Time: tm}, true),
				Entry("time same in different time zone", &data.Selector{Time: tm}, &data.Selector{Time: pointer.FromTime(now.In(time.FixedZone("Etc/GMT-1", int(-time.Hour.Seconds()))))}, true),
				Entry("time later", &data.Selector{Time: tm}, &data.Selector{Time: pointer.FromTime(now.Add(-time.Hour))}, true),
				Entry("selector deduplicator is nil", &data.Selector{}, &data.Selector{Deduplicator: &data.SelectorDeduplicator{}}, true),
				Entry("other selector deduplicator is nil", &data.Selector{Deduplicator: &data.SelectorDeduplicator{}}, &data.Selector{}, false),
				Entry("both selector deduplicator are not nil", &data.Selector{Deduplicator: &data.SelectorDeduplicator{}}, &data.Selector{Deduplicator: &data.SelectorDeduplicator{}}, true),
				Entry("selector origin is nil", &data.Selector{}, &data.Selector{Origin: &data.SelectorOrigin{}}, true),
				Entry("other selector origin is nil", &data.Selector{Origin: &data.SelectorOrigin{}}, &data.Selector{}, false),
				Entry("selector origin time is nil", &data.Selector{}, &data.Selector{Origin: &data.SelectorOrigin{Time: tmString}}, true),
				Entry("other selector origin time is nil", &data.Selector{Origin: &data.SelectorOrigin{Time: tmString}}, &data.Selector{}, false),
				Entry("selector origin time earlier", &data.Selector{Origin: &data.SelectorOrigin{Time: tmString}}, &data.Selector{Origin: &data.SelectorOrigin{Time: pointer.FromString(now.Add(time.Hour).Format(time.RFC3339Nano))}}, false),
				Entry("selector origin time same", &data.Selector{Origin: &data.SelectorOrigin{Time: tmString}}, &data.Selector{Origin: &data.SelectorOrigin{Time: tmString}}, true),
				Entry("selector origin time same in different time zone", &data.Selector{Origin: &data.SelectorOrigin{Time: tmString}}, &data.Selector{Origin: &data.SelectorOrigin{Time: pointer.FromString(now.In(time.FixedZone("Etc/GMT-1", int(-time.Hour.Seconds()))).Format(time.RFC3339Nano))}}, true),
				Entry("selector origin time later", &data.Selector{Origin: &data.SelectorOrigin{Time: tmString}}, &data.Selector{Origin: &data.SelectorOrigin{Time: pointer.FromString(now.Add(-time.Hour).Format(time.RFC3339Nano))}}, true),
			)
		})
	})

	Context("NewID", func() {
		It("returns a string of 32 lowercase hexadecimal characters", func() {
			Expect(data.NewID()).To(MatchRegexp("^[0-9a-f]{32}$"))
		})

		It("returns different IDs for each invocation", func() {
			Expect(data.NewID()).ToNot(Equal(data.NewID()))
		})
	})

	Context("IsValidID, IDValidator, and ValidateID", func() {
		DescribeTable("return the expected results when the input",
			func(value string, expectedErrors ...error) {
				Expect(data.IsValidID(value)).To(Equal(len(expectedErrors) == 0))
				errorReporter := structureTest.NewErrorReporter()
				data.IDValidator(value, errorReporter)
				errorsTest.ExpectEqual(errorReporter.Error(), expectedErrors...)
				errorsTest.ExpectEqual(data.ValidateID(value), expectedErrors...)
			},
			Entry("is empty", "", structureValidator.ErrorValueEmpty()),
			Entry("has string length out of range (lower)", "0123456789abcdef0123456789abcde", data.ErrorValueStringAsIDNotValid("0123456789abcdef0123456789abcde")),
			Entry("has string length in range", test.RandomStringFromRangeAndCharset(32, 32, test.CharsetHexadecimalLowercase)),
			Entry("has string length in range for Jellyfish", test.RandomStringFromRangeAndCharset(32, 32, test.CharsetNumeric+test.CharsetLowercase)),
			Entry("has string length out of range (upper)", "0123456789abcdef0123456789abcdef0", data.ErrorValueStringAsIDNotValid("0123456789abcdef0123456789abcdef0")),
			Entry("has uppercase characters", "0123456789ABCDEF0123456789abcdef", data.ErrorValueStringAsIDNotValid("0123456789ABCDEF0123456789abcdef")),
			Entry("has symbols", "0123456789$%^&*(0123456789abcdef", data.ErrorValueStringAsIDNotValid("0123456789$%^&*(0123456789abcdef")),
			Entry("has whitespace", "0123456789      0123456789abcdef", data.ErrorValueStringAsIDNotValid("0123456789      0123456789abcdef")),
		)
	})

	Context("Errors", func() {
		DescribeTable("have expected details when error",
			errorsTest.ExpectErrorDetails,
			Entry("is ErrorValueStringAsIDNotValid with empty string", data.ErrorValueStringAsIDNotValid(""), "value-not-valid", "value is not valid", `value "" is not valid as data id`),
			Entry("is ErrorValueStringAsIDNotValid with non-empty string", data.ErrorValueStringAsIDNotValid("0123456789abcdef0123456789abcdef"), "value-not-valid", "value is not valid", `value "0123456789abcdef0123456789abcdef" is not valid as data id`),
		)
	})
})
