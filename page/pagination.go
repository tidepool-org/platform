package page

import (
	"net/http"
	"strconv"

	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/structure"
)

const (
	PaginationPageMinimum = 0
	PaginationSizeMinimum = 1
	PaginationSizeMaximum = 100
)

type Pagination struct {
	Page int `json:"page,omitempty"`
	Size int `json:"size,omitempty"`
}

func NewPagination() *Pagination {
	return &Pagination{
		Size: PaginationSizeMaximum,
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

func (p *Pagination) Mutate(req *http.Request) error {
	parameters := map[string]string{
		"page": strconv.Itoa(p.Page),
		"size": strconv.Itoa(p.Size),
	}
	return request.NewParametersMutator(parameters).Mutate(req)
}
