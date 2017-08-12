package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"
	"net/http"
	"time"

	"github.com/onsi/gomega/ghttp"

	testAuth "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/id"
)

type RequestObject struct {
	Request string `json:"request"`
}

type ResponseObject struct {
	Response string `json:"response"`
}

var _ = Describe("Client", func() {
	Context("NewClient", func() {
		var config *client.Config

		BeforeEach(func() {
			config = client.NewConfig()
			Expect(config).ToNot(BeNil())
			config.Address = "http://localhost:1234"
			config.Timeout = 30 * time.Second
		})

		It("returns an error if config is missing", func() {
			clnt, err := client.NewClient(nil)
			Expect(err).To(MatchError("client: config is missing"))
			Expect(clnt).To(BeNil())
		})

		It("returns an error if config address is missing", func() {
			config.Address = ""
			clnt, err := client.NewClient(config)
			Expect(err).To(MatchError("client: config is invalid; client: address is missing"))
			Expect(clnt).To(BeNil())
		})

		It("returns success", func() {
			clnt, err := client.NewClient(config)
			Expect(err).ToNot(HaveOccurred())
			Expect(clnt).ToNot(BeNil())
		})
	})

	Context("with new client", func() {
		var clnt *client.Client

		BeforeEach(func() {
			config := client.NewConfig()
			Expect(config).ToNot(BeNil())
			config.Address = "http://localhost:1234"
			config.Timeout = 30 * time.Second
			var err error
			clnt, err = client.NewClient(config)
			Expect(err).ToNot(HaveOccurred())
			Expect(clnt).ToNot(BeNil())
		})

		Context("HTTPClient", func() {
			It("returns not nil", func() {
				Expect(clnt.HTTPClient()).ToNot(BeNil())
			})
		})

		Context("BuildURL", func() {
			It("returns a valid URL with one path", func() {
				Expect(clnt.BuildURL("one")).To(Equal("http://localhost:1234/one"))
			})

			It("returns a valid URL with multiple paths", func() {
				Expect(clnt.BuildURL("one", "two", "three")).To(Equal("http://localhost:1234/one/two/three"))
			})

			It("returns a valid URL with multiple paths that need to be escaped", func() {
				Expect(clnt.BuildURL("o n e", "t/w/o")).To(Equal("http://localhost:1234/o%20n%20e/t%2Fw%2Fo"))
			})
		})
	})

	Context("with started server and new client", func() {
		var server *ghttp.Server
		var clnt *client.Client
		var context *testAuth.Context
		var path string
		var url string
		var requestObject *RequestObject
		var responseObject *ResponseObject

		BeforeEach(func() {
			server = ghttp.NewServer()
			config := client.NewConfig()
			Expect(config).ToNot(BeNil())
			config.Address = server.URL()
			config.Timeout = 30 * time.Second
			var err error
			clnt, err = client.NewClient(config)
			Expect(err).ToNot(HaveOccurred())
			Expect(clnt).ToNot(BeNil())
			context = testAuth.NewContext()
			Expect(context).ToNot(BeNil())
			path = "/a/bb/ccc"
			url = server.URL() + path
			requestObject = &RequestObject{"alpha"}
			responseObject = &ResponseObject{}
		})

		AfterEach(func() {
			if server != nil {
				server.Close()
			}
			Expect(context.UnusedOutputsCount()).To(Equal(0))
		})

		Context("SendRequestWithAuthToken", func() {
			It("returns error if context is missing", func() {
				Expect(clnt.SendRequestWithAuthToken(nil, "GET", url, requestObject, responseObject)).To(MatchError("client: context is missing"))
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			Context("with auth token", func() {
				var token string

				BeforeEach(func() {
					token = id.New()
					context.AuthDetailsImpl.TokenOutputs = []string{token}
				})

				It("returns error if method is missing", func() {
					Expect(clnt.SendRequestWithAuthToken(context, "", url, requestObject, responseObject)).To(MatchError("client: method is missing"))
					Expect(server.ReceivedRequests()).To(BeEmpty())
				})

				It("returns error if url is missing", func() {
					Expect(clnt.SendRequestWithAuthToken(context, "GET", "", requestObject, responseObject)).To(MatchError("client: url is missing"))
					Expect(server.ReceivedRequests()).To(BeEmpty())
				})

				It("returns error if token is missing", func() {
					context.AuthDetailsImpl.TokenOutputs = []string{""}
					Expect(clnt.SendRequestWithAuthToken(context, "GET", url, requestObject, responseObject)).To(MatchError("client: token is missing"))
					Expect(server.ReceivedRequests()).To(BeEmpty())
				})

				It("returns error if the request object cannot be encoded", func() {
					invalidRequestObject := struct{ Func interface{} }{func() {}}
					Expect(clnt.SendRequestWithAuthToken(context, "GET", url, invalidRequestObject, responseObject).Error()).To(HavePrefix("client: error encoding JSON request to"))
					Expect(server.ReceivedRequests()).To(BeEmpty())
				})

				It("returns an error if unable to copy request trace", func() {
					context.RequestImpl = nil
					Expect(clnt.SendRequestWithAuthToken(context, "GET", url, requestObject, responseObject)).To(MatchError("client: unable to copy request trace; service: source request is missing"))
					Expect(server.ReceivedRequests()).To(BeEmpty())
				})

				It("returns error if the server is not reachable", func() {
					server.Close()
					server = nil
					err := clnt.SendRequestWithAuthToken(context, "GET", url, requestObject, responseObject)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(HavePrefix("client: unable to perform request GET "))
				})

				Context("with an unexpected response 400", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", path),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", token),
								ghttp.VerifyBody([]byte("{\"request\":\"alpha\"}\n")),
								ghttp.RespondWith(http.StatusBadRequest, nil, nil)),
						)
					})

					It("returns an error", func() {
						err := clnt.SendRequestWithAuthToken(context, "GET", url, requestObject, responseObject)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(HavePrefix("client: unexpected response status code 400 from GET "))
						Expect(server.ReceivedRequests()).To(HaveLen(1))
					})
				})

				Context("with an unauthorized response 401", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", path),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", token),
								ghttp.VerifyBody([]byte("{\"request\":\"alpha\"}\n")),
								ghttp.RespondWith(http.StatusUnauthorized, nil, nil)),
						)
					})

					It("returns an error", func() {
						err := clnt.SendRequestWithAuthToken(context, "GET", url, requestObject, responseObject)
						Expect(err).To(MatchError("client: unauthorized"))
						Expect(server.ReceivedRequests()).To(HaveLen(1))
					})
				})

				Context("with an unparseable response", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", path),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", token),
								ghttp.VerifyBody([]byte("{\"request\":\"alpha\"}\n")),
								ghttp.RespondWith(http.StatusOK, []byte("{\"response\":"), nil)),
						)
					})

					It("returns an error", func() {
						err := clnt.SendRequestWithAuthToken(context, "GET", url, requestObject, responseObject)
						Expect(err.Error()).To(HavePrefix("client: error decoding JSON response from GET "))
						Expect(server.ReceivedRequests()).To(HaveLen(1))
					})
				})

				Context("with a successful response 200", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", path),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", token),
								ghttp.VerifyBody([]byte("{\"request\":\"alpha\"}\n")),
								ghttp.RespondWith(http.StatusOK, []byte("{\"response\":\"beta\"}"), nil)),
						)
					})

					It("returns success", func() {
						Expect(clnt.SendRequestWithAuthToken(context, "GET", url, requestObject, responseObject)).To(Succeed())
						Expect(server.ReceivedRequests()).To(HaveLen(1))
						Expect(responseObject).ToNot(BeNil())
						Expect(responseObject.Response).To(Equal("beta"))
					})
				})

				Context("with a successful response 201", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", path),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", token),
								ghttp.VerifyBody([]byte("{\"request\":\"alpha\"}\n")),
								ghttp.RespondWith(http.StatusCreated, []byte("{\"response\":\"beta\"}"), nil)),
						)
					})

					It("returns success", func() {
						Expect(clnt.SendRequestWithAuthToken(context, "GET", url, requestObject, responseObject)).To(Succeed())
						Expect(server.ReceivedRequests()).To(HaveLen(1))
						Expect(responseObject).ToNot(BeNil())
						Expect(responseObject.Response).To(Equal("beta"))
					})
				})

				Context("with a successful response, but no request object", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", path),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", token),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusOK, []byte("{\"response\":\"beta\"}"), nil)),
						)
					})

					It("returns success", func() {
						Expect(clnt.SendRequestWithAuthToken(context, "GET", url, nil, responseObject)).To(Succeed())
						Expect(server.ReceivedRequests()).To(HaveLen(1))
						Expect(responseObject).ToNot(BeNil())
						Expect(responseObject.Response).To(Equal("beta"))
					})
				})

				Context("with a successful response, but no response object", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", path),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", token),
								ghttp.VerifyBody([]byte("{\"request\":\"alpha\"}\n")),
								ghttp.RespondWith(http.StatusOK, []byte("{\"response\":\"beta\"}"), nil)),
						)
					})

					It("returns success without parsing response body", func() {
						Expect(clnt.SendRequestWithAuthToken(context, "GET", url, requestObject, nil)).To(Succeed())
						Expect(server.ReceivedRequests()).To(HaveLen(1))
					})
				})
			})
		})

		Context("SendRequestWithServerToken", func() {
			It("returns error if context is missing", func() {
				Expect(clnt.SendRequestWithServerToken(nil, "GET", url, requestObject, responseObject)).To(MatchError("client: context is missing"))
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			Context("with server token", func() {
				var token string

				BeforeEach(func() {
					token = id.New()
					context.AuthClientImpl.ServerTokenOutputs = []testAuth.ServerTokenOutput{{Token: token, Error: nil}}
				})

				It("returns error if server token returns error", func() {
					context.AuthClientImpl.ServerTokenOutputs = []testAuth.ServerTokenOutput{{Token: "", Error: errors.New("test")}}
					Expect(clnt.SendRequestWithServerToken(context, "", url, requestObject, responseObject)).To(MatchError("test"))
					Expect(server.ReceivedRequests()).To(BeEmpty())
				})

				It("returns error if method is missing", func() {
					Expect(clnt.SendRequestWithServerToken(context, "", url, requestObject, responseObject)).To(MatchError("client: method is missing"))
					Expect(server.ReceivedRequests()).To(BeEmpty())
				})

				It("returns error if url is missing", func() {
					Expect(clnt.SendRequestWithServerToken(context, "GET", "", requestObject, responseObject)).To(MatchError("client: url is missing"))
					Expect(server.ReceivedRequests()).To(BeEmpty())
				})

				It("returns error if token is missing", func() {
					context.AuthClientImpl.ServerTokenOutputs = []testAuth.ServerTokenOutput{{Token: "", Error: nil}}
					Expect(clnt.SendRequestWithServerToken(context, "GET", url, requestObject, responseObject)).To(MatchError("client: token is missing"))
					Expect(server.ReceivedRequests()).To(BeEmpty())
				})

				It("returns error if the request object cannot be encoded", func() {
					invalidRequestObject := struct{ Cycle interface{} }{func() {}}
					Expect(clnt.SendRequestWithServerToken(context, "GET", url, invalidRequestObject, responseObject).Error()).To(HavePrefix("client: error encoding JSON request to"))
					Expect(server.ReceivedRequests()).To(BeEmpty())
				})

				It("returns an error if unable to copy request trace", func() {
					context.RequestImpl = nil
					Expect(clnt.SendRequestWithServerToken(context, "GET", url, requestObject, responseObject)).To(MatchError("client: unable to copy request trace; service: source request is missing"))
					Expect(server.ReceivedRequests()).To(BeEmpty())
				})

				It("returns error if the server is not reachable", func() {
					server.Close()
					server = nil
					err := clnt.SendRequestWithServerToken(context, "GET", url, requestObject, responseObject)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(HavePrefix("client: unable to perform request GET "))
				})

				Context("with an unexpected response 400", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", path),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", token),
								ghttp.VerifyBody([]byte("{\"request\":\"alpha\"}\n")),
								ghttp.RespondWith(http.StatusBadRequest, nil, nil)),
						)
					})

					It("returns an error", func() {
						err := clnt.SendRequestWithServerToken(context, "GET", url, requestObject, responseObject)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(HavePrefix("client: unexpected response status code 400 from GET "))
						Expect(server.ReceivedRequests()).To(HaveLen(1))
					})
				})

				Context("with an unauthorized response 401", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", path),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", token),
								ghttp.VerifyBody([]byte("{\"request\":\"alpha\"}\n")),
								ghttp.RespondWith(http.StatusUnauthorized, nil, nil)),
						)
					})

					It("returns an error", func() {
						err := clnt.SendRequestWithServerToken(context, "GET", url, requestObject, responseObject)
						Expect(err).To(MatchError("client: unauthorized"))
						Expect(server.ReceivedRequests()).To(HaveLen(1))
					})
				})

				Context("with an unparseable response", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", path),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", token),
								ghttp.VerifyBody([]byte("{\"request\":\"alpha\"}\n")),
								ghttp.RespondWith(http.StatusOK, []byte("{\"response\":"), nil)),
						)
					})

					It("returns an error", func() {
						err := clnt.SendRequestWithServerToken(context, "GET", url, requestObject, responseObject)
						Expect(err.Error()).To(HavePrefix("client: error decoding JSON response from GET "))
						Expect(server.ReceivedRequests()).To(HaveLen(1))
					})
				})

				Context("with a successful response 200", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", path),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", token),
								ghttp.VerifyBody([]byte("{\"request\":\"alpha\"}\n")),
								ghttp.RespondWith(http.StatusOK, []byte("{\"response\":\"beta\"}"), nil)),
						)
					})

					It("returns success", func() {
						Expect(clnt.SendRequestWithServerToken(context, "GET", url, requestObject, responseObject)).To(Succeed())
						Expect(server.ReceivedRequests()).To(HaveLen(1))
						Expect(responseObject).ToNot(BeNil())
						Expect(responseObject.Response).To(Equal("beta"))
					})
				})

				Context("with a successful response 201", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", path),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", token),
								ghttp.VerifyBody([]byte("{\"request\":\"alpha\"}\n")),
								ghttp.RespondWith(http.StatusCreated, []byte("{\"response\":\"beta\"}"), nil)),
						)
					})

					It("returns success", func() {
						Expect(clnt.SendRequestWithServerToken(context, "GET", url, requestObject, responseObject)).To(Succeed())
						Expect(server.ReceivedRequests()).To(HaveLen(1))
						Expect(responseObject).ToNot(BeNil())
						Expect(responseObject.Response).To(Equal("beta"))
					})
				})

				Context("with a successful response, but no request object", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", path),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", token),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusOK, []byte("{\"response\":\"beta\"}"), nil)),
						)
					})

					It("returns success", func() {
						Expect(clnt.SendRequestWithServerToken(context, "GET", url, nil, responseObject)).To(Succeed())
						Expect(server.ReceivedRequests()).To(HaveLen(1))
						Expect(responseObject).ToNot(BeNil())
						Expect(responseObject.Response).To(Equal("beta"))
					})
				})

				Context("with a successful response, but no response object", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", path),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", token),
								ghttp.VerifyBody([]byte("{\"request\":\"alpha\"}\n")),
								ghttp.RespondWith(http.StatusOK, []byte("{\"response\":\"beta\"}"), nil)),
						)
					})

					It("returns success without parsing response body", func() {
						Expect(clnt.SendRequestWithServerToken(context, "GET", url, requestObject, nil)).To(Succeed())
						Expect(server.ReceivedRequests()).To(HaveLen(1))
					})
				})
			})
		})
	})
})
