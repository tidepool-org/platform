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
	"github.com/tidepool-org/platform/crypto"
	cryptoTest "github.com/tidepool-org/platform/crypto/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/image"
	imageServiceClient "github.com/tidepool-org/platform/image/service/client"
	imageServiceClientTest "github.com/tidepool-org/platform/image/service/client/test"
	imageStoreStructured "github.com/tidepool-org/platform/image/store/structured"
	imageStoreStructuredTest "github.com/tidepool-org/platform/image/store/structured/test"
	imageStoreUnstructured "github.com/tidepool-org/platform/image/store/unstructured"
	imageStoreUnstructuredTest "github.com/tidepool-org/platform/image/store/unstructured/test"
	imageTest "github.com/tidepool-org/platform/image/test"
	imageTransform "github.com/tidepool-org/platform/image/transform"
	imageTransformTest "github.com/tidepool-org/platform/image/transform/test"
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
	var imageStructuredStore *imageStoreStructuredTest.Store
	var imageStructuredSession *imageStoreStructuredTest.Session
	var imageUnstructuredStore *imageStoreUnstructuredTest.Store
	var imageTransformer *imageTransformTest.Transformer
	var provider *imageServiceClientTest.Provider

	BeforeEach(func() {
		authClient = authTest.NewClient()
		imageStructuredStore = imageStoreStructuredTest.NewStore()
		imageStructuredSession = imageStoreStructuredTest.NewSession()
		imageStructuredSession.CloseOutput = func(err error) *error { return &err }(nil)
		imageStructuredStore.NewSessionOutput = func(s imageStoreStructured.Session) *imageStoreStructured.Session { return &s }(imageStructuredSession)
		imageUnstructuredStore = imageStoreUnstructuredTest.NewStore()
		imageTransformer = imageTransformTest.NewTransformer()
		provider = imageServiceClientTest.NewProvider()
		provider.AuthClientOutput = func(u auth.Client) *auth.Client { return &u }(authClient)
		provider.ImageStructuredStoreOutput = func(s imageStoreStructured.Store) *imageStoreStructured.Store { return &s }(imageStructuredStore)
		provider.ImageUnstructuredStoreOutput = func(s imageStoreUnstructured.Store) *imageStoreUnstructured.Store { return &s }(imageUnstructuredStore)
		provider.ImageTransformerOutput = func(t imageTransform.Transformer) *imageTransform.Transformer { return &t }(imageTransformer)
	})

	AfterEach(func() {
		provider.AssertOutputsEmpty()
		imageUnstructuredStore.AssertOutputsEmpty()
		imageStructuredStore.AssertOutputsEmpty()
		authClient.AssertOutputsEmpty()
	})

	Context("New", func() {
		It("returns an error when the provider is missing", func() {
			client, err := imageServiceClient.New(nil)
			errorsTest.ExpectEqual(err, errors.New("provider is missing"))
			Expect(client).To(BeNil())
		})

		It("returns successfully", func() {
			Expect(imageServiceClient.New(provider)).ToNot(BeNil())
		})
	})

	Context("with new client", func() {
		var client *imageServiceClient.Client
		var logger *logTest.Logger
		var ctx context.Context

		BeforeEach(func() {
			var err error
			client, err = imageServiceClient.New(provider)
			Expect(err).ToNot(HaveOccurred())
			Expect(client).ToNot(BeNil())
			logger = logTest.NewLogger()
			ctx = log.NewContextWithLogger(context.Background(), logger)
		})

		Context("with user id", func() {
			var userID string

			BeforeEach(func() {
				userID = userTest.RandomID()
			})

			Context("List", func() {
				var filter *image.Filter
				var pagination *page.Pagination

				BeforeEach(func() {
					filter = imageTest.RandomFilter()
					pagination = pageTest.RandomPagination()
				})

				AfterEach(func() {
					Expect(authClient.EnsureAuthorizedUserInputs).To(Equal([]authTest.EnsureAuthorizedUserInput{{TargetUserID: userID, AuthorizedPermission: permission.Read}}))
				})

				It("returns an error when the user client ensure authorized user returns an error", func() {
					responseErr := errorsTest.RandomError()
					authClient.EnsureAuthorizedUserOutputs = []authTest.EnsureAuthorizedUserOutput{{AuthorizedUserID: "", Error: responseErr}}
					result, err := client.List(ctx, userID, filter, pagination)
					errorsTest.ExpectEqual(err, responseErr)
					Expect(result).To(BeNil())
				})

				When("the user client ensure authorized service returns successfully", func() {
					BeforeEach(func() {
						authClient.EnsureAuthorizedUserOutputs = []authTest.EnsureAuthorizedUserOutput{{AuthorizedUserID: userTest.RandomID(), Error: nil}}
					})

					AfterEach(func() {
						Expect(imageStructuredSession.ListInputs).To(Equal([]imageStoreStructuredTest.ListInput{{UserID: userID, Filter: filter, Pagination: pagination}}))
					})

					It("returns an error when the image structured session list returns an error", func() {
						responseErr := errorsTest.RandomError()
						imageStructuredSession.ListOutputs = []imageStoreStructuredTest.ListOutput{{ImageArray: nil, Error: responseErr}}
						result, err := client.List(ctx, userID, filter, pagination)
						errorsTest.ExpectEqual(err, responseErr)
						Expect(result).To(BeNil())
					})

					It("returns successfully when the image structured session list returns successfully", func() {
						responseResult := imageTest.RandomImageArray(1, 3)
						imageStructuredSession.ListOutputs = []imageStoreStructuredTest.ListOutput{{ImageArray: responseResult, Error: nil}}
						result, err := client.List(ctx, userID, filter, pagination)
						Expect(err).ToNot(HaveOccurred())
						Expect(result).To(Equal(responseResult))
					})
				})
			})

			Context("Create", func() {
				var metadata *image.Metadata
				var contentIntent string
				var width int
				var height int
				var content *image.Content

				BeforeEach(func() {
					metadata = imageTest.RandomMetadata()
					contentIntent = imageTest.RandomContentIntent()
					width = test.RandomIntFromRange(10, 20)
					height = test.RandomIntFromRange(10, 20)
					content = imageTest.RandomContentFromDimensions(width, height)
				})

				AfterEach(func() {
					Expect(authClient.EnsureAuthorizedUserInputs).To(Equal([]authTest.EnsureAuthorizedUserInput{{TargetUserID: userID, AuthorizedPermission: permission.Write}}))
				})

				It("returns an error when the user client ensure authorized user returns an error", func() {
					responseErr := errorsTest.RandomError()
					authClient.EnsureAuthorizedUserOutputs = []authTest.EnsureAuthorizedUserOutput{{AuthorizedUserID: "", Error: responseErr}}
					result, err := client.Create(ctx, userID, metadata, contentIntent, content)
					errorsTest.ExpectEqual(err, responseErr)
					Expect(result).To(BeNil())
				})

				When("the user client ensure authorized user returns successfully", func() {
					BeforeEach(func() {
						authClient.EnsureAuthorizedUserOutputs = []authTest.EnsureAuthorizedUserOutput{{AuthorizedUserID: userTest.RandomID(), Error: nil}}
					})

					It("returns an error when metadata is missing", func() {
						metadata = nil
						result, err := client.Create(ctx, userID, metadata, contentIntent, content)
						errorsTest.ExpectEqual(err, errors.New("metadata is missing"))
						Expect(result).To(BeNil())
					})

					It("returns an error when metadata is invalid", func() {
						metadata.Name = pointer.FromString("")
						result, err := client.Create(ctx, userID, metadata, contentIntent, content)
						errorsTest.ExpectEqual(err, errors.New("metadata is invalid"))
						Expect(result).To(BeNil())
					})

					It("returns an error when content intent is missing", func() {
						contentIntent = ""
						result, err := client.Create(ctx, userID, metadata, contentIntent, content)
						errorsTest.ExpectEqual(err, errors.New("content intent is missing"))
						Expect(result).To(BeNil())
					})

					It("returns an error when content intent is invalid", func() {
						contentIntent = "invalid"
						result, err := client.Create(ctx, userID, metadata, contentIntent, content)
						errorsTest.ExpectEqual(err, errors.New("content intent is invalid"))
						Expect(result).To(BeNil())
					})

					It("returns an error when content is missing", func() {
						content = nil
						result, err := client.Create(ctx, userID, metadata, contentIntent, content)
						errorsTest.ExpectEqual(err, errors.New("content is missing"))
						Expect(result).To(BeNil())
					})

					It("returns an error when content is invalid", func() {
						content.Body = nil
						result, err := client.Create(ctx, userID, metadata, contentIntent, content)
						errorsTest.ExpectEqual(err, errors.New("content is invalid"))
						Expect(result).To(BeNil())
					})

					When("the image is created", func() {
						AfterEach(func() {
							Expect(imageStructuredSession.CreateInputs).To(Equal([]imageStoreStructuredTest.CreateInput{{UserID: userID, Metadata: metadata}}))
						})

						It("returns an error when the image structured session create returns an error", func() {
							responseErr := errorsTest.RandomError()
							imageStructuredSession.CreateOutputs = []imageStoreStructuredTest.CreateOutput{{Image: nil, Error: responseErr}}
							result, err := client.Create(ctx, userID, metadata, contentIntent, content)
							errorsTest.ExpectEqual(err, responseErr)
							Expect(result).To(BeNil())
						})

						When("the image structured session create returns successfully", func() {
							var createImage *image.Image
							var destroyImageErr error

							BeforeEach(func() {
								createImage = imageTest.RandomImage()
								createImage.UserID = pointer.FromString(userID)
								createImage.Status = pointer.FromString(image.StatusCreated)
								createImage.ModifiedTime = nil
								imageStructuredSession.CreateOutputs = []imageStoreStructuredTest.CreateOutput{{Image: createImage, Error: nil}}
								destroyImageErr = errorsTest.RandomError()
							})

							JustBeforeEach(func() {
								if destroyImageErr != nil {
									imageStructuredSession.DestroyOutputs = []imageStoreStructuredTest.DestroyOutput{{Destroyed: false, Error: destroyImageErr}}
								}
							})

							AfterEach(func() {
								if destroyImageErr != nil {
									Expect(imageStructuredSession.DestroyInputs).To(Equal([]imageStoreStructuredTest.DestroyInput{{ID: *createImage.ID, Condition: nil}}))
									logger.AssertError("Unable to destroy image after failure to put image content", log.Fields{"userId": userID, "id": *createImage.ID, "error": errors.NewSerializable(destroyImageErr)})
								}
							})

							It("returns an error when content is invalid", func() {
								content.Body = ioutil.NopCloser(bytes.NewReader(test.RandomBytes()))
								result, err := client.Create(ctx, userID, metadata, contentIntent, content)
								errorsTest.ExpectEqual(err, image.ErrorImageMalformed("unable to decode image; image: unknown format"))
								Expect(result).To(BeNil())
							})

							It("returns an error when content does not match media type", func() {
								switch *content.MediaType {
								case image.MediaTypeImageJPEG:
									content.MediaType = pointer.FromString(image.MediaTypeImagePNG)
								case image.MediaTypeImagePNG:
									content.MediaType = pointer.FromString(image.MediaTypeImageJPEG)
								}
								result, err := client.Create(ctx, userID, metadata, contentIntent, content)
								errorsTest.ExpectEqual(err, image.ErrorImageMalformed("header does not match media type"))
								Expect(result).To(BeNil())
							})

							When("the inmage unstructured store put content is invoked", func() {
								AfterEach(func() {
									Expect(imageUnstructuredStore.PutContentInputs).To(HaveLen(1))
									Expect(imageUnstructuredStore.PutContentInputs[0].UserID).To(Equal(userID))
									Expect(imageUnstructuredStore.PutContentInputs[0].ImageID).To(Equal(*createImage.ID))
									Expect(imageUnstructuredStore.PutContentInputs[0].ContentIntent).To(Equal(contentIntent))
									Expect(imageUnstructuredStore.PutContentInputs[0].Reader).ToNot(BeNil())
									Expect(imageUnstructuredStore.PutContentInputs[0].Options).To(Equal(&storeUnstructured.Options{MediaType: content.MediaType}))
								})

								It("returns an error when the image unstructured store put content returns an error", func() {
									responseErr := errorsTest.RandomError()
									imageUnstructuredStore.PutContentOutputs = []error{responseErr}
									result, err := client.Create(ctx, userID, metadata, contentIntent, content)
									errorsTest.ExpectEqual(err, responseErr)
									Expect(result).To(BeNil())
								})

								When("the image unstructured store put content returns successfully", func() {
									var putContentID string
									var size int64

									BeforeEach(func() {
										putContentID = ""
										imageUnstructuredStore.PutContentStub = func(ctx context.Context, userID string, imageID string, contentID string, contentIntent string, reader io.Reader, options *storeUnstructured.Options) error {
											putContentID = contentID
											var err error
											size, err = io.Copy(ioutil.Discard, reader)
											return err
										}
									})

									When("the size exceeds maximum", func() {
										BeforeEach(func() {
											contentBytes := make([]byte, 104857601)
											content.Body.Read(contentBytes)
											content.Body = ioutil.NopCloser(bytes.NewReader(contentBytes))
											content.DigestMD5 = pointer.FromString(crypto.Base64EncodedMD5Hash(contentBytes))
										})

										AfterEach(func() {
											Expect(imageUnstructuredStore.DeleteContentInputs).To(Equal([]imageStoreUnstructuredTest.DeleteContentInput{{UserID: userID, ImageID: *createImage.ID, ContentID: putContentID}}))
											Expect(imageStructuredSession.DestroyInputs).To(Equal([]imageStoreStructuredTest.DestroyInput{{ID: *createImage.ID}}))
										})

										It("returns an error", func() {
											imageUnstructuredStore.DeleteContentOutputs = []error{nil}
											result, err := client.Create(ctx, userID, metadata, contentIntent, content)
											errorsTest.ExpectEqual(err, request.ErrorResourceTooLarge())
											Expect(result).To(BeNil())
											imageUnstructuredStore.DeleteContentInputs = []imageStoreUnstructuredTest.DeleteContentInput{{UserID: *createImage.UserID, ImageID: *createImage.ID, ContentID: putContentID}}
										})

										It("returns an error and logs an error when both the unstructured store delete content", func() {
											responseErr := errorsTest.RandomError()
											imageUnstructuredStore.DeleteContentOutputs = []error{responseErr}
											result, err := client.Create(ctx, userID, metadata, contentIntent, content)
											errorsTest.ExpectEqual(err, request.ErrorResourceTooLarge())
											Expect(result).To(BeNil())
											logger.AssertError("Unable to delete image content exceeding maximum size", log.Fields{"userId": userID, "id": *createImage.ID, "error": errors.NewSerializable(responseErr)})
											imageUnstructuredStore.DeleteContentInputs = []imageStoreUnstructuredTest.DeleteContentInput{{UserID: *createImage.UserID, ImageID: *createImage.ID, ContentID: putContentID}}
										})
									})

									When("the digest does not match", func() {
										var digestMD5 string

										BeforeEach(func() {
											digestMD5 = *content.DigestMD5
											content.DigestMD5 = pointer.FromString(cryptoTest.RandomBase64EncodedMD5Hash())
										})

										AfterEach(func() {
											Expect(imageUnstructuredStore.DeleteContentInputs).To(Equal([]imageStoreUnstructuredTest.DeleteContentInput{{UserID: userID, ImageID: *createImage.ID, ContentID: putContentID}}))
										})

										It("returns an error", func() {
											imageUnstructuredStore.DeleteContentOutputs = []error{nil}
											result, err := client.Create(ctx, userID, metadata, contentIntent, content)
											errorsTest.ExpectEqual(err, errorsTest.WithPointerSource(request.ErrorDigestsNotEqual(*content.DigestMD5, digestMD5), "/digestMD5"))
											Expect(result).To(BeNil())
											imageUnstructuredStore.DeleteContentInputs = []imageStoreUnstructuredTest.DeleteContentInput{{UserID: *createImage.UserID, ImageID: *createImage.ID, ContentID: putContentID}}
										})

										It("returns an error and logs an error when both the unstructured and structured store returns an error", func() {
											responseErr := errorsTest.RandomError()
											imageUnstructuredStore.DeleteContentOutputs = []error{responseErr}
											result, err := client.Create(ctx, userID, metadata, contentIntent, content)
											errorsTest.ExpectEqual(err, errorsTest.WithPointerSource(request.ErrorDigestsNotEqual(*content.DigestMD5, digestMD5), "/digestMD5"))
											Expect(result).To(BeNil())
											logger.AssertError("Unable to delete image content with incorrect MD5 digest", log.Fields{"userId": userID, "id": *createImage.ID, "error": errors.NewSerializable(responseErr)})
											imageUnstructuredStore.DeleteContentInputs = []imageStoreUnstructuredTest.DeleteContentInput{{UserID: *createImage.UserID, ImageID: *createImage.ID, ContentID: putContentID}}
										})
									})

									When("the size and digest are valid", func() {
										AfterEach(func() {
											update := imageStoreStructured.NewUpdate()
											update.ContentID = pointer.FromString(putContentID)
											update.ContentIntent = pointer.FromString(contentIntent)
											update.ContentAttributes = imageStoreStructured.NewContentAttributes()
											update.ContentAttributes.DigestMD5 = pointer.CloneString(content.DigestMD5)
											update.ContentAttributes.MediaType = pointer.CloneString(content.MediaType)
											update.ContentAttributes.Width = pointer.FromInt(width)
											update.ContentAttributes.Height = pointer.FromInt(height)
											update.ContentAttributes.Size = pointer.FromInt(int(size))
											Expect(imageStructuredSession.UpdateInputs).To(Equal([]imageStoreStructuredTest.UpdateInput{{ID: *createImage.ID, Condition: &request.Condition{Revision: createImage.Revision}, Update: update}}))
										})

										It("returns an error when image structured session update returns an error", func() {
											responseErr := errorsTest.RandomError()
											imageStructuredSession.UpdateOutputs = []imageStoreStructuredTest.UpdateOutput{{Image: nil, Error: responseErr}}
											imageUnstructuredStore.DeleteContentOutputs = []error{nil}
											result, err := client.Create(ctx, userID, metadata, contentIntent, content)
											errorsTest.ExpectEqual(err, responseErr)
											Expect(result).To(BeNil())
											imageUnstructuredStore.DeleteContentInputs = []imageStoreUnstructuredTest.DeleteContentInput{{UserID: *createImage.UserID, ImageID: *createImage.ID, ContentID: putContentID}}
										})

										When("the image structured session update returns successfully", func() {
											var updateImage *image.Image

											BeforeEach(func() {
												destroyImageErr = nil
												updateImage = imageTest.CloneImage(createImage)
												updateImage.Status = pointer.FromString(image.StatusAvailable)
												updateImage.ContentIntent = pointer.FromString(contentIntent)
												updateImage.ContentAttributes = image.NewContentAttributes()
												updateImage.ContentAttributes.DigestMD5 = pointer.CloneString(content.DigestMD5)
												updateImage.ContentAttributes.MediaType = pointer.CloneString(content.MediaType)
												updateImage.ContentAttributes.Width = pointer.FromInt(width)
												updateImage.ContentAttributes.Height = pointer.FromInt(height)
												updateImage.ContentAttributes.Size = pointer.FromInt(int(size))
												updateImage.ModifiedTime = pointer.FromTime(test.RandomTimeFromRange(*createImage.CreatedTime, time.Now()).Truncate(time.Second))
												imageStructuredSession.UpdateOutputs = []imageStoreStructuredTest.UpdateOutput{{Image: updateImage, Error: nil}}
											})

											When("the content intent is alternate", func() {
												BeforeEach(func() {
													contentIntent = image.ContentIntentAlternate
												})

												It("returns successfully", func() {
													Expect(client.Create(ctx, userID, metadata, contentIntent, content)).To(Equal(updateImage))
												})
											})

											When("the content intent is original", func() {
												BeforeEach(func() {
													contentIntent = image.ContentIntentOriginal
												})

												It("returns successfully", func() {
													Expect(client.Create(ctx, userID, metadata, contentIntent, content)).To(Equal(updateImage))
												})
											})

											When("the size is maximum", func() {
												BeforeEach(func() {
													contentBytes := make([]byte, 104857600)
													content.Body.Read(contentBytes)
													content.Body = ioutil.NopCloser(bytes.NewReader(contentBytes))
													content.DigestMD5 = pointer.FromString(crypto.Base64EncodedMD5Hash(contentBytes))
												})

												It("returns successfully", func() {
													Expect(client.Create(ctx, userID, metadata, contentIntent, content)).To(Equal(updateImage))
												})
											})
										})
									})
								})
							})
						})
					})
				})
			})

			Context("CreateWithMetadata", func() {
				var metadata *image.Metadata

				BeforeEach(func() {
					metadata = imageTest.RandomMetadata()
				})

				AfterEach(func() {
					Expect(authClient.EnsureAuthorizedUserInputs).To(Equal([]authTest.EnsureAuthorizedUserInput{{TargetUserID: userID, AuthorizedPermission: permission.Write}}))
				})

				It("returns an error when the user client ensure authorized user returns an error", func() {
					responseErr := errorsTest.RandomError()
					authClient.EnsureAuthorizedUserOutputs = []authTest.EnsureAuthorizedUserOutput{{AuthorizedUserID: "", Error: responseErr}}
					result, err := client.CreateWithMetadata(ctx, userID, metadata)
					errorsTest.ExpectEqual(err, responseErr)
					Expect(result).To(BeNil())
				})

				When("the user client ensure authorized service returns successfully", func() {
					BeforeEach(func() {
						authClient.EnsureAuthorizedUserOutputs = []authTest.EnsureAuthorizedUserOutput{{AuthorizedUserID: userTest.RandomID(), Error: nil}}
					})

					AfterEach(func() {
						Expect(imageStructuredSession.CreateInputs).To(Equal([]imageStoreStructuredTest.CreateInput{{UserID: userID, Metadata: metadata}}))
					})

					It("returns an error when the image structured session create returns an error", func() {
						responseErr := errorsTest.RandomError()
						imageStructuredSession.CreateOutputs = []imageStoreStructuredTest.CreateOutput{{Image: nil, Error: responseErr}}
						result, err := client.CreateWithMetadata(ctx, userID, metadata)
						errorsTest.ExpectEqual(err, responseErr)
						Expect(result).To(BeNil())
					})

					It("returns successfully when the image structured session create returns successfully", func() {
						responseResult := imageTest.RandomImage()
						imageStructuredSession.CreateOutputs = []imageStoreStructuredTest.CreateOutput{{Image: responseResult, Error: nil}}
						result, err := client.CreateWithMetadata(ctx, userID, metadata)
						Expect(err).ToNot(HaveOccurred())
						Expect(result).To(Equal(responseResult))
					})
				})
			})

			Context("CreateWithContent", func() {
				var contentIntent string
				var width int
				var height int
				var content *image.Content

				BeforeEach(func() {
					contentIntent = imageTest.RandomContentIntent()
					width = test.RandomIntFromRange(10, 20)
					height = test.RandomIntFromRange(10, 20)
					content = imageTest.RandomContentFromDimensions(width, height)
				})

				AfterEach(func() {
					Expect(authClient.EnsureAuthorizedUserInputs).To(Equal([]authTest.EnsureAuthorizedUserInput{{TargetUserID: userID, AuthorizedPermission: permission.Write}}))
				})

				It("returns an error when the user client ensure authorized user returns an error", func() {
					responseErr := errorsTest.RandomError()
					authClient.EnsureAuthorizedUserOutputs = []authTest.EnsureAuthorizedUserOutput{{AuthorizedUserID: "", Error: responseErr}}
					result, err := client.CreateWithContent(ctx, userID, contentIntent, content)
					errorsTest.ExpectEqual(err, responseErr)
					Expect(result).To(BeNil())
				})

				When("the user client ensure authorized user returns successfully", func() {
					BeforeEach(func() {
						authClient.EnsureAuthorizedUserOutputs = []authTest.EnsureAuthorizedUserOutput{{AuthorizedUserID: userTest.RandomID(), Error: nil}}
					})

					It("returns an error when content intent is missing", func() {
						contentIntent = ""
						result, err := client.CreateWithContent(ctx, userID, contentIntent, content)
						errorsTest.ExpectEqual(err, errors.New("content intent is missing"))
						Expect(result).To(BeNil())
					})

					It("returns an error when content intent is invalid", func() {
						contentIntent = "invalid"
						result, err := client.CreateWithContent(ctx, userID, contentIntent, content)
						errorsTest.ExpectEqual(err, errors.New("content intent is invalid"))
						Expect(result).To(BeNil())
					})

					It("returns an error when content is missing", func() {
						content = nil
						result, err := client.CreateWithContent(ctx, userID, contentIntent, content)
						errorsTest.ExpectEqual(err, errors.New("content is missing"))
						Expect(result).To(BeNil())
					})

					It("returns an error when content is invalid", func() {
						content.Body = nil
						result, err := client.CreateWithContent(ctx, userID, contentIntent, content)
						errorsTest.ExpectEqual(err, errors.New("content is invalid"))
						Expect(result).To(BeNil())
					})

					When("the image is created", func() {
						AfterEach(func() {
							Expect(imageStructuredSession.CreateInputs).To(Equal([]imageStoreStructuredTest.CreateInput{{UserID: userID, Metadata: image.NewMetadata()}}))
						})

						It("returns an error when the image structured session create returns an error", func() {
							responseErr := errorsTest.RandomError()
							imageStructuredSession.CreateOutputs = []imageStoreStructuredTest.CreateOutput{{Image: nil, Error: responseErr}}
							result, err := client.CreateWithContent(ctx, userID, contentIntent, content)
							errorsTest.ExpectEqual(err, responseErr)
							Expect(result).To(BeNil())
						})

						When("the image structured session create returns successfully", func() {
							var createImage *image.Image
							var destroyImageErr error

							BeforeEach(func() {
								createImage = imageTest.RandomImage()
								createImage.UserID = pointer.FromString(userID)
								createImage.Status = pointer.FromString(image.StatusCreated)
								createImage.ModifiedTime = nil
								imageStructuredSession.CreateOutputs = []imageStoreStructuredTest.CreateOutput{{Image: createImage, Error: nil}}
								destroyImageErr = errorsTest.RandomError()
							})

							JustBeforeEach(func() {
								if destroyImageErr != nil {
									imageStructuredSession.DestroyOutputs = []imageStoreStructuredTest.DestroyOutput{{Destroyed: false, Error: destroyImageErr}}
								}
							})

							AfterEach(func() {
								if destroyImageErr != nil {
									Expect(imageStructuredSession.DestroyInputs).To(Equal([]imageStoreStructuredTest.DestroyInput{{ID: *createImage.ID, Condition: nil}}))
									logger.AssertError("Unable to destroy image after failure to put image content", log.Fields{"userId": userID, "id": *createImage.ID, "error": errors.NewSerializable(destroyImageErr)})
								}
							})

							It("returns an error when content is invalid", func() {
								content.Body = ioutil.NopCloser(bytes.NewReader(test.RandomBytes()))
								result, err := client.CreateWithContent(ctx, userID, contentIntent, content)
								errorsTest.ExpectEqual(err, image.ErrorImageMalformed("unable to decode image; image: unknown format"))
								Expect(result).To(BeNil())
							})

							It("returns an error when content does not match media type", func() {
								switch *content.MediaType {
								case image.MediaTypeImageJPEG:
									content.MediaType = pointer.FromString(image.MediaTypeImagePNG)
								case image.MediaTypeImagePNG:
									content.MediaType = pointer.FromString(image.MediaTypeImageJPEG)
								}
								result, err := client.CreateWithContent(ctx, userID, contentIntent, content)
								errorsTest.ExpectEqual(err, image.ErrorImageMalformed("header does not match media type"))
								Expect(result).To(BeNil())
							})

							When("the image unstructured store put content is invoked", func() {
								AfterEach(func() {
									Expect(imageUnstructuredStore.PutContentInputs).To(HaveLen(1))
									Expect(imageUnstructuredStore.PutContentInputs[0].UserID).To(Equal(userID))
									Expect(imageUnstructuredStore.PutContentInputs[0].ImageID).To(Equal(*createImage.ID))
									Expect(imageUnstructuredStore.PutContentInputs[0].ContentIntent).To(Equal(contentIntent))
									Expect(imageUnstructuredStore.PutContentInputs[0].Reader).ToNot(BeNil())
									Expect(imageUnstructuredStore.PutContentInputs[0].Options).To(Equal(&storeUnstructured.Options{MediaType: content.MediaType}))
								})

								It("returns an error when the image unstructured store put content returns an error", func() {
									responseErr := errorsTest.RandomError()
									imageUnstructuredStore.PutContentOutputs = []error{responseErr}
									result, err := client.CreateWithContent(ctx, userID, contentIntent, content)
									errorsTest.ExpectEqual(err, responseErr)
									Expect(result).To(BeNil())
								})

								When("the image unstructured store put content returns successfully", func() {
									var putContentID string
									var size int64

									BeforeEach(func() {
										putContentID = ""
										imageUnstructuredStore.PutContentStub = func(ctx context.Context, userID string, imageID string, contentID string, contentIntent string, reader io.Reader, options *storeUnstructured.Options) error {
											putContentID = contentID
											var err error
											size, err = io.Copy(ioutil.Discard, reader)
											return err
										}
									})

									When("the size exceeds maximum", func() {
										BeforeEach(func() {
											contentBytes := make([]byte, 104857601)
											content.Body.Read(contentBytes)
											content.Body = ioutil.NopCloser(bytes.NewReader(contentBytes))
											content.DigestMD5 = pointer.FromString(crypto.Base64EncodedMD5Hash(contentBytes))
										})

										AfterEach(func() {
											Expect(imageUnstructuredStore.DeleteContentInputs).To(Equal([]imageStoreUnstructuredTest.DeleteContentInput{{UserID: userID, ImageID: *createImage.ID, ContentID: putContentID}}))
										})

										It("returns an error", func() {
											imageUnstructuredStore.DeleteContentOutputs = []error{nil}
											result, err := client.CreateWithContent(ctx, userID, contentIntent, content)
											errorsTest.ExpectEqual(err, request.ErrorResourceTooLarge())
											Expect(result).To(BeNil())
											imageUnstructuredStore.DeleteContentInputs = []imageStoreUnstructuredTest.DeleteContentInput{{UserID: *createImage.UserID, ImageID: *createImage.ID, ContentID: putContentID}}
										})

										It("returns an error and logs an error when both the unstructured store delete content", func() {
											responseErr := errorsTest.RandomError()
											imageUnstructuredStore.DeleteContentOutputs = []error{responseErr}
											result, err := client.CreateWithContent(ctx, userID, contentIntent, content)
											errorsTest.ExpectEqual(err, request.ErrorResourceTooLarge())
											Expect(result).To(BeNil())
											logger.AssertError("Unable to delete image content exceeding maximum size", log.Fields{"userId": userID, "id": *createImage.ID, "error": errors.NewSerializable(responseErr)})
											imageUnstructuredStore.DeleteContentInputs = []imageStoreUnstructuredTest.DeleteContentInput{{UserID: *createImage.UserID, ImageID: *createImage.ID, ContentID: putContentID}}
										})
									})

									When("the digest does not match", func() {
										var digestMD5 string

										BeforeEach(func() {
											digestMD5 = *content.DigestMD5
											content.DigestMD5 = pointer.FromString(cryptoTest.RandomBase64EncodedMD5Hash())
										})

										AfterEach(func() {
											Expect(imageUnstructuredStore.DeleteContentInputs).To(Equal([]imageStoreUnstructuredTest.DeleteContentInput{{UserID: userID, ImageID: *createImage.ID, ContentID: putContentID}}))
										})

										It("returns an error", func() {
											imageUnstructuredStore.DeleteContentOutputs = []error{nil}
											result, err := client.CreateWithContent(ctx, userID, contentIntent, content)
											errorsTest.ExpectEqual(err, errorsTest.WithPointerSource(request.ErrorDigestsNotEqual(*content.DigestMD5, digestMD5), "/digestMD5"))
											Expect(result).To(BeNil())
											imageUnstructuredStore.DeleteContentInputs = []imageStoreUnstructuredTest.DeleteContentInput{{UserID: *createImage.UserID, ImageID: *createImage.ID, ContentID: putContentID}}
										})

										It("returns an error and logs an error when both the unstructured and structured store returns an error", func() {
											responseErr := errorsTest.RandomError()
											imageUnstructuredStore.DeleteContentOutputs = []error{responseErr}
											result, err := client.CreateWithContent(ctx, userID, contentIntent, content)
											errorsTest.ExpectEqual(err, errorsTest.WithPointerSource(request.ErrorDigestsNotEqual(*content.DigestMD5, digestMD5), "/digestMD5"))
											Expect(result).To(BeNil())
											logger.AssertError("Unable to delete image content with incorrect MD5 digest", log.Fields{"userId": userID, "id": *createImage.ID, "error": errors.NewSerializable(responseErr)})
											imageUnstructuredStore.DeleteContentInputs = []imageStoreUnstructuredTest.DeleteContentInput{{UserID: *createImage.UserID, ImageID: *createImage.ID, ContentID: putContentID}}
										})
									})

									When("the size and digest are valid", func() {
										AfterEach(func() {
											update := imageStoreStructured.NewUpdate()
											update.ContentID = pointer.FromString(putContentID)
											update.ContentIntent = pointer.FromString(contentIntent)
											update.ContentAttributes = imageStoreStructured.NewContentAttributes()
											update.ContentAttributes.DigestMD5 = pointer.CloneString(content.DigestMD5)
											update.ContentAttributes.MediaType = pointer.CloneString(content.MediaType)
											update.ContentAttributes.Width = pointer.FromInt(width)
											update.ContentAttributes.Height = pointer.FromInt(height)
											update.ContentAttributes.Size = pointer.FromInt(int(size))
											Expect(imageStructuredSession.UpdateInputs).To(Equal([]imageStoreStructuredTest.UpdateInput{{ID: *createImage.ID, Condition: &request.Condition{Revision: createImage.Revision}, Update: update}}))
										})

										It("returns an error when image structured session update returns an error", func() {
											responseErr := errorsTest.RandomError()
											imageStructuredSession.UpdateOutputs = []imageStoreStructuredTest.UpdateOutput{{Image: nil, Error: responseErr}}
											imageUnstructuredStore.DeleteContentOutputs = []error{nil}
											result, err := client.CreateWithContent(ctx, userID, contentIntent, content)
											errorsTest.ExpectEqual(err, responseErr)
											Expect(result).To(BeNil())
											imageUnstructuredStore.DeleteContentInputs = []imageStoreUnstructuredTest.DeleteContentInput{{UserID: *createImage.UserID, ImageID: *createImage.ID, ContentID: putContentID}}
										})

										When("the image structured session update returns successfully", func() {
											var updateImage *image.Image

											BeforeEach(func() {
												destroyImageErr = nil
												updateImage = imageTest.CloneImage(createImage)
												updateImage.Status = pointer.FromString(image.StatusAvailable)
												updateImage.ContentIntent = pointer.FromString(contentIntent)
												updateImage.ContentAttributes = image.NewContentAttributes()
												updateImage.ContentAttributes.DigestMD5 = pointer.CloneString(content.DigestMD5)
												updateImage.ContentAttributes.MediaType = pointer.CloneString(content.MediaType)
												updateImage.ContentAttributes.Width = pointer.FromInt(width)
												updateImage.ContentAttributes.Height = pointer.FromInt(height)
												updateImage.ContentAttributes.Size = pointer.FromInt(int(size))
												updateImage.ModifiedTime = pointer.FromTime(test.RandomTimeFromRange(*createImage.CreatedTime, time.Now()).Truncate(time.Second))
												imageStructuredSession.UpdateOutputs = []imageStoreStructuredTest.UpdateOutput{{Image: updateImage, Error: nil}}
											})

											When("the content intent is alternate", func() {
												BeforeEach(func() {
													contentIntent = image.ContentIntentAlternate
												})

												It("returns successfully", func() {
													Expect(client.CreateWithContent(ctx, userID, contentIntent, content)).To(Equal(updateImage))
												})
											})

											When("the content intent is original", func() {
												BeforeEach(func() {
													contentIntent = image.ContentIntentOriginal
												})

												It("returns successfully", func() {
													Expect(client.CreateWithContent(ctx, userID, contentIntent, content)).To(Equal(updateImage))
												})
											})

											When("the size is maximum", func() {
												BeforeEach(func() {
													contentBytes := make([]byte, 104857600)
													content.Body.Read(contentBytes)
													content.Body = ioutil.NopCloser(bytes.NewReader(contentBytes))
													content.DigestMD5 = pointer.FromString(crypto.Base64EncodedMD5Hash(contentBytes))
												})

												It("returns successfully", func() {
													Expect(client.CreateWithContent(ctx, userID, contentIntent, content)).To(Equal(updateImage))
												})
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
						Expect(imageStructuredSession.DeleteAllInputs).To(Equal([]string{userID}))
					})

					It("returns an error when the image structured session delete returns an error", func() {
						responseErr := errorsTest.RandomError()
						imageStructuredSession.DeleteAllOutputs = []imageStoreStructuredTest.DeleteAllOutput{{Deleted: false, Error: responseErr}}
						errorsTest.ExpectEqual(client.DeleteAll(ctx, userID), responseErr)
					})

					It("returns successfully when the image structured session delete returns successfully without deleted", func() {
						imageStructuredSession.DeleteAllOutputs = []imageStoreStructuredTest.DeleteAllOutput{{Deleted: false, Error: nil}}
						Expect(client.DeleteAll(ctx, userID)).To(Succeed())
					})

					When("the image structured session delete returns successfully with deleted", func() {
						BeforeEach(func() {
							imageStructuredSession.DeleteAllOutputs = []imageStoreStructuredTest.DeleteAllOutput{{Deleted: true, Error: nil}}
						})

						AfterEach(func() {
							Expect(imageUnstructuredStore.DeleteAllInputs).To(Equal([]string{userID}))
						})

						It("returns an error when the image unstructured store delete all returns an error", func() {
							responseErr := errorsTest.RandomError()
							imageUnstructuredStore.DeleteAllOutputs = []error{responseErr}
							errorsTest.ExpectEqual(client.DeleteAll(ctx, userID), responseErr)
						})

						When("the image unstructured store delete all returns successfully", func() {
							BeforeEach(func() {
								imageUnstructuredStore.DeleteAllOutputs = []error{nil}
							})

							AfterEach(func() {
								Expect(imageStructuredSession.DestroyAllInputs).To(Equal([]string{userID}))
							})

							It("returns an error when the image structured session destroy returns an error", func() {
								responseErr := errorsTest.RandomError()
								imageStructuredSession.DestroyAllOutputs = []imageStoreStructuredTest.DestroyAllOutput{{Destroyed: false, Error: responseErr}}
								errorsTest.ExpectEqual(client.DeleteAll(ctx, userID), responseErr)
							})

							It("returns successfully when the image structured session destroy returns false", func() {
								imageStructuredSession.DestroyAllOutputs = []imageStoreStructuredTest.DestroyAllOutput{{Destroyed: false, Error: nil}}
								Expect(client.DeleteAll(ctx, userID)).To(Succeed())
							})

							It("returns successfully when the image structured session destroy returns true", func() {
								imageStructuredSession.DestroyAllOutputs = []imageStoreStructuredTest.DestroyAllOutput{{Destroyed: true, Error: nil}}
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
				id = imageTest.RandomID()
			})

			Context("Get", func() {
				AfterEach(func() {
					Expect(imageStructuredSession.GetInputs).To(Equal([]imageStoreStructuredTest.GetInput{{ID: id, Condition: nil}}))
				})

				It("returns an error when the image structured session get returns an error", func() {
					responseErr := errorsTest.RandomError()
					imageStructuredSession.GetOutputs = []imageStoreStructuredTest.GetOutput{{Image: nil, Error: responseErr}}
					result, err := client.Get(ctx, id)
					errorsTest.ExpectEqual(err, responseErr)
					Expect(result).To(BeNil())
				})

				It("returns successfully when the image structured session get returns nil", func() {
					imageStructuredSession.GetOutputs = []imageStoreStructuredTest.GetOutput{{Image: nil, Error: nil}}
					deleted, err := client.Get(ctx, id)
					Expect(err).ToNot(HaveOccurred())
					Expect(deleted).To(BeNil())
				})

				When("the image structure session get returns an image", func() {
					var responseResult *image.Image

					BeforeEach(func() {
						responseResult = imageTest.RandomImage()
						imageStructuredSession.GetOutputs = []imageStoreStructuredTest.GetOutput{{Image: responseResult, Error: nil}}
					})

					AfterEach(func() {
						Expect(authClient.EnsureAuthorizedUserInputs).To(Equal([]authTest.EnsureAuthorizedUserInput{{TargetUserID: *responseResult.UserID, AuthorizedPermission: permission.Write}}))
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
							authClient.EnsureAuthorizedUserOutputs = []authTest.EnsureAuthorizedUserOutput{{AuthorizedUserID: *responseResult.UserID, Error: nil}}
						})

						It("returns successfully", func() {
							result, err := client.Get(ctx, id)
							Expect(err).ToNot(HaveOccurred())
							Expect(result).To(Equal(responseResult))
						})
					})
				})
			})

			Context("GetMetadata", func() {
				AfterEach(func() {
					Expect(imageStructuredSession.GetInputs).To(Equal([]imageStoreStructuredTest.GetInput{{ID: id, Condition: nil}}))
				})

				It("returns an error when the image structured session get returns an error", func() {
					responseErr := errorsTest.RandomError()
					imageStructuredSession.GetOutputs = []imageStoreStructuredTest.GetOutput{{Image: nil, Error: responseErr}}
					result, err := client.GetMetadata(ctx, id)
					errorsTest.ExpectEqual(err, responseErr)
					Expect(result).To(BeNil())
				})

				It("returns successfully when the image structured session get returns nil", func() {
					imageStructuredSession.GetOutputs = []imageStoreStructuredTest.GetOutput{{Image: nil, Error: nil}}
					deleted, err := client.GetMetadata(ctx, id)
					Expect(err).ToNot(HaveOccurred())
					Expect(deleted).To(BeNil())
				})

				When("the image structure session get returns an image", func() {
					var responseResult *image.Image

					BeforeEach(func() {
						responseResult = imageTest.RandomImage()
						imageStructuredSession.GetOutputs = []imageStoreStructuredTest.GetOutput{{Image: responseResult, Error: nil}}
					})

					AfterEach(func() {
						Expect(authClient.EnsureAuthorizedUserInputs).To(Equal([]authTest.EnsureAuthorizedUserInput{{TargetUserID: *responseResult.UserID, AuthorizedPermission: permission.Write}}))
					})

					It("returns an error when the user client ensure authorized service returns an error", func() {
						responseErr := errorsTest.RandomError()
						authClient.EnsureAuthorizedUserOutputs = []authTest.EnsureAuthorizedUserOutput{{AuthorizedUserID: "", Error: responseErr}}
						result, err := client.GetMetadata(ctx, id)
						errorsTest.ExpectEqual(err, responseErr)
						Expect(result).To(BeNil())
					})

					When("the user client ensure authorized service returns successfully", func() {
						BeforeEach(func() {
							authClient.EnsureAuthorizedUserOutputs = []authTest.EnsureAuthorizedUserOutput{{AuthorizedUserID: *responseResult.UserID, Error: nil}}
						})

						When("the image has metadata", func() {
							BeforeEach(func() {
								responseResult.Metadata = imageTest.RandomMetadata()
							})

							It("returns successfully", func() {
								result, err := client.GetMetadata(ctx, id)
								Expect(err).ToNot(HaveOccurred())
								Expect(result).To(Equal(responseResult.Metadata))
							})
						})

						When("the image does not have metadata", func() {
							BeforeEach(func() {
								responseResult.Metadata = nil
							})

							It("returns successfully", func() {
								result, err := client.GetMetadata(ctx, id)
								Expect(err).ToNot(HaveOccurred())
								Expect(result).To(Equal(image.NewMetadata()))
							})
						})
					})
				})
			})

			Context("GetContent", func() {
				var mediaType *string

				mediaTypeAssertions := func() {
					AfterEach(func() {
						Expect(imageStructuredSession.GetInputs).To(Equal([]imageStoreStructuredTest.GetInput{{ID: id, Condition: nil}}))
					})

					It("returns an error when the image structured session get returns an error", func() {
						responseErr := errorsTest.RandomError()
						imageStructuredSession.GetOutputs = []imageStoreStructuredTest.GetOutput{{Image: nil, Error: responseErr}}
						result, err := client.GetContent(ctx, id, mediaType)
						errorsTest.ExpectEqual(err, responseErr)
						Expect(result).To(BeNil())
					})

					It("returns successfully when the image structured session get returns nil", func() {
						imageStructuredSession.GetOutputs = []imageStoreStructuredTest.GetOutput{{Image: nil, Error: nil}}
						deleted, err := client.GetContent(ctx, id, mediaType)
						Expect(err).ToNot(HaveOccurred())
						Expect(deleted).To(BeNil())
					})

					When("the image structure session get returns an image", func() {
						var responseResult *image.Image

						BeforeEach(func() {
							responseResult = imageTest.RandomImage()
							responseResult.ID = pointer.FromString(id)
							responseResult.Status = pointer.FromString(image.StatusAvailable)
							responseResult.ContentID = pointer.FromString(imageTest.RandomContentID())
							responseResult.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
							responseResult.ContentAttributes = imageTest.RandomContentAttributes()
							if mediaType != nil {
								responseResult.ContentAttributes.MediaType = pointer.FromString(*mediaType)
							} else {
								responseResult.ContentAttributes.MediaType = pointer.FromString(imageTest.RandomMediaType())
							}
							imageStructuredSession.GetOutputs = []imageStoreStructuredTest.GetOutput{{Image: responseResult, Error: nil}}
						})

						AfterEach(func() {
							Expect(authClient.EnsureAuthorizedUserInputs).To(Equal([]authTest.EnsureAuthorizedUserInput{{TargetUserID: *responseResult.UserID, AuthorizedPermission: permission.Read}}))
						})

						It("returns an error when the user client ensure authorized service returns an error", func() {
							responseErr := errorsTest.RandomError()
							authClient.EnsureAuthorizedUserOutputs = []authTest.EnsureAuthorizedUserOutput{{AuthorizedUserID: "", Error: responseErr}}
							result, err := client.GetContent(ctx, id, mediaType)
							errorsTest.ExpectEqual(err, responseErr)
							Expect(result).To(BeNil())
						})

						When("the user client ensure authorized service returns successfully", func() {
							BeforeEach(func() {
								authClient.EnsureAuthorizedUserOutputs = []authTest.EnsureAuthorizedUserOutput{{AuthorizedUserID: *responseResult.UserID, Error: nil}}
							})

							It("returns nil when the image has no content", func() {
								responseResult.Status = pointer.FromString(image.StatusCreated)
								responseResult.ContentIntent = nil
								responseResult.ContentAttributes = nil
								result, err := client.GetContent(ctx, id, mediaType)
								Expect(err).ToNot(HaveOccurred())
								Expect(result).To(BeNil())
							})

							It("returns nil when the image does not have matching media type", func() {
								if mediaType != nil {
									switch *mediaType {
									case image.MediaTypeImageJPEG:
										responseResult.ContentAttributes.MediaType = pointer.FromString(image.MediaTypeImagePNG)
									case image.MediaTypeImagePNG:
										responseResult.ContentAttributes.MediaType = pointer.FromString(image.MediaTypeImageJPEG)
									}
								} else {
									responseResult.Status = pointer.FromString(image.StatusCreated)
									responseResult.ContentIntent = nil
									responseResult.ContentAttributes = nil
								}
								result, err := client.GetContent(ctx, id, mediaType)
								Expect(err).ToNot(HaveOccurred())
								Expect(result).To(BeNil())
							})

							When("the image store unstructured get content is invoked", func() {
								AfterEach(func() {
									Expect(imageUnstructuredStore.GetContentInputs).To(Equal([]imageStoreUnstructuredTest.GetContentInput{{UserID: *responseResult.UserID, ImageID: id, ContentID: *responseResult.ContentID, ContentIntent: *responseResult.ContentIntent}}))
								})

								It("returns an error when the image unstructured store get content returns an error", func() {
									responseErr := errorsTest.RandomError()
									imageUnstructuredStore.GetContentOutputs = []imageStoreUnstructuredTest.GetContentOutput{{Reader: nil, Error: responseErr}}
									result, err := client.GetContent(ctx, id, mediaType)
									errorsTest.ExpectEqual(err, responseErr)
									Expect(result).To(BeNil())
								})

								It("returns successfully when the image unstructured store get content returns successfully", func() {
									body := imageTest.RandomContentBytes()
									reader := ioutil.NopCloser(bytes.NewReader(body))
									imageUnstructuredStore.GetContentOutputs = []imageStoreUnstructuredTest.GetContentOutput{{Reader: reader, Error: nil}}
									result, err := client.GetContent(ctx, id, mediaType)
									Expect(err).ToNot(HaveOccurred())
									Expect(result).To(Equal(&image.Content{
										Body:      reader,
										DigestMD5: responseResult.ContentAttributes.DigestMD5,
										MediaType: responseResult.ContentAttributes.MediaType,
									}))
								})
							})
						})
					})
				}

				When("the media type is missing", func() {
					BeforeEach(func() {
						mediaType = nil
					})

					mediaTypeAssertions()
				})

				When("the media type is image/jpeg", func() {
					BeforeEach(func() {
						mediaType = pointer.FromString(image.MediaTypeImageJPEG)
					})

					mediaTypeAssertions()
				})

				When("the media type is image/png", func() {
					BeforeEach(func() {
						mediaType = pointer.FromString(image.MediaTypeImagePNG)
					})

					mediaTypeAssertions()
				})
			})

			Context("GetRenditionContent", func() {
				var rendition *image.Rendition

				BeforeEach(func() {
					rendition = imageTest.RandomRendition()
				})

				AfterEach(func() {
					Expect(imageStructuredSession.GetInputs).To(Equal([]imageStoreStructuredTest.GetInput{{ID: id, Condition: nil}}))
				})

				It("returns an error when the image structured session get returns an error", func() {
					responseErr := errorsTest.RandomError()
					imageStructuredSession.GetOutputs = []imageStoreStructuredTest.GetOutput{{Image: nil, Error: responseErr}}
					result, err := client.GetRenditionContent(ctx, id, rendition)
					errorsTest.ExpectEqual(err, responseErr)
					Expect(result).To(BeNil())
				})

				It("returns successfully when the image structured session get returns nil", func() {
					imageStructuredSession.GetOutputs = []imageStoreStructuredTest.GetOutput{{Image: nil, Error: nil}}
					deleted, err := client.GetRenditionContent(ctx, id, rendition)
					Expect(err).ToNot(HaveOccurred())
					Expect(deleted).To(BeNil())
				})

				When("the image structure session get returns an image", func() {
					var responseResult *image.Image

					BeforeEach(func() {
						responseResult = imageTest.RandomImage()
						responseResult.ID = pointer.FromString(id)
						responseResult.Status = pointer.FromString(image.StatusAvailable)
						responseResult.ContentID = pointer.FromString(imageTest.RandomContentID())
						responseResult.ContentIntent = pointer.FromString(image.ContentIntentOriginal)
						responseResult.ContentAttributes = imageTest.RandomContentAttributes()
						imageStructuredSession.GetOutputs = []imageStoreStructuredTest.GetOutput{{Image: responseResult, Error: nil}}
					})

					AfterEach(func() {
						Expect(authClient.EnsureAuthorizedUserInputs).To(Equal([]authTest.EnsureAuthorizedUserInput{{TargetUserID: *responseResult.UserID, AuthorizedPermission: permission.Read}}))
					})

					It("returns an error when the user client ensure authorized service returns an error", func() {
						responseErr := errorsTest.RandomError()
						authClient.EnsureAuthorizedUserOutputs = []authTest.EnsureAuthorizedUserOutput{{AuthorizedUserID: "", Error: responseErr}}
						result, err := client.GetRenditionContent(ctx, id, rendition)
						errorsTest.ExpectEqual(err, responseErr)
						Expect(result).To(BeNil())
					})

					When("the user client ensure authorized service returns successfully", func() {
						BeforeEach(func() {
							authClient.EnsureAuthorizedUserOutputs = []authTest.EnsureAuthorizedUserOutput{{AuthorizedUserID: *responseResult.UserID, Error: nil}}
						})

						It("returns nil when the image has no content", func() {
							responseResult.Status = pointer.FromString(image.StatusCreated)
							responseResult.ContentIntent = nil
							responseResult.ContentAttributes = nil
							result, err := client.GetRenditionContent(ctx, id, rendition)
							Expect(err).ToNot(HaveOccurred())
							Expect(result).To(BeNil())
						})

						When("the the transform is calculated", func() {
							AfterEach(func() {
								Expect(imageTransformer.CalculateTransformInputs).To(Equal([]imageTransformTest.CalculateTransformInput{{ContentAttributes: responseResult.ContentAttributes, Rendition: rendition}}))
							})

							It("returns an error when the image transform calculate transform returns an error", func() {
								responseErr := errorsTest.RandomError()
								imageTransformer.CalculateTransformOutputs = []imageTransformTest.CalculateTransformOutput{{Transform: nil, Error: responseErr}}
								result, err := client.GetRenditionContent(ctx, id, rendition)
								errorsTest.ExpectEqual(err, responseErr)
								Expect(result).To(BeNil())
							})

							When("a transform is returned", func() {
								var transform *imageTransform.Transform

								BeforeEach(func() {
									transform = imageTransformTest.RandomTransform()
									imageTransformer.CalculateTransformOutputs = []imageTransformTest.CalculateTransformOutput{{Transform: transform, Error: nil}}
								})

								When("the image does not have the transformed rendition", func() {
									AfterEach(func() {
										Expect(imageUnstructuredStore.GetContentInputs).To(Equal([]imageStoreUnstructuredTest.GetContentInput{{UserID: *responseResult.UserID, ImageID: id, ContentID: *responseResult.ContentID, ContentIntent: *responseResult.ContentIntent}}))
									})

									It("returns an error when the image unstructured store get content returns an error", func() {
										responseErr := errorsTest.RandomError()
										imageUnstructuredStore.GetContentOutputs = []imageStoreUnstructuredTest.GetContentOutput{{Reader: nil, Error: responseErr}}
										result, err := client.GetRenditionContent(ctx, id, rendition)
										errorsTest.ExpectEqual(err, responseErr)
										Expect(result).To(BeNil())
									})

									When("the image unstructured store get content return successfully", func() {
										var contentReader io.ReadCloser

										BeforeEach(func() {
											contentReader = ioutil.NopCloser(bytes.NewReader(imageTest.RandomContentBytes()))
											imageUnstructuredStore.GetContentOutputs = []imageStoreUnstructuredTest.GetContentOutput{{Reader: contentReader, Error: nil}}
										})

										AfterEach(func() {
											Expect(imageTransformer.TransformContentInputs).To(Equal([]imageTransformTest.TransformContentInput{{Reader: contentReader, Transform: transform}}))
										})

										It("returns an error when the image transform transform content returns an error", func() {
											responseErr := errorsTest.RandomError()
											imageTransformer.TransformContentOutputs = []imageTransformTest.TransformContentOutput{{Reader: nil, Error: responseErr}}
											result, err := client.GetRenditionContent(ctx, id, rendition)
											errorsTest.ExpectEqual(err, errors.Wrap(responseErr, "unable to transform image content"))
											Expect(result).To(BeNil())
										})

										When("the the image transform transform content returns successfully", func() {
											var transformRenditionReader io.ReadCloser
											var putRenditionContentRenditionsID string
											var putRenditionContentError error

											BeforeEach(func() {
												transformRenditionReader = ioutil.NopCloser(bytes.NewReader(imageTest.RandomContentBytes()))
												imageTransformer.TransformContentOutputs = []imageTransformTest.TransformContentOutput{{Reader: transformRenditionReader, Error: nil}}
												putRenditionContentRenditionsID = ""
												putRenditionContentError = nil
												imageUnstructuredStore.PutRenditionContentStub = func(ctx context.Context, userID string, imageID string, contentID string, renditionsID string, rendition string, reader io.Reader, options *storeUnstructured.Options) error {
													putRenditionContentRenditionsID = renditionsID
													return putRenditionContentError
												}
											})

											AfterEach(func() {
												options := storeUnstructured.NewOptions()
												options.MediaType = transform.Rendition.MediaType
												Expect(imageUnstructuredStore.PutRenditionContentInputs).To(Equal([]imageStoreUnstructuredTest.PutRenditionContentInput{{UserID: *responseResult.UserID, ImageID: id, ContentID: *responseResult.ContentID, RenditionsID: putRenditionContentRenditionsID, Rendition: transform.Rendition.String(), Reader: transformRenditionReader, Options: options}}))
											})

											It("returns an error when the image unstructured store put rendition content returns an error", func() {
												putRenditionContentError = errorsTest.RandomError()
												result, err := client.GetRenditionContent(ctx, id, rendition)
												errorsTest.ExpectEqual(err, putRenditionContentError)
												Expect(result).To(BeNil())
											})

											When("the image unstructured store put rendition content returns successfully", func() {
												AfterEach(func() {
													if responseResult.RenditionsID != nil && *responseResult.RenditionsID == putRenditionContentRenditionsID {
														Expect(imageStructuredSession.UpdateInputs).To(Equal([]imageStoreStructuredTest.UpdateInput{{ID: id, Condition: &request.Condition{Revision: responseResult.Revision}, Update: &imageStoreStructured.Update{Rendition: pointer.FromString(transform.Rendition.String())}}}))
													} else {
														Expect(imageStructuredSession.UpdateInputs).To(Equal([]imageStoreStructuredTest.UpdateInput{{ID: id, Condition: &request.Condition{Revision: responseResult.Revision}, Update: &imageStoreStructured.Update{RenditionsID: pointer.FromString(putRenditionContentRenditionsID), Rendition: pointer.FromString(transform.Rendition.String())}}}))
													}
												})

												When("the image structured store update returns successfully", func() {
													BeforeEach(func() {
														imageStructuredSession.UpdateOutputs = []imageStoreStructuredTest.UpdateOutput{{Image: imageTest.RandomImage(), Error: nil}}
													})

													When("an old renditions id did not exist", func() {
														BeforeEach(func() {
															responseResult.RenditionsID = nil
															responseResult.Renditions = nil
														})

														AfterEach(func() {
															Expect(imageUnstructuredStore.GetRenditionContentInputs).To(Equal([]imageStoreUnstructuredTest.GetRenditionContentInput{{UserID: *responseResult.UserID, ImageID: id, ContentID: *responseResult.ContentID, RenditionsID: putRenditionContentRenditionsID, Rendition: transform.Rendition.String()}}))
														})

														It("returns an error when the image unstructured store get rendition content returns an error", func() {
															responseErr := errorsTest.RandomError()
															imageUnstructuredStore.GetRenditionContentOutputs = append(imageUnstructuredStore.GetRenditionContentOutputs, imageStoreUnstructuredTest.GetRenditionContentOutput{Reader: nil, Error: responseErr})
															result, err := client.GetRenditionContent(ctx, id, rendition)
															errorsTest.ExpectEqual(err, responseErr)
															Expect(result).To(BeNil())
														})

														It("returns successfully when the image unstructured store get rendition content returns successfully", func() {
															contentRenditionReader := ioutil.NopCloser(bytes.NewReader(imageTest.RandomContentBytes()))
															imageUnstructuredStore.GetRenditionContentOutputs = append(imageUnstructuredStore.GetRenditionContentOutputs, imageStoreUnstructuredTest.GetRenditionContentOutput{Reader: contentRenditionReader, Error: nil})
															result, err := client.GetRenditionContent(ctx, id, rendition)
															Expect(err).ToNot(HaveOccurred())
															Expect(result).To(Equal(&image.Content{
																Body:      contentRenditionReader,
																MediaType: rendition.MediaType,
															}))
														})

														It("returns successfully when the image unstructured store get rendition content returns successfully and logs an error if the image structured store update returns an error", func() {
															responseErr := errorsTest.RandomError()
															imageStructuredSession.UpdateOutputs = []imageStoreStructuredTest.UpdateOutput{{Image: nil, Error: responseErr}}
															contentRenditionReader := ioutil.NopCloser(bytes.NewReader(imageTest.RandomContentBytes()))
															imageUnstructuredStore.GetRenditionContentOutputs = append(imageUnstructuredStore.GetRenditionContentOutputs, imageStoreUnstructuredTest.GetRenditionContentOutput{Reader: contentRenditionReader, Error: nil})
															result, err := client.GetRenditionContent(ctx, id, rendition)
															Expect(err).ToNot(HaveOccurred())
															Expect(result).To(Equal(&image.Content{
																Body:      contentRenditionReader,
																MediaType: rendition.MediaType,
															}))
															logger.AssertError("Unable to update image with rendition; orphaned rendition", log.Fields{"id": *responseResult.ID, "error": errors.NewSerializable(responseErr)})
														})
													})

													When("an old renditions id was not replaced", func() {
														BeforeEach(func() {
															renditions := make([]string, image.RenditionsLengthMaximum-1)
															for index := range renditions {
																renditions[index] = imageTest.RandomRenditionString()
															}
															responseResult.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
															responseResult.Renditions = pointer.FromStringArray(renditions)
														})

														AfterEach(func() {
															Expect(imageUnstructuredStore.GetRenditionContentInputs).To(Equal([]imageStoreUnstructuredTest.GetRenditionContentInput{{UserID: *responseResult.UserID, ImageID: id, ContentID: *responseResult.ContentID, RenditionsID: putRenditionContentRenditionsID, Rendition: transform.Rendition.String()}}))
														})

														It("returns an error when the image unstructured store get rendition content returns an error", func() {
															responseErr := errorsTest.RandomError()
															imageUnstructuredStore.GetRenditionContentOutputs = append(imageUnstructuredStore.GetRenditionContentOutputs, imageStoreUnstructuredTest.GetRenditionContentOutput{Reader: nil, Error: responseErr})
															result, err := client.GetRenditionContent(ctx, id, rendition)
															errorsTest.ExpectEqual(err, responseErr)
															Expect(result).To(BeNil())
														})

														It("returns successfully when the image unstructured store get rendition content returns successfully", func() {
															contentRenditionReader := ioutil.NopCloser(bytes.NewReader(imageTest.RandomContentBytes()))
															imageUnstructuredStore.GetRenditionContentOutputs = append(imageUnstructuredStore.GetRenditionContentOutputs, imageStoreUnstructuredTest.GetRenditionContentOutput{Reader: contentRenditionReader, Error: nil})
															result, err := client.GetRenditionContent(ctx, id, rendition)
															Expect(err).ToNot(HaveOccurred())
															Expect(result).To(Equal(&image.Content{
																Body:      contentRenditionReader,
																MediaType: rendition.MediaType,
															}))
														})

														It("returns successfully when the image unstructured store get rendition content returns successfully and logs an error if the image structured store update returns an error", func() {
															responseErr := errorsTest.RandomError()
															imageStructuredSession.UpdateOutputs = []imageStoreStructuredTest.UpdateOutput{{Image: nil, Error: responseErr}}
															contentRenditionReader := ioutil.NopCloser(bytes.NewReader(imageTest.RandomContentBytes()))
															imageUnstructuredStore.GetRenditionContentOutputs = append(imageUnstructuredStore.GetRenditionContentOutputs, imageStoreUnstructuredTest.GetRenditionContentOutput{Reader: contentRenditionReader, Error: nil})
															result, err := client.GetRenditionContent(ctx, id, rendition)
															Expect(err).ToNot(HaveOccurred())
															Expect(result).To(Equal(&image.Content{
																Body:      contentRenditionReader,
																MediaType: rendition.MediaType,
															}))
															logger.AssertError("Unable to update image with rendition; orphaned rendition", log.Fields{"id": *responseResult.ID, "error": errors.NewSerializable(responseErr)})
														})
													})

													When("an old renditions id was replaced", func() {
														BeforeEach(func() {
															renditions := make([]string, image.RenditionsLengthMaximum)
															for index := range renditions {
																renditions[index] = imageTest.RandomRenditionString()
															}
															responseResult.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
															responseResult.Renditions = pointer.FromStringArray(renditions)
														})

														When("the image unstructured store delete rendition content is invoked", func() {
															BeforeEach(func() {
																imageUnstructuredStore.DeleteRenditionContentOutputs = []error{nil}
															})

															AfterEach(func() {
																Expect(imageUnstructuredStore.DeleteRenditionContentInputs).To(Equal([]imageStoreUnstructuredTest.DeleteRenditionContentInput{{UserID: *responseResult.UserID, ImageID: id, ContentID: *responseResult.ContentID, RenditionsID: *responseResult.RenditionsID}}))
																Expect(imageUnstructuredStore.GetRenditionContentInputs).To(Equal([]imageStoreUnstructuredTest.GetRenditionContentInput{{UserID: *responseResult.UserID, ImageID: id, ContentID: *responseResult.ContentID, RenditionsID: putRenditionContentRenditionsID, Rendition: transform.Rendition.String()}}))
															})

															It("returns an error when the image unstructured store get rendition content returns an error", func() {
																responseErr := errorsTest.RandomError()
																imageUnstructuredStore.GetRenditionContentOutputs = append(imageUnstructuredStore.GetRenditionContentOutputs, imageStoreUnstructuredTest.GetRenditionContentOutput{Reader: nil, Error: responseErr})
																result, err := client.GetRenditionContent(ctx, id, rendition)
																errorsTest.ExpectEqual(err, responseErr)
																Expect(result).To(BeNil())
																logger.AssertError("Deleting excess image rendition content", log.Fields{"id": *responseResult.ID})
															})

															It("returns successfully when the image unstructured store get rendition content returns successfully and logs an error if the delete rendition content return an error", func() {
																responseErr := errorsTest.RandomError()
																imageUnstructuredStore.DeleteRenditionContentOutputs = []error{responseErr}
																contentRenditionReader := ioutil.NopCloser(bytes.NewReader(imageTest.RandomContentBytes()))
																imageUnstructuredStore.GetRenditionContentOutputs = append(imageUnstructuredStore.GetRenditionContentOutputs, imageStoreUnstructuredTest.GetRenditionContentOutput{Reader: contentRenditionReader, Error: nil})
																result, err := client.GetRenditionContent(ctx, id, rendition)
																Expect(err).ToNot(HaveOccurred())
																Expect(result).To(Equal(&image.Content{
																	Body:      contentRenditionReader,
																	MediaType: rendition.MediaType,
																}))
																logger.AssertError("Deleting excess image rendition content", log.Fields{"id": *responseResult.ID})
																logger.AssertError("Unable to delete excess image rendition content", log.Fields{"id": *responseResult.ID, "error": errors.NewSerializable(responseErr)})
															})

															It("returns successfully when the image unstructured store get rendition content returns successfully", func() {
																contentRenditionReader := ioutil.NopCloser(bytes.NewReader(imageTest.RandomContentBytes()))
																imageUnstructuredStore.GetRenditionContentOutputs = append(imageUnstructuredStore.GetRenditionContentOutputs, imageStoreUnstructuredTest.GetRenditionContentOutput{Reader: contentRenditionReader, Error: nil})
																result, err := client.GetRenditionContent(ctx, id, rendition)
																Expect(err).ToNot(HaveOccurred())
																Expect(result).To(Equal(&image.Content{
																	Body:      contentRenditionReader,
																	MediaType: rendition.MediaType,
																}))
																logger.AssertError("Deleting excess image rendition content", log.Fields{"id": *responseResult.ID})
															})
														})

														It("returns successfully when the image unstructured store get rendition content returns successfully and logs an error if the image structured store update returns an error", func() {
															responseErr := errorsTest.RandomError()
															imageStructuredSession.UpdateOutputs = []imageStoreStructuredTest.UpdateOutput{{Image: nil, Error: responseErr}}
															contentRenditionReader := ioutil.NopCloser(bytes.NewReader(imageTest.RandomContentBytes()))
															imageUnstructuredStore.GetRenditionContentOutputs = append(imageUnstructuredStore.GetRenditionContentOutputs, imageStoreUnstructuredTest.GetRenditionContentOutput{Reader: contentRenditionReader, Error: nil})
															result, err := client.GetRenditionContent(ctx, id, rendition)
															Expect(err).ToNot(HaveOccurred())
															Expect(result).To(Equal(&image.Content{
																Body:      contentRenditionReader,
																MediaType: rendition.MediaType,
															}))
															logger.AssertError("Unable to update image with rendition; orphaned rendition", log.Fields{"id": *responseResult.ID, "error": errors.NewSerializable(responseErr)})
														})
													})
												})
											})
										})
									})
								})

								When("the image has the transformed rendition", func() {
									BeforeEach(func() {
										responseResult.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
										responseResult.Renditions = pointer.FromStringArray(append(imageTest.RandomRenditionStrings(), transform.Rendition.String()))
									})

									AfterEach(func() {
										Expect(imageUnstructuredStore.GetRenditionContentInputs).To(Equal([]imageStoreUnstructuredTest.GetRenditionContentInput{{UserID: *responseResult.UserID, ImageID: id, ContentID: *responseResult.ContentID, RenditionsID: *responseResult.RenditionsID, Rendition: transform.Rendition.String()}}))
									})

									It("returns an error when the image unstructured store get rendition content returns an error", func() {
										responseErr := errorsTest.RandomError()
										imageUnstructuredStore.GetRenditionContentOutputs = []imageStoreUnstructuredTest.GetRenditionContentOutput{{Reader: nil, Error: responseErr}}
										result, err := client.GetRenditionContent(ctx, id, rendition)
										errorsTest.ExpectEqual(err, responseErr)
										Expect(result).To(BeNil())
									})

									It("returns successfully when the image unstructured store get rendition content returns successfully", func() {
										contentRenditionReader := ioutil.NopCloser(bytes.NewReader(imageTest.RandomContentBytes()))
										imageUnstructuredStore.GetRenditionContentOutputs = []imageStoreUnstructuredTest.GetRenditionContentOutput{{Reader: contentRenditionReader, Error: nil}}
										result, err := client.GetRenditionContent(ctx, id, rendition)
										Expect(err).ToNot(HaveOccurred())
										Expect(result).To(Equal(&image.Content{
											Body:      contentRenditionReader,
											MediaType: rendition.MediaType,
										}))
									})
								})
							})
						})
					})
				})
			})

			Context("PutMetadata", func() {
				var condition *request.Condition
				var metadata *image.Metadata

				BeforeEach(func() {
					condition = requestTest.RandomCondition()
					metadata = imageTest.RandomMetadata()
				})

				AfterEach(func() {
					Expect(imageStructuredSession.GetInputs).To(Equal([]imageStoreStructuredTest.GetInput{{ID: id, Condition: condition}}))
				})

				It("returns an error when the image structured session get returns an error", func() {
					responseErr := errorsTest.RandomError()
					imageStructuredSession.GetOutputs = []imageStoreStructuredTest.GetOutput{{Image: nil, Error: responseErr}}
					result, err := client.PutMetadata(ctx, id, condition, metadata)
					errorsTest.ExpectEqual(err, responseErr)
					Expect(result).To(BeNil())
				})

				It("returns successfully when the image structured session get returns nil", func() {
					imageStructuredSession.GetOutputs = []imageStoreStructuredTest.GetOutput{{Image: nil, Error: nil}}
					result, err := client.PutMetadata(ctx, id, condition, metadata)
					Expect(err).ToNot(HaveOccurred())
					Expect(result).To(BeNil())
				})

				When("the image structure session get returns an image", func() {
					var responseResult *image.Image

					BeforeEach(func() {
						responseResult = imageTest.RandomImage()
						responseResult.ID = pointer.FromString(id)
						imageStructuredSession.GetOutputs = []imageStoreStructuredTest.GetOutput{{Image: responseResult, Error: nil}}
					})

					AfterEach(func() {
						Expect(authClient.EnsureAuthorizedUserInputs).To(Equal([]authTest.EnsureAuthorizedUserInput{{TargetUserID: *responseResult.UserID, AuthorizedPermission: permission.Write}}))
					})

					It("returns an error when the user client ensure authorized service returns an error", func() {
						responseErr := errorsTest.RandomError()
						authClient.EnsureAuthorizedUserOutputs = []authTest.EnsureAuthorizedUserOutput{{AuthorizedUserID: "", Error: responseErr}}
						result, err := client.PutMetadata(ctx, id, condition, metadata)
						errorsTest.ExpectEqual(err, responseErr)
						Expect(result).To(BeNil())
					})

					When("the user client ensure authorized service returns successfully", func() {
						BeforeEach(func() {
							authClient.EnsureAuthorizedUserOutputs = []authTest.EnsureAuthorizedUserOutput{{AuthorizedUserID: *responseResult.UserID, Error: nil}}
						})

						It("returns an error when the metadata is missing", func() {
							metadata = nil
							result, err := client.PutMetadata(ctx, id, condition, metadata)
							errorsTest.ExpectEqual(err, errors.New("metadata is missing"))
							Expect(result).To(BeNil())
						})

						It("returns an error when the metadata is invalid", func() {
							metadata.Name = pointer.FromString("")
							result, err := client.PutMetadata(ctx, id, condition, metadata)
							errorsTest.ExpectEqual(err, errors.New("metadata is invalid"))
							Expect(result).To(BeNil())
						})

						When("the image is updated", func() {
							AfterEach(func() {
								update := imageStoreStructured.NewUpdate()
								update.Metadata = metadata
								Expect(imageStructuredSession.UpdateInputs).To(Equal([]imageStoreStructuredTest.UpdateInput{{ID: id, Condition: condition, Update: update}}))
							})

							It("returns an error when the image structured session create returns an error", func() {
								responseErr := errorsTest.RandomError()
								imageStructuredSession.UpdateOutputs = []imageStoreStructuredTest.UpdateOutput{{Image: nil, Error: responseErr}}
								result, err := client.PutMetadata(ctx, id, condition, metadata)
								errorsTest.ExpectEqual(err, responseErr)
								Expect(result).To(BeNil())
							})

							It("returns successfully when the image structured session create returns successfully", func() {
								updateImage := imageTest.CloneImage(responseResult)
								updateImage.ModifiedTime = pointer.FromTime(test.RandomTimeFromRange(*responseResult.CreatedTime, time.Now()).Truncate(time.Second))
								imageStructuredSession.UpdateOutputs = []imageStoreStructuredTest.UpdateOutput{{Image: updateImage, Error: nil}}
								result, err := client.PutMetadata(ctx, id, condition, metadata)
								Expect(err).ToNot(HaveOccurred())
								Expect(result).To(Equal(updateImage))
							})
						})
					})
				})
			})

			Context("PutContent", func() {
				var condition *request.Condition
				var contentIntent string
				var width int
				var height int
				var content *image.Content

				BeforeEach(func() {
					condition = requestTest.RandomCondition()
					contentIntent = imageTest.RandomContentIntent()
					width = test.RandomIntFromRange(10, 20)
					height = test.RandomIntFromRange(10, 20)
					content = imageTest.RandomContentFromDimensions(width, height)
				})

				AfterEach(func() {
					Expect(imageStructuredSession.GetInputs).To(Equal([]imageStoreStructuredTest.GetInput{{ID: id, Condition: condition}}))
				})

				It("returns an error when the image structured session get returns an error", func() {
					responseErr := errorsTest.RandomError()
					imageStructuredSession.GetOutputs = []imageStoreStructuredTest.GetOutput{{Image: nil, Error: responseErr}}
					result, err := client.PutContent(ctx, id, condition, contentIntent, content)
					errorsTest.ExpectEqual(err, responseErr)
					Expect(result).To(BeNil())
				})

				It("returns successfully when the image structured session get returns nil", func() {
					imageStructuredSession.GetOutputs = []imageStoreStructuredTest.GetOutput{{Image: nil, Error: nil}}
					result, err := client.PutContent(ctx, id, condition, contentIntent, content)
					Expect(err).ToNot(HaveOccurred())
					Expect(result).To(BeNil())
				})

				When("the image structure session get returns an image", func() {
					var responseResult *image.Image

					JustBeforeEach(func() {
						responseResult = imageTest.RandomImage()
						responseResult.ID = pointer.FromString(id)
						responseResult.Status = pointer.FromString(image.StatusAvailable)
						switch contentIntent {
						case image.ContentIntentAlternate:
							responseResult.ContentID = nil
							responseResult.ContentIntent = nil
							responseResult.ContentAttributes = nil
							responseResult.RenditionsID = nil
							responseResult.Renditions = nil
						case image.ContentIntentOriginal:
							responseResult.ContentID = pointer.FromString(imageTest.RandomContentID())
							responseResult.ContentIntent = pointer.FromString(image.ContentIntentAlternate)
							responseResult.ContentAttributes = imageTest.RandomContentAttributes()
							responseResult.RenditionsID = pointer.FromString(imageTest.RandomRenditionsID())
							responseResult.Renditions = pointer.FromStringArray(imageTest.RandomRenditionStrings())
						}
						imageStructuredSession.GetOutputs = []imageStoreStructuredTest.GetOutput{{Image: responseResult, Error: nil}}
					})

					AfterEach(func() {
						Expect(authClient.EnsureAuthorizedUserInputs).To(Equal([]authTest.EnsureAuthorizedUserInput{{TargetUserID: *responseResult.UserID, AuthorizedPermission: permission.Write}}))
					})

					It("returns an error when the user client ensure authorized service returns an error", func() {
						responseErr := errorsTest.RandomError()
						authClient.EnsureAuthorizedUserOutputs = []authTest.EnsureAuthorizedUserOutput{{AuthorizedUserID: "", Error: responseErr}}
						result, err := client.PutContent(ctx, id, condition, contentIntent, content)
						errorsTest.ExpectEqual(err, responseErr)
						Expect(result).To(BeNil())
					})

					When("the user client ensure authorized service returns successfully", func() {
						JustBeforeEach(func() {
							authClient.EnsureAuthorizedUserOutputs = []authTest.EnsureAuthorizedUserOutput{{AuthorizedUserID: *responseResult.UserID, Error: nil}}
						})

						It("returns an error when content intent is missing", func() {
							contentIntent = ""
							result, err := client.PutContent(ctx, id, condition, contentIntent, content)
							errorsTest.ExpectEqual(err, errors.New("content intent is missing"))
							Expect(result).To(BeNil())
						})

						It("returns an error when content intent is invalid", func() {
							contentIntent = "invalid"
							result, err := client.PutContent(ctx, id, condition, contentIntent, content)
							errorsTest.ExpectEqual(err, errors.New("content intent is invalid"))
							Expect(result).To(BeNil())
						})

						It("returns an error when current content intent is alternate and new content intent is alternate", func() {
							responseResult.Status = pointer.FromString(image.StatusAvailable)
							responseResult.ContentID = pointer.FromString(imageTest.RandomContentID())
							responseResult.ContentIntent = pointer.FromString(image.ContentIntentAlternate)
							responseResult.ContentAttributes = imageTest.RandomContentAttributes()
							contentIntent = image.ContentIntentAlternate
							result, err := client.PutContent(ctx, id, condition, contentIntent, content)
							errorsTest.ExpectEqual(err, image.ErrorImageContentIntentUnexpected(contentIntent))
							Expect(result).To(BeNil())
						})

						It("returns an error when current content intent is original and new content intent is alternate", func() {
							responseResult.Status = pointer.FromString(image.StatusAvailable)
							responseResult.ContentID = pointer.FromString(imageTest.RandomContentID())
							responseResult.ContentIntent = pointer.FromString(image.ContentIntentOriginal)
							responseResult.ContentAttributes = imageTest.RandomContentAttributes()
							contentIntent = image.ContentIntentAlternate
							result, err := client.PutContent(ctx, id, condition, contentIntent, content)
							errorsTest.ExpectEqual(err, image.ErrorImageContentIntentUnexpected(contentIntent))
							Expect(result).To(BeNil())
						})

						It("returns an error when current content intent is original and new content intent is original", func() {
							responseResult.Status = pointer.FromString(image.StatusAvailable)
							responseResult.ContentID = pointer.FromString(imageTest.RandomContentID())
							responseResult.ContentIntent = pointer.FromString(image.ContentIntentOriginal)
							responseResult.ContentAttributes = imageTest.RandomContentAttributes()
							contentIntent = image.ContentIntentOriginal
							result, err := client.PutContent(ctx, id, condition, contentIntent, content)
							errorsTest.ExpectEqual(err, image.ErrorImageContentIntentUnexpected(contentIntent))
							Expect(result).To(BeNil())
						})

						It("returns an error when content is missing", func() {
							content = nil
							result, err := client.PutContent(ctx, id, condition, contentIntent, content)
							errorsTest.ExpectEqual(err, errors.New("content is missing"))
							Expect(result).To(BeNil())
						})

						It("returns an error when content is invalid", func() {
							content.Body = nil
							result, err := client.PutContent(ctx, id, condition, contentIntent, content)
							errorsTest.ExpectEqual(err, errors.New("content is invalid"))
							Expect(result).To(BeNil())
						})

						It("returns an error when content body is invalid", func() {
							content.Body = ioutil.NopCloser(bytes.NewReader(test.RandomBytes()))
							result, err := client.PutContent(ctx, id, condition, contentIntent, content)
							errorsTest.ExpectEqual(err, image.ErrorImageMalformed("unable to decode image; image: unknown format"))
							Expect(result).To(BeNil())
						})

						It("returns an error when content does not match media type", func() {
							switch *content.MediaType {
							case image.MediaTypeImageJPEG:
								content.MediaType = pointer.FromString(image.MediaTypeImagePNG)
							case image.MediaTypeImagePNG:
								content.MediaType = pointer.FromString(image.MediaTypeImageJPEG)
							}
							result, err := client.PutContent(ctx, id, condition, contentIntent, content)
							errorsTest.ExpectEqual(err, image.ErrorImageMalformed("header does not match media type"))
							Expect(result).To(BeNil())
						})

						When("the image unstructured store put content is invoked", func() {
							AfterEach(func() {
								Expect(imageUnstructuredStore.PutContentInputs).To(HaveLen(1))
								Expect(imageUnstructuredStore.PutContentInputs[0].UserID).To(Equal(*responseResult.UserID))
								Expect(imageUnstructuredStore.PutContentInputs[0].ImageID).To(Equal(*responseResult.ID))
								Expect(imageUnstructuredStore.PutContentInputs[0].ContentIntent).To(Equal(contentIntent))
								Expect(imageUnstructuredStore.PutContentInputs[0].Reader).ToNot(BeNil())
								Expect(imageUnstructuredStore.PutContentInputs[0].Options).To(Equal(&storeUnstructured.Options{MediaType: content.MediaType}))
							})

							It("returns an error when the image unstructured store put content returns an error", func() {
								responseErr := errorsTest.RandomError()
								imageUnstructuredStore.PutContentOutputs = []error{responseErr}
								result, err := client.PutContent(ctx, id, condition, contentIntent, content)
								errorsTest.ExpectEqual(err, responseErr)
								Expect(result).To(BeNil())
							})

							When("the image unstructured store put content returns successfully", func() {
								var putContentID string
								var size int64

								BeforeEach(func() {
									putContentID = ""
									imageUnstructuredStore.PutContentStub = func(ctx context.Context, userID string, imageID string, contentID string, contentIntent string, reader io.Reader, options *storeUnstructured.Options) error {
										putContentID = contentID
										var err error
										size, err = io.Copy(ioutil.Discard, reader)
										return err
									}
								})

								When("the size exceeds maximum", func() {
									BeforeEach(func() {
										contentBytes := make([]byte, 104857601)
										content.Body.Read(contentBytes)
										content.Body = ioutil.NopCloser(bytes.NewReader(contentBytes))
										content.DigestMD5 = pointer.FromString(crypto.Base64EncodedMD5Hash(contentBytes))
									})

									AfterEach(func() {
										Expect(imageUnstructuredStore.DeleteContentInputs).To(Equal([]imageStoreUnstructuredTest.DeleteContentInput{{UserID: *responseResult.UserID, ImageID: *responseResult.ID, ContentID: putContentID}}))
									})

									It("returns an error", func() {
										imageUnstructuredStore.DeleteContentOutputs = []error{nil}
										result, err := client.PutContent(ctx, id, condition, contentIntent, content)
										errorsTest.ExpectEqual(err, request.ErrorResourceTooLarge())
										Expect(result).To(BeNil())
										imageUnstructuredStore.DeleteContentInputs = []imageStoreUnstructuredTest.DeleteContentInput{{UserID: *responseResult.UserID, ImageID: *responseResult.ID, ContentID: putContentID}}
									})

									It("returns an error and logs an error when both the unstructured store delete content", func() {
										responseErr := errorsTest.RandomError()
										imageUnstructuredStore.DeleteContentOutputs = []error{responseErr}
										result, err := client.PutContent(ctx, id, condition, contentIntent, content)
										errorsTest.ExpectEqual(err, request.ErrorResourceTooLarge())
										Expect(result).To(BeNil())
										logger.AssertError("Unable to delete image content exceeding maximum size", log.Fields{"userId": *responseResult.UserID, "id": *responseResult.ID, "error": errors.NewSerializable(responseErr)})
										imageUnstructuredStore.DeleteContentInputs = []imageStoreUnstructuredTest.DeleteContentInput{{UserID: *responseResult.UserID, ImageID: *responseResult.ID, ContentID: putContentID}}
									})
								})

								When("the digest does not match", func() {
									var digestMD5 string

									BeforeEach(func() {
										digestMD5 = *content.DigestMD5
										content.DigestMD5 = pointer.FromString(cryptoTest.RandomBase64EncodedMD5Hash())
									})

									AfterEach(func() {
										Expect(imageUnstructuredStore.DeleteContentInputs).To(Equal([]imageStoreUnstructuredTest.DeleteContentInput{{UserID: *responseResult.UserID, ImageID: *responseResult.ID, ContentID: putContentID}}))
									})

									It("returns an error", func() {
										imageUnstructuredStore.DeleteContentOutputs = []error{nil}
										result, err := client.PutContent(ctx, id, condition, contentIntent, content)
										errorsTest.ExpectEqual(err, errorsTest.WithPointerSource(request.ErrorDigestsNotEqual(*content.DigestMD5, digestMD5), "/digestMD5"))
										Expect(result).To(BeNil())
										imageUnstructuredStore.DeleteContentInputs = []imageStoreUnstructuredTest.DeleteContentInput{{UserID: *responseResult.UserID, ImageID: *responseResult.ID, ContentID: putContentID}}
									})

									It("returns an error and logs an error when both the unstructured and structured store returns an error", func() {
										responseErr := errorsTest.RandomError()
										imageUnstructuredStore.DeleteContentOutputs = []error{responseErr}
										result, err := client.PutContent(ctx, id, condition, contentIntent, content)
										errorsTest.ExpectEqual(err, errorsTest.WithPointerSource(request.ErrorDigestsNotEqual(*content.DigestMD5, digestMD5), "/digestMD5"))
										Expect(result).To(BeNil())
										logger.AssertError("Unable to delete image content with incorrect MD5 digest", log.Fields{"userId": *responseResult.UserID, "id": *responseResult.ID, "error": errors.NewSerializable(responseErr)})
										imageUnstructuredStore.DeleteContentInputs = []imageStoreUnstructuredTest.DeleteContentInput{{UserID: *responseResult.UserID, ImageID: *responseResult.ID, ContentID: putContentID}}
									})
								})

								When("the size and digest are valid", func() {
									AfterEach(func() {
										update := imageStoreStructured.NewUpdate()
										update.ContentID = pointer.FromString(putContentID)
										update.ContentIntent = pointer.FromString(contentIntent)
										update.ContentAttributes = imageStoreStructured.NewContentAttributes()
										update.ContentAttributes.DigestMD5 = pointer.CloneString(content.DigestMD5)
										update.ContentAttributes.MediaType = pointer.CloneString(content.MediaType)
										update.ContentAttributes.Width = pointer.FromInt(width)
										update.ContentAttributes.Height = pointer.FromInt(height)
										update.ContentAttributes.Size = pointer.FromInt(int(size))
										Expect(imageStructuredSession.UpdateInputs).To(Equal([]imageStoreStructuredTest.UpdateInput{{ID: *responseResult.ID, Condition: &request.Condition{Revision: responseResult.Revision}, Update: update}}))
									})

									It("returns an error when image structured session update returns an error", func() {
										responseErr := errorsTest.RandomError()
										imageStructuredSession.UpdateOutputs = []imageStoreStructuredTest.UpdateOutput{{Image: nil, Error: responseErr}}
										imageUnstructuredStore.DeleteContentOutputs = []error{nil}
										result, err := client.PutContent(ctx, id, condition, contentIntent, content)
										errorsTest.ExpectEqual(err, responseErr)
										Expect(result).To(BeNil())
										imageUnstructuredStore.DeleteContentInputs = []imageStoreUnstructuredTest.DeleteContentInput{{UserID: *responseResult.UserID, ImageID: *responseResult.ID, ContentID: putContentID}}
									})

									It("returns an error when image structured session update returns an error and logs an error if image unstructured delete content returns an err", func() {
										responseErr := errorsTest.RandomError()
										deleteErr := errorsTest.RandomError()
										imageStructuredSession.UpdateOutputs = []imageStoreStructuredTest.UpdateOutput{{Image: nil, Error: responseErr}}
										imageUnstructuredStore.DeleteContentOutputs = []error{deleteErr}
										result, err := client.PutContent(ctx, id, condition, contentIntent, content)
										errorsTest.ExpectEqual(err, responseErr)
										Expect(result).To(BeNil())
										logger.AssertError("Unable to delete image content for failed update", log.Fields{"userId": *responseResult.UserID, "id": *responseResult.ID, "error": errors.NewSerializable(deleteErr)})
										imageUnstructuredStore.DeleteContentInputs = []imageStoreUnstructuredTest.DeleteContentInput{{UserID: *responseResult.UserID, ImageID: *responseResult.ID, ContentID: putContentID}}
									})

									When("the image structured session update returns successfully", func() {
										var updateImage *image.Image

										JustBeforeEach(func() {
											updateImage = imageTest.CloneImage(responseResult)
											updateImage.Status = pointer.FromString(image.StatusAvailable)
											updateImage.ContentIntent = pointer.FromString(contentIntent)
											updateImage.ContentAttributes = image.NewContentAttributes()
											updateImage.ContentAttributes.DigestMD5 = pointer.CloneString(content.DigestMD5)
											updateImage.ContentAttributes.MediaType = pointer.CloneString(content.MediaType)
											updateImage.ContentAttributes.Width = pointer.FromInt(width)
											updateImage.ContentAttributes.Height = pointer.FromInt(height)
											updateImage.ContentAttributes.Size = pointer.FromInt(int(size))
											updateImage.ModifiedTime = pointer.FromTime(test.RandomTimeFromRange(*responseResult.CreatedTime, time.Now()).Truncate(time.Second))
											imageStructuredSession.UpdateOutputs = []imageStoreStructuredTest.UpdateOutput{{Image: updateImage, Error: nil}}
										})

										When("the content intent is alternate", func() {
											BeforeEach(func() {
												contentIntent = image.ContentIntentAlternate
											})

											It("returns successfully", func() {
												Expect(client.PutContent(ctx, id, condition, contentIntent, content)).To(Equal(updateImage))
											})
										})

										When("the content intent is original", func() {
											BeforeEach(func() {
												contentIntent = image.ContentIntentOriginal
											})

											It("returns successfully and logs an error if image unstructured store delete contents returns an error", func() {
												responseErr := errorsTest.RandomError()
												imageUnstructuredStore.DeleteContentOutputs = []error{responseErr}
												Expect(client.PutContent(ctx, id, condition, contentIntent, content)).To(Equal(updateImage))
												logger.AssertError("Unable to delete image content for previous content intent", log.Fields{"userId": *responseResult.UserID, "id": *responseResult.ID, "error": errors.NewSerializable(responseErr)})
												imageUnstructuredStore.DeleteContentInputs = []imageStoreUnstructuredTest.DeleteContentInput{{UserID: *responseResult.UserID, ImageID: *responseResult.ID, ContentID: *responseResult.ContentID}}
											})

											It("returns successfully", func() {
												imageUnstructuredStore.DeleteContentOutputs = []error{nil}
												Expect(client.PutContent(ctx, id, condition, contentIntent, content)).To(Equal(updateImage))
												imageUnstructuredStore.DeleteContentInputs = []imageStoreUnstructuredTest.DeleteContentInput{{UserID: *responseResult.UserID, ImageID: *responseResult.ID, ContentID: *responseResult.ContentID}}
											})
										})

										When("the size is maximum", func() {
											BeforeEach(func() {
												contentIntent = image.ContentIntentAlternate
												contentBytes := make([]byte, 104857600)
												content.Body.Read(contentBytes)
												content.Body = ioutil.NopCloser(bytes.NewReader(contentBytes))
												content.DigestMD5 = pointer.FromString(crypto.Base64EncodedMD5Hash(contentBytes))
											})

											It("returns successfully", func() {
												Expect(client.PutContent(ctx, id, condition, contentIntent, content)).To(Equal(updateImage))
											})
										})
									})
								})
							})
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
					Expect(imageStructuredSession.GetInputs).To(Equal([]imageStoreStructuredTest.GetInput{{ID: id, Condition: condition}}))
				})

				It("returns an error when the image structured session get returns an error", func() {
					responseErr := errorsTest.RandomError()
					imageStructuredSession.GetOutputs = []imageStoreStructuredTest.GetOutput{{Image: nil, Error: responseErr}}
					deleted, err := client.Delete(ctx, id, condition)
					errorsTest.ExpectEqual(err, responseErr)
					Expect(deleted).To(BeFalse())
				})

				It("returns successfully when the image structured session get returns nil", func() {
					imageStructuredSession.GetOutputs = []imageStoreStructuredTest.GetOutput{{Image: nil, Error: nil}}
					deleted, err := client.Delete(ctx, id, condition)
					Expect(err).ToNot(HaveOccurred())
					Expect(deleted).To(BeFalse())
				})

				When("the image structure session get returns an image", func() {
					var responseResult *image.Image

					BeforeEach(func() {
						responseResult = imageTest.RandomImage()
						responseResult.ID = pointer.FromString(id)
						imageStructuredSession.GetOutputs = []imageStoreStructuredTest.GetOutput{{Image: responseResult, Error: nil}}
					})

					AfterEach(func() {
						Expect(authClient.EnsureAuthorizedUserInputs).To(Equal([]authTest.EnsureAuthorizedUserInput{{TargetUserID: *responseResult.UserID, AuthorizedPermission: permission.Write}}))
					})

					It("returns an error when the user client ensure authorized service returns an error", func() {
						responseErr := errorsTest.RandomError()
						authClient.EnsureAuthorizedUserOutputs = []authTest.EnsureAuthorizedUserOutput{{AuthorizedUserID: "", Error: responseErr}}
						deleted, err := client.Delete(ctx, id, condition)
						errorsTest.ExpectEqual(err, responseErr)
						Expect(deleted).To(BeFalse())
					})

					When("the user client ensure authorized service returns successfully", func() {
						BeforeEach(func() {
							authClient.EnsureAuthorizedUserOutputs = []authTest.EnsureAuthorizedUserOutput{{AuthorizedUserID: *responseResult.UserID, Error: nil}}
						})

						AfterEach(func() {
							Expect(imageStructuredSession.DeleteInputs).To(Equal([]imageStoreStructuredTest.DeleteInput{{ID: id, Condition: condition}}))
						})

						It("returns an error when the image structured session delete returns an error", func() {
							responseErr := errorsTest.RandomError()
							imageStructuredSession.DeleteOutputs = []imageStoreStructuredTest.DeleteOutput{{Deleted: false, Error: responseErr}}
							deleted, err := client.Delete(ctx, id, condition)
							errorsTest.ExpectEqual(err, responseErr)
							Expect(deleted).To(BeFalse())
						})

						It("returns successfully when the image structured session delete returns false", func() {
							imageStructuredSession.DeleteOutputs = []imageStoreStructuredTest.DeleteOutput{{Deleted: false, Error: nil}}
							deleted, err := client.Delete(ctx, id, condition)
							Expect(err).ToNot(HaveOccurred())
							Expect(deleted).To(BeFalse())
						})

						When("the image structured session delete returns successfully", func() {
							BeforeEach(func() {
								imageStructuredSession.DeleteOutputs = []imageStoreStructuredTest.DeleteOutput{{Deleted: true, Error: nil}}
							})

							AfterEach(func() {
								Expect(imageUnstructuredStore.DeleteInputs).To(Equal([]imageStoreUnstructuredTest.DeleteInput{{UserID: *responseResult.UserID, ImageID: id}}))
							})

							It("returns an error when the image unstructured store delete returns an error", func() {
								responseErr := errorsTest.RandomError()
								imageUnstructuredStore.DeleteOutputs = []error{responseErr}
								deleted, err := client.Delete(ctx, id, condition)
								errorsTest.ExpectEqual(err, responseErr)
								Expect(deleted).To(BeFalse())
							})

							When("the image unstructured store delete returns successfully", func() {
								BeforeEach(func() {
									imageUnstructuredStore.DeleteOutputs = []error{nil}
								})

								AfterEach(func() {
									Expect(imageStructuredSession.DestroyInputs).To(Equal([]imageStoreStructuredTest.DestroyInput{{ID: id}}))
								})

								It("returns an error when the image structured session destroy returns an error", func() {
									responseErr := errorsTest.RandomError()
									imageStructuredSession.DestroyOutputs = []imageStoreStructuredTest.DestroyOutput{{Destroyed: false, Error: responseErr}}
									deleted, err := client.Delete(ctx, id, condition)
									errorsTest.ExpectEqual(err, responseErr)
									Expect(deleted).To(BeFalse())
								})

								It("returns false when the image structured session destroy returns false", func() {
									imageStructuredSession.DestroyOutputs = []imageStoreStructuredTest.DestroyOutput{{Destroyed: false, Error: nil}}
									deleted, err := client.Delete(ctx, id, condition)
									Expect(err).ToNot(HaveOccurred())
									Expect(deleted).To(BeFalse())
								})

								It("returns true when the image structured session destroy returns true", func() {
									imageStructuredSession.DestroyOutputs = []imageStoreStructuredTest.DestroyOutput{{Destroyed: true, Error: nil}}
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
