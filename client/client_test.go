package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/tidepool-org/platform/client"
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
		var config *client.Config

		BeforeEach(func() {
			config = client.NewConfig()
			Expect(config).ToNot(BeNil())
			config.Address = testHTTP.NewAddress()
			config.UserAgent = testHTTP.NewUserAgent()
		})

		It("returns an error if config is missing", func() {
			clnt, err := client.New(nil)
			Expect(err).To(MatchError("config is missing"))
			Expect(clnt).To(BeNil())
		})

		It("returns an error if config address is missing", func() {
			config.Address = ""
			clnt, err := client.New(config)
			Expect(err).To(MatchError("config is invalid; address is missing"))
			Expect(clnt).To(BeNil())
		})

		It("returns an error if config user agent is missing", func() {
			config.UserAgent = ""
			clnt, err := client.New(config)
			Expect(err).To(MatchError("config is invalid; user agent is missing"))
			Expect(clnt).To(BeNil())
		})

		It("returns success", func() {
			clnt, err := client.New(config)
			Expect(err).ToNot(HaveOccurred())
			Expect(clnt).ToNot(BeNil())
		})
	})

	Context("with new client", func() {
		var address string
		var clnt *client.Client

		BeforeEach(func() {
			address = testHTTP.NewAddress()
			config := client.NewConfig()
			Expect(config).ToNot(BeNil())
			config.Address = address + "///"
			config.UserAgent = testHTTP.NewUserAgent()
			var err error
			clnt, err = client.New(config)
			Expect(err).ToNot(HaveOccurred())
			Expect(clnt).ToNot(BeNil())
		})

		Context("ConstructURL", func() {
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
		})

		Context("AppendURLQuery", func() {
			var urlString string

			BeforeEach(func() {
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

			It("returns a URL with associated query ven if it already has a query string", func() {
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
		var userAgent string
		var clnt *client.Client
		var ctx context.Context
		var method string
		var path string
		var url string
		var headerMutator *client.HeaderMutator
		var parameterMutator *client.ParameterMutator
		var mutators []client.Mutator
		var requestBodyString string
		var requestBody *RequestBody
		var responseBodyString string
		var responseBody *ResponseBody
		var httpClient *http.Client

		BeforeEach(func() {
			server = NewServer()
			userAgent = testHTTP.NewUserAgent()
			config := client.NewConfig()
			Expect(config).ToNot(BeNil())
			config.Address = server.URL()
			config.UserAgent = userAgent
			var err error
			clnt, err = client.New(config)
			Expect(err).ToNot(HaveOccurred())
			Expect(clnt).ToNot(BeNil())
			ctx = context.Background()
			method = testHTTP.NewMethod()
			path = testHTTP.NewPath()
			url = server.URL() + path
			headerMutator = client.NewHeaderMutator(testHTTP.NewHeaderKey(), testHTTP.NewHeaderValue())
			parameterMutator = client.NewParameterMutator(testHTTP.NewParameterKey(), testHTTP.NewParameterValue())
			mutators = []client.Mutator{headerMutator, parameterMutator}
			requestBodyString = test.NewVariableString(0, 32, test.CharsetAlphaNumeric)
			requestBody = &RequestBody{requestBodyString}
			responseBodyString = test.NewVariableString(0, 32, test.CharsetAlphaNumeric)
			responseBody = &ResponseBody{}
			httpClient = http.DefaultClient
		})

		AfterEach(func() {
			if server != nil {
				server.Close()
			}
		})

		Context("SendRequest", func() {
			It("returns error if context is missing", func() {
				Expect(clnt.SendRequest(nil, method, url, mutators, requestBody, responseBody, httpClient)).To(MatchError("context is missing"))
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			It("returns error if method is missing", func() {
				Expect(clnt.SendRequest(ctx, "", url, mutators, requestBody, responseBody, httpClient)).To(MatchError("method is missing"))
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			It("returns error if url is missing", func() {
				Expect(clnt.SendRequest(ctx, method, "", mutators, requestBody, responseBody, httpClient)).To(MatchError("url is missing"))
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			It("returns error if the request object cannot be encoded", func() {
				invalidRequestBody := struct{ Func interface{} }{func() {}}
				Expect(clnt.SendRequest(ctx, method, url, mutators, invalidRequestBody, responseBody, httpClient).Error()).To(HavePrefix("error encoding JSON request to"))
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			It("returns error if mutator is missing", func() {
				invalidMutators := []client.Mutator{headerMutator, nil, parameterMutator}
				Expect(clnt.SendRequest(ctx, method, url, invalidMutators, requestBody, responseBody, httpClient)).To(MatchError("mutator is missing"))
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			It("returns error if mutator returns an error", func() {
				errorMutator := client.NewHeaderMutator("", "")
				invalidMutators := []client.Mutator{headerMutator, errorMutator, parameterMutator}
				Expect(clnt.SendRequest(ctx, method, url, invalidMutators, requestBody, responseBody, httpClient)).To(MatchError("unable to mutate request; key is missing"))
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			It("returns error if http client is missing", func() {
				Expect(clnt.SendRequest(ctx, method, url, mutators, requestBody, responseBody, nil)).To(MatchError("http client is missing"))
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			It("returns error if the server is not reachable", func() {
				server.Close()
				server = nil
				err := clnt.SendRequest(ctx, method, url, mutators, requestBody, responseBody, httpClient)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(HavePrefix(fmt.Sprintf("unable to perform request %s %s", method, url)))
			})

			Context("with an unexpected response 500", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV("User-Agent", userAgent),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody([]byte("{\"request\":\""+requestBodyString+"\"}\n")),
							RespondWith(http.StatusInternalServerError, nil, nil)),
					)
				})

				It("returns an error", func() {
					err := clnt.SendRequest(ctx, method, url, mutators, requestBody, responseBody, httpClient)
					Expect(err).To(MatchError(fmt.Sprintf(`unexpected response status code 500 from %s "%s?%s=%s"`, method, url, parameterMutator.Key, parameterMutator.Value)))
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})

			Context("with an unauthorized response 401", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV("User-Agent", userAgent),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody([]byte("{\"request\":\""+requestBodyString+"\"}\n")),
							RespondWith(http.StatusUnauthorized, nil, nil)),
					)
				})

				It("returns an error", func() {
					err := clnt.SendRequest(ctx, method, url, mutators, requestBody, responseBody, httpClient)
					Expect(err).To(MatchError("authentication token is invalid"))
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})

			Context("with an unparseable response", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV("User-Agent", userAgent),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody([]byte("{\"request\":\""+requestBodyString+"\"}\n")),
							RespondWith(http.StatusOK, []byte("{\"response\":"), nil)),
					)
				})

				It("returns an error", func() {
					err := clnt.SendRequest(ctx, method, url, mutators, requestBody, responseBody, httpClient)
					Expect(err).To(MatchError("json is malformed; unexpected EOF"))
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})

			Context("with a successful response 200", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV("User-Agent", userAgent),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody([]byte("{\"request\":\""+requestBodyString+"\"}\n")),
							RespondWith(http.StatusOK, []byte("{\"response\":\""+responseBodyString+"\"}"), nil)),
					)
				})

				It("returns success", func() {
					Expect(clnt.SendRequest(ctx, method, url, mutators, requestBody, responseBody, httpClient)).To(Succeed())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
					Expect(responseBody).ToNot(BeNil())
					Expect(responseBody.Response).To(Equal(responseBodyString))
				})
			})

			Context("with a successful response 201", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV("User-Agent", userAgent),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody([]byte("{\"request\":\""+requestBodyString+"\"}\n")),
							RespondWith(http.StatusCreated, []byte("{\"response\":\""+responseBodyString+"\"}"), nil)),
					)
				})

				It("returns success", func() {
					Expect(clnt.SendRequest(ctx, method, url, mutators, requestBody, responseBody, httpClient)).To(Succeed())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
					Expect(responseBody).ToNot(BeNil())
					Expect(responseBody.Response).To(Equal(responseBodyString))
				})
			})

			Context("with a successful response, but no request object", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV("User-Agent", userAgent),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody([]byte{}),
							RespondWith(http.StatusOK, []byte("{\"response\":\""+responseBodyString+"\"}"), nil)),
					)
				})

				It("returns success", func() {
					Expect(clnt.SendRequest(ctx, method, url, mutators, nil, responseBody, httpClient)).To(Succeed())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
					Expect(responseBody).ToNot(BeNil())
					Expect(responseBody.Response).To(Equal(responseBodyString))
				})
			})

			Context("with a successful response, but no response object", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(method, path, fmt.Sprintf("%s=%s", parameterMutator.Key, parameterMutator.Value)),
							VerifyHeaderKV("User-Agent", userAgent),
							VerifyHeaderKV(headerMutator.Key, headerMutator.Value),
							VerifyBody([]byte("{\"request\":\""+requestBodyString+"\"}\n")),
							RespondWith(http.StatusOK, []byte("{\"response\":\""+responseBodyString+"\"}"), nil)),
					)
				})

				It("returns success without parsing response body", func() {
					Expect(clnt.SendRequest(ctx, method, url, mutators, requestBody, nil, httpClient)).To(Succeed())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})
		})
	})
})
