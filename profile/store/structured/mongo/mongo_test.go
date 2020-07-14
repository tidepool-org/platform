package mongo_test

import (
	"context"
	"math/rand"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/profile"
	profileStoreStructured "github.com/tidepool-org/platform/profile/store/structured"
	profileStoreStructuredMongo "github.com/tidepool-org/platform/profile/store/structured/mongo"
	profileTest "github.com/tidepool-org/platform/profile/test"
	"github.com/tidepool-org/platform/request"
	requestTest "github.com/tidepool-org/platform/request/test"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
	"github.com/tidepool-org/platform/test"
	userTest "github.com/tidepool-org/platform/user/test"
)

func AsInterfaceArray(profileArray profile.ProfileArray) []interface{} {
	if profileArray == nil {
		return nil
	}
	array := make([]interface{}, len(profileArray))
	for index, profile := range profileArray {
		array[index] = profile
	}
	return array
}

var _ = Describe("Mongo", func() {
	var config *storeStructuredMongo.Config
	var logger *logTest.Logger
	var store *profileStoreStructuredMongo.Store
	var session profileStoreStructured.MetaRepository

	BeforeEach(func() {
		config = storeStructuredMongoTest.NewConfig()
		logger = logTest.NewLogger()
	})

	AfterEach(func() {
		if store != nil {
			store.Terminate(nil)
		}
	})

	Context("NewStore", func() {
		It("returns an error when unsuccessful", func() {
			var err error
			params := storeStructuredMongo.Params{DatabaseConfig: nil}
			store, err = profileStoreStructuredMongo.NewStore(params)
			errorsTest.ExpectEqual(err, errors.New("database config is empty"))
			Expect(store).To(BeNil())
		})

		It("returns a new store and no error when successful", func() {
			var err error
			params := storeStructuredMongo.Params{DatabaseConfig: config}
			store, err = profileStoreStructuredMongo.NewStore(params)
			Expect(err).ToNot(HaveOccurred())
			Expect(store).ToNot(BeNil())
		})
	})

	Context("with a new store", func() {
		var collection *mongo.Collection

		BeforeEach(func() {
			var err error
			params := storeStructuredMongo.Params{DatabaseConfig: config}
			store, err = profileStoreStructuredMongo.NewStore(params)
			Expect(err).ToNot(HaveOccurred())
			Expect(store).ToNot(BeNil())
			collection = store.GetCollection("seagull")
		})

		Context("NewSession", func() {
			It("returns a new session", func() {
				session = store.NewMetaRepository()
				Expect(session).ToNot(BeNil())
			})
		})

		Context("with a new session", func() {
			var ctx context.Context

			BeforeEach(func() {
				session = store.NewMetaRepository()
				ctx = log.NewContextWithLogger(context.Background(), logger)
			})

			Context("Get", func() {
				var userID string
				var condition *request.Condition

				BeforeEach(func() {
					userID = userTest.RandomID()
					condition = requestTest.RandomCondition()
				})

				It("returns an error when the context is missing", func() {
					ctx = nil
					result, err := session.Get(ctx, userID, condition)
					errorsTest.ExpectEqual(err, errors.New("context is missing"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the user id is missing", func() {
					userID = ""
					result, err := session.Get(ctx, userID, condition)
					errorsTest.ExpectEqual(err, errors.New("user id is missing"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the user id is invalid", func() {
					userID = "invalid"
					result, err := session.Get(ctx, userID, condition)
					errorsTest.ExpectEqual(err, errors.New("user id is invalid"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the condition is invalid", func() {
					condition.Revision = pointer.FromInt(-1)
					result, err := session.Get(ctx, userID, condition)
					errorsTest.ExpectEqual(err, errors.New("condition is invalid"))
					Expect(result).To(BeNil())
				})

				Context("with data", func() {
					var allResult profile.ProfileArray
					var result *profile.Profile

					BeforeEach(func() {
						allResult = profileTest.RandomProfileArray(3, 3)
						result = allResult[0]
						result.UserID = pointer.FromString(userID)
						rand.Shuffle(len(allResult), func(i, j int) { allResult[i], allResult[j] = allResult[j], allResult[i] })
					})

					JustBeforeEach(func() {
						_, err := collection.InsertMany(context.Background(), AsInterfaceArray(allResult))
						Expect(err).ToNot(HaveOccurred())
					})

					AfterEach(func() {
						logger.AssertDebug("Get", log.Fields{"userId": userID})
					})

					It("returns nil when the id does not exist", func() {
						userID = userTest.RandomID()
						Expect(session.Get(ctx, userID, condition)).To(BeNil())
					})

					When("the condition revision does not match", func() {
						BeforeEach(func() {
							condition.Revision = pointer.FromInt(*result.Revision + 1)
						})

						It("returns nil", func() {
							Expect(session.Get(ctx, userID, condition)).To(BeNil())
						})
					})

					conditionAssertions := func() {
						It("returns the result when the user id exists", func() {
							Expect(session.Get(ctx, userID, condition)).To(Equal(result))
						})

						Context("when the result is marked as deleted", func() {
							BeforeEach(func() {
								result.ModifiedTime = pointer.FromTime(test.RandomTimeFromRange(*result.CreatedTime, time.Now()).Truncate(time.Second))
								result.DeletedTime = pointer.CloneTime(result.ModifiedTime)
							})

							It("returns nil", func() {
								Expect(session.Get(ctx, userID, condition)).To(BeNil())
							})
						})
					}

					When("the condition is missing", func() {
						BeforeEach(func() {
							condition = nil
						})

						conditionAssertions()

						Context("when the revision is missing", func() {
							BeforeEach(func() {
								result.Revision = nil
							})

							It("returns the result with revision 0", func() {
								result.Revision = pointer.FromInt(0)
								Expect(session.Get(ctx, userID, condition)).To(Equal(result))
							})
						})
					})

					When("the condition revision is missing", func() {
						BeforeEach(func() {
							condition.Revision = nil
						})

						conditionAssertions()

						Context("when the revision is missing", func() {
							BeforeEach(func() {
								result.Revision = nil
							})

							It("returns the result with revision 0", func() {
								result.Revision = pointer.FromInt(0)
								Expect(session.Get(ctx, userID, condition)).To(Equal(result))
							})
						})
					})

					When("the condition revision matches", func() {
						BeforeEach(func() {
							condition.Revision = pointer.CloneInt(result.Revision)
						})

						conditionAssertions()

						Context("when the revision is missing", func() {
							BeforeEach(func() {
								result.Revision = nil
							})

							It("returns nil", func() {
								Expect(session.Get(ctx, userID, condition)).To(BeNil())
							})
						})
					})
				})
			})

			Context("Delete", func() {
				var userID string
				var condition *request.Condition

				BeforeEach(func() {
					userID = userTest.RandomID()
					condition = requestTest.RandomCondition()
				})

				It("returns an error when the context is missing", func() {
					ctx = nil
					result, err := session.Delete(ctx, userID, condition)
					errorsTest.ExpectEqual(err, errors.New("context is missing"))
					Expect(result).To(BeFalse())
				})

				It("returns an error when the user id is missing", func() {
					userID = ""
					result, err := session.Delete(ctx, userID, condition)
					errorsTest.ExpectEqual(err, errors.New("user id is missing"))
					Expect(result).To(BeFalse())
				})

				It("returns an error when the user id is invalid", func() {
					userID = "invalid"
					result, err := session.Delete(ctx, userID, condition)
					errorsTest.ExpectEqual(err, errors.New("user id is invalid"))
					Expect(result).To(BeFalse())
				})

				It("returns an error when the condition is invalid", func() {
					condition.Revision = pointer.FromInt(-1)
					result, err := session.Delete(ctx, userID, condition)
					errorsTest.ExpectEqual(err, errors.New("condition is invalid"))
					Expect(result).To(BeFalse())
				})

				Context("with data", func() {
					var original *profile.Profile

					BeforeEach(func() {
						original = profileTest.RandomProfile()
						original.UserID = pointer.FromString(userID)
					})

					JustBeforeEach(func() {
						_, err := collection.InsertOne(context.Background(), original)
						Expect(err).ToNot(HaveOccurred())
					})

					AfterEach(func() {
						if condition != nil {
							logger.AssertDebug("Delete", log.Fields{"userId": userID, "condition": condition})
						} else {
							logger.AssertDebug("Delete", log.Fields{"userId": userID})
						}
					})

					When("the original is marked as deleted", func() {
						BeforeEach(func() {
							original.ModifiedTime = pointer.FromTime(test.RandomTimeFromRange(*original.CreatedTime, time.Now()).Truncate(time.Second))
							original.DeletedTime = pointer.CloneTime(original.ModifiedTime)
						})

						It("returns false", func() {
							Expect(session.Delete(ctx, userID, condition)).To(BeFalse())
						})
					})

					When("the condition revision does not match", func() {
						BeforeEach(func() {
							condition.Revision = pointer.FromInt(*original.Revision + 1)
						})

						It("returns false", func() {
							Expect(session.Delete(ctx, userID, condition)).To(BeFalse())
						})
					})

					conditionAssertions := func() {
						Context("with updates", func() {
							It("returns true when the user id exists", func() {
								matchAllFields := MatchAllFields(Fields{
									"UserID":       Equal(original.UserID),
									"Value":        Equal(original.Value),
									"CreatedTime":  Equal(original.CreatedTime),
									"ModifiedTime": PointTo(BeTemporally("~", time.Now(), time.Second)),
									"DeletedTime":  PointTo(BeTemporally("~", time.Now(), time.Second)),
									"Revision":     PointTo(Equal(*original.Revision + 1)),
									"FullName":     BeNil(),
								})
								Expect(session.Delete(ctx, userID, condition)).To(BeTrue())
								storeResult := profile.ProfileArray{}
								cursor, err := collection.Find(context.Background(), bson.M{"userId": userID})
								Expect(err).ToNot(HaveOccurred())
								Expect(cursor).ToNot(BeNil())
								Expect(cursor.All(context.Background(), &storeResult)).To(Succeed())
								Expect(storeResult).To(HaveLen(1))
								Expect(*storeResult[0]).To(matchAllFields)
							})

							It("returns false when the user id does not exist", func() {
								userID = userTest.RandomID()
								Expect(session.Delete(ctx, userID, condition)).To(BeFalse())
							})
						})
					}

					When("the condition is missing", func() {
						BeforeEach(func() {
							condition = nil
						})

						conditionAssertions()
					})

					When("the condition revision is missing", func() {
						BeforeEach(func() {
							condition.Revision = nil
						})

						conditionAssertions()
					})

					When("the condition revision matches", func() {
						BeforeEach(func() {
							condition.Revision = pointer.CloneInt(original.Revision)
						})

						conditionAssertions()
					})
				})
			})

			Context("Destroy", func() {
				var userID string
				var condition *request.Condition

				BeforeEach(func() {
					userID = userTest.RandomID()
					condition = requestTest.RandomCondition()
				})

				It("returns an error when the context is missing", func() {
					ctx = nil
					destroyed, err := session.Destroy(ctx, userID, condition)
					errorsTest.ExpectEqual(err, errors.New("context is missing"))
					Expect(destroyed).To(BeFalse())
				})

				It("returns an error when the user id is missing", func() {
					userID = ""
					destroyed, err := session.Destroy(ctx, userID, condition)
					errorsTest.ExpectEqual(err, errors.New("user id is missing"))
					Expect(destroyed).To(BeFalse())
				})

				It("returns an error when the user id is invalid", func() {
					userID = "invalid"
					destroyed, err := session.Destroy(ctx, userID, condition)
					errorsTest.ExpectEqual(err, errors.New("user id is invalid"))
					Expect(destroyed).To(BeFalse())
				})

				It("returns an error when the condition is invalid", func() {
					condition.Revision = pointer.FromInt(-1)
					destroyed, err := session.Destroy(ctx, userID, condition)
					errorsTest.ExpectEqual(err, errors.New("condition is invalid"))
					Expect(destroyed).To(BeFalse())
				})

				Context("with data", func() {
					var original *profile.Profile

					BeforeEach(func() {
						original = profileTest.RandomProfile()
						original.UserID = pointer.FromString(userID)
					})

					JustBeforeEach(func() {
						_, err := collection.InsertOne(context.Background(), original)
						Expect(err).ToNot(HaveOccurred())
					})

					AfterEach(func() {
						if condition != nil {
							logger.AssertDebug("Destroy", log.Fields{"userId": userID, "condition": condition})
						} else {
							logger.AssertDebug("Destroy", log.Fields{"userId": userID})
						}
					})

					deletedAssertions := func() {
						It("returns false and does not destroy the original when the user id does not exist", func() {
							userID = userTest.RandomID()
							Expect(session.Destroy(ctx, userID, condition)).To(BeFalse())
							Expect(collection.CountDocuments(context.Background(), bson.M{"userId": original.UserID})).To(Equal(int64(1)))
						})

						It("returns false and does not destroy the original when the user id exists, but the condition revision does not match", func() {
							condition.Revision = pointer.FromInt(*original.Revision + 1)
							Expect(session.Destroy(ctx, userID, condition)).To(BeFalse())
							Expect(collection.CountDocuments(context.Background(), bson.M{"userId": original.UserID})).To(Equal(int64(1)))
						})

						It("returns true and destroys the original when the user id exists and the condition is missing", func() {
							condition = nil
							Expect(session.Destroy(ctx, userID, condition)).To(BeTrue())
							Expect(collection.CountDocuments(context.Background(), bson.M{"userId": original.UserID})).To(Equal(int64(0)))
						})

						It("returns true and destroys the original when the user id exists and the condition revision is missing", func() {
							condition.Revision = nil
							Expect(session.Destroy(ctx, userID, condition)).To(BeTrue())
							Expect(collection.CountDocuments(context.Background(), bson.M{"userId": original.UserID})).To(Equal(int64(0)))
						})

						It("returns true and destroys the original when the user id exists and the condition revision matches", func() {
							condition.Revision = pointer.CloneInt(original.Revision)
							Expect(session.Destroy(ctx, userID, condition)).To(BeTrue())
							Expect(collection.CountDocuments(context.Background(), bson.M{"userId": original.UserID})).To(Equal(int64(0)))
						})
					}

					When("the original is not deleted", func() {
						BeforeEach(func() {
							original.DeletedTime = nil
						})

						deletedAssertions()
					})

					When("the original is deleted", func() {
						BeforeEach(func() {
							original.ModifiedTime = pointer.FromTime(test.RandomTimeFromRange(*original.CreatedTime, time.Now()).Truncate(time.Second))
							original.DeletedTime = pointer.CloneTime(original.ModifiedTime)
						})

						deletedAssertions()
					})
				})
			})
		})
	})
})
