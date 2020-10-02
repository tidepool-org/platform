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
	"github.com/tidepool-org/platform/request"
	requestTest "github.com/tidepool-org/platform/request/test"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
	"github.com/tidepool-org/platform/test"
	"github.com/tidepool-org/platform/user"
	userStoreStructured "github.com/tidepool-org/platform/user/store/structured"
	userStoreStructuredMongo "github.com/tidepool-org/platform/user/store/structured/mongo"
	userTest "github.com/tidepool-org/platform/user/test"
)

func AsInterfaceArray(userArray user.UserArray) []interface{} {
	if userArray == nil {
		return nil
	}
	array := make([]interface{}, len(userArray))
	for index, user := range userArray {
		array[index] = user
	}
	return array
}

var _ = Describe("Mongo", func() {
	var config *storeStructuredMongo.Config
	var logger *logTest.Logger
	var store *userStoreStructuredMongo.Store
	var repository userStoreStructured.UserRepository

	BeforeEach(func() {
		config = storeStructuredMongoTest.NewConfig()
		logger = logTest.NewLogger()
	})

	AfterEach(func() {
		if store != nil {
			err := store.Terminate(context.Background())
			Expect(err).ToNot(HaveOccurred())
		}
	})

	Context("NewStore", func() {
		It("returns an error when unsuccessful", func() {
			var err error
			params := storeStructuredMongo.Params{DatabaseConfig: nil}
			store, err = userStoreStructuredMongo.NewStore(params)
			errorsTest.ExpectEqual(err, errors.New("database config is empty"))
			Expect(store).To(BeNil())
		})

		It("returns a new store and no error when successful", func() {
			var err error
			params := storeStructuredMongo.Params{DatabaseConfig: config}
			store, err = userStoreStructuredMongo.NewStore(params)
			Expect(err).ToNot(HaveOccurred())
			Expect(store).ToNot(BeNil())
		})
	})

	Context("with a new store", func() {
		var mongoCollection *mongo.Collection

		BeforeEach(func() {
			var err error
			params := storeStructuredMongo.Params{DatabaseConfig: config}
			store, err = userStoreStructuredMongo.NewStore(params)
			Expect(err).ToNot(HaveOccurred())
			Expect(store).ToNot(BeNil())
			mongoCollection = store.GetCollection("users")
		})

		Context("EnsureIndexes", func() {
			It("returns successfully", func() {
				Expect(store.EnsureIndexes()).To(Succeed())
			})
		})

		Context("NewUserRepository", func() {
			It("returns a new repository", func() {
				repository = store.NewUserRepository()
				Expect(repository).ToNot(BeNil())
			})
		})

		Context("with a new repository", func() {
			var ctx context.Context

			BeforeEach(func() {
				Expect(store.EnsureIndexes()).To(Succeed())
				repository = store.NewUserRepository()
				ctx = log.NewContextWithLogger(context.Background(), logger)
			})

			Context("Get", func() {
				var id string
				var condition *request.Condition

				BeforeEach(func() {
					id = userTest.RandomID()
					condition = requestTest.RandomCondition()
				})

				It("returns an error when the context is missing", func() {
					ctx = nil
					result, err := repository.Get(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("context is missing"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the id is missing", func() {
					id = ""
					result, err := repository.Get(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("id is missing"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the id is invalid", func() {
					id = "invalid"
					result, err := repository.Get(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("id is invalid"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the condition is invalid", func() {
					condition.Revision = pointer.FromInt(-1)
					result, err := repository.Get(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("condition is invalid"))
					Expect(result).To(BeNil())
				})

				Context("with data", func() {
					var allResult user.UserArray
					var result *user.User

					BeforeEach(func() {
						allResult = userTest.RandomUserArray(3, 3)
						result = allResult[0]
						result.UserID = pointer.FromString(id)
						rand.Shuffle(len(allResult), func(i, j int) { allResult[i], allResult[j] = allResult[j], allResult[i] })
					})

					JustBeforeEach(func() {
						_, err := mongoCollection.InsertMany(ctx, AsInterfaceArray(allResult))
						Expect(err).ToNot(HaveOccurred())
					})

					AfterEach(func() {
						logger.AssertDebug("Get", log.Fields{"id": id})
					})

					It("returns nil when the id does not exist", func() {
						id = userTest.RandomID()
						Expect(repository.Get(ctx, id, condition)).To(BeNil())
					})

					When("the condition revision does not match", func() {
						BeforeEach(func() {
							condition.Revision = pointer.FromInt(*result.Revision + 1)
						})

						It("returns nil", func() {
							Expect(repository.Get(ctx, id, condition)).To(BeNil())
						})
					})

					conditionAssertions := func() {
						It("returns the result when the id exists", func() {
							Expect(repository.Get(ctx, id, condition)).To(Equal(result))
						})

						Context("when the result is marked as deleted", func() {
							BeforeEach(func() {
								result.ModifiedTime = pointer.FromTime(test.RandomTimeFromRange(*result.CreatedTime, time.Now()).Truncate(time.Second))
								result.DeletedTime = pointer.CloneTime(result.ModifiedTime)
							})

							It("returns nil", func() {
								Expect(repository.Get(ctx, id, condition)).To(BeNil())
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
								Expect(repository.Get(ctx, id, condition)).To(Equal(result))
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
								Expect(repository.Get(ctx, id, condition)).To(Equal(result))
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
								Expect(repository.Get(ctx, id, condition)).To(BeNil())
							})
						})
					})
				})
			})

			Context("Delete", func() {
				var id string
				var condition *request.Condition

				BeforeEach(func() {
					id = userTest.RandomID()
					condition = requestTest.RandomCondition()
				})

				It("returns an error when the context is missing", func() {
					ctx = nil
					result, err := repository.Delete(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("context is missing"))
					Expect(result).To(BeFalse())
				})

				It("returns an error when the id is missing", func() {
					id = ""
					result, err := repository.Delete(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("id is missing"))
					Expect(result).To(BeFalse())
				})

				It("returns an error when the id is invalid", func() {
					id = "invalid"
					result, err := repository.Delete(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("id is invalid"))
					Expect(result).To(BeFalse())
				})

				It("returns an error when the condition is invalid", func() {
					condition.Revision = pointer.FromInt(-1)
					result, err := repository.Delete(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("condition is invalid"))
					Expect(result).To(BeFalse())
				})

				Context("with data", func() {
					var original *user.User

					BeforeEach(func() {
						original = userTest.RandomUser()
						original.UserID = pointer.FromString(id)
					})

					JustBeforeEach(func() {
						_, err := mongoCollection.InsertOne(ctx, original)
						Expect(err).ToNot(HaveOccurred())
					})

					AfterEach(func() {
						if condition != nil {
							logger.AssertDebug("Delete", log.Fields{"id": id, "condition": condition})
						} else {
							logger.AssertDebug("Delete", log.Fields{"id": id})
						}
					})

					When("the original is marked as deleted", func() {
						BeforeEach(func() {
							original.ModifiedTime = pointer.FromTime(test.RandomTimeFromRange(*original.CreatedTime, time.Now()).Truncate(time.Second))
							original.DeletedTime = pointer.CloneTime(original.ModifiedTime)
						})

						It("returns false", func() {
							Expect(repository.Delete(ctx, id, condition)).To(BeFalse())
						})
					})

					When("the condition revision does not match", func() {
						BeforeEach(func() {
							condition.Revision = pointer.FromInt(*original.Revision + 1)
						})

						It("returns false", func() {
							Expect(repository.Delete(ctx, id, condition)).To(BeFalse())
						})
					})

					conditionAssertions := func() {
						Context("with updates", func() {
							It("returns true when the id exists", func() {
								matchAllFields := MatchAllFields(Fields{
									"UserID":        PointTo(Equal(id)),
									"Username":      Equal(original.Username),
									"PasswordHash":  Equal(original.PasswordHash),
									"Authenticated": Equal(original.Authenticated),
									"TermsAccepted": Equal(original.TermsAccepted),
									"Roles":         Equal(original.Roles),
									"CreatedTime":   Equal(original.CreatedTime),
									"ModifiedTime":  PointTo(BeTemporally("~", time.Now(), time.Second)),
									"DeletedTime":   PointTo(BeTemporally("~", time.Now(), time.Second)),
									"Revision":      PointTo(Equal(*original.Revision + 1)),
								})
								Expect(repository.Delete(ctx, id, condition)).To(BeTrue())
								storeResult := user.UserArray{}
								cursor, err := mongoCollection.Find(ctx, bson.M{"userid": id})
								Expect(err).ToNot(HaveOccurred())
								Expect(cursor).ToNot(BeNil())
								Expect(cursor.All(ctx, &storeResult)).To(Succeed())
								Expect(storeResult).To(HaveLen(1))
								Expect(*storeResult[0]).To(matchAllFields)
							})

							It("returns false when the id does not exist", func() {
								id = userTest.RandomID()
								Expect(repository.Delete(ctx, id, condition)).To(BeFalse())
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
				var id string
				var condition *request.Condition

				BeforeEach(func() {
					id = userTest.RandomID()
					condition = requestTest.RandomCondition()
				})

				It("returns an error when the context is missing", func() {
					ctx = nil
					destroyed, err := repository.Destroy(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("context is missing"))
					Expect(destroyed).To(BeFalse())
				})

				It("returns an error when the id is missing", func() {
					id = ""
					destroyed, err := repository.Destroy(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("id is missing"))
					Expect(destroyed).To(BeFalse())
				})

				It("returns an error when the id is invalid", func() {
					id = "invalid"
					destroyed, err := repository.Destroy(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("id is invalid"))
					Expect(destroyed).To(BeFalse())
				})

				It("returns an error when the condition is invalid", func() {
					condition.Revision = pointer.FromInt(-1)
					destroyed, err := repository.Destroy(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("condition is invalid"))
					Expect(destroyed).To(BeFalse())
				})

				Context("with data", func() {
					var original *user.User

					BeforeEach(func() {
						original = userTest.RandomUser()
						original.UserID = pointer.FromString(id)
					})

					JustBeforeEach(func() {
						_, err := mongoCollection.InsertOne(ctx, original)
						Expect(err).ToNot(HaveOccurred())
					})

					AfterEach(func() {
						if condition != nil {
							logger.AssertDebug("Destroy", log.Fields{"id": id, "condition": condition})
						} else {
							logger.AssertDebug("Destroy", log.Fields{"id": id})
						}
					})

					deletedAssertions := func() {
						It("returns false and does not destroy the original when the id does not exist", func() {
							id = userTest.RandomID()
							Expect(repository.Destroy(ctx, id, condition)).To(BeFalse())
							Expect(mongoCollection.CountDocuments(ctx, bson.M{"userid": original.UserID})).To(Equal(int64(1)))
						})

						It("returns false and does not destroy the original when the id exists, but the condition revision does not match", func() {
							condition.Revision = pointer.FromInt(*original.Revision + 1)
							Expect(repository.Destroy(ctx, id, condition)).To(BeFalse())
							Expect(mongoCollection.CountDocuments(ctx, bson.M{"userid": original.UserID})).To(Equal(int64(1)))
						})

						It("returns true and destroys the original when the id exists and the condition is missing", func() {
							condition = nil
							Expect(repository.Destroy(ctx, id, condition)).To(BeTrue())
							Expect(mongoCollection.CountDocuments(ctx, bson.M{"userid": original.UserID})).To(Equal(int64(0)))
						})

						It("returns true and destroys the original when the id exists and the condition revision is missing", func() {
							condition.Revision = nil
							Expect(repository.Destroy(ctx, id, condition)).To(BeTrue())
							Expect(mongoCollection.CountDocuments(ctx, bson.M{"userid": original.UserID})).To(Equal(int64(0)))
						})

						It("returns true and destroys the original when the id exists and the condition revision matches", func() {
							condition.Revision = pointer.CloneInt(original.Revision)
							Expect(repository.Destroy(ctx, id, condition)).To(BeTrue())
							Expect(mongoCollection.CountDocuments(ctx, bson.M{"userid": original.UserID})).To(Equal(int64(0)))
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
