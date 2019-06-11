package client_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/auth"
	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/blob"
	blobTest "github.com/tidepool-org/platform/blob/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	requestTest "github.com/tidepool-org/platform/request/test"
	"github.com/tidepool-org/platform/test"
	"github.com/tidepool-org/platform/user"
	userServiceClient "github.com/tidepool-org/platform/user/service/client"
	userServiceClientTest "github.com/tidepool-org/platform/user/service/client/test"
	userStoreStructured "github.com/tidepool-org/platform/user/store/structured"
	userStoreStructuredTest "github.com/tidepool-org/platform/user/store/structured/test"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("Client", func() {
	var authClient *authTest.Client
	var blobClient *blobTest.Client
	var userStructuredStore *userStoreStructuredTest.Store
	var userStructuredSession *userStoreStructuredTest.Session
	var passwordHasher *userServiceClientTest.PasswordHasher
	var provider *userServiceClientTest.Provider

	BeforeEach(func() {
		authClient = authTest.NewClient()
		blobClient = blobTest.NewClient()

		userStructuredStore = userStoreStructuredTest.NewStore()
		userStructuredSession = userStoreStructuredTest.NewSession()
		userStructuredSession.CloseOutput = func(err error) *error { return &err }(nil)
		userStructuredStore.NewSessionOutput = func(s userStoreStructured.Session) *userStoreStructured.Session { return &s }(userStructuredSession)

		passwordHasher = userServiceClientTest.NewPasswordHasher()

		provider = userServiceClientTest.NewProvider()
		provider.AuthClientOutput = func(c auth.Client) *auth.Client { return &c }(authClient)
		provider.BlobClientOutput = func(c blob.Client) *blob.Client { return &c }(blobClient)
		provider.UserStructuredStoreOutput = func(s userStoreStructured.Store) *userStoreStructured.Store { return &s }(userStructuredStore)
		provider.PasswordHasherOutput = func(p userServiceClient.PasswordHasher) *userServiceClient.PasswordHasher { return &p }(passwordHasher)
	})

	AfterEach(func() {
		provider.AssertOutputsEmpty()
		userStructuredStore.AssertOutputsEmpty()
		blobClient.AssertOutputsEmpty()
		authClient.AssertOutputsEmpty()
	})

	Context("New", func() {
		It("returns an error when the provider is missing", func() {
			client, err := userServiceClient.New(nil)
			errorsTest.ExpectEqual(err, errors.New("provider is missing"))
			Expect(client).To(BeNil())
		})

		It("returns successfully", func() {
			Expect(userServiceClient.New(provider)).ToNot(BeNil())
		})
	})

	Context("with new client", func() {
		var client *userServiceClient.Client
		var logger *logTest.Logger
		var ctx context.Context

		BeforeEach(func() {
			var err error
			client, err = userServiceClient.New(provider)
			Expect(err).ToNot(HaveOccurred())
			Expect(client).ToNot(BeNil())
			logger = logTest.NewLogger()
			ctx = log.NewContextWithLogger(context.Background(), logger)
		})

		Context("with id", func() {
			var id string

			BeforeEach(func() {
				id = userTest.RandomID()
			})

			Context("Get", func() {
				AfterEach(func() {
					Expect(authClient.EnsureAuthorizedUserInputs).To(Equal([]authTest.EnsureAuthorizedUserInput{{TargetUserID: id, AuthorizedPermission: permission.Owner}}))
				})

				It("returns an error when the user client ensure authorized service returns an error", func() {
					responseErr := errorsTest.RandomError()
					authClient.EnsureAuthorizedUserOutputs = []authTest.EnsureAuthorizedUserOutput{{AuthorizedUserID: "", Error: responseErr}}
					result, err := client.Get(ctx, id)
					errorsTest.ExpectEqual(err, responseErr)
					Expect(result).To(BeNil())
				})

				When("the user client ensure authorized service returns successfully", func() {
					BeforeEach(func() {
						authClient.EnsureAuthorizedUserOutputs = []authTest.EnsureAuthorizedUserOutput{{AuthorizedUserID: id, Error: nil}}
					})

					AfterEach(func() {
						Expect(userStructuredSession.GetInputs).To(Equal([]userStoreStructuredTest.GetInput{{ID: id, Condition: nil}}))
					})

					It("returns an error when the user structured session get returns an error", func() {
						responseErr := errorsTest.RandomError()
						userStructuredSession.GetOutputs = []userStoreStructuredTest.GetOutput{{User: nil, Error: responseErr}}
						result, err := client.Get(ctx, id)
						errorsTest.ExpectEqual(err, responseErr)
						Expect(result).To(BeNil())
					})

					It("returns successfully when the user structured session get returns nil", func() {
						userStructuredSession.GetOutputs = []userStoreStructuredTest.GetOutput{{User: nil, Error: nil}}
						deleted, err := client.Get(ctx, id)
						Expect(err).ToNot(HaveOccurred())
						Expect(deleted).To(BeNil())
					})

					When("the user structure session get returns an user", func() {
						var responseResult *user.User

						BeforeEach(func() {
							responseResult = userTest.RandomUser()
							userStructuredSession.GetOutputs = []userStoreStructuredTest.GetOutput{{User: responseResult, Error: nil}}
						})

						It("returns successfully", func() {
							result, err := client.Get(ctx, id)
							Expect(err).ToNot(HaveOccurred())
							Expect(result).To(Equal(responseResult))
						})
					})
				})
			})

			Context("Delete", func() {
				var requiresPassword bool
				var deleet *user.Delete
				var condition *request.Condition

				BeforeEach(func() {
					requiresPassword = true
					deleet = userTest.RandomDelete()
					condition = requestTest.RandomCondition()
				})

				authorizedAssertions := func() {
					AfterEach(func() {
						Expect(userStructuredSession.GetInputs).To(Equal([]userStoreStructuredTest.GetInput{{ID: id, Condition: condition}}))
					})

					It("returns an error when the user structured session get returns an error", func() {
						responseErr := errorsTest.RandomError()
						userStructuredSession.GetOutputs = []userStoreStructuredTest.GetOutput{{User: nil, Error: responseErr}}
						deleted, err := client.Delete(ctx, id, deleet, condition)
						errorsTest.ExpectEqual(err, responseErr)
						Expect(deleted).To(BeFalse())
					})

					It("returns successfully when the user structured session get returns nil", func() {
						userStructuredSession.GetOutputs = []userStoreStructuredTest.GetOutput{{User: nil, Error: nil}}
						deleted, err := client.Delete(ctx, id, deleet, condition)
						Expect(err).ToNot(HaveOccurred())
						Expect(deleted).To(BeFalse())
					})

					When("the user structure session get returns an user", func() {
						var passwordHash string
						var responseResult *user.User

						BeforeEach(func() {
							passwordHash = test.RandomString()
							responseResult = userTest.RandomUser()
							responseResult.UserID = pointer.FromString(id)
							responseResult.PasswordHash = pointer.FromString(passwordHash)
							userStructuredSession.GetOutputs = []userStoreStructuredTest.GetOutput{{User: responseResult, Error: nil}}
						})

						It("returns an error if the user has the clinic role", func() {
							responseResult.Roles = pointer.FromStringArray([]string{user.RoleClinic})
							deleted, err := client.Delete(ctx, id, deleet, condition)
							errorsTest.ExpectEqual(err, request.ErrorUnauthorized())
							Expect(deleted).To(BeFalse())
						})

						It("returns an error when the delete is not specified", func() {
							if requiresPassword {
								deleet = nil
							} else {
								passwordHasher.HashPasswordOutputs = []string{test.RandomString()}
							}
							deleted, err := client.Delete(ctx, id, deleet, condition)
							errorsTest.ExpectEqual(err, request.ErrorUnauthorized())
							Expect(deleted).To(BeFalse())
						})

						It("returns an error when the delete password is not specified", func() {
							if requiresPassword {
								deleet.Password = nil
							} else {
								passwordHasher.HashPasswordOutputs = []string{test.RandomString()}
							}
							deleted, err := client.Delete(ctx, id, deleet, condition)
							errorsTest.ExpectEqual(err, request.ErrorUnauthorized())
							Expect(deleted).To(BeFalse())
						})

						It("returns an error if the password is missing", func() {
							responseResult.PasswordHash = nil
							deleted, err := client.Delete(ctx, id, deleet, condition)
							errorsTest.ExpectEqual(err, request.ErrorUnauthorized())
							Expect(deleted).To(BeFalse())
						})

						When("the passwords are compared", func() {
							AfterEach(func() {
								Expect(passwordHasher.HashPasswordInputs).To(Equal([]userServiceClientTest.HashPasswordInput{{UserID: id, Password: *deleet.Password}}))
							})

							It("returns an error if the password does not match", func() {
								passwordHasher.HashPasswordOutputs = []string{test.RandomString()}
								deleted, err := client.Delete(ctx, id, deleet, condition)
								errorsTest.ExpectEqual(err, request.ErrorUnauthorized())
								Expect(deleted).To(BeFalse())
							})

							When("the password matches", func() {
								BeforeEach(func() {
									passwordHasher.HashPasswordOutputs = []string{passwordHash}
								})

								AfterEach(func() {
									Expect(userStructuredSession.DeleteInputs).To(Equal([]userStoreStructuredTest.DeleteInput{{ID: id, Condition: condition}}))
								})

								It("returns an error when the user structured session delete returns an error", func() {
									responseErr := errorsTest.RandomError()
									userStructuredSession.DeleteOutputs = []userStoreStructuredTest.DeleteOutput{{Deleted: false, Error: responseErr}}
									deleted, err := client.Delete(ctx, id, deleet, condition)
									errorsTest.ExpectEqual(err, responseErr)
									Expect(deleted).To(BeFalse())
								})

								It("returns successfully when the user structured session delete returns false", func() {
									userStructuredSession.DeleteOutputs = []userStoreStructuredTest.DeleteOutput{{Deleted: false, Error: nil}}
									deleted, err := client.Delete(ctx, id, deleet, condition)
									Expect(err).ToNot(HaveOccurred())
									Expect(deleted).To(BeFalse())
								})

								When("the user structured session delete returns successfully", func() {
									BeforeEach(func() {
										userStructuredSession.DeleteOutputs = []userStoreStructuredTest.DeleteOutput{{Deleted: true, Error: nil}}
									})

									// TODO: Complete tests here after each client and store have appropriate mocks
								})
							})
						})
					})
				}

				When("the request is authorized as a service", func() {
					BeforeEach(func() {
						requiresPassword = false
						authClient.EnsureAuthorizedServiceOutputs = []error{nil}
					})

					AfterEach(func() {
						Expect(authClient.EnsureAuthorizedServiceInvocations).To(Equal(1))
					})

					authorizedAssertions()
				})

				When("the request is authorized as an owner", func() {
					BeforeEach(func() {
						authClient.EnsureAuthorizedServiceOutputs = []error{errorsTest.RandomError()}
						authClient.EnsureAuthorizedUserOutputs = []authTest.EnsureAuthorizedUserOutput{{AuthorizedUserID: id, Error: nil}}
					})

					AfterEach(func() {
						Expect(authClient.EnsureAuthorizedUserInputs).To(Equal([]authTest.EnsureAuthorizedUserInput{{TargetUserID: id, AuthorizedPermission: permission.Owner}}))
						Expect(authClient.EnsureAuthorizedServiceInvocations).To(Equal(1))
					})

					authorizedAssertions()
				})

				When("the request is authorized as a custodian", func() {
					BeforeEach(func() {
						requiresPassword = false
						authClient.EnsureAuthorizedServiceOutputs = []error{errorsTest.RandomError()}
						authClient.EnsureAuthorizedUserOutputs = []authTest.EnsureAuthorizedUserOutput{
							{AuthorizedUserID: "", Error: errorsTest.RandomError()},
							{AuthorizedUserID: userTest.RandomID(), Error: nil},
						}
					})

					AfterEach(func() {
						Expect(authClient.EnsureAuthorizedUserInputs).To(Equal([]authTest.EnsureAuthorizedUserInput{
							{TargetUserID: id, AuthorizedPermission: permission.Owner},
							{TargetUserID: id, AuthorizedPermission: permission.Custodian},
						}))
						Expect(authClient.EnsureAuthorizedServiceInvocations).To(Equal(1))
					})

					authorizedAssertions()
				})

				When("the request is not authorized", func() {
					var responseErr error

					BeforeEach(func() {
						responseErr = errorsTest.RandomError()
						authClient.EnsureAuthorizedServiceOutputs = []error{errorsTest.RandomError()}
						authClient.EnsureAuthorizedUserOutputs = []authTest.EnsureAuthorizedUserOutput{
							{AuthorizedUserID: "", Error: errorsTest.RandomError()},
							{AuthorizedUserID: "", Error: responseErr},
						}
					})

					AfterEach(func() {
						Expect(authClient.EnsureAuthorizedUserInputs).To(Equal([]authTest.EnsureAuthorizedUserInput{
							{TargetUserID: id, AuthorizedPermission: permission.Owner},
							{TargetUserID: id, AuthorizedPermission: permission.Custodian},
						}))
						Expect(authClient.EnsureAuthorizedServiceInvocations).To(Equal(1))
					})

					It("returns an error", func() {
						deleted, err := client.Delete(ctx, id, deleet, condition)
						errorsTest.ExpectEqual(err, responseErr)
						Expect(deleted).To(BeFalse())
					})
				})
			})
		})
	})
})
