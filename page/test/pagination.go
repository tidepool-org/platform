package test

import (
	"math"

	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/test"
)

func RandomPagination() *page.Pagination {
	pagination := page.NewPagination()
	pagination.Page = test.RandomIntFromRange(0, math.MaxInt32)
	pagination.Size = test.RandomIntFromRange(1, 100)
	return pagination
}
