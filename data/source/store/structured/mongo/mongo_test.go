package mongo_test

import (
	"context"
	"math/rand"
	"sort"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

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
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	requestTest "github.com/tidepool-org/platform/request/test"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
	userTest "github.com/tidepool-org/platform/user/test"
)

type CreatedTimeDescending dataSource.Sources

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

func SelectAndSort(sources dataSource.Sources, selector func(s *dataSource.Source) bool) dataSource.Sources {
	var selected dataSource.Sources
	for _, s := range sources {
		if selector(s) {
			selected = append(selected, s)
		}
	}
	sort.Sort(CreatedTimeDescending(selected))
	return selected
}

func AsInterfaceArray(sources dataSource.Sources) []interface{} {
	if sources == nil {
		return nil
	}
	array := make([]interface{}, len(sources))
	for index, source := range sources {
		array[index] = source
	}
	return array
}

var _ = Describe("Mongo", func() {
	var config *storeStructuredMongo.Config
	var logger *logTest.Logger
	var store *dataSourceStoreStructuredMongo.Store
	var session dataSourceStoreStructured.Session

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
			store, err = dataSourceStoreStructuredMongo.NewStore(nil, logger)
			errorsTest.ExpectEqual(err, errors.New("config is missing"))
			Expect(store).To(BeNil())
		})

		It("returns a new store and no error when successful", func() {
			var err error
			store, err = dataSourceStoreStructuredMongo.NewStore(config, logger)
			Expect(err).ToNot(HaveOccurred())
			Expect(store).ToNot(BeNil())
		})
	})

	Context("with a new store", func() {
		var mgoSession *mgo.Session
		var mgoCollection *mgo.Collection

		BeforeEach(func() {
			var err error
			store, err = dataSourceStoreStructuredMongo.NewStore(config, logger)
			Expect(err).ToNot(HaveOccurred())
			Expect(store).ToNot(BeNil())
			mgoSession = storeStructuredMongoTest.Session().Copy()
			mgoCollection = mgoSession.DB(config.Database).C(config.CollectionPrefix + "data_sources")
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
					var filter *dataSource.Filter
					var pagination *page.Pagination

					BeforeEach(func() {
						filter = dataSource.NewFilter()
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
						filter.ProviderType = pointer.FromStringArray([]string{""})
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
						var providerType string
						var providerName string
						var providerSessionID string
						var allResult dataSource.Sources

						BeforeEach(func() {
							providerType = auth.ProviderTypeOAuth
							providerName = authTest.RandomProviderName()
							providerSessionID = authTest.RandomProviderSessionID()
							allResult = dataSource.Sources{}
							for index, randomResult := range dataSourceTest.RandomSources(12, 12) {
								if index < 4 {
									randomResult.State = pointer.FromString(dataSource.StateConnected)
								} else if index < 8 {
									randomResult.State = pointer.FromString(dataSource.StateDisconnected)
								} else {
									randomResult.State = pointer.FromString(dataSource.StateError)
								}
								if index%2 == 0 {
									randomResult.ProviderName = pointer.FromString(providerName)
								}
								if (index/2)%2 == 0 {
									randomResult.ProviderSessionID = pointer.FromString(providerSessionID)
								}
								userResult := dataSourceTest.CloneSource(randomResult)
								userResult.ID = pointer.FromString(dataSourceTest.RandomID())
								userResult.UserID = pointer.FromString(userID)
								allResult = append(allResult, randomResult, userResult)
							}
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
								func(s *dataSource.Source) bool { return *s.UserID == userID },
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "pagination": pagination, "count": 12})
						})

						It("returns expected result when the filter provider type is missing", func() {
							filter.ProviderType = nil
							Expect(session.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(s *dataSource.Source) bool { return *s.UserID == userID },
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 12})
						})

						It("returns expected result when the filter provider type is specified", func() {
							filter.ProviderType = pointer.FromStringArray([]string{providerType})
							Expect(session.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(s *dataSource.Source) bool { return *s.UserID == userID },
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 12})
						})

						It("returns expected result when the filter provider name is missing", func() {
							filter.ProviderName = nil
							Expect(session.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(s *dataSource.Source) bool { return *s.UserID == userID },
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 12})
						})

						It("returns expected result when the filter provider name is specified", func() {
							filter.ProviderName = pointer.FromStringArray([]string{providerName})
							Expect(session.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(s *dataSource.Source) bool {
									return *s.UserID == userID && *s.ProviderName == providerName
								},
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 6})
						})

						It("returns expected result when the filter provider session id is missing", func() {
							filter.ProviderSessionID = nil
							Expect(session.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(s *dataSource.Source) bool { return *s.UserID == userID },
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 12})
						})

						It("returns expected result when the filter provider session id is specified", func() {
							filter.ProviderSessionID = pointer.FromStringArray([]string{providerSessionID})
							Expect(session.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(s *dataSource.Source) bool {
									return *s.UserID == userID && *s.ProviderSessionID == providerSessionID
								},
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 6})
						})

						It("returns expected result when the filter state is missing", func() {
							filter.State = nil
							Expect(session.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(s *dataSource.Source) bool { return *s.UserID == userID },
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 12})
						})

						It("returns expected result when the filter state is set to connected", func() {
							filter.State = pointer.FromStringArray([]string{dataSource.StateConnected})
							Expect(session.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(s *dataSource.Source) bool {
									return *s.UserID == userID && *s.State == dataSource.StateConnected
								},
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 4})
						})

						It("returns expected result when the filter state is set to disconnected", func() {
							filter.State = pointer.FromStringArray([]string{dataSource.StateDisconnected})
							Expect(session.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(s *dataSource.Source) bool {
									return *s.UserID == userID && *s.State == dataSource.StateDisconnected
								},
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 4})
						})

						It("returns expected result when the filter state is set to error", func() {
							filter.State = pointer.FromStringArray([]string{dataSource.StateError})
							Expect(session.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(s *dataSource.Source) bool {
									return *s.UserID == userID && *s.State == dataSource.StateError
								},
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 4})
						})

						It("returns expected result when the filter state is set to both connected and disconnected", func() {
							filter.State = pointer.FromStringArray([]string{dataSource.StateConnected, dataSource.StateDisconnected})
							Expect(session.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(s *dataSource.Source) bool {
									return *s.UserID == userID && (*s.State == dataSource.StateConnected || *s.State == dataSource.StateDisconnected)
								},
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 8})
						})

						It("returns expected result when the filter state is set to both disconnected and error", func() {
							filter.State = pointer.FromStringArray([]string{dataSource.StateDisconnected, dataSource.StateError})
							Expect(session.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(s *dataSource.Source) bool {
									return *s.UserID == userID && (*s.State == dataSource.StateDisconnected || *s.State == dataSource.StateError)
								},
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 8})
						})

						It("returns expected result when the filter state is set to all states", func() {
							filter.State = pointer.FromStringArray(dataSource.States())
							Expect(session.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(s *dataSource.Source) bool { return *s.UserID == userID },
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 12})
						})

						It("returns expected result when the filter provider type, provider name, provider session id, and state is set to connected and disconnected", func() {
							filter.ProviderType = pointer.FromStringArray([]string{providerType})
							filter.ProviderName = pointer.FromStringArray([]string{providerName})
							filter.ProviderSessionID = pointer.FromStringArray([]string{providerSessionID})
							filter.State = pointer.FromStringArray([]string{dataSource.StateConnected, dataSource.StateDisconnected})
							Expect(session.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(s *dataSource.Source) bool {
									return *s.UserID == userID && *s.ProviderName == providerName && *s.ProviderSessionID == providerSessionID && (*s.State == dataSource.StateConnected || *s.State == dataSource.StateDisconnected)
								},
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 2})
						})

						It("returns expected result when the pagination is missing", func() {
							pagination = nil
							Expect(session.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(s *dataSource.Source) bool { return *s.UserID == userID },
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "count": 12})
						})

						It("returns expected result when the pagination limits result", func() {
							pagination.Page = 1
							pagination.Size = 2
							Expect(session.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(s *dataSource.Source) bool { return *s.UserID == userID },
							)[2:4]))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 2})
						})
					})
				})

				Context("Create", func() {
					var create *dataSource.Create

					BeforeEach(func() {
						create = dataSourceTest.RandomCreate()
					})

					It("returns an error when the context is missing", func() {
						ctx = nil
						result, err := session.Create(ctx, userID, create)
						errorsTest.ExpectEqual(err, errors.New("context is missing"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the user id is missing", func() {
						userID = ""
						result, err := session.Create(ctx, userID, create)
						errorsTest.ExpectEqual(err, errors.New("user id is missing"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the user id is invalid", func() {
						userID = "invalid"
						result, err := session.Create(ctx, userID, create)
						errorsTest.ExpectEqual(err, errors.New("user id is invalid"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the create is missing", func() {
						create = nil
						result, err := session.Create(ctx, userID, create)
						errorsTest.ExpectEqual(err, errors.New("create is missing"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the create is invalid", func() {
						create.ProviderType = pointer.FromString("")
						result, err := session.Create(ctx, userID, create)
						errorsTest.ExpectEqual(err, errors.New("create is invalid"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the session is closed", func() {
						session.Close()
						result, err := session.Create(ctx, userID, create)
						errorsTest.ExpectEqual(err, errors.New("session closed"))
						Expect(result).To(BeNil())
					})

					It("returns the result after creating", func() {
						matchAllFields := MatchAllFields(Fields{
							"ID":                PointTo(Not(BeEmpty())),
							"UserID":            PointTo(Equal(userID)),
							"ProviderType":      Equal(create.ProviderType),
							"ProviderName":      Equal(create.ProviderName),
							"ProviderSessionID": Equal(create.ProviderSessionID),
							"State":             Equal(create.State),
							"Error":             BeNil(),
							"DataSetIDs":        BeNil(),
							"EarliestDataTime":  BeNil(),
							"LatestDataTime":    BeNil(),
							"LastImportTime":    BeNil(),
							"CreatedTime":       PointTo(BeTemporally("~", time.Now(), time.Second)),
							"ModifiedTime":      BeNil(),
							"Revision":          PointTo(Equal(0)),
						})
						result, err := session.Create(ctx, userID, create)
						Expect(err).ToNot(HaveOccurred())
						Expect(result).ToNot(BeNil())
						Expect(*result).To(matchAllFields)
						storeResult := dataSource.Sources{}
						Expect(mgoCollection.Find(bson.M{"id": result.ID}).All(&storeResult)).To(Succeed())
						Expect(storeResult).To(HaveLen(1))
						Expect(*storeResult[0]).To(matchAllFields)
						logger.AssertDebug("Create", log.Fields{"userId": userID, "create": create, "id": *storeResult[0].ID})
					})
				})
			})

			Context("Get", func() {
				var id string

				BeforeEach(func() {
					id = dataSourceTest.RandomID()
				})

				It("returns an error when the context is missing", func() {
					ctx = nil
					result, err := session.Get(ctx, id)
					errorsTest.ExpectEqual(err, errors.New("context is missing"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the id is missing", func() {
					id = ""
					result, err := session.Get(ctx, id)
					errorsTest.ExpectEqual(err, errors.New("id is missing"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the id is invalid", func() {
					id = "invalid"
					result, err := session.Get(ctx, id)
					errorsTest.ExpectEqual(err, errors.New("id is invalid"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the session is closed", func() {
					session.Close()
					result, err := session.Get(ctx, id)
					errorsTest.ExpectEqual(err, errors.New("session closed"))
					Expect(result).To(BeNil())
				})

				Context("with data", func() {
					var allResult dataSource.Sources
					var result *dataSource.Source

					BeforeEach(func() {
						allResult = dataSourceTest.RandomSources(4, 4)
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
						id = dataSourceTest.RandomID()
						Expect(session.Get(ctx, id)).To(BeNil())
					})

					It("returns the result when the id exists", func() {
						Expect(session.Get(ctx, id)).To(Equal(result))
					})

					Context("when the revision is missing", func() {
						BeforeEach(func() {
							result.Revision = nil
						})

						It("returns the result with revision 0", func() {
							result.Revision = pointer.FromInt(0)
							Expect(session.Get(ctx, id)).To(Equal(result))
						})
					})
				})
			})

			Context("Update", func() {
				var id string
				var condition *request.Condition
				var update *dataSource.Update

				BeforeEach(func() {
					id = dataSourceTest.RandomID()
					condition = requestTest.RandomCondition()
					update = dataSourceTest.RandomUpdate()
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
					update.State = pointer.FromString("")
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
					var original *dataSource.Source

					BeforeEach(func() {
						original = dataSourceTest.RandomSource()
						original.ID = pointer.FromString(id)
						Expect(mgoCollection.Insert(original)).To(Succeed())
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
							condition.Revision = pointer.FromInt(*original.Revision + 1)
						})

						It("returns nil", func() {
							Expect(session.Update(ctx, id, condition, update)).To(BeNil())
						})
					})

					updateAssertions := func() {
						Context("with updates", func() {
							It("returns updated result when the id exists and state is connected without error", func() {
								update.ProviderSessionID = pointer.FromString(authTest.RandomProviderSessionID())
								update.State = pointer.FromString(dataSource.StateConnected)
								update.Error = nil
								matchAllFields := MatchAllFields(Fields{
									"ID":                PointTo(Equal(id)),
									"UserID":            Equal(original.UserID),
									"ProviderType":      Equal(original.ProviderType),
									"ProviderName":      Equal(original.ProviderName),
									"ProviderSessionID": Equal(update.ProviderSessionID),
									"State":             Equal(update.State),
									"Error":             Equal(update.Error),
									"DataSetIDs":        Equal(update.DataSetIDs),
									"EarliestDataTime":  PointTo(Equal(update.EarliestDataTime.Truncate(time.Second))),
									"LatestDataTime":    PointTo(Equal(update.LatestDataTime.Truncate(time.Second))),
									"LastImportTime":    PointTo(Equal(update.LastImportTime.Truncate(time.Second))),
									"CreatedTime":       Equal(original.CreatedTime),
									"ModifiedTime":      PointTo(BeTemporally("~", time.Now(), time.Second)),
									"Revision":          PointTo(Equal(*original.Revision + 1)),
								})
								result, err := session.Update(ctx, id, condition, update)
								Expect(err).ToNot(HaveOccurred())
								Expect(result).ToNot(BeNil())
								Expect(*result).To(matchAllFields)
								storeResult := dataSource.Sources{}
								Expect(mgoCollection.Find(bson.M{"id": id}).All(&storeResult)).To(Succeed())
								Expect(storeResult).To(HaveLen(1))
								Expect(*storeResult[0]).To(matchAllFields)
							})

							It("returns updated result when the id exists and state is disconnected without error", func() {
								update.ProviderSessionID = nil
								update.State = pointer.FromString(dataSource.StateDisconnected)
								update.Error = nil
								matchAllFields := MatchAllFields(Fields{
									"ID":                PointTo(Equal(id)),
									"UserID":            Equal(original.UserID),
									"ProviderType":      Equal(original.ProviderType),
									"ProviderName":      Equal(original.ProviderName),
									"ProviderSessionID": BeNil(),
									"State":             Equal(update.State),
									"Error":             Equal(update.Error),
									"DataSetIDs":        Equal(update.DataSetIDs),
									"EarliestDataTime":  PointTo(Equal(update.EarliestDataTime.Truncate(time.Second))),
									"LatestDataTime":    PointTo(Equal(update.LatestDataTime.Truncate(time.Second))),
									"LastImportTime":    PointTo(Equal(update.LastImportTime.Truncate(time.Second))),
									"CreatedTime":       Equal(original.CreatedTime),
									"ModifiedTime":      PointTo(BeTemporally("~", time.Now(), time.Second)),
									"Revision":          PointTo(Equal(*original.Revision + 1)),
								})
								result, err := session.Update(ctx, id, condition, update)
								Expect(err).ToNot(HaveOccurred())
								Expect(result).ToNot(BeNil())
								Expect(*result).To(matchAllFields)
								storeResult := dataSource.Sources{}
								Expect(mgoCollection.Find(bson.M{"id": id}).All(&storeResult)).To(Succeed())
								Expect(storeResult).To(HaveLen(1))
								Expect(*storeResult[0]).To(matchAllFields)
							})

							It("returns updated result when the id exists and state is error with error", func() {
								update.ProviderSessionID = nil
								update.State = pointer.FromString(dataSource.StateError)
								matchAllFields := MatchAllFields(Fields{
									"ID":                PointTo(Equal(id)),
									"UserID":            Equal(original.UserID),
									"ProviderType":      Equal(original.ProviderType),
									"ProviderName":      Equal(original.ProviderName),
									"ProviderSessionID": Equal(original.ProviderSessionID),
									"State":             Equal(update.State),
									"Error":             Equal(update.Error),
									"DataSetIDs":        Equal(update.DataSetIDs),
									"EarliestDataTime":  PointTo(Equal(update.EarliestDataTime.Truncate(time.Second))),
									"LatestDataTime":    PointTo(Equal(update.LatestDataTime.Truncate(time.Second))),
									"LastImportTime":    PointTo(Equal(update.LastImportTime.Truncate(time.Second))),
									"CreatedTime":       Equal(original.CreatedTime),
									"ModifiedTime":      PointTo(BeTemporally("~", time.Now(), time.Second)),
									"Revision":          PointTo(Equal(*original.Revision + 1)),
								})
								result, err := session.Update(ctx, id, condition, update)
								Expect(err).ToNot(HaveOccurred())
								Expect(result).ToNot(BeNil())
								Expect(*result).To(matchAllFields)
								storeResult := dataSource.Sources{}
								Expect(mgoCollection.Find(bson.M{"id": id}).All(&storeResult)).To(Succeed())
								Expect(storeResult).To(HaveLen(1))
								Expect(*storeResult[0]).To(matchAllFields)
							})

							It("returns nil when the id does not exist", func() {
								id = dataSourceTest.RandomID()
								Expect(session.Update(ctx, id, condition, update)).To(BeNil())
							})
						})

						Context("without updates", func() {
							BeforeEach(func() {
								update = dataSource.NewUpdate()
							})

							It("returns original when the id exists", func() {
								Expect(session.Update(ctx, id, condition, update)).To(Equal(original))
							})

							It("returns nil when the id does not exist", func() {
								id = dataSourceTest.RandomID()
								Expect(session.Update(ctx, id, condition, update)).To(BeNil())
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
							condition.Revision = pointer.CloneInt(original.Revision)
						})

						updateAssertions()
					})
				})
			})

			Context("Destroy", func() {
				var id string
				var condition *request.Condition

				BeforeEach(func() {
					id = dataSourceTest.RandomID()
					condition = requestTest.RandomCondition()
				})

				It("returns an error when the context is missing", func() {
					ctx = nil
					deleted, err := session.Destroy(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("context is missing"))
					Expect(deleted).To(BeFalse())
				})

				It("returns an error when the id is missing", func() {
					id = ""
					deleted, err := session.Destroy(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("id is missing"))
					Expect(deleted).To(BeFalse())
				})

				It("returns an error when the id is invalid", func() {
					id = "invalid"
					deleted, err := session.Destroy(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("id is invalid"))
					Expect(deleted).To(BeFalse())
				})

				It("returns an error when the condition is invalid", func() {
					condition.Revision = pointer.FromInt(-1)
					deleted, err := session.Destroy(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("condition is invalid"))
					Expect(deleted).To(BeFalse())
				})

				It("returns an error when the session is closed", func() {
					session.Close()
					deleted, err := session.Destroy(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("session closed"))
					Expect(deleted).To(BeFalse())
				})

				Context("with data", func() {
					var original *dataSource.Source

					BeforeEach(func() {
						original = dataSourceTest.RandomSource()
						original.ID = pointer.FromString(id)
						Expect(mgoCollection.Insert(original)).To(Succeed())
					})

					AfterEach(func() {
						if condition != nil {
							logger.AssertDebug("Destroy", log.Fields{"id": id, "condition": condition})
						} else {
							logger.AssertDebug("Destroy", log.Fields{"id": id})
						}
					})

					It("returns false and does not destroy the original when the id does not exist", func() {
						id = dataSourceTest.RandomID()
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
				})
			})
		})
	})
})
