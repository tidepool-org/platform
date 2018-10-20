package client_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/auth"
	authTest "github.com/tidepool-org/platform/auth/test"
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
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/request"
	requestTest "github.com/tidepool-org/platform/request/test"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("Client", func() {
	var authClient *authTest.Client
	var dataSourceStructuredStore *dataSourceStoreStructuredTest.Store
	var dataSourceStructuredSession *dataSourceStoreStructuredTest.Session
	var provider *dataSourceServiceClientTest.Provider

	BeforeEach(func() {
		authClient = authTest.NewClient()
		dataSourceStructuredStore = dataSourceStoreStructuredTest.NewStore()
		dataSourceStructuredSession = dataSourceStoreStructuredTest.NewSession()
		dataSourceStructuredSession.CloseOutput = func(err error) *error { return &err }(nil)
		dataSourceStructuredStore.NewSessionOutput = func(s dataSourceStoreStructured.Session) *dataSourceStoreStructured.Session { return &s }(dataSourceStructuredSession)
		provider = dataSourceServiceClientTest.NewProvider()
		provider.AuthClientOutput = func(u auth.Client) *auth.Client { return &u }(authClient)
		provider.DataSourceStructuredStoreOutput = func(s dataSourceStoreStructured.Store) *dataSourceStoreStructured.Store { return &s }(dataSourceStructuredStore)
	})

	AfterEach(func() {
		provider.AssertOutputsEmpty()
		dataSourceStructuredStore.AssertOutputsEmpty()
		authClient.AssertOutputsEmpty()
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
				userID = userTest.RandomID()
			})

			Context("List", func() {
				var filter *dataSource.Filter
				var pagination *page.Pagination

				BeforeEach(func() {
					filter = dataSourceTest.RandomFilter()
					pagination = pageTest.RandomPagination()
				})

				AfterEach(func() {
					Expect(authClient.EnsureAuthorizedUserInputs).To(Equal([]authTest.EnsureAuthorizedUserInput{{Context: ctx, TargetUserID: userID, AuthorizedPermission: permission.Owner}}))
				})

				It("return an error when the user client ensure authorized user returns an error", func() {
					responseErr := errorsTest.RandomError()
					authClient.EnsureAuthorizedUserOutputs = []authTest.EnsureAuthorizedUserOutput{{AuthorizedUserID: "", Error: responseErr}}
					result, err := client.List(ctx, userID, filter, pagination)
					errorsTest.ExpectEqual(err, responseErr)
					Expect(result).To(BeNil())
				})

				When("user client ensure authorized user returns successfully", func() {
					BeforeEach(func() {
						authClient.EnsureAuthorizedUserOutputs = []authTest.EnsureAuthorizedUserOutput{{AuthorizedUserID: userTest.RandomID(), Error: nil}}
					})

					AfterEach(func() {
						Expect(dataSourceStructuredSession.ListInputs).To(Equal([]dataSourceStoreStructuredTest.ListInput{{Context: ctx, UserID: userID, Filter: filter, Pagination: pagination}}))
					})

					It("returns an error when the data source structured session list returns an error", func() {
						responseErr := errorsTest.RandomError()
						dataSourceStructuredSession.ListOutputs = []dataSourceStoreStructuredTest.ListOutput{{Sources: nil, Error: responseErr}}
						result, err := client.List(ctx, userID, filter, pagination)
						errorsTest.ExpectEqual(err, responseErr)
						Expect(result).To(BeNil())
					})

					It("returns successfully when the data source structured session list returns successfully", func() {
						responseResult := dataSourceTest.RandomSources(1, 3)
						dataSourceStructuredSession.ListOutputs = []dataSourceStoreStructuredTest.ListOutput{{Sources: responseResult, Error: nil}}
						result, err := client.List(ctx, userID, filter, pagination)
						Expect(err).ToNot(HaveOccurred())
						Expect(result).To(Equal(responseResult))
					})
				})
			})

			Context("Create", func() {
				var create *dataSource.Create

				BeforeEach(func() {
					create = dataSourceTest.RandomCreate()
				})

				AfterEach(func() {
					Expect(authClient.EnsureAuthorizedServiceInputs).To(Equal([]context.Context{ctx}))
				})

				It("return an error when the user client ensure authorized service returns an error", func() {
					responseErr := errorsTest.RandomError()
					authClient.EnsureAuthorizedServiceOutputs = []error{responseErr}
					result, err := client.Create(ctx, userID, create)
					errorsTest.ExpectEqual(err, responseErr)
					Expect(result).To(BeNil())
				})

				When("user client ensure authorized service returns successfully", func() {
					BeforeEach(func() {
						authClient.EnsureAuthorizedServiceOutputs = []error{nil}
					})

					AfterEach(func() {
						Expect(dataSourceStructuredSession.CreateInputs).To(Equal([]dataSourceStoreStructuredTest.CreateInput{{Context: ctx, UserID: userID, Create: create}}))
					})

					It("returns an error when the data source structured session create returns an error", func() {
						responseErr := errorsTest.RandomError()
						dataSourceStructuredSession.CreateOutputs = []dataSourceStoreStructuredTest.CreateOutput{{Source: nil, Error: responseErr}}
						result, err := client.Create(ctx, userID, create)
						errorsTest.ExpectEqual(err, responseErr)
						Expect(result).To(BeNil())
					})

					It("returns successfully when the data source structured session create returns successfully", func() {
						responseResult := dataSourceTest.RandomSource()
						dataSourceStructuredSession.CreateOutputs = []dataSourceStoreStructuredTest.CreateOutput{{Source: responseResult, Error: nil}}
						result, err := client.Create(ctx, userID, create)
						Expect(err).ToNot(HaveOccurred())
						Expect(result).To(Equal(responseResult))
					})
				})
			})
		})

		Context("with id", func() {
			var id string

			BeforeEach(func() {
				id = dataSourceTest.RandomID()
			})

			Context("Get", func() {
				AfterEach(func() {
					Expect(authClient.EnsureAuthorizedInputs).To(Equal([]context.Context{ctx}))
				})

				It("returns an error when the user client ensure authorized returns an error", func() {
					responseErr := errorsTest.RandomError()
					authClient.EnsureAuthorizedOutputs = []error{responseErr}
					result, err := client.Get(ctx, id)
					errorsTest.ExpectEqual(err, responseErr)
					Expect(result).To(BeNil())
				})

				When("user client ensure authorized returns successfully", func() {
					BeforeEach(func() {
						authClient.EnsureAuthorizedOutputs = []error{nil}
					})

					AfterEach(func() {
						Expect(dataSourceStructuredSession.GetInputs).To(Equal([]dataSourceStoreStructuredTest.GetInput{{Context: ctx, ID: id}}))
					})

					It("returns an error when the data source structured session get returns an error", func() {
						responseErr := errorsTest.RandomError()
						dataSourceStructuredSession.GetOutputs = []dataSourceStoreStructuredTest.GetOutput{{Source: nil, Error: responseErr}}
						result, err := client.Get(ctx, id)
						errorsTest.ExpectEqual(err, responseErr)
						Expect(result).To(BeNil())
					})

					When("data source structured session get returns successfully", func() {
						var responseResult *dataSource.Source

						BeforeEach(func() {
							responseResult = dataSourceTest.RandomSource()
							dataSourceStructuredSession.GetOutputs = []dataSourceStoreStructuredTest.GetOutput{{Source: responseResult, Error: nil}}
						})

						AfterEach(func() {
							Expect(authClient.EnsureAuthorizedUserInputs).To(Equal([]authTest.EnsureAuthorizedUserInput{{Context: ctx, TargetUserID: *responseResult.UserID, AuthorizedPermission: permission.Owner}}))
						})

						It("returns an error when the user client ensure authorized user returns an error", func() {
							responseErr := errorsTest.RandomError()
							authClient.EnsureAuthorizedUserOutputs = []authTest.EnsureAuthorizedUserOutput{{AuthorizedUserID: "", Error: responseErr}}
							result, err := client.Get(ctx, id)
							errorsTest.ExpectEqual(err, responseErr)
							Expect(result).To(BeNil())
						})

						It("returns successfully when the user client ensure authorized user returns successfully", func() {
							authClient.EnsureAuthorizedUserOutputs = []authTest.EnsureAuthorizedUserOutput{{AuthorizedUserID: userTest.RandomID(), Error: nil}}
							result, err := client.Get(ctx, id)
							Expect(err).ToNot(HaveOccurred())
							Expect(result).To(Equal(responseResult))
						})
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
					Expect(authClient.EnsureAuthorizedServiceInputs).To(Equal([]context.Context{ctx}))
				})

				It("return an error when the user client ensure authorized service returns an error", func() {
					responseErr := errorsTest.RandomError()
					authClient.EnsureAuthorizedServiceOutputs = []error{responseErr}
					result, err := client.Update(ctx, id, condition, update)
					errorsTest.ExpectEqual(err, responseErr)
					Expect(result).To(BeNil())
				})

				When("user client ensure authorized service returns successfully", func() {
					BeforeEach(func() {
						authClient.EnsureAuthorizedServiceOutputs = []error{nil}
					})

					AfterEach(func() {
						Expect(dataSourceStructuredSession.UpdateInputs).To(Equal([]dataSourceStoreStructuredTest.UpdateInput{{Context: ctx, ID: id, Condition: condition, Update: update}}))
					})

					It("returns an error when the data source structured session update returns an error", func() {
						responseErr := errorsTest.RandomError()
						dataSourceStructuredSession.UpdateOutputs = []dataSourceStoreStructuredTest.UpdateOutput{{Source: nil, Error: responseErr}}
						result, err := client.Update(ctx, id, condition, update)
						errorsTest.ExpectEqual(err, responseErr)
						Expect(result).To(BeNil())
					})

					It("returns successfully when the data source structured session update returns successfully", func() {
						responseResult := dataSourceTest.RandomSource()
						dataSourceStructuredSession.UpdateOutputs = []dataSourceStoreStructuredTest.UpdateOutput{{Source: responseResult, Error: nil}}
						result, err := client.Update(ctx, id, condition, update)
						Expect(err).ToNot(HaveOccurred())
						Expect(result).To(Equal(responseResult))
					})
				})
			})

			Context("Delete", func() {
				var condition *request.Condition

				BeforeEach(func() {
					condition = requestTest.RandomCondition()
				})

				AfterEach(func() {
					Expect(authClient.EnsureAuthorizedServiceInputs).To(Equal([]context.Context{ctx}))
				})

				It("return an error when the user client ensure authorized service returns an error", func() {
					responseErr := errorsTest.RandomError()
					authClient.EnsureAuthorizedServiceOutputs = []error{responseErr}
					deleted, err := client.Delete(ctx, id, condition)
					errorsTest.ExpectEqual(err, responseErr)
					Expect(deleted).To(BeFalse())
				})

				When("user client ensure authorized service returns successfully", func() {
					BeforeEach(func() {
						authClient.EnsureAuthorizedServiceOutputs = []error{nil}
					})

					AfterEach(func() {
						Expect(dataSourceStructuredSession.DeleteInputs).To(Equal([]dataSourceStoreStructuredTest.DeleteInput{{Context: ctx, ID: id, Condition: condition}}))
					})

					It("returns an error when the data source structured session delete returns an error", func() {
						responseErr := errorsTest.RandomError()
						dataSourceStructuredSession.DeleteOutputs = []dataSourceStoreStructuredTest.DeleteOutput{{Deleted: false, Error: responseErr}}
						deleted, err := client.Delete(ctx, id, condition)
						errorsTest.ExpectEqual(err, responseErr)
						Expect(deleted).To(BeFalse())
					})

					It("returns successfully when the data source structured session delete returns successfully without deleted", func() {
						dataSourceStructuredSession.DeleteOutputs = []dataSourceStoreStructuredTest.DeleteOutput{{Deleted: false, Error: nil}}
						deleted, err := client.Delete(ctx, id, condition)
						Expect(err).ToNot(HaveOccurred())
						Expect(deleted).To(BeFalse())
					})

					It("returns successfully when the data source structured session delete returns successfully with deleted", func() {
						dataSourceStructuredSession.DeleteOutputs = []dataSourceStoreStructuredTest.DeleteOutput{{Deleted: true, Error: nil}}
						deleted, err := client.Delete(ctx, id, condition)
						Expect(err).ToNot(HaveOccurred())
						Expect(deleted).To(BeTrue())
					})
				})
			})
		})
	})
})
