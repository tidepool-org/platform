package types

type BGMStats struct {
	SomethingElse int
}

func (BGMStats) GetType() string {
	return SummaryTypeBGM
}
