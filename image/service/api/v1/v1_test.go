package v1_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strconv"

	"github.com/mdblp/go-json-rest/rest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	cryptoTest "github.com/tidepool-org/platform/crypto/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/image"
	imageMultipart "github.com/tidepool-org/platform/image/multipart"
	imageMultipartTest "github.com/tidepool-org/platform/image/multipart/test"
	imageServiceApiV1 "github.com/tidepool-org/platform/image/service/api/v1"
	imageServiceApiV1Test "github.com/tidepool-org/platform/image/service/api/v1/test"
	imageTest "github.com/tidepool-org/platform/image/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	netTest "github.com/tidepool-org/platform/net/test"
	"github.com/tidepool-org/platform/page"
	pageTest "github.com/tidepool-org/platform/page/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	requestTest "github.com/tidepool-org/platform/request/test"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
	testRest "github.com/tidepool-org/platform/test/rest"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("V1", func() {
	var provider *imageServiceApiV1Test.Provider

	BeforeEach(func() {
		provider = imageServiceApiV1Test.NewProvider()
	})

	AfterEach(func() {
		provider.AssertOutputsEmpty()
	})

	Context("NewRouter", func() {
		It("returns an error when the provider is missing", func() {
			router, err := imageServiceApiV1.NewRouter(nil)
			errorsTest.ExpectEqual(err, errors.New("provider is missing"))
			Expect(router).To(BeNil())
		})

		It("returns successfully", func() {
			Expect(imageServiceApiV1.NewRouter(provider)).ToNot(BeNil())
		})
	})

	Context("with new router", func() {
		var router *imageServiceApiV1.Router

		BeforeEach(func() {
			var err error
			router, err = imageServiceApiV1.NewRouter(provider)
			Expect(err).ToNot(HaveOccurred())
			Expect(router).ToNot(BeNil())
		})

		Context("Routes", func() {
			It("returns the expected routes", func() {
				Expect(router.Routes()).To(ConsistOf(
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodGet), "PathExp": Equal("/v1/users/:userId/images")})),
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodPost), "PathExp": Equal("/v1/users/:userId/images")})),
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodPost), "PathExp": Equal("/v1/users/:userId/images/metadata")})),
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodPost), "PathExp": Equal("/v1/users/:userId/images/content/:contentIntent")})),
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodDelete), "PathExp": Equal("/v1/users/:userId/images")})),
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodGet), "PathExp": Equal("/v1/images/:id")})),
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodGet), "PathExp": Equal("/v1/images/:id/metadata")})),
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodGet), "PathExp": Equal("/v1/images/:id/content")})),
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodGet), "PathExp": Equal("/v1/images/:id/content/*suffix")})),
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodGet), "PathExp": Equal("/v1/images/:id/rendition/*suffix")})),
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodPut), "PathExp": Equal("/v1/images/:id/metadata")})),
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodPut), "PathExp": Equal("/v1/images/:id/content/:contentIntent")})),
					PointTo(MatchFields(IgnoreExtras, Fields{"HttpMethod": Equal(http.MethodDelete), "PathExp": Equal("/v1/images/:id")})),
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
						req.URL.Path = fmt.Sprintf("/v1/users/%s/images", userID)
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
								req.URL.Path = "/v1/users//images"
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
								req.URL.Path = "/v1/users/invalid/images"
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
								errorsTest.ExpectErrorJSON(errorsTest.WithParameterSource(structureValidator.ErrorValueStringNotOneOf("invalid", image.Statuses()), "status"), res.WriteInputs[0])
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

						Context("with client", func() {
							var client *imageTest.Client

							BeforeEach(func() {
								client = imageTest.NewClient()
								provider.ImageClientOutputs = []image.Client{client}
							})

							AfterEach(func() {
								client.AssertOutputsEmpty()
							})

							parameterAssertions := func() {
								It("responds with an unauthorized error when the client returns an unauthorized error", func() {
									client.ListOutputs = []imageTest.ListOutput{{ImageArray: nil, Error: request.ErrorUnauthorized()}}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorUnauthorized(), res.WriteInputs[0])
								})

								It("responds with an internal server error when the client returns an unknown error", func() {
									client.ListOutputs = []imageTest.ListOutput{{ImageArray: nil, Error: errorsTest.RandomError()}}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusInternalServerError}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorInternalServerError(nil), res.WriteInputs[0])
								})

								It("responds successfully when the client does not return images", func() {
									client.ListOutputs = []imageTest.ListOutput{{ImageArray: image.ImageArray{}, Error: nil}}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
									Expect(res.WriteInputs).To(HaveLen(1))
									Expect(res.WriteInputs[0]).To(MatchJSON("[]"))
								})

								It("responds successfully when the client returns images", func() {
									imageArray := imageTest.RandomImageArray(1, 4)
									client.ListOutputs = []imageTest.ListOutput{{ImageArray: imageArray, Error: nil}}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
									Expect(res.WriteInputs).To(HaveLen(1))
									Expect(json.Marshal(imageArray)).To(MatchJSON(res.WriteInputs[0]))
								})
							}

							When("the filter and pagination query parameters are not specified", func() {
								AfterEach(func() {
									Expect(client.ListInputs).To(Equal([]imageTest.ListInput{{
										UserID:     userID,
										Filter:     image.NewFilter(),
										Pagination: page.NewPagination(),
									}}))
								})

								parameterAssertions()
							})

							When("the filter and pagination query parameters are specified", func() {
								var statuses []string
								var contentIntents []string
								var paige int
								var size int

								BeforeEach(func() {
									statuses = imageTest.RandomStatuses()
									contentIntents = imageTest.RandomContentIntents()
									paige = pageTest.RandomPage()
									size = pageTest.RandomSize()
									query := url.Values{
										"status":        statuses,
										"contentIntent": contentIntents,
										"page":          []string{strconv.Itoa(paige)},
										"size":          []string{strconv.Itoa(size)},
									}
									req.URL.RawQuery = query.Encode()
								})

								AfterEach(func() {
									Expect(client.ListInputs).To(Equal([]imageTest.ListInput{{
										UserID: userID,
										Filter: &image.Filter{
											Status:        pointer.FromStringArray(statuses),
											ContentIntent: pointer.FromStringArray(contentIntents),
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

				Context("Create", func() {
					var contentType string

					BeforeEach(func() {
						contentType = fmt.Sprintf("multipart/form-data; boundary=%s", test.RandomStringFromRangeAndCharset(8, 32, test.CharsetHexidecimalLowercase))
						req.Method = http.MethodPost
						req.URL.Path = fmt.Sprintf("/v1/users/%s/images", userID)
						req.Header.Set("Content-Type", contentType)
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
								req.URL.Path = fmt.Sprintf("/v1/users//images")
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
								req.URL.Path = fmt.Sprintf("/v1/users/invalid/images")
							})

							It("responds with bad request and expected error in body", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorParameterInvalid("userId"), res.WriteInputs[0])
							})
						})

						When("the content type header is specified more than once", func() {
							BeforeEach(func() {
								req.Header.Add("Content-Type", contentType)
							})

							It("responds with bad request and expected error in body", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorHeaderInvalid("Content-Type"), res.WriteInputs[0])
							})
						})

						When("the content type header is missing", func() {
							BeforeEach(func() {
								req.Header.Del("Content-Type")
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
								req.Header.Set("Content-Type", "/")
							})

							It("responds with bad request and expected error in body", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorHeaderInvalid("Content-Type"), res.WriteInputs[0])
							})
						})

						When("the content type header is not multipart", func() {
							var mediaType string

							BeforeEach(func() {
								mediaType = netTest.RandomMediaType()
								req.Header.Set("Content-Type", mediaType)
							})

							It("responds with bad request and expected error in body", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorMediaTypeNotSupported(mediaType), res.WriteInputs[0])
							})
						})

						When("the content type header is multipart, but boundary is missing", func() {

							BeforeEach(func() {
								req.Header.Set("Content-Type", "multipart/form-data")
							})

							It("responds with bad request and expected error in body", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorHeaderInvalid("Content-Type"), res.WriteInputs[0])
							})
						})

						Context("with form decoder", func() {
							var formDecoder *imageMultipartTest.FormDecoder

							BeforeEach(func() {
								formDecoder = imageMultipartTest.NewFormDecoder()
								provider.ImageMultipartFormDecoderOutputs = []imageMultipart.FormDecoder{formDecoder}
								req.Body = ioutil.NopCloser(bytes.NewReader(imageTest.RandomContentBytes()))
							})

							AfterEach(func() {
								Expect(formDecoder.DecodeFormInputs).To(Equal([]imageMultipartTest.DecodeFormInput{{Reader: req.Body, ContentType: contentType}}))
								formDecoder.AssertOutputsEmpty()
							})

							It("responds with bad request and expected error in body when form decoder returns an error", func() {
								err := errorsTest.RandomError()
								formDecoder.DecodeFormOutputs = []imageMultipartTest.DecodeFormOutput{{Metadata: nil, ContentIntent: "", Content: nil, Error: err}}
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(err, res.WriteInputs[0])
							})

							Context("with client", func() {
								var client *imageTest.Client
								var metadata *image.Metadata
								var contentIntent string
								var content *image.Content

								BeforeEach(func() {
									client = imageTest.NewClient()
									provider.ImageClientOutputs = []image.Client{client}
									metadata = imageTest.RandomMetadata()
									contentIntent = imageTest.RandomContentIntent()
									content = imageTest.RandomContent()
									formDecoder.DecodeFormOutputs = []imageMultipartTest.DecodeFormOutput{{Metadata: metadata, ContentIntent: contentIntent, Content: content, Error: nil}}
								})

								AfterEach(func() {
									Expect(client.CreateInputs).To(Equal([]imageTest.CreateInput{{UserID: userID, Metadata: metadata, ContentIntent: contentIntent, Content: content}}))
									client.AssertOutputsEmpty()
								})

								It("responds with a bad request error when the client returns a digests not equal error", func() {
									err := request.ErrorDigestsNotEqual(cryptoTest.RandomBase64EncodedMD5Hash(), cryptoTest.RandomBase64EncodedMD5Hash())
									client.CreateOutputs = []imageTest.CreateOutput{{Image: nil, Error: err}}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(err, res.WriteInputs[0])
								})

								It("responds with a bad request error when the client returns an image content intent unexpected error", func() {
									err := image.ErrorImageContentIntentUnexpected(imageTest.RandomContentIntent())
									client.CreateOutputs = []imageTest.CreateOutput{{Image: nil, Error: err}}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(err, res.WriteInputs[0])
								})

								It("responds with a bad request error when the client returns an image malformed error", func() {
									err := image.ErrorImageMalformed(test.RandomString())
									client.CreateOutputs = []imageTest.CreateOutput{{Image: nil, Error: err}}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(err, res.WriteInputs[0])
								})

								It("responds with an unauthorized error when the client returns an unauthorized error", func() {
									client.CreateOutputs = []imageTest.CreateOutput{{Image: nil, Error: request.ErrorUnauthorized()}}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorUnauthorized(), res.WriteInputs[0])
								})

								It("responds with an internal server error when the client returns an unknown error", func() {
									client.CreateOutputs = []imageTest.CreateOutput{{Image: nil, Error: errorsTest.RandomError()}}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusInternalServerError}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorInternalServerError(nil), res.WriteInputs[0])
								})

								It("responds successfully", func() {
									responseResult := imageTest.RandomImage()
									client.CreateOutputs = []imageTest.CreateOutput{{Image: responseResult, Error: nil}}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusCreated}))
									Expect(res.WriteInputs).To(HaveLen(1))
									Expect(json.Marshal(responseResult)).To(MatchJSON(res.WriteInputs[0]))
								})
							})
						})
					})
				})

				Context("CreateWithMetadata", func() {
					BeforeEach(func() {
						req.Method = http.MethodPost
						req.URL.Path = fmt.Sprintf("/v1/users/%s/images/metadata", userID)
					})

					It("panics when the response is missing", func() {
						Expect(func() { router.CreateWithMetadata(nil, req) }).To(Panic())
					})

					It("panics when the request is missing", func() {
						Expect(func() { router.CreateWithMetadata(res, nil) }).To(Panic())
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
								req.URL.Path = "/v1/users//images/metadata"
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
								req.URL.Path = "/v1/users/invalid/images/metadata"
							})

							It("responds with bad request and expected error in body", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorParameterInvalid("userId"), res.WriteInputs[0])
							})
						})

						When("the body contains malformed JSON", func() {
							BeforeEach(func() {
								req.Body = ioutil.NopCloser(bytes.NewBuffer([]byte("][")))
							})

							It("responds with an bad request error", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorJSONMalformed(), res.WriteInputs[0])
							})
						})

						Context("with metadata", func() {
							var metadata *image.Metadata

							BeforeEach(func() {
								metadata = imageTest.RandomMetadata()
							})

							JustBeforeEach(func() {
								req.Body = ioutil.NopCloser(bytes.NewBuffer(test.MarshalRequestBody(metadata)))
							})

							When("the body contains invalid metadata", func() {
								BeforeEach(func() {
									metadata.Name = pointer.FromString("")
								})

								It("responds with an bad request error", func() {
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/name"), res.WriteInputs[0])
								})
							})

							Context("with client", func() {
								var client *imageTest.Client

								BeforeEach(func() {
									client = imageTest.NewClient()
									provider.ImageClientOutputs = []image.Client{client}
								})

								AfterEach(func() {
									client.AssertOutputsEmpty()
								})

								It("responds with an unauthorized error when the client returns an unauthorized error", func() {
									client.CreateWithMetadataOutputs = []imageTest.CreateWithMetadataOutput{{Image: nil, Error: request.ErrorUnauthorized()}}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorUnauthorized(), res.WriteInputs[0])
								})

								It("responds with an internal server error when the client returns an unknown error", func() {
									client.CreateWithMetadataOutputs = []imageTest.CreateWithMetadataOutput{{Image: nil, Error: errorsTest.RandomError()}}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusInternalServerError}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorInternalServerError(nil), res.WriteInputs[0])
								})

								It("responds successfully", func() {
									responseResult := imageTest.RandomImage()
									client.CreateWithMetadataOutputs = []imageTest.CreateWithMetadataOutput{{Image: responseResult, Error: nil}}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusCreated}))
									Expect(res.WriteInputs).To(HaveLen(1))
									Expect(json.Marshal(responseResult)).To(MatchJSON(res.WriteInputs[0]))
								})
							})
						})
					})
				})

				Context("CreateWithContent", func() {
					var contentIntent string

					contentIntentAssertions := func() {
						BeforeEach(func() {
							req.Method = http.MethodPost
							req.URL.Path = fmt.Sprintf("/v1/users/%s/images/content/%s", userID, contentIntent)
						})

						It("panics when the response is missing", func() {
							Expect(func() { router.CreateWithContent(nil, req) }).To(Panic())
						})

						It("panics when the request is missing", func() {
							Expect(func() { router.CreateWithContent(res, nil) }).To(Panic())
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
									req.URL.Path = fmt.Sprintf("/v1/users//images/content/%s", contentIntent)
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
									req.URL.Path = fmt.Sprintf("/v1/users/invalid/images/content/%s", contentIntent)
								})

								It("responds with bad request and expected error in body", func() {
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorParameterInvalid("userId"), res.WriteInputs[0])
								})
							})

							When("the path contains an invalid content intent", func() {
								BeforeEach(func() {
									req.URL.Path = fmt.Sprintf("/v1/users/%s/images/content/invalid", userID)
								})

								It("responds with bad request and expected error in body", func() {
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorParameterInvalid("contentIntent"), res.WriteInputs[0])
								})
							})

							Context("with content", func() {
								var content *image.Content

								BeforeEach(func() {
									content = imageTest.RandomContent()
								})

								JustBeforeEach(func() {
									req.Body = ioutil.NopCloser(content.Body)
									if content.DigestMD5 != nil {
										req.Header.Set("Digest", fmt.Sprintf("md5=%s", *content.DigestMD5))
									}
									if content.MediaType != nil {
										req.Header.Set("Content-Type", *content.MediaType)
									}
								})

								When("the digest is invalid", func() {
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

								When("the digest is invalid with multiple values", func() {
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

								When("the content type header is unsupported", func() {
									BeforeEach(func() {
										content.MediaType = pointer.FromString("application/octet-stream")
									})

									It("responds with bad request and expected error in body", func() {
										handlerFunc(res, req)
										Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
										Expect(res.WriteInputs).To(HaveLen(1))
										errorsTest.ExpectErrorJSON(request.ErrorMediaTypeNotSupported("application/octet-stream"), res.WriteInputs[0])
									})
								})

								Context("with client", func() {
									var client *imageTest.Client

									BeforeEach(func() {
										client = imageTest.NewClient()
										provider.ImageClientOutputs = []image.Client{client}
									})

									AfterEach(func() {
										client.AssertOutputsEmpty()
									})

									digestAssertions := func() {
										It("responds with a bad request error when the client returns a digests not equal error", func() {
											err := request.ErrorDigestsNotEqual(cryptoTest.RandomBase64EncodedMD5Hash(), cryptoTest.RandomBase64EncodedMD5Hash())
											client.CreateWithContentOutputs = []imageTest.CreateWithContentOutput{{Image: nil, Error: err}}
											handlerFunc(res, req)
											Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
											Expect(res.WriteInputs).To(HaveLen(1))
											errorsTest.ExpectErrorJSON(err, res.WriteInputs[0])
										})

										It("responds with an unauthorized error when the client returns an unauthorized error", func() {
											client.CreateWithContentOutputs = []imageTest.CreateWithContentOutput{{Image: nil, Error: request.ErrorUnauthorized()}}
											handlerFunc(res, req)
											Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
											Expect(res.WriteInputs).To(HaveLen(1))
											errorsTest.ExpectErrorJSON(request.ErrorUnauthorized(), res.WriteInputs[0])
										})

										It("responds with an internal server error when the client returns an unknown error", func() {
											client.CreateWithContentOutputs = []imageTest.CreateWithContentOutput{{Image: nil, Error: errorsTest.RandomError()}}
											handlerFunc(res, req)
											Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusInternalServerError}))
											Expect(res.WriteInputs).To(HaveLen(1))
											errorsTest.ExpectErrorJSON(request.ErrorInternalServerError(nil), res.WriteInputs[0])
										})

										It("responds successfully", func() {
											responseResult := imageTest.RandomImage()
											client.CreateWithContentOutputs = []imageTest.CreateWithContentOutput{{Image: responseResult, Error: nil}}
											handlerFunc(res, req)
											Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusCreated}))
											Expect(res.WriteInputs).To(HaveLen(1))
											Expect(json.Marshal(responseResult)).To(MatchJSON(res.WriteInputs[0]))
										})
									}

									When("the digest is not specified", func() {
										BeforeEach(func() {
											content.DigestMD5 = nil
										})

										AfterEach(func() {
											Expect(client.CreateWithContentInputs).To(Equal([]imageTest.CreateWithContentInput{{
												UserID:        userID,
												ContentIntent: contentIntent,
												Content: &image.Content{
													Body:      ioutil.NopCloser(content.Body),
													DigestMD5: nil,
													MediaType: content.MediaType,
												},
											}}))
										})

										digestAssertions()
									})

									When("the digest is specified", func() {
										AfterEach(func() {
											Expect(client.CreateWithContentInputs).To(Equal([]imageTest.CreateWithContentInput{{
												UserID:        userID,
												ContentIntent: contentIntent,
												Content: &image.Content{
													Body:      ioutil.NopCloser(content.Body),
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
					}

					When("the content intent is alternate", func() {
						BeforeEach(func() {
							contentIntent = image.ContentIntentAlternate
						})

						contentIntentAssertions()
					})

					When("the content intent is original", func() {
						BeforeEach(func() {
							contentIntent = image.ContentIntentOriginal
						})

						contentIntentAssertions()
					})
				})

				Context("DeleteAll", func() {
					BeforeEach(func() {
						req.Method = http.MethodDelete
						req.URL.Path = fmt.Sprintf("/v1/users/%s/images", userID)
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
								req.URL.Path = "/v1/users//images"
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
								req.URL.Path = "/v1/users/invalid/images"
							})

							It("responds with bad request and expected error in body", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
								Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorParameterInvalid("userId"), res.WriteInputs[0])
							})
						})

						Context("with client", func() {
							var client *imageTest.Client

							BeforeEach(func() {
								client = imageTest.NewClient()
								provider.ImageClientOutputs = []image.Client{client}
							})

							AfterEach(func() {
								client.AssertOutputsEmpty()
							})

							It("responds with an unauthorized error when the client returns an unauthorized error", func() {
								client.DeleteAllOutputs = []error{request.ErrorUnauthorized()}
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
								Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorUnauthorized(), res.WriteInputs[0])
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
					id = imageTest.RandomID()
				})

				Context("Get", func() {
					BeforeEach(func() {
						req.Method = http.MethodGet
						req.URL.Path = fmt.Sprintf("/v1/images/%s", id)
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
								req.URL.Path = "/v1/images/invalid"
							})

							It("responds with bad request and expected error in body", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorParameterInvalid("id"), res.WriteInputs[0])
							})
						})

						Context("with client", func() {
							var client *imageTest.Client

							BeforeEach(func() {
								client = imageTest.NewClient()
								provider.ImageClientOutputs = []image.Client{client}
							})

							AfterEach(func() {
								Expect(client.GetInputs).To(Equal([]string{id}))
								client.AssertOutputsEmpty()
							})

							It("responds with an unauthorized error when the client returns an unauthorized error", func() {
								client.GetOutputs = []imageTest.GetOutput{{Image: nil, Error: request.ErrorUnauthorized()}}
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorUnauthorized(), res.WriteInputs[0])
							})

							It("responds with an internal server error when the client returns an unknown error", func() {
								client.GetOutputs = []imageTest.GetOutput{{Image: nil, Error: errorsTest.RandomError()}}
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusInternalServerError}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorInternalServerError(nil), res.WriteInputs[0])
							})

							It("responds with not found error when the client does not return an image", func() {
								client.GetOutputs = []imageTest.GetOutput{{Image: nil, Error: nil}}
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusNotFound}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorResourceNotFoundWithID(id), res.WriteInputs[0])
							})

							It("responds successfully", func() {
								responseResult := imageTest.RandomImage()
								client.GetOutputs = []imageTest.GetOutput{{Image: responseResult, Error: nil}}
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
								Expect(res.WriteInputs).To(HaveLen(1))
								Expect(json.Marshal(responseResult)).To(MatchJSON(res.WriteInputs[0]))
							})
						})
					})
				})

				Context("GetMetadata", func() {
					BeforeEach(func() {
						req.Method = http.MethodGet
						req.URL.Path = fmt.Sprintf("/v1/images/%s/metadata", id)
					})

					It("panics when the response is missing", func() {
						Expect(func() { router.GetMetadata(nil, req) }).To(Panic())
					})

					It("panics when the request is missing", func() {
						Expect(func() { router.GetMetadata(res, nil) }).To(Panic())
					})

					Context("responds with JSON", func() {
						BeforeEach(func() {
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
						})

						AfterEach(func() {
							Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
						})

						When("the path does not contain an id", func() {
							BeforeEach(func() {
								req.URL.Path = "/v1/images//metadata"
							})

							It("responds with bad request and expected error in body", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorParameterMissing("id"), res.WriteInputs[0])
							})
						})

						When("the path contains an invalid id", func() {
							BeforeEach(func() {
								req.URL.Path = "/v1/images/invalid/metadata"
							})

							It("responds with bad request and expected error in body", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorParameterInvalid("id"), res.WriteInputs[0])
							})
						})

						Context("with client", func() {
							var client *imageTest.Client

							BeforeEach(func() {
								client = imageTest.NewClient()
								provider.ImageClientOutputs = []image.Client{client}
							})

							AfterEach(func() {
								Expect(client.GetMetadataInputs).To(Equal([]string{id}))
								client.AssertOutputsEmpty()
							})

							It("responds with an unauthorized error when the client returns an unauthorized error", func() {
								client.GetMetadataOutputs = []imageTest.GetMetadataOutput{{Metadata: nil, Error: request.ErrorUnauthorized()}}
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorUnauthorized(), res.WriteInputs[0])
							})

							It("responds with an internal server error when the client returns an unknown error", func() {
								client.GetMetadataOutputs = []imageTest.GetMetadataOutput{{Metadata: nil, Error: errorsTest.RandomError()}}
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusInternalServerError}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorInternalServerError(nil), res.WriteInputs[0])
							})

							It("responds with not found error when the client does not return an image", func() {
								client.GetMetadataOutputs = []imageTest.GetMetadataOutput{{Metadata: nil, Error: nil}}
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusNotFound}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorResourceNotFoundWithID(id), res.WriteInputs[0])
							})

							It("responds successfully", func() {
								responseResult := imageTest.RandomMetadata()
								client.GetMetadataOutputs = []imageTest.GetMetadataOutput{{Metadata: responseResult, Error: nil}}
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
								Expect(res.WriteInputs).To(HaveLen(1))
								Expect(json.Marshal(responseResult)).To(MatchJSON(res.WriteInputs[0]))
							})
						})
					})
				})

				Context("GetContent", func() {
					var extension *string

					extensionAssertions := func() {
						BeforeEach(func() {
							req.Method = http.MethodGet
						})

						JustBeforeEach(func() {
							if extension != nil {
								req.URL.Path = fmt.Sprintf("/v1/images/%s/content/%s.%s", id, test.RandomStringFromRangeAndCharset(1, 16, test.CharsetAlphaNumeric), *extension)
							} else {
								req.URL.Path = fmt.Sprintf("/v1/images/%s/content", id)
							}
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
									id = ""
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
									id = "invalid"
								})

								It("responds with bad request and expected error in body", func() {
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
									Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorParameterInvalid("id"), res.WriteInputs[0])
								})
							})

							When("the path contains an invalid extension", func() {
								BeforeEach(func() {
									extension = pointer.FromString("bin")
								})

								It("responds with bad request and expected error in body", func() {
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
									Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorExtensionNotSupported("bin"), res.WriteInputs[0])
								})
							})

							Context("with client", func() {
								var client *imageTest.Client

								BeforeEach(func() {
									client = imageTest.NewClient()
									provider.ImageClientOutputs = []image.Client{client}
								})

								AfterEach(func() {
									var mediaType *string
									if extension != nil {
										if value, ok := image.MediaTypeFromExtension(*extension); ok {
											mediaType = pointer.FromString(value)
										}
									}
									Expect(client.GetContentInputs).To(Equal([]imageTest.GetContentInput{{ID: id, MediaType: mediaType}}))
									client.AssertOutputsEmpty()
								})

								It("responds with an unauthorized error when the client returns an unauthorized error", func() {
									client.GetContentOutputs = []imageTest.GetContentOutput{{Content: nil, Error: request.ErrorUnauthorized()}}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
									Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorUnauthorized(), res.WriteInputs[0])
								})

								It("responds with an internal server error when the client returns an unknown error", func() {
									client.GetContentOutputs = []imageTest.GetContentOutput{{Content: nil, Error: errorsTest.RandomError()}}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusInternalServerError}))
									Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorInternalServerError(nil), res.WriteInputs[0])
								})

								It("responds with not found error when the client does not return an image", func() {
									client.GetContentOutputs = []imageTest.GetContentOutput{{Content: nil, Error: nil}}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusNotFound}))
									Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorResourceNotFoundWithID(id), res.WriteInputs[0])
								})

								It("responds successfully with minimum headers", func() {
									body := imageTest.RandomContentBytes()
									content := image.NewContent()
									content.Body = ioutil.NopCloser(bytes.NewReader(body))
									client.GetContentOutputs = []imageTest.GetContentOutput{{Content: content, Error: nil}}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
									Expect(res.WriteInputs).To(Equal([][]byte{body}))
									Expect(res.HeaderOutput).To(Equal(&http.Header{}))
								})

								It("responds successfully with full headers", func() {
									body := imageTest.RandomContentBytes()
									content := image.NewContent()
									content.Body = ioutil.NopCloser(bytes.NewReader(body))
									content.DigestMD5 = pointer.FromString(cryptoTest.RandomBase64EncodedMD5Hash())
									content.MediaType = pointer.FromString(imageTest.RandomMediaType())
									client.GetContentOutputs = []imageTest.GetContentOutput{{Content: content, Error: nil}}
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
					}

					When("the extension is missing", func() {
						BeforeEach(func() {
							extension = nil
						})

						extensionAssertions()
					})

					When("the extension is jpeg", func() {
						BeforeEach(func() {
							extension = pointer.FromString("jpeg")
						})

						extensionAssertions()
					})

					When("the extension is jpg", func() {
						BeforeEach(func() {
							extension = pointer.FromString("jpg")
						})

						extensionAssertions()
					})

					When("the extension is png", func() {
						BeforeEach(func() {
							extension = pointer.FromString("png")
						})

						extensionAssertions()
					})
				})

				Context("GetRenditionContent", func() {
					var rendition *image.Rendition
					var renditionString string

					BeforeEach(func() {
						req.Method = http.MethodGet
						rendition = imageTest.RandomRendition()
						renditionString = rendition.String()
					})

					suffixAssertions := func() {
						It("panics when the response is missing", func() {
							Expect(func() { router.GetRenditionContent(nil, req) }).To(Panic())
						})

						It("panics when the request is missing", func() {
							Expect(func() { router.GetRenditionContent(res, nil) }).To(Panic())
						})

						When("responds", func() {
							BeforeEach(func() {
								res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
							})

							When("the path does not contain an id", func() {
								BeforeEach(func() {
									id = ""
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
									id = "invalid"
								})

								It("responds with bad request and expected error in body", func() {
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
									Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorParameterInvalid("id"), res.WriteInputs[0])
								})
							})

							When("the path contains an invalid rendition", func() {
								BeforeEach(func() {
									renditionString = "invalid.bin"
								})

								It("responds with bad request and expected error in body", func() {
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
									Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(image.ErrorValueRenditionNotParsable("invalid.bin"), res.WriteInputs[0])
								})
							})

							Context("with client", func() {
								var client *imageTest.Client

								BeforeEach(func() {
									client = imageTest.NewClient()
									provider.ImageClientOutputs = []image.Client{client}
								})

								AfterEach(func() {
									Expect(client.GetRenditionContentInputs).To(Equal([]imageTest.GetRenditionContentInput{{ID: id, Rendition: rendition}}))
									client.AssertOutputsEmpty()
								})

								It("responds with an unauthorized error when the client returns an unauthorized error", func() {
									client.GetRenditionContentOutputs = []imageTest.GetRenditionContentOutput{{Content: nil, Error: request.ErrorUnauthorized()}}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
									Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorUnauthorized(), res.WriteInputs[0])
								})

								It("responds with an internal server error when the client returns an unknown error", func() {
									client.GetRenditionContentOutputs = []imageTest.GetRenditionContentOutput{{Content: nil, Error: errorsTest.RandomError()}}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusInternalServerError}))
									Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorInternalServerError(nil), res.WriteInputs[0])
								})

								It("responds with not found error when the client does not return an image", func() {
									client.GetRenditionContentOutputs = []imageTest.GetRenditionContentOutput{{Content: nil, Error: nil}}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusNotFound}))
									Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorResourceNotFoundWithID(id), res.WriteInputs[0])
								})

								It("responds successfully with minimum headers", func() {
									body := imageTest.RandomContentBytes()
									content := image.NewContent()
									content.Body = ioutil.NopCloser(bytes.NewReader(body))
									client.GetRenditionContentOutputs = []imageTest.GetRenditionContentOutput{{Content: content, Error: nil}}
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
									Expect(res.WriteInputs).To(Equal([][]byte{body}))
									Expect(res.HeaderOutput).To(Equal(&http.Header{}))
								})

								It("responds successfully with full headers", func() {
									body := imageTest.RandomContentBytes()
									content := image.NewContent()
									content.Body = ioutil.NopCloser(bytes.NewReader(body))
									content.DigestMD5 = pointer.FromString(cryptoTest.RandomBase64EncodedMD5Hash())
									content.MediaType = pointer.FromString(imageTest.RandomMediaType())
									client.GetRenditionContentOutputs = []imageTest.GetRenditionContentOutput{{Content: content, Error: nil}}
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
					}

					When("there is not a suffix", func() {
						JustBeforeEach(func() {
							req.URL.Path = fmt.Sprintf("/v1/images/%s/rendition/%s", id, renditionString)
						})

						suffixAssertions()
					})

					When("there is a suffix", func() {
						JustBeforeEach(func() {
							extension := path.Ext(renditionString)
							filename := test.RandomStringFromRangeAndCharset(1, 16, test.CharsetAlphaNumeric)
							renditionString = renditionString[:len(renditionString)-len(extension)]
							req.URL.Path = fmt.Sprintf("/v1/images/%s/rendition/%s/%s%s", id, renditionString, filename, extension)
						})

						suffixAssertions()
					})
				})

				Context("PutMetadata", func() {
					var revision *int

					BeforeEach(func() {
						revision = pointer.FromInt(requestTest.RandomRevision())
						req.Method = http.MethodPut
						req.URL.Path = fmt.Sprintf("/v1/images/%s/metadata", id)
					})

					JustBeforeEach(func() {
						if revision != nil {
							query := url.Values{"revision": []string{strconv.Itoa(*revision)}}
							req.URL.RawQuery = query.Encode()
						}
					})

					It("panics when the response is missing", func() {
						Expect(func() { router.PutMetadata(nil, req) }).To(Panic())
					})

					It("panics when the request is missing", func() {
						Expect(func() { router.PutMetadata(res, nil) }).To(Panic())
					})

					Context("responds with JSON", func() {
						BeforeEach(func() {
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
						})

						AfterEach(func() {
							Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
						})

						When("the path does not contain an id", func() {
							BeforeEach(func() {
								req.URL.Path = "/v1/images//metadata"
							})

							It("responds with bad request and expected error in body", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorParameterMissing("id"), res.WriteInputs[0])
							})
						})

						When("the path contains an invalid id", func() {
							BeforeEach(func() {
								req.URL.Path = "/v1/images/invalid/metadata"
							})

							It("responds with bad request and expected error in body", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
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
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(errorsTest.WithParameterSource(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "revision"), res.WriteInputs[0])
							})
						})

						When("the body contains malformed JSON", func() {
							BeforeEach(func() {
								req.Body = ioutil.NopCloser(bytes.NewBuffer([]byte("][")))
							})

							It("responds with an bad request error", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
								Expect(res.WriteInputs).To(HaveLen(1))
								errorsTest.ExpectErrorJSON(request.ErrorJSONMalformed(), res.WriteInputs[0])
							})
						})

						Context("with metadata", func() {
							var metadata *image.Metadata

							BeforeEach(func() {
								metadata = imageTest.RandomMetadata()
							})

							JustBeforeEach(func() {
								req.Body = ioutil.NopCloser(bytes.NewBuffer(test.MarshalRequestBody(metadata)))
							})

							When("the body contains invalid metadata", func() {
								BeforeEach(func() {
									metadata.Name = pointer.FromString("")
								})

								It("responds with an bad request error", func() {
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/name"), res.WriteInputs[0])
								})
							})

							Context("with client", func() {
								var client *imageTest.Client

								BeforeEach(func() {
									client = imageTest.NewClient()
									provider.ImageClientOutputs = []image.Client{client}
								})

								AfterEach(func() {
									Expect(client.PutMetadataInputs).To(Equal([]imageTest.PutMetadataInput{{
										ID: id,
										Condition: &request.Condition{
											Revision: revision,
										},
										Metadata: metadata,
									}}))
									client.AssertOutputsEmpty()
								})

								revisionAssertions := func() {
									It("responds with an unauthorized error when the client returns an unauthorized error", func() {
										client.PutMetadataOutputs = []imageTest.PutMetadataOutput{{Image: nil, Error: request.ErrorUnauthorized()}}
										handlerFunc(res, req)
										Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
										Expect(res.WriteInputs).To(HaveLen(1))
										errorsTest.ExpectErrorJSON(request.ErrorUnauthorized(), res.WriteInputs[0])
									})

									It("responds with an internal server error when the client returns an unknown error", func() {
										client.PutMetadataOutputs = []imageTest.PutMetadataOutput{{Image: nil, Error: errorsTest.RandomError()}}
										handlerFunc(res, req)
										Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusInternalServerError}))
										Expect(res.WriteInputs).To(HaveLen(1))
										errorsTest.ExpectErrorJSON(request.ErrorInternalServerError(nil), res.WriteInputs[0])
									})

									It("responds with not found error when the client does not return an image", func() {
										client.PutMetadataOutputs = []imageTest.PutMetadataOutput{{Image: nil, Error: nil}}
										handlerFunc(res, req)
										Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusNotFound}))
										Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
										Expect(res.WriteInputs).To(HaveLen(1))
										errorsTest.ExpectErrorJSON(request.ErrorResourceNotFoundWithIDAndOptionalRevision(id, revision), res.WriteInputs[0])
									})

									It("responds successfully", func() {
										responseResult := imageTest.RandomImage()
										client.PutMetadataOutputs = []imageTest.PutMetadataOutput{{Image: responseResult, Error: nil}}
										handlerFunc(res, req)
										Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
										Expect(res.WriteInputs).To(HaveLen(1))
										Expect(json.Marshal(responseResult)).To(MatchJSON(res.WriteInputs[0]))
									})
								}

								When("revision is missing", func() {
									BeforeEach(func() {
										revision = nil
									})

									revisionAssertions()
								})

								When("revision is present", func() {
									BeforeEach(func() {
										revision = pointer.FromInt(requestTest.RandomRevision())
									})

									revisionAssertions()
								})
							})
						})
					})
				})

				Context("PutContent", func() {
					var revision *int
					var contentIntent string

					BeforeEach(func() {
						revision = pointer.FromInt(requestTest.RandomRevision())
					})

					contentIntentAssertions := func() {
						BeforeEach(func() {
							req.Method = http.MethodPut
							req.URL.Path = fmt.Sprintf("/v1/images/%s/content/%s", id, contentIntent)
						})

						JustBeforeEach(func() {
							if revision != nil {
								query := url.Values{"revision": []string{strconv.Itoa(*revision)}}
								req.URL.RawQuery = query.Encode()
							}
						})

						It("panics when the response is missing", func() {
							Expect(func() { router.PutContent(nil, req) }).To(Panic())
						})

						It("panics when the request is missing", func() {
							Expect(func() { router.PutContent(res, nil) }).To(Panic())
						})

						Context("responds with JSON", func() {
							BeforeEach(func() {
								res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
							})

							AfterEach(func() {
								Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
							})

							When("the path does not contain an id", func() {
								BeforeEach(func() {
									req.URL.Path = fmt.Sprintf("/v1/images//content/%s", contentIntent)
								})

								It("responds with bad request and expected error in body", func() {
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorParameterMissing("id"), res.WriteInputs[0])
								})
							})

							When("the path contains an invalid id", func() {
								BeforeEach(func() {
									req.URL.Path = fmt.Sprintf("/v1/images/invalid/content/%s", contentIntent)
								})

								It("responds with bad request and expected error in body", func() {
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
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
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(errorsTest.WithParameterSource(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "revision"), res.WriteInputs[0])
								})
							})

							When("the path contains an invalid content intent", func() {
								BeforeEach(func() {
									req.URL.Path = fmt.Sprintf("/v1/images/%s/content/invalid", id)
								})

								It("responds with bad request and expected error in body", func() {
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
									Expect(res.WriteInputs).To(HaveLen(1))
									errorsTest.ExpectErrorJSON(request.ErrorParameterInvalid("contentIntent"), res.WriteInputs[0])
								})
							})

							Context("with content", func() {
								var content *image.Content

								BeforeEach(func() {
									content = imageTest.RandomContent()
								})

								JustBeforeEach(func() {
									req.Body = ioutil.NopCloser(content.Body)
									if content.DigestMD5 != nil {
										req.Header.Set("Digest", fmt.Sprintf("md5=%s", *content.DigestMD5))
									}
									if content.MediaType != nil {
										req.Header.Set("Content-Type", *content.MediaType)
									}
								})

								When("the digest is invalid", func() {
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

								When("the digest is invalid with multiple values", func() {
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

								When("the content type header is unsupported", func() {
									BeforeEach(func() {
										content.MediaType = pointer.FromString("application/octet-stream")
									})

									It("responds with bad request and expected error in body", func() {
										handlerFunc(res, req)
										Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
										Expect(res.WriteInputs).To(HaveLen(1))
										errorsTest.ExpectErrorJSON(request.ErrorMediaTypeNotSupported("application/octet-stream"), res.WriteInputs[0])
									})
								})

								Context("with client", func() {
									var client *imageTest.Client

									BeforeEach(func() {
										client = imageTest.NewClient()
										provider.ImageClientOutputs = []image.Client{client}
									})

									AfterEach(func() {
										client.AssertOutputsEmpty()
									})

									revisionAssertions := func() {
										digestAssertions := func() {
											It("responds with a bad request error when the client returns a digests not equal error", func() {
												err := request.ErrorDigestsNotEqual(cryptoTest.RandomBase64EncodedMD5Hash(), cryptoTest.RandomBase64EncodedMD5Hash())
												client.PutContentOutputs = []imageTest.PutContentOutput{{Image: nil, Error: err}}
												handlerFunc(res, req)
												Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusBadRequest}))
												Expect(res.WriteInputs).To(HaveLen(1))
												errorsTest.ExpectErrorJSON(err, res.WriteInputs[0])
											})

											It("responds with an unauthorized error when the client returns an unauthorized error", func() {
												client.PutContentOutputs = []imageTest.PutContentOutput{{Image: nil, Error: request.ErrorUnauthorized()}}
												handlerFunc(res, req)
												Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
												Expect(res.WriteInputs).To(HaveLen(1))
												errorsTest.ExpectErrorJSON(request.ErrorUnauthorized(), res.WriteInputs[0])
											})

											It("responds with an internal server error when the client returns an unknown error", func() {
												client.PutContentOutputs = []imageTest.PutContentOutput{{Image: nil, Error: errorsTest.RandomError()}}
												handlerFunc(res, req)
												Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusInternalServerError}))
												Expect(res.WriteInputs).To(HaveLen(1))
												errorsTest.ExpectErrorJSON(request.ErrorInternalServerError(nil), res.WriteInputs[0])
											})

											It("responds with not found error when the client does not return an image", func() {
												client.PutContentOutputs = []imageTest.PutContentOutput{{Image: nil, Error: nil}}
												handlerFunc(res, req)
												Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusNotFound}))
												Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
												Expect(res.WriteInputs).To(HaveLen(1))
												errorsTest.ExpectErrorJSON(request.ErrorResourceNotFoundWithIDAndOptionalRevision(id, revision), res.WriteInputs[0])
											})

											It("responds successfully", func() {
												responseResult := imageTest.RandomImage()
												client.PutContentOutputs = []imageTest.PutContentOutput{{Image: responseResult, Error: nil}}
												handlerFunc(res, req)
												Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
												Expect(res.WriteInputs).To(HaveLen(1))
												Expect(json.Marshal(responseResult)).To(MatchJSON(res.WriteInputs[0]))
											})
										}

										When("the digest is not specified", func() {
											BeforeEach(func() {
												content.DigestMD5 = nil
											})

											AfterEach(func() {
												Expect(client.PutContentInputs).To(Equal([]imageTest.PutContentInput{{
													ID: id,
													Condition: &request.Condition{
														Revision: revision,
													},
													ContentIntent: contentIntent,
													Content: &image.Content{
														Body:      ioutil.NopCloser(content.Body),
														DigestMD5: nil,
														MediaType: content.MediaType,
													},
												}}))
											})

											digestAssertions()
										})

										When("the digest is specified", func() {
											AfterEach(func() {
												Expect(client.PutContentInputs).To(Equal([]imageTest.PutContentInput{{
													ID: id,
													Condition: &request.Condition{
														Revision: revision,
													},
													ContentIntent: contentIntent,
													Content: &image.Content{
														Body:      ioutil.NopCloser(content.Body),
														DigestMD5: content.DigestMD5,
														MediaType: content.MediaType,
													},
												}}))
											})

											digestAssertions()
										})
									}

									When("revision is missing", func() {
										BeforeEach(func() {
											revision = nil
										})

										revisionAssertions()
									})

									When("revision is present", func() {
										BeforeEach(func() {
											revision = pointer.FromInt(requestTest.RandomRevision())
										})

										revisionAssertions()
									})
								})
							})
						})
					}

					When("the content intent is alternate", func() {
						BeforeEach(func() {
							contentIntent = image.ContentIntentAlternate
						})

						contentIntentAssertions()
					})

					When("the content intent is original", func() {
						BeforeEach(func() {
							contentIntent = image.ContentIntentOriginal
						})

						contentIntentAssertions()
					})
				})

				Context("Delete", func() {
					var revision *int

					BeforeEach(func() {
						revision = pointer.FromInt(requestTest.RandomRevision())
						req.Method = http.MethodDelete
						req.URL.Path = fmt.Sprintf("/v1/images/%s", id)
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

					When("responds", func() {
						BeforeEach(func() {
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
						})

						When("the path contains an invalid id", func() {
							BeforeEach(func() {
								req.URL.Path = "/v1/images/invalid"
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
							var client *imageTest.Client

							BeforeEach(func() {
								client = imageTest.NewClient()
								provider.ImageClientOutputs = []image.Client{client}
							})

							AfterEach(func() {
								client.AssertOutputsEmpty()
							})

							revisionAssertions := func() {
								Context("deletes", func() {
									It("responds with an unauthorized error when the client returns an unauthorized error", func() {
										client.DeleteOutputs = []imageTest.DeleteOutput{{Deleted: false, Error: request.ErrorUnauthorized()}}
										handlerFunc(res, req)
										Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
										Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
										Expect(res.WriteInputs).To(HaveLen(1))
										errorsTest.ExpectErrorJSON(request.ErrorUnauthorized(), res.WriteInputs[0])
									})

									It("responds with an internal server error when the client returns an unknown error", func() {
										client.DeleteOutputs = []imageTest.DeleteOutput{{Deleted: false, Error: errorsTest.RandomError()}}
										handlerFunc(res, req)
										Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusInternalServerError}))
										Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
										Expect(res.WriteInputs).To(HaveLen(1))
										errorsTest.ExpectErrorJSON(request.ErrorInternalServerError(nil), res.WriteInputs[0])
									})

									It("responds with not found error when the client does not return an image", func() {
										client.DeleteOutputs = []imageTest.DeleteOutput{{Deleted: false, Error: nil}}
										handlerFunc(res, req)
										Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusNotFound}))
										Expect(res.HeaderOutput).To(Equal(&http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}))
										Expect(res.WriteInputs).To(HaveLen(1))
										errorsTest.ExpectErrorJSON(request.ErrorResourceNotFoundWithIDAndOptionalRevision(id, revision), res.WriteInputs[0])
									})

									It("responds successfully", func() {
										res.WriteOutputs = nil
										client.DeleteOutputs = []imageTest.DeleteOutput{{Deleted: true, Error: nil}}
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
									Expect(client.DeleteInputs).To(Equal([]imageTest.DeleteInput{{ID: id, Condition: &request.Condition{}}}))
								})

								revisionAssertions()
							})

							When("revision is present", func() {
								BeforeEach(func() {
									revision = pointer.FromInt(requestTest.RandomRevision())
								})

								AfterEach(func() {
									Expect(client.DeleteInputs).To(Equal([]imageTest.DeleteInput{{ID: id, Condition: &request.Condition{Revision: revision}}}))
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
