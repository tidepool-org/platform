package v1_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"strconv"

	"github.com/ant0ine/go-json-rest/rest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"github.com/tidepool-org/platform/blob"
	blobServiceAPIV1 "github.com/tidepool-org/platform/blob/service/api/v1"
	blobServiceAPIV1Test "github.com/tidepool-org/platform/blob/service/api/v1/test"
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
	var provider *blobServiceAPIV1Test.Provider

	BeforeEach(func() {
		provider = blobServiceAPIV1Test.NewProvider()
	})

	AfterEach(func() {
		provider.AssertOutputsEmpty()
	})

	Context("NewRouter", func() {
		It("returns an error when the provider is missing", func() {
			router, err := blobServiceAPIV1.NewRouter(nil)
			errorsTest.ExpectEqual(err, errors.New("provider is missing"))
			Expect(router).To(BeNil())
		})

		It("returns successfully", func() {
			Expect(blobServiceAPIV1.NewRouter(provider)).ToNot(BeNil())
		})
	})

	Context("with new router", func() {
		var router *blobServiceAPIV1.Router

		BeforeEach(func() {
			var err error
			router, err = blobServiceAPIV1.NewRouter(provider)
			Expect(err).ToNot(HaveOccurred())
			Expect(router).ToNot(BeNil())
		})

		Context("Routes", func() {
			It("returns the expected routes", func() {
				Expect(router.Routes()).To(ConsistOf(
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodGet), "PathExp": Equal("/v1/users/:userId/blobs")})),
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodPost), "PathExp": Equal("/v1/users/:userId/blobs")})),
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodGet), "PathExp": Equal("/v1/blobs/:id")})),
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodGet), "PathExp": Equal("/v1/blobs/:id/content")})),
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
						AfterEach(func() {
							Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
						})

						When("the path does not contain a user id", func() {
							BeforeEach(func() {
								req.URL.Path = "/v1/users//blobs"
							})

							It("responds with bad request and expected error in body", func() {
								res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
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
								res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
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
								res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
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
								res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
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
								res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(errorsTest.WithParameterSource(structureValidator.ErrorValueNotInRange(0, 1, 100), "size"), res.WriteInputs[0])
							})
						})

						Context("with client", func() {
							var client *blobTest.Client

							BeforeEach(func() {
								client = blobTest.NewClient()
								provider.BlobClientOutputs = []blob.Client{client}
							})

							AfterEach(func() {
								client.AssertOutputsEmpty()
							})

							clientAssertions := func() {
								It("responds with an unauthorized error when the client returns an unauthorized error", func() {
									client.ListOutputs = []blobTest.ListOutput{{Blobs: nil, Error: request.ErrorUnauthorized()}}
									res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorUnauthorized(), res.WriteInputs[0])
								})

								It("responds with an internal server error when the client returns an unknown error", func() {
									client.ListOutputs = []blobTest.ListOutput{{Blobs: nil, Error: errorsTest.RandomError()}}
									res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusInternalServerError}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorInternalServerError(nil), res.WriteInputs[0])
								})

								It("responds successfully when the client does not return blobs", func() {
									client.ListOutputs = []blobTest.ListOutput{{Blobs: blob.Blobs{}, Error: nil}}
									res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
									Expect(res.WriteInputs).To(HaveLen(1))
									Expect(res.WriteInputs[0]).To(MatchJSON("[]"))
								})

								It("responds successfully when the client returns blobs", func() {
									blobs := blobTest.RandomBlobs(1, 4)
									client.ListOutputs = []blobTest.ListOutput{{Blobs: blobs, Error: nil}}
									res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
									Expect(res.WriteInputs).To(HaveLen(1))
									Expect(json.Marshal(blobs)).To(MatchJSON(res.WriteInputs[0]))
								})
							}

							When("the filter and pagination query parameters are not specified", func() {
								AfterEach(func() {
									Expect(client.ListInputs).To(Equal([]blobTest.ListInput{{
										Context:    ctx,
										UserID:     userID,
										Filter:     blob.NewFilter(),
										Pagination: page.NewPagination(),
									}}))
								})

								clientAssertions()
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
										Context: ctx,
										UserID:  userID,
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

								clientAssertions()
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
						AfterEach(func() {
							Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
						})

						When("the path does not contain a user id", func() {
							BeforeEach(func() {
								req.URL.Path = "/v1/users//blobs"
							})

							It("responds with bad request and expected error in body", func() {
								res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
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
								res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorParameterInvalid("userId"), res.WriteInputs[0])
							})
						})

						Context("with create", func() {
							var create *blob.Create

							BeforeEach(func() {
								create = blobTest.RandomCreate()
							})

							JustBeforeEach(func() {
								req.Body = ioutil.NopCloser(create.Body)
								if create.DigestMD5 != nil {
									req.Header.Add("Digest", fmt.Sprintf("md5=%s", *create.DigestMD5))
								}
								if create.MediaType != nil {
									req.Header.Add("Content-Type", *create.MediaType)
								}
							})

							When("the digest header is invalid", func() {
								BeforeEach(func() {
									create.DigestMD5 = pointer.FromString("invalid")
								})

								It("responds with bad request and expected error in body", func() {
									res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorHeaderInvalid("Digest"), res.WriteInputs[0])
								})
							})

							When("the digest header is invalid with multiple values", func() {
								BeforeEach(func() {
									req.Header.Add("Digest", fmt.Sprintf("md5=%s", *create.DigestMD5))
								})

								It("responds with bad request and expected error in body", func() {
									res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorHeaderInvalid("Digest"), res.WriteInputs[0])
								})
							})

							When("the content type header is missing", func() {
								BeforeEach(func() {
									create.MediaType = nil
								})

								It("responds with bad request and expected error in body", func() {
									res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorHeaderMissing("Content-Type"), res.WriteInputs[0])
								})
							})

							When("the content type header is invalid", func() {
								BeforeEach(func() {
									create.MediaType = pointer.FromString("/")
								})

								It("responds with bad request and expected error in body", func() {
									res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorHeaderInvalid("Content-Type"), res.WriteInputs[0])
								})
							})

							When("the content type header is invalid with multiple values", func() {
								BeforeEach(func() {
									req.Header.Add("Content-Type", *create.MediaType)
								})

								It("responds with bad request and expected error in body", func() {
									res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorHeaderInvalid("Content-Type"), res.WriteInputs[0])
								})
							})

							Context("with client", func() {
								var client *blobTest.Client

								BeforeEach(func() {
									client = blobTest.NewClient()
									provider.BlobClientOutputs = []blob.Client{client}
								})

								AfterEach(func() {
									client.AssertOutputsEmpty()
								})

								clientAssertions := func() {
									It("responds with a bad request error when the client returns a digests not equal error", func() {
										err := blob.ErrorDigestsNotEqual(cryptoTest.RandomBase64EncodedMD5Hash(), cryptoTest.RandomBase64EncodedMD5Hash())
										client.CreateOutputs = []blobTest.CreateOutput{{Blob: nil, Error: err}}
										res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
										handlerFunc(res, req)
										Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
										Expect(res.WriteInputs).To(HaveLen(1))
										errorsTest.ExpectErrorJSON(err, res.WriteInputs[0])
									})

									It("responds with an unauthorized error when the client returns an unauthorized error", func() {
										client.CreateOutputs = []blobTest.CreateOutput{{Blob: nil, Error: request.ErrorUnauthorized()}}
										res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
										handlerFunc(res, req)
										Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
										Expect(res.WriteInputs).To(HaveLen(1))
										errorsTest.ExpectErrorJSON(request.ErrorUnauthorized(), res.WriteInputs[0])
									})

									It("responds with an internal server error when the client returns an unknown error", func() {
										client.CreateOutputs = []blobTest.CreateOutput{{Blob: nil, Error: errorsTest.RandomError()}}
										res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
										handlerFunc(res, req)
										Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusInternalServerError}))
										Expect(res.WriteInputs).To(HaveLen(1))
										errorsTest.ExpectErrorJSON(request.ErrorInternalServerError(nil), res.WriteInputs[0])
									})

									It("responds successfully", func() {
										responseResult := blobTest.RandomBlob()
										client.CreateOutputs = []blobTest.CreateOutput{{Blob: responseResult, Error: nil}}
										res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
										handlerFunc(res, req)
										Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusCreated}))
										Expect(res.WriteInputs).To(HaveLen(1))
										Expect(json.Marshal(responseResult)).To(MatchJSON(res.WriteInputs[0]))
									})
								}

								When("the digest header is not specified", func() {
									BeforeEach(func() {
										create.DigestMD5 = nil
									})

									AfterEach(func() {
										Expect(client.CreateInputs).To(Equal([]blobTest.CreateInput{{
											Context: ctx,
											UserID:  userID,
											Create: &blob.Create{
												Body:      ioutil.NopCloser(create.Body),
												DigestMD5: nil,
												MediaType: create.MediaType,
											},
										}}))
									})

									clientAssertions()
								})

								When("the digest header is specified", func() {
									AfterEach(func() {
										Expect(client.CreateInputs).To(Equal([]blobTest.CreateInput{{
											Context: ctx,
											UserID:  userID,
											Create: &blob.Create{
												Body:      ioutil.NopCloser(create.Body),
												DigestMD5: create.DigestMD5,
												MediaType: create.MediaType,
											},
										}}))
									})

									clientAssertions()
								})
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
						AfterEach(func() {
							Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
						})

						When("the path contains an invalid id", func() {
							BeforeEach(func() {
								req.URL.Path = "/v1/blobs/invalid"
							})

							It("responds with bad request and expected error in body", func() {
								res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorParameterInvalid("id"), res.WriteInputs[0])
							})
						})

						Context("with client", func() {
							var client *blobTest.Client

							BeforeEach(func() {
								client = blobTest.NewClient()
								provider.BlobClientOutputs = []blob.Client{client}
							})

							AfterEach(func() {
								Expect(client.GetInputs).To(Equal([]blobTest.GetInput{{Context: ctx, ID: id}}))
								client.AssertOutputsEmpty()
							})

							It("responds with an unauthorized error when the client returns an unauthorized error", func() {
								client.GetOutputs = []blobTest.GetOutput{{Blob: nil, Error: request.ErrorUnauthorized()}}
								res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorUnauthorized(), res.WriteInputs[0])
							})

							It("responds with an internal server error when the client returns an unknown error", func() {
								client.GetOutputs = []blobTest.GetOutput{{Blob: nil, Error: errorsTest.RandomError()}}
								res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusInternalServerError}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorInternalServerError(nil), res.WriteInputs[0])
							})

							It("responds with not found error when the client does not return a blob", func() {
								client.GetOutputs = []blobTest.GetOutput{{Blob: nil, Error: nil}}
								res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusNotFound}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorResourceNotFoundWithID(id), res.WriteInputs[0])
							})

							It("responds successfully", func() {
								responseResult := blobTest.RandomBlob()
								client.GetOutputs = []blobTest.GetOutput{{Blob: responseResult, Error: nil}}
								res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
								Expect(res.WriteInputs).To(HaveLen(1))
								Expect(json.Marshal(responseResult)).To(MatchJSON(res.WriteInputs[0]))
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

					When("the path does not contain an id", func() {
						BeforeEach(func() {
							req.URL.Path = "/v1/blobs//content"
						})

						It("responds with bad request and expected error in body", func() {
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
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
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
							handlerFunc(res, req)
							Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
							Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
							Expect(res.WriteInputs).To(HaveLen(1))
							errorsTest.ExpectErrorJSON(request.ErrorParameterInvalid("id"), res.WriteInputs[0])
						})
					})

					Context("with client", func() {
						var client *blobTest.Client

						BeforeEach(func() {
							client = blobTest.NewClient()
							provider.BlobClientOutputs = []blob.Client{client}
						})

						AfterEach(func() {
							Expect(client.GetContentInputs).To(Equal([]blobTest.GetContentInput{{Context: ctx, ID: id}}))
							client.AssertOutputsEmpty()
						})

						It("responds with an unauthorized error when the client returns an unauthorized error", func() {
							client.GetContentOutputs = []blobTest.GetContentOutput{{Content: nil, Error: request.ErrorUnauthorized()}}
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
							handlerFunc(res, req)
							Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
							Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
							Expect(res.WriteInputs).To(HaveLen(1))
							errorsTest.ExpectErrorJSON(request.ErrorUnauthorized(), res.WriteInputs[0])
						})

						It("responds with an internal server error when the client returns an unknown error", func() {
							client.GetContentOutputs = []blobTest.GetContentOutput{{Content: nil, Error: errorsTest.RandomError()}}
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
							handlerFunc(res, req)
							Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusInternalServerError}))
							Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
							Expect(res.WriteInputs).To(HaveLen(1))
							errorsTest.ExpectErrorJSON(request.ErrorInternalServerError(nil), res.WriteInputs[0])
						})

						It("responds with not found error when the client does not return a blob", func() {
							client.GetContentOutputs = []blobTest.GetContentOutput{{Content: nil, Error: nil}}
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
							handlerFunc(res, req)
							Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusNotFound}))
							Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
							Expect(res.WriteInputs).To(HaveLen(1))
							errorsTest.ExpectErrorJSON(request.ErrorResourceNotFoundWithID(id), res.WriteInputs[0])
						})

						It("responds successfully without headers", func() {
							body := test.RandomBytes()
							content := blob.NewContent()
							content.Body = ioutil.NopCloser(bytes.NewReader(body))
							client.GetContentOutputs = []blobTest.GetContentOutput{{Content: content, Error: nil}}
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
							handlerFunc(res, req)
							Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
							Expect(res.WriteInputs).To(Equal([][]byte{body}))
							Expect(res.HeaderOutput).To(Equal(&http.Header{}))
						})

						It("responds successfully with headers", func() {
							body := test.RandomBytes()
							content := blob.NewContent()
							content.Body = ioutil.NopCloser(bytes.NewReader(body))
							content.DigestMD5 = pointer.FromString(cryptoTest.RandomBase64EncodedMD5Hash())
							content.MediaType = pointer.FromString(netTest.RandomMediaType())
							content.Size = pointer.FromInt(test.RandomIntFromRange(1, 100*1024*1024))
							client.GetContentOutputs = []blobTest.GetContentOutput{{Content: content, Error: nil}}
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
							handlerFunc(res, req)
							Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
							Expect(res.WriteInputs).To(Equal([][]byte{body}))
							Expect(res.HeaderOutput).To(Equal(&http.Header{
								"Content-Length": []string{strconv.Itoa(*content.Size)},
								"Content-Type":   []string{*content.MediaType},
								"Digest":         []string{fmt.Sprintf("MD5=%s", *content.DigestMD5)},
							}))
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

					When("the path contains an invalid id", func() {
						BeforeEach(func() {
							req.URL.Path = "/v1/blobs/invalid"
						})

						It("responds with bad request and expected error in body", func() {
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
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
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
							handlerFunc(res, req)
							Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
							Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
							Expect(res.WriteInputs).To(HaveLen(1))
							errorsTest.ExpectErrorJSON(errorsTest.WithParameterSource(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "revision"), res.WriteInputs[0])
						})
					})

					Context("with client", func() {
						var client *blobTest.Client

						BeforeEach(func() {
							client = blobTest.NewClient()
							provider.BlobClientOutputs = []blob.Client{client}
						})

						AfterEach(func() {
							client.AssertOutputsEmpty()
						})

						deleteAssertions := func() {
							Context("deletes", func() {
								It("responds with an unauthorized error when the client returns an unauthorized error", func() {
									client.DeleteOutputs = []blobTest.DeleteOutput{{Deleted: false, Error: request.ErrorUnauthorized()}}
									res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
									Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorUnauthorized(), res.WriteInputs[0])
								})

								It("responds with an internal server error when the client returns an unknown error", func() {
									client.DeleteOutputs = []blobTest.DeleteOutput{{Deleted: false, Error: errorsTest.RandomError()}}
									res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusInternalServerError}))
									Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorInternalServerError(nil), res.WriteInputs[0])
								})

								It("responds with not found error when the client does not return a blob", func() {
									client.DeleteOutputs = []blobTest.DeleteOutput{{Deleted: false, Error: nil}}
									res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusNotFound}))
									Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorResourceNotFoundWithIDAndOptionalRevision(id, revision), res.WriteInputs[0])
								})

								It("responds successfully", func() {
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
								Expect(client.DeleteInputs).To(Equal([]blobTest.DeleteInput{{Context: ctx, ID: id, Condition: &request.Condition{}}}))
							})

							deleteAssertions()
						})

						When("revision is present", func() {
							BeforeEach(func() {
								revision = pointer.FromInt(requestTest.RandomRevision())
							})

							AfterEach(func() {
								Expect(client.DeleteInputs).To(Equal([]blobTest.DeleteInput{{Context: ctx, ID: id, Condition: &request.Condition{Revision: revision}}}))
							})

							deleteAssertions()
						})
					})
				})
			})
		})
	})
})
