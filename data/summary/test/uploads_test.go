package test_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataStore "github.com/tidepool-org/platform/data/store"
	dataStoreMongo "github.com/tidepool-org/platform/data/store/mongo"
	"github.com/tidepool-org/platform/data/summary"
	dataStoreSummary "github.com/tidepool-org/platform/data/summary/store"
	. "github.com/tidepool-org/platform/data/summary/test/generators"
	"github.com/tidepool-org/platform/data/summary/types"
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	"github.com/tidepool-org/platform/data/types/blood/glucose/selfmonitored"
	"github.com/tidepool-org/platform/data/types/food"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
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
	var cgmStore *dataStoreSummary.Summaries[*types.CGMPeriods, *types.GlucoseBucket, types.CGMPeriods, types.GlucoseBucket]
	var bgmStore *dataStoreSummary.Summaries[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket]
	var continuousStore *dataStoreSummary.Summaries[*types.ContinuousStats, *types.ContinuousBucket, types.ContinuousStats, types.ContinuousBucket]

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
		registry = summary.New(summaryRepo, bucketsRepo, dataRepo)
		userId = userTest.RandomID()

		cgmStore = dataStoreSummary.NewSummaries[*types.CGMPeriods, *types.GlucoseBucket](summaryRepo)
		bgmStore = dataStoreSummary.NewSummaries[*types.BGMPeriods, *types.GlucoseBucket](summaryRepo)
		continuousStore = dataStoreSummary.NewSummaries[*types.ContinuousStats, *types.ContinuousBucket](summaryRepo)
	})

	Context("MaybeUpdateSummary", func() {

		It("with all summary types outdated", func() {
			updatesSummary := map[string]struct{}{
				"cgm": empty,
				"bgm": empty,
				"con": empty,
			}

			outdatedSinceMap := summary.MaybeUpdateSummary(ctx, registry, updatesSummary, userId, types.OutdatedReasonDataAdded)
			Expect(outdatedSinceMap).To(HaveLen(3))
			Expect(outdatedSinceMap).To(HaveKey(types.SummaryTypeCGM))
			Expect(outdatedSinceMap).To(HaveKey(types.SummaryTypeBGM))
			Expect(outdatedSinceMap).To(HaveKey(types.SummaryTypeContinuous))

			userCgmSummary, err := cgmStore.GetSummary(ctx, userId)
			Expect(err).ToNot(HaveOccurred())
			Expect(*userCgmSummary.Dates.OutdatedSince).To(Equal(*outdatedSinceMap[types.SummaryTypeCGM]))

			userBgmSummary, err := bgmStore.GetSummary(ctx, userId)
			Expect(err).ToNot(HaveOccurred())
			Expect(*userBgmSummary.Dates.OutdatedSince).To(Equal(*outdatedSinceMap[types.SummaryTypeBGM]))

			userContinuousSummary, err := continuousStore.GetSummary(ctx, userId)
			Expect(err).ToNot(HaveOccurred())
			Expect(*userContinuousSummary.Dates.OutdatedSince).To(Equal(*outdatedSinceMap[types.SummaryTypeContinuous]))
		})

		It("with cgm summary type outdated", func() {
			updatesSummary := map[string]struct{}{
				"cgm": empty,
			}

			outdatedSinceMap := summary.MaybeUpdateSummary(ctx, registry, updatesSummary, userId, types.OutdatedReasonDataAdded)
			Expect(outdatedSinceMap).To(HaveLen(1))
			Expect(outdatedSinceMap).To(HaveKey(types.SummaryTypeCGM))

			userCgmSummary, err := cgmStore.GetSummary(ctx, userId)
			Expect(err).ToNot(HaveOccurred())
			Expect(*userCgmSummary.Dates.OutdatedSince).To(Equal(*outdatedSinceMap[types.SummaryTypeCGM]))

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

			outdatedSinceMap := summary.MaybeUpdateSummary(ctx, registry, updatesSummary, userId, types.OutdatedReasonDataAdded)
			Expect(outdatedSinceMap).To(HaveLen(1))
			Expect(outdatedSinceMap).To(HaveKey(types.SummaryTypeBGM))

			userCgmSummary, err := cgmStore.GetSummary(ctx, userId)
			Expect(err).ToNot(HaveOccurred())
			Expect(userCgmSummary).To(BeNil())

			userBgmSummary, err := bgmStore.GetSummary(ctx, userId)
			Expect(err).ToNot(HaveOccurred())
			Expect(*userBgmSummary.Dates.OutdatedSince).To(Equal(*outdatedSinceMap[types.SummaryTypeBGM]))

			userContinuousSummary, err := continuousStore.GetSummary(ctx, userId)
			Expect(err).ToNot(HaveOccurred())
			Expect(userContinuousSummary).To(BeNil())
		})

		It("with continuous summary type outdated", func() {
			updatesSummary := map[string]struct{}{
				"con": empty,
			}

			outdatedSinceMap := summary.MaybeUpdateSummary(ctx, registry, updatesSummary, userId, types.OutdatedReasonDataAdded)
			Expect(outdatedSinceMap).To(HaveLen(1))
			Expect(outdatedSinceMap).To(HaveKey(types.SummaryTypeContinuous))

			userCgmSummary, err := cgmStore.GetSummary(ctx, userId)
			Expect(err).ToNot(HaveOccurred())
			Expect(userCgmSummary).To(BeNil())

			userBgmSummary, err := bgmStore.GetSummary(ctx, userId)
			Expect(err).ToNot(HaveOccurred())
			Expect(userBgmSummary).To(BeNil())

			userContinuousSummary, err := continuousStore.GetSummary(ctx, userId)
			Expect(err).ToNot(HaveOccurred())
			Expect(*userContinuousSummary.Dates.OutdatedSince).To(Equal(*outdatedSinceMap[types.SummaryTypeContinuous]))
		})

		It("with unknown summary type outdated", func() {
			updatesSummary := map[string]struct{}{
				"food": empty,
			}

			outdatedSinceMap := summary.MaybeUpdateSummary(ctx, registry, updatesSummary, userId, types.OutdatedReasonDataAdded)
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
			Expect(updatesSummary).To(HaveKey(types.SummaryTypeCGM))
			Expect(updatesSummary).To(HaveKey(types.SummaryTypeContinuous))
		})

		It("with BGM summary affecting record", func() {
			updatesSummary := make(map[string]struct{})
			datum := NewDatum(selfmonitored.Type)

			summary.CheckDatumUpdatesSummary(updatesSummary, datum)
			Expect(updatesSummary).To(HaveLen(2))
			Expect(updatesSummary).To(HaveKey(types.SummaryTypeBGM))
			Expect(updatesSummary).To(HaveKey(types.SummaryTypeContinuous))
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
