package client_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/request"
	requestTest "github.com/tidepool-org/platform/request/test"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
	testHttp "github.com/tidepool-org/platform/test/http"
)

type RequestBody struct {
	Request string `json:"request"`
}

type ResponseBody struct {
	Response string `json:"response"`
}

var _ = Describe("Client", func() {
	var config *client.Config
	var errorResponseParser client.ErrorResponseParser

	BeforeEach(func() {
		config = client.NewConfig()
		errorResponseParser = client.NewSerializableErrorResponseParser()
	})

	Context("New", func() {
		BeforeEach(func() {
			config.Address = testHttp.NewAddress()
		})

		It("returns an error if config is missing", func() {
			clnt, err := client.New(nil)
			Expect(err).To(MatchError("config is missing"))
			Expect(clnt).To(BeNil())
		})

		It("returns an error if config is invalid", func() {
			config.Address = ""
			clnt, err := client.New(config)
			Expect(err).To(MatchError("config is invalid; address is missing"))
			Expect(clnt).To(BeNil())
		})

		It("returns successfully", func() {
			Expect(client.New(config)).ToNot(BeNil())
		})
	})

	Context("NewWithErrorParser", func() {
		BeforeEach(func() {
			config.Address = testHttp.NewAddress()
		})

		It("returns an error if config is missing", func() {
			clnt, err := client.NewWithErrorParser(nil, errorResponseParser)
			Expect(err).To(MatchError("config is missing"))
			Expect(clnt).To(BeNil())
		})

		It("returns an error if config is invalid", func() {
			config.Address = ""
			clnt, err := client.NewWithErrorParser(config, errorResponseParser)
			Expect(err).To(MatchError("config is invalid; address is missing"))
			Expect(clnt).To(BeNil())
		})

		It("returns successfully", func() {
			Expect(client.NewWithErrorParser(config, errorResponseParser)).ToNot(BeNil())
		})

		It("returns successfully without an error response parser", func() {
			Expect(client.NewWithErrorParser(config, nil)).ToNot(BeNil())
		})
	})

	Context("with new client", func() {
		var address string
		var clnt *client.Client

		BeforeEach(func() {
			address = testHttp.NewAddress()
			config.Address = address
		})

		JustBeforeEach(func() {
			var err error
			clnt, err = client.NewWithErrorParser(config, errorResponseParser)
			Expect(err).ToNot(HaveOccurred())
			Expect(clnt).ToNot(BeNil())
		})

		Context("ConstructURL", func() {
			constructURLAssertions := func() {
				It("returns a valid URL with no paths", func() {
					Expect(clnt.ConstructURL()).To(Equal(address + "/"))
				})

				It("returns a valid URL with one path", func() {
					path := test.RandomStringFromRangeAndCharset(1, 8, test.CharsetAlphaNumeric)
					Expect(clnt.ConstructURL(path)).To(Equal(fmt.Sprintf("%s/%s", address, path)))
				})

				It("returns a valid URL with multiple paths", func() {
					path1 := test.RandomStringFromRangeAndCharset(1, 8, test.CharsetAlphaNumeric)
					path2 := test.RandomStringFromRangeAndCharset(1, 8, test.CharsetAlphaNumeric)
					path3 := test.RandomStringFromRangeAndCharset(1, 8, test.CharsetAlphaNumeric)
					Expect(clnt.ConstructURL(path1, path2, path3)).To(Equal(fmt.Sprintf("%s/%s/%s/%s", address, path1, path2, path3)))
				})

				It("returns a valid URL with multiple paths that need to be escaped", func() {
					path1 := test.RandomStringFromRangeAndCharset(1, 4, test.CharsetAlphaNumeric) + test.RandomStringFromRangeAndCharset(1, 4, " /;,?") + test.RandomStringFromRangeAndCharset(1, 4, test.CharsetAlphaNumeric)
					path2 := test.RandomStringFromRangeAndCharset(1, 4, test.CharsetAlphaNumeric) + test.RandomStringFromRangeAndCharset(1, 4, " /;,?") + test.RandomStringFromRangeAndCharset(1, 4, test.CharsetAlphaNumeric)
					Expect(clnt.ConstructURL(path1, path2)).To(Equal(fmt.Sprintf("%s/%s/%s", address, url.PathEscape(path1), url.PathEscape(path2))))
				})

				It("returns a valid URL with multiple paths with surrounding slashes", func() {
					path1 := test.RandomStringFromRangeAndCharset(1, 8, test.CharsetAlphaNumeric)
					path2 := test.RandomStringFromRangeAndCharset(1, 8, test.CharsetAlphaNumeric)
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
				urlString = clnt.ConstructURL(test.RandomStringFromRangeAndCharset(1, 8, test.CharsetAlphaNumeric), test.RandomStringFromRangeAndCharset(1, 8, test.CharsetAlphaNumeric))
			})

			It("returns a URL without change if the query is nil", func() {
				Expect(clnt.AppendURLQuery(urlString, nil)).To(Equal(urlString))
			})

			It("returns a URL without change if the query is empty", func() {
				Expect(clnt.AppendURLQuery(urlString, map[string]string{})).To(Equal(urlString))
			})

			It("returns a URL with associated query", func() {
				key1 := testHttp.NewParameterKey()
				value1 := testHttp.NewParameterValue()
				key2 := key1 + testHttp.NewParameterKey()
				value2 := testHttp.NewParameterValue()
				query := map[string]string{
					key1: value1,
					key2: value2,
				}
				Expect(clnt.AppendURLQuery(urlString, query)).To(Equal(fmt.Sprintf("%s?%s=%s&%s=%s", urlString, key1, value1, key2, value2)))
			})

			It("returns a URL with associated query even if it already has a query string", func() {
				urlString += "?" + testHttp.NewParameterKey() + "=" + testHttp.NewParameterValue()
				key1 := testHttp.NewParameterKey()
				value1 := testHttp.NewParameterValue()
				key2 := key1 + testHttp.NewParameterKey()
				value2 := testHttp.NewParameterValue()
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
		var inspectors []request.ResponseInspector
		var httpClient *http.Client
		var clnt *client.Client

		BeforeEach(func() {
			server = NewServer()
			responseHeaders = http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}
			ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
			method = testHttp.NewMethod()
			path = testHttp.NewPath()
			url = server.URL() + path
			headerMutator = request.NewHeaderMutator(testHttp.NewHeaderKey(), testHttp.NewHeaderValue())
			parameterMutator = request.NewParameterMutator(testHttp.NewParameterKey(), testHttp.NewParameterValue())
			mutators = []request.RequestMutator{headerMutator, parameterMutator}
			requestString = test.RandomStringFromRangeAndCharset(0, 32, test.CharsetText)
			requestBody = &RequestBody{Request: requestString}
			responseString = test.RandomStringFromRangeAndCharset(0, 32, test.CharsetText)
			inspectors = []request.ResponseInspector{request.NewHeadersInspector()}
			httpClient = http.DefaultClient
		})

		JustBeforeEach(func() {
			config.Address = server.URL()
			var err error
			clnt, err = client.NewWithErrorParser(config, errorResponseParser)
			Expect(err).ToNot(HaveOccurred())
			Expect(clnt).ToNot(BeNil())
		})

		AfterEach(func() {
			if server != nil {
				server.Close()
			}
		})

		Context("RequestStreamWithHTTPClient", func() {
			var reader io.ReadCloser
			var err error

			AfterEach(func() {
				if reader != nil {
					reader.Close()
				}
			})

			Context("with a user agent", func() {
				var userAgent string

				BeforeEach(func() {
					userAgent = testHttp.NewUserAgent()
					config.UserAgent = userAgent
				})

				It("sets the User-Agent header in requests", func() {
					server.AppendHandlers(CombineHandlers(
						VerifyHeaderKV("User-Agent", userAgent),
						RespondWith(http.StatusNoContent, nil)))

					_, err = clnt.RequestStreamWithHTTPClient(ctx, method, url, mutators, nil, inspectors, httpClient)

					Expect(err).ToNot(HaveOccurred())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})

			Context("without a user agent", func() {
				BeforeEach(func() {
					config.UserAgent = ""
				})

				It("doesn't set one (and the Go default is used)", func() {
					server.AppendHandlers(CombineHandlers(
						VerifyHeaderKV("User-Agent", "Go-http-client/1.1"),
						RespondWith(http.StatusNoContent, nil)))

					_, err = clnt.RequestStreamWithHTTPClient(ctx, method, url, mutators, nil, inspectors, httpClient)

					Expect(err).ToNot(HaveOccurred())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})

			It("returns error if http client is missing", func() {
				reader, err = clnt.RequestStreamWithHTTPClient(ctx, method, url, mutators, requestBody, inspectors, nil)
				Expect(err).To(MatchError("http client is missing"))
				Expect(reader).To(BeNil())
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			It("returns error if context is missing", func() {
				reader, err = clnt.RequestStreamWithHTTPClient(nil, method, url, mutators, requestBody, inspectors, httpClient)
				Expect(err).To(MatchError("context is missing"))
				Expect(reader).To(BeNil())
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			It("returns error if method is missing", func() {
				reader, err = clnt.RequestStreamWithHTTPClient(ctx, "", url, mutators, requestBody, inspectors, httpClient)
				Expect(err).To(MatchError("method is missing"))
				Expect(reader).To(BeNil())
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			It("returns error if url is missing", func() {
				reader, err = clnt.RequestStreamWithHTTPClient(ctx, method, "", mutators, requestBody, inspectors, httpClient)
				Expect(err).To(MatchError("url is missing"))
				Expect(reader).To(BeNil())
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			It("returns error if the request object cannot be encoded", func() {
				invalidRequestBody := struct{ Func interface{} }{func() {}}
				reader, err = clnt.RequestStreamWithHTTPClient(ctx, method, url, mutators, invalidRequestBody, inspectors, httpClient)
				Expect(err.Error()).To(MatchRegexp("unable to serialize request to .*; json: unsupported type: func()"))
				Expect(reader).To(BeNil())
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			It("returns error if method is invalid", func() {
				reader, err = clnt.RequestStreamWithHTTPClient(ctx, "/", url, mutators, requestBody, inspectors, httpClient)
				Expect(err.Error()).To(MatchRegexp("unable to create request to / .*; net/http: invalid method \"/\""))
				Expect(reader).To(BeNil())
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			It("returns error if mutator returns an error", func() {
				errorMutator := request.NewHeaderMutator("", "")
				invalidMutators := []request.RequestMutator{headerMutator, errorMutator, parameterMutator}
				reader, err = clnt.RequestStreamWithHTTPClient(ctx, method, url, invalidMutators, requestBody, inspectors, httpClient)
				Expect(err.Error()).To(MatchRegexp("unable to mutate request to .*; key is missing"))
				Expect(reader).To(BeNil())
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			It("returns error if the server is not reachable", func() {
				server.Close()
				server = nil
				reader, err = clnt.RequestStreamWithHTTPClient(ctx, method, url, mutators, requestBody, inspectors, httpClient)
				Expect(err.Error()).To(MatchRegexp("unable to perform request to .*: connect: connection refused"))
				Expect(reader).To(BeNil())
			})

			Context("with a successful response and no request body, but inspector returns error", func() {
				var responseErr error
				var errorInspector *requestTest.ResponseInspector

				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(nil),
							RespondWith(http.StatusOK, []byte(responseString), responseHeaders),
						),
					)
					responseErr = errorsTest.RandomError()
					errorInspector = requestTest.NewResponseInspector()
					errorInspector.InspectResponseOutputs = []error{responseErr}
				})

				AfterEach(func() {
					errorInspector.AssertOutputsEmpty()
				})
			})

			Context("with an bad request 400 without deserializable error body", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(test.MarshalRequestBody(requestBody)),
							RespondWith(http.StatusBadRequest, "NOT JSON"),
						),
					)
				})

				It("returns an error", func() {
					reader, err = clnt.RequestStreamWithHTTPClient(ctx, method, url, mutators, requestBody, inspectors, httpClient)
					errorsTest.ExpectEqual(err, request.ErrorBadRequest())
					Expect(reader).To(BeNil())
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
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(test.MarshalRequestBody(requestBody)),
							RespondWithJSONEncoded(http.StatusBadRequest, errors.NewSerializable(responseErr), responseHeaders),
						),
					)
				})

				It("returns an error", func() {
					reader, err = clnt.RequestStreamWithHTTPClient(ctx, method, url, mutators, requestBody, inspectors, httpClient)
					errorsTest.ExpectEqual(err, responseErr)
					Expect(reader).To(BeNil())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})

			Context("with an bad request 400 with deserializable error body without deserializing", func() {
				var responseErr error

				BeforeEach(func() {
					errorResponseParser = nil
					responseErr = errors.Append(structureValidator.ErrorValueNotEmpty(), structureValidator.ErrorValueBoolNotTrue(), structureValidator.ErrorValueIntNotOneOf(1, []int{0, 2, 4}))
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(test.MarshalRequestBody(requestBody)),
							RespondWithJSONEncoded(http.StatusBadRequest, errors.NewSerializable(responseErr), responseHeaders),
						),
					)
				})

				It("returns an error", func() {
					reader, err = clnt.RequestStreamWithHTTPClient(ctx, method, url, mutators, requestBody, inspectors, httpClient)
					errorsTest.ExpectEqual(err, request.ErrorBadRequest())
					Expect(reader).To(BeNil())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})

			Context("with an unauthorized response 401", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(test.MarshalRequestBody(requestBody)),
							RespondWith(http.StatusUnauthorized, "NOT JSON", responseHeaders),
						),
					)
				})

				It("returns an error", func() {
					reader, err = clnt.RequestStreamWithHTTPClient(ctx, method, url, mutators, requestBody, inspectors, httpClient)
					errorsTest.ExpectEqual(err, request.ErrorUnauthenticated())
					Expect(reader).To(BeNil())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})

			Context("with an forbidden response 403", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(test.MarshalRequestBody(requestBody)),
							RespondWith(http.StatusForbidden, "NOT JSON", responseHeaders),
						),
					)
				})

				It("returns an error", func() {
					reader, err = clnt.RequestStreamWithHTTPClient(ctx, method, url, mutators, requestBody, inspectors, httpClient)
					errorsTest.ExpectEqual(err, request.ErrorUnauthorized())
					Expect(reader).To(BeNil())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})

			Context("with an resource not found 404 without deserializable error body", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(test.MarshalRequestBody(requestBody)),
							RespondWith(http.StatusNotFound, "NOT JSON", responseHeaders),
						),
					)
				})

				It("returns an error", func() {
					reader, err = clnt.RequestStreamWithHTTPClient(ctx, method, url, mutators, requestBody, inspectors, httpClient)
					errorsTest.ExpectEqual(err, request.ErrorResourceNotFound())
					Expect(reader).To(BeNil())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})

			Context("with an resource not found 404 with deserializable error body", func() {
				var responseErr error

				BeforeEach(func() {
					responseErr = request.ErrorResourceNotFoundWithID(test.RandomStringFromRangeAndCharset(1, 16, test.CharsetHexadecimalLowercase))
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(test.MarshalRequestBody(requestBody)),
							RespondWithJSONEncoded(http.StatusNotFound, errors.NewSerializable(responseErr), responseHeaders),
						),
					)
				})

				It("returns an error", func() {
					reader, err = clnt.RequestStreamWithHTTPClient(ctx, method, url, mutators, requestBody, inspectors, httpClient)
					errorsTest.ExpectEqual(err, responseErr)
					Expect(reader).To(BeNil())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})

			Context("with an resource not found 404 with deserializable error body without deserializing", func() {
				var responseErr error

				BeforeEach(func() {
					errorResponseParser = nil
					responseErr = request.ErrorResourceNotFoundWithID(test.RandomStringFromRangeAndCharset(1, 16, test.CharsetHexadecimalLowercase))
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(test.MarshalRequestBody(requestBody)),
							RespondWithJSONEncoded(http.StatusNotFound, errors.NewSerializable(responseErr), responseHeaders),
						),
					)
				})

				It("returns an error", func() {
					reader, err = clnt.RequestStreamWithHTTPClient(ctx, method, url, mutators, requestBody, inspectors, httpClient)
					errorsTest.ExpectEqual(err, request.ErrorResourceNotFound())
					Expect(reader).To(BeNil())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})

			Context("with a request entity too large response 413", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(test.MarshalRequestBody(requestBody)),
							RespondWith(http.StatusRequestEntityTooLarge, "NOT JSON", responseHeaders),
						),
					)
				})

				It("returns an error", func() {
					reader, err = clnt.RequestStreamWithHTTPClient(ctx, method, url, mutators, requestBody, inspectors, httpClient)
					errorsTest.ExpectEqual(err, request.ErrorResourceTooLarge())
					Expect(reader).To(BeNil())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})

			Context("with a too many requests response 429", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(test.MarshalRequestBody(requestBody)),
							RespondWith(http.StatusTooManyRequests, "NOT JSON", responseHeaders),
						),
					)
				})

				It("returns an error", func() {
					reader, err = clnt.RequestStreamWithHTTPClient(ctx, method, url, mutators, requestBody, inspectors, httpClient)
					errorsTest.ExpectEqual(err, request.ErrorTooManyRequests())
					Expect(reader).To(BeNil())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})

			Context("with an unexpected response 500 without deserializable error body", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(test.MarshalRequestBody(requestBody)),
							RespondWith(http.StatusInternalServerError, []byte("[]"), responseHeaders),
						),
					)
				})

				It("returns an error", func() {
					reader, err = clnt.RequestStreamWithHTTPClient(ctx, method, url, mutators, requestBody, inspectors, httpClient)
					Expect(err).To(MatchError(fmt.Sprintf(`unexpected response status code 500 from %s "%s?%s=%s"`, method, url, parameterMutator.Key, parameterMutator.Value)))
					Expect(reader).To(BeNil())
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
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(test.MarshalRequestBody(requestBody)),
							RespondWithJSONEncoded(http.StatusInternalServerError, errors.NewSerializable(responseErr), responseHeaders),
						),
					)
				})

				It("returns an error", func() {
					reader, err = clnt.RequestStreamWithHTTPClient(ctx, method, url, mutators, requestBody, inspectors, httpClient)
					errorsTest.ExpectEqual(err, responseErr)
					Expect(reader).To(BeNil())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})

			Context("with an unexpected response 500 with deserializable error body without deserializing", func() {
				var responseErr error

				BeforeEach(func() {
					errorResponseParser = nil
					responseErr = errorsTest.RandomError()
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(test.MarshalRequestBody(requestBody)),
							RespondWithJSONEncoded(http.StatusInternalServerError, errors.NewSerializable(responseErr), responseHeaders),
						),
					)
				})

				It("returns an error", func() {
					reader, err = clnt.RequestStreamWithHTTPClient(ctx, method, url, mutators, requestBody, inspectors, httpClient)
					Expect(err).To(MatchError(fmt.Sprintf(`unexpected response status code 500 from %s "%s?%s=%s"`, method, url, parameterMutator.Key, parameterMutator.Value)))
					Expect(reader).To(BeNil())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})

			Context("with a successful response 204 without parsing content", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(test.MarshalRequestBody(requestBody)),
							RespondWith(http.StatusNoContent, nil),
						),
					)
				})

				It("returns success", func() {
					reader, err = clnt.RequestStreamWithHTTPClient(ctx, method, url, mutators, requestBody, inspectors, httpClient)
					Expect(err).ToNot(HaveOccurred())
					Expect(reader).To(BeNil())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})

			Context("with a successful response 205 without parsing content", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(test.MarshalRequestBody(requestBody)),
							RespondWith(http.StatusResetContent, nil),
						),
					)
				})

				It("returns success", func() {
					reader, err = clnt.RequestStreamWithHTTPClient(ctx, method, url, mutators, requestBody, inspectors, httpClient)
					Expect(err).ToNot(HaveOccurred())
					Expect(reader).To(BeNil())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})

			Context("with a successful response and no request body", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(nil),
							RespondWith(http.StatusOK, []byte(responseString)),
						),
					)
				})

				It("returns success", func() {
					reader, err = clnt.RequestStreamWithHTTPClient(ctx, method, url, mutators, nil, inspectors, httpClient)
					Expect(err).ToNot(HaveOccurred())
					Expect(reader).ToNot(BeNil())
					Expect(io.ReadAll(reader)).To(Equal([]byte(responseString)))
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})

			Context("with a successful response and request body reader", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody([]byte(requestString)),
							RespondWith(http.StatusOK, []byte(responseString)),
						),
					)
				})

				It("returns success", func() {
					reader, err = clnt.RequestStreamWithHTTPClient(ctx, method, url, mutators, strings.NewReader(requestString), inspectors, httpClient)
					Expect(err).ToNot(HaveOccurred())
					Expect(reader).ToNot(BeNil())
					Expect(io.ReadAll(reader)).To(Equal([]byte(responseString)))
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})

			Context("with a successful response and request body object", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(test.MarshalRequestBody(requestBody)),
							RespondWith(http.StatusOK, []byte(responseString)),
						),
					)
				})

				It("returns success", func() {
					reader, err = clnt.RequestStreamWithHTTPClient(ctx, method, url, mutators, requestBody, inspectors, httpClient)
					Expect(err).ToNot(HaveOccurred())
					Expect(reader).ToNot(BeNil())
					Expect(io.ReadAll(reader)).To(Equal([]byte(responseString)))
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})
		})

		Context("RequestDataWithHTTPClient", func() {
			var responseBody *ResponseBody

			BeforeEach(func() {
				responseBody = &ResponseBody{}
			})

			It("returns error if http client is missing", func() {
				Expect(clnt.RequestDataWithHTTPClient(ctx, method, url, mutators, requestBody, responseBody, inspectors, nil)).To(MatchError("http client is missing"))
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			It("returns error if context is missing", func() {
				Expect(clnt.RequestDataWithHTTPClient(nil, method, url, mutators, requestBody, responseBody, inspectors, httpClient)).To(MatchError("context is missing"))
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			It("returns error if method is missing", func() {
				Expect(clnt.RequestDataWithHTTPClient(ctx, "", url, mutators, requestBody, responseBody, inspectors, httpClient)).To(MatchError("method is missing"))
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			It("returns error if url is missing", func() {
				Expect(clnt.RequestDataWithHTTPClient(ctx, method, "", mutators, requestBody, responseBody, inspectors, httpClient)).To(MatchError("url is missing"))
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			It("returns error if the request object cannot be encoded", func() {
				invalidRequestBody := struct{ Func interface{} }{func() {}}
				Expect(clnt.RequestDataWithHTTPClient(ctx, method, url, mutators, invalidRequestBody, responseBody, inspectors, httpClient).Error()).To(MatchRegexp("unable to serialize request to .*; json: unsupported type: func()"))
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			It("returns error if mutator returns an error", func() {
				errorMutator := request.NewHeaderMutator("", "")
				invalidMutators := []request.RequestMutator{headerMutator, errorMutator, parameterMutator}
				Expect(clnt.RequestDataWithHTTPClient(ctx, method, url, invalidMutators, requestBody, responseBody, inspectors, httpClient).Error()).To(MatchRegexp("unable to mutate request to .*; key is missing"))
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			It("returns error if the server is not reachable", func() {
				server.Close()
				server = nil
				Expect(clnt.RequestDataWithHTTPClient(ctx, method, url, mutators, requestBody, responseBody, inspectors, httpClient).Error()).To(MatchRegexp("unable to perform request to .*: connect: connection refused"))
			})

			Context("with a successful response and no request body, but inspector returns error", func() {
				var responseErr error
				var errorInspector *requestTest.ResponseInspector

				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(nil),
							RespondWith(http.StatusOK, test.MarshalResponseBody(&ResponseBody{Response: responseString}), responseHeaders),
						),
					)
					responseErr = errorsTest.RandomError()
					errorInspector = requestTest.NewResponseInspector()
					errorInspector.InspectResponseOutputs = []error{responseErr}
				})

				AfterEach(func() {
					errorInspector.AssertOutputsEmpty()
				})
			})

			Context("with a successful response and no request body", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(nil),
							RespondWith(http.StatusOK, test.MarshalResponseBody(&ResponseBody{Response: responseString}), responseHeaders),
						),
					)
				})

				It("returns success", func() {
					Expect(clnt.RequestDataWithHTTPClient(ctx, method, url, mutators, nil, responseBody, inspectors, httpClient)).To(Succeed())
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
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(test.MarshalRequestBody(requestBody)),
							RespondWith(http.StatusBadRequest, []byte{255, 255, 255}, responseHeaders),
						),
					)
				})

				It("returns an error", func() {
					err := clnt.RequestDataWithHTTPClient(ctx, method, url, mutators, requestBody, responseBody, inspectors, httpClient)
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
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(test.MarshalRequestBody(requestBody)),
							RespondWithJSONEncoded(http.StatusBadRequest, errors.NewSerializable(responseErr), responseHeaders),
						),
					)
				})

				It("returns an error", func() {
					err := clnt.RequestDataWithHTTPClient(ctx, method, url, mutators, requestBody, responseBody, inspectors, httpClient)
					errorsTest.ExpectEqual(err, responseErr)
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})

			Context("with an bad request 400 with deserializable error body without deserializing", func() {
				var responseErr error

				BeforeEach(func() {
					errorResponseParser = nil
					responseErr = errors.Append(structureValidator.ErrorValueNotEmpty(), structureValidator.ErrorValueBoolNotTrue(), structureValidator.ErrorValueIntNotOneOf(1, []int{0, 2, 4}))
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(test.MarshalRequestBody(requestBody)),
							RespondWithJSONEncoded(http.StatusBadRequest, errors.NewSerializable(responseErr), responseHeaders),
						),
					)
				})

				It("returns an error", func() {
					err := clnt.RequestDataWithHTTPClient(ctx, method, url, mutators, requestBody, responseBody, inspectors, httpClient)
					errorsTest.ExpectEqual(err, request.ErrorBadRequest())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})

			Context("with an unauthorized response 401", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(test.MarshalRequestBody(requestBody)),
							RespondWith(http.StatusUnauthorized, "NOT JSON", responseHeaders),
						),
					)
				})

				It("returns an error", func() {
					err := clnt.RequestDataWithHTTPClient(ctx, method, url, mutators, requestBody, responseBody, inspectors, httpClient)
					errorsTest.ExpectEqual(err, request.ErrorUnauthenticated())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})

			Context("with an forbidden response 403", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(test.MarshalRequestBody(requestBody)),
							RespondWith(http.StatusForbidden, "NOT JSON", responseHeaders),
						),
					)
				})

				It("returns an error", func() {
					err := clnt.RequestDataWithHTTPClient(ctx, method, url, mutators, requestBody, responseBody, inspectors, httpClient)
					errorsTest.ExpectEqual(err, request.ErrorUnauthorized())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})

			Context("with an resource not found 404 without deserializable error body", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(test.MarshalRequestBody(requestBody)),
							RespondWith(http.StatusNotFound, "NOT JSON", responseHeaders),
						),
					)
				})

				It("returns an error", func() {
					err := clnt.RequestDataWithHTTPClient(ctx, method, url, mutators, requestBody, responseBody, inspectors, httpClient)
					errorsTest.ExpectEqual(err, request.ErrorResourceNotFound())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})

			Context("with an resource not found 404 with deserializable error body", func() {
				var responseErr error

				BeforeEach(func() {
					responseErr = request.ErrorResourceNotFoundWithID(test.RandomStringFromRangeAndCharset(1, 16, test.CharsetHexadecimalLowercase))
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(test.MarshalRequestBody(requestBody)),
							RespondWithJSONEncoded(http.StatusNotFound, errors.NewSerializable(responseErr), responseHeaders),
						),
					)
				})

				It("returns an error", func() {
					err := clnt.RequestDataWithHTTPClient(ctx, method, url, mutators, requestBody, responseBody, inspectors, httpClient)
					errorsTest.ExpectEqual(err, responseErr)
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})

			Context("with an resource not found 404 with deserializable error body without deserializing", func() {
				var responseErr error

				BeforeEach(func() {
					errorResponseParser = nil
					responseErr = request.ErrorResourceNotFoundWithID(test.RandomStringFromRangeAndCharset(1, 16, test.CharsetHexadecimalLowercase))
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(test.MarshalRequestBody(requestBody)),
							RespondWithJSONEncoded(http.StatusNotFound, errors.NewSerializable(responseErr), responseHeaders),
						),
					)
				})

				It("returns an error", func() {
					err := clnt.RequestDataWithHTTPClient(ctx, method, url, mutators, requestBody, responseBody, inspectors, httpClient)
					errorsTest.ExpectEqual(err, request.ErrorResourceNotFound())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})

			Context("with a resource too large response 413", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(test.MarshalRequestBody(requestBody)),
							RespondWith(http.StatusRequestEntityTooLarge, "NOT JSON", responseHeaders),
						),
					)
				})

				It("returns an error", func() {
					err := clnt.RequestDataWithHTTPClient(ctx, method, url, mutators, requestBody, responseBody, inspectors, httpClient)
					errorsTest.ExpectEqual(err, request.ErrorResourceTooLarge())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})

			Context("with a too many requests response 429", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(test.MarshalRequestBody(requestBody)),
							RespondWith(http.StatusTooManyRequests, "NOT JSON", responseHeaders),
						),
					)
				})

				It("returns an error", func() {
					err := clnt.RequestDataWithHTTPClient(ctx, method, url, mutators, requestBody, responseBody, inspectors, httpClient)
					errorsTest.ExpectEqual(err, request.ErrorTooManyRequests())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})

			Context("with an unexpected response 500 without deserializable error body", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(test.MarshalRequestBody(requestBody)),
							RespondWith(http.StatusInternalServerError, nil, responseHeaders),
						),
					)
				})

				It("returns an error", func() {
					err := clnt.RequestDataWithHTTPClient(ctx, method, url, mutators, requestBody, responseBody, inspectors, httpClient)
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
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(test.MarshalRequestBody(requestBody)),
							RespondWithJSONEncoded(http.StatusInternalServerError, errors.NewSerializable(responseErr), responseHeaders),
						),
					)
				})

				It("returns an error", func() {
					err := clnt.RequestDataWithHTTPClient(ctx, method, url, mutators, requestBody, responseBody, inspectors, httpClient)
					errorsTest.ExpectEqual(err, responseErr)
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})

			Context("with an unexpected response 500 with deserializable error body without deserializing", func() {
				var responseErr error

				BeforeEach(func() {
					errorResponseParser = nil
					responseErr = errorsTest.RandomError()
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(test.MarshalRequestBody(requestBody)),
							RespondWithJSONEncoded(http.StatusInternalServerError, errors.NewSerializable(responseErr), responseHeaders),
						),
					)
				})

				It("returns an error", func() {
					err := clnt.RequestDataWithHTTPClient(ctx, method, url, mutators, requestBody, responseBody, inspectors, httpClient)
					Expect(err).To(MatchError(fmt.Sprintf(`unexpected response status code 500 from %s "%s?%s=%s"`, method, url, parameterMutator.Key, parameterMutator.Value)))
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})

			Context("with an unparsable response", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(test.MarshalRequestBody(requestBody)),
							RespondWith(http.StatusOK, []byte("{\"response\":"), responseHeaders),
						),
					)
				})

				It("returns an error", func() {
					err := clnt.RequestDataWithHTTPClient(ctx, method, url, mutators, requestBody, responseBody, inspectors, httpClient)
					errorsTest.ExpectEqual(err, request.ErrorJSONMalformed())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})

			Context("with a successful response 204 without parsing content", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(test.MarshalRequestBody(requestBody)),
							RespondWith(http.StatusNoContent, nil),
						),
					)
				})

				It("returns success", func() {
					Expect(clnt.RequestDataWithHTTPClient(ctx, method, url, mutators, requestBody, responseBody, inspectors, httpClient)).To(Succeed())
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
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(test.MarshalRequestBody(requestBody)),
							RespondWith(http.StatusResetContent, nil),
						),
					)
				})

				It("returns success", func() {
					Expect(clnt.RequestDataWithHTTPClient(ctx, method, url, mutators, requestBody, responseBody, inspectors, httpClient)).To(Succeed())
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
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(nil),
							RespondWith(http.StatusOK, test.MarshalResponseBody(&ResponseBody{Response: responseString}), responseHeaders),
						),
					)
				})

				It("returns success", func() {
					Expect(clnt.RequestDataWithHTTPClient(ctx, method, url, mutators, nil, responseBody, inspectors, httpClient)).To(Succeed())
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
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody([]byte(requestString)),
							RespondWith(http.StatusOK, test.MarshalResponseBody(&ResponseBody{Response: responseString}), responseHeaders),
						),
					)
				})

				It("returns success", func() {
					Expect(clnt.RequestDataWithHTTPClient(ctx, method, url, mutators, strings.NewReader(requestString), responseBody, inspectors, httpClient)).To(Succeed())
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
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(test.MarshalRequestBody(requestBody)),
							RespondWith(http.StatusOK, test.MarshalResponseBody(&ResponseBody{Response: responseString}), responseHeaders),
						),
					)
				})

				It("returns success", func() {
					Expect(clnt.RequestDataWithHTTPClient(ctx, method, url, mutators, requestBody, responseBody, inspectors, httpClient)).To(Succeed())
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
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(test.MarshalRequestBody(requestBody)),
							RespondWith(http.StatusOK, test.MarshalResponseBody(&ResponseBody{Response: responseString}), responseHeaders),
						),
					)
				})

				It("returns success without parsing response body", func() {
					Expect(clnt.RequestDataWithHTTPClient(ctx, method, url, mutators, requestBody, nil, inspectors, httpClient)).To(Succeed())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})
		})
	})

	Context("NewSerializableErrorResponseParser", func() {
		It("returns success", func() {
			Expect(client.NewSerializableErrorResponseParser()).ToNot(BeNil())
		})
	})

	Context("SerializableErrorResponseParser", func() {
		It("returns nil if response body is not parsable", func() {
			serializableErrorResponseParser := client.NewSerializableErrorResponseParser()
			err := serializableErrorResponseParser.ParseErrorResponse(context.Background(), &http.Response{Body: io.NopCloser(bytes.NewReader([]byte("NOT JSON")))}, testHttp.NewRequest())
			Expect(err).To(BeNil())
		})

		It("returns deserialized error if response body is parsable", func() {
			responseErr := request.ErrorResourceNotFoundWithID(test.RandomStringFromRangeAndCharset(1, 16, test.CharsetHexadecimalLowercase))
			body, err := json.Marshal(errors.Serializable{Error: responseErr})
			Expect(err).ToNot(HaveOccurred())
			Expect(body).ToNot(BeNil())
			serializableErrorResponseParser := client.NewSerializableErrorResponseParser()
			err = serializableErrorResponseParser.ParseErrorResponse(context.Background(), &http.Response{Body: io.NopCloser(bytes.NewReader(body))}, testHttp.NewRequest())
			Expect(err).To(Equal(responseErr))
		})
	})
})
