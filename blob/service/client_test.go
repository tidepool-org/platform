package service_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"bytes"
	"context"
	"io"
	"io/ioutil"
	"time"

	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/blob"
	blobService "github.com/tidepool-org/platform/blob/service"
	blobServiceTest "github.com/tidepool-org/platform/blob/service/test"
	blobStoreStructured "github.com/tidepool-org/platform/blob/store/structured"
	blobStoreStructuredTest "github.com/tidepool-org/platform/blob/store/structured/test"
	blobStoreUnstructured "github.com/tidepool-org/platform/blob/store/unstructured"
	blobStoreUnstructuredTest "github.com/tidepool-org/platform/blob/store/unstructured/test"
	blobTest "github.com/tidepool-org/platform/blob/test"
	cryptoTest "github.com/tidepool-org/platform/crypto/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/page"
	pageTest "github.com/tidepool-org/platform/page/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/test"
	"github.com/tidepool-org/platform/user"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("Client", func() {
	var blobStructuredStore *blobStoreStructuredTest.Store
	var blobStructuredSession *blobStoreStructuredTest.Session
	var blobUnstructuredStore *blobStoreUnstructuredTest.Store
	var userClient *userTest.Client
	var clientProvider *blobServiceTest.ClientProvider

	BeforeEach(func() {
		blobStructuredStore = blobStoreStructuredTest.NewStore()
		blobStructuredSession = blobStoreStructuredTest.NewSession()
		blobStructuredSession.CloseOutput = func(err error) *error { return &err }(nil)
		blobStructuredStore.NewSessionOutput = func(s blobStoreStructured.Session) *blobStoreStructured.Session { return &s }(blobStructuredSession)
		blobUnstructuredStore = blobStoreUnstructuredTest.NewStore()
		userClient = userTest.NewClient()
		clientProvider = blobServiceTest.NewClientProvider()
		clientProvider.BlobStructuredStoreOutput = func(s blobStoreStructured.Store) *blobStoreStructured.Store { return &s }(blobStructuredStore)
		clientProvider.BlobUnstructuredStoreOutput = func(s blobStoreUnstructured.Store) *blobStoreUnstructured.Store { return &s }(blobUnstructuredStore)
		clientProvider.UserClientOutput = func(u user.Client) *user.Client { return &u }(userClient)
	})

	AfterEach(func() {
		clientProvider.AssertOutputsEmpty()
		blobUnstructuredStore.AssertOutputsEmpty()
		blobStructuredStore.AssertOutputsEmpty()
	})

	Context("NewClient", func() {
		It("returns an error when the client provider is missing", func() {
			client, err := blobService.NewClient(nil)
			errorsTest.ExpectEqual(err, errors.New("client provider is missing"))
			Expect(client).To(BeNil())
		})

		It("returns successfully", func() {
			Expect(blobService.NewClient(clientProvider)).ToNot(BeNil())
		})
	})

	Context("with new client", func() {
		var client *blobService.Client
		var logger *logTest.Logger
		var details request.Details
		var ctx context.Context

		BeforeEach(func() {
			var err error
			client, err = blobService.NewClient(clientProvider)
			Expect(err).ToNot(HaveOccurred())
			Expect(client).ToNot(BeNil())
			logger = logTest.NewLogger()
			details = request.NewDetails(request.MethodServiceSecret, "", authTest.NewSessionToken())
			ctx = context.Background()
			ctx = log.NewContextWithLogger(ctx, logger)
			ctx = request.NewContextWithDetails(ctx, details)
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
					Expect(userClient.EnsureAuthorizedServiceInputs).To(Equal([]context.Context{ctx}))
				})

				It("return an error when the user client ensure authorized service returns an error", func() {
					responseErr := errorsTest.RandomError()
					userClient.EnsureAuthorizedServiceOutputs = []error{responseErr}
					blbs, err := client.List(ctx, userID, filter, pagination)
					errorsTest.ExpectEqual(err, responseErr)
					Expect(blbs).To(BeNil())
				})

				When("user client ensure authorized service returns successfully", func() {
					BeforeEach(func() {
						userClient.EnsureAuthorizedServiceOutputs = []error{nil}
					})

					AfterEach(func() {
						Expect(blobStructuredSession.ListInputs).To(Equal([]blobStoreStructuredTest.ListInput{{Context: ctx, UserID: userID, Filter: filter, Pagination: pagination}}))
					})

					It("returns an error if the blob structured session list returns an error", func() {
						responseErr := errorsTest.RandomError()
						blobStructuredSession.ListOutputs = []blobStoreStructuredTest.ListOutput{{Blobs: nil, Error: responseErr}}
						blbs, err := client.List(ctx, userID, filter, pagination)
						errorsTest.ExpectEqual(err, responseErr)
						Expect(blbs).To(BeNil())
					})

					It("returns successfully if the blob structured session list returns successfully", func() {
						responseBlobs := blobTest.RandomBlobs(1, 3)
						blobStructuredSession.ListOutputs = []blobStoreStructuredTest.ListOutput{{Blobs: responseBlobs, Error: nil}}
						blbs, err := client.List(ctx, userID, filter, pagination)
						Expect(err).ToNot(HaveOccurred())
						Expect(blbs).To(Equal(responseBlobs))
					})
				})
			})

			Context("Create", func() {
				var create *blob.Create

				BeforeEach(func() {
					create = blobTest.RandomCreate()
				})

				AfterEach(func() {
					Expect(userClient.EnsureAuthorizedUserInputs).To(Equal([]userTest.EnsureAuthorizedUserInput{{Context: ctx, TargetUserID: userID, Permission: user.UploadPermission}}))
				})

				It("returns an error if the user client ensure authorized user returns an error", func() {
					responseErr := errorsTest.RandomError()
					userClient.EnsureAuthorizedUserOutputs = []userTest.EnsureAuthorizedUserOutput{{AuthorizedUserID: "", Error: responseErr}}
					blb, err := client.Create(ctx, userID, create)
					errorsTest.ExpectEqual(err, responseErr)
					Expect(blb).To(BeNil())
				})

				When("user client ensure authorized user returns successfully", func() {
					BeforeEach(func() {
						userClient.EnsureAuthorizedUserOutputs = []userTest.EnsureAuthorizedUserOutput{{AuthorizedUserID: userTest.RandomID(), Error: nil}}
					})

					It("returns an error if create is missing", func() {
						create = nil
						blb, err := client.Create(ctx, userID, create)
						errorsTest.ExpectEqual(err, errors.New("create is missing"))
						Expect(blb).To(BeNil())
					})

					It("returns an error if create is invalid", func() {
						create.Body = nil
						blb, err := client.Create(ctx, userID, create)
						errorsTest.ExpectEqual(err, errors.New("create is invalid"))
						Expect(blb).To(BeNil())
					})

					When("the blob is created", func() {
						AfterEach(func() {
							structuredCreate := blobStoreStructured.NewCreate()
							structuredCreate.MediaType = create.MediaType
							Expect(blobStructuredSession.CreateInputs).To(Equal([]blobStoreStructuredTest.CreateInput{{Context: ctx, UserID: userID, Create: structuredCreate}}))
						})

						It("returns an error if the blob structured session create returns an error", func() {
							responseErr := errorsTest.RandomError()
							blobStructuredSession.CreateOutputs = []blobStoreStructuredTest.CreateOutput{{Blob: nil, Error: responseErr}}
							blb, err := client.Create(ctx, userID, create)
							errorsTest.ExpectEqual(err, responseErr)
							Expect(blb).To(BeNil())
						})

						When("the blob structured session create returns successfully", func() {
							var createBlob *blob.Blob

							BeforeEach(func() {
								createBlob = blobTest.RandomBlob()
								createBlob.UserID = pointer.FromString(userID)
								createBlob.DigestMD5 = nil
								createBlob.MediaType = create.MediaType
								createBlob.Size = nil
								createBlob.Status = pointer.FromString(blob.StatusCreated)
								createBlob.ModifiedTime = nil
								blobStructuredSession.CreateOutputs = []blobStoreStructuredTest.CreateOutput{{Blob: createBlob, Error: nil}}
							})

							AfterEach(func() {
								Expect(blobUnstructuredStore.PutInputs).To(HaveLen(1))
								Expect(blobUnstructuredStore.PutInputs[0].Context).To(Equal(ctx))
								Expect(blobUnstructuredStore.PutInputs[0].UserID).To(Equal(userID))
								Expect(blobUnstructuredStore.PutInputs[0].ID).To(Equal(*createBlob.ID))
								Expect(blobUnstructuredStore.PutInputs[0].Reader).ToNot(BeNil())
							})

							It("returns an error if the blob unstructured store put returns an error", func() {
								responseErr := errorsTest.RandomError()
								blobUnstructuredStore.PutOutputs = []error{responseErr}
								blobStructuredSession.DeleteOutputs = []blobStoreStructuredTest.DeleteOutput{{Deleted: true, Error: nil}}
								blb, err := client.Create(ctx, userID, create)
								errorsTest.ExpectEqual(err, responseErr)
								Expect(blb).To(BeNil())
							})

							It("returns an error if the blob unstructured store put returns an error and logs an error if the blob structured session delete returns error", func() {
								responseErr := errorsTest.RandomError()
								blobUnstructuredStore.PutOutputs = []error{responseErr}
								blobStructuredSession.DeleteOutputs = []blobStoreStructuredTest.DeleteOutput{{Deleted: true, Error: responseErr}}
								blb, err := client.Create(ctx, userID, create)
								errorsTest.ExpectEqual(err, responseErr)
								Expect(blb).To(BeNil())
								logger.AssertError("Unable to delete blob after failure to put blob content", log.Fields{"userId": userID, "id": *createBlob.ID, "error": errors.NewSerializable(responseErr)})
							})

							When("the blob unstructured store put returns successfully", func() {
								var size int64

								BeforeEach(func() {
									blobUnstructuredStore.PutStub = func(ctx context.Context, userID string, id string, reader io.Reader) error {
										size, _ = io.Copy(ioutil.Discard, reader)
										return nil
									}
								})

								When("the digest does not match", func() {
									var digestMD5 string

									BeforeEach(func() {
										digestMD5 = *create.DigestMD5
										create.DigestMD5 = pointer.FromString(cryptoTest.RandomBase64EncodedMD5Hash())
									})

									AfterEach(func() {
										Expect(blobUnstructuredStore.DeleteInputs).To(Equal([]blobStoreUnstructuredTest.DeleteInput{{Context: ctx, UserID: userID, ID: *createBlob.ID}}))
										Expect(blobStructuredSession.DeleteInputs).To(Equal([]blobStoreStructuredTest.DeleteInput{{Context: ctx, ID: *createBlob.ID}}))
									})

									It("returns an error", func() {
										blobUnstructuredStore.DeleteOutputs = []blobStoreUnstructuredTest.DeleteOutput{{Deleted: true, Error: nil}}
										blobStructuredSession.DeleteOutputs = []blobStoreStructuredTest.DeleteOutput{{Deleted: true, Error: nil}}
										blb, err := client.Create(ctx, userID, create)
										errorsTest.ExpectEqual(err, errorsTest.WithPointerSource(blob.ErrorDigestsNotEqual(*create.DigestMD5, digestMD5), "/digestMD5"))
										Expect(blb).To(BeNil())
									})

									It("returns an error and logs an error if the unstructured store returns an error", func() {
										responseErr := errorsTest.RandomError()
										blobUnstructuredStore.DeleteOutputs = []blobStoreUnstructuredTest.DeleteOutput{{Deleted: false, Error: responseErr}}
										blobStructuredSession.DeleteOutputs = []blobStoreStructuredTest.DeleteOutput{{Deleted: true, Error: nil}}
										blb, err := client.Create(ctx, userID, create)
										errorsTest.ExpectEqual(err, errorsTest.WithPointerSource(blob.ErrorDigestsNotEqual(*create.DigestMD5, digestMD5), "/digestMD5"))
										Expect(blb).To(BeNil())
										logger.AssertError("Unable to delete blob content with incorrect MD5 digest", log.Fields{"userId": userID, "id": *createBlob.ID, "error": errors.NewSerializable(responseErr)})
									})

									It("returns an error and logs an error if the structured store returns an error", func() {
										responseErr := errorsTest.RandomError()
										blobUnstructuredStore.DeleteOutputs = []blobStoreUnstructuredTest.DeleteOutput{{Deleted: true, Error: nil}}
										blobStructuredSession.DeleteOutputs = []blobStoreStructuredTest.DeleteOutput{{Deleted: false, Error: responseErr}}
										blb, err := client.Create(ctx, userID, create)
										errorsTest.ExpectEqual(err, errorsTest.WithPointerSource(blob.ErrorDigestsNotEqual(*create.DigestMD5, digestMD5), "/digestMD5"))
										Expect(blb).To(BeNil())
										logger.AssertError("Unable to delete blob with incorrect MD5 digest", log.Fields{"userId": userID, "id": *createBlob.ID, "error": errors.NewSerializable(responseErr)})
									})

									It("returns an error and logs an error if both the unstructured and structured store returns an error", func() {
										responseErr := errorsTest.RandomError()
										blobUnstructuredStore.DeleteOutputs = []blobStoreUnstructuredTest.DeleteOutput{{Deleted: false, Error: responseErr}}
										blobStructuredSession.DeleteOutputs = []blobStoreStructuredTest.DeleteOutput{{Deleted: false, Error: responseErr}}
										blb, err := client.Create(ctx, userID, create)
										errorsTest.ExpectEqual(err, errorsTest.WithPointerSource(blob.ErrorDigestsNotEqual(*create.DigestMD5, digestMD5), "/digestMD5"))
										Expect(blb).To(BeNil())
										logger.AssertError("Unable to delete blob content with incorrect MD5 digest", log.Fields{"userId": userID, "id": *createBlob.ID, "error": errors.NewSerializable(responseErr)})
										logger.AssertError("Unable to delete blob with incorrect MD5 digest", log.Fields{"userId": userID, "id": *createBlob.ID, "error": errors.NewSerializable(responseErr)})
									})
								})

								When("the digest matches", func() {
									AfterEach(func() {
										update := blobStoreStructured.NewUpdate()
										update.DigestMD5 = pointer.CloneString(create.DigestMD5)
										update.Size = pointer.FromInt(int(size))
										update.Status = pointer.FromString(blob.StatusAvailable)
										Expect(blobStructuredSession.UpdateInputs).To(Equal([]blobStoreStructuredTest.UpdateInput{{Context: ctx, ID: *createBlob.ID, Update: update}}))
									})

									It("returns an error if blob structured session update returns an error", func() {
										responseErr := errorsTest.RandomError()
										blobStructuredSession.UpdateOutputs = []blobStoreStructuredTest.UpdateOutput{{Blob: nil, Error: responseErr}}
										blb, err := client.Create(ctx, userID, create)
										errorsTest.ExpectEqual(err, responseErr)
										Expect(blb).To(BeNil())
									})

									When("the blob structured session update returns successfully", func() {
										var updateBlob *blob.Blob

										BeforeEach(func() {
											updateBlob = blobTest.CloneBlob(createBlob)
											updateBlob.DigestMD5 = pointer.CloneString(create.DigestMD5)
											updateBlob.Size = pointer.FromInt(int(size))
											updateBlob.Status = pointer.FromString(blob.StatusAvailable)
											updateBlob.ModifiedTime = pointer.FromTime(test.RandomTimeFromRange(*createBlob.CreatedTime, time.Now()).Truncate(time.Second))
											blobStructuredSession.UpdateOutputs = []blobStoreStructuredTest.UpdateOutput{{Blob: updateBlob, Error: nil}}
										})

										It("returns successfully", func() {
											Expect(client.Create(ctx, userID, create)).To(Equal(updateBlob))
										})
									})
								})
							})
						})
					})
				})
			})
		})

		Context("with id", func() {
			var id string

			BeforeEach(func() {
				id = blob.NewID()
			})

			Context("Get", func() {
				AfterEach(func() {
					Expect(userClient.EnsureAuthorizedServiceInputs).To(Equal([]context.Context{ctx}))
				})

				It("returns an error if the user client ensure authorized service returns an error", func() {
					responseErr := errorsTest.RandomError()
					userClient.EnsureAuthorizedServiceOutputs = []error{responseErr}
					blbs, err := client.Get(ctx, id)
					errorsTest.ExpectEqual(err, responseErr)
					Expect(blbs).To(BeNil())
				})

				When("user client ensure authorized service returns successfully", func() {
					BeforeEach(func() {
						userClient.EnsureAuthorizedServiceOutputs = []error{nil}
					})

					AfterEach(func() {
						Expect(blobStructuredSession.GetInputs).To(Equal([]blobStoreStructuredTest.GetInput{{Context: ctx, ID: id}}))
					})

					It("returns an error if the blob structured session get returns an error", func() {
						responseErr := errorsTest.RandomError()
						blobStructuredSession.GetOutputs = []blobStoreStructuredTest.GetOutput{{Blob: nil, Error: responseErr}}
						blbs, err := client.Get(ctx, id)
						errorsTest.ExpectEqual(err, responseErr)
						Expect(blbs).To(BeNil())
					})

					It("returns successfully if the blob structured session get returns successfully", func() {
						responseBlob := blobTest.RandomBlob()
						blobStructuredSession.GetOutputs = []blobStoreStructuredTest.GetOutput{{Blob: responseBlob, Error: nil}}
						blbs, err := client.Get(ctx, id)
						Expect(err).ToNot(HaveOccurred())
						Expect(blbs).To(Equal(responseBlob))
					})
				})
			})

			Context("GetContent", func() {
				AfterEach(func() {
					Expect(userClient.EnsureAuthorizedServiceInputs).To(Equal([]context.Context{ctx}))
				})

				It("returns an error if the user client ensure authorized service returns an error", func() {
					responseErr := errorsTest.RandomError()
					userClient.EnsureAuthorizedServiceOutputs = []error{responseErr}
					content, err := client.GetContent(ctx, id)
					errorsTest.ExpectEqual(err, responseErr)
					Expect(content).To(BeNil())
				})

				When("user client ensure authorized service returns successfully", func() {
					BeforeEach(func() {
						userClient.EnsureAuthorizedServiceOutputs = []error{nil}
					})

					AfterEach(func() {
						Expect(blobStructuredSession.GetInputs).To(Equal([]blobStoreStructuredTest.GetInput{{Context: ctx, ID: id}}))
					})

					It("returns an error if the blob structured session get returns an error", func() {
						responseErr := errorsTest.RandomError()
						blobStructuredSession.GetOutputs = []blobStoreStructuredTest.GetOutput{{Blob: nil, Error: responseErr}}
						content, err := client.GetContent(ctx, id)
						errorsTest.ExpectEqual(err, responseErr)
						Expect(content).To(BeNil())
					})

					It("returns successfully if the blob structured session get returns nil", func() {
						blobStructuredSession.GetOutputs = []blobStoreStructuredTest.GetOutput{{Blob: nil, Error: nil}}
						content, err := client.GetContent(ctx, id)
						Expect(err).ToNot(HaveOccurred())
						Expect(content).To(BeNil())
					})

					When("the blob structure session get returns a blob", func() {
						var blb *blob.Blob

						BeforeEach(func() {
							blb = blobTest.RandomBlob()
							blb.ID = pointer.FromString(id)
							blobStructuredSession.GetOutputs = []blobStoreStructuredTest.GetOutput{{Blob: blb, Error: nil}}
						})

						AfterEach(func() {
							Expect(blobUnstructuredStore.GetInputs).To(Equal([]blobStoreUnstructuredTest.GetInput{{Context: ctx, UserID: *blb.UserID, ID: id}}))
						})

						It("returns an error if the blob unstructured store get returns an error", func() {
							responseErr := errorsTest.RandomError()
							blobUnstructuredStore.GetOutputs = []blobStoreUnstructuredTest.GetOutput{{Reader: nil, Error: responseErr}}
							content, err := client.GetContent(ctx, id)
							errorsTest.ExpectEqual(err, responseErr)
							Expect(content).To(BeNil())
						})

						It("returns successfully if the blob unstructured store get returns successfully", func() {
							body := test.RandomBytes()
							reader := ioutil.NopCloser(bytes.NewReader(body))
							blobUnstructuredStore.GetOutputs = []blobStoreUnstructuredTest.GetOutput{{Reader: reader, Error: nil}}
							content, err := client.GetContent(ctx, id)
							Expect(err).ToNot(HaveOccurred())
							Expect(content).To(Equal(&blob.Content{
								Body:      reader,
								DigestMD5: blb.DigestMD5,
								MediaType: blb.MediaType,
								Size:      blb.Size,
							}))
						})
					})
				})
			})

			Context("Delete", func() {
				AfterEach(func() {
					Expect(userClient.EnsureAuthorizedServiceInputs).To(Equal([]context.Context{ctx}))
				})

				It("returns an error if the user client ensure authorized service returns an error", func() {
					responseErr := errorsTest.RandomError()
					userClient.EnsureAuthorizedServiceOutputs = []error{responseErr}
					deleted, err := client.Delete(ctx, id)
					errorsTest.ExpectEqual(err, responseErr)
					Expect(deleted).To(BeFalse())
				})

				When("user client ensure authorized service returns successfully", func() {
					BeforeEach(func() {
						userClient.EnsureAuthorizedServiceOutputs = []error{nil}
					})

					AfterEach(func() {
						Expect(blobStructuredSession.GetInputs).To(Equal([]blobStoreStructuredTest.GetInput{{Context: ctx, ID: id}}))
					})

					It("returns an error if the blob structured session get returns an error", func() {
						responseErr := errorsTest.RandomError()
						blobStructuredSession.GetOutputs = []blobStoreStructuredTest.GetOutput{{Blob: nil, Error: responseErr}}
						deleted, err := client.Delete(ctx, id)
						errorsTest.ExpectEqual(err, responseErr)
						Expect(deleted).To(BeFalse())
					})

					It("returns successfully if the blob structured session get returns nil", func() {
						blobStructuredSession.GetOutputs = []blobStoreStructuredTest.GetOutput{{Blob: nil, Error: nil}}
						deleted, err := client.Delete(ctx, id)
						Expect(err).ToNot(HaveOccurred())
						Expect(deleted).To(BeFalse())
					})

					When("the blob structure session get returns a blob", func() {
						var blb *blob.Blob

						BeforeEach(func() {
							blb = blobTest.RandomBlob()
							blb.ID = pointer.FromString(id)
							blobStructuredSession.GetOutputs = []blobStoreStructuredTest.GetOutput{{Blob: blb, Error: nil}}
						})

						AfterEach(func() {
							Expect(blobUnstructuredStore.DeleteInputs).To(Equal([]blobStoreUnstructuredTest.DeleteInput{{Context: ctx, UserID: *blb.UserID, ID: id}}))
						})

						It("returns an error if the blob unstructured store delete returns an error", func() {
							responseErr := errorsTest.RandomError()
							blobUnstructuredStore.DeleteOutputs = []blobStoreUnstructuredTest.DeleteOutput{{Deleted: false, Error: responseErr}}
							deleted, err := client.Delete(ctx, id)
							errorsTest.ExpectEqual(err, responseErr)
							Expect(deleted).To(BeFalse())
						})

						When("the blob unstructured store delete returns successfully", func() {
							BeforeEach(func() {
								blobUnstructuredStore.DeleteOutputs = []blobStoreUnstructuredTest.DeleteOutput{{Deleted: true, Error: nil}}
							})

							AfterEach(func() {
								Expect(blobStructuredSession.DeleteInputs).To(Equal([]blobStoreStructuredTest.DeleteInput{{Context: ctx, ID: id}}))
							})

							It("returns an error if the blob structured session delete returns an error", func() {
								responseErr := errorsTest.RandomError()
								blobStructuredSession.DeleteOutputs = []blobStoreStructuredTest.DeleteOutput{{Deleted: false, Error: responseErr}}
								deleted, err := client.Delete(ctx, id)
								errorsTest.ExpectEqual(err, responseErr)
								Expect(deleted).To(BeFalse())
							})

							It("returns false if the blob structured session delete returns false", func() {
								blobStructuredSession.DeleteOutputs = []blobStoreStructuredTest.DeleteOutput{{Deleted: false, Error: nil}}
								deleted, err := client.Delete(ctx, id)
								Expect(err).ToNot(HaveOccurred())
								Expect(deleted).To(BeFalse())
							})

							It("returns true if the blob structured session delete returns true", func() {
								blobStructuredSession.DeleteOutputs = []blobStoreStructuredTest.DeleteOutput{{Deleted: true, Error: nil}}
								deleted, err := client.Delete(ctx, id)
								Expect(err).ToNot(HaveOccurred())
								Expect(deleted).To(BeTrue())
							})

							It("logs a warning if the unstructured store returns false", func() {
								blobUnstructuredStore.DeleteOutputs = []blobStoreUnstructuredTest.DeleteOutput{{Deleted: false, Error: nil}}
								blobStructuredSession.DeleteOutputs = []blobStoreStructuredTest.DeleteOutput{{Deleted: true, Error: nil}}
								deleted, err := client.Delete(ctx, id)
								Expect(err).ToNot(HaveOccurred())
								Expect(deleted).To(BeTrue())
								logger.AssertError("Deleting blob with no content", log.Fields{"id": id})
							})
						})
					})
				})
			})
		})
	})
})
