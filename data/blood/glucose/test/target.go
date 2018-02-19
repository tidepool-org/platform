package test

import (
	"github.com/tidepool-org/platform/data/blood/glucose"

	// Pull in init function to ensure test environment
	_ "github.com/tidepool-org/platform/test"
)

func NewTarget(high interface{}, low interface{}, rng interface{}, target interface{}) *glucose.Target {
	datum := &glucose.Target{}
	if value, ok := high.(float64); ok {
		datum.High = &value
	}
	if value, ok := low.(float64); ok {
		datum.Low = &value
	}
	if value, ok := rng.(float64); ok {
		datum.Range = &value
	}
	if value, ok := target.(float64); ok {
		datum.Target = &value
	}
	return datum
}
