package test

import (
	"io"

	"github.com/tidepool-org/platform/image"
)

type EncodeFormInput struct {
	Metadata      *image.Metadata
	ContentIntent string
	Content       *image.Content
}

type EncodeFormOutput struct {
	Reader      io.ReadCloser
	ContentType string
}

type FormEncoder struct {
	EncodeFormInvocations int
	EncodeFormInputs      []EncodeFormInput
	EncodeFormStub        func(metadata *image.Metadata, contentIntent string, content *image.Content) (io.ReadCloser, string)
	EncodeFormOutputs     []EncodeFormOutput
	EncodeFormOutput      *EncodeFormOutput
}

func NewFormEncoder() *FormEncoder {
	return &FormEncoder{}
}

func (f *FormEncoder) EncodeForm(metadata *image.Metadata, contentIntent string, content *image.Content) (io.ReadCloser, string) {
	f.EncodeFormInvocations++
	f.EncodeFormInputs = append(f.EncodeFormInputs, EncodeFormInput{Metadata: metadata, ContentIntent: contentIntent, Content: content})
	if f.EncodeFormStub != nil {
		return f.EncodeFormStub(metadata, contentIntent, content)
	}
	if len(f.EncodeFormOutputs) > 0 {
		output := f.EncodeFormOutputs[0]
		f.EncodeFormOutputs = f.EncodeFormOutputs[1:]
		return output.Reader, output.ContentType
	}
	if f.EncodeFormOutput != nil {
		return f.EncodeFormOutput.Reader, f.EncodeFormOutput.ContentType
	}
	panic("EncodeForm has no output")
}

func (f *FormEncoder) AssertOutputsEmpty() {
	if len(f.EncodeFormOutputs) > 0 {
		panic("EncodeFormOutputs is not empty")
	}
}
