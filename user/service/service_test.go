package service_test

import (
	"context"
	"net/http"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"

	applicationTest "github.com/tidepool-org/platform/application/test"
	authTest "github.com/tidepool-org/platform/auth/test"
	configTest "github.com/tidepool-org/platform/config/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	kafkaTest "github.com/tidepool-org/platform/kafka/test"
	"github.com/tidepool-org/platform/test"
	testHttp "github.com/tidepool-org/platform/test/http"
	userService "github.com/tidepool-org/platform/user/service"
)

var _ = Describe("Service", func() {
	Context("New", func() {
		It("returns successfully", func() {
			Expect(userService.New()).ToNot(BeNil())
		})
	})

	Context("with started server, config reporter, and new service", func() {
		var provider *applicationTest.Provider
		var serverSecret string
		var sessionToken string
		var server *Server
		var authClientConfig map[string]interface{}
		var blobClientConfig map[string]interface{}
		var dataClientConfig map[string]interface{}
		var dataSourceClientConfig map[string]interface{}
		var imageClientConfig map[string]interface{}
		var metricClientConfig map[string]interface{}
		var permissionClientConfig map[string]interface{}
		var permissionStoreConfig map[string]interface{}
		var confirmationStoreConfig map[string]interface{}
		var messageStoreConfig map[string]interface{}
		var profileStoreConfig map[string]interface{}
		var sessionStoreConfig map[string]interface{}
		var passwordConfig map[string]interface{}
		var userStructuredStoreConfig map[string]interface{}
		var userServiceConfig map[string]interface{}
		var service *userService.Service

		BeforeEach(func() {
			provider = applicationTest.NewProviderWithDefaults()

			serverSecret = authTest.NewServiceSecret()
			sessionToken = authTest.NewSessionToken()
			server = NewServer()
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/auth/serverlogin"),
					VerifyHeaderKV("X-Tidepool-Server-Name", *provider.NameOutput),
					VerifyHeaderKV("X-Tidepool-Server-Secret", serverSecret),
					VerifyBody(nil),
					RespondWith(http.StatusOK, nil, http.Header{"X-Tidepool-Session-Token": []string{sessionToken}})),
			)

			authClientConfig = map[string]interface{}{
				"address":             server.URL(),
				"server_token_secret": authTest.NewServiceSecret(),
				"external": map[string]interface{}{
					"address":                     server.URL(),
					"server_session_token_secret": serverSecret,
				},
			}
			blobClientConfig = map[string]interface{}{
				"address": server.URL(),
			}
			dataClientConfig = map[string]interface{}{
				"address": server.URL(),
			}
			dataSourceClientConfig = map[string]interface{}{
				"address": server.URL(),
			}
			imageClientConfig = map[string]interface{}{
				"address": server.URL(),
			}
			metricClientConfig = map[string]interface{}{
				"address": server.URL(),
			}
			permissionClientConfig = map[string]interface{}{
				"address": server.URL(),
			}
			permissionStoreConfig = map[string]interface{}{
				"addresses": os.Getenv("TIDEPOOL_STORE_ADDRESSES"),
				"database":  test.RandomStringFromRangeAndCharset(4, 8, test.CharsetLowercase),
				"tls":       "false",
				"secret":    test.RandomString(),
			}
			confirmationStoreConfig = map[string]interface{}{
				"addresses": os.Getenv("TIDEPOOL_STORE_ADDRESSES"),
				"database":  test.RandomStringFromRangeAndCharset(4, 8, test.CharsetLowercase),
				"tls":       "false",
			}
			messageStoreConfig = map[string]interface{}{
				"addresses": os.Getenv("TIDEPOOL_STORE_ADDRESSES"),
				"database":  test.RandomStringFromRangeAndCharset(4, 8, test.CharsetLowercase),
				"tls":       "false",
			}
			profileStoreConfig = map[string]interface{}{
				"addresses": os.Getenv("TIDEPOOL_STORE_ADDRESSES"),
				"database":  test.RandomStringFromRangeAndCharset(4, 8, test.CharsetLowercase),
				"tls":       "false",
			}
			sessionStoreConfig = map[string]interface{}{
				"addresses": os.Getenv("TIDEPOOL_STORE_ADDRESSES"),
				"database":  test.RandomStringFromRangeAndCharset(4, 8, test.CharsetLowercase),
				"tls":       "false",
			}
			passwordConfig = map[string]interface{}{
				"salt": test.RandomString(),
			}
			userStructuredStoreConfig = map[string]interface{}{
				"addresses": os.Getenv("TIDEPOOL_STORE_ADDRESSES"),
				"database":  test.RandomStringFromRangeAndCharset(4, 8, test.CharsetLowercase),
				"tls":       "false",
				"password":  passwordConfig,
			}
			userServiceConfig = map[string]interface{}{
				"auth": map[string]interface{}{
					"client": authClientConfig,
				},
				"blob": map[string]interface{}{
					"client": blobClientConfig,
				},
				"data": map[string]interface{}{
					"client": dataClientConfig,
				},
				"data_source": map[string]interface{}{
					"client": dataSourceClientConfig,
				},
				"image": map[string]interface{}{
					"client": imageClientConfig,
				},
				"metric": map[string]interface{}{
					"client": metricClientConfig,
				},
				"permission": map[string]interface{}{
					"client": permissionClientConfig,
					"store":  permissionStoreConfig,
				},
				"confirmation": map[string]interface{}{
					"store": confirmationStoreConfig,
				},
				"message": map[string]interface{}{
					"store": messageStoreConfig,
				},
				"profile": map[string]interface{}{
					"store": profileStoreConfig,
				},
				"session": map[string]interface{}{
					"store": sessionStoreConfig,
				},
				"user": map[string]interface{}{
					"store": userStructuredStoreConfig,
				},
				"secret": authTest.NewServiceSecret(),
				"server": map[string]interface{}{
					"address": testHttp.NewAddress(),
					"tls":     "false",
				},
			}
			(*provider.ConfigReporterOutput).(*configTest.Reporter).Config = userServiceConfig
			kafkaTest.SetTestEnvironmentVariables()

			service = userService.New()
			Expect(service).ToNot(BeNil())
		})

		AfterEach(func() {
			if server != nil {
				server.Close()
			}
			provider.AssertOutputsEmpty()
			kafkaTest.RestoreOldEnvironmentVariables(map[string]string{})
		})

		Context("with Terminate after", func() {
			AfterEach(func() {
				service.Terminate()
			})

			Context("Initialize", func() {
				It("returns an error when the provider is missing", func() {
					errorsTest.ExpectEqual(service.Initialize(nil), errors.New("provider is missing"))
				})

				It("returns an error when the auth client returns an error", func() {
					authClientConfig["address"] = ""
					errorsTest.ExpectEqual(service.Initialize(provider), errors.New("unable to create auth client"))
				})

				It("returns an error when the blob client returns an error", func() {
					blobClientConfig["address"] = ""
					errorsTest.ExpectEqual(service.Initialize(provider), errors.New("unable to create blob client"))
				})

				It("returns an error when the data client returns an error", func() {
					dataClientConfig["address"] = ""
					errorsTest.ExpectEqual(service.Initialize(provider), errors.New("unable to create data client"))
				})

				It("returns an error when the data source client returns an error", func() {
					dataSourceClientConfig["address"] = ""
					errorsTest.ExpectEqual(service.Initialize(provider), errors.New("unable to create data source client"))
				})

				It("returns an error when the image client returns an error", func() {
					imageClientConfig["address"] = ""
					errorsTest.ExpectEqual(service.Initialize(provider), errors.New("unable to create image client"))
				})

				It("returns an error when the metric client returns an error", func() {
					metricClientConfig["address"] = ""
					errorsTest.ExpectEqual(service.Initialize(provider), errors.New("unable to create metric client"))
				})

				It("returns an error when the permission client returns an error", func() {
					permissionClientConfig["address"] = ""
					errorsTest.ExpectEqual(service.Initialize(provider), errors.New("unable to create permission client"))
				})

				It("returns an error when the password hasher returns an error", func() {
					passwordConfig["salt"] = ""
					errorsTest.ExpectEqual(service.Initialize(provider), errors.New("unable to create password hasher"))
				})

				It("returns successfully", func() {
					Expect(service.Initialize(provider)).To(Succeed())
				})
			})

			Context("with Initialize before", func() {
				BeforeEach(func() {
					Expect(service.Initialize(provider)).To(Succeed())
				})

				Context("Terminate", func() {
					It("returns successfully", func() {
						service.Terminate()
					})
				})

				Context("Status", func() {
					It("returns successfully", func() {
						Expect(service.Status(context.Background())).ToNot(BeNil())
					})
				})

				Context("AuthClient", func() {
					It("returns successfully", func() {
						Expect(service.AuthClient()).ToNot(BeNil())
					})
				})

				Context("BlobClient", func() {
					It("returns successfully", func() {
						Expect(service.BlobClient()).ToNot(BeNil())
					})
				})

				Context("DataClient", func() {
					It("returns successfully", func() {
						Expect(service.DataClient()).ToNot(BeNil())
					})
				})

				Context("DataSourceClient", func() {
					It("returns successfully", func() {
						Expect(service.DataSourceClient()).ToNot(BeNil())
					})
				})

				Context("ImageClient", func() {
					It("returns successfully", func() {
						Expect(service.ImageClient()).ToNot(BeNil())
					})
				})

				Context("MetricClient", func() {
					It("returns successfully", func() {
						Expect(service.MetricClient()).ToNot(BeNil())
					})
				})

				Context("PermissionClient", func() {
					It("returns successfully", func() {
						Expect(service.PermissionClient()).ToNot(BeNil())
					})
				})

				Context("ConfirmationStore", func() {
					It("returns successfully", func() {
						Expect(service.ConfirmationStore()).ToNot(BeNil())
					})
				})

				Context("MessageStore", func() {
					It("returns successfully", func() {
						Expect(service.MessageStore()).ToNot(BeNil())
					})
				})

				Context("PermissionStore", func() {
					It("returns successfully", func() {
						Expect(service.PermissionStore()).ToNot(BeNil())
					})
				})

				Context("ProfileStore", func() {
					It("returns successfully", func() {
						Expect(service.ProfileStore()).ToNot(BeNil())
					})
				})

				Context("SessionStore", func() {
					It("returns successfully", func() {
						Expect(service.SessionStore()).ToNot(BeNil())
					})
				})

				Context("UserStructuredStore", func() {
					It("returns successfully", func() {
						Expect(service.UserStructuredStore()).ToNot(BeNil())
					})
				})

				Context("PasswordHasher", func() {
					It("returns successfully", func() {
						Expect(service.PasswordHasher()).ToNot(BeNil())
					})
				})

				Context("UserClient", func() {
					It("returns successfully", func() {
						Expect(service.UserClient()).ToNot(BeNil())
					})
				})
			})
		})
	})
})
