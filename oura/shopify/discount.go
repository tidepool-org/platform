package shopify

import "crypto/rand"

const (
	//OuraSizingKitProductID         = "9122899853526"
	//OuraRingProductID         = "9112952373462"

	OuraSizingKitProductID         = "15536573219203" // Todd's test gstore
	OuraSizingKitDiscountCodeTitle = "Oura Sizing Kit Discount Code"

	OuraRingProductID         = "15496765964675" // Todd's test store
	OuraRingDiscountCodeTitle = "Oura Ring Discount Code"

	DiscountCodeLength = 12
)

func RandomDiscountCode() string {
	code := rand.Text()
	return code[:DiscountCodeLength]
}
