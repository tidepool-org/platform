package mongo_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"context"
	"math/rand"
	"sort"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/blob"
	blobStoreStructured "github.com/tidepool-org/platform/blob/store/structured"
	blobStoreStructuredMongo "github.com/tidepool-org/platform/blob/store/structured/mongo"
	blobStoreStructuredTest "github.com/tidepool-org/platform/blob/store/structured/test"
	blobTest "github.com/tidepool-org/platform/blob/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	netTest "github.com/tidepool-org/platform/net/test"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
	"github.com/tidepool-org/platform/user"
)

type CreatedTimeDescending blob.Blobs

func (c CreatedTimeDescending) Len() int {
	return len(c)
}

func (c CreatedTimeDescending) Swap(left int, right int) {
	c[left], c[right] = c[right], c[left]
}

func (c CreatedTimeDescending) Less(left int, right int) bool {
	if c[left].CreatedTime == nil {
		return true
	} else if c[right].CreatedTime == nil {
		return false
	}
	return c[right].CreatedTime.Before(*c[left].CreatedTime)
}

func SelectAndSort(blbs blob.Blobs, selector func(b *blob.Blob) bool) blob.Blobs {
	var selected blob.Blobs
	for _, b := range blbs {
		if selector(b) {
			selected = append(selected, b)
		}
	}
	sort.Sort(CreatedTimeDescending(selected))
	return selected
}

func AsInterfaceArray(blbs blob.Blobs) []interface{} {
	if blbs == nil {
		return nil
	}
	array := make([]interface{}, len(blbs))
	for index, blb := range blbs {
		array[index] = blb
	}
	return array
}

var _ = Describe("Mongo", func() {
	var config *storeStructuredMongo.Config
	var logger *logTest.Logger
	var store *blobStoreStructuredMongo.Store
	var session blobStoreStructured.Session

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
			store, err = blobStoreStructuredMongo.NewStore(nil, logger)
			errorsTest.ExpectEqual(err, errors.New("config is missing"))
			Expect(store).To(BeNil())
		})

		It("returns a new store and no error when successful", func() {
			var err error
			store, err = blobStoreStructuredMongo.NewStore(config, logger)
			Expect(err).ToNot(HaveOccurred())
			Expect(store).ToNot(BeNil())
		})
	})

	Context("with a new store", func() {
		var mgoSession *mgo.Session
		var mgoCollection *mgo.Collection

		BeforeEach(func() {
			var err error
			store, err = blobStoreStructuredMongo.NewStore(config, logger)
			Expect(err).ToNot(HaveOccurred())
			Expect(store).ToNot(BeNil())
			mgoSession = storeStructuredMongoTest.Session().Copy()
			mgoCollection = mgoSession.DB(config.Database).C(config.CollectionPrefix + "blobs")
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
					MatchFields(IgnoreExtras, Fields{"Key": ConsistOf("userId"), "Background": Equal(true)}),
					MatchFields(IgnoreExtras, Fields{"Key": ConsistOf("mediaType"), "Background": Equal(true)}),
					MatchFields(IgnoreExtras, Fields{"Key": ConsistOf("status"), "Background": Equal(true)}),
				))
			})
		})

		Context("NewBlobsSession", func() {
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
					userID = user.NewID()
				})

				Context("List", func() {
					var filter *blob.Filter
					var pagination *page.Pagination

					BeforeEach(func() {
						filter = blob.NewFilter()
						pagination = page.NewPagination()
					})

					It("returns an error when the context is missing", func() {
						ctx = nil
						blbs, err := session.List(ctx, userID, filter, pagination)
						errorsTest.ExpectEqual(err, errors.New("context is missing"))
						Expect(blbs).To(BeNil())
					})

					It("returns an error when the user id is missing", func() {
						userID = ""
						blbs, err := session.List(ctx, userID, filter, pagination)
						errorsTest.ExpectEqual(err, errors.New("user id is missing"))
						Expect(blbs).To(BeNil())
					})

					It("returns an error when the user id is invalid", func() {
						userID = "invalid"
						blbs, err := session.List(ctx, userID, filter, pagination)
						errorsTest.ExpectEqual(err, errors.New("user id is invalid"))
						Expect(blbs).To(BeNil())
					})

					It("returns an error when the filter is invalid", func() {
						filter.MediaType = pointer.FromStringArray([]string{""})
						blbs, err := session.List(ctx, userID, filter, pagination)
						errorsTest.ExpectEqual(err, errors.New("filter is invalid"))
						Expect(blbs).To(BeNil())
					})

					It("returns an error when the pagination is invalid", func() {
						pagination.Page = -1
						blbs, err := session.List(ctx, userID, filter, pagination)
						errorsTest.ExpectEqual(err, errors.New("pagination is invalid"))
						Expect(blbs).To(BeNil())
					})

					It("returns an error when the session is closed", func() {
						session.Close()
						blbs, err := session.List(ctx, userID, filter, pagination)
						errorsTest.ExpectEqual(err, errors.New("session closed"))
						Expect(blbs).To(BeNil())
					})

					Context("with data", func() {
						var mediaType string
						var allBlobs blob.Blobs

						BeforeEach(func() {
							mediaType = netTest.RandomMediaType()
							allBlobs = blob.Blobs{}
							for index, randomBlob := range blobTest.RandomBlobs(4, 4) {
								if index < 2 {
									randomBlob.Status = pointer.FromString(blob.StatusAvailable)
								} else {
									randomBlob.Status = pointer.FromString(blob.StatusCreated)
								}
								if index%2 == 0 {
									randomBlob.MediaType = pointer.FromString(mediaType)
								}
								userBlob := blobTest.CloneBlob(randomBlob)
								userBlob.ID = pointer.FromString(blob.NewID())
								userBlob.UserID = pointer.FromString(userID)
								allBlobs = append(allBlobs, randomBlob, userBlob)
							}
							rand.Shuffle(len(allBlobs), func(i, j int) { allBlobs[i], allBlobs[j] = allBlobs[j], allBlobs[i] })
							Expect(mgoCollection.Insert(AsInterfaceArray(allBlobs)...)).To(Succeed())
						})

						It("returns no blobs when the user id is unknown", func() {
							userID = user.NewID()
							Expect(session.List(ctx, userID, filter, pagination)).To(SatisfyAll(Not(BeNil()), BeEmpty()))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 0})
						})

						It("returns expected blobs when the filter is missing", func() {
							filter = nil
							Expect(session.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allBlobs,
								func(b *blob.Blob) bool { return *b.UserID == userID && *b.Status == blob.StatusAvailable },
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "pagination": pagination, "count": 2})
						})

						It("returns expected blobs when the filter media type is missing", func() {
							filter.MediaType = nil
							Expect(session.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allBlobs,
								func(b *blob.Blob) bool { return *b.UserID == userID && *b.Status == blob.StatusAvailable },
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 2})
						})

						It("returns expected blobs when the filter media type is specified", func() {
							filter.MediaType = pointer.FromStringArray([]string{netTest.RandomMediaType(), mediaType})
							Expect(session.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allBlobs,
								func(b *blob.Blob) bool {
									return *b.UserID == userID && *b.MediaType == mediaType && *b.Status == blob.StatusAvailable
								},
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 1})
						})

						It("returns expected blobs when the filter status is missing", func() {
							filter.Status = nil
							Expect(session.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allBlobs,
								func(b *blob.Blob) bool { return *b.UserID == userID && *b.Status == blob.StatusAvailable },
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 2})
						})

						It("returns expected blobs when the filter status is set to available", func() {
							filter.Status = pointer.FromStringArray([]string{blob.StatusAvailable})
							Expect(session.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allBlobs,
								func(b *blob.Blob) bool { return *b.UserID == userID && *b.Status == blob.StatusAvailable },
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 2})
						})

						It("returns expected blobs when the filter status is set to created", func() {
							filter.Status = pointer.FromStringArray([]string{blob.StatusCreated})
							Expect(session.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allBlobs,
								func(b *blob.Blob) bool { return *b.UserID == userID && *b.Status == blob.StatusCreated },
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 2})
						})

						It("returns expected blobs when the filter status is set to both available and created", func() {
							filter.Status = pointer.FromStringArray(blob.Statuses())
							Expect(session.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allBlobs,
								func(b *blob.Blob) bool { return *b.UserID == userID },
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 4})
						})

						It("returns expected blobs when the filter media type and status available are specified", func() {
							filter.MediaType = pointer.FromStringArray([]string{netTest.RandomMediaType(), mediaType})
							filter.Status = pointer.FromStringArray([]string{blob.StatusAvailable})
							Expect(session.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allBlobs,
								func(b *blob.Blob) bool {
									return *b.UserID == userID && *b.MediaType == mediaType && *b.Status == blob.StatusAvailable
								},
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 1})
						})

						It("returns expected blobs when the filter media type and status created are specified", func() {
							filter.MediaType = pointer.FromStringArray([]string{netTest.RandomMediaType(), mediaType})
							filter.Status = pointer.FromStringArray([]string{blob.StatusCreated})
							Expect(session.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allBlobs,
								func(b *blob.Blob) bool {
									return *b.UserID == userID && *b.MediaType == mediaType && *b.Status == blob.StatusCreated
								},
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 1})
						})

						It("returns expected blobs when the pagination is missing", func() {
							filter.Status = pointer.FromStringArray(blob.Statuses())
							pagination = nil
							Expect(session.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allBlobs,
								func(b *blob.Blob) bool { return *b.UserID == userID },
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "count": 4})
						})

						It("returns expected blobs when the pagination limits blobs", func() {
							filter.Status = pointer.FromStringArray(blob.Statuses())
							pagination.Page = 1
							pagination.Size = 2
							Expect(session.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allBlobs,
								func(b *blob.Blob) bool { return *b.UserID == userID },
							)[2:4]))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 2})
						})
					})
				})

				Context("Create", func() {
					var create *blobStoreStructured.Create

					BeforeEach(func() {
						create = blobStoreStructuredTest.RandomCreate()
					})

					It("returns an error when the context is missing", func() {
						ctx = nil
						blb, err := session.Create(ctx, userID, create)
						errorsTest.ExpectEqual(err, errors.New("context is missing"))
						Expect(blb).To(BeNil())
					})

					It("returns an error when the user id is missing", func() {
						userID = ""
						blb, err := session.Create(ctx, userID, create)
						errorsTest.ExpectEqual(err, errors.New("user id is missing"))
						Expect(blb).To(BeNil())
					})

					It("returns an error when the user id is invalid", func() {
						userID = "invalid"
						blb, err := session.Create(ctx, userID, create)
						errorsTest.ExpectEqual(err, errors.New("user id is invalid"))
						Expect(blb).To(BeNil())
					})

					It("returns an error when the create is missing", func() {
						create = nil
						blb, err := session.Create(ctx, userID, create)
						errorsTest.ExpectEqual(err, errors.New("create is missing"))
						Expect(blb).To(BeNil())
					})

					It("returns an error when the create is invalid", func() {
						create.MediaType = pointer.FromString("")
						blb, err := session.Create(ctx, userID, create)
						errorsTest.ExpectEqual(err, errors.New("create is invalid"))
						Expect(blb).To(BeNil())
					})

					It("returns an error when the session is closed", func() {
						session.Close()
						blb, err := session.Create(ctx, userID, create)
						errorsTest.ExpectEqual(err, errors.New("session closed"))
						Expect(blb).To(BeNil())
					})

					It("returns the blob after creating", func() {
						matchAllFields := MatchAllFields(Fields{
							"ID":           PointTo(Not(BeEmpty())),
							"UserID":       PointTo(Equal(userID)),
							"DigestMD5":    BeNil(),
							"MediaType":    Equal(create.MediaType),
							"Size":         BeNil(),
							"Status":       PointTo(Equal(blob.StatusCreated)),
							"CreatedTime":  PointTo(BeTemporally("~", time.Now(), time.Second)),
							"ModifiedTime": BeNil(),
						})
						blb, err := session.Create(ctx, userID, create)
						Expect(err).ToNot(HaveOccurred())
						Expect(blb).ToNot(BeNil())
						Expect(*blb).To(matchAllFields)
						blbs := blob.Blobs{}
						Expect(mgoCollection.Find(bson.M{"id": blb.ID}).All(&blbs)).To(Succeed())
						Expect(blbs).To(HaveLen(1))
						Expect(*blbs[0]).To(matchAllFields)
						logger.AssertDebug("Create", log.Fields{"userId": userID, "create": create, "id": *blbs[0].ID})
					})

					It("returns the blob after creating without media type", func() {
						create.MediaType = nil
						matchAllFields := MatchAllFields(Fields{
							"ID":           PointTo(Not(BeEmpty())),
							"UserID":       PointTo(Equal(userID)),
							"DigestMD5":    BeNil(),
							"MediaType":    BeNil(),
							"Size":         BeNil(),
							"Status":       PointTo(Equal(blob.StatusCreated)),
							"CreatedTime":  PointTo(BeTemporally("~", time.Now(), time.Second)),
							"ModifiedTime": BeNil(),
						})
						blb, err := session.Create(ctx, userID, create)
						Expect(err).ToNot(HaveOccurred())
						Expect(blb).ToNot(BeNil())
						Expect(*blb).To(matchAllFields)
						blbs := blob.Blobs{}
						Expect(mgoCollection.Find(bson.M{"id": blb.ID}).All(&blbs)).To(Succeed())
						Expect(blbs).To(HaveLen(1))
						Expect(*blbs[0]).To(matchAllFields)
						logger.AssertDebug("Create", log.Fields{"userId": userID, "create": create, "id": *blbs[0].ID})
					})
				})
			})

			Context("Get", func() {
				var id string

				BeforeEach(func() {
					id = blob.NewID()
				})

				It("returns an error when the context is missing", func() {
					ctx = nil
					blb, err := session.Get(ctx, id)
					errorsTest.ExpectEqual(err, errors.New("context is missing"))
					Expect(blb).To(BeNil())
				})

				It("returns an error when the id is missing", func() {
					id = ""
					blb, err := session.Get(ctx, id)
					errorsTest.ExpectEqual(err, errors.New("id is missing"))
					Expect(blb).To(BeNil())
				})

				It("returns an error when the id is invalid", func() {
					id = "invalid"
					blb, err := session.Get(ctx, id)
					errorsTest.ExpectEqual(err, errors.New("id is invalid"))
					Expect(blb).To(BeNil())
				})

				It("returns an error when the session is closed", func() {
					session.Close()
					blb, err := session.Get(ctx, id)
					errorsTest.ExpectEqual(err, errors.New("session closed"))
					Expect(blb).To(BeNil())
				})

				Context("with data", func() {
					var allBlobs blob.Blobs
					var blb *blob.Blob

					BeforeEach(func() {
						allBlobs = blobTest.RandomBlobs(4, 4)
						blb = allBlobs[0]
						blb.ID = pointer.FromString(id)
						rand.Shuffle(len(allBlobs), func(i, j int) { allBlobs[i], allBlobs[j] = allBlobs[j], allBlobs[i] })
						Expect(mgoCollection.Insert(AsInterfaceArray(allBlobs)...)).To(Succeed())
					})

					AfterEach(func() {
						logger.AssertDebug("Get", log.Fields{"id": id})
					})

					It("returns nil when the id does not exist", func() {
						id = blob.NewID()
						Expect(session.Get(ctx, id)).To(BeNil())
					})

					It("returns the blob when the id exists", func() {
						Expect(session.Get(ctx, id)).To(Equal(blb))
					})
				})
			})

			Context("Update", func() {
				var id string
				var update *blobStoreStructured.Update

				BeforeEach(func() {
					id = blob.NewID()
					update = blobStoreStructuredTest.RandomUpdate()
				})

				It("returns an error when the context is missing", func() {
					ctx = nil
					blb, err := session.Update(ctx, id, update)
					errorsTest.ExpectEqual(err, errors.New("context is missing"))
					Expect(blb).To(BeNil())
				})

				It("returns an error when the id is missing", func() {
					id = ""
					blb, err := session.Update(ctx, id, update)
					errorsTest.ExpectEqual(err, errors.New("id is missing"))
					Expect(blb).To(BeNil())
				})

				It("returns an error when the id is invalid", func() {
					id = "invalid"
					blb, err := session.Update(ctx, id, update)
					errorsTest.ExpectEqual(err, errors.New("id is invalid"))
					Expect(blb).To(BeNil())
				})

				It("returns an error when the update is missing", func() {
					update = nil
					blb, err := session.Update(ctx, id, update)
					errorsTest.ExpectEqual(err, errors.New("update is missing"))
					Expect(blb).To(BeNil())
				})

				It("returns an error when the update is invalid", func() {
					update.DigestMD5 = pointer.FromString("")
					blb, err := session.Update(ctx, id, update)
					errorsTest.ExpectEqual(err, errors.New("update is invalid"))
					Expect(blb).To(BeNil())
				})

				It("returns an error when the session is closed", func() {
					session.Close()
					blb, err := session.Update(ctx, id, update)
					errorsTest.ExpectEqual(err, errors.New("session closed"))
					Expect(blb).To(BeNil())
				})

				Context("with data", func() {
					var originalBlb *blob.Blob

					BeforeEach(func() {
						originalBlb = blobTest.RandomBlob()
						originalBlb.ID = pointer.FromString(id)
						Expect(mgoCollection.Insert(originalBlb)).To(Succeed())
					})

					AfterEach(func() {
						logger.AssertDebug("Update", log.Fields{"id": id, "update": update})
					})

					It("returns updated blob when the id exists", func() {
						matchAllFields := MatchAllFields(Fields{
							"ID":           PointTo(Equal(id)),
							"UserID":       Equal(originalBlb.UserID),
							"DigestMD5":    Equal(update.DigestMD5),
							"MediaType":    Equal(update.MediaType),
							"Size":         Equal(update.Size),
							"Status":       Equal(update.Status),
							"CreatedTime":  PointTo(Not(BeZero())),
							"ModifiedTime": PointTo(BeTemporally("~", time.Now(), time.Second)),
						})
						blb, err := session.Update(ctx, id, update)
						Expect(err).ToNot(HaveOccurred())
						Expect(blb).ToNot(BeNil())
						Expect(*blb).To(matchAllFields)
						blbs := blob.Blobs{}
						Expect(mgoCollection.Find(bson.M{"id": id}).All(&blbs)).To(Succeed())
						Expect(blbs).To(HaveLen(1))
						Expect(*blbs[0]).To(matchAllFields)
					})

					It("returns nil when the id does not exist", func() {
						id = blob.NewID()
						Expect(session.Update(ctx, id, update)).To(BeNil())
					})

					Context("without updates", func() {
						BeforeEach(func() {
							update = blobStoreStructured.NewUpdate()
						})

						It("returns original blob when the id exists", func() {
							Expect(session.Update(ctx, id, update)).To(Equal(originalBlb))
						})

						It("returns nil when the id does not exist", func() {
							id = blob.NewID()
							Expect(session.Update(ctx, id, update)).To(BeNil())
						})
					})
				})
			})

			Context("Delete", func() {
				var id string

				BeforeEach(func() {
					id = blob.NewID()
				})

				It("returns an error when the context is missing", func() {
					ctx = nil
					deleted, err := session.Delete(ctx, id)
					errorsTest.ExpectEqual(err, errors.New("context is missing"))
					Expect(deleted).To(BeFalse())
				})

				It("returns an error when the id is missing", func() {
					id = ""
					deleted, err := session.Delete(ctx, id)
					errorsTest.ExpectEqual(err, errors.New("id is missing"))
					Expect(deleted).To(BeFalse())
				})

				It("returns an error when the id is invalid", func() {
					id = "invalid"
					deleted, err := session.Delete(ctx, id)
					errorsTest.ExpectEqual(err, errors.New("id is invalid"))
					Expect(deleted).To(BeFalse())
				})

				It("returns an error when the session is closed", func() {
					session.Close()
					deleted, err := session.Delete(ctx, id)
					errorsTest.ExpectEqual(err, errors.New("session closed"))
					Expect(deleted).To(BeFalse())
				})

				Context("with data", func() {
					var blb *blob.Blob

					BeforeEach(func() {
						blb = blobTest.RandomBlob()
						blb.ID = pointer.FromString(id)
						Expect(mgoCollection.Insert(blb)).To(Succeed())
					})

					AfterEach(func() {
						logger.AssertDebug("Delete", log.Fields{"id": id})
					})

					It("returns true and deletes the blob when the id exists", func() {
						Expect(session.Delete(ctx, id)).To(BeTrue())
						Expect(mgoCollection.Find(bson.M{"id": id}).Count()).To(Equal(0))
					})

					It("returns false when the id does not exist", func() {
						id = blob.NewID()
						Expect(session.Delete(ctx, id)).To(BeFalse())
					})
				})
			})
		})
	})
})
