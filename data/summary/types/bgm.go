package types

import (
	"github.com/tidepool-org/platform/data/types/blood/glucose/selfmonitored"
)

type BGMStats struct {
	GlucoseStats
}

func (*BGMStats) GetType() string {
	return SummaryTypeBGM
}

func (*BGMStats) GetDeviceDataTypes() []string {
	return []string{selfmonitored.Type}
}
