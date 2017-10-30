package test

type Mock struct {
	ID string
}

func NewMock() *Mock {
	return &Mock{
		ID: NewString(32, CharsetAlphaNumeric),
	}
}

func (m *Mock) Expectations() {
}
