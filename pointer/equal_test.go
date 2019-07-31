package pointer_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Equal", func() {
	DescribeTable("EqualBool returns expected",
		func(a *bool, b *bool, expected bool) {
			Expect(pointer.EqualBool(a, b)).To(Equal(expected))
		},
		Entry("a is missing; b is missing", nil, nil, true),
		Entry("a is missing; b is false", nil, pointer.FromBool(false), false),
		Entry("a is missing; b is true", nil, pointer.FromBool(true), false),
		Entry("a is false; b is missing", pointer.FromBool(false), nil, false),
		Entry("a is false; b is false", pointer.FromBool(false), pointer.FromBool(false), true),
		Entry("a is false; b is true", pointer.FromBool(false), pointer.FromBool(true), false),
		Entry("a is true; b is missing", pointer.FromBool(true), nil, false),
		Entry("a is true; b is false", pointer.FromBool(true), pointer.FromBool(false), false),
		Entry("a is true; b is true", pointer.FromBool(true), pointer.FromBool(true), true),
	)

	DescribeTable("EqualDuration returns expected",
		func(a *time.Duration, b *time.Duration, expected bool) {
			Expect(pointer.EqualDuration(a, b)).To(Equal(expected))
		},
		Entry("a is missing; b is missing", nil, nil, true),
		Entry("a is missing; b is present", nil, pointer.FromDuration(456*time.Second), false),
		Entry("a is present; b is missing", pointer.FromDuration(123*time.Second), nil, false),
		Entry("a is present; b is present", pointer.FromDuration(123*time.Second), pointer.FromDuration(456*time.Second), false),
		Entry("a is present; b is present and match", pointer.FromDuration(123*time.Second), pointer.FromDuration(123*time.Second), true),
	)

	DescribeTable("EqualFloat64 returns expected",
		func(a *float64, b *float64, expected bool) {
			Expect(pointer.EqualFloat64(a, b)).To(Equal(expected))
		},
		Entry("a is missing; b is missing", nil, nil, true),
		Entry("a is missing; b is present", nil, pointer.FromFloat64(4.56), false),
		Entry("a is present; b is missing", pointer.FromFloat64(1.23), nil, false),
		Entry("a is present; b is present", pointer.FromFloat64(1.23), pointer.FromFloat64(4.56), false),
		Entry("a is present; b is present and match", pointer.FromFloat64(1.23), pointer.FromFloat64(1.23), true),
	)

	DescribeTable("EqualString returns expected",
		func(a *int, b *int, expected bool) {
			Expect(pointer.EqualInt(a, b)).To(Equal(expected))
		},
		Entry("a is missing; b is missing", nil, nil, true),
		Entry("a is missing; b is present", nil, pointer.FromInt(456), false),
		Entry("a is present; b is missing", pointer.FromInt(123), nil, false),
		Entry("a is present; b is present", pointer.FromInt(123), pointer.FromInt(456), false),
		Entry("a is present; b is present and match", pointer.FromInt(123), pointer.FromInt(123), true),
	)

	DescribeTable("EqualString returns expected",
		func(a *string, b *string, expected bool) {
			Expect(pointer.EqualString(a, b)).To(Equal(expected))
		},
		Entry("a is missing; b is missing", nil, nil, true),
		Entry("a is missing; b is present", nil, pointer.FromString("def"), false),
		Entry("a is present; b is missing", pointer.FromString("abc"), nil, false),
		Entry("a is present; b is present", pointer.FromString("abc"), pointer.FromString("def"), false),
		Entry("a is present; b is present and match", pointer.FromString("abc"), pointer.FromString("abc"), true),
	)

	DescribeTable("EqualStringArray returns expected",
		func(a *[]string, b *[]string, expected bool) {
			Expect(pointer.EqualStringArray(a, b)).To(Equal(expected))
		},
		Entry("a is missing; b is missing", nil, nil, true),
		Entry("a is missing; b is empty", nil, pointer.FromStringArray([]string{}), false),
		Entry("a is missing; b is present", nil, pointer.FromStringArray([]string{"d", "e", "f"}), false),
		Entry("a is empty; b is missing", pointer.FromStringArray([]string{}), nil, false),
		Entry("a is empty; b is empty", pointer.FromStringArray([]string{}), pointer.FromStringArray([]string{}), true),
		Entry("a is empty; b is present", pointer.FromStringArray([]string{}), pointer.FromStringArray([]string{"d", "e", "f"}), false),
		Entry("a is present; b is missing", pointer.FromStringArray([]string{"a", "b", "c"}), nil, false),
		Entry("a is present; b is empty", pointer.FromStringArray([]string{"a", "b", "c"}), pointer.FromStringArray([]string{}), false),
		Entry("a is present; b is present", pointer.FromStringArray([]string{"a", "b", "c"}), pointer.FromStringArray([]string{"d", "e", "f"}), false),
		Entry("a is present; b is present and subset", pointer.FromStringArray([]string{"a", "b", "c"}), pointer.FromStringArray([]string{"a", "b"}), false),
		Entry("a is present; b is present and superset", pointer.FromStringArray([]string{"a", "b", "c"}), pointer.FromStringArray([]string{"a", "b", "c", "d"}), false),
		Entry("a is present; b is present and match", pointer.FromStringArray([]string{"a", "b", "c"}), pointer.FromStringArray([]string{"a", "b", "c"}), true),
	)

	DescribeTable("EqualTime returns expected",
		func(a *time.Time, b *time.Time, expected bool) {
			Expect(pointer.EqualTime(a, b)).To(Equal(expected))
		},
		Entry("a is missing; b is missing", nil, nil, true),
		Entry("a is missing; b is present", nil, pointer.FromTime(test.PastNearTime()), false),
		Entry("a is present; b is missing", pointer.FromTime(test.PastNearTime()), nil, false),
		Entry("a is present; b is present", pointer.FromTime(test.PastNearTime()), pointer.FromTime(test.FutureNearTime()), false),
		Entry("a is present; b is present and match", pointer.FromTime(test.PastNearTime()), pointer.FromTime(test.PastNearTime()), true),
		Entry("a is present; b is present and match with Local time zone", pointer.FromTime(test.PastNearTime().UTC()), pointer.FromTime(test.PastNearTime()), true),
		Entry("a is present; b is present and match with UTC time zone", pointer.FromTime(test.PastNearTime()), pointer.FromTime(test.PastNearTime().UTC()), true),
	)
})
