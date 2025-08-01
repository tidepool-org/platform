package client_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataSource "github.com/tidepool-org/platform/data/source"
	dataSourceServiceClient "github.com/tidepool-org/platform/data/source/service/client"
	dataSourceServiceClientTest "github.com/tidepool-org/platform/data/source/service/client/test"
	dataSourceStoreStructured "github.com/tidepool-org/platform/data/source/store/structured"
	dataSourceStoreStructuredTest "github.com/tidepool-org/platform/data/source/store/structured/test"
	dataSourceTest "github.com/tidepool-org/platform/data/source/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/page"
	pageTest "github.com/tidepool-org/platform/page/test"
	"github.com/tidepool-org/platform/request"
	requestTest "github.com/tidepool-org/platform/request/test"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("Client", func() {
	var dataSourceStructuredStore *dataSourceStoreStructuredTest.Store
	var dataSourceStructuredRepository *dataSourceStoreStructuredTest.DataRepository
	var provider *dataSourceServiceClientTest.Provider

	BeforeEach(func() {
		dataSourceStructuredStore = dataSourceStoreStructuredTest.NewStore()
		dataSourceStructuredRepository = dataSourceStoreStructuredTest.NewDataSourcesRepository()
		dataSourceStructuredRepository.CloseOutput = func(err error) *error { return &err }(nil)
		dataSourceStructuredStore.NewDataSourcesOutput = func(s dataSourceStoreStructured.DataSourcesRepository) *dataSourceStoreStructured.DataSourcesRepository {
			return &s
		}(dataSourceStructuredRepository)
		provider = dataSourceServiceClientTest.NewProvider()
		provider.DataSourceStructuredStoreOutput = func(s dataSourceStoreStructured.Store) *dataSourceStoreStructured.Store { return &s }(dataSourceStructuredStore)
	})

	AfterEach(func() {
		provider.AssertOutputsEmpty()
		dataSourceStructuredStore.AssertOutputsEmpty()
	})

	Context("New", func() {
		It("returns an error when the client provider is missing", func() {
			client, err := dataSourceServiceClient.New(nil)
			errorsTest.ExpectEqual(err, errors.New("provider is missing"))
			Expect(client).To(BeNil())
		})

		It("returns successfully", func() {
			Expect(dataSourceServiceClient.New(provider)).ToNot(BeNil())
		})
	})

	Context("with new client", func() {
		var client *dataSourceServiceClient.Client
		var logger *logTest.Logger
		var ctx context.Context

		BeforeEach(func() {
			var err error
			client, err = dataSourceServiceClient.New(provider)
			Expect(err).ToNot(HaveOccurred())
			Expect(client).ToNot(BeNil())
			logger = logTest.NewLogger()
			ctx = context.Background()
			ctx = log.NewContextWithLogger(ctx, logger)
		})

		Context("with user id", func() {
			var userID string

			BeforeEach(func() {
				userID = userTest.RandomUserID()
			})

			Context("List", func() {
				var filter *dataSource.Filter
				var pagination *page.Pagination

				BeforeEach(func() {
					filter = dataSourceTest.RandomFilter()
					pagination = pageTest.RandomPagination()
				})

				AfterEach(func() {
					Expect(dataSourceStructuredRepository.ListInputs).To(Equal([]dataSourceStoreStructuredTest.ListInput{{UserID: userID, Filter: filter, Pagination: pagination}}))
				})

				It("returns an error when the data source structured repository list returns an error", func() {
					responseErr := errorsTest.RandomError()
					dataSourceStructuredRepository.ListOutputs = []dataSourceStoreStructuredTest.ListOutput{{SourceArray: nil, Error: responseErr}}
					result, err := client.List(ctx, userID, filter, pagination)
					errorsTest.ExpectEqual(err, responseErr)
					Expect(result).To(BeNil())
				})

				It("returns successfully when the data source structured repository list returns successfully", func() {
					responseResult := dataSourceTest.RandomSourceArray(1, 3)
					dataSourceStructuredRepository.ListOutputs = []dataSourceStoreStructuredTest.ListOutput{{SourceArray: responseResult, Error: nil}}
					result, err := client.List(ctx, userID, filter, pagination)
					Expect(err).ToNot(HaveOccurred())
					Expect(result).To(Equal(responseResult))
				})
			})

			Context("Create", func() {
				var create *dataSource.Create

				BeforeEach(func() {
					create = dataSourceTest.RandomCreate()
				})

				AfterEach(func() {
					Expect(dataSourceStructuredRepository.CreateInputs).To(Equal([]dataSourceStoreStructuredTest.CreateInput{{UserID: userID, Create: create}}))
				})

				It("returns an error when the data source structured repository create returns an error", func() {
					responseErr := errorsTest.RandomError()
					dataSourceStructuredRepository.CreateOutputs = []dataSourceStoreStructuredTest.CreateOutput{{Source: nil, Error: responseErr}}
					result, err := client.Create(ctx, userID, create)
					errorsTest.ExpectEqual(err, responseErr)
					Expect(result).To(BeNil())
				})

				It("returns successfully when the data source structured repository create returns successfully", func() {
					responseResult := dataSourceTest.RandomSource()
					dataSourceStructuredRepository.CreateOutputs = []dataSourceStoreStructuredTest.CreateOutput{{Source: responseResult, Error: nil}}
					result, err := client.Create(ctx, userID, create)
					Expect(err).ToNot(HaveOccurred())
					Expect(result).To(Equal(responseResult))
				})
			})

			Context("DeleteAll", func() {
				AfterEach(func() {
					Expect(dataSourceStructuredRepository.DestroyAllInputs).To(Equal([]string{userID}))
				})

				It("returns an error when the data source structured repository destroy returns an error", func() {
					responseErr := errorsTest.RandomError()
					dataSourceStructuredRepository.DestroyAllOutputs = []dataSourceStoreStructuredTest.DestroyAllOutput{{Destroyed: false, Error: responseErr}}
					errorsTest.ExpectEqual(client.DeleteAll(ctx, userID), responseErr)
				})

				It("returns successfully when the data source structured repository destroy returns false", func() {
					dataSourceStructuredRepository.DestroyAllOutputs = []dataSourceStoreStructuredTest.DestroyAllOutput{{Destroyed: false, Error: nil}}
					Expect(client.DeleteAll(ctx, userID)).To(Succeed())
				})

				It("returns successfully when the data source structured repository destroy returns true", func() {
					dataSourceStructuredRepository.DestroyAllOutputs = []dataSourceStoreStructuredTest.DestroyAllOutput{{Destroyed: true, Error: nil}}
					Expect(client.DeleteAll(ctx, userID)).To(Succeed())
				})
			})
		})

		Context("with id", func() {
			var id string

			BeforeEach(func() {
				id = dataSourceTest.RandomDataSourceID()
			})

			Context("Get", func() {
				AfterEach(func() {
					Expect(dataSourceStructuredRepository.GetInputs).To(Equal([]string{id}))
				})

				It("returns an error when the data source structured repository get returns an error", func() {
					responseErr := errorsTest.RandomError()
					dataSourceStructuredRepository.GetOutputs = []dataSourceStoreStructuredTest.GetOutput{{Source: nil, Error: responseErr}}
					result, err := client.Get(ctx, id)
					errorsTest.ExpectEqual(err, responseErr)
					Expect(result).To(BeNil())
				})

				When("data source structured repository get returns successfully", func() {
					var responseResult *dataSource.Source

					BeforeEach(func() {
						responseResult = dataSourceTest.RandomSource()
						dataSourceStructuredRepository.GetOutputs = []dataSourceStoreStructuredTest.GetOutput{{Source: responseResult, Error: nil}}
					})

					It("returns successfully when the user client ensure authorized user returns successfully", func() {
						result, err := client.Get(ctx, id)
						Expect(err).ToNot(HaveOccurred())
						Expect(result).To(Equal(responseResult))
					})
				})
			})

			Context("Update", func() {
				var condition *request.Condition
				var update *dataSource.Update

				BeforeEach(func() {
					condition = requestTest.RandomCondition()
					update = dataSourceTest.RandomUpdate()
				})

				AfterEach(func() {
					Expect(dataSourceStructuredRepository.UpdateInputs).To(Equal([]dataSourceStoreStructuredTest.UpdateInput{{ID: id, Condition: condition, Update: update}}))
				})

				It("returns an error when the data source structured repository update returns an error", func() {
					responseErr := errorsTest.RandomError()
					dataSourceStructuredRepository.UpdateOutputs = []dataSourceStoreStructuredTest.UpdateOutput{{Source: nil, Error: responseErr}}
					result, err := client.Update(ctx, id, condition, update)
					errorsTest.ExpectEqual(err, responseErr)
					Expect(result).To(BeNil())
				})

				It("returns successfully when the data source structured repository update returns successfully", func() {
					responseResult := dataSourceTest.RandomSource()
					dataSourceStructuredRepository.UpdateOutputs = []dataSourceStoreStructuredTest.UpdateOutput{{Source: responseResult, Error: nil}}
					result, err := client.Update(ctx, id, condition, update)
					Expect(err).ToNot(HaveOccurred())
					Expect(result).To(Equal(responseResult))
				})
			})

			Context("Delete", func() {
				var condition *request.Condition

				BeforeEach(func() {
					condition = requestTest.RandomCondition()
				})

				AfterEach(func() {
					Expect(dataSourceStructuredRepository.DestroyInputs).To(Equal([]dataSourceStoreStructuredTest.DestroyInput{{ID: id, Condition: condition}}))
				})

				It("returns an error when the data source structured repository delete returns an error", func() {
					responseErr := errorsTest.RandomError()
					dataSourceStructuredRepository.DestroyOutputs = []dataSourceStoreStructuredTest.DestroyOutput{{Destroyed: false, Error: responseErr}}
					deleted, err := client.Delete(ctx, id, condition)
					errorsTest.ExpectEqual(err, responseErr)
					Expect(deleted).To(BeFalse())
				})

				It("returns successfully when the data source structured repository delete returns successfully without destroyed", func() {
					dataSourceStructuredRepository.DestroyOutputs = []dataSourceStoreStructuredTest.DestroyOutput{{Destroyed: false, Error: nil}}
					deleted, err := client.Delete(ctx, id, condition)
					Expect(err).ToNot(HaveOccurred())
					Expect(deleted).To(BeFalse())
				})

				It("returns successfully when the data source structured repository delete returns successfully with destroyed", func() {
					dataSourceStructuredRepository.DestroyOutputs = []dataSourceStoreStructuredTest.DestroyOutput{{Destroyed: true, Error: nil}}
					deleted, err := client.Delete(ctx, id, condition)
					Expect(err).ToNot(HaveOccurred())
					Expect(deleted).To(BeTrue())
				})
			})
		})

		Context("ListAll", func() {
			var filter *dataSource.Filter
			var pagination *page.Pagination

			BeforeEach(func() {
				filter = dataSourceTest.RandomFilter()
				pagination = pageTest.RandomPagination()
			})

			When("the user client ensure authorized user returns successfully", func() {
				AfterEach(func() {
					Expect(dataSourceStructuredRepository.ListAllInputs).To(Equal([]dataSourceStoreStructuredTest.ListAllInput{{Filter: filter, Pagination: pagination}}))
				})

				It("returns an error when the data source structured repository list returns an error", func() {
					responseErr := errorsTest.RandomError()
					dataSourceStructuredRepository.ListAllOutputs = []dataSourceStoreStructuredTest.ListAllOutput{{SourceArray: nil, Error: responseErr}}
					result, err := client.ListAll(ctx, filter, pagination)
					errorsTest.ExpectEqual(err, responseErr)
					Expect(result).To(BeNil())
				})

				It("returns successfully when the data source structured repository list returns successfully", func() {
					responseResult := dataSourceTest.RandomSourceArray(1, 3)
					dataSourceStructuredRepository.ListAllOutputs = []dataSourceStoreStructuredTest.ListAllOutput{{SourceArray: responseResult, Error: nil}}
					result, err := client.ListAll(ctx, filter, pagination)
					Expect(err).ToNot(HaveOccurred())
					Expect(result).To(Equal(responseResult))
				})
			})
		})
	})
})
