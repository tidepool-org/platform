package test

import (
	"github.com/tidepool-org/platform/auth"
	imageStoreStructured "github.com/tidepool-org/platform/image/store/structured"
	imageStoreUnstructured "github.com/tidepool-org/platform/image/store/unstructured"
	imageTransform "github.com/tidepool-org/platform/image/transform"
)

type Provider struct {
	AuthClientInvocations             int
	AuthClientStub                    func() auth.Client
	AuthClientOutputs                 []auth.Client
	AuthClientOutput                  *auth.Client
	ImageStructuredStoreInvocations   int
	ImageStructuredStoreStub          func() imageStoreStructured.Store
	ImageStructuredStoreOutputs       []imageStoreStructured.Store
	ImageStructuredStoreOutput        *imageStoreStructured.Store
	ImageUnstructuredStoreInvocations int
	ImageUnstructuredStoreStub        func() imageStoreUnstructured.Store
	ImageUnstructuredStoreOutputs     []imageStoreUnstructured.Store
	ImageUnstructuredStoreOutput      *imageStoreUnstructured.Store
	ImageTransformerInvocations       int
	ImageTransformerStub              func() imageTransform.Transformer
	ImageTransformerOutputs           []imageTransform.Transformer
	ImageTransformerOutput            *imageTransform.Transformer
}

func NewProvider() *Provider {
	return &Provider{}
}

func (p *Provider) AuthClient() auth.Client {
	p.AuthClientInvocations++
	if p.AuthClientStub != nil {
		return p.AuthClientStub()
	}
	if len(p.AuthClientOutputs) > 0 {
		output := p.AuthClientOutputs[0]
		p.AuthClientOutputs = p.AuthClientOutputs[1:]
		return output
	}
	if p.AuthClientOutput != nil {
		return *p.AuthClientOutput
	}
	panic("AuthClient has no output")
}

func (p *Provider) ImageStructuredStore() imageStoreStructured.Store {
	p.ImageStructuredStoreInvocations++
	if p.ImageStructuredStoreStub != nil {
		return p.ImageStructuredStoreStub()
	}
	if len(p.ImageStructuredStoreOutputs) > 0 {
		output := p.ImageStructuredStoreOutputs[0]
		p.ImageStructuredStoreOutputs = p.ImageStructuredStoreOutputs[1:]
		return output
	}
	if p.ImageStructuredStoreOutput != nil {
		return *p.ImageStructuredStoreOutput
	}
	panic("ImageStructuredStore has no output")
}

func (p *Provider) ImageUnstructuredStore() imageStoreUnstructured.Store {
	p.ImageUnstructuredStoreInvocations++
	if p.ImageUnstructuredStoreStub != nil {
		return p.ImageUnstructuredStoreStub()
	}
	if len(p.ImageUnstructuredStoreOutputs) > 0 {
		output := p.ImageUnstructuredStoreOutputs[0]
		p.ImageUnstructuredStoreOutputs = p.ImageUnstructuredStoreOutputs[1:]
		return output
	}
	if p.ImageUnstructuredStoreOutput != nil {
		return *p.ImageUnstructuredStoreOutput
	}
	panic("ImageUnstructuredStore has no output")
}

func (p *Provider) ImageTransformer() imageTransform.Transformer {
	p.ImageTransformerInvocations++
	if p.ImageTransformerStub != nil {
		return p.ImageTransformerStub()
	}
	if len(p.ImageTransformerOutputs) > 0 {
		output := p.ImageTransformerOutputs[0]
		p.ImageTransformerOutputs = p.ImageTransformerOutputs[1:]
		return output
	}
	if p.ImageTransformerOutput != nil {
		return *p.ImageTransformerOutput
	}
	panic("ImageTransformer has no output")
}

func (p *Provider) AssertOutputsEmpty() {
	if len(p.AuthClientOutputs) > 0 {
		panic("AuthClientOutputs is not empty")
	}
	if len(p.ImageStructuredStoreOutputs) > 0 {
		panic("ImageStructuredStoreOutputs is not empty")
	}
	if len(p.ImageUnstructuredStoreOutputs) > 0 {
		panic("ImageUnstructuredStoreOutputs is not empty")
	}
	if len(p.ImageTransformerOutputs) > 0 {
		panic("ImageTransformerOutputs is not empty")
	}
}
