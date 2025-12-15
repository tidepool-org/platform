package client

import (
	"context"

	"github.com/tidepool-org/platform/oura/shopify"
	"github.com/tidepool-org/platform/oura/shopify/generated"
	"github.com/tidepool-org/platform/pointer"
)

//go:generate go run github.com/Khan/genqlient
var _ = `# @genqlient
query GetOrder($identifier: OrderIdentifierInput!) {
  orderByIdentifier(identifier: $identifier) {
    discountCode
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

func (c *defaultClient) GetDeliveredProducts(ctx context.Context, orderID string) (*shopify.DeliveredProducts, error) {
	resp, err := generated.GetOrder(ctx, c.gql, &generated.OrderIdentifierInput{
		Id: pointer.FromAny(orderID),
	})
	if err != nil {
		return nil, err
	}
	if resp.GetOrderByIdentifier() == nil {
		return nil, nil
	}
	ids := make([]string, 0)
	for _, fulfillment := range resp.GetOrderByIdentifier().Fulfillments {
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
			ids = append(ids, lineItem.GetLineItem().GetProduct().GetId())
		}
	}

	return &shopify.DeliveredProducts{IDs: ids}, nil
}
