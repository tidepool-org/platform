package client_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/crypto"
	cryptoTest "github.com/tidepool-org/platform/crypto/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/image"
	imageClient "github.com/tidepool-org/platform/image/client"
	imageMultipartTest "github.com/tidepool-org/platform/image/multipart/test"
	imageTest "github.com/tidepool-org/platform/image/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	netTest "github.com/tidepool-org/platform/net/test"
	"github.com/tidepool-org/platform/page"
	pageTest "github.com/tidepool-org/platform/page/test"
	"github.com/tidepool-org/platform/platform"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	requestTest "github.com/tidepool-org/platform/request/test"
	"github.com/tidepool-org/platform/test"
	testHttp "github.com/tidepool-org/platform/test/http"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("Client", func() {
	var config *platform.Config
	var authorizeAs platform.AuthorizeAs
	var formEncoder *imageMultipartTest.FormEncoder

	BeforeEach(func() {
		config = platform.NewConfig()
		config.UserAgent = testHttp.NewUserAgent()
		formEncoder = imageMultipartTest.NewFormEncoder()
	})

	AfterEach(func() {
		formEncoder.AssertOutputsEmpty()
	})

	Context("New", func() {
		BeforeEach(func() {
			config.Address = testHttp.NewAddress()
			authorizeAs = platform.AuthorizeAsService
		})

		It("returns an error when the config is missing", func() {
			config = nil
			client, err := imageClient.New(nil, authorizeAs, formEncoder)
			errorsTest.ExpectEqual(err, errors.New("config is missing"))
			Expect(client).To(BeNil())
		})

		It("returns an error when the authorize as is invalid", func() {
			authorizeAs = platform.AuthorizeAs(-1)
			client, err := imageClient.New(config, authorizeAs, formEncoder)
			errorsTest.ExpectEqual(err, errors.New("authorize as is invalid"))
			Expect(client).To(BeNil())
		})

		It("returns an error when the form encoder is missing", func() {
			client, err := imageClient.New(config, authorizeAs, nil)
			errorsTest.ExpectEqual(err, errors.New("form encoder is missing"))
			Expect(client).To(BeNil())
		})

		It("returns success", func() {
			Expect(imageClient.New(config, authorizeAs, formEncoder)).ToNot(BeNil())
		})
	})

	Context("with server and new client", func() {
		var server *Server
		var requestHandlers []http.HandlerFunc
		var responseHeaders http.Header
		var ctx context.Context
		var client image.Client

		BeforeEach(func() {
			server = NewServer()
			requestHandlers = nil
			responseHeaders = http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}
			ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
		})

		JustBeforeEach(func() {
			server.AppendHandlers(CombineHandlers(requestHandlers...))
			config.Address = server.URL()
			var err error
			client, err = imageClient.New(config, authorizeAs, formEncoder)
			Expect(err).ToNot(HaveOccurred())
			Expect(client).ToNot(BeNil())
		})

		AfterEach(func() {
			if server != nil {
				server.Close()
			}
		})

		authorizeAssertions := func() {
			Context("with user id", func() {
				var userID string

				BeforeEach(func() {
					userID = userTest.RandomID()
				})

				Context("List", func() {
					var filter *image.Filter
					var pagination *page.Pagination

					parameterAssertions := func() {
						Context("without server response", func() {
							AfterEach(func() {
								Expect(server.ReceivedRequests()).To(BeEmpty())
							})

							It("returns an error when the context is missing", func() {
								ctx = nil
								result, err := client.List(ctx, userID, filter, pagination)
								errorsTest.ExpectEqual(err, errors.New("context is missing"))
								Expect(result).To(BeNil())
							})

							It("returns an error when the user id is missing", func() {
								userID = ""
								result, err := client.List(ctx, userID, filter, pagination)
								errorsTest.ExpectEqual(err, errors.New("user id is missing"))
								Expect(result).To(BeNil())
							})

							It("returns an error when the user id is invalid", func() {
								userID = "invalid"
								result, err := client.List(ctx, userID, filter, pagination)
								errorsTest.ExpectEqual(err, errors.New("user id is invalid"))
								Expect(result).To(BeNil())
							})

							It("returns an error when the filter is invalid", func() {
								filter = image.NewFilter()
								filter.Status = pointer.FromStringArray([]string{""})
								result, err := client.List(ctx, userID, filter, pagination)
								errorsTest.ExpectEqual(err, errors.New("filter is invalid"))
								Expect(result).To(BeNil())
							})

							It("returns an error when the pagination is invalid", func() {
								pagination = page.NewPagination()
								pagination.Page = -1
								result, err := client.List(ctx, userID, filter, pagination)
								errorsTest.ExpectEqual(err, errors.New("pagination is invalid"))
								Expect(result).To(BeNil())
							})
						})

						Context("with server response", func() {
							BeforeEach(func() {
								requestHandlers = append(requestHandlers,
									VerifyContentType(""),
									VerifyBody(nil),
								)
							})

							AfterEach(func() {
								Expect(server.ReceivedRequests()).To(HaveLen(1))
							})

							When("the server responds with an unauthenticated error", func() {
								BeforeEach(func() {
									requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusUnauthorized, errors.NewSerializable(request.ErrorUnauthenticated()), responseHeaders))
								})

								It("returns an error", func() {
									result, err := client.List(ctx, userID, filter, pagination)
									errorsTest.ExpectEqual(err, request.ErrorUnauthenticated())
									Expect(result).To(BeNil())
								})
							})

							When("the server responds with an unauthorized error", func() {
								BeforeEach(func() {
									requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusForbidden, errors.NewSerializable(request.ErrorUnauthorized()), responseHeaders))
								})

								It("returns an error", func() {
									result, err := client.List(ctx, userID, filter, pagination)
									errorsTest.ExpectEqual(err, request.ErrorUnauthorized())
									Expect(result).To(BeNil())
								})
							})

							When("the server responds with a not found error", func() {
								BeforeEach(func() {
									requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusNotFound, errors.NewSerializable(request.ErrorResourceNotFoundWithID(userID)), responseHeaders))
								})

								It("returns an error", func() {
									result, err := client.List(ctx, userID, filter, pagination)
									errorsTest.ExpectEqual(err, request.ErrorResourceNotFoundWithID(userID))
									Expect(result).To(BeNil())
								})
							})

							When("the server responds with no result", func() {
								BeforeEach(func() {
									requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusOK, image.Images{}, responseHeaders))
								})

								It("returns successfully", func() {
									result, err := client.List(ctx, userID, filter, pagination)
									Expect(err).ToNot(HaveOccurred())
									Expect(result).To(Equal(image.Images{}))
								})
							})

							When("the server responds with result", func() {
								var responseResult image.Images

								BeforeEach(func() {
									responseResult = imageTest.RandomImages(1, 4)
									requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusOK, responseResult, responseHeaders))
								})

								It("returns successfully", func() {
									Expect(client.List(ctx, userID, filter, pagination)).To(imageTest.MatchImages(responseResult))
								})
							})
						})
					}

					When("the request has no filter or pagination parameters", func() {
						BeforeEach(func() {
							filter = nil
							pagination = nil
							requestHandlers = append(requestHandlers, VerifyRequest(http.MethodGet, fmt.Sprintf("/v1/users/%s/images", userID), ""))
						})

						parameterAssertions()
					})

					When("the request has random filter and pagination parameters", func() {
						BeforeEach(func() {
							filter = imageTest.RandomFilter()
							pagination = pageTest.RandomPagination()
							query := url.Values{
								"status":        *filter.Status,
								"contentIntent": *filter.ContentIntent,
								"page":          []string{strconv.Itoa(pagination.Page)},
								"size":          []string{strconv.Itoa(pagination.Size)},
							}
							requestHandlers = append(requestHandlers, VerifyRequest(http.MethodGet, fmt.Sprintf("/v1/users/%s/images", userID), query.Encode()))
						})

						parameterAssertions()
					})
				})

				Context("Create", func() {
					var contentIntent string

					contentIntentAssertions := func() {
						var metadata *image.Metadata
						var body []byte
						var content *image.Content

						BeforeEach(func() {
							metadata = imageTest.RandomMetadata()
							body = imageTest.RandomContentBytes()
							content = imageTest.RandomContent()
							content.Body = ioutil.NopCloser(bytes.NewReader(body))
						})

						Context("without server response", func() {
							AfterEach(func() {
								Expect(server.ReceivedRequests()).To(BeEmpty())
							})

							It("returns an error when the context is missing", func() {
								ctx = nil
								result, err := client.Create(ctx, userID, metadata, contentIntent, content)
								errorsTest.ExpectEqual(err, errors.New("context is missing"))
								Expect(result).To(BeNil())
							})

							It("returns an error when the user id is missing", func() {
								userID = ""
								result, err := client.Create(ctx, userID, metadata, contentIntent, content)
								errorsTest.ExpectEqual(err, errors.New("user id is missing"))
								Expect(result).To(BeNil())
							})

							It("returns an error when the user id is invalid", func() {
								userID = "invalid"
								result, err := client.Create(ctx, userID, metadata, contentIntent, content)
								errorsTest.ExpectEqual(err, errors.New("user id is invalid"))
								Expect(result).To(BeNil())
							})

							It("returns an error when the metadata is missing", func() {
								metadata = nil
								result, err := client.Create(ctx, userID, metadata, contentIntent, content)
								errorsTest.ExpectEqual(err, errors.New("metadata is missing"))
								Expect(result).To(BeNil())
							})

							It("returns an error when the metadata is invalid", func() {
								metadata.Name = pointer.FromString("")
								result, err := client.Create(ctx, userID, metadata, contentIntent, content)
								errorsTest.ExpectEqual(err, errors.New("metadata is invalid"))
								Expect(result).To(BeNil())
							})

							It("returns an error when the content intent is missing", func() {
								contentIntent = ""
								result, err := client.Create(ctx, userID, metadata, contentIntent, content)
								errorsTest.ExpectEqual(err, errors.New("content intent is missing"))
								Expect(result).To(BeNil())
							})

							It("returns an error when the content intent is invalid", func() {
								contentIntent = "invalid"
								result, err := client.Create(ctx, userID, metadata, contentIntent, content)
								errorsTest.ExpectEqual(err, errors.New("content intent is invalid"))
								Expect(result).To(BeNil())
							})

							It("returns an error when the content is missing", func() {
								content = nil
								result, err := client.Create(ctx, userID, metadata, contentIntent, content)
								errorsTest.ExpectEqual(err, errors.New("content is missing"))
								Expect(result).To(BeNil())
							})

							It("returns an error when the content is invalid", func() {
								content.Body = nil
								result, err := client.Create(ctx, userID, metadata, contentIntent, content)
								errorsTest.ExpectEqual(err, errors.New("content is invalid"))
								Expect(result).To(BeNil())
							})
						})

						Context("with multipart", func() {
							var multipartBody []byte
							var multipartReader io.ReadCloser
							var multipartContentType string

							BeforeEach(func() {
								multipartBody = imageTest.RandomContentBytes()
								multipartReader = ioutil.NopCloser(bytes.NewReader(multipartBody))
								multipartContentType = netTest.RandomMediaType()
							})

							AfterEach(func() {
								Expect(formEncoder.EncodeFormInputs).To(Equal([]imageMultipartTest.EncodeFormInput{{Metadata: metadata, ContentIntent: contentIntent, Content: content}}))
							})

							It("returns an error when the encode multipart reader is missing", func() {
								formEncoder.EncodeFormOutputs = []imageMultipartTest.EncodeFormOutput{{Reader: nil, ContentType: multipartContentType}}
								result, err := client.Create(ctx, userID, metadata, contentIntent, content)
								errorsTest.ExpectEqual(err, errors.New("multipart reader is missing"))
								Expect(result).To(BeNil())
								Expect(server.ReceivedRequests()).To(BeEmpty())
							})

							It("returns an error when the encode multipart content type is missing", func() {
								formEncoder.EncodeFormOutputs = []imageMultipartTest.EncodeFormOutput{{Reader: multipartReader, ContentType: ""}}
								result, err := client.Create(ctx, userID, metadata, contentIntent, content)
								errorsTest.ExpectEqual(err, errors.New("multipart content type is missing"))
								Expect(result).To(BeNil())
								Expect(server.ReceivedRequests()).To(BeEmpty())
							})

							digestAssertions := func() {
								Context("with server response", func() {
									BeforeEach(func() {
										formEncoder.EncodeFormOutputs = []imageMultipartTest.EncodeFormOutput{{Reader: multipartReader, ContentType: multipartContentType}}
										requestHandlers = append(requestHandlers,
											VerifyRequest(http.MethodPost, fmt.Sprintf("/v1/users/%s/images", userID)),
											VerifyContentType(multipartContentType),
											VerifyBody(multipartBody),
										)
									})

									AfterEach(func() {
										Expect(server.ReceivedRequests()).To(HaveLen(1))
									})

									When("the server responds with an unauthenticated error", func() {
										BeforeEach(func() {
											requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusUnauthorized, errors.NewSerializable(request.ErrorUnauthenticated()), responseHeaders))
										})

										It("returns an error", func() {
											result, err := client.Create(ctx, userID, metadata, contentIntent, content)
											errorsTest.ExpectEqual(err, request.ErrorUnauthenticated())
											Expect(result).To(BeNil())
										})
									})

									When("the server responds with an unauthorized error", func() {
										BeforeEach(func() {
											requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusForbidden, errors.NewSerializable(request.ErrorUnauthorized()), responseHeaders))
										})

										It("returns an error", func() {
											result, err := client.Create(ctx, userID, metadata, contentIntent, content)
											errorsTest.ExpectEqual(err, request.ErrorUnauthorized())
											Expect(result).To(BeNil())
										})
									})

									When("the server responds with a not found error", func() {
										BeforeEach(func() {
											requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusNotFound, errors.NewSerializable(request.ErrorResourceNotFoundWithID(userID)), responseHeaders))
										})

										It("returns successfully without result", func() {
											result, err := client.Create(ctx, userID, metadata, contentIntent, content)
											errorsTest.ExpectEqual(err, request.ErrorResourceNotFoundWithID(userID))
											Expect(result).To(BeNil())
										})
									})

									When("the server responds with the result", func() {
										var responseResult *image.Image

										BeforeEach(func() {
											responseResult = imageTest.RandomImage()
											requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusOK, responseResult, responseHeaders))
										})

										It("returns successfully", func() {
											Expect(client.Create(ctx, userID, metadata, contentIntent, content)).To(imageTest.MatchImage(responseResult))
										})
									})
								})
							}

							When("the content has no digest", func() {
								BeforeEach(func() {
									content.DigestMD5 = nil
								})

								digestAssertions()
							})

							When("the content has a digest", func() {
								BeforeEach(func() {
									content.DigestMD5 = pointer.FromString(crypto.Base64EncodedMD5Hash(body))
								})

								digestAssertions()
							})
						})
					}

					When("the content intent is alternate", func() {
						BeforeEach(func() {
							contentIntent = image.ContentIntentAlternate
						})

						contentIntentAssertions()
					})

					When("the content intent is original", func() {
						BeforeEach(func() {
							contentIntent = image.ContentIntentOriginal
						})

						contentIntentAssertions()
					})
				})

				Context("CreateWithMetadata", func() {
					var metadata *image.Metadata

					BeforeEach(func() {
						metadata = imageTest.RandomMetadata()
					})

					Context("without server response", func() {
						AfterEach(func() {
							Expect(server.ReceivedRequests()).To(BeEmpty())
						})

						It("returns an error when the context is missing", func() {
							ctx = nil
							result, err := client.CreateWithMetadata(ctx, userID, metadata)
							errorsTest.ExpectEqual(err, errors.New("context is missing"))
							Expect(result).To(BeNil())
						})

						It("returns an error when the user id is missing", func() {
							userID = ""
							result, err := client.CreateWithMetadata(ctx, userID, metadata)
							errorsTest.ExpectEqual(err, errors.New("user id is missing"))
							Expect(result).To(BeNil())
						})

						It("returns an error when the user id is invalid", func() {
							userID = "invalid"
							result, err := client.CreateWithMetadata(ctx, userID, metadata)
							errorsTest.ExpectEqual(err, errors.New("user id is invalid"))
							Expect(result).To(BeNil())
						})

						It("returns an error when the metadata is missing", func() {
							metadata = nil
							result, err := client.CreateWithMetadata(ctx, userID, metadata)
							errorsTest.ExpectEqual(err, errors.New("metadata is missing"))
							Expect(result).To(BeNil())
						})

						It("returns an error when the metadata is invalid", func() {
							metadata.Name = pointer.FromString("")
							result, err := client.CreateWithMetadata(ctx, userID, metadata)
							errorsTest.ExpectEqual(err, errors.New("metadata is invalid"))
							Expect(result).To(BeNil())
						})
					})

					Context("with server response", func() {
						BeforeEach(func() {
							requestHandlers = append(requestHandlers,
								VerifyRequest(http.MethodPost, fmt.Sprintf("/v1/users/%s/images/metadata", userID)),
								VerifyContentType("application/json; charset=utf-8"),
								VerifyBody(test.MustBytes(test.MarshalRequestBody(metadata))),
							)
						})

						AfterEach(func() {
							Expect(server.ReceivedRequests()).To(HaveLen(1))
						})

						When("the server responds with an unauthenticated error", func() {
							BeforeEach(func() {
								requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusUnauthorized, errors.NewSerializable(request.ErrorUnauthenticated()), responseHeaders))
							})

							It("returns an error", func() {
								result, err := client.CreateWithMetadata(ctx, userID, metadata)
								errorsTest.ExpectEqual(err, request.ErrorUnauthenticated())
								Expect(result).To(BeNil())
							})
						})

						When("the server responds with an unauthorized error", func() {
							BeforeEach(func() {
								requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusForbidden, errors.NewSerializable(request.ErrorUnauthorized()), responseHeaders))
							})

							It("returns an error", func() {
								result, err := client.CreateWithMetadata(ctx, userID, metadata)
								errorsTest.ExpectEqual(err, request.ErrorUnauthorized())
								Expect(result).To(BeNil())
							})
						})

						When("the server responds with a not found error", func() {
							BeforeEach(func() {
								requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusNotFound, errors.NewSerializable(request.ErrorResourceNotFoundWithID(userID)), responseHeaders))
							})

							It("returns an error", func() {
								result, err := client.CreateWithMetadata(ctx, userID, metadata)
								errorsTest.ExpectEqual(err, request.ErrorResourceNotFoundWithID(userID))
								Expect(result).To(BeNil())
							})
						})

						When("the server responds with the result", func() {
							var responseResult *image.Image

							BeforeEach(func() {
								responseResult = imageTest.RandomImage()
								requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusOK, responseResult, responseHeaders))
							})

							It("returns successfully", func() {
								Expect(client.CreateWithMetadata(ctx, userID, metadata)).To(imageTest.MatchImage(responseResult))
							})
						})
					})
				})

				Context("CreateWithContent", func() {
					var contentIntent string

					contentIntentAssertions := func() {
						var body []byte
						var content *image.Content

						BeforeEach(func() {
							body = imageTest.RandomContentBytes()
							content = imageTest.RandomContent()
							content.Body = ioutil.NopCloser(bytes.NewReader(body))
						})

						Context("without server response", func() {
							AfterEach(func() {
								Expect(server.ReceivedRequests()).To(BeEmpty())
							})

							It("returns an error when the context is missing", func() {
								ctx = nil
								result, err := client.CreateWithContent(ctx, userID, contentIntent, content)
								errorsTest.ExpectEqual(err, errors.New("context is missing"))
								Expect(result).To(BeNil())
							})

							It("returns an error when the user id is missing", func() {
								userID = ""
								result, err := client.CreateWithContent(ctx, userID, contentIntent, content)
								errorsTest.ExpectEqual(err, errors.New("user id is missing"))
								Expect(result).To(BeNil())
							})

							It("returns an error when the user id is invalid", func() {
								userID = "invalid"
								result, err := client.CreateWithContent(ctx, userID, contentIntent, content)
								errorsTest.ExpectEqual(err, errors.New("user id is invalid"))
								Expect(result).To(BeNil())
							})

							It("returns an error when the content intent is missing", func() {
								contentIntent = ""
								result, err := client.CreateWithContent(ctx, userID, contentIntent, content)
								errorsTest.ExpectEqual(err, errors.New("content intent is missing"))
								Expect(result).To(BeNil())
							})

							It("returns an error when the content intent is invalid", func() {
								contentIntent = "invalid"
								result, err := client.CreateWithContent(ctx, userID, contentIntent, content)
								errorsTest.ExpectEqual(err, errors.New("content intent is invalid"))
								Expect(result).To(BeNil())
							})

							It("returns an error when the content is missing", func() {
								content = nil
								result, err := client.CreateWithContent(ctx, userID, contentIntent, content)
								errorsTest.ExpectEqual(err, errors.New("content is missing"))
								Expect(result).To(BeNil())
							})

							It("returns an error when the content is invalid", func() {
								content.Body = nil
								result, err := client.CreateWithContent(ctx, userID, contentIntent, content)
								errorsTest.ExpectEqual(err, errors.New("content is invalid"))
								Expect(result).To(BeNil())
							})
						})

						digestAssertions := func() {
							Context("with server response", func() {
								BeforeEach(func() {
									requestHandlers = append(requestHandlers,
										VerifyRequest(http.MethodPost, fmt.Sprintf("/v1/users/%s/images/content/%s", userID, contentIntent)),
										VerifyContentType(*content.MediaType),
										VerifyBody(body),
									)
								})

								AfterEach(func() {
									Expect(server.ReceivedRequests()).To(HaveLen(1))
								})

								When("the server responds with an unauthenticated error", func() {
									BeforeEach(func() {
										requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusUnauthorized, errors.NewSerializable(request.ErrorUnauthenticated()), responseHeaders))
									})

									It("returns an error", func() {
										result, err := client.CreateWithContent(ctx, userID, contentIntent, content)
										errorsTest.ExpectEqual(err, request.ErrorUnauthenticated())
										Expect(result).To(BeNil())
									})
								})

								When("the server responds with an unauthorized error", func() {
									BeforeEach(func() {
										requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusForbidden, errors.NewSerializable(request.ErrorUnauthorized()), responseHeaders))
									})

									It("returns an error", func() {
										result, err := client.CreateWithContent(ctx, userID, contentIntent, content)
										errorsTest.ExpectEqual(err, request.ErrorUnauthorized())
										Expect(result).To(BeNil())
									})
								})

								When("the server responds with a not found error", func() {
									BeforeEach(func() {
										requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusNotFound, errors.NewSerializable(request.ErrorResourceNotFoundWithID(userID)), responseHeaders))
									})

									It("returns successfully without result", func() {
										result, err := client.CreateWithContent(ctx, userID, contentIntent, content)
										errorsTest.ExpectEqual(err, request.ErrorResourceNotFoundWithID(userID))
										Expect(result).To(BeNil())
									})
								})

								When("the server responds with the result", func() {
									var responseResult *image.Image

									BeforeEach(func() {
										responseResult = imageTest.RandomImage()
										requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusOK, responseResult, responseHeaders))
									})

									It("returns successfully", func() {
										Expect(client.CreateWithContent(ctx, userID, contentIntent, content)).To(imageTest.MatchImage(responseResult))
									})
								})
							})
						}

						When("the request has no digest header", func() {
							BeforeEach(func() {
								content.DigestMD5 = nil
							})

							digestAssertions()
						})

						When("the request has a digest header", func() {
							BeforeEach(func() {
								requestHandlers = append(requestHandlers, VerifyHeaderKV("Digest", fmt.Sprintf("MD5=%s", *content.DigestMD5)))
							})

							digestAssertions()
						})
					}

					When("the content intent is alternate", func() {
						BeforeEach(func() {
							contentIntent = image.ContentIntentAlternate
						})

						contentIntentAssertions()
					})

					When("the content intent is original", func() {
						BeforeEach(func() {
							contentIntent = image.ContentIntentOriginal
						})

						contentIntentAssertions()
					})
				})

				Context("DeleteAll", func() {
					Context("without server response", func() {
						AfterEach(func() {
							Expect(server.ReceivedRequests()).To(BeEmpty())
						})

						It("returns an error when the context is missing", func() {
							ctx = nil
							errorsTest.ExpectEqual(client.DeleteAll(ctx, userID), errors.New("context is missing"))
						})

						It("returns an error when the user id is missing", func() {
							userID = ""
							errorsTest.ExpectEqual(client.DeleteAll(ctx, userID), errors.New("user id is missing"))
						})

						It("returns an error when the user id is invalid", func() {
							userID = "invalid"
							errorsTest.ExpectEqual(client.DeleteAll(ctx, userID), errors.New("user id is invalid"))
						})
					})

					Context("with server response", func() {
						BeforeEach(func() {
							requestHandlers = append(requestHandlers,
								VerifyRequest(http.MethodDelete, fmt.Sprintf("/v1/users/%s/images", userID)),
								VerifyContentType(""),
								VerifyBody(nil),
							)
						})

						AfterEach(func() {
							Expect(server.ReceivedRequests()).To(HaveLen(1))
						})

						When("the server responds with an unauthenticated error", func() {
							BeforeEach(func() {
								requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusUnauthorized, errors.NewSerializable(request.ErrorUnauthenticated()), responseHeaders))
							})

							It("returns an error", func() {
								errorsTest.ExpectEqual(client.DeleteAll(ctx, userID), request.ErrorUnauthenticated())
							})
						})

						When("the server responds with an unauthorized error", func() {
							BeforeEach(func() {
								requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusForbidden, errors.NewSerializable(request.ErrorUnauthorized()), responseHeaders))
							})

							It("returns an error", func() {
								errorsTest.ExpectEqual(client.DeleteAll(ctx, userID), request.ErrorUnauthorized())
							})
						})

						When("the server responds successfully", func() {
							BeforeEach(func() {
								requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusNoContent, nil, responseHeaders))
							})

							It("returns successfully", func() {
								Expect(client.DeleteAll(ctx, userID)).To(Succeed())
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
					Context("without server response", func() {
						AfterEach(func() {
							Expect(server.ReceivedRequests()).To(BeEmpty())
						})

						It("returns an error when the context is missing", func() {
							ctx = nil
							result, err := client.Get(ctx, id)
							errorsTest.ExpectEqual(err, errors.New("context is missing"))
							Expect(result).To(BeNil())
						})

						It("returns an error when the id is missing", func() {
							id = ""
							result, err := client.Get(ctx, id)
							errorsTest.ExpectEqual(err, errors.New("id is missing"))
							Expect(result).To(BeNil())
						})

						It("returns an error when the id is invalid", func() {
							id = "invalid"
							result, err := client.Get(ctx, id)
							errorsTest.ExpectEqual(err, errors.New("id is invalid"))
							Expect(result).To(BeNil())
						})
					})

					Context("with server response", func() {
						BeforeEach(func() {
							requestHandlers = append(requestHandlers,
								VerifyRequest(http.MethodGet, fmt.Sprintf("/v1/images/%s", id)),
								VerifyContentType(""),
								VerifyBody(nil),
							)
						})

						AfterEach(func() {
							Expect(server.ReceivedRequests()).To(HaveLen(1))
						})

						When("the server responds with an unauthenticated error", func() {
							BeforeEach(func() {
								requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusUnauthorized, errors.NewSerializable(request.ErrorUnauthenticated()), responseHeaders))
							})

							It("returns an error", func() {
								result, err := client.Get(ctx, id)
								errorsTest.ExpectEqual(err, request.ErrorUnauthenticated())
								Expect(result).To(BeNil())
							})
						})

						When("the server responds with an unauthorized error", func() {
							BeforeEach(func() {
								requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusForbidden, errors.NewSerializable(request.ErrorUnauthorized()), responseHeaders))
							})

							It("returns an error", func() {
								result, err := client.Get(ctx, id)
								errorsTest.ExpectEqual(err, request.ErrorUnauthorized())
								Expect(result).To(BeNil())
							})
						})

						When("the server responds with a not found error", func() {
							BeforeEach(func() {
								requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusNotFound, errors.NewSerializable(request.ErrorResourceNotFoundWithID(id)), responseHeaders))
							})

							It("returns successfully without result", func() {
								result, err := client.Get(ctx, id)
								Expect(err).ToNot(HaveOccurred())
								Expect(result).To(BeNil())
							})
						})

						When("the server responds with the result", func() {
							var responseResult *image.Image

							BeforeEach(func() {
								responseResult = imageTest.RandomImage()
								requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusOK, responseResult, responseHeaders))
							})

							It("returns successfully with result", func() {
								Expect(client.Get(ctx, id)).To(imageTest.MatchImage(responseResult))
							})
						})
					})
				})

				Context("GetMetadata", func() {
					Context("without server response", func() {
						AfterEach(func() {
							Expect(server.ReceivedRequests()).To(BeEmpty())
						})

						It("returns an error when the context is missing", func() {
							ctx = nil
							result, err := client.GetMetadata(ctx, id)
							errorsTest.ExpectEqual(err, errors.New("context is missing"))
							Expect(result).To(BeNil())
						})

						It("returns an error when the id is missing", func() {
							id = ""
							result, err := client.GetMetadata(ctx, id)
							errorsTest.ExpectEqual(err, errors.New("id is missing"))
							Expect(result).To(BeNil())
						})

						It("returns an error when the id is invalid", func() {
							id = "invalid"
							result, err := client.GetMetadata(ctx, id)
							errorsTest.ExpectEqual(err, errors.New("id is invalid"))
							Expect(result).To(BeNil())
						})
					})

					Context("with server response", func() {
						BeforeEach(func() {
							requestHandlers = append(requestHandlers,
								VerifyRequest(http.MethodGet, fmt.Sprintf("/v1/images/%s/metadata", id)),
								VerifyContentType(""),
								VerifyBody(nil),
							)
						})

						AfterEach(func() {
							Expect(server.ReceivedRequests()).To(HaveLen(1))
						})

						When("the server responds with an unauthenticated error", func() {
							BeforeEach(func() {
								requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusUnauthorized, errors.NewSerializable(request.ErrorUnauthenticated()), responseHeaders))
							})

							It("returns an error", func() {
								result, err := client.GetMetadata(ctx, id)
								errorsTest.ExpectEqual(err, request.ErrorUnauthenticated())
								Expect(result).To(BeNil())
							})
						})

						When("the server responds with an unauthorized error", func() {
							BeforeEach(func() {
								requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusForbidden, errors.NewSerializable(request.ErrorUnauthorized()), responseHeaders))
							})

							It("returns an error", func() {
								result, err := client.GetMetadata(ctx, id)
								errorsTest.ExpectEqual(err, request.ErrorUnauthorized())
								Expect(result).To(BeNil())
							})
						})

						When("the server responds with a not found error", func() {
							BeforeEach(func() {
								requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusNotFound, errors.NewSerializable(request.ErrorResourceNotFoundWithID(id)), responseHeaders))
							})

							It("returns successfully without result", func() {
								result, err := client.GetMetadata(ctx, id)
								Expect(err).ToNot(HaveOccurred())
								Expect(result).To(BeNil())
							})
						})

						When("the server responds with the result", func() {
							var responseResult *image.Metadata

							BeforeEach(func() {
								responseResult = imageTest.RandomMetadata()
								requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusOK, responseResult, responseHeaders))
							})

							It("returns successfully with result", func() {
								result, err := client.GetMetadata(ctx, id)
								Expect(err).ToNot(HaveOccurred())
								Expect(result).To(Equal(responseResult))
							})
						})
					})
				})

				Context("GetContent", func() {
					var contentIntent *string

					contentIntentAssertions := func() {
						Context("without server response", func() {
							AfterEach(func() {
								Expect(server.ReceivedRequests()).To(BeEmpty())
							})

							It("returns an error when the context is missing", func() {
								ctx = nil
								result, err := client.GetContent(ctx, id, contentIntent)
								errorsTest.ExpectEqual(err, errors.New("context is missing"))
								Expect(result).To(BeNil())
							})

							It("returns an error when the id is missing", func() {
								id = ""
								result, err := client.GetContent(ctx, id, contentIntent)
								errorsTest.ExpectEqual(err, errors.New("id is missing"))
								Expect(result).To(BeNil())
							})

							It("returns an error when the id is invalid", func() {
								id = "invalid"
								result, err := client.GetContent(ctx, id, contentIntent)
								errorsTest.ExpectEqual(err, errors.New("id is invalid"))
								Expect(result).To(BeNil())
							})

							It("returns an error when the content intent is invalid", func() {
								contentIntent = pointer.FromString("invalid")
								result, err := client.GetContent(ctx, id, contentIntent)
								errorsTest.ExpectEqual(err, errors.New("content intent is invalid"))
								Expect(result).To(BeNil())
							})
						})

						Context("with server response", func() {
							BeforeEach(func() {
								requestHandlers = append(requestHandlers,
									VerifyContentType(""),
									VerifyBody(nil),
								)
							})

							AfterEach(func() {
								Expect(server.ReceivedRequests()).To(HaveLen(1))
							})

							When("the server responds with an unauthenticated error", func() {
								BeforeEach(func() {
									requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusUnauthorized, errors.NewSerializable(request.ErrorUnauthenticated()), responseHeaders))
								})

								It("returns an error", func() {
									result, err := client.GetContent(ctx, id, contentIntent)
									errorsTest.ExpectEqual(err, request.ErrorUnauthenticated())
									Expect(result).To(BeNil())
								})
							})

							When("the server responds with an unauthorized error", func() {
								BeforeEach(func() {
									requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusForbidden, errors.NewSerializable(request.ErrorUnauthorized()), responseHeaders))
								})

								It("returns an error", func() {
									result, err := client.GetContent(ctx, id, contentIntent)
									errorsTest.ExpectEqual(err, request.ErrorUnauthorized())
									Expect(result).To(BeNil())
								})
							})

							When("the server responds with a not found error", func() {
								BeforeEach(func() {
									requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusNotFound, errors.NewSerializable(request.ErrorResourceNotFoundWithID(id)), responseHeaders))
								})

								It("returns successfully without result", func() {
									result, err := client.GetContent(ctx, id, contentIntent)
									Expect(err).ToNot(HaveOccurred())
									Expect(result).To(BeNil())
								})
							})

							When("the server responds with an invalid digest header", func() {
								BeforeEach(func() {
									responseHeaders = http.Header{
										"Digest": []string{"invalid"},
									}
									requestHandlers = append(requestHandlers, RespondWith(http.StatusOK, nil, responseHeaders))
								})

								It("returns successfully", func() {
									result, err := client.GetContent(ctx, id, contentIntent)
									errorsTest.ExpectEqual(err, request.ErrorHeaderInvalid("Digest"))
									Expect(result).To(BeNil())
								})
							})

							When("the server responds with an invalid content type header", func() {
								var digestMD5 string

								BeforeEach(func() {
									digestMD5 = cryptoTest.RandomBase64EncodedMD5Hash()
									responseHeaders = http.Header{
										"Digest":       []string{fmt.Sprintf("MD5=%s", digestMD5)},
										"Content-Type": []string{"/"},
									}
									requestHandlers = append(requestHandlers, RespondWith(http.StatusOK, nil, responseHeaders))
								})

								It("returns successfully", func() {
									result, err := client.GetContent(ctx, id, contentIntent)
									errorsTest.ExpectEqual(err, request.ErrorHeaderInvalid("Content-Type"))
									Expect(result).To(BeNil())
								})
							})

							When("the server responds with the result", func() {
								var body []byte
								var digestMD5 string
								var mediaType string

								BeforeEach(func() {
									body = imageTest.RandomContentBytes()
									digestMD5 = cryptoTest.RandomBase64EncodedMD5Hash()
									mediaType = imageTest.RandomMediaType()
									responseHeaders = http.Header{
										"Digest":       []string{fmt.Sprintf("MD5=%s", digestMD5)},
										"Content-Type": []string{mediaType},
									}
									requestHandlers = append(requestHandlers, RespondWith(http.StatusOK, body, responseHeaders))
								})

								It("returns successfully", func() {
									result, err := client.GetContent(ctx, id, contentIntent)
									Expect(err).ToNot(HaveOccurred())
									Expect(result).ToNot(BeNil())
									Expect(result.Body).ToNot(BeNil())
									defer result.Body.Close()
									Expect(result.DigestMD5).To(Equal(&digestMD5))
									Expect(result.MediaType).To(Equal(&mediaType))
								})
							})
						})
					}

					When("the content intent is missing", func() {
						BeforeEach(func() {
							contentIntent = nil
							requestHandlers = append(requestHandlers, VerifyRequest(http.MethodGet, fmt.Sprintf("/v1/images/%s/content", id)))
						})

						contentIntentAssertions()
					})

					When("the content intent is alternate", func() {
						BeforeEach(func() {
							contentIntent = pointer.FromString(image.ContentIntentAlternate)
							requestHandlers = append(requestHandlers, VerifyRequest(http.MethodGet, fmt.Sprintf("/v1/images/%s/content/%s", id, *contentIntent)))
						})

						contentIntentAssertions()
					})

					When("the content intent is original", func() {
						BeforeEach(func() {
							contentIntent = pointer.FromString(image.ContentIntentOriginal)
							requestHandlers = append(requestHandlers, VerifyRequest(http.MethodGet, fmt.Sprintf("/v1/images/%s/content/%s", id, *contentIntent)))
						})

						contentIntentAssertions()
					})
				})

				Context("GetRenditionContent", func() {
					var rendition *image.Rendition

					BeforeEach(func() {
						rendition = imageTest.RandomRendition()
					})

					Context("without server response", func() {
						AfterEach(func() {
							Expect(server.ReceivedRequests()).To(BeEmpty())
						})

						It("returns an error when the context is missing", func() {
							ctx = nil
							result, err := client.GetRenditionContent(ctx, id, rendition)
							errorsTest.ExpectEqual(err, errors.New("context is missing"))
							Expect(result).To(BeNil())
						})

						It("returns an error when the id is missing", func() {
							id = ""
							result, err := client.GetRenditionContent(ctx, id, rendition)
							errorsTest.ExpectEqual(err, errors.New("id is missing"))
							Expect(result).To(BeNil())
						})

						It("returns an error when the id is invalid", func() {
							id = "invalid"
							result, err := client.GetRenditionContent(ctx, id, rendition)
							errorsTest.ExpectEqual(err, errors.New("id is invalid"))
							Expect(result).To(BeNil())
						})

						It("returns an error when the rendition is missing", func() {
							rendition = nil
							result, err := client.GetRenditionContent(ctx, id, rendition)
							errorsTest.ExpectEqual(err, errors.New("rendition is missing"))
							Expect(result).To(BeNil())
						})

						It("returns an error when the rendition is invalid", func() {
							rendition.MediaType = pointer.FromString("")
							result, err := client.GetRenditionContent(ctx, id, rendition)
							errorsTest.ExpectEqual(err, errors.New("rendition is invalid"))
							Expect(result).To(BeNil())
						})
					})

					Context("with server response", func() {
						BeforeEach(func() {
							requestHandlers = append(requestHandlers,
								VerifyContentType(""),
								VerifyBody(nil),
							)
						})

						AfterEach(func() {
							Expect(server.ReceivedRequests()).To(HaveLen(1))
						})

						When("the server responds with an unauthenticated error", func() {
							BeforeEach(func() {
								requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusUnauthorized, errors.NewSerializable(request.ErrorUnauthenticated()), responseHeaders))
							})

							It("returns an error", func() {
								result, err := client.GetRenditionContent(ctx, id, rendition)
								errorsTest.ExpectEqual(err, request.ErrorUnauthenticated())
								Expect(result).To(BeNil())
							})
						})

						When("the server responds with an unauthorized error", func() {
							BeforeEach(func() {
								requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusForbidden, errors.NewSerializable(request.ErrorUnauthorized()), responseHeaders))
							})

							It("returns an error", func() {
								result, err := client.GetRenditionContent(ctx, id, rendition)
								errorsTest.ExpectEqual(err, request.ErrorUnauthorized())
								Expect(result).To(BeNil())
							})
						})

						When("the server responds with a not found error", func() {
							BeforeEach(func() {
								requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusNotFound, errors.NewSerializable(request.ErrorResourceNotFoundWithID(id)), responseHeaders))
							})

							It("returns successfully without result", func() {
								result, err := client.GetRenditionContent(ctx, id, rendition)
								Expect(err).ToNot(HaveOccurred())
								Expect(result).To(BeNil())
							})
						})

						When("the server responds with an invalid digest header", func() {
							BeforeEach(func() {
								responseHeaders = http.Header{
									"Digest": []string{"invalid"},
								}
								requestHandlers = append(requestHandlers, RespondWith(http.StatusOK, nil, responseHeaders))
							})

							It("returns successfully", func() {
								result, err := client.GetRenditionContent(ctx, id, rendition)
								errorsTest.ExpectEqual(err, request.ErrorHeaderInvalid("Digest"))
								Expect(result).To(BeNil())
							})
						})

						When("the server responds with an invalid content type header", func() {
							var digestMD5 string

							BeforeEach(func() {
								digestMD5 = cryptoTest.RandomBase64EncodedMD5Hash()
								responseHeaders = http.Header{
									"Digest":       []string{fmt.Sprintf("MD5=%s", digestMD5)},
									"Content-Type": []string{"/"},
								}
								requestHandlers = append(requestHandlers, RespondWith(http.StatusOK, nil, responseHeaders))
							})

							It("returns successfully", func() {
								result, err := client.GetRenditionContent(ctx, id, rendition)
								errorsTest.ExpectEqual(err, request.ErrorHeaderInvalid("Content-Type"))
								Expect(result).To(BeNil())
							})
						})

						When("the server responds with the result", func() {
							var body []byte
							var digestMD5 string
							var mediaType string

							BeforeEach(func() {
								body = imageTest.RandomContentBytes()
								digestMD5 = cryptoTest.RandomBase64EncodedMD5Hash()
								mediaType = imageTest.RandomMediaType()
								responseHeaders = http.Header{
									"Digest":       []string{fmt.Sprintf("MD5=%s", digestMD5)},
									"Content-Type": []string{mediaType},
								}
								requestHandlers = append(requestHandlers, RespondWith(http.StatusOK, body, responseHeaders))
							})

							It("returns successfully", func() {
								result, err := client.GetRenditionContent(ctx, id, rendition)
								Expect(err).ToNot(HaveOccurred())
								Expect(result).ToNot(BeNil())
								Expect(result.Body).ToNot(BeNil())
								defer result.Body.Close()
								Expect(result.DigestMD5).To(Equal(&digestMD5))
								Expect(result.MediaType).To(Equal(&mediaType))
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

					Context("without server response", func() {
						AfterEach(func() {
							Expect(server.ReceivedRequests()).To(BeEmpty())
						})

						It("returns an error when the context is missing", func() {
							ctx = nil
							result, err := client.PutMetadata(ctx, id, condition, metadata)
							errorsTest.ExpectEqual(err, errors.New("context is missing"))
							Expect(result).To(BeNil())
						})

						It("returns an error when the id is missing", func() {
							id = ""
							result, err := client.PutMetadata(ctx, id, condition, metadata)
							errorsTest.ExpectEqual(err, errors.New("id is missing"))
							Expect(result).To(BeNil())
						})

						It("returns an error when the id is invalid", func() {
							id = "invalid"
							result, err := client.PutMetadata(ctx, id, condition, metadata)
							errorsTest.ExpectEqual(err, errors.New("id is invalid"))
							Expect(result).To(BeNil())
						})

						It("returns an error when the condition is invalid", func() {
							condition.Revision = pointer.FromInt(-1)
							result, err := client.PutMetadata(ctx, id, condition, metadata)
							errorsTest.ExpectEqual(err, errors.New("condition is invalid"))
							Expect(result).To(BeNil())
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
					})

					conditionAssertions := func() {
						Context("with server response", func() {
							BeforeEach(func() {
								requestHandlers = append(requestHandlers,
									VerifyContentType("application/json; charset=utf-8"),
									VerifyBody(test.MustBytes(test.MarshalRequestBody(metadata))),
								)
							})

							AfterEach(func() {
								Expect(server.ReceivedRequests()).To(HaveLen(1))
							})

							When("the server responds with an unauthenticated error", func() {
								BeforeEach(func() {
									requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusUnauthorized, errors.NewSerializable(request.ErrorUnauthenticated()), responseHeaders))
								})

								It("returns an error", func() {
									result, err := client.PutMetadata(ctx, id, condition, metadata)
									errorsTest.ExpectEqual(err, request.ErrorUnauthenticated())
									Expect(result).To(BeNil())
								})
							})

							When("the server responds with an unauthorized error", func() {
								BeforeEach(func() {
									requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusForbidden, errors.NewSerializable(request.ErrorUnauthorized()), responseHeaders))
								})

								It("returns an error", func() {
									result, err := client.PutMetadata(ctx, id, condition, metadata)
									errorsTest.ExpectEqual(err, request.ErrorUnauthorized())
									Expect(result).To(BeNil())
								})
							})

							When("the server responds with a not found error", func() {
								BeforeEach(func() {
									requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusNotFound, errors.NewSerializable(request.ErrorResourceNotFoundWithID(id)), responseHeaders))
								})

								It("returns an error", func() {
									result, err := client.PutMetadata(ctx, id, condition, metadata)
									errorsTest.ExpectEqual(err, request.ErrorResourceNotFoundWithID(id))
									Expect(result).To(BeNil())
								})
							})

							When("the server responds with the result", func() {
								var responseResult *image.Image

								BeforeEach(func() {
									responseResult = imageTest.RandomImage()
									requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusOK, responseResult, responseHeaders))
								})

								It("returns successfully", func() {
									Expect(client.PutMetadata(ctx, id, condition, metadata)).To(imageTest.MatchImage(responseResult))
								})
							})
						})
					}

					When("condition is missing", func() {
						BeforeEach(func() {
							condition = nil
							requestHandlers = append(requestHandlers,
								VerifyRequest(http.MethodPut, fmt.Sprintf("/v1/images/%s/metadata", id)),
							)
						})

						conditionAssertions()
					})

					When("condition revision is missing", func() {
						BeforeEach(func() {
							condition.Revision = nil
							requestHandlers = append(requestHandlers,
								VerifyRequest(http.MethodPut, fmt.Sprintf("/v1/images/%s/metadata", id)),
							)
						})

						conditionAssertions()
					})

					When("condition revision is present", func() {
						BeforeEach(func() {
							query := url.Values{
								"revision": []string{strconv.Itoa(*condition.Revision)},
							}
							requestHandlers = append(requestHandlers,
								VerifyRequest(http.MethodPut, fmt.Sprintf("/v1/images/%s/metadata", id), query.Encode()),
							)
						})

						conditionAssertions()
					})
				})

				Context("PutContent", func() {
					var condition *request.Condition
					var contentIntent string

					BeforeEach(func() {
						condition = requestTest.RandomCondition()
					})

					contentIntentAssertions := func() {
						var body []byte
						var content *image.Content

						BeforeEach(func() {
							body = imageTest.RandomContentBytes()
							content = imageTest.RandomContent()
							content.Body = ioutil.NopCloser(bytes.NewReader(body))
						})

						Context("without server response", func() {
							AfterEach(func() {
								Expect(server.ReceivedRequests()).To(BeEmpty())
							})

							It("returns an error when the context is missing", func() {
								ctx = nil
								result, err := client.PutContent(ctx, id, condition, contentIntent, content)
								errorsTest.ExpectEqual(err, errors.New("context is missing"))
								Expect(result).To(BeNil())
							})

							It("returns an error when the id is missing", func() {
								id = ""
								result, err := client.PutContent(ctx, id, condition, contentIntent, content)
								errorsTest.ExpectEqual(err, errors.New("id is missing"))
								Expect(result).To(BeNil())
							})

							It("returns an error when the id is invalid", func() {
								id = "invalid"
								result, err := client.PutContent(ctx, id, condition, contentIntent, content)
								errorsTest.ExpectEqual(err, errors.New("id is invalid"))
								Expect(result).To(BeNil())
							})

							It("returns an error when the condition is invalid", func() {
								condition.Revision = pointer.FromInt(-1)
								result, err := client.PutContent(ctx, id, condition, contentIntent, content)
								errorsTest.ExpectEqual(err, errors.New("condition is invalid"))
								Expect(result).To(BeNil())
							})

							It("returns an error when the content intent is missing", func() {
								contentIntent = ""
								result, err := client.PutContent(ctx, id, condition, contentIntent, content)
								errorsTest.ExpectEqual(err, errors.New("content intent is missing"))
								Expect(result).To(BeNil())
							})

							It("returns an error when the content intent is invalid", func() {
								contentIntent = "invalid"
								result, err := client.PutContent(ctx, id, condition, contentIntent, content)
								errorsTest.ExpectEqual(err, errors.New("content intent is invalid"))
								Expect(result).To(BeNil())
							})

							It("returns an error when the content is missing", func() {
								content = nil
								result, err := client.PutContent(ctx, id, condition, contentIntent, content)
								errorsTest.ExpectEqual(err, errors.New("content is missing"))
								Expect(result).To(BeNil())
							})

							It("returns an error when the content is invalid", func() {
								content.Body = nil
								result, err := client.PutContent(ctx, id, condition, contentIntent, content)
								errorsTest.ExpectEqual(err, errors.New("content is invalid"))
								Expect(result).To(BeNil())
							})
						})

						conditionAssertions := func() {
							digestAssertions := func() {
								Context("with server response", func() {
									BeforeEach(func() {
										requestHandlers = append(requestHandlers,
											VerifyContentType(*content.MediaType),
											VerifyBody(body),
										)
									})

									AfterEach(func() {
										Expect(server.ReceivedRequests()).To(HaveLen(1))
									})

									When("the server responds with an unauthenticated error", func() {
										BeforeEach(func() {
											requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusUnauthorized, errors.NewSerializable(request.ErrorUnauthenticated()), responseHeaders))
										})

										It("returns an error", func() {
											result, err := client.PutContent(ctx, id, condition, contentIntent, content)
											errorsTest.ExpectEqual(err, request.ErrorUnauthenticated())
											Expect(result).To(BeNil())
										})
									})

									When("the server responds with an unauthorized error", func() {
										BeforeEach(func() {
											requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusForbidden, errors.NewSerializable(request.ErrorUnauthorized()), responseHeaders))
										})

										It("returns an error", func() {
											result, err := client.PutContent(ctx, id, condition, contentIntent, content)
											errorsTest.ExpectEqual(err, request.ErrorUnauthorized())
											Expect(result).To(BeNil())
										})
									})

									When("the server responds with a not found error", func() {
										BeforeEach(func() {
											requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusNotFound, errors.NewSerializable(request.ErrorResourceNotFoundWithID(id)), responseHeaders))
										})

										It("returns an error", func() {
											result, err := client.PutContent(ctx, id, condition, contentIntent, content)
											errorsTest.ExpectEqual(err, request.ErrorResourceNotFoundWithID(id))
											Expect(result).To(BeNil())
										})
									})

									When("the server responds with the result", func() {
										var responseResult *image.Image

										BeforeEach(func() {
											responseResult = imageTest.RandomImage()
											requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusOK, responseResult, responseHeaders))
										})

										It("returns successfully", func() {
											Expect(client.PutContent(ctx, id, condition, contentIntent, content)).To(imageTest.MatchImage(responseResult))
										})
									})
								})
							}

							When("the request has no digest header", func() {
								BeforeEach(func() {
									content.DigestMD5 = nil
								})

								digestAssertions()
							})

							When("the request has a digest header", func() {
								BeforeEach(func() {
									requestHandlers = append(requestHandlers, VerifyHeaderKV("Digest", fmt.Sprintf("MD5=%s", *content.DigestMD5)))
								})

								digestAssertions()
							})
						}

						When("condition is missing", func() {
							BeforeEach(func() {
								condition = nil
								requestHandlers = append(requestHandlers,
									VerifyRequest(http.MethodPut, fmt.Sprintf("/v1/images/%s/content/%s", id, contentIntent)),
								)
							})

							conditionAssertions()
						})

						When("condition revision is missing", func() {
							BeforeEach(func() {
								condition.Revision = nil
								requestHandlers = append(requestHandlers,
									VerifyRequest(http.MethodPut, fmt.Sprintf("/v1/images/%s/content/%s", id, contentIntent)),
								)
							})

							conditionAssertions()
						})

						When("condition revision is present", func() {
							BeforeEach(func() {
								query := url.Values{
									"revision": []string{strconv.Itoa(*condition.Revision)},
								}
								requestHandlers = append(requestHandlers,
									VerifyRequest(http.MethodPut, fmt.Sprintf("/v1/images/%s/content/%s", id, contentIntent), query.Encode()),
								)
							})

							conditionAssertions()
						})
					}

					When("the content intent is alternate", func() {
						BeforeEach(func() {
							contentIntent = image.ContentIntentAlternate
						})

						contentIntentAssertions()
					})

					When("the content intent is original", func() {
						BeforeEach(func() {
							contentIntent = image.ContentIntentOriginal
						})

						contentIntentAssertions()
					})
				})

				Context("Delete", func() {
					var condition *request.Condition

					BeforeEach(func() {
						condition = requestTest.RandomCondition()
					})

					Context("without server response", func() {
						AfterEach(func() {
							Expect(server.ReceivedRequests()).To(BeEmpty())
						})

						It("returns an error when the context is missing", func() {
							ctx = nil
							deleted, err := client.Delete(ctx, id, condition)
							errorsTest.ExpectEqual(err, errors.New("context is missing"))
							Expect(deleted).To(BeFalse())
						})

						It("returns an error when the id is missing", func() {
							id = ""
							deleted, err := client.Delete(ctx, id, condition)
							errorsTest.ExpectEqual(err, errors.New("id is missing"))
							Expect(deleted).To(BeFalse())
						})

						It("returns an error when the id is invalid", func() {
							id = "invalid"
							deleted, err := client.Delete(ctx, id, condition)
							errorsTest.ExpectEqual(err, errors.New("id is invalid"))
							Expect(deleted).To(BeFalse())
						})

						It("returns an error when the condition is invalid", func() {
							condition.Revision = pointer.FromInt(-1)
							deleted, err := client.Delete(ctx, id, condition)
							errorsTest.ExpectEqual(err, errors.New("condition is invalid"))
							Expect(deleted).To(BeFalse())
						})
					})

					conditionAssertions := func() {
						Context("with server response", func() {
							AfterEach(func() {
								Expect(server.ReceivedRequests()).To(HaveLen(1))
							})

							When("the server responds with an unauthenticated error", func() {
								BeforeEach(func() {
									requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusUnauthorized, errors.NewSerializable(request.ErrorUnauthenticated()), responseHeaders))
								})

								It("returns an error", func() {
									deleted, err := client.Delete(ctx, id, condition)
									errorsTest.ExpectEqual(err, request.ErrorUnauthenticated())
									Expect(deleted).To(BeFalse())
								})
							})

							When("the server responds with an unauthorized error", func() {
								BeforeEach(func() {
									requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusForbidden, errors.NewSerializable(request.ErrorUnauthorized()), responseHeaders))
								})

								It("returns an error", func() {
									deleted, err := client.Delete(ctx, id, condition)
									errorsTest.ExpectEqual(err, request.ErrorUnauthorized())
									Expect(deleted).To(BeFalse())
								})
							})

							When("the server responds with a not found error", func() {
								BeforeEach(func() {
									requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusNotFound, errors.NewSerializable(request.ErrorResourceNotFoundWithID(id)), responseHeaders))
								})

								It("returns successfully with deleted false", func() {
									deleted, err := client.Delete(ctx, id, condition)
									Expect(err).ToNot(HaveOccurred())
									Expect(deleted).To(BeFalse())
								})
							})

							When("the server responds successfully", func() {
								BeforeEach(func() {
									requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusNoContent, nil, responseHeaders))
								})

								It("returns successfully with delete true", func() {
									deleted, err := client.Delete(ctx, id, condition)
									Expect(err).ToNot(HaveOccurred())
									Expect(deleted).To(BeTrue())
								})
							})
						})
					}

					When("condition is missing", func() {
						BeforeEach(func() {
							condition = nil
							requestHandlers = append(requestHandlers,
								VerifyRequest(http.MethodDelete, fmt.Sprintf("/v1/images/%s", id)),
								VerifyContentType(""),
								VerifyBody(nil),
							)
						})

						conditionAssertions()
					})

					When("condition revision is missing", func() {
						BeforeEach(func() {
							condition.Revision = nil
							requestHandlers = append(requestHandlers,
								VerifyRequest(http.MethodDelete, fmt.Sprintf("/v1/images/%s", id)),
								VerifyContentType(""),
								VerifyBody(nil),
							)
						})

						conditionAssertions()
					})

					When("condition revision is present", func() {
						BeforeEach(func() {
							query := url.Values{
								"revision": []string{strconv.Itoa(*condition.Revision)},
							}
							requestHandlers = append(requestHandlers,
								VerifyRequest(http.MethodDelete, fmt.Sprintf("/v1/images/%s", id), query.Encode()),
								VerifyContentType(""),
								VerifyBody(nil),
							)
						})

						conditionAssertions()
					})
				})
			})
		}

		When("client must authorize as service", func() {
			BeforeEach(func() {
				config.ServiceSecret = authTest.NewServiceSecret()
				authorizeAs = platform.AuthorizeAsService
				requestHandlers = append(requestHandlers, VerifyHeaderKV("X-Tidepool-Service-Secret", config.ServiceSecret))
			})

			authorizeAssertions()
		})

		When("client must authorize as user", func() {
			BeforeEach(func() {
				sessionToken := authTest.NewSessionToken()
				authorizeAs = platform.AuthorizeAsUser
				requestHandlers = append(requestHandlers, VerifyHeaderKV("X-Tidepool-Session-Token", sessionToken))
				ctx = request.NewContextWithDetails(ctx, request.NewDetails(request.MethodAccessToken, userTest.RandomID(), sessionToken))
			})

			authorizeAssertions()
		})
	})
})
