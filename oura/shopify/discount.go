package shopify

import "crypto/rand"

const (
	OuraSizingKitProductID         = "9122899853526"
	OuraSizingKitDiscountCodeTitle = "Oura Sizing Kit Discount Code"

	OuraRingProductID         = "9112952373462"
	OuraRingDiscountCodeTitle = "Oura Ring Discount Code"

	DiscountCodeLength = 12
)

func RandomDiscountCode() string {
	code := rand.Text()
	return code[:DiscountCodeLength]
}
