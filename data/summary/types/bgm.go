package types

import (
	"github.com/tidepool-org/platform/data/types/blood/glucose/selfmonitored"
)

type BGMPeriods struct {
	GlucosePeriods
}

func (*BGMPeriods) GetType() string {
	return SummaryTypeBGM
}

func (*BGMPeriods) GetDeviceDataTypes() []string {
	return []string{selfmonitored.Type}
}
