package store_test

import (
	"context"

	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson"

	dataStoreMongo "github.com/tidepool-org/platform/data/store/mongo"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	dataStoreSummary "github.com/tidepool-org/platform/summary/store"
	"github.com/tidepool-org/platform/summary/test"
	"github.com/tidepool-org/platform/summary/types"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("Continuous", Label("mongodb", "slow", "integration"), func() {
	var logger *logTest.Logger
	var err error
	var ctx context.Context
	var store *dataStoreMongo.Store
	var summaryRepository *storeStructuredMongo.Repository

	BeforeEach(func() {
		logger = logTest.NewLogger()
		ctx = log.NewContextWithLogger(context.Background(), logger)
	})

	Context("Create repo and store", func() {
		var createStore *dataStoreMongo.Store

		It("Repo", func() {
			createStore = GetSuiteStore()

			summaryRepository = createStore.NewSummaryRepository().GetStore()
			Expect(summaryRepository).ToNot(BeNil())

			bgmStore := dataStoreSummary.NewSummaries[*types.ContinuousPeriods, *types.ContinuousBucket](summaryRepository)
			Expect(bgmStore).ToNot(BeNil())
		})
	})

	Context("Store", func() {
		var userId string
		var userIdOther string
		var userContinuousSummary *types.Summary[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket]
		var continuousStore *dataStoreSummary.Summaries[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket]

		BeforeEach(func() {
			store = GetSuiteStore()
			summaryRepository = store.NewSummaryRepository().GetStore()
			Expect(summaryRepository).ToNot(BeNil())

			continuousStore = dataStoreSummary.NewSummaries[*types.ContinuousPeriods, *types.ContinuousBucket](summaryRepository)

			userId = userTest.RandomUserID()
			userIdOther = userTest.RandomUserID()
		})

		AfterEach(func() {
			if summaryRepository != nil {
				_, err = summaryRepository.DeleteMany(ctx, bson.D{})
				Expect(err).To(Succeed())
			}
		})

		Context("ReplaceSummary", func() {
			It("Insert Summary with missing Type", func() {
				userContinuousSummary = test.RandomContinuousSummary(userId)
				userContinuousSummary.Type = ""

				err = continuousStore.ReplaceSummary(ctx, userContinuousSummary)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("invalid summary type '', expected 'con'"))
			})

			It("Insert Summary with invalid Type", func() {
				userContinuousSummary = test.RandomContinuousSummary(userId)
				userContinuousSummary.Type = "asdf"

				err = continuousStore.ReplaceSummary(ctx, userContinuousSummary)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("invalid summary type 'asdf', expected 'con'"))
			})

			It("Insert Summary", func() {
				userContinuousSummary = test.RandomContinuousSummary(userId)
				Expect(userContinuousSummary.Type).To(Equal("con"))

				err = continuousStore.ReplaceSummary(ctx, userContinuousSummary)
				Expect(err).ToNot(HaveOccurred())

				userContinuousSummaryWritten, err := continuousStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())

				// copy id, as that was mongo generated
				userContinuousSummary.ID = userContinuousSummaryWritten.ID
				Expect(userContinuousSummaryWritten).To(Equal(userContinuousSummary))
			})

			It("Update Summary", func() {
				var userContinuousSummaryTwo *types.Summary[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket]
				var userContinuousSummaryWritten *types.Summary[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket]
				var userContinuousSummaryWrittenTwo *types.Summary[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket]

				// generate and insert first summary
				userContinuousSummary = test.RandomContinuousSummary(userId)
				Expect(userContinuousSummary.Type).To(Equal("con"))

				err = continuousStore.ReplaceSummary(ctx, userContinuousSummary)
				Expect(err).ToNot(HaveOccurred())

				// confirm first summary was written, get ID
				userContinuousSummaryWritten, err = continuousStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())

				// copy id, as that was mongo generated
				userContinuousSummary.ID = userContinuousSummaryWritten.ID
				Expect(userContinuousSummaryWritten).To(Equal(userContinuousSummary))

				// generate a new summary with same type and user, and upsert
				userContinuousSummaryTwo = test.RandomContinuousSummary(userId)
				err = continuousStore.ReplaceSummary(ctx, userContinuousSummaryTwo)
				Expect(err).ToNot(HaveOccurred())

				userContinuousSummaryWrittenTwo, err = continuousStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())

				// confirm the ID was unchanged
				Expect(userContinuousSummaryWrittenTwo.ID).To(Equal(userContinuousSummaryWritten.ID))

				// confirm the written summary matches the new summary
				userContinuousSummaryWrittenTwo.ID = userContinuousSummaryTwo.ID
				opts := cmpopts.IgnoreUnexported(types.ContinuousPeriod{})
				Expect(userContinuousSummaryWrittenTwo).To(BeComparableTo(userContinuousSummaryTwo, opts))
			})
		})

		Context("DeleteSummary", func() {
			It("Delete Summary with empty context", func() {
				err = continuousStore.DeleteSummary(nil, userId)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("context is missing"))
			})

			It("Delete Summary with empty userId", func() {
				err = continuousStore.DeleteSummary(ctx, "")
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("userId is missing"))
			})

			It("Delete Summary", func() {
				var userContinuousSummaryWritten *types.Summary[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket]

				userContinuousSummary = test.RandomContinuousSummary(userId)
				Expect(userContinuousSummary.Type).To(Equal("con"))

				err = continuousStore.ReplaceSummary(ctx, userContinuousSummary)
				Expect(err).ToNot(HaveOccurred())

				// confirm writes
				userContinuousSummaryWritten, err = continuousStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())
				Expect(userContinuousSummaryWritten).ToNot(BeNil())

				// delete
				err = continuousStore.DeleteSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())

				// confirm delete
				userContinuousSummaryWritten, err = continuousStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())
				Expect(userContinuousSummaryWritten).To(BeNil())
			})
		})

		Context("CreateSummaries", func() {
			It("Create summaries with missing context", func() {
				var summaries = []*types.Summary[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket]{
					test.RandomContinuousSummary(userId),
					test.RandomContinuousSummary(userIdOther),
				}

				_, err = continuousStore.CreateSummaries(nil, summaries)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("context is missing"))
			})

			It("Create summaries with missing summaries", func() {
				_, err = continuousStore.CreateSummaries(ctx, nil)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("summaries for create missing"))
			})

			It("Create summaries with an invalid type", func() {
				var summaries = []*types.Summary[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket]{
					test.RandomContinuousSummary(userId),
					test.RandomContinuousSummary(userIdOther),
				}

				summaries[0].Type = "bgm"

				_, err = continuousStore.CreateSummaries(ctx, summaries)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("invalid summary type 'bgm', expected 'con' at index 0"))
			})

			It("Create summaries with an empty userId", func() {
				var summaries = []*types.Summary[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket]{
					test.RandomContinuousSummary(userId),
					test.RandomContinuousSummary(userIdOther),
				}

				summaries[0].UserID = ""

				_, err = continuousStore.CreateSummaries(ctx, summaries)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("userId is missing at index 0"))
			})

			It("Create summaries", func() {
				var count int
				var summaries = []*types.Summary[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket]{
					test.RandomContinuousSummary(userId),
					test.RandomContinuousSummary(userIdOther),
				}

				count, err = continuousStore.CreateSummaries(ctx, summaries)
				Expect(err).ToNot(HaveOccurred())
				Expect(count).To(Equal(2))

				for i := 0; i < 2; i++ {
					userContinuousSummary, err = continuousStore.GetSummary(ctx, summaries[0].UserID)
					Expect(err).ToNot(HaveOccurred())
					Expect(userContinuousSummary).ToNot(BeNil())
					summaries[i].ID = userContinuousSummary.ID
					Expect(userContinuousSummary).To(Equal(summaries[0]))
				}
			})
		})

		Context("GetSummary", func() {
			It("With missing context", func() {
				userContinuousSummary, err = continuousStore.GetSummary(nil, userId)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("context is missing"))
				Expect(userContinuousSummary).To(BeNil())
			})

			It("With missing userId", func() {
				userContinuousSummary, err = continuousStore.GetSummary(ctx, "")
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("userId is missing"))
				Expect(userContinuousSummary).To(BeNil())
			})

			It("With no summary", func() {
				userContinuousSummary, err = continuousStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())
				Expect(userContinuousSummary).To(BeNil())
			})

			It("With multiple summaries", func() {
				var summaries = []*types.Summary[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket]{
					test.RandomContinuousSummary(userId),
					test.RandomContinuousSummary(userIdOther),
				}

				_, err = continuousStore.CreateSummaries(ctx, summaries)
				Expect(err).ToNot(HaveOccurred())

				userContinuousSummary, err = continuousStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())
				Expect(userContinuousSummary).ToNot(BeNil())

				summaries[0].ID = userContinuousSummary.ID
				Expect(userContinuousSummary).To(Equal(summaries[0]))
			})

			It("Get with multiple summaries of different type", func() {
				cgmStore := dataStoreSummary.NewSummaries[*types.CGMPeriods, *types.GlucoseBucket](summaryRepository)
				bgmStore := dataStoreSummary.NewSummaries[*types.BGMPeriods, *types.GlucoseBucket](summaryRepository)

				var cgmSummaries = []*types.Summary[*types.CGMPeriods, *types.GlucoseBucket, types.CGMPeriods, types.GlucoseBucket]{
					test.RandomCGMSummary(userId),
					test.RandomCGMSummary(userIdOther),
				}

				var bgmSummaries = []*types.Summary[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket]{
					test.RandomBGMSummary(userId),
					test.RandomBGMSummary(userIdOther),
				}

				var continuousSummaries = []*types.Summary[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket]{
					test.RandomContinuousSummary(userId),
					test.RandomContinuousSummary(userIdOther),
				}

				_, err = cgmStore.CreateSummaries(ctx, cgmSummaries)
				Expect(err).ToNot(HaveOccurred())

				_, err = bgmStore.CreateSummaries(ctx, bgmSummaries)
				Expect(err).ToNot(HaveOccurred())

				_, err = continuousStore.CreateSummaries(ctx, continuousSummaries)
				Expect(err).ToNot(HaveOccurred())

				userContinuousSummary, err = continuousStore.GetSummary(ctx, userId)
				Expect(err).ToNot(HaveOccurred())
				Expect(userContinuousSummary).ToNot(BeNil())

				continuousSummaries[0].ID = userContinuousSummary.ID
				opts := cmpopts.IgnoreUnexported(types.ContinuousPeriod{})
				Expect(userContinuousSummary).To(BeComparableTo(continuousSummaries[0], opts))
			})
		})

	})
})
