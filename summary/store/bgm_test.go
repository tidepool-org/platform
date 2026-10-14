package store_test

import (
	"context"

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

var _ = Describe("BGM", Label("mongodb", "slow", "integration"), func() {
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

			bgmStore := dataStoreSummary.NewSummaries[*types.BGMPeriods, *types.GlucoseBucket](summaryRepository)
			Expect(bgmStore).ToNot(BeNil())
		})
	})

	Context("Store", func() {
		var userID string
		var userIDOther string
		var userBGMSummary *types.Summary[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket]
		var bgmStore *dataStoreSummary.Summaries[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket]

		BeforeEach(func() {
			store = GetSuiteStore()
			summaryRepository = store.NewSummaryRepository().GetStore()
			Expect(summaryRepository).ToNot(BeNil())

			bgmStore = dataStoreSummary.NewSummaries[*types.BGMPeriods, *types.GlucoseBucket](summaryRepository)

			userID = userTest.RandomUserID()
			userIDOther = userTest.RandomUserID()
		})

		AfterEach(func() {
			if summaryRepository != nil {
				_, err = summaryRepository.DeleteMany(ctx, bson.D{})
				Expect(err).To(Succeed())
			}
		})

		Context("ReplaceSummary", func() {
			It("Insert Summary with missing Type", func() {
				userBGMSummary = test.RandomBGMSummary(userID)
				userBGMSummary.Type = ""

				err = bgmStore.ReplaceSummary(ctx, userBGMSummary)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("invalid summary type '', expected 'bgm'"))
			})

			It("Insert Summary with invalid Type", func() {
				userBGMSummary = test.RandomBGMSummary(userID)
				userBGMSummary.Type = "asdf"

				err = bgmStore.ReplaceSummary(ctx, userBGMSummary)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("invalid summary type 'asdf', expected 'bgm'"))
			})

			It("Insert Summary", func() {
				userBGMSummary = test.RandomBGMSummary(userID)
				Expect(userBGMSummary.Type).To(Equal("bgm"))

				err = bgmStore.ReplaceSummary(ctx, userBGMSummary)
				Expect(err).ToNot(HaveOccurred())

				userBGMSummaryWritten, getErr := bgmStore.GetSummary(ctx, userID)
				Expect(getErr).ToNot(HaveOccurred())

				// the id generated on insert is reported on the summary written
				Expect(userBGMSummary.ID.IsZero()).To(BeFalse())
				Expect(userBGMSummaryWritten).To(Equal(userBGMSummary))
			})

			It("Update Summary", func() {
				var userBGMSummaryTwo *types.Summary[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket]
				var userBGMSummaryWritten *types.Summary[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket]
				var userBGMSummaryWrittenTwo *types.Summary[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket]

				// generate and insert first summary
				userBGMSummary = test.RandomBGMSummary(userID)
				Expect(userBGMSummary.Type).To(Equal("bgm"))

				err = bgmStore.ReplaceSummary(ctx, userBGMSummary)
				Expect(err).ToNot(HaveOccurred())

				// confirm first summary was written, get ID
				userBGMSummaryWritten, err = bgmStore.GetSummary(ctx, userID)
				Expect(err).ToNot(HaveOccurred())

				// the id generated on insert is reported on the summary written
				Expect(userBGMSummary.ID.IsZero()).To(BeFalse())
				Expect(userBGMSummaryWritten).To(Equal(userBGMSummary))

				// generate a new summary with same type and user, and upsert
				userBGMSummaryTwo = test.RandomBGMSummary(userID)
				err = bgmStore.ReplaceSummary(ctx, userBGMSummaryTwo)
				Expect(err).ToNot(HaveOccurred())

				userBGMSummaryWrittenTwo, err = bgmStore.GetSummary(ctx, userID)
				Expect(err).ToNot(HaveOccurred())

				// confirm the ID was unchanged and reported on the replacement summary
				Expect(userBGMSummaryWrittenTwo.ID).To(Equal(userBGMSummaryWritten.ID))
				Expect(userBGMSummaryTwo.ID).To(Equal(userBGMSummaryWritten.ID))

				// confirm the written summary matches the new summary
				Expect(userBGMSummaryWrittenTwo).To(Equal(userBGMSummaryTwo))
			})
		})

		Context("DeleteSummary", func() {
			It("Delete Summary with empty context", func() {
				err = bgmStore.DeleteSummary(context.Context(nil), userID)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("context is missing"))
			})

			It("Delete Summary with empty user id", func() {
				err = bgmStore.DeleteSummary(ctx, "")
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("user id is missing"))
			})

			It("Delete Summary", func() {
				var userBGMSummaryWritten *types.Summary[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket]

				userBGMSummary = test.RandomBGMSummary(userID)
				Expect(userBGMSummary.Type).To(Equal("bgm"))

				err = bgmStore.ReplaceSummary(ctx, userBGMSummary)
				Expect(err).ToNot(HaveOccurred())

				// confirm writes
				userBGMSummaryWritten, err = bgmStore.GetSummary(ctx, userID)
				Expect(err).ToNot(HaveOccurred())
				Expect(userBGMSummaryWritten).ToNot(BeNil())

				// delete
				err = bgmStore.DeleteSummary(ctx, userID)
				Expect(err).ToNot(HaveOccurred())

				// confirm delete
				userBGMSummaryWritten, err = bgmStore.GetSummary(ctx, userID)
				Expect(err).ToNot(HaveOccurred())
				Expect(userBGMSummaryWritten).To(BeNil())
			})
		})

		Context("CreateSummaries", func() {
			It("Create summaries with missing context", func() {
				var summaries = []*types.Summary[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket]{
					test.RandomBGMSummary(userID),
					test.RandomBGMSummary(userIDOther),
				}

				_, err = bgmStore.CreateSummaries(context.Context(nil), summaries)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("context is missing"))
			})

			It("Create summaries with missing summaries", func() {
				_, err = bgmStore.CreateSummaries(ctx, nil)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("summaries for create missing"))
			})

			It("Create summaries with an invalid type", func() {
				var summaries = []*types.Summary[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket]{
					test.RandomBGMSummary(userID),
					test.RandomBGMSummary(userIDOther),
				}

				summaries[0].Type = "cgm"

				_, err = bgmStore.CreateSummaries(ctx, summaries)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("invalid summary type 'cgm', expected 'bgm' at index 0"))
			})

			It("Create summaries with an invalid type", func() {
				var summaries = []*types.Summary[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket]{
					test.RandomBGMSummary(userID),
					test.RandomBGMSummary(userIDOther),
				}

				summaries[0].Type = "cgm"

				_, err = bgmStore.CreateSummaries(ctx, summaries)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("invalid summary type 'cgm', expected 'bgm' at index 0"))
			})

			It("Create summaries with an empty user id", func() {
				var summaries = []*types.Summary[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket]{
					test.RandomBGMSummary(userID),
					test.RandomBGMSummary(userIDOther),
				}

				summaries[0].UserID = ""

				_, err = bgmStore.CreateSummaries(ctx, summaries)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("user id is missing at index 0"))
			})

			It("Create summaries", func() {
				var count int
				var summaries = []*types.Summary[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket]{
					test.RandomBGMSummary(userID),
					test.RandomBGMSummary(userIDOther),
				}

				count, err = bgmStore.CreateSummaries(ctx, summaries)
				Expect(err).ToNot(HaveOccurred())
				Expect(count).To(Equal(2))

				for i := 0; i < 2; i++ {
					userBGMSummary, err = bgmStore.GetSummary(ctx, summaries[0].UserID)
					Expect(err).ToNot(HaveOccurred())
					Expect(userBGMSummary).ToNot(BeNil())
					summaries[i].ID = userBGMSummary.ID
					Expect(userBGMSummary).To(Equal(summaries[0]))
				}
			})
		})

		Context("GetSummary", func() {
			It("With missing context", func() {
				userBGMSummary, err = bgmStore.GetSummary(context.Context(nil), userID)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("context is missing"))
				Expect(userBGMSummary).To(BeNil())
			})

			It("With missing user id", func() {
				userBGMSummary, err = bgmStore.GetSummary(ctx, "")
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("user id is missing"))
				Expect(userBGMSummary).To(BeNil())
			})

			It("With no summary", func() {
				userBGMSummary, err = bgmStore.GetSummary(ctx, userID)
				Expect(err).ToNot(HaveOccurred())
				Expect(userBGMSummary).To(BeNil())
			})

			It("With multiple summaries", func() {
				var summaries = []*types.Summary[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket]{
					test.RandomBGMSummary(userID),
					test.RandomBGMSummary(userIDOther),
				}

				_, err = bgmStore.CreateSummaries(ctx, summaries)
				Expect(err).ToNot(HaveOccurred())

				userBGMSummary, err = bgmStore.GetSummary(ctx, userID)
				Expect(err).ToNot(HaveOccurred())
				Expect(userBGMSummary).ToNot(BeNil())

				summaries[0].ID = userBGMSummary.ID
				Expect(userBGMSummary).To(Equal(summaries[0]))
			})

			It("Get with multiple summaries of different type a", func() {
				cgmStore := dataStoreSummary.NewSummaries[*types.CGMPeriods, *types.GlucoseBucket](summaryRepository)
				continuousStore := dataStoreSummary.NewSummaries[*types.ContinuousPeriods, *types.ContinuousBucket](summaryRepository)

				var cgmSummaries = []*types.Summary[*types.CGMPeriods, *types.GlucoseBucket, types.CGMPeriods, types.GlucoseBucket]{
					test.RandomCGMSummary(userID),
					test.RandomCGMSummary(userIDOther),
				}

				var bgmSummaries = []*types.Summary[*types.BGMPeriods, *types.GlucoseBucket, types.BGMPeriods, types.GlucoseBucket]{
					test.RandomBGMSummary(userID),
					test.RandomBGMSummary(userIDOther),
				}

				var continuousSummaries = []*types.Summary[*types.ContinuousPeriods, *types.ContinuousBucket, types.ContinuousPeriods, types.ContinuousBucket]{
					test.RandomContinuousSummary(userID),
					test.RandomContinuousSummary(userIDOther),
				}

				_, err = cgmStore.CreateSummaries(ctx, cgmSummaries)
				Expect(err).ToNot(HaveOccurred())

				_, err = bgmStore.CreateSummaries(ctx, bgmSummaries)
				Expect(err).ToNot(HaveOccurred())

				_, err = continuousStore.CreateSummaries(ctx, continuousSummaries)
				Expect(err).ToNot(HaveOccurred())

				userBGMSummary, err = bgmStore.GetSummary(ctx, userID)
				Expect(err).ToNot(HaveOccurred())
				Expect(userBGMSummary).ToNot(BeNil())

				bgmSummaries[0].ID = userBGMSummary.ID
				Expect(userBGMSummary).To(Equal(bgmSummaries[0]))
			})
		})

	})
})
