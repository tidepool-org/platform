package mongo_test

import (
	"context"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/auth"
	authStore "github.com/tidepool-org/platform/auth/store"
	authStoreMongo "github.com/tidepool-org/platform/auth/store/mongo"
	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
	"github.com/tidepool-org/platform/user"
)

var _ = Describe("DeviceAuthorizationSession", func() {
	var mongoConfig *storeStructuredMongo.Config
	var mongoStore *authStoreMongo.Store
	var mongoSession authStore.DeviceAuthorizationSession

	BeforeEach(func() {
		mongoConfig = storeStructuredMongoTest.NewConfig()
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
			mongoStore, err = authStoreMongo.NewStore(nil, nil)
			Expect(err).To(HaveOccurred())
			Expect(mongoStore).To(BeNil())
		})

		It("returns a new store and no error if successful", func() {
			var err error
			mongoStore, err = authStoreMongo.NewStore(mongoConfig, logTest.NewLogger())
			Expect(err).ToNot(HaveOccurred())
			Expect(mongoStore).ToNot(BeNil())
		})
	})

	Context("with a new store", func() {
		BeforeEach(func() {
			var err error
			mongoStore, err = authStoreMongo.NewStore(mongoConfig, logTest.NewLogger())
			Expect(err).ToNot(HaveOccurred())
			Expect(mongoStore).ToNot(BeNil())
		})

		Context("NewSyncTaskSession", func() {
			It("returns a new session", func() {
				mongoSession = mongoStore.NewDeviceAuthorizationSession()
				Expect(mongoSession).ToNot(BeNil())
			})
		})

		Context("with a new session", func() {
			var testMongoSession *mgo.Session
			var testMongoCollection *mgo.Collection
			var ctx context.Context

			BeforeEach(func() {
				mongoSession = mongoStore.NewDeviceAuthorizationSession()
				Expect(mongoSession).ToNot(BeNil())
				testMongoSession = storeStructuredMongoTest.Session().Copy()
				testMongoCollection = testMongoSession.DB(mongoConfig.Database).C(mongoConfig.CollectionPrefix + "device_authorizations")
				ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
			})

			AfterEach(func() {
				if testMongoSession != nil {
					testMongoSession.Close()
				}
			})

			Context("CreateUserDeviceAuthorization", func() {
				var userID string
				var create *auth.DeviceAuthorizationCreate
				var authorization *auth.DeviceAuthorization

				BeforeEach(func() {
					userID = user.NewID()
					create = authTest.RandomDeviceAuthorizationCreate()
				})

				AfterEach(func() {
					if authorization != nil {
						Expect(testMongoCollection.Remove(bson.M{"id": authorization.ID})).To(Succeed())
						authorization = nil
					}
				})

				It("returns an error if the context is missing", func() {
					_, err := mongoSession.CreateUserDeviceAuthorization(nil, userID, create)
					Expect(err).To(MatchError("context is missing"))
				})

				It("returns an error if the session is closed", func() {
					mongoSession.Close()
					_, err := mongoSession.CreateUserDeviceAuthorization(ctx, userID, create)
					Expect(err).To(MatchError("session is closed"))
				})

				It("succeeds if it successfully creates the authorization", func() {
					authorization, err := mongoSession.CreateUserDeviceAuthorization(ctx, userID, create)
					Expect(err).To(Not(HaveOccurred()))
					Expect(authorization).To(Not(BeNil()))
				})

				It("has the correct stored authorizations", func() {
					created, _ := mongoSession.CreateUserDeviceAuthorization(ctx, userID, create)
					authorization = created
					Expect(created).NotTo(BeNil())

					expected := &auth.DeviceAuthorization{}
					Expect(testMongoCollection.Find(bson.M{"id": created.ID}).One(expected)).To(Succeed())
					Expect(expected.ID).NotTo(BeEmpty())
					Expect(created.ID).To(Equal(expected.ID))
				})
			})

			Context("GetUserDeviceAuthorization", func() {
				var userID string
				var authorization *auth.DeviceAuthorization

				BeforeEach(func() {
					userID = user.NewID()
					create := authTest.RandomDeviceAuthorizationCreate()
					created, err := mongoSession.CreateUserDeviceAuthorization(ctx, userID, create)
					Expect(err).To(Not(HaveOccurred()))
					authorization = created
				})

				AfterEach(func() {
					if authorization != nil {
						Expect(testMongoCollection.Remove(bson.M{"id": authorization.ID})).To(Succeed())
						authorization = nil
					}
				})

				It("returns an error if the context is missing", func() {
					_, err := mongoSession.GetUserDeviceAuthorization(nil, userID, authorization.ID)
					Expect(err).To(MatchError("context is missing"))
				})

				It("returns an error if the session is closed", func() {
					mongoSession.Close()
					_, err := mongoSession.GetUserDeviceAuthorization(ctx, userID, authorization.ID)
					Expect(err).To(MatchError("session is closed"))
				})

				It("succeeds if it finds the authorization", func() {
					found, err := mongoSession.GetUserDeviceAuthorization(ctx, userID, authorization.ID)
					Expect(err).To(Not(HaveOccurred()))
					Expect(found).To(Not(BeNil()))
					Expect(found.ID).To(Equal(authorization.ID))
					Expect(found.UserID).To(Equal(userID))
				})

				It("succeeds if it doesn't find the authorization", func() {
					found, err := mongoSession.GetUserDeviceAuthorization(ctx, user.NewID(), authTest.RandomDeviceAuthorizationID())
					Expect(err).To(Not(HaveOccurred()))
					Expect(found).To(BeNil())
				})

				It("doesn't find the authorization if the user id doesn't match", func() {
					found, err := mongoSession.GetUserDeviceAuthorization(ctx, user.NewID(), authorization.ID)
					Expect(err).To(Not(HaveOccurred()))
					Expect(found).To(BeNil())
				})
			})

			Context("GetDeviceAuthorization", func() {
				var authorization *auth.DeviceAuthorization

				BeforeEach(func() {
					create := authTest.RandomDeviceAuthorizationCreate()
					created, err := mongoSession.CreateUserDeviceAuthorization(ctx, user.NewID(), create)
					Expect(err).To(Not(HaveOccurred()))
					authorization = created
				})

				AfterEach(func() {
					if authorization != nil {
						Expect(testMongoCollection.Remove(bson.M{"id": authorization.ID})).To(Succeed())
						authorization = nil
					}
				})

				It("returns an error if the context is missing", func() {
					_, err := mongoSession.GetDeviceAuthorizationByToken(nil, authorization.Token)
					Expect(err).To(MatchError("context is missing"))
				})

				It("returns an error if the session is closed", func() {
					mongoSession.Close()
					_, err := mongoSession.GetDeviceAuthorizationByToken(ctx, authorization.Token)
					Expect(err).To(MatchError("session is closed"))
				})

				It("succeeds if it finds the authorization", func() {
					found, err := mongoSession.GetDeviceAuthorizationByToken(ctx, authorization.Token)
					Expect(found).To(Not(BeNil()))
					Expect(err).To(Not(HaveOccurred()))
					Expect(found.ID).To(Equal(authorization.ID))
					Expect(found.Token).To(Equal(authorization.Token))
				})

				It("succeeds if it doesn't find the authorization", func() {
					found, err := mongoSession.GetDeviceAuthorizationByToken(ctx, authTest.RandomDeviceAuthorizationToken())
					Expect(err).To(Not(HaveOccurred()))
					Expect(found).To(BeNil())
				})
			})

			Context("UpdateDeviceAuthorization", func() {
				var authorization *auth.DeviceAuthorization
				var update *auth.DeviceAuthorizationUpdate

				BeforeEach(func() {
					create := authTest.RandomDeviceAuthorizationCreate()
					update = authTest.RandomDeviceAuthorizationUpdate()
					update.Status = auth.DeviceAuthorizationSuccessful
					created, err := mongoSession.CreateUserDeviceAuthorization(ctx, user.NewID(), create)
					Expect(err).To(Not(HaveOccurred()))
					authorization = created
				})

				AfterEach(func() {
					if authorization != nil {
						Expect(testMongoCollection.Remove(bson.M{"id": authorization.ID})).To(Succeed())
						authorization = nil
					}
				})

				It("returns an error if the context is missing", func() {
					_, err := mongoSession.UpdateDeviceAuthorization(nil, authorization.ID, update)
					Expect(err).To(MatchError("context is missing"))
				})

				It("returns an error if the session is closed", func() {
					mongoSession.Close()
					_, err := mongoSession.UpdateDeviceAuthorization(ctx, authorization.ID, update)
					Expect(err).To(MatchError("session is closed"))
				})

				It("succeeds if it updates the authorization", func() {
					updated, err := mongoSession.UpdateDeviceAuthorization(ctx, authorization.ID, update)
					Expect(updated).To(Not(BeNil()))
					Expect(err).To(Not(HaveOccurred()))
					Expect(updated.ID).To(Equal(authorization.ID))
				})

				It("updates the expected attributes", func() {
					updated, err := mongoSession.UpdateDeviceAuthorization(ctx, authorization.ID, update)
					Expect(updated).To(Not(BeNil()))
					Expect(err).To(Not(HaveOccurred()))
					Expect(updated.ID).To(Equal(authorization.ID))
					Expect(updated.BundleID).To(Equal(update.BundleID))
					Expect(updated.VerificationCode).To(Equal(update.VerificationCode))
					Expect(updated.DeviceCheckToken).To(Equal(update.DeviceCheckToken))
					Expect(updated.Status).To(Equal(update.Status))
				})

				It("returns an error if the authorization is already completed", func() {
					updated, err := mongoSession.UpdateDeviceAuthorization(ctx, authorization.ID, update)
					Expect(updated).To(Not(BeNil()))
					Expect(err).To(Not(HaveOccurred()))
					_, err = mongoSession.UpdateDeviceAuthorization(ctx, authorization.ID, update)
					Expect(err).To(HaveOccurred())
				})

				It("returns an error if it doesn't find the authorization", func() {
					updated, err := mongoSession.UpdateDeviceAuthorization(ctx, authTest.RandomDeviceAuthorizationID(), update)
					Expect(err).To(HaveOccurred())
					Expect(updated).To(BeNil())
				})
			})
		})
	})
})
