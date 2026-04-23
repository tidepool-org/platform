package mongo_test

import (
	"context"
	"math/rand"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"github.com/tidepool-org/platform/auth"
	authTest "github.com/tidepool-org/platform/auth/test"
	dataSource "github.com/tidepool-org/platform/data/source"
	dataSourceStoreStructured "github.com/tidepool-org/platform/data/source/store/structured"
	dataSourceStoreStructuredMongo "github.com/tidepool-org/platform/data/source/store/structured/mongo"
	dataSourceTest "github.com/tidepool-org/platform/data/source/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/oauth"
	"github.com/tidepool-org/platform/page"
	pageTest "github.com/tidepool-org/platform/page/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	requestTest "github.com/tidepool-org/platform/request/test"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
	"github.com/tidepool-org/platform/test"
	userTest "github.com/tidepool-org/platform/user/test"
)

type CreatedTimeDescending dataSource.SourceArray

func (c CreatedTimeDescending) Len() int {
	return len(c)
}

func (c CreatedTimeDescending) Swap(left int, right int) {
	c[left], c[right] = c[right], c[left]
}

func (c CreatedTimeDescending) Less(left int, right int) bool {
	return c[right].CreatedTime.Before(c[left].CreatedTime)
}

func SelectAndSort(sources dataSource.SourceArray, selector func(s *dataSource.Source) bool) dataSource.SourceArray {
	var selected dataSource.SourceArray
	for _, s := range sources {
		if selector(s) {
			selected = append(selected, s)
		}
	}
	sort.Sort(CreatedTimeDescending(selected))
	return selected
}

func AsInterfaceArray(sources dataSource.SourceArray) []any {
	if sources == nil {
		return nil
	}
	array := make([]any, len(sources))
	for index, source := range sources {
		array[index] = source
	}
	return array
}

var _ = Describe("Mongo", func() {
	var config *storeStructuredMongo.Config
	var logger *logTest.Logger
	var store *dataSourceStoreStructuredMongo.Store
	var repository dataSourceStoreStructured.DataSourcesRepository

	BeforeEach(func() {
		config = storeStructuredMongoTest.NewConfig()
		logger = logTest.NewLogger()
	})

	AfterEach(func() {
		if store != nil {
			Expect(store.Terminate(context.Background())).ToNot(HaveOccurred())
		}
	})

	Context("NewStore", func() {
		It("returns an error when unsuccessful", func() {
			var err error
			store, err = dataSourceStoreStructuredMongo.NewStore(nil)
			errorsTest.ExpectEqual(err, errors.New("database config is empty"))
			Expect(store).To(BeNil())
		})

		It("returns a new store and no error when successful", func() {
			var err error
			store, err = dataSourceStoreStructuredMongo.NewStore(config)
			Expect(err).ToNot(HaveOccurred())
			Expect(store).ToNot(BeNil())
		})
	})

	Context("with a new store", func() {
		var mongoCollection *mongo.Collection

		BeforeEach(func() {
			var err error
			store, err = dataSourceStoreStructuredMongo.NewStore(config)
			Expect(err).ToNot(HaveOccurred())
			Expect(store).ToNot(BeNil())
			mongoCollection = store.GetCollection("data_sources")
		})

		Context("EnsureIndexes", func() {
			It("returns successfully", func() {
				Expect(store.EnsureIndexes()).To(Succeed())
				cursor, err := mongoCollection.Indexes().List(context.Background())
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
						"Key":    Equal(storeStructuredMongoTest.MakeKeySlice("id")),
						"Unique": Equal(true),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Key": Equal(storeStructuredMongoTest.MakeKeySlice("userId")),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Key": Equal(storeStructuredMongoTest.MakeKeySlice("providerName", "providerExternalId")),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Key": Equal(storeStructuredMongoTest.MakeKeySlice("userId", "providerType", "providerName")),
						"PartialFilterExpression": ConsistOf(
							MatchAllFields(Fields{
								"Key": Equal("state"),
								"Value": ConsistOf(
									MatchAllFields(Fields{
										"Key":   Equal("$in"),
										"Value": Equal(bson.A{"connected", "error"}),
									}),
								),
							}),
						),
						"Unique": Equal(true),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Key": Equal(storeStructuredMongoTest.MakeKeySlice("userId", "providerType", "providerName", "providerExternalId")),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Key": Equal(storeStructuredMongoTest.MakeKeySlice("providerSessionId")),
						"PartialFilterExpression": ConsistOf(
							MatchAllFields(Fields{
								"Key": Equal("providerSessionId"),
								"Value": ConsistOf(
									MatchAllFields(Fields{
										"Key":   Equal("$exists"),
										"Value": Equal(true),
									}),
								),
							}),
						),
						"Unique": Equal(true),
					}),
				))
			})
		})

		Context("NewDataSourcesRepository", func() {
			It("returns a new repository", func() {
				repository = store.NewDataSourcesRepository()
				Expect(repository).ToNot(BeNil())
			})
		})

		Context("with a new repository", func() {
			var ctx context.Context

			BeforeEach(func() {
				Expect(store.EnsureIndexes()).To(Succeed())
				repository = store.NewDataSourcesRepository()
				ctx = log.NewContextWithLogger(context.Background(), logger)
			})

			Context("with user id", func() {
				var userID string

				BeforeEach(func() {
					userID = userTest.RandomUserID()
				})

				Context("List", func() {
					It("returns an error when the context is missing", func() {
						result, err := repository.List(context.Context(nil), userID, dataSourceTest.RandomFilter(test.AllowOptionals()), pageTest.RandomPagination())
						errorsTest.ExpectEqual(err, errors.New("context is missing"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the user id is missing", func() {
						userID = ""
						result, err := repository.List(ctx, userID, dataSourceTest.RandomFilter(test.AllowOptionals()), pageTest.RandomPagination())
						errorsTest.ExpectEqual(err, errors.New("user id is missing"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the user id is invalid", func() {
						userID = "invalid"
						result, err := repository.List(ctx, userID, dataSourceTest.RandomFilter(test.AllowOptionals()), pageTest.RandomPagination())
						errorsTest.ExpectEqual(err, errors.New("user id is invalid"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the filter is invalid", func() {
						filter := dataSourceTest.RandomFilter(test.AllowOptionals())
						filter.ProviderType = pointer.FromString("")
						result, err := repository.List(ctx, userID, filter, pageTest.RandomPagination())
						errorsTest.ExpectEqual(err, errors.New("filter is invalid"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the pagination is invalid", func() {
						pagination := pageTest.RandomPagination()
						pagination.Page = -1
						result, err := repository.List(ctx, userID, dataSourceTest.RandomFilter(test.AllowOptionals()), pagination)
						errorsTest.ExpectEqual(err, errors.New("pagination is invalid"))
						Expect(result).To(BeNil())
					})

					Context("with data", func() {
						var providerType string
						var providerName string
						var providerSessionID string
						var providerExternalID string
						var allResult dataSource.SourceArray
						var filter *dataSource.Filter
						var pagination *page.Pagination

						BeforeEach(func() {
							providerType = auth.ProviderTypeOAuth
							providerName = authTest.RandomProviderName()
							providerSessionID = authTest.RandomProviderSessionID()
							providerExternalID = authTest.RandomProviderExternalID()
							allResult = dataSource.SourceArray{}
							for index, randomResult := range dataSourceTest.RandomSourceArray(12, 12) {
								if index < 4 {
									randomResult.State = dataSource.StateConnected
								} else if index < 8 {
									randomResult.State = dataSource.StateDisconnected
								} else {
									randomResult.State = dataSource.StateError
								}
								if index == 0 || (index >= 4 && index < 6) {
									randomResult.ProviderType = providerType
									randomResult.ProviderName = providerName
								}
								if randomResult.State != dataSource.StateDisconnected {
									randomResult.ProviderSessionID = pointer.FromString(authTest.RandomProviderSessionID())
								} else {
									randomResult.ProviderSessionID = nil
								}
								if index%2 == 0 {
									randomResult.ProviderExternalID = pointer.FromString(providerExternalID)
								}
								userResult := dataSourceTest.CloneSource(randomResult)
								userResult.ID = dataSourceTest.RandomDataSourceID()
								userResult.UserID = userID
								if index == 0 {
									userResult.ProviderSessionID = pointer.FromString(providerSessionID)
								} else if randomResult.State != dataSource.StateDisconnected {
									userResult.ProviderSessionID = pointer.FromString(authTest.RandomProviderSessionID())
								}
								allResult = append(allResult, randomResult, userResult)
							}
							rand.Shuffle(len(allResult), func(i, j int) { allResult[i], allResult[j] = allResult[j], allResult[i] })
							_, err := mongoCollection.InsertMany(context.Background(), AsInterfaceArray(allResult))
							Expect(err).ToNot(HaveOccurred())
							filter = &dataSource.Filter{}
							pagination = page.NewPagination()
						})

						It("returns no result when the user id is unknown", func() {
							userID = userTest.RandomUserID()
							Expect(repository.List(ctx, userID, filter, pagination)).To(SatisfyAll(Not(BeNil()), BeEmpty()))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 0})
						})

						It("returns expected result when the filter is missing", func() {
							filter = nil
							Expect(repository.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(s *dataSource.Source) bool { return s.UserID == userID },
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "pagination": pagination, "count": 12})
						})

						It("returns expected result when the filter provider type is missing", func() {
							filter.ProviderType = nil
							Expect(repository.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(s *dataSource.Source) bool { return s.UserID == userID },
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 12})
						})

						It("returns expected result when the filter provider type is specified", func() {
							filter.ProviderType = pointer.FromString(providerType)
							Expect(repository.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(s *dataSource.Source) bool { return s.UserID == userID },
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 12})
						})

						It("returns expected result when the filter provider name is missing", func() {
							filter.ProviderName = nil
							Expect(repository.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(s *dataSource.Source) bool { return s.UserID == userID },
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 12})
						})

						It("returns expected result when the filter provider name is specified", func() {
							filter.ProviderName = pointer.FromString(providerName)
							Expect(repository.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(s *dataSource.Source) bool {
									return s.UserID == userID && s.ProviderName == providerName
								},
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 3})
						})

						It("returns expected result when the filter provider type, provider name, and provider external id are set", func() {
							filter.ProviderType = pointer.FromString(oauth.ProviderType)
							filter.ProviderName = pointer.FromString(providerName)
							filter.ProviderExternalID = pointer.FromString(providerExternalID)
							Expect(repository.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(s *dataSource.Source) bool {
									return s.UserID == userID && s.ProviderType == oauth.ProviderType && s.ProviderName == providerName &&
										s.ProviderExternalID != nil && *s.ProviderExternalID == providerExternalID
								},
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 2})
						})

						It("returns expected result when the pagination is missing", func() {
							Expect(repository.List(ctx, userID, filter, nil)).To(Equal(SelectAndSort(allResult,
								func(s *dataSource.Source) bool { return s.UserID == userID },
							)))
							logger.AssertDebug("List", log.Fields{"filter": filter, "count": 12})
						})

						It("returns expected result when the pagination limits result", func() {
							pagination.Page = 1
							pagination.Size = 2
							Expect(repository.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(s *dataSource.Source) bool { return s.UserID == userID },
							)[2:4]))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 2})
						})
					})
				})

				Context("Create", func() {
					var create *dataSource.Create

					BeforeEach(func() {
						create = dataSourceTest.RandomCreate(test.AllowOptionals())
					})

					It("returns an error when the context is missing", func() {
						ctx = nil
						result, err := repository.Create(ctx, userID, create)
						errorsTest.ExpectEqual(err, errors.New("context is missing"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the user id is missing", func() {
						userID = ""
						result, err := repository.Create(ctx, userID, create)
						errorsTest.ExpectEqual(err, errors.New("user id is missing"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the user id is invalid", func() {
						userID = "invalid"
						result, err := repository.Create(ctx, userID, create)
						errorsTest.ExpectEqual(err, errors.New("user id is invalid"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the create is missing", func() {
						create = nil
						result, err := repository.Create(ctx, userID, create)
						errorsTest.ExpectEqual(err, errors.New("create is missing"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the create is invalid", func() {
						create.ProviderType = ""
						result, err := repository.Create(ctx, userID, create)
						errorsTest.ExpectEqual(err, errors.New("create is invalid"))
						Expect(result).To(BeNil())
					})

					It("returns the result after creating", func() {
						matchAllFields := MatchAllFields(Fields{
							"ID":                 Not(BeEmpty()),
							"UserID":             Equal(userID),
							"ProviderType":       Equal(create.ProviderType),
							"ProviderName":       Equal(create.ProviderName),
							"ProviderSessionID":  BeNil(),
							"ProviderExternalID": Equal(create.ProviderExternalID),
							"State":              Equal(dataSource.StateDisconnected),
							"Metadata":           Equal(create.Metadata),
							"Error":              BeNil(),
							"DataSetID":          BeNil(),
							"EarliestDataTime":   BeNil(),
							"LatestDataTime":     BeNil(),
							"LastImportTime":     BeNil(),
							"CreatedTime":        BeTemporally("~", time.Now(), time.Second),
							"ModifiedTime":       BeNil(),
							"Revision":           Equal(0),
						})
						result, err := repository.Create(ctx, userID, create)
						Expect(err).ToNot(HaveOccurred())
						Expect(result).ToNot(BeNil())
						Expect(*result).To(matchAllFields)
						storeResult := dataSource.SourceArray{}
						cursor, err := mongoCollection.Find(context.Background(), bson.M{"id": result.ID})
						Expect(err).ToNot(HaveOccurred())
						Expect(cursor).ToNot(BeNil())
						Expect(cursor.All(context.Background(), &storeResult)).To(Succeed())
						Expect(storeResult).To(HaveLen(1))
						Expect(*storeResult[0]).To(matchAllFields)
						logger.AssertDebug("Create", log.Fields{"create": create, "id": storeResult[0].ID})
					})
				})

				Context("DeleteAll", func() {
					It("returns an error when the context is missing", func() {
						ctx = nil
						err := repository.DeleteAll(ctx, userID)
						errorsTest.ExpectEqual(err, errors.New("context is missing"))
					})

					It("returns an error when the user id is missing", func() {
						userID = ""
						err := repository.DeleteAll(ctx, userID)
						errorsTest.ExpectEqual(err, errors.New("user id is missing"))
					})

					It("returns an error when the user id is invalid", func() {
						userID = "invalid"
						err := repository.DeleteAll(ctx, userID)
						errorsTest.ExpectEqual(err, errors.New("user id is invalid"))
					})

					Context("with data", func() {
						var originals dataSource.SourceArray

						BeforeEach(func() {
							originals = dataSourceTest.RandomSourceArray(2, 4)
							for _, original := range originals {
								original.UserID = userID
							}
							_, err := mongoCollection.InsertMany(context.Background(), AsInterfaceArray(originals))
							Expect(err).ToNot(HaveOccurred())
							_, err = mongoCollection.InsertMany(context.Background(), []interface{}{dataSourceTest.RandomSource(test.AllowOptionals()), dataSourceTest.RandomSource(test.AllowOptionals())})
							Expect(err).ToNot(HaveOccurred())
						})

						AfterEach(func() {
							logger.AssertDebug("DeleteAll", log.Fields{"userId": userID})
						})

						It("returns successfully and does not destroy the original when the id does not exist", func() {
							originalUserID := userID
							userID = userTest.RandomUserID()
							Expect(repository.DeleteAll(ctx, userID)).To(Succeed())
							Expect(mongoCollection.CountDocuments(context.Background(), bson.M{"userId": originalUserID})).To(Equal(int64(len(originals))))
							Expect(mongoCollection.CountDocuments(context.Background(), bson.M{})).To(Equal(int64(len(originals) + 2)))
						})

						It("returns successfully and destroys the original when the id exists and the condition is missing", func() {
							Expect(repository.DeleteAll(ctx, userID)).To(Succeed())
							Expect(mongoCollection.CountDocuments(context.Background(), bson.M{"userId": userID})).To(Equal(int64(0)))
							Expect(mongoCollection.CountDocuments(context.Background(), bson.M{})).To(Equal(int64(2)))
						})
					})
				})
			})

			Context("Get", func() {
				var id string

				BeforeEach(func() {
					id = dataSourceTest.RandomDataSourceID()
				})

				It("returns an error when the context is missing", func() {
					ctx = nil
					result, err := repository.Get(ctx, id)
					errorsTest.ExpectEqual(err, errors.New("context is missing"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the id is missing", func() {
					id = ""
					result, err := repository.Get(ctx, id)
					errorsTest.ExpectEqual(err, errors.New("id is missing"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the id is invalid", func() {
					id = "invalid"
					result, err := repository.Get(ctx, id)
					errorsTest.ExpectEqual(err, errors.New("id is invalid"))
					Expect(result).To(BeNil())
				})

				Context("with data", func() {
					var allResult dataSource.SourceArray
					var result *dataSource.Source

					BeforeEach(func() {
						allResult = dataSourceTest.RandomSourceArray(4, 4)
						result = allResult[0]
						result.ID = id
						rand.Shuffle(len(allResult), func(i, j int) { allResult[i], allResult[j] = allResult[j], allResult[i] })
					})

					JustBeforeEach(func() {
						_, err := mongoCollection.InsertMany(context.Background(), AsInterfaceArray(allResult))
						Expect(err).ToNot(HaveOccurred())
					})

					AfterEach(func() {
						logger.AssertDebug("Get", log.Fields{"id": id})
					})

					It("returns nil when the id does not exist", func() {
						id = dataSourceTest.RandomDataSourceID()
						Expect(repository.Get(ctx, id)).To(BeNil())
					})

					It("returns the result when the id exists", func() {
						Expect(repository.Get(ctx, id)).To(Equal(result))
					})
				})
			})

			Context("Update", func() {
				var id string
				var condition *request.Condition
				var update *dataSource.Update

				BeforeEach(func() {
					id = dataSourceTest.RandomDataSourceID()
					condition = requestTest.RandomCondition()
					update = dataSourceTest.RandomUpdate(test.AllowOptionals())
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
					update.State = pointer.FromString("")
					result, err := repository.Update(ctx, id, condition, update)
					errorsTest.ExpectEqual(err, errors.New("update is invalid"))
					Expect(result).To(BeNil())
				})

				Context("with data", func() {
					var original *dataSource.Source

					BeforeEach(func() {
						original = dataSourceTest.RandomSource(test.AllowOptionals())
						original.ID = id
						_, err := mongoCollection.InsertOne(context.Background(), original)
						Expect(err).ToNot(HaveOccurred())
					})

					AfterEach(func() {
						if condition != nil {
							logger.AssertDebug("Update", log.Fields{"id": id, "condition": condition, "update": update})
						} else {
							logger.AssertDebug("Update", log.Fields{"id": id, "update": update})
						}
					})

					When("the condition revision does not match", func() {
						BeforeEach(func() {
							condition.Revision = pointer.FromInt(original.Revision + 1)
						})

						It("returns nil", func() {
							Expect(repository.Update(ctx, id, condition, update)).To(BeNil())
						})
					})

					updateAssertions := func() {
						Context("with updates", func() {
							It("returns updated result when the id exists and state is connected without error", func() {
								update.ProviderSessionID = pointer.FromString(authTest.RandomProviderSessionID())
								update.ProviderExternalID = pointer.FromString(authTest.RandomProviderExternalID())
								update.State = pointer.FromString(dataSource.StateConnected)
								update.Error = &errors.Serializable{Error: nil}
								matchAllFields := MatchAllFields(Fields{
									"ID":                 Equal(id),
									"UserID":             Equal(original.UserID),
									"ProviderType":       Equal(original.ProviderType),
									"ProviderName":       Equal(original.ProviderName),
									"ProviderSessionID":  Equal(update.ProviderSessionID),
									"ProviderExternalID": Equal(update.ProviderExternalID),
									"State":              Equal(*update.State),
									"Metadata":           Equal(pointer.Default(update.Metadata, original.Metadata)),
									"Error":              BeNil(),
									"DataSetID":          Equal(pointer.DefaultPointer(update.DataSetID, original.DataSetID)),
									"EarliestDataTime":   Equal(pointer.DefaultPointer(update.EarliestDataTime, original.EarliestDataTime)),
									"LatestDataTime":     Equal(pointer.DefaultPointer(update.LatestDataTime, original.LatestDataTime)),
									"LastImportTime":     Equal(pointer.DefaultPointer(update.LastImportTime, original.LastImportTime)),
									"CreatedTime":        Equal(original.CreatedTime),
									"ModifiedTime":       PointTo(BeTemporally("~", time.Now(), time.Second)),
									"Revision":           Equal(original.Revision + 1),
								})
								result, err := repository.Update(ctx, id, condition, update)
								Expect(err).ToNot(HaveOccurred())
								Expect(result).ToNot(BeNil())
								Expect(*result).To(matchAllFields)
								storeResult := dataSource.SourceArray{}
								cursor, err := mongoCollection.Find(context.Background(), bson.M{"id": id})
								Expect(err).ToNot(HaveOccurred())
								Expect(cursor).ToNot(BeNil())
								Expect(cursor.All(context.Background(), &storeResult)).To(Succeed())
								Expect(storeResult).To(HaveLen(1))
								Expect(*storeResult[0]).To(matchAllFields)
							})

							It("returns updated result when the id exists and state is disconnected without error", func() {
								update.ProviderSessionID = nil
								update.ProviderExternalID = pointer.FromString(authTest.RandomProviderExternalID())
								update.State = pointer.FromString(dataSource.StateDisconnected)
								update.Error = &errors.Serializable{Error: nil}
								matchAllFields := MatchAllFields(Fields{
									"ID":                 Equal(id),
									"UserID":             Equal(original.UserID),
									"ProviderType":       Equal(original.ProviderType),
									"ProviderName":       Equal(original.ProviderName),
									"ProviderSessionID":  BeNil(),
									"ProviderExternalID": Equal(update.ProviderExternalID),
									"State":              Equal(*update.State),
									"Metadata":           Equal(pointer.Default(update.Metadata, original.Metadata)),
									"Error":              BeNil(),
									"DataSetID":          Equal(pointer.DefaultPointer(update.DataSetID, original.DataSetID)),
									"EarliestDataTime":   Equal(pointer.DefaultPointer(update.EarliestDataTime, original.EarliestDataTime)),
									"LatestDataTime":     Equal(pointer.DefaultPointer(update.LatestDataTime, original.LatestDataTime)),
									"LastImportTime":     Equal(pointer.DefaultPointer(update.LastImportTime, original.LastImportTime)),
									"CreatedTime":        Equal(original.CreatedTime),
									"ModifiedTime":       PointTo(BeTemporally("~", time.Now(), time.Second)),
									"Revision":           Equal(original.Revision + 1),
								})
								result, err := repository.Update(ctx, id, condition, update)
								Expect(err).ToNot(HaveOccurred())
								Expect(result).ToNot(BeNil())
								Expect(*result).To(matchAllFields)
								storeResult := dataSource.SourceArray{}
								cursor, err := mongoCollection.Find(context.Background(), bson.M{"id": id})
								Expect(err).ToNot(HaveOccurred())
								Expect(cursor).ToNot(BeNil())
								Expect(cursor.All(context.Background(), &storeResult)).To(Succeed())
								Expect(storeResult).To(HaveLen(1))
								Expect(*storeResult[0]).To(matchAllFields)
							})

							It("returns updated result when the id exists and state is error with error", func() {
								update.ProviderSessionID = nil
								update.ProviderExternalID = pointer.FromString(authTest.RandomProviderExternalID())
								update.State = pointer.FromString(dataSource.StateError)
								update.Error = errorsTest.RandomSerializable()
								matchAllFields := MatchAllFields(Fields{
									"ID":                 Equal(id),
									"UserID":             Equal(original.UserID),
									"ProviderType":       Equal(original.ProviderType),
									"ProviderName":       Equal(original.ProviderName),
									"ProviderSessionID":  Equal(original.ProviderSessionID),
									"ProviderExternalID": Equal(update.ProviderExternalID),
									"State":              Equal(*update.State),
									"Metadata":           Equal(pointer.Default(update.Metadata, original.Metadata)),
									"Error":              Equal(update.Error),
									"DataSetID":          Equal(pointer.DefaultPointer(update.DataSetID, original.DataSetID)),
									"EarliestDataTime":   Equal(pointer.DefaultPointer(update.EarliestDataTime, original.EarliestDataTime)),
									"LatestDataTime":     Equal(pointer.DefaultPointer(update.LatestDataTime, original.LatestDataTime)),
									"LastImportTime":     Equal(pointer.DefaultPointer(update.LastImportTime, original.LastImportTime)),
									"CreatedTime":        Equal(original.CreatedTime),
									"ModifiedTime":       PointTo(BeTemporally("~", time.Now(), time.Second)),
									"Revision":           Equal(original.Revision + 1),
								})
								result, err := repository.Update(ctx, id, condition, update)
								Expect(err).ToNot(HaveOccurred())
								Expect(result).ToNot(BeNil())
								Expect(*result).To(matchAllFields)
								storeResult := dataSource.SourceArray{}
								cursor, err := mongoCollection.Find(context.Background(), bson.M{"id": id})
								Expect(err).ToNot(HaveOccurred())
								Expect(cursor).ToNot(BeNil())
								Expect(cursor.All(context.Background(), &storeResult)).To(Succeed())
								Expect(storeResult).To(HaveLen(1))
								Expect(*storeResult[0]).To(matchAllFields)
							})

							It("returns nil when the id does not exist", func() {
								id = dataSourceTest.RandomDataSourceID()
								Expect(repository.Update(ctx, id, condition, update)).To(BeNil())
							})
						})

						Context("without updates", func() {
							BeforeEach(func() {
								update = dataSource.NewUpdate()
							})

							It("returns original when the id exists", func() {
								Expect(repository.Update(ctx, id, condition, update)).To(Equal(original))
							})

							It("returns nil when the id does not exist", func() {
								id = dataSourceTest.RandomDataSourceID()
								Expect(repository.Update(ctx, id, condition, update)).To(BeNil())
							})
						})
					}

					When("the condition is missing", func() {
						BeforeEach(func() {
							condition = nil
						})

						updateAssertions()
					})

					When("the condition revision is missing", func() {
						BeforeEach(func() {
							condition.Revision = nil
						})

						updateAssertions()
					})

					When("the condition revision matches", func() {
						BeforeEach(func() {
							condition.Revision = pointer.FromInt(original.Revision)
						})

						updateAssertions()
					})
				})
			})

			Context("Delete", func() {
				var id string
				var condition *request.Condition

				BeforeEach(func() {
					id = dataSourceTest.RandomDataSourceID()
					condition = requestTest.RandomCondition()
				})

				It("returns an error when the context is missing", func() {
					ctx = nil
					deleted, err := repository.Delete(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("context is missing"))
					Expect(deleted).To(BeFalse())
				})

				It("returns an error when the id is missing", func() {
					id = ""
					deleted, err := repository.Delete(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("id is missing"))
					Expect(deleted).To(BeFalse())
				})

				It("returns an error when the id is invalid", func() {
					id = "invalid"
					deleted, err := repository.Delete(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("id is invalid"))
					Expect(deleted).To(BeFalse())
				})

				It("returns an error when the condition is invalid", func() {
					condition.Revision = pointer.FromInt(-1)
					deleted, err := repository.Delete(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("condition is invalid"))
					Expect(deleted).To(BeFalse())
				})

				Context("with data", func() {
					var original *dataSource.Source

					BeforeEach(func() {
						original = dataSourceTest.RandomSource(test.AllowOptionals())
						original.ID = id
						_, err := mongoCollection.InsertMany(context.Background(), []any{original, dataSourceTest.RandomSource(test.AllowOptionals()), dataSourceTest.RandomSource(test.AllowOptionals())})
						Expect(err).ToNot(HaveOccurred())
					})

					AfterEach(func() {
						if condition != nil {
							logger.AssertDebug("Delete", log.Fields{"id": id, "condition": condition})
						} else {
							logger.AssertDebug("Delete", log.Fields{"id": id})
						}
					})

					It("returns false and does not delete the original when the id does not exist", func() {
						id = dataSourceTest.RandomDataSourceID()
						Expect(repository.Delete(ctx, id, condition)).To(BeFalse())
						Expect(mongoCollection.CountDocuments(context.Background(), bson.M{"id": original.ID})).To(Equal(int64(1)))
					})

					It("returns false and does not delete the original when the id exists, but the condition revision does not match", func() {
						condition.Revision = pointer.FromInt(original.Revision + 1)
						Expect(repository.Delete(ctx, id, condition)).To(BeFalse())
						Expect(mongoCollection.CountDocuments(context.Background(), bson.M{"id": original.ID})).To(Equal(int64(1)))
					})

					It("returns true and deletes the original when the id exists and the condition is missing", func() {
						condition = nil
						Expect(repository.Delete(ctx, id, condition)).To(BeTrue())
						Expect(mongoCollection.CountDocuments(context.Background(), bson.M{"id": original.ID})).To(Equal(int64(0)))
					})

					It("returns true and deletes the original when the id exists and the condition revision is missing", func() {
						condition.Revision = nil
						Expect(repository.Delete(ctx, id, condition)).To(BeTrue())
						Expect(mongoCollection.CountDocuments(context.Background(), bson.M{"id": original.ID})).To(Equal(int64(0)))
					})

					It("returns true and deletes the original when the id exists and the condition revision matches", func() {
						condition.Revision = pointer.FromInt(original.Revision)
						Expect(repository.Delete(ctx, id, condition)).To(BeTrue())
						Expect(mongoCollection.CountDocuments(context.Background(), bson.M{"id": original.ID})).To(Equal(int64(0)))
					})
				})
			})

			Context("GetFromProviderSession", func() {
				var providerSessionID string

				BeforeEach(func() {
					providerSessionID = authTest.RandomProviderSessionID()
				})

				It("returns an error when the context is missing", func() {
					ctx = nil
					result, err := repository.GetFromProviderSession(ctx, providerSessionID)
					errorsTest.ExpectEqual(err, errors.New("context is missing"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the provider session id is missing", func() {
					providerSessionID = ""
					result, err := repository.GetFromProviderSession(ctx, providerSessionID)
					errorsTest.ExpectEqual(err, errors.New("provider session id is missing"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the provider session id is invalid", func() {
					providerSessionID = "invalid"
					result, err := repository.GetFromProviderSession(ctx, providerSessionID)
					errorsTest.ExpectEqual(err, errors.New("provider session id is invalid"))
					Expect(result).To(BeNil())
				})

				Context("with data", func() {
					var allResult dataSource.SourceArray
					var result *dataSource.Source

					BeforeEach(func() {
						allResult = dataSourceTest.RandomSourceArray(4, 4)
						result = allResult[0]
						result.ProviderSessionID = pointer.FromString(providerSessionID)
						rand.Shuffle(len(allResult), func(i, j int) { allResult[i], allResult[j] = allResult[j], allResult[i] })
					})

					JustBeforeEach(func() {
						_, err := mongoCollection.InsertMany(context.Background(), AsInterfaceArray(allResult))
						Expect(err).ToNot(HaveOccurred())
					})

					AfterEach(func() {
						logger.AssertDebug("GetFromProviderSession", log.Fields{"providerSessionId": providerSessionID})
					})

					It("returns nil when the provider session id does not exist", func() {
						providerSessionID = dataSourceTest.RandomDataSourceID()
						Expect(repository.GetFromProviderSession(ctx, providerSessionID)).To(BeNil())
					})

					It("returns the result when the provider session id exists", func() {
						Expect(repository.GetFromProviderSession(ctx, providerSessionID)).To(Equal(result))
					})
				})
			})
		})
	})
})
