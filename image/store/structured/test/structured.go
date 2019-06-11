package test

import (
	cryptoTest "github.com/tidepool-org/platform/crypto/test"
	imageStoreStructured "github.com/tidepool-org/platform/image/store/structured"
	imageTest "github.com/tidepool-org/platform/image/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomUpdate() *imageStoreStructured.Update {
	datum := imageStoreStructured.NewUpdate()
	datum.Metadata = imageTest.RandomMetadata()
	datum.ContentID = pointer.FromString(imageTest.RandomContentID())
	datum.ContentIntent = pointer.FromString(imageTest.RandomContentIntent())
	datum.ContentAttributes = RandomContentAttributes()
	datum.RenditionsID = nil
	datum.Rendition = nil
	return datum
}

func RandomContentAttributes() *imageStoreStructured.ContentAttributes {
	datum := imageStoreStructured.NewContentAttributes()
	datum.DigestMD5 = pointer.FromString(cryptoTest.RandomBase64EncodedMD5Hash())
	datum.MediaType = pointer.FromString(imageTest.RandomMediaType())
	datum.Width = pointer.FromInt(imageTest.RandomWidth())
	datum.Height = pointer.FromInt(imageTest.RandomHeight())
	datum.Size = pointer.FromInt(test.RandomIntFromRange(1, 100*1024*1024))
	return datum
}
