package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"
	"net/http"
	"time"

	"github.com/onsi/gomega/ghttp"

	testAuth "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/client"
	dataClient "github.com/tidepool-org/platform/data/client"
	"github.com/tidepool-org/platform/id"
)

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
			clnt, err := dataClient.NewClient(nil)
			Expect(err).To(MatchError("client: config is missing"))
			Expect(clnt).To(BeNil())
		})

		It("returns an error if config address is missing", func() {
			config.Address = ""
			clnt, err := dataClient.NewClient(config)
			Expect(err).To(MatchError("client: config is invalid; client: address is missing"))
			Expect(clnt).To(BeNil())
		})

		It("returns success", func() {
			clnt, err := dataClient.NewClient(config)
			Expect(err).ToNot(HaveOccurred())
			Expect(clnt).ToNot(BeNil())
		})
	})

	Context("with started server and new client", func() {
		var server *ghttp.Server
		var clnt dataClient.Client
		var context *testAuth.Context

		BeforeEach(func() {
			server = ghttp.NewServer()
			config := client.NewConfig()
			Expect(config).ToNot(BeNil())
			config.Address = server.URL()
			config.Timeout = 30 * time.Second
			var err error
			clnt, err = dataClient.NewClient(config)
			Expect(err).ToNot(HaveOccurred())
			Expect(clnt).ToNot(BeNil())
			context = testAuth.NewContext()
			Expect(context).ToNot(BeNil())
		})

		AfterEach(func() {
			if server != nil {
				server.Close()
			}
			Expect(context.UnusedOutputsCount()).To(Equal(0))
		})

		Context("DestroyDataForUserByID", func() {
			var userID string

			BeforeEach(func() {
				userID = id.New()
			})

			It("returns error if context is missing", func() {
				Expect(clnt.DestroyDataForUserByID(nil, userID)).To(MatchError("client: context is missing"))
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			It("returns error if user id is missing", func() {
				Expect(clnt.DestroyDataForUserByID(context, "")).To(MatchError("client: user id is missing"))
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			Context("with server token", func() {
				var token string

				BeforeEach(func() {
					token = id.New()
					context.AuthClientImpl.ServerTokenOutputs = []testAuth.ServerTokenOutput{{Token: token, Error: nil}}
				})

				Context("with an unauthorized response", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("DELETE", fmt.Sprintf("/dataservices/v1/users/%s/data", userID)),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", token),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusUnauthorized, nil, nil)),
						)
					})

					It("returns an error", func() {
						err := clnt.DestroyDataForUserByID(context, userID)
						Expect(err).To(MatchError("client: unauthorized"))
						Expect(server.ReceivedRequests()).To(HaveLen(1))
					})
				})

				Context("with a successful response", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("DELETE", fmt.Sprintf("/dataservices/v1/users/%s/data", userID)),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", token),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusOK, nil, nil)),
						)
					})

					It("returns success", func() {
						Expect(clnt.DestroyDataForUserByID(context, userID)).To(Succeed())
						Expect(server.ReceivedRequests()).To(HaveLen(1))
					})
				})
			})
		})
	})
})
