package test

import (
	"github.com/tidepool-org/platform/image"
	imageMultipart "github.com/tidepool-org/platform/image/multipart"
)

type Provider struct {
	ImageClientInvocations               int
	ImageClientStub                      func() image.Client
	ImageClientOutputs                   []image.Client
	ImageClientOutput                    *image.Client
	ImageMultipartFormDecoderInvocations int
	ImageMultipartFormDecoderStub        func() imageMultipart.FormDecoder
	ImageMultipartFormDecoderOutputs     []imageMultipart.FormDecoder
	ImageMultipartFormDecoderOutput      *imageMultipart.FormDecoder
}

func NewProvider() *Provider {
	return &Provider{}
}

func (p *Provider) ImageClient() image.Client {
	p.ImageClientInvocations++
	if p.ImageClientStub != nil {
		return p.ImageClientStub()
	}
	if len(p.ImageClientOutputs) > 0 {
		output := p.ImageClientOutputs[0]
		p.ImageClientOutputs = p.ImageClientOutputs[1:]
		return output
	}
	if p.ImageClientOutput != nil {
		return *p.ImageClientOutput
	}
	panic("ImageClient has no output")
}

func (p *Provider) ImageMultipartFormDecoder() imageMultipart.FormDecoder {
	p.ImageMultipartFormDecoderInvocations++
	if p.ImageMultipartFormDecoderStub != nil {
		return p.ImageMultipartFormDecoderStub()
	}
	if len(p.ImageMultipartFormDecoderOutputs) > 0 {
		output := p.ImageMultipartFormDecoderOutputs[0]
		p.ImageMultipartFormDecoderOutputs = p.ImageMultipartFormDecoderOutputs[1:]
		return output
	}
	if p.ImageMultipartFormDecoderOutput != nil {
		return *p.ImageMultipartFormDecoderOutput
	}
	panic("ImageMultipartFormDecoder has no output")
}

func (p *Provider) AssertOutputsEmpty() {
	if len(p.ImageMultipartFormDecoderOutputs) > 0 {
		panic("ImageMultipartFormDecoderOutputs is not empty")
	}
	if len(p.ImageClientOutputs) > 0 {
		panic("ImageClientOutputs is not empty")
	}
}
