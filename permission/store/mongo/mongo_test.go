package mongo_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"encoding/base64"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/permission/store"
	"github.com/tidepool-org/platform/permission/store/mongo"
	baseMongo "github.com/tidepool-org/platform/store/mongo"
	testMongo "github.com/tidepool-org/platform/test/mongo"
)

func NewPermission(groupID string, userID string) bson.M {
	encryptedGroupID, err := app.EncryptWithAES256UsingPassphrase([]byte(groupID), []byte("secret"))
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
	permissions = append(permissions, NewPermission(app.NewID(), userID), NewPermission(userID, app.NewID()))
	return permissions
}

func ValidatePermissions(testMongoCollection *mgo.Collection, selector bson.M, expectedPermissions []interface{}) {
	var actualPermissions []interface{}
	Expect(testMongoCollection.Find(selector).Select(bson.M{"_id": 0}).All(&actualPermissions)).To(Succeed())
	Expect(actualPermissions).To(ConsistOf(expectedPermissions...))
}

var _ = Describe("Mongo", func() {
	var mongoConfig *mongo.Config
	var mongoStore *mongo.Store
	var mongoSession store.Session

	BeforeEach(func() {
		mongoConfig = &mongo.Config{
			Config: &baseMongo.Config{
				Addresses:  testMongo.Address(),
				Database:   testMongo.Database(),
				Collection: testMongo.NewCollectionName(),
				Timeout:    app.DurationAsPointer(5 * time.Second),
			},
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
		It("returns an error if logger is missing", func() {
			var err error
			mongoStore, err = mongo.New(nil, mongoConfig)
			Expect(err).To(MatchError("mongo: logger is missing"))
			Expect(mongoStore).To(BeNil())
		})

		It("returns an error if base config is missing", func() {
			var err error
			mongoConfig.Config = nil
			mongoStore, err = mongo.New(log.NewNull(), mongoConfig)
			Expect(err).To(MatchError("mongo: config is missing"))
			Expect(mongoStore).To(BeNil())
		})

		It("returns an error if base config is invalid", func() {
			var err error
			mongoConfig.Config.Addresses = ""
			mongoStore, err = mongo.New(log.NewNull(), mongoConfig)
			Expect(err).To(MatchError("mongo: config is invalid; mongo: addresses is missing"))
			Expect(mongoStore).To(BeNil())
		})

		It("returns an error if config is missing", func() {
			var err error
			mongoStore, err = mongo.New(log.NewNull(), nil)
			Expect(err).To(MatchError("mongo: config is missing"))
			Expect(mongoStore).To(BeNil())
		})

		It("returns an error if config is invalid", func() {
			var err error
			mongoConfig.Secret = ""
			mongoStore, err = mongo.New(log.NewNull(), mongoConfig)
			Expect(err).To(MatchError("mongo: config is invalid; mongo: secret is missing"))
			Expect(mongoStore).To(BeNil())
		})

		It("returns a new store and no error if successful", func() {
			var err error
			mongoStore, err = mongo.New(log.NewNull(), mongoConfig)
			Expect(err).ToNot(HaveOccurred())
			Expect(mongoStore).ToNot(BeNil())
		})
	})

	Context("with a new store", func() {
		BeforeEach(func() {
			var err error
			mongoStore, err = mongo.New(log.NewNull(), mongoConfig)
			Expect(err).ToNot(HaveOccurred())
			Expect(mongoStore).ToNot(BeNil())
		})

		Context("NewSession", func() {
			It("returns an error if unsuccessful", func() {
				var err error
				mongoSession, err = mongoStore.NewSession(nil)
				Expect(err).To(HaveOccurred())
				Expect(mongoSession).To(BeNil())
			})

			It("returns a new session and no error if successful", func() {
				var err error
				mongoSession, err = mongoStore.NewSession(log.NewNull())
				Expect(err).ToNot(HaveOccurred())
				Expect(mongoSession).ToNot(BeNil())
			})
		})

		Context("with a new session", func() {
			BeforeEach(func() {
				var err error
				mongoSession, err = mongoStore.NewSession(log.NewNull())
				Expect(err).ToNot(HaveOccurred())
				Expect(mongoSession).ToNot(BeNil())
			})

			Context("with persisted data", func() {
				var testMongoSession *mgo.Session
				var testMongoCollection *mgo.Collection
				var permissions []interface{}

				BeforeEach(func() {
					testMongoSession = testMongo.Session().Copy()
					testMongoCollection = testMongoSession.DB(mongoConfig.Database).C(mongoConfig.Collection)
					permissions = NewPermissions(app.NewID())
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
						destroyUserID = app.NewID()
						destroyPermissions = NewPermissions(destroyUserID)
					})

					JustBeforeEach(func() {
						Expect(testMongoCollection.Insert(destroyPermissions...)).To(Succeed())
					})

					It("succeeds if it successfully removes permissions", func() {
						Expect(mongoSession.DestroyPermissionsForUserByID(destroyUserID)).To(Succeed())
					})

					It("returns an error if the user id is missing", func() {
						Expect(mongoSession.DestroyPermissionsForUserByID("")).To(MatchError("mongo: user id is missing"))
					})

					It("returns an error if the session is closed", func() {
						mongoSession.Close()
						Expect(mongoSession.DestroyPermissionsForUserByID(destroyUserID)).To(MatchError("mongo: session closed"))
					})

					It("has the correct stored permissions", func() {
						ValidatePermissions(testMongoCollection, bson.M{}, append(permissions, destroyPermissions...))
						Expect(mongoSession.DestroyPermissionsForUserByID(destroyUserID)).To(Succeed())
						ValidatePermissions(testMongoCollection, bson.M{}, permissions)
					})
				})
			})

			Context("GroupIDFromUserID", func() {
				It("returns an error if the user id is missing", func() {
					groupID, err := mongoSession.(*mongo.Session).GroupIDFromUserID("")
					Expect(err).To(MatchError("mongo: user id is missing"))
					Expect(groupID).To(BeEmpty())
				})

				DescribeTable("is successful for",
					func(userID string, expectedGroupID string) {
						groupID, err := mongoSession.(*mongo.Session).GroupIDFromUserID(userID)
						Expect(err).ToNot(HaveOccurred())
						Expect(groupID).To(Equal(expectedGroupID))
					},
					Entry("is example #1", "0cd1a5d68b", "NEHqFs6tA/2NRZ9oTPAHMA=="),
					Entry("is example #2", "b52201f96b", "rsWDsFcmDE2BgNfkNoiCnQ=="),
					Entry("is example #3", "46267a83eb", "cDuye1AVYPyAKvPy18+RqQ=="),
					Entry("is example #4", "982f600045", "1uO1mX4bFJ3hAC8g20l8fw=="),
					Entry("is example #5", "a06176bed7", "pMsbWdlanJldEYjkTokydA=="),
					Entry("is example #6", "d23b0a8786", "K35VY5wP6LVTpBTMUXv5OA=="),
					Entry("is example #7", "a011c16df7", "I/RdKRn3wMcaKtC/TRUIhg=="),
					Entry("is example #8", "8ea2d078f6", "AMFipBBZSHW0pP+985buzg=="),
					Entry("is example #9", "6128ef12fc", "X7DU5wxZYR9UDh780y+J9w=="),
					Entry("is example #10", "806d315a0b", "MgBbUF8XsHkj5ndZsJ0PmQ=="),
					Entry("is example #11", "aa16056cee", "iaR6v0jAWWXbDt4qS4s9HA=="),
					Entry("is example #12", "b4ba07dab4", "ARD9NlydxJZj7sJfz1UjOA=="),
					Entry("is example #13", "b4cae0bcbd", "YZGtYTIrvgSH8e7r9klFCw=="),
					Entry("is example #14", "7a1f209635", "CPzI+gdipBRYrl4ABZav4Q=="),
					Entry("is example #15", "68e70b285e", "k7kXyy3XBtoPKw9TwjLyew=="),
					Entry("is example #16", "bf33f09e3b", "HhLoSXNns8xVJh4YChWVEA=="),
					Entry("is example #17", "bb98bafa52", "4X10Q6lWGPnz2vmH7oc/6w=="),
					Entry("is example #18", "593f506db1", "ABGQBmS1eq08lnNzhMrVyg=="),
					Entry("is example #19", "480e0d76cb", "j21FL0lWNS1DU2A2dEwgMg=="),
					Entry("is example #20", "970d79a164", "3CyaEVxSX0HgvBCwEHiSBg=="),
				)
			})
		})
	})
})
