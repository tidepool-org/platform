package client

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/tidepool-org/platform/oura/shopify/generated"
)

//go:generate go run github.com/Khan/genqlient
var _ = `# @genqlient
mutation CreateDiscountCode($basicCodeDiscount: DiscountCodeBasicInput!) {
  discountCodeBasicCreate(basicCodeDiscount: $basicCodeDiscount) {
    codeDiscountNode {
      id
      codeDiscount {
        ... on DiscountCodeBasic {
          title
          codes(first: 1) {
            nodes {
              code
            }
          }
        }
      }
    }
    userErrors {
      field
      message
    }
  }
}
`

type DiscountCodeInput struct {
	Title     string
	Code      string
	ProductID string
}

func (c *Client) CreateDiscountCode(ctx context.Context, discountCodeInput DiscountCodeInput) error {
	input := ptr(generated.DiscountCodeBasicInput{
		Title:                  ptr(discountCodeInput.Title),
		AppliesOncePerCustomer: ptr(true),
		Code:                   ptr(discountCodeInput.Code),
		UsageLimit:             ptr(1),
		Context: ptr(generated.DiscountContextInput{
			All: ptr(generated.DiscountBuyerSelectionAll),
		}),
		CustomerGets: ptr(generated.DiscountCustomerGetsInput{
			Value: ptr(generated.DiscountCustomerGetsValueInput{
				DiscountOnQuantity: ptr(generated.DiscountOnQuantityInput{
					Quantity: ptr("1"),
					Effect: ptr(generated.DiscountEffectInput{
						Percentage: ptr(float64(1)),
					}),
				}),
			}),
			Items: ptr(generated.DiscountItemsInput{
				Products: ptr(generated.DiscountProductsInput{
					ProductsToAdd: []string{discountCodeInput.ProductID},
				}),
			}),
		}),
		MinimumRequirement: ptr(generated.DiscountMinimumRequirementInput{
			Quantity: ptr(generated.DiscountMinimumQuantityInput{
				GreaterThanOrEqualToQuantity: ptr("1"),
			}),
		}),
		StartsAt: ptr(time.Now()),
	})

	resp, err := generated.CreateDiscountCode(ctx, c.gql, input)
	if err != nil {
		return err
	}

	userErrors := resp.DiscountCodeBasicCreate.UserErrors
	if len(userErrors) > 0 {
		for _, e := range userErrors {
			slog.Error("user error", "field", e.Field, "message", e.Message)
		}
		return errors.New("user errors")
	}

	return nil
}
