package client

import (
	"context"
	"strings"

	"github.com/tidepool-org/platform/oura/shopify"
	"github.com/tidepool-org/platform/oura/shopify/generated"
	"github.com/tidepool-org/platform/pointer"
)

const (
	productGIDPrefix = "gid://shopify/Product/"
)

//go:generate go run github.com/Khan/genqlient
var _ = `# @genqlient
query GetOrder($identifier: OrderIdentifierInput!) {
  orderByIdentifier(identifier: $identifier) {
	createdAt
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

func (c *defaultClient) GetOrder(ctx context.Context, orderID string) (*generated.GetOrderOrderByIdentifierOrder, error) {
	resp, err := generated.GetOrder(ctx, c.gql, &generated.OrderIdentifierInput{
		Id: pointer.FromAny(orderID),
	})
	if err != nil {
		return nil, err
	}
	return resp.GetOrderByIdentifier(), nil
}

func (c *defaultClient) GetProductsFromOrder(order *generated.GetOrderOrderByIdentifierOrder) *shopify.Products {
	if order == nil {
		return nil
	}

	ids := make([]string, 0)
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
			if strings.HasPrefix(id, productGIDPrefix) {
				id = strings.TrimPrefix(id, productGIDPrefix)
			}
			ids = append(ids, id)
		}
	}

	var discountCode string
	if order.GetDiscountCode() != nil {
		discountCode = *order.GetDiscountCode()
	}
	return &shopify.Products{
		IDs:          ids,
		DiscountCode: discountCode,
	}
}
