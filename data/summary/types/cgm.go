package types

import (
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
)

type CGMStats struct {
	GlucosePeriods
}

func (*CGMStats) GetType() string {
	return SummaryTypeCGM
}

func (*CGMStats) GetDeviceDataTypes() []string {
	return []string{continuous.Type}
}
