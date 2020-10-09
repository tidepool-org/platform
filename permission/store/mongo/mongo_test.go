package mongo_test

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	permissionStore "github.com/tidepool-org/platform/permission/store"
	permissionStoreMongo "github.com/tidepool-org/platform/permission/store/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
	"github.com/tidepool-org/platform/user"
)

func NewPermission(sharerID string, userID string) bson.M {
	return bson.M{
		"sharerId": sharerID,
		"userId":   userID,
		"permissions": bson.M{
			"upload": bson.M{},
			"view":   bson.M{},
		},
	}
}

func NewPermissions(userID string) []interface{} {
	permissions := []interface{}{}
	permissions = append(permissions, NewPermission(user.NewID(), userID), NewPermission(userID, user.NewID()))
	return permissions
}

func ValidatePermissions(testMongoCollection *mongo.Collection, selector bson.M, expectedPermissions []interface{}) {
	var actualPermissions []bson.M
	opts := options.Find().SetProjection(bson.M{"_id": 0})
	cursor, err := testMongoCollection.Find(context.Background(), selector, opts)
	Expect(err).ToNot(HaveOccurred())
	Expect(cursor).ToNot(BeNil())
	Expect(cursor.All(context.Background(), &actualPermissions)).To(Succeed())
	Expect(actualPermissions).To(ConsistOf(expectedPermissions...))
}

var _ = Describe("Mongo", func() {
	var mongoConfig *permissionStoreMongo.Config
	var mongoStore *permissionStoreMongo.Store
	var permissionRepository permissionStore.PermissionsRepository

	BeforeEach(func() {
		mongoConfig = &permissionStoreMongo.Config{
			Config: storeStructuredMongoTest.NewConfig(),
			Secret: "secret",
		}
	})

	AfterEach(func() {
		if mongoStore != nil {
			mongoStore.Terminate(context.Background())
		}
	})

	Context("New", func() {
		It("returns an error if config is missing", func() {
			var err error
			mongoStore, err = permissionStoreMongo.NewStore(nil, logTest.NewLogger())
			Expect(err).To(MatchError("config is missing"))
			Expect(mongoStore).To(BeNil())
		})

		It("returns an error if base config is missing", func() {
			var err error
			mongoConfig.Config = nil
			mongoStore, err = permissionStoreMongo.NewStore(mongoConfig, logTest.NewLogger())
			Expect(err).To(MatchError("database config is empty"))
			Expect(mongoStore).To(BeNil())
		})

		It("returns an error if base config is invalid", func() {
			var err error
			mongoConfig.Config.Addresses = nil
			mongoStore, err = permissionStoreMongo.NewStore(mongoConfig, logTest.NewLogger())
			Expect(err).To(MatchError("connection options are invalid; error parsing uri: must have at least 1 host"))
			Expect(mongoStore).To(BeNil())
		})

		It("returns an error if config is invalid", func() {
			var err error
			mongoConfig.Secret = ""
			mongoStore, err = permissionStoreMongo.NewStore(mongoConfig, logTest.NewLogger())
			Expect(err).To(MatchError("config is invalid; secret is missing"))
			Expect(mongoStore).To(BeNil())
		})

		It("returns a new store and no error if successful", func() {
			var err error
			mongoStore, err = permissionStoreMongo.NewStore(mongoConfig, logTest.NewLogger())
			Expect(err).ToNot(HaveOccurred())
			Expect(mongoStore).ToNot(BeNil())
		})
	})

	Context("with a new store", func() {
		BeforeEach(func() {
			var err error
			mongoStore, err = permissionStoreMongo.NewStore(mongoConfig, logTest.NewLogger())
			Expect(err).ToNot(HaveOccurred())
			Expect(mongoStore).ToNot(BeNil())
		})

		Context("NewPermissionsRepository", func() {
			It("returns a new repository", func() {
				permissionRepository = mongoStore.NewPermissionsRepository()
				Expect(permissionRepository).ToNot(BeNil())
			})
		})

		Context("with a new repository", func() {
			BeforeEach(func() {
				permissionRepository = mongoStore.NewPermissionsRepository()
				Expect(permissionRepository).ToNot(BeNil())
			})

			Context("with persisted data", func() {
				var testMongoCollection *mongo.Collection
				var permissions []interface{}
				var ctx context.Context

				BeforeEach(func() {
					testMongoCollection = mongoStore.GetCollection("perms")
					permissions = NewPermissions(user.NewID())
					ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
				})

				JustBeforeEach(func() {
					_, err := testMongoCollection.InsertMany(context.Background(), permissions)
					Expect(err).ToNot(HaveOccurred())
				})

				Context("DestroyPermissionsForUserByID", func() {
					var destroyUserID string
					var destroyPermissions []interface{}

					BeforeEach(func() {
						destroyUserID = user.NewID()
						destroyPermissions = NewPermissions(destroyUserID)
					})

					JustBeforeEach(func() {
						_, err := testMongoCollection.InsertMany(context.Background(), destroyPermissions)
						Expect(err).ToNot(HaveOccurred())
					})

					It("succeeds if it successfully removes permissions", func() {
						Expect(permissionRepository.DestroyPermissionsForUserByID(ctx, destroyUserID)).To(Succeed())
					})

					It("returns an error if the context is missing", func() {
						Expect(permissionRepository.DestroyPermissionsForUserByID(nil, destroyUserID)).To(MatchError("context is missing"))
					})

					It("returns an error if the user id is missing", func() {
						Expect(permissionRepository.DestroyPermissionsForUserByID(ctx, "")).To(MatchError("user id is missing"))
					})

					It("has the correct stored permissions", func() {
						ValidatePermissions(testMongoCollection, bson.M{}, append(permissions, destroyPermissions...))
						Expect(permissionRepository.DestroyPermissionsForUserByID(ctx, destroyUserID)).To(Succeed())
						ValidatePermissions(testMongoCollection, bson.M{}, permissions)
					})
				})
			})
		})
	})
})
