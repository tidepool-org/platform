package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/request"
	requestTest "github.com/tidepool-org/platform/request/test"
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
	Context("New", func() {
		var cfg *client.Config

		BeforeEach(func() {
			cfg = client.NewConfig()
			Expect(cfg).ToNot(BeNil())
			cfg.Address = testHTTP.NewAddress()
			cfg.UserAgent = testHTTP.NewUserAgent()
		})

		It("returns an error if config is missing", func() {
			clnt, err := client.New(nil)
			Expect(err).To(MatchError("config is missing"))
			Expect(clnt).To(BeNil())
		})

		It("returns an error if config is invalid", func() {
			cfg.Address = ""
			clnt, err := client.New(cfg)
			Expect(err).To(MatchError("config is invalid; address is missing"))
			Expect(clnt).To(BeNil())
		})

		It("returns success", func() {
			Expect(client.New(cfg)).ToNot(BeNil())
		})
	})

	Context("with new client", func() {
		var address string
		var cfg *client.Config
		var clnt *client.Client

		BeforeEach(func() {
			address = testHTTP.NewAddress()
			cfg = client.NewConfig()
			Expect(cfg).ToNot(BeNil())
			cfg.Address = address
			cfg.UserAgent = testHTTP.NewUserAgent()
		})

		JustBeforeEach(func() {
			var err error
			clnt, err = client.New(cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(clnt).ToNot(BeNil())
		})

		Context("ConstructURL", func() {
			AssertExpectedBehavior := func() {
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
				AssertExpectedBehavior()
			})

			Context("with trailing slashes on address", func() {
				BeforeEach(func() {
					cfg.Address += "///"
				})

				AssertExpectedBehavior()
			})
		})

		Context("AppendURLQuery", func() {
			var urlString string

			JustBeforeEach(func() {
				urlString = clnt.ConstructURL(test.NewVariableString(1, 8, test.CharsetAlphaNumeric), test.NewVariableString(1, 8, test.CharsetAlphaNumeric))
			})

			It("returns a URL without change if the query is nil", func() {
				Expect(clnt.AppendURLQuery(urlString, nil)).To(Equal(urlString))
			})

			It("returns a URL without change if the query is empty", func() {
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

			It("returns a URL with associated query even if it already has a query string", func() {
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
		var userAgent string
		var clnt *client.Client
		var ctx context.Context
		var method string
		var path string
		var url string
		var headerMutator *request.HeaderMutator
		var parameterMutator *request.ParameterMutator
		var mutators []request.RequestMutator
		var requestString string
		var requestBody *RequestBody
		var requestJSON []byte
		var responseString string
		var headersInspector *request.HeadersInspector
		var inspectors []request.ResponseInspector
		var httpClient *http.Client

		BeforeEach(func() {
			server = NewServer()
			responseHeaders = http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}
			userAgent = testHTTP.NewUserAgent()
			cfg := client.NewConfig()
			Expect(cfg).ToNot(BeNil())
			cfg.Address = server.URL()
			cfg.UserAgent = userAgent
			var err error
			clnt, err = client.New(cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(clnt).ToNot(BeNil())
			ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
			method = testHTTP.NewMethod()
			path = testHTTP.NewPath()
			url = server.URL() + path
			headerMutator = request.NewHeaderMutator(testHTTP.NewHeaderKey(), testHTTP.NewHeaderValue())
			parameterMutator = request.NewParameterMutator(testHTTP.NewParameterKey(), testHTTP.NewParameterValue())
			mutators = []request.RequestMutator{headerMutator, parameterMutator}
			requestString = test.NewVariableString(0, 32, test.CharsetText)
			requestBody = &RequestBody{Request: requestString}
			requestJSON, err = json.Marshal(requestBody)
			Expect(err).ToNot(HaveOccurred())
			Expect(requestJSON).ToNot(BeNil())
			requestJSON = append(requestJSON, []byte("\n")...)
			responseString = test.NewVariableString(0, 32, test.CharsetText)
			headersInspector = request.NewHeadersInspector()
			inspectors = []request.ResponseInspector{}
			httpClient = http.DefaultClient
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
							VerifyHeaderKV("User-Agent", userAgent),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(nil),
							RespondWith(http.StatusOK, []byte(responseString), responseHeaders),
						),
					)
					responseErr = errorsTest.NewError()
					errorInspector = requestTest.NewResponseInspector()
					errorInspector.InspectResponseOutputs = []error{responseErr}
				})

				AfterEach(func() {
					errorInspector.AssertOutputsEmpty()
				})

				It("returns error if inspector returns an error", func() {
					reader, err = clnt.RequestStreamWithHTTPClient(ctx, method, url, mutators, nil, append(inspectors, errorInspector), httpClient)
					Expect(err).To(MatchError(responseErr))
					Expect(reader).To(BeNil())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
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
							VerifyBody(requestJSON),
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
					responseErr = errors.Append(structureValidator.ErrorValueNotEmpty(), structureValidator.ErrorValueBooleanNotTrue(), structureValidator.ErrorValueIntNotOneOf(1, []int{0, 2, 4}))
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV("User-Agent", userAgent),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(requestJSON),
							RespondWithJSONEncoded(http.StatusBadRequest, errors.Serializable{Error: responseErr}, responseHeaders),
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

			Context("with an unauthorized response 401", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV("User-Agent", userAgent),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(requestJSON),
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
							VerifyHeaderKV("User-Agent", userAgent),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(requestJSON),
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
							VerifyHeaderKV("User-Agent", userAgent),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(requestJSON),
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
					responseErr = request.ErrorResourceNotFoundWithID(test.RandomStringFromRangeAndCharset(1, 16, test.CharsetHexidecimalLowercase))
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV("User-Agent", userAgent),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(requestJSON),
							RespondWithJSONEncoded(http.StatusNotFound, errors.Serializable{Error: responseErr}, responseHeaders),
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

			Context("with a too many requests response 429", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV("User-Agent", userAgent),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(requestJSON),
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
							VerifyHeaderKV("User-Agent", userAgent),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(requestJSON),
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
					responseErr = errorsTest.NewError()
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV("User-Agent", userAgent),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(requestJSON),
							RespondWithJSONEncoded(http.StatusInternalServerError, errors.Serializable{Error: responseErr}, responseHeaders),
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

			Context("with a successful response 204 without parsing content", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV("User-Agent", userAgent),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(requestJSON),
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
							VerifyHeaderKV("User-Agent", userAgent),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(requestJSON),
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
							VerifyHeaderKV("User-Agent", userAgent),
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
					Expect(ioutil.ReadAll(reader)).To(Equal([]byte(responseString)))
					Expect(server.ReceivedRequests()).To(HaveLen(1))
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
							RespondWith(http.StatusOK, []byte(responseString)),
						),
					)
				})

				It("returns success", func() {
					reader, err = clnt.RequestStreamWithHTTPClient(ctx, method, url, mutators, strings.NewReader(requestString), inspectors, httpClient)
					Expect(err).ToNot(HaveOccurred())
					Expect(reader).ToNot(BeNil())
					Expect(ioutil.ReadAll(reader)).To(Equal([]byte(responseString)))
					Expect(server.ReceivedRequests()).To(HaveLen(1))
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
							VerifyBody(requestJSON),
							RespondWith(http.StatusOK, []byte(responseString)),
						),
					)
				})

				It("returns success", func() {
					reader, err = clnt.RequestStreamWithHTTPClient(ctx, method, url, mutators, requestBody, inspectors, httpClient)
					Expect(err).ToNot(HaveOccurred())
					Expect(reader).ToNot(BeNil())
					Expect(ioutil.ReadAll(reader)).To(Equal([]byte(responseString)))
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})
		})

		Context("RequestDataWithHTTPClient", func() {
			var responseBody *ResponseBody
			var responseJSON []byte

			BeforeEach(func() {
				responseBody = &ResponseBody{}
				var err error
				responseJSON, err = json.Marshal(&ResponseBody{Response: responseString})
				Expect(err).ToNot(HaveOccurred())
				Expect(responseJSON).ToNot(BeNil())
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
							VerifyHeaderKV("User-Agent", userAgent),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(nil),
							RespondWith(http.StatusOK, responseJSON, responseHeaders),
						),
					)
					responseErr = errorsTest.NewError()
					errorInspector = requestTest.NewResponseInspector()
					errorInspector.InspectResponseOutputs = []error{responseErr}
				})

				AfterEach(func() {
					errorInspector.AssertOutputsEmpty()
				})

				It("returns error if inspector returns an error", func() {
					Expect(clnt.RequestDataWithHTTPClient(ctx, method, url, mutators, nil, responseBody, append(inspectors, errorInspector), httpClient)).To(MatchError(responseErr))
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
							RespondWith(http.StatusOK, responseJSON, responseHeaders),
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
							VerifyHeaderKV("User-Agent", userAgent),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(requestJSON),
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
					responseErr = errors.Append(structureValidator.ErrorValueNotEmpty(), structureValidator.ErrorValueBooleanNotTrue(), structureValidator.ErrorValueIntNotOneOf(1, []int{0, 2, 4}))
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV("User-Agent", userAgent),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(requestJSON),
							RespondWithJSONEncoded(http.StatusBadRequest, errors.Serializable{Error: responseErr}, responseHeaders),
						),
					)
				})

				It("returns an error", func() {
					err := clnt.RequestDataWithHTTPClient(ctx, method, url, mutators, requestBody, responseBody, inspectors, httpClient)
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
							VerifyBody(requestJSON),
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
							VerifyHeaderKV("User-Agent", userAgent),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(requestJSON),
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
							VerifyHeaderKV("User-Agent", userAgent),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(requestJSON),
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
					responseErr = request.ErrorResourceNotFoundWithID(test.RandomStringFromRangeAndCharset(1, 16, test.CharsetHexidecimalLowercase))
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV("User-Agent", userAgent),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(requestJSON),
							RespondWithJSONEncoded(http.StatusNotFound, errors.Serializable{Error: responseErr}, responseHeaders),
						),
					)
				})

				It("returns an error", func() {
					err := clnt.RequestDataWithHTTPClient(ctx, method, url, mutators, requestBody, responseBody, inspectors, httpClient)
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
							VerifyBody(requestJSON),
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
							VerifyHeaderKV("User-Agent", userAgent),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(requestJSON),
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
					responseErr = errorsTest.NewError()
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV("User-Agent", userAgent),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(requestJSON),
							RespondWithJSONEncoded(http.StatusInternalServerError, errors.Serializable{Error: responseErr}, responseHeaders),
						),
					)
				})

				It("returns an error", func() {
					err := clnt.RequestDataWithHTTPClient(ctx, method, url, mutators, requestBody, responseBody, inspectors, httpClient)
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
							VerifyBody(requestJSON),
							RespondWith(http.StatusOK, []byte("{\"response\":"), responseHeaders),
						),
					)
				})

				It("returns an error", func() {
					err := clnt.RequestDataWithHTTPClient(ctx, method, url, mutators, requestBody, responseBody, inspectors, httpClient)
					Expect(err).To(MatchError("json is malformed; unexpected EOF"))
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})

			Context("without a content type", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV("User-Agent", userAgent),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(requestJSON),
							RespondWith(http.StatusOK, responseJSON),
						),
					)
				})

				It("returns an error", func() {
					err := clnt.RequestDataWithHTTPClient(ctx, method, url, mutators, requestBody, responseBody, inspectors, httpClient)
					Expect(err).To(MatchError(`header "Content-Type" is invalid`))
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})

			Context("with an invalid content type", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV("User-Agent", userAgent),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(requestJSON),
							RespondWith(http.StatusOK, responseJSON, http.Header{"Content-Type": []string{"/"}}),
						),
					)
				})

				It("returns an error", func() {
					err := clnt.RequestDataWithHTTPClient(ctx, method, url, mutators, requestBody, responseBody, inspectors, httpClient)
					Expect(err).To(MatchError(`header "Content-Type" is invalid`))
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})

			Context("with an unexpected content type", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV("User-Agent", userAgent),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(requestJSON),
							RespondWith(http.StatusOK, responseJSON, http.Header{"Content-Type": []string{"application/json"}}),
						),
					)
				})

				It("returns an error", func() {
					err := clnt.RequestDataWithHTTPClient(ctx, method, url, mutators, requestBody, responseBody, inspectors, httpClient)
					Expect(err).To(MatchError(`header "Content-Type" is invalid`))
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
							VerifyBody(requestJSON),
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
							VerifyHeaderKV("User-Agent", userAgent),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(requestJSON),
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
							VerifyHeaderKV("User-Agent", userAgent),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(nil),
							RespondWith(http.StatusOK, responseJSON, responseHeaders),
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
							VerifyHeaderKV("User-Agent", userAgent),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody([]byte(requestString)),
							RespondWith(http.StatusOK, responseJSON, responseHeaders),
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
							VerifyHeaderKV("User-Agent", userAgent),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(requestJSON),
							RespondWith(http.StatusOK, responseJSON, responseHeaders),
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
							VerifyHeaderKV("User-Agent", userAgent),
							VerifyHeaderKV("Content-Type", "application/json; charset=utf-8"),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody(requestJSON),
							RespondWith(http.StatusOK, responseJSON, responseHeaders),
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
})
