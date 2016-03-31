package validate

type Error struct {
	Source map[string]string `json:"source"`
	Title  string            `json:"title"`
	Detail string            `json:"detail"`
}

type ErrorsArray struct {
	Errors []*Error `json:"errors"`
}

type ErrorProcessing struct {
	BasePath string
	*ErrorsArray
}

func (e ErrorProcessing) AppendPointerError(fieldName, title, detail string) {
	e.ErrorsArray.Append(NewPointerError(e.BasePath+"/"+fieldName, title, detail))
}

func NewErrorsArray() *ErrorsArray {
	return &ErrorsArray{Errors: []*Error{}}
}

func NewPointerError(path, title, detail string) *Error {
	return &Error{Source: map[string]string{"pointer": path}, Title: title, Detail: detail}
}

func NewParameterError(name, title, detail string) *Error {
	return &Error{Source: map[string]string{"parameter": name}, Title: title, Detail: detail}
}

func (e *ErrorsArray) Append(err *Error) {
	e.Errors = append(e.Errors, err)
}

func (e *ErrorsArray) HasErrors() bool {
	if e == nil || len(e.Errors) == 0 {
		return false
	}
	return true
}
