package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"net/http"
	"time"

	"github.com/onsi/gomega/ghttp"

	testAuth "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/id"
	userClient "github.com/tidepool-org/platform/user/client"
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
			clnt, err := userClient.NewClient(nil)
			Expect(err).To(MatchError("client: config is missing"))
			Expect(clnt).To(BeNil())
		})

		It("returns an error if config address is missing", func() {
			config.Address = ""
			clnt, err := userClient.NewClient(config)
			Expect(err).To(MatchError("client: config is invalid; client: address is missing"))
			Expect(clnt).To(BeNil())
		})

		It("returns success", func() {
			clnt, err := userClient.NewClient(config)
			Expect(err).ToNot(HaveOccurred())
			Expect(clnt).ToNot(BeNil())
		})
	})

	Context("with started server and new client", func() {
		var server *ghttp.Server
		var clnt userClient.Client
		var context *testAuth.Context

		BeforeEach(func() {
			server = ghttp.NewServer()
			config := client.NewConfig()
			Expect(config).ToNot(BeNil())
			config.Address = server.URL()
			config.Timeout = 30 * time.Second
			var err error
			clnt, err = userClient.NewClient(config)
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

		Context("GetUserPermissions", func() {
			var requestUserID string
			var targetUserID string

			BeforeEach(func() {
				requestUserID = id.New()
				targetUserID = id.New()
			})

			It("returns error if context is missing", func() {
				permissions, err := clnt.GetUserPermissions(nil, requestUserID, targetUserID)
				Expect(err).To(MatchError("client: context is missing"))
				Expect(permissions).To(BeNil())
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			It("returns error if request user id is missing", func() {
				permissions, err := clnt.GetUserPermissions(context, "", targetUserID)
				Expect(err).To(MatchError("client: request user id is missing"))
				Expect(permissions).To(BeNil())
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})

			It("returns error if target user id is missing", func() {
				permissions, err := clnt.GetUserPermissions(context, requestUserID, "")
				Expect(err).To(MatchError("client: target user id is missing"))
				Expect(permissions).To(BeNil())
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
								ghttp.VerifyRequest("GET", "/access/"+targetUserID+"/"+requestUserID),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", token),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusUnauthorized, nil, nil)),
						)
					})

					It("returns an error", func() {
						permissions, err := clnt.GetUserPermissions(context, requestUserID, targetUserID)
						Expect(err).To(MatchError("client: unauthorized"))
						Expect(permissions).To(BeNil())
						Expect(server.ReceivedRequests()).To(HaveLen(1))
					})
				})

				Context("with a not found response, which is the same as unauthorized", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/access/"+targetUserID+"/"+requestUserID),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", token),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusNotFound, nil, nil)),
						)
					})

					It("returns an unauthorized error", func() {
						permissions, err := clnt.GetUserPermissions(context, requestUserID, targetUserID)
						Expect(err).To(MatchError("client: unauthorized"))
						Expect(permissions).To(BeNil())
						Expect(server.ReceivedRequests()).To(HaveLen(1))
					})
				})

				Context("with a successful response, but with no permissions", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/access/"+targetUserID+"/"+requestUserID),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", token),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusOK, "{}", nil)),
						)
					})

					It("returns an error", func() {
						Expect(clnt.GetUserPermissions(context, requestUserID, targetUserID)).To(BeEmpty())
						Expect(server.ReceivedRequests()).To(HaveLen(1))
					})
				})

				Context("with a successful response with upload and view permissions", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/access/"+targetUserID+"/"+requestUserID),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", token),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusOK, `{"upload": {}, "view": {}}`, nil)),
						)
					})

					It("returns an error", func() {
						Expect(clnt.GetUserPermissions(context, requestUserID, targetUserID)).To(Equal(userClient.Permissions{
							userClient.UploadPermission: userClient.Permission{},
							userClient.ViewPermission:   userClient.Permission{},
						}))
						Expect(server.ReceivedRequests()).To(HaveLen(1))
					})
				})

				Context("with a successful response with owner permissions that already includes upload permissions", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/access/"+targetUserID+"/"+requestUserID),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", token),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusOK, `{"root": {"root-inner": "unused"}, "upload": {}}`, nil)),
						)
					})

					It("returns an error", func() {
						Expect(clnt.GetUserPermissions(context, requestUserID, targetUserID)).To(Equal(userClient.Permissions{
							userClient.OwnerPermission:  userClient.Permission{"root-inner": "unused"},
							userClient.UploadPermission: userClient.Permission{},
							userClient.ViewPermission:   userClient.Permission{"root-inner": "unused"},
						}))
						Expect(server.ReceivedRequests()).To(HaveLen(1))
					})
				})

				Context("with a successful response with owner permissions that already includes view permissions", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/access/"+targetUserID+"/"+requestUserID),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", token),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusOK, `{"root": {"root-inner": "unused"}, "view": {}}`, nil)),
						)
					})

					It("returns an error", func() {
						Expect(clnt.GetUserPermissions(context, requestUserID, targetUserID)).To(Equal(userClient.Permissions{
							userClient.OwnerPermission:  userClient.Permission{"root-inner": "unused"},
							userClient.UploadPermission: userClient.Permission{"root-inner": "unused"},
							userClient.ViewPermission:   userClient.Permission{},
						}))
						Expect(server.ReceivedRequests()).To(HaveLen(1))
					})
				})

				Context("with a successful response with owner permissions that already includes upload and view permissions", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", "/access/"+targetUserID+"/"+requestUserID),
								ghttp.VerifyHeaderKV("X-Tidepool-Session-Token", token),
								ghttp.VerifyBody([]byte{}),
								ghttp.RespondWith(http.StatusOK, `{"root": {"root-inner": "unused"}, "upload": {}, "view": {}}`, nil)),
						)
					})

					It("returns an error", func() {
						Expect(clnt.GetUserPermissions(context, requestUserID, targetUserID)).To(Equal(userClient.Permissions{
							userClient.OwnerPermission:  userClient.Permission{"root-inner": "unused"},
							userClient.UploadPermission: userClient.Permission{},
							userClient.ViewPermission:   userClient.Permission{},
						}))
						Expect(server.ReceivedRequests()).To(HaveLen(1))
					})
				})
			})
		})
	})
})
