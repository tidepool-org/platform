package mongo_test

import (
	"context"
	"math/rand"
	"sort"
	"time"

	mgo "github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	. "github.com/onsi/gomega/types"

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
	var session imageStoreStructured.Session

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
			store, err = imageStoreStructuredMongo.NewStore(nil, logger)
			errorsTest.ExpectEqual(err, errors.New("config is missing"))
			Expect(store).To(BeNil())
		})

		It("returns a new store and no error when successful", func() {
			var err error
			store, err = imageStoreStructuredMongo.NewStore(config, logger)
			Expect(err).ToNot(HaveOccurred())
			Expect(store).ToNot(BeNil())
		})
	})

	Context("with a new store", func() {
		var mgoSession *mgo.Session
		var mgoCollection *mgo.Collection

		BeforeEach(func() {
			var err error
			store, err = imageStoreStructuredMongo.NewStore(config, logger)
			Expect(err).ToNot(HaveOccurred())
			Expect(store).ToNot(BeNil())
			mgoSession = storeStructuredMongoTest.Session().Copy()
			mgoCollection = mgoSession.DB(config.Database).C(config.CollectionPrefix + "images")
		})

		AfterEach(func() {
			if mgoSession != nil {
				mgoSession.Close()
			}
		})

		Context("EnsureIndexes", func() {
			It("returns successfully", func() {
				Expect(store.EnsureIndexes()).To(Succeed())
				indexes, err := mgoCollection.Indexes()
				Expect(err).ToNot(HaveOccurred())
				Expect(indexes).To(ConsistOf(
					MatchFields(IgnoreExtras, Fields{"Key": ConsistOf("_id")}),
					MatchFields(IgnoreExtras, Fields{"Key": ConsistOf("id"), "Background": Equal(true), "Unique": Equal(true)}),
					MatchFields(IgnoreExtras, Fields{"Key": ConsistOf("userId", "status"), "Background": Equal(true)}),
				))
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
						result, err := session.List(ctx, userID, filter, pagination)
						errorsTest.ExpectEqual(err, errors.New("context is missing"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the user id is missing", func() {
						userID = ""
						result, err := session.List(ctx, userID, filter, pagination)
						errorsTest.ExpectEqual(err, errors.New("user id is missing"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the user id is invalid", func() {
						userID = "invalid"
						result, err := session.List(ctx, userID, filter, pagination)
						errorsTest.ExpectEqual(err, errors.New("user id is invalid"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the filter is invalid", func() {
						filter.Status = pointer.FromStringArray([]string{""})
						result, err := session.List(ctx, userID, filter, pagination)
						errorsTest.ExpectEqual(err, errors.New("filter is invalid"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the pagination is invalid", func() {
						pagination.Page = -1
						result, err := session.List(ctx, userID, filter, pagination)
						errorsTest.ExpectEqual(err, errors.New("pagination is invalid"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the session is closed", func() {
						session.Close()
						result, err := session.List(ctx, userID, filter, pagination)
						errorsTest.ExpectEqual(err, errors.New("session closed"))
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
							Expect(mgoCollection.Insert(AsInterfaceArray(allResult)...)).To(Succeed())
						})

						It("returns no result when the user id is unknown", func() {
							userID = userTest.RandomID()
							Expect(session.List(ctx, userID, filter, pagination)).To(SatisfyAll(Not(BeNil()), BeEmpty()))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 0})
						})

						It("returns expected result when the filter is missing", func() {
							filter = nil
							Expect(session.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(i *image.Image) bool {
									return *i.UserID == userID && i.DeletedTime == nil && *i.Status == image.StatusAvailable
								},
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "pagination": pagination, "count": 2})
						})

						It("returns expected result when the filter status is missing", func() {
							filter.Status = nil
							Expect(session.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(i *image.Image) bool {
									return *i.UserID == userID && i.DeletedTime == nil && *i.Status == image.StatusAvailable
								},
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 2})
						})

						It("returns expected result when the filter status is set to available", func() {
							filter.Status = pointer.FromStringArray([]string{image.StatusAvailable})
							Expect(session.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(i *image.Image) bool {
									return *i.UserID == userID && i.DeletedTime == nil && *i.Status == image.StatusAvailable
								},
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 2})
						})

						It("returns expected result when the filter status is set to created", func() {
							filter.Status = pointer.FromStringArray([]string{image.StatusCreated})
							Expect(session.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(i *image.Image) bool {
									return *i.UserID == userID && i.DeletedTime == nil && *i.Status == image.StatusCreated
								},
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 2})
						})

						It("returns expected result when the filter status is set to both available and original", func() {
							filter.Status = pointer.FromStringArray(image.Statuses())
							Expect(session.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(i *image.Image) bool { return *i.UserID == userID && i.DeletedTime == nil },
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 4})
						})

						It("returns expected result when the filter content intent is missing", func() {
							filter.ContentIntent = nil
							Expect(session.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(i *image.Image) bool {
									return *i.UserID == userID && i.DeletedTime == nil && *i.Status == image.StatusAvailable
								},
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 2})
						})

						It("returns expected result when the filter content intent is set to alternate", func() {
							filter.ContentIntent = pointer.FromStringArray([]string{image.ContentIntentAlternate})
							Expect(session.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(i *image.Image) bool {
									return *i.UserID == userID && i.DeletedTime == nil && *i.Status == image.StatusAvailable && *i.ContentIntent == image.ContentIntentAlternate
								},
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 1})
						})

						It("returns expected result when the filter content intent is set to original", func() {
							filter.ContentIntent = pointer.FromStringArray([]string{image.ContentIntentOriginal})
							Expect(session.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(i *image.Image) bool {
									return *i.UserID == userID && i.DeletedTime == nil && *i.Status == image.StatusAvailable && *i.ContentIntent == image.ContentIntentOriginal
								},
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 1})
						})

						It("returns expected result when the filter content intent is set to both alternate and original", func() {
							filter.ContentIntent = pointer.FromStringArray(image.ContentIntents())
							Expect(session.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(i *image.Image) bool {
									return *i.UserID == userID && i.DeletedTime == nil && *i.Status == image.StatusAvailable
								},
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 2})
						})

						It("returns expected result when the filter status is set to both available and created and filter content intent is set to alternate", func() {
							filter.Status = pointer.FromStringArray(image.Statuses())
							filter.ContentIntent = pointer.FromStringArray([]string{image.ContentIntentAlternate})
							Expect(session.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(i *image.Image) bool {
									return *i.UserID == userID && i.DeletedTime == nil && *i.Status == image.StatusAvailable && *i.ContentIntent == image.ContentIntentAlternate
								},
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 1})
						})

						It("returns expected result when the filter status is set to both available and created and filter content intent is set to original", func() {
							filter.Status = pointer.FromStringArray(image.Statuses())
							filter.ContentIntent = pointer.FromStringArray([]string{image.ContentIntentOriginal})
							Expect(session.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(i *image.Image) bool {
									return *i.UserID == userID && i.DeletedTime == nil && *i.Status == image.StatusAvailable && *i.ContentIntent == image.ContentIntentOriginal
								},
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 1})
						})

						It("returns expected result when the filter status is set to both available and created and filter content intent is set to alternate and original", func() {
							filter.Status = pointer.FromStringArray(image.Statuses())
							filter.ContentIntent = pointer.FromStringArray(image.ContentIntents())
							Expect(session.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(i *image.Image) bool {
									return *i.UserID == userID && i.DeletedTime == nil && *i.Status == image.StatusAvailable
								},
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 2})
						})

						It("returns expected result when the pagination is missing", func() {
							filter.Status = pointer.FromStringArray(image.Statuses())
							pagination = nil
							Expect(session.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(i *image.Image) bool { return *i.UserID == userID && i.DeletedTime == nil },
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "count": 4})
						})

						It("returns expected result when the pagination limits result", func() {
							filter.Status = pointer.FromStringArray(image.Statuses())
							pagination.Page = 1
							pagination.Size = 2
							Expect(session.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
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
						result, err := session.Create(ctx, userID, metadata)
						errorsTest.ExpectEqual(err, errors.New("context is missing"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the user id is missing", func() {
						userID = ""
						result, err := session.Create(ctx, userID, metadata)
						errorsTest.ExpectEqual(err, errors.New("user id is missing"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the user id is invalid", func() {
						userID = "invalid"
						result, err := session.Create(ctx, userID, metadata)
						errorsTest.ExpectEqual(err, errors.New("user id is invalid"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the metadata is missing", func() {
						metadata = nil
						result, err := session.Create(ctx, userID, metadata)
						errorsTest.ExpectEqual(err, errors.New("metadata is missing"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the metadata is invalid", func() {
						metadata.Name = pointer.FromString("")
						result, err := session.Create(ctx, userID, metadata)
						errorsTest.ExpectEqual(err, errors.New("metadata is invalid"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the session is closed", func() {
						session.Close()
						result, err := session.Create(ctx, userID, metadata)
						errorsTest.ExpectEqual(err, errors.New("session closed"))
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
						result, err := session.Create(ctx, userID, metadata)
						Expect(err).ToNot(HaveOccurred())
						Expect(result).ToNot(BeNil())
						Expect(*result).To(matchAllFields)
						storeResult := image.ImageArray{}
						Expect(mgoCollection.Find(bson.M{"id": result.ID}).All(&storeResult)).To(Succeed())
						Expect(storeResult).To(HaveLen(1))
						Expect(*storeResult[0]).To(matchAllFields)
						logger.AssertDebug("Create", log.Fields{"userId": userID, "metadata": metadata, "id": *storeResult[0].ID})
					})
				})

				Context("DeleteAll", func() {
					It("returns an error when the context is missing", func() {
						ctx = nil
						deleted, err := session.DeleteAll(ctx, userID)
						errorsTest.ExpectEqual(err, errors.New("context is missing"))
						Expect(deleted).To(BeFalse())
					})

					It("returns an error when the user id is missing", func() {
						userID = ""
						deleted, err := session.DeleteAll(ctx, userID)
						errorsTest.ExpectEqual(err, errors.New("user id is missing"))
						Expect(deleted).To(BeFalse())
					})

					It("returns an error when the user id is invalid", func() {
						userID = "invalid"
						deleted, err := session.DeleteAll(ctx, userID)
						errorsTest.ExpectEqual(err, errors.New("user id is invalid"))
						Expect(deleted).To(BeFalse())
					})

					It("returns an error when the session is closed", func() {
						session.Close()
						deleted, err := session.DeleteAll(ctx, userID)
						errorsTest.ExpectEqual(err, errors.New("session closed"))
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
								Expect(mgoCollection.Insert(original)).To(Succeed())
							}
							Expect(mgoCollection.Insert(imageTest.RandomImage(), imageTest.RandomImage())).To(Succeed())
						})

						AfterEach(func() {
							logger.AssertDebug("DeleteAll", log.Fields{"userId": userID})
						})

						It("returns false and does not delete the originals when the user id does not match", func() {
							originalUserID := userID
							userID = userTest.RandomID()
							Expect(session.DeleteAll(ctx, userID)).To(BeFalse())
							Expect(mgoCollection.Find(bson.M{"userId": originalUserID, "deletedTime": bson.M{"$exists": true}}).Count()).To(Equal(2))
							Expect(mgoCollection.Find(bson.M{"deletedTime": bson.M{"$exists": true}}).Count()).To(Equal(2))
						})

						It("returns true and deletes the originals when the user id matches", func() {
							Expect(session.DeleteAll(ctx, userID)).To(BeTrue())
							Expect(mgoCollection.Find(bson.M{"userId": userID, "deletedTime": bson.M{"$exists": true}}).Count()).To(Equal(4))
							Expect(mgoCollection.Find(bson.M{"deletedTime": bson.M{"$exists": true}}).Count()).To(Equal(4))
						})
					})
				})

				Context("DestroyAll", func() {
					It("returns an error when the context is missing", func() {
						ctx = nil
						destroyed, err := session.DestroyAll(ctx, userID)
						errorsTest.ExpectEqual(err, errors.New("context is missing"))
						Expect(destroyed).To(BeFalse())
					})

					It("returns an error when the user id is missing", func() {
						userID = ""
						destroyed, err := session.DestroyAll(ctx, userID)
						errorsTest.ExpectEqual(err, errors.New("user id is missing"))
						Expect(destroyed).To(BeFalse())
					})

					It("returns an error when the user id is invalid", func() {
						userID = "invalid"
						destroyed, err := session.DestroyAll(ctx, userID)
						errorsTest.ExpectEqual(err, errors.New("user id is invalid"))
						Expect(destroyed).To(BeFalse())
					})

					It("returns an error when the session is closed", func() {
						session.Close()
						destroyed, err := session.DestroyAll(ctx, userID)
						errorsTest.ExpectEqual(err, errors.New("session closed"))
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
								Expect(mgoCollection.Insert(original)).To(Succeed())
							}
							Expect(mgoCollection.Insert(imageTest.RandomImage(), imageTest.RandomImage())).To(Succeed())
						})

						AfterEach(func() {
							logger.AssertDebug("DestroyAll", log.Fields{"userId": userID})
						})

						It("returns false and does not destroy the originals when the user id does not match", func() {
							originalUserID := userID
							userID = userTest.RandomID()
							Expect(session.DestroyAll(ctx, userID)).To(BeFalse())
							Expect(mgoCollection.Find(bson.M{"userId": originalUserID}).Count()).To(Equal(4))
							Expect(mgoCollection.Find(bson.M{}).Count()).To(Equal(6))
						})

						It("returns true and destroys the originals when the user id matches", func() {
							Expect(session.DestroyAll(ctx, userID)).To(BeTrue())
							Expect(mgoCollection.Find(bson.M{"userId": userID}).Count()).To(Equal(0))
							Expect(mgoCollection.Find(bson.M{}).Count()).To(Equal(2))
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
					var allResult image.ImageArray
					var result *image.Image

					BeforeEach(func() {
						allResult = imageTest.RandomImageArray(3, 3)
						result = allResult[0]
						result.ID = pointer.FromString(id)
						rand.Shuffle(len(allResult), func(i, j int) { allResult[i], allResult[j] = allResult[j], allResult[i] })
					})

					JustBeforeEach(func() {
						Expect(mgoCollection.Insert(AsInterfaceArray(allResult)...)).To(Succeed())
					})

					AfterEach(func() {
						logger.AssertDebug("Get", log.Fields{"id": id})
					})

					It("returns nil when the id does not exist", func() {
						id = imageTest.RandomID()
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
					result, err := session.Update(ctx, id, condition, update)
					errorsTest.ExpectEqual(err, errors.New("context is missing"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the id is missing", func() {
					id = ""
					result, err := session.Update(ctx, id, condition, update)
					errorsTest.ExpectEqual(err, errors.New("id is missing"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the id is invalid", func() {
					id = "invalid"
					result, err := session.Update(ctx, id, condition, update)
					errorsTest.ExpectEqual(err, errors.New("id is invalid"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the condition is invalid", func() {
					condition.Revision = pointer.FromInt(-1)
					result, err := session.Update(ctx, id, condition, update)
					errorsTest.ExpectEqual(err, errors.New("condition is invalid"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the update is missing", func() {
					update = nil
					result, err := session.Update(ctx, id, condition, update)
					errorsTest.ExpectEqual(err, errors.New("update is missing"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the update is invalid", func() {
					update.ContentIntent = pointer.FromString("")
					result, err := session.Update(ctx, id, condition, update)
					errorsTest.ExpectEqual(err, errors.New("update is invalid"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the session is closed", func() {
					session.Close()
					result, err := session.Update(ctx, id, condition, update)
					errorsTest.ExpectEqual(err, errors.New("session closed"))
					Expect(result).To(BeNil())
				})

				Context("with data", func() {
					var original *image.Image

					BeforeEach(func() {
						original = imageTest.RandomImage()
						original.ID = pointer.FromString(id)
					})

					JustBeforeEach(func() {
						Expect(mgoCollection.Insert(original)).To(Succeed())
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
							Expect(session.Update(ctx, id, condition, update)).To(BeNil())
						})
					})

					When("the condition revision does not match", func() {
						BeforeEach(func() {
							condition.Revision = pointer.FromInt(*original.Revision + 1)
						})

						It("returns nil", func() {
							Expect(session.Update(ctx, id, condition, update)).To(BeNil())
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
										result, err := session.Update(ctx, id, condition, update)
										Expect(err).ToNot(HaveOccurred())
										Expect(result).ToNot(BeNil())
										Expect(*result).To(matchAllFields)
										storeResult := image.ImageArray{}
										Expect(mgoCollection.Find(bson.M{"id": id}).All(&storeResult)).To(Succeed())
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
									var contentAttributesCreatedTimeMatcher GomegaMatcher
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
									result, err := session.Update(ctx, id, condition, update)
									Expect(err).ToNot(HaveOccurred())
									Expect(result).ToNot(BeNil())
									Expect(*result).To(matchAllFields)
									storeResult := image.ImageArray{}
									Expect(mgoCollection.Find(bson.M{"id": id}).All(&storeResult)).To(Succeed())
									Expect(storeResult).To(HaveLen(1))
									Expect(*storeResult[0]).To(matchAllFields)
								})

								It("returns nil when the id does not exist", func() {
									id = imageTest.RandomID()
									Expect(session.Update(ctx, id, condition, update)).To(BeNil())
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
									result, err := session.Update(ctx, id, condition, update)
									Expect(err).ToNot(HaveOccurred())
									Expect(result).ToNot(BeNil())
									Expect(*result).To(matchAllFields)
									storeResult := image.ImageArray{}
									Expect(mgoCollection.Find(bson.M{"id": id}).All(&storeResult)).To(Succeed())
									Expect(storeResult).To(HaveLen(1))
									Expect(*storeResult[0]).To(matchAllFields)
								})

								It("returns nil when the id does not exist", func() {
									id = imageTest.RandomID()
									Expect(session.Update(ctx, id, condition, update)).To(BeNil())
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
									result, err := session.Update(ctx, id, condition, update)
									Expect(err).ToNot(HaveOccurred())
									Expect(result).ToNot(BeNil())
									Expect(*result).To(matchAllFields)
									storeResult := image.ImageArray{}
									Expect(mgoCollection.Find(bson.M{"id": id}).All(&storeResult)).To(Succeed())
									Expect(storeResult).To(HaveLen(1))
									Expect(*storeResult[0]).To(matchAllFields)
								})

								It("returns nil when the id does not exist", func() {
									id = imageTest.RandomID()
									Expect(session.Update(ctx, id, condition, update)).To(BeNil())
								})
							})
						})

						Context("without updates", func() {
							BeforeEach(func() {
								update = imageStoreStructured.NewUpdate()
							})

							It("returns original when the id exists", func() {
								Expect(session.Update(ctx, id, condition, update)).To(Equal(original))
							})

							It("returns nil when the id does not exist", func() {
								id = imageTest.RandomID()
								Expect(session.Update(ctx, id, condition, update)).To(BeNil())
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
					var original *image.Image

					BeforeEach(func() {
						original = imageTest.RandomImage()
						original.ID = pointer.FromString(id)
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
								Expect(session.Delete(ctx, id, condition)).To(BeTrue())
								storeResult := image.ImageArray{}
								Expect(mgoCollection.Find(bson.M{"id": id}).All(&storeResult)).To(Succeed())
								Expect(storeResult).To(HaveLen(1))
								Expect(*storeResult[0]).To(matchAllFields)
							})

							It("returns false when the id does not exist", func() {
								id = imageTest.RandomID()
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
					id = imageTest.RandomID()
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
					var original *image.Image

					BeforeEach(func() {
						original = imageTest.RandomImage()
						original.ID = pointer.FromString(id)
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
							id = imageTest.RandomID()
							Expect(session.Destroy(ctx, id, condition)).To(BeFalse())
							Expect(mgoCollection.Find(bson.M{"id": original.ID}).Count()).To(Equal(1))
						})

						It("returns false and does not destroy the original when the id exists, but the condition revision does not match", func() {
							condition.Revision = pointer.FromInt(*original.Revision + 1)
							Expect(session.Destroy(ctx, id, condition)).To(BeFalse())
							Expect(mgoCollection.Find(bson.M{"id": original.ID}).Count()).To(Equal(1))
						})

						It("returns true and destroys the original when the id exists and the condition is missing", func() {
							condition = nil
							Expect(session.Destroy(ctx, id, condition)).To(BeTrue())
							Expect(mgoCollection.Find(bson.M{"id": original.ID}).Count()).To(Equal(0))
						})

						It("returns true and destroys the original when the id exists and the condition revision is missing", func() {
							condition.Revision = nil
							Expect(session.Destroy(ctx, id, condition)).To(BeTrue())
							Expect(mgoCollection.Find(bson.M{"id": original.ID}).Count()).To(Equal(0))
						})

						It("returns true and destroys the original when the id exists and the condition revision matches", func() {
							condition.Revision = pointer.CloneInt(original.Revision)
							Expect(session.Destroy(ctx, id, condition)).To(BeTrue())
							Expect(mgoCollection.Find(bson.M{"id": original.ID}).Count()).To(Equal(0))
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
