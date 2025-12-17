package client

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/tidepool-org/platform/oura/shopify"
	"github.com/tidepool-org/platform/oura/shopify/generated"
	"github.com/tidepool-org/platform/pointer"
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

func (c *defaultClient) CreateDiscountCode(ctx context.Context, discountCodeInput shopify.DiscountCodeInput) error {
	input := pointer.FromAny(generated.DiscountCodeBasicInput{
		Title:                  pointer.FromAny(discountCodeInput.Title),
		AppliesOncePerCustomer: pointer.FromAny(true),
		Code:                   pointer.FromAny(discountCodeInput.Code),
		UsageLimit:             pointer.FromAny(1),
		Context: pointer.FromAny(generated.DiscountContextInput{
			All: pointer.FromAny(generated.DiscountBuyerSelectionAll),
		}),
		CustomerGets: pointer.FromAny(generated.DiscountCustomerGetsInput{
			Value: pointer.FromAny(generated.DiscountCustomerGetsValueInput{
				DiscountOnQuantity: pointer.FromAny(generated.DiscountOnQuantityInput{
					Quantity: pointer.FromAny("1"),
					Effect: pointer.FromAny(generated.DiscountEffectInput{
						Percentage: pointer.FromAny(float64(1)),
					}),
				}),
			}),
			Items: pointer.FromAny(generated.DiscountItemsInput{
				Products: pointer.FromAny(generated.DiscountProductsInput{
					ProductsToAdd: []string{fmt.Sprintf("gid://shopify/Product/%s", discountCodeInput.ProductID)},
				}),
			}),
		}),
		MinimumRequirement: pointer.FromAny(generated.DiscountMinimumRequirementInput{
			Quantity: pointer.FromAny(generated.DiscountMinimumQuantityInput{
				GreaterThanOrEqualToQuantity: pointer.FromAny("1"),
			}),
		}),
		StartsAt: pointer.FromAny(time.Now()),
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
