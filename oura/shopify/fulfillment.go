package shopify

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/oura/customerio"
	"github.com/tidepool-org/platform/oura/shopify/client"
)

const ()

type FulfillmentEventCreated struct {
	ID                int       `json:"id"`
	FulfillmentID     int       `json:"fulfillment_id"`
	Status            string    `json:"status"`
	Message           string    `json:"message"`
	HappenedAt        time.Time `json:"happened_at"`
	Country           string    `json:"country"`
	ShopID            int       `json:"shop_id"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	OrderID           int64     `json:"order_id"`
	AdminGraphqlApiID string    `json:"admin_graphql_api_id"`
}

type FulfillmentCreatedEventProcessor struct {
	logger log.Logger

	customerIOClient customerio.Client
	shopifyClient    client.Client
}

func (f *FulfillmentCreatedEventProcessor) Process(ctx context.Context, event FulfillmentEventCreated) error {
	logger := f.logger.WithField("fulfillmentId", event.FulfillmentID)
	if event.Status != "delivered" {
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

	logger = logger.WithField("orderId", orderId).WithField("productId", deliveredProducts.IDs[0])

	customers, err := f.customerIOClient.FindCustomers(ctx, map[string]any{
		"filter": map[string]any{
			"or": []any{
				map[string]any{
					"field":    "oura_sizing_kit_discount_code",
					"operator": "eq",
					"value":    deliveredProducts.DiscountCode,
				},
				map[string]any{
					"field":    "oura_ring_discount_code",
					"operator": "eq",
					"value":    deliveredProducts.DiscountCode,
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

	// There shouldn't be more than one product for a single delivery, but adding this loop to be more flexible with the
	// 3PL integration or future changes.
	for _, productId := range deliveredProducts.IDs {
		switch productId {
		case OuraSizingKitProductID:
			err := f.onSizingKitDelivered(ctx, customers.Identifiers[0], event)
			if err != nil {
				logger.WithError(err).Warn("unable to send sizing kit delivered event")
			}
		default:
			logger.WithField("productId", productId).Warn("ignoring fulfillment event for unknown product")
		}
	}

	return nil
}

func (f *FulfillmentCreatedEventProcessor) onSizingKitDelivered(ctx context.Context, identifiers customerio.Identifiers, event FulfillmentEventCreated) error {
	discountCode := RandomDiscountCode()
	err := f.shopifyClient.CreateDiscountCode(ctx, client.DiscountCodeInput{
		Title:     OuraRingDiscountCodeTitle,
		Code:      discountCode,
		ProductID: OuraRingProductID,
	})
	if err != nil {
		return errors.Wrap(err, "unable to create oura discount code")
	}

	sizingKitDelivered := customerio.Event{
		Name: customerio.OuraSizingKitDeliveredEventType,
		ID:   strconv.Itoa(event.ID),
		Data: customerio.OuraSizingKitDeliveredData{
			OuraRingDiscountCode: discountCode,
		},
	}

	err = f.customerIOClient.SendEvent(ctx, identifiers.CID, sizingKitDelivered)
	if err != nil {
		return errors.Wrap(err, "unable to send sizing kit delivered event")
	}

	return nil
}
