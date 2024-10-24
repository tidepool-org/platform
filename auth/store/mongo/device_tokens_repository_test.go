package mongo

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/auth/store"
	"github.com/tidepool-org/platform/devicetokens"
	"github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
)

const testUserID = "857ec1d7-8777-4877-a308-96a23c066524"

var _ = Describe("deviceTokenRepo", Label("mongodb", "slow", "integration"), func() {
	It("retrieves all for the given user id", func() {
		test := newDeviceTokensRepoTest()

		docs, err := test.Repo.GetAllByUserID(test.Ctx, testUserID)
		Expect(err).To(Succeed())

		if Expect(docs).To(HaveLen(2)) {
			for _, doc := range docs {
				Expect(doc.UserID).To(Equal(testUserID))
			}
		}
	})

	It("ensures indexes", func() {
		test := newDeviceTokensRepoTest()
		Expect(test.Repo.EnsureIndexes()).To(Succeed())
	})
})

type deviceTokensRepoTest struct {
	Ctx    context.Context
	Repo   store.DeviceTokenRepository
	Config *mongo.Config
	Store  *Store
}

func newDeviceTokensRepoTest() *deviceTokensRepoTest {
	test := &deviceTokensRepoTest{
		Ctx:    context.Background(),
		Config: storeStructuredMongoTest.NewConfig(),
	}
	store, err := NewStore(test.Config)
	Expect(err).To(Succeed())
	test.Store = store
	test.Repo = store.NewDeviceTokenRepository()

	testDocs := []*devicetokens.Document{
		{
			UserID:      testUserID,
			TokenKey:    "a",
			DeviceToken: devicetokens.DeviceToken{},
		},
		{
			UserID:      testUserID,
			TokenKey:    "b",
			DeviceToken: devicetokens.DeviceToken{},
		},
		{
			UserID:      "not" + testUserID,
			TokenKey:    "c",
			DeviceToken: devicetokens.DeviceToken{},
		},
	}
	for _, testDoc := range testDocs {
		Expect(test.Repo.Upsert(test.Ctx, testDoc)).To(Succeed())
	}

	return test
}
