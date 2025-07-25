package types_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	. "github.com/tidepool-org/platform/summary/test"
	. "github.com/tidepool-org/platform/summary/types"
)

var _ = Describe("Buckets", func() {
	var bucketTime time.Time
	var err error
	var userId string
	var bucketFactory BucketFactoryFn[*GlucoseBucket, GlucoseBucket]

	BeforeEach(func() {
		now := time.Now()
		userId = "1234"
		bucketTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		bucketFactory = CreateBucketForUser[*GlucoseBucket](userId, SummaryTypeCGM)
	})

	Context("bucketsByTime.Update", func() {
		var userBuckets BucketsByTime[*GlucoseBucket, GlucoseBucket]
		var cgmDatums []data.Datum

		It("With no existing buckets", func() {
			datumTime := bucketTime.Add(5 * time.Minute)
			userBuckets = BucketsByTime[*GlucoseBucket, GlucoseBucket]{}
			cgmDatums = []data.Datum{NewContinuousGlucoseWithValue(datumTime, InTargetBloodGlucose)}

			err = userBuckets.Update(bucketFactory, cgmDatums)
			Expect(err).ToNot(HaveOccurred())
			Expect(userBuckets).To(HaveKey(bucketTime))
			Expect(userBuckets[bucketTime].Type).To(Equal(SummaryTypeCGM))
		})

		It("Adding to existing buckets", func() {
			datumTime := bucketTime.Add(5 * time.Minute)
			userBuckets = BucketsByTime[*GlucoseBucket, GlucoseBucket]{}
			cgmDatums = []data.Datum{NewContinuousGlucoseWithValue(datumTime, InTargetBloodGlucose)}

			err = userBuckets.Update(bucketFactory, cgmDatums)
			Expect(err).ToNot(HaveOccurred())
			Expect(userBuckets).To(HaveKey(bucketTime))
			Expect(userBuckets[bucketTime].Data.Target.Records).To(Equal(1))

			cgmDatums = []data.Datum{NewContinuousGlucoseWithValue(datumTime.Add(5*time.Minute), InTargetBloodGlucose)}

			err = userBuckets.Update(bucketFactory, cgmDatums)
			Expect(err).ToNot(HaveOccurred())
			Expect(userBuckets[bucketTime].Data.Target.Records).To(Equal(2))
		})

		It("Adding to two different buckets at once", func() {
			datumTime := bucketTime.Add(5 * time.Minute)
			userBuckets = BucketsByTime[*GlucoseBucket, GlucoseBucket]{}
			cgmDatums = []data.Datum{
				NewContinuousGlucoseWithValue(datumTime, InTargetBloodGlucose),
				NewContinuousGlucoseWithValue(datumTime.Add(time.Hour), LowBloodGlucose-0.1),
			}

			err = userBuckets.Update(bucketFactory, cgmDatums)
			Expect(err).ToNot(HaveOccurred())
			Expect(userBuckets).To(HaveKey(bucketTime))
			Expect(userBuckets[bucketTime].Data.Target.Records).To(Equal(1))
			Expect(userBuckets[bucketTime.Add(time.Hour)].Data.Low.Records).To(Equal(1))
		})

		It("Adding two records to the same bucket at once", func() {
			datumTime := bucketTime.Add(5 * time.Minute)
			userBuckets = BucketsByTime[*GlucoseBucket, GlucoseBucket]{}
			cgmDatums = []data.Datum{
				NewContinuousGlucoseWithValue(datumTime, InTargetBloodGlucose),
				NewContinuousGlucoseWithValue(datumTime.Add(5*time.Minute), LowBloodGlucose-0.1),
			}

			err = userBuckets.Update(bucketFactory, cgmDatums)
			Expect(err).ToNot(HaveOccurred())
			Expect(userBuckets).To(HaveKey(bucketTime))
			Expect(userBuckets[bucketTime].Data.Target.Records).To(Equal(1))
			Expect(userBuckets[bucketTime].Data.Low.Records).To(Equal(1))
		})
	})
})
