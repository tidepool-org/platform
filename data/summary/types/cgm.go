package types

import (
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
)

type CGMPeriods struct {
	GlucosePeriods
}

func (*CGMPeriods) GetType() string {
	return SummaryTypeCGM
}

func (*CGMPeriods) GetDeviceDataTypes() []string {
	return []string{continuous.Type}
}
