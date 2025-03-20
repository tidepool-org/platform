package types_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataStore "github.com/tidepool-org/platform/data/store"
	dataStoreMongo "github.com/tidepool-org/platform/data/store/mongo"
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	"github.com/tidepool-org/platform/data/types/blood/glucose/selfmonitored"
	"github.com/tidepool-org/platform/data/types/food"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
	"github.com/tidepool-org/platform/summary"
	dataStoreSummary "github.com/tidepool-org/platform/summary/store"
	. "github.com/tidepool-org/platform/summary/test"
	. "github.com/tidepool-org/platform/summary/types"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("Upload Helpers", func() {
	var err error
	var empty struct{}
	var logger log.Logger
	var ctx context.Context
	var registry *summary.SummarizerRegistry
	var config *storeStructuredMongo.Config
	var store *dataStoreMongo.Store
	var summaryRepo *storeStructuredMongo.Repository
	var bucketsRepo *storeStructuredMongo.Repository
	var dataRepo dataStore.DataRepository
	var userId string
	var cgmStore *dataStoreSummary.Summaries[*CGMPeriods, *GlucoseBucket, CGMPeriods, GlucoseBucket]
	var bgmStore *dataStoreSummary.Summaries[*BGMPeriods, *GlucoseBucket, BGMPeriods, GlucoseBucket]
	var continuousStore *dataStoreSummary.Summaries[*ContinuousPeriods, *ContinuousBucket, ContinuousPeriods, ContinuousBucket]

	BeforeEach(func() {
		logger = logTest.NewLogger()
		ctx = log.NewContextWithLogger(context.Background(), logger)
		config = storeStructuredMongoTest.NewConfig()

		store, err = dataStoreMongo.NewStore(config)
		Expect(err).ToNot(HaveOccurred())
		Expect(store.EnsureIndexes()).To(Succeed())

		summaryRepo = store.NewSummaryRepository().GetStore()
		bucketsRepo = store.NewBucketsRepository().GetStore()
		dataRepo = store.NewDataRepository()
		registry = summary.New(summaryRepo, bucketsRepo, dataRepo, store.GetClient())
		userId = userTest.RandomID()

		cgmStore = dataStoreSummary.NewSummaries[*CGMPeriods, *GlucoseBucket](summaryRepo)
		bgmStore = dataStoreSummary.NewSummaries[*BGMPeriods, *GlucoseBucket](summaryRepo)
		continuousStore = dataStoreSummary.NewSummaries[*ContinuousPeriods, *ContinuousBucket](summaryRepo)
	})

	Context("MaybeUpdateSummary", func() {

		It("with all summary types outdated", func() {
			updatesSummary := map[string]struct{}{
				"cgm": empty,
				"bgm": empty,
				"con": empty,
			}

			outdatedSinceMap := summary.MaybeUpdateSummary(ctx, registry, updatesSummary, userId, OutdatedReasonDataAdded)
			Expect(outdatedSinceMap).To(HaveLen(3))
			Expect(outdatedSinceMap).To(HaveKey(SummaryTypeCGM))
			Expect(outdatedSinceMap).To(HaveKey(SummaryTypeBGM))
			Expect(outdatedSinceMap).To(HaveKey(SummaryTypeContinuous))

			userCgmSummary, err := cgmStore.GetSummary(ctx, userId)
			Expect(err).ToNot(HaveOccurred())
			Expect(*userCgmSummary.Dates.OutdatedSince).To(Equal(*outdatedSinceMap[SummaryTypeCGM]))

			userBgmSummary, err := bgmStore.GetSummary(ctx, userId)
			Expect(err).ToNot(HaveOccurred())
			Expect(*userBgmSummary.Dates.OutdatedSince).To(Equal(*outdatedSinceMap[SummaryTypeBGM]))

			userContinuousSummary, err := continuousStore.GetSummary(ctx, userId)
			Expect(err).ToNot(HaveOccurred())
			Expect(*userContinuousSummary.Dates.OutdatedSince).To(Equal(*outdatedSinceMap[SummaryTypeContinuous]))
		})

		It("with cgm summary type outdated", func() {
			updatesSummary := map[string]struct{}{
				"cgm": empty,
			}

			outdatedSinceMap := summary.MaybeUpdateSummary(ctx, registry, updatesSummary, userId, OutdatedReasonDataAdded)
			Expect(outdatedSinceMap).To(HaveLen(1))
			Expect(outdatedSinceMap).To(HaveKey(SummaryTypeCGM))

			userCgmSummary, err := cgmStore.GetSummary(ctx, userId)
			Expect(err).ToNot(HaveOccurred())
			Expect(*userCgmSummary.Dates.OutdatedSince).To(Equal(*outdatedSinceMap[SummaryTypeCGM]))

			userBgmSummary, err := bgmStore.GetSummary(ctx, userId)
			Expect(err).ToNot(HaveOccurred())
			Expect(userBgmSummary).To(BeNil())

			userContinuousSummary, err := continuousStore.GetSummary(ctx, userId)
			Expect(err).ToNot(HaveOccurred())
			Expect(userContinuousSummary).To(BeNil())
		})

		It("with bgm summary type outdated", func() {
			updatesSummary := map[string]struct{}{
				"bgm": empty,
			}

			outdatedSinceMap := summary.MaybeUpdateSummary(ctx, registry, updatesSummary, userId, OutdatedReasonDataAdded)
			Expect(outdatedSinceMap).To(HaveLen(1))
			Expect(outdatedSinceMap).To(HaveKey(SummaryTypeBGM))

			userCgmSummary, err := cgmStore.GetSummary(ctx, userId)
			Expect(err).ToNot(HaveOccurred())
			Expect(userCgmSummary).To(BeNil())

			userBgmSummary, err := bgmStore.GetSummary(ctx, userId)
			Expect(err).ToNot(HaveOccurred())
			Expect(*userBgmSummary.Dates.OutdatedSince).To(Equal(*outdatedSinceMap[SummaryTypeBGM]))

			userContinuousSummary, err := continuousStore.GetSummary(ctx, userId)
			Expect(err).ToNot(HaveOccurred())
			Expect(userContinuousSummary).To(BeNil())
		})

		It("with continuous summary type outdated", func() {
			updatesSummary := map[string]struct{}{
				"con": empty,
			}

			outdatedSinceMap := summary.MaybeUpdateSummary(ctx, registry, updatesSummary, userId, OutdatedReasonDataAdded)
			Expect(outdatedSinceMap).To(HaveLen(1))
			Expect(outdatedSinceMap).To(HaveKey(SummaryTypeContinuous))

			userCgmSummary, err := cgmStore.GetSummary(ctx, userId)
			Expect(err).ToNot(HaveOccurred())
			Expect(userCgmSummary).To(BeNil())

			userBgmSummary, err := bgmStore.GetSummary(ctx, userId)
			Expect(err).ToNot(HaveOccurred())
			Expect(userBgmSummary).To(BeNil())

			userContinuousSummary, err := continuousStore.GetSummary(ctx, userId)
			Expect(err).ToNot(HaveOccurred())
			Expect(*userContinuousSummary.Dates.OutdatedSince).To(Equal(*outdatedSinceMap[SummaryTypeContinuous]))
		})

		It("with unknown summary type outdated", func() {
			updatesSummary := map[string]struct{}{
				"food": empty,
			}

			outdatedSinceMap := summary.MaybeUpdateSummary(ctx, registry, updatesSummary, userId, OutdatedReasonDataAdded)
			Expect(outdatedSinceMap).To(BeEmpty())
		})
	})

	Context("CheckDatumUpdatesSummary", func() {
		It("with non-summary type", func() {
			var updatesSummary map[string]struct{}
			datum := NewDatum(food.Type)

			summary.CheckDatumUpdatesSummary(updatesSummary, datum)
			Expect(updatesSummary).To(BeEmpty())
		})

		It("with too old summary affecting record", func() {
			updatesSummary := make(map[string]struct{})
			datum := NewOldDatum(continuous.Type)

			summary.CheckDatumUpdatesSummary(updatesSummary, datum)
			Expect(updatesSummary).To(HaveLen(0))
		})

		It("with future summary affecting record", func() {
			updatesSummary := make(map[string]struct{})
			datum := NewNewDatum(continuous.Type)

			summary.CheckDatumUpdatesSummary(updatesSummary, datum)
			Expect(updatesSummary).To(HaveLen(0))
		})

		It("with CGM summary affecting record", func() {
			updatesSummary := make(map[string]struct{})
			datum := NewDatum(continuous.Type)

			summary.CheckDatumUpdatesSummary(updatesSummary, datum)
			Expect(updatesSummary).To(HaveLen(2))
			Expect(updatesSummary).To(HaveKey(SummaryTypeCGM))
			Expect(updatesSummary).To(HaveKey(SummaryTypeContinuous))
		})

		It("with BGM summary affecting record", func() {
			updatesSummary := make(map[string]struct{})
			datum := NewDatum(selfmonitored.Type)

			summary.CheckDatumUpdatesSummary(updatesSummary, datum)
			Expect(updatesSummary).To(HaveLen(2))
			Expect(updatesSummary).To(HaveKey(SummaryTypeBGM))
			Expect(updatesSummary).To(HaveKey(SummaryTypeContinuous))
		})

		It("with inactive BGM summary affecting record", func() {
			updatesSummary := make(map[string]struct{})
			datum := NewDatum(selfmonitored.Type)
			datum.Active = false

			summary.CheckDatumUpdatesSummary(updatesSummary, datum)
			Expect(updatesSummary).To(HaveLen(0))
		})

		It("with inactive CGM summary affecting record", func() {
			updatesSummary := make(map[string]struct{})
			datum := NewDatum(continuous.Type)
			datum.Active = false

			summary.CheckDatumUpdatesSummary(updatesSummary, datum)
			Expect(updatesSummary).To(HaveLen(0))
		})
	})
})
