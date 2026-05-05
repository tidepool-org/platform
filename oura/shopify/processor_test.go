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

	"go.uber.org/mock/gomock"

	"github.com/tidepool-org/platform/auth"
	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/customerio"
	dataSource "github.com/tidepool-org/platform/data/source"
	dataSourceTest "github.com/tidepool-org/platform/data/source/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/oura"
	"github.com/tidepool-org/platform/oura/shopify"
	shopifyTest "github.com/tidepool-org/platform/oura/shopify/test"
	ouraTest "github.com/tidepool-org/platform/oura/test"
	"github.com/tidepool-org/platform/pointer"
)

var _ = Describe("OrderProcessor", func() {
	var (
		ctx       context.Context
		processor *shopify.OrderProcessor
		logger    log.Logger

		mockController *gomock.Controller

		authClient       *authTest.MockClient
		dataSourceClient *dataSourceTest.MockClient
		shopifyClnt      *shopifyTest.MockClient

		appAPIServer    *httptest.Server
		appAPIResponses *ouraTest.StubResponses

		trackAPIServer    *httptest.Server
		trackAPIResponses *ouraTest.StubResponses
	)

	BeforeEach(func() {
		ctx = context.Background()
		logger = logTest.NewLogger()
		mockController, ctx = gomock.WithContext(ctx, GinkgoT())

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

		shopifyClnt = shopifyTest.NewMockClient(mockController)
		authClient = authTest.NewMockClient(mockController)
		dataSourceClient = dataSourceTest.NewMockClient(mockController)

		config := shopify.Config{
			Enabled: true,
		}

		processor, err = shopify.NewOrderProcessor(logger, config, customerIOClient, shopifyClnt, authClient, dataSourceClient, GetSuiteStore())
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(trackAPIResponses.UnmatchedResponses()).To(Equal(0))
		Expect(appAPIResponses.UnmatchedResponses()).To(Equal(0))

		appAPIServer.Close()
		trackAPIServer.Close()
	})

	Context("ProcessFulfillment", func() {
		It("should successfully process a sizing kit delivery", func() {
			id := "1aacb960-430c-4081-8b3b-a32688807dc5"
			sizingKitDiscountCode := shopify.RandomDiscountCode()

			event := shopify.FulfillmentEvent{
				ID:             9876543,
				CreatedAt:      time.Now(),
				Status:         "success",
				ShipmentStatus: pointer.FromAny("delivered"),
				OrderID:        rand.Int63n(999999999999),
			}
			orderID := fmt.Sprintf("gid://shopify/Order/%d", event.OrderID)
			orderSummary := &shopify.OrderSummary{
				GID:                 orderID,
				CreatedTime:         time.Now(),
				OrderedProductIDs:   []string{shopify.OuraSizingKitProductID},
				DeliveredProductIDs: []string{shopify.OuraSizingKitProductID},
				DiscountCode:        sizingKitDiscountCode,
			}

			deduplicationID, err := customerio.CreateUlid(&orderSummary.CreatedTime, "oura_sizing_kit_delivered"+":"+orderSummary.GID)
			Expect(err).ToNot(HaveOccurred())

			shopifyClnt.EXPECT().
				GetOrderSummary(gomock.Any(), orderID).
				Return(orderSummary, nil)

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
				DoAndReturn(func(ctx context.Context, input shopify.DiscountCodeInput) error {
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
				})

			err = processor.ProcessFulfillment(ctx, event)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should successfully process a ring delivery", func() {
			id := "1aacb960-430c-4081-8b3b-a32688807dc5"
			discountCode := shopify.RandomDiscountCode()

			event := shopify.FulfillmentEvent{
				ID:             9876543,
				CreatedAt:      time.Now(),
				Status:         "success",
				ShipmentStatus: pointer.FromAny("delivered"),
				OrderID:        rand.Int63n(999999999999),
			}
			orderID := fmt.Sprintf("gid://shopify/Order/%d", event.OrderID)
			orderSummary := &shopify.OrderSummary{
				GID:                 orderID,
				CreatedTime:         time.Now(),
				OrderedProductIDs:   []string{shopify.OuraRingProductID},
				DeliveredProductIDs: []string{shopify.OuraRingProductID},
				DiscountCode:        discountCode,
			}

			deduplicationID, err := customerio.CreateUlid(&orderSummary.CreatedTime, "oura_ring_delivered"+":"+orderID)
			Expect(err).ToNot(HaveOccurred())

			dataSourceClient.EXPECT().
				List(gomock.Any(), id, gomock.Any(), gomock.Any()).
				Return(dataSource.SourceArray{}, nil)
			dataSourceClient.EXPECT().Create(gomock.Any(), id, gomock.Any()).DoAndReturn(func(ctx context.Context, userID string, create *dataSource.Create) (*dataSource.Source, error) {
				Expect(create.ProviderName).To(Equal("oura"))
				Expect(create.ProviderType).To(Equal("oauth"))

				source := dataSourceTest.RandomSource()
				source.UserID = userID
				source.ProviderName = oura.ProviderName
				source.ProviderType = auth.ProviderTypeOAuth
				source.State = dataSource.StateDisconnected
				return source, nil
			})

			shopifyClnt.EXPECT().
				GetOrderSummary(gomock.Any(), orderID).
				Return(orderSummary, nil)

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

			err = processor.ProcessFulfillment(ctx, event)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("ProcessOrderCreate", func() {
		It("should successfully process a sizing kit order placements", func() {
			id := "1aacb960-430c-4081-8b3b-a32688807dc5"
			sizingKitDiscountCode := shopify.RandomDiscountCode()
			productId, err := strconv.Atoi(shopify.OuraSizingKitProductID)
			Expect(err).ToNot(HaveOccurred())

			event := shopify.OrdersCreateEvent{
				ID:                9999999999,
				CreatedAt:         time.Now(),
				AdminGraphQLAPIID: "gid://shopify/Order/9999999999",
				DiscountCodes: []shopify.DiscountCode{{
					Code:   sizingKitDiscountCode,
					Type:   "discount",
					Amount: "10.00",
				}},
				LineItems: []shopify.LineItem{{
					ID:                rand.Int63(),
					AdminGraphQLAPIID: fmt.Sprintf("gid://shopify/Product/%d", productId),
					ProductID:         int64(productId),
				}},
			}
			orderSummary := &shopify.OrderSummary{
				GID:               event.AdminGraphQLAPIID,
				CreatedTime:       time.Now(),
				OrderedProductIDs: []string{fmt.Sprintf("gid://shopify/Product/%d", productId)},
				DiscountCode:      sizingKitDiscountCode,
			}

			deduplicationID, err := customerio.CreateUlid(&orderSummary.CreatedTime, sizingKitDiscountCode)
			Expect(err).ToNot(HaveOccurred())

			shopifyClnt.EXPECT().
				GetOrderSummary(gomock.Any(), orderSummary.GID).
				Return(orderSummary, nil)

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

			trackAPIResponses.AddResponse(
				[]ouraTest.RequestMatcher{
					ouraTest.NewRequestMethodAndPathMatcher(http.MethodPost, "/api/v1/customers/"+id+"/events"),
					ouraTest.NewRequestJSONBodyMatcher(`{
					  	        "name": "oura_sizing_kit_ordered",
						        "id": "` + deduplicationID.String() + `",
						        "data": {
                                    "oura_sizing_kit_discount_code": "` + sizingKitDiscountCode + `"
                                }
					        }`),
				},
				ouraTest.Response{StatusCode: http.StatusOK, Body: "{}"},
			)

			err = processor.ProcessOrderCreate(ctx, event)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should successfully process a ring order placements", func() {
			id := "1aacb960-430c-4081-8b3b-a32688807dc5"
			ringDiscountCode := shopify.RandomDiscountCode()
			productId, err := strconv.Atoi(shopify.OuraRingProductID)
			Expect(err).ToNot(HaveOccurred())

			event := shopify.OrdersCreateEvent{
				ID:                9999999999,
				CreatedAt:         time.Now(),
				AdminGraphQLAPIID: "gid://shopify/Order/9999999999",
				DiscountCodes: []shopify.DiscountCode{{
					Code:   ringDiscountCode,
					Type:   "discount",
					Amount: "10.00",
				}},
				LineItems: []shopify.LineItem{{
					ID:                rand.Int63(),
					AdminGraphQLAPIID: fmt.Sprintf("gid://shopify/Product/%d", productId),
					ProductID:         int64(productId),
				}},
			}
			orderSummary := &shopify.OrderSummary{
				GID:               event.AdminGraphQLAPIID,
				CreatedTime:       time.Now(),
				OrderedProductIDs: []string{fmt.Sprintf("gid://shopify/Product/%d", productId)},
				DiscountCode:      ringDiscountCode,
			}

			deduplicationID, err := customerio.CreateUlid(&event.CreatedAt, ringDiscountCode)
			Expect(err).ToNot(HaveOccurred())

			shopifyClnt.EXPECT().
				GetOrderSummary(gomock.Any(), orderSummary.GID).
				Return(orderSummary, nil)

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
							        	"value": "` + ringDiscountCode + `"
                                    }
						        }
						    ]
					    }
					}`),
				},
				ouraTest.Response{StatusCode: http.StatusOK, Body: customers},
			)

			trackAPIResponses.AddResponse(
				[]ouraTest.RequestMatcher{
					ouraTest.NewRequestMethodAndPathMatcher(http.MethodPost, "/api/v1/customers/"+id+"/events"),
					ouraTest.NewRequestJSONBodyMatcher(`{
					  	        "name": "oura_ring_ordered",
						        "id": "` + deduplicationID.String() + `",
						        "data": {
                                    "oura_ring_discount_code": "` + ringDiscountCode + `"
                                }
					        }`),
				},
				ouraTest.Response{StatusCode: http.StatusOK, Body: "{}"},
			)

			err = processor.ProcessOrderCreate(ctx, event)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
