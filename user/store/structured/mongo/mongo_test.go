package mongo_test

import (
	"context"
	"math/rand"
	"time"

	mgo "github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
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
	var session userStoreStructured.Session

	BeforeEach(func() {
		config = storeStructuredMongoTest.NewConfig()
		logger = logTest.NewLogger()
	})

	AfterEach(func() {
		if session != nil {
			session.Close()
		}
		if store != nil {
			store.Close()
		}
	})

	Context("NewStore", func() {
		It("returns an error when unsuccessful", func() {
			var err error
			store, err = userStoreStructuredMongo.NewStore(nil, logger)
			errorsTest.ExpectEqual(err, errors.New("config is missing"))
			Expect(store).To(BeNil())
		})

		It("returns a new store and no error when successful", func() {
			var err error
			store, err = userStoreStructuredMongo.NewStore(config, logger)
			Expect(err).ToNot(HaveOccurred())
			Expect(store).ToNot(BeNil())
		})
	})

	Context("with a new store", func() {
		var mgoSession *mgo.Session
		var mgoCollection *mgo.Collection

		BeforeEach(func() {
			var err error
			store, err = userStoreStructuredMongo.NewStore(config, logger)
			Expect(err).ToNot(HaveOccurred())
			Expect(store).ToNot(BeNil())
			mgoSession = storeStructuredMongoTest.Session().Copy()
			mgoCollection = mgoSession.DB(config.Database).C(config.CollectionPrefix + "users")
		})

		AfterEach(func() {
			if mgoSession != nil {
				mgoSession.Close()
			}
		})

		Context("EnsureIndexes", func() {
			It("returns successfully", func() {
				Expect(store.EnsureIndexes()).To(Succeed())
			})
		})

		Context("NewSession", func() {
			It("returns a new session", func() {
				session = store.NewSession()
				Expect(session).ToNot(BeNil())
			})
		})

		Context("with a new session", func() {
			var ctx context.Context

			BeforeEach(func() {
				Expect(store.EnsureIndexes()).To(Succeed())
				session = store.NewSession()
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
					result, err := session.Get(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("context is missing"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the id is missing", func() {
					id = ""
					result, err := session.Get(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("id is missing"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the id is invalid", func() {
					id = "invalid"
					result, err := session.Get(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("id is invalid"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the condition is invalid", func() {
					condition.Revision = pointer.FromInt(-1)
					result, err := session.Get(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("condition is invalid"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the session is closed", func() {
					session.Close()
					result, err := session.Get(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("session closed"))
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
						Expect(mgoCollection.Insert(AsInterfaceArray(allResult)...)).To(Succeed())
					})

					AfterEach(func() {
						logger.AssertDebug("Get", log.Fields{"id": id})
					})

					It("returns nil when the id does not exist", func() {
						id = userTest.RandomID()
						Expect(session.Get(ctx, id, condition)).To(BeNil())
					})

					When("the condition revision does not match", func() {
						BeforeEach(func() {
							condition.Revision = pointer.FromInt(*result.Revision + 1)
						})

						It("returns nil", func() {
							Expect(session.Get(ctx, id, condition)).To(BeNil())
						})
					})

					conditionAssertions := func() {
						It("returns the result when the id exists", func() {
							Expect(session.Get(ctx, id, condition)).To(Equal(result))
						})

						Context("when the result is marked as deleted", func() {
							BeforeEach(func() {
								result.ModifiedTime = pointer.FromTime(test.RandomTimeFromRange(*result.CreatedTime, time.Now()).Truncate(time.Second))
								result.DeletedTime = pointer.CloneTime(result.ModifiedTime)
							})

							It("returns nil", func() {
								Expect(session.Get(ctx, id, condition)).To(BeNil())
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
								Expect(session.Get(ctx, id, condition)).To(Equal(result))
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
								Expect(session.Get(ctx, id, condition)).To(Equal(result))
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
								Expect(session.Get(ctx, id, condition)).To(BeNil())
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
					result, err := session.Delete(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("context is missing"))
					Expect(result).To(BeFalse())
				})

				It("returns an error when the id is missing", func() {
					id = ""
					result, err := session.Delete(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("id is missing"))
					Expect(result).To(BeFalse())
				})

				It("returns an error when the id is invalid", func() {
					id = "invalid"
					result, err := session.Delete(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("id is invalid"))
					Expect(result).To(BeFalse())
				})

				It("returns an error when the condition is invalid", func() {
					condition.Revision = pointer.FromInt(-1)
					result, err := session.Delete(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("condition is invalid"))
					Expect(result).To(BeFalse())
				})

				It("returns an error when the session is closed", func() {
					session.Close()
					result, err := session.Delete(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("session closed"))
					Expect(result).To(BeFalse())
				})

				Context("with data", func() {
					var original *user.User

					BeforeEach(func() {
						original = userTest.RandomUser()
						original.UserID = pointer.FromString(id)
					})

					JustBeforeEach(func() {
						Expect(mgoCollection.Insert(original)).To(Succeed())
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
							Expect(session.Delete(ctx, id, condition)).To(BeFalse())
						})
					})

					When("the condition revision does not match", func() {
						BeforeEach(func() {
							condition.Revision = pointer.FromInt(*original.Revision + 1)
						})

						It("returns false", func() {
							Expect(session.Delete(ctx, id, condition)).To(BeFalse())
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
								Expect(session.Delete(ctx, id, condition)).To(BeTrue())
								storeResult := user.UserArray{}
								Expect(mgoCollection.Find(bson.M{"userid": id}).All(&storeResult)).To(Succeed())
								Expect(storeResult).To(HaveLen(1))
								Expect(*storeResult[0]).To(matchAllFields)
							})

							It("returns false when the id does not exist", func() {
								id = userTest.RandomID()
								Expect(session.Delete(ctx, id, condition)).To(BeFalse())
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
					destroyed, err := session.Destroy(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("context is missing"))
					Expect(destroyed).To(BeFalse())
				})

				It("returns an error when the id is missing", func() {
					id = ""
					destroyed, err := session.Destroy(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("id is missing"))
					Expect(destroyed).To(BeFalse())
				})

				It("returns an error when the id is invalid", func() {
					id = "invalid"
					destroyed, err := session.Destroy(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("id is invalid"))
					Expect(destroyed).To(BeFalse())
				})

				It("returns an error when the condition is invalid", func() {
					condition.Revision = pointer.FromInt(-1)
					destroyed, err := session.Destroy(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("condition is invalid"))
					Expect(destroyed).To(BeFalse())
				})

				It("returns an error when the session is closed", func() {
					session.Close()
					destroyed, err := session.Destroy(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("session closed"))
					Expect(destroyed).To(BeFalse())
				})

				Context("with data", func() {
					var original *user.User

					BeforeEach(func() {
						original = userTest.RandomUser()
						original.UserID = pointer.FromString(id)
					})

					JustBeforeEach(func() {
						Expect(mgoCollection.Insert(original)).To(Succeed())
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
							Expect(session.Destroy(ctx, id, condition)).To(BeFalse())
							Expect(mgoCollection.Find(bson.M{"userid": original.UserID}).Count()).To(Equal(1))
						})

						It("returns false and does not destroy the original when the id exists, but the condition revision does not match", func() {
							condition.Revision = pointer.FromInt(*original.Revision + 1)
							Expect(session.Destroy(ctx, id, condition)).To(BeFalse())
							Expect(mgoCollection.Find(bson.M{"userid": original.UserID}).Count()).To(Equal(1))
						})

						It("returns true and destroys the original when the id exists and the condition is missing", func() {
							condition = nil
							Expect(session.Destroy(ctx, id, condition)).To(BeTrue())
							Expect(mgoCollection.Find(bson.M{"userid": original.UserID}).Count()).To(Equal(0))
						})

						It("returns true and destroys the original when the id exists and the condition revision is missing", func() {
							condition.Revision = nil
							Expect(session.Destroy(ctx, id, condition)).To(BeTrue())
							Expect(mgoCollection.Find(bson.M{"userid": original.UserID}).Count()).To(Equal(0))
						})

						It("returns true and destroys the original when the id exists and the condition revision matches", func() {
							condition.Revision = pointer.CloneInt(original.Revision)
							Expect(session.Destroy(ctx, id, condition)).To(BeTrue())
							Expect(mgoCollection.Find(bson.M{"userid": original.UserID}).Count()).To(Equal(0))
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
