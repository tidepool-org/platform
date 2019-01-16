package test

import (
	"io"

	"github.com/tidepool-org/platform/image"
	imageTransform "github.com/tidepool-org/platform/image/transform"
)

type CalculateTransformInput struct {
	ContentAttributes *image.ContentAttributes
	Rendition         *image.Rendition
}

type CalculateTransformOutput struct {
	Transform *imageTransform.Transform
	Error     error
}

type TransformContentInput struct {
	Reader    io.Reader
	Transform *imageTransform.Transform
}

type TransformContentOutput struct {
	Reader io.ReadCloser
	Error  error
}

type Transformer struct {
	CalculateTransformInvocations int
	CalculateTransformInputs      []CalculateTransformInput
	CalculateTransformStub        func(contentAttributes *image.ContentAttributes, rendition *image.Rendition) (*imageTransform.Transform, error)
	CalculateTransformOutputs     []CalculateTransformOutput
	CalculateTransformOutput      *CalculateTransformOutput
	TransformContentInvocations   int
	TransformContentInputs        []TransformContentInput
	TransformContentStub          func(reader io.Reader, transform *imageTransform.Transform) (io.ReadCloser, error)
	TransformContentOutputs       []TransformContentOutput
	TransformContentOutput        *TransformContentOutput
}

func NewTransformer() *Transformer {
	return &Transformer{}
}

func (t *Transformer) CalculateTransform(contentAttributes *image.ContentAttributes, rendition *image.Rendition) (*imageTransform.Transform, error) {
	t.CalculateTransformInvocations++
	t.CalculateTransformInputs = append(t.CalculateTransformInputs, CalculateTransformInput{ContentAttributes: contentAttributes, Rendition: rendition})
	if t.CalculateTransformStub != nil {
		return t.CalculateTransformStub(contentAttributes, rendition)
	}
	if len(t.CalculateTransformOutputs) > 0 {
		output := t.CalculateTransformOutputs[0]
		t.CalculateTransformOutputs = t.CalculateTransformOutputs[1:]
		return output.Transform, output.Error
	}
	if t.CalculateTransformOutput != nil {
		return t.CalculateTransformOutput.Transform, t.CalculateTransformOutput.Error
	}
	panic("CalculateTransform has no output")
}

func (t *Transformer) TransformContent(reader io.Reader, transform *imageTransform.Transform) (io.ReadCloser, error) {
	t.TransformContentInvocations++
	t.TransformContentInputs = append(t.TransformContentInputs, TransformContentInput{Reader: reader, Transform: transform})
	if t.TransformContentStub != nil {
		return t.TransformContentStub(reader, transform)
	}
	if len(t.TransformContentOutputs) > 0 {
		output := t.TransformContentOutputs[0]
		t.TransformContentOutputs = t.TransformContentOutputs[1:]
		return output.Reader, output.Error
	}
	if t.TransformContentOutput != nil {
		return t.TransformContentOutput.Reader, t.TransformContentOutput.Error
	}
	panic("TransformContent has no output")
}

func (t *Transformer) AssertOutputsEmpty() {
	if len(t.CalculateTransformOutputs) > 0 {
		panic("CalculateTransformOutputs is not empty")
	}
	if len(t.TransformContentOutputs) > 0 {
		panic("TransformContentOutputs is not empty")
	}
}
