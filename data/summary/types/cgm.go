package types

type CGMStats struct {
	PercentInLow float64
}

func (CGMStats) GetType() string {
	return SummaryTypeCGM
}
