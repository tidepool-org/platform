package page

import (
	"net/http"
	"strconv"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/structure"
)

const (
	PaginationPageDefault = 0
	PaginationPageMinimum = 0
	PaginationSizeDefault = 100
	PaginationSizeMaximum = 1000
	PaginationSizeMinimum = 1
)

type Pagination struct {
	Page int `json:"page,omitempty"`
	Size int `json:"size,omitempty"`
}

func NewPagination() *Pagination {
	return &Pagination{
		Page: PaginationPageDefault,
		Size: PaginationSizeDefault,
	}
}

func NewPaginationMinimum() *Pagination {
	return &Pagination{
		Page: PaginationPageMinimum,
		Size: PaginationSizeMinimum,
	}
}

func (p *Pagination) Parse(parser structure.ObjectParser) {
	if page := parser.Int("page"); page != nil {
		p.Page = *page
	}
	if size := parser.Int("size"); size != nil {
		p.Size = *size
	}
}

func (p *Pagination) Validate(validator structure.Validator) {
	validator.Int("page", &p.Page).GreaterThanOrEqualTo(PaginationPageMinimum)
	validator.Int("size", &p.Size).InRange(PaginationSizeMinimum, PaginationSizeMaximum)
}

func (p *Pagination) MutateRequest(req *http.Request) error {
	parameters := map[string]string{}
	if p.Page != PaginationPageDefault {
		parameters["page"] = strconv.Itoa(p.Page)
	}
	if p.Size != PaginationSizeDefault {
		parameters["size"] = strconv.Itoa(p.Size)
	}
	return request.NewParametersMutator(parameters).MutateRequest(req)
}

type Paginator func(pagination Pagination) (done bool, err error)

func Paginate(paginator Paginator) error {
	return PaginateWithSize(PaginationSizeDefault, paginator)
}

func PaginateWithSize(size int, paginator Paginator) error {
	if size < PaginationSizeMinimum {
		return errors.New("size is less than minimum")
	}
	if paginator == nil {
		return errors.New("paginator is missing")
	}

	for page := 0; ; page++ {
		if done, err := paginator(Pagination{Page: page, Size: size}); err != nil {
			return err
		} else if done {
			return nil
		}
	}
}

type Pager[T any, U ~[]T] func(pagination Pagination) (U, error)

func Collect[T any, U ~[]T](pager Pager[T, U]) (U, error) {
	return CollectWithSize(PaginationSizeDefault, pager)
}

func CollectWithSize[T any, U ~[]T](size int, pager Pager[T, U]) (U, error) {
	if size < PaginationSizeMinimum {
		return nil, errors.New("size is less than minimum")
	}
	if pager == nil {
		return nil, errors.New("pager is missing")
	}

	var result U
	for page := 0; ; page++ {
		paged, err := pager(Pagination{Page: page, Size: size})
		if err != nil {
			return nil, err
		} else if paged == nil {
			return result, nil
		}
		if result == nil {
			result = U{}
		}
		result = append(result, paged...)
		if len(paged) < size {
			return result, nil
		}
	}
}

func First[T any, U ~[]T](pager Pager[T, U]) (value T, err error) {
	if pager == nil {
		return value, errors.New("pager is missing")
	}

	paged, err := pager(Pagination{Page: PaginationPageMinimum, Size: PaginationSizeMinimum})
	if err != nil || len(paged) <= 0 {
		return value, err
	}
	return paged[0], nil
}

type Processor[T any, U any] func(element T) (U, error)

func Process[T any, U any, V ~[]T, W []U](pager Pager[T, V], processor Processor[T, U]) (W, error) {
	return ProcessWithSize[T, U, V, W](PaginationSizeDefault, pager, processor)
}

func ProcessWithSize[T any, U any, V ~[]T, W []U](size int, pager Pager[T, V], processor Processor[T, U]) (W, error) {
	if size < PaginationSizeMinimum {
		return nil, errors.New("size is less than minimum")
	}
	if pager == nil {
		return nil, errors.New("pager is missing")
	}
	if processor == nil {
		return nil, errors.New("processor is missing")
	}

	var result W
	for page := 0; ; page++ {
		paged, err := pager(Pagination{Page: page, Size: size})
		if err != nil {
			return result, err
		} else if paged == nil {
			return result, nil
		}
		if result == nil {
			result = W{}
		}
		for _, processing := range paged {
			if processed, err := processor(processing); err != nil {
				return result, err
			} else {
				result = append(result, processed)
			}
		}
		if len(paged) < size {
			return result, nil
		}
	}
}
