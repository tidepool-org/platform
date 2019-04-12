package transform

import (
	"image/png"
	"io"
	"math"

	"github.com/disintegration/imaging"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/image"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

type Transformer interface {
	CalculateTransform(contentAttributes *image.ContentAttributes, rendition *image.Rendition) (*Transform, error)
	TransformContent(reader io.Reader, transform *Transform) (io.ReadCloser, error)
}

type Transform struct {
	Rendition     image.Rendition `json:"rendition,omitempty"`
	ContentWidth  int             `json:"contentWidth,omitempty"`
	ContentHeight int             `json:"contentHeight,omitempty"`
	Resize        bool            `json:"resize,omitempty"`
	Crop          bool            `json:"crop,omitempty"`
}

func NewTransform() *Transform {
	return &Transform{}
}

func NewTransformWithRendition(rendition *image.Rendition) (*Transform, error) {
	if rendition == nil {
		return nil, errors.New("rendition is missing")
	}
	if rendition.Width == nil {
		return nil, errors.New("rendition width is missing")
	}
	if rendition.Height == nil {
		return nil, errors.New("rendition height is missing")
	}

	return &Transform{
		Rendition:     *rendition,
		ContentWidth:  *rendition.Width,
		ContentHeight: *rendition.Height,
	}, nil
}

func (t *Transform) Validate(validator structure.Validator) {
	t.Rendition.Validate(validator.WithReference("rendition"))
	validator.Int("width", &t.ContentWidth).InRange(image.WidthMinimum, image.WidthMaximum)
	validator.Int("height", &t.ContentHeight).InRange(image.HeightMinimum, image.HeightMaximum)
	if !t.Resize {
		validator.Bool("crop", &t.Crop).False()
	}
}

func (t *Transform) ConstrainContentWidth(aspectRatio float64) {
	t.ContentWidth = int(math.Round(float64(t.ContentHeight) * aspectRatio))
}

func (t *Transform) ConstrainContentHeight(aspectRatio float64) {
	t.ContentHeight = int(math.Round(float64(t.ContentWidth) / aspectRatio))
}

func (t *Transform) ConstrainWidth(aspectRatio float64) {
	t.ConstrainContentWidth(aspectRatio)
	t.Rendition.Width = pointer.FromInt(t.ContentWidth)
}

func (t *Transform) ConstrainHeight(aspectRatio float64) {
	t.ConstrainContentHeight(aspectRatio)
	t.Rendition.Height = pointer.FromInt(t.ContentHeight)
}

func (t *Transform) Reset(contentAttributes *image.ContentAttributes) error {
	if contentAttributes == nil {
		return errors.New("content attributes is missing")
	}
	if contentAttributes.Width == nil {
		return errors.New("content attributes width is missing")
	}
	if contentAttributes.Height == nil {
		return errors.New("content attributes height is missing")
	}

	t.Rendition.Width = contentAttributes.Width
	t.Rendition.Height = contentAttributes.Height
	t.Rendition.Mode = pointer.FromString(image.ModeScale)
	if t.Rendition.SupportsQuality() {
		t.Rendition.Quality = pointer.FromInt(image.QualityDefault)
	}
	t.ContentWidth = *contentAttributes.Width
	t.ContentHeight = *contentAttributes.Height
	t.Resize = false
	t.Crop = false
	return nil
}

type TransformerImpl struct{}

func NewTransformer() *TransformerImpl {
	return &TransformerImpl{}
}

func (t *TransformerImpl) CalculateTransform(contentAttributes *image.ContentAttributes, rendition *image.Rendition) (*Transform, error) {
	if contentAttributes == nil {
		return nil, errors.New("content attributes is missing")
	} else if err := structureValidator.New().Validate(contentAttributes); err != nil {
		return nil, errors.Wrap(err, "content attributes is invalid")
	}
	if rendition == nil {
		return nil, errors.New("rendition is missing")
	} else if err := structureValidator.New().Validate(rendition); err != nil {
		return nil, errors.Wrap(err, "rendition is invalid")
	}

	contentAttributesAspectRatio := float64(*contentAttributes.Width) / float64(*contentAttributes.Height)
	transform, err := NewTransformWithRendition(rendition.WithDefaults(contentAttributesAspectRatio))
	if err != nil {
		return nil, err
	}
	transformRenditionAspectRatio := float64(*transform.Rendition.Width) / float64(*transform.Rendition.Height)

	if transformRenditionAspectRatio < contentAttributesAspectRatio {
		switch *transform.Rendition.Mode {
		case image.ModeFill, image.ModeFillDown:
			transform.ConstrainContentWidth(contentAttributesAspectRatio)
		case image.ModeFit, image.ModeFitDown:
			transform.ConstrainHeight(contentAttributesAspectRatio)
		case image.ModePad, image.ModePadDown:
			transform.ConstrainContentHeight(contentAttributesAspectRatio)
		}
	} else if transformRenditionAspectRatio > contentAttributesAspectRatio {
		switch *transform.Rendition.Mode {
		case image.ModeFill, image.ModeFillDown:
			transform.ConstrainContentHeight(contentAttributesAspectRatio)
		case image.ModeFit, image.ModeFitDown:
			transform.ConstrainWidth(contentAttributesAspectRatio)
		case image.ModePad, image.ModePadDown:
			transform.ConstrainContentWidth(contentAttributesAspectRatio)
		}
	}

	switch *transform.Rendition.Mode {
	case image.ModeFillDown, image.ModeFitDown, image.ModePadDown, image.ModeScaleDown:
		if transform.ContentWidth <= *contentAttributes.Width && transform.ContentHeight <= *contentAttributes.Height {
			transform.Rendition.Mode = pointer.FromString(image.NormalizeMode(*transform.Rendition.Mode))
		} else {
			transform.Reset(contentAttributes)
		}
	}

	if *transform.Rendition.Width == transform.ContentWidth && *transform.Rendition.Height == transform.ContentHeight {
		transform.Rendition.Mode = pointer.FromString(image.ModeScale)
	}

	if *transform.Rendition.Mode != image.ModePad && (!contentAttributes.SupportsTransparency() || transform.Rendition.SupportsTransparency()) {
		transform.Rendition.Background = nil
	}

	transform.Resize = transform.ContentWidth != *contentAttributes.Width || transform.ContentHeight != *contentAttributes.Height
	transform.Crop = *transform.Rendition.Mode == image.ModeFill

	return transform, nil
}

func (t *TransformerImpl) TransformContent(reader io.Reader, transform *Transform) (io.ReadCloser, error) {
	if reader == nil {
		return nil, errors.New("reader is missing")
	}
	if transform == nil {
		return nil, errors.New("transform is missing")
	} else if err := structureValidator.New().Validate(transform); err != nil {
		return nil, errors.Wrap(err, "transform is invalid")
	}

	var format imaging.Format
	var encodeOptions []imaging.EncodeOption
	switch *transform.Rendition.MediaType {
	case image.MediaTypeImageJPEG:
		format = imaging.JPEG
		if transform.Rendition.Quality != nil {
			encodeOptions = append(encodeOptions, imaging.JPEGQuality(*transform.Rendition.Quality))
		} else {
			encodeOptions = append(encodeOptions, imaging.JPEGQuality(image.QualityDefault))
		}
	case image.MediaTypeImagePNG:
		format = imaging.PNG
		encodeOptions = append(encodeOptions, imaging.PNGCompressionLevel(pngCompressionLevelDefault))
	}

	content, err := imaging.Decode(reader)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to decode content")
	}

	if transform.Resize {
		content = imaging.Resize(content, transform.ContentWidth, transform.ContentHeight, resampleFilterDefault)
	}

	if transform.Rendition.Background != nil {
		if transform.Rendition.SupportsTransparency() {
			content = imaging.PasteCenter(imaging.New(*transform.Rendition.Width, *transform.Rendition.Height, *transform.Rendition.Background), content)
		} else {
			content = imaging.OverlayCenter(imaging.New(*transform.Rendition.Width, *transform.Rendition.Height, *transform.Rendition.Background), content, 1.0)
		}
	} else if transform.Crop {
		content = imaging.CropCenter(content, *transform.Rendition.Width, *transform.Rendition.Height)
	}

	pipeReader, pipeWriter := io.Pipe()
	go func() {
		var encodeErr error
		defer func() {
			pipeWriter.CloseWithError(encodeErr)
		}()

		if encodeErr = imaging.Encode(pipeWriter, content, format, encodeOptions...); encodeErr != nil {
			encodeErr = errors.Wrapf(encodeErr, "unable to encode content")
		}
	}()

	return pipeReader, nil
}

const pngCompressionLevelDefault = png.DefaultCompression

var resampleFilterDefault = imaging.Lanczos
