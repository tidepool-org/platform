package client_test

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/platform"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	requestTest "github.com/tidepool-org/platform/request/test"
	"github.com/tidepool-org/platform/test"
	testHttp "github.com/tidepool-org/platform/test/http"
	"github.com/tidepool-org/platform/user"
	userClient "github.com/tidepool-org/platform/user/client"
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
			client, err := userClient.New(nil, authorizeAs)
			errorsTest.ExpectEqual(err, errors.New("config is missing"))
			Expect(client).To(BeNil())
		})

		It("returns an error when the authorize as is invalid", func() {
			authorizeAs = platform.AuthorizeAs(-1)
			client, err := userClient.New(config, authorizeAs)
			errorsTest.ExpectEqual(err, errors.New("authorize as is invalid"))
			Expect(client).To(BeNil())
		})

		It("returns success", func() {
			Expect(userClient.New(config, authorizeAs)).ToNot(BeNil())
		})
	})

	Context("with server and new client", func() {
		var server *Server
		var requestHandlers []http.HandlerFunc
		var responseHeaders http.Header
		var ctx context.Context
		var client user.Client

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
			client, err = userClient.New(config, authorizeAs)
			Expect(err).ToNot(HaveOccurred())
			Expect(client).ToNot(BeNil())
		})

		AfterEach(func() {
			if server != nil {
				server.Close()
			}
		})

		authorizeAssertions := func() {
			Context("with id", func() {
				var id string

				BeforeEach(func() {
					id = userTest.RandomID()
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
								VerifyRequest(http.MethodGet, fmt.Sprintf("/v1/users/%s", id)),
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
							var responseResult *user.User

							BeforeEach(func() {
								responseResult = userTest.RandomUser()
								requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusOK, responseResult, responseHeaders))
							})

							It("returns successfully with result", func() {
								Expect(client.Get(ctx, id)).To(userTest.MatchUser(responseResult))
							})
						})
					})
				})

				Context("Delete", func() {
					var deleet *user.Delete
					var condition *request.Condition

					BeforeEach(func() {
						deleet = userTest.RandomDelete()
						condition = requestTest.RandomCondition()
					})

					Context("without server response", func() {
						AfterEach(func() {
							Expect(server.ReceivedRequests()).To(BeEmpty())
						})

						It("returns an error when the context is missing", func() {
							ctx = nil
							deleted, err := client.Delete(ctx, id, deleet, condition)
							errorsTest.ExpectEqual(err, errors.New("context is missing"))
							Expect(deleted).To(BeFalse())
						})

						It("returns an error when the id is missing", func() {
							id = ""
							deleted, err := client.Delete(ctx, id, deleet, condition)
							errorsTest.ExpectEqual(err, errors.New("id is missing"))
							Expect(deleted).To(BeFalse())
						})

						It("returns an error when the id is invalid", func() {
							id = "invalid"
							deleted, err := client.Delete(ctx, id, deleet, condition)
							errorsTest.ExpectEqual(err, errors.New("id is invalid"))
							Expect(deleted).To(BeFalse())
						})

						It("returns an error when the delete is invalid", func() {
							deleet.Password = pointer.FromString("")
							deleted, err := client.Delete(ctx, id, deleet, condition)
							errorsTest.ExpectEqual(err, errors.New("delete is invalid"))
							Expect(deleted).To(BeFalse())
						})

						It("returns an error when the condition is invalid", func() {
							condition.Revision = pointer.FromInt(-1)
							deleted, err := client.Delete(ctx, id, deleet, condition)
							errorsTest.ExpectEqual(err, errors.New("condition is invalid"))
							Expect(deleted).To(BeFalse())
						})
					})

					conditionAssertions := func() {
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
										deleted, err := client.Delete(ctx, id, deleet, condition)
										errorsTest.ExpectEqual(err, request.ErrorUnauthenticated())
										Expect(deleted).To(BeFalse())
									})
								})

								When("the server responds with an unauthorized error", func() {
									BeforeEach(func() {
										requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusForbidden, errors.NewSerializable(request.ErrorUnauthorized()), responseHeaders))
									})

									It("returns an error", func() {
										deleted, err := client.Delete(ctx, id, deleet, condition)
										errorsTest.ExpectEqual(err, request.ErrorUnauthorized())
										Expect(deleted).To(BeFalse())
									})
								})

								When("the server responds with a not found error", func() {
									BeforeEach(func() {
										requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusNotFound, errors.NewSerializable(request.ErrorResourceNotFoundWithID(id)), responseHeaders))
									})

									It("returns successfully with deleted false", func() {
										deleted, err := client.Delete(ctx, id, deleet, condition)
										Expect(err).ToNot(HaveOccurred())
										Expect(deleted).To(BeFalse())
									})
								})

								When("the server responds successfully", func() {
									BeforeEach(func() {
										requestHandlers = append(requestHandlers, RespondWithJSONEncoded(http.StatusNoContent, nil, responseHeaders))
									})

									It("returns successfully with delete true", func() {
										deleted, err := client.Delete(ctx, id, deleet, condition)
										Expect(err).ToNot(HaveOccurred())
										Expect(deleted).To(BeTrue())
									})
								})
							})
						}

						When("delete is missing", func() {
							BeforeEach(func() {
								deleet = nil
								requestHandlers = append(requestHandlers,
									VerifyContentType(""),
									VerifyBody(nil),
								)
							})

							deleteAssertions()
						})

						When("delete password is missing", func() {
							BeforeEach(func() {
								deleet.Password = nil
								requestHandlers = append(requestHandlers,
									VerifyContentType("application/json; charset=utf-8"),
									VerifyBody(test.MarshalRequestBody(deleet)),
								)
							})

							deleteAssertions()
						})

						When("delete password is present", func() {
							BeforeEach(func() {
								requestHandlers = append(requestHandlers,
									VerifyContentType("application/json; charset=utf-8"),
									VerifyBody(test.MarshalRequestBody(deleet)),
								)
							})

							deleteAssertions()
						})
					}

					When("condition is missing", func() {
						BeforeEach(func() {
							condition = nil
							requestHandlers = append(requestHandlers, VerifyRequest(http.MethodDelete, fmt.Sprintf("/v1/users/%s", id)))
						})

						conditionAssertions()
					})

					When("condition revision is missing", func() {
						BeforeEach(func() {
							condition.Revision = nil
							requestHandlers = append(requestHandlers, VerifyRequest(http.MethodDelete, fmt.Sprintf("/v1/users/%s", id)))
						})

						conditionAssertions()
					})

					When("condition revision is present", func() {
						BeforeEach(func() {
							query := url.Values{
								"revision": []string{strconv.Itoa(*condition.Revision)},
							}
							requestHandlers = append(requestHandlers, VerifyRequest(http.MethodDelete, fmt.Sprintf("/v1/users/%s", id), query.Encode()))
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
