package summary_test

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"

	dataStore "github.com/tidepool-org/platform/data/store"
	dataStoreMongo "github.com/tidepool-org/platform/data/store/mongo"
	"github.com/tidepool-org/platform/data/summary"
	dataStoreSummary "github.com/tidepool-org/platform/data/summary/store"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"

	userTest "github.com/tidepool-org/platform/user/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/summary/types"
)

var _ = Describe("Summary", func() {
	Context("MaybeUpdateSummary", func() {
		var err error
		var empty struct{}
		var logger log.Logger
		var ctx context.Context
		var registry *summary.SummarizerRegistry
		var config *storeStructuredMongo.Config
		var store *dataStoreMongo.Store
		var summaryRepository *storeStructuredMongo.Repository
		var dataStore dataStore.DataRepository
		var userId string
		var cgmStore *dataStoreSummary.Repo[types.CGMStats, *types.CGMStats]
		var bgmStore *dataStoreSummary.Repo[types.BGMStats, *types.BGMStats]

		BeforeEach(func() {
			logger = logTest.NewLogger()
			ctx = log.NewContextWithLogger(context.Background(), logger)
			config = storeStructuredMongoTest.NewConfig()

			store, err = dataStoreMongo.NewStore(config)
			Expect(err).ToNot(HaveOccurred())

			summaryRepository = store.NewSummaryRepository().GetStore()
			dataStore = store.NewDataRepository()
			registry = summary.New(summaryRepository, dataStore)
			userId = userTest.RandomID()

			cgmStore = dataStoreSummary.New[types.CGMStats, *types.CGMStats](summaryRepository)
			bgmStore = dataStoreSummary.New[types.BGMStats, *types.BGMStats](summaryRepository)
		})

		AfterEach(func() {
			_, err = summaryRepository.DeleteMany(ctx, bson.D{})
			Expect(err).ToNot(HaveOccurred())
		})

		It("with all summary types outdated", func() {
			updatesSummary := map[string]struct{}{
				"cgm": empty,
				"bgm": empty,
			}

			outdatedSinceMap := summary.MaybeUpdateSummary(ctx, registry, updatesSummary, userId, types.OutdatedReasonDataAdded)
			Expect(outdatedSinceMap).To(HaveLen(2))
			Expect(outdatedSinceMap).To(HaveKey(types.SummaryTypeCGM))
			Expect(outdatedSinceMap).To(HaveKey(types.SummaryTypeBGM))

			userCgmSummary, err := cgmStore.GetSummary(ctx, userId)
			Expect(err).ToNot(HaveOccurred())
			Expect(*userCgmSummary.Dates.OutdatedSince).To(Equal(*outdatedSinceMap[types.SummaryTypeCGM]))

			userBgmSummary, err := bgmStore.GetSummary(ctx, userId)
			Expect(err).ToNot(HaveOccurred())
			Expect(*userBgmSummary.Dates.OutdatedSince).To(Equal(*outdatedSinceMap[types.SummaryTypeBGM]))
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
		})

		It("with unknown summary type outdated", func() {
			updatesSummary := map[string]struct{}{
				"food": empty,
			}

			outdatedSinceMap := summary.MaybeUpdateSummary(ctx, registry, updatesSummary, userId, types.OutdatedReasonDataAdded)
			Expect(outdatedSinceMap).To(BeEmpty())
		})
	})
})
