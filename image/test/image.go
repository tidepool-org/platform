package test

import (
	"bytes"
	"image/png"
	"io/ioutil"
	"math"
	"time"

	"github.com/disintegration/imaging"
	"github.com/onsi/gomega"
	gomegaGstruct "github.com/onsi/gomega/gstruct"
	gomegaTypes "github.com/onsi/gomega/types"

	"github.com/tidepool-org/platform/crypto"
	cryptoTest "github.com/tidepool-org/platform/crypto/test"
	"github.com/tidepool-org/platform/image"
	"github.com/tidepool-org/platform/pointer"
	requestTest "github.com/tidepool-org/platform/request/test"
	"github.com/tidepool-org/platform/test"
	userTest "github.com/tidepool-org/platform/user/test"
)

func RandomContentIntent() string {
	return test.RandomStringFromArray(image.ContentIntents())
}

func RandomContentIntents() []string {
	return test.RandomStringArrayFromRangeAndArrayWithoutDuplicates(1, len(image.ContentIntents()), image.ContentIntents())
}

func RandomMediaType() string {
	return test.RandomStringFromArray(image.MediaTypes())
}

func RandomMode() string {
	return test.RandomStringFromArray(image.Modes())
}

func RandomStatus() string {
	return test.RandomStringFromArray(image.Statuses())
}

func RandomStatuses() []string {
	return test.RandomStringArrayFromRangeAndArrayWithoutDuplicates(1, len(image.Statuses()), image.Statuses())
}

func RandomFilter() *image.Filter {
	datum := image.NewFilter()
	datum.Status = pointer.FromStringArray(RandomStatuses())
	datum.ContentIntent = pointer.FromStringArray(RandomContentIntents())
	return datum
}

func NewObjectFromFilter(datum *image.Filter, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.Status != nil {
		object["status"] = test.NewObjectFromStringArray(*datum.Status, objectFormat)
	}
	if datum.ContentIntent != nil {
		object["contentIntent"] = test.NewObjectFromStringArray(*datum.ContentIntent, objectFormat)
	}
	return object
}

func RandomMetadata() *image.Metadata {
	datum := image.NewMetadata()
	datum.Name = pointer.FromString(RandomName())
	return datum
}

func NewObjectFromMetadata(datum *image.Metadata, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.Name != nil {
		object["name"] = test.NewObjectFromString(*datum.Name, objectFormat)
	}
	return object
}

func RandomName() string {
	return test.RandomStringFromRange(1, image.NameLengthMaximum)
}

func RandomContent() *image.Content {
	mediaType := RandomMediaType()
	contentBytes := RandomContentBytesFromMediaType(mediaType)
	datum := image.NewContent()
	datum.Body = ioutil.NopCloser(bytes.NewReader(contentBytes))
	datum.DigestMD5 = pointer.FromString(crypto.Base64EncodedMD5Hash(contentBytes))
	datum.MediaType = pointer.FromString(mediaType)
	return datum
}

func RandomContentFromDimensions(width int, height int) *image.Content {
	mediaType := RandomMediaType()
	contentBytes := RandomContentBytesFromDimensionsAndMediaType(width, height, mediaType)
	datum := image.NewContent()
	datum.Body = ioutil.NopCloser(bytes.NewReader(contentBytes))
	datum.DigestMD5 = pointer.FromString(crypto.Base64EncodedMD5Hash(contentBytes))
	datum.MediaType = pointer.FromString(mediaType)
	return datum
}

func RandomContentBytes() []byte {
	return RandomContentBytesFromMediaType(RandomMediaType())
}

func RandomContentBytesFromMediaType(mediaType string) []byte {
	return RandomContentBytesFromDimensionsAndMediaType(test.RandomIntFromRange(10, 20), test.RandomIntFromRange(10, 20), mediaType)
}

func RandomContentBytesFromDimensionsAndMediaType(width int, height int, mediaType string) []byte {
	var format imaging.Format
	var options []imaging.EncodeOption

	switch mediaType {
	case image.MediaTypeImageJPEG:
		format = imaging.JPEG
		options = append(options, imaging.JPEGQuality(95))
	case image.MediaTypeImagePNG:
		format = imaging.PNG
		options = append(options, imaging.PNGCompressionLevel(png.DefaultCompression))
	default:
		panic("RandomContentBytesFromDimensionsAndMediaType: unexpected media type")
	}

	contentBytes := &bytes.Buffer{}
	if err := imaging.Encode(contentBytes, imaging.New(width, height, RandomColor().NRGBA), format, options...); err != nil {
		panic("RandomContentBytesFromDimensionsAndMediaType: unable to encode image")
	}
	return contentBytes.Bytes()
}

func RandomGeneration() int {
	return test.RandomIntFromRange(0, math.MaxInt32)
}

func RandomRendition() *image.Rendition {
	datum := image.NewRendition()
	datum.MediaType = pointer.FromString(RandomMediaType())
	datum.Width = pointer.FromInt(RandomWidth())
	datum.Height = pointer.FromInt(RandomHeight())
	datum.Mode = pointer.FromString(RandomMode())
	datum.Background = RandomColor()
	if *datum.MediaType == image.MediaTypeImageJPEG {
		datum.Quality = pointer.FromInt(RandomQuality())
	}
	return datum
}

func CloneRendition(datum *image.Rendition) *image.Rendition {
	if datum == nil {
		return nil
	}
	clone := image.NewRendition()
	clone.MediaType = pointer.CloneString(datum.MediaType)
	clone.Width = pointer.CloneInt(datum.Width)
	clone.Height = pointer.CloneInt(datum.Height)
	clone.Mode = pointer.CloneString(datum.Mode)
	clone.Background = CloneColor(datum.Background)
	clone.Quality = pointer.CloneInt(datum.Quality)
	return clone
}

func NewObjectFromRendition(datum *image.Rendition, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.MediaType != nil {
		object["mediaType"] = test.NewObjectFromString(*datum.MediaType, objectFormat)
	}
	if datum.Width != nil {
		object["width"] = test.NewObjectFromInt(*datum.Width, objectFormat)
	}
	if datum.Height != nil {
		object["height"] = test.NewObjectFromInt(*datum.Height, objectFormat)
	}
	if datum.Mode != nil {
		object["mode"] = test.NewObjectFromString(*datum.Mode, objectFormat)
	}
	if datum.Background != nil {
		object["background"] = test.NewObjectFromString(datum.Background.String(), objectFormat)
	}
	if datum.Quality != nil {
		object["quality"] = test.NewObjectFromInt(*datum.Quality, objectFormat)
	}
	return object
}

func RandomRenditionAsString() string {
	return RandomRendition().String()
}

func RandomRenditionsAsStrings() []string {
	datum := make([]string, test.RandomIntFromRange(2, 3))
	for index := range datum {
		datum[index] = RandomRenditionAsString()
	}
	return datum
}

func RandomWidth() int {
	return test.RandomIntFromRange(image.WidthMinimum, image.WidthMaximum)
}

func RandomHeight() int {
	return test.RandomIntFromRange(image.HeightMinimum, image.HeightMaximum)
}

func RandomColor() *image.Color {
	return image.NewColor(uint8(test.RandomInt()), uint8(test.RandomInt()), uint8(test.RandomInt()), uint8(test.RandomInt()))
}

func CloneColor(datum *image.Color) *image.Color {
	if datum == nil {
		return nil
	}
	return &image.Color{
		NRGBA: datum.NRGBA,
	}
}

func RandomQuality() int {
	return test.RandomIntFromRange(image.QualityMinimum, image.QualityMaximum)
}

func RandomContentAttributes() *image.ContentAttributes {
	datum := image.NewContentAttributes()
	datum.DigestMD5 = pointer.FromString(cryptoTest.RandomBase64EncodedMD5Hash())
	datum.MediaType = pointer.FromString(RandomMediaType())
	datum.Width = pointer.FromInt(RandomWidth())
	datum.Height = pointer.FromInt(RandomHeight())
	datum.Size = pointer.FromInt(test.RandomIntFromRange(1, 100*1024*1024))
	datum.CreatedTime = pointer.FromTime(test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now()).Truncate(time.Second))
	datum.ModifiedTime = pointer.FromTime(test.RandomTimeFromRange(*datum.CreatedTime, time.Now()).Truncate(time.Second))
	return datum
}

func CloneContentAttributes(datum *image.ContentAttributes) *image.ContentAttributes {
	if datum == nil {
		return nil
	}
	clone := image.NewContentAttributes()
	clone.DigestMD5 = pointer.CloneString(datum.DigestMD5)
	clone.MediaType = pointer.CloneString(datum.MediaType)
	clone.Width = pointer.CloneInt(datum.Width)
	clone.Height = pointer.CloneInt(datum.Height)
	clone.Size = pointer.CloneInt(datum.Size)
	clone.CreatedTime = pointer.CloneTime(datum.CreatedTime)
	clone.ModifiedTime = pointer.CloneTime(datum.ModifiedTime)
	return clone
}

func NewObjectFromContentAttributes(datum *image.ContentAttributes, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.DigestMD5 != nil {
		object["digestMD5"] = test.NewObjectFromString(*datum.DigestMD5, objectFormat)
	}
	if datum.MediaType != nil {
		object["mediaType"] = test.NewObjectFromString(*datum.MediaType, objectFormat)
	}
	if datum.Width != nil {
		object["width"] = test.NewObjectFromInt(*datum.Width, objectFormat)
	}
	if datum.Height != nil {
		object["height"] = test.NewObjectFromInt(*datum.Height, objectFormat)
	}
	if datum.Size != nil {
		object["size"] = test.NewObjectFromInt(*datum.Size, objectFormat)
	}
	if datum.CreatedTime != nil {
		object["createdTime"] = test.NewObjectFromTime(*datum.CreatedTime, objectFormat)
	}
	if datum.ModifiedTime != nil {
		object["modifiedTime"] = test.NewObjectFromTime(*datum.ModifiedTime, objectFormat)
	}
	return object
}

func MatchContentAttributes(datum *image.ContentAttributes) gomegaTypes.GomegaMatcher {
	if datum == nil {
		return gomega.BeNil()
	}
	return gomegaGstruct.PointTo(gomegaGstruct.MatchAllFields(gomegaGstruct.Fields{
		"DigestMD5":    gomega.Equal(datum.DigestMD5),
		"MediaType":    gomega.Equal(datum.MediaType),
		"Width":        gomega.Equal(datum.Width),
		"Height":       gomega.Equal(datum.Height),
		"Size":         gomega.Equal(datum.Size),
		"CreatedTime":  test.MatchTime(datum.CreatedTime),
		"ModifiedTime": test.MatchTime(datum.ModifiedTime),
	}))
}

func RandomImage() *image.Image {
	datum := &image.Image{}
	datum.ID = pointer.FromString(RandomID())
	datum.UserID = pointer.FromString(userTest.RandomID())
	datum.Status = pointer.FromString(RandomStatus())
	datum.Name = pointer.FromString(RandomName())
	if *datum.Status == image.StatusAvailable {
		datum.ContentID = pointer.FromString(RandomContentID())
		datum.ContentIntent = pointer.FromString(RandomContentIntent())
		datum.ContentAttributes = RandomContentAttributes()
		datum.RenditionsID = pointer.FromString(RandomRenditionsID())
		datum.Renditions = pointer.FromStringArray(RandomRenditionsAsStrings())
	}
	datum.CreatedTime = pointer.FromTime(test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now()).Truncate(time.Second))
	if *datum.Status == image.StatusAvailable {
		datum.ModifiedTime = pointer.FromTime(test.RandomTimeFromRange(*datum.CreatedTime, time.Now()).Truncate(time.Second))
	}
	datum.Revision = pointer.FromInt(requestTest.RandomRevision())
	return datum
}

func CloneImage(datum *image.Image) *image.Image {
	if datum == nil {
		return nil
	}
	clone := &image.Image{}
	clone.ID = pointer.CloneString(datum.ID)
	clone.UserID = pointer.CloneString(datum.UserID)
	clone.Status = pointer.CloneString(datum.Status)
	clone.Name = pointer.CloneString(datum.Name)
	clone.ContentID = pointer.CloneString(datum.ContentID)
	clone.ContentIntent = pointer.CloneString(datum.ContentIntent)
	clone.ContentAttributes = CloneContentAttributes(datum.ContentAttributes)
	clone.RenditionsID = pointer.CloneString(datum.RenditionsID)
	clone.Renditions = pointer.CloneStringArray(datum.Renditions)
	clone.CreatedTime = pointer.CloneTime(datum.CreatedTime)
	clone.ModifiedTime = pointer.CloneTime(datum.ModifiedTime)
	clone.DeletedTime = pointer.CloneTime(datum.DeletedTime)
	clone.Revision = pointer.CloneInt(datum.Revision)
	return clone
}

func NewObjectFromImage(datum *image.Image, objectFormat test.ObjectFormat) map[string]interface{} {
	if datum == nil {
		return nil
	}
	object := map[string]interface{}{}
	if datum.ID != nil {
		object["id"] = test.NewObjectFromString(*datum.ID, objectFormat)
	}
	if datum.UserID != nil {
		object["userId"] = test.NewObjectFromString(*datum.UserID, objectFormat)
	}
	if datum.Status != nil {
		object["status"] = test.NewObjectFromString(*datum.Status, objectFormat)
	}
	if datum.Name != nil {
		object["name"] = test.NewObjectFromString(*datum.Name, objectFormat)
	}
	if datum.ContentID != nil {
		object["contentId"] = test.NewObjectFromString(*datum.ContentID, objectFormat)
	}
	if datum.ContentIntent != nil {
		object["contentIntent"] = test.NewObjectFromString(*datum.ContentIntent, objectFormat)
	}
	if datum.ContentAttributes != nil {
		object["contentAttributes"] = NewObjectFromContentAttributes(datum.ContentAttributes, objectFormat)
	}
	if datum.RenditionsID != nil {
		object["renditionsId"] = test.NewObjectFromString(*datum.RenditionsID, objectFormat)
	}
	if datum.Renditions != nil {
		object["renditions"] = test.NewObjectFromStringArray(*datum.Renditions, objectFormat)
	}
	if datum.CreatedTime != nil {
		object["createdTime"] = test.NewObjectFromTime(*datum.CreatedTime, objectFormat)
	}
	if datum.ModifiedTime != nil {
		object["modifiedTime"] = test.NewObjectFromTime(*datum.ModifiedTime, objectFormat)
	}
	if datum.DeletedTime != nil {
		object["deletedTime"] = test.NewObjectFromTime(*datum.DeletedTime, objectFormat)
	}
	if datum.Revision != nil {
		object["revision"] = test.NewObjectFromInt(*datum.Revision, objectFormat)
	}
	return object
}

func MatchImage(datum *image.Image) gomegaTypes.GomegaMatcher {
	if datum == nil {
		return gomega.BeNil()
	}
	return gomegaGstruct.PointTo(gomegaGstruct.MatchAllFields(gomegaGstruct.Fields{
		"ID":                gomega.Equal(datum.ID),
		"UserID":            gomega.Equal(datum.UserID),
		"Status":            gomega.Equal(datum.Status),
		"Name":              gomega.Equal(datum.Name),
		"ContentID":         gomega.Equal(datum.ContentID),
		"ContentIntent":     gomega.Equal(datum.ContentIntent),
		"ContentAttributes": MatchContentAttributes(datum.ContentAttributes),
		"RenditionsID":      gomega.Equal(datum.RenditionsID),
		"Renditions":        gomega.Equal(datum.Renditions),
		"CreatedTime":       test.MatchTime(datum.CreatedTime),
		"ModifiedTime":      test.MatchTime(datum.ModifiedTime),
		"DeletedTime":       test.MatchTime(datum.DeletedTime),
		"Revision":          gomega.Equal(datum.Revision),
	}))
}

func RandomImages(minimumLength int, maximumLength int) image.Images {
	datum := make(image.Images, test.RandomIntFromRange(minimumLength, maximumLength))
	for index := range datum {
		datum[index] = RandomImage()
	}
	return datum
}

func CloneImages(datum image.Images) image.Images {
	if datum == nil {
		return nil
	}
	clone := make(image.Images, len(datum))
	for index := range datum {
		clone[index] = CloneImage(datum[index])
	}
	return clone
}

func MatchImages(datum image.Images) gomegaTypes.GomegaMatcher {
	matchers := []gomegaTypes.GomegaMatcher{}
	for _, d := range datum {
		matchers = append(matchers, MatchImage(d))
	}
	return test.MatchArray(matchers)
}

func RandomID() string {
	return image.NewID()
}

func RandomContentID() string {
	return image.NewContentID()
}

func RandomRenditionsID() string {
	return image.NewRenditionsID()
}
