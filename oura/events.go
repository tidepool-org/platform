package oura

const (
	OuraEligibilitySurveyCompletedEventType = "oura_eligibility_survey_completed"
	OuraSizingKitOrderedEventType           = "oura_sizing_kit_ordered"
	OuraSizingKitDeliveredEventType         = "oura_sizing_kit_delivered"
	OuraRingOrderedEventType                = "oura_ring_ordered"
	OuraRingDeliveredEventType              = "oura_ring_delivered"
)

type OuraEligibilitySurveyCompletedData struct {
	OuraEligibilitySurveyID       string `json:"oura_eligibility_survey_id"`
	OuraEligibilitySurveyEligible bool   `json:"oura_eligibility_survey_eligible"`
	OuraSizingKitDiscountCode     string `json:"oura_sizing_kit_discount_code,omitempty"`
}

type OuraSizingKitOrderedData struct {
	OuraSizingKitDiscountCode string `json:"oura_sizing_kit_discount_code"`
}

type OuraSizingKitDeliveredData struct {
	OuraSizingKitDiscountCode string `json:"oura_sizing_kit_discount_code"`
	OuraRingDiscountCode      string `json:"oura_ring_discount_code"`
}

type OuraRingOrderedData struct {
	OuraRingDiscountCode string `json:"oura_ring_discount_code"`
}
type OuraRingDeliveredData struct {
	OuraRingDiscountCode                  string `json:"oura_ring_discount_code"`
	OuraAccountLinkingToken               string `json:"oura_account_linking_token"`
	OuraAccountLinkingTokenExpirationTime int64  `json:"oura_account_linking_token_expiration_time"`
}
