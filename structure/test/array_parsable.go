package test

import "github.com/tidepool-org/platform/structure"

type ArrayParsable struct {
	ParseInvocations int
	ParseInputs      []structure.ArrayParser
}

func NewArrayParsable() *ArrayParsable {
	return &ArrayParsable{}
}

func (a *ArrayParsable) Parse(objectParser structure.ArrayParser) {
	a.ParseInvocations++

	a.ParseInputs = append(a.ParseInputs, objectParser)
}

func (a *ArrayParsable) Expectations() {}
