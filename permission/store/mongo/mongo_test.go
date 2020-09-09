package mongo_test

import (
	"context"
	"encoding/base64"

	mgo "github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/crypto"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	permissionStore "github.com/tidepool-org/platform/permission/store"
	permissionStoreMongo "github.com/tidepool-org/platform/permission/store/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
	"github.com/tidepool-org/platform/user"
)

func NewPermission(groupID string, userID string) bson.M {
	encryptedGroupID, err := crypto.EncryptWithAES256UsingPassphrase([]byte(groupID), []byte("secret"))
	Expect(err).ToNot(HaveOccurred())

	return bson.M{
		"groupId": base64.StdEncoding.EncodeToString(encryptedGroupID),
		"userId":  userID,
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

func ValidatePermissions(testMongoCollection *mgo.Collection, selector bson.M, expectedPermissions []interface{}) {
	var actualPermissions []interface{}
	Expect(testMongoCollection.Find(selector).Select(bson.M{"_id": 0}).All(&actualPermissions)).To(Succeed())
	Expect(actualPermissions).To(ConsistOf(expectedPermissions...))
}

var _ = Describe("Mongo", func() {
	var mongoConfig *permissionStoreMongo.Config
	var mongoStore *permissionStoreMongo.Store
	var mongoSession permissionStore.PermissionsSession

	BeforeEach(func() {
		mongoConfig = &permissionStoreMongo.Config{
			Config: storeStructuredMongoTest.NewConfig(),
			Secret: "secret",
		}
	})

	AfterEach(func() {
		if mongoSession != nil {
			mongoSession.Close()
		}
		if mongoStore != nil {
			mongoStore.Close()
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
			Expect(err).To(MatchError("config is missing"))
			Expect(mongoStore).To(BeNil())
		})

		It("returns an error if base config is invalid", func() {
			var err error
			mongoConfig.Config.SetAddresses(nil)
			mongoStore, err = permissionStoreMongo.NewStore(mongoConfig, logTest.NewLogger())
			Expect(err).To(MatchError("config is invalid; addresses is missing"))
			Expect(mongoStore).To(BeNil())
		})

		It("returns an error if config is invalid", func() {
			var err error
			mongoConfig.Secret = ""
			mongoStore, err = permissionStoreMongo.NewStore(mongoConfig, logTest.NewLogger())
			Expect(err).To(MatchError("config is invalid; secret is missing"))
			Expect(mongoStore).To(BeNil())
		})

		It("returns an error if logger is missing", func() {
			var err error
			mongoStore, err = permissionStoreMongo.NewStore(mongoConfig, nil)
			Expect(err).To(MatchError("logger is missing"))
			Expect(mongoStore).To(BeNil())
		})

		It("returns a new store and no error if successful", func() {
			var err error
			mongoStore, err = permissionStoreMongo.NewStore(mongoConfig, logTest.NewLogger())
			mongoStore.WaitUntilStarted()
			Expect(err).ToNot(HaveOccurred())
			Expect(mongoStore).ToNot(BeNil())
		})
	})

	Context("with a new store", func() {
		BeforeEach(func() {
			var err error
			mongoStore, err = permissionStoreMongo.NewStore(mongoConfig, logTest.NewLogger())
			mongoStore.WaitUntilStarted()
			Expect(err).ToNot(HaveOccurred())
			Expect(mongoStore).ToNot(BeNil())
		})

		Context("NewPermissionsSession", func() {
			It("returns a new session", func() {
				mongoSession = mongoStore.NewPermissionsSession()
				Expect(mongoSession).ToNot(BeNil())
			})
		})

		Context("with a new session", func() {
			BeforeEach(func() {
				mongoSession = mongoStore.NewPermissionsSession()
				Expect(mongoSession).ToNot(BeNil())
			})

			Context("with persisted data", func() {
				var testMongoSession *mgo.Session
				var testMongoCollection *mgo.Collection
				var permissions []interface{}
				var ctx context.Context

				BeforeEach(func() {
					testMongoSession = storeStructuredMongoTest.Session().Copy()
					testMongoCollection = testMongoSession.DB(mongoConfig.Database).C(mongoConfig.CollectionPrefix + "perms")
					permissions = NewPermissions(user.NewID())
					ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
				})

				JustBeforeEach(func() {
					Expect(testMongoCollection.Insert(permissions...)).To(Succeed())
				})

				AfterEach(func() {
					if testMongoSession != nil {
						testMongoSession.Close()
					}
				})

				Context("DestroyPermissionsForUserByID", func() {
					var destroyUserID string
					var destroyPermissions []interface{}

					BeforeEach(func() {
						destroyUserID = user.NewID()
						destroyPermissions = NewPermissions(destroyUserID)
					})

					JustBeforeEach(func() {
						Expect(testMongoCollection.Insert(destroyPermissions...)).To(Succeed())
					})

					It("succeeds if it successfully removes permissions", func() {
						Expect(mongoSession.DestroyPermissionsForUserByID(ctx, destroyUserID)).To(Succeed())
					})

					It("returns an error if the context is missing", func() {
						Expect(mongoSession.DestroyPermissionsForUserByID(nil, destroyUserID)).To(MatchError("context is missing"))
					})

					It("returns an error if the user id is missing", func() {
						Expect(mongoSession.DestroyPermissionsForUserByID(ctx, "")).To(MatchError("user id is missing"))
					})

					It("returns an error if the session is closed", func() {
						mongoSession.Close()
						Expect(mongoSession.DestroyPermissionsForUserByID(ctx, destroyUserID)).To(MatchError("session closed"))
					})

					It("has the correct stored permissions", func() {
						ValidatePermissions(testMongoCollection, bson.M{}, append(permissions, destroyPermissions...))
						Expect(mongoSession.DestroyPermissionsForUserByID(ctx, destroyUserID)).To(Succeed())
						ValidatePermissions(testMongoCollection, bson.M{}, permissions)
					})
				})
			})
		})
	})
})
