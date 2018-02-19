package test

import (
	"github.com/tidepool-org/platform/data/types/basal"

	// Pull in init function to ensure test environment
	_ "github.com/tidepool-org/platform/test"
)

func NewSuppressed(typ interface{}, deliveryType interface{}, annotations *[]map[string]interface{}, rate interface{}, scheduleName interface{}, suppressed *basal.Suppressed) *basal.Suppressed {
	datum := &basal.Suppressed{}
	if val, ok := typ.(string); ok {
		datum.Type = &val
	}
	if val, ok := deliveryType.(string); ok {
		datum.DeliveryType = &val
	}
	datum.Annotations = annotations
	if val, ok := rate.(float64); ok {
		datum.Rate = &val
	}
	if val, ok := scheduleName.(string); ok {
		datum.ScheduleName = &val
	}
	datum.Suppressed = suppressed
	return datum
}
