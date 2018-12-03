package mongo_test

import (
	. "github.com/onsi/ginkgo"
)

// func NewProfile(profileID string, fullName string) bson.M {
// 	return bson.M{
// 		"_id":   profileID,
// 		"value": `{"profile":{"fullName":"` + fullName + `","patient":{"birthday":"2000-01-01","diagnosisDate":"2010-12-31","targetDevices":["dexcom","tandem"],"targetTimezone":"US/Pacific"}},"private":{"uploads":{"name":"","id":"1234567890","hash":"1234567890abcdef"}}}`,
// 	}
// }

// func NewProfiles() []interface{} {
// 	profiles := []interface{}{}
// 	profiles = append(profiles, NewProfile(id.New(), id.New()), NewProfile(id.New(), id.New()), NewProfile(id.New(), id.New()))
// 	return profiles
// }

// func ValidateProfiles(testMongoCollection *mgo.Collection, selector bson.M, expectedProfiles []interface{}) {
// 	var actualProfiles []interface{}
// 	Expect(testMongoCollection.Find(selector).All(&actualProfiles)).To(Succeed())
// 	Expect(actualProfiles).To(ConsistOf(expectedProfiles...))
// }

var _ = Describe("Mongo", func() {
	// 	var mongoConfig *storeStructuredMongo.Config
	// 	var mongoStore *mongo.Store
	// 	var mongoSession store.ProfilesSession

	// 	BeforeEach(func() {
	// 		mongoConfig = storeStructuredMongoTest.NewConfig()
	// 	})

	// 	AfterEach(func() {
	// 		if mongoSession != nil {
	// 			mongoSession.Close()
	// 		}
	// 		if mongoStore != nil {
	// 			mongoStore.Close()
	// 		}
	// 	})

	// 	Context("New", func() {
	// 		It("returns an error if unsuccessful", func() {
	// 			var err error
	// 			mongoStore, err = mongo.NewStore(nil, nil)
	// 			Expect(err).To(HaveOccurred())
	// 			Expect(mongoStore).To(BeNil())
	// 		})

	// 		It("returns a new store and no error if successful", func() {
	// 			var err error
	// 			mongoStore, err = mongo.NewStore(mongoConfig, null.NewLogger())
	// 			Expect(err).ToNot(HaveOccurred())
	// 			Expect(mongoStore).ToNot(BeNil())
	// 		})
	// 	})

	// 	Context("with a new store", func() {
	// 		BeforeEach(func() {
	// 			var err error
	// 			mongoStore, err = mongo.NewStore(mongoConfig, null.NewLogger())
	// 			Expect(err).ToNot(HaveOccurred())
	// 			Expect(mongoStore).ToNot(BeNil())
	// 		})

	// 		Context("NewProfilesSession", func() {
	// 			It("returns a new session", func() {
	// 				mongoSession = mongoStore.NewProfilesSession()
	// 				Expect(mongoSession).ToNot(BeNil())
	// 			})
	// 		})

	// 		Context("with a new session", func() {
	// 			BeforeEach(func() {
	// 				mongoSession = mongoStore.NewProfilesSession()
	// 				Expect(mongoSession).ToNot(BeNil())
	// 			})

	// 			Context("with persisted data", func() {
	// 				var testMongoSession *mgo.Session
	// 				var testMongoCollection *mgo.Collection
	// 				var profiles []interface{}

	// 				BeforeEach(func() {
	// 					testMongoSession = storeStructuredMongoTest.Session().Copy()
	// 					testMongoCollection = testMongoSession.DB(mongoConfig.Database).C(mongoConfig.CollectionPrefix + "seagull")
	// 					profiles = NewProfiles()
	// 				})

	// 				JustBeforeEach(func() {
	// 					Expect(testMongoCollection.Insert(profiles...)).To(Succeed())
	// 				})

	// 				AfterEach(func() {
	// 					if testMongoSession != nil {
	// 						testMongoSession.Close()
	// 					}
	// 				})

	// 				Context("GetProfileByID", func() {
	// 					var getProfileID string
	// 					var getProfileFullName string
	// 					var getProfile interface{}

	// 					BeforeEach(func() {
	// 						getProfileID = id.New()
	// 						getProfileFullName = id.New()
	// 						getProfile = NewProfile(getProfileID, getProfileFullName)
	// 					})

	// 					JustBeforeEach(func() {
	// 						Expect(testMongoCollection.Insert(getProfile)).To(Succeed())
	// 					})

	// 					It("succeeds if it successfully gets the profile by id", func() {
	// 						profile, err := mongoSession.GetProfileByID(getProfileID)
	// 						Expect(err).ToNot(HaveOccurred())
	// 						Expect(profile).ToNot(BeNil())
	// 						Expect(profile.FullName).ToNot(BeNil())
	// 						Expect(*profile.FullName).To(Equal(getProfileFullName))
	// 					})

	// 					It("returns no error and no profile if the profile id is not found", func() {
	// 						profile, err := mongoSession.GetProfileByID(id.New())
	// 						Expect(err).ToNot(HaveOccurred())
	// 						Expect(profile).To(BeNil())
	// 					})

	// 					It("returns an error if the profile id is missing", func() {
	// 						profile, err := mongoSession.GetProfileByID("")
	// 						Expect(err).To(MatchError("profile id is missing"))
	// 						Expect(profile).To(BeNil())
	// 					})

	// 					It("returns an error if the session is closed", func() {
	// 						mongoSession.Close()
	// 						profile, err := mongoSession.GetProfileByID(getProfileID)
	// 						Expect(err).To(MatchError("session closed"))
	// 						Expect(profile).To(BeNil())
	// 					})

	// 					Context("with no value", func() {
	// 						BeforeEach(func() {
	// 							getProfile.(bson.M)["value"] = nil
	// 						})

	// 						It("succeeds, but does not fill in the full name", func() {
	// 							profile, err := mongoSession.GetProfileByID(getProfileID)
	// 							Expect(err).ToNot(HaveOccurred())
	// 							Expect(profile).ToNot(BeNil())
	// 							Expect(profile.FullName).To(BeNil())
	// 						})
	// 					})

	// 					Context("with empty value", func() {
	// 						BeforeEach(func() {
	// 							getProfile.(bson.M)["value"] = ``
	// 						})

	// 						It("succeeds, but does not fill in the full name", func() {
	// 							profile, err := mongoSession.GetProfileByID(getProfileID)
	// 							Expect(err).ToNot(HaveOccurred())
	// 							Expect(profile).ToNot(BeNil())
	// 							Expect(profile.FullName).To(BeNil())
	// 						})
	// 					})

	// 					Context("with invalid JSON value", func() {
	// 						BeforeEach(func() {
	// 							getProfile.(bson.M)["value"] = `{`
	// 						})

	// 						It("succeeds, but does not fill in the full name", func() {
	// 							profile, err := mongoSession.GetProfileByID(getProfileID)
	// 							Expect(err).ToNot(HaveOccurred())
	// 							Expect(profile).ToNot(BeNil())
	// 							Expect(profile.FullName).To(BeNil())
	// 						})
	// 					})

	// 					Context("with valid value that does not contain profile", func() {
	// 						BeforeEach(func() {
	// 							getProfile.(bson.M)["value"] = `{}`
	// 						})

	// 						It("succeeds, but does not fill in the full name", func() {
	// 							profile, err := mongoSession.GetProfileByID(getProfileID)
	// 							Expect(err).ToNot(HaveOccurred())
	// 							Expect(profile).ToNot(BeNil())
	// 							Expect(profile.FullName).To(BeNil())
	// 						})
	// 					})

	// 					Context("with valid value that does not contain full name in profile", func() {
	// 						BeforeEach(func() {
	// 							getProfile.(bson.M)["value"] = `{"profile":{}}`
	// 						})

	// 						It("succeeds, but does not fill in the full name", func() {
	// 							profile, err := mongoSession.GetProfileByID(getProfileID)
	// 							Expect(err).ToNot(HaveOccurred())
	// 							Expect(profile).ToNot(BeNil())
	// 							Expect(profile.FullName).To(BeNil())
	// 						})
	// 					})
	// 				})

	// 				Context("DestroyProfileByID", func() {
	// 					var destroyProfileID string
	// 					var destroyProfile interface{}

	// 					BeforeEach(func() {
	// 						destroyProfileID = id.New()
	// 						destroyProfile = NewProfile(destroyProfileID, id.New())
	// 					})

	// 					JustBeforeEach(func() {
	// 						Expect(testMongoCollection.Insert(destroyProfile)).To(Succeed())
	// 					})

	// 					It("succeeds if it successfully removes profiles", func() {
	// 						Expect(mongoSession.DestroyProfileByID(destroyProfileID)).To(Succeed())
	// 					})

	// 					It("returns an error if the profile id is missing", func() {
	// 						Expect(mongoSession.DestroyProfileByID("")).To(MatchError("profile id is missing"))
	// 					})

	// 					It("returns an error if the session is closed", func() {
	// 						mongoSession.Close()
	// 						Expect(mongoSession.DestroyProfileByID(destroyProfileID)).To(MatchError("session closed"))
	// 					})

	// 					It("has the correct stored profiles", func() {
	// 						ValidateProfiles(testMongoCollection, bson.M{}, append(profiles, destroyProfile))
	// 						Expect(mongoSession.DestroyProfileByID(destroyProfileID)).To(Succeed())
	// 						ValidateProfiles(testMongoCollection, bson.M{}, profiles)
	// 					})
	// 				})
	// 			})
	// 		})
	// 	})
})
