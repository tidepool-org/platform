package test

import (
	imageTest "github.com/tidepool-org/platform/image/test"
	imageTransform "github.com/tidepool-org/platform/image/transform"
	"github.com/tidepool-org/platform/test"
)

func RandomTransform() *imageTransform.Transform {
	datum := imageTransform.NewTransform()
	datum.Rendition = *imageTest.RandomRendition()
	datum.ContentWidth = imageTest.RandomWidth()
	datum.ContentHeight = imageTest.RandomHeight()
	datum.Resize = test.RandomBool()
	datum.Crop = datum.Resize && test.RandomBool()
	return datum
}

func CloneTransform(datum *imageTransform.Transform) *imageTransform.Transform {
	if datum == nil {
		return nil
	}
	clone := imageTransform.NewTransform()
	clone.Rendition = *imageTest.CloneRendition(&datum.Rendition)
	clone.ContentWidth = datum.ContentWidth
	clone.ContentHeight = datum.ContentHeight
	clone.Resize = datum.Resize
	clone.Crop = datum.Crop
	return clone
}
