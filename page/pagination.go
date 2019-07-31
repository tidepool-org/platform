package page

import (
	"net/http"
	"strconv"

	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/structure"
)

const (
	PaginationPageDefault = 0
	PaginationPageMinimum = 0
	PaginationSizeDefault = 100
	PaginationSizeMinimum = 1
	PaginationSizeMaximum = 1000
)

// FUTURE: Use pointers to Page and Size

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
