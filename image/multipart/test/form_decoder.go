package test

import (
	"io"

	"github.com/tidepool-org/platform/image"
)

type DecodeFormInput struct {
	Reader      io.Reader
	ContentType string
}

type DecodeFormOutput struct {
	Metadata      *image.Metadata
	ContentIntent string
	Content       *image.Content
	Error         error
}

type FormDecoder struct {
	DecodeFormInvocations int
	DecodeFormInputs      []DecodeFormInput
	DecodeFormStub        func(reader io.Reader, contentType string) (*image.Metadata, string, *image.Content, error)
	DecodeFormOutputs     []DecodeFormOutput
	DecodeFormOutput      *DecodeFormOutput
}

func NewFormDecoder() *FormDecoder {
	return &FormDecoder{}
}

func (f *FormDecoder) DecodeForm(reader io.Reader, contentType string) (*image.Metadata, string, *image.Content, error) {
	f.DecodeFormInvocations++
	f.DecodeFormInputs = append(f.DecodeFormInputs, DecodeFormInput{Reader: reader, ContentType: contentType})
	if f.DecodeFormStub != nil {
		return f.DecodeFormStub(reader, contentType)
	}
	if len(f.DecodeFormOutputs) > 0 {
		output := f.DecodeFormOutputs[0]
		f.DecodeFormOutputs = f.DecodeFormOutputs[1:]
		return output.Metadata, output.ContentIntent, output.Content, output.Error
	}
	if f.DecodeFormOutput != nil {
		return f.DecodeFormOutput.Metadata, f.DecodeFormOutput.ContentIntent, f.DecodeFormOutput.Content, f.DecodeFormOutput.Error
	}
	panic("DecodeForm has no output")
}

func (f *FormDecoder) AssertOutputsEmpty() {
	if len(f.DecodeFormOutputs) > 0 {
		panic("DecodeFormOutputs is not empty")
	}
}
