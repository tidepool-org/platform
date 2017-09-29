package test

import (
	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/test"
)

type ArrayParsable struct {
	*test.Mock
	ParseInvocations int
	ParseInputs      []structure.ArrayParser
}

func NewArrayParsable() *ArrayParsable {
	return &ArrayParsable{
		Mock: test.NewMock(),
	}
}

func (a *ArrayParsable) Parse(objectParser structure.ArrayParser) {
	a.ParseInvocations++

	a.ParseInputs = append(a.ParseInputs, objectParser)
}

func (a *ArrayParsable) Expectations() {
	a.Mock.Expectations()
}
