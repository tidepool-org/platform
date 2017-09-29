package test

import (
	"github.com/onsi/gomega"

	"github.com/tidepool-org/platform/provider"
	"github.com/tidepool-org/platform/test"
)

type GetInput struct {
	Type string
	Name string
}

type GetOutput struct {
	Provider provider.Provider
	Error    error
}

type Factory struct {
	*test.Mock
	GetInvocations int
	GetInputs      []GetInput
	GetOutputs     []GetOutput
}

func NewFactory() *Factory {
	return &Factory{
		Mock: test.NewMock(),
	}
}

func (f *Factory) Get(typ string, name string) (provider.Provider, error) {
	f.GetInvocations++

	f.GetInputs = append(f.GetInputs, GetInput{Type: typ, Name: name})

	gomega.Expect(f.GetOutputs).ToNot(gomega.BeEmpty())

	output := f.GetOutputs[0]
	f.GetOutputs = f.GetOutputs[1:]
	return output.Provider, output.Error
}

func (f *Factory) Expectations() {
	f.Mock.Expectations()
	gomega.Expect(f.GetOutputs).To(gomega.BeEmpty())
}
