package client_test

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/auth"
	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/blob"
	blobServiceClient "github.com/tidepool-org/platform/blob/service/client"
	blobServiceClientTest "github.com/tidepool-org/platform/blob/service/client/test"
	blobStoreStructured "github.com/tidepool-org/platform/blob/store/structured"
	blobStoreStructuredTest "github.com/tidepool-org/platform/blob/store/structured/test"
	blobStoreUnstructured "github.com/tidepool-org/platform/blob/store/unstructured"
	blobStoreUnstructuredTest "github.com/tidepool-org/platform/blob/store/unstructured/test"
	blobTest "github.com/tidepool-org/platform/blob/test"
	"github.com/tidepool-org/platform/crypto"
	cryptoTest "github.com/tidepool-org/platform/crypto/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/page"
	pageTest "github.com/tidepool-org/platform/page/test"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	requestTest "github.com/tidepool-org/platform/request/test"
	storeUnstructured "github.com/tidepool-org/platform/store/unstructured"
	"github.com/tidepool-org/platform/test"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("Client", func() {
	var authClient *authTest.Client
	var blobStructuredStore *blobStoreStructuredTest.Store
	var blobStructuredRepository *blobStoreStructuredTest.BlobRepository
	var blobUnstructuredStore *blobStoreUnstructuredTest.Store
	var provider *blobServiceClientTest.Provider

	BeforeEach(func() {
		authClient = authTest.NewClient()
		blobStructuredStore = blobStoreStructuredTest.NewStore()
		blobStructuredRepository = blobStoreStructuredTest.NewBlobRepository()
		blobStructuredStore.NewRepositoryOutput = func(s blobStoreStructured.BlobRepository) *blobStoreStructured.BlobRepository { return &s }(blobStructuredRepository)
		blobUnstructuredStore = blobStoreUnstructuredTest.NewStore()
		provider = blobServiceClientTest.NewProvider()
		provider.AuthClientOutput = func(u auth.Client) *auth.Client { return &u }(authClient)
		provider.BlobStructuredStoreOutput = func(s blobStoreStructured.Store) *blobStoreStructured.Store { return &s }(blobStructuredStore)
		provider.BlobUnstructuredStoreOutput = func(s blobStoreUnstructured.Store) *blobStoreUnstructured.Store { return &s }(blobUnstructuredStore)
	})

	AfterEach(func() {
		provider.AssertOutputsEmpty()
		blobUnstructuredStore.AssertOutputsEmpty()
		blobStructuredStore.AssertOutputsEmpty()
		authClient.AssertOutputsEmpty()
	})

	Context("New", func() {
		It("returns an error when the provider is missing", func() {
			client, err := blobServiceClient.New(nil)
			errorsTest.ExpectEqual(err, errors.New("provider is missing"))
			Expect(client).To(BeNil())
		})

		It("returns successfully", func() {
			Expect(blobServiceClient.New(provider)).ToNot(BeNil())
		})
	})

	Context("with new client", func() {
		var client *blobServiceClient.Client
		var logger *logTest.Logger
		var ctx context.Context

		BeforeEach(func() {
			var err error
			client, err = blobServiceClient.New(provider)
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
				var filter *blob.Filter
				var pagination *page.Pagination

				BeforeEach(func() {
					filter = blobTest.RandomFilter()
					pagination = pageTest.RandomPagination()
				})

				AfterEach(func() {
					Expect(authClient.EnsureAuthorizedServiceInvocations).To(Equal(1))
				})

				It("return an error when the user client ensure authorized service returns an error", func() {
					responseErr := errorsTest.RandomError()
					authClient.EnsureAuthorizedServiceOutputs = []error{responseErr}
					result, err := client.List(ctx, userID, filter, pagination)
					errorsTest.ExpectEqual(err, responseErr)
					Expect(result).To(BeNil())
				})

				When("the user client ensure authorized service returns successfully", func() {
					BeforeEach(func() {
						authClient.EnsureAuthorizedServiceOutputs = []error{nil}
					})

					AfterEach(func() {
						Expect(blobStructuredRepository.ListInputs).To(Equal([]blobStoreStructuredTest.ListInput{{UserID: userID, Filter: filter, Pagination: pagination}}))
					})

					It("returns an error when the blob repository list returns an error", func() {
						responseErr := errorsTest.RandomError()
						blobStructuredRepository.ListOutputs = []blobStoreStructuredTest.ListOutput{{BlobArray: nil, Error: responseErr}}
						result, err := client.List(ctx, userID, filter, pagination)
						errorsTest.ExpectEqual(err, responseErr)
						Expect(result).To(BeNil())
					})

					It("returns successfully when the blob repository list returns successfully", func() {
						responseResult := blobTest.RandomBlobArray(1, 3)
						blobStructuredRepository.ListOutputs = []blobStoreStructuredTest.ListOutput{{BlobArray: responseResult, Error: nil}}
						result, err := client.List(ctx, userID, filter, pagination)
						Expect(err).ToNot(HaveOccurred())
						Expect(result).To(Equal(responseResult))
					})
				})
			})

			Context("Create", func() {
				var content *blob.Content

				BeforeEach(func() {
					content = blobTest.RandomContent()
				})

				AfterEach(func() {
					Expect(authClient.EnsureAuthorizedUserInputs).To(Equal([]authTest.EnsureAuthorizedUserInput{{TargetUserID: userID, AuthorizedPermission: permission.Write}}))
				})

				It("returns an error when the user client ensure authorized user returns an error", func() {
					responseErr := errorsTest.RandomError()
					authClient.EnsureAuthorizedUserOutputs = []authTest.EnsureAuthorizedUserOutput{{AuthorizedUserID: "", Error: responseErr}}
					result, err := client.Create(ctx, userID, content)
					errorsTest.ExpectEqual(err, responseErr)
					Expect(result).To(BeNil())
				})

				When("the user client ensure authorized user returns successfully", func() {
					BeforeEach(func() {
						authClient.EnsureAuthorizedUserOutputs = []authTest.EnsureAuthorizedUserOutput{{AuthorizedUserID: userTest.RandomID(), Error: nil}}
					})

					It("returns an error when content is missing", func() {
						content = nil
						result, err := client.Create(ctx, userID, content)
						errorsTest.ExpectEqual(err, errors.New("content is missing"))
						Expect(result).To(BeNil())
					})

					It("returns an error when content is invalid", func() {
						content.Body = nil
						result, err := client.Create(ctx, userID, content)
						errorsTest.ExpectEqual(err, errors.New("content is invalid"))
						Expect(result).To(BeNil())
					})

					When("the blob is created", func() {
						AfterEach(func() {
							structuredCreate := blobStoreStructured.NewCreate()
							structuredCreate.MediaType = content.MediaType
							Expect(blobStructuredRepository.CreateInputs).To(Equal([]blobStoreStructuredTest.CreateInput{{UserID: userID, Create: structuredCreate}}))
						})

						It("returns an error when the blob repository create returns an error", func() {
							responseErr := errorsTest.RandomError()
							blobStructuredRepository.CreateOutputs = []blobStoreStructuredTest.CreateOutput{{Blob: nil, Error: responseErr}}
							result, err := client.Create(ctx, userID, content)
							errorsTest.ExpectEqual(err, responseErr)
							Expect(result).To(BeNil())
						})

						When("the blob repository create returns successfully", func() {
							var createBlob *blob.Blob

							BeforeEach(func() {
								createBlob = blobTest.RandomBlob()
								createBlob.UserID = pointer.FromString(userID)
								createBlob.DigestMD5 = nil
								createBlob.MediaType = content.MediaType
								createBlob.Size = nil
								createBlob.Status = pointer.FromString(blob.StatusCreated)
								createBlob.ModifiedTime = nil
								blobStructuredRepository.CreateOutputs = []blobStoreStructuredTest.CreateOutput{{Blob: createBlob, Error: nil}}
							})

							AfterEach(func() {
								Expect(blobUnstructuredStore.PutInputs).To(HaveLen(1))
								Expect(blobUnstructuredStore.PutInputs[0].UserID).To(Equal(userID))
								Expect(blobUnstructuredStore.PutInputs[0].ID).To(Equal(*createBlob.ID))
								Expect(blobUnstructuredStore.PutInputs[0].Reader).ToNot(BeNil())
							})

							It("returns an error when the blob unstructured store put returns an error", func() {
								responseErr := errorsTest.RandomError()
								blobUnstructuredStore.PutOutputs = []error{responseErr}
								blobStructuredRepository.DestroyOutputs = []blobStoreStructuredTest.DestroyOutput{{Destroyed: true, Error: nil}}
								result, err := client.Create(ctx, userID, content)
								errorsTest.ExpectEqual(err, responseErr)
								Expect(result).To(BeNil())
							})

							It("returns an error when the blob unstructured store put returns an error and logs an error when the blob repository destroy returns error", func() {
								responseErr := errorsTest.RandomError()
								blobUnstructuredStore.PutOutputs = []error{responseErr}
								blobStructuredRepository.DestroyOutputs = []blobStoreStructuredTest.DestroyOutput{{Destroyed: true, Error: responseErr}}
								result, err := client.Create(ctx, userID, content)
								errorsTest.ExpectEqual(err, responseErr)
								Expect(result).To(BeNil())
								logger.AssertError("Unable to destroy blob after failure to put blob content", log.Fields{"userId": userID, "id": *createBlob.ID, "error": errors.NewSerializable(responseErr)})
							})

							When("the blob unstructured store put returns successfully", func() {
								var size int64

								BeforeEach(func() {
									blobUnstructuredStore.PutStub = func(ctx context.Context, userID string, id string, reader io.Reader, options *storeUnstructured.Options) error {
										size, _ = io.Copy(ioutil.Discard, reader)
										return nil
									}
								})

								When("the size exceeds maximum", func() {
									BeforeEach(func() {
										body := make([]byte, 104857601)
										content.Body = ioutil.NopCloser(bytes.NewReader(body))
										content.DigestMD5 = pointer.FromString(crypto.Base64EncodedMD5Hash(body))
									})

									AfterEach(func() {
										Expect(blobUnstructuredStore.DeleteInputs).To(Equal([]blobStoreUnstructuredTest.DeleteInput{{UserID: userID, ID: *createBlob.ID}}))
										Expect(blobStructuredRepository.DestroyInputs).To(Equal([]blobStoreStructuredTest.DestroyInput{{ID: *createBlob.ID}}))
									})

									It("returns an error", func() {
										blobUnstructuredStore.DeleteOutputs = []blobStoreUnstructuredTest.DeleteOutput{{Deleted: true, Error: nil}}
										blobStructuredRepository.DestroyOutputs = []blobStoreStructuredTest.DestroyOutput{{Destroyed: true, Error: nil}}
										result, err := client.Create(ctx, userID, content)
										errorsTest.ExpectEqual(err, request.ErrorResourceTooLarge())
										Expect(result).To(BeNil())
									})

									It("returns an error and logs an error when the unstructured store returns an error", func() {
										responseErr := errorsTest.RandomError()
										blobUnstructuredStore.DeleteOutputs = []blobStoreUnstructuredTest.DeleteOutput{{Deleted: false, Error: responseErr}}
										blobStructuredRepository.DestroyOutputs = []blobStoreStructuredTest.DestroyOutput{{Destroyed: true, Error: nil}}
										result, err := client.Create(ctx, userID, content)
										errorsTest.ExpectEqual(err, request.ErrorResourceTooLarge())
										Expect(result).To(BeNil())
										logger.AssertError("Unable to delete blob content exceeding maximum size", log.Fields{"userId": userID, "id": *createBlob.ID, "error": errors.NewSerializable(responseErr)})
									})

									It("returns an error and logs an error when the structured store returns an error", func() {
										responseErr := errorsTest.RandomError()
										blobUnstructuredStore.DeleteOutputs = []blobStoreUnstructuredTest.DeleteOutput{{Deleted: true, Error: nil}}
										blobStructuredRepository.DestroyOutputs = []blobStoreStructuredTest.DestroyOutput{{Destroyed: false, Error: responseErr}}
										result, err := client.Create(ctx, userID, content)
										errorsTest.ExpectEqual(err, request.ErrorResourceTooLarge())
										Expect(result).To(BeNil())
										logger.AssertError("Unable to destroy blob exceeding maximum size", log.Fields{"userId": userID, "id": *createBlob.ID, "error": errors.NewSerializable(responseErr)})
									})

									It("returns an error and logs an error when both the unstructured and structured store returns an error", func() {
										responseErr := errorsTest.RandomError()
										blobUnstructuredStore.DeleteOutputs = []blobStoreUnstructuredTest.DeleteOutput{{Deleted: false, Error: responseErr}}
										blobStructuredRepository.DestroyOutputs = []blobStoreStructuredTest.DestroyOutput{{Destroyed: false, Error: responseErr}}
										result, err := client.Create(ctx, userID, content)
										errorsTest.ExpectEqual(err, request.ErrorResourceTooLarge())
										Expect(result).To(BeNil())
										logger.AssertError("Unable to delete blob content exceeding maximum size", log.Fields{"userId": userID, "id": *createBlob.ID, "error": errors.NewSerializable(responseErr)})
										logger.AssertError("Unable to destroy blob exceeding maximum size", log.Fields{"userId": userID, "id": *createBlob.ID, "error": errors.NewSerializable(responseErr)})
									})
								})

								When("the digest does not match", func() {
									var digestMD5 string

									BeforeEach(func() {
										digestMD5 = *content.DigestMD5
										content.DigestMD5 = pointer.FromString(cryptoTest.RandomBase64EncodedMD5Hash())
									})

									AfterEach(func() {
										Expect(blobUnstructuredStore.DeleteInputs).To(Equal([]blobStoreUnstructuredTest.DeleteInput{{UserID: userID, ID: *createBlob.ID}}))
										Expect(blobStructuredRepository.DestroyInputs).To(Equal([]blobStoreStructuredTest.DestroyInput{{ID: *createBlob.ID}}))
									})

									It("returns an error", func() {
										blobUnstructuredStore.DeleteOutputs = []blobStoreUnstructuredTest.DeleteOutput{{Deleted: true, Error: nil}}
										blobStructuredRepository.DestroyOutputs = []blobStoreStructuredTest.DestroyOutput{{Destroyed: true, Error: nil}}
										result, err := client.Create(ctx, userID, content)
										errorsTest.ExpectEqual(err, errorsTest.WithPointerSource(request.ErrorDigestsNotEqual(*content.DigestMD5, digestMD5), "/digestMD5"))
										Expect(result).To(BeNil())
									})

									It("returns an error and logs an error when the unstructured store returns an error", func() {
										responseErr := errorsTest.RandomError()
										blobUnstructuredStore.DeleteOutputs = []blobStoreUnstructuredTest.DeleteOutput{{Deleted: false, Error: responseErr}}
										blobStructuredRepository.DestroyOutputs = []blobStoreStructuredTest.DestroyOutput{{Destroyed: true, Error: nil}}
										result, err := client.Create(ctx, userID, content)
										errorsTest.ExpectEqual(err, errorsTest.WithPointerSource(request.ErrorDigestsNotEqual(*content.DigestMD5, digestMD5), "/digestMD5"))
										Expect(result).To(BeNil())
										logger.AssertError("Unable to delete blob content with incorrect MD5 digest", log.Fields{"userId": userID, "id": *createBlob.ID, "error": errors.NewSerializable(responseErr)})
									})

									It("returns an error and logs an error when the structured store returns an error", func() {
										responseErr := errorsTest.RandomError()
										blobUnstructuredStore.DeleteOutputs = []blobStoreUnstructuredTest.DeleteOutput{{Deleted: true, Error: nil}}
										blobStructuredRepository.DestroyOutputs = []blobStoreStructuredTest.DestroyOutput{{Destroyed: false, Error: responseErr}}
										result, err := client.Create(ctx, userID, content)
										errorsTest.ExpectEqual(err, errorsTest.WithPointerSource(request.ErrorDigestsNotEqual(*content.DigestMD5, digestMD5), "/digestMD5"))
										Expect(result).To(BeNil())
										logger.AssertError("Unable to destroy blob with incorrect MD5 digest", log.Fields{"userId": userID, "id": *createBlob.ID, "error": errors.NewSerializable(responseErr)})
									})

									It("returns an error and logs an error when both the unstructured and structured store returns an error", func() {
										responseErr := errorsTest.RandomError()
										blobUnstructuredStore.DeleteOutputs = []blobStoreUnstructuredTest.DeleteOutput{{Deleted: false, Error: responseErr}}
										blobStructuredRepository.DestroyOutputs = []blobStoreStructuredTest.DestroyOutput{{Destroyed: false, Error: responseErr}}
										result, err := client.Create(ctx, userID, content)
										errorsTest.ExpectEqual(err, errorsTest.WithPointerSource(request.ErrorDigestsNotEqual(*content.DigestMD5, digestMD5), "/digestMD5"))
										Expect(result).To(BeNil())
										logger.AssertError("Unable to delete blob content with incorrect MD5 digest", log.Fields{"userId": userID, "id": *createBlob.ID, "error": errors.NewSerializable(responseErr)})
										logger.AssertError("Unable to destroy blob with incorrect MD5 digest", log.Fields{"userId": userID, "id": *createBlob.ID, "error": errors.NewSerializable(responseErr)})
									})
								})

								When("the digest and size are valid", func() {
									AfterEach(func() {
										update := blobStoreStructured.NewUpdate()
										update.DigestMD5 = pointer.CloneString(content.DigestMD5)
										update.Size = pointer.FromInt(int(size))
										update.Status = pointer.FromString(blob.StatusAvailable)
										Expect(blobStructuredRepository.UpdateInputs).To(Equal([]blobStoreStructuredTest.UpdateInput{{ID: *createBlob.ID, Update: update}}))
									})

									It("returns an error when blob repository update returns an error", func() {
										responseErr := errorsTest.RandomError()
										blobStructuredRepository.UpdateOutputs = []blobStoreStructuredTest.UpdateOutput{{Blob: nil, Error: responseErr}}
										result, err := client.Create(ctx, userID, content)
										errorsTest.ExpectEqual(err, responseErr)
										Expect(result).To(BeNil())
									})

									When("the blob repository update returns successfully", func() {
										var updateBlob *blob.Blob

										BeforeEach(func() {
											updateBlob = blobTest.CloneBlob(createBlob)
											updateBlob.DigestMD5 = pointer.CloneString(content.DigestMD5)
											updateBlob.Size = pointer.FromInt(int(size))
											updateBlob.Status = pointer.FromString(blob.StatusAvailable)
											updateBlob.ModifiedTime = pointer.FromTime(test.RandomTimeFromRange(*createBlob.CreatedTime, time.Now()))
											blobStructuredRepository.UpdateOutputs = []blobStoreStructuredTest.UpdateOutput{{Blob: updateBlob, Error: nil}}
										})

										It("returns successfully", func() {
											Expect(client.Create(ctx, userID, content)).To(Equal(updateBlob))
										})

										When("the size is maximum", func() {
											BeforeEach(func() {
												body := make([]byte, 104857600)
												content.Body = ioutil.NopCloser(bytes.NewReader(body))
												content.DigestMD5 = pointer.FromString(crypto.Base64EncodedMD5Hash(body))
											})

											It("returns successfully", func() {
												Expect(client.Create(ctx, userID, content)).To(Equal(updateBlob))
											})
										})
									})
								})
							})
						})
					})
				})
			})

			Context("DeleteAll", func() {
				AfterEach(func() {
					Expect(authClient.EnsureAuthorizedServiceInvocations).To(Equal(1))
				})

				It("returns an error when the user client ensure authorized service returns an error", func() {
					responseErr := errorsTest.RandomError()
					authClient.EnsureAuthorizedServiceOutputs = []error{responseErr}
					errorsTest.ExpectEqual(client.DeleteAll(ctx, userID), responseErr)
				})

				When("the user client ensure authorized service returns successfully", func() {
					BeforeEach(func() {
						authClient.EnsureAuthorizedServiceOutputs = []error{nil}
					})

					AfterEach(func() {
						Expect(blobStructuredRepository.DeleteAllInputs).To(Equal([]string{userID}))
					})

					It("returns an error when the blob repository delete returns an error", func() {
						responseErr := errorsTest.RandomError()
						blobStructuredRepository.DeleteAllOutputs = []blobStoreStructuredTest.DeleteAllOutput{{Deleted: false, Error: responseErr}}
						errorsTest.ExpectEqual(client.DeleteAll(ctx, userID), responseErr)
					})

					It("returns successfully when the blob repository delete returns successfully without deleted", func() {
						blobStructuredRepository.DeleteAllOutputs = []blobStoreStructuredTest.DeleteAllOutput{{Deleted: false, Error: nil}}
						Expect(client.DeleteAll(ctx, userID)).To(Succeed())
					})

					When("the blob repository delete returns successfully with deleted", func() {
						BeforeEach(func() {
							blobStructuredRepository.DeleteAllOutputs = []blobStoreStructuredTest.DeleteAllOutput{{Deleted: true, Error: nil}}
						})

						AfterEach(func() {
							Expect(blobUnstructuredStore.DeleteAllInputs).To(Equal([]string{userID}))
						})

						It("returns an error when the blob unstructured store delete all returns an error", func() {
							responseErr := errorsTest.RandomError()
							blobUnstructuredStore.DeleteAllOutputs = []error{responseErr}
							errorsTest.ExpectEqual(client.DeleteAll(ctx, userID), responseErr)
						})

						When("the blob unstructured store delete all returns successfully", func() {
							BeforeEach(func() {
								blobUnstructuredStore.DeleteAllOutputs = []error{nil}
							})

							AfterEach(func() {
								Expect(blobStructuredRepository.DestroyAllInputs).To(Equal([]string{userID}))
							})

							It("returns an error when the blob repository destroy returns an error", func() {
								responseErr := errorsTest.RandomError()
								blobStructuredRepository.DestroyAllOutputs = []blobStoreStructuredTest.DestroyAllOutput{{Destroyed: false, Error: responseErr}}
								errorsTest.ExpectEqual(client.DeleteAll(ctx, userID), responseErr)
							})

							It("returns successfully when the blob repository destroy returns false", func() {
								blobStructuredRepository.DestroyAllOutputs = []blobStoreStructuredTest.DestroyAllOutput{{Destroyed: false, Error: nil}}
								Expect(client.DeleteAll(ctx, userID)).To(Succeed())
							})

							It("returns successfully when the blob repository destroy returns true", func() {
								blobStructuredRepository.DestroyAllOutputs = []blobStoreStructuredTest.DestroyAllOutput{{Destroyed: true, Error: nil}}
								Expect(client.DeleteAll(ctx, userID)).To(Succeed())
							})
						})
					})
				})
			})
		})

		Context("with id", func() {
			var id string

			BeforeEach(func() {
				id = blobTest.RandomID()
			})

			Context("Get", func() {
				AfterEach(func() {
					Expect(authClient.EnsureAuthorizedServiceInvocations).To(Equal(1))
				})

				It("returns an error when the user client ensure authorized service returns an error", func() {
					responseErr := errorsTest.RandomError()
					authClient.EnsureAuthorizedServiceOutputs = []error{responseErr}
					result, err := client.Get(ctx, id)
					errorsTest.ExpectEqual(err, responseErr)
					Expect(result).To(BeNil())
				})

				When("the user client ensure authorized service returns successfully", func() {
					BeforeEach(func() {
						authClient.EnsureAuthorizedServiceOutputs = []error{nil}
					})

					AfterEach(func() {
						Expect(blobStructuredRepository.GetInputs).To(Equal([]blobStoreStructuredTest.GetInput{{ID: id, Condition: nil}}))
					})

					It("returns an error when the blob repository get returns an error", func() {
						responseErr := errorsTest.RandomError()
						blobStructuredRepository.GetOutputs = []blobStoreStructuredTest.GetOutput{{Blob: nil, Error: responseErr}}
						result, err := client.Get(ctx, id)
						errorsTest.ExpectEqual(err, responseErr)
						Expect(result).To(BeNil())
					})

					It("returns successfully when the blob repository get returns successfully", func() {
						responseResult := blobTest.RandomBlob()
						blobStructuredRepository.GetOutputs = []blobStoreStructuredTest.GetOutput{{Blob: responseResult, Error: nil}}
						result, err := client.Get(ctx, id)
						Expect(err).ToNot(HaveOccurred())
						Expect(result).To(Equal(responseResult))
					})
				})
			})

			Context("GetContent", func() {
				AfterEach(func() {
					Expect(authClient.EnsureAuthorizedServiceInvocations).To(Equal(1))
				})

				It("returns an error when the user client ensure authorized service returns an error", func() {
					responseErr := errorsTest.RandomError()
					authClient.EnsureAuthorizedServiceOutputs = []error{responseErr}
					result, err := client.GetContent(ctx, id)
					errorsTest.ExpectEqual(err, responseErr)
					Expect(result).To(BeNil())
				})

				When("the user client ensure authorized service returns successfully", func() {
					BeforeEach(func() {
						authClient.EnsureAuthorizedServiceOutputs = []error{nil}
					})

					AfterEach(func() {
						Expect(blobStructuredRepository.GetInputs).To(Equal([]blobStoreStructuredTest.GetInput{{ID: id, Condition: nil}}))
					})

					It("returns an error when the blob repository get returns an error", func() {
						responseErr := errorsTest.RandomError()
						blobStructuredRepository.GetOutputs = []blobStoreStructuredTest.GetOutput{{Blob: nil, Error: responseErr}}
						result, err := client.GetContent(ctx, id)
						errorsTest.ExpectEqual(err, responseErr)
						Expect(result).To(BeNil())
					})

					It("returns successfully when the blob repository get returns nil", func() {
						blobStructuredRepository.GetOutputs = []blobStoreStructuredTest.GetOutput{{Blob: nil, Error: nil}}
						result, err := client.GetContent(ctx, id)
						Expect(err).ToNot(HaveOccurred())
						Expect(result).To(BeNil())
					})

					When("the blob structure repository get returns a blob", func() {
						var responseResult *blob.Blob

						BeforeEach(func() {
							responseResult = blobTest.RandomBlob()
							responseResult.ID = pointer.FromString(id)
							blobStructuredRepository.GetOutputs = []blobStoreStructuredTest.GetOutput{{Blob: responseResult, Error: nil}}
						})

						AfterEach(func() {
							Expect(blobUnstructuredStore.GetInputs).To(Equal([]blobStoreUnstructuredTest.GetInput{{UserID: *responseResult.UserID, ID: id}}))
						})

						It("returns an error when the blob unstructured store get returns an error", func() {
							responseErr := errorsTest.RandomError()
							blobUnstructuredStore.GetOutputs = []blobStoreUnstructuredTest.GetOutput{{Reader: nil, Error: responseErr}}
							result, err := client.GetContent(ctx, id)
							errorsTest.ExpectEqual(err, responseErr)
							Expect(result).To(BeNil())
						})

						It("returns successfully when the blob unstructured store get returns successfully", func() {
							body := test.RandomBytes()
							reader := ioutil.NopCloser(bytes.NewReader(body))
							blobUnstructuredStore.GetOutputs = []blobStoreUnstructuredTest.GetOutput{{Reader: reader, Error: nil}}
							result, err := client.GetContent(ctx, id)
							Expect(err).ToNot(HaveOccurred())
							Expect(result).To(Equal(&blob.Content{
								Body:      reader,
								DigestMD5: responseResult.DigestMD5,
								MediaType: responseResult.MediaType,
							}))
						})
					})
				})
			})

			Context("Delete", func() {
				var condition *request.Condition

				BeforeEach(func() {
					condition = requestTest.RandomCondition()
				})

				AfterEach(func() {
					Expect(authClient.EnsureAuthorizedServiceInvocations).To(Equal(1))
				})

				It("returns an error when the user client ensure authorized service returns an error", func() {
					responseErr := errorsTest.RandomError()
					authClient.EnsureAuthorizedServiceOutputs = []error{responseErr}
					deleted, err := client.Delete(ctx, id, condition)
					errorsTest.ExpectEqual(err, responseErr)
					Expect(deleted).To(BeFalse())
				})

				When("the user client ensure authorized service returns successfully", func() {
					BeforeEach(func() {
						authClient.EnsureAuthorizedServiceOutputs = []error{nil}
					})

					AfterEach(func() {
						Expect(blobStructuredRepository.GetInputs).To(Equal([]blobStoreStructuredTest.GetInput{{ID: id, Condition: condition}}))
					})

					It("returns an error when the blob repository get returns an error", func() {
						responseErr := errorsTest.RandomError()
						blobStructuredRepository.GetOutputs = []blobStoreStructuredTest.GetOutput{{Blob: nil, Error: responseErr}}
						deleted, err := client.Delete(ctx, id, condition)
						errorsTest.ExpectEqual(err, responseErr)
						Expect(deleted).To(BeFalse())
					})

					It("returns successfully when the blob repository get returns nil", func() {
						blobStructuredRepository.GetOutputs = []blobStoreStructuredTest.GetOutput{{Blob: nil, Error: nil}}
						deleted, err := client.Delete(ctx, id, condition)
						Expect(err).ToNot(HaveOccurred())
						Expect(deleted).To(BeFalse())
					})

					When("the blob structure repository get returns a blob", func() {
						var responseResult *blob.Blob

						BeforeEach(func() {
							responseResult = blobTest.RandomBlob()
							responseResult.ID = pointer.FromString(id)
							blobStructuredRepository.GetOutputs = []blobStoreStructuredTest.GetOutput{{Blob: responseResult, Error: nil}}
						})

						AfterEach(func() {
							Expect(blobStructuredRepository.DeleteInputs).To(Equal([]blobStoreStructuredTest.DeleteInput{{ID: id, Condition: condition}}))
						})

						It("returns an error when the blob repository delete returns an error", func() {
							responseErr := errorsTest.RandomError()
							blobStructuredRepository.DeleteOutputs = []blobStoreStructuredTest.DeleteOutput{{Deleted: false, Error: responseErr}}
							deleted, err := client.Delete(ctx, id, condition)
							errorsTest.ExpectEqual(err, responseErr)
							Expect(deleted).To(BeFalse())
						})

						It("returns successfully when the blob repository delete returns false", func() {
							blobStructuredRepository.DeleteOutputs = []blobStoreStructuredTest.DeleteOutput{{Deleted: false, Error: nil}}
							deleted, err := client.Delete(ctx, id, condition)
							Expect(err).ToNot(HaveOccurred())
							Expect(deleted).To(BeFalse())
						})

						When("the blob repository delete returns successfully", func() {
							BeforeEach(func() {
								blobStructuredRepository.DeleteOutputs = []blobStoreStructuredTest.DeleteOutput{{Deleted: true, Error: nil}}
							})

							AfterEach(func() {
								Expect(blobUnstructuredStore.DeleteInputs).To(Equal([]blobStoreUnstructuredTest.DeleteInput{{UserID: *responseResult.UserID, ID: id}}))
							})

							It("returns an error when the blob unstructured store delete returns an error", func() {
								responseErr := errorsTest.RandomError()
								blobUnstructuredStore.DeleteOutputs = []blobStoreUnstructuredTest.DeleteOutput{{Deleted: false, Error: responseErr}}
								deleted, err := client.Delete(ctx, id, condition)
								errorsTest.ExpectEqual(err, responseErr)
								Expect(deleted).To(BeFalse())
							})

							When("the blob unstructured store delete returns successfully", func() {
								BeforeEach(func() {
									blobUnstructuredStore.DeleteOutputs = []blobStoreUnstructuredTest.DeleteOutput{{Deleted: true, Error: nil}}
								})

								AfterEach(func() {
									Expect(blobStructuredRepository.DestroyInputs).To(Equal([]blobStoreStructuredTest.DestroyInput{{ID: id}}))
								})

								It("returns an error when the blob repository destroy returns an error", func() {
									responseErr := errorsTest.RandomError()
									blobStructuredRepository.DestroyOutputs = []blobStoreStructuredTest.DestroyOutput{{Destroyed: false, Error: responseErr}}
									deleted, err := client.Delete(ctx, id, condition)
									errorsTest.ExpectEqual(err, responseErr)
									Expect(deleted).To(BeFalse())
								})

								It("returns false when the blob repository destroy returns false", func() {
									blobStructuredRepository.DestroyOutputs = []blobStoreStructuredTest.DestroyOutput{{Destroyed: false, Error: nil}}
									deleted, err := client.Delete(ctx, id, condition)
									Expect(err).ToNot(HaveOccurred())
									Expect(deleted).To(BeFalse())
								})

								It("returns true when the blob repository destroy returns true", func() {
									blobStructuredRepository.DestroyOutputs = []blobStoreStructuredTest.DestroyOutput{{Destroyed: true, Error: nil}}
									deleted, err := client.Delete(ctx, id, condition)
									Expect(err).ToNot(HaveOccurred())
									Expect(deleted).To(BeTrue())
								})
							})
						})
					})
				})
			})
		})
	})
})
