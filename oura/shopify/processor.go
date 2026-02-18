package shopify

import (
	"context"
	"strings"
	"time"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/customerio"
	dataSource "github.com/tidepool-org/platform/data/source"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/oura"
	"github.com/tidepool-org/platform/oura/shopify/store"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
)

const (
	ouraAccountLinkingTokenPath = "/v1/oauth/oura"
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

type OrderProcessor struct {
	logger log.Logger

	customerIOClient      *customerio.Client
	dataSourceClient      dataSource.Client
	restrictedTokenClient auth.RestrictedTokenAccessor
	shopifyClient         Client
	store                 store.Store
}

func NewOrderProcessor(logger log.Logger, customerIOClient *customerio.Client, shopifyClient Client, restrictedTokenClient auth.RestrictedTokenAccessor, dataSourceClient dataSource.Client, store store.Store) (*OrderProcessor, error) {
	return &OrderProcessor{
		logger:                logger,
		customerIOClient:      customerIOClient,
		dataSourceClient:      dataSourceClient,
		restrictedTokenClient: restrictedTokenClient,
		shopifyClient:         shopifyClient,
		store:                 store,
	}, nil
}

func (p *OrderProcessor) ProcessFulfillment(ctx context.Context, event FulfillmentEvent) error {
	orderGID := GetOrderGID(event.OrderID)
	logger := p.logger.WithField("orderGID", orderGID)

	if event.ShipmentStatus == nil || !strings.EqualFold(*event.ShipmentStatus, "delivered") {
		logger.Warn("ignoring non-delivery fulfillment event")
		return nil
	}

	return p.processDeliveredOrder(ctx, orderGID)
}

func (p *OrderProcessor) processDeliveredOrder(ctx context.Context, orderGID string) error {
	logger := p.logger.WithField("orderGID", orderGID)

	if event, err := p.store.GetShopifyOrderEvent(ctx, orderGID, store.OrderEventTypeDelivered); err != nil {
		return errors.Wrap(err, "unable to retrieve shopify order event")
	} else if event != nil {
		logger.Info("ignoring order create event because it was already processed")
		return nil
	}

	order, err := p.shopifyClient.GetOrderSummary(ctx, orderGID)
	if err != nil {
		return err
	} else if order == nil {
		logger.Warn("order not found")
		return nil
	}

	if count := len(order.DeliveredProductIDs); count == 0 {
		logger.Info("ignoring fulfillment event with no delivered products")
		return nil
	} else if count > 1 {
		logger.Warn("ignoring fulfillment event with multiple delivered products")
		return nil
	}

	deliveredProductID := order.DeliveredProductIDs[0]
	logger = logger.WithField("productId", deliveredProductID)

	attribute, ok := productIDToOuraDiscountAttribute[deliveredProductID]
	if !ok {
		logger.Warn("unable to find discount attribute for delivered product")
		return nil
	}

	customers, err := p.customerIOClient.FindCustomers(ctx, map[string]any{
		"filter": map[string]any{
			"and": []any{
				map[string]any{
					"attribute": map[string]any{
						"field":    attribute,
						"operator": "eq",
						"value":    order.DiscountCode,
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
		if err := p.onSizingKitDelivered(ctx, customers.Identifiers[0], *order); err != nil {
			logger.WithError(err).Warn("unable to send sizing kit delivered event")
			return err
		}
	case OuraRingProductID:
		if err := p.onRingDelivered(ctx, customers.Identifiers[0], *order); err != nil {
			logger.WithError(err).Warn("unable to send ring delivered event")
			return err
		}
	default:
		logger.Warn("ignoring fulfillment event for unknown product")
		return nil
	}

	err = p.store.CreateShopifyOrderEvent(ctx, store.ShopifyOrderEvent{
		OrderGID:   orderGID,
		UserID:     customers.Identifiers[0].ID,
		Type:       store.OrderEventTypeDelivered,
		CreateTime: time.Now(),
	})
	if err != nil {
		return errors.Wrap(err, "unable to create shopify order event")
	}

	return nil
}

func (p *OrderProcessor) ProcessOrderCreate(ctx context.Context, event OrdersCreateEvent) error {
	return p.processNewOrder(ctx, GetOrderGID(event.ID))
}

func (p *OrderProcessor) processNewOrder(ctx context.Context, orderGID string) error {
	logger := p.logger.WithField("orderGID", orderGID)

	if event, err := p.store.GetShopifyOrderEvent(ctx, orderGID, store.OrderEventTypeCreated); err != nil {
		return errors.Wrap(err, "unable to retrieve shopify order event")
	} else if event != nil {
		logger.Info("ignoring order create event because it was already processed")
		return nil
	}

	order, err := p.shopifyClient.GetOrderSummary(ctx, orderGID)
	if err != nil {
		return errors.Wrap(err, "unable to retrieve order")
	} else if order == nil {
		logger.Warn("order not found")
		return nil
	}

	if count := len(order.OrderedProductIDs); count == 0 {
		logger.Info("ignoring fulfillment event with no delivered products")
		return nil
	} else if count > 1 {
		logger.Warn("ignoring fulfillment event with multiple delivered products")
		return nil
	}

	productID := order.OrderedProductIDs[0]
	logger = logger.WithField("productId", productID)

	attribute, ok := productIDToOuraDiscountAttribute[productID]
	if !ok {
		logger.Warn("unable to find discount attribute for product")
		return nil
	}

	customers, err := p.customerIOClient.FindCustomers(ctx, map[string]any{
		"filter": map[string]any{
			"and": []any{
				map[string]any{
					"attribute": map[string]any{
						"field":    attribute,
						"operator": "eq",
						"value":    order.DiscountCode,
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
		if err := p.onSizingKitOrdered(ctx, customers.Identifiers[0], *order); err != nil {
			logger.WithError(err).Warn("unable to send sizing kit ordered event")
			return err
		}
	case OuraRingProductID:
		if err := p.onRingOrdered(ctx, customers.Identifiers[0], *order); err != nil {
			logger.WithError(err).Warn("unable to send ring ordered event")
			return err
		}
	default:
		logger.Warn("ignoring orders create event for unknown product")
		return nil
	}

	err = p.store.CreateShopifyOrderEvent(ctx, store.ShopifyOrderEvent{
		OrderGID:   orderGID,
		UserID:     customers.Identifiers[0].ID,
		Type:       store.OrderEventTypeCreated,
		CreateTime: time.Now(),
	})
	if err != nil {
		return errors.Wrap(err, "unable to create shopify order event")
	}

	return nil
}

func (p *OrderProcessor) onSizingKitDelivered(ctx context.Context, identifiers customerio.Identifiers, order OrderSummary) error {
	discountCode := RandomDiscountCode()
	err := p.shopifyClient.CreateDiscountCode(ctx, DiscountCodeInput{
		Title:     OuraRingDiscountCodeTitle,
		Code:      discountCode,
		ProductID: OuraRingProductID,
	})
	if err != nil {
		return errors.Wrap(err, "unable to create oura discount code")
	}

	sizingKitDelivered := &customerio.Event{
		Name: oura.OuraSizingKitDeliveredEventType,
		Data: oura.OuraSizingKitDeliveredData{
			OuraRingDiscountCode:      discountCode,
			OuraSizingKitDiscountCode: order.DiscountCode,
		},
	}

	if err = sizingKitDelivered.SetDeduplicationID(&order.CreatedTime, order.GID); err != nil {
		return err
	}

	return p.customerIOClient.SendEvent(ctx, identifiers.ID, sizingKitDelivered)
}

func (p *OrderProcessor) onRingDelivered(ctx context.Context, identifiers customerio.Identifiers, order OrderSummary) error {
	// A user must have a data source to be able to link their account
	sources, err := p.dataSourceClient.List(ctx, identifiers.ID, &dataSource.Filter{
		ProviderName: pointer.FromAny([]string{oura.ProviderName}),
		ProviderType: pointer.FromAny([]string{auth.ProviderTypeOAuth}),
	}, page.NewPaginationMinimum())
	if err != nil {
		return errors.Wrap(err, "unable to list data sources")
	}
	if len(sources) == 0 {
		p.logger.WithField("userId", identifiers.ID).Info("creating oura data source")
		create := dataSource.NewCreate()
		create.ProviderName = pointer.FromAny(oura.ProviderName)
		create.ProviderType = pointer.FromAny(auth.ProviderTypeOAuth)

		_, err := p.dataSourceClient.Create(ctx, identifiers.ID, create)
		if err != nil {
			return errors.Wrap(err, "unable to create oura data source")
		}
	}

	create := auth.NewRestrictedTokenCreate()
	create.Paths = pointer.FromAny([]string{ouraAccountLinkingTokenPath})
	create.ExpirationTime = pointer.FromTime(time.Now().Add(time.Hour * 24 * 30))

	token, err := p.restrictedTokenClient.CreateUserRestrictedToken(ctx, identifiers.ID, create)
	if err != nil {
		return errors.Wrap(err, "unable to create restricted token")
	}

	ringDelivered := &customerio.Event{
		Name: oura.OuraRingDeliveredEventType,
		Data: oura.OuraRingDeliveredData{
			OuraRingDiscountCode:                  order.DiscountCode,
			OuraAccountLinkingToken:               token.ID,
			OuraAccountLinkingTokenExpirationTime: token.ExpirationTime.Unix(),
		},
	}

	if err = ringDelivered.SetDeduplicationID(&order.CreatedTime, order.GID); err != nil {
		return err
	}

	return p.customerIOClient.SendEvent(ctx, identifiers.ID, ringDelivered)
}

func (p *OrderProcessor) onSizingKitOrdered(ctx context.Context, identifiers customerio.Identifiers, order OrderSummary) error {
	sizingKitOrdered := &customerio.Event{
		Name: oura.OuraSizingKitOrderedEventType,
		Data: oura.OuraSizingKitOrderedData{
			OuraSizingKitDiscountCode: order.DiscountCode,
		},
	}
	if err := sizingKitOrdered.SetDeduplicationID(&order.CreatedTime, order.GID); err != nil {
		return errors.Wrap(err, "unable to set event id")
	}

	return p.customerIOClient.SendEvent(ctx, identifiers.ID, sizingKitOrdered)
}

func (p *OrderProcessor) onRingOrdered(ctx context.Context, identifiers customerio.Identifiers, order OrderSummary) error {
	ringOrdered := &customerio.Event{
		Name: oura.OuraRingOrderedEventType,
		Data: oura.OuraRingOrderedData{
			OuraRingDiscountCode: order.DiscountCode,
		},
	}

	if err := ringOrdered.SetDeduplicationID(&order.CreatedTime, order.GID); err != nil {
		return errors.Wrap(err, "unable to set event id")
	}

	return p.customerIOClient.SendEvent(ctx, identifiers.ID, ringOrdered)
}
