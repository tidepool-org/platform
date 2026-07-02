package test

import (
	"github.com/onsi/gomega"

	"github.com/tidepool-org/platform/provider"
)

type GetInput struct {
	Type string
	Name string
}

type Factory struct {
	GetInvocations int
	GetInputs      []GetInput
	GetOutputs     []provider.Provider
}

func NewFactory() *Factory {
	return &Factory{}
}

func (f *Factory) Get(typ string, name string) provider.Provider {
	f.GetInvocations++

	f.GetInputs = append(f.GetInputs, GetInput{Type: typ, Name: name})

	gomega.Expect(f.GetOutputs).ToNot(gomega.BeEmpty())

	output := f.GetOutputs[0]
	f.GetOutputs = f.GetOutputs[1:]
	return output
}

func (f *Factory) Expectations() {
	gomega.Expect(f.GetOutputs).To(gomega.BeEmpty())
}
