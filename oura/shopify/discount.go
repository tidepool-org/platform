package shopify

import "crypto/rand"

const (
	OuraRingProductID         = "9112952373462"
	OuraRingDiscountCodeTitle = "Oura Ring Discount Code"

	OuraSizingKitProductID         = "9122899853526"
	OuraSizingKitDiscountCodeTitle = "Oura Sizing Kit Discount Code"

	DiscountCodeLength = 12
)

func RandomDiscountCode() string {
	code := rand.Text()
	return code[:DiscountCodeLength]
}
