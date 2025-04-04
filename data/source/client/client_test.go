package client_test

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	authTest "github.com/tidepool-org/platform/auth/test"
	dataSource "github.com/tidepool-org/platform/data/source"
	dataSourceClient "github.com/tidepool-org/platform/data/source/client"
	dataSourceTest "github.com/tidepool-org/platform/data/source/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
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
			client, err := dataSourceClient.New(nil, authorizeAs)
			errorsTest.ExpectEqual(err, errors.New("config is missing"))
			Expect(client).To(BeNil())
		})

		It("returns an error when the authorize as is invalid", func() {
			authorizeAs = platform.AuthorizeAs(-1)
			client, err := dataSourceClient.New(config, authorizeAs)
			errorsTest.ExpectEqual(err, errors.New("authorize as is invalid"))
			Expect(client).To(BeNil())
		})

		It("returns success", func() {
			Expect(dataSourceClient.New(config, authorizeAs)).ToNot(BeNil())
		})
	})

	Context("with server and new client", func() {
		var server *Server
		var requestHandlers []http.HandlerFunc
		var responseHeaders http.Header
		var logger *logTest.Logger
		var ctx context.Context
		var client dataSource.Client

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
			client, err = dataSourceClient.New(config, authorizeAs)
			Expect(err).ToNot(HaveOccurred())
			Expect(client).ToNot(BeNil())
		})

		AfterEach(func() {
			if server != nil {
				server.Close()
			}
		})

		clientAssertions := func() {
			Context("with user id", func() {
				var userID string

				BeforeEach(func() {
					userID = userTest.RandomID()
				})

				Context("List", func() {
					var filter *dataSource.Filter
					var pagination *page.Pagination

					listAssertions := func() {
						Context("without server response", func() {
							AfterEach(func() {
								Expect(server.ReceivedRequests()).To(BeEmpty())
							})

							It("returns an error when the context is missing", func() {
								ctx = nil
								result, err := client.List(ctx, filter, pagination)
								errorsTest.ExpectEqual(err, errors.New("context is missing"))
								Expect(result).To(BeNil())
							})

							It("returns an error when the user id is missing", func() {
								filter = dataSource.NewFilter()
								result, err := client.List(ctx, filter, pagination)
								errorsTest.ExpectEqual(err, errors.New("user id is missing"))
								Expect(result).To(BeNil())
							})

							It("returns an error when the user id is invalid", func() {
								filter = dataSource.NewFilter()
								filter.UserID = pointer.FromString("invalid")
								result, err := client.List(ctx, filter, pagination)
								errorsTest.ExpectEqual(err, errors.New("filter is invalid"))
								Expect(result).To(BeNil())
							})

							It("returns an error when the filter is invalid", func() {
								filter = dataSource.NewFilter()
								filter.ProviderType = pointer.FromStringArray([]string{""})
								result, err := client.List(ctx, filter, pagination)
								errorsTest.ExpectEqual(err, errors.New("filter is invalid"))
								Expect(result).To(BeNil())
							})

							It("returns an error when the pagination is invalid", func() {
								pagination = page.NewPagination()
								pagination.Page = -1
								result, err := client.List(ctx, filter, pagination)
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
									result, err := client.List(ctx, filter, pagination)
									errorsTest.ExpectEqual(err, request.ErrorUnauthenticated())
									Expect(result).To(BeNil())
								})
							})

							When("the server responds with an unauthorized error", func() {
								BeforeEach(func() {
									requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusForbidden, errors.NewSerializable(request.ErrorUnauthorized()), responseHeaders))
								})

								It("returns an error", func() {
									result, err := client.List(ctx, filter, pagination)
									errorsTest.ExpectEqual(err, request.ErrorUnauthorized())
									Expect(result).To(BeNil())
								})
							})

							When("the server responds with a not found error", func() {
								BeforeEach(func() {
									requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusNotFound, errors.NewSerializable(request.ErrorResourceNotFoundWithID(userID)), responseHeaders))
								})

								It("returns an error", func() {
									result, err := client.List(ctx, filter, pagination)
									errorsTest.ExpectEqual(err, request.ErrorResourceNotFoundWithID(userID))
									Expect(result).To(BeNil())
								})
							})

							When("the server responds with no result", func() {
								BeforeEach(func() {
									requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusOK, dataSource.SourceArray{}, responseHeaders))
								})

								It("returns successfully", func() {
									result, err := client.List(ctx, filter, pagination)
									Expect(err).ToNot(HaveOccurred())
									Expect(result).To(Equal(dataSource.SourceArray{}))
								})
							})

							When("the server responds with result", func() {
								var responseResult dataSource.SourceArray

								BeforeEach(func() {
									responseResult = dataSourceTest.RandomSourceArray(1, 4)
									requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusOK, responseResult, responseHeaders))
								})

								It("returns successfully", func() {
									Expect(client.List(ctx, filter, pagination)).To(dataSourceTest.MatchSourceArray(responseResult))
								})
							})
						})
					}

					When("the request has no pagination parameters", func() {
						BeforeEach(func() {
							filter = dataSource.NewFilter()
							filter.UserID = pointer.FromString(userID)
							pagination = nil
							requestHandlers = append(requestHandlers, VerifyRequest(http.MethodGet, fmt.Sprintf("/v1/users/%s/data_sources", userID), ""))
						})

						listAssertions()
					})

					When("the request has random filter and pagination parameters", func() {
						BeforeEach(func() {
							filter = dataSourceTest.RandomFilter()
							filter.UserID = pointer.FromString(userID)
							pagination = pageTest.RandomPagination()
							query := url.Values{
								"providerType":       *filter.ProviderType,
								"providerName":       *filter.ProviderName,
								"providerSessionId":  *filter.ProviderSessionID,
								"providerExternalId": *filter.ProviderExternalID,
								"state":              *filter.State,
								"page":               []string{strconv.Itoa(pagination.Page)},
								"size":               []string{strconv.Itoa(pagination.Size)},
							}
							requestHandlers = append(requestHandlers, VerifyRequest(http.MethodGet, fmt.Sprintf("/v1/users/%s/data_sources", userID), query.Encode()))
						})

						listAssertions()
					})
				})

				Context("Create", func() {
					var create *dataSource.Create

					BeforeEach(func() {
						create = dataSourceTest.RandomCreate()
					})

					Context("without server response", func() {
						AfterEach(func() {
							Expect(server.ReceivedRequests()).To(BeEmpty())
						})

						It("returns an error when the context is missing", func() {
							ctx = nil
							result, err := client.Create(ctx, userID, create)
							errorsTest.ExpectEqual(err, errors.New("context is missing"))
							Expect(result).To(BeNil())
						})

						It("returns an error when the user id is missing", func() {
							userID = ""
							result, err := client.Create(ctx, userID, create)
							errorsTest.ExpectEqual(err, errors.New("user id is missing"))
							Expect(result).To(BeNil())
						})

						It("returns an error when the user id is invalid", func() {
							userID = "invalid"
							result, err := client.Create(ctx, userID, create)
							errorsTest.ExpectEqual(err, errors.New("user id is invalid"))
							Expect(result).To(BeNil())
						})

						It("returns an error when the create is missing", func() {
							create = nil
							result, err := client.Create(ctx, userID, create)
							errorsTest.ExpectEqual(err, errors.New("create is missing"))
							Expect(result).To(BeNil())
						})

						It("returns an error when the create is invalid", func() {
							create.ProviderType = pointer.FromString("")
							result, err := client.Create(ctx, userID, create)
							errorsTest.ExpectEqual(err, errors.New("create is invalid"))
							Expect(result).To(BeNil())
						})
					})

					Context("with server response", func() {
						BeforeEach(func() {
							requestHandlers = append(requestHandlers,
								VerifyRequest("POST", fmt.Sprintf("/v1/users/%s/data_sources", userID)),
								VerifyContentType("application/json; charset=utf-8"),
								VerifyBody(test.MarshalRequestBody(create)),
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
								result, err := client.Create(ctx, userID, create)
								errorsTest.ExpectEqual(err, request.ErrorUnauthenticated())
								Expect(result).To(BeNil())
							})
						})

						When("the server responds with an unauthorized error", func() {
							BeforeEach(func() {
								requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusForbidden, errors.NewSerializable(request.ErrorUnauthorized()), responseHeaders))
							})

							It("returns an error", func() {
								result, err := client.Create(ctx, userID, create)
								errorsTest.ExpectEqual(err, request.ErrorUnauthorized())
								Expect(result).To(BeNil())
							})
						})

						When("the server responds with a not found error", func() {
							BeforeEach(func() {
								requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusNotFound, errors.NewSerializable(request.ErrorResourceNotFoundWithID(userID)), responseHeaders))
							})

							It("returns an error", func() {
								result, err := client.Create(ctx, userID, create)
								errorsTest.ExpectEqual(err, request.ErrorResourceNotFoundWithID(userID))
								Expect(result).To(BeNil())
							})
						})

						When("the server responds with the result", func() {
							var responseResult *dataSource.Source

							BeforeEach(func() {
								responseResult = dataSourceTest.RandomSource()
								requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusOK, responseResult, responseHeaders))
							})

							It("returns successfully", func() {
								Expect(client.Create(ctx, userID, create)).To(dataSourceTest.MatchSource(responseResult))
							})
						})
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
								VerifyRequest(http.MethodDelete, fmt.Sprintf("/v1/users/%s/data_sources", userID)),
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
					id = dataSourceTest.RandomID()
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
								VerifyRequest(http.MethodGet, fmt.Sprintf("/v1/data_sources/%s", id)),
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
							var responseResult *dataSource.Source

							BeforeEach(func() {
								responseResult = dataSourceTest.RandomSource()
								requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusOK, responseResult, responseHeaders))
							})

							It("returns successfully with result", func() {
								Expect(client.Get(ctx, id)).To(dataSourceTest.MatchSource(responseResult))
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

					Context("without server response", func() {
						AfterEach(func() {
							Expect(server.ReceivedRequests()).To(BeEmpty())
						})

						It("returns an error when the context is missing", func() {
							ctx = nil
							result, err := client.Update(ctx, id, condition, update)
							errorsTest.ExpectEqual(err, errors.New("context is missing"))
							Expect(result).To(BeNil())
						})

						It("returns an error when the id is missing", func() {
							id = ""
							result, err := client.Update(ctx, id, condition, update)
							errorsTest.ExpectEqual(err, errors.New("id is missing"))
							Expect(result).To(BeNil())
						})

						It("returns an error when the id is invalid", func() {
							id = "invalid"
							result, err := client.Update(ctx, id, condition, update)
							errorsTest.ExpectEqual(err, errors.New("id is invalid"))
							Expect(result).To(BeNil())
						})

						It("returns an error when the condition is invalid", func() {
							condition.Revision = pointer.FromInt(-1)
							result, err := client.Update(ctx, id, condition, update)
							errorsTest.ExpectEqual(err, errors.New("condition is invalid"))
							Expect(result).To(BeNil())
						})

						It("returns an error when the update is missing", func() {
							update = nil
							result, err := client.Update(ctx, id, condition, update)
							errorsTest.ExpectEqual(err, errors.New("update is missing"))
							Expect(result).To(BeNil())
						})

						It("returns an error when the update is invalid", func() {
							update.State = pointer.FromString("")
							result, err := client.Update(ctx, id, condition, update)
							errorsTest.ExpectEqual(err, errors.New("update is invalid"))
							Expect(result).To(BeNil())
						})
					})

					updateAssertions := func() {
						Context("with server response", func() {
							AfterEach(func() {
								Expect(server.ReceivedRequests()).To(HaveLen(1))
							})

							When("the server responds with an unauthenticated error", func() {
								BeforeEach(func() {
									requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusUnauthorized, errors.NewSerializable(request.ErrorUnauthenticated()), responseHeaders))
								})

								It("returns an error", func() {
									result, err := client.Update(ctx, id, condition, update)
									errorsTest.ExpectEqual(err, request.ErrorUnauthenticated())
									Expect(result).To(BeNil())
								})
							})

							When("the server responds with an unauthorized error", func() {
								BeforeEach(func() {
									requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusForbidden, errors.NewSerializable(request.ErrorUnauthorized()), responseHeaders))
								})

								It("returns an error", func() {
									result, err := client.Update(ctx, id, condition, update)
									errorsTest.ExpectEqual(err, request.ErrorUnauthorized())
									Expect(result).To(BeNil())
								})
							})

							When("the server responds with a not found error", func() {
								BeforeEach(func() {
									requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusNotFound, errors.NewSerializable(request.ErrorResourceNotFoundWithID(id)), responseHeaders))
								})

								It("returns successfully without result", func() {
									result, err := client.Update(ctx, id, condition, update)
									Expect(err).ToNot(HaveOccurred())
									Expect(result).To(BeNil())
								})
							})

							When("the server responds with the result", func() {
								var responseResult *dataSource.Source

								BeforeEach(func() {
									responseResult = dataSourceTest.RandomSource()
									requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusOK, responseResult, responseHeaders))
								})

								It("returns successfully", func() {
									Expect(client.Update(ctx, id, condition, update)).To(dataSourceTest.MatchSource(responseResult))
								})
							})
						})
					}

					When("condition is missing", func() {
						BeforeEach(func() {
							condition = nil
							requestHandlers = append(requestHandlers,
								VerifyRequest(http.MethodPut, fmt.Sprintf("/v1/data_sources/%s", id)),
								VerifyContentType("application/json; charset=utf-8"),
								VerifyBody(test.MarshalRequestBody(update)),
							)
						})

						updateAssertions()
					})

					When("condition revision is missing", func() {
						BeforeEach(func() {
							condition.Revision = nil
							requestHandlers = append(requestHandlers,
								VerifyRequest(http.MethodPut, fmt.Sprintf("/v1/data_sources/%s", id)),
								VerifyContentType("application/json; charset=utf-8"),
								VerifyBody(test.MarshalRequestBody(update)),
							)
						})

						updateAssertions()
					})

					When("condition revision is present", func() {
						BeforeEach(func() {
							query := url.Values{
								"revision": []string{strconv.Itoa(*condition.Revision)},
							}
							requestHandlers = append(requestHandlers,
								VerifyRequest(http.MethodPut, fmt.Sprintf("/v1/data_sources/%s", id), query.Encode()),
								VerifyContentType("application/json; charset=utf-8"),
								VerifyBody(test.MarshalRequestBody(update)),
							)
						})

						updateAssertions()
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

					deleteAssertions := func() {
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
								VerifyRequest("DELETE", fmt.Sprintf("/v1/data_sources/%s", id)),
								VerifyContentType(""),
								VerifyBody(nil),
							)
						})

						deleteAssertions()
					})

					When("condition revision is missing", func() {
						BeforeEach(func() {
							condition.Revision = nil
							requestHandlers = append(requestHandlers,
								VerifyRequest("DELETE", fmt.Sprintf("/v1/data_sources/%s", id)),
								VerifyContentType(""),
								VerifyBody(nil),
							)
						})

						deleteAssertions()
					})

					When("condition revision is present", func() {
						BeforeEach(func() {
							query := url.Values{
								"revision": []string{strconv.Itoa(*condition.Revision)},
							}
							requestHandlers = append(requestHandlers,
								VerifyRequest("DELETE", fmt.Sprintf("/v1/data_sources/%s", id), query.Encode()),
								VerifyContentType(""),
								VerifyBody(nil),
							)
						})

						deleteAssertions()
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

			clientAssertions()
		})

		When("client must authorize as user", func() {
			BeforeEach(func() {
				sessionToken := authTest.NewSessionToken()
				authorizeAs = platform.AuthorizeAsUser
				requestHandlers = append(requestHandlers, VerifyHeaderKV("X-Tidepool-Session-Token", sessionToken))
				ctx = request.NewContextWithAuthDetails(ctx, request.NewAuthDetails(request.MethodAccessToken, userTest.RandomID(), sessionToken))
			})

			clientAssertions()
		})
	})
})
