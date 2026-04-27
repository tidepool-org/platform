package mongo_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"slices"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"go.mongodb.org/mongo-driver/bson"
	bsonPrimitive "go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	compressTest "github.com/tidepool-org/platform/compress/test"
	"github.com/tidepool-org/platform/crypto"
	dataRaw "github.com/tidepool-org/platform/data/raw"
	dataRawStoreStructuredMongo "github.com/tidepool-org/platform/data/raw/store/structured/mongo"
	dataRawStoreStructuredMongoTest "github.com/tidepool-org/platform/data/raw/store/structured/mongo/test"
	dataRawTest "github.com/tidepool-org/platform/data/raw/test"
	dataTest "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/page"
	pageTest "github.com/tidepool-org/platform/page/test"
	"github.com/tidepool-org/platform/pointer"
	storeStructured "github.com/tidepool-org/platform/store/structured"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
	storeStructuredTest "github.com/tidepool-org/platform/store/structured/test"
	"github.com/tidepool-org/platform/test"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("Mongo", func() {
	Context("Store", func() {
		var config *storeStructuredMongo.Config
		var logger *logTest.Logger
		var store *dataRawStoreStructuredMongo.Store

		BeforeEach(func() {
			config = storeStructuredMongoTest.NewConfig()
			logger = logTest.NewLogger()
		})

		AfterEach(func() {
			if store != nil {
				Expect(store.Terminate(context.Background())).To(Succeed())
			}
		})

		Context("NewStore", func() {
			It("returns an error when unsuccessful", func() {
				var err error
				store, err = dataRawStoreStructuredMongo.NewStore(nil)
				errorsTest.ExpectEqual(err, errors.New("database config is empty"))
				Expect(store).To(BeNil())
			})

			It("returns a new store when successful", func() {
				var err error
				store, err = dataRawStoreStructuredMongo.NewStore(config)
				Expect(err).ToNot(HaveOccurred())
				Expect(store).ToNot(BeNil())
			})
		})

		Context("with a new store", func() {
			var mongoCollection *mongo.Collection

			BeforeEach(func() {
				var err error
				store, err = dataRawStoreStructuredMongo.NewStore(config)
				Expect(err).ToNot(HaveOccurred())
				Expect(store).ToNot(BeNil())
				mongoCollection = store.GetCollection("raw")
			})

			Context("EnsureIndexes", func() {
				It("returns successfully and creates the expected index", func() {
					Expect(store.EnsureIndexes()).To(Succeed())
					cursor, err := mongoCollection.Indexes().List(context.Background())
					Expect(err).ToNot(HaveOccurred())
					Expect(cursor).ToNot(BeNil())
					var indexes []storeStructuredMongoTest.MongoIndex
					Expect(cursor.All(context.Background(), &indexes)).To(Succeed())
					Expect(indexes).To(ConsistOf(
						MatchFields(IgnoreExtras, Fields{
							"Key": Equal(storeStructuredMongoTest.MakeKeySlice("_id")),
						}),
						MatchFields(IgnoreExtras, Fields{
							"Key":  Equal(storeStructuredMongoTest.MakeKeySlice("userId", "dataSetId", "createdTime")),
							"Name": Equal("UserIdDataSetIdCreatedTime"),
						}),
					))
				})
			})

			Context("with store, ctx, and user", func() {
				var ctx context.Context
				var userID string

				BeforeEach(func() {
					Expect(store.EnsureIndexes()).To(Succeed())
					ctx = log.NewContextWithLogger(context.Background(), logger)
					userID = userTest.RandomUserID()
				})

				Context("List", func() {
					It("returns an error when the user id is missing", func() {
						result, err := store.List(ctx, "", nil, nil)
						errorsTest.ExpectEqual(err, errors.New("user id is missing"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the user id is invalid", func() {
						result, err := store.List(ctx, "invalid", nil, nil)
						errorsTest.ExpectEqual(err, errors.New("user id is invalid"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the filter is invalid", func() {
						filter := dataRawTest.RandomFilter(test.RandomOptionals())
						filter.DataSetID = pointer.From("")
						result, err := store.List(ctx, userID, filter, nil)
						errorsTest.ExpectEqual(err, errors.New("filter is invalid"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the pagination is invalid", func() {
						pagination := pageTest.RandomPagination()
						pagination.Page = -1
						result, err := store.List(ctx, userID, nil, pagination)
						errorsTest.ExpectEqual(err, errors.New("pagination is invalid"))
						Expect(result).To(BeNil())
					})

					Context("with data", func() {
						var now time.Time
						var otherUserID string
						var dataSetID string
						var userDocs dataRawStoreStructuredMongo.Documents
						var allDocs dataRawStoreStructuredMongo.Documents
						var sortedUserDocs dataRawStoreStructuredMongo.Documents

						BeforeEach(func() {
							now = time.Now().UTC()

							past := now.Add(-2 * time.Hour)
							future := now.Add(2 * time.Hour)

							otherUserID = userTest.RandomUserID()
							dataSetID = dataTest.RandomDataSetID()
							otherDataSetID := dataTest.RandomDataSetID()

							// document0: target user, target dataSetID, no processedTime, archivableTime in future, no archivedTime
							document0 := dataRawStoreStructuredMongoTest.RandomDocumentWithUserIDAndDataSetID(userID, dataSetID, test.RandomOptionals())
							document0.CreatedTime = test.RandomTimeBefore(past)
							document0.ProcessedTime = nil
							document0.ArchivableTime = &future
							document0.ArchivedTime = nil

							// document1: target user, target dataSetID, processedTime set, archivableTime in past, no archivedTime
							document1 := dataRawStoreStructuredMongoTest.RandomDocumentWithUserIDAndDataSetID(userID, dataSetID, test.RandomOptionals())
							document1.CreatedTime = test.RandomTimeBefore(document0.CreatedTime.Add(-24 * time.Hour))
							document1.ProcessedTime = &past
							document1.ArchivableTime = &past
							document1.ArchivedTime = nil

							// document2: target user, different dataSetID, processedTime set, no archivableTime, archivedTime set
							document2 := dataRawStoreStructuredMongoTest.RandomDocumentWithUserIDAndDataSetID(userID, otherDataSetID, test.RandomOptionals())
							document2.CreatedTime = test.RandomTimeBefore(document1.CreatedTime.Add(-24 * time.Hour))
							document2.ProcessedTime = &past
							document2.ArchivableTime = nil
							document2.ArchivedTime = &past

							// document3: target user, target dataSetID, no processedTime, no archivableTime, no archivedTime
							document3 := dataRawStoreStructuredMongoTest.RandomDocumentWithUserIDAndDataSetID(userID, dataSetID, test.RandomOptionals())
							document3.CreatedTime = test.RandomTimeBefore(document2.CreatedTime.Add(-24 * time.Hour))
							document3.ProcessedTime = nil
							document3.ArchivableTime = nil
							document3.ArchivedTime = nil

							// document4, document5: different user
							document4 := dataRawStoreStructuredMongoTest.RandomDocumentWithUserIDAndDataSetID(otherUserID, dataSetID, test.RandomOptionals())
							document5 := dataRawStoreStructuredMongoTest.RandomDocumentWithUserIDAndDataSetID(otherUserID, dataTest.RandomDataSetID(), test.RandomOptionals())

							userDocs = []*dataRawStoreStructuredMongo.Document{document0, document1, document2, document3}
							allDocs = append(userDocs, document4, document5)

							sortedUserDocs = slices.Clone(userDocs)
							slices.SortFunc(sortedUserDocs, func(a, b *dataRawStoreStructuredMongo.Document) int { return a.CreatedTime.Compare(b.CreatedTime) })

							_, err := mongoCollection.InsertMany(context.Background(), test.AsAnyArray(allDocs))
							Expect(err).ToNot(HaveOccurred())
						})

						It("returns empty when the user id is unknown", func() {
							result, err := store.List(ctx, userTest.RandomUserID(), nil, nil)
							Expect(err).ToNot(HaveOccurred())
							Expect(result).To(BeNil())
						})

						It("returns all user records when filter is nil", func() {
							result, err := store.List(ctx, userID, nil, nil)
							Expect(err).ToNot(HaveOccurred())
							Expect(result).To(HaveLen(4))
						})

						It("returns all user records when filter is empty", func() {
							result, err := store.List(ctx, userID, &dataRaw.Filter{}, nil)
							Expect(err).ToNot(HaveOccurred())
							Expect(result).To(HaveLen(4))
						})

						It("returns only records matching the created date", func() {
							createdDate := userDocs[0].CreatedTime.Format(dataRawStoreStructuredMongo.IDDateFormat)
							filter := &dataRaw.Filter{CreatedDate: pointer.From(createdDate)}
							result, err := store.List(ctx, userID, filter, nil)
							Expect(err).ToNot(HaveOccurred())
							Expect(result).To(HaveLen(1))
							for _, r := range result {
								Expect(r.CreatedTime.Format(dataRawStoreStructuredMongo.IDDateFormat)).To(Equal(createdDate))
							}
						})

						It("returns only records matching the dataSetID filter", func() {
							filter := &dataRaw.Filter{DataSetID: pointer.From(dataSetID)}
							result, err := store.List(ctx, userID, filter, nil)
							Expect(err).ToNot(HaveOccurred())
							Expect(result).To(HaveLen(3))
							for _, r := range result {
								Expect(r.DataSetID).To(Equal(dataSetID))
							}
						})

						It("returns only processed records when processed is true", func() {
							filter := &dataRaw.Filter{Processed: pointer.From(true)}
							result, err := store.List(ctx, userID, filter, nil)
							Expect(err).ToNot(HaveOccurred())
							Expect(result).To(HaveLen(2))
							for _, r := range result {
								Expect(r.ProcessedTime).ToNot(BeNil())
							}
						})

						It("returns only unprocessed records when processed is false", func() {
							filter := &dataRaw.Filter{Processed: pointer.From(false)}
							result, err := store.List(ctx, userID, filter, nil)
							Expect(err).ToNot(HaveOccurred())
							Expect(result).To(HaveLen(2))
							for _, r := range result {
								Expect(r.ProcessedTime).To(BeNil())
							}
						})

						It("returns only archivable records when archivable is true", func() {
							filter := &dataRaw.Filter{Archivable: pointer.From(true)}
							result, err := store.List(ctx, userID, filter, nil)
							Expect(err).ToNot(HaveOccurred())
							Expect(result).To(HaveLen(1))
							for _, r := range result {
								Expect(r.ArchivableTime).ToNot(BeNil())
								Expect(r.ArchivableTime.Before(now)).To(BeTrue())
							}
						})

						It("returns only archivable records when archivable is false", func() {
							filter := &dataRaw.Filter{Archivable: pointer.From(false)}
							result, err := store.List(ctx, userID, filter, nil)
							Expect(err).ToNot(HaveOccurred())
							Expect(result).To(HaveLen(3))
							for _, r := range result {
								if r.ArchivableTime != nil {
									Expect(r.ArchivableTime.Before(now)).To(BeFalse())
								}
							}
						})

						It("returns only archivable records when archivable is true", func() {
							filter := &dataRaw.Filter{Archivable: pointer.From(true)}
							result, err := store.List(ctx, userID, filter, nil)
							Expect(err).ToNot(HaveOccurred())
							Expect(result).To(HaveLen(1))
							for _, r := range result {
								Expect(r.ArchivableTime).ToNot(BeNil())
								Expect(r.ArchivableTime.Before(now)).To(BeTrue())
							}
						})

						It("returns only archived records when archived is true", func() {
							filter := &dataRaw.Filter{Archived: pointer.From(true)}
							result, err := store.List(ctx, userID, filter, nil)
							Expect(err).ToNot(HaveOccurred())
							Expect(result).To(HaveLen(1))
							for _, r := range result {
								Expect(r.ArchivedTime).ToNot(BeNil())
							}
						})

						It("returns only non-archived records when archived is false", func() {
							filter := &dataRaw.Filter{Archived: pointer.From(false)}
							result, err := store.List(ctx, userID, filter, nil)
							Expect(err).ToNot(HaveOccurred())
							Expect(result).To(HaveLen(3))
							for _, r := range result {
								Expect(r.ArchivedTime).To(BeNil())
							}
						})

						It("returns results ordered by createdTime ascending", func() {
							result, err := store.List(ctx, userID, nil, nil)
							Expect(err).ToNot(HaveOccurred())
							Expect(result).To(HaveLen(len(userDocs)))
							for i := 1; i < len(result); i++ {
								Expect(result[i-1].CreatedTime.Before(result[i].CreatedTime) || result[i-1].CreatedTime.Equal(result[i].CreatedTime)).To(BeTrue())
							}
						})

						It("returns paginated results", func() {
							pagination := page.NewPagination()
							pagination.Page = 1
							pagination.Size = 2
							result, err := store.List(ctx, userID, nil, pagination)
							Expect(err).ToNot(HaveOccurred())
							Expect(result).To(HaveLen(2))

							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": &dataRaw.Filter{}, "pagination": pagination, "count": 2})
						})
					})
				})

				Context("Create", func() {
					var dataSetID string
					var create *dataRaw.Create

					BeforeEach(func() {
						dataSetID = dataTest.RandomDataSetID()
						create = dataRawTest.RandomCreate(test.RandomOptionals())
						create.DigestMD5 = nil
						create.DigestSHA256 = nil
					})

					It("returns an error when the user id is missing", func() {
						result, err := store.Create(ctx, "", dataSetID, create, bytes.NewReader(nil))
						errorsTest.ExpectEqual(err, errors.New("user id is missing"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the user id is invalid", func() {
						result, err := store.Create(ctx, "invalid", dataSetID, create, bytes.NewReader(nil))
						errorsTest.ExpectEqual(err, errors.New("user id is invalid"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the data set id is missing", func() {
						result, err := store.Create(ctx, userID, "", create, bytes.NewReader(nil))
						errorsTest.ExpectEqual(err, errors.New("data set id is missing"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the data set id is invalid", func() {
						result, err := store.Create(ctx, userID, "invalid", create, bytes.NewReader(nil))
						errorsTest.ExpectEqual(err, errors.New("data set id is invalid"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the create is missing", func() {
						result, err := store.Create(ctx, userID, dataSetID, nil, bytes.NewReader(nil))
						errorsTest.ExpectEqual(err, errors.New("create is missing"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the create is invalid", func() {
						invalidCreate := &dataRaw.Create{MediaType: pointer.From("")}
						result, err := store.Create(ctx, userID, dataSetID, invalidCreate, bytes.NewReader(nil))
						errorsTest.ExpectEqual(err, errors.New("create is invalid"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the reader is missing", func() {
						result, err := store.Create(ctx, userID, dataSetID, create, nil)
						errorsTest.ExpectEqual(err, errors.New("reader is missing"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the reader returns an error", func() {
						testErr := errorsTest.RandomError()
						result, err := store.Create(ctx, userID, dataSetID, create, test.ErrorReader(testErr))
						errorsTest.ExpectEqual(err, errors.New("unable to read data"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the reader returns too much data", func() {
						result, err := store.Create(ctx, userID, dataSetID, create, test.ZeroReader())
						errorsTest.ExpectEqual(err, errors.New("data size exceeds maximum"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the reader returns too much stored data", func() {
						result, err := store.Create(ctx, userID, dataSetID, create, io.LimitReader(test.RandReader(), dataRaw.SizeStoredMaximum+1))
						errorsTest.ExpectEqual(err, errors.New("data size stored exceeds maximum"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the MD5 digest does not match", func() {
						create.DigestMD5 = pointer.From(crypto.Base64EncodedMD5Hash(test.RandomBytes()))
						result, err := store.Create(ctx, userID, dataSetID, create, bytes.NewReader(test.RandomBytes()))
						errorsTest.ExpectEqual(err, errors.New("calculated MD5 digest does not match expected"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the SHA256 digest does not match", func() {
						create.DigestSHA256 = pointer.From(crypto.Base64EncodedSHA256Hash(test.RandomBytes()))
						result, err := store.Create(ctx, userID, dataSetID, create, bytes.NewReader(test.RandomBytes()))
						errorsTest.ExpectEqual(err, errors.New("calculated SHA256 digest does not match expected"))
						Expect(result).To(BeNil())
					})

					It("returns the result after creating without digests", func() {
						data := test.RandomBytes()
						result, err := store.Create(ctx, userID, dataSetID, create, bytes.NewReader(data))
						Expect(err).ToNot(HaveOccurred())
						Expect(result).To(PointTo(MatchAllFields(Fields{
							"ID":             Not(BeEmpty()),
							"UserID":         Equal(userID),
							"DataSetID":      Equal(dataSetID),
							"Metadata":       Equal(create.Metadata),
							"DigestMD5":      Equal(crypto.Base64EncodedMD5Hash(data)),
							"DigestSHA256":   PointTo(Equal(crypto.Base64EncodedSHA256Hash(data))),
							"MediaType":      Equal(pointer.Default(create.MediaType, dataRaw.MediaTypeDefault)),
							"Size":           Equal(len(data)),
							"ProcessedTime":  BeNil(),
							"ArchivableTime": Equal(create.ArchivableTime),
							"ArchivedTime":   BeNil(),
							"CreatedTime":    BeTemporally("~", time.Now(), time.Second),
							"ModifiedTime":   BeNil(),
							"Revision":       Equal(1),
						})))

						objectID, _, err := dataRawStoreStructuredMongo.ObjectIDAndDateFromID(result.ID)
						Expect(err).ToNot(HaveOccurred())

						var storedDocuments dataRawStoreStructuredMongo.Documents
						cursor, err := mongoCollection.Find(context.Background(), bson.M{"_id": objectID})
						Expect(err).ToNot(HaveOccurred())
						Expect(cursor.All(context.Background(), &storedDocuments)).To(Succeed())
						Expect(storedDocuments).To(HaveLen(1))
					})

					It("returns the result after creating with matching digests", func() {
						data := test.RandomBytes()
						create.DigestMD5 = pointer.From(crypto.Base64EncodedMD5Hash(data))
						create.DigestSHA256 = pointer.From(crypto.Base64EncodedSHA256Hash(data))
						result, err := store.Create(ctx, userID, dataSetID, create, bytes.NewReader(data))
						Expect(err).ToNot(HaveOccurred())
						Expect(result).ToNot(BeNil())
						Expect(result.DigestMD5).To(Equal(*create.DigestMD5))
						Expect(result.DigestSHA256).To(Equal(create.DigestSHA256))
					})

					It("stores original data when there is no data", func() {
						result, err := store.Create(ctx, userID, dataSetID, create, bytes.NewReader(nil))
						Expect(err).ToNot(HaveOccurred())
						Expect(result).ToNot(BeNil())

						objectID, _, err := dataRawStoreStructuredMongo.ObjectIDAndDateFromID(result.ID)
						Expect(err).ToNot(HaveOccurred())
						var storedDocument dataRawStoreStructuredMongo.Document
						Expect(mongoCollection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&storedDocument)).To(Succeed())
						Expect(storedDocument.Compressed).To(BeFalse())
					})

					It("stores original data when data is to random to compress", func() {
						data := test.RandomBytes()
						result, err := store.Create(ctx, userID, dataSetID, create, bytes.NewReader(data))
						Expect(err).ToNot(HaveOccurred())
						Expect(result).ToNot(BeNil())

						objectID, _, err := dataRawStoreStructuredMongo.ObjectIDAndDateFromID(result.ID)
						Expect(err).ToNot(HaveOccurred())
						var storedDocument dataRawStoreStructuredMongo.Document
						Expect(mongoCollection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&storedDocument)).To(Succeed())
						Expect(storedDocument.Compressed).To(BeFalse())
					})

					It("stores compressed data when data is large enough to benefit from compression", func() {
						data := slices.Repeat(test.RandomBytesFromRange(1, 10), 100)
						result, err := store.Create(ctx, userID, dataSetID, create, bytes.NewReader(data))
						Expect(err).ToNot(HaveOccurred())
						Expect(result).ToNot(BeNil())

						objectID, _, err := dataRawStoreStructuredMongo.ObjectIDAndDateFromID(result.ID)
						Expect(err).ToNot(HaveOccurred())
						var storedDocument dataRawStoreStructuredMongo.Document
						Expect(mongoCollection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&storedDocument)).To(Succeed())
						Expect(storedDocument.Compressed).To(BeTrue())
					})
				})

				Context("Get", func() {
					var id string
					var condition *storeStructured.Condition

					BeforeEach(func() {
						id = dataRawTest.RandomDataRawID()
						condition = storeStructuredTest.RandomCondition()
					})

					It("returns an error when the id is missing", func() {
						result, err := store.Get(ctx, "", condition)
						errorsTest.ExpectEqual(err, errors.New("id is missing"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the id is invalid", func() {
						result, err := store.Get(ctx, "invalid", condition)
						errorsTest.ExpectEqual(err, errors.New("id is invalid"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the condition is invalid", func() {
						condition.Revision = pointer.From(-1)
						result, err := store.Get(ctx, id, condition)
						errorsTest.ExpectEqual(err, errors.New("condition is invalid"))
						Expect(result).To(BeNil())
					})

					Context("with data", func() {
						var document *dataRawStoreStructuredMongo.Document

						BeforeEach(func() {
							document = dataRawStoreStructuredMongoTest.RandomDocumentWithUserIDAndDataSetID(userID, dataTest.RandomDataSetID(), test.RandomOptionals())
							id = dataRawStoreStructuredMongo.IDFromObjectIDAndDate(document.ID, document.CreatedTime)
							condition = &storeStructured.Condition{Revision: pointer.From(document.Revision)}

							otherDocuments := []*dataRawStoreStructuredMongo.Document{
								dataRawStoreStructuredMongoTest.RandomDocumentWithUserIDAndDataSetID(userID, dataTest.RandomDataSetID(), test.RandomOptionals()),
								dataRawStoreStructuredMongoTest.RandomDocumentWithUserIDAndDataSetID(userID, dataTest.RandomDataSetID(), test.RandomOptionals()),
							}
							_, err := mongoCollection.InsertMany(context.Background(), test.AsAnyArray(append(otherDocuments, document)))
							Expect(err).ToNot(HaveOccurred())
						})

						AfterEach(func() {
							if condition != nil {
								logger.AssertDebug("Get", log.Fields{"id": id, "condition": condition})
							} else {
								logger.AssertDebug("Get", log.Fields{"id": id})
							}
						})

						It("returns nil when the id does not match", func() {
							id = dataRawTest.RandomDataRawID()

							result, err := store.Get(ctx, id, condition)
							Expect(err).ToNot(HaveOccurred())
							Expect(result).To(BeNil())
						})

						It("returns nil when the condition revision does not match", func() {
							condition = &storeStructured.Condition{Revision: pointer.From(document.Revision + 1)}

							result, err := store.Get(ctx, id, condition)
							Expect(err).ToNot(HaveOccurred())
							Expect(result).To(BeNil())
						})

						It("returns the result when the id exists and condition is nil", func() {
							condition = nil

							result, err := store.Get(ctx, id, condition)
							Expect(err).ToNot(HaveOccurred())
							Expect(result).ToNot(BeNil())
							Expect(result.ID).To(Equal(id))
							Expect(result.UserID).To(Equal(document.UserID))
						})

						It("returns the result when the id condition matches", func() {
							result, err := store.Get(ctx, id, condition)
							Expect(err).ToNot(HaveOccurred())
							Expect(result).ToNot(BeNil())
							Expect(result.ID).To(Equal(id))
						})
					})
				})

				Context("GetContent", func() {
					var id string
					var condition *storeStructured.Condition

					BeforeEach(func() {
						id = dataRawTest.RandomDataRawID()
						condition = storeStructuredTest.RandomCondition()
					})

					It("returns an error when the id is missing", func() {
						result, err := store.GetContent(ctx, "", condition)
						errorsTest.ExpectEqual(err, errors.New("id is missing"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the id is invalid", func() {
						result, err := store.GetContent(ctx, "invalid", condition)
						errorsTest.ExpectEqual(err, errors.New("id is invalid"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the condition is invalid", func() {
						condition.Revision = pointer.From(-1)
						result, err := store.GetContent(ctx, id, condition)
						errorsTest.ExpectEqual(err, errors.New("condition is invalid"))
						Expect(result).To(BeNil())
					})

					Context("with data", func() {
						var document *dataRawStoreStructuredMongo.Document
						var originalData []byte

						BeforeEach(func() {
							document = dataRawStoreStructuredMongoTest.RandomDocumentWithCompressed(false, test.RandomOptionals())
							id = dataRawStoreStructuredMongo.IDFromObjectIDAndDate(document.ID, document.CreatedTime)
							condition = &storeStructured.Condition{Revision: pointer.From(document.Revision)}
							originalData = document.Data.Data

							otherDocuments := []*dataRawStoreStructuredMongo.Document{
								dataRawStoreStructuredMongoTest.RandomDocumentWithUserIDAndDataSetID(userID, dataTest.RandomDataSetID(), test.RandomOptionals()),
								dataRawStoreStructuredMongoTest.RandomDocumentWithUserIDAndDataSetID(userID, dataTest.RandomDataSetID(), test.RandomOptionals()),
							}
							_, err := mongoCollection.InsertMany(context.Background(), test.AsAnyArray(append(otherDocuments, document)))
							Expect(err).ToNot(HaveOccurred())
						})

						AfterEach(func() {
							if condition != nil {
								logger.AssertDebug("GetContent", log.Fields{"id": id, "condition": condition})
							} else {
								logger.AssertDebug("GetContent", log.Fields{"id": id})
							}
						})

						It("returns nil when the id does not exist", func() {
							id = dataRawTest.RandomDataRawID()

							result, err := store.GetContent(ctx, id, condition)
							Expect(err).ToNot(HaveOccurred())
							Expect(result).To(BeNil())
						})

						It("returns content when the id exists and condition is nil", func() {
							condition = nil

							result, err := store.GetContent(ctx, id, condition)
							Expect(err).ToNot(HaveOccurred())
							Expect(result).ToNot(BeNil())
							Expect(result.DigestMD5).To(Equal(document.DigestMD5))
							Expect(result.MediaType).To(Equal(document.MediaType))
							Expect(result.ReadCloser).ToNot(BeNil())

							data, readErr := io.ReadAll(result.ReadCloser)
							Expect(readErr).ToNot(HaveOccurred())
							Expect(data).To(Equal(originalData))
						})

						It("returns content when the id and the condition match", func() {
							result, err := store.GetContent(ctx, id, condition)
							Expect(err).ToNot(HaveOccurred())
							Expect(result).ToNot(BeNil())
							Expect(result.DigestMD5).To(Equal(document.DigestMD5))
							Expect(result.MediaType).To(Equal(document.MediaType))
							Expect(result.ReadCloser).ToNot(BeNil())

							data, readErr := io.ReadAll(result.ReadCloser)
							Expect(readErr).ToNot(HaveOccurred())
							Expect(data).To(Equal(originalData))
						})

						Context("when the document is compressed", func() {
							BeforeEach(func() {
								document.Compressed = true
								document.Data = bsonPrimitive.Binary{Data: compressTest.Compress(originalData)}
							})

							It("returns content when the id exists and condition is nil", func() {
								condition = nil

								result, err := store.GetContent(ctx, id, condition)
								Expect(err).ToNot(HaveOccurred())
								Expect(result).ToNot(BeNil())
								Expect(result.DigestMD5).To(Equal(document.DigestMD5))
								Expect(result.MediaType).To(Equal(document.MediaType))
								Expect(result.ReadCloser).ToNot(BeNil())

								data, readErr := io.ReadAll(result.ReadCloser)
								Expect(readErr).ToNot(HaveOccurred())
								Expect(data).To(Equal(originalData))
							})

							It("returns content when the id and the condition match", func() {
								result, err := store.GetContent(ctx, id, condition)
								Expect(err).ToNot(HaveOccurred())
								Expect(result).ToNot(BeNil())
								Expect(result.DigestMD5).To(Equal(document.DigestMD5))
								Expect(result.MediaType).To(Equal(document.MediaType))
								Expect(result.ReadCloser).ToNot(BeNil())

								data, readErr := io.ReadAll(result.ReadCloser)
								Expect(readErr).ToNot(HaveOccurred())
								Expect(data).To(Equal(originalData))
							})
						})
					})
				})

				Context("Update", func() {
					var id string
					var condition *storeStructured.Condition
					var update *dataRaw.Update

					BeforeEach(func() {
						id = dataRawTest.RandomDataRawID()
						condition = storeStructuredTest.RandomCondition()
						update = dataRawTest.RandomUpdate(test.RandomOptionals())
					})

					It("returns an error when the id is missing", func() {
						result, err := store.Update(ctx, "", condition, update)
						errorsTest.ExpectEqual(err, errors.New("id is missing"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the id is invalid", func() {
						result, err := store.Update(ctx, "invalid", condition, update)
						errorsTest.ExpectEqual(err, errors.New("id is invalid"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the condition is invalid", func() {
						condition.Revision = pointer.From(-1)
						result, err := store.Update(ctx, id, condition, update)
						errorsTest.ExpectEqual(err, errors.New("condition is invalid"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the update is missing", func() {
						result, err := store.Update(ctx, id, condition, nil)
						errorsTest.ExpectEqual(err, errors.New("update is missing"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the update is invalid", func() {
						update.ProcessedTime = pointer.From(time.Time{})
						result, err := store.Update(ctx, id, condition, update)
						errorsTest.ExpectEqual(err, errors.New("update is invalid"))
						Expect(result).To(BeNil())
					})

					Context("with data", func() {
						var document *dataRawStoreStructuredMongo.Document

						BeforeEach(func() {
							document = dataRawStoreStructuredMongoTest.RandomDocumentWithUserIDAndDataSetID(userID, dataTest.RandomDataSetID(), test.RandomOptionals())
							id = dataRawStoreStructuredMongo.IDFromObjectIDAndDate(document.ID, document.CreatedTime)
							condition = &storeStructured.Condition{Revision: pointer.From(document.Revision)}

							otherDocuments := []*dataRawStoreStructuredMongo.Document{
								dataRawStoreStructuredMongoTest.RandomDocumentWithUserIDAndDataSetID(userID, dataTest.RandomDataSetID(), test.RandomOptionals()),
								dataRawStoreStructuredMongoTest.RandomDocumentWithUserIDAndDataSetID(userID, dataTest.RandomDataSetID(), test.RandomOptionals()),
							}
							_, err := mongoCollection.InsertMany(context.Background(), test.AsAnyArray(append(otherDocuments, document)))
							Expect(err).ToNot(HaveOccurred())
						})

						AfterEach(func() {
							if condition != nil {
								logger.AssertDebug("Update", log.Fields{"id": id, "condition": condition, "update": update})
							} else {
								logger.AssertDebug("Update", log.Fields{"id": id, "update": update})
							}
						})

						It("returns nil when the id does not exist", func() {
							id = dataRawTest.RandomDataRawID()

							result, err := store.Update(ctx, id, condition, update)
							Expect(err).ToNot(HaveOccurred())
							Expect(result).To(BeNil())
						})

						It("returns nil when the condition revision does not match", func() {
							condition = &storeStructured.Condition{Revision: pointer.From(document.Revision + 1)}

							result, err := store.Update(ctx, id, condition, update)
							Expect(err).ToNot(HaveOccurred())
							Expect(result).To(BeNil())
						})

						It("returns updated result when condition is nil", func() {
							condition = nil

							result, err := store.Update(ctx, id, condition, update)
							Expect(err).ToNot(HaveOccurred())
							Expect(result).ToNot(BeNil())
							Expect(result.ProcessedTime).To(Equal(pointer.DefaultPointer(update.ProcessedTime, document.ProcessedTime)))
							Expect(result.ArchivableTime).To(Equal(pointer.DefaultPointer(update.ArchivableTime, document.ArchivableTime)))
							Expect(result.ArchivedTime).To(Equal(pointer.DefaultPointer(update.ArchivedTime, document.ArchivedTime)))
							if update.Metadata != nil {
								Expect(result.Metadata).To(Equal(*update.Metadata))
							} else {
								Expect(result.Metadata).To(Equal(document.Metadata))
							}
							Expect(result.ModifiedTime).To(PointTo(BeTemporally("~", time.Now(), time.Second)))
						})

						It("returns updated result when condition revision matches", func() {
							result, err := store.Update(ctx, id, condition, update)
							Expect(err).ToNot(HaveOccurred())
							Expect(result).ToNot(BeNil())
							Expect(result.ProcessedTime).To(Equal(pointer.DefaultPointer(update.ProcessedTime, document.ProcessedTime)))
							Expect(result.ArchivableTime).To(Equal(pointer.DefaultPointer(update.ArchivableTime, document.ArchivableTime)))
							Expect(result.ArchivedTime).To(Equal(pointer.DefaultPointer(update.ArchivedTime, document.ArchivedTime)))
							if update.Metadata != nil {
								Expect(result.Metadata).To(Equal(*update.Metadata))
							} else {
								Expect(result.Metadata).To(Equal(document.Metadata))
							}
							Expect(result.ModifiedTime).To(PointTo(BeTemporally("~", time.Now(), time.Second)))
						})
					})
				})

				Context("Delete", func() {
					var id string
					var condition *storeStructured.Condition

					BeforeEach(func() {
						id = dataRawTest.RandomDataRawID()
						condition = storeStructuredTest.RandomCondition()
					})

					It("returns an error when the id is missing", func() {
						result, err := store.Delete(ctx, "", condition)
						errorsTest.ExpectEqual(err, errors.New("id is missing"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the id is invalid", func() {
						result, err := store.Delete(ctx, "invalid", condition)
						errorsTest.ExpectEqual(err, errors.New("id is invalid"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the condition is invalid", func() {
						condition.Revision = pointer.From(-1)
						result, err := store.Delete(ctx, id, condition)
						errorsTest.ExpectEqual(err, errors.New("condition is invalid"))
						Expect(result).To(BeNil())
					})

					Context("with data", func() {
						var document *dataRawStoreStructuredMongo.Document

						BeforeEach(func() {
							document = dataRawStoreStructuredMongoTest.RandomDocumentWithUserIDAndDataSetID(userID, dataTest.RandomDataSetID(), test.RandomOptionals())
							id = dataRawStoreStructuredMongo.IDFromObjectIDAndDate(document.ID, document.CreatedTime)
							condition = &storeStructured.Condition{Revision: pointer.From(document.Revision)}

							otherDocuments := []*dataRawStoreStructuredMongo.Document{
								dataRawStoreStructuredMongoTest.RandomDocumentWithUserIDAndDataSetID(userID, dataTest.RandomDataSetID(), test.RandomOptionals()),
								dataRawStoreStructuredMongoTest.RandomDocumentWithUserIDAndDataSetID(userID, dataTest.RandomDataSetID(), test.RandomOptionals()),
							}
							_, err := mongoCollection.InsertMany(context.Background(), test.AsAnyArray(append(otherDocuments, document)))
							Expect(err).ToNot(HaveOccurred())
						})

						AfterEach(func() {
							if condition != nil {
								logger.AssertDebug("Delete", log.Fields{"id": id, "condition": condition})
							} else {
								logger.AssertDebug("Delete", log.Fields{"id": id})
							}
						})

						It("returns nil when the id does not exist", func() {
							id = dataRawTest.RandomDataRawID()

							result, err := store.Delete(ctx, id, condition)
							Expect(err).ToNot(HaveOccurred())
							Expect(result).To(BeNil())

							count, err := mongoCollection.CountDocuments(context.Background(), bson.M{"_id": document.ID})
							Expect(err).ToNot(HaveOccurred())
							Expect(count).To(Equal(int64(1)))

							// Other documents not affected
							total, err := mongoCollection.CountDocuments(context.Background(), bson.M{})
							Expect(err).ToNot(HaveOccurred())
							Expect(total).To(Equal(int64(3)))
						})

						It("returns nil when the condition revision does not match", func() {
							condition = &storeStructured.Condition{Revision: pointer.From(document.Revision + 1)}

							result, err := store.Delete(ctx, id, condition)
							Expect(err).ToNot(HaveOccurred())
							Expect(result).To(BeNil())

							count, err := mongoCollection.CountDocuments(context.Background(), bson.M{"_id": document.ID})
							Expect(err).ToNot(HaveOccurred())
							Expect(count).To(Equal(int64(1)))

							// Other documents not affected
							total, err := mongoCollection.CountDocuments(context.Background(), bson.M{})
							Expect(err).ToNot(HaveOccurred())
							Expect(total).To(Equal(int64(3)))
						})

						It("returns the deleted document when condition is nil", func() {
							condition = nil

							result, err := store.Delete(ctx, id, condition)
							Expect(err).ToNot(HaveOccurred())
							Expect(result).ToNot(BeNil())
							Expect(result.ID).To(Equal(id))

							count, err := mongoCollection.CountDocuments(context.Background(), bson.M{"_id": document.ID})
							Expect(err).ToNot(HaveOccurred())
							Expect(count).To(BeZero())

							// Other documents not affected
							total, err := mongoCollection.CountDocuments(context.Background(), bson.M{})
							Expect(err).ToNot(HaveOccurred())
							Expect(total).To(Equal(int64(2)))
						})

						It("returns the deleted document when condition revision matches", func() {
							result, err := store.Delete(ctx, id, condition)
							Expect(err).ToNot(HaveOccurred())
							Expect(result).ToNot(BeNil())
							Expect(result.ID).To(Equal(id))

							count, err := mongoCollection.CountDocuments(context.Background(), bson.M{"_id": document.ID})
							Expect(err).ToNot(HaveOccurred())
							Expect(count).To(BeZero())

							// Other documents not affected
							total, err := mongoCollection.CountDocuments(context.Background(), bson.M{})
							Expect(err).ToNot(HaveOccurred())
							Expect(total).To(Equal(int64(2)))
						})
					})
				})

				Context("DeleteMultiple", func() {
					It("returns an error when any id is invalid", func() {
						count, err := store.DeleteMultiple(ctx, []string{dataRawTest.RandomDataRawID(), "invalid-id"})
						errorsTest.ExpectEqual(err, errors.New("id is invalid"))
						Expect(count).To(BeZero())
					})

					Context("with data", func() {
						var documents []*dataRawStoreStructuredMongo.Document

						BeforeEach(func() {
							documents = []*dataRawStoreStructuredMongo.Document{
								dataRawStoreStructuredMongoTest.RandomDocumentWithUserIDAndDataSetID(userID, dataTest.RandomDataSetID(), test.RandomOptionals()),
								dataRawStoreStructuredMongoTest.RandomDocumentWithUserIDAndDataSetID(userID, dataTest.RandomDataSetID(), test.RandomOptionals()),
								dataRawStoreStructuredMongoTest.RandomDocumentWithUserIDAndDataSetID(userID, dataTest.RandomDataSetID(), test.RandomOptionals()),
							}
							_, err := mongoCollection.InsertMany(context.Background(), test.AsAnyArray(documents))
							Expect(err).ToNot(HaveOccurred())
						})

						It("returns zero when ids is nil", func() {
							count, err := store.DeleteMultiple(ctx, nil)
							Expect(err).ToNot(HaveOccurred())
							Expect(count).To(BeZero())

							// Other documents not affected
							total, err := mongoCollection.CountDocuments(context.Background(), bson.M{})
							Expect(err).ToNot(HaveOccurred())
							Expect(total).To(Equal(int64(3)))
						})

						It("returns zero when ids is nil", func() {
							count, err := store.DeleteMultiple(ctx, []string{})
							Expect(err).ToNot(HaveOccurred())
							Expect(count).To(BeZero())

							// Other documents not affected
							total, err := mongoCollection.CountDocuments(context.Background(), bson.M{})
							Expect(err).ToNot(HaveOccurred())
							Expect(total).To(Equal(int64(3)))
						})

						It("returns zero when none of the ids exist", func() {
							ids := []string{dataRawTest.RandomDataRawID(), dataRawTest.RandomDataRawID()}
							count, err := store.DeleteMultiple(ctx, ids)
							Expect(err).ToNot(HaveOccurred())
							Expect(count).To(BeZero())

							// Other documents not affected
							total, err := mongoCollection.CountDocuments(context.Background(), bson.M{})
							Expect(err).ToNot(HaveOccurred())
							Expect(total).To(Equal(int64(3)))
						})

						It("returns the count of deleted documents and leaves others intact", func() {
							ids := []string{
								dataRawStoreStructuredMongo.IDFromObjectIDAndDate(documents[0].ID, documents[0].CreatedTime),
								dataRawStoreStructuredMongo.IDFromObjectIDAndDate(documents[2].ID, documents[2].CreatedTime),
							}

							count, err := store.DeleteMultiple(ctx, ids)
							Expect(err).ToNot(HaveOccurred())
							Expect(count).To(Equal(2))

							remaining, err := mongoCollection.CountDocuments(context.Background(), bson.M{})
							Expect(err).ToNot(HaveOccurred())
							Expect(remaining).To(Equal(int64(1)))

							logger.AssertDebug("DeleteMultiple", log.Fields{"ids": ids})
						})
					})
				})

				Context("DeleteAllByDataSetID", func() {
					var dataSetID string

					BeforeEach(func() {
						dataSetID = dataTest.RandomDataSetID()
					})

					It("returns an error when the user id is missing", func() {
						count, err := store.DeleteAllByDataSetID(ctx, "", dataSetID)
						errorsTest.ExpectEqual(err, errors.New("user id is missing"))
						Expect(count).To(BeZero())
					})

					It("returns an error when the user id is invalid", func() {
						count, err := store.DeleteAllByDataSetID(ctx, "invalid", dataSetID)
						errorsTest.ExpectEqual(err, errors.New("user id is invalid"))
						Expect(count).To(BeZero())
					})

					It("returns an error when the data set id is missing", func() {
						count, err := store.DeleteAllByDataSetID(ctx, userID, "")
						errorsTest.ExpectEqual(err, errors.New("data set id is missing"))
						Expect(count).To(BeZero())
					})

					It("returns an error when the data set id is invalid", func() {
						count, err := store.DeleteAllByDataSetID(ctx, userID, "invalid")
						errorsTest.ExpectEqual(err, errors.New("data set id is invalid"))
						Expect(count).To(BeZero())
					})

					Context("with data", func() {
						BeforeEach(func() {
							documents := []*dataRawStoreStructuredMongo.Document{
								dataRawStoreStructuredMongoTest.RandomDocumentWithUserIDAndDataSetID(userID, dataSetID, test.RandomOptionals()),
								dataRawStoreStructuredMongoTest.RandomDocumentWithUserIDAndDataSetID(userID, dataSetID, test.RandomOptionals()),
								dataRawStoreStructuredMongoTest.RandomDocumentWithUserIDAndDataSetID(userID, dataTest.RandomDataSetID(), test.RandomOptionals()),
								dataRawStoreStructuredMongoTest.RandomDocumentWithUserIDAndDataSetID(userTest.RandomUserID(), dataSetID, test.RandomOptionals()),
							}
							_, err := mongoCollection.InsertMany(context.Background(), test.AsAnyArray(documents))
							Expect(err).ToNot(HaveOccurred())
						})

						It("returns zero when no matching documents exist by user id", func() {
							anotherUserID := userTest.RandomUserID()

							count, err := store.DeleteAllByDataSetID(ctx, anotherUserID, dataSetID)
							Expect(err).ToNot(HaveOccurred())
							Expect(count).To(BeZero())

							// Other documents not affected
							total, err := mongoCollection.CountDocuments(context.Background(), bson.M{})
							Expect(err).ToNot(HaveOccurred())
							Expect(total).To(Equal(int64(4)))

							logger.AssertDebug("DeleteAllByDataSetID", log.Fields{"userId": anotherUserID, "dataSetId": dataSetID})
						})

						It("returns zero when no matching documents exist by data set id", func() {
							anotherDataSetID := dataTest.RandomDataSetID()

							count, err := store.DeleteAllByDataSetID(ctx, userID, anotherDataSetID)
							Expect(err).ToNot(HaveOccurred())
							Expect(count).To(BeZero())

							// Other documents not affected
							total, err := mongoCollection.CountDocuments(context.Background(), bson.M{})
							Expect(err).ToNot(HaveOccurred())
							Expect(total).To(Equal(int64(4)))

							logger.AssertDebug("DeleteAllByDataSetID", log.Fields{"userId": userID, "dataSetId": anotherDataSetID})
						})

						It("deletes only matching documents and returns the count", func() {
							count, err := store.DeleteAllByDataSetID(ctx, userID, dataSetID)
							Expect(err).ToNot(HaveOccurred())
							Expect(count).To(Equal(2))

							remaining, err := mongoCollection.CountDocuments(context.Background(), bson.M{"userId": userID, "dataSetId": dataSetID})
							Expect(err).ToNot(HaveOccurred())
							Expect(remaining).To(BeZero())

							// Other documents not affected
							total, err := mongoCollection.CountDocuments(context.Background(), bson.M{})
							Expect(err).ToNot(HaveOccurred())
							Expect(total).To(Equal(int64(2)))

							logger.AssertDebug("DeleteAllByDataSetID", log.Fields{"userId": userID, "dataSetId": dataSetID})
						})
					})
				})

				Context("DeleteAllByUserID", func() {
					It("returns an error when the user id is missing", func() {
						count, err := store.DeleteAllByUserID(ctx, "")
						errorsTest.ExpectEqual(err, errors.New("user id is missing"))
						Expect(count).To(BeZero())
					})

					It("returns an error when the user id is invalid", func() {
						count, err := store.DeleteAllByUserID(ctx, "invalid")
						errorsTest.ExpectEqual(err, errors.New("user id is invalid"))
						Expect(count).To(BeZero())
					})

					Context("with data", func() {
						var otherUserID string

						BeforeEach(func() {
							otherUserID = userTest.RandomUserID()
							documents := []*dataRawStoreStructuredMongo.Document{
								dataRawStoreStructuredMongoTest.RandomDocumentWithUserIDAndDataSetID(userID, dataTest.RandomDataSetID(), test.RandomOptionals()),
								dataRawStoreStructuredMongoTest.RandomDocumentWithUserIDAndDataSetID(userID, dataTest.RandomDataSetID(), test.RandomOptionals()),
								dataRawStoreStructuredMongoTest.RandomDocumentWithUserIDAndDataSetID(otherUserID, dataTest.RandomDataSetID(), test.RandomOptionals()),
								dataRawStoreStructuredMongoTest.RandomDocumentWithUserIDAndDataSetID(otherUserID, dataTest.RandomDataSetID(), test.RandomOptionals()),
							}
							_, err := mongoCollection.InsertMany(context.Background(), test.AsAnyArray(documents))
							Expect(err).ToNot(HaveOccurred())
						})

						It("returns zero when no matching documents exist", func() {
							anotherUserID := userTest.RandomUserID()

							count, err := store.DeleteAllByUserID(ctx, anotherUserID)
							Expect(err).ToNot(HaveOccurred())
							Expect(count).To(BeZero())

							remaining, err := mongoCollection.CountDocuments(context.Background(), bson.M{"userId": userID})
							Expect(err).ToNot(HaveOccurred())
							Expect(remaining).To(Equal(int64(2)))

							// Other user documents not affected
							otherRemaining, err := mongoCollection.CountDocuments(context.Background(), bson.M{})
							Expect(err).ToNot(HaveOccurred())
							Expect(otherRemaining).To(Equal(int64(4)))

							logger.AssertDebug("DeleteAllByUserID", log.Fields{"userId": anotherUserID})
						})

						It("deletes only the user documents and returns the count", func() {
							count, err := store.DeleteAllByUserID(ctx, userID)
							Expect(err).ToNot(HaveOccurred())
							Expect(count).To(Equal(2))

							remaining, err := mongoCollection.CountDocuments(context.Background(), bson.M{"userId": userID})
							Expect(err).ToNot(HaveOccurred())
							Expect(remaining).To(BeZero())

							// Other user documents not affected
							otherRemaining, err := mongoCollection.CountDocuments(context.Background(), bson.M{})
							Expect(err).ToNot(HaveOccurred())
							Expect(otherRemaining).To(Equal(int64(2)))

							logger.AssertDebug("DeleteAllByUserID", log.Fields{"userId": userID})
						})
					})
				})
			})
		})
	})

	Context("Document", func() {
		var document *dataRawStoreStructuredMongo.Document

		BeforeEach(func() {
			document = dataRawStoreStructuredMongoTest.RandomDocument(test.RandomOptionals())
		})

		Context("AsRaw", func() {
			It("returns the expected", func() {
				Expect(document.AsRaw()).To(PointTo(MatchAllFields(Fields{
					"ID":             Equal(dataRawStoreStructuredMongo.IDFromObjectIDAndDate(document.ID, document.CreatedTime)),
					"UserID":         Equal(document.UserID),
					"DataSetID":      Equal(document.DataSetID),
					"Metadata":       Equal(document.Metadata),
					"DigestMD5":      Equal(document.DigestMD5),
					"DigestSHA256":   Equal(document.DigestSHA256),
					"MediaType":      Equal(document.MediaType),
					"Size":           Equal(document.Size),
					"ProcessedTime":  Equal(document.ProcessedTime),
					"ArchivableTime": Equal(document.ArchivableTime),
					"ArchivedTime":   Equal(document.ArchivedTime),
					"CreatedTime":    Equal(document.CreatedTime),
					"ModifiedTime":   Equal(document.ModifiedTime),
					"Revision":       Equal(document.Revision),
				})))
			})
		})

		Context("AsContent", func() {
			It("returns content with uncompressed data when Compressed is false", func() {
				expectedData := test.RandomBytes()

				document.Compressed = false
				document.Data = bsonPrimitive.Binary{Data: expectedData}

				content := document.AsContent()
				Expect(content).ToNot(BeNil())
				Expect(content.DigestMD5).To(Equal(document.DigestMD5))
				Expect(content.DigestSHA256).To(Equal(document.DigestSHA256))
				Expect(content.MediaType).To(Equal(document.MediaType))
				Expect(content.ReadCloser).ToNot(BeNil())
				defer content.ReadCloser.Close()

				data, err := io.ReadAll(content.ReadCloser)
				Expect(err).ToNot(HaveOccurred())
				Expect(data).To(Equal(expectedData))
			})

			It("returns decompressed content when Compressed is true", func() {
				expectedData := test.RandomBytes()

				document.Compressed = true
				document.Data = bsonPrimitive.Binary{Data: compressTest.Compress(expectedData)}

				content := document.AsContent()
				Expect(content).ToNot(BeNil())
				Expect(content.DigestMD5).To(Equal(document.DigestMD5))
				Expect(content.DigestSHA256).To(Equal(document.DigestSHA256))
				Expect(content.MediaType).To(Equal(document.MediaType))
				Expect(content.ReadCloser).ToNot(BeNil())
				defer content.ReadCloser.Close()

				data, err := io.ReadAll(content.ReadCloser)
				Expect(err).ToNot(HaveOccurred())
				Expect(data).To(Equal(expectedData))
			})
		})
	})

	Context("Documents", func() {
		Context("AsRaw", func() {
			It("returns nil when documents is nil", func() {
				var documents dataRawStoreStructuredMongo.Documents
				Expect(documents.AsRaw()).To(BeNil())
			})

			It("returns empty slice when documents is empty", func() {
				documents := dataRawStoreStructuredMongo.Documents{}
				raws := documents.AsRaw()
				Expect(raws).ToNot(BeNil())
				Expect(raws).To(BeEmpty())
			})

			It("returns the expected Raws", func() {
				documents := make(dataRawStoreStructuredMongo.Documents, test.RandomIntFromRange(1, 5))
				expectedRaws := make([]*dataRaw.Raw, len(documents))
				for index := range documents {
					documents[index] = dataRawStoreStructuredMongoTest.RandomDocument(test.RandomOptionals())
					expectedRaws[index] = documents[index].AsRaw()
				}
				raws := documents.AsRaw()
				Expect(raws).To(Equal(expectedRaws))
			})
		})
	})

	Context("ObjectIDsFromIDs", func() {
		It("returns nil when ids is nil", func() {
			result, err := dataRawStoreStructuredMongo.ObjectIDsFromIDs(nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(BeNil())
		})

		It("returns an error when an id is missing", func() {
			ids := []string{dataRawTest.RandomDataRawID(), "", dataRawTest.RandomDataRawID()}
			result, err := dataRawStoreStructuredMongo.ObjectIDsFromIDs(ids)
			errorsTest.ExpectEqual(err, errors.New("id is missing"))
			Expect(result).To(BeNil())
		})

		It("returns an error when an id is invalid", func() {
			ids := []string{dataRawTest.RandomDataRawID(), "invalid"}
			result, err := dataRawStoreStructuredMongo.ObjectIDsFromIDs(ids)
			errorsTest.ExpectEqual(err, errors.New("id is invalid"))
			Expect(result).To(BeNil())
		})

		It("returns the expected ObjectIDs for valid ids", func() {
			expectedObjectIDs := make([]bsonPrimitive.ObjectID, test.RandomIntFromRange(1, 5))
			ids := make([]string, len(expectedObjectIDs))
			for index := range ids {
				expectedObjectIDs[index] = bsonPrimitive.NewObjectID()
				ids[index] = dataRawStoreStructuredMongo.IDFromObjectIDAndDate(expectedObjectIDs[index], test.RandomTime())
			}
			result, err := dataRawStoreStructuredMongo.ObjectIDsFromIDs(ids)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(expectedObjectIDs))
		})
	})

	Context("ObjectIDAndDateFromID", func() {
		It("returns an error when id is missing", func() {
			objectID, date, err := dataRawStoreStructuredMongo.ObjectIDAndDateFromID("")
			errorsTest.ExpectEqual(err, errors.New("id is missing"))
			Expect(objectID).To(Equal(bsonPrimitive.NilObjectID))
			Expect(date).To(Equal(time.Time{}))
		})

		It("returns an error when id has no separator", func() {
			objectID, date, err := dataRawStoreStructuredMongo.ObjectIDAndDateFromID("invalid")
			errorsTest.ExpectEqual(err, errors.New("id is invalid"))
			Expect(objectID).To(Equal(bsonPrimitive.NilObjectID))
			Expect(date).To(Equal(time.Time{}))
		})

		It("returns an error when the first part is not a valid ObjectID", func() {
			objectID, date, err := dataRawStoreStructuredMongo.ObjectIDAndDateFromID("invalid:2024-01-15")
			errorsTest.ExpectEqual(err, errors.New("id is invalid"))
			Expect(objectID).To(Equal(bsonPrimitive.NilObjectID))
			Expect(date).To(Equal(time.Time{}))
		})

		It("returns an error when the second part is not a valid date", func() {
			objectIDHex := bsonPrimitive.NewObjectID().Hex()
			objectID, date, err := dataRawStoreStructuredMongo.ObjectIDAndDateFromID(objectIDHex + ":not-a-date")
			errorsTest.ExpectEqual(err, errors.New("id is invalid"))
			Expect(objectID).To(Equal(bsonPrimitive.NilObjectID))
			Expect(date).To(Equal(time.Time{}))
		})

		It("returns the correct ObjectID and date for a valid id", func() {
			expectedObjectID := bsonPrimitive.NewObjectID()
			expectedDate := test.RandomTime()
			id := dataRawStoreStructuredMongo.IDFromObjectIDAndDate(expectedObjectID, expectedDate)
			objectID, date, err := dataRawStoreStructuredMongo.ObjectIDAndDateFromID(id)
			Expect(err).ToNot(HaveOccurred())
			Expect(objectID).To(Equal(expectedObjectID))
			Expect(date).To(Equal(expectedDate.UTC().Truncate(24 * time.Hour)))
		})
	})

	Context("IDFromObjectIDAndDate", func() {
		It("returns expected", func() {
			objectID := bsonPrimitive.NewObjectID()
			date := test.RandomTime()
			id := dataRawStoreStructuredMongo.IDFromObjectIDAndDate(objectID, date)
			Expect(id).To(Equal(fmt.Sprintf("%s:%s", objectID.Hex(), date.UTC().Format(time.DateOnly))))
		})

		It("round trip as expected through ObjectIDAndDateFromID", func() {
			originalObjectID := bsonPrimitive.NewObjectID()
			originalDate := test.RandomTime()
			id := dataRawStoreStructuredMongo.IDFromObjectIDAndDate(originalObjectID, originalDate)
			parsedObjectID, parsedDate, err := dataRawStoreStructuredMongo.ObjectIDAndDateFromID(id)
			Expect(err).ToNot(HaveOccurred())
			Expect(parsedObjectID).To(Equal(originalObjectID))
			Expect(parsedDate).To(Equal(originalDate.UTC().Truncate(24 * time.Hour)))
		})
	})
})
