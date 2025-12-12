package customerio

const (
	OuraEligibilitySurveyCompletedEventType = "oura_eligibility_survey_completed"
	OuraSizingKitDeliveredEventType         = "oura_sizing_kit_delivered"
)

type OuraEligibilitySurveyCompletedData struct {
	Eligible                  bool   `json:"eligible"`
	OuraSizingKitDiscountCode string `json:"oura_sizing_kit_discount_code,omitempty"`
}

type OuraSizingKitDeliveredData struct {
	OuraRingDiscountCode string `json:"oura_ring_discount_code"`
}
