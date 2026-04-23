package client_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.uber.org/mock/gomock"

	authTest "github.com/tidepool-org/platform/auth/test"
	dataSource "github.com/tidepool-org/platform/data/source"
	dataSourceServiceClient "github.com/tidepool-org/platform/data/source/service/client"
	dataSourceServiceClientTest "github.com/tidepool-org/platform/data/source/service/client/test"
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
	"github.com/tidepool-org/platform/test"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("Client", func() {
	var ctx context.Context
	var mockController *gomock.Controller
	var mockProvider *dataSourceServiceClientTest.MockProvider

	BeforeEach(func() {
		ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
		mockController, ctx = gomock.WithContext(ctx, GinkgoT())
		mockProvider = dataSourceServiceClientTest.NewMockProvider(mockController)
	})

	Context("New", func() {
		It("returns an error when the client provider is missing", func() {
			client, err := dataSourceServiceClient.New(nil)
			errorsTest.ExpectEqual(err, errors.New("provider is missing"))
			Expect(client).To(BeNil())
		})

		It("returns successfully", func() {
			Expect(dataSourceServiceClient.New(mockProvider)).ToNot(BeNil())
		})
	})

	Context("with new client", func() {
		var mockDataSourcesRepository *dataSourceStoreStructuredTest.MockDataSourcesRepository
		var mockStore *dataSourceStoreStructuredTest.MockStore
		var client *dataSourceServiceClient.Client

		BeforeEach(func() {
			mockDataSourcesRepository = dataSourceStoreStructuredTest.NewMockDataSourcesRepository(mockController)
			mockStore = dataSourceStoreStructuredTest.NewMockStore(mockController)

			mockStore.EXPECT().NewDataSourcesRepository().Return(mockDataSourcesRepository)
			mockProvider.EXPECT().DataSourceStructuredStore().Return(mockStore)

			var err error
			client, err = dataSourceServiceClient.New(mockProvider)
			Expect(err).ToNot(HaveOccurred())
			Expect(client).ToNot(BeNil())
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
					filter = dataSourceTest.RandomFilter(test.AllowOptionals())
					pagination = pageTest.RandomPagination()
				})

				It("returns an error when the data source structured repository list returns an error", func() {
					responseErr := errorsTest.RandomError()
					mockDataSourcesRepository.EXPECT().List(ctx, userID, filter, pagination).Return(nil, responseErr)
					result, err := client.List(ctx, userID, filter, pagination)
					errorsTest.ExpectEqual(err, responseErr)
					Expect(result).To(BeNil())
				})

				It("returns successfully when the data source structured repository list returns successfully", func() {
					responseResult := dataSourceTest.RandomSourceArray(1, 3, test.AllowOptionals())
					mockDataSourcesRepository.EXPECT().List(ctx, userID, filter, pagination).Return(responseResult, nil)
					result, err := client.List(ctx, userID, filter, pagination)
					Expect(err).ToNot(HaveOccurred())
					Expect(result).To(Equal(responseResult))
				})
			})

			Context("Create", func() {
				var create *dataSource.Create

				BeforeEach(func() {
					create = dataSourceTest.RandomCreate(test.AllowOptionals())
				})

				It("returns an error when the data source structured repository create returns an error", func() {
					responseErr := errorsTest.RandomError()
					mockDataSourcesRepository.EXPECT().Create(ctx, userID, create).Return(nil, responseErr)
					result, err := client.Create(ctx, userID, create)
					errorsTest.ExpectEqual(err, responseErr)
					Expect(result).To(BeNil())
				})

				It("returns successfully when the data source structured repository create returns successfully", func() {
					responseResult := dataSourceTest.RandomSource(test.AllowOptionals())
					mockDataSourcesRepository.EXPECT().Create(ctx, userID, create).Return(responseResult, nil)
					result, err := client.Create(ctx, userID, create)
					Expect(err).ToNot(HaveOccurred())
					Expect(result).To(Equal(responseResult))
				})
			})

			Context("DeleteAll", func() {
				It("returns an error when the data source structured repository delete all returns an error", func() {
					responseErr := errorsTest.RandomError()
					mockDataSourcesRepository.EXPECT().DeleteAll(ctx, userID).Return(responseErr)
					errorsTest.ExpectEqual(client.DeleteAll(ctx, userID), responseErr)
				})

				It("returns successfully when the data source structured repository delete all returns successfully", func() {
					mockDataSourcesRepository.EXPECT().DeleteAll(ctx, userID).Return(nil)
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
				It("returns an error when the data source structured repository get returns an error", func() {
					responseErr := errorsTest.RandomError()
					mockDataSourcesRepository.EXPECT().Get(ctx, id).Return(nil, responseErr)
					result, err := client.Get(ctx, id)
					errorsTest.ExpectEqual(err, responseErr)
					Expect(result).To(BeNil())
				})

				It("returns successfully when the user client ensure authorized user returns successfully", func() {
					responseResult := dataSourceTest.RandomSource(test.AllowOptionals())
					mockDataSourcesRepository.EXPECT().Get(ctx, id).Return(responseResult, nil)
					result, err := client.Get(ctx, id)
					Expect(err).ToNot(HaveOccurred())
					Expect(result).To(Equal(responseResult))
				})
			})

			Context("Update", func() {
				var condition *request.Condition
				var update *dataSource.Update

				BeforeEach(func() {
					condition = requestTest.RandomCondition()
					update = dataSourceTest.RandomUpdate(test.AllowOptionals())
				})

				It("returns an error when the data source structured repository update returns an error", func() {
					responseErr := errorsTest.RandomError()
					mockDataSourcesRepository.EXPECT().Update(ctx, id, condition, update).Return(nil, responseErr)
					result, err := client.Update(ctx, id, condition, update)
					errorsTest.ExpectEqual(err, responseErr)
					Expect(result).To(BeNil())
				})

				It("returns successfully when the data source structured repository update returns successfully", func() {
					responseResult := dataSourceTest.RandomSource(test.AllowOptionals())
					mockDataSourcesRepository.EXPECT().Update(ctx, id, condition, update).Return(responseResult, nil)
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

				It("returns an error when the data source structured repository delete returns an error", func() {
					responseErr := errorsTest.RandomError()
					mockDataSourcesRepository.EXPECT().Delete(ctx, id, condition).Return(false, responseErr)
					deleted, err := client.Delete(ctx, id, condition)
					errorsTest.ExpectEqual(err, responseErr)
					Expect(deleted).To(BeFalse())
				})

				It("returns successfully when the data source structured repository delete returns successfully without deleted", func() {
					mockDataSourcesRepository.EXPECT().Delete(ctx, id, condition).Return(false, nil)
					deleted, err := client.Delete(ctx, id, condition)
					Expect(err).ToNot(HaveOccurred())
					Expect(deleted).To(BeFalse())
				})

				It("returns successfully when the data source structured repository delete returns successfully with deleted", func() {
					mockDataSourcesRepository.EXPECT().Delete(ctx, id, condition).Return(true, nil)
					deleted, err := client.Delete(ctx, id, condition)
					Expect(err).ToNot(HaveOccurred())
					Expect(deleted).To(BeTrue())
				})
			})
		})

		Context("with provider session id", func() {
			var providerSessionID string

			BeforeEach(func() {
				providerSessionID = authTest.RandomProviderSessionID()
			})

			Context("GetFromProviderSession", func() {
				It("returns an error when the data source structured repository get returns an error", func() {
					responseErr := errorsTest.RandomError()
					mockDataSourcesRepository.EXPECT().GetFromProviderSession(ctx, providerSessionID).Return(nil, responseErr)
					result, err := client.GetFromProviderSession(ctx, providerSessionID)
					errorsTest.ExpectEqual(err, responseErr)
					Expect(result).To(BeNil())
				})

				It("returns successfully when the user client ensure authorized user returns successfully", func() {
					responseResult := dataSourceTest.RandomSource(test.AllowOptionals())
					mockDataSourcesRepository.EXPECT().GetFromProviderSession(ctx, providerSessionID).Return(responseResult, nil)
					result, err := client.GetFromProviderSession(ctx, providerSessionID)
					Expect(err).ToNot(HaveOccurred())
					Expect(result).To(Equal(responseResult))
				})
			})
		})
	})
})
