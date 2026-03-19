package shopify

import "crypto/rand"

const (
	//OuraSizingKitProductID         = "9122899853526"
	//OuraRingProductID         = "9112952373462"

	OuraSizingKitProductID         = "9280563445974" // Dummy Sizing Kit
	OuraSizingKitDiscountCodeTitle = "Oura Sizing Kit Discount Code"

	OuraRingProductID         = "9280563708118" // Dummy Oura ring
	OuraRingDiscountCodeTitle = "Oura Ring Discount Code"

	DiscountCodeLength = 12
)

func RandomDiscountCode() string {
	code := rand.Text()
	return code[:DiscountCodeLength]
}
