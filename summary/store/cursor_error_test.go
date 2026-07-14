package store_test

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"

	"github.com/tidepool-org/platform/page"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	summaryStore "github.com/tidepool-org/platform/summary/store"
	"github.com/tidepool-org/platform/summary/types"
)

func TestGetOutdatedUserIDsWithFailingCursor(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("returns the cursor error instead of a truncated response", func(mt *mtest.T) {
		g := NewWithT(mt.T)
		summaries := summaryStore.NewSummaries[*types.CGMPeriods,
			*types.GlucoseBucket](storeStructuredMongo.NewRepository(mt.Coll))

		ns := mt.DB.Name() + "." + mt.Coll.Name()
		mt.AddMockResponses(
			// one outdated summary arrives before the cursor fails mid-iteration
			mtest.CreateCursorResponse(1, ns, mtest.FirstBatch, bson.D{
				{
					Key: "userId", Value: "user-1",
				},
				{
					Key: "dates",
					Value: bson.D{
						{
							Key:   "outdatedSince",
							Value: time.Now().UTC().Truncate(time.Millisecond),
						},
					},
				}}),
			mtest.CreateCommandErrorResponse(mtest.CommandError{
				Code: 43, Message: "cursor killed mid-iteration"}),
			mtest.CreateCursorResponse(0, ns, mtest.FirstBatch, bson.D{
				{Key: "n", Value: int32(3)},
			}),
		)

		_, err := summaries.GetOutdatedUserIDs(context.Background(), page.NewPagination())
		g.Expect(err).To(MatchError(ContainSubstring("cursor killed mid-iteration")))
	})
}
