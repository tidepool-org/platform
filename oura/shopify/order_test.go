package shopify_test

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/oura/shopify"

	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/oura/customerio"
	ouraTest "github.com/tidepool-org/platform/oura/test"
)

var _ = Describe("FulfillmentEventProcessor", func() {
	var (
		ctx       context.Context
		processor *shopify.OrdersCreateEventProcessor
		logger    log.Logger

		appAPIServer    *httptest.Server
		appAPIResponses *ouraTest.StubResponses

		trackAPIServer    *httptest.Server
		trackAPIResponses *ouraTest.StubResponses
	)

	BeforeEach(func() {
		ctx = context.Background()
		logger = logTest.NewLogger()

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

		processor, err = shopify.NewOrdersCreateEventProcessor(logger, customerIOClient)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(trackAPIResponses.UnmatchedResponses()).To(Equal(0))
		Expect(appAPIResponses.UnmatchedResponses()).To(Equal(0))

		appAPIServer.Close()
		trackAPIServer.Close()
	})

	Context("Process", func() {
		It("should successfully process a sizing kit order placements", func() {
			id := "1aacb960-430c-4081-8b3b-a32688807dc5"
			sizingKitDiscountCode := shopify.RandomDiscountCode()
			productId, err := strconv.Atoi(shopify.OuraSizingKitProductID)
			Expect(err).ToNot(HaveOccurred())

			event := shopify.OrdersCreateEvent{
				ID:                9999999999,
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
						        "id": "` + sizingKitDiscountCode + `",
						        "data": {
                                    "oura_sizing_kit_discount_code": "` + sizingKitDiscountCode + `"
                                }
					        }`),
				},
				ouraTest.Response{StatusCode: http.StatusOK, Body: "{}"},
			)

			err = processor.Process(ctx, event)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should successfully process a ring order placements", func() {
			id := "1aacb960-430c-4081-8b3b-a32688807dc5"
			ringDiscountCode := shopify.RandomDiscountCode()
			productId, err := strconv.Atoi(shopify.OuraRingProductID)
			Expect(err).ToNot(HaveOccurred())

			event := shopify.OrdersCreateEvent{
				ID:                9999999999,
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
						        "id": "` + ringDiscountCode + `",
						        "data": {
                                    "oura_ring_discount_code": "` + ringDiscountCode + `"
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
