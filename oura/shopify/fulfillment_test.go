package shopify_test

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"go.uber.org/mock/gomock"

	dataSource "github.com/tidepool-org/platform/data/source"
	"github.com/tidepool-org/platform/oura"

	dataSourceTest "github.com/tidepool-org/platform/data/source/test"

	"github.com/tidepool-org/platform/customerio"

	"github.com/tidepool-org/platform/auth"
	authTest "github.com/tidepool-org/platform/auth/test"

	"github.com/tidepool-org/platform/pointer"

	"github.com/tidepool-org/platform/oura/shopify"

	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	shopfiyTest "github.com/tidepool-org/platform/oura/shopify/test"
	ouraTest "github.com/tidepool-org/platform/oura/test"
)

var _ = Describe("FulfillmentEventProcessor", func() {
	var (
		ctx       context.Context
		processor *shopify.FulfillmentEventProcessor
		logger    log.Logger

		ctrl *gomock.Controller

		authClient       *authTest.MockClient
		dataSourceClient *dataSourceTest.MockClient
		shopifyClnt      *shopfiyTest.MockClient

		appAPIServer    *httptest.Server
		appAPIResponses *ouraTest.StubResponses

		trackAPIServer    *httptest.Server
		trackAPIResponses *ouraTest.StubResponses
	)

	BeforeEach(func() {
		ctx = context.Background()
		logger = logTest.NewLogger()
		ctrl, ctx = gomock.WithContext(ctx, GinkgoT())

		appAPIResponses = ouraTest.NewStubResponses()
		appAPIServer = ouraTest.NewStubServer(appAPIResponses)

		trackAPIResponses = ouraTest.NewStubResponses()
		trackAPIServer = ouraTest.NewStubServer(trackAPIResponses)

		customerIOConfig := customerio.Config{
			AppAPIBaseURL:   appAPIServer.URL,
			TrackAPIBaseURL: trackAPIServer.URL,
		}
		customerIOClient, err := customerio.NewClient(customerIOConfig, logger)
		Expect(err).ToNot(HaveOccurred())

		shopifyClnt = shopfiyTest.NewMockClient(ctrl)
		authClient = authTest.NewMockClient(ctrl)
		dataSourceClient = dataSourceTest.NewMockClient(ctrl)

		processor, err = shopify.NewFulfillmentEventProcessor(logger, customerIOClient, shopifyClnt, authClient, dataSourceClient)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(trackAPIResponses.UnmatchedResponses()).To(Equal(0))
		Expect(appAPIResponses.UnmatchedResponses()).To(Equal(0))

		appAPIServer.Close()
		trackAPIServer.Close()
		ctrl.Finish()
	})

	Context("Process", func() {
		It("should successfully process a sizing kit delivery", func() {
			id := "1aacb960-430c-4081-8b3b-a32688807dc5"
			sizingKitDiscountCode := shopify.RandomDiscountCode()

			event := shopify.FulfillmentEvent{
				ID:             9876543,
				CreatedAt:      time.Now(),
				ShipmentStatus: pointer.FromAny("delivered"),
				OrderID:        rand.Int63n(999999999999),
			}

			deduplicationID, err := customerio.CreateUlid(&event.CreatedAt, strconv.FormatInt(event.OrderID, 10))
			Expect(err).ToNot(HaveOccurred())

			shopifyClnt.EXPECT().
				GetDeliveredProducts(gomock.Any(), fmt.Sprintf("gid://shopify/Order/%d", event.OrderID)).
				Return(&shopify.DeliveredProducts{
					IDs:          []string{shopify.OuraSizingKitProductID},
					DiscountCode: sizingKitDiscountCode,
				}, nil)

			customers, err := ouraTest.LoadFixture("./test/fixtures/customers.json")
			Expect(err).ToNot(HaveOccurred())
			appAPIResponses.AddResponse(
				[]ouraTest.RequestMatcher{
					ouraTest.NewRequestMethodAndPathMatcher(http.MethodPost, "/v1/customers"),
					ouraTest.NewRequestJSONBodyMatcher(`{
					    "filter": {
				  		    "and": [
				 		        {
									"attribute": {
										"field": "oura_sizing_kit_discount_code",
							        	"operator": "eq",
							        	"value": "` + sizingKitDiscountCode + `"
                                    }
						        }
						    ]
					    }
					}`),
				},
				ouraTest.Response{StatusCode: http.StatusOK, Body: customers},
			)

			shopifyClnt.EXPECT().
				CreateDiscountCode(gomock.Any(), gomock.Any()).
				Do(func(ctx context.Context, input shopify.DiscountCodeInput) error {
					Expect(input.Title).To(Equal(shopify.OuraRingDiscountCodeTitle))
					Expect(len(input.Code)).To(BeNumerically(">=", 12))
					Expect(input.ProductID).To(Equal(shopify.OuraRingProductID))

					trackAPIResponses.AddResponse(
						[]ouraTest.RequestMatcher{
							ouraTest.NewRequestMethodAndPathMatcher(http.MethodPost, "/api/v1/customers/"+id+"/events"),
							ouraTest.NewRequestJSONBodyMatcher(`{
					  	        "name": "oura_sizing_kit_delivered",
						        "id": "` + deduplicationID.String() + `",
						        "data": {
                                    "oura_ring_discount_code": "` + input.Code + `",
                                    "oura_sizing_kit_discount_code": "` + sizingKitDiscountCode + `"
                                }
					        }`),
						},
						ouraTest.Response{StatusCode: http.StatusOK, Body: "{}"},
					)

					return nil
				}).
				Return(nil)

			err = processor.Process(ctx, event)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should successfully process a ring delivery", func() {
			id := "1aacb960-430c-4081-8b3b-a32688807dc5"
			discountCode := shopify.RandomDiscountCode()

			event := shopify.FulfillmentEvent{
				ID:             9876543,
				CreatedAt:      time.Now(),
				ShipmentStatus: pointer.FromAny("delivered"),
				OrderID:        rand.Int63n(999999999999),
			}

			deduplicationID, err := customerio.CreateUlid(&event.CreatedAt, strconv.FormatInt(event.OrderID, 10))
			Expect(err).ToNot(HaveOccurred())

			dataSourceClient.EXPECT().
				List(gomock.Any(), id, gomock.Any(), gomock.Any()).
				Return(dataSource.SourceArray{}, nil)
			dataSourceClient.EXPECT().Create(gomock.Any(), id, gomock.Any()).DoAndReturn(func(ctx context.Context, userID string, create *dataSource.Create) (*dataSource.Source, error) {
				Expect(create.ProviderName).To(PointTo(Equal("oura")))
				Expect(create.ProviderType).To(PointTo(Equal("oauth")))

				source := dataSourceTest.RandomSource()
				source.UserID = pointer.FromAny(userID)
				source.ProviderName = pointer.FromAny(oura.ProviderName)
				source.ProviderType = pointer.FromAny(auth.ProviderTypeOAuth)
				source.State = pointer.FromAny(dataSource.StateDisconnected)
				return source, nil
			})

			shopifyClnt.EXPECT().
				GetDeliveredProducts(gomock.Any(), fmt.Sprintf("gid://shopify/Order/%d", event.OrderID)).
				Return(&shopify.DeliveredProducts{
					IDs:          []string{shopify.OuraRingProductID},
					DiscountCode: discountCode,
				}, nil)

			customers, err := ouraTest.LoadFixture("./test/fixtures/customers.json")
			Expect(err).ToNot(HaveOccurred())
			appAPIResponses.AddResponse(
				[]ouraTest.RequestMatcher{
					ouraTest.NewRequestMethodAndPathMatcher(http.MethodPost, "/v1/customers"),
					ouraTest.NewRequestJSONBodyMatcher(`{
					    "filter": {
				  		    "and": [
				 		        {
									"attribute": {
										"field": "oura_ring_discount_code",
										"operator": "eq",
										"value": "` + discountCode + `"
									}
						        }
						    ]
					    }
					}`),
				},
				ouraTest.Response{StatusCode: http.StatusOK, Body: customers},
			)

			tokenID := authTest.RandomRestrictedTokenID()
			tokenExpirationTime := time.Now().Add(time.Hour * 24 * 30)
			token := auth.RestrictedToken{
				ID:             tokenID,
				UserID:         id,
				Paths:          pointer.FromAny([]string{"/v1/oauth/oura"}),
				ExpirationTime: tokenExpirationTime,
				CreatedTime:    time.Now(),
			}
			authClient.EXPECT().
				CreateUserRestrictedToken(gomock.Any(), id, gomock.Any()).
				Return(&token, nil)

			trackAPIResponses.AddResponse(
				[]ouraTest.RequestMatcher{
					ouraTest.NewRequestMethodAndPathMatcher(http.MethodPost, "/api/v1/customers/"+id+"/events"),
					ouraTest.NewRequestJSONBodyMatcher(`{
					  	"name": "oura_ring_delivered",
						"id": "` + deduplicationID.String() + `",
                        "data": {
                          "oura_ring_discount_code": "` + discountCode + `",
                          "oura_account_linking_token": "` + tokenID + `",
                          "oura_account_linking_token_expiration_time": ` + fmt.Sprintf("%d", tokenExpirationTime.Unix()) + `
                        }
					}`),
				},
				ouraTest.Response{StatusCode: http.StatusOK, Body: "{}"},
			)

			err = processor.Process(ctx, event)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
