package client_test

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	"go.uber.org/mock/gomock"

	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/blob"
	blobClient "github.com/tidepool-org/platform/blob/client"
	blobTest "github.com/tidepool-org/platform/blob/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
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

	BeforeEach(func() {
		config = platform.NewConfig()
		config.UserAgent = testHttp.NewUserAgent()
	})

	Context("New", func() {
		BeforeEach(func() {
			config.Address = testHttp.NewAddress()
			authorizeAs = platform.AuthorizeAsService
		})

		It("returns an error when the config is missing", func() {
			config = nil
			client, err := blobClient.New(nil, authorizeAs)
			errorsTest.ExpectEqual(err, errors.New("config is missing"))
			Expect(client).To(BeNil())
		})

		It("returns an error when the authorize as is invalid", func() {
			authorizeAs = platform.AuthorizeAs(-1)
			client, err := blobClient.New(config, authorizeAs)
			errorsTest.ExpectEqual(err, errors.New("authorize as is invalid"))
			Expect(client).To(BeNil())
		})

		It("returns success", func() {
			Expect(blobClient.New(config, authorizeAs)).ToNot(BeNil())
		})
	})

	Context("with server and new client", func() {
		var server *Server
		var requestHandlers []http.HandlerFunc
		var responseHeaders http.Header
		var logger *logTest.Logger
		var ctx context.Context
		var client blob.Client

		BeforeEach(func() {
			server = NewServer()
			requestHandlers = nil
			responseHeaders = http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}
			logger = logTest.NewLogger()
			ctx = log.NewContextWithLogger(context.Background(), logger)
		})

		JustBeforeEach(func() {
			server.AppendHandlers(CombineHandlers(requestHandlers...))
			config.Address = server.URL()
			var err error
			client, err = blobClient.New(config, authorizeAs)
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
					userID = userTest.RandomUserID()
				})

				Context("List", func() {
					var filter *blob.Filter
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
								filter = blob.NewFilter()
								filter.MediaType = pointer.FromStringArray([]string{""})
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
									requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusOK, blob.BlobArray{}, responseHeaders))
								})

								It("returns successfully", func() {
									result, err := client.List(ctx, userID, filter, pagination)
									Expect(err).ToNot(HaveOccurred())
									Expect(result).To(Equal(blob.BlobArray{}))
								})
							})

							When("the server responds with result", func() {
								var responseResult blob.BlobArray

								BeforeEach(func() {
									responseResult = blobTest.RandomBlobArray(1, 4)
									requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusOK, responseResult, responseHeaders))
								})

								It("returns successfully", func() {
									Expect(client.List(ctx, userID, filter, pagination)).To(blobTest.MatchBlobArray(responseResult))
								})
							})
						})
					}

					When("the request has no filter or pagination parameters", func() {
						BeforeEach(func() {
							filter = nil
							pagination = nil
							requestHandlers = append(requestHandlers, VerifyRequest(http.MethodGet, fmt.Sprintf("/v1/users/%s/blobs", userID), ""))
						})

						parameterAssertions()
					})

					When("the request has random filter and pagination parameters", func() {
						BeforeEach(func() {
							filter = blobTest.RandomFilter()
							pagination = pageTest.RandomPagination()
							query := url.Values{
								"mediaType": *filter.MediaType,
								"status":    *filter.Status,
								"page":      []string{strconv.Itoa(pagination.Page)},
								"size":      []string{strconv.Itoa(pagination.Size)},
							}
							requestHandlers = append(requestHandlers, VerifyRequest(http.MethodGet, fmt.Sprintf("/v1/users/%s/blobs", userID), query.Encode()))
						})

						parameterAssertions()
					})
				})

				Context("Create", func() {
					var body []byte
					var content *blob.Content

					BeforeEach(func() {
						body = test.RandomBytes()
						content = blobTest.RandomContent()
						content.Body = ioutil.NopCloser(bytes.NewReader(body))
					})

					Context("without server response", func() {
						AfterEach(func() {
							Expect(server.ReceivedRequests()).To(BeEmpty())
						})

						It("returns an error when the context is missing", func() {
							ctx = nil
							result, err := client.Create(ctx, userID, content)
							errorsTest.ExpectEqual(err, errors.New("context is missing"))
							Expect(result).To(BeNil())
						})

						It("returns an error when the user id is missing", func() {
							userID = ""
							result, err := client.Create(ctx, userID, content)
							errorsTest.ExpectEqual(err, errors.New("user id is missing"))
							Expect(result).To(BeNil())
						})

						It("returns an error when the user id is invalid", func() {
							userID = "invalid"
							result, err := client.Create(ctx, userID, content)
							errorsTest.ExpectEqual(err, errors.New("user id is invalid"))
							Expect(result).To(BeNil())
						})

						It("returns an error when the content is missing", func() {
							content = nil
							result, err := client.Create(ctx, userID, content)
							errorsTest.ExpectEqual(err, errors.New("content is missing"))
							Expect(result).To(BeNil())
						})

						It("returns an error when the content is invalid", func() {
							content.Body = nil
							result, err := client.Create(ctx, userID, content)
							errorsTest.ExpectEqual(err, errors.New("content is invalid"))
							Expect(result).To(BeNil())
						})
					})

					digestAssertions := func() {
						Context("with server response", func() {
							BeforeEach(func() {
								requestHandlers = append(requestHandlers,
									VerifyRequest(http.MethodPost, fmt.Sprintf("/v1/users/%s/blobs", userID)),
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
									result, err := client.Create(ctx, userID, content)
									errorsTest.ExpectEqual(err, request.ErrorUnauthenticated())
									Expect(result).To(BeNil())
								})
							})

							When("the server responds with an unauthorized error", func() {
								BeforeEach(func() {
									requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusForbidden, errors.NewSerializable(request.ErrorUnauthorized()), responseHeaders))
								})

								It("returns an error", func() {
									result, err := client.Create(ctx, userID, content)
									errorsTest.ExpectEqual(err, request.ErrorUnauthorized())
									Expect(result).To(BeNil())
								})
							})

							When("the server responds with a not found error", func() {
								BeforeEach(func() {
									requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusNotFound, errors.NewSerializable(request.ErrorResourceNotFoundWithID(userID)), responseHeaders))
								})

								It("returns an error", func() {
									result, err := client.Create(ctx, userID, content)
									errorsTest.ExpectEqual(err, request.ErrorResourceNotFoundWithID(userID))
									Expect(result).To(BeNil())
								})
							})

							When("the server responds with the result", func() {
								var responseResult *blob.Blob

								BeforeEach(func() {
									responseResult = blobTest.RandomBlob()
									requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusOK, responseResult, responseHeaders))
								})

								It("returns successfully", func() {
									Expect(client.Create(ctx, userID, content)).To(blobTest.MatchBlob(responseResult))
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
								VerifyRequest(http.MethodDelete, fmt.Sprintf("/v1/users/%s/blobs", userID)),
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
					id = blobTest.RandomBlobID()
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
								VerifyRequest(http.MethodGet, fmt.Sprintf("/v1/blobs/%s", id)),
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
							var responseResult *blob.Blob

							BeforeEach(func() {
								responseResult = blobTest.RandomBlob()
								requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusOK, responseResult, responseHeaders))
							})

							It("returns successfully with result", func() {
								Expect(client.Get(ctx, id)).To(blobTest.MatchBlob(responseResult))
							})
						})
					})
				})

				Context("GetContent", func() {
					Context("without server response", func() {
						AfterEach(func() {
							Expect(server.ReceivedRequests()).To(BeEmpty())
						})

						It("returns an error when the context is missing", func() {
							ctx = nil
							result, err := client.GetContent(ctx, id)
							errorsTest.ExpectEqual(err, errors.New("context is missing"))
							Expect(result).To(BeNil())
						})

						It("returns an error when the id is missing", func() {
							id = ""
							result, err := client.GetContent(ctx, id)
							errorsTest.ExpectEqual(err, errors.New("id is missing"))
							Expect(result).To(BeNil())
						})

						It("returns an error when the id is invalid", func() {
							id = "invalid"
							result, err := client.GetContent(ctx, id)
							errorsTest.ExpectEqual(err, errors.New("id is invalid"))
							Expect(result).To(BeNil())
						})
					})

					Context("with server response", func() {
						BeforeEach(func() {
							requestHandlers = append(requestHandlers,
								VerifyRequest(http.MethodGet, fmt.Sprintf("/v1/blobs/%s/content", id)),
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
								result, err := client.GetContent(ctx, id)
								errorsTest.ExpectEqual(err, request.ErrorUnauthenticated())
								Expect(result).To(BeNil())
							})
						})

						When("the server responds with an unauthorized error", func() {
							BeforeEach(func() {
								requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusForbidden, errors.NewSerializable(request.ErrorUnauthorized()), responseHeaders))
							})

							It("returns an error", func() {
								result, err := client.GetContent(ctx, id)
								errorsTest.ExpectEqual(err, request.ErrorUnauthorized())
								Expect(result).To(BeNil())
							})
						})

						When("the server responds with a not found error", func() {
							BeforeEach(func() {
								requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusNotFound, errors.NewSerializable(request.ErrorResourceNotFoundWithID(id)), responseHeaders))
							})

							It("returns successfully without result", func() {
								result, err := client.GetContent(ctx, id)
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
								result, err := client.GetContent(ctx, id)
								errorsTest.ExpectEqual(err, request.ErrorHeaderInvalid("Digest"))
								Expect(result).To(BeNil())
							})
						})

						When("the server responds with an invalid content type header", func() {
							var digestMD5 string

							BeforeEach(func() {
								digestMD5 = netTest.RandomDigestMD5()
								responseHeaders = http.Header{
									"Digest":       []string{fmt.Sprintf("MD5=%s", digestMD5)},
									"Content-Type": []string{"/"},
								}
								requestHandlers = append(requestHandlers, RespondWith(http.StatusOK, nil, responseHeaders))
							})

							It("returns successfully", func() {
								result, err := client.GetContent(ctx, id)
								errorsTest.ExpectEqual(err, request.ErrorHeaderInvalid("Content-Type"))
								Expect(result).To(BeNil())
							})
						})

						When("the server responds with the content", func() {
							var body []byte
							var digestMD5 string
							var mediaType string

							BeforeEach(func() {
								body = test.RandomBytes()
								digestMD5 = netTest.RandomDigestMD5()
								mediaType = netTest.RandomMediaType()
								responseHeaders = http.Header{
									"Digest":       []string{fmt.Sprintf("MD5=%s", digestMD5)},
									"Content-Type": []string{mediaType},
								}
								requestHandlers = append(requestHandlers, RespondWith(http.StatusOK, body, responseHeaders))
							})

							It("returns successfully", func() {
								result, err := client.GetContent(ctx, id)
								Expect(err).ToNot(HaveOccurred())
								Expect(result).ToNot(BeNil())
								Expect(result.Body).ToNot(BeNil())
								defer result.Body.Close()
								Expect(*result.DigestMD5).To(Equal(digestMD5))
								Expect(*result.MediaType).To(Equal(mediaType))
							})
						})
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
								VerifyRequest(http.MethodDelete, fmt.Sprintf("/v1/blobs/%s", id)),
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
								VerifyRequest(http.MethodDelete, fmt.Sprintf("/v1/blobs/%s", id)),
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
								VerifyRequest(http.MethodDelete, fmt.Sprintf("/v1/blobs/%s", id), query.Encode()),
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
			var serverSessionTokenProviderController *gomock.Controller
			var serverSessionTokenProvider *authTest.MockServerSessionTokenProvider

			BeforeEach(func() {
				serverSessionTokenProviderController = gomock.NewController(GinkgoT())
				serverSessionTokenProvider = authTest.NewMockServerSessionTokenProvider(serverSessionTokenProviderController)
				sessionToken := authTest.NewSessionToken()
				serverSessionTokenProvider.EXPECT().ServerSessionToken().Return(sessionToken, nil).AnyTimes()
				authorizeAs = platform.AuthorizeAsUser
				requestHandlers = append(requestHandlers, VerifyHeaderKV("X-Tidepool-Session-Token", sessionToken))
				ctx = request.NewContextWithAuthDetails(ctx, request.NewAuthDetails(request.MethodAccessToken, userTest.RandomUserID(), sessionToken))
			})

			AfterEach(func() {
				serverSessionTokenProviderController.Finish()
			})

			authorizeAssertions()
		})
	})
})
