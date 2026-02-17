package shopify

import (
	"context"
	"fmt"
	"time"

	"github.com/tidepool-org/platform/customerio"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/oura"
	"github.com/tidepool-org/platform/oura/shopify/store"
)

type OrdersCreateEvent struct {
	ID                int64          `json:"id"`
	CreatedAt         time.Time      `json:"created_at"`
	AdminGraphQLAPIID string         `json:"admin_graphql_api_id"`
	DiscountCodes     []DiscountCode `json:"discount_codes"`
	LineItems         []LineItem     `json:"line_items"`
}

type DiscountCode struct {
	Code   string `json:"code"`
	Type   string `json:"type"`
	Amount string `json:"amount"`
}

type LineItem struct {
	ID                int64  `json:"id"`
	AdminGraphQLAPIID string `json:"admin_graphql_api_id"`
	ProductID         int64  `json:"product_id"`
}

type OrdersCreateEventProcessor struct {
	logger log.Logger

	customerIOClient *customerio.Client
	store            store.Store
}

func NewOrdersCreateEventProcessor(logger log.Logger, customerIOClient *customerio.Client, store store.Store) (*OrdersCreateEventProcessor, error) {
	return &OrdersCreateEventProcessor{
		logger:           logger,
		customerIOClient: customerIOClient,
		store:            store,
	}, nil
}

func (o *OrdersCreateEventProcessor) Process(ctx context.Context, event OrdersCreateEvent) error {
	logger := o.logger.WithField("orderId", event.ID)

	if event, err := o.store.GetShopifyOrderEvent(ctx, event.AdminGraphQLAPIID, store.OrderEventTypeCreated); err != nil {
		return errors.Wrap(err, "unable to retrieve shopify order event")
	} else if event != nil {
		logger.Info("ignoring order create event because it was already processed")
		return nil
	}

	var products []string
	for _, lineItem := range event.LineItems {
		products = append(products, fmt.Sprintf("%d", lineItem.ProductID))
	}

	if len(products) == 0 {
		logger.Info("ignoring orders create event with no products")
		return nil
	} else if len(products) > 1 {
		logger.Warn("ignoring orders create event with multiple products")
		return nil
	}

	productID := products[0]
	logger = logger.WithField("productId", productID)

	attribute, ok := productIDToOuraDiscountAttribute[productID]
	if !ok {
		logger.Warn("unable to find discount attribute for product")
		return nil
	}

	var discountCodes []string
	for _, discountCode := range event.DiscountCodes {
		discountCodes = append(discountCodes, discountCode.Code)
	}

	if len(discountCodes) == 0 {
		logger.Warn("ignoring orders create event with no discount codes")
		return nil
	} else if len(discountCodes) > 1 {
		logger.Warn("ignoring orders create event with multiple discount codes")
		return nil
	}

	discountCode := discountCodes[0]
	customers, err := o.customerIOClient.FindCustomers(ctx, map[string]any{
		"filter": map[string]any{
			"and": []any{
				map[string]any{
					"attribute": map[string]any{
						"field":    attribute,
						"operator": "eq",
						"value":    discountCode,
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

	switch productID {
	case OuraSizingKitProductID:
		if err := o.onSizingKitOrdered(ctx, customers.Identifiers[0], event, discountCode); err != nil {
			logger.WithError(err).Warn("unable to send sizing kit ordered event")
			return err
		}
	case OuraRingProductID:
		if err := o.onRingOrdered(ctx, customers.Identifiers[0], event, discountCode); err != nil {
			logger.WithError(err).Warn("unable to send ring ordered event")
			return err
		}
	default:
		logger.Warn("ignoring orders create event for unknown product")
		return nil
	}

	err = o.store.CreateShopifyOrderEvent(ctx, store.ShopifyOrderEvent{
		OrderID:    event.AdminGraphQLAPIID,
		UserID:     customers.Identifiers[0].ID,
		Type:       store.OrderEventTypeCreated,
		CreateTime: time.Now(),
	})
	if err != nil {
		return errors.Wrap(err, "unable to create shopify order event")
	}

	return nil
}

func (o *OrdersCreateEventProcessor) onSizingKitOrdered(ctx context.Context, identifiers customerio.Identifiers, event OrdersCreateEvent, discountCode string) error {
	sizingKitOrdered := &customerio.Event{
		Name: oura.OuraSizingKitOrderedEventType,
		Data: oura.OuraSizingKitOrderedData{
			OuraSizingKitDiscountCode: discountCode,
		},
	}
	if err := sizingKitOrdered.SetDeduplicationID(&event.CreatedAt, discountCode); err != nil {
		return errors.Wrap(err, "unable to set event id")
	}

	return o.customerIOClient.SendEvent(ctx, identifiers.ID, sizingKitOrdered)
}

func (o *OrdersCreateEventProcessor) onRingOrdered(ctx context.Context, identifiers customerio.Identifiers, event OrdersCreateEvent, discountCode string) error {
	ringOrdered := &customerio.Event{
		Name: oura.OuraRingOrderedEventType,
		Data: oura.OuraRingOrderedData{
			OuraRingDiscountCode: discountCode,
		},
	}

	if err := ringOrdered.SetDeduplicationID(&event.CreatedAt, discountCode); err != nil {
		return errors.Wrap(err, "unable to set event id")
	}

	return o.customerIOClient.SendEvent(ctx, identifiers.ID, ringOrdered)
}
