package mongo_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"

	consentStoreMongo "github.com/tidepool-org/platform/consent/store/mongo"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

func TestListWithFailingCursor(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("returns the cursor error instead of an empty result", func(mt *mtest.T) {
		g := NewWithT(mt.T)
		ctx := log.NewContextWithLogger(context.Background(), logTest.NewLogger())
		repo := &consentStoreMongo.ConsentRepository{
			Repository: storeStructuredMongo.NewRepository(mt.Coll),
		}

		ns := mt.DB.Name() + "." + mt.Coll.Name()
		mt.AddMockResponses(
			// the aggregation cursor fails before delivering its single result document
			mtest.CreateCursorResponse(1, ns, mtest.FirstBatch),
			mtest.CreateCommandErrorResponse(mtest.CommandError{
				Code:    43,
				Message: "cursor killed mid-iteration",
			}),
		)

		_, err := repo.List(ctx, nil, nil)
		g.Expect(err).To(MatchError(ContainSubstring("cursor killed mid-iteration")))
	})
}

func TestListConsentRecordsWithFailingCursor(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("returns the cursor error instead of an empty result", func(mt *mtest.T) {
		g := NewWithT(mt.T)
		ctx := log.NewContextWithLogger(context.Background(), logTest.NewLogger())
		repo := &consentStoreMongo.ConsentRecordRepository{
			Repository: storeStructuredMongo.NewRepository(mt.Coll),
		}

		ns := mt.DB.Name() + "." + mt.Coll.Name()
		mt.AddMockResponses(
			// the aggregation cursor fails before delivering its single result document
			mtest.CreateCursorResponse(1, ns, mtest.FirstBatch),
			mtest.CreateCommandErrorResponse(mtest.CommandError{
				Code:    43,
				Message: "cursor killed mid-iteration",
			}),
		)

		_, err := repo.ListConsentRecords(ctx, "user-1", nil, nil)
		g.Expect(err).To(MatchError(ContainSubstring("cursor killed mid-iteration")))
	})
}
