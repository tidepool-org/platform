package test

import "github.com/tidepool-org/platform/id"

type Mock struct {
	ID string
}

func NewMock() *Mock {
	return &Mock{
		ID: id.New(),
	}
}
