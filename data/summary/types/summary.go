package types

import "time"

const (
	SummaryTypeCGM = "cgm"
	SummaryTypeBGM = "bgm"
)

type Summary[T Stats] struct {
	Type   string
	UserID string

	Dates Dates
	Stats T
}

type Dates struct {
	OutdatedSince time.Time
}

type Stats interface {
	CGMStats | BGMStats
	GetType() string
}

func Create[T Stats]() Summary[T] {
	stats := new(T)
	return Summary[T]{
		Type:  (*stats).GetType(),
		Stats: *stats,
	}
}

func GetTypeString[T Stats]() string {
	t := new(T)
	return (*t).GetType()
}
