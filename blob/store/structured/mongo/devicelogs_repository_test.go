package mongo_test

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"github.com/tidepool-org/platform/blob"
	blobStoreStructured "github.com/tidepool-org/platform/blob/store/structured"
	blobStoreStructuredMongo "github.com/tidepool-org/platform/blob/store/structured/mongo"
	blobStoreStructuredTest "github.com/tidepool-org/platform/blob/store/structured/test"
	blobTest "github.com/tidepool-org/platform/blob/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	requestTest "github.com/tidepool-org/platform/request/test"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("Mongo", func() {
	var logger *logTest.Logger
	var store *blobStoreStructuredMongo.Store

	BeforeEach(func() {
		logger = logTest.NewLogger()
	})

	Context("with a new store", func() {
		var deviceLogsCollection *mongo.Collection

		BeforeEach(func() {
			store = GetSuiteStore()
			deviceLogsCollection = store.GetCollection("deviceLogs")
		})

		AfterEach(func() {
			if deviceLogsCollection != nil {
				deviceLogsCollection.DeleteMany(context.Background(), bson.D{})
			}
		})

		Context("NewDeviceLogsRepository", func() {
			var deviceLogsRepository blobStoreStructured.DeviceLogsRepository

			It("returns a new session", func() {
				deviceLogsRepository = store.NewDeviceLogsRepository()
				Expect(deviceLogsRepository).ToNot(BeNil())
			})
		})

		Context("NewDeviceLogsRepository with a new session", func() {
			var ctx context.Context
			var deviceLogsRepository blobStoreStructured.DeviceLogsRepository

			BeforeEach(func() {
				Expect(store.EnsureIndexes()).To(Succeed())
				deviceLogsRepository = store.NewDeviceLogsRepository()
				ctx = log.NewContextWithLogger(context.Background(), logger)
			})

			Context("with user id", func() {
				var userID string

				BeforeEach(func() {
					userID = userTest.RandomID()
				})

				Context("Create", func() {
					var create *blobStoreStructured.Create

					BeforeEach(func() {
						create = blobStoreStructuredTest.RandomCreate()
					})

					It("returns an error when the context is missing", func() {
						ctx = nil
						result, err := deviceLogsRepository.Create(ctx, userID, create)
						errorsTest.ExpectEqual(err, errors.New("context is missing"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the user id is missing", func() {
						userID = ""
						result, err := deviceLogsRepository.Create(ctx, userID, create)
						errorsTest.ExpectEqual(err, errors.New("user id is missing"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the user id is invalid", func() {
						userID = "invalid"
						result, err := deviceLogsRepository.Create(ctx, userID, create)
						errorsTest.ExpectEqual(err, errors.New("user id is invalid"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the create is missing", func() {
						create = nil
						result, err := deviceLogsRepository.Create(ctx, userID, create)
						errorsTest.ExpectEqual(err, errors.New("create is missing"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the create is invalid", func() {
						create.MediaType = pointer.FromString("")
						result, err := deviceLogsRepository.Create(ctx, userID, create)
						errorsTest.ExpectEqual(err, errors.New("create is invalid"))
						Expect(result).To(BeNil())
					})

					It("returns the result after creating", func() {
						matchAllFields := MatchAllFields(Fields{
							"ID":          PointTo(Not(BeEmpty())),
							"UserID":      PointTo(Equal(userID)),
							"DigestMD5":   BeNil(),
							"MediaType":   Equal(create.MediaType),
							"Size":        BeNil(),
							"CreatedTime": PointTo(BeTemporally("~", time.Now(), time.Second)),
							"StartAtTime": BeNil(),
							"EndAtTime":   BeNil(),
							"Revision":    PointTo(Equal(0)),
						})
						result, err := deviceLogsRepository.Create(ctx, userID, create)
						Expect(err).ToNot(HaveOccurred())
						Expect(result).ToNot(BeNil())
						Expect(*result).To(matchAllFields)
						storeResult := blob.DeviceLogsBlobArray{}
						cursor, err := deviceLogsCollection.Find(context.Background(), bson.M{"id": result.ID})
						Expect(err).ToNot(HaveOccurred())
						Expect(cursor).ToNot(BeNil())
						Expect(cursor.All(context.Background(), &storeResult)).To(Succeed())
						Expect(storeResult).To(HaveLen(1))
						Expect(*storeResult[0]).To(matchAllFields)
						logger.AssertDebug("Create", log.Fields{"userId": userID, "create": create, "id": *storeResult[0].ID})
					})

					It("returns the result after creating without media type", func() {
						create.MediaType = nil
						matchAllFields := MatchAllFields(Fields{
							"ID":          PointTo(Not(BeEmpty())),
							"UserID":      PointTo(Equal(userID)),
							"DigestMD5":   BeNil(),
							"MediaType":   BeNil(),
							"Size":        BeNil(),
							"CreatedTime": PointTo(BeTemporally("~", time.Now(), time.Second)),
							"StartAtTime": BeNil(),
							"EndAtTime":   BeNil(),
							"Revision":    PointTo(Equal(0)),
						})
						result, err := deviceLogsRepository.Create(ctx, userID, create)
						Expect(err).ToNot(HaveOccurred())
						Expect(result).ToNot(BeNil())
						Expect(*result).To(matchAllFields)
						storeResult := blob.DeviceLogsBlobArray{}
						cursor, err := deviceLogsCollection.Find(context.Background(), bson.M{"id": result.ID})
						Expect(err).ToNot(HaveOccurred())
						Expect(cursor).ToNot(BeNil())
						Expect(cursor.All(context.Background(), &storeResult)).To(Succeed())
						Expect(storeResult).To(HaveLen(1))
						Expect(*storeResult[0]).To(matchAllFields)
						logger.AssertDebug("Create", log.Fields{"userId": userID, "create": create, "id": *storeResult[0].ID})
					})
				})
			})

			Context("Update", func() {
				var id string
				var condition *request.Condition
				var update *blobStoreStructured.DeviceLogsUpdate

				BeforeEach(func() {
					id = blobTest.RandomID()
					condition = requestTest.RandomCondition()
					update = blobStoreStructuredTest.RandomDeviceLogsUpdate()
				})

				It("returns an error when the context is missing", func() {
					ctx = nil
					result, err := deviceLogsRepository.Update(ctx, id, condition, update)
					errorsTest.ExpectEqual(err, errors.New("context is missing"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the id is missing", func() {
					id = ""
					result, err := deviceLogsRepository.Update(ctx, id, condition, update)
					errorsTest.ExpectEqual(err, errors.New("id is missing"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the id is invalid", func() {
					id = "invalid"
					result, err := deviceLogsRepository.Update(ctx, id, condition, update)
					errorsTest.ExpectEqual(err, errors.New("id is invalid"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the condition is invalid", func() {
					condition.Revision = pointer.FromInt(-1)
					result, err := deviceLogsRepository.Update(ctx, id, condition, update)
					errorsTest.ExpectEqual(err, errors.New("condition is invalid"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the update is missing", func() {
					update = nil
					result, err := deviceLogsRepository.Update(ctx, id, condition, update)
					errorsTest.ExpectEqual(err, errors.New("update is missing"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the update is invalid", func() {
					update.StartAt = nil
					result, err := deviceLogsRepository.Update(ctx, id, condition, update)
					errorsTest.ExpectEqual(err, errors.New("update is invalid"))
					Expect(result).To(BeNil())
				})

				Context("with data", func() {
					var original *blob.DeviceLogsBlob

					BeforeEach(func() {
						update = blobStoreStructuredTest.RandomDeviceLogsUpdate()
						original = blobTest.RandomDeviceLogsBlob()
						original.ID = pointer.FromString(id)
						_, err := deviceLogsCollection.InsertOne(context.Background(), original)
						Expect(err).ToNot(HaveOccurred())
					})

					AfterEach(func() {
						if condition != nil {
							logger.AssertDebug("Update", log.Fields{"id": id, "condition": condition, "update": update})
						} else {
							logger.AssertDebug("Update", log.Fields{"id": id, "update": update})
						}
					})

					conditionAssertions := func() {
						Context("with updates", func() {
							It("returns updated result when the id exists", func() {
								matchAllFields := MatchAllFields(Fields{
									"ID":          PointTo(Equal(id)),
									"UserID":      Equal(original.UserID),
									"DigestMD5":   Equal(update.DigestMD5),
									"MediaType":   Equal(update.MediaType),
									"Size":        Equal(update.Size),
									"StartAtTime": Equal(update.StartAt),
									"EndAtTime":   Equal(update.EndAt),
									"CreatedTime": Equal(original.CreatedTime),
									"Revision":    PointTo(Equal(*original.Revision + 1)),
								})
								result, err := deviceLogsRepository.Update(ctx, id, condition, update)
								Expect(err).ToNot(HaveOccurred())
								Expect(result).ToNot(BeNil())
								Expect(*result).To(matchAllFields)
								storeResult := blob.DeviceLogsBlobArray{}
								cursor, err := deviceLogsCollection.Find(context.Background(), bson.M{"id": id})
								Expect(err).ToNot(HaveOccurred())
								Expect(cursor).ToNot(BeNil())
								Expect(cursor.All(context.Background(), &storeResult)).To(Succeed())
								Expect(storeResult).To(HaveLen(1))
								Expect(*storeResult[0]).To(matchAllFields)
							})

							It("returns nil when the id does not exist", func() {
								id = blobTest.RandomID()
								Expect(deviceLogsRepository.Update(ctx, id, condition, update)).To(BeNil())
							})
						})

						Context("without updates", func() {
							BeforeEach(func() {
								update = blobStoreStructured.NewDeviceLogsUpdate()
							})

							It("returns original when the id exists", func() {
								Expect(deviceLogsRepository.Update(ctx, id, condition, update)).To(Equal(original))
							})

							It("returns nil when the id does not exist", func() {
								id = blobTest.RandomID()
								Expect(deviceLogsRepository.Update(ctx, id, condition, update)).To(BeNil())
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
					id = blobTest.RandomID()
					condition = requestTest.RandomCondition()
				})

				It("returns an error when the context is missing", func() {
					ctx = nil
					deleted, err := deviceLogsRepository.Destroy(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("context is missing"))
					Expect(deleted).To(BeFalse())
				})

				It("returns an error when the id is missing", func() {
					id = ""
					deleted, err := deviceLogsRepository.Destroy(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("id is missing"))
					Expect(deleted).To(BeFalse())
				})

				It("returns an error when the id is invalid", func() {
					id = "invalid"
					deleted, err := deviceLogsRepository.Destroy(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("id is invalid"))
					Expect(deleted).To(BeFalse())
				})

				It("returns an error when the condition is invalid", func() {
					condition.Revision = pointer.FromInt(-1)
					deleted, err := deviceLogsRepository.Destroy(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("condition is invalid"))
					Expect(deleted).To(BeFalse())
				})

				Context("with data", func() {
					var original *blob.DeviceLogsBlob

					BeforeEach(func() {
						original = blobTest.RandomDeviceLogsBlob()
						original.ID = pointer.FromString(id)
						_, err := deviceLogsCollection.InsertOne(context.Background(), original)
						Expect(err).ToNot(HaveOccurred())
					})

					AfterEach(func() {
						if condition != nil {
							logger.AssertDebug("Destroy", log.Fields{"id": id, "condition": condition})
						} else {
							logger.AssertDebug("Destroy", log.Fields{"id": id})
						}
					})

					It("returns false and does not delete the original when the id does not exist", func() {
						id = blobTest.RandomID()
						Expect(deviceLogsRepository.Destroy(ctx, id, condition)).To(BeFalse())
						Expect(deviceLogsCollection.CountDocuments(context.Background(), bson.M{"id": original.ID})).To(Equal(int64(1)))
					})

					It("returns false and does not delete the original when the id exists, but the condition revision does not match", func() {
						condition.Revision = pointer.FromInt(*original.Revision + 1)
						Expect(deviceLogsRepository.Destroy(ctx, id, condition)).To(BeFalse())
						Expect(deviceLogsCollection.CountDocuments(context.Background(), bson.M{"id": original.ID})).To(Equal(int64(1)))
					})

					It("returns true and deletes the original when the id exists and the condition is missing", func() {
						condition = nil
						Expect(deviceLogsRepository.Destroy(ctx, id, condition)).To(BeTrue())
						Expect(deviceLogsCollection.CountDocuments(context.Background(), bson.M{"id": original.ID})).To(Equal(int64(0)))
					})

					It("returns true and deletes the original when the id exists and the condition revision is missing", func() {
						condition.Revision = nil
						Expect(deviceLogsRepository.Destroy(ctx, id, condition)).To(BeTrue())
						Expect(deviceLogsCollection.CountDocuments(context.Background(), bson.M{"id": original.ID})).To(Equal(int64(0)))
					})

					It("returns true and deletes the original when the id exists and the condition revision matches", func() {
						condition.Revision = pointer.CloneInt(original.Revision)
						Expect(deviceLogsRepository.Destroy(ctx, id, condition)).To(BeTrue())
						Expect(deviceLogsCollection.CountDocuments(context.Background(), bson.M{"id": original.ID})).To(Equal(int64(0)))
					})
				})
			})
		})
	})
})
