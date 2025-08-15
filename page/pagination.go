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
