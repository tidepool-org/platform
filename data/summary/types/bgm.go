package types

import (
	"github.com/tidepool-org/platform/data/types/blood/glucose/selfmonitored"
)

type BGMObservations struct {
	GlucoseStats
}

func (*BGMObservations) GetType() string {
	return SummaryTypeBGM
}

func (*BGMObservations) GetDeviceDataTypes() []string {
	return []string{selfmonitored.Type}
}
