package test

import (
	"math"

	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/test"
)

func RandomPagination() *page.Pagination {
	pagination := page.NewPagination()
	pagination.Page = RandomPage()
	pagination.Size = RandomSize()
	return pagination
}

func ClonePagination(datum *page.Pagination) *page.Pagination {
	if datum == nil {
		return nil
	}
	clone := page.NewPagination()
	clone.Page = datum.Page
	clone.Size = datum.Size
	return clone
}

func RandomPage() int {
	datum := page.PaginationPageDefault
	for datum == page.PaginationPageDefault {
		datum = test.RandomIntFromRange(page.PaginationPageMinimum, math.MaxInt32)
	}
	return datum
}

func RandomSize() int {
	datum := page.PaginationSizeDefault
	for datum == page.PaginationSizeDefault {
		datum = test.RandomIntFromRange(page.PaginationSizeMinimum, page.PaginationSizeMaximum)
	}
	return datum
}
