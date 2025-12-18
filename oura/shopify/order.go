package shopify

import (
	"context"
	"fmt"
	"time"

	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/oura/customerio"
)

type OrdersCreateEvent struct {
	ID                 int64          `json:"id"`
	AdminGraphQLAPIID  string         `json:"admin_graphql_api_id"`
	ConfirmationNumber interface{}    `json:"confirmation_number"`
	Confirmed          bool           `json:"confirmed"`
	ContactEmail       string         `json:"contact_email"`
	CreatedAt          time.Time      `json:"created_at"`
	DiscountCodes      []DiscountCode `json:"discount_codes"`
	Email              string         `json:"email"`
	LineItems          []LineItem     `json:"line_items"`
	Name               string         `json:"name"`
	Note               interface{}    `json:"note"`
	NoteAttributes     []interface{}  `json:"note_attributes"`
	OrderNumber        int            `json:"order_number"`
	OrderStatusUrl     string         `json:"order_status_url"`
	ProcessedAt        time.Time      `json:"processed_at"`
	Reference          interface{}    `json:"reference"`
	UpdatedAt          time.Time      `json:"updated_at"`
	UserId             interface{}    `json:"user_id"`
	Returns            []interface{}  `json:"returns"`
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
	shopifyClient    Client
}

func NewOrdersCreateEventProcessor(logger log.Logger, customerIOClient *customerio.Client, shopifyClient Client) (*OrdersCreateEventProcessor, error) {
	return &OrdersCreateEventProcessor{
		logger:           logger,
		customerIOClient: customerIOClient,
		shopifyClient:    shopifyClient,
	}, nil
}

func (f *OrdersCreateEventProcessor) Process(ctx context.Context, event OrdersCreateEvent) error {
	logger := f.logger.WithField("orderId", event.ID)

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
	customers, err := f.customerIOClient.FindCustomers(ctx, map[string]any{
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
		if err := f.onSizingKitOrdered(ctx, customers.Identifiers[0], discountCode); err != nil {
			logger.WithError(err).Warn("unable to send sizing kit ordered event")
			return err
		}
	case OuraRingProductID:
		if err := f.onRingOrdered(ctx, customers.Identifiers[0], discountCode); err != nil {
			logger.WithError(err).Warn("unable to send ring ordered event")
			return err
		}
	default:
		logger.Warn("ignoring orders create event for unknown product")
	}

	return nil
}

func (f *OrdersCreateEventProcessor) onSizingKitOrdered(ctx context.Context, identifiers customerio.Identifiers, discountCode string) error {
	sizingKitOrdered := customerio.Event{
		Name: customerio.OuraSizingKitOrderedEventType,
		ID:   discountCode,
		Data: customerio.OuraSizingKitOrderedData{
			OuraSizingKitDiscountCode: discountCode,
		},
	}

	return f.customerIOClient.SendEvent(ctx, identifiers.ID, sizingKitOrdered)
}

func (f *OrdersCreateEventProcessor) onRingOrdered(ctx context.Context, identifiers customerio.Identifiers, discountCode string) error {
	ringOrdered := customerio.Event{
		Name: customerio.OuraRingOrderedEventType,
		ID:   discountCode,
		Data: customerio.OuraRingOrderedData{
			OuraRingDiscountCode: discountCode,
		},
	}

	return f.customerIOClient.SendEvent(ctx, identifiers.ID, ringOrdered)
}
