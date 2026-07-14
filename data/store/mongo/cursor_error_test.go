package mongo_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"

	"github.com/tidepool-org/platform/data"
	dataStoreMongo "github.com/tidepool-org/platform/data/store/mongo"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

func TestUnarchiveDeviceDataWithFailingAggregate(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("returns the aggregate error instead of panicking", func(mt *mtest.T) {
		g := NewWithT(mt.T)
		ctx := log.NewContextWithLogger(context.Background(), logTest.NewLogger())
		repo := &dataStoreMongo.DatumRepository{
			Repository: storeStructuredMongo.NewRepository(mt.Coll),
		}

		dataSet := &data.DataSet{
			UserID:   pointer.FromString("user-1"),
			UploadID: pointer.FromString("upload-1"),
			DeviceID: pointer.FromString("device-1"),
		}

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{
			Code:    43,
			Message: "aggregate failed",
		}))

		err := repo.UnarchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet)
		g.Expect(err).To(MatchError(ContainSubstring("aggregate failed")))
	})
}
