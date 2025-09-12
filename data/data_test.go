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
				Entry("both hashes are nil", &data.SelectorDeduplicator{}, &data.SelectorDeduplicator{}, false),
				Entry("deduplicator hash is nil", &data.SelectorDeduplicator{}, &data.SelectorDeduplicator{Hash: hash}, false),
				Entry("other deduplicator hash is nil", &data.SelectorDeduplicator{Hash: hash}, &data.SelectorDeduplicator{}, false),
				Entry("hash mismatch", &data.SelectorDeduplicator{Hash: hash}, &data.SelectorDeduplicator{Hash: pointer.FromString("mismatch")}, false),
				Entry("hash match", &data.SelectorDeduplicator{Hash: hash}, &data.SelectorDeduplicator{Hash: hash}, true),
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
				Entry("id is nil", &data.SelectorOrigin{}, &data.SelectorOrigin{}, false),
				Entry("origin id is nil", &data.SelectorOrigin{}, &data.SelectorOrigin{ID: id}, false),
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
				Entry("both times are nil", &data.SelectorOrigin{}, &data.SelectorOrigin{}, false),
				Entry("origin is nil", nil, &data.SelectorOrigin{Time: tm}, false),
				Entry("origin time is nil", &data.SelectorOrigin{}, &data.SelectorOrigin{Time: tm}, false),
				Entry("origin time is unparsable", &data.SelectorOrigin{Time: pointer.FromString("unparsable")}, &data.SelectorOrigin{Time: tm}, false),
				Entry("other origin is nil", &data.SelectorOrigin{Time: tm}, nil, true),
				Entry("other origin time is nil", &data.SelectorOrigin{Time: tm}, &data.SelectorOrigin{}, true),
				Entry("other origin time is unparsable", &data.SelectorOrigin{Time: tm}, &data.SelectorOrigin{Time: pointer.FromString("unparsable")}, true),
				Entry("time earlier", &data.SelectorOrigin{Time: tm}, &data.SelectorOrigin{Time: pointer.FromString(now.Add(time.Hour).Format(time.RFC3339Nano))}, false),
				Entry("time same", &data.SelectorOrigin{Time: tm}, &data.SelectorOrigin{Time: tm}, false),
				Entry("time same in different time zone", &data.SelectorOrigin{Time: tm}, &data.SelectorOrigin{Time: pointer.FromString(now.In(time.FixedZone("Etc/GMT-1", int(-time.Hour.Seconds()))).Format(time.RFC3339Nano))}, false),
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
				Entry("id, deduplicator, and origin are nil", &data.Selector{}, &data.Selector{}, false),
				Entry("other selector id is nil", &data.Selector{ID: id}, &data.Selector{}, false),
				Entry("other selector id mismatch", &data.Selector{ID: id}, &data.Selector{ID: pointer.FromString("mismatch")}, false),
				Entry("other selector id match", &data.Selector{ID: id}, &data.Selector{ID: id}, true),
				Entry("other selector deduplicator is nil", &data.Selector{Deduplicator: &data.SelectorDeduplicator{Hash: deduplicatorHash}}, &data.Selector{}, false),
				Entry("other selector deduplicator id mismatch", &data.Selector{Deduplicator: &data.SelectorDeduplicator{Hash: deduplicatorHash}}, &data.Selector{Deduplicator: &data.SelectorDeduplicator{Hash: pointer.FromString("mismatch")}}, false),
				Entry("other selector deduplicator match", &data.Selector{Deduplicator: &data.SelectorDeduplicator{Hash: deduplicatorHash}}, &data.Selector{Deduplicator: &data.SelectorDeduplicator{Hash: deduplicatorHash}}, true),
				Entry("other selector origin is nil", &data.Selector{Origin: &data.SelectorOrigin{ID: originID}}, &data.Selector{}, false),
				Entry("other selector origin id mismatch", &data.Selector{Origin: &data.SelectorOrigin{ID: originID}}, &data.Selector{Origin: &data.SelectorOrigin{ID: pointer.FromString("mismatch")}}, false),
				Entry("other selector origin match", &data.Selector{Origin: &data.SelectorOrigin{ID: originID}}, &data.Selector{Origin: &data.SelectorOrigin{ID: originID}}, true),
			)
		})

		Context("NewerThan", func() {
			now := time.Now()
			id := pointer.FromString(data.NewID())
			tm := pointer.FromTime(now)
			tmString := pointer.FromString(tm.Format(time.RFC3339Nano))

			DescribeTable("return the expected results when the selector",
				func(deduplicator *data.Selector, otherDeduplicator *data.Selector, expectedResult bool) {
					Expect(deduplicator.NewerThan(otherDeduplicator)).To(Equal(expectedResult))
				},
				Entry("both are nil", nil, nil, false),
				Entry("selector is nil", nil, &data.Selector{}, false),
				Entry("other selector is nil", &data.Selector{}, nil, true),
				Entry("id, deduplicator, and origin are nil", &data.Selector{}, &data.Selector{}, false),
				Entry("selector time is nil", &data.Selector{ID: id}, &data.Selector{Time: tm}, false),
				Entry("other selector time is nil", &data.Selector{ID: id, Time: tm}, &data.Selector{}, true),
				Entry("selector time earlier", &data.Selector{ID: id, Time: tm}, &data.Selector{Time: pointer.FromTime(now.Add(time.Hour))}, false),
				Entry("selector time same", &data.Selector{ID: id, Time: tm}, &data.Selector{Time: tm}, false),
				Entry("selector time same in different time zone", &data.Selector{ID: id, Time: tm}, &data.Selector{Time: pointer.FromTime(now.In(time.FixedZone("Etc/GMT-1", int(-time.Hour.Seconds()))))}, false),
				Entry("selector time later", &data.Selector{ID: id, Time: tm}, &data.Selector{Time: pointer.FromTime(now.Add(-time.Hour))}, true),
				Entry("other selector deduplicator is nil", &data.Selector{Deduplicator: &data.SelectorDeduplicator{}}, &data.Selector{}, true),
				Entry("other selector deduplicator is not nil", &data.Selector{Deduplicator: &data.SelectorDeduplicator{}}, &data.Selector{Deduplicator: &data.SelectorDeduplicator{}}, true),
				Entry("other selector origin is nil", &data.Selector{Origin: &data.SelectorOrigin{}}, &data.Selector{}, true),
				Entry("selector origin time is nil", &data.Selector{}, &data.Selector{Origin: &data.SelectorOrigin{Time: tmString}}, false),
				Entry("other selector origin time is nil", &data.Selector{Origin: &data.SelectorOrigin{Time: tmString}}, &data.Selector{}, true),
				Entry("selector origin time earlier", &data.Selector{Origin: &data.SelectorOrigin{Time: tmString}}, &data.Selector{Origin: &data.SelectorOrigin{Time: pointer.FromString(now.Add(time.Hour).Format(time.RFC3339Nano))}}, false),
				Entry("selector origin time same", &data.Selector{Origin: &data.SelectorOrigin{Time: tmString}}, &data.Selector{Origin: &data.SelectorOrigin{Time: tmString}}, false),
				Entry("selector origin time same in different time zone", &data.Selector{Origin: &data.SelectorOrigin{Time: tmString}}, &data.Selector{Origin: &data.SelectorOrigin{Time: pointer.FromString(now.In(time.FixedZone("Etc/GMT-1", int(-time.Hour.Seconds()))).Format(time.RFC3339Nano))}}, false),
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
