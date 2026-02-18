package client

import (
	"context"
	"strings"

	"github.com/tidepool-org/platform/oura/shopify"
	"github.com/tidepool-org/platform/oura/shopify/generated"
	"github.com/tidepool-org/platform/pointer"
)

//go:generate go run github.com/Khan/genqlient
var _ = `# @genqlient
query GetOrder($identifier: OrderIdentifierInput!) {
  orderByIdentifier(identifier: $identifier) {
	createdAt
    discountCode
    lineItems {
      nodes { 
        product {
          id
        }
      }
    }
    fulfillments(first: 10) {
      deliveredAt
      displayStatus
      fulfillmentLineItems(first: 10) {
        nodes { 
          lineItem {
            product {
              id
            }
          }
        }
      }
	}
    id
  }
}
`

func (c *defaultClient) GetOrderSummary(ctx context.Context, orderID string) (*shopify.OrderSummary, error) {
	resp, err := generated.GetOrder(ctx, c.gql, &generated.OrderIdentifierInput{
		Id: pointer.FromAny(orderID),
	})
	if err != nil || resp.GetOrderByIdentifier() == nil {
		return nil, err
	}

	order := resp.GetOrderByIdentifier()
	summary := shopify.OrderSummary{
		CreatedTime:  order.CreatedAt,
		DiscountCode: pointer.Default(order.GetDiscountCode(), ""),
	}

	for _, lineItem := range order.GetLineItems().GetNodes() {
		if lineItem == nil {
			continue
		}
		id := lineItem.GetProduct().GetId()
		if strings.HasPrefix(id, shopify.ProductGIDPrefix) {
			id = strings.TrimPrefix(id, shopify.ProductGIDPrefix)
		}
		summary.OrderedProductIDs = append(summary.OrderedProductIDs, id)
	}

	for _, fulfillment := range order.Fulfillments {
		if fulfillment == nil {
			continue
		}

		lineItems := fulfillment.GetFulfillmentLineItems()
		if lineItems == nil {
			continue
		}
		for _, lineItem := range lineItems.GetNodes() {
			if lineItem == nil || lineItem.GetLineItem() == nil {
				continue
			}
			id := lineItem.GetLineItem().GetProduct().GetId()
			if strings.HasPrefix(id, shopify.ProductGIDPrefix) {
				id = strings.TrimPrefix(id, shopify.ProductGIDPrefix)
			}
			summary.DeliveredProductIDs = append(summary.DeliveredProductIDs, id)
		}
	}

	return &summary, nil
}
