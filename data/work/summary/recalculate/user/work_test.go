package user_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	dataWorkPostprocess "github.com/tidepool-org/platform/data/work/postprocess"
	dataWorkSummaryRecalculateUser "github.com/tidepool-org/platform/data/work/summary/recalculate/user"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/metadata"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
	userTest "github.com/tidepool-org/platform/user/test"
	userWork "github.com/tidepool-org/platform/user/work"
)

var _ = Describe("work", func() {
	It("Type is expected", func() {
		Expect(dataWorkSummaryRecalculateUser.Type).
			To(Equal("org.tidepool.data.summary.recalculate.user"))
	})

	It("ProcessingTimeout is expected", func() {
		Expect(dataWorkSummaryRecalculateUser.ProcessingTimeout).To(Equal(5 * time.Minute))
	})

	It("DeduplicationID combines the recalculation and the user", func() {
		Expect(dataWorkSummaryRecalculateUser.DeduplicationID("recalculation", "user")).
			To(Equal("recalculation:user"))
	})

	Context("NewWorkCreate", func() {
		var recalculationID string
		var userID string

		BeforeEach(func() {
			recalculationID = test.RandomStringFromCharset(test.CharsetAlpha)
			userID = userTest.RandomUserID()
		})

		It("returns an error if the recalculation id is missing", func() {
			workCreate, err := dataWorkSummaryRecalculateUser.NewWorkCreate("", userID)
			Expect(err).To(MatchError("recalculation id is missing"))
			Expect(workCreate).To(BeNil())
		})

		It("returns an error if the user id is missing", func() {
			workCreate, err :=
				dataWorkSummaryRecalculateUser.NewWorkCreate(recalculationID, "")
			Expect(err).To(MatchError("user id is missing"))
			Expect(workCreate).To(BeNil())
		})

		// The serial id is that of the postprocess work of the user, so that the two are
		// never processed at once
		It("returns work for the user, serialized with their postprocess work", func() {
			workCreate, err := dataWorkSummaryRecalculateUser.NewWorkCreate(recalculationID,
				userID)
			Expect(err).ToNot(HaveOccurred())
			deduplicationID :=
				dataWorkSummaryRecalculateUser.DeduplicationID(recalculationID, userID)
			serialID := dataWorkPostprocess.IDFromUserID(userID)
			Expect(workCreate).To(PointTo(MatchAllFields(Fields{
				"Type":                    Equal(dataWorkSummaryRecalculateUser.Type),
				"GroupID":                 BeNil(),
				"DeduplicationID":         PointTo(Equal(deduplicationID)),
				"SerialID":                PointTo(Equal(serialID)),
				"ProcessingAvailableTime": BeTemporally("~", time.Now(), time.Second),
				"ProcessingPriority":      Equal(0),
				"ProcessingTimeout":       Equal(300),
				"Metadata":                Equal(map[string]any{"userId": userID}),
			})))
		})
	})

	Context("Metadata", func() {
		type Metadata = dataWorkSummaryRecalculateUser.Metadata

		newUserMetadata := func() userWork.Metadata {
			return userWork.Metadata{UserID: pointer.From(userTest.RandomUserID())}
		}

		It("encodes and decodes", func() {
			ctx := log.NewContextWithLogger(GinkgoT().Context(), logTest.NewLogger())
			datum := &Metadata{
				Metadata:              newUserMetadata(),
				PendingSummaryUpdates: []string{"bgm", "cgm"},
				PendingSummaryDeletes: []string{"68ad3f439bd2caa1a5758a9c"},
			}

			encoded, err := metadata.Encode(datum)
			Expect(err).ToNot(HaveOccurred())
			decoded, err := metadata.Decode[Metadata](ctx, encoded)
			Expect(err).ToNot(HaveOccurred())
			Expect(decoded).To(Equal(datum))
		})

		DescribeTable("is invalid",
			func(mutate func(datum *Metadata)) {
				ctx := log.NewContextWithLogger(GinkgoT().Context(), logTest.NewLogger())
				datum := &Metadata{Metadata: newUserMetadata()}
				mutate(datum)

				encoded, err := metadata.Encode(datum)
				Expect(err).ToNot(HaveOccurred())
				_, err = metadata.Decode[Metadata](ctx, encoded)
				Expect(err).To(HaveOccurred())
			},
			Entry("with a pending update of a summary type that isn't recalculated",
				func(datum *Metadata) {
					datum.PendingSummaryUpdates = []string{"con"}
				}),
			Entry("with a duplicate pending update",
				func(datum *Metadata) {
					datum.PendingSummaryUpdates = []string{"cgm", "cgm"}
				}),
			Entry("with an empty pending delete",
				func(datum *Metadata) {
					datum.PendingSummaryDeletes = []string{""}
				}),
		)
	})
})
