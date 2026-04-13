package v1_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ant0ine/go-json-rest/rest"
	"go.uber.org/mock/gomock"

	"github.com/tidepool-org/platform/auth"
	authTest "github.com/tidepool-org/platform/auth/test"
	dataServiceTest "github.com/tidepool-org/platform/data/service/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/oauth"
	"github.com/tidepool-org/platform/oura"
	ouraDataWorkEvent "github.com/tidepool-org/platform/oura/data/work/event"
	ouraServiceApiV1 "github.com/tidepool-org/platform/oura/service/api/v1"
	ouraTest "github.com/tidepool-org/platform/oura/test"
	ouraWebhook "github.com/tidepool-org/platform/oura/webhook"
	ouraWebhookTest "github.com/tidepool-org/platform/oura/webhook/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
	testHttp "github.com/tidepool-org/platform/test/http"
	"github.com/tidepool-org/platform/work"
	workTest "github.com/tidepool-org/platform/work/test"
)

var _ = Describe("event", func() {
	It("RequestHeaderSignature is expected", func() {
		Expect(ouraServiceApiV1.RequestHeaderSignature).To(Equal("X-Oura-Signature"))
	})

	It("RequestHeaderTimestamp is expected", func() {
		Expect(ouraServiceApiV1.RequestHeaderTimestamp).To(Equal("X-Oura-Timestamp"))
	})

	It("RequestBodySizeMaximum is expected", func() {
		Expect(ouraServiceApiV1.RequestBodySizeMaximum).To(Equal(1048576))
	})

	Context("with request, response, and router", func() {
		var (
			logger              *logTest.Logger
			ctx                 context.Context
			req                 *rest.Request
			res                 *testHttp.ResponseWriter
			mockController      *gomock.Controller
			mockAuthClient      *authTest.MockClient
			mockOuraClient      *ouraTest.MockClient
			mockWorkClient      *workTest.MockClient
			dependencies        ouraServiceApiV1.Dependencies
			handler             rest.HandlerFunc
			secret              string
			timestamp           string
			event               *ouraWebhook.Event
			body                []byte
			signature           string
			providerSessions    auth.ProviderSessions
			expectedFilter      *auth.ProviderSessionFilter
			expectedWorkCreates []*work.Create
			header              http.Header
		)

		BeforeEach(func() {
			var err error
			logger = logTest.NewLogger()
			ctx = log.NewContextWithLogger(context.Background(), logger)
			req = &rest.Request{}
			res = testHttp.NewResponseWriter()
			mockController, ctx = gomock.WithContext(ctx, GinkgoT())
			mockAuthClient = authTest.NewMockClient(mockController)
			mockOuraClient = ouraTest.NewMockClient(mockController)
			mockWorkClient = workTest.NewMockClient(mockController)
			dependencies = ouraServiceApiV1.Dependencies{
				AuthClient: mockAuthClient,
				OuraClient: mockOuraClient,
				WorkClient: mockWorkClient,
			}
			secret = test.RandomString()
			timestamp = test.RandomTime().Format(time.RFC3339Nano)
			event = ouraWebhookTest.RandomEvent()
			body, err = json.Marshal(event)
			Expect(err).NotTo(HaveOccurred())
			Expect(body).ToNot(BeNil())
			signature, err = ouraServiceApiV1.CalculateSignature(secret, timestamp, body)
			Expect(err).ToNot(HaveOccurred())
			Expect(signature).ToNot(BeEmpty())
			providerSessions = authTest.RandomProviderSessions()
			expectedFilter = &auth.ProviderSessionFilter{
				Type:       pointer.From(oauth.ProviderType),
				Name:       pointer.From(oura.ProviderName),
				ExternalID: pointer.From(*event.UserID),
			}
			expectedWorkCreates = []*work.Create{}
			for _, providerSession := range providerSessions {
				workCreate, err := ouraDataWorkEvent.NewWorkCreate(providerSession.ID, event)
				Expect(err).ToNot(HaveOccurred())
				Expect(workCreate).ToNot(BeNil())
				expectedWorkCreates = append(expectedWorkCreates, workCreate)
			}
			header = http.Header{}
			header.Set(ouraServiceApiV1.RequestHeaderSignature, signature)
			header.Set(ouraServiceApiV1.RequestHeaderTimestamp, timestamp)
			req.Request = httptest.NewRequestWithContext(ctx, "POST", "/", bytes.NewReader(body))
			req.Header = header
		})

		withHandler := func() {
			It("returns http.StatusBadRequest when the signature is missing", func() {
				req.Header.Del(ouraServiceApiV1.RequestHeaderSignature)
				handler(res, req)
				Expect(res.ResponseRecorder).To(HaveHTTPStatus(http.StatusBadRequest))
				Expect(res.ResponseRecorder).To(HaveHTTPBody("signature is missing"))
				logger.AssertError("signature is missing", log.Fields{"signature": ""})
			})

			It("returns http.StatusBadRequest when the timestamp is missing", func() {
				req.Header.Del(ouraServiceApiV1.RequestHeaderTimestamp)
				handler(res, req)
				Expect(res.ResponseRecorder).To(HaveHTTPStatus(http.StatusBadRequest))
				Expect(res.ResponseRecorder).To(HaveHTTPBody("timestamp is missing"))
				logger.AssertError("timestamp is missing", log.Fields{"signature": signature, "timestamp": ""})
			})

			It("returns http.StatusInternalServerError when the body cannot be read", func() {
				testErr := errorsTest.RandomError()
				req.Request = httptest.NewRequestWithContext(ctx, "POST", "/", test.ErrorReader(testErr))
				req.Header = header
				handler(res, req)
				Expect(res.ResponseRecorder).To(HaveHTTPStatus(http.StatusInternalServerError))
				Expect(res.ResponseRecorder).To(HaveHTTPBody(MatchJSON(internalServerErrorJSON)))
				logger.AssertError("unable to read request body", log.Fields{"signature": signature, "timestamp": timestamp})
			})

			It("returns http.StatusBadRequest when the body size exceed the maximum", func() {
				body = test.RandomBytesFromRange(ouraServiceApiV1.RequestBodySizeMaximum+1, ouraServiceApiV1.RequestBodySizeMaximum+1)
				req.Request = httptest.NewRequestWithContext(ctx, "POST", "/", bytes.NewReader(body))
				req.Header = header
				handler(res, req)
				Expect(res.ResponseRecorder).To(HaveHTTPStatus(http.StatusBadRequest))
				Expect(res.ResponseRecorder).To(HaveHTTPBody("request body size exceeds maximum allowed size"))
				logger.AssertError("request body size exceeds maximum allowed size", log.Fields{"signature": signature, "timestamp": timestamp, "bodySize": len(body)})
			})

			Context("with client secret", func() {
				BeforeEach(func() {
					mockOuraClient.EXPECT().ClientSecret().Return(secret)
				})

				It("returns http.StatusForbidden when signature does not match", func() {
					req.Header.Set(ouraServiceApiV1.RequestHeaderSignature, "invalid")
					handler(res, req)
					Expect(res.ResponseRecorder).To(HaveHTTPStatus(http.StatusForbidden))
					Expect(res.ResponseRecorder).To(HaveHTTPBody("signature is invalid"))
					logger.AssertError("signature is invalid", log.Fields{"signature": "invalid", "timestamp": timestamp, "bodySize": len(body), "calculatedSignature": signature})
				})

				It("returns http.StatusBadRequest when the body cannot be parsed", func() {
					var err error
					body = []byte(`["invalid"]`)
					signature, err = ouraServiceApiV1.CalculateSignature(secret, timestamp, body)
					Expect(err).ToNot(HaveOccurred())
					Expect(signature).ToNot(BeEmpty())
					header.Set(ouraServiceApiV1.RequestHeaderSignature, signature)
					req.Request = httptest.NewRequestWithContext(ctx, "POST", "/", bytes.NewReader(body))
					req.Header = header
					handler(res, req)
					Expect(res.ResponseRecorder).To(HaveHTTPStatus(http.StatusBadRequest))
					Expect(res.ResponseRecorder).To(HaveHTTPBody("unable to parse request body"))
					logger.AssertError("unable to parse request body", log.Fields{"signature": signature, "timestamp": timestamp, "bodySize": len(body)})
				})

				It("returns http.StatusInternalServerError when the provider sessions cannot be retrieved", func() {
					testErr := errorsTest.RandomError()
					mockAuthClient.EXPECT().
						ListProviderSessions(gomock.Not(gomock.Nil()), expectedFilter, gomock.Not(gomock.Nil())).
						Return(nil, testErr)
					handler(res, req)
					Expect(res.ResponseRecorder).To(HaveHTTPStatus(http.StatusInternalServerError))
					Expect(res.ResponseRecorder).To(HaveHTTPBody(MatchJSON(internalServerErrorJSON)))
					logger.AssertError("unable to get provider sessions", log.Fields{"signature": signature, "timestamp": timestamp, "bodySize": len(body), "event": event})
				})

				It("returns http.StatusInternalServerError when there are no provider sessions", func() {
					mockAuthClient.EXPECT().
						ListProviderSessions(gomock.Not(gomock.Nil()), expectedFilter, gomock.Not(gomock.Nil())).
						Return(auth.ProviderSessions{}, nil)
					handler(res, req)
					Expect(res.ResponseRecorder).To(HaveHTTPStatus(http.StatusInternalServerError))
					Expect(res.ResponseRecorder).To(HaveHTTPBody(MatchJSON(internalServerErrorJSON)))
					logger.AssertError("provider session is missing", log.Fields{"signature": signature, "timestamp": timestamp, "bodySize": len(body), "event": event})
				})

				Context("with provider sessions", func() {
					BeforeEach(func() {
						mockAuthClient.EXPECT().
							ListProviderSessions(gomock.Not(gomock.Nil()), expectedFilter, gomock.Not(gomock.Nil())).
							Return(providerSessions, nil)
					})

					It("returns http.StatusInternalServerError when work create cannot be created", func() {
						providerSessions[0].ID = ""
						handler(res, req)
						Expect(res.ResponseRecorder).To(HaveHTTPStatus(http.StatusInternalServerError))
						Expect(res.ResponseRecorder).To(HaveHTTPBody(MatchJSON(internalServerErrorJSON)))
						logger.AssertError("unable to create work create", log.Fields{"signature": signature, "timestamp": timestamp, "bodySize": len(body), "event": event, "providerSessionId": providerSessions[0].ID})
					})

					It("returns http.StatusInternalServerError when work cannot be created", func() {
						testErr := errorsTest.RandomError()
						expectedWorkCreate := expectedWorkCreates[0]
						mockWorkClient.EXPECT().
							Create(gomock.Not(gomock.Nil()), expectedWorkCreate).
							Return(nil, testErr)
						handler(res, req)
						Expect(res.ResponseRecorder).To(HaveHTTPStatus(http.StatusInternalServerError))
						Expect(res.ResponseRecorder).To(HaveHTTPBody(MatchJSON(internalServerErrorJSON)))
						logger.AssertError("unable to create work", log.Fields{"signature": signature, "timestamp": timestamp, "bodySize": len(body), "event": event, "providerSessionId": providerSessions[0].ID})
					})

					It("returns http.StatusOK", func() {
						for _, expectedWorkCreate := range expectedWorkCreates {
							mockWorkClient.EXPECT().
								Create(gomock.Not(gomock.Nil()), expectedWorkCreate).
								Return(workTest.NewWorkFromCreateWithState(expectedWorkCreate, work.StatePending), nil)
						}
						handler(res, req)
						Expect(res.ResponseRecorder).To(HaveHTTPStatus(http.StatusOK))
						Expect(res.ResponseRecorder).To(HaveHTTPBody("OK"))
					})
				})
			})
		}

		Context("with modern router", func() {
			BeforeEach(func() {
				router, err := ouraServiceApiV1.NewRouter(dependencies)
				Expect(err).ToNot(HaveOccurred())
				Expect(router).ToNot(BeNil())
				handler = func(res rest.ResponseWriter, req *rest.Request) {
					router.Event(res, req)
				}
			})

			withHandler()
		})

		Context("with legacy router", func() {
			var mockDataServiceContext *dataServiceTest.MockContext

			BeforeEach(func() {
				mockDataServiceContext = dataServiceTest.NewMockContext(mockController)
				handler = func(res rest.ResponseWriter, req *rest.Request) {
					mockDataServiceContext.EXPECT().Request().Return(req)
					mockDataServiceContext.EXPECT().Response().Return(res)
					ouraServiceApiV1.Event(mockDataServiceContext)
				}
			})

			It("returns http.StatusInternalServerError when the dependencies are invalid", func() {
				mockDataServiceContext.EXPECT().AuthClient().Return(nil)
				mockDataServiceContext.EXPECT().OuraClient().Return(nil)
				mockDataServiceContext.EXPECT().WorkClient().Return(nil)
				handler(res, req)
				Expect(res.ResponseRecorder).To(HaveHTTPStatus(http.StatusInternalServerError))
				Expect(res.ResponseRecorder).To(HaveHTTPBody(MatchJSON(internalServerErrorJSON)))
			})

			Context("with valid dependencies", func() {
				BeforeEach(func() {
					mockDataServiceContext.EXPECT().AuthClient().Return(mockAuthClient)
					mockDataServiceContext.EXPECT().OuraClient().Return(mockOuraClient)
					mockDataServiceContext.EXPECT().WorkClient().Return(mockWorkClient)
				})

				withHandler()
			})
		})
	})

	Context("CalculateSignature", func() {
		DescribeTable("return the expected results when the input",
			func(secret string, timestamp string, bytes []byte, expectedSignature string) {
				Expect(ouraServiceApiV1.CalculateSignature(secret, timestamp, bytes)).To(Equal(expectedSignature))
			},
			Entry("all missing", "", "", []byte{}, "B613679A0814D9EC772F95D778C35FC5FF1697C493715653C6C712144292C5AD"),
			Entry("only secret", "test-secret", "", []byte{}, "A41BC6D81D6413576AE0994995E0AD89A416EC97389515C3604F47722122EEEB"),
			Entry("only timestamp", "", "test-timestamp", []byte{}, "D2B4047E610AF25419403D37AF372CF98AC79BFF6A95348FFC46EA9A04C85399"),
			Entry("bytes nil", "", "", nil, "B613679A0814D9EC772F95D778C35FC5FF1697C493715653C6C712144292C5AD"),
			Entry("bytes empty", "", "", []byte(""), "B613679A0814D9EC772F95D778C35FC5FF1697C493715653C6C712144292C5AD"),
			Entry("only bytes", "", "", []byte("test-bytes"), "7FAFF8948F0C5DE1DB8F16C5CCCB93EECC880C325008DA6D0E16A805B061523D"),
			Entry("all exist", "test-secret", "test-timestamp", []byte("test-bytes"), "1FA9466B814527E16CC069DB6CA11CCC0592DC388137C458CEA4D3E329A89647"),
		)
	})
})

var internalServerErrorJSON = `{"code": "internal-server-error", "title": "internal server error", "detail": "internal server error"}`
