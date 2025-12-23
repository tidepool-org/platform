package shopify

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/oura/customerio"
)

var (
	productIDToOuraDiscountAttribute = map[string]string{
		OuraSizingKitProductID: "oura_sizing_kit_discount_code",
		OuraRingProductID:      "oura_ring_discount_code",
	}
)

type FulfillmentEvent struct {
	ID                int64     `json:"id"`
	OrderID           int64     `json:"order_id"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	Service           *string   `json:"service"`
	UpdatedAt         time.Time `json:"updated_at"`
	TrackingCompany   string    `json:"tracking_company"`
	ShipmentStatus    *string   `json:"shipment_status"`
	LocationID        *int64    `json:"location_id"`
	Email             string    `json:"email"`
	TrackingNumber    string    `json:"tracking_number"`
	TrackingNumbers   []string  `json:"tracking_numbers"`
	TrackingURL       string    `json:"tracking_url"`
	TrackingURLs      []string  `json:"tracking_urls"`
	Name              string    `json:"name"`
	AdminGraphQLAPIID string    `json:"admin_graphql_api_id"`
}

type FulfillmentEventProcessor struct {
	logger log.Logger

	customerIOClient *customerio.Client
	shopifyClient    Client
}

func NewFulfillmentEventProcessor(logger log.Logger, customerIOClient *customerio.Client, shopifyClient Client) (*FulfillmentEventProcessor, error) {
	return &FulfillmentEventProcessor{
		logger:           logger,
		customerIOClient: customerIOClient,
		shopifyClient:    shopifyClient,
	}, nil
}

func (f *FulfillmentEventProcessor) Process(ctx context.Context, event FulfillmentEvent) error {
	logger := f.logger.WithField("fulfillmentId", event.ID)
	if event.ShipmentStatus == nil || !strings.EqualFold(*event.ShipmentStatus, "delivered") {
		logger.Warn("ignoring non-delivery fulfillment event")
		return nil
	}

	orderId := fmt.Sprintf("gid://shopify/Order/%d", event.OrderID)
	deliveredProducts, err := f.shopifyClient.GetDeliveredProducts(ctx, orderId)
	if err != nil {
		return err
	}
	if deliveredProducts == nil || len(deliveredProducts.IDs) == 0 {
		logger.Info("ignoring fulfillment event with no delivered products")
		return nil
	} else if len(deliveredProducts.IDs) > 1 {
		logger.Warn("ignoring fulfillment event with multiple delivered products")
		return nil
	}

	deliveredProductID := deliveredProducts.IDs[0]
	logger = logger.WithField("orderId", orderId).WithField("productId", deliveredProductID)

	attribute, ok := productIDToOuraDiscountAttribute[deliveredProductID]
	if !ok {
		logger.Warn("unable to find discount attribute for delivered product")
		return nil
	}

	customers, err := f.customerIOClient.FindCustomers(ctx, map[string]any{
		"filter": map[string]any{
			"and": []any{
				map[string]any{
					"attribute": map[string]any{
						"field":    attribute,
						"operator": "eq",
						"value":    deliveredProducts.DiscountCode,
					},
				},
			},
		},
	})
	if err != nil {
		logger.WithError(err).Warn("unable to find customers")
		return nil
	}

	if len(customers.Identifiers) == 0 {
		logger.Warn("no customers found for delivered products")
		return nil
	} else if len(customers.Identifiers) > 1 {
		userIds := make([]string, 0, len(customers.Identifiers))
		for _, id := range customers.Identifiers {
			userIds = append(userIds, id.ID)
		}
		logger.WithField("userIds", userIds).Warn("multiple customers found for delivered products")
		return nil
	}

	switch deliveredProductID {
	case OuraSizingKitProductID:
		if err := f.onSizingKitDelivered(ctx, customers.Identifiers[0], event, deliveredProducts.DiscountCode); err != nil {
			logger.WithError(err).Warn("unable to send sizing kit delivered event")
			return err
		}
	case OuraRingProductID:
		if err := f.onRingDelivered(ctx, customers.Identifiers[0], event, deliveredProducts.DiscountCode); err != nil {
			logger.WithError(err).Warn("unable to send ring delivered event")
			return err
		}
	default:
		logger.Warn("ignoring fulfillment event for unknown product")
	}

	return nil
}

func (f *FulfillmentEventProcessor) onSizingKitDelivered(ctx context.Context, identifiers customerio.Identifiers, event FulfillmentEvent, sizingKitDiscountCode string) error {
	discountCode := RandomDiscountCode()
	err := f.shopifyClient.CreateDiscountCode(ctx, DiscountCodeInput{
		Title:     OuraRingDiscountCodeTitle,
		Code:      discountCode,
		ProductID: OuraRingProductID,
	})
	if err != nil {
		return errors.Wrap(err, "unable to create oura discount code")
	}

	sizingKitDelivered := customerio.Event{
		Name: customerio.OuraSizingKitDeliveredEventType,
		ID:   fmt.Sprintf("%d", event.ID),
		Data: customerio.OuraSizingKitDeliveredData{
			OuraRingDiscountCode:      discountCode,
			OuraSizingKitDiscountCode: sizingKitDiscountCode,
		},
	}

	return f.customerIOClient.SendEvent(ctx, identifiers.ID, sizingKitDelivered)
}

func (f *FulfillmentEventProcessor) onRingDelivered(ctx context.Context, identifiers customerio.Identifiers, event FulfillmentEvent, ringDiscountCode string) error {
	ringDelivered := customerio.Event{
		Name: customerio.OuraRingDeliveredEventType,
		ID:   fmt.Sprintf("%d", event.ID),
		Data: customerio.OuraRingDeliveredData{
			OuraRingDiscountCode: ringDiscountCode,
		},
	}

	return f.customerIOClient.SendEvent(ctx, identifiers.ID, ringDelivered)
}
