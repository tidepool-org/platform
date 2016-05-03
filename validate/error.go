package validate

import (
	"fmt"

	"github.com/tidepool-org/platform/service"
)

// TODO: ErrorProcessing should be renamed Context as it is the context of validation, not just errors

type ErrorProcessing struct {
	*service.Errors
	pointer string
}

func NewErrorProcessing(pointer string) ErrorProcessing {
	return ErrorProcessing{
		Errors:  service.NewErrors(),
		pointer: pointer,
	}
}

func (e *ErrorProcessing) Pointer() string {
	return e.pointer
}

func (e *ErrorProcessing) AppendPointerError(pointer string, title string, detail string) {
	e.AppendError(&service.Error{
		Source: &service.Source{
			Pointer: fmt.Sprintf("%s/%s", e.pointer, pointer),
		},
		Title:  title,
		Detail: detail,
	})
}
