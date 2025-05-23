package test

import (
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/test"
)

func RandomPage() int {
	datum := page.PaginationPageDefault
	for datum == page.PaginationPageDefault {
		datum = test.RandomIntFromRange(page.PaginationPageMinimum, test.RandomIntMaximum())
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

func NewObjectFromPagination(datum *page.Pagination, objectFormat test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := map[string]any{}
	if datum.Page != page.PaginationPageDefault {
		object["page"] = test.NewObjectFromInt(datum.Page, objectFormat)
	}
	if datum.Size != page.PaginationSizeDefault {
		object["size"] = test.NewObjectFromInt(datum.Size, objectFormat)
	}
	return object
}
