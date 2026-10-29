package work_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
	"github.com/tidepool-org/platform/work"
	workTest "github.com/tidepool-org/platform/work/test"
)

var _ = Describe("Work", func() {
	Context("Work", func() {
		var wrk *work.Work
		var deduplicationID string
		var releasedDeduplicationID string

		BeforeEach(func() {
			wrk = workTest.RandomWork()
			deduplicationID = workTest.RandomDeduplicationID()
			wrk.DeduplicationID = pointer.FromString(deduplicationID)
			releasedDeduplicationID = deduplicationID + work.ReleasedSuffix + wrk.ID
		})

		Context("IsDeduplicationIDReleased", func() {
			// Work without a deduplication id holds nothing to release
			It("returns true when the work has no deduplication id", func() {
				wrk.DeduplicationID = nil
				Expect(wrk.IsDeduplicationIDReleased()).To(BeTrue())
			})

			It("returns false when the deduplication id does not end with the work id", func() {
				Expect(wrk.IsDeduplicationIDReleased()).To(BeFalse())
			})

			// The id of another work is not a release, the suffix must name this work
			It("returns false when the deduplication id ends with the id of another work", func() {
				wrk.DeduplicationID = pointer.FromString(deduplicationID + work.ReleasedSuffix + workTest.RandomID())
				Expect(wrk.IsDeduplicationIDReleased()).To(BeFalse())
			})

			It("returns true when the deduplication id ends with the work id", func() {
				wrk.DeduplicationID = pointer.FromString(releasedDeduplicationID)
				Expect(wrk.IsDeduplicationIDReleased()).To(BeTrue())
			})
		})

		Context("ReleasedDeduplicationID", func() {
			It("returns nil when the work has no deduplication id", func() {
				wrk.DeduplicationID = nil
				Expect(wrk.ReleasedDeduplicationID()).To(BeNil())
			})

			It("returns nil when the deduplication id is already released", func() {
				wrk.DeduplicationID = pointer.FromString(releasedDeduplicationID)
				Expect(wrk.ReleasedDeduplicationID()).To(BeNil())
			})

			It("returns the deduplication id suffixed by the work id", func() {
				Expect(wrk.ReleasedDeduplicationID()).To(PointTo(Equal(releasedDeduplicationID)))
			})

			It("returns an id that reports as released", func() {
				wrk.DeduplicationID = wrk.ReleasedDeduplicationID()
				Expect(wrk.IsDeduplicationIDReleased()).To(BeTrue())
			})

			// The work id alone makes the released id unique, so the original is sacrificed to keep
			// the whole within the maximum length
			It("truncates the deduplication id to fit the work id within the maximum length", func() {
				longDeduplicationID := test.RandomStringFromRangeAndCharset(work.DeduplicationIDLengthMaximum, work.DeduplicationIDLengthMaximum, test.CharsetAlphaNumeric)
				wrk.DeduplicationID = pointer.FromString(longDeduplicationID)

				released := wrk.ReleasedDeduplicationID()
				Expect(released).To(PointTo(HaveLen(work.DeduplicationIDLengthMaximum)))
				Expect(*released).To(HavePrefix(longDeduplicationID[:work.DeduplicationIDLengthMaximum-len(work.ReleasedSuffix)-len(wrk.ID)]))
				Expect(*released).To(HaveSuffix(work.ReleasedSuffix + wrk.ID))
			})

			// Length is measured in runes, as the validator does, so a multibyte original is not
			// truncated by byte count
			It("truncates by rune rather than byte", func() {
				longDeduplicationID := test.RandomStringFromRangeAndCharset(work.DeduplicationIDLengthMaximum, work.DeduplicationIDLengthMaximum, test.CharsetUnicode)
				wrk.DeduplicationID = pointer.FromString(longDeduplicationID)

				released := wrk.ReleasedDeduplicationID()
				Expect(released).ToNot(BeNil())
				Expect([]rune(*released)).To(HaveLen(work.DeduplicationIDLengthMaximum))
				Expect(*released).To(HaveSuffix(work.ReleasedSuffix + wrk.ID))
			})
		})
	})
})
