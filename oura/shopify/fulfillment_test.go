package shopify_test

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/tidepool-org/platform/pointer"

	"github.com/tidepool-org/platform/oura/shopify"

	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/oura/customerio"
	jotformTest "github.com/tidepool-org/platform/oura/jotform/test"
	shopfiyTest "github.com/tidepool-org/platform/oura/shopify/test"
	ouraTest "github.com/tidepool-org/platform/oura/test"
)

var _ = Describe("FulfillmentEventProcessor", func() {
	var (
		ctx       context.Context
		processor *shopify.FulfillmentEventProcessor
		logger    log.Logger

		shopifyCtrl *gomock.Controller
		shopifyClnt *shopfiyTest.MockClient

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

		shopifyCtrl = gomock.NewController(GinkgoT())
		shopifyClnt = shopfiyTest.NewMockClient(shopifyCtrl)

		processor, err = shopify.NewFulfillmentEventProcessor(logger, customerIOClient, shopifyClnt)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(trackAPIResponses.UnmatchedResponses()).To(Equal(0))
		Expect(appAPIResponses.UnmatchedResponses()).To(Equal(0))

		appAPIServer.Close()
		trackAPIServer.Close()
		shopifyCtrl.Finish()
	})

	Context("ProcessSubmission", func() {
		It("should successfully process a sizing kit delivery", func() {
			id := "1aacb960-430c-4081-8b3b-a32688807dc5"
			sizingKitDiscountCode := shopify.RandomDiscountCode()

			event := shopify.FulfillmentEvent{
				ID:             9876543,
				ShipmentStatus: pointer.FromAny("delivered"),
				OrderID:        rand.Int63n(999999999999),
			}

			shopifyClnt.EXPECT().
				GetDeliveredProducts(gomock.Any(), fmt.Sprintf("gid://shopify/Order/%d", event.OrderID)).
				Return(&shopify.DeliveredProducts{
					OrderID:      fmt.Sprintf("%d", event.OrderID),
					IDs:          []string{shopify.OuraSizingKitProductID},
					DiscountCode: sizingKitDiscountCode,
				}, nil)

			customers, err := jotformTest.LoadFixture("./test/fixtures/customers.json")
			Expect(err).ToNot(HaveOccurred())
			appAPIResponses.AddResponse(
				[]ouraTest.RequestMatcher{
					ouraTest.NewRequestMethodAndPathMatcher(http.MethodPost, "/v1/customers"),
					ouraTest.NewRequestJSONBodyMatcher(`{
					    "filter": {
				  		    "and": [
				 		        {
							        "field": "oura_sizing_kit_discount_code",
							        "operator": "eq",
							        "value": "` + sizingKitDiscountCode + `"
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
						        "id": "` + fmt.Sprintf("%d", event.ID) + `",
						        "data": {
                                    "oura_ring_discount_code": "` + input.Code + `"
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
				ShipmentStatus: pointer.FromAny("delivered"),
				OrderID:        rand.Int63n(999999999999),
			}

			shopifyClnt.EXPECT().
				GetDeliveredProducts(gomock.Any(), fmt.Sprintf("gid://shopify/Order/%d", event.OrderID)).
				Return(&shopify.DeliveredProducts{
					OrderID:      fmt.Sprintf("%d", event.OrderID),
					IDs:          []string{shopify.OuraRingProductID},
					DiscountCode: discountCode,
				}, nil)

			customers, err := jotformTest.LoadFixture("./test/fixtures/customers.json")
			Expect(err).ToNot(HaveOccurred())
			appAPIResponses.AddResponse(
				[]ouraTest.RequestMatcher{
					ouraTest.NewRequestMethodAndPathMatcher(http.MethodPost, "/v1/customers"),
					ouraTest.NewRequestJSONBodyMatcher(`{
					    "filter": {
				  		    "and": [
				 		        {
							        "field": "oura_ring_discount_code",
							        "operator": "eq",
							        "value": "` + discountCode + `"
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
					  	"name": "oura_ring_delivered",
						"id": "` + fmt.Sprintf("%d", event.ID) + `",
						"data": {}
					}`),
				},
				ouraTest.Response{StatusCode: http.StatusOK, Body: "{}"},
			)

			err = processor.Process(ctx, event)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
