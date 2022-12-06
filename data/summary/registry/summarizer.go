package registry

import (
	"context"
	"github.com/tidepool-org/platform/data/summary/types"
	"time"
)

type Summarizer[T types.Stats] interface {
	GetSummary(ctx context.Context, userId string) (*types.Summary[T], error)
	UpdateSummary(ctx context.Context, userId string) (*types.Summary[T], error)
}

func SkipUntil[T types.RecordTypes](date time.Time, userData []*T) ([]*T, error) {
	var skip int
	for i := 0; i < len(userData); i++ {
		recordTime := (*userData[i]).GetTime()

		if recordTime.Before(date) {
			skip = i + 1
		} else {
			break
		}
	}

	if skip > 0 {
		userData = userData[skip:]
	}

	return userData, nil
}
