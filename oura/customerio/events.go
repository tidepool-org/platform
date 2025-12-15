package customerio

const (
	OuraEligibilitySurveyCompletedEventType = "oura_eligibility_survey_completed"
	OuraSizingKitDeliveredEventType         = "oura_sizing_kit_delivered"
	OuraRingDeliveredEventType              = "oura_ring_delivered"
)

type OuraEligibilitySurveyCompletedData struct {
	OuraEligibilitySurveyID       string `json:"oura_eligibility_survey_id"`
	OuraEligibilitySurveyEligible bool   `json:"oura_eligibility_survey_eligible"`
	OuraSizingKitDiscountCode     string `json:"oura_sizing_kit_discount_code,omitempty"`
}

type OuraSizingKitDeliveredData struct {
	OuraRingDiscountCode string `json:"oura_ring_discount_code"`
}

type OuraRingDeliveredData struct{}
