package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	"net/http"

	testAuth "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/task"
	taskClient "github.com/tidepool-org/platform/task/client"
)

var _ = Describe("Client", func() {
	var cfg *client.Config

	BeforeEach(func() {
		cfg = client.NewConfig()
		Expect(cfg).ToNot(BeNil())
	})

	Context("New", func() {
		BeforeEach(func() {
			cfg.Address = "http://localhost:1234"
		})

		It("returns an error if unsuccessful", func() {
			clnt, err := taskClient.New(nil)
			Expect(err).To(HaveOccurred())
			Expect(clnt).To(BeNil())
		})

		It("returns success", func() {
			clnt, err := taskClient.New(cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(clnt).ToNot(BeNil())
		})
	})

	Context("with server and new client", func() {
		var svr *Server
		var clnt task.Client
		var ctx *testAuth.Context

		BeforeEach(func() {
			svr = NewServer()
			Expect(svr).ToNot(BeNil())
			cfg.Address = svr.URL()
			var err error
			clnt, err = taskClient.New(cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(clnt).ToNot(BeNil())
			ctx = testAuth.NewContext()
			Expect(ctx).ToNot(BeNil())
		})

		AfterEach(func() {
			if svr != nil {
				svr.Close()
			}
			Expect(ctx.UnusedOutputsCount()).To(Equal(0))
		})

		Context("GetStatus", func() {
			It("returns an error if unsuccessful", func() {
				sts, err := clnt.GetStatus(nil)
				Expect(err).To(HaveOccurred())
				Expect(sts).To(BeNil())
				Expect(svr.ReceivedRequests()).To(BeEmpty())
			})

			Context("with server token", func() {
				var serverToken string

				BeforeEach(func() {
					serverToken = id.New()
					ctx.AuthClientImpl.ServerTokenOutputs = []testAuth.ServerTokenOutput{{Token: serverToken, Error: nil}}
				})

				Context("with an empty body", func() {
					BeforeEach(func() {
						svr.AppendHandlers(
							CombineHandlers(
								VerifyRequest("GET", "/status"),
								VerifyHeaderKV("X-Tidepool-Session-Token", serverToken),
								VerifyBody([]byte{}),
								RespondWith(http.StatusOK, nil, nil)),
						)
					})

					It("returns an error", func() {
						sts, err := clnt.GetStatus(ctx)
						Expect(err.Error()).To(HavePrefix("client: error decoding JSON response from GET "))
						Expect(sts).To(BeNil())
						Expect(svr.ReceivedRequests()).To(HaveLen(1))
					})
				})

				Context("with a successful, but empty response", func() {
					BeforeEach(func() {
						svr.AppendHandlers(
							CombineHandlers(
								VerifyRequest("GET", "/status"),
								VerifyHeaderKV("X-Tidepool-Session-Token", serverToken),
								VerifyBody([]byte{}),
								RespondWith(http.StatusOK, `{}`, nil)),
						)
					})

					It("returns successfully", func() {
						sts, err := clnt.GetStatus(ctx)
						Expect(err).ToNot(HaveOccurred())
						Expect(sts).ToNot(BeNil())
						Expect(sts.Version).To(BeEmpty())
						Expect(svr.ReceivedRequests()).To(HaveLen(1))
					})
				})

				Context("with a successful response", func() {
					BeforeEach(func() {
						svr.AppendHandlers(
							CombineHandlers(
								VerifyRequest("GET", "/status"),
								VerifyHeaderKV("X-Tidepool-Session-Token", serverToken),
								VerifyBody([]byte{}),
								RespondWith(http.StatusOK, `{"version": "1.2.3"}`, nil)),
						)
					})

					It("returns successfully", func() {
						sts, err := clnt.GetStatus(ctx)
						Expect(err).ToNot(HaveOccurred())
						Expect(sts).ToNot(BeNil())
						Expect(sts.Version).To(Equal("1.2.3"))
						Expect(svr.ReceivedRequests()).To(HaveLen(1))
					})
				})
			})
		})
	})
})
