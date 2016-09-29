package mongo_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/profile/store"
	"github.com/tidepool-org/platform/profile/store/mongo"
	baseMongo "github.com/tidepool-org/platform/store/mongo"
	testMongo "github.com/tidepool-org/platform/test/mongo"
)

func NewProfile(profileID string, fullName string) bson.M {
	return bson.M{
		"_id":   profileID,
		"value": `{"profile":{"fullName":"` + fullName + `","patient":{"birthday":"2000-01-01","diagnosisDate":"2010-12-31","targetDevices":["dexcom","tandem"],"targetTimezone":"US/Pacific"}},"private":{"uploads":{"name":"","id":"1234567890","hash":"1234567890abcdef"}}}`,
	}
}

func NewProfiles() []interface{} {
	profiles := []interface{}{}
	profiles = append(profiles, NewProfile(app.NewID(), app.NewID()), NewProfile(app.NewID(), app.NewID()), NewProfile(app.NewID(), app.NewID()))
	return profiles
}

func ValidateProfiles(testMongoCollection *mgo.Collection, selector bson.M, expectedProfiles []interface{}) {
	var actualProfiles []interface{}
	Expect(testMongoCollection.Find(selector).All(&actualProfiles)).To(Succeed())
	Expect(actualProfiles).To(ConsistOf(expectedProfiles...))
}

var _ = Describe("Mongo", func() {
	var mongoConfig *baseMongo.Config
	var mongoStore *mongo.Store
	var mongoSession store.Session

	BeforeEach(func() {
		mongoConfig = &baseMongo.Config{
			Addresses:  testMongo.Address(),
			Database:   testMongo.Database(),
			Collection: testMongo.NewCollectionName(),
			Timeout:    app.DurationAsPointer(5 * time.Second),
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
		It("returns an error if unsuccessful", func() {
			var err error
			mongoStore, err = mongo.New(nil, nil)
			Expect(err).To(HaveOccurred())
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
				var profiles []interface{}

				BeforeEach(func() {
					testMongoSession = testMongo.Session().Copy()
					testMongoCollection = testMongoSession.DB(mongoConfig.Database).C(mongoConfig.Collection)
					profiles = NewProfiles()
				})

				JustBeforeEach(func() {
					Expect(testMongoCollection.Insert(profiles...)).To(Succeed())
				})

				AfterEach(func() {
					if testMongoSession != nil {
						testMongoSession.Close()
					}
				})

				Context("GetProfileByID", func() {
					var getProfileID string
					var getProfileFullName string
					var getProfile interface{}

					BeforeEach(func() {
						getProfileID = app.NewID()
						getProfileFullName = app.NewID()
						getProfile = NewProfile(getProfileID, getProfileFullName)
					})

					JustBeforeEach(func() {
						Expect(testMongoCollection.Insert(getProfile)).To(Succeed())
					})

					It("succeeds if it successfully gets the profile by id", func() {
						profile, err := mongoSession.GetProfileByID(getProfileID)
						Expect(err).ToNot(HaveOccurred())
						Expect(profile).ToNot(BeNil())
						Expect(profile.FullName).ToNot(BeNil())
						Expect(*profile.FullName).To(Equal(getProfileFullName))
					})

					It("returns no error and no profile if the profile id is not found", func() {
						profile, err := mongoSession.GetProfileByID(app.NewID())
						Expect(err).ToNot(HaveOccurred())
						Expect(profile).To(BeNil())
					})

					It("returns an error if the profile id is missing", func() {
						profile, err := mongoSession.GetProfileByID("")
						Expect(err).To(MatchError("mongo: profile id is missing"))
						Expect(profile).To(BeNil())
					})

					It("returns an error if the session is closed", func() {
						mongoSession.Close()
						profile, err := mongoSession.GetProfileByID(getProfileID)
						Expect(err).To(MatchError("mongo: session closed"))
						Expect(profile).To(BeNil())
					})

					Context("with no value", func() {
						BeforeEach(func() {
							getProfile.(bson.M)["value"] = nil
						})

						It("succeeds, but does not fill in the full name", func() {
							profile, err := mongoSession.GetProfileByID(getProfileID)
							Expect(err).ToNot(HaveOccurred())
							Expect(profile).ToNot(BeNil())
							Expect(profile.FullName).To(BeNil())
						})
					})

					Context("with empty value", func() {
						BeforeEach(func() {
							getProfile.(bson.M)["value"] = ``
						})

						It("succeeds, but does not fill in the full name", func() {
							profile, err := mongoSession.GetProfileByID(getProfileID)
							Expect(err).ToNot(HaveOccurred())
							Expect(profile).ToNot(BeNil())
							Expect(profile.FullName).To(BeNil())
						})
					})

					Context("with invalid JSON value", func() {
						BeforeEach(func() {
							getProfile.(bson.M)["value"] = `{`
						})

						It("succeeds, but does not fill in the full name", func() {
							profile, err := mongoSession.GetProfileByID(getProfileID)
							Expect(err).ToNot(HaveOccurred())
							Expect(profile).ToNot(BeNil())
							Expect(profile.FullName).To(BeNil())
						})
					})

					Context("with valid value that does not contain profile", func() {
						BeforeEach(func() {
							getProfile.(bson.M)["value"] = `{}`
						})

						It("succeeds, but does not fill in the full name", func() {
							profile, err := mongoSession.GetProfileByID(getProfileID)
							Expect(err).ToNot(HaveOccurred())
							Expect(profile).ToNot(BeNil())
							Expect(profile.FullName).To(BeNil())
						})
					})

					Context("with valid value that does not contain full name in profile", func() {
						BeforeEach(func() {
							getProfile.(bson.M)["value"] = `{"profile":{}}`
						})

						It("succeeds, but does not fill in the full name", func() {
							profile, err := mongoSession.GetProfileByID(getProfileID)
							Expect(err).ToNot(HaveOccurred())
							Expect(profile).ToNot(BeNil())
							Expect(profile.FullName).To(BeNil())
						})
					})
				})

				Context("DestroyProfileByID", func() {
					var destroyProfileID string
					var destroyProfile interface{}

					BeforeEach(func() {
						destroyProfileID = app.NewID()
						destroyProfile = NewProfile(destroyProfileID, app.NewID())
					})

					JustBeforeEach(func() {
						Expect(testMongoCollection.Insert(destroyProfile)).To(Succeed())
					})

					It("succeeds if it successfully removes profiles", func() {
						Expect(mongoSession.DestroyProfileByID(destroyProfileID)).To(Succeed())
					})

					It("returns an error if the profile id is missing", func() {
						Expect(mongoSession.DestroyProfileByID("")).To(MatchError("mongo: profile id is missing"))
					})

					It("returns an error if the session is closed", func() {
						mongoSession.Close()
						Expect(mongoSession.DestroyProfileByID(destroyProfileID)).To(MatchError("mongo: session closed"))
					})

					It("has the correct stored profiles", func() {
						ValidateProfiles(testMongoCollection, bson.M{}, append(profiles, destroyProfile))
						Expect(mongoSession.DestroyProfileByID(destroyProfileID)).To(Succeed())
						ValidateProfiles(testMongoCollection, bson.M{}, profiles)
					})
				})
			})
		})
	})
})
