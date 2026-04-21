package client

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/tidepool-org/platform/oura/shopify"
	"github.com/tidepool-org/platform/oura/shopify/generated"
	"github.com/tidepool-org/platform/pointer"
)

//go:generate go run github.com/Khan/genqlient
var _ = `# @genqlient
query GetOrder($identifier: OrderIdentifierInput!) {
  orderByIdentifier(identifier: $identifier) {
	createdAt
	updatedAt
    discountCode
    lineItems(first: 10) {
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
		GID:          orderID,
		CreatedTime:  order.CreatedAt,
		DiscountCode: pointer.Default(order.GetDiscountCode(), ""),
		UpdatedTime:  order.UpdatedAt,
	}

	for _, lineItem := range order.GetLineItems().GetNodes() {
		if lineItem == nil {
			continue
		}
		id := strings.TrimPrefix(lineItem.GetProduct().GetId(), shopify.ProductGIDPrefix)
		summary.OrderedProductIDs = append(summary.OrderedProductIDs, id)
	}

	deliveredProducts := map[string]struct{}{}
	for _, fulfillment := range order.Fulfillments {
		if fulfillment == nil {
			continue
		}
		if fulfillment.DisplayStatus != nil && *fulfillment.DisplayStatus == generated.FulfillmentDisplayStatusDelivered {
			summary.IsDelivered = true
		}

		lineItems := fulfillment.GetFulfillmentLineItems()
		if lineItems == nil {
			continue
		}
		for _, lineItem := range lineItems.GetNodes() {
			if lineItem == nil || lineItem.GetLineItem() == nil {
				continue
			}
			id := strings.TrimPrefix(lineItem.GetLineItem().GetProduct().GetId(), shopify.ProductGIDPrefix)
			deliveredProducts[id] = struct{}{}
		}
	}
	summary.DeliveredProductIDs = slices.Collect(maps.Keys(deliveredProducts))
	return &summary, nil
}

//go:generate go run github.com/Khan/genqlient
var _ = `# @genqlient
query GetGIDsOfUpdatedOrders($query: String!, $count: Int!) {
  orders(query: $query, first: $count, sortKey: UPDATED_AT) {
    edges {
	  node {
		id
	  }
	}
  }
}
`

func (c *defaultClient) GetGIDsOfUpdatedOrders(ctx context.Context, updatedSince time.Time, count int) ([]string, error) {
	query := fmt.Sprintf("updated_at:>='%s'", updatedSince.Format(time.RFC3339))
	resp, err := generated.GetGIDsOfUpdatedOrders(ctx, c.gql, query, count)
	if err != nil || resp.GetOrders() == nil {
		return nil, err
	}

	gids := make([]string, 0, len(resp.GetOrders().GetEdges()))
	for _, edge := range resp.GetOrders().GetEdges() {
		if edge.GetNode() == nil {
			continue
		}

		gids = append(gids, edge.GetNode().GetId())
	}

	return gids, nil
}
