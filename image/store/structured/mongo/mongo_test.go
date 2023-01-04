package mongo_test

import (
	"context"
	"math/rand"
	"sort"
	"time"

	gomegaTypes "github.com/onsi/gomega/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	associationTest "github.com/tidepool-org/platform/association/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/image"
	imageStoreStructured "github.com/tidepool-org/platform/image/store/structured"
	imageStoreStructuredMongo "github.com/tidepool-org/platform/image/store/structured/mongo"
	imageStoreStructuredTest "github.com/tidepool-org/platform/image/store/structured/test"
	imageTest "github.com/tidepool-org/platform/image/test"
	locationTest "github.com/tidepool-org/platform/location/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	metadataTest "github.com/tidepool-org/platform/metadata/test"
	originTest "github.com/tidepool-org/platform/origin/test"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	requestTest "github.com/tidepool-org/platform/request/test"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
	"github.com/tidepool-org/platform/test"
	userTest "github.com/tidepool-org/platform/user/test"
)

type CreatedTimeDescending image.ImageArray

func (c CreatedTimeDescending) Len() int {
	return len(c)
}

func (c CreatedTimeDescending) Less(left int, right int) bool {
	if c[left].CreatedTime == nil {
		return true
	} else if c[right].CreatedTime == nil {
		return false
	}
	return c[right].CreatedTime.Before(*c[left].CreatedTime)
}

func (c CreatedTimeDescending) Swap(left int, right int) {
	c[left], c[right] = c[right], c[left]
}

func SelectAndSort(imageArray image.ImageArray, selector func(i *image.Image) bool) image.ImageArray {
	var selected image.ImageArray
	for _, image := range imageArray {
		if selector(image) {
			selected = append(selected, image)
		}
	}
	sort.Sort(CreatedTimeDescending(selected))
	return selected
}

func AsInterfaceArray(imageArray image.ImageArray) []interface{} {
	if imageArray == nil {
		return nil
	}
	array := make([]interface{}, len(imageArray))
	for index, image := range imageArray {
		array[index] = image
	}
	return array
}

var _ = Describe("Mongo", func() {
	var config *storeStructuredMongo.Config
	var logger *logTest.Logger
	var store *imageStoreStructuredMongo.Store
	var repository imageStoreStructured.ImageRepository

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
			store, err = imageStoreStructuredMongo.NewStore(nil)
			errorsTest.ExpectEqual(err, errors.New("database config is empty"))
			Expect(store).To(BeNil())
		})

		It("returns a new store and no error when successful", func() {
			var err error
			store, err = imageStoreStructuredMongo.NewStore(config)
			Expect(err).ToNot(HaveOccurred())
			Expect(store).ToNot(BeNil())
		})
	})

	Context("with a new store", func() {
		var collection *mongo.Collection

		BeforeEach(func() {
			var err error
			store, err = imageStoreStructuredMongo.NewStore(config)
			Expect(err).ToNot(HaveOccurred())
			Expect(store).ToNot(BeNil())
			collection = store.GetCollection("images")
		})

		Context("EnsureIndexes", func() {
			It("returns successfully", func() {
				Expect(store.EnsureIndexes()).To(Succeed())
				cursor, err := collection.Indexes().List(context.Background())
				Expect(err).ToNot(HaveOccurred())
				Expect(cursor).ToNot(BeNil())
				var indexes []storeStructuredMongoTest.MongoIndex
				err = cursor.All(context.Background(), &indexes)
				Expect(err).ToNot(HaveOccurred())

				Expect(indexes).To(ConsistOf(
					MatchFields(IgnoreExtras, Fields{
						"Key": Equal(storeStructuredMongoTest.MakeKeySlice("_id")),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Key":        Equal(storeStructuredMongoTest.MakeKeySlice("id")),
						"Background": Equal(true),
						"Unique":     Equal(true),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Key":        Equal(storeStructuredMongoTest.MakeKeySlice("userId", "status")),
						"Background": Equal(true),
					}),
				))
			})
		})

		Context("NewImageRepository", func() {
			It("returns a new session", func() {
				repository = store.NewImageRepository()
				Expect(repository).ToNot(BeNil())
			})
		})

		Context("with a new session", func() {
			var ctx context.Context

			BeforeEach(func() {
				Expect(store.EnsureIndexes()).To(Succeed())
				repository = store.NewImageRepository()
				ctx = log.NewContextWithLogger(context.Background(), logger)
			})

			Context("with user id", func() {
				var userID string

				BeforeEach(func() {
					userID = userTest.RandomID()
				})

				Context("List", func() {
					var filter *image.Filter
					var pagination *page.Pagination

					BeforeEach(func() {
						filter = image.NewFilter()
						pagination = page.NewPagination()
					})

					It("returns an error when the context is missing", func() {
						ctx = nil
						result, err := repository.List(ctx, userID, filter, pagination)
						errorsTest.ExpectEqual(err, errors.New("context is missing"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the user id is missing", func() {
						userID = ""
						result, err := repository.List(ctx, userID, filter, pagination)
						errorsTest.ExpectEqual(err, errors.New("user id is missing"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the user id is invalid", func() {
						userID = "invalid"
						result, err := repository.List(ctx, userID, filter, pagination)
						errorsTest.ExpectEqual(err, errors.New("user id is invalid"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the filter is invalid", func() {
						filter.Status = pointer.FromStringArray([]string{""})
						result, err := repository.List(ctx, userID, filter, pagination)
						errorsTest.ExpectEqual(err, errors.New("filter is invalid"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the pagination is invalid", func() {
						pagination.Page = -1
						result, err := repository.List(ctx, userID, filter, pagination)
						errorsTest.ExpectEqual(err, errors.New("pagination is invalid"))
						Expect(result).To(BeNil())
					})

					Context("with data", func() {
						var allResult image.ImageArray

						BeforeEach(func() {
							allResult = imageTest.RandomImageArray(8, 8)
							for index, result := range allResult {
								result.ID = pointer.FromString(imageTest.RandomID())
								result.UserID = pointer.FromString(userID)
								result.Status = pointer.FromString(image.Statuses()[(index/2)%2])
								switch *result.Status {
								case image.StatusAvailable:
									result.ContentIntent = pointer.FromString(image.ContentIntents()[(index/4)%2])
									result.ContentAttributes = imageTest.RandomContentAttributes()
								case image.StatusCreated:
									result.ContentIntent = nil
									result.ContentAttributes = nil
								}
								if index%2 == 0 {
									result.ModifiedTime = pointer.FromTime(test.RandomTimeFromRange(*result.CreatedTime, time.Now()).Truncate(time.Second))
									result.DeletedTime = pointer.CloneTime(result.ModifiedTime)
								}
							}
							allResult = append(allResult, imageTest.RandomImage(), imageTest.RandomImage())
							rand.Shuffle(len(allResult), func(i, j int) { allResult[i], allResult[j] = allResult[j], allResult[i] })
							_, err := collection.InsertMany(ctx, AsInterfaceArray(allResult))
							Expect(err).ToNot(HaveOccurred())
						})

						It("returns no result when the user id is unknown", func() {
							userID = userTest.RandomID()
							Expect(repository.List(ctx, userID, filter, pagination)).To(SatisfyAll(Not(BeNil()), BeEmpty()))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 0})
						})

						It("returns expected result when the filter is missing", func() {
							filter = nil
							Expect(repository.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(i *image.Image) bool {
									return *i.UserID == userID && i.DeletedTime == nil && *i.Status == image.StatusAvailable
								},
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "pagination": pagination, "count": 2})
						})

						It("returns expected result when the filter status is missing", func() {
							filter.Status = nil
							Expect(repository.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(i *image.Image) bool {
									return *i.UserID == userID && i.DeletedTime == nil && *i.Status == image.StatusAvailable
								},
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 2})
						})

						It("returns expected result when the filter status is set to available", func() {
							filter.Status = pointer.FromStringArray([]string{image.StatusAvailable})
							Expect(repository.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(i *image.Image) bool {
									return *i.UserID == userID && i.DeletedTime == nil && *i.Status == image.StatusAvailable
								},
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 2})
						})

						It("returns expected result when the filter status is set to created", func() {
							filter.Status = pointer.FromStringArray([]string{image.StatusCreated})
							Expect(repository.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(i *image.Image) bool {
									return *i.UserID == userID && i.DeletedTime == nil && *i.Status == image.StatusCreated
								},
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 2})
						})

						It("returns expected result when the filter status is set to both available and original", func() {
							filter.Status = pointer.FromStringArray(image.Statuses())
							Expect(repository.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(i *image.Image) bool { return *i.UserID == userID && i.DeletedTime == nil },
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 4})
						})

						It("returns expected result when the filter content intent is missing", func() {
							filter.ContentIntent = nil
							Expect(repository.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(i *image.Image) bool {
									return *i.UserID == userID && i.DeletedTime == nil && *i.Status == image.StatusAvailable
								},
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 2})
						})

						It("returns expected result when the filter content intent is set to alternate", func() {
							filter.ContentIntent = pointer.FromStringArray([]string{image.ContentIntentAlternate})
							Expect(repository.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(i *image.Image) bool {
									return *i.UserID == userID && i.DeletedTime == nil && *i.Status == image.StatusAvailable && *i.ContentIntent == image.ContentIntentAlternate
								},
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 1})
						})

						It("returns expected result when the filter content intent is set to original", func() {
							filter.ContentIntent = pointer.FromStringArray([]string{image.ContentIntentOriginal})
							Expect(repository.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(i *image.Image) bool {
									return *i.UserID == userID && i.DeletedTime == nil && *i.Status == image.StatusAvailable && *i.ContentIntent == image.ContentIntentOriginal
								},
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 1})
						})

						It("returns expected result when the filter content intent is set to both alternate and original", func() {
							filter.ContentIntent = pointer.FromStringArray(image.ContentIntents())
							Expect(repository.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(i *image.Image) bool {
									return *i.UserID == userID && i.DeletedTime == nil && *i.Status == image.StatusAvailable
								},
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 2})
						})

						It("returns expected result when the filter status is set to both available and created and filter content intent is set to alternate", func() {
							filter.Status = pointer.FromStringArray(image.Statuses())
							filter.ContentIntent = pointer.FromStringArray([]string{image.ContentIntentAlternate})
							Expect(repository.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(i *image.Image) bool {
									return *i.UserID == userID && i.DeletedTime == nil && *i.Status == image.StatusAvailable && *i.ContentIntent == image.ContentIntentAlternate
								},
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 1})
						})

						It("returns expected result when the filter status is set to both available and created and filter content intent is set to original", func() {
							filter.Status = pointer.FromStringArray(image.Statuses())
							filter.ContentIntent = pointer.FromStringArray([]string{image.ContentIntentOriginal})
							Expect(repository.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(i *image.Image) bool {
									return *i.UserID == userID && i.DeletedTime == nil && *i.Status == image.StatusAvailable && *i.ContentIntent == image.ContentIntentOriginal
								},
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 1})
						})

						It("returns expected result when the filter status is set to both available and created and filter content intent is set to alternate and original", func() {
							filter.Status = pointer.FromStringArray(image.Statuses())
							filter.ContentIntent = pointer.FromStringArray(image.ContentIntents())
							Expect(repository.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(i *image.Image) bool {
									return *i.UserID == userID && i.DeletedTime == nil && *i.Status == image.StatusAvailable
								},
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 2})
						})

						It("returns expected result when the pagination is missing", func() {
							filter.Status = pointer.FromStringArray(image.Statuses())
							pagination = nil
							Expect(repository.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(i *image.Image) bool { return *i.UserID == userID && i.DeletedTime == nil },
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "count": 4})
						})

						It("returns expected result when the pagination limits result", func() {
							filter.Status = pointer.FromStringArray(image.Statuses())
							pagination.Page = 1
							pagination.Size = 2
							Expect(repository.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(i *image.Image) bool { return *i.UserID == userID && i.DeletedTime == nil },
							)[2:4]))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 2})
						})
					})
				})

				Context("Create", func() {
					var metadata *image.Metadata

					BeforeEach(func() {
						metadata = imageTest.RandomMetadata()
					})

					It("returns an error when the context is missing", func() {
						ctx = nil
						result, err := repository.Create(ctx, userID, metadata)
						errorsTest.ExpectEqual(err, errors.New("context is missing"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the user id is missing", func() {
						userID = ""
						result, err := repository.Create(ctx, userID, metadata)
						errorsTest.ExpectEqual(err, errors.New("user id is missing"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the user id is invalid", func() {
						userID = "invalid"
						result, err := repository.Create(ctx, userID, metadata)
						errorsTest.ExpectEqual(err, errors.New("user id is invalid"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the metadata is missing", func() {
						metadata = nil
						result, err := repository.Create(ctx, userID, metadata)
						errorsTest.ExpectEqual(err, errors.New("metadata is missing"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the metadata is invalid", func() {
						metadata.Name = pointer.FromString("")
						result, err := repository.Create(ctx, userID, metadata)
						errorsTest.ExpectEqual(err, errors.New("metadata is invalid"))
						Expect(result).To(BeNil())
					})

					It("returns the result after creating", func() {
						matchAllFields := MatchAllFields(Fields{
							"ID":                PointTo(Not(BeEmpty())),
							"UserID":            PointTo(Equal(userID)),
							"Status":            PointTo(Equal(image.StatusCreated)),
							"Metadata":          Equal(metadata),
							"ContentID":         BeNil(),
							"ContentIntent":     BeNil(),
							"ContentAttributes": BeNil(),
							"RenditionsID":      BeNil(),
							"Renditions":        BeNil(),
							"CreatedTime":       PointTo(BeTemporally("~", time.Now(), time.Second)),
							"ModifiedTime":      BeNil(),
							"DeletedTime":       BeNil(),
							"Revision":          PointTo(Equal(0)),
						})
						result, err := repository.Create(ctx, userID, metadata)
						Expect(err).ToNot(HaveOccurred())
						Expect(result).ToNot(BeNil())
						Expect(*result).To(matchAllFields)
						storeResult := image.ImageArray{}
						cursor, err := collection.Find(context.Background(), bson.M{"id": result.ID})
						Expect(err).ToNot(HaveOccurred())
						Expect(cursor).ToNot(BeNil())
						Expect(cursor.All(context.Background(), &storeResult)).To(Succeed())
						Expect(storeResult).To(HaveLen(1))
						Expect(*storeResult[0]).To(matchAllFields)
						logger.AssertDebug("Create", log.Fields{"userId": userID, "metadata": metadata, "id": *storeResult[0].ID})
					})
				})

				Context("DeleteAll", func() {
					It("returns an error when the context is missing", func() {
						ctx = nil
						deleted, err := repository.DeleteAll(ctx, userID)
						errorsTest.ExpectEqual(err, errors.New("context is missing"))
						Expect(deleted).To(BeFalse())
					})

					It("returns an error when the user id is missing", func() {
						userID = ""
						deleted, err := repository.DeleteAll(ctx, userID)
						errorsTest.ExpectEqual(err, errors.New("user id is missing"))
						Expect(deleted).To(BeFalse())
					})

					It("returns an error when the user id is invalid", func() {
						userID = "invalid"
						deleted, err := repository.DeleteAll(ctx, userID)
						errorsTest.ExpectEqual(err, errors.New("user id is invalid"))
						Expect(deleted).To(BeFalse())
					})

					Context("with data", func() {
						var originals image.ImageArray

						BeforeEach(func() {
							originals = imageTest.RandomImageArray(4, 4)
							for index, original := range originals {
								original.UserID = pointer.FromString(userID)
								if index%2 == 0 {
									original.ModifiedTime = pointer.FromTime(test.RandomTimeFromRange(*original.CreatedTime, time.Now()).Truncate(time.Second))
									original.DeletedTime = pointer.CloneTime(original.ModifiedTime)
								}
								_, err := collection.InsertOne(ctx, original)
								Expect(err).ToNot(HaveOccurred())
							}
							_, err := collection.InsertMany(ctx, []interface{}{imageTest.RandomImage(), imageTest.RandomImage()})
							Expect(err).ToNot(HaveOccurred())
						})

						AfterEach(func() {
							logger.AssertDebug("DeleteAll", log.Fields{"userId": userID})
						})

						It("returns false and does not delete the originals when the user id does not match", func() {
							originalUserID := userID
							userID = userTest.RandomID()
							Expect(repository.DeleteAll(ctx, userID)).To(BeFalse())
							Expect(collection.CountDocuments(ctx, bson.M{"userId": originalUserID, "deletedTime": bson.M{"$exists": true}})).To(Equal(int64(2)))
							Expect(collection.CountDocuments(ctx, bson.M{"deletedTime": bson.M{"$exists": true}})).To(Equal(int64(2)))
						})

						It("returns true and deletes the originals when the user id matches", func() {
							Expect(repository.DeleteAll(ctx, userID)).To(BeTrue())
							Expect(collection.CountDocuments(ctx, bson.M{"userId": userID, "deletedTime": bson.M{"$exists": true}})).To(Equal(int64(4)))
							Expect(collection.CountDocuments(ctx, bson.M{"deletedTime": bson.M{"$exists": true}})).To(Equal(int64(4)))
						})
					})
				})

				Context("DestroyAll", func() {
					It("returns an error when the context is missing", func() {
						ctx = nil
						destroyed, err := repository.DestroyAll(ctx, userID)
						errorsTest.ExpectEqual(err, errors.New("context is missing"))
						Expect(destroyed).To(BeFalse())
					})

					It("returns an error when the user id is missing", func() {
						userID = ""
						destroyed, err := repository.DestroyAll(ctx, userID)
						errorsTest.ExpectEqual(err, errors.New("user id is missing"))
						Expect(destroyed).To(BeFalse())
					})

					It("returns an error when the user id is invalid", func() {
						userID = "invalid"
						destroyed, err := repository.DestroyAll(ctx, userID)
						errorsTest.ExpectEqual(err, errors.New("user id is invalid"))
						Expect(destroyed).To(BeFalse())
					})

					Context("with data", func() {
						var originals image.ImageArray

						BeforeEach(func() {
							originals = imageTest.RandomImageArray(4, 4)
							for index, original := range originals {
								original.UserID = pointer.FromString(userID)
								if index%2 == 0 {
									original.ModifiedTime = pointer.FromTime(test.RandomTimeFromRange(*original.CreatedTime, time.Now()).Truncate(time.Second))
									original.DeletedTime = pointer.CloneTime(original.ModifiedTime)
								}
								_, err := collection.InsertOne(ctx, original)
								Expect(err).ToNot(HaveOccurred())
							}
							_, err := collection.InsertMany(ctx, []interface{}{imageTest.RandomImage(), imageTest.RandomImage()})
							Expect(err).ToNot(HaveOccurred())
						})

						AfterEach(func() {
							logger.AssertDebug("DestroyAll", log.Fields{"userId": userID})
						})

						It("returns false and does not destroy the originals when the user id does not match", func() {
							originalUserID := userID
							userID = userTest.RandomID()
							Expect(repository.DestroyAll(ctx, userID)).To(BeFalse())
							Expect(collection.CountDocuments(ctx, bson.M{"userId": originalUserID})).To(Equal(int64(4)))
							Expect(collection.CountDocuments(ctx, bson.M{})).To(Equal(int64(6)))
						})

						It("returns true and destroys the originals when the user id matches", func() {
							Expect(repository.DestroyAll(ctx, userID)).To(BeTrue())
							Expect(collection.CountDocuments(ctx, bson.M{"userId": userID})).To(Equal(int64(0)))
							Expect(collection.CountDocuments(ctx, bson.M{})).To(Equal(int64(2)))
						})
					})
				})
			})

			Context("Get", func() {
				var id string
				var condition *request.Condition

				BeforeEach(func() {
					id = imageTest.RandomID()
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
					var allResult image.ImageArray
					var result *image.Image

					BeforeEach(func() {
						allResult = imageTest.RandomImageArray(3, 3)
						result = allResult[0]
						result.ID = pointer.FromString(id)
						rand.Shuffle(len(allResult), func(i, j int) { allResult[i], allResult[j] = allResult[j], allResult[i] })
					})

					JustBeforeEach(func() {
						_, err := collection.InsertMany(ctx, AsInterfaceArray(allResult))
						Expect(err).ToNot(HaveOccurred())
					})

					AfterEach(func() {
						logger.AssertDebug("Get", log.Fields{"id": id})
					})

					It("returns nil when the id does not exist", func() {
						id = imageTest.RandomID()
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

			Context("Update", func() {
				var id string
				var condition *request.Condition
				var update *imageStoreStructured.Update

				BeforeEach(func() {
					id = imageTest.RandomID()
					condition = requestTest.RandomCondition()
					update = imageStoreStructuredTest.RandomUpdate()
				})

				It("returns an error when the context is missing", func() {
					ctx = nil
					result, err := repository.Update(ctx, id, condition, update)
					errorsTest.ExpectEqual(err, errors.New("context is missing"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the id is missing", func() {
					id = ""
					result, err := repository.Update(ctx, id, condition, update)
					errorsTest.ExpectEqual(err, errors.New("id is missing"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the id is invalid", func() {
					id = "invalid"
					result, err := repository.Update(ctx, id, condition, update)
					errorsTest.ExpectEqual(err, errors.New("id is invalid"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the condition is invalid", func() {
					condition.Revision = pointer.FromInt(-1)
					result, err := repository.Update(ctx, id, condition, update)
					errorsTest.ExpectEqual(err, errors.New("condition is invalid"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the update is missing", func() {
					update = nil
					result, err := repository.Update(ctx, id, condition, update)
					errorsTest.ExpectEqual(err, errors.New("update is missing"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the update is invalid", func() {
					update.ContentIntent = pointer.FromString("")
					result, err := repository.Update(ctx, id, condition, update)
					errorsTest.ExpectEqual(err, errors.New("update is invalid"))
					Expect(result).To(BeNil())
				})

				Context("with data", func() {
					var original *image.Image

					BeforeEach(func() {
						original = imageTest.RandomImage()
						original.ID = pointer.FromString(id)
					})

					JustBeforeEach(func() {
						_, err := collection.InsertOne(ctx, original)
						Expect(err).ToNot(HaveOccurred())
					})

					AfterEach(func() {
						if condition != nil {
							logger.AssertDebug("Update", log.Fields{"id": id, "condition": condition, "update": update})
						} else {
							logger.AssertDebug("Update", log.Fields{"id": id, "update": update})
						}
					})

					When("the original is marked as deleted", func() {
						BeforeEach(func() {
							original.ModifiedTime = pointer.FromTime(test.RandomTimeFromRange(*original.CreatedTime, time.Now()).Truncate(time.Second))
							original.DeletedTime = pointer.CloneTime(original.ModifiedTime)
						})

						It("returns nil", func() {
							Expect(repository.Update(ctx, id, condition, update)).To(BeNil())
						})
					})

					When("the condition revision does not match", func() {
						BeforeEach(func() {
							condition.Revision = pointer.FromInt(*original.Revision + 1)
						})

						It("returns nil", func() {
							Expect(repository.Update(ctx, id, condition, update)).To(BeNil())
						})
					})

					conditionAssertions := func() {
						Context("with updates", func() {
							Context("with metadata", func() {
								var resultMetadata *image.Metadata

								BeforeEach(func() {
									update = imageStoreStructured.NewUpdate()
									update.Metadata = image.NewMetadata()
									resultMetadata = imageTest.CloneMetadata(original.Metadata)
								})

								metadataAssertions := func() {
									It("returns updated result when the id exists", func() {
										matchAllFields := MatchAllFields(Fields{
											"ID":                PointTo(Equal(id)),
											"UserID":            Equal(original.UserID),
											"Status":            Equal(original.Status),
											"Metadata":          Equal(resultMetadata),
											"ContentID":         Equal(original.ContentID),
											"ContentIntent":     Equal(original.ContentIntent),
											"ContentAttributes": Equal(original.ContentAttributes),
											"RenditionsID":      Equal(original.RenditionsID),
											"Renditions":        Equal(original.Renditions),
											"CreatedTime":       Equal(original.CreatedTime),
											"ModifiedTime":      PointTo(BeTemporally("~", time.Now(), time.Second)),
											"DeletedTime":       Equal(original.DeletedTime),
											"Revision":          PointTo(Equal(*original.Revision + 1)),
										})
										result, err := repository.Update(ctx, id, condition, update)
										Expect(err).ToNot(HaveOccurred())
										Expect(result).ToNot(BeNil())
										Expect(*result).To(matchAllFields)
										storeResult := image.ImageArray{}
										cursor, err := collection.Find(context.Background(), bson.M{"id": id})
										Expect(err).ToNot(HaveOccurred())
										Expect(cursor).ToNot(BeNil())
										Expect(cursor.All(context.Background(), &storeResult)).To(Succeed())
										Expect(storeResult).To(HaveLen(1))
										Expect(*storeResult[0]).To(matchAllFields)
									})
								}

								When("only associations is specified", func() {
									BeforeEach(func() {
										update.Metadata.Associations = associationTest.RandomAssociationArray()
										resultMetadata.Associations = update.Metadata.Associations
									})

									metadataAssertions()
								})

								When("only location is specified", func() {
									BeforeEach(func() {
										update.Metadata.Location = locationTest.RandomLocation()
										resultMetadata.Location = update.Metadata.Location
									})

									metadataAssertions()
								})

								When("only metadata is specified", func() {
									BeforeEach(func() {
										update.Metadata.Metadata = metadataTest.RandomMetadata()
										resultMetadata.Metadata = update.Metadata.Metadata
									})

									metadataAssertions()
								})

								When("only name is specified", func() {
									BeforeEach(func() {
										update.Metadata.Name = pointer.FromString(imageTest.RandomName())
										resultMetadata.Name = update.Metadata.Name
									})

									metadataAssertions()
								})

								When("only origin is specified", func() {
									BeforeEach(func() {
										update.Metadata.Origin = originTest.RandomOrigin()
										resultMetadata.Origin = update.Metadata.Origin
									})

									metadataAssertions()
								})

								When("all are specified", func() {
									BeforeEach(func() {
										update.Metadata = imageTest.RandomMetadata()
										resultMetadata = update.Metadata
									})

									metadataAssertions()
								})
							})

							Context("with content id, content intent, and content attributes", func() {
								BeforeEach(func() {
									update = imageStoreStructured.NewUpdate()
									update.ContentID = pointer.FromString(imageTest.RandomContentID())
									update.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
									update.ContentAttributes = imageStoreStructuredTest.RandomContentAttributes()
								})

								It("returns updated result when the id exists", func() {
									var contentAttributesCreatedTimeMatcher gomegaTypes.GomegaMatcher
									if original.ContentAttributes != nil {
										contentAttributesCreatedTimeMatcher = Equal(original.ContentAttributes.CreatedTime)
									} else {
										contentAttributesCreatedTimeMatcher = PointTo(BeTemporally("~", time.Now(), time.Second))
									}
									matchAllFields := MatchAllFields(Fields{
										"ID":            PointTo(Equal(id)),
										"UserID":        Equal(original.UserID),
										"Status":        PointTo(Equal(image.StatusAvailable)),
										"Metadata":      Equal(original.Metadata),
										"ContentID":     Equal(update.ContentID),
										"ContentIntent": Equal(update.ContentIntent),
										"ContentAttributes": PointTo(MatchAllFields(Fields{
											"DigestMD5":    Equal(update.ContentAttributes.DigestMD5),
											"MediaType":    Equal(update.ContentAttributes.MediaType),
											"Width":        Equal(update.ContentAttributes.Width),
											"Height":       Equal(update.ContentAttributes.Height),
											"Size":         Equal(update.ContentAttributes.Size),
											"CreatedTime":  contentAttributesCreatedTimeMatcher,
											"ModifiedTime": PointTo(BeTemporally("~", time.Now(), time.Second)),
										})),
										"RenditionsID": BeNil(),
										"Renditions":   BeNil(),
										"CreatedTime":  Equal(original.CreatedTime),
										"ModifiedTime": PointTo(BeTemporally("~", time.Now(), time.Second)),
										"DeletedTime":  Equal(original.DeletedTime),
										"Revision":     PointTo(Equal(*original.Revision + 1)),
									})
									result, err := repository.Update(ctx, id, condition, update)
									Expect(err).ToNot(HaveOccurred())
									Expect(result).ToNot(BeNil())
									Expect(*result).To(matchAllFields)
									storeResult := image.ImageArray{}
									cursor, err := collection.Find(context.Background(), bson.M{"id": id})
									Expect(err).ToNot(HaveOccurred())
									Expect(cursor).ToNot(BeNil())
									Expect(cursor.All(context.Background(), &storeResult)).To(Succeed())
									Expect(storeResult).To(HaveLen(1))
									Expect(*storeResult[0]).To(matchAllFields)
								})

								It("returns nil when the id does not exist", func() {
									id = imageTest.RandomID()
									Expect(repository.Update(ctx, id, condition, update)).To(BeNil())
								})
							})

							Context("with renditions id and rendition", func() {
								var renditionsID string
								var renditionString string

								BeforeEach(func() {
									renditionsID = imageTest.RandomRenditionsID()
									renditionString = imageTest.RandomRenditionString()
									original.Status = pointer.FromString(image.StatusAvailable)
									original.ContentID = pointer.FromString(imageTest.RandomContentID())
									original.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
									original.ContentAttributes = imageTest.RandomContentAttributes()
									original.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
									original.Renditions = pointer.FromStringArray(imageTest.RandomRenditionStrings())
									update = imageStoreStructured.NewUpdate()
									update.RenditionsID = pointer.FromString(renditionsID)
									update.Rendition = pointer.FromString(renditionString)
								})

								It("returns updated result when the id exists", func() {
									matchAllFields := MatchAllFields(Fields{
										"ID":                PointTo(Equal(id)),
										"UserID":            Equal(original.UserID),
										"Status":            Equal(original.Status),
										"Metadata":          Equal(original.Metadata),
										"ContentID":         Equal(original.ContentID),
										"ContentIntent":     Equal(original.ContentIntent),
										"ContentAttributes": Equal(original.ContentAttributes),
										"RenditionsID":      Equal(update.RenditionsID),
										"Renditions":        PointTo(Equal([]string{renditionString})),
										"CreatedTime":       Equal(original.CreatedTime),
										"ModifiedTime":      PointTo(BeTemporally("~", time.Now(), time.Second)),
										"DeletedTime":       Equal(original.DeletedTime),
										"Revision":          PointTo(Equal(*original.Revision + 1)),
									})
									result, err := repository.Update(ctx, id, condition, update)
									Expect(err).ToNot(HaveOccurred())
									Expect(result).ToNot(BeNil())
									Expect(*result).To(matchAllFields)
									storeResult := image.ImageArray{}
									cursor, err := collection.Find(context.Background(), bson.M{"id": id})
									Expect(err).ToNot(HaveOccurred())
									Expect(cursor).ToNot(BeNil())
									Expect(cursor.All(context.Background(), &storeResult)).To(Succeed())
									Expect(storeResult).To(HaveLen(1))
									Expect(*storeResult[0]).To(matchAllFields)
								})

								It("returns nil when the id does not exist", func() {
									id = imageTest.RandomID()
									Expect(repository.Update(ctx, id, condition, update)).To(BeNil())
								})
							})

							Context("with rendition", func() {
								var renditionString string

								BeforeEach(func() {
									renditionString = imageTest.RandomRenditionString()
									original.Status = pointer.FromString(image.StatusAvailable)
									original.ContentID = pointer.FromString(imageTest.RandomContentID())
									original.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
									original.ContentAttributes = imageTest.RandomContentAttributes()
									original.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
									original.Renditions = pointer.FromStringArray(imageTest.RandomRenditionStrings())
									update = imageStoreStructured.NewUpdate()
									update.Rendition = pointer.FromString(renditionString)
								})

								It("returns updated result when the id exists", func() {
									matchAllFields := MatchAllFields(Fields{
										"ID":                PointTo(Equal(id)),
										"UserID":            Equal(original.UserID),
										"Status":            Equal(original.Status),
										"Metadata":          Equal(original.Metadata),
										"ContentID":         Equal(original.ContentID),
										"ContentIntent":     Equal(original.ContentIntent),
										"ContentAttributes": Equal(original.ContentAttributes),
										"RenditionsID":      Equal(original.RenditionsID),
										"Renditions":        PointTo(Equal(append(*original.Renditions, renditionString))),
										"CreatedTime":       Equal(original.CreatedTime),
										"ModifiedTime":      PointTo(BeTemporally("~", time.Now(), time.Second)),
										"DeletedTime":       Equal(original.DeletedTime),
										"Revision":          PointTo(Equal(*original.Revision + 1)),
									})
									result, err := repository.Update(ctx, id, condition, update)
									Expect(err).ToNot(HaveOccurred())
									Expect(result).ToNot(BeNil())
									Expect(*result).To(matchAllFields)
									storeResult := image.ImageArray{}
									cursor, err := collection.Find(context.Background(), bson.M{"id": id})
									Expect(err).ToNot(HaveOccurred())
									Expect(cursor).ToNot(BeNil())
									Expect(cursor.All(context.Background(), &storeResult)).To(Succeed())
									Expect(storeResult).To(HaveLen(1))
									Expect(*storeResult[0]).To(matchAllFields)
								})

								It("returns nil when the id does not exist", func() {
									id = imageTest.RandomID()
									Expect(repository.Update(ctx, id, condition, update)).To(BeNil())
								})
							})
						})

						Context("without updates", func() {
							BeforeEach(func() {
								update = imageStoreStructured.NewUpdate()
							})

							It("returns original when the id exists", func() {
								Expect(repository.Update(ctx, id, condition, update)).To(Equal(original))
							})

							It("returns nil when the id does not exist", func() {
								id = imageTest.RandomID()
								Expect(repository.Update(ctx, id, condition, update)).To(BeNil())
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

			Context("Delete", func() {
				var id string
				var condition *request.Condition

				BeforeEach(func() {
					id = imageTest.RandomID()
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
					var original *image.Image

					BeforeEach(func() {
						original = imageTest.RandomImage()
						original.ID = pointer.FromString(id)
					})

					JustBeforeEach(func() {
						_, err := collection.InsertOne(ctx, original)
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
									"ID":                PointTo(Equal(id)),
									"UserID":            Equal(original.UserID),
									"Status":            Equal(original.Status),
									"Metadata":          Equal(original.Metadata),
									"ContentID":         Equal(original.ContentID),
									"ContentIntent":     Equal(original.ContentIntent),
									"ContentAttributes": Equal(original.ContentAttributes),
									"RenditionsID":      Equal(original.RenditionsID),
									"Renditions":        Equal(original.Renditions),
									"CreatedTime":       Equal(original.CreatedTime),
									"ModifiedTime":      PointTo(BeTemporally("~", time.Now(), time.Second)),
									"DeletedTime":       PointTo(BeTemporally("~", time.Now(), time.Second)),
									"Revision":          PointTo(Equal(*original.Revision + 1)),
								})
								Expect(repository.Delete(ctx, id, condition)).To(BeTrue())
								storeResult := image.ImageArray{}
								cursor, err := collection.Find(context.Background(), bson.M{"id": id})
								Expect(err).ToNot(HaveOccurred())
								Expect(cursor).ToNot(BeNil())
								Expect(cursor.All(context.Background(), &storeResult)).To(Succeed())
								Expect(storeResult).To(HaveLen(1))
								Expect(*storeResult[0]).To(matchAllFields)
							})

							It("returns false when the id does not exist", func() {
								id = imageTest.RandomID()
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
					id = imageTest.RandomID()
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
					var original *image.Image

					BeforeEach(func() {
						original = imageTest.RandomImage()
						original.ID = pointer.FromString(id)
					})

					JustBeforeEach(func() {
						_, err := collection.InsertOne(ctx, original)
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
							id = imageTest.RandomID()
							Expect(repository.Destroy(ctx, id, condition)).To(BeFalse())
							Expect(collection.CountDocuments(ctx, bson.M{"id": original.ID})).To(Equal(int64(1)))
						})

						It("returns false and does not destroy the original when the id exists, but the condition revision does not match", func() {
							condition.Revision = pointer.FromInt(*original.Revision + 1)
							Expect(repository.Destroy(ctx, id, condition)).To(BeFalse())
							Expect(collection.CountDocuments(ctx, bson.M{"id": original.ID})).To(Equal(int64(1)))
						})

						It("returns true and destroys the original when the id exists and the condition is missing", func() {
							condition = nil
							Expect(repository.Destroy(ctx, id, condition)).To(BeTrue())
							Expect(collection.CountDocuments(ctx, bson.M{"id": original.ID})).To(Equal(int64(0)))
						})

						It("returns true and destroys the original when the id exists and the condition revision is missing", func() {
							condition.Revision = nil
							Expect(repository.Destroy(ctx, id, condition)).To(BeTrue())
							Expect(collection.CountDocuments(ctx, bson.M{"id": original.ID})).To(Equal(int64(0)))
						})

						It("returns true and destroys the original when the id exists and the condition revision matches", func() {
							condition.Revision = pointer.CloneInt(original.Revision)
							Expect(repository.Destroy(ctx, id, condition)).To(BeTrue())
							Expect(collection.CountDocuments(ctx, bson.M{"id": original.ID})).To(Equal(int64(0)))
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
