package client_test

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	oauthClient "github.com/tidepool-org/platform/oauth/client"
	oauthTest "github.com/tidepool-org/platform/oauth/test"
	"github.com/tidepool-org/platform/request"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
	testHTTP "github.com/tidepool-org/platform/test/http"
)

type RequestBody struct {
	Request string `json:"request"`
}

type ResponseBody struct {
	Response string `json:"response"`
}

var _ = Describe("Client", func() {
	var userAgent string
	var config *client.Config
	var tokenSourceSource *oauthTest.TokenSourceSource

	BeforeEach(func() {
		userAgent = testHTTP.NewUserAgent()
		config = client.NewConfig()
		config.UserAgent = userAgent
		tokenSourceSource = oauthTest.NewTokenSourceSource()
	})

	AfterEach(func() {
		tokenSourceSource.AssertOutputsEmpty()
	})

	Context("New", func() {
		BeforeEach(func() {
			config.Address = testHTTP.NewAddress()
		})

		It("returns an error when config is missing", func() {
			clnt, err := oauthClient.New(nil, tokenSourceSource)
			Expect(err).To(MatchError("config is missing"))
			Expect(clnt).To(BeNil())
		})

		It("returns an error when token source source is missing", func() {
			clnt, err := oauthClient.New(config, nil)
			Expect(err).To(MatchError("token source source is missing"))
			Expect(clnt).To(BeNil())
		})

		It("returns an error when config is invalid", func() {
			config.Address = ""
			clnt, err := oauthClient.New(config, tokenSourceSource)
			Expect(err).To(MatchError("config is invalid; address is missing"))
			Expect(clnt).To(BeNil())
		})

		It("returns successfully", func() {
			Expect(oauthClient.New(config, tokenSourceSource)).ToNot(BeNil())
		})
	})

	Context("with new client", func() {
		var address string
		var clnt *oauthClient.Client

		BeforeEach(func() {
			address = testHTTP.NewAddress()
			config.Address = address
		})

		JustBeforeEach(func() {
			var err error
			clnt, err = oauthClient.New(config, tokenSourceSource)
			Expect(err).ToNot(HaveOccurred())
			Expect(clnt).ToNot(BeNil())
		})

		Context("ConstructURL", func() {
			constructURLAssertions := func() {
				It("returns a valid URL with no paths", func() {
					Expect(clnt.ConstructURL()).To(Equal(address + "/"))
				})

				It("returns a valid URL with one path", func() {
					path := test.NewVariableString(1, 8, test.CharsetAlphaNumeric)
					Expect(clnt.ConstructURL(path)).To(Equal(fmt.Sprintf("%s/%s", address, path)))
				})

				It("returns a valid URL with multiple paths", func() {
					path1 := test.NewVariableString(1, 8, test.CharsetAlphaNumeric)
					path2 := test.NewVariableString(1, 8, test.CharsetAlphaNumeric)
					path3 := test.NewVariableString(1, 8, test.CharsetAlphaNumeric)
					Expect(clnt.ConstructURL(path1, path2, path3)).To(Equal(fmt.Sprintf("%s/%s/%s/%s", address, path1, path2, path3)))
				})

				It("returns a valid URL with multiple paths that need to be escaped", func() {
					path1 := test.NewVariableString(1, 4, test.CharsetAlphaNumeric) + test.NewVariableString(1, 4, " /;,?") + test.NewVariableString(1, 4, test.CharsetAlphaNumeric)
					path2 := test.NewVariableString(1, 4, test.CharsetAlphaNumeric) + test.NewVariableString(1, 4, " /;,?") + test.NewVariableString(1, 4, test.CharsetAlphaNumeric)
					Expect(clnt.ConstructURL(path1, path2)).To(Equal(fmt.Sprintf("%s/%s/%s", address, url.PathEscape(path1), url.PathEscape(path2))))
				})

				It("returns a valid URL with multiple paths with surrounding slashes", func() {
					path1 := test.NewVariableString(1, 8, test.CharsetAlphaNumeric)
					path2 := test.NewVariableString(1, 8, test.CharsetAlphaNumeric)
					Expect(clnt.ConstructURL("/"+path1+"/", "/"+path2+"/")).To(Equal(fmt.Sprintf("%s/%s/%s", address, path1, path2)))
				})
			}

			Context("without trailing slash on address", func() {
				constructURLAssertions()
			})

			Context("with trailing slashes on address", func() {
				BeforeEach(func() {
					config.Address += "///"
				})

				constructURLAssertions()
			})
		})

		Context("AppendURLQuery", func() {
			var urlString string

			JustBeforeEach(func() {
				urlString = clnt.ConstructURL(test.NewVariableString(1, 8, test.CharsetAlphaNumeric), test.NewVariableString(1, 8, test.CharsetAlphaNumeric))
			})

			It("returns a URL without change when the query is nil", func() {
				Expect(clnt.AppendURLQuery(urlString, nil)).To(Equal(urlString))
			})

			It("returns a URL without change when the query is empty", func() {
				Expect(clnt.AppendURLQuery(urlString, map[string]string{})).To(Equal(urlString))
			})

			It("returns a URL with associated query", func() {
				key1 := testHTTP.NewParameterKey()
				value1 := testHTTP.NewParameterValue()
				key2 := key1 + testHTTP.NewParameterKey()
				value2 := testHTTP.NewParameterValue()
				query := map[string]string{
					key1: value1,
					key2: value2,
				}
				Expect(clnt.AppendURLQuery(urlString, query)).To(Equal(fmt.Sprintf("%s?%s=%s&%s=%s", urlString, key1, value1, key2, value2)))
			})

			It("returns a URL with associated query even when it already has a query string", func() {
				urlString += "?" + testHTTP.NewParameterKey() + "=" + testHTTP.NewParameterValue()
				key1 := testHTTP.NewParameterKey()
				value1 := testHTTP.NewParameterValue()
				key2 := key1 + testHTTP.NewParameterKey()
				value2 := testHTTP.NewParameterValue()
				query := map[string]string{
					key1: value1,
					key2: value2,
				}
				Expect(clnt.AppendURLQuery(urlString, query)).To(Equal(fmt.Sprintf("%s&%s=%s&%s=%s", urlString, key1, value1, key2, value2)))
			})
		})
	})

	Context("with started server and new client", func() {
		var server *Server
		var responseHeaders http.Header
		var ctx context.Context
		var method string
		var path string
		var url string
		var headerMutator *request.HeaderMutator
		var parameterMutator *request.ParameterMutator
		var mutators []request.RequestMutator
		var requestString string
		var requestBody *RequestBody
		var responseString string
		var httpClientSource *oauthTest.HTTPClientSource
		var clnt *oauthClient.Client

		BeforeEach(func() {
			server = NewServer()
			responseHeaders = http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}
			ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
			method = testHTTP.NewMethod()
			path = testHTTP.NewPath()
			url = server.URL() + path
			headerMutator = request.NewHeaderMutator(testHTTP.NewHeaderKey(), testHTTP.NewHeaderValue())
			parameterMutator = request.NewParameterMutator(testHTTP.NewParameterKey(), testHTTP.NewParameterValue())
			mutators = []request.RequestMutator{headerMutator, parameterMutator}
			requestString = test.NewVariableString(0, 32, test.CharsetText)
			requestBody = &RequestBody{Request: requestString}
			responseString = test.NewVariableString(0, 32, test.CharsetText)
			httpClientSource = oauthTest.NewHTTPClientSource()
		})

		JustBeforeEach(func() {
			config.Address = server.URL()
			var err error
			clnt, err = oauthClient.New(config, tokenSourceSource)
			Expect(err).ToNot(HaveOccurred())
			Expect(clnt).ToNot(BeNil())
		})

		AfterEach(func() {
			if server != nil {
				server.Close()
			}
			httpClientSource.AssertOutputsEmpty()
		})

		Context("SendOAuthRequest", func() {
			var responseBody *ResponseBody

			BeforeEach(func() {
				responseBody = &ResponseBody{}
			})

			It("returns error when http client source is missing", func() {
				Expect(clnt.SendOAuthRequest(ctx, method, url, mutators, requestBody, responseBody, nil)).To(MatchError("http client source is missing"))
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			When("http client source is not missing", func() {
				AfterEach(func() {
					Expect(httpClientSource.HTTPClientInputs).To(Equal([]oauthTest.HTTPClientInput{{Context: ctx, TokenSourceSource: tokenSourceSource}}))
				})

				It("returns error when http client source returns an error", func() {
					responseErr := errorsTest.RandomError()
					httpClientSource.HTTPClientOutputs = []oauthTest.HTTPClientOutput{{HTTPClient: nil, Error: responseErr}}
					Expect(clnt.SendOAuthRequest(ctx, method, url, mutators, requestBody, responseBody, httpClientSource)).To(Equal(responseErr))
					Expect(server.ReceivedRequests()).To(BeEmpty())
				})

				When("http client source returns successfully", func() {
					var httpClient *http.Client

					BeforeEach(func() {
						httpClient = http.DefaultClient
						httpClientSource.HTTPClientOutputs = []oauthTest.HTTPClientOutput{{HTTPClient: httpClient, Error: nil}}
					})

					It("returns error when context is missing", func() {
						ctx = nil
						Expect(clnt.SendOAuthRequest(ctx, method, url, mutators, requestBody, responseBody, httpClientSource)).To(MatchError("context is missing"))
						Expect(server.ReceivedRequests()).To(BeEmpty())
					})

					It("returns error when method is missing", func() {
						Expect(clnt.SendOAuthRequest(ctx, "", url, mutators, requestBody, responseBody, httpClientSource)).To(MatchError("method is missing"))
						Expect(server.ReceivedRequests()).To(BeEmpty())
					})

					It("returns error when url is missing", func() {
						Expect(clnt.SendOAuthRequest(ctx, method, "", mutators, requestBody, responseBody, httpClientSource)).To(MatchError("url is missing"))
						Expect(server.ReceivedRequests()).To(BeEmpty())
					})

					It("returns error when the request object cannot be encoded", func() {
						invalidRequestBody := struct{ Func interface{} }{func() {}}
						Expect(clnt.SendOAuthRequest(ctx, method, url, mutators, invalidRequestBody, responseBody, httpClientSource).Error()).To(MatchRegexp("unable to serialize request to .*; json: unsupported type: func()"))
						Expect(server.ReceivedRequests()).To(BeEmpty())
					})

					It("returns error when mutator returns an error", func() {
						errorMutator := request.NewHeaderMutator("", "")
						invalidMutators := []request.RequestMutator{headerMutator, errorMutator, parameterMutator}
						Expect(clnt.SendOAuthRequest(ctx, method, url, invalidMutators, requestBody, responseBody, httpClientSource).Error()).To(MatchRegexp("unable to mutate request to .*; key is missing"))
						Expect(server.ReceivedRequests()).To(BeEmpty())
					})

					It("returns error when the server is not reachable", func() {
						server.Close()
						server = nil
						Expect(clnt.SendOAuthRequest(ctx, method, url, mutators, requestBody, responseBody, httpClientSource).Error()).To(MatchRegexp("unable to perform request to .*: connect: connection refused"))
					})

					Context("with a successful response and no request body", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
									VerifyBody(nil),
									RespondWith(http.StatusOK, test.MustBytes(test.MarshalResponseBody(&ResponseBody{Response: responseString})), responseHeaders),
								),
							)
						})

						It("returns success", func() {
							Expect(clnt.SendOAuthRequest(ctx, method, url, mutators, nil, responseBody, httpClientSource)).To(Succeed())
							Expect(server.ReceivedRequests()).To(HaveLen(1))
							Expect(responseBody).ToNot(BeNil())
							Expect(responseBody.Response).To(Equal(responseString))
						})
					})

					Context("with an bad request 400 without deserializable error body", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
									VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
									VerifyBody(test.MustBytes(test.MarshalRequestBody(requestBody))),
									RespondWith(http.StatusBadRequest, []byte{255, 255, 255}, responseHeaders),
								),
							)
						})

						It("returns an error", func() {
							err := clnt.SendOAuthRequest(ctx, method, url, mutators, requestBody, responseBody, httpClientSource)
							errorsTest.ExpectEqual(err, request.ErrorBadRequest())
							Expect(server.ReceivedRequests()).To(HaveLen(1))
						})
					})

					Context("with an bad request 400 with deserializable error body", func() {
						var responseErr error

						BeforeEach(func() {
							responseErr = errors.Append(structureValidator.ErrorValueNotEmpty(), structureValidator.ErrorValueBoolNotTrue(), structureValidator.ErrorValueIntNotOneOf(1, []int{0, 2, 4}))
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
									VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
									VerifyBody(test.MustBytes(test.MarshalRequestBody(requestBody))),
									RespondWithJSONEncoded(http.StatusBadRequest, errors.NewSerializable(responseErr), responseHeaders),
								),
							)
						})

						It("returns an error", func() {
							err := clnt.SendOAuthRequest(ctx, method, url, mutators, requestBody, responseBody, httpClientSource)
							errorsTest.ExpectEqual(err, responseErr)
							Expect(server.ReceivedRequests()).To(HaveLen(1))
						})
					})

					Context("with an unauthorized response 401", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
									VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
									VerifyBody(test.MustBytes(test.MarshalRequestBody(requestBody))),
									RespondWith(http.StatusUnauthorized, "NOT JSON", responseHeaders),
								),
							)
						})

						It("returns an error", func() {
							err := clnt.SendOAuthRequest(ctx, method, url, mutators, requestBody, responseBody, httpClientSource)
							errorsTest.ExpectEqual(err, request.ErrorUnauthenticated())
							Expect(server.ReceivedRequests()).To(HaveLen(1))
						})
					})

					Context("with an forbidden response 403", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
									VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
									VerifyBody(test.MustBytes(test.MarshalRequestBody(requestBody))),
									RespondWith(http.StatusForbidden, "NOT JSON", responseHeaders),
								),
							)
						})

						It("returns an error", func() {
							err := clnt.SendOAuthRequest(ctx, method, url, mutators, requestBody, responseBody, httpClientSource)
							errorsTest.ExpectEqual(err, request.ErrorUnauthorized())
							Expect(server.ReceivedRequests()).To(HaveLen(1))
						})
					})

					Context("with an resource not found 404 without deserializable error body", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
									VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
									VerifyBody(test.MustBytes(test.MarshalRequestBody(requestBody))),
									RespondWith(http.StatusNotFound, "NOT JSON", responseHeaders),
								),
							)
						})

						It("returns an error", func() {
							err := clnt.SendOAuthRequest(ctx, method, url, mutators, requestBody, responseBody, httpClientSource)
							errorsTest.ExpectEqual(err, request.ErrorResourceNotFound())
							Expect(server.ReceivedRequests()).To(HaveLen(1))
						})
					})

					Context("with an resource not found 404 with deserializable error body", func() {
						var responseErr error

						BeforeEach(func() {
							responseErr = request.ErrorResourceNotFoundWithID(test.RandomStringFromRangeAndCharset(1, 16, test.CharsetHexidecimalLowercase))
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
									VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
									VerifyBody(test.MustBytes(test.MarshalRequestBody(requestBody))),
									RespondWithJSONEncoded(http.StatusNotFound, errors.NewSerializable(responseErr), responseHeaders),
								),
							)
						})

						It("returns an error", func() {
							err := clnt.SendOAuthRequest(ctx, method, url, mutators, requestBody, responseBody, httpClientSource)
							errorsTest.ExpectEqual(err, responseErr)
							Expect(server.ReceivedRequests()).To(HaveLen(1))
						})
					})

					Context("with a too many requests response 429", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
									VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
									VerifyBody(test.MustBytes(test.MarshalRequestBody(requestBody))),
									RespondWith(http.StatusTooManyRequests, "NOT JSON", responseHeaders),
								),
							)
						})

						It("returns an error", func() {
							err := clnt.SendOAuthRequest(ctx, method, url, mutators, requestBody, responseBody, httpClientSource)
							errorsTest.ExpectEqual(err, request.ErrorTooManyRequests())
							Expect(server.ReceivedRequests()).To(HaveLen(1))
						})
					})

					Context("with an unexpected response 500 without deserializable error body", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
									VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
									VerifyBody(test.MustBytes(test.MarshalRequestBody(requestBody))),
									RespondWith(http.StatusInternalServerError, nil, responseHeaders),
								),
							)
						})

						It("returns an error", func() {
							err := clnt.SendOAuthRequest(ctx, method, url, mutators, requestBody, responseBody, httpClientSource)
							Expect(err).To(MatchError(fmt.Sprintf(`unexpected response status code 500 from %s "%s?%s=%s"`, method, url, parameterMutator.Key, parameterMutator.Value)))
							Expect(server.ReceivedRequests()).To(HaveLen(1))
						})
					})

					Context("with an unexpected response 500 with deserializable error body", func() {
						var responseErr error

						BeforeEach(func() {
							responseErr = errorsTest.RandomError()
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
									VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
									VerifyBody(test.MustBytes(test.MarshalRequestBody(requestBody))),
									RespondWithJSONEncoded(http.StatusInternalServerError, errors.NewSerializable(responseErr), responseHeaders),
								),
							)
						})

						It("returns an error", func() {
							err := clnt.SendOAuthRequest(ctx, method, url, mutators, requestBody, responseBody, httpClientSource)
							errorsTest.ExpectEqual(err, responseErr)
							Expect(server.ReceivedRequests()).To(HaveLen(1))
						})
					})

					Context("with an unparseable response", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
									VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
									VerifyBody(test.MustBytes(test.MarshalRequestBody(requestBody))),
									RespondWith(http.StatusOK, []byte("{\"response\":"), responseHeaders),
								),
							)
						})

						It("returns an error", func() {
							err := clnt.SendOAuthRequest(ctx, method, url, mutators, requestBody, responseBody, httpClientSource)
							Expect(err).To(MatchError("json is malformed; unexpected EOF"))
							Expect(server.ReceivedRequests()).To(HaveLen(1))
						})
					})

					Context("with a successful response 204 without parsing content", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
									VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
									VerifyBody(test.MustBytes(test.MarshalRequestBody(requestBody))),
									RespondWith(http.StatusNoContent, nil),
								),
							)
						})

						It("returns success", func() {
							Expect(clnt.SendOAuthRequest(ctx, method, url, mutators, requestBody, responseBody, httpClientSource)).To(Succeed())
							Expect(server.ReceivedRequests()).To(HaveLen(1))
							Expect(responseBody).ToNot(BeNil())
							Expect(responseBody.Response).To(BeEmpty())
						})
					})

					Context("with a successful response 205 without parsing content", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
									VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
									VerifyBody(test.MustBytes(test.MarshalRequestBody(requestBody))),
									RespondWith(http.StatusResetContent, nil),
								),
							)
						})

						It("returns success", func() {
							Expect(clnt.SendOAuthRequest(ctx, method, url, mutators, requestBody, responseBody, httpClientSource)).To(Succeed())
							Expect(server.ReceivedRequests()).To(HaveLen(1))
							Expect(responseBody).ToNot(BeNil())
							Expect(responseBody.Response).To(BeEmpty())
						})
					})

					Context("with a successful response and no request body", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
									VerifyBody(nil),
									RespondWith(http.StatusOK, test.MustBytes(test.MarshalResponseBody(&ResponseBody{Response: responseString})), responseHeaders),
								),
							)
						})

						It("returns success", func() {
							Expect(clnt.SendOAuthRequest(ctx, method, url, mutators, nil, responseBody, httpClientSource)).To(Succeed())
							Expect(server.ReceivedRequests()).To(HaveLen(1))
							Expect(responseBody).ToNot(BeNil())
							Expect(responseBody.Response).To(Equal(responseString))
						})
					})

					Context("with a successful response and request body reader", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
									VerifyBody([]byte(requestString)),
									RespondWith(http.StatusOK, test.MustBytes(test.MarshalResponseBody(&ResponseBody{Response: responseString})), responseHeaders),
								),
							)
						})

						It("returns success", func() {
							Expect(clnt.SendOAuthRequest(ctx, method, url, mutators, strings.NewReader(requestString), responseBody, httpClientSource)).To(Succeed())
							Expect(server.ReceivedRequests()).To(HaveLen(1))
							Expect(responseBody).ToNot(BeNil())
							Expect(responseBody.Response).To(Equal(responseString))
						})
					})

					Context("with a successful response and request body object", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
									VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
									VerifyBody(test.MustBytes(test.MarshalRequestBody(requestBody))),
									RespondWith(http.StatusOK, test.MustBytes(test.MarshalResponseBody(&ResponseBody{Response: responseString})), responseHeaders),
								),
							)
						})

						It("returns success", func() {
							Expect(clnt.SendOAuthRequest(ctx, method, url, mutators, requestBody, responseBody, httpClientSource)).To(Succeed())
							Expect(server.ReceivedRequests()).To(HaveLen(1))
							Expect(responseBody).ToNot(BeNil())
							Expect(responseBody.Response).To(Equal(responseString))
						})
					})

					Context("with a successful response and no response body object", func() {
						BeforeEach(func() {
							server.AppendHandlers(
								CombineHandlers(
									VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
									VerifyHeaderKV("User-Agent", userAgent),
									VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
									VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
									VerifyBody(test.MustBytes(test.MarshalRequestBody(requestBody))),
									RespondWith(http.StatusOK, test.MustBytes(test.MarshalResponseBody(&ResponseBody{Response: responseString})), responseHeaders),
								),
							)
						})

						It("returns success without parsing response body", func() {
							Expect(clnt.SendOAuthRequest(ctx, method, url, mutators, requestBody, nil, httpClientSource)).To(Succeed())
							Expect(server.ReceivedRequests()).To(HaveLen(1))
						})
					})
				})
			})
		})
	})
})
