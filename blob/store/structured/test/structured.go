package test

import (
	"github.com/tidepool-org/platform/blob"
	blobStoreStructured "github.com/tidepool-org/platform/blob/store/structured"
	cryptoTest "github.com/tidepool-org/platform/crypto/test"
	netTest "github.com/tidepool-org/platform/net/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomCreate() *blobStoreStructured.Create {
	datum := blobStoreStructured.NewCreate()
	datum.MediaType = pointer.FromString(netTest.RandomMediaType())
	return datum
}

func RandomUpdate() *blobStoreStructured.Update {
	datum := blobStoreStructured.NewUpdate()
	datum.DigestMD5 = pointer.FromString(cryptoTest.RandomBase64EncodedMD5Hash())
	datum.MediaType = pointer.FromString(netTest.RandomMediaType())
	datum.Size = pointer.FromInt(test.RandomIntFromRange(1, 100*1024*1024))
	datum.Status = pointer.FromString(test.RandomStringFromArray(blob.Statuses()))
	return datum
}
