package v1_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/tidepool-org/platform/auth"
	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/permission"

	"github.com/ant0ine/go-json-rest/rest"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"github.com/tidepool-org/platform/blob"
	blobServiceApiV1 "github.com/tidepool-org/platform/blob/service/api/v1"
	blobServiceApiV1Test "github.com/tidepool-org/platform/blob/service/api/v1/test"
	blobTest "github.com/tidepool-org/platform/blob/test"
	cryptoTest "github.com/tidepool-org/platform/crypto/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/net"
	netTest "github.com/tidepool-org/platform/net/test"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	requestTest "github.com/tidepool-org/platform/request/test"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
	testRest "github.com/tidepool-org/platform/test/rest"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("V1", func() {
	var provider *blobServiceApiV1Test.Provider

	BeforeEach(func() {
		provider = blobServiceApiV1Test.NewProvider()
	})

	AfterEach(func() {
		provider.AssertOutputsEmpty()
	})

	Context("NewRouter", func() {
		It("returns an error when the provider is missing", func() {
			router, err := blobServiceApiV1.NewRouter(nil)
			errorsTest.ExpectEqual(err, errors.New("provider is missing"))
			Expect(router).To(BeNil())
		})

		It("returns successfully", func() {
			Expect(blobServiceApiV1.NewRouter(provider)).ToNot(BeNil())
		})
	})

	Context("with new router", func() {
		var router *blobServiceApiV1.Router

		BeforeEach(func() {
			var err error
			router, err = blobServiceApiV1.NewRouter(provider)
			Expect(err).ToNot(HaveOccurred())
			Expect(router).ToNot(BeNil())
		})

		Context("Routes", func() {
			It("returns the expected routes", func() {
				Expect(router.Routes()).To(ConsistOf(
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodGet), "PathExp": Equal("/v1/users/:userId/blobs")})),
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodPost), "PathExp": Equal("/v1/users/:userId/blobs")})),
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodDelete), "PathExp": Equal("/v1/users/:userId/blobs")})),
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodPost), "PathExp": Equal("/v1/users/:userId/device_logs")})),
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodGet), "PathExp": Equal("/v1/users/:userId/device_logs")})),
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodGet), "PathExp": Equal("/v1/blobs/:id")})),

					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodGet), "PathExp": Equal("/v1/blobs/:id/content")})),
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodGet), "PathExp": Equal("/v1/device_logs/:id")})),
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodGet), "PathExp": Equal("/v1/device_logs/:id/content")})),
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodDelete), "PathExp": Equal("/v1/blobs/:id")})),
				))
			})
		})

		Context("with response and request", func() {
			var res *testRest.ResponseWriter
			var req *rest.Request
			var ctx context.Context
			var handlerFunc rest.HandlerFunc

			BeforeEach(func() {
				res = testRest.NewResponseWriter()
				res.HeaderOutput = &http.Header{}
				req = testRest.NewRequest()
				ctx = log.NewContextWithLogger(req.Context(), logTest.NewLogger())
				req.Request = req.WithContext(ctx)
			})

			JustBeforeEach(func() {
				app, err := rest.MakeRouter(router.Routes()...)
				Expect(err).ToNot(HaveOccurred())
				Expect(app).ToNot(BeNil())
				handlerFunc = app.AppFunc()
			})

			AfterEach(func() {
				res.AssertOutputsEmpty()
			})

			Context("with user id", func() {
				var userID string

				BeforeEach(func() {
					userID = userTest.RandomID()
				})

				Context("List", func() {
					BeforeEach(func() {
						req.Method = http.MethodGet
						req.URL.Path = fmt.Sprintf("/v1/users/%s/blobs", userID)
					})

					It("panics when the response is missing", func() {
						Expect(func() { router.List(nil, req) }).To(Panic())
					})

					It("panics when the request is missing", func() {
						Expect(func() { router.List(res, nil) }).To(Panic())
					})

					Context("responds with JSON", func() {
						BeforeEach(func() {
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
						})

						AfterEach(func() {
							Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
						})

						When("the path does not contain a user id", func() {
							BeforeEach(func() {
								req.URL.Path = "/v1/users//blobs"
							})

							It("responds with bad request and expected error in body", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorParameterMissing("userId"), res.WriteInputs[0])
							})
						})

						When("the path contains an invalid user id", func() {
							BeforeEach(func() {
								req.URL.Path = "/v1/users/invalid/blobs"
							})

							It("responds with bad request and expected error in body", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorParameterInvalid("userId"), res.WriteInputs[0])
							})
						})

						When("a filter status query parameter is invalid", func() {
							BeforeEach(func() {
								query := url.Values{"status": []string{"created", "invalid"}}
								req.URL.RawQuery = query.Encode()
							})

							It("responds with bad request and expected error in body", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(errorsTest.WithParameterSource(structureValidator.ErrorValueStringNotOneOf("invalid", blob.Statuses()), "status"), res.WriteInputs[0])
							})
						})

						When("a filter media type query parameter is invalid", func() {
							BeforeEach(func() {
								query := url.Values{"mediaType": []string{netTest.RandomMediaType(), "/"}}
								req.URL.RawQuery = query.Encode()
							})

							It("responds with bad request and expected error in body", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(errorsTest.WithParameterSource(net.ErrorValueStringAsMediaTypeNotValid("/"), "mediaType"), res.WriteInputs[0])
							})
						})

						When("a pagination query parameter is invalid", func() {
							BeforeEach(func() {
								query := url.Values{"size": []string{"0"}}
								req.URL.RawQuery = query.Encode()
							})

							It("responds with bad request and expected error in body", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(errorsTest.WithParameterSource(structureValidator.ErrorValueNotInRange(0, 1, 1000), "size"), res.WriteInputs[0])
							})
						})

						Context("with clients", func() {
							var authClient *authTest.Client
							var client *blobTest.Client

							BeforeEach(func() {
								authClient = authTest.NewClient()
								client = blobTest.NewClient()
								provider.BlobClientOutputs = []blob.Client{client}
								provider.AuthClientOutputs = []auth.Client{authClient}
							})

							BeforeEach(func() {
								authClient.EnsureAuthorizedServiceOutputs = []error{nil}
							})

							AfterEach(func() {
								authClient.AssertOutputsEmpty()
								client.AssertOutputsEmpty()
							})

							When("the client returns an unauthorized error", func() {
								It("responds with an unauthorized error", func() {
									provider.BlobClientOutputs = nil
									authClient.EnsureAuthorizedServiceOutputs = []error{request.ErrorUnauthorized()}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorUnauthorized(), res.WriteInputs[0])
								})
							})

							parameterAssertions := func() {
								It("responds with an internal server error when the client returns an unknown error", func() {
									client.ListOutputs = []blobTest.ListOutput{{BlobArray: nil, Error: errorsTest.RandomError()}}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusInternalServerError}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorInternalServerError(nil), res.WriteInputs[0])
								})

								It("responds successfully when the client does not return blobs", func() {
									client.ListOutputs = []blobTest.ListOutput{{BlobArray: blob.BlobArray{}, Error: nil}}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
									Expect(res.WriteInputs).To(HaveLen(1))
									Expect(res.WriteInputs[0]).To(MatchJSON("[]"))
								})

								It("responds successfully when the client returns blobs", func() {
									blobs := blobTest.RandomBlobArray(1, 4)
									client.ListOutputs = []blobTest.ListOutput{{BlobArray: blobs, Error: nil}}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
									Expect(res.WriteInputs).To(HaveLen(1))
									Expect(json.Marshal(blobs)).To(MatchJSON(res.WriteInputs[0]))
								})
							}

							When("the filter and pagination query parameters are not specified", func() {
								AfterEach(func() {
									Expect(client.ListInputs).To(Equal([]blobTest.ListInput{{
										UserID:     userID,
										Filter:     blob.NewFilter(),
										Pagination: page.NewPagination(),
									}}))
								})

								parameterAssertions()
							})

							When("the filter and pagination query parameters are specified", func() {
								var mediaType []string
								var status []string
								var paige int
								var size int

								BeforeEach(func() {
									mediaType = netTest.RandomMediaTypes(1, 3)
									status = blobTest.RandomStatuses()
									paige = test.RandomIntFromRange(0, math.MaxInt32)
									size = test.RandomIntFromRange(1, 100)
									query := url.Values{
										"mediaType": mediaType,
										"status":    status,
										"page":      []string{strconv.Itoa(paige)},
										"size":      []string{strconv.Itoa(size)},
									}
									req.URL.RawQuery = query.Encode()
								})

								AfterEach(func() {
									Expect(client.ListInputs).To(Equal([]blobTest.ListInput{{
										UserID: userID,
										Filter: &blob.Filter{
											MediaType: pointer.FromStringArray(mediaType),
											Status:    pointer.FromStringArray(status),
										},
										Pagination: &page.Pagination{
											Page: paige,
											Size: size,
										},
									}}))
								})

								parameterAssertions()
							})
						})
					})
				})

				Context("ListDeviceLogs", func() {
					BeforeEach(func() {
						req.Method = http.MethodGet
						req.URL.Path = fmt.Sprintf("/v1/users/%s/device_logs", userID)
					})

					Context("responds with JSON", func() {
						BeforeEach(func() {
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
						})
						AfterEach(func() {
							if res.WriteHeaderInputs[0] < 300 {
								Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
							}
						})

						Context("with clients", func() {
							var authClient *authTest.Client
							var client *blobTest.Client
							var details request.AuthDetails

							BeforeEach(func() {
								authClient = authTest.NewClient()
								client = blobTest.NewClient()
								provider.BlobClientOutputs = []blob.Client{client}
								provider.AuthClientOutputs = []auth.Client{authClient}
							})

							AfterEach(func() {
								authClient.AssertOutputsEmpty()
								client.AssertOutputsEmpty()
							})

							Context("non-logged in user", func() {
								It("responds with unauthenticated if not logged in", func() {
									provider.BlobClientOutputs = nil
									provider.AuthClientOutputs = nil
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusUnauthorized}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorUnauthenticated(), res.WriteInputs[0])

								})
							})

							Context("with server details", func() {
								BeforeEach(func() {
									details = request.NewAuthDetails(request.MethodServiceSecret, "", authTest.NewSessionToken())
									req.Request = req.WithContext(request.NewContextWithAuthDetails(req.Context(), details))
								})

								It("responds successfully when the client does not return device logs", func() {
									client.ListDeviceLogsOutputs = []blobTest.ListDeviceLogsOutput{{DeviceLogs: blob.DeviceLogsBlobArray{}, Error: nil}}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
									Expect(res.WriteInputs).To(HaveLen(1))
									Expect(res.WriteInputs[0]).To(MatchJSON("[]"))
								})

								It("responds successfully when the client returns device logs", func() {
									logs := blobTest.RandomDeviceLogsArray(1, 4)
									client.ListDeviceLogsOutputs = []blobTest.ListDeviceLogsOutput{{DeviceLogs: logs, Error: nil}}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
									Expect(res.WriteInputs).To(HaveLen(1))
									Expect(json.Marshal(logs)).To(MatchJSON(res.WriteInputs[0]))
								})
							})

							Context("with user details", func() {
								var sharerUserID string
								BeforeEach(func() {
									sharerUserID = userTest.RandomID()
									details = request.NewAuthDetails(request.MethodSessionToken, userID, authTest.NewSessionToken())
									req.Request = req.WithContext(request.NewContextWithAuthDetails(req.Context(), details))
									granteeToGrantorPerms := map[string]map[string]permission.Permissions{
										userID: {
											userID: permission.Permissions{
												permission.Owner: permission.Permission{},
											},
											sharerUserID: permission.Permissions{
												permission.Read: permission.Permission{},
											},
										},
									}
									authClient.GetUserPermissionsStub = func(ctx context.Context, requestUserID string, targetUserID string) (permission.Permissions, error) {
										if perms, ok := granteeToGrantorPerms[requestUserID]; ok {
											return perms[targetUserID], nil
										}
										return nil, nil
									}
								})

								It("responds successfully when the own user doesn't have device logs", func() {
									client.ListDeviceLogsOutputs = []blobTest.ListDeviceLogsOutput{{DeviceLogs: blob.DeviceLogsBlobArray{}, Error: nil}}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
									Expect(res.WriteInputs).To(HaveLen(1))
									Expect(res.WriteInputs[0]).To(MatchJSON("[]"))
								})

								It("responds successfully with own user's device logs", func() {
									logs := blobTest.RandomDeviceLogsArray(1, 4)
									client.ListDeviceLogsOutputs = []blobTest.ListDeviceLogsOutput{{DeviceLogs: logs, Error: nil}}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
									Expect(res.WriteInputs).To(HaveLen(1))
									Expect(json.Marshal(logs)).To(MatchJSON(res.WriteInputs[0]))
								})

								It("responds successfully when user has access to another person's device logs", func() {
									req.Method = http.MethodGet
									req.URL.Path = fmt.Sprintf("/v1/users/%s/device_logs", sharerUserID)
									logs := blobTest.RandomDeviceLogsArray(1, 4)
									client.ListDeviceLogsOutputs = []blobTest.ListDeviceLogsOutput{{DeviceLogs: logs, Error: nil}}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
									Expect(res.WriteInputs).To(HaveLen(1))
									Expect(json.Marshal(logs)).To(MatchJSON(res.WriteInputs[0]))
								})

								It("responds with forbidden when user doesn't have access to another person's device logs", func() {
									otherUserID := userTest.RandomID()
									provider.BlobClientOutputs = nil
									req.Method = http.MethodGet
									req.URL.Path = fmt.Sprintf("/v1/users/%s/device_logs", otherUserID)
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
									res.WriteOutputs = nil
								})
							})
						})
					})
				})

				Context("GetDeviceLogsContent", func() {
					var id string
					BeforeEach(func() {
						id = blobTest.RandomID()
						req.Method = http.MethodGet
						req.URL.Path = fmt.Sprintf("/v1/device_logs/%s/content", id)
					})

					When("responds", func() {
						BeforeEach(func() {
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
						})

						Context("with clients", func() {
							var authClient *authTest.Client
							var client *blobTest.Client
							var details request.AuthDetails

							BeforeEach(func() {
								authClient = authTest.NewClient()
								client = blobTest.NewClient()
								provider.BlobClientOutputs = []blob.Client{client}
								provider.AuthClientOutputs = []auth.Client{authClient}
							})

							AfterEach(func() {
								authClient.AssertOutputsEmpty()
								client.AssertOutputsEmpty()
							})

							When("unauthenticated", func() {
								It("responds with an unauthenticated error", func() {
									provider.BlobClientOutputs = nil
									provider.AuthClientOutputs = nil
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusUnauthorized}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorUnauthenticated(), res.WriteInputs[0])
								})
							})

							When("with server details", func() {

								BeforeEach(func() {
									details = request.NewAuthDetails(request.MethodServiceSecret, "", authTest.NewSessionToken())
									req.Request = req.WithContext(request.NewContextWithAuthDetails(req.Context(), details))
								})

								It("responds with not found error when the client does not return a blob", func() {
									client.GetDeviceLogsBlobOutputs = []blobTest.GetDeviceLogsBlobOutput{{}}
									res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
									provider.AuthClientOutputs = nil

									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusNotFound}))
									Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorResourceNotFoundWithID(id), res.WriteInputs[0])
								})

								It("responds successfully with headers", func() {
									deviceLogsBlob := blobTest.RandomDeviceLogsBlob()
									content := blob.NewDeviceLogsContent()
									body := test.RandomBytes()
									content.Body = io.NopCloser(bytes.NewReader(body))
									content.DigestMD5 = pointer.FromString(cryptoTest.RandomBase64EncodedMD5Hash())
									content.MediaType = pointer.FromString(netTest.RandomMediaType())
									content.StartAt = pointer.FromTime(test.RandomTime())
									content.EndAt = pointer.FromTime(test.RandomTime())
									client.GetDeviceLogsBlobOutputs = []blobTest.GetDeviceLogsBlobOutput{{Blob: deviceLogsBlob}}
									client.GetDeviceLogsContentOutputs = []blobTest.GetDeviceLogsContentOutput{{Content: content}}
									res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}

									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
									Expect(res.WriteInputs).To(Equal([][]byte{body}))
									Expect(res.HeaderOutput).To(Equal(&http.Header{
										"Content-Type": []string{*content.MediaType},
										"Digest":       []string{fmt.Sprintf("MD5=%s", *content.DigestMD5)},
										"Start-At":     []string{content.StartAt.Format(time.RFC3339Nano)},
										"End-At":       []string{content.EndAt.Format(time.RFC3339Nano)},
									}))
								})
								It("responds successfully with headers for text/plain", func() {
									deviceLogsBlob := blobTest.RandomDeviceLogsBlobMediaType("text/plain; charset=utf-8")
									content := blob.NewDeviceLogsContent()
									body := test.RandomBytes()
									content.Body = io.NopCloser(bytes.NewReader(body))
									content.DigestMD5 = pointer.FromString(cryptoTest.RandomBase64EncodedMD5Hash())
									content.MediaType = pointer.FromString("text/plain; charset=utf-8")
									content.StartAt = pointer.FromTime(test.RandomTime())
									content.EndAt = pointer.FromTime(test.RandomTime())
									client.GetDeviceLogsBlobOutputs = []blobTest.GetDeviceLogsBlobOutput{{Blob: deviceLogsBlob}}
									client.GetDeviceLogsContentOutputs = []blobTest.GetDeviceLogsContentOutput{{Content: content}}
									res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}

									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
									Expect(res.WriteInputs).To(Equal([][]byte{body}))
									Expect(res.HeaderOutput).To(Equal(&http.Header{
										"Content-Type": []string{"text/plain; charset=utf-8"},
										"Digest":       []string{fmt.Sprintf("MD5=%s", *content.DigestMD5)},
										"Start-At":     []string{content.StartAt.Format(time.RFC3339Nano)},
										"End-At":       []string{content.EndAt.Format(time.RFC3339Nano)},
									}))
								})
							})

							When("with user details", func() {
								var userID string
								var sharerUserID string
								BeforeEach(func() {
									userID = userTest.RandomID()
									sharerUserID = userTest.RandomID()
									details = request.NewAuthDetails(request.MethodSessionToken, userID, authTest.NewSessionToken())
									req.Request = req.WithContext(request.NewContextWithAuthDetails(req.Context(), details))
									granteeToGrantorPerms := map[string]map[string]permission.Permissions{
										userID: {
											userID: permission.Permissions{
												permission.Owner: permission.Permission{},
											},
											sharerUserID: permission.Permissions{
												permission.Read: permission.Permission{},
											},
										},
									}
									authClient.GetUserPermissionsStub = func(ctx context.Context, requestUserID string, targetUserID string) (permission.Permissions, error) {
										if perms, ok := granteeToGrantorPerms[requestUserID]; ok {
											return perms[targetUserID], nil
										}
										return nil, nil
									}
								})

								It("responds with not found error when the client does not return a blob", func() {
									client.GetDeviceLogsBlobOutputs = []blobTest.GetDeviceLogsBlobOutput{{}}
									res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
									provider.AuthClientOutputs = nil

									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusNotFound}))
									Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorResourceNotFoundWithID(id), res.WriteInputs[0])
								})

								It("responds successfully with headers for user's own logs content", func() {
									deviceLogsBlob := blobTest.RandomDeviceLogsBlob()
									deviceLogsBlob.UserID = pointer.FromString(userID)
									content := blob.NewDeviceLogsContent()
									body := test.RandomBytes()
									content.Body = io.NopCloser(bytes.NewReader(body))
									content.DigestMD5 = pointer.FromString(cryptoTest.RandomBase64EncodedMD5Hash())
									content.MediaType = pointer.FromString(netTest.RandomMediaType())
									content.StartAt = pointer.FromTime(test.RandomTime())
									content.EndAt = pointer.FromTime(test.RandomTime())
									client.GetDeviceLogsBlobOutputs = []blobTest.GetDeviceLogsBlobOutput{{Blob: deviceLogsBlob}}
									client.GetDeviceLogsContentOutputs = []blobTest.GetDeviceLogsContentOutput{{Content: content}}
									res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}

									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
									Expect(res.WriteInputs).To(Equal([][]byte{body}))
									Expect(res.HeaderOutput).To(Equal(&http.Header{
										"Content-Type": []string{*content.MediaType},
										"Digest":       []string{fmt.Sprintf("MD5=%s", *content.DigestMD5)},
										"Start-At":     []string{content.StartAt.Format(time.RFC3339Nano)},
										"End-At":       []string{content.EndAt.Format(time.RFC3339Nano)},
									}))
								})
								It("responds successfully with headers for user's own logs content of type text/plain", func() {
									deviceLogsBlob := blobTest.RandomDeviceLogsBlob()
									deviceLogsBlob.UserID = pointer.FromString(userID)
									content := blob.NewDeviceLogsContent()
									body := test.RandomBytes()
									content.Body = io.NopCloser(bytes.NewReader(body))
									content.DigestMD5 = pointer.FromString(cryptoTest.RandomBase64EncodedMD5Hash())
									content.MediaType = pointer.FromString("text/plain; charset=utf-8")
									content.StartAt = pointer.FromTime(test.RandomTime())
									content.EndAt = pointer.FromTime(test.RandomTime())
									client.GetDeviceLogsBlobOutputs = []blobTest.GetDeviceLogsBlobOutput{{Blob: deviceLogsBlob}}
									client.GetDeviceLogsContentOutputs = []blobTest.GetDeviceLogsContentOutput{{Content: content}}
									res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}

									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
									Expect(res.WriteInputs).To(Equal([][]byte{body}))
									Expect(res.HeaderOutput).To(Equal(&http.Header{
										"Content-Type": []string{*content.MediaType},
										"Digest":       []string{fmt.Sprintf("MD5=%s", *content.DigestMD5)},
										"Start-At":     []string{content.StartAt.Format(time.RFC3339Nano)},
										"End-At":       []string{content.EndAt.Format(time.RFC3339Nano)},
									}))
								})

								It("responds successfully when user has access to another person's device logs content", func() {
									deviceLogsBlob := blobTest.RandomDeviceLogsBlob()
									deviceLogsBlob.UserID = pointer.FromString(sharerUserID)
									content := blob.NewDeviceLogsContent()
									body := test.RandomBytes()
									content.Body = io.NopCloser(bytes.NewReader(body))
									content.DigestMD5 = pointer.FromString(cryptoTest.RandomBase64EncodedMD5Hash())
									content.MediaType = pointer.FromString(netTest.RandomMediaType())
									content.StartAt = pointer.FromTime(test.RandomTime())
									content.EndAt = pointer.FromTime(test.RandomTime())
									client.GetDeviceLogsBlobOutputs = []blobTest.GetDeviceLogsBlobOutput{{Blob: deviceLogsBlob}}
									client.GetDeviceLogsContentOutputs = []blobTest.GetDeviceLogsContentOutput{{Content: content}}
									res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}

									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
									Expect(res.WriteInputs).To(Equal([][]byte{body}))
									Expect(res.HeaderOutput).To(Equal(&http.Header{
										"Content-Type": []string{*content.MediaType},
										"Digest":       []string{fmt.Sprintf("MD5=%s", *content.DigestMD5)},
										"Start-At":     []string{content.StartAt.Format(time.RFC3339Nano)},
										"End-At":       []string{content.EndAt.Format(time.RFC3339Nano)},
									}))
								})
							})
						})
					})
				})

				Context("Create", func() {
					BeforeEach(func() {
						req.Method = http.MethodPost
						req.URL.Path = fmt.Sprintf("/v1/users/%s/blobs", userID)
					})

					It("panics when the response is missing", func() {
						Expect(func() { router.Create(nil, req) }).To(Panic())
					})

					It("panics when the request is missing", func() {
						Expect(func() { router.Create(res, nil) }).To(Panic())
					})

					Context("responds with JSON", func() {
						BeforeEach(func() {
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
						})

						AfterEach(func() {
							Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
						})

						When("the path does not contain a user id", func() {
							BeforeEach(func() {
								req.URL.Path = "/v1/users//blobs"
							})

							It("responds with bad request and expected error in body", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorParameterMissing("userId"), res.WriteInputs[0])
							})
						})

						When("the path contains an invalid user id", func() {
							BeforeEach(func() {
								req.URL.Path = "/v1/users/invalid/blobs"
							})

							It("responds with bad request and expected error in body", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorParameterInvalid("userId"), res.WriteInputs[0])
							})
						})

						Context("with content", func() {
							var content *blob.Content

							BeforeEach(func() {
								content = blobTest.RandomContent()
							})

							JustBeforeEach(func() {
								req.Body = io.NopCloser(content.Body)
								if content.DigestMD5 != nil {
									req.Header.Set("Digest", fmt.Sprintf("md5=%s", *content.DigestMD5))
								}
								if content.MediaType != nil {
									req.Header.Set("Content-Type", *content.MediaType)
								}
							})

							When("the digest header is invalid", func() {
								BeforeEach(func() {
									content.DigestMD5 = pointer.FromString("invalid")
								})

								It("responds with bad request and expected error in body", func() {
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorHeaderInvalid("Digest"), res.WriteInputs[0])
								})
							})

							When("the digest header is invalid with multiple values", func() {
								JustBeforeEach(func() {
									req.Header.Add("Digest", fmt.Sprintf("md5=%s", *content.DigestMD5))
								})

								It("responds with bad request and expected error in body", func() {
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorHeaderInvalid("Digest"), res.WriteInputs[0])
								})
							})

							When("the content type header is missing", func() {
								BeforeEach(func() {
									content.MediaType = nil
								})

								It("responds with bad request and expected error in body", func() {
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorHeaderMissing("Content-Type"), res.WriteInputs[0])
								})
							})

							When("the content type header is invalid", func() {
								BeforeEach(func() {
									content.MediaType = pointer.FromString("/")
								})

								It("responds with bad request and expected error in body", func() {
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorHeaderInvalid("Content-Type"), res.WriteInputs[0])
								})
							})

							When("the content type header is invalid with multiple values", func() {
								JustBeforeEach(func() {
									req.Header.Add("Content-Type", *content.MediaType)
								})

								It("responds with bad request and expected error in body", func() {
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorHeaderInvalid("Content-Type"), res.WriteInputs[0])
								})
							})

							Context("with clients", func() {
								var authClient *authTest.Client
								var client *blobTest.Client

								BeforeEach(func() {
									authClient = authTest.NewClient()
									client = blobTest.NewClient()
									provider.BlobClientOutputs = []blob.Client{client}
									provider.AuthClientOutputs = []auth.Client{authClient}
								})

								BeforeEach(func() {
									authClient.EnsureAuthorizedUserOutputs = []authTest.EnsureAuthorizedUserOutput{{
										AuthorizedUserID: userID,
										Error:            nil,
									}}
								})

								AfterEach(func() {
									Expect(authClient.EnsureAuthorizedUserInputs[0].TargetUserID).To(Equal(userID))
									Expect(authClient.EnsureAuthorizedUserInputs[0].AuthorizedPermission).To(Equal(permission.Write))
									authClient.AssertOutputsEmpty()
									client.AssertOutputsEmpty()
								})

								When("the client returns an unauthorized error", func() {
									It("responds with an unauthorized error", func() {
										provider.BlobClientOutputs = nil
										authClient.EnsureAuthorizedUserOutputs = []authTest.EnsureAuthorizedUserOutput{{
											AuthorizedUserID: "",
											Error:            request.ErrorUnauthorized(),
										}}
										handlerFunc(res, req)
										Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
										Expect(res.WriteInputs).To(HaveLen(1))
										errorsTest.ExpectErrorJSON(request.ErrorUnauthorized(), res.WriteInputs[0])
									})
								})

								digestAssertions := func() {
									It("responds with a bad request error when the client returns a digests not equal error", func() {
										err := request.ErrorDigestsNotEqual(cryptoTest.RandomBase64EncodedMD5Hash(), cryptoTest.RandomBase64EncodedMD5Hash())
										client.CreateOutputs = []blobTest.CreateOutput{{Blob: nil, Error: err}}
										handlerFunc(res, req)
										Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
										Expect(res.WriteInputs).To(HaveLen(1))
										errorsTest.ExpectErrorJSON(err, res.WriteInputs[0])
									})

									It("responds with an internal server error when the client returns an unknown error", func() {
										client.CreateOutputs = []blobTest.CreateOutput{{Blob: nil, Error: errorsTest.RandomError()}}
										handlerFunc(res, req)
										Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusInternalServerError}))
										Expect(res.WriteInputs).To(HaveLen(1))
										errorsTest.ExpectErrorJSON(request.ErrorInternalServerError(nil), res.WriteInputs[0])
									})

									It("responds successfully", func() {
										responseResult := blobTest.RandomBlob()
										client.CreateOutputs = []blobTest.CreateOutput{{Blob: responseResult, Error: nil}}
										handlerFunc(res, req)
										Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusCreated}))
										Expect(res.WriteInputs).To(HaveLen(1))
										Expect(json.Marshal(responseResult)).To(MatchJSON(res.WriteInputs[0]))
									})
								}

								When("the digest header is not specified", func() {
									BeforeEach(func() {
										content.DigestMD5 = nil
									})

									AfterEach(func() {
										Expect(client.CreateInputs).To(Equal([]blobTest.CreateInput{{
											UserID: userID,
											Content: &blob.Content{
												Body:      io.NopCloser(content.Body),
												DigestMD5: nil,
												MediaType: content.MediaType,
											},
										}}))
									})

									digestAssertions()
								})

								When("the digest header is specified", func() {
									AfterEach(func() {
										Expect(client.CreateInputs).To(Equal([]blobTest.CreateInput{{
											UserID: userID,
											Content: &blob.Content{
												Body:      io.NopCloser(content.Body),
												DigestMD5: content.DigestMD5,
												MediaType: content.MediaType,
											},
										}}))
									})

									digestAssertions()
								})
							})
						})
					})
				})

				Context("CreateDeviceLogs", func() {
					BeforeEach(func() {
						req.Method = http.MethodPost
						req.URL.Path = fmt.Sprintf("/v1/users/%s/device_logs", userID)
					})

					It("panics when the response is missing", func() {
						Expect(func() { router.Create(nil, req) }).To(Panic())
					})

					It("panics when the request is missing", func() {
						Expect(func() { router.Create(res, nil) }).To(Panic())
					})

					Context("responds with JSON", func() {
						BeforeEach(func() {
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
						})

						AfterEach(func() {
							Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
						})

						When("the path does not contain a user id", func() {
							BeforeEach(func() {
								req.URL.Path = "/v1/users//device_logs"
							})

							It("responds with bad request and expected error in body", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorParameterMissing("userId"), res.WriteInputs[0])
							})
						})

						When("the path contains an invalid user id", func() {
							BeforeEach(func() {
								req.URL.Path = "/v1/users/invalid/device_logs"
							})

							It("responds with bad request and expected error in body", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorParameterInvalid("userId"), res.WriteInputs[0])
							})
						})

						Context("with content", func() {
							var content *blob.DeviceLogsContent

							BeforeEach(func() {
								content = blobTest.RandomDeviceLogsContent()
							})

							JustBeforeEach(func() {
								req.Body = io.NopCloser(content.Body)
								if content.DigestMD5 != nil {
									req.Header.Set("Digest", fmt.Sprintf("md5=%s", *content.DigestMD5))
								}
								if content.MediaType != nil {
									req.Header.Set("Content-Type", *content.MediaType)
								}
								if content.EndAt != nil {
									req.Header.Set("X-Logs-End-At-Time", content.EndAt.Format(time.RFC3339Nano))
								}
								if content.StartAt != nil {
									req.Header.Set("X-Logs-Start-At-Time", content.StartAt.Format(time.RFC3339Nano))
								}
							})

							When("the digest header is invalid", func() {
								BeforeEach(func() {
									content.DigestMD5 = pointer.FromString("invalid")
								})

								It("responds with bad request and expected error in body", func() {
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorHeaderInvalid("Digest"), res.WriteInputs[0])
								})
							})

							When("the digest header is invalid with multiple values", func() {
								JustBeforeEach(func() {
									req.Header.Add("Digest", fmt.Sprintf("md5=%s", *content.DigestMD5))
								})

								It("responds with bad request and expected error in body", func() {
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorHeaderInvalid("Digest"), res.WriteInputs[0])
								})
							})
							When("the digest header is missing", func() {
								BeforeEach(func() {
									content.DigestMD5 = nil
								})

								It("responds with bad request and expected error in body", func() {
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorHeaderMissing("Digest"), res.WriteInputs[0])
								})
							})

							When("the content type header is missing", func() {
								BeforeEach(func() {
									content.MediaType = nil
								})

								It("responds with bad request and expected error in body", func() {
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorHeaderMissing("Content-Type"), res.WriteInputs[0])
								})
							})

							When("the content type header is invalid", func() {
								BeforeEach(func() {
									content.MediaType = pointer.FromString("invalid type")
								})

								It("responds with bad request and expected error in body", func() {
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorHeaderInvalid("Content-Type"), res.WriteInputs[0])
								})
							})

							When("the content type header is invalid with multiple values", func() {
								JustBeforeEach(func() {
									req.Header.Add("Content-Type", *content.MediaType)
								})

								It("responds with bad request and expected error in body", func() {
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorHeaderInvalid("Content-Type"), res.WriteInputs[0])
								})
							})

							Context("with clients", func() {
								var authClient *authTest.Client
								var client *blobTest.Client

								BeforeEach(func() {
									authClient = authTest.NewClient()
									client = blobTest.NewClient()
									provider.BlobClientOutputs = []blob.Client{client}
									provider.AuthClientOutputs = []auth.Client{authClient}
								})

								BeforeEach(func() {
									authClient.EnsureAuthorizedUserOutputs = []authTest.EnsureAuthorizedUserOutput{{
										AuthorizedUserID: userID,
										Error:            nil,
									}}
								})

								AfterEach(func() {
									Expect(authClient.EnsureAuthorizedUserInputs[0].TargetUserID).To(Equal(userID))
									Expect(authClient.EnsureAuthorizedUserInputs[0].AuthorizedPermission).To(Equal(permission.Write))
									authClient.AssertOutputsEmpty()
									client.AssertOutputsEmpty()
								})

								When("the client returns an unauthorized error", func() {
									It("responds with an unauthorized error", func() {
										provider.BlobClientOutputs = nil
										authClient.EnsureAuthorizedUserOutputs = []authTest.EnsureAuthorizedUserOutput{{
											AuthorizedUserID: "",
											Error:            request.ErrorUnauthorized(),
										}}
										handlerFunc(res, req)
										Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
										Expect(res.WriteInputs).To(HaveLen(1))
										errorsTest.ExpectErrorJSON(request.ErrorUnauthorized(), res.WriteInputs[0])
									})
								})

								When("the digest header is specified", func() {
									AfterEach(func() {
										Expect(client.CreateDeviceLogsInputs).To(Equal([]blobTest.CreateDeviceLogsInput{{
											UserID: userID,
											Content: &blob.DeviceLogsContent{
												Body:      io.NopCloser(content.Body),
												DigestMD5: content.DigestMD5,
												MediaType: content.MediaType,
												StartAt:   content.StartAt,
												EndAt:     content.EndAt,
											},
										}}))
									})
									It("responds with a bad request error when the client returns a digests not equal error", func() {
										err := request.ErrorDigestsNotEqual(cryptoTest.RandomBase64EncodedMD5Hash(), cryptoTest.RandomBase64EncodedMD5Hash())
										client.CreateDeviceLogsOutputs = []blobTest.CreateDeviceLogsOutput{{Blob: nil, Error: err}}
										handlerFunc(res, req)
										Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
										Expect(res.WriteInputs).To(HaveLen(1))
										errorsTest.ExpectErrorJSON(err, res.WriteInputs[0])
									})

									It("responds with an internal server error when the client returns an unknown error", func() {
										client.CreateDeviceLogsOutputs = []blobTest.CreateDeviceLogsOutput{{Blob: nil, Error: errorsTest.RandomError()}}
										handlerFunc(res, req)
										Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusInternalServerError}))
										Expect(res.WriteInputs).To(HaveLen(1))
										errorsTest.ExpectErrorJSON(request.ErrorInternalServerError(nil), res.WriteInputs[0])
									})

									It("responds successfully", func() {
										responseResult := blobTest.RandomDeviceLogsBlob()
										client.CreateDeviceLogsOutputs = []blobTest.CreateDeviceLogsOutput{{Blob: responseResult, Error: nil}}
										handlerFunc(res, req)
										Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusCreated}))
										Expect(res.WriteInputs).To(HaveLen(1))
										Expect(json.Marshal(responseResult)).To(MatchJSON(res.WriteInputs[0]))
									})
								})
							})
						})
					})
				})

				Context("DeleteAll", func() {
					BeforeEach(func() {
						req.Method = http.MethodDelete
						req.URL.Path = fmt.Sprintf("/v1/users/%s/blobs", userID)
					})

					It("panics when the response is missing", func() {
						Expect(func() { router.DeleteAll(nil, req) }).To(Panic())
					})

					It("panics when the request is missing", func() {
						Expect(func() { router.DeleteAll(res, nil) }).To(Panic())
					})

					When("responds", func() {
						BeforeEach(func() {
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
						})

						When("the path does not contain a user id", func() {
							BeforeEach(func() {
								req.URL.Path = "/v1/users//blobs"
							})

							It("responds with bad request and expected error in body", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
								Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorParameterMissing("userId"), res.WriteInputs[0])
							})
						})

						When("the path contains an invalid user id", func() {
							BeforeEach(func() {
								req.URL.Path = "/v1/users/invalid/blobs"
							})

							It("responds with bad request and expected error in body", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
								Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorParameterInvalid("userId"), res.WriteInputs[0])
							})
						})

						Context("with clients", func() {
							var authClient *authTest.Client
							var client *blobTest.Client

							BeforeEach(func() {
								authClient = authTest.NewClient()
								client = blobTest.NewClient()
								provider.BlobClientOutputs = []blob.Client{client}
								provider.AuthClientOutputs = []auth.Client{authClient}
							})

							BeforeEach(func() {
								authClient.EnsureAuthorizedServiceOutputs = []error{nil}
							})

							AfterEach(func() {
								authClient.AssertOutputsEmpty()
								client.AssertOutputsEmpty()
							})

							When("the client returns an unauthorized error", func() {
								It("responds with an unauthorized error", func() {
									provider.BlobClientOutputs = nil
									authClient.EnsureAuthorizedServiceOutputs = []error{request.ErrorUnauthorized()}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorUnauthorized(), res.WriteInputs[0])
								})
							})

							It("responds with an internal server error when the client returns an unknown error", func() {
								client.DeleteAllOutputs = []error{errorsTest.RandomError()}
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusInternalServerError}))
								Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorInternalServerError(nil), res.WriteInputs[0])
							})

							It("responds successfully", func() {
								res.WriteOutputs = nil
								client.DeleteAllOutputs = []error{nil}
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusNoContent}))
								Expect(res.HeaderOutput).To(Equal(&http.Header{}))
							})
						})
					})
				})
			})

			Context("with id", func() {
				var id string

				BeforeEach(func() {
					id = blobTest.RandomID()
				})

				Context("Get", func() {
					BeforeEach(func() {
						req.Method = http.MethodGet
						req.URL.Path = fmt.Sprintf("/v1/blobs/%s", id)
					})

					It("panics when the response is missing", func() {
						Expect(func() { router.Get(nil, req) }).To(Panic())
					})

					It("panics when the request is missing", func() {
						Expect(func() { router.Get(res, nil) }).To(Panic())
					})

					Context("responds with JSON", func() {
						BeforeEach(func() {
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
						})

						AfterEach(func() {
							Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
						})

						When("the path contains an invalid id", func() {
							BeforeEach(func() {
								req.URL.Path = "/v1/blobs/invalid"
							})

							It("responds with bad request and expected error in body", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorParameterInvalid("id"), res.WriteInputs[0])
							})
						})

						Context("with clients", func() {
							var authClient *authTest.Client
							var client *blobTest.Client

							BeforeEach(func() {
								authClient = authTest.NewClient()
								client = blobTest.NewClient()
								provider.BlobClientOutputs = []blob.Client{client}
								provider.AuthClientOutputs = []auth.Client{authClient}
							})

							BeforeEach(func() {
								authClient.EnsureAuthorizedServiceOutputs = []error{nil}
							})

							When("the client returns an unauthorized error", func() {
								AfterEach(func() {
									authClient.AssertOutputsEmpty()
									client.AssertOutputsEmpty()
								})

								It("responds with an unauthorized error", func() {
									provider.BlobClientOutputs = nil
									authClient.EnsureAuthorizedServiceOutputs = []error{request.ErrorUnauthorized()}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorUnauthorized(), res.WriteInputs[0])
								})
							})

							When("the user is authorized", func() {
								AfterEach(func() {
									Expect(client.GetInputs).To(Equal([]string{id}))
								})

								It("responds with an internal server error when the client returns an unknown error", func() {
									client.GetOutputs = []blobTest.GetOutput{{Blob: nil, Error: errorsTest.RandomError()}}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusInternalServerError}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorInternalServerError(nil), res.WriteInputs[0])
								})

								It("responds with not found error when the client does not return a blob", func() {
									client.GetOutputs = []blobTest.GetOutput{{Blob: nil, Error: nil}}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusNotFound}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorResourceNotFoundWithID(id), res.WriteInputs[0])
								})

								It("responds successfully", func() {
									responseResult := blobTest.RandomBlob()
									client.GetOutputs = []blobTest.GetOutput{{Blob: responseResult, Error: nil}}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
									Expect(res.WriteInputs).To(HaveLen(1))
									Expect(json.Marshal(responseResult)).To(MatchJSON(res.WriteInputs[0]))
								})
							})
						})
					})
				})

				Context("GetContent", func() {
					BeforeEach(func() {
						req.Method = http.MethodGet
						req.URL.Path = fmt.Sprintf("/v1/blobs/%s/content", id)
					})

					It("panics when the response is missing", func() {
						Expect(func() { router.GetContent(nil, req) }).To(Panic())
					})

					It("panics when the request is missing", func() {
						Expect(func() { router.GetContent(res, nil) }).To(Panic())
					})

					When("responds", func() {
						BeforeEach(func() {
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
						})

						When("the path does not contain an id", func() {
							BeforeEach(func() {
								req.URL.Path = "/v1/blobs//content"
							})

							It("responds with bad request and expected error in body", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
								Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorParameterMissing("id"), res.WriteInputs[0])
							})
						})

						When("the path contains an invalid id", func() {
							BeforeEach(func() {
								req.URL.Path = "/v1/blobs/invalid/content"
							})

							It("responds with bad request and expected error in body", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
								Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorParameterInvalid("id"), res.WriteInputs[0])
							})
						})

						Context("with clients", func() {
							var authClient *authTest.Client
							var client *blobTest.Client

							BeforeEach(func() {
								authClient = authTest.NewClient()
								client = blobTest.NewClient()
								provider.BlobClientOutputs = []blob.Client{client}
								provider.AuthClientOutputs = []auth.Client{authClient}
							})

							BeforeEach(func() {
								authClient.EnsureAuthorizedServiceOutputs = []error{nil}
							})

							AfterEach(func() {
								authClient.AssertOutputsEmpty()
								client.AssertOutputsEmpty()
							})

							When("the client returns an unauthorized error", func() {
								It("responds with an unauthorized error", func() {
									provider.BlobClientOutputs = nil
									authClient.EnsureAuthorizedServiceOutputs = []error{request.ErrorUnauthorized()}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorUnauthorized(), res.WriteInputs[0])
								})
							})

							When("the user is authorized", func() {
								AfterEach(func() {
									Expect(client.GetContentInputs).To(Equal([]string{id}))
								})

								It("responds with an internal server error when the client returns an unknown error", func() {
									client.GetContentOutputs = []blobTest.GetContentOutput{{Content: nil, Error: errorsTest.RandomError()}}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusInternalServerError}))
									Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorInternalServerError(nil), res.WriteInputs[0])
								})

								It("responds with not found error when the client does not return a blob", func() {
									client.GetContentOutputs = []blobTest.GetContentOutput{{Content: nil, Error: nil}}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusNotFound}))
									Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorResourceNotFoundWithID(id), res.WriteInputs[0])
								})

								It("responds successfully without headers", func() {
									body := test.RandomBytes()
									content := blob.NewContent()
									content.Body = io.NopCloser(bytes.NewReader(body))
									client.GetContentOutputs = []blobTest.GetContentOutput{{Content: content, Error: nil}}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
									Expect(res.WriteInputs).To(Equal([][]byte{body}))
									Expect(res.HeaderOutput).To(Equal(&http.Header{}))
								})

								It("responds successfully with headers", func() {
									body := test.RandomBytes()
									content := blob.NewContent()
									content.Body = io.NopCloser(bytes.NewReader(body))
									content.DigestMD5 = pointer.FromString(cryptoTest.RandomBase64EncodedMD5Hash())
									content.MediaType = pointer.FromString(netTest.RandomMediaType())
									client.GetContentOutputs = []blobTest.GetContentOutput{{Content: content, Error: nil}}
									res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
									Expect(res.WriteInputs).To(Equal([][]byte{body}))
									Expect(res.HeaderOutput).To(Equal(&http.Header{
										"Content-Type": []string{*content.MediaType},
										"Digest":       []string{fmt.Sprintf("MD5=%s", *content.DigestMD5)},
									}))
								})
							})
						})
					})
				})

				Context("Delete", func() {
					var revision *int

					BeforeEach(func() {
						revision = pointer.FromInt(requestTest.RandomRevision())
						req.Method = http.MethodDelete
						req.URL.Path = fmt.Sprintf("/v1/blobs/%s", id)
					})

					JustBeforeEach(func() {
						if revision != nil {
							query := url.Values{"revision": []string{strconv.Itoa(*revision)}}
							req.URL.RawQuery = query.Encode()
						}
					})

					It("panics when the response is missing", func() {
						Expect(func() { router.Delete(nil, req) }).To(Panic())
					})

					It("panics when the request is missing", func() {
						Expect(func() { router.Delete(res, nil) }).To(Panic())
					})

					Context("responds", func() {
						BeforeEach(func() {
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
						})

						When("the path contains an invalid id", func() {
							BeforeEach(func() {
								req.URL.Path = "/v1/blobs/invalid"
							})

							It("responds with bad request and expected error in body", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
								Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorParameterInvalid("id"), res.WriteInputs[0])
							})
						})

						When("the query contains an invalid revision", func() {
							BeforeEach(func() {
								revision = pointer.FromInt(-1)
							})

							It("responds with bad request and expected error in body", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
								Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(errorsTest.WithParameterSource(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "revision"), res.WriteInputs[0])
							})
						})

						Context("with client", func() {
							var authClient *authTest.Client
							var client *blobTest.Client

							BeforeEach(func() {
								authClient = authTest.NewClient()
								client = blobTest.NewClient()
								provider.BlobClientOutputs = []blob.Client{client}
								provider.AuthClientOutputs = []auth.Client{authClient}
							})

							BeforeEach(func() {
								authClient.EnsureAuthorizedServiceOutputs = []error{nil}
							})

							AfterEach(func() {
								authClient.AssertOutputsEmpty()
								client.AssertOutputsEmpty()
							})

							When("the client returns an unauthorized error", func() {
								It("responds with an unauthorized error", func() {
									provider.BlobClientOutputs = nil
									authClient.EnsureAuthorizedServiceOutputs = []error{request.ErrorUnauthorized()}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorUnauthorized(), res.WriteInputs[0])
								})
							})

							revisionAssertions := func() {
								Context("deletes", func() {
									It("responds with an internal server error when the client returns an unknown error", func() {
										client.DeleteOutputs = []blobTest.DeleteOutput{{Deleted: false, Error: errorsTest.RandomError()}}
										handlerFunc(res, req)
										Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusInternalServerError}))
										Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
										Expect(res.WriteInputs).To(HaveLen(1))
										errorsTest.ExpectErrorJSON(request.ErrorInternalServerError(nil), res.WriteInputs[0])
									})

									It("responds with not found error when the client does not return a blob", func() {
										client.DeleteOutputs = []blobTest.DeleteOutput{{Deleted: false, Error: nil}}
										handlerFunc(res, req)
										Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusNotFound}))
										Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
										Expect(res.WriteInputs).To(HaveLen(1))
										errorsTest.ExpectErrorJSON(request.ErrorResourceNotFoundWithIDAndOptionalRevision(id, revision), res.WriteInputs[0])
									})

									It("responds successfully", func() {
										res.WriteOutputs = nil
										client.DeleteOutputs = []blobTest.DeleteOutput{{Deleted: true, Error: nil}}
										handlerFunc(res, req)
										Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusNoContent}))
										Expect(res.HeaderOutput).To(Equal(&http.Header{}))
									})
								})
							}

							When("revision is missing", func() {
								BeforeEach(func() {
									revision = nil
								})

								AfterEach(func() {
									Expect(client.DeleteInputs).To(Equal([]blobTest.DeleteInput{{ID: id, Condition: &request.Condition{}}}))
								})

								revisionAssertions()
							})

							When("revision is present", func() {
								BeforeEach(func() {
									revision = pointer.FromInt(requestTest.RandomRevision())
								})

								AfterEach(func() {
									Expect(client.DeleteInputs).To(Equal([]blobTest.DeleteInput{{ID: id, Condition: &request.Condition{Revision: revision}}}))
								})

								revisionAssertions()
							})
						})
					})
				})
			})
		})
	})
})
